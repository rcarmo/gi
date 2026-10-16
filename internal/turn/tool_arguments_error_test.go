package turn

import (
	"context"
	"encoding/json"
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
	for _, name := range []string{HookToolExecutionStart, HookToolExecutionEnd, HookToolCall, HookApproveTool, HookToolResult} {
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

func TestInvalidArgumentsSurviveDatabaseReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "invalid-reopen.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateSession(ctx, "s", "s", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, "t", "s", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	call := goai.ToolCall{ID: "bad", Name: "shell", Arguments: map[string]any{}, ArgumentsError: "Expected object"}
	if err := s.AddMessage(ctx, "call-row", "s", "assistant", "", map[string]any{"kind": "tool_calls", "tool_calls": []any{map[string]any{"id": call.ID, "name": call.Name, "arguments": call.Arguments, "argumentsError": call.ArgumentsError}}}); err != nil {
		t.Fatal(err)
	}
	e := New(s)
	if err := e.runner("s").rejectInvalidToolArguments(ctx, s, "s", "t", "agent", 1, &goai.Context{}, call); err != nil {
		t.Fatal(err)
	}
	e.Close()
	s.Close()
	s, err = store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	snapshot, err := s.ContextSnapshot(ctx, "s")
	if err != nil {
		t.Fatal(err)
	}
	messages := (&sessionRunner{}).projectContextSnapshot(ctx, "s", snapshot)
	if len(messages) != 2 {
		t.Fatalf("replay pair missing: %+v", messages)
	}
	calls := goai.GetToolCalls(&messages[0])
	if len(calls) != 1 || calls[0].ArgumentsError != "Expected object" {
		t.Fatalf("replayed call became valid: %+v", calls)
	}
	if messages[1].ToolCallID != "bad" || !messages[1].IsError {
		t.Fatalf("replayed error unpaired: %+v", messages[1])
	}
}

func TestNestedInvalidArgumentsRejectBeforeHooksAndEffects(t *testing.T) {
	e := New(nil)
	hooks, effects := 0, 0
	for _, name := range []string{HookToolCall, HookApproveTool, HookToolResult} {
		if _, err := e.RegisterHook(name, "invalid-nested", func(context.Context, HookRequest) (HookResponse, error) { hooks++; return HookResponse{}, nil }); err != nil {
			t.Fatal(err)
		}
	}
	tool := tools.RegisteredTool{Name: "fixture", Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) {
		effects++
		return "unexpected", nil
	}}
	for _, raw := range []string{"null", "[]", `{"x":`, "42"} {
		if _, err := e.runNestedTool(context.Background(), tools.ToolRuntime{}, "nested", tool, json.RawMessage(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	if hooks != 0 || effects != 0 {
		t.Fatalf("nested parse reached host: hooks=%d effects=%d", hooks, effects)
	}
}

func TestMixedAndHookModifiedInvalidCalls(t *testing.T) {
	for _, phase := range []string{HookToolCall, HookApproveTool} {
		t.Run(phase, func(t *testing.T) {
			ctx := context.Background()
			s, err := store.Open(filepath.Join(t.TempDir(), "mixed.db"))
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
			effects := 0
			approvalIDs := []string{}
			counts := []int{}
			if err := e.RegisterTool(tools.RegisteredTool{Name: "fixture", Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) { effects++; return "ok", nil }}); err != nil {
				t.Fatal(err)
			}
			for _, hook := range []string{HookToolExecutionStart, HookToolExecutionEnd} {
				if _, err := e.RegisterHook(hook, "counts", func(_ context.Context, r HookRequest) (HookResponse, error) {
					counts = append(counts, r.Payload["count"].(int))
					return HookResponse{}, nil
				}); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := e.RegisterHook(HookApproveTool, "approval", func(_ context.Context, r HookRequest) (HookResponse, error) {
				approvalIDs = append(approvalIDs, r.ToolCall.ID)
				return HookResponse{}, nil
			}); err != nil {
				t.Fatal(err)
			}
			if _, err := e.RegisterHook(phase, "mark", func(_ context.Context, r HookRequest) (HookResponse, error) {
				if r.ToolCall.ID != "modified" {
					return HookResponse{}, nil
				}
				c := *r.ToolCall
				c.ArgumentsError = "hook introduced failure"
				return HookResponse{Action: "modify", ToolCall: &c}, nil
			}); err != nil {
				t.Fatal(err)
			}
			conv := &goai.Context{}
			out := e.runner("s").executeToolCallsPhase(ctx, s, "t", "s", "fixture", "agent", 1, conv, []goai.ToolCall{
				{ID: "invalid", Name: "fixture", ArgumentsError: "parser failure"},
				{ID: "modified", Name: "fixture", Arguments: map[string]any{}},
				{ID: "valid", Name: "fixture", Arguments: map[string]any{}},
			}, nil, "", 0, &goai.Usage{})
			if out.terminated || effects != 1 {
				t.Fatalf("effects=%d outcome=%+v", effects, out)
			}
			if len(counts) != 2 || counts[0] != 2 || counts[1] != 2 {
				t.Fatalf("batch counts %+v", counts)
			}
			if len(conv.Messages) != 3 || !conv.Messages[0].IsError || !conv.Messages[1].IsError || conv.Messages[2].IsError {
				t.Fatalf("result order %+v", conv.Messages)
			}
			for _, id := range approvalIDs {
				if id == "invalid" || (phase == HookToolCall && id == "modified") {
					t.Fatalf("invalid reached approval %+v", approvalIDs)
				}
			}
			events, err := s.ListTurnEvents(ctx, "t")
			if err != nil {
				t.Fatal(err)
			}
			for _, ev := range events {
				if ev.Type == "tool.started" && ev.Payload["tool_call_id"] != "valid" {
					t.Fatalf("invalid started %+v", ev)
				}
			}
		})
	}
}

func TestNestedHookModifiedArgumentsError(t *testing.T) {
	for _, phase := range []string{HookToolCall, HookApproveTool} {
		t.Run(phase, func(t *testing.T) {
			e := New(nil)
			effects := 0
			if _, err := e.RegisterHook(phase, "mark", func(_ context.Context, r HookRequest) (HookResponse, error) {
				c := *r.ToolCall
				c.ArgumentsError = "invalid mutation"
				return HookResponse{Action: "modify", ToolCall: &c}, nil
			}); err != nil {
				t.Fatal(err)
			}
			tool := tools.RegisteredTool{Name: "fixture", Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) {
				effects++
				return "unexpected", nil
			}}
			if _, err := e.runNestedTool(context.Background(), tools.ToolRuntime{}, "nested", tool, json.RawMessage(`{}`)); err == nil || effects != 0 {
				t.Fatalf("effects=%d err=%v", effects, err)
			}
		})
	}
}
