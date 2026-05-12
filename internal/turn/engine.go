package turn

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/connectivity"
	"github.com/rcarmo/gi/internal/peering"
	"github.com/rcarmo/gi/internal/routing"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
)

type Engine struct {
	store                        *store.Store
	systemPrompt                 string
	routeResolver                *routing.RouteResolver
	modelRouter                  *routing.Router
	runtimeCfg                   config.RuntimeConfig
	hooks                        *HookRegistry
	tools                        *ToolRegistry
	connectivity                 *connectivity.Registry
	topics                       *topics.Bus
	peering                      *peering.Manager
	bgCtx                        context.Context
	bgCancel                     context.CancelFunc
	extensions                   []ExtensionInfo
	extensionsMu                 sync.RWMutex
	sessions                     sync.Map // sessionID -> *sessionRunner
	subs                         map[string]map[chan map[string]any]bool
	subsMu                       sync.Mutex
	beforeSetupHook              func(context.Context, string, string)
	beforeSetupErrorHook         func(context.Context, string, string) error
	beforeCreateSubTurnErrorHook func(context.Context, string, string) error
	beforeLaunchClaimHook        func(context.Context, string, string)
	beforeCleanupNextWorkHook    func(context.Context, string)
}

type sharedSessionCoord struct {
	mu      sync.Mutex
	current *runningTurn
}

var sharedSessionCoords sync.Map // sessionCoordKey -> *sharedSessionCoord

type sessionCoordKey struct {
	store     *store.Store
	sessionID string
}

type sessionRunner struct {
	*sharedSessionCoord
	store  *store.Store
	engine *Engine
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

type DirectInput struct {
	Kind          string         `json:"kind"`
	SessionID     string         `json:"session_id,omitempty"`
	SessionKey    string         `json:"session_key,omitempty"`
	TargetAgentID string         `json:"target_agent_id,omitempty"`
	Prompt        string         `json:"prompt,omitempty"`
	Intent        string         `json:"intent,omitempty"`
	Model         string         `json:"model,omitempty"`
	ParentTurnID  string         `json:"parent_turn_id,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	Origin        DirectOrigin   `json:"origin,omitempty"`
}

type DirectOrigin struct {
	SourceKind string `json:"source_kind,omitempty"`
	SourceID   string `json:"source_id,omitempty"`
	Role       string `json:"role,omitempty"`
	Label      string `json:"label,omitempty"`
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
	cfg.Hooks = applyHookDefaultsCompat(cfg.Hooks)
	bgCtx, bgCancel := context.WithCancel(context.Background())
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
		bgCtx:         bgCtx,
		bgCancel:      bgCancel,
		subs:          map[string]map[chan map[string]any]bool{},
	}
	e.registerDefaultTools()
	e.startTopicBridge()
	if e.store != nil {
		if _, err := e.recoverInterruptedTurns(e.backgroundContext(), ""); err != nil {
			log.Printf("turn recovery: startup scan failed: %v", err)
		}
	}
	return e
}

func (e *Engine) Close() error {
	if e == nil {
		return nil
	}
	if e.bgCancel != nil {
		e.bgCancel()
	}
	if e.peering != nil {
		return e.peering.Close()
	}
	return nil
}

func (e *Engine) backgroundContext() context.Context {
	if e != nil && e.bgCtx != nil {
		return e.bgCtx
	}
	return context.Background()
}

func (e *Engine) SubmitPrompt(ctx context.Context, in RunInput) (*SubmitResult, error) {
	if in.Intent == "" {
		in.Intent = "prompt"
	}
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
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
	var parentTurn *store.Turn
	subTurnDepth := 0
	subTurnMaxDepth := defaultSubTurnMaxDepth
	subTurnMaxConcurrency := defaultSubTurnMaxConcurrency
	subTurnDeliveryMode := "sync"
	subTurnCritical := false
	subTurnToolsRestricted := false
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
		subTurnCritical = boolValueOr(in.Metadata["subturn_critical"], boolValue(in.Metadata["critical"]))
		if subTurnMaxDepth <= 0 {
			subTurnMaxDepth = defaultSubTurnMaxDepth
		}
		if subTurnMaxConcurrency <= 0 {
			subTurnMaxConcurrency = defaultSubTurnMaxConcurrency
		}
		parentTurn, err = e.store.GetTurn(ctx, in.ParentTurnID)
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
		metadata["subturn_critical"] = subTurnCritical
	}
	effectiveTools, restrictedTools, err := e.resolveEffectiveToolNames(parentTurn, in.Metadata)
	if err != nil {
		return nil, err
	}
	subTurnToolsRestricted = restrictedTools
	for k, v := range in.Metadata {
		if k == "effective_tools" || k == "subturn_tools_restricted" {
			continue
		}
		metadata[k] = v
	}
	metadata["effective_tools"] = effectiveTools
	if in.ParentTurnID != "" {
		metadata["subturn_tools_restricted"] = subTurnToolsRestricted
	}
	if _, err := e.store.CreateTurnWithStatus(ctx, turnID, in.SessionID, "queued", in.Prompt, metadata); err != nil {
		return nil, err
	}
	durableCtx := e.backgroundContext()
	if in.ParentTurnID != "" && parentSessionID != "" {
		subturnMetadata := map[string]any{"intent": in.Intent, "model": in.Model, "depth": subTurnDepth, "max_depth": subTurnMaxDepth, "max_concurrency": subTurnMaxConcurrency, "delivery_mode": subTurnDeliveryMode, "subturn_critical": subTurnCritical, "effective_tools": effectiveTools, "subturn_tools_restricted": subTurnToolsRestricted}
		if hook := e.beforeCreateSubTurnErrorHook; hook != nil {
			if err := hook(durableCtx, in.ParentTurnID, turnID); err != nil {
				warnStore("rollback turn after create subturn hook failure", e.store.DeleteTurn(durableCtx, turnID))
				return nil, err
			}
		}
		if _, err := e.store.CreateSubTurn(durableCtx, in.ParentTurnID, parentSessionID, turnID, in.SessionID, subTurnDeliveryMode, subTurnDepth, subturnMetadata); err != nil {
			warnStore("rollback turn after create subturn failure", e.store.DeleteTurn(durableCtx, turnID))
			return nil, err
		}
		e.broadcast(parentSessionID, map[string]any{
			"type":             "subturn_created",
			"chat_jid":         "gi:" + parentSessionID,
			"parent_turn_id":   in.ParentTurnID,
			"parent_session":   parentSessionID,
			"child_turn_id":    turnID,
			"child_session":    in.SessionID,
			"depth":            subTurnDepth,
			"delivery_mode":    subTurnDeliveryMode,
			"critical":         subTurnCritical,
			"restricted_tools": subTurnToolsRestricted,
			"max_depth":        subturnMetadata["max_depth"],
			"max_concurrency":  subturnMetadata["max_concurrency"],
		})
		if in.SessionID != parentSessionID {
			e.broadcast(in.SessionID, map[string]any{
				"type":             "subturn_created",
				"chat_jid":         "gi:" + in.SessionID,
				"parent_turn_id":   in.ParentTurnID,
				"parent_session":   parentSessionID,
				"child_turn_id":    turnID,
				"child_session":    in.SessionID,
				"depth":            subTurnDepth,
				"delivery_mode":    subTurnDeliveryMode,
				"critical":         subTurnCritical,
				"restricted_tools": subTurnToolsRestricted,
			})
		}
	}
	if err := e.recordRouteDecision(durableCtx, in.SessionID, turnID, metadata); err != nil {
		// Non-fatal: routing decisions are an orchestration artifact.
		log.Printf("orchestration: route decision persist failed: %v", err)
	}
	if !queued {
		launched, err := e.launchTurnLocked(durableCtx, runner, in.SessionID, turnID)
		if err != nil {
			return nil, err
		}
		if !launched {
			if steeringResult, steered, err := e.convertLaunchConflictToSteering(durableCtx, turnID, in); err != nil {
				return nil, err
			} else if steered {
				return steeringResult, nil
			}
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
		warnStore("add queued prompt system message", e.store.AddMessage(durableCtx, store.NowID("msg"), in.SessionID, "system", fmt.Sprintf("Queued prompt: %s", in.Prompt), queuePayload))
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
	warnStore("append turn.submitted event", e.store.AppendTurnEvent(durableCtx, turnID, in.SessionID, "turn.submitted", submittedPayload))
	warnStore("sync queue count after submit", e.store.SyncSessionQueueCount(durableCtx, in.SessionID))
	warnStore("touch session model after submit", e.store.TouchSessionState(durableCtx, in.SessionID, map[string]any{"model": in.Model}))
	return &SubmitResult{TurnID: turnID, SessionID: in.SessionID, Status: status, Queued: queued}, nil
}

func (e *Engine) CancelTurn(ctx context.Context, sessionID, turnID string) error {
	turn, err := e.store.GetTurn(ctx, turnID)
	if err != nil {
		return err
	}
	turnSessionID := turn.SessionID
	runner := e.runner(turnSessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if runner.current != nil && runner.current.turnID == turnID {
		if err := e.store.AppendTurnEvent(ctx, turnID, turnSessionID, "turn.cancelling", map[string]any{"phase": "cancel", "checkpoint": true}); err != nil {
			return err
		}
		if err := e.store.UpdateTurnStatusAndPhase(ctx, turnID, "cancelling", "cancelling"); err != nil {
			return err
		}
		runner.emitTurnStateHook(ctx, turnSessionID, turnID, "", "", "cancelling", "cancelling", map[string]any{"reason": "cancel_requested"})
		runner.current.cancel()
		runner.current.cmdMu.Lock()
		if runner.current.cmd != nil && runner.current.cmd.Process != nil {
			if err := syscall.Kill(-runner.current.cmd.Process.Pid, syscall.SIGKILL); err != nil {
				warnStore("kill running command process group", err)
			}
		}
		runner.current.cmdMu.Unlock()
		return nil
	}
	if turn.Status == "queued" {
		if err := e.store.UpdateTurnStatusAndPhase(ctx, turnID, "cancelled", "aborted"); err != nil {
			return err
		}
		if err := e.store.MarkTurnFinished(ctx, turnID); err != nil {
			return err
		}
		runner.emitTurnStateHook(ctx, turnSessionID, turnID, "", "", "cancelled", "aborted", map[string]any{"reason": "queued_cancel"})
		warnStore("sync queue count after queued cancel", e.store.SyncSessionQueueCount(ctx, turnSessionID))
		return e.store.AppendTurnEvent(ctx, turnID, turnSessionID, "turn.cancelled", map[string]any{"phase": "cancel", "checkpoint": true, "queued": true})
	}
	return fmt.Errorf("turn not cancellable")
}

func (e *Engine) convertLaunchConflictToSteering(ctx context.Context, turnID string, in RunInput) (*SubmitResult, bool, error) {
	activeTurnID, _, err := e.store.GetSessionActiveTurn(ctx, in.SessionID)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	res, err := e.submitSteeringPrompt(ctx, in.SessionID, activeTurnID, in)
	if err != nil {
		// Keep the already-persisted queued turn as a fallback rather than dropping the prompt.
		return nil, false, nil
	}
	if err := e.store.DeleteTurn(ctx, turnID); err != nil {
		return nil, false, err
	}
	return res, true, nil
}

func (e *Engine) runner(sessionID string) *sessionRunner {
	if v, ok := e.sessions.Load(sessionID); ok {
		return v.(*sessionRunner)
	}
	coordKey := sessionCoordKey{store: e.store, sessionID: sessionID}
	coordVal, _ := sharedSessionCoords.LoadOrStore(coordKey, &sharedSessionCoord{})
	runner := &sessionRunner{sharedSessionCoord: coordVal.(*sharedSessionCoord), store: e.store, engine: e}
	actual, _ := e.sessions.LoadOrStore(sessionID, runner)
	return actual.(*sessionRunner)
}

func (e *Engine) launchTurnLocked(ctx context.Context, runner *sessionRunner, sessionID, turnID string) (bool, error) {
	if hook := e.beforeLaunchClaimHook; hook != nil {
		hook(ctx, sessionID, turnID)
	}
	claimToken := turnID
	claimed, err := e.store.ClaimSessionActiveTurn(ctx, sessionID, turnID, "runner", claimToken)
	if err != nil {
		return false, err
	}
	if !claimed {
		return false, nil
	}
	runCtx, cancel := context.WithCancel(e.backgroundContext())
	active := &runningTurn{turnID: turnID, cancel: cancel}
	runner.current = active
	releaseClaim := func() {
		warnStore("release active claim after launch failure", e.store.ReleaseSessionActiveTurn(e.backgroundContext(), sessionID, claimToken))
		if runner.current == active {
			runner.current = nil
		}
		cancel()
	}
	if err := e.store.MarkTurnClaimed(ctx, turnID, "runner"); err != nil {
		releaseClaim()
		return false, err
	}
	if err := e.store.UpdateTurnStatusAndPhase(ctx, turnID, "running", "setup"); err != nil {
		releaseClaim()
		return false, err
	}
	if err := e.store.TouchSessionState(ctx, sessionID, map[string]any{"active_turn_id": turnID, "status": "running"}); err != nil {
		warnStore("rollback turn status to queued", e.store.UpdateTurnStatusAndPhase(ctx, turnID, "queued", "queued"))
		releaseClaim()
		return false, err
	}
	go runner.runTurn(e.store, sessionID, turnID, runCtx, cancel, active)
	return true, nil
}

func (r *sessionRunner) runTurn(s *store.Store, sessionID, turnID string, ctx context.Context, cancel context.CancelFunc, active *runningTurn) {
	claimToken := turnID
	defer cancel()
	defer func() {
		r.cleanupTurnRun(sessionID, claimToken, active)
	}()

	run, err := r.setupTurnRun(ctx, s, sessionID, turnID)
	if err != nil {
		if ctx.Err() != nil || isCancellationError(err) {
			r.finishTurn(s, turnID, sessionID, "", "", "cancelled", "Turn cancelled", "")
		} else {
			r.finishTurn(s, turnID, sessionID, "", "", "failed", fmt.Sprintf("Turn setup error: %v", err), "setup_error")
		}
		return
	}
	sessionID = run.sessionID
	go r.heartbeatActiveTurn(ctx, sessionID, claimToken)
	r.runPreparedTurn(ctx, s, run)
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
	if _, exists := m[ch]; !exists {
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
	defer e.subsMu.Unlock()
	m, ok := e.subs[sessionID]
	if !ok {
		return
	}
	for ch := range m {
		select {
		case ch <- ev:
		default:
		}
	}
}
