package turn

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/compaction"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

const SidePromptDefaultSystem = "Answer the user briefly and directly. This is a side conversation that should not affect the main chat until explicitly injected."

type SidePromptResult struct {
	Status   string `json:"status"`
	Result   string `json:"result"`
	Thinking string `json:"thinking"`
	Model    string `json:"model"`
}

// RunSidePrompt snapshots provider-safe context without acquiring the main-turn
// queue, emitting lifecycle hooks, executing tools or writing conversation state.
func (e *Engine) RunSidePrompt(ctx context.Context, sessionID, prompt, system string, emit func(string, string)) (*SidePromptResult, error) {
	if e == nil || e.store == nil {
		return nil, errors.New("side prompt engine unavailable")
	}
	e.sideMu.Lock()
	if e.closing.Load() {
		e.sideMu.Unlock()
		return nil, errors.New("side prompt engine unavailable")
	}
	e.sideRuns.Add(1)
	e.sideMu.Unlock()
	defer e.sideRuns.Done()
	prompt = strings.TrimSpace(prompt)
	if prompt == "" || len(prompt) > 64*1024 || len(system) > 16*1024 {
		return nil, errors.New("invalid side prompt")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(e.backgroundContext(), cancel)
	defer stop()
	session, err := e.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	choice := inference.SessionModel(session.State, inference.SessionModelChoice{Model: e.runtimeCfg.DefaultModel, Provider: e.runtimeCfg.DefaultProvider})
	model := choice.Label()
	if choice.Model == "bootstrap" || choice.Model == "test-model" || choice.Model == "" {
		return nil, errors.New("side prompt requires an inference model")
	}
	snapshot, err := e.store.ContextSnapshot(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	bytes := len(snapshot.Summary) + len(prompt) + len(system)
	for _, message := range snapshot.Messages {
		bytes += len(message.Content)
		// Bound media before the shared projector loads/decompresses blobs.
		// Foreign and unsupported media are skipped by that projector.
		if message.Role == "user" {
			refs, _ := store.NormalizeMediaReferences(message.Payload["media"])
			for _, ref := range refs {
				media, err := e.store.GetMedia(ctx, ref.MediaID)
				if err != nil || media.SessionID != sessionID || !providerSafeImageTypes[strings.ToLower(strings.TrimSpace(media.ContentType))] || media.OriginalSize > 10<<20 {
					continue
				}
				bytes += base64.StdEncoding.EncodedLen(media.OriginalSize)
				if bytes > 2*1024*1024 {
					return nil, errors.New("side prompt context exceeds limit")
				}
			}
		}
	}
	if bytes > 2*1024*1024 {
		return nil, errors.New("side prompt context exceeds limit")
	}
	projection := &sessionRunner{store: e.store, engine: e}
	messages := projection.projectContextSnapshot(ctx, sessionID, snapshot)
	messages = append(messages, goai.UserMessage(prompt))
	if strings.TrimSpace(system) == "" {
		system = SidePromptDefaultSystem
	}
	context := &goai.Context{SystemPrompt: system, Messages: messages} // No tools.
	window := inference.ResolveModelContextWindow(choice.Provider, choice.Model)
	// Image bytes are included in the input bound and conservatively reserved
	// in the estimate; the text estimator alone cannot account for media.
	imageTokens := 0
	bytes = len(system) + len(prompt) + len(snapshot.Summary)
	for _, message := range snapshot.Messages {
		bytes += len(message.Content)
	}
	for _, message := range messages {
		for _, block := range message.Content {
			bytes += len(block.Data)
			imageTokens += len(block.Data) / 4
		}
	}
	if bytes > 2*1024*1024 {
		return nil, errors.New("side prompt context exceeds limit")
	}
	if window <= 0 {
		return nil, fmt.Errorf("side prompt context window unavailable for model %s", model)
	}
	if compaction.EstimateMessagesTokens(messages)+compaction.EstimateTokens(system)+imageTokens+1024 > window {
		return nil, fmt.Errorf("side prompt context does not fit model %s", model)
	}
	thinking := inference.CapturedSessionThinking(session, model)
	if thinking == "" {
		thinking, _ = inference.EffectiveThinking(model, e.runtimeCfg.DefaultThinkingLevel)
	}
	if thinking == "off" {
		thinking = ""
	}
	var answer, thought strings.Builder
	var outputErr error
	output := func(event map[string]any) {
		if outputErr != nil {
			return
		}
		kind, _ := event["type"].(string)
		delta, _ := event["delta"].(string)
		if kind != "text_delta" && kind != "thinking_delta" {
			return
		}
		if answer.Len()+thought.Len()+len(delta) > 512*1024 {
			outputErr = errors.New("side prompt output exceeds limit")
			cancel()
			return
		}
		if kind == "text_delta" {
			answer.WriteString(delta)
		} else {
			thought.WriteString(delta)
		}
		if emit != nil {
			emit(kind, delta)
		}
	}
	result, err := streamWithToolsWithHooks(ctx, model, context, output, &inference.StreamHooks{Thinking: thinking, Transport: goai.Transport(e.runtimeCfg.Transport), MaxTokens: 1024, CacheRetention: goai.CacheRetentionNone})
	if outputErr != nil {
		return nil, outputErr
	}
	if err != nil {
		return nil, err
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if result == nil || result.Message == nil {
		return nil, errors.New("side prompt finished without a response")
	}
	if result.Message.StopReason == goai.StopReasonError || result.Message.StopReason == goai.StopReasonAborted {
		return nil, errors.New("side prompt provider failed")
	}
	if answer.Len() == 0 {
		if len(result.Text) > 512*1024 {
			return nil, errors.New("side prompt output exceeds limit")
		}
		answer.WriteString(result.Text)
	}
	if thought.Len() == 0 {
		for _, block := range result.Message.Content {
			if block.Type == "thinking" {
				thought.WriteString(block.Thinking)
			}
		}
	}
	if answer.Len()+thought.Len() > 512*1024 {
		return nil, errors.New("side prompt output exceeds limit")
	}
	return &SidePromptResult{Status: "success", Result: answer.String(), Thinking: thought.String(), Model: model}, nil
}
