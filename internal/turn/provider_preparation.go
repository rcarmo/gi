package turn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

// providerPreparation belongs to one turn owner. A retry replaces its request
// inputs from committed sources; no provider closure or tool call is replayed.
// Keeping the resolved choices here also carries a retry's model/thinking into
// the subsequent tool phase and iterations without rewriting admission history.
type providerPreparation struct {
	initialized               bool
	startCaptured             bool
	retryAttempt              int
	lastToolFailureSig        string
	repeatedToolFailureCount  int
	model, thinking           string
	cfg                       config.RuntimeConfig
	startPrompt, startMessage string
	startMessages, startTools json.RawMessage
	startRegistered           map[string]bool
	plan                      *sectionPlan
}

func (p *providerPreparation) initialize(ctx context.Context, s *store.Store, turnID, model string, cfg config.RuntimeConfig) error {
	if p.initialized {
		return nil
	}
	rec, err := s.GetTurn(ctx, turnID)
	if err != nil {
		return err
	}
	p.model, p.cfg = model, cfg
	if rec.Metadata["selected_thinking_model"] == model {
		p.thinking, _ = rec.Metadata["selected_thinking_level"].(string)
	}
	p.initialized = true
	return nil
}

func (r *sessionRunner) refreshProviderPreparation(ctx context.Context, s *store.Store, turnID, sessionID, agentID string, iter, attempt int, p *providerPreparation, convCtx *goai.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.CheckProviderRetryRunning(ctx, sessionID, turnID); err != nil {
		return err
	}
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	cfg, err := r.engine.runtimeCfg.ForGeneration()
	if err != nil {
		return err
	}
	policy := cfg.Retry.Policy()
	if !policy.Enabled || attempt > policy.MaxRetries {
		return fmt.Errorf("retry disabled or exhausted under current policy")
	}
	return r.rebuildProviderPreparationWithSession(ctx, s, turnID, sessionID, agentID, iter, p, convCtx, cfg, session)
}

func (r *sessionRunner) rebuildProviderPreparation(ctx context.Context, s *store.Store, turnID, sessionID, agentID string, iter int, p *providerPreparation, convCtx *goai.Context, cfg config.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	return r.rebuildProviderPreparationWithSession(ctx, s, turnID, sessionID, agentID, iter, p, convCtx, cfg, session)
}

func (r *sessionRunner) rebuildProviderPreparationWithSession(ctx context.Context, s *store.Store, turnID, sessionID, agentID string, iter int, p *providerPreparation, convCtx *goai.Context, cfg config.RuntimeConfig, session *store.Session) error {
	p.cfg = cfg
	p.model = inference.SessionModel(session.State, inference.SessionModelChoice{Model: p.model}).Label()
	if p.model == "" {
		return fmt.Errorf("retry has no current model")
	}
	p.thinking = inference.CapturedSessionThinking(session, p.model)
	if p.thinking == "" {
		p.thinking, _ = inference.EffectiveThinking(p.model, p.cfg.DefaultThinkingLevel)
	}
	fresh, err := r.buildAgentContext(ctx, s, turnID, sessionID, p.cfg, p)
	if err != nil {
		return err
	}
	// Lifecycle and queue/steering hooks are not rerun. Context and request hooks
	// are generation preparation and receive fresh committed inputs instead.
	if err := r.applyAgentStartPreparation(p, fresh); err != nil {
		return err
	}
	r.applyGenerationContextHook(ctx, sessionID, turnID, p.model, agentID, iter, fresh)
	snapshot, err := s.ContextSnapshot(ctx, sessionID)
	if err != nil {
		return err
	}
	if err = r.compactSnapshot(ctx, sessionID, turnID, p.model, agentID, fresh, snapshot, false, p); err != nil {
		return err
	}
	*convCtx = *fresh
	return nil
}

// Retain once-only lifecycle policy, not the failed provider's request. A
// message replacement remains an explicit turn policy over rebuilt history;
// registered tool definitions are rebound to the current registry on retry.
func (r *sessionRunner) captureAgentStartPreparation(p *providerPreparation, resp HookResponse) error {
	p.startPrompt, p.startMessage = resp.SystemPrompt, resp.Message
	var err error
	if resp.Messages != nil {
		p.startMessages, err = json.Marshal(resp.Messages)
		if err != nil {
			return err
		}
	}
	if resp.Tools != nil {
		p.startTools, err = json.Marshal(resp.Tools)
		if err != nil {
			return err
		}
		p.startRegistered = map[string]bool{}
		for _, tool := range resp.Tools {
			_, p.startRegistered[tool.Name] = r.engine.tools.GetRegistered(tool.Name)
		}
	}
	p.startCaptured = true
	return nil
}

func (p *providerPreparation) lifecyclePolicy() *store.ProviderLifecyclePolicy {
	if !p.startCaptured {
		return nil
	}
	return (&store.ProviderLifecyclePolicy{Version: 1, SystemPrompt: p.startPrompt, Message: p.startMessage, Messages: p.startMessages, Tools: p.startTools, RegisteredTools: p.startRegistered}).Clone()
}

func (r *sessionRunner) applyAgentStartPreparation(p *providerPreparation, conv *goai.Context) error {
	if p.startPrompt != "" {
		conv.SystemPrompt = p.startPrompt
	}
	if p.startMessages != nil {
		decoder := json.NewDecoder(bytes.NewReader(p.startMessages))
		decoder.UseNumber()
		if err := decoder.Decode(&conv.Messages); err != nil {
			return err
		}
	}
	if p.startTools != nil {
		var tools []goai.Tool
		if err := json.Unmarshal(p.startTools, &tools); err != nil {
			return err
		}
		conv.Tools = []goai.Tool{}
		for _, tool := range tools {
			if registered, ok := r.engine.tools.GetRegistered(tool.Name); ok {
				conv.Tools = append(conv.Tools, registered.Definition())
			} else if !p.startRegistered[tool.Name] {
				conv.Tools = append(conv.Tools, tool)
			}
		}
	}
	if strings.TrimSpace(p.startMessage) != "" {
		conv.Messages = append([]goai.Message{goai.UserMessage(p.startMessage)}, conv.Messages...)
	}
	return nil
}

func (r *sessionRunner) prepareProviderRequestContext(ctx context.Context, sessionID, turnID, model, agentID string, iter int, convCtx *goai.Context, plan *sectionPlan) (*goai.Context, error) {
	requestCtx := &goai.Context{SystemPrompt: convCtx.SystemPrompt, Messages: append([]goai.Message(nil), convCtx.Messages...), Tools: append([]goai.Tool(nil), convCtx.Tools...)}
	if plan != nil {
		requestCtx.SystemPrompt, requestCtx.Messages = plan.apply(requestCtx.SystemPrompt, requestCtx.Messages, inference.SupportsMidConversationSystemMessages(model))
	}
	resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookBeforeProviderRequest, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, SystemPrompt: convCtx.SystemPrompt, Messages: convCtx.Messages, Tools: convCtx.Tools, Payload: map[string]any{"model": model, "messages": len(convCtx.Messages), "tools": len(convCtx.Tools), "stage": "context"}})
	if err != nil {
		log.Printf("hook before_provider_request error: %v", err)
		return requestCtx, nil
	}
	if err = hookAbortFromResponse(resp, "aborted before provider request by hook"); err != nil {
		return nil, err
	}
	if resp.SystemPrompt != "" {
		convCtx.SystemPrompt = resp.SystemPrompt
		requestCtx.SystemPrompt = resp.SystemPrompt
	}
	if resp.Messages != nil {
		convCtx.Messages = resp.Messages
		requestCtx.Messages = append([]goai.Message(nil), resp.Messages...)
	}
	if resp.Tools != nil {
		convCtx.Tools = resp.Tools
		requestCtx.Tools = append([]goai.Tool(nil), resp.Tools...)
	}
	if strings.TrimSpace(resp.Message) != "" {
		requestCtx.Messages = append([]goai.Message{goai.UserMessage(resp.Message)}, requestCtx.Messages...)
	}
	return requestCtx, nil
}

func (r *sessionRunner) applyGenerationContextHook(ctx context.Context, sessionID, turnID, model, agentID string, iter int, convCtx *goai.Context) {
	resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookContext, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, SystemPrompt: convCtx.SystemPrompt, Messages: convCtx.Messages, Tools: convCtx.Tools})
	if err != nil {
		log.Printf("hook context error: %v", err)
		return
	}
	if resp.SystemPrompt != "" {
		convCtx.SystemPrompt = resp.SystemPrompt
	}
	if resp.Messages != nil {
		convCtx.Messages = resp.Messages
	}
	if resp.Tools != nil {
		convCtx.Tools = resp.Tools
	}
}
