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
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/connectivity"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/peering"
	"github.com/rcarmo/gi/internal/routing"
	"github.com/rcarmo/gi/internal/routing/routedsession"
	"github.com/rcarmo/gi/internal/scripting"
	"github.com/rcarmo/gi/internal/shellenv"
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
	autoCompactionOverride            atomic.Int32 // SetAutoCompaction: 0 unset, 1 on, 2 off
	hooks                             *HookRegistry
	tools                             *tools.ToolRegistry
	connectivity                      *connectivity.Registry
	topics                            *topics.Bus
	peering                           *peering.Manager
	bgCtx                             context.Context
	bgCancel                          context.CancelFunc
	maintenance                       sync.WaitGroup
	mcp                               *mcpState      // nil unless EnableMCP was called
	mcpConfigErrors                   []error        // config errors when no server is configured
	mcpNotices                        mcpNotices     // for the UI (SetMCPNotifier)
	mcpWaited                         atomic.Bool    // the first prompt waited for direct MCP servers
	codemode                          *codemodeState // nil when -builtin:codemode
	closing                           atomic.Bool    // set by Close: no new launches
	providerRecoveryEnabled           atomic.Bool    // explicit runtime-ready barrier before effectful recovery
	launchMu                          sync.Mutex     // host admission versus shutdown
	closeOnce                         sync.Once
	closeErr                          error
	hookMu                            sync.Mutex
	hooksClosed                       bool
	hookRuns                          sync.WaitGroup
	hookHosts                         map[string]*hookHostGroup
	hookProcesses                     map[*mountedProcessHook]struct{}
	runs                              sync.WaitGroup // in-flight runTurn goroutines
	sideMu                            sync.Mutex     // serialises side admission with Close
	sideRuns                          sync.WaitGroup
	extensions                        []ExtensionInfo
	extensionsMu                      sync.RWMutex
	extensionCommands                 *ExtensionCommandRegistry
	sessions                          sync.Map // sessionID -> *sessionRunner
	subs                              map[string]map[chan map[string]any]bool
	subsMu                            sync.Mutex
	beforeSetupHook                   func(context.Context, string, string)
	beforeSetupErrorHook              func(context.Context, string, string) error
	beforeCreateSubTurnErrorHook      func(context.Context, string, string) error
	beforeLaunchClaimHook             func(context.Context, string, string)
	beforeLaunchSessionStateErrorHook func(context.Context, string, string) error
	beforeCleanupNextWorkHook         func(context.Context, string)
	// Test seam; nil uses the joined native shell process runner.
	shellPromptRunner func(context.Context, string, func(string)) (string, error, bool)
}

type sharedSessionCoord struct {
	mu      sync.Mutex
	current *runningTurn
	hosts   map[string]chan struct{}
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
	owner               *Engine // immutable host owner; session coordination may be shared
	turnID              string
	claimToken          string
	cancel              context.CancelFunc
	done                chan struct{}
	cancelled           <-chan struct{}
	terminalPending     bool // guarded by sharedSessionCoord.mu
	checkpointClosing   bool // graceful retry reservation; no terminal lifecycle yet
	recoveredTerminal   *store.TurnSettlement
	recoveredRetry      *store.ProviderRetryCheckpoint
	recoveredGeneration *store.GenerationReadyCheckpoint
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
	Channel    string `json:"channel,omitempty"`
	Account    string `json:"account,omitempty"`
}

type SubmitResult struct {
	SourceCommitted bool   `json:"source_committed,omitempty"`
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
		store:             s,
		systemPrompt:      systemPrompt,
		routeResolver:     routing.NewRouteResolver(cfg.Agents, cfg.Session),
		modelRouter:       routing.NewRouter(cfg.Routing),
		runtimeCfg:        cfg,
		hooks:             NewHookRegistry(),
		tools:             tools.NewToolRegistry(),
		connectivity:      connectivity.NewRegistry(),
		topics:            topics.NewBus(),
		extensionCommands: NewExtensionCommandRegistry(),
		peering:           peering.NewManager(cfg.Peering, cfg.WorkspaceRoot),
		bgCtx:             bgCtx,
		bgCancel:          bgCancel,
		subs:              map[string]map[chan map[string]any]bool{},
	}
	e.registerDefaultTools()
	e.registerSessionPlanContext()
	e.registerCodemodeTool()
	e.startTopicBridge()
	// User-level skills live outside the workspace; the prompt tells the model
	// to load them with read (Pi), so their directories are readable.
	readable := giskills.UserSkillDirs()
	for _, s := range cfg.Discovery.Skills {
		if s.Source == "user" {
			readable = append(readable, s.BaseDir)
		}
	}
	tools.SetReadOnlyRoots(readable)
	if e.store != nil {
		e.maintenance.Add(1)
		go func() {
			defer e.maintenance.Done()
			e.runMCPOutputPruner(e.backgroundContext())
		}() // mcp-output and codemode-output
	}
	if e.store != nil {
		if _, err := e.recoverInterruptedTurns(e.backgroundContext(), ""); err != nil {
			log.Printf("turn recovery: startup scan failed: %v", err)
		}
		if err := e.handoffLegacyStoppedQueues(e.backgroundContext()); err != nil {
			log.Printf("turn recovery: legacy Stop queue handoff failed: %v", err)
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
	return tool.Executor(ctx, tools.ToolRuntime{Store: e.store, SessionID: sessionID, WorkspaceRoot: e.runtimeCfg.WorkspaceRoot, PublishMessage: e.publishWidgetMessage, PublishPlan: e.PublishPlanChanged}, goai.ToolCall{Name: name, Arguments: args})
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

// shutdownGrace bounds cancellation admission and the first join diagnostic.
// Close continues joining after it expires.
const shutdownGrace = 5 * time.Second

// Close fences launches and joins main, hook, side and maintenance hosts.
// Concurrent callers share the shutdown barrier and its result. The grace
// interval controls diagnostics only; it never releases a live owned host.
// Verified retry checkpoints are requeued only after all owned work joins.
// Other phases still use conservative abort settlement (TODO #86/#92).
func (e *Engine) Close() error {
	if e == nil {
		return nil
	}
	e.closeOnce.Do(func() {
		e.launchMu.Lock()
		e.closing.Store(true)
		e.launchMu.Unlock()
		checkpoints, err := e.stopActiveTurnsForClose(shutdownGrace)
		e.closeErr = errors.Join(e.closeErr, err)
		e.runs.Wait()
		e.closeHookHosts()
		// Fence side admission before waiting. Preserve main-turn finalisation
		// ordering, then cancel background work and independent side requests.
		e.sideMu.Lock()
		if e.bgCancel != nil {
			e.bgCancel()
		}
		e.sideMu.Unlock()
		e.sideRuns.Wait()
		e.closeMCP()
		e.maintenance.Wait() // finish DB writers before callers close/remove the store
		if e.peering != nil {
			e.closeErr = errors.Join(e.closeErr, e.peering.Close())
		}
		for _, checkpoint := range checkpoints {
			// Join even the closing lease before publishing a claim that a
			// ready peer may immediately adopt. Its heartbeat must not race
			// that legitimate ownership handoff.
			e.closeErr = errors.Join(e.closeErr, checkpoint.stopHeartbeat())
		}
		for _, checkpoint := range checkpoints {
			if checkpoint.reserved {
				e.closeErr = errors.Join(e.closeErr, e.store.CommitProviderRetryClose(context.Background(), checkpoint.sessionID, checkpoint.turnID, checkpoint.token))
			}
		}
	})
	return e.closeErr
}

// ErrTurnOwnedElsewhere reports a fresh active-turn claim held by another
// live gi process sharing this database.
var ErrTurnOwnedElsewhere = errors.New("active turn is owned by another running gi process")

// LiveActiveTurn returns the turn this engine is currently running for the
// session, or "" when no worker in this process owns one.
func (e *Engine) LiveActiveTurn(sessionID string) string {
	v, ok := e.sessions.Load(sessionID)
	if !ok {
		return ""
	}
	runner := v.(*sessionRunner)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if runner.current == nil {
		return ""
	}
	return runner.current.turnID
}

// AbortStaleActiveTurn finalizes a session's active turn that no worker in
// this process is running (left behind by a crash or killed process) as
// aborted, releases its claim and returns the session to idle, instead of
// letting stale-claim recovery replay it. Claims with a fresh heartbeat
// belong to another live process and are refused.
func (e *Engine) AbortStaleActiveTurn(ctx context.Context, sessionID string) (string, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	activeID, _, err := e.store.GetSessionActiveTurn(opCtx, sessionID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && activeID == "") {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if live := e.LiveActiveTurn(sessionID); live == activeID {
		return "", store.ErrQueueConflict // live: callers use the normal abort path
	}
	claims, err := e.store.ListStaleActiveTurnClaims(opCtx, time.Now().Add(-interruptedTurnStaleAfter), sessionID)
	if err != nil {
		return "", err
	}
	for _, claim := range claims {
		if claim.TurnID != activeID {
			continue
		}
		claim.Phase = "cancelling" // recovery finalizes cancelling claims as aborted
		if err := e.recoverInterruptedTurn(opCtx, claim); err != nil {
			return "", err
		}
		return activeID, nil
	}
	return "", ErrTurnOwnedElsewhere
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
	if opCtx == nil {
		return context.Canceled
	}
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
	if opCtx == nil {
		return context.Canceled
	}
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
	return e.submitPrompt(ctx, in, nil, "", "")
}

func (e *Engine) submitPrompt(ctx context.Context, in RunInput, retry *store.HeldRetryAdmission, composerToken, mediaToken string) (*SubmitResult, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	// Direct MCP tools must be registered before the turn's tool set is
	// admitted; wait (bounded) for their servers outside the runner lock.
	e.awaitDirectMCPTools(opCtx)
	e.providerRecoveryEnabled.Store(true)
	if in.Intent == "" {
		in.Intent = "prompt"
	}
	// Sanitize before steering too, not just before queued-turn insertion.
	// Only SubmitTUIComposer may supply the engine-owned text receipt token.
	in.Metadata = cloneMap(in.Metadata)
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	delete(in.Metadata, "tui_text_claim")
	delete(in.Metadata, "tui_media_claim")
	if composerToken != "" {
		in.Metadata["tui_text_claim"] = composerToken
	}
	if mediaToken != "" {
		in.Metadata["tui_media_claim"] = mediaToken
	}
	if opCtx == nil {
		return nil, context.Canceled
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
	activeQueue := false
	admissionConflict := false
	if activeTurnID, _, err := e.store.GetSessionActiveTurn(opCtx, in.SessionID); err == nil {
		if active, getErr := e.store.GetTurn(opCtx, activeTurnID); getErr == nil && (active.Status == "running" || active.Status == "cancelling") && active.Metadata["operation"] == "manual_compaction" {
			return nil, store.ErrQueueConflict
		}
		if in.Intent != "queue" {
			result, err := e.submitSteeringPrompt(ctx, in.SessionID, activeTurnID, in)
			if !errors.Is(err, store.ErrQueueConflict) {
				return result, err
			}
			admissionConflict = true
			// Completion can precede claim release. Keep that claim owned by
			// cleanup, but persist this prompt as a distinct queued turn.
		}
		activeQueue = true
	} else if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	// Held, unconsumed queue Steer rows are visible but must neither auto-send
	// nor prevent a new explicit user submission from starting.
	next, err := e.store.GetNextQueuedTurn(opCtx, in.SessionID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	queued := activeQueue || next != nil
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
	if rawMedia, ok := in.Metadata["media"]; ok {
		mediaRefs, err := e.resolveMediaReferences(ctx, in.SessionID, rawMedia)
		if err != nil {
			return nil, err
		}
		if len(mediaRefs) > 0 {
			metadata["media"] = mediaRefs
		}
	}
	effectiveTools, restrictedTools, err := e.resolveEffectiveToolNames(parentTurn, in.Metadata)
	if err != nil {
		return nil, err
	}
	if parentTurn == nil {
		effectiveTools = e.applySessionCodemodeToggle(opCtx, in.SessionID, effectiveTools)
	}
	subTurnToolsRestricted = restrictedTools
	for k, v := range in.Metadata {
		// Admission receipts are engine-owned, not arbitrary caller metadata.
		if retry == nil && strings.HasPrefix(k, "retry_") {
			continue
		}
		if k == "effective_tools" || k == "subturn_tools_restricted" || k == "media" || k == "operation" || k == "context_token" || k == "selected_thinking_level" || k == "selected_thinking_model" {
			continue
		}
		metadata[k] = v
	}
	// Snapshot validated session thinking once at admission. Caller metadata
	// cannot change active/queued turns, and runtime never rereads mutable settings.
	selectedSession, err := e.store.GetSession(opCtx, in.SessionID)
	if err != nil {
		return nil, err
	}
	if level := e.admissionThinking(selectedSession, in.Model); level != "" {
		metadata["selected_thinking_level"] = level
		metadata["selected_thinking_model"] = in.Model
	}
	metadata["effective_tools"] = effectiveTools
	if in.ParentTurnID != "" {
		metadata["subturn_tools_restricted"] = subTurnToolsRestricted
	}
	subturnMetadata := map[string]any{"intent": in.Intent, "model": in.Model, "depth": subTurnDepth, "max_depth": subTurnMaxDepth, "max_concurrency": subTurnMaxConcurrency, "delivery_mode": subTurnDeliveryMode, "subturn_critical": subTurnCritical, "effective_tools": effectiveTools, "subturn_tools_restricted": subTurnToolsRestricted}
	if retry != nil {
		var sub *store.RetrySubTurn
		if in.ParentTurnID != "" && parentSessionID != "" {
			if hook := e.beforeCreateSubTurnErrorHook; hook != nil {
				if err := hook(opCtx, in.ParentTurnID, turnID); err != nil {
					return nil, err
				}
			}
			sub = &store.RetrySubTurn{ParentTurnID: in.ParentTurnID, ParentSessionID: parentSessionID, DeliveryMode: subTurnDeliveryMode, Depth: subTurnDepth, Metadata: subturnMetadata}
		}
		if err := e.store.AdmitHeldRetry(opCtx, *retry, turnID, in.SessionID, in.Prompt, metadata, sub); err != nil {
			return nil, err
		}
	} else if _, err := e.store.CreateTurnWithStatus(opCtx, turnID, in.SessionID, "queued", in.Prompt, metadata); err != nil {
		return nil, err
	}
	durableCtx := e.backgroundContext()
	if in.ParentTurnID != "" && parentSessionID != "" {
		if retry == nil {
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
		if !launched && in.Intent != "queue" {
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
		// A follow-up must not make its still-running origin look idle.
		if _, _, claimErr := e.store.GetSessionActiveTurn(durableCtx, in.SessionID); claimErr == sql.ErrNoRows {
			sessionStateUpdate["status"] = "queued"
			sessionStateUpdate["active_turn_id"] = nil
		}
	}
	if len(sessionStateUpdate) > 0 {
		logutil.WarnIfErr("touch session state after submit", e.store.TouchSessionState(durableCtx, in.SessionID, sessionStateUpdate))
	}
	// Another Store/engine can finish cleanup before our fallback row exists.
	// If its claim is already gone, launch queued work after admission events
	// are durable; otherwise the claim owner's cleanup remains responsible.
	if queued && admissionConflict {
		if _, _, claimErr := e.store.GetSessionActiveTurn(durableCtx, in.SessionID); claimErr == sql.ErrNoRows {
			if _, launchErr := e.startNextQueuedTurnLocked(durableCtx, runner, in.SessionID); launchErr != nil {
				return nil, launchErr
			}
		}
	}
	return &SubmitResult{TurnID: turnID, SessionID: in.SessionID, Status: status, Queued: queued}, nil
}

func (e *Engine) CancelQueuedTurn(ctx context.Context, sessionID, turnID string) error {
	return e.cancelTurn(ctx, sessionID, turnID, true, false)
}

func (e *Engine) CancelTurn(ctx context.Context, sessionID, turnID string) error {
	return e.cancelTurn(ctx, sessionID, turnID, false, false)
}

// CancelActiveTurn never falls back to queued cancellation after a run ends.
func (e *Engine) CancelActiveTurn(ctx context.Context, sessionID, turnID string) error {
	return e.cancelTurn(ctx, sessionID, turnID, false, true)
}

// StopWebActiveTurn aborts the selected run. Queued work follows normal cleanup,
// matching Piclaw's abort/finalization path; Stop does not pause the queue.
func (e *Engine) StopWebActiveTurn(ctx context.Context, sessionID, turnID string) error {
	return e.CancelActiveTurn(ctx, sessionID, turnID)
}

// AbortActiveAndRestoreTUITextDraft holds the runner lock until the durable
// cancellation and text-only draft restoration have committed. On conflict
// the worker is not signalled and queued work remains deliverable.
func (e *Engine) AbortActiveAndRestoreTUITextDraft(ctx context.Context, sessionID, turnID string, expected store.TUITextDraft) (store.TUITextDraft, []string, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	runner := e.runner(sessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if e.closing.Load() {
		return store.TUITextDraft{}, nil, context.Canceled
	}
	if runner.current == nil || runner.current.turnID != turnID || turnID == "" {
		return store.TUITextDraft{}, nil, store.ErrQueueConflict
	}
	turn, err := e.store.GetTurn(opCtx, turnID)
	if err != nil || turn.SessionID != sessionID {
		return store.TUITextDraft{}, nil, store.ErrQueueConflict
	}
	agentID, model := runner.resolveTurnAgentAndModel(opCtx, e.store, turn, sessionID, turn.Prompt)
	restored, ids, err := e.store.AbortActiveTurnAndRestoreQueuedTUITextDraft(opCtx, sessionID, turnID, runner.current.claimToken, expected)
	if err != nil {
		return store.TUITextDraft{}, nil, err
	}
	e.PublishRuntimeTurnEvent("turn_cancelling", sessionID, turnID, agentID, "cancelling", "cancelling", map[string]any{"reason": "escape_restore", "failure_kind": ""})
	runner.emitTurnStateHook(opCtx, sessionID, turnID, agentID, model, "cancelling", "cancelling", map[string]any{"reason": "escape_restore", "failure_kind": ""})
	runner.current.cancel()
	return restored, ids, nil
}

func (r *sessionRunner) runTurn(s *store.Store, sessionID, turnID string, ctx context.Context, cancel context.CancelFunc, active *runningTurn) {
	ctx = store.WithTurnWriter(ctx, sessionID, turnID, active.claimToken)
	defer cancel()
	defer func() {
		r.engine.joinTurnHooks(turnID, true)
		r.cleanupTurnRun(sessionID, active.claimToken, active)
		r.engine.forgetTurnHooks(turnID)
	}()

	if active.recoveredGeneration != nil {
		defer r.ownTurnHeartbeat(ctx, sessionID, active.claimToken, cancel)()
		agentID, _ := r.resolveTurnIdentityForFinalize(ctx, s, sessionID, turnID)
		r.runAgentLoopFromCheckpoint(ctx, s, turnID, sessionID, active.recoveredGeneration.Model, agentID, nil, nil, active.recoveredGeneration)
		return
	}
	if active.recoveredRetry != nil {
		defer r.ownTurnHeartbeat(ctx, sessionID, active.claimToken, cancel)()
		agentID, _ := r.resolveTurnIdentityForFinalize(ctx, s, sessionID, turnID)
		r.runAgentLoop(ctx, s, turnID, sessionID, active.recoveredRetry.Wait.Model, agentID, nil, active.recoveredRetry)
		return
	}
	if active.recoveredTerminal != nil {
		defer r.ownTurnHeartbeat(ctx, sessionID, active.claimToken, cancel)()
		agentID, model := r.resolveTurnIdentityForFinalize(ctx, s, sessionID, turnID)
		decision := active.recoveredTerminal
		r.finishTurnWithPayload(s, turnID, sessionID, agentID, model, decision.Status, decision.Message, decision.FailureKind, decision.Payload)
		return
	}
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
	defer r.ownTurnHeartbeat(ctx, sessionID, active.claimToken, cancel)()
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
	if opCtx == nil {
		return nil, context.Canceled
	}
	summary = normalizeHoldResolutionSummary(summary)
	turnRec, err := e.store.GetTurn(opCtx, turnID)
	if err != nil {
		return nil, err
	}
	failureRec, err := e.store.GetTurnFailure(opCtx, turnID)
	if err != nil {
		return nil, err
	}
	if failureRec.ResolutionState == "retry_pending" || (failureRec.ResolutionState == "retried" && failureRec.RetryAdmissionVersion == 1) {
		admitted, err := e.store.ReconcileHeldRetry(opCtx, turnID, failureRec.RetryAdmissionToken, summary)
		if err != nil {
			return nil, err
		}
		result, err := e.recoveredRetryResult(opCtx, turnRec.SessionID, admitted)
		if err != nil {
			return nil, err
		}
		e.publishRetryResolution(opCtx, turnRec, result.TurnID)
		return result, nil
	}
	if failureRec.HoldState == "none" {
		return nil, fmt.Errorf("retry held turn: turn %s is not currently held", turnID)
	}
	token := store.NowID("retry")
	reserved, err := e.store.ReserveAtomicHeldRetry(opCtx, turnID, token)
	if err != nil {
		return nil, err
	}
	if !reserved {
		return nil, store.ErrFailureConflict
	}
	metadata := retryMetadata(turnRec.Metadata)
	metadata["retry_of_turn_id"] = turnID
	metadata["retry_failure_kind"] = failureRec.FailureKind
	metadata["retry_hold_state"] = failureRec.HoldState
	metadata["failure_resolution"] = "retry"
	metadata["retry_admission_token"] = token
	result, err := e.submitPrompt(opCtx, RunInput{
		SessionID: turnRec.SessionID,
		Prompt:    turnRec.Prompt,
		// A retry is a durable follow-on, never implicit steering into other work.
		Intent:       "queue",
		Model:        internalx.StringValue(turnRec.Metadata["model"], ""),
		ParentTurnID: internalx.StringValue(turnRec.Metadata["parent_turn_id"], ""),
		Metadata:     metadata,
	}, &store.HeldRetryAdmission{OriginalTurnID: turnID, Token: token, Summary: summary}, "", "")
	admitted, reconcileErr := e.store.FinishHeldRetry(opCtx, turnID, token, summary, err != nil)
	if reconcileErr != nil {
		return nil, fmt.Errorf("retry admission requires reconciliation: %w", reconcileErr)
	}
	if admitted == "" {
		if err != nil {
			return nil, err
		}
		return nil, store.ErrRetryPending
	}
	if result == nil || result.TurnID != admitted {
		result, err = e.recoveredRetryResult(opCtx, turnRec.SessionID, admitted)
		if err != nil {
			return nil, err
		}
	}
	e.publishRetryResolution(opCtx, turnRec, result.TurnID)
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

var ingressMetadataKeys = []string{"source_session_id", "source_agent_id", "target_agent_id", "routed_from_prompt", "ingress_kind", "ingress_source_kind", "ingress_source_id", "ingress_role", "ingress_label", "ingress_channel", "ingress_account", "ingress_session_key"}
var turnStartedMetadataKeys = []string{"parent_turn_id", "route_mode", "route_matched_by", "source_session_id", "source_agent_id", "target_agent_id", "routed_from_prompt", "ingress_kind", "ingress_source_kind", "ingress_source_id", "ingress_role", "ingress_label", "ingress_channel", "ingress_account", "ingress_session_key"}

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
	if value := strings.TrimSpace(in.Origin.Channel); value != "" {
		metadata["ingress_channel"] = value
	}
	if value := strings.TrimSpace(in.Origin.Account); value != "" {
		metadata["ingress_account"] = value
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
			"channel":     strings.TrimSpace(in.Origin.Channel),
			"account":     strings.TrimSpace(in.Origin.Account),
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
			Channel:    strings.TrimSpace(internalx.StringValue(origin["channel"], "")),
			Account:    strings.TrimSpace(internalx.StringValue(origin["account"], "")),
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
	case "compaction", "compaction_started", "compaction_completed", "compaction_cancelled", "compaction_failed", "compaction_suppressed":
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
	turnID = strings.TrimSpace(turnID)
	status = strings.TrimSpace(status)
	phase = strings.TrimSpace(phase)
	body["turn_id"] = turnID
	body["status"] = status
	body["phase"] = phase
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
	topic = strings.TrimSpace(topic)
	envelopeType = strings.TrimSpace(tools.FirstNonEmpty(envelopeType, "notice"))
	sessionID = strings.TrimSpace(sessionID)
	agentID = strings.TrimSpace(agentID)
	if sessionID != "" {
		body["session_id"] = sessionID
	}
	if agentID != "" {
		body["agent_id"] = agentID
	}
	env := topics.Envelope{
		Topic:     topic,
		SessionID: sessionID,
		AgentID:   agentID,
		Source:    "runtime",
		Type:      envelopeType,
		Payload:   body,
	}
	e.publishTopicEvent(env)
	if topic != "runtime" {
		aggregatePayload := cloneMap(body)
		aggregatePayload["runtime_topic"] = topic
		e.publishTopicEvent(topics.Envelope{Topic: "runtime", SessionID: env.SessionID, AgentID: env.AgentID, Source: env.Source, Type: env.Type, Payload: aggregatePayload})
	}
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
		if (claim.Phase == "generation_ready" || claim.Phase == "retry_wait" || claim.Phase == "running" || store.IsClosedProviderRetryPhase(claim.Phase)) && !e.providerRecoveryEnabled.Load() {
			continue
		}
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

// recoveryDispositionForClaim decides what happens to a turn whose claim
// went stale (gi stopped mid-turn). A turn that may have sent a request or
// run tools is held for an explicit /retry instead of being replayed, which
// could repeat side effects (#21); only turns that had not started yet are
// requeued.
func recoveryDispositionForClaim(claim store.ActiveTurnClaim) string {
	switch claim.Phase {
	case "waiting_on_tools":
		return "hold_for_retry_or_skip_after_tool_checkpoint"
	case "cancelling":
		return "abort_cancelling"
	case "completed", "failed", "aborted", "cancelled", "held_for_retry_or_skip":
		return "release_terminal"
	case "queued", "setup", "steer_returned":
		return "requeue_interrupted_turn"
	default: // inference, compacting, or an unknown running phase
		return "hold_for_retry_or_skip_after_interruption"
	}
}

// turnStartedWork reports whether a turn sent a model request or started a
// tool (unknown counts as started).
func (e *Engine) turnStartedWork(ctx context.Context, turnID string) bool {
	events, err := e.store.ListTurnEvents(ctx, turnID)
	if err != nil {
		return true
	}
	for _, ev := range events {
		if ev.Type == "inference.started" || ev.Type == "tool.started" {
			return true
		}
	}
	return false
}

// InterruptedFailureKind marks turns held because gi stopped mid-turn.
const InterruptedFailureKind = "recovery_interrupted"

func (e *Engine) recoverInterruptedTurn(ctx context.Context, claim store.ActiveTurnClaim) error {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if handled, err := e.recoverStagedTerminal(opCtx, claim); handled || err != nil {
		return err
	}
	if handled, err := e.recoverProviderRetry(opCtx, claim); handled || err != nil {
		return err
	}
	if handled, err := e.recoverGenerationReady(opCtx, claim); handled || err != nil {
		return err
	}
	disposition := recoveryDispositionForClaim(claim)
	status := claim.Status
	phase := claim.Phase
	markFinished := false
	recoveryPhase := claim.Phase
	if record, getErr := e.store.GetTurn(opCtx, claim.TurnID); getErr == nil && record.Metadata["operation"] == "manual_compaction" && status != "completed" && status != "failed" && status != "cancelled" && status != "aborted" {
		recoveryPhase = "cancelling"
		disposition = "abort_interrupted_manual_compaction"
	}
	if claim.Phase == "compacting" && disposition == "hold_for_retry_or_skip_after_interruption" && !e.turnStartedWork(opCtx, claim.TurnID) {
		// Compacting before the turn's first request: nothing was sent and
		// no tool ran, so replaying it is safe.
		disposition, recoveryPhase = "requeue_after_compaction_checkpoint", "queued"
	}
	if disposition == "hold_for_retry_or_skip_after_interruption" {
		recoveryPhase = "waiting_on_tools" // held like a tool-phase interruption
	}
	switch recoveryPhase {
	case "waiting_on_tools":
		status = "failed"
		phase = "held_for_retry_or_skip"
		markFinished = true
	case "cancelling":
		status = "aborted"
		phase = "aborted"
		markFinished = true
	case "completed", "failed", "aborted", "cancelled", "held_for_retry_or_skip":
		// Held and terminal turns are never replayed by stale-claim recovery.
	default:
		status = "queued"
		phase = "queued"
	}

	// Apply recoverable resets without replaying interrupted tool execution.
	for i := 0; i < 10; i++ {
		applied, err := e.store.ApplyPendingContextReset(ctx, claim.SessionID, claim.TurnID)
		if err != nil {
			return fmt.Errorf("recover pending reset: %w", err)
		}
		if !applied {
			break
		}
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
		switch disposition {
		case "hold_for_retry_or_skip_after_tool_checkpoint":
			if err := e.store.MarkTurnFailureWithFallbackErr(e.backgroundContext(), nil, claim.TurnID, claim.SessionID, "recovery_interrupted_tool_phase", "review", "Recovered stale turn that was interrupted while waiting on tool results"); err != nil {
				return err
			}
		case "hold_for_retry_or_skip_after_interruption":
			if err := e.store.MarkTurnFailureWithFallbackErr(e.backgroundContext(), nil, claim.TurnID, claim.SessionID, InterruptedFailureKind, "review", "Interrupted when gi stopped; held instead of resending"); err != nil {
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
	if coordCtx == nil {
		return false, context.Canceled
	}
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
	payload := map[string]any{"intent": in.Intent, "model": in.Model, "kind": "steering", "active_turn_id": activeTurnID}
	if in.ParentTurnID != "" {
		payload["parent_turn_id"] = in.ParentTurnID
	}
	for k, v := range in.Metadata {
		if k == "selected_thinking_level" || k == "selected_thinking_model" {
			continue
		}
		payload[k] = v
	}
	// A deferred steering continuation keeps the request's validated choice;
	// steering consumed by the current turn cannot alter that turn's options.
	selectedSession, err := e.store.GetSession(opCtx, sessionID)
	if err != nil {
		return nil, err
	}
	if level := e.admissionThinking(selectedSession, in.Model); level != "" {
		payload["selected_thinking_level"] = level
		payload["selected_thinking_model"] = in.Model
	}
	media := steeringMediaFromMetadata(in.Metadata)
	queueMode := internalx.StringValue(in.Metadata["steering_mode"], "one-at-a-time")
	steeringRole := normalizeSteeringRole(internalx.StringValue(in.Metadata["ingress_role"], "user"))
	if _, err := e.store.EnqueueActiveSteering(opCtx, sessionID, activeTurnID, steeringRole, in.Prompt, payload, media, queueMode); err != nil {
		if errors.Is(err, store.ErrQueueConflict) {
			return nil, err
		}
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
		for _, key := range []string{"intent", "model", "parent_turn_id", "source_session_id", "source_agent_id", "target_agent_id", "route_mode", "route_matched_by", "selected_thinking_level", "selected_thinking_model"} {
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
	return e.continueSession(ctx, sessionID)
}

func (e *Engine) continueSession(ctx context.Context, sessionID string) (bool, error) {
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

func (r *sessionRunner) dequeueSteeringMessages(ctx context.Context, sessionID, turnID string) ([]store.SteeringMessage, error) {
	msgs, err := r.store.DequeueSteeringForTurn(ctx, sessionID, turnID)
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
	persisted := 0
	for _, msg := range msgs {
		role := persistedSteeringChatRole(msg.Role)
		payload := map[string]any{"kind": "chat", "intent": internalx.StringValue(msg.Payload["intent"], "prompt"), "turn_id": turnID, "steering": true, "steering_role": normalizeSteeringRole(msg.Role)}
		for k, v := range msg.Payload {
			payload[k] = v
		}
		if len(msg.Media) > 0 {
			payload["media"] = append([]string(nil), msg.Media...)
		}
		// Identity comes from the consuming checkpoint, never queued metadata.
		payload["turn_id"] = turnID
		if msg.SourceQueueID != "" {
			if err := r.store.PersistBoundSteering(ctx, sessionID, turnID, msg, payload); err != nil {
				log.Printf("persist bound steering: %v", err)
				continue
			}
		} else {
			logutil.WarnIfErr("add steering message", r.store.AddMessage(ctx, store.NowID("msg"), sessionID, role, msg.Content, payload))
		}
		persisted++
		totalContentLen += len(msg.Content)
	}
	if persisted == 0 {
		return 0
	}
	logutil.WarnIfErr("append steering.injected event", r.store.AppendTurnEvent(ctx, turnID, sessionID, "steering.injected", map[string]any{
		"phase":             "steering",
		"checkpoint":        true,
		"count":             persisted,
		"total_content_len": totalContentLen,
	}))
	r.engine.broadcast(sessionID, map[string]any{"type": "steering_injected", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "count": persisted, "media_count": steeringMediaCount(msgs)})
	return persisted
}

func (r *sessionRunner) injectSteeringMessages(ctx context.Context, sessionID, turnID string, convCtx *goai.Context, msgs []store.SteeringMessage) int {
	if len(msgs) == 0 {
		return 0
	}
	injected := 0
	for _, msg := range msgs {
		if r.persistSteeringMessages(ctx, sessionID, turnID, []store.SteeringMessage{msg}) != 1 {
			continue
		}
		injected++
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
			convCtx.Messages = append(convCtx.Messages, r.userMessageWithProviderSafeMedia(ctx, sessionID, content, msg.Payload))
		}
	}
	return injected
}

func (r *sessionRunner) skipRemainingToolCalls(ctx context.Context, sessionID, turnID string, convCtx *goai.Context, toolCalls []goai.ToolCall, start int) error {
	for _, call := range toolCalls[start:] {
		if err := r.settleSkippedTool(ctx, sessionID, turnID, "", 0, convCtx, call, "queued user steering message", skippedDueToQueuedUserMessage, ""); err != nil {
			return err
		}
	}
	return nil
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
	trimmed := strings.TrimSpace(name)
	switch strings.ToLower(trimmed) {
	case "before_llm", HookBeforeProviderRequest:
		return HookBeforeProviderRequest
	case "after_llm", HookAfterProviderResponse:
		return HookAfterProviderResponse
	case "before_tool", HookToolCall:
		return HookToolCall
	case "after_tool", HookToolResult:
		return HookToolResult
	case HookBeforeAgentStart:
		return HookBeforeAgentStart
	case HookAgentStart:
		return HookAgentStart
	case HookAgentEnd:
		return HookAgentEnd
	case HookTurnStart:
		return HookTurnStart
	case HookTurnEnd:
		return HookTurnEnd
	case HookTurnState:
		return HookTurnState
	case HookSessionState:
		return HookSessionState
	case HookContext:
		return HookContext
	case HookMessageStart:
		return HookMessageStart
	case HookMessageUpdate:
		return HookMessageUpdate
	case HookMessageEnd:
		return HookMessageEnd
	case HookToolExecutionStart:
		return HookToolExecutionStart
	case HookApproveTool:
		return HookApproveTool
	case HookToolExecutionUpdate:
		return HookToolExecutionUpdate
	case HookToolExecutionEnd:
		return HookToolExecutionEnd
	case HookModelSelect:
		return HookModelSelect
	case HookUserBash:
		return HookUserBash
	default:
		return trimmed
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
	e.hookMu.Lock()
	defer e.hookMu.Unlock()
	if e.hooksClosed {
		return nil, context.Canceled
	}
	return e.hooks.Register(name, source, handler)
}

func (e *Engine) ClearHooks() {
	e.hookMu.Lock()
	e.hooks.Clear()
	processes := e.hookProcesses
	e.hookProcesses = nil
	e.hookMu.Unlock()
	for process := range processes {
		process.close()
	}
}

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
	hookCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMS)*time.Millisecond)
	defer cancel()
	if err := hookCtx.Err(); err != nil {
		return HookResponse{}, int(time.Since(startedAt).Milliseconds()), err
	}
	group, err := e.admitHookHost(req)
	if err != nil {
		return HookResponse{}, int(time.Since(startedAt).Milliseconds()), err
	}
	detached, err := detachHookEnvelope(req)
	if err != nil {
		e.finishHookHost(group)
		return HookResponse{}, int(time.Since(startedAt).Milliseconds()), HookExecutionError{Name: req.Name, Source: item.source, Kind: "invalid_input", Trace: req.Trace, Cause: err}
	}
	type hookResult struct {
		resp HookResponse
		err  error
	}
	resultCh := make(chan hookResult, 1)
	go func() {
		defer e.finishHookHost(group)
		defer func() {
			if recovered := recover(); recovered != nil {
				resultCh <- hookResult{err: HookExecutionError{Name: req.Name, Source: item.source, Kind: "panic", Trace: req.Trace, TimeoutMS: timeoutMS, Cause: fmt.Errorf("%v", recovered)}}
			}
		}()
		resp, err := item.handler(hookCtx, detached)
		if err != nil {
			kind := "handler_error"
			if errors.Is(err, context.DeadlineExceeded) && ctx.Err() == nil {
				kind = "timeout"
			}
			resultCh <- hookResult{err: HookExecutionError{Name: req.Name, Source: item.source, Kind: kind, Trace: req.Trace, TimeoutMS: timeoutMS, Cause: err}}
			return
		}
		if hookCtx.Err() != nil {
			return
		} // the expired invocation has no response authority
		resp, err = detachHookEnvelope(resp)
		if err != nil {
			resultCh <- hookResult{err: HookExecutionError{Name: req.Name, Source: item.source, Kind: "invalid_response", Trace: req.Trace, Cause: err}}
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
		if hookCtx.Err() != nil {
			if ctx.Err() != nil {
				return HookResponse{}, int(time.Since(startedAt).Milliseconds()), ctx.Err()
			}
			return HookResponse{}, int(time.Since(startedAt).Milliseconds()), HookExecutionError{Name: req.Name, Source: item.source, Kind: "timeout", Trace: req.Trace, TimeoutMS: timeoutMS, Cause: hookCtx.Err()}
		}
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

	mu          sync.Mutex
	started     bool
	closed      bool
	ioHosts     sync.WaitGroup
	stopContext func()
	stderr      bytes.Buffer
	nextID      int
	stderrGen   uint64
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	stdout      io.ReadCloser
	enc         *json.Encoder
	dec         *json.Decoder
	stderrMu    sync.Mutex
	stderrPipe  io.ReadCloser
}

func (e *Engine) registerProcessHook(workspaceRoot string, spec scripting.EventHookSpec) (func(), error) {
	mounted := &mountedProcessHook{workspaceRoot: workspaceRoot, spec: spec}
	e.hookMu.Lock()
	if e.hooksClosed {
		e.hookMu.Unlock()
		return nil, context.Canceled
	}
	unregister, err := e.hooks.Register(spec.Name, tools.FirstNonEmpty(spec.Source, "process"), mounted.invoke)
	if err != nil {
		e.hookMu.Unlock()
		return nil, err
	}
	if e.hookProcesses == nil {
		e.hookProcesses = make(map[*mountedProcessHook]struct{})
	}
	e.hookProcesses[mounted] = struct{}{}
	e.hookMu.Unlock()
	return func() {
		unregister()
		mounted.close()
		e.hookMu.Lock()
		delete(e.hookProcesses, mounted)
		e.hookMu.Unlock()
	}, nil
}

func (m *mountedProcessHook) close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	m.closeLocked()
}

func (m *mountedProcessHook) invoke(ctx context.Context, req HookRequest) (HookResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return HookResponse{}, context.Canceled
	}
	if err := ctx.Err(); err != nil {
		return HookResponse{}, err
	}
	if err := m.ensureStartedLocked(ctx, req); err != nil {
		return HookResponse{}, err
	}
	defer m.watchContextLocked(ctx)()
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
	configureHookProcess(cmd)
	cmd.Dir = resolveProcessHookDir(m.workspaceRoot, m.spec.CWD)
	cmd.Env = appendProcessHookEnv(os.Environ(), m.spec, req, m.workspaceRoot)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("process hook stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return fmt.Errorf("process hook stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return fmt.Errorf("process hook stderr: %w", err)
	}
	m.stderrMu.Lock()
	m.stderr.Reset()
	m.stderrGen++
	stderrGen := m.stderrGen
	m.stderrMu.Unlock()
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		_ = stderr.Close()
		return fmt.Errorf("start process hook: %w", err)
	}
	m.stderrPipe = stderr
	m.ioHosts.Add(1)
	go func(gen uint64) {
		defer m.ioHosts.Done()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stderr)
		m.stderrMu.Lock()
		defer m.stderrMu.Unlock()
		if gen != m.stderrGen {
			return
		}
		_, _ = m.stderr.Write(buf.Bytes())
	}(stderrGen)
	m.cmd = cmd
	m.stdin = stdin
	m.stdout = stdout
	m.enc = json.NewEncoder(stdin)
	m.enc.SetEscapeHTML(false)
	m.dec = json.NewDecoder(stdout)
	m.started = true
	defer m.watchContextLocked(ctx)()
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
	m.ioHosts.Add(1)
	go func(decoder *json.Decoder) {
		defer m.ioHosts.Done()
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

// Cancellation must also interrupt request writes, not just response reads.
// Join the watcher before Wait releases the PID, so it cannot signal a reused
// process identity. All mutable mount state remains protected by m.mu.
func (m *mountedProcessHook) watchContextLocked(ctx context.Context) func() {
	cmd, stdin, stdout, stderr := m.cmd, m.stdin, m.stdout, m.stderrPipe
	done := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		defer close(done)
		killHookProcess(cmd)
		_ = stdin.Close()
		_ = stdout.Close()
		_ = stderr.Close()
	})
	var once sync.Once
	join := func() {
		once.Do(func() {
			if !stop() {
				<-done
			}
		})
	}
	m.stopContext = join
	return join
}
func (m *mountedProcessHook) closeLocked() {
	if m.stopContext != nil {
		m.stopContext()
		m.stopContext = nil
	}
	m.stderrMu.Lock()
	m.stderrGen++
	m.stderrMu.Unlock()
	if m.stdin != nil {
		_ = m.stdin.Close()
	}
	if m.stdout != nil {
		_ = m.stdout.Close()
	}
	if m.stderrPipe != nil {
		_ = m.stderrPipe.Close()
	}
	if m.cmd != nil {
		killHookProcess(m.cmd)
		_ = m.cmd.Wait()
	}
	m.ioHosts.Wait()
	m.stderrPipe = nil
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
			e.PublishTopicEvent(topics.Envelope{Topic: topicName, SessionID: sessionValue, AgentID: strings.TrimSpace(agentID), Source: source, Type: strings.TrimSpace(typ), Sequence: topics.SequenceFromValue(envelope["sequence"]), Payload: payload})
			return nil
		},
		func(ctx context.Context, sessionID string, pattern string, opts scripting.TopicSubscribeOptions) (<-chan topics.Envelope, func(), error) {
			if e == nil || e.Topics() == nil {
				return nil, nil, fmt.Errorf("subscribe topic: topic bus is not available")
			}
			subOpts := topics.SubscribeOptions{Buffer: opts.Buffer, SessionID: strings.TrimSpace(opts.SessionID), AgentID: strings.TrimSpace(opts.AgentID), AfterSequence: opts.AfterSequence}
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
				return e.registerProcessHook(e.runtimeCfg.WorkspaceRoot, hook)
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
		func(ctx context.Context, sessionID string, spec scripting.CommandSpec) error {
			cmdSpec := ExtensionCommandSpec{Name: spec.Name, Description: spec.Description, Usage: spec.Usage, Source: tools.FirstNonEmpty(spec.Source, spec.Path, "script"), Engine: tools.FirstNonEmpty(spec.Engine, "js")}
			script := spec.Script
			if strings.TrimSpace(script) == "" {
				return fmt.Errorf("register command %s: script handler is required", spec.Name)
			}
			_, err := e.RegisterExtensionCommand(cmdSpec, func(ctx context.Context, cmd ExtensionCommandContext) (ExtensionCommandResult, error) {
				payload := map[string]any{"name": cmd.Name, "args": cmd.Args, "argv": cmd.Argv, "session_id": cmd.SessionID, "agent_id": cmd.AgentID}
				var commandScript string
				if cmdSpec.Engine == "joker" {
					commandScript = tools.ScriptWithPayload("joker", "command", payload, script)
				} else {
					b, _ := json.Marshal(payload)
					commandScript = fmt.Sprintf("const command = %s; const __handler = %s; JSON.stringify(__handler(command));", string(b), script)
				}
				out := scriptTool.Execute(ctx, tools.ScriptInput{Engine: cmdSpec.Engine, Path: spec.Path, SessionID: cmd.SessionID, Script: commandScript})
				if out.Error != "" {
					return ExtensionCommandResult{Type: "error", Error: out.Error}, errors.New(out.Error)
				}
				return extensionCommandResultFromScript(out.Result)
			})
			return err
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
	must(tools.MessagesTool())
	must(tools.DashboardWidgetTool())
	must(tools.PlanTool())
	must(tools.RegisteredTool{
		Name:        "read",
		Description: tools.ReadToolDescription,
		Parameters:  json.RawMessage(tools.ReadToolParameters),
		Source:      "builtin",
		Kind:        "read-only",
		Weight:      "lightweight",
		Activation:  "default",
		Executor: func(ctx context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
			return tools.ExecuteReadTool(ctx, rt, call)
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
			return tools.ExecuteWrite(ctx, e.runtimeCfg, rt.Store, call)
		},
	})
	must(tools.RegisteredTool{
		Name:        "edit",
		Description: tools.EditToolDescription,
		Parameters:  json.RawMessage(tools.EditToolParameters),
		Source:      "builtin",
		Kind:        "mutating",
		Weight:      "lightweight",
		Activation:  "default",
		Executor: func(ctx context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
			return tools.ExecuteEditTool(ctx, e.runtimeCfg, rt, call)
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
			var db *sql.DB
			if rt.Store != nil {
				db = rt.Store.DB()
			}
			prepare := shellenv.Preparer(db, e.runtimeCfg.ShellPath)
			return tools.ExecuteShellPrepared(ctx, rt.WorkspaceRoot, call, rt.OnOutput, prepare)
		},
	})
}

const repeatedToolFailureLimit = store.RepeatedToolFailureLimit

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
func (r *sessionRunner) runAgentLoop(ctx context.Context, s *store.Store, turnID, sessionID, model, agentID string, initialSteering []store.SteeringMessage, continuations ...*store.ProviderRetryCheckpoint) {
	var retry *store.ProviderRetryCheckpoint
	if len(continuations) > 0 {
		retry = continuations[0]
	}
	r.runAgentLoopFromCheckpoint(ctx, s, turnID, sessionID, model, agentID, initialSteering, retry, nil)
}

func (r *sessionRunner) runAgentLoopFromCheckpoint(ctx context.Context, s *store.Store, turnID, sessionID, model, agentID string, initialSteering []store.SteeringMessage, retry *store.ProviderRetryCheckpoint, generation *store.GenerationReadyCheckpoint) {
	maxIter := r.engine.runtimeCfg.MaxIterations
	if maxIter <= 0 {
		maxIter = 64
	}

	providerState := &providerPreparation{}
	var convCtx *goai.Context
	var err error
	firstIteration := 1
	resumed := retry != nil || generation != nil
	lifecycleStarted := resumed // the checkpoint follows this turn's original agent start
	defer func() {
		if !lifecycleStarted || r.checkpointClosing(turnID) {
			return
		}
		agentEndReason := "unsettled"
		if rec, err := s.GetTurn(r.engine.backgroundContext(), turnID); err == nil && isTerminalSubTurnStatus(rec.Status) {
			agentEndReason = rec.Status
		}
		_, _ = r.engine.emitHook(r.engine.backgroundContext(), HookRequest{Name: HookAgentEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Payload: map[string]any{"reason": agentEndReason}})
	}()
	if generation != nil {
		providerState, convCtx, err = r.restoreGenerationPreparation(ctx, s, turnID, sessionID, agentID, generation)
		if err == nil {
			model = providerState.model
		}
		firstIteration = generation.Iteration
		maxIter = min(maxIter, generation.MaxIterations)
	} else if retry != nil {
		checkpoint := retry
		providerState, err = r.restoreProviderPreparation(checkpoint.Wait)
		if err == nil {
			err = r.waitRecoveredProviderRetry(ctx, s, sessionID, turnID, checkpoint)
		}
		if err == nil {
			firstIteration = checkpoint.Wait.Iteration
			convCtx = &goai.Context{}
		}
	} else {
		convCtx, err = r.assembleAgentContext(ctx, s, turnID, sessionID, model, agentID, providerState)
		lifecycleStarted = err == nil
	}
	if err != nil {
		if ctx.Err() != nil || isCancellationError(err) {
			r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled", "")
		} else {
			r.finishTurn(s, turnID, sessionID, agentID, model, "failed", fmt.Sprintf("Turn context error: %v", err), "context_error")
		}
		return
	}
	r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "title": "Thinking…", "status": "running", "turn_id": turnID})

	lastToolFailureSig := providerState.lastToolFailureSig
	repeatedToolFailureCount := providerState.repeatedToolFailureCount
	pendingSteering := append([]store.SteeringMessage(nil), initialSteering...)

	for iter := firstIteration; iter <= maxIter; iter++ {
		if ctx.Err() != nil {
			r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled", "")
			return
		}

		var prepareErr error
		if !resumed || iter != firstIteration {
			pendingSteering, prepareErr = r.prepareAgentIteration(ctx, sessionID, turnID, model, agentID, iter, convCtx, pendingSteering, providerState)
		}
		if prepareErr != nil {
			if ctx.Err() != nil || isCancellationError(prepareErr) || errors.Is(prepareErr, store.ErrCompactionInactive) {
				r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled", "")
			} else {
				r.finishTurn(s, turnID, sessionID, agentID, model, "failed", fmt.Sprintf("Context preparation error: %v", prepareErr), "compaction_error")
			}
			return
		}
		providerState.lastToolFailureSig, providerState.repeatedToolFailureCount = lastToolFailureSig, repeatedToolFailureCount
		// A sleeping/woken retry already has its dedicated frontier. Ordinary
		// generations commit only after preparation and all prior effects settle.
		if retry == nil || iter != firstIteration {
			checkpoint := &store.GenerationReadyCheckpoint{Version: 1, Iteration: iter, MaxIterations: maxIter, Model: model, Lifecycle: providerState.lifecyclePolicy(), LastToolFailureSignature: lastToolFailureSig, RepeatedToolFailures: repeatedToolFailureCount}
			if err := s.SaveGenerationReady(ctx, sessionID, turnID, checkpoint); err != nil {
				r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Persist generation continuation: "+err.Error(), "persistence_error")
				return
			}
		}
		result, inferErr := r.runProviderIteration(ctx, s, turnID, sessionID, model, agentID, iter, maxIter, convCtx, providerState)
		if providerState.initialized {
			model = providerState.model
		}
		iterLabel := fmt.Sprintf("iter=%d/%d", iter, maxIter)
		if inferErr != nil {
			if result != nil && result.OccurrenceID != "" {
				if err := r.settleProviderFailure(ctx, s, turnID, sessionID, model, iter, result, inferErr); err != nil {
					r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Provider receipt settlement: "+err.Error(), "persistence_error")
					return
				}
			}
			if ctx.Err() != nil || isCancellationError(inferErr) {
				r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled", "")
				return
			}
			var abortErr hookAbortError
			if errors.As(inferErr, &abortErr) {
				logutil.WarnIfErr("append inference.aborted event", s.AppendTurnEvent(ctx, turnID, sessionID, "inference.aborted", map[string]any{"phase": "inference", "checkpoint": true, "error": abortErr.Error(), "iteration": iter, "hard_abort": abortErr.hard}))
				r.finishTurn(s, turnID, sessionID, agentID, model, "aborted", abortErr.Error(), "hook_abort")
				return
			}
			log.Printf("inference [%s] error: %v", iterLabel, inferErr)
			logutil.WarnIfErr("append inference.failed event", s.AppendTurnEvent(ctx, turnID, sessionID, "inference.failed", map[string]any{"phase": "inference", "checkpoint": true, "error": inferErr.Error(), "iteration": iter}))
			r.finishTurn(s, turnID, sessionID, agentID, model, "failed", fmt.Sprintf("Inference error: %v", inferErr), "provider_error")
			return
		}
		if result == nil || result.Message == nil {
			if ctx.Err() != nil {
				r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled", "")
				return
			}
			if settleErr := r.settleProviderFailure(ctx, s, turnID, sessionID, model, iter, result, fmt.Errorf("provider returned no result")); settleErr != nil {
				r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Settle empty provider receipt: "+settleErr.Error(), "persistence_error")
				return
			}
			log.Printf("inference [%s]: nil result", iterLabel)
			r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Inference returned no result", "provider_invalid_result")
			return
		}

		if result.Usage == nil {
			result.Usage = result.Message.Usage
		}
		if terminalErr := inference.ValidateAssistantResult(result); terminalErr != nil {
			if settleErr := r.settleProviderFailure(ctx, s, turnID, sessionID, model, iter, result, terminalErr); settleErr != nil {
				r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Settle invalid provider receipt: "+settleErr.Error(), "persistence_error")
				return
			}
			r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Invalid provider terminal: "+terminalErr.Error(), "provider_invalid_result")
			return
		}
		assistantMsg := result.Message
		textContent := goai.GetTextContent(assistantMsg)
		toolCalls := goai.GetToolCalls(assistantMsg)
		log.Printf("inference [%s]: stop=%q toolCalls=%d text=%d", iterLabel, assistantMsg.StopReason, len(toolCalls), len(textContent))
		goai.AppendAssistantMessage(convCtx, assistantMsg)

		needsToolExecution := len(toolCalls) > 0
		if !needsToolExecution {
			// A direct reply is a completed assistant message even when steering
			// keeps this run active for another model request. Persist it before
			// rebuilding the next context from SQLite; otherwise it vanishes from
			// both history and the visible timeline.
			msgID := store.ProviderMessageID(result.OccurrenceID)
			messagePayload := map[string]any{"kind": "chat", "source": "inference", "model": model, "turn_id": turnID, "agent_id": agentID, "iterations": iter}
			messagePayload["provider_message"] = result.Message
			r.addResponseRecord(ctx, s, turnID, model, result, messagePayload)
			r.addRecoveryMarker(ctx, sessionID, turnID, messagePayload)
			if err := r.settleProviderResponse(ctx, s, turnID, sessionID, model, iter, result, msgID, textContent, messagePayload, "reply_boundary"); err != nil {
				if errors.Is(err, context.Canceled) {
					r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled", "")
					return
				}
				r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Persist assistant response: "+err.Error(), "persistence_error")
				return
			}
			r.broadcastPost(sessionID, turnID, msgID, textContent, agentID, messagePayload["content_blocks"])
			_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookMessageEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"chars": len(textContent)}})
			if steerMsgs, err := r.dequeueSteeringMessages(ctx, sessionID, turnID); err != nil {
				log.Printf("steering dequeue error after direct response: %v", err)
			} else if len(steerMsgs) > 0 {
				pendingSteering = append(pendingSteering, steerMsgs...)
				continue
			}
			log.Printf("inference [%s]: final response (%d chars, %d iterations)", iterLabel, len(textContent), iter)
			// Iteration-end hooks can still steer/cancel the proposed completion.
			// Durable completion is published only by successful settlement below.
			_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookTurnEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"status": "completed", "settled": false}})
			r.finishTurnOK(s, turnID, sessionID, agentID, model, iter)
			return
		}

		log.Printf("inference [%s]: %d tool call(s)", iterLabel, len(toolCalls))
		if strings.TrimSpace(textContent) != "" && strings.TrimSpace(result.Text) == "" {
			r.engine.broadcast(sessionID, map[string]any{
				"type": "agent_thought_delta", "chat_jid": "gi:" + sessionID,
				"turn_id": turnID, "iteration": iter, "delta": textContent,
			})
		}
		toolCallSummary := textContent
		for _, tc := range toolCalls {
			if toolCallSummary != "" {
				toolCallSummary += "\n"
			}
			toolCallSummary += fmt.Sprintf("[tool_call: %s]", tc.Name)
		}
		// Full calls (id, name, arguments) make the history exportable to Pi's
		// session format; the summary content stays for display and context.
		recordedCalls := make([]map[string]any, 0, len(toolCalls))
		for _, tc := range toolCalls {
			recordedCalls = append(recordedCalls, map[string]any{"id": tc.ID, "name": tc.Name, "arguments": tc.Arguments, "argumentsError": tc.ArgumentsError})
		}
		callsPayload := map[string]any{
			"kind": "tool_calls", "source": "inference", "model": model,
			"turn_id": turnID, "agent_id": agentID, "display_text": strings.TrimSpace(textContent),
			"tool_calls": recordedCalls,
		}
		callsPayload["provider_message"] = result.Message
		r.addResponseRecord(ctx, s, turnID, model, result, callsPayload)
		if err := r.settleProviderResponse(ctx, s, turnID, sessionID, model, iter, result, store.ProviderMessageID(result.OccurrenceID), toolCallSummary, callsPayload, "tools"); err != nil {
			if errors.Is(err, context.Canceled) {
				r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled", "")
				return
			}
			r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Persist assistant tool calls: "+err.Error(), "persistence_error")
			return
		}

		outcome := r.executeToolCallsPhase(ctx, s, turnID, sessionID, model, agentID, iter, convCtx, toolCalls, pendingSteering, lastToolFailureSig, repeatedToolFailureCount, nil)
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
	r.finishTurnWithPayload(s, turnID, sessionID, agentID, model, "completed", fmt.Sprintf("Reached maximum iteration limit (%d). The task may be incomplete.", maxIter), "", map[string]any{"iterations": maxIter, "completion_kind": "max_iterations"})
}

// executeTool dispatches a single tool call and returns the text result.
func (r *sessionRunner) executeTool(ctx context.Context, call goai.ToolCall, sessionID, turnID string, output ...func(string) error) (string, error) {
	var onOutput func(string) error
	if len(output) > 0 {
		onOutput = output[0]
	}
	return r.executeToolWithImages(ctx, call, sessionID, turnID, onOutput, nil)
}

// executeToolWithImages also collects images the tool attaches (nil: tools
// describe images in text instead).
func (r *sessionRunner) executeToolWithImages(ctx context.Context, call goai.ToolCall, sessionID, turnID string, onOutput func(string) error, images *toolExtras, prepared ...tools.RegisteredTool) (string, error) {
	if call.ArgumentsError != "" {
		return "", fmt.Errorf("Invalid tool arguments: %s", call.ArgumentsError)
	}
	if strings.TrimSpace(turnID) != "" {
		turnRec, err := r.store.GetTurn(ctx, turnID)
		if err != nil {
			return "", err
		}
		if !toolAllowedByMetadata(turnRec.Metadata, call.Name) && !r.engine.sessionLoadedTool(ctx, sessionID, call.Name) {
			return "", fmt.Errorf("tool not allowed in this turn: %s", call.Name)
		}
	}
	tool, ok := r.engine.tools.GetRegistered(call.Name)
	if len(prepared) > 0 {
		tool = prepared[0]
		ok = tool.Executor != nil
	}
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", call.Name)
	}
	occurrenceID := ""
	if images != nil {
		occurrenceID = images.occurrenceID
	}
	return tool.Executor(ctx, tools.ToolRuntime{
		ToolOccurrenceID: occurrenceID,
		Store:            r.store,
		SessionID:        sessionID,
		TurnID:           turnID,
		WorkspaceRoot:    r.engine.runtimeCfg.WorkspaceRoot,
		OnOutput:         onOutput,
		AttachImage:      images.attachFunc(),
		AddTools:         images.addToolsFunc(),
		ToolCallID:       call.ID,
		SetDetails:       images.setDetailsFunc(),
		AddUsage:         images.addUsageFunc(),
		UnknownUsage:     images.unknownUsageFunc(),
		PublishMessage:   r.engine.publishWidgetMessage,
		PublishPlan:      r.engine.PublishPlanChanged,
	}, call)
}

// Latest request occupancy is independent of cumulative turn billing.
func responseContextMeasurement(model string, usage *goai.Usage, iteration int) map[string]any {
	if usage == nil || usage.Input < 0 || usage.CacheRead < 0 || usage.CacheWrite < 0 {
		return nil
	}
	if usage.Input == 0 && usage.CacheRead == 0 && usage.CacheWrite == 0 && usage.Output == 0 && usage.TotalTokens == 0 {
		return nil
	}
	return map[string]any{"tokens": usage.Input + usage.CacheRead + usage.CacheWrite, "input": usage.Input, "cache_read": usage.CacheRead, "cache_write": usage.CacheWrite, "model": model, "iteration": iteration, "phase": "inference"}
}

func (r *sessionRunner) settleProviderResponse(ctx context.Context, s *store.Store, turnID, sessionID, model string, iteration int, result *inference.StreamResult, messageID, content string, payload map[string]any, decision string) error {
	measurement := responseContextMeasurement(model, result.Usage, iteration)
	settled, err := s.SettleProviderAttempt(context.WithoutCancel(ctx), turnID, sessionID, result.OccurrenceID, store.ProviderReceipt{MessageID: messageID, Content: content, Payload: payload, Usage: result.Usage, Unknown: providerUsageUnknown(result.Usage), Measurement: measurement, Decision: decision, Iteration: iteration})
	if err != nil {
		return err
	}
	if settled.Applied {
		if measurement != nil {
			r.engine.PublishRuntimeTurnEvent("context_measured", sessionID, turnID, "", "running", "inference", measurement)
		}
		r.engine.PublishRuntimeTurnEvent("turn_usage", sessionID, turnID, "", "running", "inference", map[string]any{"usage": settled.Accounting.Map()})
	}
	if settled.Decision == "cancelled" {
		return context.Canceled
	}
	return nil
}

func providerUsageUnknown(u *goai.Usage) bool {
	return u == nil || (u.Input == 0 && u.Output == 0 && u.CacheRead == 0 && u.CacheWrite == 0 && u.TotalTokens == 0 && u.CacheWrite1h == 0 && u.Reasoning == 0)
}

func (r *sessionRunner) settleProviderFailure(ctx context.Context, s *store.Store, turnID, sessionID, model string, iteration int, result *inference.StreamResult, failure error, retryWait ...*store.ProviderRetryWait) error {
	if result == nil || result.OccurrenceID == "" {
		return fmt.Errorf("missing provider attempt identity")
	}
	bgCtx := context.WithoutCancel(ctx)
	message := result.Message
	partial := result.Incomplete
	if message != nil && (message.StopReason == goai.StopReasonPending || message.StopReason == goai.StopReasonAborted || message.Deferred != nil) {
		partial = true
	}
	// Keep the committed prefix separately from the provider's exact terminal
	// message: an ErrorEvent need not repeat its content or received usage.
	record, err := s.ProviderPartial(bgCtx, turnID, sessionID, result.OccurrenceID)
	if err != nil {
		return err
	}
	var checkpoint any
	if record != nil {
		checkpoint = record.Payload["partial_checkpoint"]
		if message == nil {
			message = record.Message
			partial = true
		}
	}
	usage := result.Usage
	if usage == nil && message != nil {
		usage = message.Usage
	}
	if store.ValidateReceiptUsage(usage) != nil {
		usage = nil
	}
	usageFromPartial := false
	if providerUsageUnknown(usage) && record != nil && store.ValidateReceiptUsage(record.Message.Usage) == nil {
		usage = record.Message.Usage
		usageFromPartial = usage != nil
	}
	messageID, content := "", ""
	payload := map[string]any{"kind": "chat", "source": "inference", "model": model, "turn_id": turnID, "partial": partial || result.Incomplete, "receipt_error": failure.Error()}
	if checkpoint != nil {
		payload["partial_checkpoint"] = checkpoint
	}
	if record != nil {
		payload["provider_partial"] = record.Payload["provider_message"]
	}
	if message != nil {
		encoded, encodeErr := json.Marshal(message)
		if encodeErr != nil {
			payload["provider_message_omissions"] = encodeErr.Error()
			if message.Usage != nil {
				payload["provider_usage_diagnostic"] = fmt.Sprintf("%+v", *message.Usage)
			}
			message, _ = boundedProviderPartial(message)
			partial = true
			payload["partial"] = true
			payload["partial_checkpoint"] = map[string]any{"truncated": true, "string_bytes_limit": providerPartialStringBytes, "blocks_limit": 64}
			var err error
			encoded, err = json.Marshal(message)
			if err != nil {
				return err
			}
		}
		messageID = store.ProviderMessageID(result.OccurrenceID)
		content = goai.GetTextContent(message)
		savedResult := *result
		savedResult.Message = message
		savedResult.Usage = usage
		payload["provider_message"] = json.RawMessage(encoded)
		r.addResponseRecord(bgCtx, s, turnID, model, &savedResult, payload)
	}
	if record != nil {
		// Derive history text from structured provider snapshots, never a
		// mutable UI row. Keep differing fragments visible without rewriting
		// either signed provider message or adopting them into model context.
		prefix := goai.GetTextContent(record.Message)
		switch {
		case content == "" || strings.HasPrefix(prefix, content):
			content = prefix
			payload["partial"] = true
		case prefix != "" && !strings.HasPrefix(content, prefix):
			content = prefix + "\n\nProvider response:\n\n" + content
			payload["partial"] = true
		}
		if _, hasThinking := payload["thinking_blocks"]; !hasThinking {
			if thinking, ok := record.Payload["thinking_blocks"]; ok {
				payload["thinking_blocks"] = thinking
			}
		}
	}
	var next *store.ProviderRetryWait
	if len(retryWait) > 0 {
		next = retryWait[0]
	}
	settled, err := s.SettleProviderAttempt(bgCtx, turnID, sessionID, result.OccurrenceID, store.ProviderReceipt{MessageID: messageID, Content: content, Payload: payload, Usage: usage, Unknown: partial || result.Incomplete || usageFromPartial || providerUsageUnknown(usage), Measurement: responseContextMeasurement(model, usage, iteration), Decision: "error_boundary", Error: failure.Error(), Iteration: iteration, RetryWait: next})
	if err == nil && next != nil && settled.Decision != "cancelled" {
		if settled.RetryWait == nil {
			return store.ErrQueueConflict
		}
		*next = *settled.RetryWait
	}
	if err == nil && settled.Applied && message != nil {
		r.broadcastPost(sessionID, turnID, settled.MessageID, content, "", payload["content_blocks"])
	}
	if err == nil && settled.Decision == "cancelled" {
		return context.Canceled
	}
	return err
}

// addResponseRecord keeps what a response was, as Pi's session entries do:
// its usage and stop reason, its thinking blocks and the thinking level it
// was asked for (session export, docs/internal/session-import.md).
func (r *sessionRunner) addResponseRecord(ctx context.Context, s *store.Store, turnID, model string, result *inference.StreamResult, payload map[string]any) {
	if result == nil {
		return
	}
	if result.TerminalReason != "" {
		payload["provider_terminal_reason"] = result.TerminalReason
	}
	if result.Usage != nil {
		if raw, err := json.Marshal(result.Usage); err == nil {
			var usage map[string]any
			if json.Unmarshal(raw, &usage) == nil {
				payload["usage"] = usage
			}
		}
	}
	if msg := result.Message; msg != nil {
		if msg.StopReason != "" {
			payload["stop_reason"] = string(msg.StopReason)
		}
		var thinking []any
		for _, b := range msg.Content {
			if b.Type != "thinking" {
				continue
			}
			block := map[string]any{"type": "thinking", "thinking": b.Thinking}
			if b.ThinkingSignature != "" {
				block["thinkingSignature"] = b.ThinkingSignature
			}
			if b.Redacted {
				block["redacted"] = true
			}
			thinking = append(thinking, block)
		}
		if len(thinking) > 0 {
			payload["thinking_blocks"] = thinking
		}
	}
	if result.RequestThinking != nil {
		if *result.RequestThinking != "" {
			payload["thinking_level"] = *result.RequestThinking
		}
		return
	}
	if rec, err := s.GetTurn(ctx, turnID); err == nil && rec.Metadata["selected_thinking_model"] == model {
		if level, _ := rec.Metadata["selected_thinking_level"].(string); level != "" {
			payload["thinking_level"] = level
		}
	}
}

// broadcastPost sends a new_post SSE event for the final assistant message.
func (r *sessionRunner) addRecoveryMarker(ctx context.Context, sessionID, turnID string, payload map[string]any) {
	marker, err := r.engine.store.RecoveryMarker(ctx, sessionID, turnID)
	logutil.WarnIfErr("read final response recovery marker", err)
	if err == nil && marker != nil {
		payload["content_blocks"] = []any{marker}
	}
}

func (r *sessionRunner) broadcastPost(sessionID, turnID, msgID, content, agentID string, blocks ...any) {
	var contentBlocks any
	if len(blocks) > 0 {
		contentBlocks = blocks[0]
	}
	parentID, err := r.engine.store.MessageReplyTo(context.Background(), sessionID, msgID)
	logutil.WarnIfErr("read live response reply identity", err)
	r.engine.broadcast(sessionID, map[string]any{
		"type": "new_post", "turn_id": turnID, "id": msgID, "chat_jid": "gi:" + sessionID,
		"content": content, "timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"sender": "agent", "is_bot_message": true,
		"data": map[string]any{"type": "agent_response", "content": content, "agent_id": agentID, "content_blocks": contentBlocks, "thread_id": parentID},
	})
	r.engine.broadcast(sessionID, map[string]any{"type": "agent_response", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "id": msgID})
}

func (r *sessionRunner) broadcastSystemPost(sessionID, turnID, msgID, content string) {
	r.broadcastSystemPostWithBlocks(sessionID, turnID, msgID, content, nil)
}

func (r *sessionRunner) broadcastSystemPostWithBlocks(sessionID, turnID, msgID, content string, blocks []any) {
	data := map[string]any{"type": "system_message", "content": content, "turn_id": turnID}
	if len(blocks) > 0 {
		data["content_blocks"] = blocks
	}
	r.engine.broadcast(sessionID, map[string]any{
		"type": "new_post", "id": msgID, "chat_jid": "gi:" + sessionID,
		"content": content, "timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"sender": "system",
		"data":   data,
	})
}

// turnOutcomeMarker is Piclaw's turn_outcome_marker content block: the chip
// after a terminal post's timestamp naming how the turn ended.
func turnOutcomeMarker(status, failureKind string) map[string]any {
	switch status {
	case "failed":
		return map[string]any{"type": "turn_outcome_marker", "kind": tools.FirstNonEmpty(failureKind, "error"), "severity": "error", "label": "failed"}
	case "cancelled":
		return map[string]any{"type": "turn_outcome_marker", "kind": "cancelled", "severity": "warning", "label": "cancelled"}
	}
	return nil
}

// PostSystemMessage stores a system message in a session's timeline and
// shows it to connected clients. System messages are not model context.
func (e *Engine) PostSystemMessage(ctx context.Context, sessionID, content string, payload map[string]any) (string, error) {
	msgID := store.NowID("msg")
	if err := e.store.AddMessage(ctx, msgID, sessionID, "system", content, payload); err != nil {
		return "", err
	}
	e.broadcast(sessionID, map[string]any{
		"type": "new_post", "id": msgID, "chat_jid": "gi:" + sessionID,
		"content": content, "timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"sender": "system",
		"data":   map[string]any{"type": "system_message", "content": content},
	})
	return msgID, nil
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
	if r.checkpointClosing(turnID) {
		return
	}
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
	r.setTerminalPending(turnID, true)
	r.engine.joinTurnHooks(turnID, false)
	phase := terminalPhaseForStatus(status)
	msgPayload := map[string]any{"kind": "chat", "source": "system", "turn_id": turnID, "agent_id": agentID}
	if marker := turnOutcomeMarker(status, failureKind); marker != nil {
		msgPayload["content_blocks"] = []any{marker}
	}
	decision := r.ownedTerminalDecision(turnID, store.TurnSettlement{Status: status, Phase: phase, MessageID: store.NowID("msg"), Message: systemMsg, MessagePayload: msgPayload, FailureKind: failureKind, Payload: payload})
	var settled store.TurnSettlement
	for {
		childrenStatus, pending, err := s.StageTurnTerminal(bgCtx, turnID, sessionID, decision)
		if err != nil {
			log.Printf("stage terminal decision for %s: %v", turnID, err)
			return
		}
		if !pending {
			r.setTerminalPending(turnID, false)
			return
		}
		childFailureKind := failureKind
		if childrenStatus == "cancelled" {
			childFailureKind = ""
		}
		r.propagateChildSubTurnCancellation(bgCtx, turnID, childrenStatus, childFailureKind)
		if err = r.joinRequiredChildren(bgCtx, turnID, childrenStatus, childFailureKind); err != nil {
			if errors.Is(err, store.ErrQueueConflict) {
				continue
			}
			log.Printf("join child settlement for %s: %v", turnID, err)
			return
		}
		r.appendFinalSteeringCheckpoint(s, turnID, sessionID)
		decision.ChildrenStatus = childrenStatus
		settled, err = s.SettleRunningTurn(bgCtx, turnID, sessionID, decision)
		if err != nil {
			if errors.Is(err, store.ErrQueueConflict) {
				latest, readErr := s.GetTurn(bgCtx, turnID)
				if readErr == nil && latest.Status == "cancelling" && childrenStatus != "cancelled" {
					continue
				}
			}
			log.Printf("turn settlement conflict for %s: %v", turnID, err)
			return
		}
		break
	}
	r.setTerminalPending(turnID, false)
	if !settled.Applied {
		return
	}
	status, phase, systemMsg, failureKind, payload = settled.Status, settled.Phase, settled.Message, settled.FailureKind, settled.Payload
	if systemMsg != "" {
		blocks, _ := settled.MessagePayload["content_blocks"].([]any)
		r.broadcastSystemPostWithBlocks(sessionID, turnID, settled.MessageID, systemMsg, blocks)
	}
	if settled.Accounting.Iterations > 0 || settled.Accounting.Unknown > 0 || settled.Accounting.Usage != (goai.Usage{}) {
		r.engine.PublishRuntimeTurnEvent("turn_usage", sessionID, turnID, agentID, status, phase, map[string]any{"usage": settled.Accounting.Map()})
	}
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
	r.publishSubTurnLifecycle(bgCtx, turnID, status)
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
	r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "title": "", "status": "idle"})
}

func (r *sessionRunner) publishSubTurnLifecycle(ctx context.Context, childTurnID, status string) {
	if strings.TrimSpace(childTurnID) == "" {
		return
	}
	opCtx := store.CoordinationContext(ctx, r.engine.backgroundContext())
	if opCtx == nil {
		return
	}
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
		orphaned, ending := false, false
		if parentTurn, err := r.store.GetTurn(opCtx, sub.ParentTurnID); err == nil {
			ending = parentTurn.Phase == "settling" || parentTurn.Status == "cancelling"
			orphaned = isTerminalSubTurnStatus(parentTurn.Status) || ending
		}
		eventType := "subturn_result_ready"
		if orphaned {
			eventType = "subturn_orphaned"
			orphanReason := "parent_turn_completed_before_async_result_consumption"
			if ending {
				orphanReason = "parent_terminal_decision_staged"
			}
			logutil.WarnIfErr("update async subturn orphan metadata", r.store.UpdateSubTurnMetadataByChild(opCtx, sub.ChildTurnID, map[string]any{
				"orphaned":      true,
				"orphaned_at":   time.Now().UTC().Format(time.RFC3339Nano),
				"orphan_reason": orphanReason,
			}))
			if sub.ParentSessionID != sub.ChildSessionID {
				content := fmt.Sprintf("Async sub-turn %s finished with status %s after parent turn %s had already ended.", sub.ChildTurnID, status, sub.ParentTurnID)
				if ending {
					content = fmt.Sprintf("Async sub-turn %s finished with status %s while parent turn %s was ending.", sub.ChildTurnID, status, sub.ParentTurnID)
				}
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
	case status == "failed":
		return subTurnCancellationDirective{enabled: true, cancelCritical: false, reason: "parent_failed"}
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
	visited := map[string]bool{parentTurnID: true}
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

func (r *sessionRunner) buildAgentContext(ctx context.Context, s *store.Store, turnID, sessionID string, cfg config.RuntimeConfig, p *providerPreparation) (*goai.Context, error) {
	snapshot, err := s.ContextSnapshot(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("load session messages: %w", err)
	}
	turnRec, err := s.GetTurn(ctx, turnID)
	if err != nil {
		return nil, fmt.Errorf("load turn metadata: %w", err)
	}
	turnMetadata := turnRec.Metadata
	convCtx := &goai.Context{
		Tools:    r.engine.toolDefsForMetadata(turnMetadata),
		Messages: r.projectContextSnapshot(ctx, sessionID, snapshot),
	}
	// Tools loaded by tool_search earlier in the session stay declared.
	r.engine.declareLoadedTools(convCtx, r.engine.sessionLoadedTools(ctx, sessionID))
	r.engine.applyCodemodeLoadout(ctx, convCtx, sessionID, turnMetadata)
	ids, offset := snapshotMessageIDs(snapshot)
	plan := r.engine.planPromptSections(ctx, sessionID, r.engine.promptSectionsWithConfig(convCtx.Tools, cfg), ids, offset, convCtx.Messages, turnID)
	p.plan = plan
	convCtx.SystemPrompt = plan.systemPrompt()
	return convCtx, nil
}

func (r *sessionRunner) assembleAgentContext(ctx context.Context, s *store.Store, turnID, sessionID, model, agentID string, preparations ...*providerPreparation) (*goai.Context, error) {
	p := &providerPreparation{}
	if len(preparations) > 0 && preparations[0] != nil {
		p = preparations[0]
	}
	convCtx, err := r.buildAgentContext(ctx, s, turnID, sessionID, r.engine.runtimeCfg, p)
	if err != nil {
		return nil, err
	}
	if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookBeforeAgentStart, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, SystemPrompt: convCtx.SystemPrompt, Messages: convCtx.Messages, Tools: convCtx.Tools}); err != nil {
		log.Printf("hook before_agent_start error: %v", err)
		p.startCaptured = true // no overrides were adopted; do not replay the hook
	} else {
		if len(preparations) > 0 && preparations[0] != nil {
			if err := r.captureAgentStartPreparation(preparations[0], resp); err != nil {
				return nil, fmt.Errorf("capture agent start preparation: %w", err)
			}
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
		if strings.TrimSpace(resp.Message) != "" {
			convCtx.Messages = append([]goai.Message{goai.UserMessage(resp.Message)}, convCtx.Messages...)
		}
	}
	_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookAgentStart, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model})
	return convCtx, nil
}

func (r *sessionRunner) prepareAgentIteration(ctx context.Context, sessionID, turnID, model, agentID string, iter int, convCtx *goai.Context, pendingSteering []store.SteeringMessage, preparations ...*providerPreparation) ([]store.SteeringMessage, error) {
	// A previous boundary may already have dequeued one message. Do not poll
	// again in this iteration: one-at-a-time must reach the next model request
	// before another queued message is admitted (as in Pi's agent loop).
	if len(pendingSteering) == 0 {
		if steerMsgs, err := r.dequeueSteeringMessages(ctx, sessionID, turnID); err != nil {
			log.Printf("steering dequeue error: %v", err)
		} else if len(steerMsgs) > 0 {
			pendingSteering = steerMsgs
		}
	}
	if len(pendingSteering) > 0 {
		r.injectSteeringMessages(ctx, sessionID, turnID, convCtx, pendingSteering)
		pendingSteering = nil
	}
	_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookTurnStart, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter})
	r.applyGenerationContextHook(ctx, sessionID, turnID, model, agentID, iter, convCtx)
	// Snapshot before the asynchronous before-compact hook; completion compares
	// the exact history and checkpoint version inside its write transaction.
	snapshot, snapshotErr := r.store.ContextSnapshot(ctx, sessionID)
	if snapshotErr != nil {
		return pendingSteering, snapshotErr
	}
	err := r.compactSnapshot(ctx, sessionID, turnID, model, agentID, convCtx, snapshot, false, preparations...)
	return pendingSteering, err
}

func (r *sessionRunner) runProviderIteration(ctx context.Context, s *store.Store, turnID, sessionID, model, agentID string, iter, maxIter int, convCtx *goai.Context, preparations ...*providerPreparation) (*inference.StreamResult, error) {
	prepared := &providerPreparation{}
	if len(preparations) > 0 && preparations[0] != nil {
		prepared = preparations[0]
	}
	if err := prepared.initialize(ctx, s, turnID, model, r.engine.runtimeCfg); err != nil {
		return nil, err
	}
	policy := prepared.cfg.Retry.Policy()
	var previousFailure error
	attempt := prepared.retryAttempt
	prepared.retryAttempt = 0
	var retryWait *store.ProviderRetryWait
	var retryOccurrenceID string
	result, inferErr := retryProviderRequestDynamicFrom(ctx, func() config.ProviderRetryPolicy { return policy }, attempt, func() (*inference.StreamResult, error, bool) {
		defer func() { attempt++ }()
		retryWait = nil
		if attempt > 0 {
			if err := r.refreshProviderPreparation(ctx, s, turnID, sessionID, agentID, iter, attempt, prepared, convCtx); err != nil {
				return nil, errors.Join(previousFailure, fmt.Errorf("retry preparation: %w", err)), true
			}
			policy = prepared.cfg.Retry.Policy()
		}
		model, thinking := prepared.model, prepared.thinking
		responseObserved := false
		requestCtx, err := r.prepareProviderRequestContext(ctx, sessionID, turnID, model, agentID, iter, convCtx, prepared.plan)
		if err != nil {
			return nil, err, true
		}
		if err = ctx.Err(); err != nil {
			return nil, err, true
		}
		log.Printf("inference [iter=%d/%d]: calling %s", iter, maxIter, model)
		progress, hookFailed := false, false
		requestID := store.NowID("inference")
		intent, intentErr := sanitizedProviderIntent(map[string]any{"model": model, "context": requestCtx, "thinking": thinking, "transport": prepared.cfg.Transport})
		if intentErr != nil {
			return nil, intentErr, true
		}
		if intentErr = s.AdmitTurnEffect(ctx, sessionID, turnID, "running", "inference.started", map[string]any{"phase": "inference", "model": model, "iteration": iter, "checkpoint": true, "request_id": requestID, "intent": intent}); intentErr != nil {
			return nil, intentErr, true
		}
		r.emitTurnStateHook(ctx, sessionID, turnID, agentID, model, "running", "running", map[string]any{"reason": "provider_iteration", "iteration": iter})
		r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "title": fmt.Sprintf("Thinking… (%d)", iter), "status": "running", "turn_id": turnID})
		partials := &providerPartialRecorder{store: s, ctx: context.WithoutCancel(ctx), turnID: turnID, sessionID: sessionID, occurrenceID: requestID, model: model}
		result, inferErr := streamWithToolsWithHooks(ctx, model, requestCtx, func(ev map[string]any) {
			ev["chat_jid"] = "gi:" + sessionID
			ev["turn_id"] = turnID
			ev["iteration"] = iter
			switch ev["type"] {
			case "text_delta", "thinking_delta", "tool_call_start", "tool_call_end":
				progress = true
			}
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
				// The turn owner decides retry versus terminal failure. Do not
				// publish a terminal error while a transient retry is pending.
			}
		}, &inference.StreamHooks{
			Thinking: thinking,
			OnProgress: func(message *goai.Message) error {
				if message != nil && (len(message.Content) > 0 || !providerUsageUnknown(message.Usage)) {
					progress = true
				}
				return partials.Observe(message)
			},
			Transport: goai.Transport(prepared.cfg.Transport),
			OnPayload: func(payload any, modelDef *goai.Model) (any, error) {
				hookPayload := map[string]any{"ok": true, "request": payload, "stage": "payload"}
				if modelDef != nil {
					hookPayload["provider"] = string(modelDef.Provider)
					hookPayload["api"] = string(modelDef.Api)
					hookPayload["model_id"] = modelDef.ID
				}
				resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookBeforeProviderRequest, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, SystemPrompt: requestCtx.SystemPrompt, Messages: requestCtx.Messages, Tools: requestCtx.Tools, Payload: hookPayload})
				if err != nil {
					hookFailed = true
					return nil, err
				}
				if abortErr := hookAbortFromResponse(resp, "aborted before provider request send by hook"); abortErr != nil {
					hookFailed = true
					return nil, abortErr
				}
				if replacement, ok := providerRequestReplacementFromHook(resp); ok {
					payload = replacement
				}
				prepared, err := sanitizedProviderIntent(payload)
				if err == nil {
					err = s.AdmitTurnEffect(ctx, sessionID, turnID, "running", "inference.request_prepared", map[string]any{"phase": "inference", "iteration": iter, "request_id": requestID, "intent": prepared})
				}
				if err != nil {
					hookFailed = true
					return nil, err
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
		if result == nil {
			result = &inference.StreamResult{}
		}
		result.OccurrenceID = requestID
		actualThinking := thinking
		result.RequestThinking = &actualThinking
		previousFailure = inferErr
		retryOccurrenceID = requestID
		// Some adapters return a failed partial without emitting deltas. That is
		// still observed progress, not permission to repeat the generation.
		if result.Incomplete || result.Partial != nil || (result.Message != nil && (len(result.Message.Content) > 0 || result.Message.Deferred != nil || result.Message.StopReason == goai.StopReasonPending)) {
			progress = true
		}
		if err := partials.Seal(result.Partial); err != nil {
			inferErr = fmt.Errorf("persist provider partial: %w", err)
			hookFailed = true
		}
		if inferErr != nil {
			if class, eligible := transientProviderFailure(inferErr); ctx.Err() == nil && eligible && !progress && !hookFailed {
				current, policyErr := r.engine.runtimeCfg.ForGeneration()
				if policyErr != nil {
					inferErr = errors.Join(inferErr, fmt.Errorf("retry policy reload: %w", policyErr))
					hookFailed = true
				} else {
					policy = current.Retry.Policy()
					if policy.Enabled && attempt < policy.MaxRetries {
						retryWait = newProviderRetryWait(policy, model, iter, attempt+1, providerRetryDelay(policy, attempt+1), class)
						retryWait.Lifecycle = prepared.lifecyclePolicy()
						retryWait.LastToolFailureSignature, retryWait.RepeatedToolFailures = prepared.lastToolFailureSig, prepared.repeatedToolFailureCount
					}
				}
			}
			if err := r.settleProviderFailure(ctx, s, turnID, sessionID, model, iter, result, inferErr, retryWait); err != nil {
				return result, fmt.Errorf("provider receipt settlement: %w", err), true
			}
		}
		if !responseObserved {
			if _, err := r.engine.emitHook(ctx, HookRequest{Name: HookAfterProviderResponse, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"ok": inferErr == nil}}); err != nil {
				log.Printf("hook after_provider_response error: %v", err)
			}
		}
		return result, inferErr, progress || hookFailed
	}, func(attempt int, delay time.Duration, class string) (time.Time, error) {
		if retryWait == nil || retryWait.Attempt != attempt {
			return time.Time{}, store.ErrQueueConflict
		}
		r.publishProviderRetryWait(turnID, sessionID, retryWait)
		return retryWait.RetryAt, nil
	}, func() error {
		if err := s.ResumeProviderRetry(ctx, sessionID, turnID, retryOccurrenceID); err != nil {
			return err
		}
		r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "status": "running", "title": "Retrying provider request…", "phase": "inference"})
		return nil
	})
	return result, inferErr
}

func (r *sessionRunner) executeToolCallsPhase(ctx context.Context, s *store.Store, turnID, sessionID, model, agentID string, iter int, convCtx *goai.Context, toolCalls []goai.ToolCall, pendingSteering []store.SteeringMessage, lastToolFailureSig string, repeatedToolFailureCount int, totalUsage *goai.Usage) toolExecutionOutcome {
	outcome := toolExecutionOutcome{
		pendingSteering:          pendingSteering,
		lastToolFailureSig:       lastToolFailureSig,
		repeatedToolFailureCount: repeatedToolFailureCount,
	}
	executableCalls := 0
	for _, call := range toolCalls {
		if call.ArgumentsError == "" {
			executableCalls++
		}
	}
	if executableCalls > 0 {
		_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookToolExecutionStart, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"count": executableCalls}})
		defer func() {
			_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookToolExecutionEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"count": executableCalls}})
		}()
	}
	responsePosted := false
	for _, call := range toolCalls {
		if ctx.Err() != nil {
			r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled during tool execution", "")
			outcome.terminated = true
			return outcome
		}

		if call.ArgumentsError != "" {
			if err := r.rejectInvalidToolArguments(ctx, s, sessionID, turnID, agentID, iter, convCtx, call); err != nil {
				r.finishTurn(s, turnID, sessionID, agentID, model, "failed", err.Error(), "persistence_error")
				outcome.terminated = true
				return outcome
			}
			continue
		}
		originalCall, originalErr := json.Marshal(call)
		if originalErr != nil {
			r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Pin tool provenance: "+originalErr.Error(), "persistence_error")
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
				if call.Arguments == nil {
					call.Arguments = map[string]any{}
				}
				if call.ArgumentsError != "" {
					if err := r.rejectInvalidToolArguments(ctx, s, sessionID, turnID, agentID, iter, convCtx, call); err != nil {
						r.finishTurn(s, turnID, sessionID, agentID, model, "failed", err.Error(), "persistence_error")
						outcome.terminated = true
						return outcome
					}
					continue
				}
			}
			if injectedResult, ok := directToolResultFromHook(resp); ok {
				displayResult, outputDiagnostics := boundToolOutput(injectedResult)
				// Hook results have no execution boundary or measured duration.
				hookOccurrence := store.NowID("tool")
				if err := s.SettleToolResult(ctx, turnID, sessionID, "tool.finished",
					map[string]any{"phase": "tool", "tool": call.Name, "checkpoint": true, "tool_call_id": call.ID, "occurrence_id": hookOccurrence, "output_length": len(injectedResult), "source": "hook", "hook_phase": "tool_call"},
					store.NowID("msg"), displayResult,
					withOutputDiagnostics(map[string]any{"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": false, "turn_id": turnID, "occurrence_id": hookOccurrence, "source": "hook", "hook_phase": "tool_call"}, outputDiagnostics)); err != nil {
					r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Hook result settlement failed: "+err.Error(), "persistence_error")
					outcome.terminated = true
					return outcome
				}
				r.engine.PublishRuntimeHookDecisionEvent("hook_respond", HookRequest{Name: HookToolCall, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "tool_call", "source": "hook", "output_length": len(injectedResult)})
				r.engine.broadcast(sessionID, map[string]any{"type": "tool_finished", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "output_length": len(injectedResult)})
				r.engine.PublishRuntimeToolEvent("tool_finished", sessionID, turnID, agentID, call.Name, call.ID, iter, nil, map[string]any{"phase": "tool", "occurrence_id": hookOccurrence, "output_length": len(injectedResult), "source": "hook", "hook_phase": "tool_call"})
				goai.AppendToolResult(convCtx, call.ID, call.Name, displayResult, false)
				outcome.lastToolFailureSig = ""
				outcome.repeatedToolFailureCount = 0
				continue
			}
			if resp.Block {
				reason := internalx.StringValue(resp.Reason, "tool call blocked")
				if err := r.settleSkippedTool(ctx, sessionID, turnID, agentID, iter, convCtx, call, reason, "Error: blocked by hook: "+reason, "tool_call"); err != nil {
					r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Settle denied tool: "+err.Error(), "persistence_error")
					outcome.terminated = true
					return outcome
				}
				r.engine.PublishRuntimeHookDecisionEvent("hook_deny", HookRequest{Name: HookToolCall, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "tool_call", "reason": reason})
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
				if err := r.settleSkippedTool(ctx, sessionID, turnID, agentID, iter, convCtx, call, reason, "Error: blocked by hook: "+reason, "approve_tool"); err != nil {
					r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Settle denied tool: "+err.Error(), "persistence_error")
					outcome.terminated = true
					return outcome
				}
				r.engine.PublishRuntimeHookDecisionEvent("hook_deny", HookRequest{Name: HookApproveTool, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "approve_tool", "reason": reason})
				continue
			}
			if resp.ToolCall != nil {
				r.engine.PublishRuntimeHookDecisionEvent("hook_modify", HookRequest{Name: HookApproveTool, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "approve_tool", "modified_tool": resp.ToolCall.Name, "modified_tool_call_id": resp.ToolCall.ID})
				call = *resp.ToolCall
				if call.Arguments == nil {
					call.Arguments = map[string]any{}
				}
				if call.ArgumentsError != "" {
					if err := r.rejectInvalidToolArguments(ctx, s, sessionID, turnID, agentID, iter, convCtx, call); err != nil {
						r.finishTurn(s, turnID, sessionID, agentID, model, "failed", err.Error(), "persistence_error")
						outcome.terminated = true
						return outcome
					}
					continue
				}
			}
		}

		toolOccurrenceID := store.NowID("tool")
		registered, _ := r.engine.tools.GetRegistered(call.Name)
		if err := s.AdmitTurnEffect(ctx, sessionID, turnID, "waiting_on_tools", "tool.started", map[string]any{
			"phase": "tool", "tool": call.Name, "checkpoint": true, "tool_call_id": call.ID, "occurrence_id": toolOccurrenceID, "iteration": iter, "preview": store.ToolActivityPreview(call.Arguments),
			"original_call": json.RawMessage(originalCall), "prepared_call": call, "schema": registered.Parameters, "replay_safe": registered.ReplaySafe,
		}); err != nil {
			r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Admit tool intent: "+err.Error(), "persistence_error")
			outcome.terminated = true
			return outcome
		}
		r.emitTurnStateHook(ctx, sessionID, turnID, agentID, model, "running", "waiting_on_tools", map[string]any{"reason": "tool_execution", "tool": call.Name, "iteration": iter})
		r.engine.PublishRuntimeToolEvent("tool_started", sessionID, turnID, agentID, call.Name, call.ID, iter, nil, map[string]any{"occurrence_id": toolOccurrenceID, "phase": "tool", "arguments": call.Arguments})
		r.engine.broadcast(sessionID, map[string]any{"type": "tool_activity_changed", "chat_jid": "gi:" + sessionID, "turn_id": turnID})

		r.engine.broadcast(sessionID, map[string]any{
			"type": "agent_status", "chat_jid": "gi:" + sessionID,
			"title": fmt.Sprintf("Running: %s", call.Name), "status": "running", "turn_id": turnID,
		})

		reportOutput, closeOutput := r.toolOutputReporter(ctx, turnID, sessionID, call.ID, toolOccurrenceID)
		defer closeOutput()
		images := &toolExtras{occurrenceID: toolOccurrenceID}
		executionStart := time.Now()
		toolResult, toolErr := r.executeToolWithImages(ctx, call, sessionID, turnID, func(text string) error { return reportOutput(text, false) }, images, registered)
		durationMS := time.Since(executionStart).Milliseconds()
		if totalUsage != nil {
			addUsage(totalUsage, images.usage) // model calls made by the tool (codemode models.*)
		}
		if outputErr := reportOutput(toolResult, false); outputErr != nil {
			r.persistStoppedTool(ctx, s, sessionID, turnID, call, toolOccurrenceID, "aborted", durationMS, images)
			r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Persist tool output: "+outputErr.Error(), "persistence_error")
			outcome.terminated = true
			return outcome
		}
		if toolErr != nil {
			if ctx.Err() != nil || isCancellationError(toolErr) {
				_ = reportOutput(toolResult, true)
				r.persistStoppedTool(ctx, s, sessionID, turnID, call, toolOccurrenceID, "cancelled", durationMS, images)
				r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled during tool execution", "")
				outcome.terminated = true
				return outcome
			}
			log.Printf("tool [%s] error: %v", call.Name, firstLine(toolErr.Error()))
			errText := toolErr.Error() // Pi's createErrorToolResult: the bare message
			var resultErr *tools.ResultError
			if errors.As(toolErr, &resultErr) {
				errText = resultErr.Text // the complete result, e.g. shell output + exit status (Pi)
			}
			errText, outputDiagnostics := boundToolOutput(errText)
			if outputErr := reportOutput(errText, true); outputErr != nil {
				r.persistStoppedTool(ctx, s, sessionID, turnID, call, toolOccurrenceID, "aborted", durationMS, images)
				r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Persist tool output: "+outputErr.Error(), "persistence_error")
				outcome.terminated = true
				return outcome
			}
			if err := s.SettleToolResult(ctx, turnID, sessionID, "tool.failed", map[string]any{
				"phase": "tool", "tool": call.Name, "checkpoint": true,
				"duration_ms": durationMS, "tool_call_id": call.ID, "occurrence_id": toolOccurrenceID, "error": toolErr.Error(),
			}, store.NowID("msg"), errText+images.transcriptSuffix(), images.withReplay(withOutputDiagnostics(map[string]any{
				"duration_ms": durationMS, "occurrence_id": toolOccurrenceID, "kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": true, "turn_id": turnID,
			}, outputDiagnostics), errText, true)); err != nil {
				r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Tool result settlement failed: "+err.Error(), "persistence_error")
				outcome.terminated = true
				return outcome
			}
			r.engine.broadcast(sessionID, map[string]any{"type": "tool_failed", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "error": toolErr.Error()})
			r.engine.PublishRuntimeToolEvent("tool_failed", sessionID, turnID, agentID, call.Name, call.ID, iter, toolErr, images.withDetails(map[string]any{"occurrence_id": toolOccurrenceID, "duration_ms": durationMS, "phase": "tool", "arguments": call.Arguments, "output": errText}))

			r.engine.broadcast(sessionID, map[string]any{"type": "tool_activity_changed", "chat_jid": "gi:" + sessionID, "turn_id": turnID})
			appendToolResultWithImages(convCtx, call, errText, true, images)

			outcome.lastToolFailureSig, outcome.repeatedToolFailureCount = nextRepeatedToolFailureCount(outcome.lastToolFailureSig, outcome.repeatedToolFailureCount, call, toolErr)
			if outcome.repeatedToolFailureCount >= repeatedToolFailureLimit {
				msg := fmt.Sprintf("Aborting after %d repeated identical tool failures: %v", outcome.repeatedToolFailureCount, toolErr)
				log.Printf("tool [%s] repeated failure guard tripped: %s", call.Name, msg)
				r.finishTurn(s, turnID, sessionID, agentID, model, "failed", msg, "repeated_tool_failure")
				outcome.terminated = true
				return outcome
			}
		} else {
			if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookToolResult, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, ToolCall: &call, ToolResult: toolResult, Payload: map[string]any{"tool": call.Name, "tool_call_id": call.ID, "is_error": false}}); err != nil {
				log.Printf("hook tool_result error: %v", err)
			} else {
				if abortErr := hookAbortFromResponse(resp, fmt.Sprintf("tool %s result aborted by hook", call.Name)); abortErr != nil {
					r.persistStoppedTool(ctx, s, sessionID, turnID, call, toolOccurrenceID, "aborted", durationMS, images)
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
			displayResult, outputDiagnostics := boundToolOutput(toolResult)
			if outputErr := reportOutput(displayResult, true); outputErr != nil {
				r.persistStoppedTool(ctx, s, sessionID, turnID, call, toolOccurrenceID, "aborted", durationMS, images)
				r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Persist tool output: "+outputErr.Error(), "persistence_error")
				outcome.terminated = true
				return outcome
			}
			if err := s.SettleToolResult(ctx, turnID, sessionID, "tool.finished", map[string]any{
				"phase": "tool", "tool": call.Name, "checkpoint": true,
				"duration_ms": durationMS, "tool_call_id": call.ID, "occurrence_id": toolOccurrenceID, "output_length": len(toolResult),
			}, store.NowID("msg"), displayResult+images.transcriptSuffix(), images.withReplay(withOutputDiagnostics(map[string]any{
				"duration_ms": durationMS, "occurrence_id": toolOccurrenceID, "kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": false, "turn_id": turnID,
			}, outputDiagnostics), displayResult, false)); err != nil {
				r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Tool result settlement failed: "+err.Error(), "persistence_error")
				outcome.terminated = true
				return outcome
			}
			r.engine.broadcast(sessionID, map[string]any{"type": "tool_finished", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "output_length": len(toolResult)})
			r.engine.PublishRuntimeToolEvent("tool_finished", sessionID, turnID, agentID, call.Name, call.ID, iter, nil, images.withDetails(map[string]any{"occurrence_id": toolOccurrenceID, "duration_ms": durationMS, "phase": "tool", "arguments": call.Arguments, "output_length": len(toolResult), "output": displayResult}))

			r.engine.broadcast(sessionID, map[string]any{"type": "tool_activity_changed", "chat_jid": "gi:" + sessionID, "turn_id": turnID})
			appendToolResultWithImages(convCtx, call, displayResult, false, images)
			r.engine.declareLoadedTools(convCtx, images.added)

			if tool, ok := r.engine.tools.GetRegistered(call.Name); ok && tool.CompletesTurn {
				responsePosted = true
			}
			outcome.lastToolFailureSig = ""
			outcome.repeatedToolFailureCount = 0
		}
	}
	if responsePosted {
		// Leave steering queued for the normal final checkpoint/handoff; do not
		// dequeue it into a model continuation that this UI-only response replaces.
		r.finishTurnOK(s, turnID, sessionID, agentID, model, iter)
		outcome.terminated = true
		return outcome
	}
	// Pi drains the assistant's tool-call batch before polling steering for the
	// next model turn. A message queued during one tool must not suppress later
	// calls from the same response; the queued row stays recoverable until here.
	if ctx.Err() == nil {
		if reset, err := s.ApplyPendingContextReset(ctx, sessionID, turnID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			r.finishTurn(s, turnID, sessionID, agentID, model, "failed", err.Error(), "persistence_error")
			outcome.terminated = true
			return outcome
		} else if reset {
			r.finishTurnOK(s, turnID, sessionID, agentID, model, iter)
			outcome.terminated = true
			return outcome
		}
		if steerMsgs, err := r.dequeueSteeringMessages(ctx, sessionID, turnID); err != nil {
			log.Printf("steering dequeue after tool batch: %v", err)
		} else {
			outcome.pendingSteering = append(outcome.pendingSteering, steerMsgs...)
		}
	}
	return outcome
}

func (r *sessionRunner) appendFinalSteeringCheckpoint(s *store.Store, turnID, sessionID string) {
	bgCtx := r.engine.backgroundContext()
	// The inbox is bounded. Drain consecutive empty resets here so no reset
	// remains stranded after the active claim is released. A reset with text
	// becomes ordinary follow-up work and stops this loop at the FIFO head.
	for i := 0; i < 10; i++ {
		applied, err := s.ApplyPendingContextReset(bgCtx, sessionID, turnID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("apply final context reset: %v", err)
			return
		}
		if !applied {
			break
		}
	}
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
	if active != nil && active.checkpointClosing {
		if r.current == active {
			r.current = nil
		}
		return // the Close controller retains the fenced reservation until join
	}
	if active != nil && active.terminalPending {
		// Keep the durable claim/checkpoint for recovery; releasing it would
		// let another turn pass unresolved child or settlement ownership.
		if r.current == active {
			r.current = nil
		}
		r.appendCleanupHandoffFailure(ctx, sessionID, active.turnID, "terminal_settlement_pending", fmt.Errorf("terminal decision requires recovery"))
		return
	}
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
	if rawMedia, ok := turnRec.Metadata["media"]; ok {
		userPayload["media"] = rawMedia
	}
	persistedIdlePrompt, err := s.HasIdleQueuePrompt(ctx, sessionID, turnID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(prompt) != "" && len(initialSteering) == 0 && !persistedIdlePrompt {
		if err := s.AddMessage(ctx, store.NowID("msg"), sessionID, "user", prompt, userPayload); err != nil {
			return nil, fmt.Errorf("persist user prompt message: %w", err)
		}
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
	identity, err := s.RequireSessionIdentityRuntime(opCtx, sessionID)
	agentID := ""
	if err == nil {
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
	if run.turn.Metadata["operation"] == "manual_compaction" {
		r.runManualCompaction(ctx, run)
		return
	}
	if run.model != "bootstrap" && run.model != "test-model" && run.model != "" {
		r.runAgentLoop(ctx, s, run.turnID, run.sessionID, run.model, run.agentID, run.initialSteering)
		return
	}
	r.runShellTurn(ctx, s, run)
}

func (e *Engine) admissionThinking(session *store.Session, model string) string {
	if level := inference.CapturedSessionThinking(session, model); level != "" {
		return level
	}
	level, _ := inference.EffectiveThinking(model, strings.TrimSpace(e.runtimeCfg.DefaultThinkingLevel))
	return level
}
