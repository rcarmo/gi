package turn

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/connectivity"
	"github.com/rcarmo/gi/internal/peering"
	"github.com/rcarmo/gi/internal/routing"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
)

type Engine struct {
	store         *store.Store
	systemPrompt  string
	routeResolver *routing.RouteResolver
	modelRouter   *routing.Router
	runtimeCfg    config.RuntimeConfig
	hooks         *HookRegistry
	tools         *ToolRegistry
	connectivity  *connectivity.Registry
	topics        *topics.Bus
	peering       *peering.Manager
	extensions    []ExtensionInfo
	extensionsMu  sync.RWMutex
	sessions      sync.Map // sessionID -> *sessionRunner
	subs          map[string]map[chan map[string]any]bool
	subsMu        sync.Mutex
}

type sessionRunner struct {
	mu      sync.Mutex
	store   *store.Store
	engine  *Engine
	current *runningTurn
}

type runningTurn struct {
	turnID string
	cancel context.CancelFunc
	cmd    *exec.Cmd
	cmdMu  sync.Mutex
}

type RunInput struct {
	SessionID    string
	Prompt       string
	Intent       string
	Model        string
	ParentTurnID string
	Metadata     map[string]any
}

type SubmitResult struct {
	TurnID          string `json:"turn_id"`
	SessionID       string `json:"session_id"`
	Status          string `json:"status"`
	Queued          bool   `json:"queued"`
	SourceSessionID string `json:"source_session_id,omitempty"`
	TargetAgentID   string `json:"target_agent_id,omitempty"`
	Routed          bool   `json:"routed,omitempty"`
	CreatedSession  bool   `json:"created_session,omitempty"`
}

type Summary struct {
	TurnID    string            `json:"turn_id"`
	SessionID string            `json:"session_id"`
	Status    string            `json:"status"`
	Assistant string            `json:"assistant"`
	Events    []store.TurnEvent `json:"events"`
}

const (
	defaultSubTurnMaxDepth       = 8
	defaultSubTurnMaxConcurrency = 4
)

func New(s *store.Store) *Engine {
	cfg := config.RuntimeConfig{
		DefaultModel: "bootstrap",
		Agents:       routing.AgentsConfig{List: []routing.AgentConfig{{ID: "agent", Default: true, Model: "bootstrap"}}},
		Session:      routing.SessionConfig{Dimensions: []string{"chat"}},
	}
	return NewWithRuntimeConfig(s, cfg, "")
}

func NewWithSystemPrompt(s *store.Store, systemPrompt string) *Engine {
	cfg := config.RuntimeConfig{
		DefaultModel: "bootstrap",
		Agents:       routing.AgentsConfig{List: []routing.AgentConfig{{ID: "agent", Default: true, Model: "bootstrap"}}},
		Session:      routing.SessionConfig{Dimensions: []string{"chat"}},
	}
	return NewWithRuntimeConfig(s, cfg, systemPrompt)
}

func NewWithRuntimeConfig(s *store.Store, cfg config.RuntimeConfig, systemPrompt string) *Engine {
	if strings.TrimSpace(cfg.DefaultModel) == "" {
		cfg.DefaultModel = "bootstrap"
	}
	if len(cfg.Agents.List) == 0 {
		cfg.Agents.List = []routing.AgentConfig{{ID: "agent", Default: true, Model: cfg.DefaultModel}}
	}
	if len(cfg.Session.Dimensions) == 0 {
		cfg.Session.Dimensions = []string{"chat"}
	}
	e := &Engine{
		store:         s,
		systemPrompt:  systemPrompt,
		routeResolver: routing.NewRouteResolver(cfg.Agents, cfg.Session),
		modelRouter:   routing.NewRouter(cfg.Routing),
		runtimeCfg:    cfg,
		hooks:         NewHookRegistry(),
		tools:         NewToolRegistry(),
		connectivity:  connectivity.NewRegistry(),
		topics:        topics.NewBus(),
		peering:       peering.NewManager(cfg.Peering, cfg.WorkspaceRoot),
		subs:          map[string]map[chan map[string]any]bool{},
	}
	e.registerDefaultTools()
	e.startTopicBridge()
	if e.store != nil {
		if _, err := e.recoverInterruptedTurns(context.Background(), ""); err != nil {
			log.Printf("turn recovery: startup scan failed: %v", err)
		}
	}
	return e
}

func (e *Engine) SubmitPrompt(ctx context.Context, in RunInput) (*SubmitResult, error) {
	if in.Intent == "" {
		in.Intent = "prompt"
	}
	recovered, err := e.recoverInterruptedTurns(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}
	if recovered {
		if err := e.startNextQueuedTurn(ctx, in.SessionID); err != nil {
			return nil, err
		}
	}
	turnID := store.NowID("turn")
	runner := e.runner(in.SessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if activeTurnID, _, err := e.store.GetSessionActiveTurn(ctx, in.SessionID); err == nil {
		return e.submitSteeringPrompt(ctx, in.SessionID, activeTurnID, in)
	} else if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	queued := false
	count, err := e.store.CountQueuedTurns(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}
	queued = count > 0
	metadata := map[string]any{"intent": in.Intent, "model": in.Model}
	parentSessionID := ""
	subTurnDepth := 0
	subTurnMaxDepth := defaultSubTurnMaxDepth
	subTurnMaxConcurrency := defaultSubTurnMaxConcurrency
	subTurnDeliveryMode := "sync"
	if in.ParentTurnID != "" {
		metadata["parent_turn_id"] = in.ParentTurnID
		if v, ok := in.Metadata["subturn_max_depth"]; ok {
			subTurnMaxDepth = intValueOr(v, defaultSubTurnMaxDepth)
		}
		if v, ok := in.Metadata["subturn_max_concurrency"]; ok {
			subTurnMaxConcurrency = intValueOr(v, defaultSubTurnMaxConcurrency)
		}
		if modeRaw, ok := in.Metadata["subturn_delivery_mode"]; ok {
			mode, err := normalizeSubTurnDeliveryMode(stringValue(modeRaw, "sync"))
			if err != nil {
				return nil, err
			}
			subTurnDeliveryMode = mode
		}
		if subTurnMaxDepth <= 0 {
			subTurnMaxDepth = defaultSubTurnMaxDepth
		}
		if subTurnMaxConcurrency <= 0 {
			subTurnMaxConcurrency = defaultSubTurnMaxConcurrency
		}
		parentTurn, err := e.store.GetTurn(ctx, in.ParentTurnID)
		if err != nil {
			return nil, fmt.Errorf("resolve parent turn: %w", err)
		}
		parentSessionID = parentTurn.SessionID
		subTurnDepth = intValueOr(parentTurn.Metadata["subturn_depth"], 0) + 1
		if subTurnDepth > subTurnMaxDepth {
			return nil, fmt.Errorf("subturn depth limit exceeded: depth=%d max=%d", subTurnDepth, subTurnMaxDepth)
		}
		runningChildren, err := e.store.CountRunningSubTurnsByParent(ctx, in.ParentTurnID)
		if err != nil {
			return nil, err
		}
		if runningChildren >= subTurnMaxConcurrency {
			return nil, fmt.Errorf("subturn concurrency limit exceeded: running=%d max=%d", runningChildren, subTurnMaxConcurrency)
		}
		metadata["subturn_depth"] = subTurnDepth
		metadata["subturn_parent_turn_id"] = in.ParentTurnID
		metadata["subturn_max_depth"] = subTurnMaxDepth
		metadata["subturn_max_concurrency"] = subTurnMaxConcurrency
		metadata["subturn_delivery_mode"] = subTurnDeliveryMode
	}
	for k, v := range in.Metadata {
		metadata[k] = v
	}
	if _, err := e.store.CreateTurnWithStatus(ctx, turnID, in.SessionID, "queued", in.Prompt, metadata); err != nil {
		return nil, err
	}
	if in.ParentTurnID != "" && parentSessionID != "" {
		subturnMetadata := map[string]any{"intent": in.Intent, "model": in.Model, "depth": subTurnDepth, "max_depth": subTurnMaxDepth, "max_concurrency": subTurnMaxConcurrency, "delivery_mode": subTurnDeliveryMode}
		if _, err := e.store.CreateSubTurn(ctx, in.ParentTurnID, parentSessionID, turnID, in.SessionID, subTurnDeliveryMode, subTurnDepth, subturnMetadata); err != nil {
			return nil, err
		}
		e.broadcast(parentSessionID, map[string]any{
			"type":            "subturn_created",
			"chat_jid":        "gi:" + parentSessionID,
			"parent_turn_id":  in.ParentTurnID,
			"parent_session":  parentSessionID,
			"child_turn_id":   turnID,
			"child_session":   in.SessionID,
			"depth":           subTurnDepth,
			"delivery_mode":   subTurnDeliveryMode,
			"max_depth":       subturnMetadata["max_depth"],
			"max_concurrency": subturnMetadata["max_concurrency"],
		})
		if in.SessionID != parentSessionID {
			e.broadcast(in.SessionID, map[string]any{
				"type":           "subturn_created",
				"chat_jid":       "gi:" + in.SessionID,
				"parent_turn_id": in.ParentTurnID,
				"parent_session": parentSessionID,
				"child_turn_id":  turnID,
				"child_session":  in.SessionID,
				"depth":          subTurnDepth,
				"delivery_mode":  subTurnDeliveryMode,
			})
		}
	}
	if err := e.recordRouteDecision(ctx, in.SessionID, turnID, metadata); err != nil {
		// Non-fatal: routing decisions are an orchestration artifact.
		log.Printf("orchestration: route decision persist failed: %v", err)
	}
	if !queued {
		launched, err := e.launchTurnLocked(ctx, runner, in.SessionID, turnID)
		if err != nil {
			return nil, err
		}
		queued = !launched
	}
	status := "running"
	if queued {
		status = "queued"
		queuePayload := map[string]any{"kind": "queue", "turn_id": turnID, "intent": in.Intent}
		for _, key := range []string{"source_session_id", "source_agent_id", "target_agent_id", "routed_from_prompt"} {
			if value, ok := metadata[key]; ok {
				queuePayload[key] = value
			}
		}
		if err := e.store.AddMessage(ctx, store.NowID("msg"), in.SessionID, "system", fmt.Sprintf("Queued prompt: %s", in.Prompt), queuePayload); err != nil {
			return nil, err
		}
	}
	submittedPayload := map[string]any{"phase": "queue", "intent": in.Intent, "queued": queued, "checkpoint": true}
	for _, key := range []string{"source_session_id", "source_agent_id", "target_agent_id", "routed_from_prompt"} {
		if value, ok := metadata[key]; ok {
			submittedPayload[key] = value
		}
	}
	if routeMatchedBy := metadata["route_matched_by"]; routeMatchedBy != nil {
		submittedPayload["route_matched_by"] = routeMatchedBy
	}
	if err := e.store.AppendTurnEvent(ctx, turnID, in.SessionID, "turn.submitted", submittedPayload); err != nil {
		return nil, err
	}
	_ = e.store.SyncSessionQueueCount(ctx, in.SessionID)
	_ = e.store.TouchSessionState(ctx, in.SessionID, map[string]any{"model": in.Model})
	return &SubmitResult{TurnID: turnID, SessionID: in.SessionID, Status: status, Queued: queued}, nil
}

func (e *Engine) SubmitPromptRouted(ctx context.Context, in RunInput) (*SubmitResult, error) {
	source, err := e.store.GetSession(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}
	targetAgentID, body, directed := parseDirectedPrompt(in.Prompt)
	promptBody := in.Prompt
	mentioned := false
	if directed {
		if body == "" {
			return nil, fmt.Errorf("directed prompt requires content after @%s", targetAgentID)
		}
		promptBody = body
		mentioned = true
	}
	inbound := routing.InboundContext{
		Channel:   sessionChannel(source),
		Account:   sessionAccount(source),
		ChatType:  "direct",
		ChatID:    source.ID,
		SenderID:  "user",
		Mentioned: mentioned,
		Prompt:    promptBody,
	}
	route := e.routeResolver.ResolveRoute(inbound)
	if directed && targetAgentID != "" {
		route.AgentID = routing.NormalizeAgentID(targetAgentID)
		route.MatchedBy = "mention"
	}
	target, created, err := e.ResolveOrCreateRouteSession(ctx, source, route, inbound)
	if err != nil {
		return nil, err
	}
	if target.ID != source.ID {
		return e.submitPeerRoutedPrompt(ctx, source, target, route, promptBody, in.Intent, in.Model, created, directed, in.ParentTurnID)
	}
	in.SessionID = target.ID
	in.Prompt = promptBody
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	in.Metadata["route_mode"] = "prompt"
	in.Metadata["route_matched_by"] = route.MatchedBy
	in.Metadata["target_agent_id"] = route.AgentID
	in.Metadata["target_session_id"] = target.ID
	in.Metadata["source_agent_id"] = sessionAgentID(source)
	if route.MatchedBy != "" {
		in.Metadata["routing_policy"] = route.MatchedBy
	}
	in.Metadata["requested_agent_id"] = route.AgentID
	in.Metadata["source_session_id"] = source.ID
	in.Metadata["route_created_session"] = created
	in.Metadata["routing_enabled"] = true
	return e.SubmitPrompt(ctx, in)
}

func (e *Engine) SubmitPeerMessage(ctx context.Context, sourceSessionID, targetAgentID, content, intent, model, parentTurnID string) (*SubmitResult, error) {
	source, err := e.store.GetSession(ctx, sourceSessionID)
	if err != nil {
		return nil, err
	}
	inbound := routing.InboundContext{Channel: sessionChannel(source), Account: sessionAccount(source), ChatType: "direct", ChatID: source.ID, SenderID: sessionAgentID(source), Mentioned: true, Prompt: content}
	route := e.routeResolver.ResolveRoute(inbound)
	route.AgentID = routing.NormalizeAgentID(targetAgentID)
	route.MatchedBy = "peer-message"
	target, created, err := e.ResolveOrCreateRouteSession(ctx, source, route, inbound)
	if err != nil {
		return nil, err
	}
	return e.submitPeerRoutedPrompt(ctx, source, target, route, content, intent, model, created, true, parentTurnID)
}

func (e *Engine) ResolveOrCreatePeerSession(ctx context.Context, sourceSessionID, targetAgentID string) (*store.Session, bool, error) {
	source, err := e.store.GetSession(ctx, sourceSessionID)
	if err != nil {
		return nil, false, err
	}
	inbound := routing.InboundContext{Channel: sessionChannel(source), Account: sessionAccount(source), ChatType: "direct", ChatID: source.ID, SenderID: sessionAgentID(source), Mentioned: true}
	route := e.routeResolver.ResolveRoute(inbound)
	route.AgentID = routing.NormalizeAgentID(targetAgentID)
	route.MatchedBy = "peer-session"
	return e.ResolveOrCreateRouteSession(ctx, source, route, inbound)
}

func (e *Engine) CancelTurn(ctx context.Context, sessionID, turnID string) error {
	runner := e.runner(sessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if runner.current != nil && runner.current.turnID == turnID {
		_ = e.store.AppendTurnEvent(ctx, turnID, sessionID, "turn.cancelling", map[string]any{"phase": "cancel", "checkpoint": true})
		_ = e.store.UpdateTurnStatusAndPhase(ctx, turnID, "cancelling", "cancelling")
		runner.current.cancel()
		runner.current.cmdMu.Lock()
		if runner.current.cmd != nil && runner.current.cmd.Process != nil {
			_ = syscall.Kill(-runner.current.cmd.Process.Pid, syscall.SIGKILL)
		}
		runner.current.cmdMu.Unlock()
		return nil
	}
	turn, err := e.store.GetTurn(ctx, turnID)
	if err != nil {
		return err
	}
	if turn.Status == "queued" {
		if err := e.store.UpdateTurnStatusAndPhase(ctx, turnID, "cancelled", "aborted"); err != nil {
			return err
		}
		_ = e.store.SyncSessionQueueCount(ctx, sessionID)
		return e.store.AppendTurnEvent(ctx, turnID, sessionID, "turn.cancelled", map[string]any{"phase": "cancel", "checkpoint": true, "queued": true})
	}
	return fmt.Errorf("turn not cancellable")
}

func (e *Engine) runner(sessionID string) *sessionRunner {
	v, _ := e.sessions.LoadOrStore(sessionID, &sessionRunner{store: e.store, engine: e})
	return v.(*sessionRunner)
}

func (e *Engine) launchTurnLocked(ctx context.Context, runner *sessionRunner, sessionID, turnID string) (bool, error) {
	claimed, err := e.store.ClaimSessionActiveTurn(ctx, sessionID, turnID, "runner", turnID)
	if err != nil {
		return false, err
	}
	if !claimed {
		return false, nil
	}
	_ = e.store.MarkTurnClaimed(ctx, turnID, "runner")
	_ = e.store.UpdateTurnStatusAndPhase(ctx, turnID, "running", "setup")
	_ = e.store.TouchSessionState(ctx, sessionID, map[string]any{"active_turn_id": turnID, "status": "running"})
	go runner.runTurn(e.store, sessionID, turnID)
	return true, nil
}

func (r *sessionRunner) runTurn(s *store.Store, sessionID, turnID string) {
	ctx, cancel := context.WithCancel(context.Background())
	claimToken := turnID
	r.mu.Lock()
	r.current = &runningTurn{turnID: turnID, cancel: cancel}
	r.mu.Unlock()
	defer func() {
		cancel()
		_ = s.ReleaseSessionActiveTurn(context.Background(), sessionID, claimToken)
		_ = s.SyncSessionQueueCount(context.Background(), sessionID)
		r.mu.Lock()
		r.current = nil
		r.mu.Unlock()
		if err := r.engine.startNextQueuedTurn(context.Background(), sessionID); err != nil {
			log.Printf("turn coordination: launch queued turn failed: %v", err)
			return
		}
		if _, _, err := s.GetSessionActiveTurn(context.Background(), sessionID); err == sql.ErrNoRows {
			if _, err := r.engine.continueQueuedSteering(context.Background(), sessionID); err != nil {
				log.Printf("steering continuation: %v", err)
			}
		}
	}()

	turnRec, err := s.GetTurn(ctx, turnID)
	if err != nil {
		return
	}
	if strings.TrimSpace(sessionID) == "" {
		sessionID = turnRec.SessionID
	}
	go r.heartbeatActiveTurn(ctx, sessionID, claimToken)
	initialSteering := steeringMessagesFromMetadata(turnRec.Metadata)
	prompt := turnRec.Prompt
	if strings.TrimSpace(prompt) == "" && len(initialSteering) > 0 {
		prompt = steeringPromptForShell(initialSteering)
	}
	intent := stringValue(turnRec.Metadata["intent"], "prompt")
	model := stringValue(turnRec.Metadata["model"], "bootstrap")
	agentID := "agent"
	if sess, err := s.GetSession(ctx, sessionID); err == nil && sess.Scope != nil && sess.Scope.AgentID != "" {
		agentID = sess.Scope.AgentID
	}
	// Only run the model router when no explicit model was requested
	// (i.e., the turn metadata model matches the agent's default model).
	agentModel := r.engine.modelForAgent(agentID)
	if strings.TrimSpace(model) == "" {
		model = agentModel
	}
	if model == agentModel {
		history, _ := s.ListMessages(ctx, sessionID)
		routingHistory := make([]routing.HistoryMessage, 0, len(history))
		for _, msg := range history {
			routingHistory = append(routingHistory, routing.HistoryMessage{Payload: msg.Payload})
		}
		selected, usedLight, score := r.engine.modelRouter.SelectModel(prompt, routingHistory, agentModel)
		if usedLight && strings.TrimSpace(selected) != "" {
			model = selected
			turnRec.Metadata["route_model_score"] = score
			turnRec.Metadata["route_used_light_model"] = usedLight
		}
	}
	_ = s.TouchSessionState(ctx, sessionID, map[string]any{"active_turn_id": turnID, "model": model, "status": "running"})
	userPayload := map[string]any{"kind": "chat", "intent": intent, "turn_id": turnID}
	for _, key := range []string{"source_session_id", "source_agent_id", "target_agent_id", "routed_from_prompt"} {
		if value, ok := turnRec.Metadata[key]; ok {
			userPayload[key] = value
		}
	}
	if strings.TrimSpace(prompt) != "" && len(initialSteering) == 0 {
		_ = s.AddMessage(ctx, store.NowID("msg"), sessionID, "user", prompt, userPayload)
	}
	startedPayload := map[string]any{"phase": "turn", "prompt": prompt, "intent": intent, "model": model, "checkpoint": true}
	for _, key := range []string{"source_session_id", "source_agent_id", "target_agent_id", "routed_from_prompt", "parent_turn_id", "route_mode", "route_matched_by"} {
		if value, ok := turnRec.Metadata[key]; ok {
			startedPayload[key] = value
		}
	}
	_ = s.AppendTurnEvent(ctx, turnID, sessionID, "turn.started", startedPayload)

	// Try LLM inference if model is not the bootstrap stub
	if model != "bootstrap" && model != "test-model" && model != "" {
		r.runAgentLoop(ctx, s, turnID, sessionID, model, agentID, initialSteering)
		return
	}

	// Fallback: shell stub for bootstrap/test mode
	if len(initialSteering) > 0 {
		r.persistSteeringMessages(ctx, sessionID, turnID, initialSteering)
	}
	_ = s.AppendTurnEvent(ctx, turnID, sessionID, "tool.started", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "command": []string{"sh", "-lc", "printf 'Gi received: %s' \"$GI_PROMPT\""}})

	out, runErr, cancelled := runShell(ctx, prompt, func(cmd *exec.Cmd) {
		r.mu.Lock()
		if r.current != nil && r.current.turnID == turnID {
			r.current.cmdMu.Lock()
			r.current.cmd = cmd
			r.current.cmdMu.Unlock()
		}
		r.mu.Unlock()
	}, func(delta string) {
		if strings.TrimSpace(delta) == "" {
			return
		}
		r.engine.broadcast(sessionID, map[string]any{
			"type":     "agent_draft_delta",
			"chat_jid": "gi:" + sessionID,
			"delta":    delta,
			"turn_id":  turnID,
		})
	})
	if cancelled {
		if staged, stagedTurnID, err := r.engine.stageQueuedSteeringContinuation(context.Background(), sessionID); err != nil {
			log.Printf("steering final checkpoint: %v", err)
		} else if staged {
			_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "steering.final_checkpoint", map[string]any{"phase": "steering", "checkpoint": true, "staged_turn_id": stagedTurnID})
		}
		_ = s.AppendTurnEvent(ctx, turnID, sessionID, "turn.cancelled", map[string]any{"phase": "cancel", "checkpoint": true})
		_ = s.UpdateTurnStatus(context.Background(), turnID, "cancelled")
		r.publishSubTurnLifecycle(context.Background(), turnID, "cancelled")
		_ = s.AddMessage(context.Background(), store.NowID("msg"), sessionID, "system", "Turn cancelled", map[string]any{"kind": "status", "turn_id": turnID, "clipped": true})
		_ = s.TouchSessionState(context.Background(), sessionID, map[string]any{"status": "idle", "active_turn_id": nil})
		return
	}
	if runErr != nil {
		if staged, stagedTurnID, err := r.engine.stageQueuedSteeringContinuation(context.Background(), sessionID); err != nil {
			log.Printf("steering final checkpoint: %v", err)
		} else if staged {
			_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "steering.final_checkpoint", map[string]any{"phase": "steering", "checkpoint": true, "staged_turn_id": stagedTurnID})
		}
		markTurnFailure(s, turnID, sessionID, "shell_error", runErr.Error())
		_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "tool.failed", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "error": runErr.Error()})
		_ = s.UpdateTurnStatus(context.Background(), turnID, "failed")
		r.publishSubTurnLifecycle(context.Background(), turnID, "failed")
		_ = s.TouchSessionState(context.Background(), sessionID, map[string]any{"status": "idle", "active_turn_id": nil})
		return
	}
	if staged, stagedTurnID, err := r.engine.stageQueuedSteeringContinuation(context.Background(), sessionID); err != nil {
		log.Printf("steering final checkpoint: %v", err)
	} else if staged {
		_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "steering.final_checkpoint", map[string]any{"phase": "steering", "checkpoint": true, "staged_turn_id": stagedTurnID})
	}
	_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "tool.finished", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "output": out})
	msgID := store.NowID("msg")
	_ = s.AddMessage(context.Background(), msgID, sessionID, "assistant", out, map[string]any{"kind": "chat", "source": "shell", "turn_id": turnID, "agent_id": agentID})
	r.engine.broadcast(sessionID, map[string]any{
		"type": "new_post", "id": msgID, "chat_jid": "gi:" + sessionID,
		"content": out, "timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"sender": "agent", "is_bot_message": true,
		"data": map[string]any{"type": "agent_response", "content": out, "agent_id": agentID},
	})
	_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "turn.finished", map[string]any{"phase": "turn", "checkpoint": true, "status": "completed"})
	_ = s.UpdateTurnStatus(context.Background(), turnID, "completed")
	r.publishSubTurnLifecycle(context.Background(), turnID, "completed")
	_ = s.TouchSessionState(context.Background(), sessionID, map[string]any{"status": "idle", "active_turn_id": nil})
}

func (e *Engine) Summary(ctx context.Context, turnID string) (*Summary, error) {
	turnRec, err := e.store.GetTurn(ctx, turnID)
	if err != nil {
		return nil, err
	}
	events, err := e.store.ListTurnEvents(ctx, turnID)
	if err != nil {
		return nil, err
	}
	msgs, _ := e.store.ListMessages(ctx, turnRec.SessionID)
	assistant := ""
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "assistant" {
			assistant = msgs[i].Content
			break
		}
	}
	return &Summary{TurnID: turnID, SessionID: turnRec.SessionID, Status: turnRec.Status, Assistant: assistant, Events: events}, nil
}

func runShell(ctx context.Context, prompt string, onStart func(*exec.Cmd), onDelta func(string)) (string, error, bool) {
	cmd := exec.Command("sh", "-lc", "printf 'Gi received: %s' \"$GI_PROMPT\"")
	cmd.Env = append(cmd.Environ(), "GI_PROMPT="+prompt)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err, false
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", err, false
	}
	if err := cmd.Start(); err != nil {
		return "", err, false
	}
	if onStart != nil {
		onStart(cmd)
	}
	var stdout, stderr bytes.Buffer
	var readWG sync.WaitGroup
	readWG.Add(2)
	go func() {
		defer readWG.Done()
		buf := make([]byte, 128)
		for {
			n, readErr := stdoutPipe.Read(buf)
			if n > 0 {
				chunk := string(buf[:n])
				stdout.WriteString(chunk)
				if onDelta != nil {
					onDelta(chunk)
				}
			}
			if readErr != nil {
				return
			}
		}
	}()
	go func() {
		defer readWG.Done()
		_, _ = io.Copy(&stderr, stderrPipe)
	}()
	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()
	select {
	case <-ctx.Done():
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		<-waitCh
		readWG.Wait()
		return stdout.String(), nil, true
	case err := <-waitCh:
		readWG.Wait()
		if err != nil {
			if stderr.Len() > 0 {
				return "", fmt.Errorf("%w: %s", err, stderr.String()), false
			}
			return "", err, false
		}
		if stderr.Len() > 0 {
			return stdout.String(), fmt.Errorf("stderr: %s", stderr.String()), false
		}
		return stdout.String(), nil, false
	}
}

func (e *Engine) ResolveOrCreateRouteSession(ctx context.Context, source *store.Session, route routing.ResolvedRoute, inbound routing.InboundContext) (*store.Session, bool, error) {
	if source == nil {
		return nil, false, fmt.Errorf("missing source session")
	}
	if sessionAgentID(source) == route.AgentID {
		return source, false, nil
	}
	alloc := gisession.AllocateRouteSession(gisession.AllocationInput{AgentID: route.AgentID, Context: inbound, SessionPolicy: route.SessionPolicy})
	if existing, err := e.store.FindSessionByAllocation(ctx, alloc); err == nil {
		return existing, false, nil
	} else if err != nil && err != sql.ErrNoRows {
		return nil, false, err
	}
	if existing, err := e.store.FindChildSessionByParentAndAgent(ctx, source.ID, route.AgentID); err == nil {
		return existing, false, nil
	} else if err != nil && err != sql.ErrNoRows {
		return nil, false, err
	}
	if strings.TrimSpace(source.ParentSessionID) != "" {
		if existing, err := e.store.FindChildSessionByParentAndAgent(ctx, source.ParentSessionID, route.AgentID); err == nil {
			return existing, false, nil
		} else if err != nil && err != sql.ErrNoRows {
			return nil, false, err
		}
	}
	state := map[string]any{"status": "idle", "queue_count": 0, "model": e.modelForAgent(route.AgentID), "provider": e.runtimeCfg.DefaultProvider, "thinking_level": e.runtimeCfg.DefaultThinkingLevel}
	cloned, err := e.store.CreateSessionWithMetadata(ctx, store.NowID("session"), source.ID, "@"+route.AgentID, state, &alloc.Scope, alloc.SessionAliases)
	if err != nil {
		return nil, false, err
	}
	messages, err := e.store.ListMessages(ctx, source.ID)
	if err == nil {
		for _, msg := range messages {
			payload := map[string]any{}
			for k, v := range msg.Payload {
				payload[k] = v
			}
			payload["forked_from_message_id"] = msg.ID
			_ = e.store.AddMessage(ctx, store.NowID("msg"), cloned.ID, msg.Role, msg.Content, payload)
		}
	}
	_ = e.store.AddMessage(ctx, store.NowID("msg"), cloned.ID, "system", fmt.Sprintf("Forked from @%s", sessionAgentID(source)), map[string]any{"kind": "fork", "source_session_id": source.ID, "source_agent_id": sessionAgentID(source), "route_matched_by": route.MatchedBy, "clipped": true})
	return cloned, true, nil
}

func (e *Engine) submitPeerRoutedPrompt(ctx context.Context, source, target *store.Session, route routing.ResolvedRoute, content, intent, model string, created, directed bool, parentTurnID string) (*SubmitResult, error) {
	sourceAgentID := sessionAgentID(source)
	routingContent := fmt.Sprintf("↪ routed to @%s: %s", route.AgentID, content)
	routingPayload := map[string]any{"kind": "routing", "target_agent_id": route.AgentID, "target_session_id": target.ID, "source_agent_id": sourceAgentID, "source_session_id": source.ID, "route_matched_by": route.MatchedBy, "clipped": true}
	_ = e.store.AddMessage(ctx, store.NowID("msg"), source.ID, "system", routingContent, routingPayload)
	metadata := map[string]any{
		"source_session_id":     source.ID,
		"source_agent_id":       sourceAgentID,
		"target_session_id":     target.ID,
		"target_agent_id":       route.AgentID,
		"requested_agent_id":    route.AgentID,
		"routing_policy":        route.MatchedBy,
		"route_matched_by":      route.MatchedBy,
		"routed_from_prompt":    directed,
		"route_mode":            "peer-message",
		"route_created_session": created,
		"routing_enabled":       true,
	}
	if strings.TrimSpace(parentTurnID) != "" {
		metadata["parent_turn_id"] = parentTurnID
	}
	result, err := e.SubmitPrompt(ctx, RunInput{SessionID: target.ID, Prompt: content, Intent: intent, Model: model, ParentTurnID: parentTurnID, Metadata: metadata})
	if err != nil {
		return nil, err
	}
	result.SourceSessionID = source.ID
	result.TargetAgentID = route.AgentID
	result.Routed = true
	result.CreatedSession = created
	return result, nil
}

func (e *Engine) modelForAgent(agentID string) string {
	agentID = routing.NormalizeAgentID(agentID)
	for _, agent := range e.runtimeCfg.Agents.List {
		if routing.NormalizeAgentID(agent.ID) == agentID {
			if strings.TrimSpace(agent.Model) != "" {
				return agent.Model
			}
		}
	}
	if strings.TrimSpace(e.runtimeCfg.DefaultModel) != "" {
		return e.runtimeCfg.DefaultModel
	}
	return "bootstrap"
}

func parseDirectedPrompt(prompt string) (string, string, bool) {
	trimmed := strings.TrimSpace(prompt)
	if !strings.HasPrefix(trimmed, "@") {
		return "", prompt, false
	}
	rest := trimmed[1:]
	end := strings.IndexAny(rest, " \t\r\n:")
	if end < 0 {
		return normalizeAgentID(rest), "", true
	}
	target := normalizeAgentID(rest[:end])
	body := strings.TrimSpace(rest[end:])
	body = strings.TrimPrefix(body, ":")
	body = strings.TrimSpace(body)
	return target, body, target != ""
}

func sessionAgentID(sess *store.Session) string {
	if sess != nil && sess.Scope != nil && sess.Scope.AgentID != "" {
		return sess.Scope.AgentID
	}
	return "agent"
}

func sessionChannel(sess *store.Session) string {
	if sess != nil && sess.Scope != nil && strings.TrimSpace(sess.Scope.Channel) != "" {
		return sess.Scope.Channel
	}
	return "gi"
}

func sessionAccount(sess *store.Session) string {
	if sess != nil && sess.Scope != nil && strings.TrimSpace(sess.Scope.Account) != "" {
		return sess.Scope.Account
	}
	return "default"
}

func normalizeAgentID(v string) string {
	return strings.TrimSpace(strings.TrimPrefix(strings.ToLower(v), "@"))
}

func stringValue(v any, fallback string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return fallback
}

func (e *Engine) recordRouteDecision(ctx context.Context, sourceSessionID, turnID string, metadata map[string]any) error {
	sourceSession := stringValue(metadata["source_session_id"], sourceSessionID)
	targetSession := stringValue(metadata["target_session_id"], "")
	targetAgentID := stringValue(metadata["target_agent_id"], "")
	if targetAgentID == "" {
		return nil
	}
	routeMode := stringValue(metadata["route_mode"], stringValue(metadata["mode"], "prompt"))
	sourceAgent := stringValue(metadata["source_agent_id"], "")
	if sourceAgent == "" {
		sess, err := e.store.GetSession(context.Background(), sourceSession)
		if err == nil {
			sourceAgent = sessionAgentID(sess)
		}
	}
	routingPolicy := stringValue(metadata["routing_policy"], "")
	matchedBy := stringValue(metadata["route_matched_by"], "")
	requestedAgent := stringValue(metadata["requested_agent_id"], "")
	if requestedAgent == "" {
		requestedAgent = targetAgentID
	}
	if targetSession == "" {
		targetSession = sourceSession
	}
	decision := store.RouteEvent{
		TurnID:         turnID,
		SourceSession:  sourceSession,
		TargetSession:  targetSession,
		SourceAgentID:  sourceAgent,
		TargetAgentID:  targetAgentID,
		Mode:           routeMode,
		MatchedBy:      matchedBy,
		RoutingPolicy:  routingPolicy,
		RequestedAgent: requestedAgent,
		Metadata: map[string]any{
			"routed_from_prompt": boolValue(metadata["routed_from_prompt"]),
			"created_session":    boolValue(metadata["route_created_session"]),
			"routing_enabled":    boolValueOr(metadata["routing_enabled"], true),
		},
	}
	for k, v := range metadata {
		if k == "routed_from_prompt" || k == "route_created_session" || k == "routing_enabled" || k == "route_matched_by" || k == "routing_policy" || k == "target_agent_id" || k == "source_agent_id" || k == "target_session_id" || k == "route_mode" || k == "requested_agent_id" {
			continue
		}
		decision.Metadata[k] = v
	}
	if !boolValueOr(metadata["routing_enabled"], true) {
		return nil
	}
	if _, err := e.store.RecordRouteEvent(ctx, decision); err != nil {
		return err
	}
	e.broadcast(sourceSession, map[string]any{
		"type":            "routing_decision",
		"chat_jid":        "gi:" + sourceSession,
		"turn_id":         turnID,
		"source_session":  sourceSession,
		"target_session":  targetSession,
		"source_agent_id": sourceAgent,
		"target_agent_id": targetAgentID,
		"mode":            routeMode,
		"matched_by":      matchedBy,
		"created_session": boolValue(metadata["route_created_session"]),
	})
	if targetSession != "" && targetSession != sourceSession {
		e.broadcast(targetSession, map[string]any{
			"type":            "routing_incoming",
			"chat_jid":        "gi:" + targetSession,
			"source_session":  sourceSession,
			"target_session":  targetSession,
			"source_agent_id": sourceAgent,
			"target_agent_id": targetAgentID,
			"turn_id":         turnID,
			"mode":            routeMode,
		})
	}
	return nil
}

func boolValue(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	if s, ok := v.(string); ok {
		s = strings.ToLower(strings.TrimSpace(s))
		return s == "true" || s == "1" || s == "yes"
	}
	return false
}

func boolValueOr(v any, fallback bool) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	if v == nil {
		return fallback
	}
	if s, ok := v.(string); ok {
		s = strings.TrimSpace(strings.ToLower(s))
		if s == "" {
			return fallback
		}
		return s == "true" || s == "1" || s == "yes"
	}
	switch num := v.(type) {
	case float64:
		return num != 0
	case int:
		return num != 0
	case int64:
		return num != 0
	case int32:
		return num != 0
	case uint:
		return num != 0
	case uint64:
		return num != 0
	case uint32:
		return num != 0
	default:
		return fallback
	}
}

func intValueOr(v any, fallback int) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case int32:
		return int(n)
	case uint:
		return int(n)
	case uint64:
		return int(n)
	case uint32:
		return int(n)
	case float64:
		return int(n)
	case string:
		n = strings.TrimSpace(n)
		if n == "" {
			return fallback
		}
		if parsed, err := strconv.Atoi(n); err == nil {
			return parsed
		}
	}
	return fallback
}

func normalizeSubTurnDeliveryMode(mode string) (string, error) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		return "sync", nil
	}
	switch mode {
	case "sync", "async":
		return mode, nil
	default:
		return "", fmt.Errorf("invalid subturn delivery mode: %s", mode)
	}
}

func SortQueuedTurns(turns []store.Turn) {
	sort.SliceStable(turns, func(i, j int) bool { return turns[i].CreatedAt < turns[j].CreatedAt })
}

func (e *Engine) Subscribe(sessionID string) chan map[string]any {
	ch := make(chan map[string]any, 64)
	e.subsMu.Lock()
	defer e.subsMu.Unlock()
	if e.subs == nil {
		e.subs = map[string]map[chan map[string]any]bool{}
	}
	m, ok := e.subs[sessionID]
	if !ok {
		m = make(map[chan map[string]any]bool)
		e.subs[sessionID] = m
	}
	m[ch] = true
	return ch
}

func (e *Engine) Unsubscribe(sessionID string, ch chan map[string]any) {
	e.subsMu.Lock()
	defer e.subsMu.Unlock()
	m, ok := e.subs[sessionID]
	if !ok {
		return
	}
	delete(m, ch)
	if len(m) == 0 {
		delete(e.subs, sessionID)
	}
	close(ch)
}

func (e *Engine) broadcast(sessionID string, ev map[string]any) {
	e.publishTopicFromBroadcast(sessionID, ev)
	e.subsMu.Lock()
	m, ok := e.subs[sessionID]
	if !ok {
		e.subsMu.Unlock()
		return
	}
	chs := make([]chan map[string]any, 0, len(m))
	for ch := range m {
		chs = append(chs, ch)
	}
	e.subsMu.Unlock()
	for _, ch := range chs {
		select {
		case ch <- ev:
		default:
		}
	}
}
