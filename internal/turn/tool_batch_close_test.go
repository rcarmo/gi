package turn

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func TestToolBatchCloseJoinsThenResumesWithoutLifecycleReplay(t *testing.T) {
	for _, mode := range []string{"execute", "replay", "prepare", "result", "commit-fault", "output-fault", "usage-fault"} {
		t.Run(mode, func(t *testing.T) {
			s := openTestStore(t)
			defer s.Close()
			ctx := context.Background()
			fault := strings.HasSuffix(mode, "-fault")
			if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
				t.Fatal(err)
			}
			entered, release := make(chan struct{}), make(chan struct{})
			retained := make(chan tools.ToolRuntime, 1)
			var once sync.Once
			unblock := func() { once.Do(func() { close(release) }) }
			e := NewWithRuntimeConfig(s, config.RuntimeConfig{DefaultModel: "test/model"}, "")
			defer e.Close()
			defer unblock()
			var requests, starts, ends, executions, resultHooks atomic.Int32
			registerHooks := func(engine *Engine) {
				t.Helper()
				for name, counter := range map[string]*atomic.Int32{HookAgentStart: &starts, HookAgentEnd: &ends} {
					if _, err := engine.RegisterHook(name, name, func(context.Context, HookRequest) (HookResponse, error) { counter.Add(1); return HookResponse{}, nil }); err != nil {
						t.Fatal(err)
					}
				}
			}
			registerHooks(e)
			if mode == "prepare" || mode == "result" {
				name := HookToolCall
				if mode == "result" {
					name = HookToolResult
				}
				if _, err := e.RegisterHook(name, "blocked-policy", func(context.Context, HookRequest) (HookResponse, error) {
					if mode == "result" {
						resultHooks.Add(1)
					}
					close(entered)
					<-release
					return HookResponse{}, nil
				}); err != nil {
					t.Fatal(err)
				}
			}
			tool := tools.RegisteredTool{Name: "close-probe", Parameters: json.RawMessage(`{"type":"object","properties":{}}`), ReplaySafe: true,
				Executor: func(_ context.Context, rt tools.ToolRuntime, _ goai.ToolCall) (string, error) {
					executions.Add(1)
					retained <- rt
					rt.OnOutput("committed SDK prefix")
					rt.AddUsage(goai.Usage{Input: 11})
					if mode != "result" {
						rt.OnOutput("coalesced SDK tail before cancellation")
						close(entered)
						<-release // Intentionally ignore cancellation; Close must join us.
						rt.AddUsage(goai.Usage{Input: 7})
						if err := rt.OnOutput("last SDK report after cancellation"); err != nil {
							return "", err
						}
					}
					return "unapproved callback body", nil
				}}
			if err := e.RegisterTool(tool); err != nil {
				t.Fatal(err)
			}
			withStreamWithToolsHookStub(t, func(_ context.Context, _ string, conv *goai.Context, _ func(map[string]any), _ *inference.StreamHooks) (*inference.StreamResult, error) {
				if requests.Add(1) == 1 {
					return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonToolUse, Usage: &goai.Usage{Input: 3}, Content: []goai.ContentBlock{{Type: "toolCall", ID: "wire-close", Name: "close-probe", Arguments: map[string]any{}}}}}, nil
				}
				raw, err := json.Marshal(conv)
				if err != nil {
					return nil, err
				}
				if strings.Contains(string(raw), "unapproved callback body") {
					return nil, errors.New("unapproved result entered model context")
				}
				foundError := false
				for _, message := range conv.Messages {
					if message.Role == goai.RoleToolResult && message.ToolCallID == "wire-close" && message.IsError {
						foundError = true
					}
				}
				if mode == "replay" && (!strings.Contains(string(raw), "fresh replay result") || strings.Contains(string(raw), "last SDK report after cancellation")) {
					return nil, errors.New("replay context did not isolate old physical execution")
				}
				if mode != "replay" && !foundError {
					return nil, errors.New("missing interrupted tool result")
				}
				if mode == "execute" || mode == "commit-fault" {
					if !strings.Contains(string(raw), "last SDK report after cancellation") {
						return nil, errors.New("closing recorder lost SDK tail")
					}
				}
				return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Usage: &goai.Usage{Input: 7}, Content: []goai.ContentBlock{{Type: "text", Text: "continued"}}}, Text: "continued"}, nil
			})
			run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "one original prompt", Model: "test/model"})
			if err != nil {
				t.Fatal(err)
			}
			select {
			case <-entered:
			case <-time.After(3 * time.Second):
				row, _ := s.GetTurn(ctx, run.TurnID)
				t.Fatalf("tool boundary not reached: turn=%+v requests=%d", row, requests.Load())
			}
			_, old, err := s.GetSessionActiveTurn(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			if mode == "commit-fault" {
				if _, err = s.DB().Exec(`create trigger reject_tool_close before update on turns when new.phase='checkpoint_tool_batch' begin select raise(abort,'reject close');end`); err != nil {
					t.Fatal(err)
				}
			} else if fault {
				kind := "tool.output"
				if mode == "usage-fault" {
					kind = "tool.usage_received"
				}
				if _, err = s.DB().Exec(`create trigger reject_tool_close before insert on turn_events when new.event_type='` + kind + `' begin select raise(abort,'reject snapshot');end`); err != nil {
					t.Fatal(err)
				}
			}
			closed, duplicate := make(chan error, 1), make(chan error, 1)
			go func() { closed <- e.Close() }()
			go func() { duplicate <- e.Close() }()
			defer func() {
				if t.Failed() {
					events, _ := s.ListTurnEvents(ctx, run.TurnID)
					for _, ev := range events {
						t.Logf("event %d: %s", ev.Seq, ev.Type)
					}
				}
			}()
			waitForCondition(t, 3*time.Second, func() bool {
				var worker string
				if s.DB().QueryRow(`select worker_id from session_active_turns where session_id='A'`).Scan(&worker) == nil && worker == "tool-closing" {
					return true
				}
				select {
				case err := <-closed:
					t.Fatalf("Close returned before reservation: %v", err)
				default:
				}
				return false
			}, "tool close reservation")
			select {
			case err := <-closed:
				t.Fatalf("Close did not join callback: %v", err)
			default:
			}
			row, err := s.GetTurn(ctx, run.TurnID)
			if err != nil || row.Status != "running" || row.Phase != "waiting_on_tools" || row.FinishedAt != "" || ends.Load() != 0 {
				t.Fatal(row, err, ends.Load())
			}
			oldCtx := store.WithTurnWriter(ctx, "A", run.TurnID, old)
			if err = s.AdmitTurnEffect(oldCtx, "A", run.TurnID, "running", "inference.started", map[string]any{"request_id": "stale"}); err == nil {
				t.Fatalf("old writer not fenced: %v", err)
			}
			unblock()
			for _, done := range []chan error{closed, duplicate} {
				select {
				case err := <-done:
					if (err != nil) != fault {
						t.Fatal(err)
					}
				case <-time.After(4 * time.Second):
					t.Fatal("Close failed to join")
				}
			}
			row, err = s.GetTurn(ctx, run.TurnID)
			if err != nil || row.FinishedAt != "" || ends.Load() != 0 {
				t.Fatal(row, err)
			}
			if fault {
				if row.Status != "running" {
					t.Fatal(row)
				}
				if _, err = s.DB().Exec(`drop trigger reject_tool_close;update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
					t.Fatal(err)
				}
			} else if row.Status != "queued" || row.Phase != store.ClosedToolBatchPhase {
				t.Fatal(row)
			}
			// Escaped recorders must not revive authority after the joined host.
			if mode != "prepare" {
				rt := <-retained
				before, _ := s.ListTurnEvents(ctx, run.TurnID)
				rt.OnOutput("escaped output")
				rt.AddUsage(goai.Usage{Input: 99})
				after, _ := s.ListTurnEvents(ctx, run.TurnID)
				if len(after) != len(before) {
					t.Fatal("escaped recorder mutated joined checkpoint")
				}
			}
			resumed := NewWithRuntimeConfig(s, config.RuntimeConfig{DefaultModel: "test/model"}, "")
			defer resumed.Close()
			registerHooks(resumed)
			tool.ReplaySafe = mode == "replay" // Both registrations must permit replay.
			tool.Executor = func(_ context.Context, rt tools.ToolRuntime, _ goai.ToolCall) (string, error) {
				executions.Add(1)
				if mode == "replay" {
					rt.AddUsage(goai.Usage{Input: 5})
					return "fresh replay result", nil
				}
				return "", errors.New("returned or preparing callback must not repeat")
			}
			if err = resumed.RegisterTool(tool); err != nil {
				t.Fatal(err)
			}
			if _, err = resumed.RegisterHook(HookToolResult, "no-repeat", func(context.Context, HookRequest) (HookResponse, error) {
				resultHooks.Add(1)
				return HookResponse{}, nil
			}); err != nil {
				t.Fatal(err)
			}
			if requests.Load() != 1 {
				t.Fatal("dispatched before runtime ready")
			}
			if err = resumed.ResumeInterruptedWork(ctx); err != nil {
				t.Fatal(err)
			}
			waitRetryDone(t, s, run.TurnID)
			resumed.runs.Wait()
			row, err = s.GetTurn(ctx, run.TurnID)
			if err != nil || row.Status != "completed" || starts.Load() != 1 || ends.Load() != 1 || requests.Load() != 2 {
				t.Fatal(row, err, starts.Load(), ends.Load(), requests.Load())
			}
			wantExecutions, wantUsage, wantUnknown, wantHooks := int32(1), int64(28), 1, int32(0)
			if mode == "prepare" {
				wantExecutions, wantUsage, wantUnknown = 0, 10, 0
			}
			if mode == "replay" {
				wantExecutions, wantUsage, wantHooks = 2, 33, 1
			}
			if mode == "result" {
				wantUsage, wantUnknown, wantHooks = 21, 0, 1
			}
			if mode == "usage-fault" {
				wantUsage = 21
			}
			accounting, err := s.TurnAccounting(ctx, run.TurnID)
			if err != nil || int64(accounting.Usage.Input) != wantUsage || accounting.Unknown != wantUnknown || executions.Load() != wantExecutions || resultHooks.Load() != wantHooks {
				t.Fatal(accounting, err, executions.Load(), resultHooks.Load())
			}
		})
	}
}
