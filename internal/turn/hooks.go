package turn

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rcarmo/gi/internal/logutil"
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
	"github.com/rcarmo/gi/internal/scripting"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

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
	_, recordErr := e.store.RecordHookInvocation(ctx, req.TurnID, req.SessionID, req.Name, req.Name, item.source, action, req, response, hookInvocationErrorText(err), durationMS)
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
		if action := store.StringValue(raw["action"], ""); strings.TrimSpace(action) != "" {
			resp.Action = strings.ToLower(strings.TrimSpace(action))
		}
	}
	switch resp.Action {
	case "continue":
	case "modify":
	case "respond":
		resp.Handled = true
		if strings.TrimSpace(resp.Message) == "" {
			if response := store.StringValue(raw["response"], ""); strings.TrimSpace(response) != "" {
				resp.Message = response
			} else if payload, ok := raw["payload"].(map[string]any); ok {
				resp.Message = store.StringValue(payload["response"], "")
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
