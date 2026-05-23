// Package turn implements Gi's append-only turn engine.
package turn

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rcarmo/gi/internal/compaction"
	"github.com/rcarmo/gi/internal/logutil"
	giskills "github.com/rcarmo/gi/internal/skills"
	"github.com/rcarmo/gi/internal/tools"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/connectivity"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/peering"
	"github.com/rcarmo/gi/internal/routing"
	"github.com/rcarmo/gi/internal/routing/routedsession"
	"github.com/rcarmo/gi/internal/scripting"
	"github.com/rcarmo/gi/internal/store"
	storeaudit "github.com/rcarmo/gi/internal/store/audit"
	"github.com/rcarmo/gi/internal/store/internalx"
	"github.com/rcarmo/gi/internal/store/queue"
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
		if model := strings.TrimSpace(internalx.StringValue(turnRec.Metadata["model"], "")); model != "" {
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
			subTurnMaxDepth = internalx.IntValueOr(v, defaultSubTurnMaxDepth)
		}
		if v, ok := in.Metadata["subturn_max_concurrency"]; ok {
			subTurnMaxConcurrency = internalx.IntValueOr(v, defaultSubTurnMaxConcurrency)
		}
		if modeRaw, ok := in.Metadata["subturn_delivery_mode"]; ok {
			mode, err := store.NormalizeSubTurnDeliveryMode(internalx.StringValue(modeRaw, "sync"))
			if err != nil {
				return nil, err
			}
			subTurnDeliveryMode = mode
		}
		subTurnCritical = internalx.BoolValueOr(in.Metadata["subturn_critical"], internalx.BoolValue(in.Metadata["critical"]))
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
		subTurnDepth = internalx.IntValueOr(parentTurn.Metadata["subturn_depth"], 0) + 1
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
	if err := routing.RecordDecision(durableCtx, in.SessionID, turnID, metadata, routing.Options{ResolveSourceAgentID: func(ctx context.Context, sessionID string) string {
		identity, err := e.store.RequireSessionIdentityRuntime(ctx, sessionID)
		if err != nil {
			return ""
		}
		return identity.AgentID
	}, RecordRouteEvent: func(ctx context.Context, event routing.Event) (int64, error) {
		return storeaudit.RecordRouteEvent(ctx, e.store.DB(), storeaudit.RouteEvent(event))
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
		if model := strings.TrimSpace(internalx.StringValue(turnRec.Metadata["model"], "")); model != "" {
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
	sourceIdentity, err := e.store.RequireSessionIdentityRuntime(opCtx, in.SessionID)
	if err != nil {
		return nil, err
	}
	inbound, err := routedsession.RequireInboundContextFromSession(opCtx, e.store, in.SessionID)
	if err != nil {
		return nil, err
	}
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
	in.Metadata = routing.ApplyPromptRouteMetadata(in.Metadata, sourceSessionID, targetSessionID, sourceIdentity.AgentID, route, created)
	return e.SubmitPrompt(opCtx, in)
}

func (e *Engine) SubmitPeerMessage(ctx context.Context, sourceSessionID, targetAgentID, content, intent, model, parentTurnID string) (*SubmitResult, error) {
	return e.submitPeerMessageWithMetadata(ctx, sourceSessionID, targetAgentID, content, intent, model, parentTurnID, nil)
}

func (e *Engine) submitPeerMessageWithMetadata(ctx context.Context, sourceSessionID, targetAgentID, content, intent, model, parentTurnID string, extraMetadata map[string]any) (*SubmitResult, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	sourceIdentity, err := e.store.RequireSessionIdentityRuntime(opCtx, sourceSessionID)
	if err != nil {
		return nil, err
	}
	inbound, err := routedsession.RequireInboundContextFromSession(opCtx, e.store, sourceSessionID)
	if err != nil {
		return nil, err
	}
	inbound.SenderID = sourceIdentity.AgentID
	route := routing.PreparePeerRoutedInput(targetAgentID, "peer-message", content, inbound, e.routeResolver)
	targetSessionID, created, err := routedsession.ResolveOrCreate(opCtx, e.store, sourceSessionID, route, inbound, routedsession.ResolveOptions{ModelForAgent: e.modelForAgent, DefaultProvider: e.runtimeCfg.DefaultProvider, DefaultThinking: e.runtimeCfg.DefaultThinkingLevel})
	if err != nil {
		return nil, err
	}
	return e.submitPeerRoutedPrompt(opCtx, sourceSessionID, targetSessionID, route, content, intent, model, created, true, parentTurnID, extraMetadata)
}

func (e *Engine) ResolveOrCreatePeerSessionID(ctx context.Context, sourceSessionID, targetAgentID string) (string, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	sourceIdentity, err := e.store.RequireSessionIdentityRuntime(opCtx, sourceSessionID)
	if err != nil {
		return "", err
	}
	inbound, err := routedsession.RequireInboundContextFromSession(opCtx, e.store, sourceSessionID)
	if err != nil {
		return "", err
	}
	inbound.SenderID = sourceIdentity.AgentID
	route := routing.PreparePeerRoutedInput(targetAgentID, "peer-session", "", inbound, e.routeResolver)
	targetSessionID, _, err := routedsession.ResolveOrCreate(opCtx, e.store, sourceSessionID, route, inbound, routedsession.ResolveOptions{ModelForAgent: e.modelForAgent, DefaultProvider: e.runtimeCfg.DefaultProvider, DefaultThinking: e.runtimeCfg.DefaultThinkingLevel})
	if err != nil {
		return "", err
	}
	return targetSessionID, nil
}

func (e *Engine) submitPeerRoutedPrompt(ctx context.Context, sourceSessionID, targetSessionID string, route routing.ResolvedRoute, content, intent, model string, created, directed bool, parentTurnID string, extraMetadata map[string]any) (*SubmitResult, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	sourceIdentity, err := e.store.RequireSessionIdentityRuntime(opCtx, sourceSessionID)
	if err != nil {
		return nil, err
	}
	sourceAgentID := sourceIdentity.AgentID
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
	holdState = internalx.NormalizedLowerString(holdState)
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
		Intent:       internalx.StringValue(turnRec.Metadata["intent"], "prompt"),
		Model:        internalx.StringValue(turnRec.Metadata["model"], ""),
		ParentTurnID: internalx.StringValue(turnRec.Metadata["parent_turn_id"], ""),
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

var ingressMetadataKeys = []string{"source_session_id", "source_agent_id", "target_agent_id", "routed_from_prompt", "ingress_kind", "ingress_source_kind", "ingress_source_id", "ingress_role", "ingress_label", "ingress_session_key"}
var turnStartedMetadataKeys = []string{"parent_turn_id", "route_mode", "route_matched_by", "source_session_id", "source_agent_id", "target_agent_id", "routed_from_prompt", "ingress_kind", "ingress_source_kind", "ingress_source_id", "ingress_role", "ingress_label", "ingress_session_key"}

func normalizeDirectKind(kind string) string {
	kind = internalx.NormalizedLowerString(kind)
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
	kind = internalx.NormalizedLowerString(kind)
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
	metadata := cloneMap(in.Metadata)
	if metadata == nil {
		metadata = map[string]any{}
	}
	envelope := map[string]any{
		"kind":            normalizeDirectKind(in.Kind),
		"session_id":      strings.TrimSpace(in.SessionID),
		"session_key":     strings.TrimSpace(in.SessionKey),
		"target_agent_id": strings.TrimSpace(in.TargetAgentID),
		"prompt":          in.Prompt,
		"intent":          in.Intent,
		"model":           in.Model,
		"parent_turn_id":  strings.TrimSpace(in.ParentTurnID),
		"metadata":        metadata,
		"origin": map[string]any{
			"source_kind": normalizeDirectSourceKind(in.Origin.SourceKind),
			"source_id":   strings.TrimSpace(in.Origin.SourceID),
			"role":        strings.TrimSpace(in.Origin.Role),
			"label":       strings.TrimSpace(in.Origin.Label),
		},
	}
	return envelope
}

func directInputFromEnvelope(envelope map[string]any) DirectInput {
	in := DirectInput{
		Kind:          normalizeDirectKind(internalx.StringValue(envelope["kind"], "")),
		SessionID:     strings.TrimSpace(internalx.StringValue(envelope["session_id"], "")),
		SessionKey:    strings.TrimSpace(internalx.StringValue(envelope["session_key"], "")),
		TargetAgentID: strings.TrimSpace(internalx.StringValue(envelope["target_agent_id"], "")),
		Prompt:        internalx.StringValue(envelope["prompt"], ""),
		Intent:        internalx.StringValue(envelope["intent"], ""),
		Model:         internalx.StringValue(envelope["model"], ""),
		ParentTurnID:  strings.TrimSpace(internalx.StringValue(envelope["parent_turn_id"], "")),
		Metadata:      map[string]any{},
	}
	if metadata, ok := envelope["metadata"].(map[string]any); ok && metadata != nil {
		in.Metadata = cloneMap(metadata)
	}
	if origin, ok := envelope["origin"].(map[string]any); ok && origin != nil {
		in.Origin = DirectOrigin{
			SourceKind: normalizeDirectSourceKind(internalx.StringValue(origin["source_kind"], "")),
			SourceID:   strings.TrimSpace(internalx.StringValue(origin["source_id"], "")),
			Role:       strings.TrimSpace(internalx.StringValue(origin["role"], "")),
			Label:      strings.TrimSpace(internalx.StringValue(origin["label"], "")),
		}
	}
	return in
}

const (
	inboundWorkMaxAttempts = 3
	inboundWorkRetryDelay  = 2 * time.Second
)

func (e *Engine) EnqueueDirectInbound(ctx context.Context, in DirectInput) (*queue.InboundWorkItem, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if e.store == nil {
		return nil, fmt.Errorf("direct inbound queue requires store")
	}
	sourceKind := normalizeDirectSourceKind(in.Origin.SourceKind)
	envelope := directEnvelopeFromInput(in)
	item, err := queue.EnqueueInboundWork(opCtx, e.store.DB(), sourceKind, strings.TrimSpace(in.SessionID), strings.TrimSpace(in.SessionKey), envelope)
	if err != nil {
		return nil, err
	}
	e.PublishRuntimeInboundWorkEvent("inbound_work_enqueued", item, nil)
	return item, nil
}

func (e *Engine) ProcessNextInboundWork(ctx context.Context, claimedBy string) (*queue.InboundWorkItem, *SubmitResult, error) {
	if e.store == nil {
		return nil, nil, fmt.Errorf("inbound work processing requires store")
	}
	item, err := queue.ClaimNextInboundWork(ctx, e.store.DB(), claimedBy)
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

func (e *Engine) finalizeInboundWorkAttempt(ctx context.Context, item *queue.InboundWorkItem, result *SubmitResult, processErr error) (*queue.InboundWorkItem, *SubmitResult, error) {
	postCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if processErr != nil {
		attemptCount := item.AttemptCount + 1
		var updateErr error
		if attemptCount >= inboundWorkMaxAttempts {
			updateErr = queue.RecordInboundWorkFailure(postCtx, e.store.DB(), item.ID, attemptCount, processErr.Error())
		} else {
			updateErr = queue.RecordInboundWorkRetry(postCtx, e.store.DB(), item.ID, attemptCount, processErr.Error(), inboundWorkRetryDelay*time.Duration(attemptCount))
		}
		statusEvent := map[bool]string{true: "inbound_work_failed", false: "inbound_work_retry_scheduled"}[attemptCount >= inboundWorkMaxAttempts]
		if updateErr != nil {
			return item, result, updateErr
		}
		updated, getErr := queue.GetInboundWork(postCtx, e.store.DB(), item.ID)
		if getErr == nil {
			item = updated
		}
		e.PublishRuntimeInboundWorkEvent(statusEvent, item, map[string]any{"error": processErr.Error()})
		return item, result, processErr
	}
	if err := queue.UpdateInboundWorkStatus(postCtx, e.store.DB(), item.ID, "completed"); err != nil {
		return item, result, err
	}
	updated, getErr := queue.GetInboundWork(postCtx, e.store.DB(), item.ID)
	if getErr == nil {
		item = updated
	}
	e.PublishRuntimeInboundWorkEvent("inbound_work_completed", item, nil)
	return item, result, nil
}

func (e *Engine) ProcessNextInboundWorkIfQueued(ctx context.Context, claimedBy string) (*queue.InboundWorkItem, *SubmitResult, bool, error) {
	item, result, err := e.ProcessNextInboundWork(ctx, claimedBy)
	if err == sql.ErrNoRows {
		return nil, nil, false, nil
	}
	if err != nil {
		return item, result, true, err
	}
	return item, result, true, nil
}

func (e *Engine) ProcessQueuedInboundWork(ctx context.Context, claimedBy string, limit int) ([]*queue.InboundWorkItem, []*SubmitResult, error) {
	if limit <= 0 {
		limit = 1
	}
	items := make([]*queue.InboundWorkItem, 0, limit)
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

func (e *Engine) PublishRuntimeInboundWorkEvent(eventType string, item *queue.InboundWorkItem, extra map[string]any) {
	if e == nil || e.topics == nil || item == nil {
		return
	}
	payload := cloneMap(extra)
	if payload == nil {
		payload = map[string]any{}
	}
	payload["id"] = item.ID
	payload["status"] = strings.TrimSpace(item.Status)
	payload["source_kind"] = normalizeDirectSourceKind(item.SourceKind)
	payload["session_id"] = strings.TrimSpace(item.SessionID)
	payload["explicit_session_key"] = strings.TrimSpace(item.ExplicitSessionKey)
	payload["attempt_count"] = item.AttemptCount
	payload["last_error"] = strings.TrimSpace(item.LastError)
	payload["next_attempt_at"] = strings.TrimSpace(item.NextAttemptAt)
	payload["claimed_by"] = strings.TrimSpace(item.ClaimedBy)
	payload["claimed_at"] = strings.TrimSpace(item.ClaimedAt)
	payload["created_at"] = strings.TrimSpace(item.CreatedAt)
	payload["updated_at"] = strings.TrimSpace(item.UpdatedAt)
	e.publishRuntimeTopicEvent("runtime.inbound_work", strings.TrimSpace(item.SessionID), "", "notice", eventType, payload)
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

func copySelectedMetadata(dst, src map[string]any, keys []string) {
	if dst == nil || src == nil {
		return
	}
	for _, key := range keys {
		if value, ok := src[key]; ok {
			dst[key] = value
		}
	}
}

var activeTurnHeartbeatInterval = 5 * time.Second

const interruptedTurnStaleAfter = 30 * time.Second

func (e *Engine) appendRecoveryFailureEvent(ctx context.Context, claim store.ActiveTurnClaim, err error) {
	if e == nil || e.store == nil || err == nil || strings.TrimSpace(claim.SessionID) == "" || strings.TrimSpace(claim.TurnID) == "" {
		return
	}
	payload := map[string]any{
		"phase":                "recovery",
		"checkpoint":           true,
		"reason":               "recovery_failed",
		"error":                err.Error(),
		"previous_status":      claim.Status,
		"previous_phase":       claim.Phase,
		"recovery_disposition": recoveryDispositionForClaim(claim),
		"stale_claim":          true,
	}
	logutil.WarnIfErr("append recovery failure event", e.store.AppendTurnEvent(ctx, claim.TurnID, claim.SessionID, "turn.recovery_failed", payload))
	runner := e.runner(claim.SessionID)
	agentID, model := "", ""
	if turnRec, turnErr := e.store.GetTurn(ctx, claim.TurnID); turnErr == nil {
		agentID, model = runner.resolveTurnAgentAndModel(ctx, e.store, turnRec, claim.SessionID, turnRec.Prompt)
	}
	e.PublishRuntimeTurnEvent("turn_recovery_failed", claim.SessionID, claim.TurnID, agentID, claim.Status, claim.Phase, cloneMap(payload))
	runner.emitTurnStateHook(ctx, claim.SessionID, claim.TurnID, agentID, model, claim.Status, claim.Phase, cloneMap(payload))
	runner.emitSessionStateHook(ctx, claim.SessionID, agentID, model, "running", map[string]any{
		"reason":               "recovery_failed",
		"error":                err.Error(),
		"recovery_disposition": recoveryDispositionForClaim(claim),
		"stale_claim":          true,
		"active_turn_id":       claim.TurnID,
		"turn_id":              claim.TurnID,
		"turn_status":          claim.Status,
		"turn_phase":           claim.Phase,
	})
}

func (e *Engine) recoverInterruptedTurns(ctx context.Context, sessionID string) (bool, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	claims, err := e.store.ListStaleActiveTurnClaims(opCtx, time.Now().Add(-interruptedTurnStaleAfter), sessionID)
	if err != nil {
		return false, err
	}
	type recoveryScanCounts struct {
		recovered int
		failed    int
	}
	recovered := false
	var firstErr error
	sessionsToRestart := map[string]bool{}
	sessionCounts := map[string]*recoveryScanCounts{}
	failedClaimsBySession := map[string][]store.ActiveTurnClaim{}
	for _, claim := range claims {
		counts := sessionCounts[claim.SessionID]
		if counts == nil {
			counts = &recoveryScanCounts{}
			sessionCounts[claim.SessionID] = counts
		}
		if err := e.recoverInterruptedTurn(opCtx, claim); err != nil {
			log.Printf("turn recovery: recover %s/%s failed: %v", claim.SessionID, claim.TurnID, err)
			e.appendRecoveryFailureEvent(opCtx, claim, err)
			failedClaimsBySession[claim.SessionID] = append(failedClaimsBySession[claim.SessionID], claim)
			counts.failed++
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		counts.recovered++
		recovered = true
		queueCount, err := e.store.CountQueuedTurns(opCtx, claim.SessionID)
		if err != nil {
			return recovered, err
		}
		if queueCount > 0 {
			sessionsToRestart[claim.SessionID] = true
		}
	}
	for failedSessionID, counts := range sessionCounts {
		if counts.failed > 0 {
			e.emitRecoveryScanFailureSessionState(opCtx, failedSessionID, counts.recovered, counts.failed)
			for _, claim := range failedClaimsBySession[failedSessionID] {
				logutil.WarnIfErr("append recovery scan summary event", e.store.AppendTurnEvent(opCtx, claim.TurnID, failedSessionID, "turn.recovery_scan_failed", map[string]any{
					"phase":                 "recovery",
					"checkpoint":            true,
					"reason":                "recovery_scan_failed",
					"recovery_disposition":  recoveryDispositionForClaim(claim),
					"recovered_claim_count": counts.recovered,
					"failed_claim_count":    counts.failed,
				}))
			}
		}
	}
	for recoveredSessionID := range sessionsToRestart {
		if err := e.startNextQueuedTurn(opCtx, recoveredSessionID); err != nil {
			e.emitRecoveryRestartFailureSessionState(opCtx, recoveredSessionID, err)
			return recovered, err
		}
	}
	if firstErr != nil {
		return recovered, firstErr
	}
	return recovered, nil
}

func (e *Engine) emitRecoveryScanFailureSessionState(ctx context.Context, sessionID string, recoveredCount, failedCount int) {
	if e == nil || e.store == nil || strings.TrimSpace(sessionID) == "" || failedCount == 0 {
		return
	}
	runner := e.runner(sessionID)
	agentID, model := "", ""
	activeTurnID, _, activeErr := e.store.GetSessionActiveTurn(ctx, sessionID)
	if activeErr != nil && activeErr != sql.ErrNoRows {
		return
	}
	queueCount, err := e.store.CountQueuedTurns(ctx, sessionID)
	if err != nil {
		return
	}
	status := "idle"
	activeTurnValue := any(nil)
	if activeErr == nil {
		status = "running"
		activeTurnValue = activeTurnID
		if turnRec, turnErr := e.store.GetTurn(ctx, activeTurnID); turnErr == nil {
			agentID, model = runner.resolveTurnAgentAndModel(ctx, e.store, turnRec, sessionID, turnRec.Prompt)
		}
	} else if queueCount > 0 {
		status = "queued"
	}
	if model == "" {
		if stateModel, stateErr := e.store.SessionStateString(ctx, sessionID, "model"); stateErr == nil {
			model = stateModel
		}
	}
	runner.emitSessionStateHook(ctx, sessionID, agentID, model, status, map[string]any{
		"reason":                "recovery_scan_failed",
		"active_turn_id":        activeTurnValue,
		"queue_count":           queueCount,
		"recovered_claim_count": recoveredCount,
		"failed_claim_count":    failedCount,
	})
}

func (e *Engine) emitRecoveryRestartFailureSessionState(ctx context.Context, sessionID string, err error) {
	if e == nil || e.store == nil || strings.TrimSpace(sessionID) == "" || err == nil {
		return
	}
	runner := e.runner(sessionID)
	agentID, model := "", ""
	activeTurnID, _, activeErr := e.store.GetSessionActiveTurn(ctx, sessionID)
	if activeErr != nil && activeErr != sql.ErrNoRows {
		return
	}
	queueCount, countErr := e.store.CountQueuedTurns(ctx, sessionID)
	if countErr != nil {
		return
	}
	status := "idle"
	activeTurnValue := any(nil)
	if activeErr == nil {
		status = "running"
		activeTurnValue = activeTurnID
		if turnRec, turnErr := e.store.GetTurn(ctx, activeTurnID); turnErr == nil {
			agentID, model = runner.resolveTurnAgentAndModel(ctx, e.store, turnRec, sessionID, turnRec.Prompt)
		}
	} else if queueCount > 0 {
		status = "queued"
	}
	if model == "" {
		if stateModel, stateErr := e.store.SessionStateString(ctx, sessionID, "model"); stateErr == nil {
			model = stateModel
		}
	}
	payload := map[string]any{
		"phase":          "recovery",
		"checkpoint":     true,
		"reason":         "recovery_restart_failed",
		"error":          err.Error(),
		"active_turn_id": activeTurnValue,
		"queue_count":    queueCount,
	}
	if queuedTurn, queuedErr := e.store.GetNextQueuedTurn(ctx, sessionID); queuedErr == nil {
		logutil.WarnIfErr("append recovery restart failure event", e.store.AppendTurnEvent(ctx, queuedTurn.ID, sessionID, "turn.recovery_restart_failed", cloneMap(payload)))
	}
	runner.emitSessionStateHook(ctx, sessionID, agentID, model, status, payload)
}

func recoveryDispositionForClaim(claim store.ActiveTurnClaim) string {
	switch claim.Phase {
	case "waiting_on_tools":
		return "hold_for_retry_or_skip_after_tool_checkpoint"
	case "cancelling":
		return "abort_cancelling"
	case "completed", "failed", "aborted", "cancelled":
		return "release_terminal"
	default:
		if claim.Phase == "compacting" {
			return "requeue_after_compaction_checkpoint"
		}
		return "requeue_interrupted_turn"
	}
}

func (e *Engine) recoverInterruptedTurn(ctx context.Context, claim store.ActiveTurnClaim) error {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	disposition := recoveryDispositionForClaim(claim)
	status := claim.Status
	phase := claim.Phase
	markFinished := false

	switch claim.Phase {
	case "waiting_on_tools":
		status = "failed"
		phase = "held_for_retry_or_skip"
		markFinished = true
	case "cancelling":
		status = "aborted"
		phase = "aborted"
		markFinished = true
	case "completed", "failed", "aborted", "cancelled":
		// Terminal turn with a stale claim: just release the claim.
	default:
		status = "queued"
		phase = "queued"
	}

	payload := map[string]any{
		"phase":                "recovery",
		"checkpoint":           true,
		"reason":               "recovery",
		"previous_status":      claim.Status,
		"previous_phase":       claim.Phase,
		"recovery_disposition": disposition,
		"stale_claim":          true,
	}
	if staleFor, err := staleDurationSeconds(claim.UpdatedAt); err == nil {
		payload["stale_for_seconds"] = staleFor
	}

	if disposition != "release_terminal" {
		if disposition == "hold_for_retry_or_skip_after_tool_checkpoint" {
			if err := e.store.MarkTurnFailureWithFallbackErr(e.backgroundContext(), nil, claim.TurnID, claim.SessionID, "recovery_interrupted_tool_phase", "review", "Recovered stale turn that was interrupted while waiting on tool results"); err != nil {
				return err
			}
		}
		if err := e.store.UpdateTurnStatusAndPhase(opCtx, claim.TurnID, status, phase); err != nil {
			return err
		}
		if markFinished {
			if err := e.store.MarkTurnFinished(opCtx, claim.TurnID); err != nil {
				return err
			}
		}
	}
	logutil.WarnIfErr("turn recovered event append", e.store.AppendTurnEvent(opCtx, claim.TurnID, claim.SessionID, "turn.recovered", payload))
	if err := e.store.ReleaseSessionActiveTurn(opCtx, claim.SessionID, claim.ClaimToken); err != nil {
		return err
	}
	if err := e.store.SyncSessionQueueCount(opCtx, claim.SessionID); err != nil {
		return err
	}
	queueCount, err := e.store.CountQueuedTurns(opCtx, claim.SessionID)
	if err != nil {
		return err
	}
	sessionStatus := "idle"
	if queueCount > 0 {
		sessionStatus = "queued"
	}
	if err := e.store.TouchSessionState(opCtx, claim.SessionID, map[string]any{"active_turn_id": nil, "status": sessionStatus}); err != nil {
		return err
	}
	turnRec, err := e.store.GetTurn(opCtx, claim.TurnID)
	if err != nil {
		return err
	}
	runner := e.runner(claim.SessionID)
	agentID, model := runner.resolveTurnAgentAndModel(opCtx, e.store, turnRec, claim.SessionID, turnRec.Prompt)
	e.PublishRuntimeTurnEvent("turn_recovered", claim.SessionID, claim.TurnID, agentID, status, phase, cloneMap(payload))
	runner.emitTurnStateHook(opCtx, claim.SessionID, claim.TurnID, agentID, model, status, phase, cloneMap(payload))
	runner.emitSessionStateHook(opCtx, claim.SessionID, agentID, model, sessionStatus, map[string]any{"reason": "recovery", "recovery_disposition": disposition, "stale_claim": true, "active_turn_id": nil, "turn_id": claim.TurnID, "turn_status": status, "turn_phase": phase})
	return nil
}

func (e *Engine) startNextQueuedTurnLocked(ctx context.Context, runner *sessionRunner, sessionID string) (bool, error) {
	if sessionID == "" {
		return false, nil
	}
	coordCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if activeTurnID, _, err := e.store.GetSessionActiveTurn(coordCtx, sessionID); err == nil {
		if err := e.normalizeRunningSessionState(coordCtx, sessionID, activeTurnID, true, ""); err != nil {
			return false, err
		}
		return false, nil
	} else if err != sql.ErrNoRows {
		return false, err
	}
	next, err := e.store.GetNextQueuedTurn(coordCtx, sessionID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if activeTurnID, _, err := e.store.GetSessionActiveTurn(coordCtx, sessionID); err == nil {
		if err := e.normalizeRunningSessionState(coordCtx, sessionID, activeTurnID, true, ""); err != nil {
			return false, err
		}
		return false, nil
	} else if err != sql.ErrNoRows {
		return false, err
	}
	launched, err := e.launchTurnLocked(coordCtx, runner, sessionID, next.ID)
	if err != nil {
		return false, err
	}
	return launched, nil
}

func (e *Engine) startNextQueuedTurn(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	runner := e.runner(sessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	_, err := e.startNextQueuedTurnLocked(ctx, runner, sessionID)
	return err
}

func (r *sessionRunner) heartbeatActiveTurn(ctx context.Context, sessionID, claimToken string, cancel context.CancelFunc) {
	bgCtx := r.engine.backgroundContext()
	ticker := time.NewTicker(activeTurnHeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.store.TouchSessionActiveTurn(bgCtx, sessionID, claimToken); err != nil {
				log.Printf("turn heartbeat: %v", err)
				if errors.Is(err, sql.ErrNoRows) {
					cancel()
					return
				}
			}
		}
	}
}

func staleDurationSeconds(updatedAt string) (float64, error) {
	ts, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return 0, err
	}
	return time.Since(ts).Seconds(), nil
}

const skippedDueToQueuedUserMessage = "Skipped due to queued user message."

func normalizeSteeringRole(role string) string {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "assistant":
		return "assistant"
	case "system":
		return "system"
	default:
		return "user"
	}
}

func persistedSteeringChatRole(role string) string {
	if normalizeSteeringRole(role) == "assistant" {
		return "assistant"
	}
	return "user"
}

func steeringMessagesToMetadata(msgs []store.SteeringMessage) []map[string]any {
	out := make([]map[string]any, 0, len(msgs))
	for _, msg := range msgs {
		out = append(out, map[string]any{
			"role":       msg.Role,
			"content":    msg.Content,
			"payload":    msg.Payload,
			"media":      msg.Media,
			"queue_mode": msg.QueueMode,
		})
	}
	return out
}

func steeringMessagesFromMetadata(metadata map[string]any) []store.SteeringMessage {
	if metadata == nil {
		return nil
	}
	raw, ok := metadata["initial_steering"]
	if !ok || raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]store.SteeringMessage, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		payload, _ := m["payload"].(map[string]any)
		media := []string{}
		if rawMedia, ok := m["media"].([]any); ok {
			for _, v := range rawMedia {
				if s, ok := v.(string); ok && s != "" {
					media = append(media, s)
				}
			}
		}
		out = append(out, store.SteeringMessage{
			Role:      normalizeSteeringRole(internalx.StringValue(m["role"], "user")),
			Content:   internalx.StringValue(m["content"], ""),
			Payload:   payload,
			Media:     media,
			QueueMode: internalx.StringValue(m["queue_mode"], "one-at-a-time"),
		})
	}
	return out
}

func (e *Engine) submitSteeringPrompt(ctx context.Context, sessionID, activeTurnID string, in RunInput) (*SubmitResult, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if strings.TrimSpace(sessionID) != "" && strings.TrimSpace(activeTurnID) != "" {
		e.normalizeRunningSessionState(opCtx, sessionID, activeTurnID, true, "")
	}
	payload := map[string]any{"intent": in.Intent, "model": in.Model, "kind": "steering", "active_turn_id": activeTurnID}
	if in.ParentTurnID != "" {
		payload["parent_turn_id"] = in.ParentTurnID
	}
	for k, v := range in.Metadata {
		payload[k] = v
	}
	media := steeringMediaFromMetadata(in.Metadata)
	queueMode := internalx.StringValue(in.Metadata["steering_mode"], "one-at-a-time")
	steeringRole := normalizeSteeringRole(internalx.StringValue(in.Metadata["ingress_role"], "user"))
	if _, err := e.store.EnqueueSteering(opCtx, sessionID, activeTurnID, steeringRole, in.Prompt, payload, media, queueMode); err != nil {
		logutil.WarnIfErr("append steering.rejected event", e.store.AppendTurnEvent(e.backgroundContext(), activeTurnID, sessionID, "steering.rejected", map[string]any{
			"phase":       "steering",
			"checkpoint":  true,
			"reason":      err.Error(),
			"queue_mode":  queueMode,
			"content":     in.Prompt,
			"media_count": len(media),
		}))
		e.broadcast(sessionID, map[string]any{
			"type":        "steering_rejected",
			"chat_jid":    "gi:" + sessionID,
			"turn_id":     activeTurnID,
			"queue_mode":  queueMode,
			"content_len": len(in.Prompt),
			"media_count": len(media),
			"reason":      err.Error(),
		})
		return nil, err
	}
	logutil.WarnIfErr("append steering.enqueued event", e.store.AppendTurnEvent(opCtx, activeTurnID, sessionID, "steering.enqueued", map[string]any{
		"phase":       "steering",
		"checkpoint":  true,
		"content":     in.Prompt,
		"queue_mode":  queueMode,
		"media_count": len(media),
	}))
	e.broadcast(sessionID, map[string]any{
		"type":        "steering_enqueued",
		"chat_jid":    "gi:" + sessionID,
		"turn_id":     activeTurnID,
		"queue_mode":  queueMode,
		"content_len": len(in.Prompt),
		"media_count": len(media),
	})
	return &SubmitResult{TurnID: activeTurnID, SessionID: sessionID, Status: "running", Queued: false}, nil
}

func steeringMetadataFromMessages(msgs []store.SteeringMessage) map[string]any {
	metadata := map[string]any{
		"initial_steering": steeringMessagesToMetadata(msgs),
		"continue":         true,
	}
	if len(msgs) > 0 && msgs[0].Payload != nil {
		for _, key := range []string{"intent", "model", "parent_turn_id", "source_session_id", "source_agent_id", "target_agent_id", "route_mode", "route_matched_by"} {
			if value, ok := msgs[0].Payload[key]; ok {
				metadata[key] = value
			}
		}
	}
	return metadata
}

func (e *Engine) stageQueuedSteeringContinuation(ctx context.Context, sessionID string) (bool, string, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	turnID := store.NowID("turn")
	turnRec, msgs, err := e.store.StageSteeringContinuation(opCtx, sessionID, turnID)
	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	metadata := turnRec.Metadata
	submittedPayload := map[string]any{"phase": "queue", "intent": internalx.StringValue(metadata["intent"], "continue"), "queued": true, "checkpoint": true, "continue": true}
	logutil.WarnIfErr("append continued turn.submitted event", e.store.AppendTurnEvent(opCtx, turnID, sessionID, "turn.submitted", submittedPayload))
	e.PublishRuntimeTurnEvent("turn_submitted", sessionID, turnID, "", "queued", "queued", submittedPayload)
	if err := e.normalizeInactiveSessionState(opCtx, sessionID, "queued", internalx.StringValue(metadata["model"], ""), true); err != nil {
		return false, "", err
	}
	logutil.WarnIfErr("append steering.continue_staged event", e.store.AppendTurnEvent(opCtx, turnID, sessionID, "steering.continue_staged", map[string]any{"phase": "steering", "checkpoint": true, "count": len(msgs)}))
	e.broadcast(sessionID, map[string]any{"type": "steering_continue_staged", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "count": len(msgs)})
	return true, turnID, nil
}

func (e *Engine) continueQueuedSteeringLocked(ctx context.Context, runner *sessionRunner, sessionID string) (bool, error) {
	staged, turnID, err := e.stageQueuedSteeringContinuation(ctx, sessionID)
	if err != nil {
		return false, err
	}
	if !staged {
		return false, nil
	}
	coordCtx := store.CoordinationContext(ctx, e.backgroundContext())
	launched, err := e.startNextQueuedTurnLocked(coordCtx, runner, sessionID)
	if err != nil {
		return false, err
	}
	if launched {
		if err := e.store.SyncSessionQueueCount(coordCtx, sessionID); err != nil {
			return false, err
		}
		logutil.WarnIfErr("append steering.continued event", e.store.AppendTurnEvent(coordCtx, turnID, sessionID, "steering.continued", map[string]any{"phase": "steering", "checkpoint": true, "handoff": "launched"}))
		e.broadcast(sessionID, map[string]any{"type": "steering_continued", "chat_jid": "gi:" + sessionID, "turn_id": turnID})
		return true, nil
	}
	activeTurnID, _, err := e.store.GetSessionActiveTurn(coordCtx, sessionID)
	if err == nil {
		if err := e.normalizeRunningSessionState(coordCtx, sessionID, activeTurnID, true, ""); err != nil {
			return false, err
		}
		logutil.WarnIfErr("append steering.continued event", e.store.AppendTurnEvent(coordCtx, turnID, sessionID, "steering.continued", map[string]any{"phase": "steering", "checkpoint": true, "handoff": "active_claim"}))
		return true, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return true, nil
}

func (e *Engine) continueQueuedSteering(ctx context.Context, sessionID string) (bool, error) {
	runner := e.runner(sessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	return e.continueQueuedSteeringLocked(ctx, runner, sessionID)
}

func (e *Engine) ContinueSession(ctx context.Context, sessionID string) (bool, error) {
	runner := e.runner(sessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	coordCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if activeTurnID, _, err := e.store.GetSessionActiveTurn(coordCtx, sessionID); err == nil {
		if err := e.normalizeRunningSessionState(coordCtx, sessionID, activeTurnID, true, ""); err != nil {
			return false, err
		}
		return false, nil
	} else if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	launched, err := e.startNextQueuedTurnLocked(coordCtx, runner, sessionID)
	if err != nil {
		return false, err
	}
	if launched {
		return true, nil
	}
	if activeTurnID, _, err := e.store.GetSessionActiveTurn(coordCtx, sessionID); err == nil {
		if err := e.normalizeRunningSessionState(coordCtx, sessionID, activeTurnID, true, ""); err != nil {
			return false, err
		}
		return true, nil
	} else if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	continued, err := e.continueQueuedSteeringLocked(coordCtx, runner, sessionID)
	if err != nil {
		return false, err
	}
	if continued {
		return true, nil
	}
	if err := e.normalizeInactiveSessionState(coordCtx, sessionID, "idle", "", true); err != nil {
		return false, err
	}
	return false, nil
}

func (r *sessionRunner) dequeueSteeringMessages(ctx context.Context, sessionID string) ([]store.SteeringMessage, error) {
	msgs, err := r.store.DequeueSteering(ctx, sessionID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(msgs) > 0 {
		r.engine.broadcast(sessionID, map[string]any{"type": "steering_dequeued", "chat_jid": "gi:" + sessionID, "count": len(msgs)})
	}
	return msgs, nil
}

func (r *sessionRunner) persistSteeringMessages(ctx context.Context, sessionID, turnID string, msgs []store.SteeringMessage) int {
	if len(msgs) == 0 {
		return 0
	}
	totalContentLen := 0
	for _, msg := range msgs {
		role := persistedSteeringChatRole(msg.Role)
		payload := map[string]any{"kind": "chat", "intent": internalx.StringValue(msg.Payload["intent"], "prompt"), "turn_id": turnID, "steering": true, "steering_role": normalizeSteeringRole(msg.Role)}
		for k, v := range msg.Payload {
			payload[k] = v
		}
		if len(msg.Media) > 0 {
			payload["media"] = append([]string(nil), msg.Media...)
		}
		logutil.WarnIfErr("add steering message", r.store.AddMessage(ctx, store.NowID("msg"), sessionID, role, msg.Content, payload))
		totalContentLen += len(msg.Content)
	}
	logutil.WarnIfErr("append steering.injected event", r.store.AppendTurnEvent(ctx, turnID, sessionID, "steering.injected", map[string]any{
		"phase":             "steering",
		"checkpoint":        true,
		"count":             len(msgs),
		"total_content_len": totalContentLen,
	}))
	r.engine.broadcast(sessionID, map[string]any{"type": "steering_injected", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "count": len(msgs), "media_count": steeringMediaCount(msgs)})
	return len(msgs)
}

func (r *sessionRunner) injectSteeringMessages(ctx context.Context, sessionID, turnID string, convCtx *goai.Context, msgs []store.SteeringMessage) int {
	if len(msgs) == 0 {
		return 0
	}
	for _, msg := range msgs {
		role := normalizeSteeringRole(msg.Role)
		content := msg.Content
		if len(msg.Media) > 0 {
			if strings.TrimSpace(content) == "" {
				content = "[user provided media attachments]"
			} else {
				content += "\n\n[media attachments included]"
			}
		}
		switch role {
		case "assistant":
			convCtx.Messages = append(convCtx.Messages, goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: content}}})
		default:
			convCtx.Messages = append(convCtx.Messages, goai.UserMessage(content))
		}
	}
	return r.persistSteeringMessages(ctx, sessionID, turnID, msgs)
}

func (r *sessionRunner) skipRemainingToolCalls(ctx context.Context, sessionID, turnID string, convCtx *goai.Context, toolCalls []goai.ToolCall, start int) {
	for i := start; i < len(toolCalls); i++ {
		call := toolCalls[i]
		goai.AppendToolResult(convCtx, call.ID, call.Name, skippedDueToQueuedUserMessage, true)
		logutil.WarnIfErr("append tool.skipped event", r.store.AppendTurnEvent(ctx, turnID, sessionID, "tool.skipped", map[string]any{
			"phase":        "tool",
			"checkpoint":   true,
			"tool":         call.Name,
			"tool_call_id": call.ID,
			"reason":       "queued user steering message",
		}))
		logutil.WarnIfErr("add skipped tool_result message", r.store.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", skippedDueToQueuedUserMessage, map[string]any{
			"kind":         "tool_result",
			"tool_call_id": call.ID,
			"tool_name":    call.Name,
			"is_error":     true,
			"turn_id":      turnID,
			"skipped":      true,
			"skip_reason":  "queued user steering message",
		}))
		r.engine.broadcast(sessionID, map[string]any{"type": "tool_skipped", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "reason": "queued user steering message"})
		r.engine.PublishRuntimeToolEvent("tool_skipped", sessionID, turnID, "", call.Name, call.ID, 0, nil, map[string]any{"reason": "queued user steering message", "phase": "tool"})
	}
}

func steeringMediaFromMetadata(metadata map[string]any) []string {
	if metadata == nil {
		return nil
	}
	raw, ok := metadata["media"]
	if !ok || raw == nil {
		return nil
	}
	switch m := raw.(type) {
	case []string:
		out := make([]string, 0, len(m))
		for _, item := range m {
			if strings.TrimSpace(item) != "" {
				out = append(out, item)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(m))
		for _, item := range m {
			s, ok := item.(string)
			if !ok || strings.TrimSpace(s) == "" {
				continue
			}
			out = append(out, s)
		}
		return out
	default:
		return nil
	}
}

func steeringMediaCount(msgs []store.SteeringMessage) int {
	total := 0
	for _, msg := range msgs {
		total += len(msg.Media)
	}
	return total
}

func steeringPromptForShell(msgs []store.SteeringMessage) string {
	parts := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		if strings.TrimSpace(msg.Content) != "" {
			parts = append(parts, msg.Content)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

// Hook names model the non-UX lifecycle surface exposed by the turn engine.
const (
	HookSessionStart          = "session_start"
	HookSessionShutdown       = "session_shutdown"
	HookSessionBeforeSwitch   = "session_before_switch"
	HookSessionBeforeFork     = "session_before_fork"
	HookSessionBeforeTree     = "session_before_tree"
	HookSessionBeforeCompact  = "session_before_compact"
	HookSessionCompact        = "session_compact"
	HookSessionTree           = "session_tree"
	HookResourcesDiscover     = "resources_discover"
	HookInput                 = "input"
	HookBeforeAgentStart      = "before_agent_start"
	HookAgentStart            = "agent_start"
	HookAgentEnd              = "agent_end"
	HookTurnStart             = "turn_start"
	HookTurnEnd               = "turn_end"
	HookTurnState             = "turn_state"
	HookSessionState          = "session_state"
	HookContext               = "context"
	HookBeforeProviderRequest = "before_provider_request"
	HookAfterProviderResponse = "after_provider_response"
	HookMessageStart          = "message_start"
	HookMessageUpdate         = "message_update"
	HookMessageEnd            = "message_end"
	HookToolExecutionStart    = "tool_execution_start"
	HookToolCall              = "tool_call"
	HookApproveTool           = "approve_tool"
	HookToolExecutionUpdate   = "tool_execution_update"
	HookToolResult            = "tool_result"
	HookToolExecutionEnd      = "tool_execution_end"
	HookModelSelect           = "model_select"
	HookUserBash              = "user_bash"
)

type HookTrace struct {
	ID        string `json:"id,omitempty"`
	EmittedAt string `json:"emitted_at,omitempty"`
}

// HookRequest is the typed envelope delivered to engine hooks. Fields are
// intentionally broad so the same structure can cover observation, gates, and
// mutation hooks without adding a new Go type per hook name.
type HookRequest struct {
	Name          string         `json:"name"`
	SessionID     string         `json:"session_id,omitempty"`
	TurnID        string         `json:"turn_id,omitempty"`
	AgentID       string         `json:"agent_id,omitempty"`
	Model         string         `json:"model,omitempty"`
	Iteration     int            `json:"iteration,omitempty"`
	SessionStatus string         `json:"session_status,omitempty"`
	TurnStatus    string         `json:"turn_status,omitempty"`
	TurnPhase     string         `json:"turn_phase,omitempty"`
	Payload       map[string]any `json:"payload,omitempty"`
	Trace         HookTrace      `json:"trace,omitempty"`

	SystemPrompt string         `json:"system_prompt,omitempty"`
	Messages     []goai.Message `json:"messages,omitempty"`
	Tools        []goai.Tool    `json:"tools,omitempty"`
	ToolCall     *goai.ToolCall `json:"tool_call,omitempty"`
	ToolResult   string         `json:"tool_result,omitempty"`
	ToolError    bool           `json:"tool_error,omitempty"`
}

// HookResponse is merged into the running turn. Mutation hooks are chained in
// registration order; gate hooks stop at the first blocking response.
type HookResponse struct {
	Action       string         `json:"action,omitempty"`
	Cancel       bool           `json:"cancel,omitempty"`
	Block        bool           `json:"block,omitempty"`
	Handled      bool           `json:"handled,omitempty"`
	Reason       string         `json:"reason,omitempty"`
	Payload      map[string]any `json:"payload,omitempty"`
	Message      string         `json:"message,omitempty"`
	SystemPrompt string         `json:"system_prompt,omitempty"`
	Messages     []goai.Message `json:"messages,omitempty"`
	Tools        []goai.Tool    `json:"tools,omitempty"`
	ToolCall     *goai.ToolCall `json:"tool_call,omitempty"`
	ToolResult   *string        `json:"tool_result,omitempty"`
}

// HookHandler is a synchronous engine hook callback.
type HookHandler func(context.Context, HookRequest) (HookResponse, error)

type registeredHook struct {
	id      uint64
	source  string
	handler HookHandler
}

type HookExecutionError struct {
	Name      string
	Source    string
	Kind      string
	Trace     HookTrace
	TimeoutMS int
	Cause     error
}

func (e HookExecutionError) Error() string {
	switch e.Kind {
	case "timeout":
		return fmt.Sprintf("hook %s from %s timed out after %dms", e.Name, e.Source, e.TimeoutMS)
	case "panic":
		return fmt.Sprintf("hook %s from %s panicked: %v", e.Name, e.Source, e.Cause)
	default:
		return fmt.Sprintf("hook %s from %s failed: %v", e.Name, e.Source, e.Cause)
	}
}

func (e HookExecutionError) Unwrap() error { return e.Cause }

// HookRegistry stores hook callbacks. Handlers are copied before invocation so
// hooks can register/unregister safely outside the call path.
type HookRegistry struct {
	mu     sync.RWMutex
	nextID uint64
	hooks  map[string][]registeredHook
}

var hookTraceSeq atomic.Uint64

func applyHookDefaultsCompat(settings config.HookSettings) config.HookSettings {
	if settings.TimeoutMS <= 0 {
		settings.TimeoutMS = 1500
	}
	settings.OnError = config.NormalizeHookPolicy(settings.OnError, "error")
	settings.OnTimeout = config.NormalizeHookPolicy(settings.OnTimeout, "continue")
	return settings
}

func NewHookRegistry() *HookRegistry {
	return &HookRegistry{hooks: make(map[string][]registeredHook)}
}

func normalizeHookName(name string) string {
	switch strings.TrimSpace(strings.ToLower(name)) {
	case "before_llm":
		return HookBeforeProviderRequest
	case "after_llm":
		return HookAfterProviderResponse
	case "before_tool":
		return HookToolCall
	case "after_tool":
		return HookToolResult
	default:
		return strings.TrimSpace(name)
	}
}

func (r *HookRegistry) Register(name, source string, handler HookHandler) (func(), error) {
	name = normalizeHookName(name)
	if name == "" {
		return nil, fmt.Errorf("hook name is required")
	}
	if handler == nil {
		return nil, fmt.Errorf("hook handler is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	id := r.nextID
	r.hooks[name] = append(r.hooks[name], registeredHook{id: id, source: source, handler: handler})
	return func() { r.Unregister(name, id) }, nil
}

func (r *HookRegistry) Unregister(name string, id uint64) {
	name = normalizeHookName(name)
	r.mu.Lock()
	defer r.mu.Unlock()
	items := r.hooks[name]
	for i, item := range items {
		if item.id == id {
			r.hooks[name] = append(items[:i], items[i+1:]...)
			break
		}
	}
	if len(r.hooks[name]) == 0 {
		delete(r.hooks, name)
	}
}

func (r *HookRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks = make(map[string][]registeredHook)
}

func (r *HookRegistry) Handlers(name string) []registeredHook {
	name = normalizeHookName(name)
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []registeredHook
	out = append(out, r.hooks[name]...)
	if name != "*" {
		out = append(out, r.hooks["*"]...)
	}
	return out
}

func (r *HookRegistry) Infos() []HookInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []HookInfo
	for name, hooks := range r.hooks {
		for _, hook := range hooks {
			out = append(out, HookInfo{Name: name, Source: hook.source, ID: hook.id})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Name == out[j].Name {
			return out[i].ID < out[j].ID
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func (e *Engine) RegisterHook(name, source string, handler HookHandler) (func(), error) {
	return e.hooks.Register(name, source, handler)
}

func (e *Engine) ClearHooks() { e.hooks.Clear() }

func nextHookTrace() HookTrace {
	seq := hookTraceSeq.Add(1)
	now := time.Now().UTC()
	return HookTrace{ID: fmt.Sprintf("hook_%d_%d", now.UnixNano(), seq), EmittedAt: now.Format(time.RFC3339Nano)}
}

func hookScriptPayload(req HookRequest) map[string]any {
	payload := map[string]any{
		"hook":           req.Name,
		"name":           req.Name,
		"session_id":     req.SessionID,
		"turn_id":        req.TurnID,
		"agent_id":       req.AgentID,
		"model":          req.Model,
		"iteration":      req.Iteration,
		"session_status": req.SessionStatus,
		"turn_status":    req.TurnStatus,
		"turn_phase":     req.TurnPhase,
		"system_prompt":  req.SystemPrompt,
		"messages":       req.Messages,
		"tools":          req.Tools,
		"tool_call":      req.ToolCall,
		"tool_result":    req.ToolResult,
		"tool_error":     req.ToolError,
		"payload":        req.Payload,
		"trace":          req.Trace,
	}
	return payload
}

func (e *Engine) invokeHookHandler(ctx context.Context, req HookRequest, item registeredHook) (HookResponse, int, error) {
	startedAt := time.Now()
	timeoutMS := e.runtimeCfg.Hooks.TimeoutMS
	if timeoutMS <= 0 {
		timeoutMS = 1500
	}
	hookCtx := ctx
	var cancel context.CancelFunc
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		hookCtx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMS)*time.Millisecond)
		defer cancel()
	}
	type hookResult struct {
		resp HookResponse
		err  error
	}
	resultCh := make(chan hookResult, 1)
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				resultCh <- hookResult{err: HookExecutionError{Name: req.Name, Source: item.source, Kind: "panic", Trace: req.Trace, TimeoutMS: timeoutMS, Cause: fmt.Errorf("%v", recovered)}}
			}
		}()
		resp, err := item.handler(hookCtx, req)
		if err != nil {
			kind := "handler_error"
			if errors.Is(err, context.DeadlineExceeded) && ctx.Err() == nil {
				kind = "timeout"
			}
			resultCh <- hookResult{err: HookExecutionError{Name: req.Name, Source: item.source, Kind: kind, Trace: req.Trace, TimeoutMS: timeoutMS, Cause: err}}
			return
		}
		resultCh <- hookResult{resp: resp}
	}()
	select {
	case <-hookCtx.Done():
		durationMS := int(time.Since(startedAt).Milliseconds())
		if ctx.Err() != nil {
			return HookResponse{}, durationMS, ctx.Err()
		}
		return HookResponse{}, durationMS, HookExecutionError{Name: req.Name, Source: item.source, Kind: "timeout", Trace: req.Trace, TimeoutMS: timeoutMS, Cause: hookCtx.Err()}
	case res := <-resultCh:
		return res.resp, int(time.Since(startedAt).Milliseconds()), res.err
	}
}

func (e *Engine) applyHookFailurePolicy(req HookRequest, item registeredHook, err error) error {
	execErr := HookExecutionError{Name: req.Name, Source: item.source, Kind: "handler_error", Trace: req.Trace, Cause: err}
	if typed, ok := err.(HookExecutionError); ok {
		execErr = typed
	}
	policy := e.runtimeCfg.Hooks.OnError
	if execErr.Kind == "timeout" {
		policy = e.runtimeCfg.Hooks.OnTimeout
	}
	if config.NormalizeHookPolicy(policy, "error") == "continue" {
		log.Printf("hook %s from %s: %v (continuing by policy)", req.Name, item.source, execErr)
		return nil
	}
	return execErr
}

func hookInvocationAction(resp HookResponse, err error, continued bool) string {
	if err != nil {
		if continued {
			return "continue"
		}
		return "error"
	}
	if strings.TrimSpace(resp.Action) != "" {
		return strings.ToLower(strings.TrimSpace(resp.Action))
	}
	if resp.Cancel {
		return "abort_turn"
	}
	if resp.Block {
		return "deny"
	}
	if resp.Handled {
		return "respond"
	}
	return "continue"
}

func hookInvocationErrorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (e *Engine) recordHookInvocation(ctx context.Context, req HookRequest, item registeredHook, action string, response HookResponse, err error, durationMS int) {
	e.PublishRuntimeHookEvent("hook_invocation", req, item.source, action, durationMS, err)
	if e.store == nil {
		return
	}
	_, recordErr := storeaudit.RecordHookInvocation(ctx, e.store.DB(), req.TurnID, req.SessionID, req.Name, req.Name, item.source, action, req, response, hookInvocationErrorText(err), durationMS)
	logutil.WarnIfErr("record hook invocation", recordErr)
}

func (e *Engine) emitHook(ctx context.Context, req HookRequest) (HookResponse, error) {
	req.Name = normalizeHookName(req.Name)
	if req.Name == "" {
		return HookResponse{}, fmt.Errorf("hook name is required")
	}
	if req.Trace.ID == "" || req.Trace.EmittedAt == "" {
		req.Trace = nextHookTrace()
	}
	var merged HookResponse
	for _, item := range e.hooks.Handlers(req.Name) {
		resp, durationMS, err := e.invokeHookHandler(ctx, req, item)
		if err != nil {
			policyErr := e.applyHookFailurePolicy(req, item, err)
			e.recordHookInvocation(e.backgroundContext(), req, item, hookInvocationAction(HookResponse{}, err, policyErr == nil), HookResponse{}, err, durationMS)
			if policyErr != nil {
				return merged, policyErr
			}
			continue
		}
		e.recordHookInvocation(e.backgroundContext(), req, item, hookInvocationAction(resp, nil, false), resp, nil, durationMS)
		if resp.Action != "" {
			merged.Action = strings.ToLower(strings.TrimSpace(resp.Action))
		}
		if resp.Payload != nil {
			merged.Payload = resp.Payload
			req.Payload = resp.Payload
		}
		if resp.Message != "" {
			merged.Message = resp.Message
		}
		if resp.SystemPrompt != "" {
			merged.SystemPrompt = resp.SystemPrompt
			req.SystemPrompt = resp.SystemPrompt
		}
		if resp.Messages != nil {
			merged.Messages = resp.Messages
			req.Messages = resp.Messages
		}
		if resp.Tools != nil {
			merged.Tools = resp.Tools
			req.Tools = resp.Tools
		}
		if resp.ToolCall != nil {
			merged.ToolCall = resp.ToolCall
			req.ToolCall = resp.ToolCall
		}
		if resp.ToolResult != nil {
			merged.ToolResult = resp.ToolResult
			req.ToolResult = *resp.ToolResult
		}
		if resp.Handled {
			merged.Handled = true
		}
		if resp.Cancel || resp.Block {
			merged.Cancel = resp.Cancel
			merged.Block = resp.Block
			merged.Reason = resp.Reason
			return merged, nil
		}
		if resp.Reason != "" {
			merged.Reason = resp.Reason
		}
	}
	return merged, nil
}

func hookResponseFromScript(result string) (HookResponse, error) {
	var resp HookResponse
	if strings.TrimSpace(result) == "" {
		return resp, nil
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(result), &raw); err != nil {
		return resp, nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return resp, nil
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return resp, nil
	}
	if tc, ok := raw["tool_call"]; ok {
		if call := decodeToolCall(tc); call != nil {
			resp.ToolCall = call
		}
	}
	if payload, ok := raw["payload"].(map[string]any); ok {
		if tc, ok := payload["tool_call"]; ok {
			if call := decodeToolCall(tc); call != nil {
				resp.ToolCall = call
			}
		}
	}
	resp.Action = strings.ToLower(strings.TrimSpace(resp.Action))
	if resp.Action == "" {
		if action := internalx.StringValue(raw["action"], ""); strings.TrimSpace(action) != "" {
			resp.Action = strings.ToLower(strings.TrimSpace(action))
		}
	}
	switch resp.Action {
	case "continue":
	case "modify":
	case "respond":
		resp.Handled = true
		if strings.TrimSpace(resp.Message) == "" {
			if response := internalx.StringValue(raw["response"], ""); strings.TrimSpace(response) != "" {
				resp.Message = response
			} else if payload, ok := raw["payload"].(map[string]any); ok {
				resp.Message = internalx.StringValue(payload["response"], "")
			}
		}
	case "deny":
		resp.Block = true
		if strings.TrimSpace(resp.Reason) == "" {
			resp.Reason = "denied by hook"
		}
	case "abort_turn":
		resp.Cancel = true
		resp.Block = true
		if strings.TrimSpace(resp.Reason) == "" {
			resp.Reason = "aborted by hook"
		}
	case "hard_abort":
		resp.Cancel = true
		resp.Block = true
		if resp.Payload == nil {
			resp.Payload = map[string]any{}
		}
		resp.Payload["hard_abort"] = true
		if strings.TrimSpace(resp.Reason) == "" {
			resp.Reason = "hard aborted by hook"
		}
	}
	return resp, nil
}

func decodeToolCall(raw any) *goai.ToolCall {
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var call goai.ToolCall
	if err := json.Unmarshal(b, &call); err != nil {
		return nil
	}
	if call.Name == "" {
		var alt struct {
			ID        string         `json:"id"`
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(b, &alt); err != nil || alt.Name == "" {
			return nil
		}
		call.ID = alt.ID
		call.Name = alt.Name
		call.Arguments = alt.Arguments
	}
	if call.Arguments == nil {
		call.Arguments = map[string]any{}
	}
	if call.Type == "" {
		call.Type = "toolCall"
	}
	return &call
}

func (r *sessionRunner) emitTurnStateHook(ctx context.Context, sessionID, turnID, agentID, model, status, phase string, payload map[string]any) {
	r.emitTurnState(ctx, sessionID, turnID, agentID, model, status, phase, payload, true)
}

func (r *sessionRunner) emitTurnStateHookOnly(ctx context.Context, sessionID, turnID, agentID, model, status, phase string, payload map[string]any) {
	r.emitTurnState(ctx, sessionID, turnID, agentID, model, status, phase, payload, false)
}

func (r *sessionRunner) emitTurnState(ctx context.Context, sessionID, turnID, agentID, model, status, phase string, payload map[string]any, publishTopic bool) {
	if stringsMapEmpty(payload) {
		payload = map[string]any{}
	}
	if publishTopic {
		r.engine.PublishRuntimeTurnEvent("turn_state", sessionID, turnID, agentID, status, phase, payload)
	}
	if _, err := r.engine.emitHook(ctx, HookRequest{
		Name:       HookTurnState,
		SessionID:  sessionID,
		TurnID:     turnID,
		AgentID:    agentID,
		Model:      model,
		TurnStatus: status,
		TurnPhase:  phase,
		Payload:    payload,
	}); err != nil {
		log.Printf("hook turn_state error: %v", err)
	}
}

func (r *sessionRunner) emitSessionStateHook(ctx context.Context, sessionID, agentID, model, status string, payload map[string]any) {
	r.emitSessionState(ctx, sessionID, agentID, model, status, payload, true)
}

func (r *sessionRunner) emitSessionStateHookOnly(ctx context.Context, sessionID, agentID, model, status string, payload map[string]any) {
	r.emitSessionState(ctx, sessionID, agentID, model, status, payload, false)
}

func (r *sessionRunner) emitSessionState(ctx context.Context, sessionID, agentID, model, status string, payload map[string]any, publishTopic bool) {
	if stringsMapEmpty(payload) {
		payload = map[string]any{}
	}
	if publishTopic {
		r.engine.PublishRuntimeSessionEvent("session_state", sessionID, agentID, status, payload)
	}
	if _, err := r.engine.emitHook(ctx, HookRequest{
		Name:          HookSessionState,
		SessionID:     sessionID,
		AgentID:       agentID,
		Model:         model,
		SessionStatus: status,
		Payload:       payload,
	}); err != nil {
		log.Printf("hook session_state error: %v", err)
	}
}

func stringsMapEmpty(v map[string]any) bool {
	return len(v) == 0
}

const processHookProtocol = "hook-jsonrpc-v1"

type processHookRPCRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      int            `json:"id"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
}

type processHookRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type processHookRPCResponse struct {
	JSONRPC string               `json:"jsonrpc"`
	ID      int                  `json:"id"`
	Result  json.RawMessage      `json:"result,omitempty"`
	Error   *processHookRPCError `json:"error,omitempty"`
}

type processHookHelloResult struct {
	OK   bool   `json:"ok"`
	Name string `json:"name,omitempty"`
}

func isProcessHookSpec(spec scripting.EventHookSpec) bool {
	if strings.EqualFold(strings.TrimSpace(spec.Engine), "process") {
		return true
	}
	return strings.TrimSpace(spec.Command) != ""
}

type mountedProcessHook struct {
	workspaceRoot string
	spec          scripting.EventHookSpec

	mu        sync.Mutex
	started   bool
	nextID    int
	stderrGen uint64
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	enc       *json.Encoder
	dec       *json.Decoder
	stderrMu  sync.Mutex
	stderr    bytes.Buffer
}

func newProcessHookHandler(workspaceRoot string, spec scripting.EventHookSpec) HookHandler {
	mounted := &mountedProcessHook{workspaceRoot: workspaceRoot, spec: spec}
	return func(ctx context.Context, req HookRequest) (HookResponse, error) {
		return mounted.invoke(ctx, req)
	}
}

func invokeProcessHook(ctx context.Context, workspaceRoot string, spec scripting.EventHookSpec, req HookRequest) (HookResponse, error) {
	mounted := &mountedProcessHook{workspaceRoot: workspaceRoot, spec: spec}
	return mounted.invoke(ctx, req)
}

func (m *mountedProcessHook) invoke(ctx context.Context, req HookRequest) (HookResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.ensureStartedLocked(ctx, req); err != nil {
		return HookResponse{}, err
	}
	m.nextID++
	invokeReq := processHookRPCRequest{
		JSONRPC: "2.0",
		ID:      m.nextID,
		Method:  processHookMethodName(req.Name),
		Params:  processHookPayload(req),
	}
	if err := m.enc.Encode(invokeReq); err != nil {
		stderrOut := m.stderrStringLocked()
		m.closeLocked()
		if stderrOut != "" {
			return HookResponse{}, fmt.Errorf("write process hook request: %w (%s)", err, stderrOut)
		}
		return HookResponse{}, fmt.Errorf("write process hook request: %w", err)
	}
	invokeResp, err := m.decodeResponseLocked(ctx)
	if err != nil {
		stderrOut := m.stderrStringLocked()
		m.closeLocked()
		if stderrOut != "" {
			return HookResponse{}, fmt.Errorf("read process hook response: %w (%s)", err, stderrOut)
		}
		return HookResponse{}, fmt.Errorf("read process hook response: %w", err)
	}
	if invokeResp.Error != nil {
		return HookResponse{}, fmt.Errorf("process hook %s failed: %s", req.Name, invokeResp.Error.Message)
	}
	resp, err := hookResponseFromScript(strings.TrimSpace(string(invokeResp.Result)))
	if err != nil {
		return HookResponse{}, err
	}
	return resp, nil
}

func (m *mountedProcessHook) ensureStartedLocked(ctx context.Context, req HookRequest) error {
	transport := strings.TrimSpace(m.spec.Transport)
	if transport == "" {
		transport = "stdio"
	}
	if transport != "stdio" {
		return fmt.Errorf("process hook transport not supported: %s", transport)
	}
	protocol := strings.TrimSpace(m.spec.Protocol)
	if protocol == "" {
		protocol = processHookProtocol
	}
	if protocol != processHookProtocol {
		return fmt.Errorf("process hook protocol not supported: %s", protocol)
	}
	if m.started && m.cmd != nil && m.cmd.Process != nil {
		return nil
	}
	command, err := resolveProcessHookCommand(m.workspaceRoot, m.spec.Command)
	if err != nil {
		return err
	}
	cmd := exec.Command(command, m.spec.Args...)
	cmd.Dir = resolveProcessHookDir(m.workspaceRoot, m.spec.CWD)
	cmd.Env = appendProcessHookEnv(os.Environ(), m.spec, req, m.workspaceRoot)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("process hook stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("process hook stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("process hook stderr: %w", err)
	}
	m.stderrMu.Lock()
	m.stderr.Reset()
	m.stderrGen++
	stderrGen := m.stderrGen
	m.stderrMu.Unlock()
	go func(gen uint64) {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stderr)
		m.stderrMu.Lock()
		defer m.stderrMu.Unlock()
		if gen != m.stderrGen {
			return
		}
		_, _ = m.stderr.Write(buf.Bytes())
	}(stderrGen)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start process hook: %w", err)
	}
	m.cmd = cmd
	m.stdin = stdin
	m.stdout = stdout
	m.enc = json.NewEncoder(stdin)
	m.enc.SetEscapeHTML(false)
	m.dec = json.NewDecoder(stdout)
	m.started = true
	m.nextID = 1
	helloReq := processHookRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "hook.hello",
		Params: map[string]any{
			"name":     tools.FirstNonEmpty(strings.TrimSpace(m.spec.Source), strings.TrimSpace(m.spec.Name), filepath.Base(command)),
			"version":  1,
			"modes":    processHookModes(req.Name),
			"protocol": processHookProtocol,
		},
	}
	if err := m.enc.Encode(helloReq); err != nil {
		m.closeLocked()
		return fmt.Errorf("write process hook hello: %w", err)
	}
	helloResp, err := m.decodeResponseLocked(ctx)
	if err != nil {
		m.closeLocked()
		return fmt.Errorf("read process hook hello: %w", err)
	}
	if helloResp.Error != nil {
		m.closeLocked()
		return fmt.Errorf("process hook hello failed: %s", helloResp.Error.Message)
	}
	var helloResult processHookHelloResult
	if err := json.Unmarshal(helloResp.Result, &helloResult); err != nil {
		m.closeLocked()
		return fmt.Errorf("decode process hook hello: %w", err)
	}
	if !helloResult.OK {
		m.closeLocked()
		return fmt.Errorf("process hook hello rejected hook %s", req.Name)
	}
	return nil
}

func (m *mountedProcessHook) decodeResponseLocked(ctx context.Context) (processHookRPCResponse, error) {
	type decoded struct {
		resp processHookRPCResponse
		err  error
	}
	dec := m.dec
	if dec == nil {
		return processHookRPCResponse{}, fmt.Errorf("process hook decoder not initialized")
	}
	ch := make(chan decoded, 1)
	go func(decoder *json.Decoder) {
		var resp processHookRPCResponse
		err := decoder.Decode(&resp)
		ch <- decoded{resp: resp, err: err}
	}(dec)
	select {
	case <-ctx.Done():
		m.closeLocked()
		return processHookRPCResponse{}, ctx.Err()
	case res := <-ch:
		return res.resp, res.err
	}
}

func (m *mountedProcessHook) stderrStringLocked() string {
	m.stderrMu.Lock()
	defer m.stderrMu.Unlock()
	return strings.TrimSpace(m.stderr.String())
}

func (m *mountedProcessHook) closeLocked() {
	m.stderrMu.Lock()
	m.stderrGen++
	m.stderrMu.Unlock()
	if m.stdin != nil {
		_ = m.stdin.Close()
	}
	if m.stdout != nil {
		_ = m.stdout.Close()
	}
	if m.cmd != nil {
		if m.cmd.Process != nil {
			_ = m.cmd.Process.Kill()
		}
		_ = m.cmd.Wait()
	}
	m.started = false
	m.cmd = nil
	m.stdin = nil
	m.stdout = nil
	m.enc = nil
	m.dec = nil
}

func resolveProcessHookCommand(workspaceRoot, command string) (string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", fmt.Errorf("process hook command is required")
	}
	if filepath.IsAbs(command) {
		return command, nil
	}
	if strings.ContainsRune(command, filepath.Separator) {
		return filepath.Join(workspaceRoot, command), nil
	}
	return command, nil
}

func resolveProcessHookDir(workspaceRoot, cwd string) string {
	cwd = strings.TrimSpace(cwd)
	if cwd == "" {
		return workspaceRoot
	}
	if filepath.IsAbs(cwd) {
		return cwd
	}
	return filepath.Join(workspaceRoot, cwd)
}

func appendProcessHookEnv(base []string, spec scripting.EventHookSpec, req HookRequest, workspaceRoot string) []string {
	env := append([]string{}, base...)
	env = append(env,
		"GI_HOOK_NAME="+req.Name,
		"GI_SESSION_ID="+req.SessionID,
		"GI_TURN_ID="+req.TurnID,
		"GI_AGENT_ID="+req.AgentID,
		"GI_MODEL="+req.Model,
		"GI_WORKSPACE_ROOT="+workspaceRoot,
	)
	for k, v := range spec.Env {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		env = append(env, key+"="+v)
	}
	return env
}

func processHookMethodName(name string) string {
	switch normalizeHookName(name) {
	case HookBeforeProviderRequest:
		return "hook.before_llm"
	case HookAfterProviderResponse:
		return "hook.after_llm"
	case HookToolCall:
		return "hook.before_tool"
	case HookToolResult:
		return "hook.after_tool"
	default:
		return "hook." + normalizeHookName(name)
	}
}

func processHookModes(name string) []string {
	switch normalizeHookName(name) {
	case HookToolCall, HookToolResult:
		return []string{"tool"}
	case HookApproveTool:
		return []string{"approve"}
	case HookBeforeProviderRequest, HookAfterProviderResponse:
		return []string{"llm"}
	default:
		return []string{"observe"}
	}
}

func processHookPayload(req HookRequest) map[string]any {
	payload := hookScriptPayload(req)
	payload["hook_phase"] = req.Name
	payload["hook_method"] = processHookMethodName(req.Name)
	if req.ToolCall != nil {
		payload["tool"] = req.ToolCall.Name
		payload["arguments"] = req.ToolCall.Arguments
	}
	if strings.TrimSpace(req.ToolResult) != "" || req.ToolError {
		payload["result"] = map[string]any{
			"for_llm":  req.ToolResult,
			"is_error": req.ToolError,
		}
	}
	if req.Name == HookBeforeProviderRequest || req.Name == normalizeHookName("before_llm") {
		if req.Payload != nil {
			if request, ok := req.Payload["request"]; ok && request != nil {
				payload["request"] = request
			} else {
				payload["request"] = map[string]any{
					"model":         req.Model,
					"system_prompt": req.SystemPrompt,
					"messages":      req.Messages,
					"tools":         req.Tools,
				}
			}
		} else {
			payload["request"] = map[string]any{
				"model":         req.Model,
				"system_prompt": req.SystemPrompt,
				"messages":      req.Messages,
				"tools":         req.Tools,
			}
		}
	}
	return payload
}

func (e *Engine) registerDefaultTools() {
	scriptTool := tools.NewScriptTool(e.store, e.runtimeCfg)
	scriptTool.SetConnectivityCallbacks(
		func(ctx context.Context, sessionID string, spec connectivity.RouteSpec) (connectivity.RouteInfo, error) {
			if spec.SessionID == "" {
				spec.SessionID = sessionID
			}
			return e.connectivity.Register(ctx, spec, func(ctx context.Context, event connectivity.EventEnvelope) (connectivity.RouteResponse, error) {
				payload := map[string]any{"event": event, "route": spec, "payload": event.Payload}
				input := tools.ScriptInput{Engine: spec.Engine, Path: spec.Path, SessionID: event.SessionID, Script: tools.ScriptWithPayload(spec.Engine, "event", payload, spec.Script)}
				out := scriptTool.Execute(ctx, input)
				if out.Error != "" {
					return connectivity.RouteResponse{Status: 500, Body: out.Result}, errors.New(out.Error)
				}
				resp := connectivity.RouteResponse{Status: 200, Body: out.Result}
				if strings.TrimSpace(out.Result) != "" {
					if err := json.Unmarshal([]byte(out.Result), &resp); err != nil {
						resp = connectivity.RouteResponse{Status: 200, Body: out.Result}
					} else if resp.Status == 0 && resp.Body == "" {
						resp.Body = out.Result
					}
				}
				if resp.Status == 0 {
					resp.Status = 200
				}
				return resp, nil
			})
		},
		func(ctx context.Context, sessionID, id string) error { return e.connectivity.Unregister(ctx, id) },
		func(ctx context.Context, sessionID string, filter map[string]any) ([]connectivity.RouteInfo, error) {
			return e.connectivity.List(ctx, filter)
		},
		func(ctx context.Context, sessionID, topic string, payload map[string]any) error {
			if payload == nil {
				payload = map[string]any{}
			}
			payload["session_id"] = sessionID
			return e.connectivity.Emit(ctx, topic, payload)
		},
		func(ctx context.Context, sessionID string, envelope map[string]any) error {
			if e == nil || e.Topics() == nil {
				return fmt.Errorf("publish topic: topic bus is not available")
			}
			topicName, _ := envelope["topic"].(string)
			topicName = strings.TrimSpace(topicName)
			if topicName == "" {
				return fmt.Errorf("publish topic: topic is required")
			}
			payload, _ := envelope["payload"].(map[string]any)
			if payload == nil {
				payload = map[string]any{}
			}
			agentID, _ := envelope["agent_id"].(string)
			source, _ := envelope["source"].(string)
			if strings.TrimSpace(source) == "" {
				source = "script"
			}
			typ, _ := envelope["type"].(string)
			sessionValue, _ := envelope["session_id"].(string)
			if strings.TrimSpace(sessionValue) == "" {
				sessionValue = sessionID
			}
			e.PublishTopicEvent(topics.Envelope{Topic: topicName, SessionID: sessionValue, AgentID: strings.TrimSpace(agentID), Source: source, Type: strings.TrimSpace(typ), Payload: payload})
			return nil
		},
		func(ctx context.Context, sessionID string, pattern string, opts scripting.TopicSubscribeOptions) (<-chan topics.Envelope, func(), error) {
			if e == nil || e.Topics() == nil {
				return nil, nil, fmt.Errorf("subscribe topic: topic bus is not available")
			}
			subOpts := topics.SubscribeOptions{Buffer: opts.Buffer, SessionID: strings.TrimSpace(opts.SessionID), AgentID: strings.TrimSpace(opts.AgentID)}
			if subOpts.SessionID == "" {
				subOpts.SessionID = sessionID
			}
			ch, unsubscribe := e.Topics().Subscribe(ctx, pattern, subOpts)
			return ch, unsubscribe, nil
		},
	)
	scriptTool.SetAgenticCallbacks(
		func(ctx context.Context, sessionID string, hook scripting.EventHookSpec) (func(), error) {
			if isProcessHookSpec(hook) {
				return e.RegisterHook(hook.Name, tools.FirstNonEmpty(hook.Source, "process"), newProcessHookHandler(e.runtimeCfg.WorkspaceRoot, hook))
			}
			return e.RegisterHook(hook.Name, tools.FirstNonEmpty(hook.Source, "script"), func(ctx context.Context, req HookRequest) (HookResponse, error) {
				payload := hookScriptPayload(req)
				input := tools.ScriptInput{Engine: hook.Engine, Path: hook.Path, SessionID: req.SessionID, Script: tools.ScriptWithPayload(hook.Engine, "hook", payload, hook.Script)}
				out := scriptTool.Execute(ctx, input)
				if out.Error != "" {
					return HookResponse{}, errors.New(out.Error)
				}
				return hookResponseFromScript(out.Result)
			})
		},
		func(ctx context.Context, sessionID string, spec scripting.ToolSpec) error {
			params := spec.Parameters
			if len(params) == 0 {
				params = json.RawMessage(`{"type":"object","properties":{}}`)
			}
			return e.RegisterTool(tools.RegisteredTool{Name: spec.Name, Description: spec.Description, Parameters: params, Source: tools.FirstNonEmpty(spec.Source, "script"), Kind: "mixed", Weight: "standard", Activation: "on-demand", Executor: func(ctx context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
				payload := map[string]any{"tool_call_id": call.ID, "name": call.Name, "arguments": call.Arguments, "session_id": rt.SessionID}
				input := tools.ScriptInput{Engine: spec.Engine, Path: spec.Path, SessionID: rt.SessionID, Script: tools.ScriptWithPayload(spec.Engine, "tool", payload, spec.Script)}
				out := scriptTool.Execute(ctx, input)
				if out.Error != "" {
					return out.Result, errors.New(out.Error)
				}
				return out.Result, nil
			}})
		},
		func(ctx context.Context, sessionID string, names []string) error { return e.SetActiveTools(names) },
		func(ctx context.Context, sessionID string) ([]string, error) { return e.ActiveTools(), nil },
	)
	for _, ext := range tools.LoadWorkspaceExtensions(e.backgroundContext(), e.runtimeCfg.WorkspaceRoot, scriptTool) {
		if ext.Error != "" {
			log.Printf("extension load failed path=%s engine=%s: %s", ext.Path, ext.Engine, ext.Error)
			e.recordExtension(ExtensionInfo{Engine: ext.Engine, Path: ext.Path, Status: "failed", Error: ext.Error})
			continue
		}
		e.recordExtension(ExtensionInfo{Engine: ext.Engine, Path: ext.Path, Status: "loaded"})
		log.Printf("extension loaded path=%s engine=%s", ext.Path, ext.Engine)
	}
	registerDiscoveredTools := func() {
		giskills.RegisterDiscoveredTools(e.runtimeCfg.Discovery.Tools, func(tool giskills.ToolManifest, params json.RawMessage) error {
			return e.RegisterTool(tools.RegisteredTool{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  params,
				Source:      "workspace-manifest",
				Kind:        "mixed",
				Weight:      "standard",
				Activation:  "on-demand",
				Executor: func(ctx context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
					payload := map[string]any{"tool_call_id": call.ID, "name": call.Name, "arguments": call.Arguments, "session_id": rt.SessionID}
					input := tools.ScriptInput{Engine: tool.Engine, Path: tool.Path, SessionID: rt.SessionID, Script: tools.ScriptWithPayload(tool.Engine, "tool", payload, tool.Script)}
					out := scriptTool.Execute(ctx, input)
					if out.Error != "" {
						return out.Result, errors.New(out.Error)
					}
					return out.Result, nil
				},
			})
		}, func(name string, err error) {
			log.Printf("register discovered tool %q: %v", name, err)
		})
	}
	must := func(t tools.RegisteredTool) {
		if err := e.RegisterTool(t); err != nil {
			panic(err)
		}
	}
	must(tools.RegisteredTool{
		Name:        "tools",
		Description: "List available tools or get details about a specific tool. Use with no arguments to list all tools (names + short descriptions). Pass a tool name via the `name` argument to get its full schema and usage. Use `query` to filter tools by keyword.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"name":{"type":"string","description":"Exact tool name to get full details for"},"query":{"type":"string","description":"Filter tools by keyword in name/description/metadata"},"intent":{"type":"string","description":"Natural-language goal for staged discovery"},"include_parameters":{"type":"boolean","description":"Include parameter schemas in list results"},"include_inactive":{"type":"boolean","description":"Include inactive tools in discovery results"},"activate":{"type":"array","items":{"type":"string"},"description":"Set active tools by name; tools remains active"},"reset_active":{"type":"boolean","description":"Reset active tools to all default registry tools"}}}`),
		Source:      "builtin",
		Kind:        "read-only",
		Weight:      "lightweight",
		Activation:  "default",
		Executor: func(ctx context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
			return e.executeToolsTool(call.Arguments)
		},
	})
	must(tools.RegisteredTool{
		Name:        "skills",
		Description: "List workspace-discovered skills or read a skill's SKILL.md. Skills are discovered from .gi/skills and .pi/skills.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"name":{"type":"string","description":"Exact skill name to read; omit to list skills"},"query":{"type":"string","description":"Filter listed skills by name or description"}}}`),
		Source:      "builtin",
		Kind:        "read-only",
		Weight:      "lightweight",
		Activation:  "default",
		Executor: func(ctx context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
			return giskills.ExecuteTool(rt.WorkspaceRoot, call.Arguments)
		},
	})
	registerDiscoveredTools()
	must(tools.RegisteredTool{
		Name:        "read",
		Description: "Read text content from a workspace file. Supports workspace-relative paths and vfs:// paths.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"Workspace-relative path or vfs://namespace/path"}},"required":["path"]}`),
		Source:      "builtin",
		Kind:        "read-only",
		Weight:      "lightweight",
		Activation:  "default",
		Executor: func(ctx context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
			return tools.ExecuteRead(ctx, rt.WorkspaceRoot, rt.Store, call)
		},
	})
	must(tools.RegisteredTool{
		Name:        "write",
		Description: "Write text content to a workspace file. Creates parent directories for workspace paths.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"Workspace-relative path or vfs://namespace/path"},"content":{"type":"string","description":"File content to write"}},"required":["path","content"]}`),
		Source:      "builtin",
		Kind:        "mutating",
		Weight:      "lightweight",
		Activation:  "default",
		Executor: func(ctx context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
			return tools.ExecuteWrite(ctx, rt.WorkspaceRoot, rt.Store, call)
		},
	})
	if def := scriptTool.Definition(); def != nil {
		params, _ := json.Marshal(def["parameters"])
		must(tools.RegisteredTool{
			Name:        "script",
			Description: fmt.Sprint(def["description"]),
			Parameters:  params,
			Source:      "builtin",
			Kind:        "mixed",
			Weight:      "standard",
			Activation:  "default",
			Executor: func(ctx context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
				input := tools.ScriptInput{SessionID: rt.SessionID}
				b, _ := json.Marshal(call.Arguments)
				if err := json.Unmarshal(b, &input); err != nil {
					return "", err
				}
				out := scriptTool.Execute(ctx, input)
				if out.Error != "" {
					return out.Result, errors.New(out.Error)
				}
				return out.Result, nil
			},
		})
	}
	must(tools.RegisteredTool{
		Name:        "compact",
		Description: "Inspect compaction thresholds or estimate whether the current session should compact. Supports smart compaction plugins via session_before_compact hooks.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"session_id":{"type":"string","description":"Session to inspect; defaults to current session"},"dry_run":{"type":"boolean","description":"Return estimate/preparation without changing context"}}}`),
		Source:      "builtin",
		Kind:        "read-only",
		Weight:      "standard",
		Activation:  "on-demand",
		Executor: func(ctx context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
			sessionID, _ := call.Arguments["session_id"].(string)
			if sessionID == "" {
				sessionID = rt.SessionID
			}
			msgs, err := rt.Store.ListMessages(ctx, sessionID)
			if err != nil {
				return "", err
			}
			var aiMsgs []goai.Message
			for _, m := range msgs {
				switch m.Role {
				case "user":
					aiMsgs = append(aiMsgs, goai.UserMessage(m.Content))
				case "assistant":
					aiMsgs = append(aiMsgs, goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: m.Content}}})
				}
			}
			settings := e.runtimeCfg.Compaction
			tokens := compaction.EstimateMessagesTokens(aiMsgs)
			prep := compaction.Prepare(aiMsgs, tokens, settings.KeepRecentTokens, settings.ReserveTokens, settings.ThresholdTokens, settings.Strategy)
			b, _ := json.MarshalIndent(map[string]any{"enabled": settings.Enabled, "context_tokens": tokens, "threshold_tokens": settings.ThresholdTokens, "should_compact": settings.Enabled && tokens > settings.ThresholdTokens, "preparation": prep}, "", "  ")
			return string(b), nil
		},
	})
	must(tools.RegisteredTool{
		Name:        "peering",
		Description: "Inspect gi peer-discovery status. The backend is tsnet/Tailscale and is disabled by default until configured.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"action":{"type":"string","description":"status (default); future: start/stop/discover"}}}`),
		Source:      "builtin",
		Kind:        "read-only",
		Weight:      "lightweight",
		Activation:  "on-demand",
		Executor: func(ctx context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
			b, _ := json.MarshalIndent(e.PeeringStatus(), "", "  ")
			return string(b), nil
		},
	})
	must(tools.RegisteredTool{
		Name:        "rtk",
		Description: "Run a shell command and return RTK-style compact output using gi's native Go filters for git/search/listing/test output.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"command":{"type":"string","description":"Shell command to execute and compact"},"filter_only":{"type":"boolean","description":"Filter the supplied output instead of executing"},"output":{"type":"string","description":"Raw output to filter when filter_only is true"}},"required":["command"]}`),
		Source:      "builtin",
		Kind:        "mixed",
		Weight:      "standard",
		Activation:  "on-demand",
		Executor: func(ctx context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
			return tools.ExecuteRTK(ctx, rt.WorkspaceRoot, call)
		},
	})
	must(tools.RegisteredTool{
		Name:        "shell",
		Description: "Execute a shell command and return stdout/stderr. Use for running tests, installing packages, searching files, etc.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"command":{"type":"string","description":"Shell command to execute"}},"required":["command"]}`),
		Source:      "builtin",
		Kind:        "mixed",
		Weight:      "heavy",
		Activation:  "default",
		Executor: func(ctx context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
			return tools.ExecuteShell(ctx, rt.WorkspaceRoot, call)
		},
	})
}

const repeatedToolFailureLimit = 4

func toolFailureSignature(call goai.ToolCall, err error) string {
	if err == nil {
		return ""
	}
	argKeys := make([]string, 0, len(call.Arguments))
	for k := range call.Arguments {
		argKeys = append(argKeys, k)
	}
	sort.Strings(argKeys)
	parts := make([]string, 0, len(argKeys))
	for _, k := range argKeys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, call.Arguments[k]))
	}
	return fmt.Sprintf("%s|%s|%s", call.Name, strings.Join(parts, ","), err.Error())
}

func nextRepeatedToolFailureCount(lastSig string, lastCount int, call goai.ToolCall, err error) (string, int) {
	sig := toolFailureSignature(call, err)
	if sig == "" {
		return "", 0
	}
	if sig == lastSig {
		return sig, lastCount + 1
	}
	return sig, 1
}

// runAgentLoop runs the core tool-use loop: call LLM, execute any tool calls,
// feed results back, repeat until the LLM produces a final text response or
// the iteration budget is exhausted.
func (r *sessionRunner) runAgentLoop(ctx context.Context, s *store.Store, turnID, sessionID, model, agentID string, initialSteering []store.SteeringMessage) {
	maxIter := r.engine.runtimeCfg.MaxIterations
	if maxIter <= 0 {
		maxIter = 64
	}

	convCtx := r.assembleAgentContext(ctx, s, turnID, sessionID, model, agentID)
	agentEndReason := "completed"
	defer func() {
		_, _ = r.engine.emitHook(r.engine.backgroundContext(), HookRequest{Name: HookAgentEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Payload: map[string]any{"reason": agentEndReason}})
	}()

	r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "title": "Thinking…", "status": "running", "turn_id": turnID})

	var totalUsage goai.Usage
	lastToolFailureSig := ""
	repeatedToolFailureCount := 0
	pendingSteering := append([]store.SteeringMessage(nil), initialSteering...)

	for iter := 1; iter <= maxIter; iter++ {
		if ctx.Err() != nil {
			agentEndReason = "cancelled"
			r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled", "")
			return
		}

		pendingSteering = r.prepareAgentIteration(ctx, sessionID, turnID, model, agentID, iter, convCtx, pendingSteering)
		result, inferErr := r.runProviderIteration(ctx, s, turnID, sessionID, model, agentID, iter, maxIter, convCtx)
		iterLabel := fmt.Sprintf("iter=%d/%d", iter, maxIter)
		if inferErr != nil {
			if ctx.Err() != nil || isCancellationError(inferErr) {
				agentEndReason = "cancelled"
				r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled", "")
				return
			}
			var abortErr hookAbortError
			if errors.As(inferErr, &abortErr) {
				logutil.WarnIfErr("append inference.aborted event", s.AppendTurnEvent(ctx, turnID, sessionID, "inference.aborted", map[string]any{"phase": "inference", "checkpoint": true, "error": abortErr.Error(), "iteration": iter, "hard_abort": abortErr.hard}))
				agentEndReason = "aborted"
				r.finishTurn(s, turnID, sessionID, agentID, model, "aborted", abortErr.Error(), "hook_abort")
				return
			}
			log.Printf("inference [%s] error: %v", iterLabel, inferErr)
			logutil.WarnIfErr("append inference.failed event", s.AppendTurnEvent(ctx, turnID, sessionID, "inference.failed", map[string]any{"phase": "inference", "checkpoint": true, "error": inferErr.Error(), "iteration": iter}))
			agentEndReason = "failed"
			r.finishTurn(s, turnID, sessionID, agentID, model, "failed", fmt.Sprintf("Inference error: %v", inferErr), "provider_error")
			return
		}
		if result == nil || result.Message == nil {
			if ctx.Err() != nil {
				agentEndReason = "cancelled"
				r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled", "")
				return
			}
			log.Printf("inference [%s]: nil result", iterLabel)
			agentEndReason = "failed"
			r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Inference returned no result", "provider_invalid_result")
			return
		}

		if result.Usage != nil {
			totalUsage.Input += result.Usage.Input
			totalUsage.Output += result.Usage.Output
			totalUsage.TotalTokens += result.Usage.TotalTokens
			totalUsage.CacheRead += result.Usage.CacheRead
			totalUsage.CacheWrite += result.Usage.CacheWrite
			totalUsage.Cost.Input += result.Usage.Cost.Input
			totalUsage.Cost.Output += result.Usage.Cost.Output
			totalUsage.Cost.Total += result.Usage.Cost.Total
		}

		assistantMsg := result.Message
		textContent := goai.GetTextContent(assistantMsg)
		toolCalls := goai.GetToolCalls(assistantMsg)
		log.Printf("inference [%s]: stop=%q toolCalls=%d text=%d", iterLabel, assistantMsg.StopReason, len(toolCalls), len(textContent))
		goai.AppendAssistantMessage(convCtx, assistantMsg)

		needsToolExecution := goai.NeedsToolExecution(assistantMsg) || len(toolCalls) > 0
		if !needsToolExecution {
			if steerMsgs, err := r.dequeueSteeringMessages(ctx, sessionID); err != nil {
				log.Printf("steering dequeue error after direct response: %v", err)
			} else if len(steerMsgs) > 0 {
				pendingSteering = append(pendingSteering, steerMsgs...)
				continue
			}
			log.Printf("inference [%s]: final response (%d chars, %d iterations)", iterLabel, len(textContent), iter)
			agentEndReason = "completed"
			r.persistUsage(s, turnID, sessionID, &totalUsage, iter)

			msgID := store.NowID("msg")
			logutil.WarnIfErr("add assistant inference message", s.AddMessage(ctx, msgID, sessionID, "assistant", textContent, map[string]any{
				"kind": "chat", "source": "inference", "model": model,
				"turn_id": turnID, "agent_id": agentID, "iterations": iter,
			}))

			r.broadcastPost(sessionID, turnID, msgID, textContent, agentID)
			_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookMessageEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"chars": len(textContent)}})
			_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookTurnEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"status": "completed"}})
			r.finishTurnOK(s, turnID, sessionID, agentID, model, iter)
			return
		}

		log.Printf("inference [%s]: %d tool call(s)", iterLabel, len(toolCalls))
		toolCallSummary := textContent
		for _, tc := range toolCalls {
			if toolCallSummary != "" {
				toolCallSummary += "\n"
			}
			toolCallSummary += fmt.Sprintf("[tool_call: %s]", tc.Name)
		}
		logutil.WarnIfErr("add assistant tool_calls summary", s.AddMessage(ctx, store.NowID("msg"), sessionID, "assistant", toolCallSummary, map[string]any{
			"kind": "tool_calls", "source": "inference", "model": model,
			"turn_id": turnID, "agent_id": agentID,
		}))

		outcome := r.executeToolCallsPhase(ctx, s, turnID, sessionID, model, agentID, iter, convCtx, toolCalls, pendingSteering, lastToolFailureSig, repeatedToolFailureCount, &totalUsage)
		if outcome.terminated {
			return
		}
		pendingSteering = outcome.pendingSteering
		lastToolFailureSig = outcome.lastToolFailureSig
		repeatedToolFailureCount = outcome.repeatedToolFailureCount
		if outcome.skipRemainingTools || len(pendingSteering) > 0 {
			continue
		}
		_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookTurnEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"status": "tools"}})
	}

	log.Printf("inference: max iterations (%d) reached for turn %s", maxIter, turnID)
	agentEndReason = "completed"
	r.persistUsage(s, turnID, sessionID, &totalUsage, maxIter)
	r.finishTurnWithPayload(s, turnID, sessionID, agentID, model, "completed", fmt.Sprintf("Reached maximum iteration limit (%d). The task may be incomplete.", maxIter), "", map[string]any{"iterations": maxIter, "completion_kind": "max_iterations"})
}

// executeTool dispatches a single tool call and returns the text result.
func (r *sessionRunner) executeTool(ctx context.Context, call goai.ToolCall, sessionID, turnID string) (string, error) {
	if strings.TrimSpace(turnID) != "" {
		turnRec, err := r.store.GetTurn(ctx, turnID)
		if err != nil {
			return "", err
		}
		if !toolAllowedByMetadata(turnRec.Metadata, call.Name) {
			return "", fmt.Errorf("tool not allowed in this turn: %s", call.Name)
		}
	}
	tool, ok := r.engine.tools.GetRegistered(call.Name)
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", call.Name)
	}
	return tool.Executor(ctx, tools.ToolRuntime{
		Store:         r.store,
		SessionID:     sessionID,
		TurnID:        turnID,
		WorkspaceRoot: r.engine.runtimeCfg.WorkspaceRoot,
	}, call)
}

// persistUsage records cumulative usage for the turn.
func (r *sessionRunner) persistUsage(s *store.Store, turnID, sessionID string, usage *goai.Usage, iterations int) {
	bgCtx := r.engine.backgroundContext()
	usageMap := map[string]any{
		"input": usage.Input, "output": usage.Output,
		"total":      usage.TotalTokens,
		"cache_read": usage.CacheRead, "cache_write": usage.CacheWrite,
		"cost_input": usage.Cost.Input, "cost_output": usage.Cost.Output,
		"cost_total": usage.Cost.Total,
		"iterations": iterations,
	}
	logutil.WarnIfErr("append inference.finished event", s.AppendTurnEvent(bgCtx, turnID, sessionID, "inference.finished", map[string]any{
		"phase": "inference", "checkpoint": true, "usage": usageMap, "iterations": iterations,
	}))
	log.Printf("inference: usage input=%d output=%d total=%d cost=%.6f iterations=%d",
		usage.Input, usage.Output, usage.TotalTokens, usage.Cost.Total, iterations)
}

// broadcastPost sends a new_post SSE event for the final assistant message.
func (r *sessionRunner) broadcastPost(sessionID, turnID, msgID, content, agentID string) {
	r.engine.broadcast(sessionID, map[string]any{
		"type": "new_post", "id": msgID, "chat_jid": "gi:" + sessionID,
		"content": content, "timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"sender": "agent", "is_bot_message": true,
		"data": map[string]any{"type": "agent_response", "content": content, "agent_id": agentID},
	})
	r.engine.broadcast(sessionID, map[string]any{"type": "agent_response", "chat_jid": "gi:" + sessionID, "id": msgID})
}

func (r *sessionRunner) broadcastSystemPost(sessionID, turnID, msgID, content string) {
	r.engine.broadcast(sessionID, map[string]any{
		"type": "new_post", "id": msgID, "chat_jid": "gi:" + sessionID,
		"content": content, "timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"sender": "system",
		"data":   map[string]any{"type": "system_message", "content": content, "turn_id": turnID},
	})
}

func terminalPhaseForStatus(status string) string {
	switch status {
	case "completed":
		return "completed"
	case "failed":
		return "failed"
	case "aborted", "cancelled":
		return "aborted"
	case "cancelling":
		return "cancelling"
	case "queued":
		return "queued"
	default:
		return "running"
	}
}

// finishTurnOK marks a turn as successfully completed.
func (r *sessionRunner) finishTurnOK(s *store.Store, turnID, sessionID, agentID, model string, iterations int) {
	r.finishTurnWithPayload(s, turnID, sessionID, agentID, model, "completed", "", "", map[string]any{"iterations": iterations, "completion_kind": "response"})
}

// finishTurn persists a terminal status and optional system message.
func (r *sessionRunner) finishTurn(s *store.Store, turnID, sessionID, agentID, model, status, systemMsg, failureKind string) {
	r.finishTurnWithPayload(s, turnID, sessionID, agentID, model, status, systemMsg, failureKind, nil)
}

func (r *sessionRunner) finishTurnWithPayload(s *store.Store, turnID, sessionID, agentID, model, status, systemMsg, failureKind string, payload map[string]any) {
	bgCtx := r.engine.backgroundContext()
	if strings.TrimSpace(agentID) == "" || strings.TrimSpace(model) == "" {
		resolvedAgentID, resolvedModel := r.resolveTurnIdentityForFinalize(bgCtx, s, sessionID, turnID)
		if strings.TrimSpace(agentID) == "" {
			agentID = resolvedAgentID
		}
		if strings.TrimSpace(model) == "" {
			model = resolvedModel
		}
	}
	r.appendFinalSteeringCheckpoint(s, turnID, sessionID)
	if systemMsg != "" {
		msgID := store.NowID("msg")
		logutil.WarnIfErr("add terminal system message", s.AddMessage(bgCtx, msgID, sessionID, "system", systemMsg, map[string]any{
			"kind": "chat", "source": "system", "turn_id": turnID, "agent_id": agentID,
		}))
		r.broadcastSystemPost(sessionID, turnID, msgID, systemMsg)
	}
	if failureKind != "" {
		logutil.WarnIfErr("turn failure mark", s.MarkTurnFailureWithFallbackErr(r.engine.backgroundContext(), nil, turnID, sessionID, failureKind, "none", systemMsg))
	}
	finishedPayload := cloneMap(payload)
	if finishedPayload == nil {
		finishedPayload = map[string]any{}
	}
	finishedPayload["phase"] = "turn"
	finishedPayload["checkpoint"] = true
	finishedPayload["status"] = status
	finishedPayload["reason"] = tools.FirstNonEmpty(failureKind, status)
	finishedPayload["failure_kind"] = failureKind
	logutil.WarnIfErr("append turn.finished event", s.AppendTurnEvent(bgCtx, turnID, sessionID, "turn.finished", finishedPayload))
	phase := terminalPhaseForStatus(status)
	logutil.WarnIfErr("update turn status and phase terminal", s.UpdateTurnStatusAndPhase(bgCtx, turnID, status, phase))
	logutil.WarnIfErr("mark turn finished", s.MarkTurnFinished(bgCtx, turnID))
	turnEventType := "turn_terminal"
	sessionIdleReason := "turn_terminal"
	if status == "completed" && failureKind == "" {
		turnEventType = "turn_completed"
		sessionIdleReason = "turn_completed"
	}
	turnPayload := cloneMap(payload)
	if turnPayload == nil {
		turnPayload = map[string]any{}
	}
	turnPayload["reason"] = tools.FirstNonEmpty(failureKind, status)
	turnPayload["failure_kind"] = failureKind
	r.engine.PublishRuntimeTurnEvent(turnEventType, sessionID, turnID, agentID, status, phase, turnPayload)
	hookPayload := cloneMap(payload)
	if hookPayload == nil {
		hookPayload = map[string]any{}
	}
	hookPayload["reason"] = tools.FirstNonEmpty(failureKind, status)
	hookPayload["failure_kind"] = failureKind
	r.emitTurnStateHookOnly(bgCtx, sessionID, turnID, agentID, model, status, phase, hookPayload)
	r.propagateChildSubTurnCancellation(bgCtx, turnID, status, failureKind)
	r.publishSubTurnLifecycle(bgCtx, turnID, status)
	logutil.WarnIfErr("touch session idle", s.TouchSessionState(bgCtx, sessionID, map[string]any{"status": "idle", "active_turn_id": nil}))
	sessionPayload := cloneMap(payload)
	if sessionPayload == nil {
		sessionPayload = map[string]any{}
	}
	sessionPayload["reason"] = sessionIdleReason
	sessionPayload["failure_kind"] = failureKind
	sessionPayload["active_turn_id"] = nil
	sessionPayload["turn_id"] = turnID
	sessionPayload["turn_status"] = status
	sessionPayload["turn_phase"] = phase
	sessionPayload["model"] = model
	r.engine.PublishRuntimeSessionEvent("session_idle", sessionID, agentID, "idle", sessionPayload)
	sessionHookPayload := cloneMap(payload)
	if sessionHookPayload == nil {
		sessionHookPayload = map[string]any{}
	}
	sessionHookPayload["reason"] = sessionIdleReason
	sessionHookPayload["failure_kind"] = failureKind
	sessionHookPayload["active_turn_id"] = nil
	sessionHookPayload["turn_id"] = turnID
	sessionHookPayload["turn_status"] = status
	sessionHookPayload["turn_phase"] = phase
	r.emitSessionStateHookOnly(bgCtx, sessionID, agentID, model, "idle", sessionHookPayload)
	r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "title": "", "status": "idle"})
}

func (r *sessionRunner) publishSubTurnLifecycle(ctx context.Context, childTurnID, status string) {
	if strings.TrimSpace(childTurnID) == "" {
		return
	}
	opCtx := store.CoordinationContext(ctx, r.engine.backgroundContext())
	sub, err := r.store.GetSubTurnByChild(opCtx, childTurnID)
	if err != nil {
		return
	}
	payload := map[string]any{
		"type":           "subturn_status",
		"chat_jid":       "gi:" + sub.ParentSessionID,
		"parent_turn_id": sub.ParentTurnID,
		"parent_session": sub.ParentSessionID,
		"child_turn_id":  sub.ChildTurnID,
		"child_session":  sub.ChildSessionID,
		"status":         status,
		"depth":          sub.Depth,
		"delivery_mode":  sub.DeliveryMode,
	}
	r.engine.broadcast(sub.ParentSessionID, payload)
	if sub.ChildSessionID != sub.ParentSessionID {
		mirror := cloneMap(payload)
		mirror["chat_jid"] = "gi:" + sub.ChildSessionID
		r.engine.broadcast(sub.ChildSessionID, mirror)
	}
	if !isTerminalSubTurnStatus(status) {
		return
	}
	summary := r.subTurnResultSummary(opCtx, sub.ChildSessionID, sub.ChildTurnID)
	if sub.DeliveryMode == "async" {
		orphaned := false
		if parentTurn, err := r.store.GetTurn(opCtx, sub.ParentTurnID); err == nil {
			orphaned = isTerminalSubTurnStatus(parentTurn.Status)
		}
		eventType := "subturn_result_ready"
		if orphaned {
			eventType = "subturn_orphaned"
			logutil.WarnIfErr("update async subturn orphan metadata", r.store.UpdateSubTurnMetadataByChild(opCtx, sub.ChildTurnID, map[string]any{
				"orphaned":      true,
				"orphaned_at":   time.Now().UTC().Format(time.RFC3339Nano),
				"orphan_reason": "parent_turn_completed_before_async_result_consumption",
			}))
			if sub.ParentSessionID != sub.ChildSessionID {
				content := fmt.Sprintf("Async sub-turn %s finished with status %s after parent turn %s had already ended.", sub.ChildTurnID, status, sub.ParentTurnID)
				if strings.TrimSpace(summary) != "" {
					content += "\n\n" + summary
				}
				logutil.WarnIfErr("add async orphan result message", r.store.AddMessage(opCtx, store.NowID("msg"), sub.ParentSessionID, "system", content, map[string]any{
					"kind":             "subturn_orphan_result",
					"parent_turn_id":   sub.ParentTurnID,
					"child_turn_id":    sub.ChildTurnID,
					"child_session_id": sub.ChildSessionID,
					"status":           status,
					"delivery_mode":    "async",
					"orphaned":         true,
					"summary":          summary,
				}))
			}
		}
		r.engine.broadcast(sub.ParentSessionID, map[string]any{
			"type":           eventType,
			"chat_jid":       "gi:" + sub.ParentSessionID,
			"parent_turn_id": sub.ParentTurnID,
			"parent_session": sub.ParentSessionID,
			"child_turn_id":  sub.ChildTurnID,
			"child_session":  sub.ChildSessionID,
			"status":         status,
			"summary":        summary,
			"delivery_mode":  "async",
			"orphaned":       orphaned,
		})
		return
	}
	if sub.ParentSessionID != sub.ChildSessionID {
		content := fmt.Sprintf("Sub-turn %s finished with status %s.", sub.ChildTurnID, status)
		if strings.TrimSpace(summary) != "" {
			content += "\n\n" + summary
		}
		logutil.WarnIfErr("add sync subturn result message", r.store.AddMessage(opCtx, store.NowID("msg"), sub.ParentSessionID, "system", content, map[string]any{
			"kind":             "subturn_result",
			"parent_turn_id":   sub.ParentTurnID,
			"child_turn_id":    sub.ChildTurnID,
			"child_session_id": sub.ChildSessionID,
			"status":           status,
			"delivery_mode":    "sync",
			"summary":          summary,
		}))
	}
	r.engine.broadcast(sub.ParentSessionID, map[string]any{
		"type":           "subturn_result_delivered",
		"chat_jid":       "gi:" + sub.ParentSessionID,
		"parent_turn_id": sub.ParentTurnID,
		"parent_session": sub.ParentSessionID,
		"child_turn_id":  sub.ChildTurnID,
		"child_session":  sub.ChildSessionID,
		"status":         status,
		"summary":        summary,
		"delivery_mode":  "sync",
	})
}

func isTerminalSubTurnStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "completed", "failed", "aborted", "cancelled":
		return true
	default:
		return false
	}
}

func (r *sessionRunner) subTurnResultSummary(ctx context.Context, childSessionID, childTurnID string) string {
	if strings.TrimSpace(childSessionID) == "" {
		return ""
	}
	msgs, err := r.store.ListMessages(ctx, childSessionID)
	if err != nil {
		return ""
	}
	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		if msg.Role != "assistant" {
			continue
		}
		msgTurnID := internalx.StringValue(msg.Payload["turn_id"], "")
		if msgTurnID != "" && msgTurnID != childTurnID {
			continue
		}
		text := strings.TrimSpace(msg.Content)
		if text == "" {
			continue
		}
		if len(text) > 500 {
			return text[:500] + "..."
		}
		return text
	}
	return ""
}

type subTurnCancellationDirective struct {
	enabled        bool
	cancelCritical bool
	reason         string
}

func directiveForParentTerminalStatus(parentStatus, failureKind string) subTurnCancellationDirective {
	status := strings.ToLower(strings.TrimSpace(parentStatus))
	failureKind = strings.ToLower(strings.TrimSpace(failureKind))
	switch {
	case isTimeoutFailureKind(failureKind):
		return subTurnCancellationDirective{enabled: true, cancelCritical: true, reason: "parent_timeout"}
	case status == "cancelled" || status == "aborted":
		return subTurnCancellationDirective{enabled: true, cancelCritical: true, reason: "parent_hard_abort"}
	case status == "completed":
		return subTurnCancellationDirective{enabled: true, cancelCritical: false, reason: "parent_finished_gracefully"}
	default:
		return subTurnCancellationDirective{}
	}
}

func isTimeoutFailureKind(failureKind string) bool {
	failureKind = strings.ToLower(strings.TrimSpace(failureKind))
	if failureKind == "" {
		return false
	}
	return strings.Contains(failureKind, "timeout") || strings.Contains(failureKind, "deadline")
}

func isCriticalSubTurn(sub store.SubTurn) bool {
	return internalx.BoolValue(sub.Metadata["subturn_critical"]) || internalx.BoolValue(sub.Metadata["critical"])
}

func isCancellableSubTurnStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "queued", "running", "cancelling":
		return true
	default:
		return false
	}
}

func (r *sessionRunner) propagateChildSubTurnCancellation(ctx context.Context, parentTurnID, parentStatus, failureKind string) {
	parentTurnID = strings.TrimSpace(parentTurnID)
	if parentTurnID == "" {
		return
	}
	directive := directiveForParentTerminalStatus(parentStatus, failureKind)
	if !directive.enabled {
		return
	}
	visited := map[string]bool{}
	var walk func(string)
	walk = func(turnID string) {
		subs, err := r.store.ListSubTurnsByParent(ctx, turnID)
		if err != nil {
			logutil.WarnIfErr("list child subturns", err)
			return
		}
		for _, sub := range subs {
			if visited[sub.ChildTurnID] {
				continue
			}
			visited[sub.ChildTurnID] = true
			walk(sub.ChildTurnID)
			if !isCancellableSubTurnStatus(sub.Status) {
				continue
			}
			if isCriticalSubTurn(sub) && !directive.cancelCritical {
				continue
			}
			logutil.WarnIfErr("update child subturn cancellation metadata", r.store.UpdateSubTurnMetadataByChild(ctx, sub.ChildTurnID, map[string]any{
				"cancel_requested_by_parent":     true,
				"cancel_requested_at":            time.Now().UTC().Format(time.RFC3339Nano),
				"cancel_requested_parent_turn":   turnID,
				"cancel_requested_parent_status": parentStatus,
				"cancel_requested_failure_kind":  failureKind,
				"cancel_reason":                  directive.reason,
			}))
			if err := r.engine.CancelTurn(ctx, sub.ChildSessionID, sub.ChildTurnID); err != nil {
				logutil.WarnIfErr("cancel child subturn", err)
				continue
			}
			r.engine.broadcast(sub.ParentSessionID, map[string]any{
				"type":                "subturn_cancel_requested",
				"chat_jid":            "gi:" + sub.ParentSessionID,
				"parent_turn_id":      sub.ParentTurnID,
				"parent_session":      sub.ParentSessionID,
				"child_turn_id":       sub.ChildTurnID,
				"child_session":       sub.ChildSessionID,
				"reason":              directive.reason,
				"parent_status":       parentStatus,
				"parent_failure_kind": failureKind,
			})
		}
	}
	walk(parentTurnID)
}

type toolExecutionOutcome struct {
	terminated               bool
	skipRemainingTools       bool
	pendingSteering          []store.SteeringMessage
	lastToolFailureSig       string
	repeatedToolFailureCount int
}

var streamWithToolsWithHooks = inference.StreamWithToolsWithHooks

func isCancellationError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

type hookAbortError struct {
	reason string
	hard   bool
}

func (e hookAbortError) Error() string {
	if strings.TrimSpace(e.reason) == "" {
		if e.hard {
			return "turn hard-aborted by hook"
		}
		return "turn aborted by hook"
	}
	return e.reason
}

func hookAbortFromResponse(resp HookResponse, fallback string) error {
	if !resp.Cancel {
		return nil
	}
	reason := strings.TrimSpace(resp.Reason)
	if reason == "" {
		reason = fallback
	}
	return hookAbortError{reason: reason, hard: internalx.BoolValue(resp.Payload["hard_abort"])}
}

func directToolResultFromHook(resp HookResponse) (string, bool) {
	if resp.ToolResult != nil {
		return *resp.ToolResult, true
	}
	if resp.Handled && strings.TrimSpace(resp.Message) != "" {
		return resp.Message, true
	}
	return "", false
}

func providerRequestReplacementFromHook(resp HookResponse) (any, bool) {
	if resp.Payload == nil {
		return nil, false
	}
	replacement, ok := resp.Payload["request"]
	if !ok || replacement == nil {
		return nil, false
	}
	return replacement, true
}

func (r *sessionRunner) assembleAgentContext(ctx context.Context, s *store.Store, turnID, sessionID, model, agentID string) *goai.Context {
	sysPrompt := r.engine.systemPrompt
	if sysPrompt == "" {
		sysPrompt = "You are a helpful coding assistant."
	}

	var turnMetadata map[string]any
	if turnRec, err := s.GetTurn(ctx, turnID); err == nil {
		turnMetadata = turnRec.Metadata
	}

	msgs, _ := s.ListMessages(ctx, sessionID)
	convCtx := &goai.Context{
		SystemPrompt: sysPrompt,
		Tools:        r.engine.toolDefsForMetadata(turnMetadata),
	}
	for _, m := range msgs {
		switch m.Role {
		case "user":
			convCtx.Messages = append(convCtx.Messages, goai.UserMessage(m.Content))
		case "assistant":
			convCtx.Messages = append(convCtx.Messages, goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: m.Content}}})
		}
	}

	if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookBeforeAgentStart, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, SystemPrompt: convCtx.SystemPrompt, Messages: convCtx.Messages, Tools: convCtx.Tools}); err != nil {
		log.Printf("hook before_agent_start error: %v", err)
	} else {
		if resp.SystemPrompt != "" {
			convCtx.SystemPrompt = resp.SystemPrompt
		}
		if resp.Messages != nil {
			convCtx.Messages = resp.Messages
		}
		if resp.Tools != nil {
			convCtx.Tools = resp.Tools
		}
		if strings.TrimSpace(resp.Message) != "" {
			convCtx.Messages = append([]goai.Message{goai.UserMessage(resp.Message)}, convCtx.Messages...)
		}
	}
	_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookAgentStart, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model})
	return convCtx
}

func (r *sessionRunner) prepareAgentIteration(ctx context.Context, sessionID, turnID, model, agentID string, iter int, convCtx *goai.Context, pendingSteering []store.SteeringMessage) []store.SteeringMessage {
	if steerMsgs, err := r.dequeueSteeringMessages(ctx, sessionID); err != nil {
		log.Printf("steering dequeue error: %v", err)
	} else if len(steerMsgs) > 0 {
		pendingSteering = append(pendingSteering, steerMsgs...)
	}
	if len(pendingSteering) > 0 {
		r.injectSteeringMessages(ctx, sessionID, turnID, convCtx, pendingSteering)
		pendingSteering = nil
	}
	_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookTurnStart, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter})
	if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookContext, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, SystemPrompt: convCtx.SystemPrompt, Messages: convCtx.Messages, Tools: convCtx.Tools}); err != nil {
		log.Printf("hook context error: %v", err)
	} else {
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
	compaction.MaybeCompactContext(ctx, compaction.RuntimeRequest{SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Settings: r.engine.runtimeCfg.Compaction}, convCtx, compaction.RuntimeOps{BackgroundContext: r.engine.backgroundContext, BeforeCompact: func(ctx context.Context, payload map[string]any, messages []goai.Message) (compaction.HookDecision, error) {
		resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookSessionBeforeCompact, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Payload: payload, Messages: messages})
		return compaction.HookDecision{Cancel: resp.Cancel, Block: resp.Block, Payload: resp.Payload}, err
	}, AfterCompact: func(ctx context.Context, payload map[string]any) {
		_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookSessionCompact, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Payload: payload})
	}, UpdateTurnStatusAndPhase: r.store.UpdateTurnStatusAndPhase, AppendTurnEvent: r.store.AppendTurnEvent, TouchSessionActiveTurn: r.store.TouchSessionActiveTurn, AddMessage: r.store.AddMessage, Broadcast: r.engine.broadcast, Warn: logutil.WarnIfErr})
	return pendingSteering
}

func (r *sessionRunner) runProviderIteration(ctx context.Context, s *store.Store, turnID, sessionID, model, agentID string, iter, maxIter int, convCtx *goai.Context) (*inference.StreamResult, error) {
	requestCtx := &goai.Context{SystemPrompt: convCtx.SystemPrompt, Messages: append([]goai.Message(nil), convCtx.Messages...), Tools: append([]goai.Tool(nil), convCtx.Tools...)}
	if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookBeforeProviderRequest, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, SystemPrompt: convCtx.SystemPrompt, Messages: convCtx.Messages, Tools: convCtx.Tools, Payload: map[string]any{"model": model, "messages": len(convCtx.Messages), "tools": len(convCtx.Tools), "stage": "context"}}); err != nil {
		log.Printf("hook before_provider_request error: %v", err)
	} else {
		if abortErr := hookAbortFromResponse(resp, "aborted before provider request by hook"); abortErr != nil {
			return nil, abortErr
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
	}
	logutil.WarnIfErr("update turn running phase", s.UpdateTurnStatusAndPhase(ctx, turnID, "running", "running"))
	r.emitTurnStateHook(ctx, sessionID, turnID, agentID, model, "running", "running", map[string]any{"reason": "provider_iteration", "iteration": iter})
	logutil.WarnIfErr("append inference.started event", s.AppendTurnEvent(ctx, turnID, sessionID, "inference.started", map[string]any{"phase": "inference", "model": model, "iteration": iter, "checkpoint": true}))
	iterLabel := fmt.Sprintf("iter=%d/%d", iter, maxIter)
	log.Printf("inference [%s]: calling %s", iterLabel, model)

	r.engine.broadcast(sessionID, map[string]any{
		"type": "agent_status", "chat_jid": "gi:" + sessionID,
		"title": fmt.Sprintf("Thinking… (%d)", iter), "status": "running", "turn_id": turnID,
	})

	responseObserved := false
	result, inferErr := streamWithToolsWithHooks(ctx, model, requestCtx, func(ev map[string]any) {
		ev["chat_jid"] = "gi:" + sessionID
		ev["turn_id"] = turnID
		ev["iteration"] = iter
		switch ev["type"] {
		case "text_delta":
			_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookMessageUpdate, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"delta": ev["delta"]}})
			ev["type"] = "agent_draft_delta"
			r.engine.broadcast(sessionID, ev)
		case "thinking_delta":
			ev["type"] = "agent_thought_delta"
			r.engine.broadcast(sessionID, ev)
		case "tool_call_start":
			r.engine.broadcast(sessionID, map[string]any{
				"type": "agent_status", "chat_jid": "gi:" + sessionID,
				"title": fmt.Sprintf("Tool: %s", ev["name"]), "status": "running", "turn_id": turnID,
			})
		case "error":
			r.engine.broadcast(sessionID, ev)
		}
	}, &inference.StreamHooks{
		OnPayload: func(payload any, modelDef *goai.Model) (any, error) {
			hookPayload := map[string]any{"ok": true, "request": payload, "stage": "payload"}
			if modelDef != nil {
				hookPayload["provider"] = string(modelDef.Provider)
				hookPayload["api"] = string(modelDef.Api)
				hookPayload["model_id"] = modelDef.ID
			}
			resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookBeforeProviderRequest, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, SystemPrompt: requestCtx.SystemPrompt, Messages: requestCtx.Messages, Tools: requestCtx.Tools, Payload: hookPayload})
			if err != nil {
				return nil, err
			}
			if abortErr := hookAbortFromResponse(resp, "aborted before provider request send by hook"); abortErr != nil {
				return nil, abortErr
			}
			if replacement, ok := providerRequestReplacementFromHook(resp); ok {
				return replacement, nil
			}
			return payload, nil
		},
		OnResponse: func(status int, headers map[string]string, modelDef *goai.Model) {
			responseObserved = true
			payload := map[string]any{"ok": true, "status": status, "headers": headers}
			if modelDef != nil {
				payload["provider"] = string(modelDef.Provider)
				payload["api"] = string(modelDef.Api)
				payload["model_id"] = modelDef.ID
			}
			if _, err := r.engine.emitHook(ctx, HookRequest{Name: HookAfterProviderResponse, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: payload}); err != nil {
				log.Printf("hook after_provider_response error: %v", err)
			}
		},
	})
	if !responseObserved {
		if _, err := r.engine.emitHook(ctx, HookRequest{Name: HookAfterProviderResponse, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"ok": inferErr == nil}}); err != nil {
			log.Printf("hook after_provider_response error: %v", err)
		}
	}
	return result, inferErr
}

func (r *sessionRunner) executeToolCallsPhase(ctx context.Context, s *store.Store, turnID, sessionID, model, agentID string, iter int, convCtx *goai.Context, toolCalls []goai.ToolCall, pendingSteering []store.SteeringMessage, lastToolFailureSig string, repeatedToolFailureCount int, totalUsage *goai.Usage) toolExecutionOutcome {
	outcome := toolExecutionOutcome{
		pendingSteering:          pendingSteering,
		lastToolFailureSig:       lastToolFailureSig,
		repeatedToolFailureCount: repeatedToolFailureCount,
	}
	_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookToolExecutionStart, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"count": len(toolCalls)}})
	defer func() {
		_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookToolExecutionEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"count": len(toolCalls)}})
	}()
	for i, call := range toolCalls {
		if ctx.Err() != nil {
			r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled during tool execution", "")
			outcome.terminated = true
			return outcome
		}

		if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookToolCall, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, ToolCall: &call, Payload: map[string]any{"tool": call.Name, "tool_call_id": call.ID, "arguments": call.Arguments}}); err != nil {
			log.Printf("hook tool_call error: %v", err)
		} else {
			if abortErr := hookAbortFromResponse(resp, fmt.Sprintf("tool %s aborted by hook", call.Name)); abortErr != nil {
				r.engine.PublishRuntimeHookDecisionEvent("hook_abort", HookRequest{Name: HookToolCall, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "tool_call", "reason": abortErr.Error()})
				r.finishTurn(s, turnID, sessionID, agentID, model, "aborted", abortErr.Error(), "hook_abort")
				outcome.terminated = true
				return outcome
			}
			if resp.ToolCall != nil {
				r.engine.PublishRuntimeHookDecisionEvent("hook_modify", HookRequest{Name: HookToolCall, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "tool_call", "modified_tool": resp.ToolCall.Name, "modified_tool_call_id": resp.ToolCall.ID})
				call = *resp.ToolCall
			}
			if injectedResult, ok := directToolResultFromHook(resp); ok {
				r.engine.PublishRuntimeHookDecisionEvent("hook_respond", HookRequest{Name: HookToolCall, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "tool_call", "source": "hook", "output_length": len(injectedResult)})
				displayResult := injectedResult
				if len(displayResult) > 100000 {
					displayResult = displayResult[:100000] + "\n... (truncated)"
				}
				r.engine.broadcast(sessionID, map[string]any{"type": "tool_finished", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "output_length": len(injectedResult)})
				r.engine.PublishRuntimeToolEvent("tool_finished", sessionID, turnID, agentID, call.Name, call.ID, iter, nil, map[string]any{"phase": "tool", "output_length": len(injectedResult), "source": "hook", "hook_phase": "tool_call"})
				logutil.WarnIfErr("append hook-responded tool.finished event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.finished", map[string]any{
					"phase":         "tool",
					"tool":          call.Name,
					"checkpoint":    true,
					"tool_call_id":  call.ID,
					"output_length": len(injectedResult),
					"source":        "hook",
					"hook_phase":    "tool_call",
				}))
				goai.AppendToolResult(convCtx, call.ID, call.Name, displayResult, false)
				logutil.WarnIfErr("add injected tool_result message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", displayResult, map[string]any{"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": false, "turn_id": turnID, "source": "hook", "hook_phase": "tool_call"}))
				outcome.lastToolFailureSig = ""
				outcome.repeatedToolFailureCount = 0
				if steerMsgs, err := r.dequeueSteeringMessages(ctx, sessionID); err != nil {
					log.Printf("steering dequeue error after hook tool response: %v", err)
				} else if len(steerMsgs) > 0 {
					outcome.pendingSteering = append(outcome.pendingSteering, steerMsgs...)
					if i+1 < len(toolCalls) {
						r.skipRemainingToolCalls(ctx, sessionID, turnID, convCtx, toolCalls, i+1)
						outcome.skipRemainingTools = true
					}
				}
				if outcome.skipRemainingTools || len(outcome.pendingSteering) > 0 {
					return outcome
				}
				continue
			}
			if resp.Block {
				reason := internalx.StringValue(resp.Reason, "tool call blocked")
				r.engine.PublishRuntimeHookDecisionEvent("hook_deny", HookRequest{Name: HookToolCall, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "tool_call", "reason": reason})
				logutil.WarnIfErr("append hook-denied tool.skipped event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.skipped", map[string]any{
					"phase":        "tool",
					"checkpoint":   true,
					"tool":         call.Name,
					"tool_call_id": call.ID,
					"reason":       reason,
					"hook_phase":   "tool_call",
				}))
				r.engine.broadcast(sessionID, map[string]any{"type": "tool_skipped", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "reason": reason})
				r.engine.PublishRuntimeToolEvent("tool_skipped", sessionID, turnID, agentID, call.Name, call.ID, iter, nil, map[string]any{"reason": reason, "phase": "tool", "hook_phase": "tool_call"})
				toolErr := fmt.Errorf("blocked by hook: %s", reason)
				errText := fmt.Sprintf("Error: %v", toolErr)
				goai.AppendToolResult(convCtx, call.ID, call.Name, errText, true)
				logutil.WarnIfErr("add blocked tool_result message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", errText, map[string]any{"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": true, "turn_id": turnID, "skipped": true, "skip_reason": reason, "hook_phase": "tool_call"}))
				continue
			}
		}
		if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookApproveTool, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, ToolCall: &call, Payload: map[string]any{"tool": call.Name, "tool_call_id": call.ID, "arguments": call.Arguments}}); err != nil {
			log.Printf("hook approve_tool error: %v", err)
		} else {
			if abortErr := hookAbortFromResponse(resp, fmt.Sprintf("tool %s approval aborted by hook", call.Name)); abortErr != nil {
				r.engine.PublishRuntimeHookDecisionEvent("hook_abort", HookRequest{Name: HookApproveTool, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "approve_tool", "reason": abortErr.Error()})
				r.finishTurn(s, turnID, sessionID, agentID, model, "aborted", abortErr.Error(), "hook_abort")
				outcome.terminated = true
				return outcome
			}
			if resp.Block {
				reason := internalx.StringValue(resp.Reason, "tool not approved")
				r.engine.PublishRuntimeHookDecisionEvent("hook_deny", HookRequest{Name: HookApproveTool, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "approve_tool", "reason": reason})
				logutil.WarnIfErr("append approve-denied tool.skipped event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.skipped", map[string]any{
					"phase":        "tool",
					"checkpoint":   true,
					"tool":         call.Name,
					"tool_call_id": call.ID,
					"reason":       reason,
					"hook_phase":   "approve_tool",
				}))
				r.engine.broadcast(sessionID, map[string]any{"type": "tool_skipped", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "reason": reason})
				r.engine.PublishRuntimeToolEvent("tool_skipped", sessionID, turnID, agentID, call.Name, call.ID, iter, nil, map[string]any{"reason": reason, "phase": "tool", "hook_phase": "approve_tool"})
				toolErr := fmt.Errorf("blocked by hook: %s", reason)
				errText := fmt.Sprintf("Error: %v", toolErr)
				goai.AppendToolResult(convCtx, call.ID, call.Name, errText, true)
				logutil.WarnIfErr("add denied tool_result message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", errText, map[string]any{"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": true, "turn_id": turnID, "skipped": true, "skip_reason": reason, "hook_phase": "approve_tool"}))
				continue
			}
			if resp.ToolCall != nil {
				r.engine.PublishRuntimeHookDecisionEvent("hook_modify", HookRequest{Name: HookApproveTool, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "approve_tool", "modified_tool": resp.ToolCall.Name, "modified_tool_call_id": resp.ToolCall.ID})
				call = *resp.ToolCall
			}
		}

		logutil.WarnIfErr("update turn waiting_on_tools phase", s.UpdateTurnStatusAndPhase(ctx, turnID, "running", "waiting_on_tools"))
		r.emitTurnStateHook(ctx, sessionID, turnID, agentID, model, "running", "waiting_on_tools", map[string]any{"reason": "tool_execution", "tool": call.Name, "iteration": iter})
		r.engine.PublishRuntimeToolEvent("tool_started", sessionID, turnID, agentID, call.Name, call.ID, iter, nil, map[string]any{"phase": "tool"})
		logutil.WarnIfErr("append tool.started event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.started", map[string]any{
			"phase": "tool", "tool": call.Name, "checkpoint": true,
			"tool_call_id": call.ID, "iteration": iter,
		}))

		r.engine.broadcast(sessionID, map[string]any{
			"type": "agent_status", "chat_jid": "gi:" + sessionID,
			"title": fmt.Sprintf("Running: %s", call.Name), "status": "running", "turn_id": turnID,
		})

		toolResult, toolErr := r.executeTool(ctx, call, sessionID, turnID)
		if toolErr != nil {
			if ctx.Err() != nil || isCancellationError(toolErr) {
				r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled during tool execution", "")
				outcome.terminated = true
				return outcome
			}
			log.Printf("tool [%s] error: %v", call.Name, toolErr)
			r.engine.broadcast(sessionID, map[string]any{"type": "tool_failed", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "error": toolErr.Error()})
			r.engine.PublishRuntimeToolEvent("tool_failed", sessionID, turnID, agentID, call.Name, call.ID, iter, toolErr, map[string]any{"phase": "tool"})
			logutil.WarnIfErr("append tool.failed event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.failed", map[string]any{
				"phase": "tool", "tool": call.Name, "checkpoint": true,
				"tool_call_id": call.ID, "error": toolErr.Error(),
			}))
			errText := fmt.Sprintf("Error: %v", toolErr)
			goai.AppendToolResult(convCtx, call.ID, call.Name, errText, true)
			logutil.WarnIfErr("add errored tool_result message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", errText, map[string]any{
				"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": true, "turn_id": turnID,
			}))
			outcome.lastToolFailureSig, outcome.repeatedToolFailureCount = nextRepeatedToolFailureCount(outcome.lastToolFailureSig, outcome.repeatedToolFailureCount, call, toolErr)
			if outcome.repeatedToolFailureCount >= repeatedToolFailureLimit {
				msg := fmt.Sprintf("Aborting after %d repeated identical tool failures: %v", outcome.repeatedToolFailureCount, toolErr)
				log.Printf("tool [%s] repeated failure guard tripped: %s", call.Name, msg)
				r.persistUsage(s, turnID, sessionID, totalUsage, iter)
				r.finishTurn(s, turnID, sessionID, agentID, model, "failed", msg, "repeated_tool_failure")
				outcome.terminated = true
				return outcome
			}
		} else {
			if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookToolResult, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, ToolCall: &call, ToolResult: toolResult, Payload: map[string]any{"tool": call.Name, "tool_call_id": call.ID, "is_error": false}}); err != nil {
				log.Printf("hook tool_result error: %v", err)
			} else {
				if abortErr := hookAbortFromResponse(resp, fmt.Sprintf("tool %s result aborted by hook", call.Name)); abortErr != nil {
					r.engine.PublishRuntimeHookDecisionEvent("hook_abort", HookRequest{Name: HookToolResult, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "tool_result", "reason": abortErr.Error()})
					r.finishTurn(s, turnID, sessionID, agentID, model, "aborted", abortErr.Error(), "hook_abort")
					outcome.terminated = true
					return outcome
				}
				if resp.ToolResult != nil {
					r.engine.PublishRuntimeHookDecisionEvent("hook_modify", HookRequest{Name: HookToolResult, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "tool_result", "output_length": len(*resp.ToolResult)})
					toolResult = *resp.ToolResult
				}
			}
			displayResult := toolResult
			if len(displayResult) > 100000 {
				displayResult = displayResult[:100000] + "\n... (truncated)"
			}
			r.engine.broadcast(sessionID, map[string]any{"type": "tool_finished", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "output_length": len(toolResult)})
			r.engine.PublishRuntimeToolEvent("tool_finished", sessionID, turnID, agentID, call.Name, call.ID, iter, nil, map[string]any{"phase": "tool", "output_length": len(toolResult)})
			logutil.WarnIfErr("append tool.finished event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.finished", map[string]any{
				"phase": "tool", "tool": call.Name, "checkpoint": true,
				"tool_call_id": call.ID, "output_length": len(toolResult),
			}))
			goai.AppendToolResult(convCtx, call.ID, call.Name, displayResult, false)
			logutil.WarnIfErr("add successful tool_result message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", displayResult, map[string]any{
				"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": false, "turn_id": turnID,
			}))
			outcome.lastToolFailureSig = ""
			outcome.repeatedToolFailureCount = 0
		}
		if steerMsgs, err := r.dequeueSteeringMessages(ctx, sessionID); err != nil {
			log.Printf("steering dequeue error after tool: %v", err)
		} else if len(steerMsgs) > 0 {
			outcome.pendingSteering = append(outcome.pendingSteering, steerMsgs...)
			if i+1 < len(toolCalls) {
				r.skipRemainingToolCalls(ctx, sessionID, turnID, convCtx, toolCalls, i+1)
				outcome.skipRemainingTools = true
			}
			return outcome
		}
	}
	return outcome
}

func (r *sessionRunner) appendFinalSteeringCheckpoint(s *store.Store, turnID, sessionID string) {
	bgCtx := r.engine.backgroundContext()
	if staged, stagedTurnID, err := r.engine.stageQueuedSteeringContinuation(bgCtx, sessionID); err != nil {
		log.Printf("steering final checkpoint: %v", err)
	} else if staged {
		logutil.WarnIfErr("append steering final checkpoint", s.AppendTurnEvent(bgCtx, turnID, sessionID, "steering.final_checkpoint", map[string]any{"phase": "steering", "checkpoint": true, "staged_turn_id": stagedTurnID}))
	}
}

type preparedTurnRun struct {
	turn            *store.Turn
	sessionID       string
	turnID          string
	prompt          string
	intent          string
	model           string
	agentID         string
	initialSteering []store.SteeringMessage
}

func (r *sessionRunner) appendCleanupHandoffFailure(ctx context.Context, sessionID, turnID, stage string, err error) {
	if strings.TrimSpace(sessionID) == "" || strings.TrimSpace(turnID) == "" || err == nil {
		return
	}
	logutil.WarnIfErr("append cleanup handoff failure event", r.store.AppendTurnEvent(ctx, turnID, sessionID, "turn.cleanup_handoff_failed", map[string]any{
		"phase":      "cleanup",
		"checkpoint": true,
		"stage":      stage,
		"error":      err.Error(),
	}))
}

func (r *sessionRunner) cleanupTurnRun(sessionID, claimToken string, active *runningTurn) {
	ctx := r.engine.backgroundContext()
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.store.ReleaseSessionActiveTurn(ctx, sessionID, claimToken); err != nil {
		log.Printf("turn coordination: release session active turn failed: %v", err)
		r.appendCleanupHandoffFailure(ctx, sessionID, claimToken, "release_active_claim", err)
		if r.current == active {
			r.current = nil
		}
		return
	}
	if r.current == active {
		r.current = nil
	}
	if err := r.store.SyncSessionQueueCount(ctx, sessionID); err != nil {
		log.Printf("turn coordination: sync session queue count failed: %v", err)
		r.appendCleanupHandoffFailure(ctx, sessionID, claimToken, "sync_queue_count", err)
		return
	}
	if hook := r.engine.beforeCleanupNextWorkHook; hook != nil {
		hook(ctx, sessionID)
	}
	launched, err := r.engine.startNextQueuedTurnLocked(ctx, r, sessionID)
	if err != nil {
		log.Printf("turn coordination: launch queued turn failed: %v", err)
	} else if launched {
		if strings.TrimSpace(claimToken) != "" {
			logutil.WarnIfErr("append cleanup handoff event", r.store.AppendTurnEvent(ctx, claimToken, sessionID, "turn.cleanup_handoff", map[string]any{"phase": "cleanup", "checkpoint": true, "handoff": "next_queued_turn"}))
		}
	} else {
		if _, _, err := r.store.GetSessionActiveTurn(ctx, sessionID); err == sql.ErrNoRows {
			if _, err := r.engine.continueQueuedSteeringLocked(ctx, r, sessionID); err != nil {
				log.Printf("steering continuation: %v", err)
				r.appendCleanupHandoffFailure(ctx, sessionID, claimToken, "continue_queued_steering", err)
			}
		} else if err != nil {
			log.Printf("turn coordination: inspect active turn after cleanup failed: %v", err)
		}
	}
	queueCount, err := r.store.CountQueuedTurns(ctx, sessionID)
	if err != nil {
		log.Printf("turn coordination: count queued turns after cleanup failed: %v", err)
		return
	}
	activeTurnID, _, err := r.store.GetSessionActiveTurn(ctx, sessionID)
	hasActiveTurn := err == nil
	if err != nil && err != sql.ErrNoRows {
		log.Printf("turn coordination: inspect active turn during cleanup normalization failed: %v", err)
		return
	}
	if hasActiveTurn {
		if err := r.engine.normalizeRunningSessionState(ctx, sessionID, activeTurnID, false, ""); err != nil {
			log.Printf("turn coordination: normalize running session state after cleanup failed: %v", err)
		}
		return
	}
	if queueCount > 0 {
		if err := r.engine.normalizeInactiveSessionState(ctx, sessionID, "queued", "", false); err != nil {
			log.Printf("turn coordination: normalize queued session state after cleanup failed: %v", err)
		}
		return
	}
	if err := r.engine.normalizeInactiveSessionState(ctx, sessionID, "idle", "", false); err != nil {
		log.Printf("turn coordination: normalize idle session state after cleanup failed: %v", err)
	}
}

func (r *sessionRunner) setupTurnRun(ctx context.Context, s *store.Store, sessionID, turnID string) (*preparedTurnRun, error) {
	turnRec, err := s.GetTurn(ctx, turnID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(sessionID) == "" {
		sessionID = turnRec.SessionID
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	initialSteering := steeringMessagesFromMetadata(turnRec.Metadata)
	prompt := turnRec.Prompt
	if strings.TrimSpace(prompt) == "" && len(initialSteering) > 0 {
		prompt = steeringPromptForShell(initialSteering)
	}
	intent := internalx.StringValue(turnRec.Metadata["intent"], "prompt")
	agentID, model := r.resolveTurnAgentAndModel(ctx, s, turnRec, sessionID, prompt)
	if hook := r.engine.beforeSetupHook; hook != nil {
		hook(ctx, sessionID, turnID)
	}
	if hook := r.engine.beforeSetupErrorHook; hook != nil {
		if err := hook(ctx, sessionID, turnID); err != nil {
			return nil, err
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	logutil.WarnIfErr("touch session state running", s.TouchSessionState(ctx, sessionID, map[string]any{"active_turn_id": turnID, "model": model, "status": "running"}))
	r.engine.PublishRuntimeSessionEvent("session_running", sessionID, agentID, "running", map[string]any{"reason": "setup", "active_turn_id": turnID, "turn_id": turnID, "turn_status": "running", "turn_phase": "setup", "model": model})
	r.emitSessionStateHookOnly(ctx, sessionID, agentID, model, "running", map[string]any{"reason": "setup", "active_turn_id": turnID, "turn_id": turnID, "turn_status": "running", "turn_phase": "setup"})
	r.engine.PublishRuntimeTurnEvent("turn_started", sessionID, turnID, agentID, "running", "setup", map[string]any{"reason": "setup", "model": model})
	r.emitTurnStateHookOnly(ctx, sessionID, turnID, agentID, model, "running", "setup", map[string]any{"reason": "setup"})
	userPayload := map[string]any{"kind": "chat", "intent": intent, "turn_id": turnID}
	copySelectedMetadata(userPayload, turnRec.Metadata, ingressMetadataKeys)
	if strings.TrimSpace(prompt) != "" && len(initialSteering) == 0 {
		logutil.WarnIfErr("add user prompt message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "user", prompt, userPayload))
	}
	startedPayload := map[string]any{"phase": "turn", "prompt": prompt, "intent": intent, "model": model, "checkpoint": true}
	copySelectedMetadata(startedPayload, turnRec.Metadata, turnStartedMetadataKeys)
	logutil.WarnIfErr("append turn.started event", s.AppendTurnEvent(ctx, turnID, sessionID, "turn.started", startedPayload))
	return &preparedTurnRun{
		turn:            turnRec,
		sessionID:       sessionID,
		turnID:          turnID,
		prompt:          prompt,
		intent:          intent,
		model:           model,
		agentID:         agentID,
		initialSteering: initialSteering,
	}, nil
}

func (r *sessionRunner) resolveTurnAgentAndModel(ctx context.Context, s *store.Store, turnRec *store.Turn, sessionID, prompt string) (string, string) {
	opCtx := store.CoordinationContext(ctx, r.engine.backgroundContext())
	model := internalx.StringValue(turnRec.Metadata["model"], "bootstrap")
	agentID := "agent"
	if identity, err := s.RequireSessionIdentityRuntime(opCtx, sessionID); err == nil {
		agentID = identity.AgentID
	}
	agentModel := r.engine.modelForAgent(agentID)
	if strings.TrimSpace(model) == "" {
		model = agentModel
	}
	if model == agentModel {
		history, _ := s.ListMessages(opCtx, sessionID)
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
	return agentID, model
}

func (r *sessionRunner) runPreparedTurn(ctx context.Context, s *store.Store, run *preparedTurnRun) {
	if run.model != "bootstrap" && run.model != "test-model" && run.model != "" {
		r.runAgentLoop(ctx, s, run.turnID, run.sessionID, run.model, run.agentID, run.initialSteering)
		return
	}
	r.runShellTurn(ctx, s, run)
}

func (r *sessionRunner) runShellTurn(ctx context.Context, s *store.Store, run *preparedTurnRun) {
	if len(run.initialSteering) > 0 {
		r.persistSteeringMessages(ctx, run.sessionID, run.turnID, run.initialSteering)
	}
	r.engine.PublishRuntimeToolEvent("tool_started", run.sessionID, run.turnID, run.agentID, "shell", "", 0, nil, map[string]any{"phase": "tool", "command": []string{"sh", "-lc", "printf 'Gi received: %s' \"$GI_PROMPT\""}})
	logutil.WarnIfErr("append shell tool.started event", s.AppendTurnEvent(ctx, run.turnID, run.sessionID, "tool.started", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "command": []string{"sh", "-lc", "printf 'Gi received: %s' \"$GI_PROMPT\""}}))

	out, runErr, cancelled := tools.RunShellPrompt(ctx, run.prompt, func(cmd *exec.Cmd) {
		r.mu.Lock()
		if r.current != nil && r.current.turnID == run.turnID {
			r.current.cmdMu.Lock()
			r.current.cmd = cmd
			r.current.cmdMu.Unlock()
		}
		r.mu.Unlock()
	}, func(delta string) {
		if strings.TrimSpace(delta) == "" {
			return
		}
		r.engine.broadcast(run.sessionID, map[string]any{
			"type":     "agent_draft_delta",
			"chat_jid": "gi:" + run.sessionID,
			"delta":    delta,
			"turn_id":  run.turnID,
		})
	})
	if cancelled {
		bgCtx := r.engine.backgroundContext()
		r.appendFinalSteeringCheckpoint(s, run.turnID, run.sessionID)
		logutil.WarnIfErr("append turn.cancelled event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "turn.cancelled", map[string]any{"phase": "cancel", "checkpoint": true, "reason": "cancelled", "status": "cancelled", "turn_phase": "aborted", "failure_kind": ""}))
		logutil.WarnIfErr("append turn.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "turn.finished", map[string]any{"phase": "turn", "checkpoint": true, "status": "cancelled", "reason": "cancelled", "failure_kind": ""}))
		logutil.WarnIfErr("update turn status cancelled", s.UpdateTurnStatus(bgCtx, run.turnID, "cancelled"))
		r.engine.PublishRuntimeTurnEvent("turn_terminal", run.sessionID, run.turnID, run.agentID, "cancelled", "aborted", map[string]any{"reason": "cancelled", "failure_kind": ""})
		r.emitTurnStateHookOnly(bgCtx, run.sessionID, run.turnID, run.agentID, run.model, "cancelled", "aborted", map[string]any{"reason": "cancelled", "failure_kind": ""})
		r.propagateChildSubTurnCancellation(bgCtx, run.turnID, "cancelled", "")
		r.publishSubTurnLifecycle(bgCtx, run.turnID, "cancelled")
		msgID := store.NowID("msg")
		logutil.WarnIfErr("add turn cancelled system message", s.AddMessage(bgCtx, msgID, run.sessionID, "system", "Turn cancelled", map[string]any{"kind": "status", "turn_id": run.turnID, "clipped": true}))
		r.broadcastSystemPost(run.sessionID, run.turnID, msgID, "Turn cancelled")
		logutil.WarnIfErr("touch session idle after cancel", s.TouchSessionState(bgCtx, run.sessionID, map[string]any{"status": "idle", "active_turn_id": nil}))
		r.engine.PublishRuntimeSessionEvent("session_idle", run.sessionID, run.agentID, "idle", map[string]any{"reason": "turn_terminal", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "cancelled", "turn_phase": "aborted", "failure_kind": "", "model": run.model})
		r.emitSessionStateHookOnly(bgCtx, run.sessionID, run.agentID, run.model, "idle", map[string]any{"reason": "turn_terminal", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "cancelled", "turn_phase": "aborted", "failure_kind": ""})
		return
	}
	if runErr != nil {
		bgCtx := r.engine.backgroundContext()
		r.appendFinalSteeringCheckpoint(s, run.turnID, run.sessionID)
		logutil.WarnIfErr("turn failure mark", s.MarkTurnFailureWithFallbackErr(bgCtx, nil, run.turnID, run.sessionID, "shell_error", "none", runErr.Error()))
		r.engine.PublishRuntimeToolEvent("tool_failed", run.sessionID, run.turnID, run.agentID, "shell", "", 0, runErr, map[string]any{"phase": "tool"})
		logutil.WarnIfErr("append shell tool.failed event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "tool.failed", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "error": runErr.Error(), "failure_kind": "shell_error"}))
		msgID := store.NowID("msg")
		logutil.WarnIfErr("add shell failure system message", s.AddMessage(bgCtx, msgID, run.sessionID, "system", fmt.Sprintf("Shell tool failed: %v", runErr), map[string]any{"kind": "status", "turn_id": run.turnID, "source": "system", "failure_kind": "shell_error"}))
		r.broadcastSystemPost(run.sessionID, run.turnID, msgID, fmt.Sprintf("Shell tool failed: %v", runErr))
		logutil.WarnIfErr("append turn.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "turn.finished", map[string]any{"phase": "turn", "checkpoint": true, "status": "failed", "reason": "shell_error", "failure_kind": "shell_error"}))
		logutil.WarnIfErr("update turn status failed", s.UpdateTurnStatus(bgCtx, run.turnID, "failed"))
		r.engine.PublishRuntimeTurnEvent("turn_terminal", run.sessionID, run.turnID, run.agentID, "failed", "failed", map[string]any{"reason": "shell_error", "failure_kind": "shell_error"})
		r.emitTurnStateHookOnly(bgCtx, run.sessionID, run.turnID, run.agentID, run.model, "failed", "failed", map[string]any{"reason": "shell_error", "failure_kind": "shell_error"})
		r.propagateChildSubTurnCancellation(bgCtx, run.turnID, "failed", "shell_error")
		r.publishSubTurnLifecycle(bgCtx, run.turnID, "failed")
		logutil.WarnIfErr("touch session idle after failure", s.TouchSessionState(bgCtx, run.sessionID, map[string]any{"status": "idle", "active_turn_id": nil}))
		r.engine.PublishRuntimeSessionEvent("session_idle", run.sessionID, run.agentID, "idle", map[string]any{"reason": "turn_terminal", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "failed", "turn_phase": "failed", "failure_kind": "shell_error", "model": run.model})
		r.emitSessionStateHookOnly(bgCtx, run.sessionID, run.agentID, run.model, "idle", map[string]any{"reason": "turn_terminal", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "failed", "turn_phase": "failed", "failure_kind": "shell_error"})
		return
	}
	bgCtx := r.engine.backgroundContext()
	r.appendFinalSteeringCheckpoint(s, run.turnID, run.sessionID)
	r.engine.PublishRuntimeToolEvent("tool_finished", run.sessionID, run.turnID, run.agentID, "shell", "", 0, nil, map[string]any{"phase": "tool", "output_length": len(out)})
	logutil.WarnIfErr("append shell tool.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "tool.finished", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "output_length": len(out)}))
	msgID := store.NowID("msg")
	logutil.WarnIfErr("add shell assistant message", s.AddMessage(bgCtx, msgID, run.sessionID, "assistant", out, map[string]any{"kind": "chat", "source": "shell", "turn_id": run.turnID, "agent_id": run.agentID}))
	r.engine.broadcast(run.sessionID, map[string]any{
		"type": "new_post", "id": msgID, "chat_jid": "gi:" + run.sessionID,
		"content": out, "timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"sender": "agent", "is_bot_message": true,
		"data": map[string]any{"type": "agent_response", "content": out, "agent_id": run.agentID},
	})
	completionPayload := map[string]any{"reason": "completed", "completion_kind": "response"}
	logutil.WarnIfErr("append turn.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "turn.finished", map[string]any{"phase": "turn", "checkpoint": true, "status": "completed", "reason": "completed", "failure_kind": "", "completion_kind": "response"}))
	logutil.WarnIfErr("update turn status completed", s.UpdateTurnStatus(bgCtx, run.turnID, "completed"))
	r.engine.PublishRuntimeTurnEvent("turn_completed", run.sessionID, run.turnID, run.agentID, "completed", "completed", completionPayload)
	r.emitTurnStateHookOnly(bgCtx, run.sessionID, run.turnID, run.agentID, run.model, "completed", "completed", completionPayload)
	r.propagateChildSubTurnCancellation(bgCtx, run.turnID, "completed", "")
	r.publishSubTurnLifecycle(bgCtx, run.turnID, "completed")
	logutil.WarnIfErr("touch session idle after completion", s.TouchSessionState(bgCtx, run.sessionID, map[string]any{"status": "idle", "active_turn_id": nil}))
	sessionCompletionPayload := map[string]any{"reason": "turn_completed", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "completed", "turn_phase": "completed", "failure_kind": "", "model": run.model, "completion_kind": "response"}
	r.engine.PublishRuntimeSessionEvent("session_idle", run.sessionID, run.agentID, "idle", sessionCompletionPayload)
	r.emitSessionStateHookOnly(bgCtx, run.sessionID, run.agentID, run.model, "idle", map[string]any{"reason": "turn_completed", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "completed", "turn_phase": "completed", "failure_kind": "", "completion_kind": "response"})
}
