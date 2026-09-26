package turn

import (
	"context"
	"errors"
	"testing"

	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func TestToolTerminalCancellationAndResultAbortPersistOccurrence(t *testing.T) {
	for _, state := range []string{"cancelled", "cancelled-write-failure", "aborted", "completed", "failed"} {
		t.Run(state, func(t *testing.T) {
			s := openTestStore(t)
			defer s.Close()
			e := New(s)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
				t.Fatal(err)
			}
			if _, err := s.CreateTurnWithStatus(ctx, "t", "A", "running", "test", map[string]any{"model": "bootstrap"}); err != nil {
				t.Fatal(err)
			}
			if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "t", "owner", "claim"); err != nil || !ok {
				t.Fatal(ok, err)
			}
			writeFailure := state == "cancelled-write-failure"
			if writeFailure {
				state = "cancelled"
				if _, err := s.DB().Exec(`create trigger reject_tool_terminal before insert on turn_events when new.event_type='tool.cancelled' begin select raise(abort,'terminal write unavailable'); end`); err != nil {
					t.Fatal(err)
				}
			}
			if err := e.RegisterTool(tools.RegisteredTool{Name: "terminal-test", Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) {
				if state == "cancelled" {
					cancel()
					return "", context.Canceled
				}
				if state == "failed" {
					return "", errors.New("tool failed")
				}
				return "ok", nil
			}}); err != nil {
				t.Fatal(err)
			}
			if state == "aborted" {
				if _, err := e.RegisterHook(HookToolResult, "abort-result", func(context.Context, HookRequest) (HookResponse, error) {
					return HookResponse{Action: "abort", Cancel: true, Reason: "stop result"}, nil
				}); err != nil {
					t.Fatal(err)
				}
			}
			runner := e.runner("A")
			out := runner.executeToolCallsPhase(ctx, s, "t", "A", "bootstrap", "agent", 1, &goai.Context{}, []goai.ToolCall{{Type: "toolCall", ID: "call", Name: "terminal-test"}}, nil, "", 0, &goai.Usage{})
			if (state == "cancelled" || state == "aborted") != out.terminated {
				t.Fatal(out)
			}
			events, err := s.ListTurnEvents(context.Background(), "t")
			if err != nil {
				t.Fatal(err)
			}
			expected := "tool." + state
			if state == "completed" {
				expected = "tool.finished"
			}
			var start, terminal map[string]any
			for _, event := range events {
				if event.Type == "tool.started" {
					start = event.Payload
				}
				if event.Type == expected {
					terminal = event.Payload
				}
			}
			if writeFailure {
				if terminal != nil {
					t.Fatal("failed terminal was persisted", terminal)
				}
				activity, err := s.SessionActivity(context.Background(), "A")
				if err != nil {
					t.Fatal(err)
				}
				tool := activity["tool"].(map[string]any)
				if tool["state"] != "interrupted" || tool["duration_ms"] != nil || tool["finished_at"] != "" {
					t.Fatal(tool)
				}
				turn, err := s.GetTurn(context.Background(), "t")
				if err != nil || turn.Status != "cancelled" {
					t.Fatal(turn, err)
				}
				return
			}
			if start == nil || terminal == nil || start["occurrence_id"] == "" || start["occurrence_id"] == nil || start["occurrence_id"] != terminal["occurrence_id"] {
				t.Fatal(events)
			}
			activity, err := s.SessionActivity(context.Background(), "A")
			if err != nil {
				t.Fatal(err)
			}
			tool := activity["tool"].(map[string]any)
			if tool["state"] != state || tool["duration_ms"] == nil || tool["tool_call_id"] != "call" {
				t.Fatal(tool)
			}
		})
	}
}

func TestToolTerminalHookResponseCannotFinishPriorLegacyCall(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "t", "A", "running", "test", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "t", "owner", "claim"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if err := s.AppendTurnEvent(ctx, "t", "A", "tool.started", map[string]any{"tool": "same", "tool_call_id": "reused"}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.RegisterHook(HookToolCall, "respond", func(context.Context, HookRequest) (HookResponse, error) {
		text := "synthetic"
		return HookResponse{ToolResult: &text}, nil
	}); err != nil {
		t.Fatal(err)
	}
	result := e.runner("A").executeToolCallsPhase(ctx, s, "t", "A", "bootstrap", "agent", 1, &goai.Context{}, []goai.ToolCall{{Type: "toolCall", ID: "reused", Name: "same"}}, nil, "", 0, &goai.Usage{})
	if result.terminated {
		t.Fatal(result)
	}
	events, err := s.ListTurnEvents(ctx, "t")
	if err != nil {
		t.Fatal(err)
	}
	starts, finishes := 0, 0
	for _, event := range events {
		if event.Type == "tool.started" {
			starts++
		}
		if event.Type == "tool.finished" {
			finishes++
			if event.Payload["occurrence_id"] == nil || event.Payload["occurrence_id"] == "" {
				t.Fatal(event)
			}
		}
	}
	if starts != 1 || finishes != 1 {
		t.Fatal(events)
	}
	activity, err := s.SessionActivity(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	tool := activity["tool"].(map[string]any)
	if tool["state"] != "running" || tool["duration_ms"] != nil {
		t.Fatal("synthetic result finished old execution", tool)
	}
}
