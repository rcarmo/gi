package turn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/routing"
	"github.com/rcarmo/gi/internal/scripting"
	"github.com/rcarmo/gi/internal/topics"
	goai "github.com/rcarmo/go-ai"
)

func TestHookNameAliasesMapToCanonicalPhases(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)

	calledBeforeLLM := 0
	if _, err := e.RegisterHook("before_llm", "test", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		calledBeforeLLM++
		return HookResponse{}, nil
	}); err != nil {
		t.Fatalf("register before_llm hook: %v", err)
	}
	if _, err := e.emitHook(context.Background(), HookRequest{Name: HookBeforeProviderRequest}); err != nil {
		t.Fatalf("emit before provider request: %v", err)
	}
	if calledBeforeLLM != 1 {
		t.Fatalf("expected before_llm alias handler to run once, got %d", calledBeforeLLM)
	}

	calledBeforeTool := 0
	if _, err := e.RegisterHook("before_tool", "test", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		calledBeforeTool++
		return HookResponse{}, nil
	}); err != nil {
		t.Fatalf("register before_tool hook: %v", err)
	}
	if _, err := e.emitHook(context.Background(), HookRequest{Name: HookToolCall}); err != nil {
		t.Fatalf("emit tool call hook: %v", err)
	}
	if calledBeforeTool != 1 {
		t.Fatalf("expected before_tool alias handler to run once, got %d", calledBeforeTool)
	}
}

func TestHookResponseFromScriptAppliesActionSemantics(t *testing.T) {
	deny, err := hookResponseFromScript(`{"action":"deny","reason":"nope"}`)
	if err != nil {
		t.Fatalf("decode deny action: %v", err)
	}
	if !deny.Block || deny.Cancel || deny.Reason != "nope" {
		t.Fatalf("unexpected deny mapping: %#v", deny)
	}

	respond, err := hookResponseFromScript(`{"action":"respond","response":"handled"}`)
	if err != nil {
		t.Fatalf("decode respond action: %v", err)
	}
	if !respond.Handled || respond.Message != "handled" {
		t.Fatalf("unexpected respond mapping: %#v", respond)
	}

	hardAbort, err := hookResponseFromScript(`{"action":"hard_abort"}`)
	if err != nil {
		t.Fatalf("decode hard_abort action: %v", err)
	}
	if !hardAbort.Block || !hardAbort.Cancel {
		t.Fatalf("expected hard_abort to block+cancel: %#v", hardAbort)
	}
	if hardAbort.Payload == nil || hardAbort.Payload["hard_abort"] != true {
		t.Fatalf("expected hard_abort payload marker: %#v", hardAbort)
	}
}

func TestNextHookTraceUsesConsistentTimestampSource(t *testing.T) {
	trace := nextHookTrace()
	if trace.ID == "" || trace.EmittedAt == "" {
		t.Fatalf("expected populated hook trace, got %#v", trace)
	}
	var nanos int64
	if _, err := fmt.Sscanf(trace.ID, "hook_%d_", &nanos); err != nil {
		t.Fatalf("parse hook trace id %q: %v", trace.ID, err)
	}
	emittedAt, err := time.Parse(time.RFC3339Nano, trace.EmittedAt)
	if err != nil {
		t.Fatalf("parse emitted_at %q: %v", trace.EmittedAt, err)
	}
	if emittedAt.UnixNano() != nanos {
		t.Fatalf("expected hook trace id timestamp %d to match emitted_at %d in %#v", nanos, emittedAt.UnixNano(), trace)
	}
}

func TestHookRequestJSONSafePayloadIncludesStructuredFields(t *testing.T) {
	req := HookRequest{
		Name:          HookToolCall,
		SessionID:     "session-json-safe",
		TurnID:        "turn-json-safe",
		AgentID:       "agent",
		Model:         "model",
		Iteration:     2,
		SessionStatus: "running",
		TurnStatus:    "running",
		TurnPhase:     "waiting_on_tools",
		Payload:       map[string]any{"k": "v"},
		Trace:         HookTrace{ID: "hook_trace_1", EmittedAt: "2026-05-09T00:00:00Z"},
		SystemPrompt:  "system",
		Messages:      []goai.Message{goai.UserMessage("hello")},
		Tools:         []goai.Tool{{Name: "read", Description: "Read"}},
		ToolCall:      &goai.ToolCall{Type: "toolCall", ID: "tc1", Name: "read", Arguments: map[string]any{"path": "README.md"}},
		ToolResult:    "ok",
		ToolError:     false,
	}
	payload := hookScriptPayload(req)
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal hook payload: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal hook payload: %v", err)
	}
	if decoded["session_status"] != "running" || decoded["turn_phase"] != "waiting_on_tools" {
		t.Fatalf("missing state fields in hook payload: %#v", decoded)
	}
	if _, ok := decoded["messages"]; !ok {
		t.Fatalf("expected messages in hook payload: %#v", decoded)
	}
	if _, ok := decoded["tools"]; !ok {
		t.Fatalf("expected tools in hook payload: %#v", decoded)
	}
	if _, ok := decoded["tool_call"]; !ok {
		t.Fatalf("expected tool_call in hook payload: %#v", decoded)
	}
	trace, ok := decoded["trace"].(map[string]any)
	if !ok || trace["id"] != "hook_trace_1" {
		t.Fatalf("expected trace metadata in hook payload: %#v", decoded)
	}
}

func TestHookResponseFromScriptDecodesStructuredFields(t *testing.T) {
	resp, err := hookResponseFromScript(`{"action":"modify","system_prompt":"patched","messages":[{"role":"user","content":[{"type":"text","text":"hello"}]}],"tools":[{"name":"read","description":"Read","parameters":{"type":"object"}}],"tool_call":{"id":"tc1","name":"read","arguments":{"path":"README.md"}},"tool_result":"done"}`)
	if err != nil {
		t.Fatalf("decode structured hook response: %v", err)
	}
	if resp.SystemPrompt != "patched" || len(resp.Messages) != 1 || len(resp.Tools) != 1 {
		t.Fatalf("expected structured fields in hook response: %#v", resp)
	}
	if resp.ToolCall == nil || resp.ToolCall.Name != "read" {
		t.Fatalf("expected tool_call decode in hook response: %#v", resp)
	}
	if resp.ToolResult == nil || *resp.ToolResult != "done" {
		t.Fatalf("expected tool_result decode in hook response: %#v", resp)
	}
}

func TestTurnAndSessionStateHooksObserveLifecycle(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_hook_state", "HookState", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_hook_state", "session_hook_state", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	var turnStates []string
	var sessionStates []string
	var sessionPayloads []map[string]any
	if _, err := e.RegisterHook(HookTurnState, "test", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		turnStates = append(turnStates, req.TurnStatus+"/"+req.TurnPhase)
		return HookResponse{}, nil
	}); err != nil {
		t.Fatalf("register turn_state hook: %v", err)
	}
	if _, err := e.RegisterHook(HookSessionState, "test", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		sessionStates = append(sessionStates, req.SessionStatus)
		sessionPayloads = append(sessionPayloads, cloneMap(req.Payload))
		return HookResponse{}, nil
	}); err != nil {
		t.Fatalf("register session_state hook: %v", err)
	}
	runner := e.runner("session_hook_state")
	_, err := runner.setupTurnRun(ctx, s, "session_hook_state", "turn_hook_state")
	if err != nil {
		t.Fatalf("setup turn run: %v", err)
	}
	runner.finishTurn(s, "turn_hook_state", "session_hook_state", "agent", "bootstrap", "failed", "nope", "provider_error")
	if len(turnStates) < 2 {
		t.Fatalf("expected setup + terminal turn_state hooks, got %#v", turnStates)
	}
	if turnStates[0] != "running/setup" {
		t.Fatalf("expected first turn state running/setup, got %#v", turnStates)
	}
	if turnStates[len(turnStates)-1] != "failed/failed" {
		t.Fatalf("expected terminal turn state failed/failed, got %#v", turnStates)
	}
	if len(sessionStates) < 2 || sessionStates[0] != "running" || sessionStates[len(sessionStates)-1] != "idle" {
		t.Fatalf("expected running→idle session states, got %#v", sessionStates)
	}
	setupPayload := sessionPayloads[0]
	if setupPayload["turn_id"] != "turn_hook_state" || setupPayload["active_turn_id"] != "turn_hook_state" || setupPayload["turn_status"] != "running" || setupPayload["turn_phase"] != "setup" || setupPayload["reason"] != "setup" {
		t.Fatalf("expected setup session_state hook payload with turn correlation, got %#v", setupPayload)
	}
	terminalPayload := sessionPayloads[len(sessionPayloads)-1]
	if terminalPayload["turn_id"] != "turn_hook_state" || terminalPayload["turn_status"] != "failed" || terminalPayload["turn_phase"] != "failed" || terminalPayload["reason"] != "turn_terminal" || terminalPayload["failure_kind"] != "provider_error" {
		t.Fatalf("expected terminal session_state hook payload with turn correlation, got %#v", terminalPayload)
	}
}

func TestQueuedCancelSessionStateHookCarriesResolvedModel(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_cancel_hook_model", "CancelHook", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_cancel_hook_model", "session_cancel_hook_model", "queued", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	payloadCh := make(chan HookRequest, 8)
	if _, err := e.RegisterHook(HookSessionState, "test", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		if req.SessionID == "session_cancel_hook_model" && req.Payload["reason"] == "queued_cancel" {
			payloadCh <- req
		}
		return HookResponse{}, nil
	}); err != nil {
		t.Fatalf("register session_state hook: %v", err)
	}
	if err := e.CancelTurn(ctx, "session_cancel_hook_model", queuedTurn.ID); err != nil {
		t.Fatalf("cancel queued turn: %v", err)
	}
	select {
	case req := <-payloadCh:
		if req.Model != "bootstrap" || req.Payload["turn_id"] != queuedTurn.ID || req.Payload["turn_status"] != "cancelled" || req.Payload["turn_phase"] != "aborted" {
			t.Fatalf("expected queued-cancel session_state hook with resolved model, got %#v", req)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected queued-cancel session_state hook")
	}
}

func TestShellTerminalSessionStateHookCarriesTurnID(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_shell_hook_turn_id", "ShellHook", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	payloadCh := make(chan map[string]any, 8)
	if _, err := e.RegisterHook(HookSessionState, "test", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		if req.SessionID == "session_shell_hook_turn_id" && req.SessionStatus == "idle" {
			payloadCh <- cloneMap(req.Payload)
		}
		return HookResponse{}, nil
	}); err != nil {
		t.Fatalf("register session_state hook: %v", err)
	}
	result, err := e.SubmitPrompt(ctx, RunInput{SessionID: "session_shell_hook_turn_id", Prompt: "hello", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit prompt: %v", err)
	}
	deadline := time.After(2 * time.Second)
	for {
		select {
		case payload := <-payloadCh:
			if payload["turn_id"] != result.TurnID || payload["turn_status"] != "completed" || payload["turn_phase"] != "completed" || payload["reason"] != "turn_completed" {
				t.Fatalf("expected terminal session_state hook payload with turn id, got %#v", payload)
			}
			return
		case <-deadline:
			t.Fatal("expected terminal session_state hook payload for shell turn")
		}
	}
}

func TestHookTimeoutPolicyContinue(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	e.runtimeCfg.Hooks.TimeoutMS = 20
	e.runtimeCfg.Hooks.OnTimeout = "continue"
	called := 0
	if _, err := e.RegisterHook(HookBeforeProviderRequest, "slow", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		<-ctx.Done()
		return HookResponse{}, ctx.Err()
	}); err != nil {
		t.Fatalf("register slow hook: %v", err)
	}
	if _, err := e.RegisterHook(HookBeforeProviderRequest, "fast", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		called++
		return HookResponse{}, nil
	}); err != nil {
		t.Fatalf("register fast hook: %v", err)
	}
	if _, err := e.emitHook(context.Background(), HookRequest{Name: HookBeforeProviderRequest}); err != nil {
		t.Fatalf("emit hook with timeout-continue policy: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected later hooks to continue after timeout, got %d", called)
	}
}

func TestHookErrorPolicyContinue(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	e.runtimeCfg.Hooks.OnError = "continue"
	called := 0
	if _, err := e.RegisterHook(HookToolCall, "broken", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		return HookResponse{}, errors.New("boom")
	}); err != nil {
		t.Fatalf("register broken hook: %v", err)
	}
	if _, err := e.RegisterHook(HookToolCall, "fast", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		called++
		return HookResponse{}, nil
	}); err != nil {
		t.Fatalf("register fast hook: %v", err)
	}
	if _, err := e.emitHook(context.Background(), HookRequest{Name: HookToolCall}); err != nil {
		t.Fatalf("emit hook with error-continue policy: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected later hooks to continue after handler error, got %d", called)
	}
}

func TestHookErrorPolicyReturnsTypedError(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	if _, err := e.RegisterHook(HookToolCall, "broken", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		return HookResponse{}, errors.New("boom")
	}); err != nil {
		t.Fatalf("register broken hook: %v", err)
	}
	_, err := e.emitHook(context.Background(), HookRequest{Name: HookToolCall})
	if err == nil {
		t.Fatal("expected hook execution error")
	}
	var hookErr HookExecutionError
	if !errors.As(err, &hookErr) {
		t.Fatalf("expected typed hook execution error, got %T %v", err, err)
	}
	if hookErr.Kind != "handler_error" || hookErr.Source != "broken" {
		t.Fatalf("unexpected hook execution error: %#v", hookErr)
	}
}

func TestApplyHookDefaultsCompat(t *testing.T) {
	settings := applyHookDefaultsCompat(config.HookSettings{})
	if settings.TimeoutMS <= 0 {
		t.Fatalf("expected default hook timeout, got %#v", settings)
	}
	if settings.OnError != "error" || settings.OnTimeout != "continue" {
		t.Fatalf("unexpected hook defaults: %#v", settings)
	}
	custom := applyHookDefaultsCompat(config.HookSettings{TimeoutMS: 25, OnError: "continue", OnTimeout: "error"})
	if custom.TimeoutMS != 25 || custom.OnError != "continue" || custom.OnTimeout != "error" {
		t.Fatalf("unexpected custom hook defaults: %#v", custom)
	}
}

func TestEmitHookPersistsHookInvocationAudit(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_audit", "Audit", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_audit", "session_audit", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if _, err := e.RegisterHook(HookToolCall, "audit-test", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		return HookResponse{Action: "modify", Payload: map[string]any{"seen": true}}, nil
	}); err != nil {
		t.Fatalf("register hook: %v", err)
	}
	_, err := e.emitHook(ctx, HookRequest{Name: HookToolCall, SessionID: "session_audit", TurnID: "turn_audit", ToolCall: &goai.ToolCall{Type: "toolCall", ID: "tc1", Name: "read", Arguments: map[string]any{"path": "README.md"}}})
	if err != nil {
		t.Fatalf("emit hook: %v", err)
	}
	items, err := s.ListHookInvocationsByTurn(ctx, "turn_audit")
	if err != nil {
		t.Fatalf("list persisted hook invocations: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one persisted hook invocation, got %#v", items)
	}
	if items[0].HookName != HookToolCall || items[0].HookSource != "audit-test" || items[0].Action != "modify" {
		t.Fatalf("unexpected persisted hook invocation: %#v", items[0])
	}
	trace, ok := items[0].Request["trace"].(map[string]any)
	if !ok || trace["id"] == "" {
		t.Fatalf("expected persisted trace metadata, got %#v", items[0])
	}
}

func TestHookInvocationAuditPersistsAfterRequestContextCancellation(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	defer e.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_audit_cancel", "AuditCancel", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_audit_cancel", "session_audit_cancel", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	e.runtimeCfg.Hooks.OnError = "continue"
	if _, err := e.RegisterHook(HookToolCall, "audit-cancel", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		<-ctx.Done()
		return HookResponse{}, ctx.Err()
	}); err != nil {
		t.Fatalf("register hook: %v", err)
	}
	reqCtx, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := e.emitHook(reqCtx, HookRequest{Name: HookToolCall, SessionID: "session_audit_cancel", TurnID: "turn_audit_cancel", ToolCall: &goai.ToolCall{Type: "toolCall", ID: "tc_cancel", Name: "read", Arguments: map[string]any{"path": "README.md"}}}); err != nil {
		t.Fatalf("emit hook with canceled context under continue policy: %v", err)
	}
	items, err := s.ListHookInvocationsByTurn(ctx, "turn_audit_cancel")
	if err != nil {
		t.Fatalf("list persisted hook invocations: %v", err)
	}
	if len(items) != 1 || items[0].HookSource != "audit-cancel" || items[0].Action != "continue" || items[0].ErrorText == "" {
		t.Fatalf("unexpected persisted hook invocations after canceled context: %#v", items)
	}
}

func TestBeforeProviderRequestCanMutateProviderContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_before_llm", "BeforeLLM", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_before_llm", "session_before_llm", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if _, err := e.RegisterHook(HookBeforeProviderRequest, "mutate-llm", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		return HookResponse{
			Action:       "modify",
			SystemPrompt: "mutated system prompt",
			Messages:     []goai.Message{goai.UserMessage("mutated message")},
			Tools:        []goai.Tool{{Name: "hook_tool", Description: "Injected", Parameters: json.RawMessage(`{"type":"object","properties":{}}`)}},
		}, nil
	}); err != nil {
		t.Fatalf("register before_provider_request hook: %v", err)
	}
	var capturedSystemPrompt string
	var capturedMessages []goai.Message
	var capturedTools []goai.Tool
	withStreamWithToolsStub(t, func(ctx context.Context, model string, convCtx *goai.Context, cb func(map[string]any)) (*inference.StreamResult, error) {
		capturedSystemPrompt = convCtx.SystemPrompt
		capturedMessages = append([]goai.Message(nil), convCtx.Messages...)
		capturedTools = append([]goai.Tool(nil), convCtx.Tools...)
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}}, nil
	})
	runner := e.runner("session_before_llm")
	convCtx := &goai.Context{
		SystemPrompt: "original system prompt",
		Messages:     []goai.Message{goai.UserMessage("original message")},
		Tools:        []goai.Tool{{Name: "read", Description: "Read", Parameters: json.RawMessage(`{"type":"object","properties":{}}`)}},
	}
	if _, err := runner.runProviderIteration(ctx, s, "turn_before_llm", "session_before_llm", "bootstrap", "agent", 1, 4, convCtx); err != nil {
		t.Fatalf("run provider iteration: %v", err)
	}
	if capturedSystemPrompt != "mutated system prompt" {
		t.Fatalf("expected mutated system prompt, got %q", capturedSystemPrompt)
	}
	if len(capturedMessages) != 1 || goai.GetTextContent(&capturedMessages[0]) != "mutated message" {
		t.Fatalf("expected mutated messages, got %#v", capturedMessages)
	}
	if len(capturedTools) != 1 || capturedTools[0].Name != "hook_tool" {
		t.Fatalf("expected mutated tools, got %#v", capturedTools)
	}
}

func TestBeforeProviderRequestMessagePrependIsRequestLocal(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_before_llm_message_only", "BeforeLLMMessageOnly", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_before_llm_message_only", "session_before_llm_message_only", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if _, err := e.RegisterHook(HookBeforeProviderRequest, "prepend-message", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		if stringValue(req.Payload["stage"], "") != "context" {
			return HookResponse{}, nil
		}
		return HookResponse{Action: "modify", Message: "temporary prepend"}, nil
	}); err != nil {
		t.Fatalf("register before_provider_request hook: %v", err)
	}
	var capturedMessages []goai.Message
	withStreamWithToolsStub(t, func(ctx context.Context, model string, convCtx *goai.Context, cb func(map[string]any)) (*inference.StreamResult, error) {
		capturedMessages = append([]goai.Message(nil), convCtx.Messages...)
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}}, nil
	})
	runner := e.runner("session_before_llm_message_only")
	convCtx := &goai.Context{SystemPrompt: "original system prompt", Messages: []goai.Message{goai.UserMessage("original message")}}
	if _, err := runner.runProviderIteration(ctx, s, "turn_before_llm_message_only", "session_before_llm_message_only", "bootstrap", "agent", 1, 4, convCtx); err != nil {
		t.Fatalf("run provider iteration: %v", err)
	}
	if len(capturedMessages) != 2 || goai.GetTextContent(&capturedMessages[0]) != "temporary prepend" || goai.GetTextContent(&capturedMessages[1]) != "original message" {
		t.Fatalf("expected request-local prepended message plus original message, got %#v", capturedMessages)
	}
	if len(convCtx.Messages) != 1 || goai.GetTextContent(&convCtx.Messages[0]) != "original message" {
		t.Fatalf("expected shared convCtx messages to remain unchanged after request-local prepend, got %#v", convCtx.Messages)
	}
}

func TestBeforeProviderRequestCanReplaceRawProviderPayload(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_before_llm_payload", "BeforeLLMPayload", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_before_llm_payload", "session_before_llm_payload", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	stageCalls := 0
	if _, err := e.RegisterHook(HookBeforeProviderRequest, "replace-payload", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		if stringValue(req.Payload["stage"], "") != "payload" {
			return HookResponse{}, nil
		}
		stageCalls++
		return HookResponse{Action: "modify", Payload: map[string]any{"request": map[string]any{"model": "replaced-model", "messages": []map[string]any{{"role": "user", "content": "replaced"}}, "stream": true}}}, nil
	}); err != nil {
		t.Fatalf("register before_provider_request hook: %v", err)
	}
	withStreamWithToolsHookStub(t, func(ctx context.Context, model string, convCtx *goai.Context, cb func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
		if hooks == nil || hooks.OnPayload == nil {
			t.Fatalf("expected provider payload hook wiring")
		}
		payload, err := hooks.OnPayload(map[string]any{"model": model, "messages": []map[string]any{{"role": "user", "content": "original"}}, "stream": true}, &goai.Model{ID: model, Provider: goai.Provider("test-provider"), Api: goai.ApiOpenAICompletions})
		if err != nil {
			t.Fatalf("invoke payload hook: %v", err)
		}
		payloadMap, ok := payload.(map[string]any)
		if !ok {
			t.Fatalf("expected replacement payload map, got %#v", payload)
		}
		if stringValue(payloadMap["model"], "") != "replaced-model" {
			t.Fatalf("expected replaced model payload, got %#v", payloadMap)
		}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}}, nil
	})
	runner := e.runner("session_before_llm_payload")
	convCtx := &goai.Context{SystemPrompt: "original", Messages: []goai.Message{goai.UserMessage("hello")}}
	if _, err := runner.runProviderIteration(ctx, s, "turn_before_llm_payload", "session_before_llm_payload", "bootstrap", "agent", 1, 4, convCtx); err != nil {
		t.Fatalf("run provider iteration: %v", err)
	}
	if stageCalls != 1 {
		t.Fatalf("expected payload-stage hook to run once, got %d", stageCalls)
	}
}

func TestAfterProviderResponseReceivesObservedStatusAndHeaders(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_after_llm", "AfterLLM", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_after_llm", "session_after_llm", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	var capturedPayload map[string]any
	if _, err := e.RegisterHook(HookAfterProviderResponse, "observe-response", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		capturedPayload = req.Payload
		return HookResponse{}, nil
	}); err != nil {
		t.Fatalf("register after_provider_response hook: %v", err)
	}
	withStreamWithToolsHookStub(t, func(ctx context.Context, model string, convCtx *goai.Context, cb func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
		if hooks == nil || hooks.OnResponse == nil {
			t.Fatalf("expected provider response hook wiring")
		}
		hooks.OnResponse(202, map[string]string{"x-test-header": "ok"}, &goai.Model{ID: model, Provider: goai.Provider("test-provider"), Api: goai.ApiOpenAICompletions})
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}}, nil
	})
	runner := e.runner("session_after_llm")
	convCtx := &goai.Context{SystemPrompt: "original", Messages: []goai.Message{goai.UserMessage("hello")}}
	if _, err := runner.runProviderIteration(ctx, s, "turn_after_llm", "session_after_llm", "bootstrap", "agent", 1, 4, convCtx); err != nil {
		t.Fatalf("run provider iteration: %v", err)
	}
	if capturedPayload == nil {
		t.Fatal("expected after_provider_response payload")
	}
	status, ok := capturedPayload["status"].(int)
	if !ok {
		if statusFloat, ok := capturedPayload["status"].(float64); ok {
			status = int(statusFloat)
		}
	}
	if status != 202 {
		t.Fatalf("expected status 202, got %#v", capturedPayload)
	}
	headers, ok := capturedPayload["headers"].(map[string]string)
	if !ok {
		generic, ok := capturedPayload["headers"].(map[string]any)
		if !ok {
			t.Fatalf("expected response headers map, got %#v", capturedPayload["headers"])
		}
		headers = map[string]string{}
		for k, v := range generic {
			headers[k] = stringValue(v, "")
		}
	}
	if headers["x-test-header"] != "ok" {
		t.Fatalf("expected response header payload, got %#v", headers)
	}
	if stringValue(capturedPayload["provider"], "") != "test-provider" {
		t.Fatalf("expected provider metadata, got %#v", capturedPayload)
	}
}

func TestToolCallHookCanMutateArgumentsDuringExecution(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_tool_mutate", "ToolMutate", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_tool_mutate", "session_tool_mutate", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	executedValue := ""
	if err := e.RegisterTool(RegisteredTool{Name: "echo_test", Description: "Echo", Executor: func(ctx context.Context, rt ToolRuntime, call goai.ToolCall) (string, error) {
		executedValue = stringValue(call.Arguments["value"], "")
		return "exec:" + executedValue, nil
	}}); err != nil {
		t.Fatalf("register tool: %v", err)
	}
	if _, err := e.RegisterHook(HookToolCall, "mutate-tool-call", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		mutated := *req.ToolCall
		mutated.Arguments = map[string]any{"value": "mutated"}
		return HookResponse{Action: "modify", ToolCall: &mutated}, nil
	}); err != nil {
		t.Fatalf("register tool_call hook: %v", err)
	}
	runner := e.runner("session_tool_mutate")
	outcome := runner.executeToolCallsPhase(ctx, s, "turn_tool_mutate", "session_tool_mutate", "bootstrap", "agent", 1, &goai.Context{}, []goai.ToolCall{{Type: "toolCall", ID: "tc_mut", Name: "echo_test", Arguments: map[string]any{"value": "original"}}}, nil, "", 0, &goai.Usage{})
	if outcome.terminated {
		t.Fatalf("expected tool phase to continue, got %#v", outcome)
	}
	if executedValue != "mutated" {
		t.Fatalf("expected mutated tool argument to execute, got %q", executedValue)
	}
	msgs, err := s.ListMessages(ctx, "session_tool_mutate")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Content != "exec:mutated" {
		t.Fatalf("expected mutated tool result message, got %#v", msgs)
	}
}

func TestToolCallHookCanRespondWithoutExecutingTool(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_tool_respond", "ToolRespond", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_tool_respond", "session_tool_respond", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	afterToolCalls := 0
	if _, err := e.RegisterHook(HookToolCall, "respond-hook", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		result := "hook injected result"
		return HookResponse{Action: "respond", Handled: true, ToolResult: &result}, nil
	}); err != nil {
		t.Fatalf("register tool_call hook: %v", err)
	}
	if _, err := e.RegisterHook(HookToolResult, "after-tool", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		afterToolCalls++
		return HookResponse{}, nil
	}); err != nil {
		t.Fatalf("register tool_result hook: %v", err)
	}
	toolCh, unsub := e.Topics().Subscribe(ctx, "runtime.tool", topics.SubscribeOptions{Buffer: 8, SessionID: "session_tool_respond"})
	defer unsub()
	runner := e.runner("session_tool_respond")
	outcome := runner.executeToolCallsPhase(ctx, s, "turn_tool_respond", "session_tool_respond", "bootstrap", "agent", 1, &goai.Context{}, []goai.ToolCall{{Type: "toolCall", ID: "tc_resp", Name: "plugin_tool", Arguments: map[string]any{"value": "x"}}}, nil, "", 0, &goai.Usage{})
	if outcome.terminated {
		t.Fatalf("expected hook response injection to continue, got %#v", outcome)
	}
	if afterToolCalls != 0 {
		t.Fatalf("expected after_tool hook to be skipped on direct response, got %d calls", afterToolCalls)
	}
	msgs, err := s.ListMessages(ctx, "session_tool_respond")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Content != "hook injected result" {
		t.Fatalf("expected injected tool result message, got %#v", msgs)
	}
	if msgs[0].Payload["source"] != "hook" || msgs[0].Payload["hook_phase"] != "tool_call" {
		t.Fatalf("expected injected tool result hook metadata, got %#v", msgs[0].Payload)
	}
	events, err := s.ListTurnEvents(ctx, "turn_tool_respond")
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundFinished := false
	for _, event := range events {
		if event.Type == "tool.finished" && event.Payload["tool"] == "plugin_tool" && event.Payload["tool_call_id"] == "tc_resp" && event.Payload["source"] == "hook" && event.Payload["hook_phase"] == "tool_call" {
			foundFinished = true
		}
	}
	if !foundFinished {
		t.Fatalf("expected tool.finished event for hook-responded tool, got %#v", events)
	}
	select {
	case env := <-toolCh:
		if env.Topic != "runtime.tool" || env.Payload["type"] != "tool_finished" || env.Payload["tool"] != "plugin_tool" || env.Payload["tool_call_id"] != "tc_resp" || env.Payload["source"] != "hook" || env.Payload["hook_phase"] != "tool_call" {
			t.Fatalf("unexpected runtime.tool hook-respond payload: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected runtime.tool tool_finished topic for hook-responded tool")
	}
}

func TestApproveToolHookCanDenyExecution(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_tool_deny", "ToolDeny", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_tool_deny", "session_tool_deny", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	executions := 0
	if err := e.RegisterTool(RegisteredTool{Name: "deny_test", Description: "Deny", Executor: func(ctx context.Context, rt ToolRuntime, call goai.ToolCall) (string, error) {
		executions++
		return "should not run", nil
	}}); err != nil {
		t.Fatalf("register tool: %v", err)
	}
	if _, err := e.RegisterHook(HookApproveTool, "deny-hook", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		return HookResponse{Action: "deny", Block: true, Reason: "policy denied"}, nil
	}); err != nil {
		t.Fatalf("register approve hook: %v", err)
	}
	toolCh, unsub := e.Topics().Subscribe(ctx, "runtime.tool", topics.SubscribeOptions{Buffer: 8, SessionID: "session_tool_deny"})
	defer unsub()
	runner := e.runner("session_tool_deny")
	outcome := runner.executeToolCallsPhase(ctx, s, "turn_tool_deny", "session_tool_deny", "bootstrap", "agent", 1, &goai.Context{}, []goai.ToolCall{{Type: "toolCall", ID: "tc_deny", Name: "deny_test", Arguments: map[string]any{"value": "x"}}}, nil, "", 0, &goai.Usage{})
	if outcome.terminated {
		t.Fatalf("expected denied tool to stay in turn loop, got %#v", outcome)
	}
	if executions != 0 {
		t.Fatalf("expected denied tool not to execute, got %d executions", executions)
	}
	msgs, err := s.ListMessages(ctx, "session_tool_deny")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Role != "tool_result" || msgs[0].Content == "" || msgs[0].Content == "should not run" {
		t.Fatalf("expected denial tool_result message, got %#v", msgs)
	}
	if msgs[0].Payload["skipped"] != true || msgs[0].Payload["skip_reason"] != "policy denied" || msgs[0].Payload["hook_phase"] != "approve_tool" {
		t.Fatalf("expected denial tool_result skip metadata, got %#v", msgs[0].Payload)
	}
	events, err := s.ListTurnEvents(ctx, "turn_tool_deny")
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundSkipped := false
	for _, event := range events {
		if event.Type == "tool.skipped" && event.Payload["tool"] == "deny_test" && event.Payload["tool_call_id"] == "tc_deny" && event.Payload["reason"] == "policy denied" && event.Payload["hook_phase"] == "approve_tool" {
			foundSkipped = true
		}
	}
	if !foundSkipped {
		t.Fatalf("expected tool.skipped event for denied tool, got %#v", events)
	}
	select {
	case env := <-toolCh:
		if env.Topic != "runtime.tool" || env.Payload["type"] != "tool_skipped" || env.Payload["tool"] != "deny_test" || env.Payload["tool_call_id"] != "tc_deny" || env.Payload["reason"] != "policy denied" || env.Payload["hook_phase"] != "approve_tool" {
			t.Fatalf("unexpected runtime.tool deny-skip payload: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected runtime.tool tool_skipped topic for denied tool")
	}
}

func TestHookAbortSemanticsAbortTurnDuringToolCall(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_hook_abort", "HookAbort", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_hook_abort", "session_hook_abort", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if _, err := e.RegisterHook(HookToolCall, "abort-hook", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		return HookResponse{Action: "abort_turn", Cancel: true, Block: true, Reason: "stop now"}, nil
	}); err != nil {
		t.Fatalf("register tool_call hook: %v", err)
	}
	runner := e.runner("session_hook_abort")
	outcome := runner.executeToolCallsPhase(ctx, s, "turn_hook_abort", "session_hook_abort", "bootstrap", "agent", 1, &goai.Context{}, []goai.ToolCall{{Type: "toolCall", ID: "tc_abort", Name: "unknown_tool", Arguments: map[string]any{}}}, nil, "", 0, &goai.Usage{})
	if !outcome.terminated {
		t.Fatalf("expected aborting hook to terminate turn, got %#v", outcome)
	}
	turnRec, err := s.GetTurn(ctx, "turn_hook_abort")
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if turnRec.Status != "aborted" || turnRec.Phase != "aborted" {
		t.Fatalf("expected aborted turn, got %#v", turnRec)
	}
	msgs, err := s.ListMessages(ctx, "session_hook_abort")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(msgs) == 0 || msgs[len(msgs)-1].Content != "stop now" {
		t.Fatalf("expected terminal abort message, got %#v", msgs)
	}
}

func TestProcessHookHandshakeAndBeforeToolProtocol(t *testing.T) {
	root := t.TempDir()
	hookPath := filepath.Join(root, "hook.sh")
	script := `#!/bin/sh
set -eu
[ "${GI_HOOK_NAME:-}" = "tool_call" ] || { echo "bad hook env" >&2; exit 1; }
[ "${GI_SESSION_ID:-}" = "session_process_hook" ] || { echo "bad session env" >&2; exit 1; }
while IFS= read -r line; do
  case "$line" in
    *'"method":"hook.hello"'*) printf '%s\n' '{"jsonrpc":"2.0","id":1,"result":{"ok":true,"name":"process-hook"}}' ;;
    *'"method":"hook.before_tool"'*) printf '%s\n' '{"jsonrpc":"2.0","id":2,"result":{"action":"modify","tool_call":{"id":"tc_proc","name":"read","arguments":{"path":"process.md"}}}}' ;;
    *) echo "bad call" >&2; exit 1 ;;
  esac
done
`
	if err := os.WriteFile(hookPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write hook script: %v", err)
	}
	s := openTestStore(t)
	defer s.Close()
	e := NewWithRuntimeConfig(s, config.RuntimeConfig{WorkspaceRoot: root, DefaultModel: "bootstrap", Agents: routing.AgentsConfig{List: []routing.AgentConfig{{ID: "agent", Default: true, Model: "bootstrap"}}}}, "")
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_process_hook", "ProcessHook", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_process_hook", "session_process_hook", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	spec := scripting.EventHookSpec{Name: HookToolCall, Engine: "process", Command: hookPath, Source: "process-test"}
	if _, err := e.RegisterHook(HookToolCall, spec.Source, newProcessHookHandler(root, spec)); err != nil {
		t.Fatalf("register process hook: %v", err)
	}
	resp, err := e.emitHook(ctx, HookRequest{Name: HookToolCall, SessionID: "session_process_hook", TurnID: "turn_process_hook", ToolCall: &goai.ToolCall{Type: "toolCall", ID: "tc_orig", Name: "read", Arguments: map[string]any{"path": "original.md"}}})
	if err != nil {
		t.Fatalf("emit process hook: %v", err)
	}
	if resp.ToolCall == nil || resp.ToolCall.Name != "read" || stringValue(resp.ToolCall.Arguments["path"], "") != "process.md" {
		t.Fatalf("expected process hook mutation, got %#v", resp)
	}
	items, err := s.ListHookInvocationsByTurn(ctx, "turn_process_hook")
	if err != nil {
		t.Fatalf("list hook invocations: %v", err)
	}
	if len(items) != 1 || items[0].HookSource != "process-test" || items[0].Action != "modify" {
		t.Fatalf("unexpected process hook audit rows: %#v", items)
	}
}

func TestProcessHookReusesMountedProcessAcrossInvocations(t *testing.T) {
	root := t.TempDir()
	hookPath := filepath.Join(root, "hook.sh")
	statePath := filepath.Join(root, "state.txt")
	script := fmt.Sprintf(`#!/bin/sh
set -eu
state_file=%q
: > "$state_file"
while IFS= read -r line; do
  case "$line" in
    *'"method":"hook.hello"'*)
      printf 'hello\n' >> "$state_file"
      printf '%%s\n' '{"jsonrpc":"2.0","id":1,"result":{"ok":true,"name":"process-hook"}}'
      ;;
    *'"method":"hook.before_tool"'*)
      printf 'call\n' >> "$state_file"
      printf '%%s\n' '{"jsonrpc":"2.0","id":2,"result":{"action":"modify","tool_call":{"id":"tc_proc","name":"read","arguments":{"path":"mounted.md"}}}}'
      ;;
    *) echo "bad call" >&2; exit 1 ;;
  esac
done
`, statePath)
	if err := os.WriteFile(hookPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write hook script: %v", err)
	}
	handler := newProcessHookHandler(root, scripting.EventHookSpec{Name: HookToolCall, Engine: "process", Command: hookPath, Source: "mounted-process-test"})
	for i := 0; i < 2; i++ {
		resp, err := handler(context.Background(), HookRequest{Name: HookToolCall, SessionID: "session_mounted_process_hook", TurnID: fmt.Sprintf("turn_%d", i), ToolCall: &goai.ToolCall{Type: "toolCall", ID: fmt.Sprintf("tc_%d", i), Name: "read", Arguments: map[string]any{"path": "original.md"}}})
		if err != nil {
			t.Fatalf("invoke mounted process hook %d: %v", i, err)
		}
		if resp.ToolCall == nil || stringValue(resp.ToolCall.Arguments["path"], "") != "mounted.md" {
			t.Fatalf("expected mounted process hook mutation, got %#v", resp)
		}
	}
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read mounted hook state: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != "hello\ncall\ncall" {
		t.Fatalf("expected one hello and two calls through mounted process, got %q", got)
	}
}
