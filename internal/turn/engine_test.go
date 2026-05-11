package turn

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/routing"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
	goai "github.com/rcarmo/go-ai"
)

func openTestStore(t *testing.T) *store.Store {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	s, err := store.Open(dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return s
}

func withStreamWithToolsStub(t *testing.T, stub func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error)) {
	t.Helper()
	withStreamWithToolsHookStub(t, func(ctx context.Context, model string, convCtx *goai.Context, cb func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
		return stub(ctx, model, convCtx, cb)
	})
}

func withStreamWithToolsHookStub(t *testing.T, stub func(context.Context, string, *goai.Context, func(map[string]any), *inference.StreamHooks) (*inference.StreamResult, error)) {
	t.Helper()
	original := streamWithToolsWithHooks
	streamWithToolsWithHooks = stub
	t.Cleanup(func() {
		streamWithToolsWithHooks = original
	})
}

func waitForCondition(t *testing.T, timeout time.Duration, check func() bool, label string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", label)
}

func TestSubmitPromptSteersSecondPromptToActiveTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_1", "Test", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)

	first, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_1", Prompt: "one", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit first: %v", err)
	}
	second, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_1", Prompt: "two", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit second: %v", err)
	}
	if first.Queued {
		t.Fatalf("first should not be queued: %#v", first)
	}
	if second.Queued || second.Status != "running" {
		t.Fatalf("second should be accepted as steering against the active turn: %#v", second)
	}
	if second.TurnID != first.TurnID {
		t.Fatalf("expected steering to target active turn %s, got %#v", first.TurnID, second)
	}
	sess, err := s.GetSession(ctx, "session_1")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if got := sess.State["queue_count"]; got != float64(0) && got != 0 {
		t.Fatalf("expected queue_count=0 while steering is queued separately, got %#v", got)
	}
	if depth, err := s.SteeringQueueLength(ctx, "session_1"); err != nil {
		t.Fatalf("steering queue length: %v", err)
	} else if depth != 1 {
		t.Fatalf("expected steering queue depth 1, got %d", depth)
	}
	activeTurnID, _, err := s.GetSessionActiveTurn(ctx, "session_1")
	if err != nil {
		t.Fatalf("get active turn: %v", err)
	}
	if activeTurnID != first.TurnID {
		t.Fatalf("expected active turn %s, got %s", first.TurnID, activeTurnID)
	}

	time.Sleep(2500 * time.Millisecond)
	turns, err := s.ListTurns(ctx, "session_1")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected 2 turns, got %d", len(turns))
	}
	if turns[0].Status != "completed" || turns[1].Status != "completed" {
		t.Fatalf("unexpected turn statuses: %#v", turns)
	}
	if turns[0].Phase != "completed" || turns[1].Phase != "completed" {
		t.Fatalf("unexpected turn phases: %#v", turns)
	}
	sess, err = s.GetSession(ctx, "session_1")
	if err != nil {
		t.Fatalf("get session after completion: %v", err)
	}
	if got := sess.State["queue_count"]; got != float64(0) && got != 0 {
		t.Fatalf("expected queue_count=0 after completion, got %#v", got)
	}
	if _, _, err := s.GetSessionActiveTurn(ctx, "session_1"); err == nil {
		t.Fatal("expected no active turn after completion")
	}
}

func TestSubmitPromptSurvivesCallerCancellationAfterTurnPersisted(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx, cancel := context.WithCancel(context.Background())
	if _, err := s.CreateSession(ctx, "session_submit_cancel", "Test", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	cancelled := false
	engine.beforeLaunchClaimHook = func(ctx context.Context, sessionID, turnID string) {
		if !cancelled {
			cancelled = true
			cancel()
		}
	}
	res, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_submit_cancel", Prompt: "hello", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit prompt should survive caller cancellation after persistence: %v", err)
	}
	if res == nil || res.TurnID == "" {
		t.Fatalf("expected submit result despite caller cancellation, got %#v", res)
	}
	turnRec, err := s.GetTurn(context.Background(), res.TurnID)
	if err != nil {
		t.Fatalf("get persisted turn: %v", err)
	}
	if turnRec.Status == "" {
		t.Fatalf("expected persisted turn status, got %#v", turnRec)
	}
}

func TestSubmitPromptRollsBackTurnWhenCreateSubTurnFails(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_subturn_fail", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_subturn_fail", "Child", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	parentTurn, err := s.CreateTurnWithStatus(ctx, "turn_parent_subturn_fail", "session_parent_subturn_fail", "completed", "parent", map[string]any{"intent": "prompt", "model": "bootstrap", "subturn_depth": 0, "effective_tools": []string{"read"}})
	if err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	engine := New(s)
	engine.beforeCreateSubTurnErrorHook = func(ctx context.Context, parentTurnID, childTurnID string) error {
		return fmt.Errorf("boom create subturn")
	}
	_, err = engine.SubmitPrompt(ctx, RunInput{SessionID: "session_child_subturn_fail", Prompt: "child", Model: "bootstrap", ParentTurnID: parentTurn.ID})
	if err == nil || !strings.Contains(err.Error(), "boom create subturn") {
		t.Fatalf("expected create subturn error, got %v", err)
	}
	turns, err := s.ListTurns(ctx, "session_child_subturn_fail")
	if err != nil {
		t.Fatalf("list child turns: %v", err)
	}
	if len(turns) != 0 {
		t.Fatalf("expected rollback to remove child turn after create subturn failure, got %#v", turns)
	}
}

func TestConcurrentSubmitAcrossEnginesCompletesWithoutConflict(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_1", "Test", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	engineA := New(s)
	engineB := New(s)

	results := make(chan *SubmitResult, 2)
	errs := make(chan error, 2)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i, engine := range []*Engine{engineA, engineB} {
		wg.Add(1)
		go func(idx int, eng *Engine) {
			defer wg.Done()
			<-start
			res, err := eng.SubmitPrompt(ctx, RunInput{SessionID: "session_1", Prompt: fmt.Sprintf("prompt-%d", idx+1), Model: "bootstrap"})
			if err != nil {
				errs <- err
				return
			}
			results <- res
		}(i, engine)
	}
	close(start)
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected submit error: %v", err)
		}
	}
	var submitted []*SubmitResult
	for res := range results {
		submitted = append(submitted, res)
	}
	if len(submitted) != 2 {
		t.Fatalf("expected 2 submit results, got %d", len(submitted))
	}

	time.Sleep(2500 * time.Millisecond)
	turns, err := s.ListTurns(ctx, "session_1")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected 2 turns, got %d", len(turns))
	}
	for _, turnRec := range turns {
		if turnRec.Status != "completed" {
			t.Fatalf("expected completed turns, got %#v", turns)
		}
	}
}

func TestSubmitPromptClaimConflictConvertsFreshTurnToSteering(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_claim_conflict", "Test", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_existing_active", "session_claim_conflict", "running", "already running", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create existing active turn: %v", err)
	}
	engine := New(s)
	engine.beforeLaunchClaimHook = func(ctx context.Context, sessionID, turnID string) {
		engine.beforeLaunchClaimHook = nil
		if _, err := s.ClaimSessionActiveTurn(ctx, sessionID, activeTurn.ID, "runner", activeTurn.ID); err != nil {
			t.Fatalf("claim existing active turn: %v", err)
		}
		if err := s.TouchSessionState(ctx, sessionID, map[string]any{"active_turn_id": activeTurn.ID, "status": "running"}); err != nil {
			t.Fatalf("touch session state: %v", err)
		}
	}
	res, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_claim_conflict", Prompt: "steer me", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit prompt: %v", err)
	}
	if res.Queued || res.Status != "running" || res.TurnID != activeTurn.ID {
		t.Fatalf("expected claim-conflict submit to steer to existing active turn, got %#v", res)
	}
	turns, err := s.ListTurns(ctx, "session_claim_conflict")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 1 || turns[0].ID != activeTurn.ID {
		t.Fatalf("expected transient queued turn to be removed after steering fallback, got %#v", turns)
	}
	if depth, err := s.SteeringQueueLength(ctx, "session_claim_conflict"); err != nil {
		t.Fatalf("steering queue length: %v", err)
	} else if depth != 1 {
		t.Fatalf("expected steering queue depth 1 after claim-conflict fallback, got %d", depth)
	}
}

func TestCleanupTurnRunDoesNotClearNewerRunningTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_cleanup_guard", "Test", map[string]any{"model": "bootstrap", "status": "running"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	runner := engine.runner("session_cleanup_guard")
	oldRunning := &runningTurn{turnID: "turn_old"}
	newRunning := &runningTurn{turnID: "turn_new"}
	runner.current = newRunning
	runner.cleanupTurnRun("session_cleanup_guard", "turn_old", oldRunning)
	if runner.current != newRunning {
		t.Fatalf("expected cleanup for older turn not to clear newer running turn, got %#v", runner.current)
	}
}

func TestClaimConflictFallsBackToQueuedTurnWhenSteeringIsFull(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_claim_conflict_full", "Test", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_existing_active_full", "session_claim_conflict_full", "running", "already running", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create existing active turn: %v", err)
	}
	for i := 0; i < 10; i++ {
		if _, err := s.EnqueueSteering(ctx, "session_claim_conflict_full", activeTurn.ID, "user", fmt.Sprintf("queued-%d", i), map[string]any{"intent": "prompt", "model": "bootstrap"}, nil, "one-at-a-time"); err != nil {
			t.Fatalf("enqueue steering %d: %v", i, err)
		}
	}
	engine := New(s)
	topicCtx, topicCancel := context.WithCancel(ctx)
	defer topicCancel()
	ch, unsub := engine.Topics().Subscribe(topicCtx, "session.steering", topics.SubscribeOptions{Buffer: 4, SessionID: "session_claim_conflict_full"})
	defer unsub()
	engine.beforeLaunchClaimHook = func(ctx context.Context, sessionID, turnID string) {
		engine.beforeLaunchClaimHook = nil
		if _, err := s.ClaimSessionActiveTurn(ctx, sessionID, activeTurn.ID, "runner", activeTurn.ID); err != nil {
			t.Fatalf("claim existing active turn: %v", err)
		}
		if err := s.TouchSessionState(ctx, sessionID, map[string]any{"active_turn_id": activeTurn.ID, "status": "running"}); err != nil {
			t.Fatalf("touch session state: %v", err)
		}
	}
	res, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_claim_conflict_full", Prompt: "preserve me", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit prompt: %v", err)
	}
	if !res.Queued || res.Status != "queued" {
		t.Fatalf("expected queued fallback when steering is full, got %#v", res)
	}
	select {
	case env := <-ch:
		if gotType, _ := env.Payload["type"].(string); gotType != "steering_rejected" {
			t.Fatalf("expected steering_rejected topic payload, got %#v", env)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected steering_rejected topic event")
	}
	turns, err := s.ListTurns(ctx, "session_claim_conflict_full")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected active turn plus preserved queued fallback turn, got %#v", turns)
	}
	if depth, err := s.SteeringQueueLength(ctx, "session_claim_conflict_full"); err != nil {
		t.Fatalf("steering queue length: %v", err)
	} else if depth != 10 {
		t.Fatalf("expected steering queue to remain full at 10, got %d", depth)
	}
}

func TestStartupRecoveryRequeuesCompactingTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_recover_compact", "Test", map[string]any{"model": "bootstrap", "status": "running"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_recover_compact", "session_recover_compact", "running", "hello", map[string]any{"intent": "prompt"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, turnRec.ID, "running", "compacting"); err != nil {
		t.Fatalf("set compacting phase: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_recover_compact", turnRec.ID, "runner", turnRec.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update session_active_turns set updated_at = '2000-01-01T00:00:00Z' where session_id = ?`, "session_recover_compact"); err != nil {
		t.Fatalf("age active turn claim: %v", err)
	}

	_ = New(s)

	recovered, err := s.GetTurn(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("get recovered turn: %v", err)
	}
	if recovered.Status != "queued" || recovered.Phase != "queued" {
		t.Fatalf("expected queued recovered turn, got %#v", recovered)
	}
	if _, _, err := s.GetSessionActiveTurn(ctx, "session_recover_compact"); err == nil {
		t.Fatal("expected stale active claim to be released")
	}
	events, err := s.ListTurnEvents(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("list recovery events: %v", err)
	}
	found := false
	for _, event := range events {
		if event.Type == "turn.recovered" && event.Payload["recovery_disposition"] == "requeue_after_compaction_checkpoint" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected compaction recovery event, got %#v", events)
	}
}

func TestStartupRecoveryHoldsToolPhaseTurnForReview(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_recover_tool", "Test", map[string]any{"model": "bootstrap", "status": "running"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_recover_tool", "session_recover_tool", "running", "hello", map[string]any{"intent": "prompt"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, turnRec.ID, "running", "waiting_on_tools"); err != nil {
		t.Fatalf("set waiting_on_tools phase: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_recover_tool", turnRec.ID, "runner", turnRec.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update session_active_turns set updated_at = '2000-01-01T00:00:00Z' where session_id = ?`, "session_recover_tool"); err != nil {
		t.Fatalf("age active turn claim: %v", err)
	}

	_ = New(s)

	recovered, err := s.GetTurn(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("get recovered turn: %v", err)
	}
	if recovered.Status != "failed" || recovered.Phase != "held_for_retry_or_skip" {
		t.Fatalf("expected held recovered turn, got %#v", recovered)
	}
	if recovered.FinishedAt == "" {
		t.Fatalf("expected finished_at on failed recovered turn, got %#v", recovered)
	}
	failureRec, err := s.GetTurnFailure(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("expected recovery failure marker: %v", err)
	}
	if failureRec.FailureKind != "recovery_interrupted_tool_phase" || failureRec.HoldState != "review" {
		t.Fatalf("unexpected recovery failure marker: %#v", failureRec)
	}
	if _, _, err := s.GetSessionActiveTurn(ctx, "session_recover_tool"); err == nil {
		t.Fatal("expected stale active claim to be released")
	}
}

func TestStageQueuedSteeringContinuationCreatesQueuedTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_stage_steering", "Test", map[string]any{"model": "bootstrap", "status": "running"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_stage_active", "session_stage_steering", "running", "active", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create active turn: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_stage_steering", activeTurn.ID, "runner", activeTurn.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_stage_steering", activeTurn.ID, "user", "late steer", map[string]any{"intent": "prompt", "model": "bootstrap"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}
	engine := New(s)
	staged, stagedTurnID, err := engine.stageQueuedSteeringContinuation(ctx, "session_stage_steering")
	if err != nil {
		t.Fatalf("stage queued steering continuation: %v", err)
	}
	if !staged || stagedTurnID == "" {
		t.Fatalf("expected staged continuation turn, got staged=%v id=%q", staged, stagedTurnID)
	}
	if depth, err := s.SteeringQueueLength(ctx, "session_stage_steering"); err != nil {
		t.Fatalf("steering queue length: %v", err)
	} else if depth != 0 {
		t.Fatalf("expected steering queue to be drained, got %d", depth)
	}
	stagedTurn, err := s.GetTurn(ctx, stagedTurnID)
	if err != nil {
		t.Fatalf("get staged turn: %v", err)
	}
	if stagedTurn.Status != "queued" || stagedTurn.Phase != "queued" {
		t.Fatalf("expected queued staged turn, got %#v", stagedTurn)
	}
	if steeringMessages := steeringMessagesFromMetadata(stagedTurn.Metadata); len(steeringMessages) != 1 || steeringMessages[0].Content != "late steer" {
		t.Fatalf("expected staged steering metadata, got %#v", stagedTurn.Metadata)
	}
}

func TestStageQueuedSteeringContinuationDoesNotDrainWhenQueuedTurnAlreadyExists(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_stage_steering_busy", "Test", map[string]any{"model": "bootstrap", "status": "running"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_stage_active_busy", "session_stage_steering_busy", "running", "active", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create active turn: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_stage_steering_busy", activeTurn.ID, "runner", activeTurn.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_stage_existing_queue", "session_stage_steering_busy", "queued", "already queued", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_stage_steering_busy", activeTurn.ID, "user", "late steer", map[string]any{"intent": "prompt", "model": "bootstrap"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}
	engine := New(s)
	staged, stagedTurnID, err := engine.stageQueuedSteeringContinuation(ctx, "session_stage_steering_busy")
	if err != nil {
		t.Fatalf("stage queued steering continuation: %v", err)
	}
	if staged || stagedTurnID != "" {
		t.Fatalf("expected no staged continuation while a queued turn already exists, got staged=%v id=%q", staged, stagedTurnID)
	}
	if depth, err := s.SteeringQueueLength(ctx, "session_stage_steering_busy"); err != nil {
		t.Fatalf("steering queue length: %v", err)
	} else if depth != 1 {
		t.Fatalf("expected steering queue to remain queued, got %d", depth)
	}
	turns, err := s.ListTurns(ctx, "session_stage_steering_busy")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected only active+queued turns, got %#v", turns)
	}
}

func TestContinueSessionStartsQueuedSteeringWhenIdle(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_continue", "Test", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_continue", "", "user", "continue please", map[string]any{"intent": "prompt", "model": "bootstrap"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}
	engine := New(s)
	continued, err := engine.ContinueSession(ctx, "session_continue")
	if err != nil {
		t.Fatalf("continue session: %v", err)
	}
	if !continued {
		t.Fatal("expected ContinueSession to start queued steering")
	}
	time.Sleep(1500 * time.Millisecond)
	turns, err := s.ListTurns(ctx, "session_continue")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected one continuation turn, got %#v", turns)
	}
	msgs, err := s.ListMessages(ctx, "session_continue")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	found := false
	for _, msg := range msgs {
		if msg.Role == "user" && msg.Content == "continue please" && msg.Payload["steering"] == true {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected continued steering message in history, got %#v", msgs)
	}
}

func TestConcurrentContinueSessionStartsSingleTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_continue_concurrent", "Test", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_continue_concurrent", "", "user", "continue please", map[string]any{"intent": "prompt", "model": "bootstrap"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}
	engineA := New(s)
	engineB := New(s)
	start := make(chan struct{})
	results := make(chan bool, 2)
	errCh := make(chan error, 2)
	for _, engine := range []*Engine{engineA, engineB} {
		go func(eng *Engine) {
			<-start
			continued, err := eng.ContinueSession(ctx, "session_continue_concurrent")
			if err != nil {
				errCh <- err
				return
			}
			results <- continued
		}(engine)
	}
	close(start)
	for i := 0; i < 2; i++ {
		select {
		case err := <-errCh:
			t.Fatalf("continue session: %v", err)
		case <-results:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for continue results")
		}
	}
	time.Sleep(1500 * time.Millisecond)
	turns, err := s.ListTurns(ctx, "session_continue_concurrent")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected one continuation turn after concurrent continue, got %#v", turns)
	}
	if depth, err := s.SteeringQueueLength(ctx, "session_continue_concurrent"); err != nil {
		t.Fatalf("steering queue length: %v", err)
	} else if depth != 0 {
		t.Fatalf("expected steering queue drained after concurrent continue, got %d", depth)
	}
}

func TestCleanupSchedulesContinuationBeforeConcurrentSubmit(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_cleanup_continue", "Test", map[string]any{"model": "bootstrap", "status": "running"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_cleanup_active", "session_cleanup_continue", "running", "active", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create active turn: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_cleanup_continue", activeTurn.ID, "runner", activeTurn.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_cleanup_continue", activeTurn.ID, "user", "late steer", map[string]any{"intent": "prompt", "model": "bootstrap"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}
	engineA := New(s)
	engineB := New(s)
	hookStarted := make(chan struct{})
	releaseHook := make(chan struct{})
	releaseSetup := make(chan struct{})
	engineA.beforeCleanupNextWorkHook = func(ctx context.Context, sessionID string) {
		select {
		case <-hookStarted:
		default:
			close(hookStarted)
		}
		<-releaseHook
	}
	engineA.beforeSetupHook = func(ctx context.Context, sessionID, turnID string) {
		if sessionID == "session_cleanup_continue" && turnID != activeTurn.ID {
			<-releaseSetup
		}
	}
	runnerA := engineA.runner("session_cleanup_continue")
	oldRunning := &runningTurn{turnID: activeTurn.ID}
	runnerA.current = oldRunning
	cleanupDone := make(chan struct{})
	go func() {
		defer close(cleanupDone)
		runnerA.cleanupTurnRun("session_cleanup_continue", activeTurn.ID, oldRunning)
	}()
	waitForCondition(t, 2*time.Second, func() bool {
		select {
		case <-hookStarted:
			return true
		default:
			return false
		}
	}, "cleanup next-work hook")
	resultCh := make(chan *SubmitResult, 1)
	errCh := make(chan error, 1)
	go func() {
		res, err := engineB.SubmitPrompt(ctx, RunInput{SessionID: "session_cleanup_continue", Prompt: "fresh submit", Model: "bootstrap"})
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- res
	}()
	select {
	case err := <-errCh:
		t.Fatalf("submit prompt while cleanup pending: %v", err)
	case <-resultCh:
		t.Fatal("expected submit to block behind cleanup session coordination")
	case <-time.After(150 * time.Millisecond):
	}
	close(releaseHook)
	var res *SubmitResult
	select {
	case err := <-errCh:
		t.Fatalf("submit prompt after cleanup release: %v", err)
	case res = <-resultCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for submit result")
	}
	if res.Queued || res.Status != "running" {
		t.Fatalf("expected submit to steer or join running continuation, got %#v", res)
	}
	turns, err := s.ListTurns(ctx, "session_cleanup_continue")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected only active turn + continuation turn after concurrent submit, got %#v", turns)
	}
	if depth, err := s.SteeringQueueLength(ctx, "session_cleanup_continue"); err != nil {
		t.Fatalf("steering queue length: %v", err)
	} else if depth != 1 {
		t.Fatalf("expected fresh submit to become steering against continuation, got depth %d", depth)
	}
	close(releaseSetup)
	select {
	case <-cleanupDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for cleanup completion")
	}
}

func TestBusySameSessionPromptCreatesSteeringNotQueuedTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_steer_busy", "Test", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	first, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_steer_busy", Prompt: "first", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit first: %v", err)
	}
	second, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_steer_busy", Prompt: "second", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit second: %v", err)
	}
	if second.TurnID != first.TurnID || second.Queued {
		t.Fatalf("expected second prompt to steer the active turn, got %#v", second)
	}
	turns, err := s.ListTurns(ctx, "session_steer_busy")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected only the active turn before continuation, got %#v", turns)
	}
	if depth, err := s.SteeringQueueLength(ctx, "session_steer_busy"); err != nil {
		t.Fatalf("steering queue length: %v", err)
	} else if depth != 1 {
		t.Fatalf("expected steering queue depth 1, got %d", depth)
	}
}

func TestFailureMarkerDoesNotBlockLaterTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_marker", "Test", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	failedTurn, err := s.CreateTurnWithStatus(ctx, "turn_failed_marker", "session_marker", "failed", "broken", map[string]any{"intent": "prompt"})
	if err != nil {
		t.Fatalf("create failed turn: %v", err)
	}
	if err := s.UpsertTurnFailure(ctx, failedTurn.ID, "session_marker", "provider_error", "none", "provider failed"); err != nil {
		t.Fatalf("upsert failure marker: %v", err)
	}
	engine := New(s)
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_marker", Prompt: "new work", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit prompt: %v", err)
	}
	if result.Queued {
		t.Fatalf("expected new turn to start despite prior failure marker: %#v", result)
	}
}

func TestRetryHeldTurnCreatesFollowOnWork(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_retry_hold", "Test", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_retry_hold", "session_retry_hold", "failed", "redo this", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if err := s.UpsertTurnFailure(ctx, turnRec.ID, turnRec.SessionID, "provider_error", "none", "provider failed"); err != nil {
		t.Fatalf("mark failure: %v", err)
	}
	engine := New(s)
	if err := engine.HoldTurnFailure(ctx, turnRec.ID, "review", "needs operator choice"); err != nil {
		t.Fatalf("hold turn failure: %v", err)
	}
	result, err := engine.RetryHeldTurn(ctx, turnRec.ID, "retry requested")
	if err != nil {
		t.Fatalf("retry held turn: %v", err)
	}
	if result.TurnID == turnRec.ID {
		t.Fatalf("expected retry to create a new turn, got %#v", result)
	}
	failureRec, err := s.GetTurnFailure(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("get resolved failure row: %v", err)
	}
	if failureRec.HoldState != "none" || failureRec.ResolutionState != "retried" || failureRec.ResolvedTurnID != result.TurnID {
		t.Fatalf("unexpected resolved failure row: %#v", failureRec)
	}
	time.Sleep(1500 * time.Millisecond)
	turns, err := s.ListTurns(ctx, "session_retry_hold")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected original + retry turn, got %#v", turns)
	}
	retryTurn, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("get retry turn: %v", err)
	}
	if retryTurn.Metadata["retry_of_turn_id"] != turnRec.ID {
		t.Fatalf("expected retry metadata on new turn, got %#v", retryTurn.Metadata)
	}
}

func TestHeldTurnDoesNotBlockLaterSubmitAfterSkip(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_skip_hold", "Test", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_skip_hold", "session_skip_hold", "failed", "skip this", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if err := s.UpsertTurnFailure(ctx, turnRec.ID, turnRec.SessionID, "provider_error", "none", "provider failed"); err != nil {
		t.Fatalf("mark failure: %v", err)
	}
	engine := New(s)
	if err := engine.HoldTurnFailure(ctx, turnRec.ID, "review", "needs operator choice"); err != nil {
		t.Fatalf("hold turn failure: %v", err)
	}
	if err := engine.SkipHeldTurn(ctx, turnRec.ID, "skip requested"); err != nil {
		t.Fatalf("skip held turn: %v", err)
	}
	failureRec, err := s.GetTurnFailure(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("get skipped failure row: %v", err)
	}
	if failureRec.HoldState != "none" || failureRec.ResolutionState != "skipped" {
		t.Fatalf("unexpected skipped failure row: %#v", failureRec)
	}
	resolvedTurn, err := s.GetTurn(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("get resolved turn: %v", err)
	}
	if resolvedTurn.Phase != "failed" {
		t.Fatalf("expected held turn to return to failed phase after skip, got %#v", resolvedTurn)
	}
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_skip_hold", Prompt: "fresh work", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit prompt after skip: %v", err)
	}
	if result.Queued {
		t.Fatalf("expected fresh turn to start despite skipped held turn: %#v", result)
	}
}

func TestCancelQueuedTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	sess, _ := s.CreateSession(ctx, "session_1", "Test", map[string]any{"model": "bootstrap"})
	engine := New(s)
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_queued_cancel", sess.ID, "queued", "two", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	if err := engine.CancelTurn(ctx, "session_1", queuedTurn.ID); err != nil {
		t.Fatalf("cancel queued: %v", err)
	}
	turnRec, err := s.GetTurn(ctx, queuedTurn.ID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if turnRec.Status != "cancelled" {
		t.Fatalf("expected cancelled, got %s", turnRec.Status)
	}
	if turnRec.FinishedAt == "" {
		t.Fatalf("expected queued cancel to set finished_at, got %#v", turnRec)
	}
}

func TestCancelQueuedTurnIgnoresCallerSessionID(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	sessA, _ := s.CreateSession(ctx, "session_cancel_a", "A", map[string]any{"model": "bootstrap"})
	if _, err := s.CreateSession(ctx, "session_cancel_b", "B", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create secondary session: %v", err)
	}
	engine := New(s)
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_queued_cancel_cross", sessA.ID, "queued", "two", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	if err := engine.CancelTurn(ctx, "session_cancel_b", queuedTurn.ID); err != nil {
		t.Fatalf("cancel queued with wrong caller session: %v", err)
	}
	turnRec, err := s.GetTurn(ctx, queuedTurn.ID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if turnRec.Status != "cancelled" {
		t.Fatalf("expected cancelled, got %s", turnRec.Status)
	}
}

func TestCancelActiveStreamingTurnMarksCancelled(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_cancel_streaming", "Streaming", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	started := make(chan struct{})
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		select {
		case <-started:
		default:
			close(started)
		}
		if broadcast != nil {
			broadcast(map[string]any{"type": "text_delta", "delta": "partial"})
		}
		<-ctx.Done()
		return nil, ctx.Err()
	})
	engine := New(s)
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_cancel_streaming", Prompt: "stream please", Model: "mock-stream"})
	if err != nil {
		t.Fatalf("submit prompt: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		select {
		case <-started:
			return true
		default:
			return false
		}
	}, "streaming turn start")
	if err := engine.CancelTurn(ctx, "session_cancel_streaming", result.TurnID); err != nil {
		t.Fatalf("cancel active streaming turn: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && turnRec.Status == "cancelled" && turnRec.FinishedAt != ""
	}, "streaming turn cancellation")
	turnRec, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if turnRec.Phase != "aborted" {
		t.Fatalf("expected aborted terminal phase after cancel, got %#v", turnRec)
	}
	if turnRec.FinishedAt == "" {
		t.Fatalf("expected finished_at after active cancel, got %#v", turnRec)
	}
	events, err := s.ListTurnEvents(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundCancelling := false
	for _, event := range events {
		if event.Type == "turn.cancelling" {
			foundCancelling = true
		}
	}
	if !foundCancelling {
		t.Fatalf("expected turn.cancelling event, got %#v", events)
	}
}

func TestCrossEngineCancelActiveStreamingTurnMarksCancelled(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_cancel_streaming_cross", "Streaming", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	started := make(chan struct{})
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		select {
		case <-started:
		default:
			close(started)
		}
		<-ctx.Done()
		return nil, ctx.Err()
	})
	engineA := New(s)
	engineB := New(s)
	result, err := engineA.SubmitPrompt(ctx, RunInput{SessionID: "session_cancel_streaming_cross", Prompt: "stream please", Model: "mock-stream"})
	if err != nil {
		t.Fatalf("submit prompt: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		select {
		case <-started:
			return true
		default:
			return false
		}
	}, "cross-engine streaming turn start")
	if err := engineB.CancelTurn(ctx, "session_cancel_streaming_cross", result.TurnID); err != nil {
		t.Fatalf("cross-engine cancel active streaming turn: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && turnRec.Status == "cancelled" && turnRec.FinishedAt != ""
	}, "cross-engine streaming turn cancellation")
}

func TestCancelTurnDuringToolExecutionMarksCancelled(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_cancel_tool", "Tool", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	toolStarted := make(chan struct{})
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonToolUse, Content: []goai.ContentBlock{{Type: "toolCall", ID: "tc_cancel", Name: "block", Arguments: map[string]any{}}}}}, nil
	})
	engine := New(s)
	if err := engine.RegisterTool(RegisteredTool{Name: "block", Description: "blocks until cancelled", Executor: func(ctx context.Context, rt ToolRuntime, call goai.ToolCall) (string, error) {
		select {
		case <-toolStarted:
		default:
			close(toolStarted)
		}
		<-ctx.Done()
		return "", ctx.Err()
	}}); err != nil {
		t.Fatalf("register tool: %v", err)
	}
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_cancel_tool", Prompt: "tool please", Model: "mock-tool"})
	if err != nil {
		t.Fatalf("submit prompt: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		select {
		case <-toolStarted:
			return true
		default:
			return false
		}
	}, "tool execution start")
	if err := engine.CancelTurn(ctx, "session_cancel_tool", result.TurnID); err != nil {
		t.Fatalf("cancel turn during tool execution: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && turnRec.Status == "cancelled" && turnRec.FinishedAt != ""
	}, "tool turn cancellation")
	turnRec, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if turnRec.Phase != "aborted" {
		t.Fatalf("expected aborted phase after tool cancellation, got %#v", turnRec)
	}
	events, err := s.ListTurnEvents(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	for _, event := range events {
		if event.Type == "tool.failed" {
			t.Fatalf("expected cancellation to avoid tool.failed event, got %#v", events)
		}
	}
}

func TestCancelActiveParentTurnPropagatesToChildSubTurns(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_cancel_runtime", "Parent", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_cancel_runtime", "Child", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	started := make(chan struct{})
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		select {
		case <-started:
		default:
			close(started)
		}
		<-ctx.Done()
		return nil, ctx.Err()
	})
	engine := New(s)
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_parent_cancel_runtime", Prompt: "stream parent", Model: "mock-stream"})
	if err != nil {
		t.Fatalf("submit parent prompt: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		select {
		case <-started:
			return true
		default:
			return false
		}
	}, "parent streaming start")
	childTurn, err := s.CreateTurnWithStatus(ctx, "turn_child_cancel_runtime", "session_child_cancel_runtime", "running", "child", map[string]any{"intent": "prompt", "parent_turn_id": result.TurnID})
	if err != nil {
		t.Fatalf("create child turn: %v", err)
	}
	if _, err := s.CreateSubTurn(ctx, result.TurnID, "session_parent_cancel_runtime", childTurn.ID, "session_child_cancel_runtime", "async", 1, map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create subturn link: %v", err)
	}
	var childCancelled atomic.Int32
	runnerChild := engine.runner("session_child_cancel_runtime")
	runnerChild.mu.Lock()
	runnerChild.current = &runningTurn{turnID: childTurn.ID, cancel: func() { childCancelled.Add(1) }}
	runnerChild.mu.Unlock()
	if err := engine.CancelTurn(ctx, "session_parent_cancel_runtime", result.TurnID); err != nil {
		t.Fatalf("cancel parent turn: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && turnRec.Status == "cancelled" && turnRec.FinishedAt != ""
	}, "parent turn cancellation")
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, childTurn.ID)
		return err == nil && turnRec.Status == "cancelling"
	}, "child turn cancellation propagation")
	waitForCondition(t, 2*time.Second, func() bool {
		return childCancelled.Load() == 1
	}, "child cancel callback propagation")
}

func TestQueuedTurnsRunInCreatedOrder(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_queue_order", "Queue", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	firstStarted := make(chan struct{})
	secondStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		prompt := ""
		for i := len(convCtx.Messages) - 1; i >= 0; i-- {
			msg := convCtx.Messages[i]
			if msg.Role == goai.RoleUser && len(msg.Content) > 0 {
				prompt = msg.Content[0].Text
				break
			}
		}
		switch prompt {
		case "first queued":
			select {
			case <-firstStarted:
			default:
				close(firstStarted)
			}
			<-releaseFirst
			return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "first done"}}}}, nil
		case "second queued":
			select {
			case <-secondStarted:
			default:
				close(secondStarted)
			}
			return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "second done"}}}}, nil
		default:
			return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "unexpected"}}}}, nil
		}
	})
	engine := New(s)
	firstTurn, err := s.CreateTurnWithStatus(ctx, "turn_queue_order_1", "session_queue_order", "queued", "first queued", map[string]any{"intent": "prompt", "model": "mock-order"})
	if err != nil {
		t.Fatalf("create first queued turn: %v", err)
	}
	secondTurn, err := s.CreateTurnWithStatus(ctx, "turn_queue_order_2", "session_queue_order", "queued", "second queued", map[string]any{"intent": "prompt", "model": "mock-order"})
	if err != nil {
		t.Fatalf("create second queued turn: %v", err)
	}
	if err := engine.startNextQueuedTurn(ctx, "session_queue_order"); err != nil {
		t.Fatalf("start next queued turn: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		select {
		case <-firstStarted:
			return true
		default:
			return false
		}
	}, "first queued turn start")
	if err := engine.startNextQueuedTurn(ctx, "session_queue_order"); err != nil {
		t.Fatalf("start next queued turn while active: %v", err)
	}
	firstState, err := s.GetTurn(ctx, firstTurn.ID)
	if err != nil {
		t.Fatalf("get first turn: %v", err)
	}
	secondState, err := s.GetTurn(ctx, secondTurn.ID)
	if err != nil {
		t.Fatalf("get second turn: %v", err)
	}
	if firstState.StartedAt == "" {
		t.Fatalf("expected first queued turn to start, got %#v", firstState)
	}
	if secondState.Status != "queued" || secondState.StartedAt != "" {
		t.Fatalf("expected second queued turn to remain queued until first completes, got %#v", secondState)
	}
	close(releaseFirst)
	waitForCondition(t, 2*time.Second, func() bool {
		select {
		case <-secondStarted:
			return true
		default:
			return false
		}
	}, "second queued turn start")
	waitForCondition(t, 2*time.Second, func() bool {
		firstDone, err1 := s.GetTurn(ctx, firstTurn.ID)
		secondDone, err2 := s.GetTurn(ctx, secondTurn.ID)
		return err1 == nil && err2 == nil && firstDone.Status == "completed" && secondDone.Status == "completed"
	}, "queued turn completion order")
	firstDone, err := s.GetTurn(ctx, firstTurn.ID)
	if err != nil {
		t.Fatalf("get completed first turn: %v", err)
	}
	secondDone, err := s.GetTurn(ctx, secondTurn.ID)
	if err != nil {
		t.Fatalf("get completed second turn: %v", err)
	}
	if !(firstDone.StartedAt < secondDone.StartedAt) {
		t.Fatalf("expected first queued turn to start before second, got first=%q second=%q", firstDone.StartedAt, secondDone.StartedAt)
	}
}

func TestSetupErrorMarksTurnFailed(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_setup_error", "Setup", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	engine.beforeSetupErrorHook = func(ctx context.Context, sessionID, turnID string) error {
		return fmt.Errorf("boom during setup")
	}
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_setup_error", Prompt: "fail setup", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit prompt: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && turnRec.Status == "failed" && turnRec.FinishedAt != ""
	}, "setup failure finalization")
	turnRec, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if turnRec.Phase != "failed" {
		t.Fatalf("expected failed phase after setup error, got %#v", turnRec)
	}
	if turnRec.FinishedAt == "" {
		t.Fatalf("expected finished_at after setup error, got %#v", turnRec)
	}
	msgs, err := s.ListMessages(ctx, "session_setup_error")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	foundFailure := false
	for _, msg := range msgs {
		if msg.Role == "assistant" && strings.Contains(msg.Content, "Turn setup error: boom during setup") {
			foundFailure = true
		}
	}
	if !foundFailure {
		t.Fatalf("expected setup failure message in history, got %#v", msgs)
	}
}

func TestCancelTurnDuringSetupMarksCancelled(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_cancel_setup", "Setup", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	setupEntered := make(chan struct{})
	engine.beforeSetupHook = func(ctx context.Context, sessionID, turnID string) {
		select {
		case <-setupEntered:
		default:
			close(setupEntered)
		}
		<-ctx.Done()
	}
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_cancel_setup", Prompt: "cancel before start", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit prompt: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		select {
		case <-setupEntered:
			return true
		default:
			return false
		}
	}, "setup phase entry")
	if err := engine.CancelTurn(ctx, "session_cancel_setup", result.TurnID); err != nil {
		t.Fatalf("cancel during setup: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && turnRec.Status == "cancelled" && turnRec.FinishedAt != ""
	}, "setup turn cancellation")
	turnRec, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if turnRec.Phase != "aborted" {
		t.Fatalf("expected aborted phase after setup cancellation, got %#v", turnRec)
	}
	if turnRec.FinishedAt == "" {
		t.Fatalf("expected finished_at after setup cancellation, got %#v", turnRec)
	}
	events, err := s.ListTurnEvents(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundCancelling := false
	foundStarted := false
	for _, event := range events {
		if event.Type == "turn.cancelling" {
			foundCancelling = true
		}
		if event.Type == "turn.started" {
			foundStarted = true
		}
	}
	if !foundCancelling {
		t.Fatalf("expected turn.cancelling during setup cancellation, got %#v", events)
	}
	if foundStarted {
		t.Fatalf("expected setup cancellation before turn.started, got %#v", events)
	}
	msgs, err := s.ListMessages(ctx, "session_cancel_setup")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	for _, msg := range msgs {
		if msg.Role == "user" {
			t.Fatalf("expected no persisted user message before setup cancellation, got %#v", msgs)
		}
	}
}

func TestSubmitPromptRoutedCreatesChildAgentSession(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	root, err := s.CreateSession(ctx, "session_root", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	result, err := engine.SubmitPromptRouted(ctx, RunInput{SessionID: root.ID, Prompt: "@agent1 hello there", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit routed prompt: %v", err)
	}
	if !result.Routed || result.TargetAgentID != "agent1" || !result.CreatedSession {
		t.Fatalf("unexpected routed result: %#v", result)
	}
	time.Sleep(1500 * time.Millisecond)
	sessions, err := s.ListSessions(ctx)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	var child *store.Session
	for i := range sessions {
		if sessions[i].ID == result.SessionID {
			child = &sessions[i]
		}
	}
	if child == nil || child.ParentSessionID != root.ID {
		t.Fatalf("unexpected child session: %#v", child)
	}
	msgs, err := s.ListMessages(ctx, child.ID)
	if err != nil {
		t.Fatalf("list child messages: %v", err)
	}
	found := false
	for _, msg := range msgs {
		if msg.Role == "assistant" && msg.Payload["agent_id"] == "agent1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected assistant reply from @agent1, got %#v", msgs)
	}
}

func TestPreparePromptRouteResolutionUsesSourceSessionScopeContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	scope := gisession.SessionScope{Version: gisession.ScopeVersionV1, AgentID: "agent", Channel: "slack", Account: "workspace", Dimensions: []string{"space", "chat", "topic"}, Values: map[string]string{"space": "room:eng", "chat": "group:thread-7", "topic": "topic:builds"}}
	source, err := s.CreateSessionWithMetadata(ctx, "session_route_scope", "", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}, &scope, []string{"agent:agent:slack:chat:group:thread-7"})
	if err != nil {
		t.Fatalf("create source session: %v", err)
	}
	engine := New(s)
	resolution, err := engine.preparePromptRouteResolution(ctx, RunInput{SessionID: source.ID, Prompt: "@agent1 hello there", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("prepare prompt route resolution: %v", err)
	}
	if resolution.inbound.Channel != "slack" || resolution.inbound.Account != "workspace" {
		t.Fatalf("expected inbound channel/account from scope, got %#v", resolution.inbound)
	}
	if resolution.inbound.ChatType != "group" || resolution.inbound.ChatID != "thread-7" {
		t.Fatalf("expected scoped chat identity, got %#v", resolution.inbound)
	}
	if resolution.inbound.SpaceType != "room" || resolution.inbound.SpaceID != "eng" {
		t.Fatalf("expected scoped space identity, got %#v", resolution.inbound)
	}
	if resolution.inbound.TopicID != "builds" {
		t.Fatalf("expected scoped topic identity, got %#v", resolution.inbound)
	}
}

func TestPreparePromptRouteResolutionPrefersSessionIdentityOverScopeSnapshot(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	scope := gisession.SessionScope{Version: gisession.ScopeVersionV1, AgentID: "agent", Channel: "slack", Account: "workspace", Dimensions: []string{"space", "chat", "topic"}, Values: map[string]string{"space": "room:eng", "chat": "group:thread-7", "topic": "topic:builds"}}
	source, err := s.CreateSessionWithMetadata(ctx, "session_route_scope_identity", "", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}, &scope, []string{"agent:agent:slack:chat:group:thread-7"})
	if err != nil {
		t.Fatalf("create source session: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update sessions set scope_json = ? where id = ?`, `{"version":1,"agent_id":"wrong","channel":"gi","account":"default","dimensions":["chat"],"values":{"chat":"direct:wrong"}}`, source.ID); err != nil {
		t.Fatalf("mutate scope snapshot: %v", err)
	}
	engine := New(s)
	resolution, err := engine.preparePromptRouteResolution(ctx, RunInput{SessionID: source.ID, Prompt: "@agent1 hello there", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("prepare prompt route resolution: %v", err)
	}
	if resolution.inbound.Channel != "slack" || resolution.inbound.Account != "workspace" {
		t.Fatalf("expected canonical inbound channel/account, got %#v", resolution.inbound)
	}
	if resolution.inbound.ChatType != "group" || resolution.inbound.ChatID != "thread-7" || resolution.inbound.SpaceType != "room" || resolution.inbound.SpaceID != "eng" || resolution.inbound.TopicID != "builds" {
		t.Fatalf("expected canonical scoped inbound context, got %#v", resolution.inbound)
	}
}

func TestResolveOrCreateRouteSessionUsesIdentityForSameAgentFastPath(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	scope := gisession.SessionScope{Version: gisession.ScopeVersionV1, AgentID: "agent", Channel: "gi", Account: "default", Dimensions: []string{"chat"}, Values: map[string]string{"chat": "direct:session_same_agent_identity"}}
	if _, err := s.CreateSessionWithMetadata(ctx, "session_same_agent_identity", "", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}, &scope, []string{"agent:agent:gi:chat:direct:session_same_agent_identity"}); err != nil {
		t.Fatalf("create source session: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update sessions set scope_json = ? where id = ?`, `{"version":1,"agent_id":"wrong","channel":"gi","account":"default","dimensions":["chat"],"values":{"chat":"direct:session_same_agent_identity"}}`, "session_same_agent_identity"); err != nil {
		t.Fatalf("mutate scope snapshot: %v", err)
	}
	source, err := s.GetSession(ctx, "session_same_agent_identity")
	if err != nil {
		t.Fatalf("reload source session: %v", err)
	}
	engine := New(s)
	target, created, err := engine.ResolveOrCreateRouteSession(ctx, source, routing.ResolvedRoute{AgentID: "agent"}, routing.InboundContext{Channel: "gi", Account: "default", ChatType: "direct", ChatID: source.ID})
	if err != nil {
		t.Fatalf("resolve route session: %v", err)
	}
	if created || target.ID != source.ID {
		t.Fatalf("expected canonical same-agent fast path reuse, got target=%#v created=%v", target, created)
	}
}

func TestSubmitPromptRoutedRejectsDirectedPromptWithoutBody(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	root, err := s.CreateSession(ctx, "session_root_directed_empty", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	_, err = engine.SubmitPromptRouted(ctx, RunInput{SessionID: root.ID, Prompt: "@agent1", Model: "bootstrap"})
	if err == nil {
		t.Fatal("expected directed prompt validation error")
	}
	if !strings.Contains(err.Error(), "directed prompt requires content") {
		t.Fatalf("unexpected directed prompt error: %v", err)
	}
}

func TestProcessDirectPromptResolvesExplicitSessionKey(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}}, nil
	})
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_direct_key")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_direct_key", Title: "@agent", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	result, err := engine.ProcessDirect(ctx, DirectInput{Kind: DirectKindPrompt, SessionKey: alloc.SessionKey, Prompt: "hello from key", Model: "bootstrap", Origin: DirectOrigin{SourceKind: DirectSourceKindIPC, SourceID: "ipc:key"}})
	if err != nil {
		t.Fatalf("process direct prompt by session key: %v", err)
	}
	waitForCondition(t, time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && strings.TrimSpace(turnRec.FinishedAt) != ""
	}, "direct prompt by session key completion")
	if result.SessionID != sess.ID {
		t.Fatalf("expected resolved session id %s, got %#v", sess.ID, result)
	}
	turnRec, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if stringValue(turnRec.Metadata["ingress_session_key"], "") != alloc.SessionKey {
		t.Fatalf("expected ingress session key metadata, got %#v", turnRec.Metadata)
	}
}

func TestProcessDirectRoutedPromptPreservesIngressMetadata(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}}, nil
	})
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_direct_route")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_direct_route", Title: "@agent", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	result, err := engine.ProcessDirect(ctx, DirectInput{Kind: DirectKindPrompt, SessionKey: alloc.SessionKey, Prompt: "@agent1 hello routed direct", Model: "bootstrap", Origin: DirectOrigin{SourceKind: DirectSourceKindIPC, SourceID: "ipc:routed"}})
	if err != nil {
		t.Fatalf("process routed direct prompt: %v", err)
	}
	if !result.Routed || result.SessionID == sess.ID || result.TurnID == "" {
		t.Fatalf("expected routed direct result, got %#v", result)
	}
	waitForCondition(t, time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && strings.TrimSpace(turnRec.FinishedAt) != ""
	}, "routed direct prompt completion")
	turnRec, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("get routed turn: %v", err)
	}
	if stringValue(turnRec.Metadata["ingress_source_kind"], "") != DirectSourceKindIPC || stringValue(turnRec.Metadata["ingress_source_id"], "") != "ipc:routed" || stringValue(turnRec.Metadata["ingress_session_key"], "") != alloc.SessionKey {
		t.Fatalf("expected ingress metadata on routed direct turn, got %#v", turnRec.Metadata)
	}
}

func TestProcessDirectRejectsMismatchedSessionIDAndSessionKey(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_direct_mismatch")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_direct_mismatch", Title: "@agent", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	_, err = engine.ProcessDirect(ctx, DirectInput{Kind: DirectKindPrompt, SessionID: "different-session", SessionKey: alloc.SessionKey, Prompt: "hello", Model: "bootstrap"})
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("expected session id/session key mismatch error, got %v (resolved session=%s)", err, sess.ID)
	}
}

func TestProcessDirectPromptUsesNormalSubmitPathAndIngressMetadata(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}}, nil
	})
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_direct_prompt")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_direct_prompt", Title: "@agent", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	result, err := engine.ProcessDirect(ctx, DirectInput{Kind: DirectKindPrompt, SessionID: sess.ID, Prompt: "hello from direct", Model: "bootstrap", Origin: DirectOrigin{SourceKind: "ipc", SourceID: "ipc:test", Role: "system", Label: "IPC test"}})
	if err != nil {
		t.Fatalf("process direct prompt: %v", err)
	}
	if result.TurnID == "" {
		t.Fatalf("unexpected direct prompt result: %#v", result)
	}
	waitForCondition(t, time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && strings.TrimSpace(turnRec.FinishedAt) != ""
	}, "direct prompt completion")
	turnRec, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if stringValue(turnRec.Metadata["ingress_kind"], "") != "direct" || stringValue(turnRec.Metadata["ingress_source_kind"], "") != "ipc" || stringValue(turnRec.Metadata["ingress_source_id"], "") != "ipc:test" {
		t.Fatalf("expected direct ingress metadata on turn, got %#v", turnRec.Metadata)
	}
	msgs, err := s.ListMessages(ctx, result.SessionID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(msgs) == 0 || stringValue(msgs[0].Payload["ingress_source_kind"], "") != "ipc" || stringValue(msgs[0].Payload["ingress_source_id"], "") != "ipc:test" {
		t.Fatalf("expected direct ingress metadata on persisted user message, got %#v", msgs)
	}
	events, err := s.ListTurnEvents(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundStarted := false
	for _, event := range events {
		if event.Type == "turn.started" {
			foundStarted = true
			if stringValue(event.Payload["ingress_source_kind"], "") != "ipc" || stringValue(event.Payload["ingress_source_id"], "") != "ipc:test" {
				t.Fatalf("expected ingress metadata on turn.started event, got %#v", event)
			}
		}
	}
	if !foundStarted {
		t.Fatalf("expected turn.started event, got %#v", events)
	}
}

func TestProcessDirectSteersSameSessionWhileActive(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_direct_steer")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_direct_steer", Title: "@agent", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	first, err := engine.SubmitPrompt(ctx, RunInput{SessionID: sess.ID, Prompt: "first", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit first prompt: %v", err)
	}
	second, err := engine.ProcessDirect(ctx, DirectInput{Kind: DirectKindPrompt, SessionID: sess.ID, Prompt: "second via direct", Model: "bootstrap", Origin: DirectOrigin{SourceKind: DirectSourceKindIPC, SourceID: "ipc:active"}})
	if err != nil {
		t.Fatalf("process direct steering prompt: %v", err)
	}
	if second.TurnID != first.TurnID || second.Status != "running" || second.Queued {
		t.Fatalf("expected same active turn steering result, got first=%#v second=%#v", first, second)
	}
	if depth, err := s.SteeringQueueLength(ctx, sess.ID); err != nil {
		t.Fatalf("steering queue length: %v", err)
	} else if depth != 1 {
		t.Fatalf("expected steering queue depth 1, got %d", depth)
	}
	msgs, err := s.ListMessages(ctx, sess.ID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(msgs) > 1 {
		t.Fatalf("expected queued steering not to persist a second chat message before injection, got %#v", msgs)
	}
	for _, msg := range msgs {
		if stringValue(msg.Payload["ingress_source_id"], "") == "ipc:active" {
			t.Fatalf("expected queued steering not to persist direct-ingress message before injection, got %#v", msgs)
		}
	}
}

func TestProcessDirectSteeringNormalizesUnexpectedIngressRole(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_direct_role_norm")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_direct_role_norm", Title: "@agent", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := engine.SubmitPrompt(ctx, RunInput{SessionID: sess.ID, Prompt: "first", Model: "bootstrap"}); err != nil {
		t.Fatalf("submit first prompt: %v", err)
	}
	if _, err := engine.ProcessDirect(ctx, DirectInput{Kind: DirectKindPrompt, SessionID: sess.ID, Prompt: "second", Model: "bootstrap", Origin: DirectOrigin{SourceKind: DirectSourceKindIPC, SourceID: "ipc:role", Role: "bogus-role"}}); err != nil {
		t.Fatalf("process direct steering prompt: %v", err)
	}
	msgs, err := s.DequeueSteering(ctx, sess.ID)
	if err != nil {
		t.Fatalf("dequeue steering: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Role != "user" {
		t.Fatalf("expected normalized user steering role, got %#v", msgs)
	}
}

func TestProcessSystemDirectWhileActiveSteersSameSession(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_system_steer")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_system_steer", Title: "@agent", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	first, err := engine.SubmitPrompt(ctx, RunInput{SessionID: sess.ID, Prompt: "first", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit first prompt: %v", err)
	}
	second, err := engine.ProcessSystemDirect(ctx, DirectInput{Kind: DirectKindPrompt, SessionID: sess.ID, Prompt: "system steer", Model: "bootstrap", Origin: DirectOrigin{SourceID: "scheduler:active"}})
	if err != nil {
		t.Fatalf("process system direct steering prompt: %v", err)
	}
	if second.TurnID != first.TurnID || second.Status != "running" || second.Queued {
		t.Fatalf("expected same active turn steering result, got first=%#v second=%#v", first, second)
	}
	if depth, err := s.SteeringQueueLength(ctx, sess.ID); err != nil {
		t.Fatalf("steering queue length: %v", err)
	} else if depth != 1 {
		t.Fatalf("expected one steering row, got depth=%d", depth)
	}
	msgs, err := s.DequeueSteering(ctx, sess.ID)
	if err != nil {
		t.Fatalf("dequeue steering: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected one steering row, got %#v", msgs)
	}
	if msgs[0].Role != "system" {
		t.Fatalf("expected system steering role, got %#v", msgs[0])
	}
	if stringValue(msgs[0].Payload["ingress_source_kind"], "") != DirectSourceKindSystem || stringValue(msgs[0].Payload["ingress_source_id"], "") != "scheduler:active" || stringValue(msgs[0].Payload["ingress_role"], "") != "system" {
		t.Fatalf("expected system ingress metadata on queued steering payload, got %#v", msgs[0])
	}
}

func TestPersistSteeringMessagesStoresSystemRoleAsUserChatHistory(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	if _, err := s.CreateSession(ctx, "session_steering_persist_role", "SteeringPersist", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_steering_persist_role", "session_steering_persist_role", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	runner := engine.runner("session_steering_persist_role")
	count := runner.persistSteeringMessages(ctx, "session_steering_persist_role", "turn_steering_persist_role", []store.SteeringMessage{{Role: "system", Content: "system steering", Payload: map[string]any{"intent": "prompt", "ingress_role": "system", "ingress_source_kind": DirectSourceKindSystem}}})
	if count != 1 {
		t.Fatalf("expected one persisted steering message, got %d", count)
	}
	msgs, err := s.ListMessages(ctx, "session_steering_persist_role")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Role != "user" {
		t.Fatalf("expected system steering to persist as user chat message, got %#v", msgs)
	}
	if stringValue(msgs[0].Payload["steering_role"], "") != "system" || stringValue(msgs[0].Payload["ingress_role"], "") != "system" {
		t.Fatalf("expected steering role audit metadata to be preserved, got %#v", msgs[0])
	}
}

func TestProcessSystemDirectDefaultsSystemOriginMetadata(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}}, nil
	})
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_system_direct")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_system_direct", Title: "@agent", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	result, err := engine.ProcessSystemDirect(ctx, DirectInput{Kind: DirectKindPrompt, SessionID: sess.ID, Prompt: "hello from system", Model: "bootstrap", Origin: DirectOrigin{SourceID: "scheduler:system", Label: "Scheduler"}})
	if err != nil {
		t.Fatalf("process system direct: %v", err)
	}
	waitForCondition(t, time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && strings.TrimSpace(turnRec.FinishedAt) != ""
	}, "system direct completion")
	turnRec, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if stringValue(turnRec.Metadata["ingress_source_kind"], "") != DirectSourceKindSystem || stringValue(turnRec.Metadata["ingress_role"], "") != "system" || stringValue(turnRec.Metadata["ingress_source_id"], "") != "scheduler:system" {
		t.Fatalf("expected system ingress metadata on turn, got %#v", turnRec.Metadata)
	}
	msgs, err := s.ListMessages(ctx, result.SessionID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(msgs) == 0 || stringValue(msgs[0].Payload["ingress_source_kind"], "") != DirectSourceKindSystem || stringValue(msgs[0].Payload["ingress_role"], "") != "system" {
		t.Fatalf("expected system ingress metadata on persisted user message, got %#v", msgs)
	}
}

func TestProcessInternalDirectDefaultsInternalOriginMetadata(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}}, nil
	})
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_internal_direct")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_internal_direct", Title: "@agent", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	result, err := engine.ProcessInternalDirect(ctx, DirectInput{Kind: DirectKindPrompt, SessionID: sess.ID, Prompt: "hello from internal", Model: "bootstrap", Origin: DirectOrigin{SourceID: "engine:internal"}})
	if err != nil {
		t.Fatalf("process internal direct: %v", err)
	}
	waitForCondition(t, time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && strings.TrimSpace(turnRec.FinishedAt) != ""
	}, "internal direct completion")
	turnRec, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if stringValue(turnRec.Metadata["ingress_source_kind"], "") != DirectSourceKindInternal || stringValue(turnRec.Metadata["ingress_role"], "") != "system" || stringValue(turnRec.Metadata["ingress_source_id"], "") != "engine:internal" {
		t.Fatalf("expected internal ingress metadata on turn, got %#v", turnRec.Metadata)
	}
}

func TestProcessDirectPeerMessageCarriesIngressMetadataIntoTargetTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}}, nil
	})
	sourceAlloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_direct_source")
	source, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_direct_source", Title: "@agent", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: sourceAlloc})
	if err != nil {
		t.Fatalf("create source session: %v", err)
	}
	result, err := engine.ProcessDirect(ctx, DirectInput{Kind: DirectKindPeerMessage, SessionID: source.ID, TargetAgentID: "agent1", Prompt: "hello peer direct", Intent: "prompt", Model: "bootstrap", Origin: DirectOrigin{SourceKind: "system", SourceID: "scheduler:1"}})
	if err != nil {
		t.Fatalf("process direct peer message: %v", err)
	}
	if !result.Routed || result.TurnID == "" || result.SessionID == source.ID {
		t.Fatalf("unexpected direct peer result: %#v", result)
	}
	waitForCondition(t, time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && strings.TrimSpace(turnRec.FinishedAt) != ""
	}, "direct peer completion")
	turnRec, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("get routed turn: %v", err)
	}
	if stringValue(turnRec.Metadata["ingress_kind"], "") != "direct" || stringValue(turnRec.Metadata["ingress_source_kind"], "") != "system" || stringValue(turnRec.Metadata["ingress_source_id"], "") != "scheduler:1" {
		t.Fatalf("expected direct ingress metadata on routed turn, got %#v", turnRec.Metadata)
	}
	events, err := s.ListTurnEvents(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundStarted := false
	for _, event := range events {
		if event.Type == "turn.started" {
			foundStarted = true
			if stringValue(event.Payload["ingress_source_kind"], "") != "system" || stringValue(event.Payload["ingress_source_id"], "") != "scheduler:1" {
				t.Fatalf("expected ingress metadata on routed turn.started event, got %#v", event)
			}
		}
	}
	if !foundStarted {
		t.Fatalf("expected turn.started event, got %#v", events)
	}
}

func TestResolveOrCreateRouteSessionReturnsSourceForSameAgent(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	source, err := s.CreateSession(ctx, "session_same_agent_route", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create source session: %v", err)
	}
	engine := New(s)
	target, created, err := engine.ResolveOrCreateRouteSession(ctx, source, routing.ResolvedRoute{AgentID: normalizeAgentID(sessionAgentID(source))}, routing.InboundContext{Channel: "gi", Account: "default", ChatType: "direct", ChatID: source.ID})
	if err != nil {
		t.Fatalf("resolve route session: %v", err)
	}
	if created {
		t.Fatalf("expected same-agent route to reuse source, got created=%v", created)
	}
	if target.ID != source.ID {
		t.Fatalf("expected source session reuse, got target=%#v source=%#v", target, source)
	}
}

func TestRunShellStreamsDraftChunks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var chunks []string
	out, err, cancelled := runShell(ctx, "hello", nil, func(delta string) {
		chunks = append(chunks, delta)
	})
	if cancelled || err != nil {
		t.Fatalf("unexpected shell result: out=%q err=%v cancelled=%v", out, err, cancelled)
	}
	if out != "Gi received: hello" {
		t.Fatalf("unexpected output: %q", out)
	}
	if len(chunks) == 0 {
		t.Fatal("expected streamed chunks")
	}
	if got := chunks[0]; got == "" {
		t.Fatalf("expected non-empty first chunk: %#v", chunks)
	}
}

func TestSubmitPeerMessageUsesExistingTargetSession(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	source, err := s.CreateSession(ctx, "session_root", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	target, err := s.CloneSession(ctx, source.ID, "session_child", "@agent1", "agent1")
	if err != nil {
		t.Fatalf("create target: %v", err)
	}
	engine := New(s)
	result, err := engine.SubmitPeerMessage(ctx, source.ID, "agent1", "hello from peer", "prompt", "bootstrap", "")
	if err != nil {
		t.Fatalf("submit peer message: %v", err)
	}
	if result.SessionID != target.ID || result.CreatedSession {
		t.Fatalf("unexpected peer result: %#v", result)
	}
}

func TestConcurrentSubmitDifferentSessionsRunsConcurrently(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_a", "A", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session a: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_b", "B", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session b: %v", err)
	}
	engine := New(s)

	start := make(chan struct{})
	errCh := make(chan error, 2)
	resCh := make(chan *SubmitResult, 2)
	for i, sessionID := range []string{"session_a", "session_b"} {
		go func(idx int, sid string) {
			<-start
			res, err := engine.SubmitPrompt(ctx, RunInput{SessionID: sid, Prompt: fmt.Sprintf("hello-%d", idx+1), Model: "bootstrap"})
			if err != nil {
				errCh <- err
				return
			}
			resCh <- res
		}(i, sessionID)
	}
	close(start)

	var results []*SubmitResult
	for i := 0; i < 2; i++ {
		select {
		case err := <-errCh:
			t.Fatalf("submit error: %v", err)
		case res := <-resCh:
			results = append(results, res)
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for submit results")
		}
	}
	if results[0].TurnID == results[1].TurnID {
		t.Fatalf("expected distinct turns across sessions, got %#v", results)
	}

	time.Sleep(1500 * time.Millisecond)
	for _, sid := range []string{"session_a", "session_b"} {
		turns, err := s.ListTurns(ctx, sid)
		if err != nil {
			t.Fatalf("list turns %s: %v", sid, err)
		}
		if len(turns) != 1 || turns[0].Status != "completed" {
			t.Fatalf("unexpected turns for %s: %#v", sid, turns)
		}
	}
}

func TestContinueSessionPreservesSteeringMediaInHistory(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_media_steering", "Test", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	media := []string{"media_1", "media_2"}
	if _, err := s.EnqueueSteering(ctx, "session_media_steering", "", "user", "check these", map[string]any{"intent": "prompt", "model": "bootstrap"}, media, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering with media: %v", err)
	}
	engine := New(s)
	continued, err := engine.ContinueSession(ctx, "session_media_steering")
	if err != nil {
		t.Fatalf("continue session: %v", err)
	}
	if !continued {
		t.Fatal("expected session continuation")
	}
	time.Sleep(1500 * time.Millisecond)

	msgs, err := s.ListMessages(ctx, "session_media_steering")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	found := false
	for _, msg := range msgs {
		if msg.Role != "user" || msg.Payload["steering"] != true {
			continue
		}
		rawMedia, ok := msg.Payload["media"]
		if !ok {
			continue
		}
		switch v := rawMedia.(type) {
		case []any:
			if len(v) == len(media) {
				found = true
			}
		case []string:
			if len(v) == len(media) {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("expected persisted steering media in history payloads, got %#v", msgs)
	}
}

func TestSkipRemainingToolCallsPersistsSkippedResults(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_skip_tools", "Test", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_skip_tools", "session_skip_tools", "running", "prompt", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	engine := New(s)
	runner := engine.runner("session_skip_tools")
	convCtx := &goai.Context{}
	calls := []goai.ToolCall{
		{ID: "tc1", Name: "tool.one", Arguments: map[string]any{}},
		{ID: "tc2", Name: "tool.two", Arguments: map[string]any{}},
	}
	runner.skipRemainingToolCalls(ctx, "session_skip_tools", "turn_skip_tools", convCtx, calls, 0)

	events, err := s.ListTurnEvents(ctx, "turn_skip_tools")
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	skippedEvents := 0
	for _, event := range events {
		if event.Type == "tool.skipped" {
			skippedEvents++
		}
	}
	if skippedEvents != 2 {
		t.Fatalf("expected 2 tool.skipped events, got %d (%#v)", skippedEvents, events)
	}

	msgs, err := s.ListMessages(ctx, "session_skip_tools")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	skippedMsgs := 0
	for _, msg := range msgs {
		if msg.Role != "tool_result" {
			continue
		}
		if msg.Payload["skipped"] == true && msg.Payload["skip_reason"] == "queued user steering message" && msg.Content == skippedDueToQueuedUserMessage {
			skippedMsgs++
		}
	}
	if skippedMsgs != 2 {
		t.Fatalf("expected 2 skipped tool_result messages, got %d (%#v)", skippedMsgs, msgs)
	}
}

func TestSubmitPromptWithParentTurnCreatesSubTurnRecord(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_turn", "Test", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	parent, err := s.CreateTurnWithStatus(ctx, "turn_parent_submit", "session_parent_turn", "completed", "parent", map[string]any{"intent": "prompt", "subturn_depth": 0})
	if err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	engine := New(s)
	res, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_parent_turn", Prompt: "child", Model: "bootstrap", ParentTurnID: parent.ID})
	if err != nil {
		t.Fatalf("submit child turn: %v", err)
	}
	link, err := s.GetSubTurnByChild(ctx, res.TurnID)
	if err != nil {
		t.Fatalf("get subturn link: %v", err)
	}
	if link.ParentTurnID != parent.ID || link.ChildTurnID != res.TurnID || link.Depth != 1 {
		t.Fatalf("unexpected subturn link: %#v", link)
	}
}

func TestSubmitPromptWithParentTurnRejectsDepthOverflow(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_depth", "Test", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	parent, err := s.CreateTurnWithStatus(ctx, "turn_parent_depth", "session_parent_depth", "completed", "parent", map[string]any{"intent": "prompt", "subturn_depth": defaultSubTurnMaxDepth})
	if err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	engine := New(s)
	_, err = engine.SubmitPrompt(ctx, RunInput{SessionID: "session_parent_depth", Prompt: "child", Model: "bootstrap", ParentTurnID: parent.ID})
	if err == nil {
		t.Fatal("expected subturn depth overflow error")
	}
	if !strings.Contains(err.Error(), "subturn depth limit exceeded") {
		t.Fatalf("unexpected depth overflow error: %v", err)
	}
}

func TestSubmitPromptWithParentTurnRejectsConcurrencyOverflow(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_concurrency", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_concurrency", "Child", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	parent, err := s.CreateTurnWithStatus(ctx, "turn_parent_concurrency", "session_parent_concurrency", "running", "parent", map[string]any{"intent": "prompt", "subturn_depth": 0})
	if err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_child_existing", "session_child_concurrency", "running", "child", map[string]any{"intent": "prompt", "subturn_depth": 1, "parent_turn_id": parent.ID}); err != nil {
		t.Fatalf("create child turn: %v", err)
	}
	if _, err := s.CreateSubTurn(ctx, parent.ID, "session_parent_concurrency", "turn_child_existing", "session_child_concurrency", "sync", 1, map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create existing subturn: %v", err)
	}
	engine := New(s)
	_, err = engine.SubmitPrompt(ctx, RunInput{SessionID: "session_child_concurrency", Prompt: "new child", Model: "bootstrap", ParentTurnID: parent.ID, Metadata: map[string]any{"subturn_max_concurrency": 1}})
	if err == nil {
		t.Fatal("expected subturn concurrency overflow error")
	}
	if !strings.Contains(err.Error(), "subturn concurrency limit exceeded") {
		t.Fatalf("unexpected concurrency overflow error: %v", err)
	}
}

func TestSubmitPromptWithParentTurnSupportsAsyncDeliveryMode(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_async", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_async", "Child", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	parent, err := s.CreateTurnWithStatus(ctx, "turn_parent_async", "session_parent_async", "completed", "parent", map[string]any{"intent": "prompt", "subturn_depth": 0})
	if err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	engine := New(s)
	res, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_child_async", Prompt: "child", Model: "bootstrap", ParentTurnID: parent.ID, Metadata: map[string]any{"subturn_delivery_mode": "async"}})
	if err != nil {
		t.Fatalf("submit async child turn: %v", err)
	}
	link, err := s.GetSubTurnByChild(ctx, res.TurnID)
	if err != nil {
		t.Fatalf("get async subturn link: %v", err)
	}
	if link.DeliveryMode != "async" {
		t.Fatalf("expected async delivery mode, got %#v", link)
	}
}

func TestSubmitPromptWithParentTurnRejectsInvalidDeliveryMode(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_invalid_mode", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_invalid_mode", "Child", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	parent, err := s.CreateTurnWithStatus(ctx, "turn_parent_invalid_mode", "session_parent_invalid_mode", "completed", "parent", map[string]any{"intent": "prompt", "subturn_depth": 0})
	if err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	engine := New(s)
	_, err = engine.SubmitPrompt(ctx, RunInput{SessionID: "session_child_invalid_mode", Prompt: "child", Model: "bootstrap", ParentTurnID: parent.ID, Metadata: map[string]any{"subturn_delivery_mode": "sideband"}})
	if err == nil {
		t.Fatal("expected invalid subturn delivery mode error")
	}
	if !strings.Contains(err.Error(), "invalid subturn delivery mode") {
		t.Fatalf("unexpected delivery mode error: %v", err)
	}
}

func TestSubTurnSyncVsAsyncResultDelivery(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_delivery", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_sync_delivery", "ChildSync", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create sync child session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_async_delivery", "ChildAsync", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create async child session: %v", err)
	}
	parent, err := s.CreateTurnWithStatus(ctx, "turn_parent_delivery", "session_parent_delivery", "running", "parent", map[string]any{"intent": "prompt", "subturn_depth": 0})
	if err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	engine := New(s)
	if _, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_child_sync_delivery", Prompt: "sync child", Model: "bootstrap", ParentTurnID: parent.ID}); err != nil {
		t.Fatalf("submit sync child: %v", err)
	}
	if _, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_child_async_delivery", Prompt: "async child", Model: "bootstrap", ParentTurnID: parent.ID, Metadata: map[string]any{"subturn_delivery_mode": "async"}}); err != nil {
		t.Fatalf("submit async child: %v", err)
	}
	time.Sleep(1200 * time.Millisecond)
	msgs, err := s.ListMessages(ctx, "session_parent_delivery")
	if err != nil {
		t.Fatalf("list parent messages: %v", err)
	}
	syncDelivered := 0
	asyncDelivered := 0
	for _, msg := range msgs {
		if msg.Role != "system" || msg.Payload["kind"] != "subturn_result" {
			continue
		}
		if msg.Payload["delivery_mode"] == "sync" {
			syncDelivered++
		}
		if msg.Payload["delivery_mode"] == "async" {
			asyncDelivered++
		}
	}
	if syncDelivered == 0 {
		t.Fatalf("expected at least one sync subturn delivery in parent messages, got %#v", msgs)
	}
	if asyncDelivered != 0 {
		t.Fatalf("expected no async subturn delivery message in parent history, got %#v", msgs)
	}
}

func TestAsyncSubTurnOrphanHandlingPersistsParentNotice(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_orphan", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_orphan", "Child", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_parent_orphan", "session_parent_orphan", "completed", "parent", map[string]any{"intent": "prompt", "subturn_depth": 0}); err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_child_orphan", "session_child_orphan", "running", "child", map[string]any{"intent": "prompt", "subturn_depth": 1, "parent_turn_id": "turn_parent_orphan"}); err != nil {
		t.Fatalf("create child turn: %v", err)
	}
	if _, err := s.CreateSubTurn(ctx, "turn_parent_orphan", "session_parent_orphan", "turn_child_orphan", "session_child_orphan", "async", 1, map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create async subturn: %v", err)
	}
	if err := s.AddMessage(ctx, "msg_child_orphan_summary", "session_child_orphan", "assistant", "child async result", map[string]any{"kind": "chat", "turn_id": "turn_child_orphan"}); err != nil {
		t.Fatalf("seed child summary message: %v", err)
	}
	engine := New(s)
	runner := engine.runner("session_child_orphan")
	runner.publishSubTurnLifecycle(ctx, "turn_child_orphan", "completed")

	sub, err := s.GetSubTurnByChild(ctx, "turn_child_orphan")
	if err != nil {
		t.Fatalf("get orphaned subturn: %v", err)
	}
	if sub.Metadata["orphaned"] != true {
		t.Fatalf("expected subturn metadata orphaned flag, got %#v", sub.Metadata)
	}

	parentMsgs, err := s.ListMessages(ctx, "session_parent_orphan")
	if err != nil {
		t.Fatalf("list parent orphan messages: %v", err)
	}
	found := false
	for _, msg := range parentMsgs {
		if msg.Role == "system" && msg.Payload["kind"] == "subturn_orphan_result" && msg.Payload["delivery_mode"] == "async" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected orphan result notice in parent session, got %#v", parentMsgs)
	}
}

func TestBroadcastConcurrentUnsubscribeDoesNotPanic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	sessionID := "session_broadcast_race"
	ch := engine.Subscribe(sessionID)

	panicCh := make(chan any, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				panicCh <- r
			}
		}()
		for i := 0; i < 4000; i++ {
			engine.broadcast(sessionID, map[string]any{"type": "agent_status", "turn_id": fmt.Sprintf("turn_%d", i)})
		}
	}()

	for i := 0; i < 256; i++ {
		engine.Unsubscribe(sessionID, ch)
		ch = engine.Subscribe(sessionID)
	}
	engine.Unsubscribe(sessionID, ch)
	<-done
	select {
	case p := <-panicCh:
		t.Fatalf("broadcast panicked during concurrent unsubscribe: %v", p)
	default:
	}
}

func TestUnsubscribeIsIdempotent(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	sessionID := "session_unsub_idempotent"
	ch := engine.Subscribe(sessionID)
	engine.Unsubscribe(sessionID, ch)
	engine.Unsubscribe(sessionID, ch)
}

func TestGracefulParentFinishCancelsOnlyNonCriticalChildSubTurns(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_graceful_cancel", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_noncritical", "ChildA", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create noncritical child session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_critical", "ChildB", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create critical child session: %v", err)
	}
	parent, err := s.CreateTurnWithStatus(ctx, "turn_parent_graceful_cancel", "session_parent_graceful_cancel", "completed", "parent", map[string]any{"intent": "prompt"})
	if err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_child_noncritical", "session_child_noncritical", "running", "child a", map[string]any{"intent": "prompt", "parent_turn_id": parent.ID}); err != nil {
		t.Fatalf("create noncritical child turn: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_child_critical", "session_child_critical", "running", "child b", map[string]any{"intent": "prompt", "parent_turn_id": parent.ID, "subturn_critical": true}); err != nil {
		t.Fatalf("create critical child turn: %v", err)
	}
	if _, err := s.CreateSubTurn(ctx, parent.ID, "session_parent_graceful_cancel", "turn_child_noncritical", "session_child_noncritical", "async", 1, map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create noncritical subturn: %v", err)
	}
	if _, err := s.CreateSubTurn(ctx, parent.ID, "session_parent_graceful_cancel", "turn_child_critical", "session_child_critical", "async", 1, map[string]any{"intent": "prompt", "subturn_critical": true}); err != nil {
		t.Fatalf("create critical subturn: %v", err)
	}
	engine := New(s)
	noncriticalCancelled := 0
	criticalCancelled := 0
	runnerA := engine.runner("session_child_noncritical")
	runnerA.mu.Lock()
	runnerA.current = &runningTurn{turnID: "turn_child_noncritical", cancel: func() { noncriticalCancelled++ }}
	runnerA.mu.Unlock()
	runnerB := engine.runner("session_child_critical")
	runnerB.mu.Lock()
	runnerB.current = &runningTurn{turnID: "turn_child_critical", cancel: func() { criticalCancelled++ }}
	runnerB.mu.Unlock()

	parentRunner := engine.runner("session_parent_graceful_cancel")
	parentRunner.propagateChildSubTurnCancellation(ctx, parent.ID, "completed", "")

	childA, err := s.GetTurn(ctx, "turn_child_noncritical")
	if err != nil {
		t.Fatalf("get noncritical child turn: %v", err)
	}
	if childA.Status != "cancelling" {
		t.Fatalf("expected noncritical child to enter cancelling, got %#v", childA)
	}
	if noncriticalCancelled != 1 {
		t.Fatalf("expected noncritical child cancel invoked once, got %d", noncriticalCancelled)
	}
	childB, err := s.GetTurn(ctx, "turn_child_critical")
	if err != nil {
		t.Fatalf("get critical child turn: %v", err)
	}
	if childB.Status != "running" {
		t.Fatalf("expected critical child to remain running after graceful finish, got %#v", childB)
	}
	if criticalCancelled != 0 {
		t.Fatalf("expected critical child not to be cancelled on graceful finish, got %d", criticalCancelled)
	}
}

func TestHardAbortParentCancelsDescendantSubTurns(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_abort_cancel", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_abort_cancel", "Child", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_grandchild_abort_cancel", "Grandchild", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create grandchild session: %v", err)
	}
	parent, err := s.CreateTurnWithStatus(ctx, "turn_parent_abort_cancel", "session_parent_abort_cancel", "cancelled", "parent", map[string]any{"intent": "prompt"})
	if err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	child, err := s.CreateTurnWithStatus(ctx, "turn_child_abort_cancel", "session_child_abort_cancel", "running", "child", map[string]any{"intent": "prompt", "parent_turn_id": parent.ID, "subturn_critical": true})
	if err != nil {
		t.Fatalf("create child turn: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_grandchild_abort_cancel", "session_grandchild_abort_cancel", "running", "grandchild", map[string]any{"intent": "prompt", "parent_turn_id": child.ID}); err != nil {
		t.Fatalf("create grandchild turn: %v", err)
	}
	if _, err := s.CreateSubTurn(ctx, parent.ID, "session_parent_abort_cancel", child.ID, "session_child_abort_cancel", "async", 1, map[string]any{"intent": "prompt", "subturn_critical": true}); err != nil {
		t.Fatalf("create child subturn: %v", err)
	}
	if _, err := s.CreateSubTurn(ctx, child.ID, "session_child_abort_cancel", "turn_grandchild_abort_cancel", "session_grandchild_abort_cancel", "async", 2, map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create grandchild subturn: %v", err)
	}
	engine := New(s)
	childCancelled := 0
	grandchildCancelled := 0
	runnerChild := engine.runner("session_child_abort_cancel")
	runnerChild.mu.Lock()
	runnerChild.current = &runningTurn{turnID: child.ID, cancel: func() { childCancelled++ }}
	runnerChild.mu.Unlock()
	runnerGrandchild := engine.runner("session_grandchild_abort_cancel")
	runnerGrandchild.mu.Lock()
	runnerGrandchild.current = &runningTurn{turnID: "turn_grandchild_abort_cancel", cancel: func() { grandchildCancelled++ }}
	runnerGrandchild.mu.Unlock()

	parentRunner := engine.runner("session_parent_abort_cancel")
	parentRunner.propagateChildSubTurnCancellation(ctx, parent.ID, "cancelled", "")

	childTurn, err := s.GetTurn(ctx, child.ID)
	if err != nil {
		t.Fatalf("get child turn: %v", err)
	}
	if childTurn.Status != "cancelling" {
		t.Fatalf("expected child turn cancelling after hard abort, got %#v", childTurn)
	}
	grandchildTurn, err := s.GetTurn(ctx, "turn_grandchild_abort_cancel")
	if err != nil {
		t.Fatalf("get grandchild turn: %v", err)
	}
	if grandchildTurn.Status != "cancelling" {
		t.Fatalf("expected grandchild turn cancelling after hard abort, got %#v", grandchildTurn)
	}
	if childCancelled != 1 || grandchildCancelled != 1 {
		t.Fatalf("expected child/grandchild cancels once each, got child=%d grandchild=%d", childCancelled, grandchildCancelled)
	}
}

func TestTimeoutParentCancelsCriticalChildSubTurns(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_timeout_cancel", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_timeout_cancel", "Child", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	parent, err := s.CreateTurnWithStatus(ctx, "turn_parent_timeout_cancel", "session_parent_timeout_cancel", "failed", "parent", map[string]any{"intent": "prompt"})
	if err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	child, err := s.CreateTurnWithStatus(ctx, "turn_child_timeout_cancel", "session_child_timeout_cancel", "running", "child", map[string]any{"intent": "prompt", "parent_turn_id": parent.ID, "subturn_critical": true})
	if err != nil {
		t.Fatalf("create child turn: %v", err)
	}
	if _, err := s.CreateSubTurn(ctx, parent.ID, "session_parent_timeout_cancel", child.ID, "session_child_timeout_cancel", "async", 1, map[string]any{"intent": "prompt", "subturn_critical": true}); err != nil {
		t.Fatalf("create critical child subturn: %v", err)
	}
	engine := New(s)
	cancelled := 0
	runnerChild := engine.runner("session_child_timeout_cancel")
	runnerChild.mu.Lock()
	runnerChild.current = &runningTurn{turnID: child.ID, cancel: func() { cancelled++ }}
	runnerChild.mu.Unlock()

	parentRunner := engine.runner("session_parent_timeout_cancel")
	parentRunner.propagateChildSubTurnCancellation(ctx, parent.ID, "failed", "parent_timeout")

	childTurn, err := s.GetTurn(ctx, child.ID)
	if err != nil {
		t.Fatalf("get child turn: %v", err)
	}
	if childTurn.Status != "cancelling" {
		t.Fatalf("expected critical child turn cancelling after timeout, got %#v", childTurn)
	}
	if cancelled != 1 {
		t.Fatalf("expected critical child cancel invoked once on timeout, got %d", cancelled)
	}
	link, err := s.GetSubTurnByChild(ctx, child.ID)
	if err != nil {
		t.Fatalf("get child subturn link: %v", err)
	}
	if got := link.Metadata["cancel_reason"]; got != "parent_timeout" {
		t.Fatalf("expected parent_timeout cancel reason, got %#v", link.Metadata)
	}
}

func TestSubmitPromptWithParentTurnInheritsEffectiveTools(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_tools_inherit", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	parent, err := s.CreateTurnWithStatus(ctx, "turn_parent_tools_inherit", "session_parent_tools_inherit", "completed", "parent", map[string]any{"intent": "prompt", "subturn_depth": 0, "effective_tools": []string{"read", "shell"}})
	if err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	engine := New(s)
	res, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_parent_tools_inherit", Prompt: "child", Model: "bootstrap", ParentTurnID: parent.ID})
	if err != nil {
		t.Fatalf("submit child turn: %v", err)
	}
	turnRec, err := s.GetTurn(ctx, res.TurnID)
	if err != nil {
		t.Fatalf("get child turn: %v", err)
	}
	got := toolNamesFromValue(turnRec.Metadata["effective_tools"])
	if strings.Join(got, ",") != "read,shell" {
		t.Fatalf("expected inherited effective tools read,shell, got %#v", turnRec.Metadata)
	}
	if turnRec.Metadata["subturn_tools_restricted"] != false {
		t.Fatalf("expected inherited tool set to be marked unrestricted, got %#v", turnRec.Metadata)
	}
}

func TestSubmitPromptWithParentTurnRestrictsToolsToSubset(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_tools_subset", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	parent, err := s.CreateTurnWithStatus(ctx, "turn_parent_tools_subset", "session_parent_tools_subset", "completed", "parent", map[string]any{"intent": "prompt", "subturn_depth": 0, "effective_tools": []string{"read", "write", "shell"}})
	if err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	engine := New(s)
	res, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_parent_tools_subset", Prompt: "child", Model: "bootstrap", ParentTurnID: parent.ID, Metadata: map[string]any{"subturn_tools": []string{"read", "write"}}})
	if err != nil {
		t.Fatalf("submit restricted child turn: %v", err)
	}
	turnRec, err := s.GetTurn(ctx, res.TurnID)
	if err != nil {
		t.Fatalf("get child turn: %v", err)
	}
	got := toolNamesFromValue(turnRec.Metadata["effective_tools"])
	if strings.Join(got, ",") != "read,write" {
		t.Fatalf("expected restricted effective tools read,write, got %#v", turnRec.Metadata)
	}
	if turnRec.Metadata["subturn_tools_restricted"] != true {
		t.Fatalf("expected restricted tool set marker, got %#v", turnRec.Metadata)
	}
}

func TestSubmitPromptWithParentTurnRejectsRestrictedToolOutsideParentSet(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_tools_reject", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	parent, err := s.CreateTurnWithStatus(ctx, "turn_parent_tools_reject", "session_parent_tools_reject", "completed", "parent", map[string]any{"intent": "prompt", "subturn_depth": 0, "effective_tools": []string{"read"}})
	if err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	engine := New(s)
	_, err = engine.SubmitPrompt(ctx, RunInput{SessionID: "session_parent_tools_reject", Prompt: "child", Model: "bootstrap", ParentTurnID: parent.ID, Metadata: map[string]any{"subturn_tools": []string{"shell"}}})
	if err == nil {
		t.Fatal("expected subturn restricted-tool subset error")
	}
	if !strings.Contains(err.Error(), "subset of parent effective tools") {
		t.Fatalf("unexpected restricted-tool error: %v", err)
	}
}

func TestExecuteToolRejectsDisallowedToolForTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_tool_restrict_exec", "Demo", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_tool_restrict_exec", "session_tool_restrict_exec", "running", "prompt", map[string]any{"intent": "prompt", "effective_tools": []string{"read"}}); err != nil {
		t.Fatalf("create restricted turn: %v", err)
	}
	engine := New(s)
	runner := engine.runner("session_tool_restrict_exec")
	_, err := runner.executeTool(ctx, goai.ToolCall{Name: "write", Arguments: map[string]any{"path": "foo.txt", "content": "hi"}}, "session_tool_restrict_exec", "turn_tool_restrict_exec")
	if err == nil {
		t.Fatal("expected disallowed tool execution error")
	}
	if !strings.Contains(err.Error(), "tool not allowed in this turn") {
		t.Fatalf("unexpected disallowed tool error: %v", err)
	}
}
