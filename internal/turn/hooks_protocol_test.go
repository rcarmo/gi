package turn

import (
	"context"
	"testing"
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
