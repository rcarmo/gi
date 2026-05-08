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

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/connectivity"
	"github.com/rcarmo/gi/internal/peering"
	"github.com/rcarmo/gi/internal/routing"
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
	metadata["effective_tools"] = effectiveTools
	if in.ParentTurnID != "" {
		metadata["subturn_tools_restricted"] = subTurnToolsRestricted
	}
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
	if in.ParentTurnID != "" && parentSessionID != "" {
		subturnMetadata := map[string]any{"intent": in.Intent, "model": in.Model, "depth": subTurnDepth, "max_depth": subTurnMaxDepth, "max_concurrency": subTurnMaxConcurrency, "delivery_mode": subTurnDeliveryMode, "subturn_critical": subTurnCritical, "effective_tools": effectiveTools, "subturn_tools_restricted": subTurnToolsRestricted}
		if _, err := e.store.CreateSubTurn(ctx, in.ParentTurnID, parentSessionID, turnID, in.SessionID, subTurnDeliveryMode, subTurnDepth, subturnMetadata); err != nil {
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
	warnStore("sync queue count after submit", e.store.SyncSessionQueueCount(ctx, in.SessionID))
	warnStore("touch session model after submit", e.store.TouchSessionState(ctx, in.SessionID, map[string]any{"model": in.Model}))
	return &SubmitResult{TurnID: turnID, SessionID: in.SessionID, Status: status, Queued: queued}, nil
}

func (e *Engine) SubmitPromptRouted(ctx context.Context, in RunInput) (*SubmitResult, error) {
	resolution, err := e.preparePromptRouteResolution(ctx, in)
	if err != nil {
		return nil, err
	}
	if err := e.resolveRoutedPromptTarget(ctx, resolution); err != nil {
		return nil, err
	}
	if resolution.target.ID != resolution.source.ID {
		return e.submitPeerRoutedPrompt(ctx, resolution.source, resolution.target, resolution.route, resolution.promptBody, in.Intent, in.Model, resolution.created, resolution.directed, in.ParentTurnID)
	}
	in.SessionID = resolution.target.ID
	in.Prompt = resolution.promptBody
	e.applyLocalRouteMetadata(&in, resolution)
	return e.SubmitPrompt(ctx, in)
}

func (e *Engine) SubmitPeerMessage(ctx context.Context, sourceSessionID, targetAgentID, content, intent, model, parentTurnID string) (*SubmitResult, error) {
	resolution, err := e.preparePeerRouteResolution(ctx, sourceSessionID, targetAgentID, content, "peer-message")
	if err != nil {
		return nil, err
	}
	if err := e.resolveRoutedPromptTarget(ctx, resolution); err != nil {
		return nil, err
	}
	return e.submitPeerRoutedPrompt(ctx, resolution.source, resolution.target, resolution.route, content, intent, model, resolution.created, resolution.directed, parentTurnID)
}

func (e *Engine) ResolveOrCreatePeerSession(ctx context.Context, sourceSessionID, targetAgentID string) (*store.Session, bool, error) {
	resolution, err := e.preparePeerRouteResolution(ctx, sourceSessionID, targetAgentID, "", "peer-session")
	if err != nil {
		return nil, false, err
	}
	if err := e.resolveRoutedPromptTarget(ctx, resolution); err != nil {
		return nil, false, err
	}
	return resolution.target, resolution.created, nil
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
		warnStore("sync queue count after queued cancel", e.store.SyncSessionQueueCount(ctx, turnSessionID))
		return e.store.AppendTurnEvent(ctx, turnID, turnSessionID, "turn.cancelled", map[string]any{"phase": "cancel", "checkpoint": true, "queued": true})
	}
	return fmt.Errorf("turn not cancellable")
}

func (e *Engine) runner(sessionID string) *sessionRunner {
	v, _ := e.sessions.LoadOrStore(sessionID, &sessionRunner{store: e.store, engine: e})
	return v.(*sessionRunner)
}

func (e *Engine) launchTurnLocked(ctx context.Context, runner *sessionRunner, sessionID, turnID string) (bool, error) {
	claimToken := turnID
	claimed, err := e.store.ClaimSessionActiveTurn(ctx, sessionID, turnID, "runner", claimToken)
	if err != nil {
		return false, err
	}
	if !claimed {
		return false, nil
	}
	releaseClaim := func() {
		warnStore("release active claim after launch failure", e.store.ReleaseSessionActiveTurn(ctx, sessionID, claimToken))
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
	go runner.runTurn(e.store, sessionID, turnID)
	return true, nil
}

func (r *sessionRunner) runTurn(s *store.Store, sessionID, turnID string) {
	ctx, cancel := context.WithCancel(context.Background())
	claimToken := turnID
	r.mu.Lock()
	r.current = &runningTurn{turnID: turnID, cancel: cancel}
	r.mu.Unlock()
	defer cancel()
	defer func() {
		r.cleanupTurnRun(sessionID, claimToken)
	}()

	run, err := r.setupTurnRun(ctx, s, sessionID, turnID)
	if err != nil {
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
		if _, err := io.Copy(&stderr, stderrPipe); err != nil {
			warnStore("copy shell stderr", err)
		}
	}()
	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()
	select {
	case <-ctx.Done():
		if cmd.Process != nil {
			if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
				warnStore("kill timed out shell process group", err)
			}
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
	if normalizeAgentID(sessionAgentID(source)) == normalizeAgentID(route.AgentID) {
		return source, false, nil
	}
	plan, err := e.prepareRouteSessionPlan(source, route, inbound)
	if err != nil {
		return nil, false, err
	}
	existing, err := e.resolveExistingRouteSession(ctx, plan)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}
	cloned, err := e.cloneRouteSession(ctx, plan)
	if err != nil {
		return nil, false, err
	}
	return cloned, true, nil
}

func (e *Engine) submitPeerRoutedPrompt(ctx context.Context, source, target *store.Session, route routing.ResolvedRoute, content, intent, model string, created, directed bool, parentTurnID string) (*SubmitResult, error) {
	sourceAgentID := sessionAgentID(source)
	routingContent := fmt.Sprintf("↪ routed to @%s: %s", route.AgentID, content)
	routingPayload := map[string]any{"kind": "routing", "target_agent_id": route.AgentID, "target_session_id": target.ID, "source_agent_id": sourceAgentID, "source_session_id": source.ID, "route_matched_by": route.MatchedBy, "clipped": true}
	warnStore("add routing message to source session", e.store.AddMessage(ctx, store.NowID("msg"), source.ID, "system", routingContent, routingPayload))
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
