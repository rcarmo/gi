package turn

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func TestAgentLoopValidatesTerminalBeforeToolExecutionOrSuccess(t *testing.T) {
	for _, stop := range []goai.StopReason{goai.StopReasonStop, goai.StopReasonLength, goai.StopReasonToolUse, goai.StopReasonPending, goai.StopReasonDeferred, goai.StopReasonError, goai.StopReasonAborted, "", "unknown"} {
		for _, calls := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/calls=%v", stop, calls), func(t *testing.T) {
				ctx := context.Background()
				s := openTestStore(t)
				defer s.Close()
				e := New(s)
				defer e.Close()
				inference.Init()
				model := "terminal-contract/fixture"
				goai.RegisterModel(&goai.Model{ID: "fixture", Provider: "terminal-contract", Api: goai.ApiOpenAICompletions, ContextWindow: 32000, MaxTokens: 1024})
				if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
					t.Fatal(err)
				}
				var effects, requests atomic.Int32
				if err := e.RegisterTool(tools.RegisteredTool{Name: "terminal-fixture", Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) {
					effects.Add(1)
					return "ok", nil
				}}); err != nil {
					t.Fatal(err)
				}
				original := streamWithToolsWithHooks
				defer func() { streamWithToolsWithHooks = original }()
				streamWithToolsWithHooks = func(context.Context, string, *goai.Context, func(map[string]any), *inference.StreamHooks) (*inference.StreamResult, error) {
					message := replayAnswer("answer")
					message.StopReason = stop
					if requests.Add(1) == 1 {
						if calls {
							message.Content = append(message.Content, goai.ContentBlock{Type: "toolCall", ID: "c", Name: "terminal-fixture", Arguments: map[string]any{}})
						}
					} else {
						message.StopReason = goai.StopReasonStop
					}
					return &inference.StreamResult{Message: &message, Text: "answer"}, nil
				}
				run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "test terminal", Model: model})
				if err != nil {
					t.Fatal(err)
				}
				valid := stop == goai.StopReasonToolUse || (!calls && (stop == goai.StopReasonStop || stop == goai.StopReasonLength))
				expectedStatus := "failed"
				if valid {
					expectedStatus = "completed"
				}
				waitForCondition(t, 3*time.Second, func() bool { row, err := s.GetTurn(ctx, run.TurnID); return err == nil && row.Status == expectedStatus }, "terminal settlement")
				expectedEffects := int32(0)
				if valid && calls {
					expectedEffects = 1
				}
				if effects.Load() != expectedEffects {
					t.Fatal("invalid terminal entered tool host", effects.Load())
				}
				if !valid && requests.Load() != 1 {
					t.Fatal("invalid terminal retried or continued", requests.Load())
				}
				waitForCondition(t, time.Second, func() bool {
					r := e.runner("A")
					r.mu.Lock()
					defer r.mu.Unlock()
					_, _, err := s.GetSessionActiveTurn(ctx, "A")
					return err != nil
				}, "runner cleanup")
			})
		}
	}
}
