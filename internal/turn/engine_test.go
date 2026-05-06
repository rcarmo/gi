package turn

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/store"
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
