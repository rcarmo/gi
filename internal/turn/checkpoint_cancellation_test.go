package turn

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func TestCheckpointCancellationDoesNotDispatchAndRecoversPartialCleanup(t *testing.T) {
	for _, mode := range []string{"direct", "partial-cleanup", "required-child"} {
		t.Run(mode, func(t *testing.T) {
			fault, withChild := mode == "partial-cleanup", mode == "required-child"
			s := openTestStore(t)
			defer s.Close()
			ctx := context.Background()
			if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
				t.Fatal(err)
			}
			e := New(s)
			defer e.Close()
			entered := make(chan struct{}, 2)
			var models, toolsRun, starts, ends atomic.Int32
			if _, err := e.RegisterHook(HookAgentStart, "start", func(context.Context, HookRequest) (HookResponse, error) { starts.Add(1); return HookResponse{}, nil }); err != nil {
				t.Fatal(err)
			}
			if err := e.RegisterTool(tools.RegisteredTool{Name: "cancel-probe", Parameters: json.RawMessage(`{"type":"object"}`), ReplaySafe: true, Executor: func(ctx context.Context, rt tools.ToolRuntime, _ goai.ToolCall) (string, error) {
				toolsRun.Add(1)
				rt.AddUsage(goai.Usage{Input: 11})
				rt.OnOutput("committed SDK prefix")
				entered <- struct{}{}
				<-ctx.Done()
				return "unapproved cancelled body", ctx.Err()
			}}); err != nil {
				t.Fatal(err)
			}
			withStreamWithToolsHookStub(t, func(context.Context, string, *goai.Context, func(map[string]any), *inference.StreamHooks) (*inference.StreamResult, error) {
				models.Add(1)
				return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonToolUse, Usage: &goai.Usage{Input: 3}, Content: []goai.ContentBlock{
					{Type: "toolCall", ID: "first", Name: "cancel-probe", Arguments: map[string]any{}},
					{Type: "toolCall", ID: "second", Name: "cancel-probe", Arguments: map[string]any{}},
					{Type: "toolCall", ID: "third", Name: "cancel-probe", Arguments: map[string]any{}},
				}}}, nil
			})
			run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "cancel original checkpoint", Model: "test/model"})
			if err != nil {
				t.Fatal(err)
			}
			select {
			case <-entered:
			case <-time.After(3 * time.Second):
				t.Fatal("executor not entered")
			}
			var child *SubmitResult
			if withChild {
				if _, err = s.CreateSession(ctx, "B", "B", nil); err != nil {
					t.Fatal(err)
				}
				child, err = e.SubmitPrompt(ctx, RunInput{SessionID: "B", Prompt: "required child", Model: "test/model", ParentTurnID: run.TurnID, Metadata: map[string]any{"subturn_critical": true}})
				if err != nil {
					t.Fatal(err)
				}
				select {
				case <-entered:
				case <-time.After(3 * time.Second):
					t.Fatal("child executor not entered")
				}
			}
			if err = e.Close(); err != nil {
				t.Fatal(err)
			}
			cancelEngine := New(s) // No tool registration or runtime-ready call.
			defer cancelEngine.Close()
			childEnding, releaseChildEnd := make(chan struct{}), make(chan struct{})
			var releaseOnce sync.Once
			unblockChild := func() { releaseOnce.Do(func() { close(releaseChildEnd) }) }
			defer unblockChild()
			if _, err = cancelEngine.RegisterHook(HookAgentEnd, "end", func(_ context.Context, req HookRequest) (HookResponse, error) {
				ends.Add(1)
				if withChild && req.TurnID == child.TurnID {
					close(childEnding)
					<-releaseChildEnd
				}
				return HookResponse{}, nil
			}); err != nil {
				t.Fatal(err)
			}
			if fault {
				if _, err = s.DB().Exec(`create trigger reject_second before insert on turn_events when new.event_type='tool.interrupted' and json_extract(new.payload_json,'$.tool_call_id')='second' begin select raise(abort,'second rejected');end`); err != nil {
					t.Fatal(err)
				}
			}
			if err = cancelEngine.CancelQueuedTurn(ctx, "A", run.TurnID); err != nil {
				t.Fatal(err)
			}
			if withChild {
				select {
				case <-childEnding:
				case <-time.After(3 * time.Second):
					t.Fatal("child terminal hook not entered")
				}
				parent, err := s.GetTurn(ctx, run.TurnID)
				if err != nil || isTerminalSubTurnStatus(parent.Status) || parent.FinishedAt != "" {
					t.Fatal("parent outran child cleanup", parent, err)
				}
				unblockChild()
			}
			cancelEngine.runs.Wait()
			if fault {
				row, getErr := s.GetTurn(ctx, run.TurnID)
				if getErr != nil || row.Status != "cancelling" || row.Phase != "settling" || ends.Load() != 0 {
					t.Fatal(row, getErr, ends.Load())
				}
				var receipts int
				if err = s.DB().QueryRow(`select count(*) from turn_receipts where turn_id=? and kind='tool'`, run.TurnID).Scan(&receipts); err != nil || receipts != 1 {
					t.Fatal(receipts, err)
				}
				if _, err = s.DB().Exec(`drop trigger reject_second;update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
					t.Fatal(err)
				}
				if err = cancelEngine.ResumeInterruptedWork(ctx); err != nil {
					t.Fatal(err)
				}
				cancelEngine.runs.Wait()
			}
			row, err := s.GetTurn(ctx, run.TurnID)
			want := int32(1)
			if withChild {
				want = 2
			}
			if err != nil || row.Status != "cancelled" || row.FinishedAt == "" || models.Load() != want || toolsRun.Load() != want || starts.Load() != want || ends.Load() != want {
				t.Fatal(row, err, models.Load(), toolsRun.Load(), starts.Load(), ends.Load())
			}
			accounting, err := s.TurnAccounting(ctx, run.TurnID)
			if err != nil || accounting.Usage.Input != 14 || accounting.Unknown != 1 {
				t.Fatal(accounting, err)
			}
			for _, id := range []string{"first", "second", "third"} {
				var results int
				if err = s.DB().QueryRow(`select count(*) from messages where session_id='A' and role='tool_result' and json_extract(payload_json,'$.tool_call_id')=?`, id).Scan(&results); err != nil || results != 1 {
					t.Fatal("dangling/duplicated tool call", id, results, err)
				}
			}
		})
	}
}
