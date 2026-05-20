package turn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/rcarmo/gi/internal/logutil"
	"github.com/rcarmo/gi/internal/tools"
	"log"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/connectivity"
	"github.com/rcarmo/gi/internal/peering"
	"github.com/rcarmo/gi/internal/routing"
	"github.com/rcarmo/gi/internal/routing/routedsession"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
	goai "github.com/rcarmo/go-ai"
)

type Engine struct {
	store                             *store.Store
	systemPrompt                      string
	routeResolver                     *routing.RouteResolver
	modelRouter                       *routing.Router
	runtimeCfg                        config.RuntimeConfig
	hooks                             *HookRegistry
	tools                             *tools.ToolRegistry
	connectivity                      *connectivity.Registry
	topics                            *topics.Bus
	peering                           *peering.Manager
	bgCtx                             context.Context
	bgCancel                          context.CancelFunc
	extensions                        []ExtensionInfo
	extensionsMu                      sync.RWMutex
	sessions                          sync.Map // sessionID -> *sessionRunner
	subs                              map[string]map[chan map[string]any]bool
	subsMu                            sync.Mutex
	beforeSetupHook                   func(context.Context, string, string)
	beforeSetupErrorHook              func(context.Context, string, string) error
	beforeCreateSubTurnErrorHook      func(context.Context, string, string) error
	beforeLaunchClaimHook             func(context.Context, string, string)
	beforeLaunchSessionStateErrorHook func(context.Context, string, string) error
	beforeCleanupNextWorkHook         func(context.Context, string)
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

type ExtensionInfo struct {
	Engine string `json:"engine"`
	Path   string `json:"path"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type HookInfo struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	ID     uint64 `json:"id"`
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
		Agents:       config.AgentsConfig{List: []config.AgentConfig{{ID: "agent", Default: true, Model: "bootstrap"}}},
		Session:      config.SessionConfig{Dimensions: []string{"chat"}},
	}
	return NewWithRuntimeConfig(s, cfg, "")
}

func NewWithSystemPrompt(s *store.Store, systemPrompt string) *Engine {
	cfg := config.RuntimeConfig{
		DefaultModel: "bootstrap",
		Agents:       config.AgentsConfig{List: []config.AgentConfig{{ID: "agent", Default: true, Model: "bootstrap"}}},
		Session:      config.SessionConfig{Dimensions: []string{"chat"}},
	}
	return NewWithRuntimeConfig(s, cfg, systemPrompt)
}

func NewWithRuntimeConfig(s *store.Store, cfg config.RuntimeConfig, systemPrompt string) *Engine {
	if strings.TrimSpace(cfg.DefaultModel) == "" {
		cfg.DefaultModel = "bootstrap"
	}
	if len(cfg.Agents.List) == 0 {
		cfg.Agents.List = []config.AgentConfig{{ID: "agent", Default: true, Model: cfg.DefaultModel}}
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
		tools:         tools.NewToolRegistry(),
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

func (e *Engine) ExtensionInfos() []ExtensionInfo {
	e.extensionsMu.RLock()
	defer e.extensionsMu.RUnlock()
	out := append([]ExtensionInfo(nil), e.extensions...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func (e *Engine) HookInfos() []HookInfo {
	return e.hooks.Infos()
}

func (e *Engine) recordExtension(info ExtensionInfo) {
	e.extensionsMu.Lock()
	e.extensions = append(e.extensions, info)
	e.extensionsMu.Unlock()
	envType := "notice"
	if info.Status == "failed" {
		envType = "error"
	}
	e.publishTopicEvent(topics.Envelope{
		Topic:  "extension." + info.Status,
		Source: "extension",
		Type:   envType,
		Payload: map[string]any{
			"engine": info.Engine,
			"path":   info.Path,
			"status": info.Status,
			"error":  info.Error,
		},
	})
}

func (e *Engine) RegisterTool(tool tools.RegisteredTool) error { return e.tools.Register(tool) }
func (e *Engine) SetActiveTools(names []string) error          { return e.tools.SetActive(names) }
func (e *Engine) ActiveTools() []string                        { return e.tools.ActiveNames() }
func (e *Engine) ResetActiveTools() {
	if err := e.tools.SetActive(nil); err != nil {
		log.Printf("reset active tools: %v", err)
	}
}
func (e *Engine) ToolEntries() []tools.RegisteredTool { return e.tools.AllEntries() }
func (e *Engine) toolDefs() []goai.Tool               { return e.tools.Definitions() }
func (e *Engine) ExecuteToolsMeta(args map[string]any) (string, error) {
	return e.executeToolsTool(args)
}
func (e *Engine) ExecuteToolByName(ctx context.Context, name, sessionID string, args map[string]any) (string, error) {
	tool, ok := e.tools.Get(name)
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return tool.Executor(ctx, tools.ToolRuntime{Store: e.store, SessionID: sessionID, WorkspaceRoot: e.runtimeCfg.WorkspaceRoot}, goai.ToolCall{Name: name, Arguments: args})
}
func (e *Engine) executeToolsTool(args map[string]any) (string, error) {
	return tools.ExecuteToolsTool(e.tools, args, e.SetActiveTools, e.ActiveTools, e.ResetActiveTools)
}

func (e *Engine) allRegisteredToolNames() []string {
	entries := e.tools.AllEntries()
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry.Name)
	}
	return tools.NormalizeToolNames(out)
}
func (e *Engine) defaultEffectiveToolNames() []string {
	return tools.NormalizeToolNames(e.ActiveTools())
}
func (e *Engine) resolveEffectiveToolNames(parentTurn *store.Turn, inputMetadata map[string]any) ([]string, bool, error) {
	var parent map[string]any
	if parentTurn != nil {
		parent = parentTurn.Metadata
	}
	return tools.ResolveEffectiveToolNames(parent, inputMetadata, e.defaultEffectiveToolNames(), e.allRegisteredToolNames())
}
func toolAllowedByMetadata(metadata map[string]any, toolName string) bool {
	return tools.ToolAllowedByMetadata(metadata, toolName)
}
func (e *Engine) toolDefsForMetadata(metadata map[string]any) []goai.Tool {
	return tools.ToolDefsForMetadata(metadata, e.tools.AllEntries(), e.toolDefs())
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

func (e *Engine) normalizeRunningSessionState(ctx context.Context, sessionID, activeTurnID string, syncQueue bool, overrideModel string) error {
	if e == nil || e.store == nil || strings.TrimSpace(sessionID) == "" || strings.TrimSpace(activeTurnID) == "" {
		return nil
	}
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	sessionState := map[string]any{"status": "running", "active_turn_id": activeTurnID}
	if turnRec, err := e.store.GetTurn(opCtx, activeTurnID); err == nil {
		if model := strings.TrimSpace(store.StringValue(turnRec.Metadata["model"], "")); model != "" {
			sessionState["model"] = model
		}
	}
	if model := strings.TrimSpace(overrideModel); model != "" {
		sessionState["model"] = model
	}
	if syncQueue {
		if err := e.store.SyncSessionQueueCount(opCtx, sessionID); err != nil {
			return fmt.Errorf("sync queue count on running session normalization: %w", err)
		}
	}
	if err := e.store.TouchSessionState(opCtx, sessionID, sessionState); err != nil {
		return fmt.Errorf("touch running session normalization: %w", err)
	}
	return nil
}

func (e *Engine) normalizeInactiveSessionState(ctx context.Context, sessionID, status, model string, syncQueue bool) error {
	if e == nil || e.store == nil || strings.TrimSpace(sessionID) == "" {
		return nil
	}
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	sessionState := map[string]any{"status": status, "active_turn_id": nil}
	if model = strings.TrimSpace(model); model != "" {
		sessionState["model"] = model
	}
	if syncQueue {
		if err := e.store.SyncSessionQueueCount(opCtx, sessionID); err != nil {
			return fmt.Errorf("sync queue count on inactive session normalization: %w", err)
		}
	}
	if err := e.store.TouchSessionState(opCtx, sessionID, sessionState); err != nil {
		return fmt.Errorf("touch inactive session normalization: %w", err)
	}
	return nil
}

// Connectivity returns the engine-wide connectivity registry. Routes are
// transport-neutral; web/socket adapters dispatch through this registry.
func (e *Engine) Connectivity() *connectivity.Registry { return e.connectivity }

func (e *Engine) PeeringStatus() peering.Status {
	if e.peering == nil {
		return peering.Status{Backend: "tsnet", State: "unavailable"}
	}
	return e.peering.Status()
}

func (e *Engine) SubmitPrompt(ctx context.Context, in RunInput) (*SubmitResult, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
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
	if activeTurnID, _, err := e.store.GetSessionActiveTurn(opCtx, in.SessionID); err == nil {
		return e.submitSteeringPrompt(ctx, in.SessionID, activeTurnID, in)
	} else if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	queued := false
	count, err := e.store.CountQueuedTurns(opCtx, in.SessionID)
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
			subTurnMaxDepth = store.IntValueOr(v, defaultSubTurnMaxDepth)
		}
		if v, ok := in.Metadata["subturn_max_concurrency"]; ok {
			subTurnMaxConcurrency = store.IntValueOr(v, defaultSubTurnMaxConcurrency)
		}
		if modeRaw, ok := in.Metadata["subturn_delivery_mode"]; ok {
			mode, err := store.NormalizeSubTurnDeliveryMode(store.StringValue(modeRaw, "sync"))
			if err != nil {
				return nil, err
			}
			subTurnDeliveryMode = mode
		}
		subTurnCritical = store.BoolValueOr(in.Metadata["subturn_critical"], store.BoolValue(in.Metadata["critical"]))
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
		subTurnDepth = store.IntValueOr(parentTurn.Metadata["subturn_depth"], 0) + 1
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
	if _, err := e.store.CreateTurnWithStatus(opCtx, turnID, in.SessionID, "queued", in.Prompt, metadata); err != nil {
		return nil, err
	}
	durableCtx := e.backgroundContext()
	if in.ParentTurnID != "" && parentSessionID != "" {
		subturnMetadata := map[string]any{"intent": in.Intent, "model": in.Model, "depth": subTurnDepth, "max_depth": subTurnMaxDepth, "max_concurrency": subTurnMaxConcurrency, "delivery_mode": subTurnDeliveryMode, "subturn_critical": subTurnCritical, "effective_tools": effectiveTools, "subturn_tools_restricted": subTurnToolsRestricted}
		if hook := e.beforeCreateSubTurnErrorHook; hook != nil {
			if err := hook(durableCtx, in.ParentTurnID, turnID); err != nil {
				logutil.WarnIfErr("rollback turn after create subturn hook failure", e.store.DeleteTurn(durableCtx, turnID))
				return nil, err
			}
		}
		if _, err := e.store.CreateSubTurn(durableCtx, in.ParentTurnID, parentSessionID, turnID, in.SessionID, subTurnDeliveryMode, subTurnDepth, subturnMetadata); err != nil {
			logutil.WarnIfErr("rollback turn after create subturn failure", e.store.DeleteTurn(durableCtx, turnID))
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
	if err := routing.RecordDecision(durableCtx, in.SessionID, turnID, metadata, routing.Options{SessionAgentID: e.store.SessionAgentID, RecordRouteEvent: func(ctx context.Context, event routing.Event) (int64, error) {
		return e.store.RecordRouteEvent(ctx, store.RouteEvent(event))
	}, PublishRuntimeRoutingEvent: e.PublishRuntimeRoutingEvent, Broadcast: e.broadcast}); err != nil {
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
		logutil.WarnIfErr("add queued prompt system message", e.store.AddMessage(durableCtx, store.NowID("msg"), in.SessionID, "system", fmt.Sprintf("Queued prompt: %s", in.Prompt), queuePayload))
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
	logutil.WarnIfErr("append turn.submitted event", e.store.AppendTurnEvent(durableCtx, turnID, in.SessionID, "turn.submitted", submittedPayload))
	e.PublishRuntimeTurnEvent("turn_submitted", in.SessionID, turnID, "", tools.FirstNonEmpty(status, "queued"), tools.FirstNonEmpty(status, "queued"), submittedPayload)
	logutil.WarnIfErr("sync queue count after submit", e.store.SyncSessionQueueCount(durableCtx, in.SessionID))
	sessionStateUpdate := map[string]any{}
	if model := strings.TrimSpace(in.Model); model != "" {
		sessionStateUpdate["model"] = model
	}
	if queued {
		sessionStateUpdate["status"] = "queued"
		sessionStateUpdate["active_turn_id"] = nil
	}
	if len(sessionStateUpdate) > 0 {
		logutil.WarnIfErr("touch session state after submit", e.store.TouchSessionState(durableCtx, in.SessionID, sessionStateUpdate))
	}
	return &SubmitResult{TurnID: turnID, SessionID: in.SessionID, Status: status, Queued: queued}, nil
}

func (e *Engine) CancelTurn(ctx context.Context, sessionID, turnID string) error {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	turn, err := e.store.GetTurn(opCtx, turnID)
	if err != nil {
		return err
	}
	turnSessionID := turn.SessionID
	if strings.TrimSpace(sessionID) != "" && turnSessionID != sessionID {
		return fmt.Errorf("turn %s does not belong to session %s", turnID, sessionID)
	}
	runner := e.runner(turnSessionID)
	agentID, model := runner.resolveTurnAgentAndModel(opCtx, e.store, turn, turnSessionID, turn.Prompt)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if runner.current != nil && runner.current.turnID == turnID {
		if err := e.store.AppendTurnEvent(opCtx, turnID, turnSessionID, "turn.cancelling", map[string]any{"phase": "cancel", "checkpoint": true, "reason": "cancel_requested", "status": "cancelling", "turn_phase": "cancelling", "failure_kind": ""}); err != nil {
			return err
		}
		if err := e.store.UpdateTurnStatusAndPhase(opCtx, turnID, "cancelling", "cancelling"); err != nil {
			return err
		}
		e.PublishRuntimeTurnEvent("turn_cancelling", turnSessionID, turnID, agentID, "cancelling", "cancelling", map[string]any{"reason": "cancel_requested", "failure_kind": ""})
		runner.emitTurnStateHook(opCtx, turnSessionID, turnID, agentID, model, "cancelling", "cancelling", map[string]any{"reason": "cancel_requested", "failure_kind": ""})
		runner.emitSessionStateHook(opCtx, turnSessionID, agentID, model, "running", map[string]any{"reason": "cancel_requested", "failure_kind": "", "active_turn_id": turnID, "turn_id": turnID, "turn_status": "cancelling", "turn_phase": "cancelling"})
		runner.current.cancel()
		runner.current.cmdMu.Lock()
		if runner.current.cmd != nil && runner.current.cmd.Process != nil {
			if err := syscall.Kill(-runner.current.cmd.Process.Pid, syscall.SIGKILL); err != nil {
				logutil.WarnIfErr("kill running command process group", err)
			}
		}
		runner.current.cmdMu.Unlock()
		return nil
	}
	if turn.Status == "queued" {
		if err := e.store.UpdateTurnStatusAndPhase(opCtx, turnID, "cancelled", "aborted"); err != nil {
			return err
		}
		if err := e.store.MarkTurnFinished(opCtx, turnID); err != nil {
			return err
		}
		runner.emitTurnStateHook(opCtx, turnSessionID, turnID, agentID, model, "cancelled", "aborted", map[string]any{"reason": "queued_cancel"})
		e.PublishRuntimeTurnEvent("turn_terminal", turnSessionID, turnID, agentID, "cancelled", "aborted", map[string]any{"reason": "queued_cancel", "failure_kind": ""})
		logutil.WarnIfErr("sync queue count after queued cancel", e.store.SyncSessionQueueCount(opCtx, turnSessionID))
		queueCount, err := e.store.CountQueuedTurns(opCtx, turnSessionID)
		if err != nil {
			return err
		}
		activeTurnID, _, err := e.store.GetSessionActiveTurn(opCtx, turnSessionID)
		hasActiveTurn := err == nil
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		sessionStatus := "idle"
		activeTurnValue := any(nil)
		if hasActiveTurn {
			sessionStatus = "running"
			activeTurnValue = activeTurnID
		} else if queueCount > 0 {
			sessionStatus = "queued"
		}
		logutil.WarnIfErr("touch session state after queued cancel", e.store.TouchSessionState(opCtx, turnSessionID, map[string]any{"status": sessionStatus, "active_turn_id": activeTurnValue}))
		runner.emitSessionStateHook(opCtx, turnSessionID, agentID, model, sessionStatus, map[string]any{"reason": "queued_cancel", "failure_kind": "", "turn_id": turnID, "turn_status": "cancelled", "turn_phase": "aborted", "active_turn_id": activeTurnValue})
		if sessionStatus == "idle" {
			e.PublishRuntimeSessionEvent("session_idle", turnSessionID, agentID, "idle", map[string]any{"reason": "turn_terminal", "active_turn_id": nil, "failure_kind": "", "turn_id": turnID, "turn_status": "cancelled", "turn_phase": "aborted", "model": model})
		}
		if err := e.store.AppendTurnEvent(opCtx, turnID, turnSessionID, "turn.cancelled", map[string]any{"phase": "cancel", "checkpoint": true, "queued": true, "reason": "queued_cancel", "status": "cancelled", "turn_phase": "aborted", "failure_kind": ""}); err != nil {
			return err
		}
		if sessionStatus == "running" {
			if err := e.store.AppendTurnEvent(opCtx, turnID, turnSessionID, "turn.cleanup_handoff", map[string]any{"phase": "cancel", "checkpoint": true, "reason": "queued_cancel", "handoff": "active_turn", "active_turn_id": activeTurnValue}); err != nil {
				return err
			}
		}
		if sessionStatus == "queued" {
			if err := e.store.AppendTurnEvent(opCtx, turnID, turnSessionID, "turn.cleanup_handoff", map[string]any{"phase": "cancel", "checkpoint": true, "reason": "queued_cancel", "handoff": "next_queued_turn"}); err != nil {
				return err
			}
			if _, err := e.startNextQueuedTurnLocked(opCtx, runner, turnSessionID); err != nil {
				return err
			}
		}
		if sessionStatus == "idle" {
			if err := e.store.AppendTurnEvent(opCtx, turnID, turnSessionID, "turn.cleanup_handoff", map[string]any{"phase": "cancel", "checkpoint": true, "reason": "queued_cancel", "handoff": "idle_session"}); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("turn not cancellable")
}

func (e *Engine) convertLaunchConflictToSteering(ctx context.Context, turnID string, in RunInput) (*SubmitResult, bool, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	activeTurnID, _, err := e.store.GetSessionActiveTurn(opCtx, in.SessionID)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	res, err := e.submitSteeringPrompt(opCtx, in.SessionID, activeTurnID, in)
	if err != nil {
		// Keep the already-persisted queued turn as a fallback rather than dropping the prompt.
		logutil.WarnIfErr("append launch-conflict fallback preserved event", e.store.AppendTurnEvent(opCtx, turnID, in.SessionID, "turn.cleanup_handoff", map[string]any{"phase": "launch", "checkpoint": true, "reason": "launch_claim_conflict", "handoff": "queued_fallback_preserved", "active_turn_id": activeTurnID}))
		return nil, false, nil
	}
	logutil.WarnIfErr("append launch-conflict steering handoff event", e.store.AppendTurnEvent(opCtx, activeTurnID, in.SessionID, "turn.cleanup_handoff", map[string]any{"phase": "launch", "checkpoint": true, "reason": "launch_claim_conflict", "handoff": "active_turn_steering", "replaced_turn_id": turnID}))
	if err := e.store.DeleteTurn(opCtx, turnID); err != nil {
		log.Printf("turn coordination: delete transient queued turn after steering fallback failed: %v", err)
	}
	if err := e.normalizeRunningSessionState(opCtx, in.SessionID, activeTurnID, true, in.Model); err != nil {
		return nil, false, err
	}
	return res, true, nil
}

func (r *sessionRunner) resolveTurnIdentityForFinalize(ctx context.Context, s *store.Store, sessionID, turnID string) (string, string) {
	opCtx := store.CoordinationContext(ctx, r.engine.backgroundContext())
	turnRec, err := s.GetTurn(opCtx, turnID)
	if err != nil {
		return "", ""
	}
	if strings.TrimSpace(sessionID) == "" {
		sessionID = turnRec.SessionID
	}
	return r.resolveTurnAgentAndModel(opCtx, s, turnRec, sessionID, turnRec.Prompt)
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
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if hook := e.beforeLaunchClaimHook; hook != nil {
		hook(opCtx, sessionID, turnID)
	}
	claimToken := turnID
	claimed, err := e.store.ClaimSessionActiveTurn(opCtx, sessionID, turnID, "runner", claimToken)
	if err != nil {
		return false, err
	}
	if !claimed {
		return false, nil
	}
	runCtx, cancel := context.WithCancel(e.backgroundContext())
	active := &runningTurn{turnID: turnID, cancel: cancel}
	runner.current = active
	claimedTurn := false
	releaseClaim := func(restoreQueued bool) error {
		var cleanupErrs []error
		if restoreQueued {
			if err := e.store.UpdateTurnStatusAndPhase(e.backgroundContext(), turnID, "queued", "queued"); err != nil {
				cleanupErrs = append(cleanupErrs, fmt.Errorf("rollback turn status to queued: %w", err))
			}
			if claimedTurn {
				if err := e.store.ResetTurnClaim(e.backgroundContext(), turnID); err != nil {
					cleanupErrs = append(cleanupErrs, fmt.Errorf("reset turn claim after launch failure: %w", err))
				}
			}
			if err := e.store.SyncSessionQueueCount(e.backgroundContext(), sessionID); err != nil {
				cleanupErrs = append(cleanupErrs, fmt.Errorf("sync queue count after launch rollback: %w", err))
			}
			if err := e.store.TouchSessionState(e.backgroundContext(), sessionID, map[string]any{"status": "queued", "active_turn_id": nil}); err != nil {
				cleanupErrs = append(cleanupErrs, fmt.Errorf("touch session queued after launch rollback: %w", err))
			}
		}
		if err := e.store.ReleaseSessionActiveTurn(e.backgroundContext(), sessionID, claimToken); err != nil {
			cleanupErrs = append(cleanupErrs, fmt.Errorf("release active claim after launch failure: %w", err))
		}
		if runner.current == active {
			runner.current = nil
		}
		cancel()
		return errors.Join(cleanupErrs...)
	}
	if err := e.store.MarkTurnClaimed(opCtx, turnID, "runner"); err != nil {
		if cleanupErr := releaseClaim(false); cleanupErr != nil {
			return false, errors.Join(err, cleanupErr)
		}
		return false, err
	}
	claimedTurn = true
	if err := e.store.UpdateTurnStatusAndPhase(opCtx, turnID, "running", "setup"); err != nil {
		if cleanupErr := releaseClaim(true); cleanupErr != nil {
			return false, errors.Join(err, cleanupErr)
		}
		return false, err
	}
	if hook := e.beforeLaunchSessionStateErrorHook; hook != nil {
		if err := hook(opCtx, sessionID, turnID); err != nil {
			if cleanupErr := releaseClaim(true); cleanupErr != nil {
				return false, errors.Join(err, cleanupErr)
			}
			return false, err
		}
	}
	sessionState := map[string]any{"active_turn_id": turnID, "status": "running"}
	if turnRec, turnErr := e.store.GetTurn(opCtx, turnID); turnErr == nil {
		if model := strings.TrimSpace(store.StringValue(turnRec.Metadata["model"], "")); model != "" {
			sessionState["model"] = model
		}
	}
	if err := e.store.TouchSessionState(opCtx, sessionID, sessionState); err != nil {
		if cleanupErr := releaseClaim(true); cleanupErr != nil {
			return false, errors.Join(err, cleanupErr)
		}
		return false, err
	}
	logutil.WarnIfErr("sync queue count after launch", e.store.SyncSessionQueueCount(opCtx, sessionID))
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
		agentID, model := r.resolveTurnIdentityForFinalize(ctx, s, sessionID, turnID)
		if ctx.Err() != nil || isCancellationError(err) {
			r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled", "")
		} else {
			r.finishTurn(s, turnID, sessionID, agentID, model, "failed", fmt.Sprintf("Turn setup error: %v", err), "setup_error")
		}
		return
	}
	sessionID = run.sessionID
	go r.heartbeatActiveTurn(ctx, sessionID, claimToken, cancel)
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

func (e *Engine) SubmitPromptRouted(ctx context.Context, in RunInput) (*SubmitResult, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if err := e.store.RequireSession(opCtx, in.SessionID); err != nil {
		return nil, err
	}
	inbound := routedsession.InboundContextFromSession(opCtx, e.store, in.SessionID)
	inbound.SenderID = "user"
	route, promptBody, directed, err := routing.PreparePromptRoutedInput(in.Prompt, inbound, e.routeResolver)
	if err != nil {
		return nil, err
	}
	targetSessionID, created, err := routedsession.ResolveOrCreate(opCtx, e.store, in.SessionID, route, inbound, routedsession.ResolveOptions{ModelForAgent: e.modelForAgent, DefaultProvider: e.runtimeCfg.DefaultProvider, DefaultThinking: e.runtimeCfg.DefaultThinkingLevel})
	if err != nil {
		return nil, err
	}
	if targetSessionID != in.SessionID {
		return e.submitPeerRoutedPrompt(opCtx, in.SessionID, targetSessionID, route, promptBody, in.Intent, in.Model, created, directed, in.ParentTurnID, in.Metadata)
	}
	sourceSessionID := in.SessionID
	in.SessionID = targetSessionID
	in.Prompt = promptBody
	in.Metadata = routing.ApplyPromptRouteMetadata(in.Metadata, sourceSessionID, targetSessionID, e.store.SessionAgentID(opCtx, sourceSessionID), route, created)
	return e.SubmitPrompt(opCtx, in)
}

func (e *Engine) SubmitPeerMessage(ctx context.Context, sourceSessionID, targetAgentID, content, intent, model, parentTurnID string) (*SubmitResult, error) {
	return e.submitPeerMessageWithMetadata(ctx, sourceSessionID, targetAgentID, content, intent, model, parentTurnID, nil)
}

func (e *Engine) submitPeerMessageWithMetadata(ctx context.Context, sourceSessionID, targetAgentID, content, intent, model, parentTurnID string, extraMetadata map[string]any) (*SubmitResult, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if err := e.store.RequireSession(opCtx, sourceSessionID); err != nil {
		return nil, err
	}
	inbound := routedsession.InboundContextFromSession(opCtx, e.store, sourceSessionID)
	inbound.SenderID = e.store.SessionAgentID(opCtx, sourceSessionID)
	route := routing.PreparePeerRoutedInput(targetAgentID, "peer-message", content, inbound, e.routeResolver)
	targetSessionID, created, err := routedsession.ResolveOrCreate(opCtx, e.store, sourceSessionID, route, inbound, routedsession.ResolveOptions{ModelForAgent: e.modelForAgent, DefaultProvider: e.runtimeCfg.DefaultProvider, DefaultThinking: e.runtimeCfg.DefaultThinkingLevel})
	if err != nil {
		return nil, err
	}
	return e.submitPeerRoutedPrompt(opCtx, sourceSessionID, targetSessionID, route, content, intent, model, created, true, parentTurnID, extraMetadata)
}

func (e *Engine) ResolveOrCreatePeerSessionID(ctx context.Context, sourceSessionID, targetAgentID string) (string, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if err := e.store.RequireSession(opCtx, sourceSessionID); err != nil {
		return "", err
	}
	inbound := routedsession.InboundContextFromSession(opCtx, e.store, sourceSessionID)
	inbound.SenderID = e.store.SessionAgentID(opCtx, sourceSessionID)
	route := routing.PreparePeerRoutedInput(targetAgentID, "peer-session", "", inbound, e.routeResolver)
	targetSessionID, _, err := routedsession.ResolveOrCreate(opCtx, e.store, sourceSessionID, route, inbound, routedsession.ResolveOptions{ModelForAgent: e.modelForAgent, DefaultProvider: e.runtimeCfg.DefaultProvider, DefaultThinking: e.runtimeCfg.DefaultThinkingLevel})
	if err != nil {
		return "", err
	}
	return targetSessionID, nil
}

func (e *Engine) submitPeerRoutedPrompt(ctx context.Context, sourceSessionID, targetSessionID string, route routing.ResolvedRoute, content, intent, model string, created, directed bool, parentTurnID string, extraMetadata map[string]any) (*SubmitResult, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	sourceAgentID := e.store.SessionAgentID(opCtx, sourceSessionID)
	routingContent := fmt.Sprintf("↪ routed to @%s: %s", route.AgentID, content)
	routingPayload := map[string]any{"kind": "routing", "target_agent_id": route.AgentID, "target_session_id": targetSessionID, "source_agent_id": sourceAgentID, "source_session_id": sourceSessionID, "route_matched_by": route.MatchedBy, "clipped": true}
	logutil.WarnIfErr("add routing message to source session", e.store.AddMessage(opCtx, store.NowID("msg"), sourceSessionID, "system", routingContent, routingPayload))
	metadata := map[string]any{
		"source_session_id":     sourceSessionID,
		"source_agent_id":       sourceAgentID,
		"target_session_id":     targetSessionID,
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
	for k, v := range extraMetadata {
		metadata[k] = v
	}
	result, err := e.SubmitPrompt(opCtx, RunInput{SessionID: targetSessionID, Prompt: content, Intent: intent, Model: model, ParentTurnID: parentTurnID, Metadata: metadata})
	if err != nil {
		return nil, err
	}
	result.SourceSessionID = sourceSessionID
	result.TargetAgentID = route.AgentID
	result.Routed = true
	result.CreatedSession = created
	return result, nil
}

func (e *Engine) modelForAgent(agentID string) string {
	return routing.ModelForAgent(agentID, e.runtimeCfg.Agents, e.runtimeCfg.DefaultModel)
}

func normalizeHoldResolutionSummary(summary string) string {
	return strings.TrimSpace(summary)
}

func normalizedHoldStateForEvent(holdState string) string {
	holdState = store.NormalizedLowerString(holdState)
	if holdState == "" {
		return "none"
	}
	return holdState
}

func holdSummaryForEvent(turnID, holdState, summary string) string {
	summary = normalizeHoldResolutionSummary(summary)
	if summary != "" {
		return summary
	}
	return fmt.Sprintf("turn %s placed on %s hold", turnID, holdState)
}

func (e *Engine) HoldTurnFailure(ctx context.Context, turnID, holdState, summary string) error {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	turnRec, err := e.store.GetTurn(opCtx, turnID)
	if err != nil {
		return err
	}
	holdState = normalizedHoldStateForEvent(holdState)
	summary = holdSummaryForEvent(turnID, holdState, summary)
	if err := e.store.HoldTurnFailure(opCtx, turnID, holdState, summary); err != nil {
		return err
	}
	phase := "held_for_retry_or_skip"
	payload := map[string]any{
		"phase":      "recovery",
		"checkpoint": true,
		"reason":     "failure_held",
		"hold_state": holdState,
		"summary":    summary,
	}
	if err := e.store.AppendTurnEvent(opCtx, turnID, turnRec.SessionID, "turn.failure_held", payload); err != nil {
		return err
	}
	e.PublishRuntimeTurnEvent("turn_failure_held", turnRec.SessionID, turnID, "", turnRec.Status, phase, payload)
	return nil
}

func (e *Engine) RetryHeldTurn(ctx context.Context, turnID, summary string) (*SubmitResult, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	summary = normalizeHoldResolutionSummary(summary)
	turnRec, err := e.store.GetTurn(opCtx, turnID)
	if err != nil {
		return nil, err
	}
	failureRec, err := e.store.GetTurnFailure(opCtx, turnID)
	if err != nil {
		return nil, err
	}
	if failureRec.HoldState == "none" {
		return nil, fmt.Errorf("retry held turn: turn %s is not currently held", turnID)
	}
	metadata := map[string]any{}
	for k, v := range turnRec.Metadata {
		metadata[k] = v
	}
	metadata["retry_of_turn_id"] = turnID
	metadata["retry_failure_kind"] = failureRec.FailureKind
	metadata["retry_hold_state"] = failureRec.HoldState
	metadata["failure_resolution"] = "retry"
	result, err := e.SubmitPrompt(opCtx, RunInput{
		SessionID:    turnRec.SessionID,
		Prompt:       turnRec.Prompt,
		Intent:       store.StringValue(turnRec.Metadata["intent"], "prompt"),
		Model:        store.StringValue(turnRec.Metadata["model"], ""),
		ParentTurnID: store.StringValue(turnRec.Metadata["parent_turn_id"], ""),
		Metadata:     metadata,
	})
	if err != nil {
		return nil, err
	}
	if err := e.store.ResolveTurnFailure(opCtx, turnID, "retried", summary, result.TurnID); err != nil {
		return nil, err
	}
	phase := turnRec.Phase
	if phase == "held_for_retry_or_skip" {
		phase = store.RuntimeTurnPhaseForStatus(turnRec.Status)
	}
	payload := map[string]any{
		"phase":              "recovery",
		"checkpoint":         true,
		"reason":             "failure_resolved",
		"resolution_state":   "retried",
		"resolution_summary": summary,
		"resolved_turn_id":   result.TurnID,
	}
	logutil.WarnIfErr("append turn.failure_resolved event", e.store.AppendTurnEvent(opCtx, turnID, turnRec.SessionID, "turn.failure_resolved", payload))
	e.PublishRuntimeTurnEvent("turn_failure_resolved", turnRec.SessionID, turnID, "", turnRec.Status, phase, payload)
	return result, nil
}

func (e *Engine) SkipHeldTurn(ctx context.Context, turnID, summary string) error {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	summary = normalizeHoldResolutionSummary(summary)
	turnRec, err := e.store.GetTurn(opCtx, turnID)
	if err != nil {
		return err
	}
	failureRec, err := e.store.GetTurnFailure(opCtx, turnID)
	if err != nil {
		return err
	}
	if failureRec.HoldState == "none" {
		return fmt.Errorf("skip held turn: turn %s is not currently held", turnID)
	}
	if err := e.store.ResolveTurnFailure(opCtx, turnID, "skipped", summary, ""); err != nil {
		return err
	}
	phase := turnRec.Phase
	if phase == "held_for_retry_or_skip" {
		phase = store.RuntimeTurnPhaseForStatus(turnRec.Status)
	}
	payload := map[string]any{
		"phase":              "recovery",
		"checkpoint":         true,
		"reason":             "failure_resolved",
		"resolution_state":   "skipped",
		"resolution_summary": summary,
	}
	if err := e.store.AppendTurnEvent(opCtx, turnID, turnRec.SessionID, "turn.failure_resolved", payload); err != nil {
		return err
	}
	e.PublishRuntimeTurnEvent("turn_failure_resolved", turnRec.SessionID, turnID, "", turnRec.Status, phase, payload)
	return nil
}

const (
	DirectKindPrompt      = "prompt"
	DirectKindPeerMessage = "peer-message"
	DirectKindContinue    = "continue"

	DirectSourceKindDirect   = "direct"
	DirectSourceKindIPC      = "ipc"
	DirectSourceKindSystem   = "system"
	DirectSourceKindInternal = "internal"
)

func normalizeDirectKind(kind string) string {
	kind = store.NormalizedLowerString(kind)
	switch kind {
	case "", DirectKindPrompt:
		return DirectKindPrompt
	case DirectKindPeerMessage:
		return DirectKindPeerMessage
	case DirectKindContinue:
		return DirectKindContinue
	default:
		return kind
	}
}

func normalizeDirectSourceKind(kind string) string {
	kind = store.NormalizedLowerString(kind)
	switch kind {
	case "", DirectSourceKindDirect:
		return DirectSourceKindDirect
	case DirectSourceKindIPC:
		return DirectSourceKindIPC
	case DirectSourceKindSystem:
		return DirectSourceKindSystem
	case DirectSourceKindInternal:
		return DirectSourceKindInternal
	default:
		return kind
	}
}

func (e *Engine) ProcessSystemDirect(ctx context.Context, in DirectInput) (*SubmitResult, error) {
	in.Origin.SourceKind = DirectSourceKindSystem
	if strings.TrimSpace(in.Origin.Role) == "" {
		in.Origin.Role = "system"
	}
	return e.ProcessDirect(ctx, in)
}

func (e *Engine) ProcessInternalDirect(ctx context.Context, in DirectInput) (*SubmitResult, error) {
	in.Origin.SourceKind = DirectSourceKindInternal
	if strings.TrimSpace(in.Origin.Role) == "" {
		in.Origin.Role = "system"
	}
	return e.ProcessDirect(ctx, in)
}

func (e *Engine) resolveDirectSessionID(ctx context.Context, in DirectInput) (string, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	sessionID := strings.TrimSpace(in.SessionID)
	sessionKey := strings.TrimSpace(in.SessionKey)
	if sessionKey != "" {
		if e.store == nil {
			return "", fmt.Errorf("direct processing requires store-backed session resolution")
		}
		resolvedSessionID, err := e.store.ResolveSessionIDByKeyOrAlias(opCtx, sessionKey)
		if err != nil {
			return "", err
		}
		if sessionID != "" && sessionID != resolvedSessionID {
			return "", fmt.Errorf("direct processing session id %q does not match session key %q", sessionID, sessionKey)
		}
		return resolvedSessionID, nil
	}
	if sessionID != "" {
		return sessionID, nil
	}
	return "", fmt.Errorf("direct processing requires session id or session key")
}

func (e *Engine) ProcessDirect(ctx context.Context, in DirectInput) (*SubmitResult, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	kind := normalizeDirectKind(in.Kind)
	metadata := map[string]any{}
	for k, v := range in.Metadata {
		metadata[k] = v
	}
	metadata["ingress_kind"] = "direct"
	metadata["ingress_source_kind"] = normalizeDirectSourceKind(in.Origin.SourceKind)
	if value := strings.TrimSpace(in.Origin.SourceID); value != "" {
		metadata["ingress_source_id"] = value
	}
	if value := strings.TrimSpace(in.Origin.Role); value != "" {
		metadata["ingress_role"] = value
	}
	if value := strings.TrimSpace(in.Origin.Label); value != "" {
		metadata["ingress_label"] = value
	}
	if value := strings.TrimSpace(in.SessionKey); value != "" {
		metadata["ingress_session_key"] = value
	}
	sessionID, err := e.resolveDirectSessionID(opCtx, in)
	if err != nil {
		return nil, err
	}
	switch kind {
	case DirectKindPrompt:
		return e.SubmitPromptRouted(opCtx, RunInput{SessionID: sessionID, Prompt: in.Prompt, Intent: in.Intent, Model: in.Model, ParentTurnID: in.ParentTurnID, Metadata: metadata})
	case DirectKindPeerMessage:
		if strings.TrimSpace(in.TargetAgentID) == "" {
			return nil, fmt.Errorf("direct peer-message requires target agent id")
		}
		return e.submitPeerMessageWithMetadata(opCtx, sessionID, in.TargetAgentID, in.Prompt, in.Intent, in.Model, in.ParentTurnID, metadata)
	case DirectKindContinue:
		continued, err := e.ContinueSession(opCtx, sessionID)
		if err != nil {
			return nil, err
		}
		return &SubmitResult{SessionID: sessionID, Status: map[bool]string{true: "continued", false: "idle"}[continued], Queued: false}, nil
	default:
		return nil, fmt.Errorf("direct input kind not supported: %s", in.Kind)
	}
}

func directEnvelopeFromInput(in DirectInput) map[string]any {
	envelope := map[string]any{
		"kind":            in.Kind,
		"session_id":      in.SessionID,
		"session_key":     in.SessionKey,
		"target_agent_id": in.TargetAgentID,
		"prompt":          in.Prompt,
		"intent":          in.Intent,
		"model":           in.Model,
		"parent_turn_id":  in.ParentTurnID,
		"metadata":        in.Metadata,
		"origin": map[string]any{
			"source_kind": in.Origin.SourceKind,
			"source_id":   in.Origin.SourceID,
			"role":        in.Origin.Role,
			"label":       in.Origin.Label,
		},
	}
	return envelope
}

func directInputFromEnvelope(envelope map[string]any) DirectInput {
	in := DirectInput{
		Kind:          store.StringValue(envelope["kind"], ""),
		SessionID:     store.StringValue(envelope["session_id"], ""),
		SessionKey:    store.StringValue(envelope["session_key"], ""),
		TargetAgentID: store.StringValue(envelope["target_agent_id"], ""),
		Prompt:        store.StringValue(envelope["prompt"], ""),
		Intent:        store.StringValue(envelope["intent"], ""),
		Model:         store.StringValue(envelope["model"], ""),
		ParentTurnID:  store.StringValue(envelope["parent_turn_id"], ""),
		Metadata:      map[string]any{},
	}
	if metadata, ok := envelope["metadata"].(map[string]any); ok && metadata != nil {
		in.Metadata = metadata
	}
	if origin, ok := envelope["origin"].(map[string]any); ok && origin != nil {
		in.Origin = DirectOrigin{
			SourceKind: store.StringValue(origin["source_kind"], ""),
			SourceID:   store.StringValue(origin["source_id"], ""),
			Role:       store.StringValue(origin["role"], ""),
			Label:      store.StringValue(origin["label"], ""),
		}
	}
	return in
}

const (
	inboundWorkMaxAttempts = 3
	inboundWorkRetryDelay  = 2 * time.Second
)

func (e *Engine) EnqueueDirectInbound(ctx context.Context, in DirectInput) (*store.InboundWorkItem, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if e.store == nil {
		return nil, fmt.Errorf("direct inbound queue requires store")
	}
	sourceKind := normalizeDirectSourceKind(in.Origin.SourceKind)
	envelope := directEnvelopeFromInput(in)
	item, err := e.store.EnqueueInboundWork(opCtx, sourceKind, strings.TrimSpace(in.SessionID), strings.TrimSpace(in.SessionKey), envelope)
	if err != nil {
		return nil, err
	}
	e.PublishRuntimeInboundWorkEvent("inbound_work_enqueued", item, nil)
	return item, nil
}

func (e *Engine) ProcessNextInboundWork(ctx context.Context, claimedBy string) (*store.InboundWorkItem, *SubmitResult, error) {
	if e.store == nil {
		return nil, nil, fmt.Errorf("inbound work processing requires store")
	}
	item, err := e.store.ClaimNextInboundWork(ctx, claimedBy)
	if err != nil {
		return nil, nil, err
	}
	in := directInputFromEnvelope(item.Envelope)
	if strings.TrimSpace(in.Origin.SourceKind) == "" {
		in.Origin.SourceKind = item.SourceKind
	}
	if strings.TrimSpace(in.SessionID) == "" {
		in.SessionID = item.SessionID
	}
	if strings.TrimSpace(in.SessionKey) == "" {
		in.SessionKey = item.ExplicitSessionKey
	}
	result, processErr := e.ProcessDirect(ctx, in)
	return e.finalizeInboundWorkAttempt(ctx, item, result, processErr)
}

func (e *Engine) finalizeInboundWorkAttempt(ctx context.Context, item *store.InboundWorkItem, result *SubmitResult, processErr error) (*store.InboundWorkItem, *SubmitResult, error) {
	postCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if processErr != nil {
		attemptCount := item.AttemptCount + 1
		var updateErr error
		if attemptCount >= inboundWorkMaxAttempts {
			updateErr = e.store.RecordInboundWorkFailure(postCtx, item.ID, attemptCount, processErr.Error())
		} else {
			updateErr = e.store.RecordInboundWorkRetry(postCtx, item.ID, attemptCount, processErr.Error(), inboundWorkRetryDelay*time.Duration(attemptCount))
		}
		statusEvent := map[bool]string{true: "inbound_work_failed", false: "inbound_work_retry_scheduled"}[attemptCount >= inboundWorkMaxAttempts]
		if updateErr != nil {
			return item, result, updateErr
		}
		updated, getErr := e.store.GetInboundWork(postCtx, item.ID)
		if getErr == nil {
			item = updated
		}
		e.PublishRuntimeInboundWorkEvent(statusEvent, item, map[string]any{"error": processErr.Error()})
		return item, result, processErr
	}
	if err := e.store.UpdateInboundWorkStatus(postCtx, item.ID, "completed"); err != nil {
		return item, result, err
	}
	updated, getErr := e.store.GetInboundWork(postCtx, item.ID)
	if getErr == nil {
		item = updated
	}
	e.PublishRuntimeInboundWorkEvent("inbound_work_completed", item, nil)
	return item, result, nil
}

func (e *Engine) ProcessNextInboundWorkIfQueued(ctx context.Context, claimedBy string) (*store.InboundWorkItem, *SubmitResult, bool, error) {
	item, result, err := e.ProcessNextInboundWork(ctx, claimedBy)
	if err == sql.ErrNoRows {
		return nil, nil, false, nil
	}
	if err != nil {
		return item, result, true, err
	}
	return item, result, true, nil
}

func (e *Engine) ProcessQueuedInboundWork(ctx context.Context, claimedBy string, limit int) ([]*store.InboundWorkItem, []*SubmitResult, error) {
	if limit <= 0 {
		limit = 1
	}
	items := make([]*store.InboundWorkItem, 0, limit)
	results := make([]*SubmitResult, 0, limit)
	for i := 0; i < limit; i++ {
		item, result, ok, err := e.ProcessNextInboundWorkIfQueued(ctx, claimedBy)
		if err != nil {
			return items, results, err
		}
		if !ok {
			break
		}
		items = append(items, item)
		results = append(results, result)
	}
	return items, results, nil
}

// Topics returns the engine-wide normalized topic bus.
func (e *Engine) Topics() *topics.Bus { return e.topics }

func (e *Engine) startTopicBridge() {
	if e == nil || e.topics == nil || e.connectivity == nil || e.connectivity.Bus() == nil || e.bgCtx == nil {
		return
	}
	ch, _ := e.connectivity.Bus().Subscribe(e.bgCtx, "*", 64)
	go func() {
		for ev := range ch {
			topic := strings.TrimSpace(ev.Topic)
			if topic == "" {
				topic = "event"
			}
			if !strings.HasPrefix(topic, "connectivity.") {
				topic = "connectivity." + topic
			}
			e.topics.Publish(topics.Envelope{
				Topic:     topic,
				SessionID: ev.SessionID,
				AgentID:   ev.AgentID,
				Source:    tools.FirstNonEmpty(ev.Source, "connectivity"),
				Type:      "event",
				Payload: map[string]any{
					"id":        ev.ID,
					"route_id":  ev.RouteID,
					"transport": ev.Transport,
					"topic":     ev.Topic,
					"payload":   ev.Payload,
				},
				Timestamp: ev.Timestamp,
			})
		}
	}()
}

func (e *Engine) publishTopicEvent(env topics.Envelope) {
	if e == nil || e.topics == nil {
		return
	}
	e.topics.Publish(env)
}

func (e *Engine) PublishTopicEvent(env topics.Envelope) {
	e.publishTopicEvent(env)
}

func (e *Engine) publishTopicFromBroadcast(sessionID string, ev map[string]any) {
	if e == nil || e.topics == nil || ev == nil {
		return
	}
	evType, _ := ev["type"].(string)
	topic, envelopeType := topicForBroadcastEvent(evType)
	if topic == "" {
		return
	}
	agentID, _ := ev["agent_id"].(string)
	e.publishTopicEvent(topics.Envelope{
		Topic:     topic,
		SessionID: sessionID,
		AgentID:   agentID,
		Source:    "turn",
		Type:      envelopeType,
		Payload:   cloneMap(ev),
	})
}

func topicForBroadcastEvent(evType string) (topic string, envelopeType string) {
	switch strings.TrimSpace(evType) {
	case "agent_status":
		return "turn.status", "status"
	case "agent_draft_delta":
		return "turn.draft", "delta"
	case "agent_thought_delta":
		return "turn.thought", "delta"
	case "tool_finished":
		return "turn.tool.end", "result"
	case "tool_failed":
		return "turn.tool.end", "error"
	case "tool_skipped":
		return "turn.tool.end", "notice"
	case "steering_enqueued", "steering_dequeued", "steering_continue_staged", "steering_continued", "steering_injected", "steering_rejected":
		return "session.steering", "notice"
	case "subturn_created", "subturn_status", "subturn_result_ready", "subturn_result_delivered", "subturn_orphaned", "subturn_cancel_requested":
		return "turn.subturn", "notice"
	case "compaction":
		return "session.compaction", "notice"
	case "routing_decision", "routing_incoming":
		return "session.routing", "notice"
	case "new_post", "agent_response":
		return "turn.response", "result"
	default:
		if strings.TrimSpace(evType) == "" {
			return "", ""
		}
		return "event." + strings.ReplaceAll(evType, "_", "."), "notice"
	}
}

func (e *Engine) PublishRuntimeInboundWorkEvent(eventType string, item *store.InboundWorkItem, extra map[string]any) {
	if e == nil || e.topics == nil || item == nil {
		return
	}
	payload := cloneMap(extra)
	if payload == nil {
		payload = map[string]any{}
	}
	payload["id"] = item.ID
	payload["status"] = item.Status
	payload["source_kind"] = item.SourceKind
	payload["session_id"] = item.SessionID
	payload["explicit_session_key"] = item.ExplicitSessionKey
	payload["attempt_count"] = item.AttemptCount
	payload["last_error"] = item.LastError
	payload["next_attempt_at"] = item.NextAttemptAt
	payload["claimed_by"] = item.ClaimedBy
	payload["claimed_at"] = item.ClaimedAt
	payload["created_at"] = item.CreatedAt
	payload["updated_at"] = item.UpdatedAt
	e.publishRuntimeTopicEvent("runtime.inbound_work", item.SessionID, "", "notice", eventType, payload)
}

func (e *Engine) PublishRuntimeDispatcherEvent(eventType string, payload map[string]any) {
	e.publishRuntimeTopicEvent("runtime.dispatcher", "", "", "notice", eventType, payload)
}

func (e *Engine) PublishRuntimeHookEvent(eventType string, req HookRequest, source string, action string, durationMS int, err error) {
	payload := runtimeHookPayload(req, nil)
	payload["source"] = strings.TrimSpace(source)
	payload["action"] = strings.TrimSpace(action)
	payload["duration_ms"] = durationMS
	payload["trace"] = map[string]any{
		"id":         req.Trace.ID,
		"emitted_at": req.Trace.EmittedAt,
	}
	if err != nil {
		payload["error"] = err.Error()
	}
	e.publishRuntimeTopicEvent("runtime.hook", req.SessionID, req.AgentID, "notice", eventType, payload)
}

func (e *Engine) PublishRuntimeHookDecisionEvent(eventType string, req HookRequest, payload map[string]any) {
	e.publishRuntimeTopicEvent("runtime.hook", req.SessionID, req.AgentID, "notice", eventType, runtimeHookPayload(req, payload))
}

func (e *Engine) PublishRuntimeTurnEvent(eventType, sessionID, turnID, agentID, status, phase string, payload map[string]any) {
	body := cloneMap(payload)
	if body == nil {
		body = map[string]any{}
	}
	body["turn_id"] = strings.TrimSpace(turnID)
	body["status"] = strings.TrimSpace(status)
	body["phase"] = strings.TrimSpace(phase)
	e.publishRuntimeTopicEvent("runtime.turn", sessionID, agentID, "notice", eventType, body)
}

func (e *Engine) PublishRuntimeSessionEvent(eventType, sessionID, agentID, status string, payload map[string]any) {
	body := cloneMap(payload)
	if body == nil {
		body = map[string]any{}
	}
	body["status"] = strings.TrimSpace(status)
	e.publishRuntimeTopicEvent("runtime.session", sessionID, agentID, "notice", eventType, body)
}

func (e *Engine) PublishRuntimeToolEvent(eventType, sessionID, turnID, agentID, toolName, toolCallID string, iteration int, err error, payload map[string]any) {
	body := cloneMap(payload)
	if body == nil {
		body = map[string]any{}
	}
	body["turn_id"] = strings.TrimSpace(turnID)
	body["tool"] = strings.TrimSpace(toolName)
	body["tool_call_id"] = strings.TrimSpace(toolCallID)
	if iteration > 0 {
		body["iteration"] = iteration
	}
	envelopeType := "notice"
	if err != nil {
		body["error"] = err.Error()
		envelopeType = "error"
	}
	e.publishRuntimeTopicEvent("runtime.tool", sessionID, agentID, envelopeType, eventType, body)
}

func (e *Engine) PublishRuntimeRoutingEvent(eventType string, decision routing.Event) {
	body := cloneMap(decision.Metadata)
	if body == nil {
		body = map[string]any{}
	}
	body["route_event_id"] = decision.ID
	body["turn_id"] = strings.TrimSpace(decision.TurnID)
	body["source_session_id"] = strings.TrimSpace(decision.SourceSession)
	body["target_session_id"] = strings.TrimSpace(decision.TargetSession)
	body["source_agent_id"] = strings.TrimSpace(decision.SourceAgentID)
	body["target_agent_id"] = strings.TrimSpace(decision.TargetAgentID)
	body["mode"] = strings.TrimSpace(decision.Mode)
	body["matched_by"] = strings.TrimSpace(decision.MatchedBy)
	body["routing_policy"] = strings.TrimSpace(decision.RoutingPolicy)
	body["requested_agent_id"] = strings.TrimSpace(decision.RequestedAgent)
	body["created_at"] = strings.TrimSpace(decision.CreatedAt)
	sessionID := strings.TrimSpace(decision.SourceSession)
	if strings.TrimSpace(eventType) == "routing_incoming" && strings.TrimSpace(decision.TargetSession) != "" {
		sessionID = strings.TrimSpace(decision.TargetSession)
	}
	e.publishRuntimeTopicEvent("runtime.routing", sessionID, strings.TrimSpace(decision.TargetAgentID), "notice", eventType, body)
}

func runtimeHookPayload(req HookRequest, payload map[string]any) map[string]any {
	body := cloneMap(payload)
	if body == nil {
		body = map[string]any{}
	}
	body["hook"] = req.Name
	body["session_id"] = req.SessionID
	body["turn_id"] = req.TurnID
	body["agent_id"] = req.AgentID
	body["iteration"] = req.Iteration
	body["turn_status"] = req.TurnStatus
	body["turn_phase"] = req.TurnPhase
	body["session_status"] = req.SessionStatus
	if req.ToolCall != nil {
		body["tool"] = req.ToolCall.Name
		body["tool_call_id"] = req.ToolCall.ID
	}
	return body
}

func (e *Engine) publishRuntimeTopicEvent(topic, sessionID, agentID, envelopeType, eventType string, payload map[string]any) {
	if e == nil || e.topics == nil {
		return
	}
	body := cloneMap(payload)
	if body == nil {
		body = map[string]any{}
	}
	body["type"] = strings.TrimSpace(eventType)
	e.publishTopicEvent(topics.Envelope{
		Topic:     strings.TrimSpace(topic),
		SessionID: strings.TrimSpace(sessionID),
		AgentID:   strings.TrimSpace(agentID),
		Source:    "runtime",
		Type:      strings.TrimSpace(tools.FirstNonEmpty(envelopeType, "notice")),
		Payload:   body,
	})
}

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
