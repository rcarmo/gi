package turn

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func TestInvalidToolArgumentsRejectBeforeCallHooksAndEffects(t *testing.T) {
	ctx := context.Background()
	s, err := store.Open(filepath.Join(t.TempDir(), "invalid-args.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e := New(s)
	defer e.Close()
	if _, err := s.CreateSession(ctx, "s", "s", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, "t", "s", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	effects, hooks := 0, 0
	if err := e.RegisterTool(tools.RegisteredTool{Name: "invalid_fixture", Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) {
		effects++
		return "unexpected", nil
	}}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{HookToolCall, HookApproveTool, HookToolResult} {
		if _, err := e.RegisterHook(name, "invalid-args-fixture", func(context.Context, HookRequest) (HookResponse, error) { hooks++; return HookResponse{}, nil }); err != nil {
			t.Fatal(err)
		}
	}
	conv := &goai.Context{}
	call := goai.ToolCall{ID: "c", Name: "invalid_fixture", Arguments: map[string]any{}, ArgumentsError: "Expected an object, got array"}
	result := e.runner("s").executeToolCallsPhase(ctx, s, "t", "s", "fixture", "agent", 1, conv, []goai.ToolCall{call}, nil, "", 0, &goai.Usage{})
	if result.terminated || effects != 0 || hooks != 0 {
		t.Fatalf("invalid args reached host: effects=%d hooks=%d result=%+v", effects, hooks, result)
	}
	if len(conv.Messages) != 1 || !conv.Messages[0].IsError {
		t.Fatalf("missing error result: %+v", conv.Messages)
	}
	rows, err := s.ListMessages(ctx, "s")
	if err != nil || len(rows) != 1 || rows[0].Payload["reason"] != "invalid_arguments" {
		t.Fatalf("saved result: %+v %v", rows, err)
	}
	events, err := s.ListTurnEvents(ctx, "t")
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.Type == "tool.started" {
			t.Fatal("invalid call was started")
		}
	}
}

func TestDirectToolDispatchRejectsArgumentsErrorBeforeLookup(t *testing.T) {
	e := New(nil)
	_, err := e.runner("unused").executeToolWithImages(context.Background(), goai.ToolCall{Name: "not_registered", ArgumentsError: "bad JSON"}, "unused", "unused", nil, nil)
	if err == nil || err.Error() != "Invalid tool arguments: bad JSON" {
		t.Fatalf("marker ignored: %v", err)
	}
}
