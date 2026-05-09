package turn

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rcarmo/gi/internal/config"
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
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return HookTrace{ID: fmt.Sprintf("hook_%d_%d", time.Now().UTC().UnixNano(), seq), EmittedAt: now}
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

func (e *Engine) invokeHookHandler(ctx context.Context, req HookRequest, item registeredHook) (HookResponse, error) {
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
		if ctx.Err() != nil {
			return HookResponse{}, ctx.Err()
		}
		return HookResponse{}, HookExecutionError{Name: req.Name, Source: item.source, Kind: "timeout", Trace: req.Trace, TimeoutMS: timeoutMS, Cause: hookCtx.Err()}
	case res := <-resultCh:
		return res.resp, res.err
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
		resp, err := e.invokeHookHandler(ctx, req, item)
		if err != nil {
			if policyErr := e.applyHookFailurePolicy(req, item, err); policyErr != nil {
				return merged, policyErr
			}
			continue
		}
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
