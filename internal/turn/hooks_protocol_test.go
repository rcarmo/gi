package turn

import (
	"context"
	"encoding/json"
	"testing"

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
	if _, err := e.RegisterHook(HookTurnState, "test", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		turnStates = append(turnStates, req.TurnStatus+"/"+req.TurnPhase)
		return HookResponse{}, nil
	}); err != nil {
		t.Fatalf("register turn_state hook: %v", err)
	}
	if _, err := e.RegisterHook(HookSessionState, "test", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		sessionStates = append(sessionStates, req.SessionStatus)
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
}
