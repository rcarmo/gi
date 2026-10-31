package turn

import (
	"context"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func TestToolOutputPolicyBoundsWholeLinesAndUTF8(t *testing.T) {
	for _, tc := range []struct{ name, text, want string }{
		{"lines", strings.Repeat("line\n", 2001), strings.Repeat("line\n", 2000)},
		{"multibyte", strings.Repeat("界", 18000), strings.Repeat("界", toolOutputMaxBytes/3)},
		{"whole-lines", "first\n" + strings.Repeat("x", toolOutputMaxBytes), "first\n"},
		{"small", "unchanged\n", "unchanged\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, diagnostics := boundToolOutput(tc.text)
			if got != tc.want || !utf8.ValidString(got) || len(got) > toolOutputMaxBytes || toolTextLines(got) > toolOutputMaxLines {
				t.Fatalf("bad output: bytes=%d diagnostics=%+v", len(got), diagnostics)
			}
			if diagnostics["truncated"] != (got != tc.text) || diagnostics["total_bytes"] != len(tc.text) {
				t.Fatal(diagnostics)
			}
		})
	}
}

func TestToolOutputPolicyAppliesToHookSuccessAndFailure(t *testing.T) {
	for _, mode := range []string{"success", "failure", "hook-response", "hook-replacement"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			s := openTestStore(t)
			defer s.Close()
			e := New(s)
			defer e.Close()
			if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
				t.Fatal(err)
			}
			if _, err := s.CreateTurnWithStatus(ctx, "t", "A", "running", "test", nil); err != nil {
				t.Fatal(err)
			}
			if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "t", "owner", "claim"); err != nil || !ok {
				t.Fatal(ok, err)
			}
			full := strings.Repeat(strings.Repeat("界", 30)+"\n", 700)
			expected, _ := boundToolOutput(full)
			effects := 0
			if err := e.RegisterTool(tools.RegisteredTool{Name: "bounded-output", Executor: func(_ context.Context, rt tools.ToolRuntime, _ goai.ToolCall) (string, error) {
				effects++
				rt.SetDetails(map[string]any{"spill_file": "vfs://workspace/full-output.txt"})
				if mode == "failure" {
					return "temporary progress", errors.New(full)
				}
				if mode == "hook-replacement" {
					return "temporary progress", nil
				}
				return full, nil
			}}); err != nil {
				t.Fatal(err)
			}
			if mode == "hook-response" || mode == "hook-replacement" {
				hook := HookToolCall
				if mode == "hook-replacement" {
					hook = HookToolResult
				}
				if _, err := e.RegisterHook(hook, "replace", func(context.Context, HookRequest) (HookResponse, error) {
					return HookResponse{Action: "respond", Handled: true, ToolResult: &full}, nil
				}); err != nil {
					t.Fatal(err)
				}
			}
			conv := &goai.Context{}
			outcome := e.runner("A").executeToolCallsPhase(ctx, s, "t", "A", "bootstrap", "agent", 1, conv, []goai.ToolCall{{Type: "toolCall", ID: "call", Name: "bounded-output", Arguments: map[string]any{}}}, nil, "", 0, &goai.Usage{})
			if outcome.terminated || len(conv.Messages) != 1 || goai.GetTextContent(&conv.Messages[0]) != expected {
				t.Fatalf("model output: %+v", conv.Messages)
			}
			rows, err := s.ListMessages(ctx, "A")
			if err != nil || len(rows) != 1 {
				t.Fatal(rows, err)
			}
			row := rows[0]
			if row.Content != expected || row.Payload["output_policy"].(map[string]any)["truncated"] != true {
				t.Fatal("unbounded persisted output", row.Payload)
			}
			if mode == "hook-response" {
				if effects != 0 {
					t.Fatal("hook response executed tool")
				}
			} else {
				if row.Payload["details"].(map[string]any)["spill_file"] != "vfs://workspace/full-output.txt" {
					t.Fatal("spill reference lost", row.Payload)
				}
				events, err := s.ListTurnEvents(ctx, "t")
				if err != nil {
					t.Fatal(err)
				}
				var last map[string]any
				for _, ev := range events {
					if ev.Type == "tool.output" {
						last = ev.Payload
					}
				}
				expectedPreview := store.ToolOutputPreview(expected)
				if last == nil || last["final"] != true || last["output_total_lines"] != float64(expectedPreview["output_total_lines"].(int)) || last["output_preview"] != expectedPreview["output_preview"] {
					t.Fatal("final output did not replace progress")
				}
			}
		})
	}
}
