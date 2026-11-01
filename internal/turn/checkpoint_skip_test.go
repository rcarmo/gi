package turn

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func TestCheckpointSkipSettlesOriginalChildAndJoinsFinalHook(t *testing.T) {
	for _, mode := range []string{"direct", "partial-cleanup"} {
		t.Run(mode, func(t *testing.T) { testCheckpointSkipChild(t, mode == "partial-cleanup") })
	}
}

func testCheckpointSkipChild(t *testing.T, fault bool) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	e := New(s)
	defer e.Close()
	entered := make(chan struct{})
	var models, executors, ends atomic.Int32
	if err := e.RegisterTool(tools.RegisteredTool{Name: "skip-probe", Parameters: json.RawMessage(`{"type":"object"}`), ReplaySafe: true, Executor: func(ctx context.Context, rt tools.ToolRuntime, _ goai.ToolCall) (string, error) {
		executors.Add(1)
		rt.AddUsage(goai.Usage{Input: 11})
		rt.OnOutput("committed SDK prefix")
		close(entered)
		<-ctx.Done()
		return "unapproved late body", ctx.Err()
	}}); err != nil {
		t.Fatal(err)
	}
	withStreamWithToolsHookStub(t, func(context.Context, string, *goai.Context, func(map[string]any), *inference.StreamHooks) (*inference.StreamResult, error) {
		models.Add(1)
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonToolUse, Usage: &goai.Usage{Input: 3}, Content: []goai.ContentBlock{{Type: "toolCall", ID: "first", Name: "skip-probe", Arguments: map[string]any{}}, {Type: "toolCall", ID: "second", Name: "skip-probe", Arguments: map[string]any{}}}}}, nil
	})
	run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "original child", Model: "test/model"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("executor not entered")
	}
	if err = e.Close(); err != nil {
		t.Fatal(err)
	}
	// Park this proven closed frontier, retaining its original invocation and
	// physical progress. No SDK host survives the Close join barrier.
	_, token, err := s.GetSessionActiveTurn(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.MarkTurnFailureWithFallbackErr(ctx, nil, run.TurnID, "A", "recovery_interrupted_tool_phase", "review", "review required"); err != nil {
		t.Fatal(err)
	}
	if err = s.UpdateTurnStatusAndPhase(ctx, run.TurnID, "failed", "held_for_retry_or_skip"); err != nil {
		t.Fatal(err)
	}
	if err = s.AppendTurnEvent(ctx, run.TurnID, "A", "turn.recovered", map[string]any{"phase": "recovery", "checkpoint": true, "stale_claim": true, "previous_status": "running", "previous_phase": "waiting_on_tools", "recovery_disposition": "hold_for_retry_or_skip_after_tool_checkpoint"}); err != nil {
		t.Fatal(err)
	}
	if err = s.ReleaseSessionActiveTurn(ctx, "A", token); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateSession(ctx, "P", "P", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateTurnWithStatus(ctx, "parent", "P", "running", "parent", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateSubTurn(ctx, "parent", "P", run.TurnID, "A", "sync", 1, nil); err != nil {
		t.Fatal(err)
	}
	next := New(s)
	defer next.Close() // No tools or runtime-ready setup.
	if _, err = next.RetryHeldTurn(ctx, run.TurnID, "revive child"); !errors.Is(err, store.ErrFailureConflict) {
		t.Fatal("parent-linked child retried", err)
	}
	if err = next.runner("P").joinRequiredChildren(ctx, "parent", "completed", ""); err == nil {
		t.Fatal("unreceipted held child pronounced settled")
	}
	ending, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	if _, err = next.RegisterHook(HookAgentEnd, "end", func(context.Context, HookRequest) (HookResponse, error) {
		ends.Add(1)
		close(ending)
		<-release
		return HookResponse{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	runner := next.runner("A")
	runner.mu.Lock()
	if runner.hosts == nil {
		runner.hosts = map[string]chan struct{}{}
	}
	runner.hosts["previous-host"] = make(chan struct{})
	runner.mu.Unlock()
	busyErr := next.SkipHeldTurn(ctx, run.TurnID, "too early")
	runner.mu.Lock()
	delete(runner.hosts, "previous-host")
	runner.mu.Unlock()
	if !errors.Is(busyErr, store.ErrRetryPending) {
		t.Fatal("skip replaced an unjoined session host", busyErr)
	}
	if fault {
		if _, err = s.DB().Exec(`create trigger reject_skip_second before insert on turn_events when new.event_type='tool.interrupted' and json_extract(new.payload_json,'$.tool_call_id')='second' begin select raise(abort,'skip second rejected');end`); err != nil {
			t.Fatal(err)
		}
	}
	if err = next.SkipHeldTurn(ctx, run.TurnID, "operator skip"); err != nil {
		t.Fatal(err)
	}
	if fault {
		deadline := time.Now().Add(3 * time.Second)
		for {
			var receipts int
			if err = s.DB().QueryRow(`select count(*) from turn_receipts where turn_id=? and kind='tool'`, run.TurnID).Scan(&receipts); err != nil {
				t.Fatal(err)
			}
			runner := next.runner("A")
			runner.mu.Lock()
			live := runner.hosts[run.TurnID] != nil
			runner.mu.Unlock()
			if receipts == 1 && !live {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("partial skip cleanup did not retire", receipts, live)
			}
			time.Sleep(time.Millisecond)
		}
		if ends.Load() != 0 {
			t.Fatal("agent end preceded pending slot")
		}
		if _, err = s.DB().Exec(`drop trigger reject_skip_second`); err != nil {
			t.Fatal(err)
		}
		if _, err = s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z' where session_id='A'`); err != nil {
			t.Fatal(err)
		}
		if err = next.ResumeInterruptedWork(ctx); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case <-ending:
	case <-time.After(3 * time.Second):
		t.Fatal("skip did not reach terminal cleanup")
	}
	terminal, err := s.HasTurnTerminalReceipt(ctx, run.TurnID)
	if err != nil || !terminal {
		t.Fatal("skip failed to settle child", terminal, err)
	}
	joined := make(chan error, 1)
	go func() { joined <- next.runner("P").joinRequiredChildren(ctx, "parent", "completed", "") }()
	select {
	case err := <-joined:
		t.Fatal("parent crossed child final hook", err)
	case <-time.After(20 * time.Millisecond):
	}
	unblock()
	select {
	case err := <-joined:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("parent did not join skipped child")
	}
	rec, err := s.GetTurn(ctx, run.TurnID)
	if err != nil || rec.Status != "failed" || rec.Phase != "failed" {
		t.Fatal(rec, err)
	}
	failure, err := s.GetTurnFailure(ctx, run.TurnID)
	if err != nil || failure.ResolutionState != "skipped" || failure.HoldState != "none" {
		t.Fatal(failure, err)
	}
	accounting, err := s.TurnAccounting(ctx, run.TurnID)
	if err != nil || accounting.Usage.Input != 14 || accounting.Unknown != 1 {
		t.Fatal(accounting, err)
	}
	var results, cancelled, prompts int
	if err = s.DB().QueryRow(`select count(*),coalesce(sum(coalesce(json_extract(payload_json,'$.cancelled'),0)),0) from messages where session_id='A' and json_extract(payload_json,'$.kind')='tool_result'`).Scan(&results, &cancelled); err != nil || results != 2 || cancelled != 0 {
		t.Fatal("skip lost wire calls or fabricated cancellation", results, cancelled, err)
	}
	if err = s.DB().QueryRow(`select count(*) from turns where session_id='A'`).Scan(&prompts); err != nil || prompts != 1 || models.Load() != 1 || executors.Load() != 1 || ends.Load() != 1 {
		t.Fatal("skip replayed work", prompts, models.Load(), executors.Load(), ends.Load(), err)
	}
	if _, err = next.RetryHeldTurn(ctx, run.TurnID, "revive consumed child"); err == nil {
		t.Fatal("consumed skipped child revived")
	}
}
