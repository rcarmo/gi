package turn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os/exec"
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

func TestSubmitPromptMarksSessionQueuedWhenWorkRemainsQueued(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	sess, err := s.CreateSession(ctx, "session_submit_queued_state", "QueueState", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_submit_queued_existing", sess.ID, "queued", "existing", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create existing queued turn: %v", err)
	}
	engine := New(s)
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: sess.ID, Prompt: "later", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit prompt: %v", err)
	}
	if !result.Queued || result.Status != "queued" {
		t.Fatalf("expected queued submit result, got %#v", result)
	}
	sessRec, err := s.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "queued" || sessRec.State["active_turn_id"] != nil {
		t.Fatalf("expected queued submit to mark session queued with no active turn, got %#v", sessRec.State)
	}
	if got := sessRec.State["queue_count"]; got != float64(2) && got != 2 {
		t.Fatalf("expected queue_count 2 after queued submit, got %#v", sessRec.State)
	}
}

func TestSubmitPromptPreservesSessionModelWhenInputModelIsEmpty(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	sess, err := s.CreateSession(ctx, "session_submit_model_preserve", "QueueState", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_submit_model_existing", sess.ID, "queued", "existing", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create existing queued turn: %v", err)
	}
	engine := New(s)
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: sess.ID, Prompt: "later"})
	if err != nil {
		t.Fatalf("submit prompt: %v", err)
	}
	if !result.Queued || result.Status != "queued" {
		t.Fatalf("expected queued submit result, got %#v", result)
	}
	sessRec, err := s.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["model"] != "bootstrap" || sessRec.State["status"] != "queued" {
		t.Fatalf("expected queued submit to preserve session model, got %#v", sessRec.State)
	}
	if got := sessRec.State["queue_count"]; got != float64(2) && got != 2 {
		t.Fatalf("expected queue_count 2 after queued submit, got %#v", sessRec.State)
	}
}

func TestSubmitPromptQueuedSubmitSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	sess, err := s.CreateSession(ctx, "session_submit_queued_cancel", "Test", map[string]any{"model": "bootstrap", "status": "queued", "queue_count": 1})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_submit_queued_cancel_existing", sess.ID, "queued", "existing", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create existing queued turn: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	engine := New(s)
	result, err := engine.SubmitPrompt(cancelCtx, RunInput{SessionID: sess.ID, Prompt: "later", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit queued prompt with canceled context: %v", err)
	}
	if !result.Queued || result.Status != "queued" {
		t.Fatalf("expected queued submit result, got %#v", result)
	}
	turns, err := s.ListTurns(ctx, sess.ID)
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected existing queued turn plus new queued turn, got %#v", turns)
	}
	sessRec, err := s.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "queued" || sessRec.State["active_turn_id"] != nil || sessRec.State["model"] != "bootstrap" {
		t.Fatalf("expected queued submit state to persist despite canceled caller context, got %#v", sessRec.State)
	}
	if got := sessRec.State["queue_count"]; got != float64(2) && got != 2 {
		t.Fatalf("expected queue_count 2 after canceled queued submit, got %#v", sessRec.State)
	}
}

func TestSubmitPromptPublishesQueuedTurnSubmittedTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_submit_topic", "Test", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	if _, err := s.CreateTurnWithStatus(ctx, "turn_submit_topic_active", "session_submit_topic", "queued", "first", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create first queued turn: %v", err)
	}
	turnTopicCh, unsub := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 16, SessionID: "session_submit_topic"})
	defer unsub()
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_submit_topic", Prompt: "second", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit queued prompt: %v", err)
	}
	if !result.Queued {
		t.Fatalf("expected queued submit result, got %#v", result)
	}
	deadline := time.After(2 * time.Second)
	for {
		select {
		case env := <-turnTopicCh:
			if env.Payload["type"] == "turn_submitted" && env.Payload["turn_id"] == result.TurnID && env.Payload["status"] == "queued" && env.Payload["phase"] == "queued" && env.Payload["queued"] == true {
				return
			}
		case <-deadline:
			t.Fatal("expected queued submit to publish turn_submitted runtime topic")
		}
	}
}

func TestStageQueuedSteeringContinuationMarksSessionQueued(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_continue_queued_state", "ContinueState", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_continue_queued_prev", "session_continue_queued_state", "completed", "previous", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create previous turn: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_continue_queued_state", "turn_continue_queued_prev", "user", "continue me", map[string]any{"intent": "continue", "model": "bootstrap"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}
	engine := New(s)
	staged, turnID, err := engine.stageQueuedSteeringContinuation(ctx, "session_continue_queued_state")
	if err != nil {
		t.Fatalf("stage queued steering continuation: %v", err)
	}
	if !staged || turnID == "" {
		t.Fatalf("expected staged continuation turn, got staged=%v id=%q", staged, turnID)
	}
	sessRec, err := s.GetSession(ctx, "session_continue_queued_state")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "queued" || sessRec.State["active_turn_id"] != nil || sessRec.State["model"] != "bootstrap" {
		t.Fatalf("expected staged continuation to mark session queued with cleared active turn, got %#v", sessRec.State)
	}
	if got := sessRec.State["queue_count"]; got != float64(1) && got != 1 {
		t.Fatalf("expected queue_count 1 after staged continuation, got %#v", sessRec.State)
	}
	stagedTurn, err := s.GetTurn(ctx, turnID)
	if err != nil {
		t.Fatalf("get staged continuation turn: %v", err)
	}
	if stagedTurn.Status != "queued" || stagedTurn.Phase != "queued" {
		t.Fatalf("expected staged continuation turn queued, got %#v", stagedTurn)
	}
}

func TestContinueSessionNormalizesRunningStateWhenContinuationTurnIsExternallyClaimed(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_continue_claimed", "ContinueState", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_continue_claimed_prev", "session_continue_claimed", "completed", "previous", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create previous turn: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_continue_claimed", "turn_continue_claimed_prev", "user", "continue me", map[string]any{"intent": "continue", "model": "bootstrap"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}
	engine := New(s)
	engine.beforeLaunchClaimHook = func(ctx context.Context, sessionID, turnID string) {
		if sessionID == "session_continue_claimed" {
			if _, err := s.ClaimSessionActiveTurn(ctx, sessionID, turnID, "external", turnID); err != nil {
				t.Fatalf("claim continuation turn inside launch hook: %v", err)
			}
		}
	}
	continued, err := engine.ContinueSession(ctx, "session_continue_claimed")
	if err != nil {
		t.Fatalf("continue session: %v", err)
	}
	if !continued {
		t.Fatal("expected continuation path to report progress")
	}
	sessRec, err := s.GetSession(ctx, "session_continue_claimed")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "running" || sessRec.State["active_turn_id"] == nil || sessRec.State["model"] != "bootstrap" {
		t.Fatalf("expected continuation handoff to normalize running session state, got %#v", sessRec.State)
	}
	if got := sessRec.State["queue_count"]; got != float64(1) && got != 1 {
		t.Fatalf("expected externally claimed continuation turn to remain queued-count-visible until it starts, got %#v", sessRec.State)
	}
	turns, err := s.ListTurns(ctx, "session_continue_claimed")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected previous turn plus staged continuation turn, got %#v", turns)
	}
	foundQueued := false
	for _, turn := range turns {
		if turn.ID != "turn_continue_claimed_prev" && turn.Status == "queued" {
			foundQueued = true
		}
	}
	if !foundQueued {
		t.Fatalf("expected staged continuation turn to remain queued when externally claimed before local launch, got %#v", turns)
	}
}

func TestStartNextQueuedTurnLockedNormalizesClaimedRunningSession(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_startnext_claimed_model", "Test", map[string]any{"model": "stale-model", "status": "queued", "active_turn_id": nil, "queue_count": 2}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_startnext_claimed_active", "session_startnext_claimed_model", "running", "active", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create active turn: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_startnext_claimed_queued", "session_startnext_claimed_model", "queued", "queued", map[string]any{"intent": "prompt", "model": "queued-model"}); err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_startnext_claimed_model", activeTurn.ID, "external", activeTurn.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	engine := New(s)
	runner := engine.runner("session_startnext_claimed_model")
	launched, err := engine.startNextQueuedTurnLocked(ctx, runner, "session_startnext_claimed_model")
	if err != nil {
		t.Fatalf("start next queued turn locked: %v", err)
	}
	if launched {
		t.Fatal("expected no local launch when another worker already owns the session")
	}
	sessRec, err := s.GetSession(ctx, "session_startnext_claimed_model")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "running" || sessRec.State["active_turn_id"] != activeTurn.ID || sessRec.State["model"] != "bootstrap" {
		t.Fatalf("expected claimed running session normalization, got %#v", sessRec.State)
	}
	if got := sessRec.State["queue_count"]; got != float64(1) && got != 1 {
		t.Fatalf("expected queue_count to normalize to remaining queued work, got %#v", sessRec.State)
	}
}

func TestContinueQueuedSteeringHandoffPreservesSessionModel(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_continue_steering_handoff_model", "Test", map[string]any{"model": "stale-model", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_continue_steering_handoff_prev", "session_continue_steering_handoff_model", "completed", "previous", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create previous turn: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_continue_steering_handoff_model", "turn_continue_steering_handoff_prev", "user", "continue me", map[string]any{"intent": "continue", "model": "bootstrap"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}
	engine := New(s)
	engine.beforeLaunchClaimHook = func(ctx context.Context, sessionID, turnID string) {
		if sessionID == "session_continue_steering_handoff_model" {
			if _, err := s.ClaimSessionActiveTurn(ctx, sessionID, turnID, "external", turnID); err != nil {
				t.Fatalf("claim continuation turn inside launch hook: %v", err)
			}
		}
	}
	continued, err := engine.ContinueSession(ctx, "session_continue_steering_handoff_model")
	if err != nil {
		t.Fatalf("continue session: %v", err)
	}
	if !continued {
		t.Fatal("expected continuation path to report progress")
	}
	sessRec, err := s.GetSession(ctx, "session_continue_steering_handoff_model")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "running" || sessRec.State["active_turn_id"] == nil || sessRec.State["model"] != "bootstrap" {
		t.Fatalf("expected steering continuation handoff to preserve active turn model, got %#v", sessRec.State)
	}
}

func TestStageQueuedSteeringContinuationPreservesSessionModelWhenSteeringOmitsModel(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_continue_model_preserve", "ContinueState", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_continue_model_preserve_prev", "session_continue_model_preserve", "completed", "previous", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create previous turn: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_continue_model_preserve", "turn_continue_model_preserve_prev", "user", "continue me", map[string]any{"intent": "continue"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}
	engine := New(s)
	if staged, turnID, err := engine.stageQueuedSteeringContinuation(ctx, "session_continue_model_preserve"); err != nil {
		t.Fatalf("stage queued steering continuation: %v", err)
	} else if !staged || turnID == "" {
		t.Fatalf("expected staged continuation turn, got staged=%v id=%q", staged, turnID)
	}
	sessRec, err := s.GetSession(ctx, "session_continue_model_preserve")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["model"] != "bootstrap" || sessRec.State["status"] != "queued" {
		t.Fatalf("expected staged continuation to preserve session model, got %#v", sessRec.State)
	}
}

func TestSubmitPromptSteeringNormalizesRunningSessionState(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_submit_steering_state", "Test", map[string]any{"model": "old-model", "status": "queued", "active_turn_id": nil, "queue_count": 1}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_submit_steering_state", "session_submit_steering_state", "running", "one", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create active turn: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_submit_steering_state", activeTurn.ID, "runner", activeTurn.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	engine := New(s)
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_submit_steering_state", Prompt: "two", Model: "ignored-model"})
	if err != nil {
		t.Fatalf("submit second prompt: %v", err)
	}
	if result.Queued || result.Status != "running" || result.TurnID != activeTurn.ID {
		t.Fatalf("expected steering result against active turn, got %#v", result)
	}
	sessRec, err := s.GetSession(ctx, "session_submit_steering_state")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "running" || sessRec.State["active_turn_id"] != activeTurn.ID || sessRec.State["model"] != "bootstrap" {
		t.Fatalf("expected steering submit to normalize running session state, got %#v", sessRec.State)
	}
	if got := sessRec.State["queue_count"]; got != float64(0) && got != 0 {
		t.Fatalf("expected steering submit to clear stale queue_count, got %#v", sessRec.State)
	}
}

func TestSubmitPromptSteeringSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_submit_steering_cancel", "Test", map[string]any{"model": "bootstrap", "status": "running"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_submit_steering_cancel", "session_submit_steering_cancel", "running", "one", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create active turn: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_submit_steering_cancel", activeTurn.ID, "runner", activeTurn.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	engine := New(s)
	result, err := engine.SubmitPrompt(cancelCtx, RunInput{SessionID: "session_submit_steering_cancel", Prompt: "two", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit prompt with canceled context: %v", err)
	}
	if result.Queued || result.Status != "running" || result.TurnID != activeTurn.ID {
		t.Fatalf("expected steering result against active turn, got %#v", result)
	}
	if depth, err := s.SteeringQueueLength(context.Background(), "session_submit_steering_cancel"); err != nil {
		t.Fatalf("steering queue length: %v", err)
	} else if depth != 1 {
		t.Fatalf("expected steering to persist despite canceled caller context, got depth %d", depth)
	}
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

func TestParsingHelpersNormalizeAsExpected(t *testing.T) {
	if got := normalizeAgentID(" @Agent-One "); got != "agent-one" {
		t.Fatalf("expected normalized agent id, got %q", got)
	}
	if got := normalizeDirectKind("  PEER-message "); got != DirectKindPeerMessage {
		t.Fatalf("expected normalized direct kind, got %q", got)
	}
	if got := normalizeDirectSourceKind("  SYSTEM "); got != DirectSourceKindSystem {
		t.Fatalf("expected normalized direct source kind, got %q", got)
	}
	if target, body, ok := parseDirectedPrompt("  @Agent-One: hello there  "); !ok || target != "agent-one" || body != "hello there" {
		t.Fatalf("expected directed prompt parse, got target=%q body=%q ok=%v", target, body, ok)
	}
	if typ, id, ok := splitScopedValue(" Room:Eng ", "space"); !ok || typ != "room" || id != "eng" {
		t.Fatalf("expected scoped value parse, got type=%q id=%q ok=%v", typ, id, ok)
	}
	if typ, id, ok := splitScopedValue("Thread-7", "group"); !ok || typ != "group" || id != "thread-7" {
		t.Fatalf("expected fallback scoped value parse, got type=%q id=%q ok=%v", typ, id, ok)
	}
}

func TestCoordinationContextPrefersActiveCallerAndFallsBackWhenNeeded(t *testing.T) {
	activeCtx := context.Background()
	fallbackCtx := context.Background()
	if got := coordinationContext(activeCtx, fallbackCtx); got != activeCtx {
		t.Fatal("expected active caller context to win")
	}
	cancelCtx, cancel := context.WithCancel(activeCtx)
	cancel()
	if got := coordinationContext(cancelCtx, fallbackCtx); got != fallbackCtx {
		t.Fatal("expected canceled caller context to fall back")
	}
	if got := coordinationContext(nil, fallbackCtx); got != fallbackCtx {
		t.Fatal("expected nil caller context to fall back")
	}
	cancelFallback, cancelFallbackFn := context.WithCancel(activeCtx)
	cancelFallbackFn()
	if got := coordinationContext(cancelCtx, cancelFallback); got != nil {
		t.Fatalf("expected nil when both caller and fallback contexts are unusable, got %#v", got)
	}
}

func TestSessionIdentityHelpersSurviveCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateDefaultSession("agentx", "web", "acctx", "session_identity_ctx")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_identity_ctx", Title: "@agentx", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session with canonical identity: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	if got := sessionAgentIDWithStoreFallback(cancelCtx, ctx, s, sess); got != "agentx" {
		t.Fatalf("expected canonical agent id under canceled caller context, got %q", got)
	}
	if got := sessionChannelWithStoreFallback(cancelCtx, ctx, s, sess); got != "web" {
		t.Fatalf("expected canonical channel under canceled caller context, got %q", got)
	}
	if got := sessionAccountWithStoreFallback(cancelCtx, ctx, s, sess); got != "acctx" {
		t.Fatalf("expected canonical account under canceled caller context, got %q", got)
	}
}

func TestResolveTurnAgentAndModelSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateDefaultSession("agentmodel", "web", "acctmodel", "session_agent_model_ctx")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_agent_model_ctx", Title: "@agentmodel", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session with canonical identity: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_agent_model_ctx", sess.ID, "queued", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	engine := New(s)
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	agentID, model := engine.runner(sess.ID).resolveTurnAgentAndModel(cancelCtx, s, turnRec, sess.ID, turnRec.Prompt)
	if agentID != "agentmodel" {
		t.Fatalf("expected canonical agent id under canceled caller context, got %q", agentID)
	}
	if model == "" {
		t.Fatal("expected non-empty model resolution")
	}
}

func TestLaunchTurnLockedSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_launch_ctx", "Test", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_launch_ctx", "session_launch_ctx", "queued", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	engine := New(s)
	runner := engine.runner("session_launch_ctx")
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	launched, err := engine.launchTurnLocked(cancelCtx, runner, "session_launch_ctx", queuedTurn.ID)
	if err != nil {
		t.Fatalf("launch turn with canceled caller context: %v", err)
	}
	if !launched {
		t.Fatalf("expected queued turn to launch despite canceled caller context")
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, queuedTurn.ID)
		return err == nil && strings.TrimSpace(turnRec.FinishedAt) != ""
	}, "launched turn completion with canceled caller context")
	turnRec, err := s.GetTurn(ctx, queuedTurn.ID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if turnRec.Status != "completed" {
		t.Fatalf("expected launched turn to complete, got %#v", turnRec)
	}
	if _, _, err := s.GetSessionActiveTurn(ctx, "session_launch_ctx"); err != sql.ErrNoRows {
		t.Fatalf("expected no lingering active turn after launched completion, got err=%v", err)
	}
}

func TestConvertLaunchConflictToSteeringSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_claim_conflict_ctx", "Test", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_existing_active_ctx", "session_claim_conflict_ctx", "running", "already running", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create existing active turn: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_transient_ctx", "session_claim_conflict_ctx", "queued", "steer me", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create transient queued turn: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_claim_conflict_ctx", activeTurn.ID, "runner", activeTurn.ID); err != nil {
		t.Fatalf("claim existing active turn: %v", err)
	}
	if err := s.TouchSessionState(ctx, "session_claim_conflict_ctx", map[string]any{"active_turn_id": activeTurn.ID, "status": "running"}); err != nil {
		t.Fatalf("touch session state: %v", err)
	}
	engine := New(s)
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	res, steered, err := engine.convertLaunchConflictToSteering(cancelCtx, queuedTurn.ID, RunInput{SessionID: "session_claim_conflict_ctx", Prompt: "steer me", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("convert launch conflict with canceled caller context: %v", err)
	}
	if !steered || res == nil || res.TurnID != activeTurn.ID || res.Status != "running" {
		t.Fatalf("expected steering fallback to existing active turn, got steered=%v res=%#v", steered, res)
	}
	turns, err := s.ListTurns(ctx, "session_claim_conflict_ctx")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 1 || turns[0].ID != activeTurn.ID {
		t.Fatalf("expected transient queued turn removed after canceled-context steering fallback, got %#v", turns)
	}
	if depth, err := s.SteeringQueueLength(ctx, "session_claim_conflict_ctx"); err != nil {
		t.Fatalf("steering queue length: %v", err)
	} else if depth != 1 {
		t.Fatalf("expected steering queue depth 1 after canceled-context claim-conflict fallback, got %d", depth)
	}
}

func TestConvertLaunchConflictToSteeringPreservesSuccessWhenTransientTurnCleanupFails(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_claim_conflict_cleanup", "Test", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_existing_active_cleanup", "session_claim_conflict_cleanup", "running", "already running", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create existing active turn: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_transient_cleanup", "session_claim_conflict_cleanup", "queued", "steer me", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create transient queued turn: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_claim_conflict_cleanup", activeTurn.ID, "runner", activeTurn.ID); err != nil {
		t.Fatalf("claim existing active turn: %v", err)
	}
	if err := s.TouchSessionState(ctx, "session_claim_conflict_cleanup", map[string]any{"active_turn_id": activeTurn.ID, "status": "running"}); err != nil {
		t.Fatalf("touch session state: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `
		create trigger fail_delete_transient_turn_cleanup
		before delete on turns
		for each row when old.id = 'turn_transient_cleanup'
		begin
			select raise(fail, 'delete blocked for test');
		end;
	`); err != nil {
		t.Fatalf("create delete-failure trigger: %v", err)
	}
	engine := New(s)
	res, steered, err := engine.convertLaunchConflictToSteering(ctx, queuedTurn.ID, RunInput{SessionID: "session_claim_conflict_cleanup", Prompt: "steer me", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("expected steering fallback success despite transient cleanup failure, got %v", err)
	}
	if !steered || res == nil || res.TurnID != activeTurn.ID || res.Status != "running" {
		t.Fatalf("expected steering fallback to existing active turn, got steered=%v res=%#v", steered, res)
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

func TestCleanupTurnRunStopsAfterActiveClaimReleaseFailure(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_cleanup_release_fail", "Test", map[string]any{"model": "bootstrap", "status": "running"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_cleanup_release_fail", "session_cleanup_release_fail", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	engine := New(s)
	hookCalled := false
	engine.beforeCleanupNextWorkHook = func(ctx context.Context, sessionID string) {
		hookCalled = true
	}
	runner := engine.runner("session_cleanup_release_fail")
	active := &runningTurn{turnID: "turn_cleanup_release_fail"}
	runner.current = active
	if _, err := s.DB().ExecContext(ctx, `
		create trigger fail_release_session_active_turn_cleanup
		before delete on session_active_turns
		for each row when old.session_id = 'session_cleanup_release_fail'
		begin
			select raise(fail, 'release blocked for test');
		end;
	`); err != nil {
		t.Fatalf("create release-failure trigger: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_cleanup_release_fail", "turn_cleanup_release_fail", "runner", "turn_cleanup_release_fail"); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	runner.cleanupTurnRun("session_cleanup_release_fail", "turn_cleanup_release_fail", active)
	if hookCalled {
		t.Fatal("expected cleanup coordination to stop before next-work hook when claim release fails")
	}
	if runner.current != nil {
		t.Fatalf("expected cleanup to still clear current running turn, got %#v", runner.current)
	}
	if _, _, err := s.GetSessionActiveTurn(ctx, "session_cleanup_release_fail"); err != nil {
		t.Fatalf("expected active claim to remain when release is blocked, got %v", err)
	}
	events, err := s.ListTurnEvents(ctx, "turn_cleanup_release_fail")
	if err != nil {
		t.Fatalf("list cleanup failure events: %v", err)
	}
	found := false
	for _, event := range events {
		if event.Type == "turn.cleanup_handoff_failed" && event.Payload["stage"] == "release_active_claim" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected cleanup handoff failure event for release failure, got %#v", events)
	}
}

func TestCleanupTurnRunStopsAfterQueueSyncFailure(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_cleanup_sync_fail", "Test", map[string]any{"model": "bootstrap", "status": "running"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_cleanup_sync_fail", "session_cleanup_sync_fail", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_cleanup_sync_fail", "turn_cleanup_sync_fail", "runner", "turn_cleanup_sync_fail"); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	engine := New(s)
	hookCalled := false
	engine.beforeCleanupNextWorkHook = func(ctx context.Context, sessionID string) {
		hookCalled = true
	}
	runner := engine.runner("session_cleanup_sync_fail")
	active := &runningTurn{turnID: "turn_cleanup_sync_fail"}
	runner.current = active
	if _, err := s.DB().ExecContext(ctx, `
		create trigger fail_cleanup_sync_session_update
		before update on sessions
		for each row when old.id = 'session_cleanup_sync_fail'
		begin
			select raise(fail, 'queue sync blocked for test');
		end;
	`); err != nil {
		t.Fatalf("create queue-sync-failure trigger: %v", err)
	}
	runner.cleanupTurnRun("session_cleanup_sync_fail", "turn_cleanup_sync_fail", active)
	if hookCalled {
		t.Fatal("expected cleanup coordination to stop before next-work hook when queue sync fails")
	}
	if runner.current != nil {
		t.Fatalf("expected cleanup to still clear current running turn, got %#v", runner.current)
	}
	events, err := s.ListTurnEvents(ctx, "turn_cleanup_sync_fail")
	if err != nil {
		t.Fatalf("list cleanup failure events: %v", err)
	}
	found := false
	for _, event := range events {
		if event.Type == "turn.cleanup_handoff_failed" && event.Payload["stage"] == "sync_queue_count" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected cleanup handoff failure event for queue sync failure, got %#v", events)
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
	started := make(chan struct{}, 1)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		select {
		case started <- struct{}{}:
		default:
		}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "recovered done"}}}}, nil
	})
	_, err := s.CreateSession(ctx, "session_recover_compact", "Test", map[string]any{"model": "bootstrap", "status": "running"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_recover_compact", "session_recover_compact", "running", "hello", map[string]any{"intent": "prompt", "model": "mock-recover-compact"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, turnRec.ID, "running", "compacting"); err != nil {
		t.Fatalf("set compacting phase: %v", err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "session_recover_compact", turnRec.ID, "runner", turnRec.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	} else if !ok {
		t.Fatal("expected active turn claim to be acquired")
	}
	if _, err := s.DB().ExecContext(ctx, `update session_active_turns set updated_at = '2000-01-01T00:00:00Z' where session_id = ?`, "session_recover_compact"); err != nil {
		t.Fatalf("age active turn claim: %v", err)
	}

	_ = New(s)

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("expected recovered compacting turn to restart")
	}
	waitForCondition(t, 2*time.Second, func() bool {
		recovered, err := s.GetTurn(ctx, turnRec.ID)
		return err == nil && recovered.Status == "completed"
	}, "recovered compacting turn completion")
	waitForCondition(t, 2*time.Second, func() bool {
		_, _, err := s.GetSessionActiveTurn(ctx, "session_recover_compact")
		return err == sql.ErrNoRows
	}, "recovered compacting turn active-claim release")
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

func TestRecoverInterruptedTurnsReturnsErrorWhenHoldMarkerPersistenceFails(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	sess, err := s.CreateSession(ctx, "session_recover_tool_marker_fail", "Test", map[string]any{"model": "bootstrap", "status": "running"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_recover_tool_marker_fail", sess.ID, "running", "hello", map[string]any{"intent": "prompt"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, turnRec.ID, "running", "waiting_on_tools"); err != nil {
		t.Fatalf("set waiting_on_tools phase: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, sess.ID, turnRec.ID, "runner", turnRec.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update session_active_turns set updated_at = '2000-01-01T00:00:00Z' where session_id = ?`, sess.ID); err != nil {
		t.Fatalf("age active turn claim: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `
		create trigger fail_recovery_hold_marker
		before insert on turn_failures
		for each row when new.turn_id = 'turn_recover_tool_marker_fail'
		begin
			select raise(fail, 'hold marker blocked for test');
		end;
	`); err != nil {
		t.Fatalf("create hold-marker trigger: %v", err)
	}
	engine := New(s)
	turnTopicCh, unsubTurn := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 8, SessionID: sess.ID})
	defer unsubTurn()
	sessionTopicCh, unsubSession := engine.Topics().Subscribe(ctx, "runtime.session", topics.SubscribeOptions{Buffer: 8, SessionID: sess.ID})
	defer unsubSession()
	recovered, err := engine.recoverInterruptedTurns(ctx, sess.ID)
	if recovered {
		t.Fatal("expected failed recovery not to report success")
	}
	if err == nil || !strings.Contains(err.Error(), "hold marker blocked for test") {
		t.Fatalf("expected recovery to surface hold-marker failure, got %v", err)
	}
	current, err := s.GetTurn(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("get turn after failed recovery: %v", err)
	}
	if current.Status != "running" || current.Phase != "waiting_on_tools" {
		t.Fatalf("expected turn state unchanged after failed recovery marker write, got %#v", current)
	}
	events, err := s.ListTurnEvents(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("list recovery failure events: %v", err)
	}
	found := false
	for _, event := range events {
		if event.Type == "turn.recovery_failed" && event.Payload["recovery_disposition"] == "hold_for_retry_or_skip_after_tool_checkpoint" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected recovery failure audit event, got %#v", events)
	}
	foundTurnFailed := false
	foundSessionFailed := false
	deadline := time.After(time.Second)
	for !(foundTurnFailed && foundSessionFailed) {
		select {
		case env := <-turnTopicCh:
			if env.Payload["type"] == "turn_recovery_failed" && env.Payload["turn_id"] == turnRec.ID && env.Payload["recovery_disposition"] == "hold_for_retry_or_skip_after_tool_checkpoint" {
				foundTurnFailed = true
			}
		case env := <-sessionTopicCh:
			if env.Payload["type"] == "session_state" && env.Payload["reason"] == "recovery_failed" && env.Payload["turn_id"] == turnRec.ID && env.Payload["recovery_disposition"] == "hold_for_retry_or_skip_after_tool_checkpoint" {
				foundSessionFailed = true
			}
		case <-deadline:
			t.Fatalf("expected runtime recovery failure notices, got turn=%v session=%v", foundTurnFailed, foundSessionFailed)
		}
	}
}

func TestRecoverInterruptedTurnsEmitsSessionScanFailureSummaryForMixedOutcomes(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	started := make(chan struct{}, 1)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		select {
		case started <- struct{}{}:
		default:
		}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "queued done"}}}}, nil
	})
	engine := New(s)

	goodSess, err := s.CreateSession(ctx, "session_recover_mixed_good", "Recover", map[string]any{"status": "running"})
	if err != nil {
		t.Fatalf("create good session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_recover_mixed_good_terminal", goodSess.ID, "cancelled", "terminal", map[string]any{"intent": "prompt", "model": "mock-recover"}); err != nil {
		t.Fatalf("create good terminal turn: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_recover_mixed_good_queued", goodSess.ID, "queued", "queued after recovery", map[string]any{"intent": "prompt", "model": "mock-recover"}); err != nil {
		t.Fatalf("create good queued turn: %v", err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, goodSess.ID, "turn_recover_mixed_good_terminal", "worker-test", "claim-recover-mixed-good"); err != nil {
		t.Fatalf("claim good active turn: %v", err)
	} else if !ok {
		t.Fatal("expected good active turn claim")
	}
	staleTime := time.Now().Add(-(interruptedTurnStaleAfter + 5*time.Second)).UTC().Format(time.RFC3339Nano)
	if err := s.TouchSessionState(ctx, goodSess.ID, map[string]any{"active_turn_id": "turn_recover_mixed_good_terminal", "status": "running"}); err != nil {
		t.Fatalf("touch good session state: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update session_active_turns set updated_at = ? where session_id = ?`, staleTime, goodSess.ID); err != nil {
		t.Fatalf("age good active turn claim: %v", err)
	}

	badSess, err := s.CreateSession(ctx, "session_recover_mixed_bad", "Recover", map[string]any{"status": "running", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create bad session: %v", err)
	}
	badTurn, err := s.CreateTurnWithStatus(ctx, "turn_recover_mixed_bad", badSess.ID, "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create bad turn: %v", err)
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, badTurn.ID, "running", "waiting_on_tools"); err != nil {
		t.Fatalf("set bad waiting_on_tools phase: %v", err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, badSess.ID, badTurn.ID, "runner", badTurn.ID); err != nil {
		t.Fatalf("claim bad active turn: %v", err)
	} else if !ok {
		t.Fatal("expected bad active turn claim")
	}
	if err := s.TouchSessionState(ctx, badSess.ID, map[string]any{"active_turn_id": badTurn.ID, "status": "running"}); err != nil {
		t.Fatalf("touch bad session state: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update session_active_turns set updated_at = ? where session_id = ?`, staleTime, badSess.ID); err != nil {
		t.Fatalf("age bad active turn claim: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `
		create trigger fail_recovery_hold_marker_mixed
		before insert on turn_failures
		for each row when new.turn_id = 'turn_recover_mixed_bad'
		begin
			select raise(fail, 'mixed hold marker blocked for test');
		end;
	`); err != nil {
		t.Fatalf("create mixed hold-marker trigger: %v", err)
	}
	sessionTopicCh, unsubSession := engine.Topics().Subscribe(ctx, "runtime.session", topics.SubscribeOptions{Buffer: 16, SessionID: badSess.ID})
	defer unsubSession()

	recovered, err := engine.recoverInterruptedTurns(ctx, "")
	if !recovered {
		t.Fatal("expected mixed recovery scan to report at least one recovered claim")
	}
	if err == nil || !strings.Contains(err.Error(), "mixed hold marker blocked for test") {
		t.Fatalf("expected mixed recovery scan to surface failing claim, got %v", err)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("expected recovered session queued work to start despite mixed outcomes")
	}
	foundSummary := false
	deadline := time.After(time.Second)
	for !foundSummary {
		select {
		case env := <-sessionTopicCh:
			if env.Payload["type"] == "session_state" && env.Payload["reason"] == "recovery_scan_failed" {
				if env.Payload["failed_claim_count"] != 1 && env.Payload["failed_claim_count"] != float64(1) {
					t.Fatalf("expected failed_claim_count 1, got %#v", env)
				}
				if env.Payload["recovered_claim_count"] != 0 && env.Payload["recovered_claim_count"] != float64(0) {
					t.Fatalf("expected recovered_claim_count 0 for failed session, got %#v", env)
				}
				foundSummary = true
			}
		case <-deadline:
			t.Fatal("expected session-level mixed recovery failure summary")
		}
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

func TestContinueSessionRejectsMissingSessionNormalization(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	continued, err := engine.ContinueSession(ctx, "missing-session")
	if continued {
		t.Fatal("expected missing-session continue to report no progress")
	}
	if err == nil || !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected missing-session continue to surface sql.ErrNoRows, got %v", err)
	}
}

func TestContinueSessionAppendsSteeringContinuedEvent(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_continue_event", "Test", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_continue_event_prev", "session_continue_event", "completed", "previous", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create previous turn: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_continue_event", "turn_continue_event_prev", "user", "continue please", map[string]any{"intent": "prompt", "model": "bootstrap"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}
	engine := New(s)
	continued, err := engine.ContinueSession(ctx, "session_continue_event")
	if err != nil {
		t.Fatalf("continue session: %v", err)
	}
	if !continued {
		t.Fatal("expected ContinueSession to continue queued steering")
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turns, err := s.ListTurns(ctx, "session_continue_event")
		if err != nil {
			return false
		}
		for _, turn := range turns {
			if turn.ID != "turn_continue_event_prev" {
				events, err := s.ListTurnEvents(ctx, turn.ID)
				if err != nil {
					return false
				}
				for _, event := range events {
					if event.Type == "steering.continued" {
						return true
					}
				}
			}
		}
		return false
	}, "steering continued audit event")
}

func TestContinueSessionClearsQueueCountAfterLaunchingContinuation(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_continue_queue_count", "Test", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_continue_queue_count_prev", "session_continue_queue_count", "completed", "previous", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create previous turn: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_continue_queue_count", "turn_continue_queue_count_prev", "user", "continue please", map[string]any{"intent": "prompt", "model": "bootstrap"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}
	engine := New(s)
	continued, err := engine.ContinueSession(ctx, "session_continue_queue_count")
	if err != nil {
		t.Fatalf("continue session: %v", err)
	}
	if !continued {
		t.Fatal("expected ContinueSession to start queued steering")
	}
	waitForCondition(t, 2*time.Second, func() bool {
		sessRec, err := s.GetSession(ctx, "session_continue_queue_count")
		if err != nil {
			return false
		}
		got := sessRec.State["queue_count"]
		return got == float64(0) || got == 0
	}, "continuation launch queue_count normalization")
	sessRec, err := s.GetSession(ctx, "session_continue_queue_count")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if got := sessRec.State["queue_count"]; got != float64(0) && got != 0 {
		t.Fatalf("expected launched continuation to clear queue_count, got %#v", sessRec.State)
	}
}

func TestContinueSessionNormalizesRunningStateWhenQueuedTurnIsExternallyClaimed(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_continue_queued_claimed", "Test", map[string]any{"model": "bootstrap", "status": "queued", "active_turn_id": nil, "queue_count": 1}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_continue_queued_claimed", "session_continue_queued_claimed", "queued", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	engine := New(s)
	engine.beforeLaunchClaimHook = func(ctx context.Context, sessionID, turnID string) {
		if sessionID == "session_continue_queued_claimed" && turnID == queuedTurn.ID {
			if _, err := s.ClaimSessionActiveTurn(ctx, sessionID, turnID, "external", turnID); err != nil {
				t.Fatalf("claim queued turn inside launch hook: %v", err)
			}
			if err := s.MarkTurnClaimed(ctx, turnID, "external"); err != nil {
				t.Fatalf("mark queued turn claimed inside launch hook: %v", err)
			}
			if err := s.UpdateTurnStatusAndPhase(ctx, turnID, "running", "setup"); err != nil {
				t.Fatalf("mark queued turn running inside launch hook: %v", err)
			}
		}
	}
	continued, err := engine.ContinueSession(ctx, "session_continue_queued_claimed")
	if err != nil {
		t.Fatalf("continue session: %v", err)
	}
	if !continued {
		t.Fatal("expected queued continue handoff to report progress")
	}
	sessRec, err := s.GetSession(ctx, "session_continue_queued_claimed")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "running" || sessRec.State["active_turn_id"] != queuedTurn.ID || sessRec.State["model"] != "bootstrap" {
		t.Fatalf("expected queued continue handoff to normalize running session state, got %#v", sessRec.State)
	}
	if got := sessRec.State["queue_count"]; got != float64(0) && got != 0 {
		t.Fatalf("expected queued continue handoff to clear stale queue_count, got %#v", sessRec.State)
	}
}

func TestContinueSessionNormalizesRunningStateWhenActiveTurnExists(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_continue_active_normalize", "Test", map[string]any{"model": "bootstrap", "status": "queued", "active_turn_id": nil, "queue_count": 0}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_continue_active_normalize", "session_continue_active_normalize", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create active turn: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_continue_active_normalize", activeTurn.ID, "runner", activeTurn.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	engine := New(s)
	continued, err := engine.ContinueSession(ctx, "session_continue_active_normalize")
	if err != nil {
		t.Fatalf("continue session: %v", err)
	}
	if continued {
		t.Fatal("expected active continue to report no new work launched")
	}
	sessRec, err := s.GetSession(ctx, "session_continue_active_normalize")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "running" || sessRec.State["active_turn_id"] != activeTurn.ID || sessRec.State["model"] != "bootstrap" {
		t.Fatalf("expected active continue to normalize running session state, got %#v", sessRec.State)
	}
	if got := sessRec.State["queue_count"]; got != float64(0) && got != 0 {
		t.Fatalf("expected active continue to clear stale queue_count, got %#v", sessRec.State)
	}
}

func TestContinueSessionNormalizesIdleStateWhenNoWorkRemains(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_continue_idle_normalize", "Test", map[string]any{"model": "bootstrap", "status": "queued", "active_turn_id": "stale-turn", "queue_count": 1}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	continued, err := engine.ContinueSession(ctx, "session_continue_idle_normalize")
	if err != nil {
		t.Fatalf("continue session: %v", err)
	}
	if continued {
		t.Fatal("expected no continuation work")
	}
	sessRec, err := s.GetSession(ctx, "session_continue_idle_normalize")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "idle" || sessRec.State["active_turn_id"] != nil {
		t.Fatalf("expected idle continue to normalize session idle state, got %#v", sessRec.State)
	}
	if got := sessRec.State["queue_count"]; got != float64(0) && got != 0 {
		t.Fatalf("expected idle continue to clear stale queue_count, got %#v", sessRec.State)
	}
}

func TestLaunchTurnLockedSetsSessionModelBeforeSetupRuns(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_launch_model_state", "Test", map[string]any{"model": "old-model", "status": "queued", "queue_count": 1}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_launch_model_state", "session_launch_model_state", "queued", "hello", map[string]any{"intent": "prompt", "model": "new-model"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	engine := New(s)
	releaseSetup := make(chan struct{})
	engine.beforeSetupHook = func(ctx context.Context, sessionID, turnID string) {
		if sessionID == "session_launch_model_state" && turnID == queuedTurn.ID {
			<-releaseSetup
		}
	}
	runner := engine.runner("session_launch_model_state")
	launched, err := engine.launchTurnLocked(ctx, runner, "session_launch_model_state", queuedTurn.ID)
	if err != nil {
		close(releaseSetup)
		t.Fatalf("launch queued turn: %v", err)
	}
	if !launched {
		close(releaseSetup)
		t.Fatal("expected queued turn to launch")
	}
	waitForCondition(t, 2*time.Second, func() bool {
		sessRec, err := s.GetSession(ctx, "session_launch_model_state")
		if err != nil {
			return false
		}
		return sessRec.State["status"] == "running" && sessRec.State["active_turn_id"] == queuedTurn.ID && sessRec.State["model"] == "new-model"
	}, "launch-time running session model")
	close(releaseSetup)
}

func TestContinueSessionClearsQueueCountAfterLaunchingQueuedTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_continue_launch_queue", "Test", map[string]any{"model": "bootstrap", "status": "queued", "queue_count": 1}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_continue_launch_queue", "session_continue_launch_queue", "queued", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	engine := New(s)
	continued, err := engine.ContinueSession(ctx, "session_continue_launch_queue")
	if err != nil {
		t.Fatalf("continue session: %v", err)
	}
	if !continued {
		t.Fatal("expected ContinueSession to launch queued turn")
	}
	waitForCondition(t, 2*time.Second, func() bool {
		sessRec, err := s.GetSession(ctx, "session_continue_launch_queue")
		if err != nil {
			return false
		}
		got := sessRec.State["queue_count"]
		return got == float64(0) || got == 0
	}, "queued turn launch queue_count normalization")
	sessRec, err := s.GetSession(ctx, "session_continue_launch_queue")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if got := sessRec.State["queue_count"]; got != float64(0) && got != 0 {
		t.Fatalf("expected launched queued turn to clear queue_count, got %#v", sessRec.State)
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

func TestContinueSessionSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_continue_ctx", "Test", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_continue_ctx", "", "user", "continue please", map[string]any{"intent": "prompt", "model": "bootstrap"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}
	engine := New(s)
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	continued, err := engine.ContinueSession(cancelCtx, "session_continue_ctx")
	if err != nil {
		t.Fatalf("continue session with canceled caller context: %v", err)
	}
	if !continued {
		t.Fatal("expected ContinueSession to stage/start queued steering despite canceled caller context")
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turns, err := s.ListTurns(ctx, "session_continue_ctx")
		return err == nil && len(turns) > 0
	}, "continued turn creation with canceled caller context")
	turns, err := s.ListTurns(ctx, "session_continue_ctx")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	foundSubmitted := false
	for _, turn := range turns {
		if turn.Status == "queued" || turn.Status == "running" || turn.Status == "completed" {
			foundSubmitted = true
		}
	}
	if !foundSubmitted {
		t.Fatalf("expected continuation turn to be staged/launched despite canceled caller context, got %#v", turns)
	}
}

func TestContinueSessionPublishesTurnSubmittedTopicForSteeringContinuation(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_continue_topics", "Test", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	turnTopicCh, unsub := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 16, SessionID: "session_continue_topics"})
	defer unsub()
	if _, err := s.EnqueueSteering(ctx, "session_continue_topics", "", "user", "continue please", map[string]any{"intent": "prompt", "model": "bootstrap"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}
	continued, err := engine.ContinueSession(ctx, "session_continue_topics")
	if err != nil {
		t.Fatalf("continue session: %v", err)
	}
	if !continued {
		t.Fatal("expected ContinueSession to start queued steering")
	}
	deadline := time.After(2 * time.Second)
	for {
		select {
		case env := <-turnTopicCh:
			if env.Payload["type"] == "turn_submitted" && env.Payload["status"] == "queued" && env.Payload["phase"] == "queued" && env.Payload["queued"] == true && env.Payload["continue"] == true {
				return
			}
		case <-deadline:
			t.Fatal("expected steering continuation to publish turn_submitted runtime topic")
		}
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

func TestCleanupTurnRunNormalizesRunningStateWhenNextTurnAlreadyClaimed(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_cleanup_claimed", "Test", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_cleanup_claimed_active", "session_cleanup_claimed", "completed", "active", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create active turn: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_cleanup_claimed_next", "session_cleanup_claimed", "running", "next", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create next turn: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_cleanup_claimed", activeTurn.ID, "runner", activeTurn.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	if err := s.TouchSessionState(ctx, "session_cleanup_claimed", map[string]any{"status": "idle", "active_turn_id": nil}); err != nil {
		t.Fatalf("seed session state: %v", err)
	}
	engine := New(s)
	engine.beforeCleanupNextWorkHook = func(ctx context.Context, sessionID string) {
		if _, err := s.ClaimSessionActiveTurn(ctx, sessionID, queuedTurn.ID, "external", queuedTurn.ID); err != nil {
			t.Fatalf("claim queued turn inside cleanup hook: %v", err)
		}
	}
	runner := engine.runner("session_cleanup_claimed")
	oldRunning := &runningTurn{turnID: activeTurn.ID}
	runner.current = oldRunning
	runner.cleanupTurnRun("session_cleanup_claimed", activeTurn.ID, oldRunning)
	sessRec, err := s.GetSession(ctx, "session_cleanup_claimed")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "running" || sessRec.State["active_turn_id"] != queuedTurn.ID || sessRec.State["model"] != "bootstrap" {
		t.Fatalf("expected cleanup to normalize session to externally claimed running turn, got %#v", sessRec.State)
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
	turnTopicCh, unsub := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 16, SessionID: turnRec.SessionID})
	defer unsub()
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
	foundResolved := false
	deadline := time.After(2 * time.Second)
	for !foundResolved {
		select {
		case env := <-turnTopicCh:
			if env.Payload["type"] == "turn_failure_resolved" && env.Payload["turn_id"] == turnRec.ID && env.Payload["resolution_state"] == "retried" && env.Payload["resolution_summary"] == "retry requested" && env.Payload["resolved_turn_id"] == result.TurnID {
				foundResolved = true
			}
		case <-deadline:
			t.Fatalf("expected retry hold resolution runtime checkpoint with resolved turn id")
		}
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
	if events, err := s.ListTurnEvents(ctx, queuedTurn.ID); err != nil {
		t.Fatalf("list turn events: %v", err)
	} else {
		found := false
		for _, event := range events {
			if event.Type == "turn.cancelled" {
				found = true
				if event.Payload["reason"] != "queued_cancel" || event.Payload["status"] != "cancelled" || event.Payload["turn_phase"] != "aborted" || event.Payload["failure_kind"] != "" {
					t.Fatalf("unexpected turn.cancelled payload, got %#v", event.Payload)
				}
			}
		}
		if !found {
			t.Fatalf("expected turn.cancelled event, got %#v", events)
		}
	}
	if turnRec.FinishedAt == "" {
		t.Fatalf("expected queued cancel to set finished_at, got %#v", turnRec)
	}
	sessRec, err := s.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "idle" {
		t.Fatalf("expected queued cancel to leave session idle, got %#v", sessRec)
	}
}

func TestCancelTurnRejectsSessionMismatch(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_cancel_mismatch_a", "A", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session a: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_cancel_mismatch_b", "B", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session b: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_cancel_mismatch", "session_cancel_mismatch_a", "queued", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	engine := New(s)
	err = engine.CancelTurn(ctx, "session_cancel_mismatch_b", queuedTurn.ID)
	if err == nil || !strings.Contains(err.Error(), "does not belong to session") {
		t.Fatalf("expected session mismatch error, got %v", err)
	}
	turnRec, err := s.GetTurn(ctx, queuedTurn.ID)
	if err != nil {
		t.Fatalf("get queued turn: %v", err)
	}
	if turnRec.Status != "queued" || turnRec.Phase != "queued" || strings.TrimSpace(turnRec.FinishedAt) != "" {
		t.Fatalf("expected mismatched cancel to leave queued turn untouched, got %#v", turnRec)
	}
}

func TestLaunchConflictSteeringFallbackClearsQueuedSessionState(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_launch_conflict_state", "Test", map[string]any{"model": "bootstrap", "status": "running", "active_turn_id": "turn_launch_conflict_active"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_launch_conflict_active", "session_launch_conflict_state", "running", "active", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create active turn: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_launch_conflict_state", activeTurn.ID, "runner", activeTurn.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	engine := New(s)
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_launch_conflict_state", Prompt: "steer me", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit prompt: %v", err)
	}
	if result.Queued || result.Status != "running" || result.TurnID != activeTurn.ID {
		t.Fatalf("expected steering fallback result against active turn, got %#v", result)
	}
	sessRec, err := s.GetSession(ctx, "session_launch_conflict_state")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "running" || sessRec.State["active_turn_id"] != activeTurn.ID {
		t.Fatalf("expected steering fallback to preserve running session state, got %#v", sessRec.State)
	}
	if got := sessRec.State["queue_count"]; got != float64(0) && got != 0 {
		t.Fatalf("expected steering fallback to clear queued count after deleting queued turn, got %#v", sessRec.State)
	}
	turns, err := s.ListTurns(ctx, "session_launch_conflict_state")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 1 || turns[0].ID != activeTurn.ID {
		t.Fatalf("expected only active turn after steering fallback deletes transient queued turn, got %#v", turns)
	}
}

func TestLaunchConflictSteeringFallbackPreservesSessionModelWhenInputModelIsEmpty(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_launch_conflict_model", "Test", map[string]any{"model": "old-model", "status": "running", "active_turn_id": "turn_launch_conflict_model_active", "queue_count": 0}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_launch_conflict_model_active", "session_launch_conflict_model", "running", "active", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create active turn: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_launch_conflict_model", activeTurn.ID, "runner", activeTurn.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	engine := New(s)
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_launch_conflict_model", Prompt: "steer me"})
	if err != nil {
		t.Fatalf("submit prompt: %v", err)
	}
	if result.Queued || result.Status != "running" || result.TurnID != activeTurn.ID {
		t.Fatalf("expected steering fallback result against active turn, got %#v", result)
	}
	sessRec, err := s.GetSession(ctx, "session_launch_conflict_model")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "running" || sessRec.State["active_turn_id"] != activeTurn.ID || sessRec.State["model"] != "bootstrap" {
		t.Fatalf("expected steering fallback to preserve active-turn model, got %#v", sessRec.State)
	}
}

func TestLaunchTurnLockedRollsBackClaimAndQueuedStateOnPreRunFailure(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_launch_rollback", "Test", map[string]any{"model": "bootstrap", "status": "queued", "queue_count": 1}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_launch_rollback", "session_launch_rollback", "queued", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	engine := New(s)
	engine.beforeLaunchSessionStateErrorHook = func(ctx context.Context, sessionID, turnID string) error {
		if sessionID == "session_launch_rollback" && turnID == queuedTurn.ID {
			return fmt.Errorf("boom")
		}
		return nil
	}
	runner := engine.runner("session_launch_rollback")
	launched, err := engine.launchTurnLocked(ctx, runner, "session_launch_rollback", queuedTurn.ID)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected injected launch failure, got launched=%v err=%v", launched, err)
	}
	if launched {
		t.Fatal("expected launch to fail")
	}
	turnRec, err := s.GetTurn(ctx, queuedTurn.ID)
	if err != nil {
		t.Fatalf("get queued turn: %v", err)
	}
	if turnRec.Status != "queued" || turnRec.Phase != "queued" || strings.TrimSpace(turnRec.ClaimedBy) != "" || strings.TrimSpace(turnRec.ClaimedAt) != "" || strings.TrimSpace(turnRec.StartedAt) != "" {
		t.Fatalf("expected launch rollback to restore queued turn bookkeeping, got %#v", turnRec)
	}
	if _, _, err := s.GetSessionActiveTurn(ctx, "session_launch_rollback"); err == nil {
		t.Fatal("expected active claim released after launch rollback")
	}
	sessRec, err := s.GetSession(ctx, "session_launch_rollback")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "queued" || sessRec.State["active_turn_id"] != nil {
		t.Fatalf("expected launch rollback to restore queued session state, got %#v", sessRec.State)
	}
	if got := sessRec.State["queue_count"]; got != float64(1) && got != 1 {
		t.Fatalf("expected launch rollback to keep queue_count 1, got %#v", sessRec.State)
	}
}

func TestLaunchTurnRollbackSurfacesCleanupFailure(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_launch_rollback_cleanup", "Test", map[string]any{"model": "bootstrap", "status": "queued", "queue_count": 1}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_launch_rollback_cleanup", "session_launch_rollback_cleanup", "queued", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	engine := New(s)
	engine.beforeLaunchSessionStateErrorHook = func(ctx context.Context, sessionID, turnID string) error {
		if sessionID == "session_launch_rollback_cleanup" && turnID == queuedTurn.ID {
			if err := s.DeleteTurn(ctx, turnID); err != nil {
				t.Fatalf("delete queued turn during rollback hook: %v", err)
			}
			return fmt.Errorf("boom")
		}
		return nil
	}
	runner := engine.runner("session_launch_rollback_cleanup")
	launched, err := engine.launchTurnLocked(ctx, runner, "session_launch_rollback_cleanup", queuedTurn.ID)
	if launched {
		t.Fatal("expected launch to fail")
	}
	if err == nil || !strings.Contains(err.Error(), "boom") || !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected launch rollback error to include setup failure and cleanup sql.ErrNoRows, got %v", err)
	}
	if _, _, err := s.GetSessionActiveTurn(ctx, "session_launch_rollback_cleanup"); err != sql.ErrNoRows {
		t.Fatalf("expected active claim released after rollback cleanup failure, got %v", err)
	}
	sessRec, err := s.GetSession(ctx, "session_launch_rollback_cleanup")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "queued" || sessRec.State["active_turn_id"] != nil {
		t.Fatalf("expected session state restored despite missing-turn rollback failure, got %#v", sessRec.State)
	}
}

func TestCancelQueuedTurnPreservesRunningSessionWithActiveClaim(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	sess, _ := s.CreateSession(ctx, "session_queued_cancel_active", "Test", map[string]any{"model": "bootstrap", "status": "running", "active_turn_id": "turn_active_running"})
	engine := New(s)
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_active_running", sess.ID, "running", "one", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create active turn: %v", err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, sess.ID, activeTurn.ID, "runner", activeTurn.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	} else if !ok {
		t.Fatal("expected active turn claim")
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_queued_cancel_active", sess.ID, "queued", "two", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	if err := engine.CancelTurn(ctx, sess.ID, queuedTurn.ID); err != nil {
		t.Fatalf("cancel queued turn: %v", err)
	}
	sessRec, err := s.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "running" || sessRec.State["active_turn_id"] != activeTurn.ID {
		t.Fatalf("expected queued cancel to preserve running session state, got %#v", sessRec.State)
	}
}

func TestCancelQueuedTurnSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	sess, _ := s.CreateSession(ctx, "session_queued_cancel_ctx", "Test", map[string]any{"model": "bootstrap"})
	engine := New(s)
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_queued_cancel_ctx", sess.ID, "queued", "two", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	if err := engine.CancelTurn(cancelCtx, sess.ID, queuedTurn.ID); err != nil {
		t.Fatalf("cancel queued with canceled caller context: %v", err)
	}
	turnRec, err := s.GetTurn(ctx, queuedTurn.ID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if turnRec.Status != "cancelled" || turnRec.FinishedAt == "" {
		t.Fatalf("expected queued cancel to persist despite canceled caller context, got %#v", turnRec)
	}
	events, err := s.ListTurnEvents(ctx, queuedTurn.ID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundCancelled := false
	for _, event := range events {
		if event.Type == "turn.cancelled" && event.Payload["reason"] == "queued_cancel" {
			foundCancelled = true
		}
	}
	if !foundCancelled {
		t.Fatalf("expected queued cancel audit row despite canceled caller context, got %#v", events)
	}
}

func TestCancelQueuedTurnStartsNextQueuedWorkWhenSessionRemainsQueued(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	started := make(chan struct{}, 1)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		select {
		case started <- struct{}{}:
		default:
		}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "next queued done"}}}}, nil
	})
	sess, _ := s.CreateSession(ctx, "session_cancel_queue_chain", "Chain", map[string]any{"model": "mock-cancel"})
	engine := New(s)
	firstQueued, err := s.CreateTurnWithStatus(ctx, "turn_queued_cancel_chain_1", sess.ID, "queued", "first queued", map[string]any{"intent": "prompt", "model": "mock-cancel"})
	if err != nil {
		t.Fatalf("create first queued turn: %v", err)
	}
	secondQueued, err := s.CreateTurnWithStatus(ctx, "turn_queued_cancel_chain_2", sess.ID, "queued", "second queued", map[string]any{"intent": "prompt", "model": "mock-cancel"})
	if err != nil {
		t.Fatalf("create second queued turn: %v", err)
	}
	if err := engine.CancelTurn(ctx, sess.ID, firstQueued.ID); err != nil {
		t.Fatalf("cancel first queued turn: %v", err)
	}
	firstState, err := s.GetTurn(ctx, firstQueued.ID)
	if err != nil {
		t.Fatalf("get cancelled first turn: %v", err)
	}
	if firstState.Status != "cancelled" {
		t.Fatalf("expected first queued turn cancelled, got %#v", firstState)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("expected next queued turn to start after queued cancel")
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, secondQueued.ID)
		return err == nil && turnRec.StartedAt != ""
	}, "next queued turn start after queued cancel")
}

func TestCancelQueuedTurnRestartPublishesSetupTopics(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	started := make(chan struct{}, 1)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		select {
		case started <- struct{}{}:
		default:
		}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "next queued done"}}}}, nil
	})
	sess, _ := s.CreateSession(ctx, "session_cancel_queue_topics", "Chain", map[string]any{"model": "mock-cancel"})
	engine := New(s)
	firstQueued, err := s.CreateTurnWithStatus(ctx, "turn_queued_cancel_topics_1", sess.ID, "queued", "first queued", map[string]any{"intent": "prompt", "model": "mock-cancel"})
	if err != nil {
		t.Fatalf("create first queued turn: %v", err)
	}
	secondQueued, err := s.CreateTurnWithStatus(ctx, "turn_queued_cancel_topics_2", sess.ID, "queued", "second queued", map[string]any{"intent": "prompt", "model": "mock-cancel"})
	if err != nil {
		t.Fatalf("create second queued turn: %v", err)
	}
	sessionTopicCh, unsubSession := engine.Topics().Subscribe(ctx, "runtime.session", topics.SubscribeOptions{Buffer: 16, SessionID: sess.ID})
	defer unsubSession()
	turnTopicCh, unsubTurn := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 16, SessionID: sess.ID})
	defer unsubTurn()

	if err := engine.CancelTurn(ctx, sess.ID, firstQueued.ID); err != nil {
		t.Fatalf("cancel first queued turn: %v", err)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("expected next queued turn to start after queued cancel")
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, secondQueued.ID)
		return err == nil && turnRec.StartedAt != ""
	}, "next queued turn start after queued cancel")

	foundSessionRunning := false
	foundTurnStarted := false
	deadline := time.After(2 * time.Second)
	for !(foundSessionRunning && foundTurnStarted) {
		select {
		case env := <-sessionTopicCh:
			if env.Payload["type"] == "session_running" && env.Payload["status"] == "running" && env.Payload["active_turn_id"] == secondQueued.ID {
				foundSessionRunning = true
			}
		case env := <-turnTopicCh:
			if env.Payload["type"] == "turn_started" && env.Payload["turn_id"] == secondQueued.ID && env.Payload["status"] == "running" && env.Payload["phase"] == "setup" {
				foundTurnStarted = true
			}
		case <-deadline:
			t.Fatalf("expected queued cancel restart to publish session_running and turn_started, got session_running=%v turn_started=%v", foundSessionRunning, foundTurnStarted)
		}
	}
}

func TestCancelQueuedTurnPublishesTerminalTopicsWhenSessionBecomesIdle(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	sess, _ := s.CreateSession(ctx, "session_cancel_topics_idle", "Test", map[string]any{"model": "bootstrap"})
	engine := New(s)
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_queued_cancel_topics_idle", sess.ID, "queued", "two", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	turnTopicCh, unsubTurn := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 8, SessionID: sess.ID})
	defer unsubTurn()
	sessionTopicCh, unsubSession := engine.Topics().Subscribe(ctx, "runtime.session", topics.SubscribeOptions{Buffer: 8, SessionID: sess.ID})
	defer unsubSession()

	if err := engine.CancelTurn(ctx, sess.ID, queuedTurn.ID); err != nil {
		t.Fatalf("cancel queued turn: %v", err)
	}

	foundTurnTerminal := false
	foundSessionIdle := false
	deadline := time.After(2 * time.Second)
	for !(foundTurnTerminal && foundSessionIdle) {
		select {
		case env := <-turnTopicCh:
			if env.Payload["type"] == "turn_terminal" && env.Payload["turn_id"] == queuedTurn.ID && env.Payload["status"] == "cancelled" && env.Payload["reason"] == "queued_cancel" {
				foundTurnTerminal = true
			}
		case env := <-sessionTopicCh:
			if env.Payload["type"] == "session_idle" && env.Payload["turn_id"] == queuedTurn.ID && env.Payload["turn_status"] == "cancelled" && env.Payload["reason"] == "turn_terminal" && env.Payload["failure_kind"] == "" {
				foundSessionIdle = true
			}
		case <-deadline:
			t.Fatalf("expected queued cancel to publish turn_terminal and session_idle, got turn_terminal=%v session_idle=%v", foundTurnTerminal, foundSessionIdle)
		}
	}
}

func TestSkipHeldTurnSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	if _, err := s.CreateSession(ctx, "session_skip_ctx", "Skip", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_skip_ctx", "session_skip_ctx", "failed", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if err := s.UpsertTurnFailure(ctx, turnRec.ID, turnRec.SessionID, "provider_error", "review", "provider failed"); err != nil {
		t.Fatalf("upsert held failure: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	if err := engine.SkipHeldTurn(cancelCtx, turnRec.ID, "skip requested"); err != nil {
		t.Fatalf("skip held turn with canceled caller context: %v", err)
	}
	failureRec, err := s.GetTurnFailure(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("get turn failure: %v", err)
	}
	if failureRec.HoldState != "none" || failureRec.ResolutionState != "skipped" {
		t.Fatalf("expected skipped resolution despite canceled caller context, got %#v", failureRec)
	}
	events, err := s.ListTurnEvents(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundResolved := false
	for _, event := range events {
		if event.Type == "turn.failure_resolved" && event.Payload["resolution_state"] == "skipped" {
			foundResolved = true
		}
	}
	if !foundResolved {
		t.Fatalf("expected turn.failure_resolved audit row despite canceled caller context, got %#v", events)
	}
}

func TestCancelQueuedTurnRejectsWrongCallerSessionID(t *testing.T) {
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
	if err := engine.CancelTurn(ctx, "session_cancel_b", queuedTurn.ID); err == nil || !strings.Contains(err.Error(), "does not belong to session") {
		t.Fatalf("expected session mismatch error, got %v", err)
	}
	turnRec, err := s.GetTurn(ctx, queuedTurn.ID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if turnRec.Status != "queued" || turnRec.Phase != "queued" || strings.TrimSpace(turnRec.FinishedAt) != "" {
		t.Fatalf("expected queued turn unchanged after wrong-session cancel attempt, got %#v", turnRec)
	}
}

func TestHeartbeatCancelsTurnWhenActiveClaimDisappears(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_heartbeat_claim_lost", "Streaming", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
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
	prevHeartbeatInterval := activeTurnHeartbeatInterval
	activeTurnHeartbeatInterval = 20 * time.Millisecond
	defer func() {
		activeTurnHeartbeatInterval = prevHeartbeatInterval
	}()
	engine := New(s)
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_heartbeat_claim_lost", Prompt: "stream please", Model: "mock-stream"})
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
	if err := s.ReleaseSessionActiveTurn(ctx, "session_heartbeat_claim_lost", result.TurnID); err != nil {
		t.Fatalf("release active claim externally: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && turnRec.Status == "cancelled"
	}, "heartbeat-driven turn cancellation")
	turnRec, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("get cancelled turn: %v", err)
	}
	if turnRec.Phase != "aborted" {
		t.Fatalf("expected lost-claim cancellation to end aborted, got %#v", turnRec)
	}
	sessRec, err := s.GetSession(ctx, "session_heartbeat_claim_lost")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "idle" || sessRec.State["active_turn_id"] != nil {
		t.Fatalf("expected session idle after lost active claim cancellation, got %#v", sessRec.State)
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
	turnTopicCh, unsubTurn := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 16, SessionID: "session_cancel_streaming"})
	defer unsubTurn()
	sessionTopicCh, unsubSession := engine.Topics().Subscribe(ctx, "runtime.session", topics.SubscribeOptions{Buffer: 16, SessionID: "session_cancel_streaming"})
	defer unsubSession()
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
	foundTurnCancelling := false
	foundSessionRunning := false
	deadline := time.After(2 * time.Second)
	for !(foundTurnCancelling && foundSessionRunning) {
		select {
		case env := <-turnTopicCh:
			if env.Payload["type"] == "turn_cancelling" && env.Payload["turn_id"] == result.TurnID && env.Payload["status"] == "cancelling" && env.Payload["phase"] == "cancelling" {
				foundTurnCancelling = true
			}
		case env := <-sessionTopicCh:
			if env.Payload["type"] == "session_state" && env.Payload["status"] == "running" && env.Payload["active_turn_id"] == result.TurnID && env.Payload["turn_id"] == result.TurnID && env.Payload["turn_status"] == "cancelling" && env.Payload["turn_phase"] == "cancelling" && env.Payload["reason"] == "cancel_requested" && env.Payload["failure_kind"] == "" {
				foundSessionRunning = true
			}
		case <-deadline:
			t.Fatalf("expected active cancel to publish turn_cancelling and session_state cancel_requested, got turn_cancelling=%v session_state=%v", foundTurnCancelling, foundSessionRunning)
		}
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
			if event.Payload["reason"] != "cancel_requested" || event.Payload["status"] != "cancelling" || event.Payload["turn_phase"] != "cancelling" || event.Payload["failure_kind"] != "" {
				t.Fatalf("unexpected turn.cancelling payload, got %#v", event.Payload)
			}
		}
	}
	if !foundCancelling {
		t.Fatalf("expected turn.cancelling event, got %#v", events)
	}
}

func TestCancelActiveStreamingTurnSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_cancel_streaming_ctx", "Streaming", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	started := make(chan struct{})
	withStreamWithToolsStub(t, func(ctx context.Context, model string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		select {
		case <-started:
		default:
			close(started)
		}
		<-ctx.Done()
		return nil, ctx.Err()
	})
	engine := New(s)
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_cancel_streaming_ctx", Prompt: "stream please", Model: "mock-stream"})
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
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	if err := engine.CancelTurn(cancelCtx, "session_cancel_streaming_ctx", result.TurnID); err != nil {
		t.Fatalf("cancel active streaming turn with canceled caller context: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && turnRec.Status == "cancelled" && turnRec.FinishedAt != ""
	}, "streaming turn cancellation with canceled caller context")
	events, err := s.ListTurnEvents(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundCancelling := false
	for _, event := range events {
		if event.Type == "turn.cancelling" && event.Payload["reason"] == "cancel_requested" {
			foundCancelling = true
		}
	}
	if !foundCancelling {
		t.Fatalf("expected turn.cancelling audit row despite canceled caller context, got %#v", events)
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

func TestRecoverInterruptedTurnReleasesCancelledClaimWithoutRequeue(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)

	sess, err := s.CreateSession(ctx, "session_recover_cancelled", "Recover", map[string]any{"status": "running"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_recover_cancelled", sess.ID, "cancelled", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create cancelled turn: %v", err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, sess.ID, "turn_recover_cancelled", "worker-test", "claim-recover-cancelled"); err != nil {
		t.Fatalf("claim active turn: %v", err)
	} else if !ok {
		t.Fatal("expected active turn claim to be acquired")
	}
	staleTime := time.Now().Add(-(interruptedTurnStaleAfter + 5*time.Second)).UTC().Format(time.RFC3339Nano)
	if err := s.TouchSessionState(ctx, sess.ID, map[string]any{"active_turn_id": "turn_recover_cancelled", "status": "running"}); err != nil {
		t.Fatalf("touch session state: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update session_active_turns set updated_at = ? where session_id = ? and turn_id = ?`, staleTime, sess.ID, "turn_recover_cancelled"); err != nil {
		t.Fatalf("age active turn claim: %v", err)
	}

	recovered, err := engine.recoverInterruptedTurns(ctx, sess.ID)
	if err != nil {
		t.Fatalf("recover interrupted turns: %v", err)
	}
	if !recovered {
		t.Fatal("expected interrupted cancelled turn to be recovered")
	}
	turnRec, err := s.GetTurn(ctx, "turn_recover_cancelled")
	if err != nil {
		t.Fatalf("get recovered turn: %v", err)
	}
	if turnRec.Status != "cancelled" || turnRec.Phase != "aborted" {
		t.Fatalf("expected cancelled turn to remain terminal, got %#v", turnRec)
	}
	if _, _, err := s.GetSessionActiveTurn(ctx, sess.ID); err != sql.ErrNoRows {
		t.Fatalf("expected active claim released, got err=%v", err)
	}
	queueCount, err := s.CountQueuedTurns(ctx, sess.ID)
	if err != nil {
		t.Fatalf("count queued turns: %v", err)
	}
	if queueCount != 0 {
		t.Fatalf("expected no queued turns after cancelled recovery, got %d", queueCount)
	}
	sessRec, err := s.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sessRec.State["status"] != "idle" {
		t.Fatalf("expected session idle after cancelled recovery, got %#v", sessRec)
	}
}

func TestRecoverInterruptedTurnsSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)

	sess, err := s.CreateSession(ctx, "session_recover_cancel_ctx", "Recover", map[string]any{"status": "running"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_recover_cancel_ctx", sess.ID, "cancelled", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create cancelled turn: %v", err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, sess.ID, "turn_recover_cancel_ctx", "worker-test", "claim-recover-cancel-ctx"); err != nil {
		t.Fatalf("claim active turn: %v", err)
	} else if !ok {
		t.Fatal("expected active turn claim to be acquired")
	}
	staleTime := time.Now().Add(-(interruptedTurnStaleAfter + 5*time.Second)).UTC().Format(time.RFC3339Nano)
	if err := s.TouchSessionState(ctx, sess.ID, map[string]any{"active_turn_id": "turn_recover_cancel_ctx", "status": "running"}); err != nil {
		t.Fatalf("touch session state: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update session_active_turns set updated_at = ? where session_id = ? and turn_id = ?`, staleTime, sess.ID, "turn_recover_cancel_ctx"); err != nil {
		t.Fatalf("age active turn claim: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	recovered, err := engine.recoverInterruptedTurns(cancelCtx, sess.ID)
	if err != nil {
		t.Fatalf("recover interrupted turns with canceled caller context: %v", err)
	}
	if !recovered {
		t.Fatal("expected interrupted cancelled turn to be recovered")
	}
	if _, _, err := s.GetSessionActiveTurn(ctx, sess.ID); err != sql.ErrNoRows {
		t.Fatalf("expected active claim released, got err=%v", err)
	}
	events, err := s.ListTurnEvents(ctx, "turn_recover_cancel_ctx")
	if err != nil {
		t.Fatalf("list recovery events: %v", err)
	}
	foundRecovered := false
	for _, event := range events {
		if event.Type == "turn.recovered" && event.Payload["reason"] == "recovery" {
			foundRecovered = true
		}
	}
	if !foundRecovered {
		t.Fatalf("expected recovery audit row despite canceled caller context, got %#v", events)
	}
}

func TestMarkTurnFailureSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_failure_ctx", "FailureCtx", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_failure_ctx", "session_failure_ctx", "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	markTurnFailureWithHoldAndFallback(cancelCtx, ctx, s, "turn_failure_ctx", "session_failure_ctx", "provider_error", "review", "failed under canceled ctx")
	failureRec, err := s.GetTurnFailure(ctx, "turn_failure_ctx")
	if err != nil {
		t.Fatalf("get turn failure: %v", err)
	}
	if failureRec.FailureKind != "provider_error" || failureRec.HoldState != "review" || failureRec.Summary != "failed under canceled ctx" {
		t.Fatalf("unexpected failure marker after canceled caller context: %#v", failureRec)
	}
	events, err := s.ListTurnEvents(ctx, "turn_failure_ctx")
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundMarked := false
	for _, event := range events {
		if event.Type == "turn.failure_marked" && event.Payload["failure_kind"] == "provider_error" && event.Payload["hold_state"] == "review" {
			foundMarked = true
		}
	}
	if !foundMarked {
		t.Fatalf("expected turn.failure_marked audit row despite canceled caller context, got %#v", events)
	}
}

func TestRecoverInterruptedTurnsStartsQueuedWorkAfterReleasingTerminalClaim(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	started := make(chan struct{}, 1)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		select {
		case started <- struct{}{}:
		default:
		}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "queued done"}}}}, nil
	})
	engine := New(s)

	sess, err := s.CreateSession(ctx, "session_recover_restart", "Recover", map[string]any{"status": "running"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_recover_terminal", sess.ID, "cancelled", "terminal", map[string]any{"intent": "prompt", "model": "mock-recover"}); err != nil {
		t.Fatalf("create cancelled turn: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_recover_queued", sess.ID, "queued", "queued after recovery", map[string]any{"intent": "prompt", "model": "mock-recover"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, sess.ID, "turn_recover_terminal", "worker-test", "claim-recover-terminal"); err != nil {
		t.Fatalf("claim active turn: %v", err)
	} else if !ok {
		t.Fatal("expected active turn claim to be acquired")
	}
	staleTime := time.Now().Add(-(interruptedTurnStaleAfter + 5*time.Second)).UTC().Format(time.RFC3339Nano)
	if err := s.TouchSessionState(ctx, sess.ID, map[string]any{"active_turn_id": "turn_recover_terminal", "status": "running"}); err != nil {
		t.Fatalf("touch session state: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update session_active_turns set updated_at = ? where session_id = ? and turn_id = ?`, staleTime, sess.ID, "turn_recover_terminal"); err != nil {
		t.Fatalf("age active turn claim: %v", err)
	}

	recovered, err := engine.recoverInterruptedTurns(ctx, sess.ID)
	if err != nil {
		t.Fatalf("recover interrupted turns: %v", err)
	}
	if !recovered {
		t.Fatal("expected interrupted terminal turn to be recovered")
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, queuedTurn.ID)
		return err == nil && turnRec.StartedAt != ""
	}, "queued turn launch after recovery")
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("expected queued turn to start after recovery released terminal claim")
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, queuedTurn.ID)
		return err == nil && turnRec.Status == "completed"
	}, "queued turn completion after recovery")
}

func TestRecoverInterruptedTurnsEmitsSessionRestartFailureSummary(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)

	sess, err := s.CreateSession(ctx, "session_recover_restart_fail", "Recover", map[string]any{"status": "running", "model": "mock-recover"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_recover_restart_fail_terminal", sess.ID, "cancelled", "terminal", map[string]any{"intent": "prompt", "model": "mock-recover"}); err != nil {
		t.Fatalf("create terminal turn: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_recover_restart_fail_queued", sess.ID, "queued", "queued after recovery", map[string]any{"intent": "prompt", "model": "mock-recover"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, sess.ID, "turn_recover_restart_fail_terminal", "worker-test", "claim-recover-restart-fail"); err != nil {
		t.Fatalf("claim active turn: %v", err)
	} else if !ok {
		t.Fatal("expected active turn claim")
	}
	staleTime := time.Now().Add(-(interruptedTurnStaleAfter + 5*time.Second)).UTC().Format(time.RFC3339Nano)
	if err := s.TouchSessionState(ctx, sess.ID, map[string]any{"active_turn_id": "turn_recover_restart_fail_terminal", "status": "running"}); err != nil {
		t.Fatalf("touch session state: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update session_active_turns set updated_at = ? where session_id = ?`, staleTime, sess.ID); err != nil {
		t.Fatalf("age active turn claim: %v", err)
	}
	engine.beforeLaunchSessionStateErrorHook = func(ctx context.Context, sessionID, turnID string) error {
		if sessionID == sess.ID && turnID == queuedTurn.ID {
			return fmt.Errorf("restart boom")
		}
		return nil
	}
	sessionTopicCh, unsubSession := engine.Topics().Subscribe(ctx, "runtime.session", topics.SubscribeOptions{Buffer: 16, SessionID: sess.ID})
	defer unsubSession()
	recovered, err := engine.recoverInterruptedTurns(ctx, sess.ID)
	if !recovered {
		t.Fatal("expected recovery pass to report released stale claim before restart failure")
	}
	if err == nil || !strings.Contains(err.Error(), "restart boom") {
		t.Fatalf("expected restart failure to surface, got %v", err)
	}
	found := false
	deadline := time.After(time.Second)
	for !found {
		select {
		case env := <-sessionTopicCh:
			if env.Payload["type"] == "session_state" && env.Payload["reason"] == "recovery_restart_failed" {
				if env.Payload["queue_count"] != 1 && env.Payload["queue_count"] != float64(1) {
					t.Fatalf("expected queue_count 1 in restart failure summary, got %#v", env)
				}
				found = true
			}
		case <-deadline:
			t.Fatal("expected session restart failure summary")
		}
	}
	events, err := s.ListTurnEvents(ctx, queuedTurn.ID)
	if err != nil {
		t.Fatalf("list queued turn recovery restart failure events: %v", err)
	}
	foundEvent := false
	for _, event := range events {
		if event.Type == "turn.recovery_restart_failed" && event.Payload["reason"] == "recovery_restart_failed" {
			foundEvent = true
		}
	}
	if !foundEvent {
		t.Fatalf("expected queued turn recovery restart failure event, got %#v", events)
	}
}

func TestRecoverInterruptedTurnPublishesRuntimeStateTopics(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)

	sess, err := s.CreateSession(ctx, "session_recover_topics", "Recover", map[string]any{"status": "running", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_recover_topics", sess.ID, "running", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, turnRec.ID, "running", "waiting_on_tools"); err != nil {
		t.Fatalf("set waiting_on_tools phase: %v", err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, sess.ID, turnRec.ID, "runner", "claim-recover-topics"); err != nil {
		t.Fatalf("claim active turn: %v", err)
	} else if !ok {
		t.Fatal("expected active turn claim to be acquired")
	}
	if _, err := s.DB().ExecContext(ctx, `update session_active_turns set updated_at = '2000-01-01T00:00:00Z' where session_id = ?`, sess.ID); err != nil {
		t.Fatalf("age active turn claim: %v", err)
	}
	turnTopicCh, unsubTurn := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 8, SessionID: sess.ID})
	defer unsubTurn()
	sessionTopicCh, unsubSession := engine.Topics().Subscribe(ctx, "runtime.session", topics.SubscribeOptions{Buffer: 8, SessionID: sess.ID})
	defer unsubSession()

	recovered, err := engine.recoverInterruptedTurns(ctx, sess.ID)
	if err != nil {
		t.Fatalf("recover interrupted turns: %v", err)
	}
	if !recovered {
		t.Fatal("expected stale turn to be recovered")
	}

	foundTurnRecovered := false
	foundTurnState := false
	turnDeadline := time.After(time.Second)
	for !(foundTurnRecovered && foundTurnState) {
		select {
		case env := <-turnTopicCh:
			switch env.Payload["type"] {
			case "turn_recovered":
				if env.Payload["status"] != "failed" || env.Payload["phase"] != "held_for_retry_or_skip" || env.Payload["recovery_disposition"] != "hold_for_retry_or_skip_after_tool_checkpoint" {
					t.Fatalf("unexpected runtime.turn recovery checkpoint payload: %#v", env)
				}
				foundTurnRecovered = true
			case "turn_state":
				if env.Payload["status"] != "failed" || env.Payload["phase"] != "held_for_retry_or_skip" || env.Payload["recovery_disposition"] != "hold_for_retry_or_skip_after_tool_checkpoint" {
					t.Fatalf("unexpected runtime.turn recovery state payload: %#v", env)
				}
				foundTurnState = true
			}
		case <-turnDeadline:
			t.Fatalf("expected runtime.turn recovery checkpoint and state topics, got turn_recovered=%v turn_state=%v", foundTurnRecovered, foundTurnState)
		}
	}
	select {
	case env := <-sessionTopicCh:
		if env.Payload["type"] != "session_state" || env.Payload["status"] != "idle" || env.Payload["reason"] != "recovery" || env.Payload["recovery_disposition"] != "hold_for_retry_or_skip_after_tool_checkpoint" || env.Payload["turn_status"] != "failed" {
			t.Fatalf("unexpected runtime.session recovery payload: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected runtime.session recovery state topic")
	}
}

func TestRecoverInterruptedTurnsRestartPublishesSetupTopics(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	started := make(chan struct{}, 1)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		select {
		case started <- struct{}{}:
		default:
		}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "queued done"}}}}, nil
	})
	engine := New(s)

	sess, err := s.CreateSession(ctx, "session_recover_restart_topics", "Recover", map[string]any{"status": "running"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_recover_restart_terminal", sess.ID, "cancelled", "terminal", map[string]any{"intent": "prompt", "model": "mock-recover"}); err != nil {
		t.Fatalf("create cancelled turn: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_recover_restart_queued", sess.ID, "queued", "queued after recovery", map[string]any{"intent": "prompt", "model": "mock-recover"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, sess.ID, "turn_recover_restart_terminal", "worker-test", "claim-recover-terminal-topics"); err != nil {
		t.Fatalf("claim active turn: %v", err)
	} else if !ok {
		t.Fatal("expected active turn claim to be acquired")
	}
	staleTime := time.Now().Add(-(interruptedTurnStaleAfter + 5*time.Second)).UTC().Format(time.RFC3339Nano)
	if err := s.TouchSessionState(ctx, sess.ID, map[string]any{"active_turn_id": "turn_recover_restart_terminal", "status": "running"}); err != nil {
		t.Fatalf("touch session state: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update session_active_turns set updated_at = ? where session_id = ? and turn_id = ?`, staleTime, sess.ID, "turn_recover_restart_terminal"); err != nil {
		t.Fatalf("age active turn claim: %v", err)
	}
	sessionTopicCh, unsubSession := engine.Topics().Subscribe(ctx, "runtime.session", topics.SubscribeOptions{Buffer: 16, SessionID: sess.ID})
	defer unsubSession()
	turnTopicCh, unsubTurn := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 16, SessionID: sess.ID})
	defer unsubTurn()

	recovered, err := engine.recoverInterruptedTurns(ctx, sess.ID)
	if err != nil {
		t.Fatalf("recover interrupted turns: %v", err)
	}
	if !recovered {
		t.Fatal("expected interrupted terminal turn to be recovered")
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("expected queued turn to start after recovery released terminal claim")
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, queuedTurn.ID)
		return err == nil && turnRec.StartedAt != ""
	}, "queued turn launch after recovery")

	foundSessionRunning := false
	foundTurnStarted := false
	deadline := time.After(2 * time.Second)
	for !(foundSessionRunning && foundTurnStarted) {
		select {
		case env := <-sessionTopicCh:
			if env.Payload["type"] == "session_running" && env.Payload["status"] == "running" && env.Payload["active_turn_id"] == queuedTurn.ID {
				foundSessionRunning = true
			}
		case env := <-turnTopicCh:
			if env.Payload["type"] == "turn_started" && env.Payload["turn_id"] == queuedTurn.ID && env.Payload["status"] == "running" && env.Payload["phase"] == "setup" {
				foundTurnStarted = true
			}
		case <-deadline:
			t.Fatalf("expected recovery restart to publish session_running and turn_started, got session_running=%v turn_started=%v", foundSessionRunning, foundTurnStarted)
		}
	}
}

func TestResolveTurnIdentityForFinalizeSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateDefaultSession("agentfinalize", "web", "acctfinalize", "session_finalize_ctx")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_finalize_ctx", Title: "@agentfinalize", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session with canonical identity: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_finalize_ctx", sess.ID, "queued", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create finalize turn: %v", err)
	}
	engine := New(s)
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	agentID, model := engine.runner(sess.ID).resolveTurnIdentityForFinalize(cancelCtx, s, sess.ID, "turn_finalize_ctx")
	if agentID != "agentfinalize" {
		t.Fatalf("expected canonical agent id under canceled caller context, got %q", agentID)
	}
	if model == "" {
		t.Fatal("expected non-empty model resolution under canceled caller context")
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
	turnTopicCh, unsubTurn := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 8, SessionID: "session_setup_error"})
	defer unsubTurn()
	sessionTopicCh, unsubSession := engine.Topics().Subscribe(ctx, "runtime.session", topics.SubscribeOptions{Buffer: 8, SessionID: "session_setup_error"})
	defer unsubSession()
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
	foundTurnTerminal := false
	foundSessionIdle := false
	deadline := time.After(2 * time.Second)
	for !(foundTurnTerminal && foundSessionIdle) {
		select {
		case env := <-turnTopicCh:
			if env.Payload["type"] == "turn_terminal" && env.Payload["status"] == "failed" && env.Payload["failure_kind"] == "setup_error" && env.Payload["reason"] == "setup_error" {
				foundTurnTerminal = true
			}
		case env := <-sessionTopicCh:
			if env.Payload["type"] == "session_idle" && env.Payload["turn_status"] == "failed" && env.Payload["failure_kind"] == "setup_error" && env.Payload["model"] == "bootstrap" {
				foundSessionIdle = true
			}
		case <-deadline:
			t.Fatalf("expected setup failure to publish enriched turn/session terminal topics, got turn_terminal=%v session_idle=%v", foundTurnTerminal, foundSessionIdle)
		}
	}
	msgs, err := s.ListMessages(ctx, "session_setup_error")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	foundFailure := false
	for _, msg := range msgs {
		if msg.Role == "system" && strings.Contains(msg.Content, "Turn setup error: boom during setup") {
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
	turnTopicCh, unsubTurn := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 16, SessionID: "session_cancel_setup"})
	defer unsubTurn()
	sessionTopicCh, unsubSession := engine.Topics().Subscribe(ctx, "runtime.session", topics.SubscribeOptions{Buffer: 16, SessionID: "session_cancel_setup"})
	defer unsubSession()
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
	foundTurnCancelling := false
	foundTurnTerminal := false
	foundSessionCancelRequested := false
	foundSessionIdle := false
	deadline := time.After(2 * time.Second)
	for !(foundTurnCancelling && foundTurnTerminal && foundSessionCancelRequested && foundSessionIdle) {
		select {
		case env := <-turnTopicCh:
			if env.Payload["type"] == "turn_cancelling" && env.Payload["turn_id"] == result.TurnID && env.Payload["status"] == "cancelling" && env.Payload["phase"] == "cancelling" && env.Payload["reason"] == "cancel_requested" {
				foundTurnCancelling = true
			}
			if env.Payload["type"] == "turn_terminal" && env.Payload["turn_id"] == result.TurnID && env.Payload["status"] == "cancelled" && env.Payload["reason"] == "cancelled" && env.Payload["failure_kind"] == "" {
				foundTurnTerminal = true
			}
		case env := <-sessionTopicCh:
			if env.Payload["type"] == "session_state" && env.Payload["status"] == "running" && env.Payload["active_turn_id"] == result.TurnID && env.Payload["turn_id"] == result.TurnID && env.Payload["turn_status"] == "cancelling" && env.Payload["turn_phase"] == "cancelling" && env.Payload["reason"] == "cancel_requested" && env.Payload["failure_kind"] == "" {
				foundSessionCancelRequested = true
			}
			if env.Payload["type"] == "session_idle" && env.Payload["turn_id"] == result.TurnID && env.Payload["turn_status"] == "cancelled" && env.Payload["reason"] == "turn_terminal" && env.Payload["failure_kind"] == "" && env.Payload["model"] == "bootstrap" {
				foundSessionIdle = true
			}
		case <-deadline:
			t.Fatalf("expected setup cancel to publish turn/session cancel + terminal topics, got turn_cancelling=%v turn_terminal=%v session_state=%v session_idle=%v", foundTurnCancelling, foundTurnTerminal, foundSessionCancelRequested, foundSessionIdle)
		}
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

func TestShellTurnPublishesTerminalRuntimeTopics(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_shell_topics", "Shell", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	turnTopicCh, unsubTurn := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 16, SessionID: "session_shell_topics"})
	defer unsubTurn()
	sessionTopicCh, unsubSession := engine.Topics().Subscribe(ctx, "runtime.session", topics.SubscribeOptions{Buffer: 16, SessionID: "session_shell_topics"})
	defer unsubSession()
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_shell_topics", Prompt: "hello shell", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit shell prompt: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && turnRec.Status == "completed" && turnRec.FinishedAt != ""
	}, "shell turn completion")
	foundTurnCompleted := false
	foundSessionIdle := false
	deadline := time.After(2 * time.Second)
	for !(foundTurnCompleted && foundSessionIdle) {
		select {
		case env := <-turnTopicCh:
			if env.Payload["type"] == "turn_completed" && env.Payload["turn_id"] == result.TurnID && env.Payload["status"] == "completed" && env.Payload["reason"] == "completed" && env.Payload["completion_kind"] == "response" {
				foundTurnCompleted = true
			}
		case env := <-sessionTopicCh:
			if env.Payload["type"] == "session_idle" && env.Payload["turn_id"] == result.TurnID && env.Payload["turn_status"] == "completed" && env.Payload["failure_kind"] == "" && env.Payload["model"] == "bootstrap" && env.Payload["completion_kind"] == "response" {
				foundSessionIdle = true
			}
		case <-deadline:
			t.Fatalf("expected shell turn to publish runtime turn/session terminal topics, got turn_completed=%v session_idle=%v", foundTurnCompleted, foundSessionIdle)
		}
	}
	events, err := s.ListTurnEvents(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundFinishedAudit := false
	foundToolFinishedAudit := false
	for _, event := range events {
		if event.Type == "turn.finished" && event.Payload["status"] == "completed" && event.Payload["completion_kind"] == "response" {
			foundFinishedAudit = true
		}
		if event.Type == "tool.finished" && event.Payload["tool"] == "shell" && event.Payload["output_length"] != nil {
			foundToolFinishedAudit = true
		}
	}
	if !foundFinishedAudit {
		t.Fatalf("expected shell completion to persist completion_kind on turn.finished, got %#v", events)
	}
	if !foundToolFinishedAudit {
		t.Fatalf("expected shell completion to persist normalized tool.finished audit row, got %#v", events)
	}
}

func TestShellCancelBroadcastsSystemMessageToTurnResponseTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_shell_cancel_topics", "Shell", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	turnResponseCh, unsub := engine.Topics().Subscribe(ctx, "turn.response", topics.SubscribeOptions{Buffer: 16, SessionID: "session_shell_cancel_topics"})
	defer unsub()
	setupEntered := make(chan struct{})
	engine.beforeSetupHook = func(ctx context.Context, sessionID, turnID string) {
		select {
		case <-setupEntered:
		default:
			close(setupEntered)
		}
		<-ctx.Done()
	}
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_shell_cancel_topics", Prompt: "cancel shell", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit shell prompt: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		select {
		case <-setupEntered:
			return true
		default:
			return false
		}
	}, "shell setup phase entry")
	if err := engine.CancelTurn(ctx, "session_shell_cancel_topics", result.TurnID); err != nil {
		t.Fatalf("cancel shell turn during setup: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && turnRec.Status == "cancelled" && turnRec.FinishedAt != ""
	}, "shell turn cancellation")
	deadline := time.After(2 * time.Second)
	foundSystemMessage := false
	for !foundSystemMessage {
		select {
		case env := <-turnResponseCh:
			if env.Payload["sender"] == "system" {
				data, _ := env.Payload["data"].(map[string]any)
				if data["type"] == "system_message" && data["content"] == "Turn cancelled" {
					foundSystemMessage = true
				}
			}
		case <-deadline:
			t.Fatal("expected shell cancel to publish turn.response system message")
		}
	}
	events, err := s.ListTurnEvents(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundFinishedAudit := false
	for _, event := range events {
		if event.Type == "turn.finished" && event.Payload["status"] == "cancelled" && event.Payload["reason"] == "cancelled" && event.Payload["failure_kind"] == "" {
			foundFinishedAudit = true
		}
	}
	if !foundFinishedAudit {
		t.Fatalf("expected shell cancel to persist turn.finished audit row, got %#v", events)
	}
}

func TestShellFailureBroadcastsSystemMessageToTurnResponseTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_shell_failure_topics", "Shell", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	turnResponseCh, unsub := engine.Topics().Subscribe(ctx, "turn.response", topics.SubscribeOptions{Buffer: 16, SessionID: "session_shell_failure_topics"})
	defer unsub()
	t.Setenv("PATH", "")
	result, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_shell_failure_topics", Prompt: "fail shell", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit shell prompt: %v", err)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && turnRec.Status == "failed" && turnRec.FinishedAt != ""
	}, "shell turn failure")
	deadline := time.After(2 * time.Second)
	foundSystemMessage := false
	for !foundSystemMessage {
		select {
		case env := <-turnResponseCh:
			if env.Payload["sender"] == "system" {
				data, _ := env.Payload["data"].(map[string]any)
				if data["type"] == "system_message" {
					content, _ := data["content"].(string)
					if strings.Contains(content, "Shell tool failed:") {
						foundSystemMessage = true
					}
				}
			}
		case <-deadline:
			t.Fatal("expected shell failure to publish turn.response system message")
		}
	}
	events, err := s.ListTurnEvents(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundFinishedAudit := false
	foundToolFailedAudit := false
	for _, event := range events {
		if event.Type == "turn.finished" && event.Payload["status"] == "failed" && event.Payload["reason"] == "shell_error" && event.Payload["failure_kind"] == "shell_error" {
			foundFinishedAudit = true
		}
		if event.Type == "tool.failed" && event.Payload["tool"] == "shell" && event.Payload["failure_kind"] == "shell_error" {
			foundToolFailedAudit = true
		}
	}
	if !foundFinishedAudit {
		t.Fatalf("expected shell failure to persist turn.finished audit row, got %#v", events)
	}
	if !foundToolFailedAudit {
		t.Fatalf("expected shell failure to persist normalized tool.failed audit row, got %#v", events)
	}
}

func TestHoldTurnFailureSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	if _, err := s.CreateSession(ctx, "session_hold_ctx", "Hold", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_hold_ctx", "session_hold_ctx", "failed", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if err := s.UpsertTurnFailure(ctx, turnRec.ID, turnRec.SessionID, "provider_error", "none", "provider failed"); err != nil {
		t.Fatalf("upsert turn failure: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	if err := engine.HoldTurnFailure(cancelCtx, turnRec.ID, "review", "needs review"); err != nil {
		t.Fatalf("hold turn failure with canceled caller context: %v", err)
	}
	failureRec, err := s.GetTurnFailure(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("get turn failure: %v", err)
	}
	if failureRec.HoldState != "review" {
		t.Fatalf("expected held failure despite canceled caller context, got %#v", failureRec)
	}
	events, err := s.ListTurnEvents(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundHeld := false
	for _, event := range events {
		if event.Type == "turn.failure_held" && event.Payload["hold_state"] == "review" {
			foundHeld = true
		}
	}
	if !foundHeld {
		t.Fatalf("expected turn.failure_held audit row despite canceled caller context, got %#v", events)
	}
}

func TestHoldResolutionPublishesNormalizedHeldPayload(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_hold_payload_topics", "Hold Payload", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_hold_payload_topics", "session_hold_payload_topics", "failed", "redo this", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create failed turn: %v", err)
	}
	engine := New(s)
	turnTopicCh, unsub := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 16, SessionID: turnRec.SessionID})
	defer unsub()
	if err := engine.HoldTurnFailure(ctx, turnRec.ID, "  REVIEW  ", "   "); err != nil {
		t.Fatalf("hold turn failure: %v", err)
	}
	failureRec, err := s.GetTurnFailure(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("get turn failure: %v", err)
	}
	deadline := time.After(2 * time.Second)
	for {
		select {
		case env := <-turnTopicCh:
			if env.Payload["type"] == "turn_failure_held" {
				if env.Payload["hold_state"] != failureRec.HoldState || env.Payload["summary"] != failureRec.Summary {
					t.Fatalf("expected normalized held payload to match store row, got payload=%#v store=%#v", env.Payload, failureRec)
				}
				return
			}
		case <-deadline:
			t.Fatal("expected turn_failure_held runtime topic")
		}
	}
}

func TestHoldResolutionPublishesUpdatedRuntimeTurnPhases(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_hold_phase_topics", "Hold Phase", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_hold_phase_topics", "session_hold_phase_topics", "failed", "redo this", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create failed turn: %v", err)
	}
	engine := New(s)
	turnTopicCh, unsub := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 16, SessionID: turnRec.SessionID})
	defer unsub()
	if err := engine.HoldTurnFailure(ctx, turnRec.ID, "review", "needs operator choice"); err != nil {
		t.Fatalf("hold turn failure: %v", err)
	}
	if err := engine.SkipHeldTurn(ctx, turnRec.ID, "skip requested"); err != nil {
		t.Fatalf("skip held turn: %v", err)
	}
	foundHeld := false
	foundResolved := false
	deadline := time.After(2 * time.Second)
	for !(foundHeld && foundResolved) {
		select {
		case env := <-turnTopicCh:
			switch env.Payload["type"] {
			case "turn_failure_held":
				if env.Payload["phase"] == "held_for_retry_or_skip" {
					foundHeld = true
				}
			case "turn_failure_resolved":
				if env.Payload["phase"] == "failed" && env.Payload["resolution_state"] == "skipped" {
					foundResolved = true
				}
			}
		case <-deadline:
			t.Fatalf("expected updated hold-resolution runtime phases, got held=%v resolved=%v", foundHeld, foundResolved)
		}
	}
}

func TestHoldResolutionPublishesRuntimeTurnCheckpoints(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	if _, err := s.CreateSession(ctx, "session_hold_topics", "Hold", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_hold_topics", "session_hold_topics", "failed", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if err := s.UpsertTurnFailure(ctx, turnRec.ID, turnRec.SessionID, "provider_error", "none", "provider failed"); err != nil {
		t.Fatalf("upsert turn failure: %v", err)
	}
	turnTopicCh, unsub := engine.Topics().Subscribe(ctx, "runtime.turn", topics.SubscribeOptions{Buffer: 16, SessionID: turnRec.SessionID})
	defer unsub()
	if err := engine.HoldTurnFailure(ctx, turnRec.ID, "review", "needs review"); err != nil {
		t.Fatalf("hold turn failure: %v", err)
	}
	if err := engine.SkipHeldTurn(ctx, turnRec.ID, "skipped on purpose"); err != nil {
		t.Fatalf("skip held turn: %v", err)
	}
	foundHeld := false
	foundResolved := false
	deadline := time.After(2 * time.Second)
	for !(foundHeld && foundResolved) {
		select {
		case env := <-turnTopicCh:
			if env.Payload["type"] == "turn_failure_held" && env.Payload["turn_id"] == turnRec.ID && env.Payload["hold_state"] == "review" && env.Payload["summary"] == "needs review" {
				foundHeld = true
			}
			if env.Payload["type"] == "turn_failure_resolved" && env.Payload["turn_id"] == turnRec.ID && env.Payload["resolution_state"] == "skipped" && env.Payload["resolution_summary"] == "skipped on purpose" {
				foundResolved = true
			}
		case <-deadline:
			t.Fatalf("expected hold resolution runtime turn checkpoints, got held=%v resolved=%v", foundHeld, foundResolved)
		}
	}
}

func TestRecordRouteDecisionSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateDefaultSession("routeagent", "web", "routeacct", "session_route_ctx")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_route_ctx", Title: "@routeagent", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session with canonical identity: %v", err)
	}
	target, err := s.CloneSession(ctx, sess.ID, "session_target_ctx", "@agent1", "agent1")
	if err != nil {
		t.Fatalf("create target session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_route_ctx", sess.ID, "queued", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create route turn: %v", err)
	}
	engine := New(s)
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	metadata := map[string]any{
		"source_session_id":  sess.ID,
		"target_session_id":  target.ID,
		"target_agent_id":    "agent1",
		"route_mode":         "prompt",
		"route_matched_by":   "mention",
		"routing_policy":     "mention",
		"requested_agent_id": "agent1",
		"routing_enabled":    true,
	}
	if err := engine.recordRouteDecision(cancelCtx, sess.ID, "turn_route_ctx", metadata); err != nil {
		t.Fatalf("record route decision with canceled caller context: %v", err)
	}
	events, err := s.ListRouteEvents(ctx, sess.ID)
	if err != nil {
		t.Fatalf("list route events: %v", err)
	}
	found := false
	for _, event := range events {
		if event.TurnID == "turn_route_ctx" && event.SourceAgentID == "routeagent" && event.TargetAgentID == "agent1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected route event with canonical source agent despite canceled caller context, got %#v", events)
	}
}

func TestSubmitPromptRoutedSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	root, err := s.CreateSession(ctx, "session_root_routed_cancel_ctx", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	result, err := engine.SubmitPromptRouted(cancelCtx, RunInput{SessionID: root.ID, Prompt: "@agent1 hello there", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit routed prompt with canceled caller context: %v", err)
	}
	if !result.Routed || result.TurnID == "" || result.SessionID == root.ID {
		t.Fatalf("expected routed prompt result despite canceled caller context, got %#v", result)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && strings.TrimSpace(turnRec.FinishedAt) != ""
	}, "routed prompt completion with canceled caller context")
	msgs, err := s.ListMessages(ctx, root.ID)
	if err != nil {
		t.Fatalf("list source messages: %v", err)
	}
	foundRoutingNotice := false
	for _, msg := range msgs {
		if msg.Role == "system" && msg.Payload["kind"] == "routing" {
			foundRoutingNotice = true
		}
	}
	if !foundRoutingNotice {
		t.Fatalf("expected source routing notice despite canceled caller context, got %#v", msgs)
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

func TestPreparePromptRouteResolutionPrefersSessionIdentityUnderCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	scope := gisession.SessionScope{Version: gisession.ScopeVersionV1, AgentID: "agent", Channel: "slack", Account: "workspace", Dimensions: []string{"space", "chat", "topic"}, Values: map[string]string{"space": "room:eng", "chat": "group:thread-7", "topic": "topic:builds"}}
	source, err := s.CreateSessionWithMetadata(ctx, "session_route_scope_identity_cancel", "", "@agent", map[string]any{"model": "bootstrap", "status": "idle"}, &scope, []string{"agent:agent:slack:chat:group:thread-7"})
	if err != nil {
		t.Fatalf("create source session: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update sessions set scope_json = ? where id = ?`, `{"version":1,"agent_id":"wrong","channel":"gi","account":"default","dimensions":["chat"],"values":{"chat":"direct:wrong"}}`, source.ID); err != nil {
		t.Fatalf("mutate scope snapshot: %v", err)
	}
	engine := New(s)
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	resolution, err := engine.preparePromptRouteResolution(cancelCtx, RunInput{SessionID: source.ID, Prompt: "@agent1 hello there", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("prepare prompt route resolution under canceled caller context: %v", err)
	}
	if resolution.inbound.Channel != "slack" || resolution.inbound.Account != "workspace" {
		t.Fatalf("expected canonical inbound channel/account under canceled caller context, got %#v", resolution.inbound)
	}
	if resolution.inbound.ChatType != "group" || resolution.inbound.ChatID != "thread-7" || resolution.inbound.SpaceType != "room" || resolution.inbound.SpaceID != "eng" || resolution.inbound.TopicID != "builds" {
		t.Fatalf("expected canonical scoped inbound context under canceled caller context, got %#v", resolution.inbound)
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

func TestEnqueueDirectInboundSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_inbound_cancel")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_inbound_cancel", Title: "@agent", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	queued, err := engine.EnqueueDirectInbound(cancelCtx, DirectInput{Kind: DirectKindPrompt, SessionID: sess.ID, Prompt: "hello from canceled enqueue", Model: "bootstrap", Origin: DirectOrigin{SourceKind: DirectSourceKindIPC, SourceID: "ipc:canceled"}})
	if err != nil {
		t.Fatalf("enqueue direct inbound with canceled context: %v", err)
	}
	if queued.Status != "queued" || queued.ID <= 0 {
		t.Fatalf("expected queued inbound item, got %#v", queued)
	}
	item, err := s.GetInboundWork(ctx, queued.ID)
	if err != nil {
		t.Fatalf("get inbound work: %v", err)
	}
	if item.Status != "queued" || item.SourceKind != DirectSourceKindIPC || item.SessionID != sess.ID {
		t.Fatalf("expected persisted queued inbound item, got %#v", item)
	}
}

func TestProcessQueuedInboundWorkDrainsMultipleItemsInOrder(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}}, nil
	})
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_inbound_drain")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_inbound_drain", Title: "@agent", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	first, err := engine.EnqueueDirectInbound(ctx, DirectInput{Kind: DirectKindPrompt, SessionID: sess.ID, Prompt: "first inbound", Model: "bootstrap", Origin: DirectOrigin{SourceKind: DirectSourceKindIPC, SourceID: "ipc:first"}})
	if err != nil {
		t.Fatalf("enqueue first inbound: %v", err)
	}
	second, err := engine.EnqueueDirectInbound(ctx, DirectInput{Kind: DirectKindPrompt, SessionID: sess.ID, Prompt: "second inbound", Model: "bootstrap", Origin: DirectOrigin{SourceKind: DirectSourceKindIPC, SourceID: "ipc:second"}})
	if err != nil {
		t.Fatalf("enqueue second inbound: %v", err)
	}
	items, results, err := engine.ProcessQueuedInboundWork(ctx, "queue-worker", 10)
	if err != nil {
		t.Fatalf("process queued inbound work: %v", err)
	}
	if len(items) != 2 || len(results) != 2 {
		t.Fatalf("expected two processed inbound items/results, got items=%#v results=%#v", items, results)
	}
	if items[0].ID != first.ID || items[1].ID != second.ID {
		t.Fatalf("expected FIFO inbound processing order, got %#v", items)
	}
	if items[0].Status != "completed" || items[1].Status != "completed" {
		t.Fatalf("expected completed inbound items, got %#v", items)
	}
}

func TestProcessQueuedInboundWorkReturnsNoopWhenEmpty(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	items, results, err := engine.ProcessQueuedInboundWork(ctx, "queue-worker", 10)
	if err != nil {
		t.Fatalf("process empty inbound queue: %v", err)
	}
	if len(items) != 0 || len(results) != 0 {
		t.Fatalf("expected empty inbound processing result, got items=%#v results=%#v", items, results)
	}
}

func TestProcessNextInboundWorkProcessesQueuedDirectPrompt(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}}, nil
	})
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_inbound_direct")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_inbound_direct", Title: "@agent", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	queued, err := engine.EnqueueDirectInbound(ctx, DirectInput{Kind: DirectKindPrompt, SessionKey: alloc.SessionKey, Prompt: "hello from inbound queue", Model: "bootstrap", Origin: DirectOrigin{SourceKind: DirectSourceKindIPC, SourceID: "ipc:queue"}})
	if err != nil {
		t.Fatalf("enqueue direct inbound: %v", err)
	}
	if queued.Status != "queued" {
		t.Fatalf("expected queued inbound work, got %#v", queued)
	}
	item, result, err := engine.ProcessNextInboundWork(ctx, "queue-worker")
	if err != nil {
		t.Fatalf("process next inbound work: %v", err)
	}
	if item.Status != "completed" || item.ClaimedBy != "" || item.ClaimedAt != "" {
		t.Fatalf("expected completed inbound work with cleared claim state, got %#v", item)
	}
	if result == nil || result.SessionID != sess.ID || result.TurnID == "" {
		t.Fatalf("unexpected inbound processing result: item=%#v result=%#v", item, result)
	}
	turnRec, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if stringValue(turnRec.Metadata["ingress_source_id"], "") != "ipc:queue" || stringValue(turnRec.Metadata["ingress_session_key"], "") != alloc.SessionKey {
		t.Fatalf("expected ingress metadata on queued direct turn, got %#v", turnRec.Metadata)
	}
}

func TestFinalizeInboundWorkAttemptSurvivesCanceledCallerContextAfterClaim(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	queued, err := s.EnqueueInboundWork(ctx, "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "missing target session"})
	if err != nil {
		t.Fatalf("enqueue bad inbound work: %v", err)
	}
	claimed, err := s.ClaimNextInboundWork(ctx, "queue-worker")
	if err != nil {
		t.Fatalf("claim inbound work: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	item, result, err := engine.finalizeInboundWorkAttempt(cancelCtx, claimed, nil, context.Canceled)
	if err == nil {
		t.Fatalf("expected processing error after canceled context, got item=%#v result=%#v", item, result)
	}
	if item == nil || item.ID != queued.ID || item.Status != "retry" || item.AttemptCount != 1 || item.LastError == "" || item.ClaimedBy != "" || item.ClaimedAt != "" {
		t.Fatalf("expected retry-marked inbound work with cleared claim state after canceled processing context, got queued=%#v item=%#v err=%v", queued, item, err)
	}
	stored, getErr := s.GetInboundWork(ctx, queued.ID)
	if getErr != nil {
		t.Fatalf("get inbound work after canceled processing: %v", getErr)
	}
	if stored.Status != "retry" || stored.AttemptCount != 1 || stored.ClaimedBy != "" || stored.ClaimedAt != "" {
		t.Fatalf("expected durable retry state after canceled processing context, got %#v", stored)
	}
}

func TestProcessNextInboundWorkMarksRetryOnFirstFailure(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	queued, err := s.EnqueueInboundWork(ctx, "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "missing target session"})
	if err != nil {
		t.Fatalf("enqueue bad inbound work: %v", err)
	}
	item, result, err := engine.ProcessNextInboundWork(ctx, "queue-worker")
	if err == nil {
		t.Fatalf("expected processing error for retryable inbound work, got item=%#v result=%#v", item, result)
	}
	if item == nil || item.ID != queued.ID || item.Status != "retry" || item.AttemptCount != 1 || item.LastError == "" || item.NextAttemptAt == "" {
		t.Fatalf("expected retry-marked inbound work item, got queued=%#v item=%#v err=%v", queued, item, err)
	}
}

func TestProcessNextInboundWorkEventuallyMarksFailedAfterRetryBudget(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	queued, err := s.EnqueueInboundWork(ctx, "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "missing target session"})
	if err != nil {
		t.Fatalf("enqueue bad inbound work: %v", err)
	}
	if err := s.RecordInboundWorkRetry(ctx, queued.ID, inboundWorkMaxAttempts-1, "previous failure", 0); err != nil {
		t.Fatalf("seed retry attempt count: %v", err)
	}
	item, result, err := engine.ProcessNextInboundWork(ctx, "queue-worker")
	if err == nil {
		t.Fatalf("expected terminal processing error for exhausted inbound work, got item=%#v result=%#v", item, result)
	}
	if item == nil || item.ID != queued.ID || item.Status != "failed" || item.AttemptCount != inboundWorkMaxAttempts || item.LastError == "" {
		t.Fatalf("expected failed inbound work item after retry budget, got queued=%#v item=%#v err=%v", queued, item, err)
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
	msgs, err := s.ListMessages(ctx, sess.ID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	foundUserIngress := false
	for _, msg := range msgs {
		if msg.Role == "user" && msg.Content == "hello from key" && stringValue(msg.Payload["ingress_session_key"], "") == alloc.SessionKey {
			foundUserIngress = true
		}
	}
	if !foundUserIngress {
		t.Fatalf("expected persisted user message ingress_session_key, got %#v", msgs)
	}
	events, err := s.ListTurnEvents(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	foundStartedIngress := false
	for _, event := range events {
		if event.Type == "turn.started" && stringValue(event.Payload["ingress_session_key"], "") == alloc.SessionKey {
			foundStartedIngress = true
		}
	}
	if !foundStartedIngress {
		t.Fatalf("expected turn.started ingress_session_key, got %#v", events)
	}
}

func TestProcessDirectPromptBySessionKeySurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	engine := New(s)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}}, nil
	})
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_direct_key_ctx")
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "session_direct_key_ctx", Title: "@agent", State: map[string]any{"status": "idle", "queue_count": 0, "model": "bootstrap"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	result, err := engine.ProcessDirect(cancelCtx, DirectInput{Kind: DirectKindPrompt, SessionKey: alloc.SessionKey, Prompt: "hello from canceled key", Model: "bootstrap", Origin: DirectOrigin{SourceKind: DirectSourceKindIPC, SourceID: "ipc:key:ctx"}})
	if err != nil {
		t.Fatalf("process direct prompt by session key with canceled caller context: %v", err)
	}
	waitForCondition(t, time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && strings.TrimSpace(turnRec.FinishedAt) != ""
	}, "direct prompt by session key completion with canceled caller context")
	turnRec, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if turnRec.SessionID != sess.ID || stringValue(turnRec.Metadata["ingress_session_key"], "") != alloc.SessionKey {
		t.Fatalf("expected canceled-caller direct prompt to resolve session key and persist turn, got %#v", turnRec)
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

func TestRunShellReportsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	out, err, cancelled := runShell(ctx, "hello", func(cmd *exec.Cmd) {
		close(started)
		cancel()
	}, nil)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("expected shell command to start before cancellation")
	}
	if !cancelled || err != nil {
		t.Fatalf("expected cancelled shell result without error, got out=%q err=%v cancelled=%v", out, err, cancelled)
	}
}

func TestCloneRouteSessionSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	source, err := s.CreateSession(ctx, "session_root_clone_ctx", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	engine := New(s)
	plan, err := engine.prepareRouteSessionPlan(source, routing.ResolvedRoute{AgentID: "agent1", MatchedBy: "mention"}, inboundContextFromSession(ctx, s, source))
	if err != nil {
		t.Fatalf("prepare route session plan: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	cloned, created, err := engine.cloneRouteSession(cancelCtx, plan)
	if err != nil {
		t.Fatalf("clone route session with canceled caller context: %v", err)
	}
	if !created || cloned == nil {
		t.Fatalf("expected cloned route session, got created=%v cloned=%#v", created, cloned)
	}
	msgs, err := s.ListMessages(ctx, cloned.ID)
	if err != nil {
		t.Fatalf("list cloned session messages: %v", err)
	}
	foundFork := false
	for _, msg := range msgs {
		if msg.Role == "system" && msg.Payload["kind"] == "fork" {
			foundFork = true
		}
	}
	if !foundFork {
		t.Fatalf("expected forked-from notice despite canceled caller context, got %#v", msgs)
	}
}

func TestSubmitPeerRoutedPromptSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	source, err := s.CreateSession(ctx, "session_root_routed_ctx", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	target, err := s.CloneSession(ctx, source.ID, "session_child_routed_ctx", "@agent1", "agent1")
	if err != nil {
		t.Fatalf("create target: %v", err)
	}
	engine := New(s)
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	result, err := engine.submitPeerRoutedPrompt(cancelCtx, source, target, routing.ResolvedRoute{AgentID: "agent1", MatchedBy: "peer-message"}, "hello from peer", "prompt", "bootstrap", false, true, "", nil)
	if err != nil {
		t.Fatalf("submit peer routed prompt with canceled caller context: %v", err)
	}
	if result.SessionID != target.ID || result.TurnID == "" {
		t.Fatalf("unexpected peer routed result: %#v", result)
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && strings.TrimSpace(turnRec.FinishedAt) != ""
	}, "routed peer prompt completion with canceled caller context")
	msgs, err := s.ListMessages(ctx, source.ID)
	if err != nil {
		t.Fatalf("list source messages: %v", err)
	}
	foundRoutingNotice := false
	for _, msg := range msgs {
		if msg.Role == "system" && msg.Payload["kind"] == "routing" {
			foundRoutingNotice = true
		}
	}
	if !foundRoutingNotice {
		t.Fatalf("expected routing notice despite canceled caller context, got %#v", msgs)
	}
	turnRec, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatalf("get target turn: %v", err)
	}
	if turnRec.SessionID != target.ID {
		t.Fatalf("expected target turn in routed target session, got %#v", turnRec)
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

func TestSubTurnSyncResultDeliverySurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_delivery_ctx", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_sync_delivery_ctx", "ChildSync", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_parent_delivery_ctx", "session_parent_delivery_ctx", "running", "parent", map[string]any{"intent": "prompt", "subturn_depth": 0}); err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_child_sync_delivery_ctx", "session_child_sync_delivery_ctx", "completed", "child", map[string]any{"intent": "prompt", "subturn_depth": 1, "parent_turn_id": "turn_parent_delivery_ctx"}); err != nil {
		t.Fatalf("create child turn: %v", err)
	}
	if err := s.AddMessage(ctx, "msg_child_sync_summary_ctx", "session_child_sync_delivery_ctx", "assistant", "child sync result", map[string]any{"kind": "chat", "turn_id": "turn_child_sync_delivery_ctx"}); err != nil {
		t.Fatalf("seed child summary message: %v", err)
	}
	if _, err := s.CreateSubTurn(ctx, "turn_parent_delivery_ctx", "session_parent_delivery_ctx", "turn_child_sync_delivery_ctx", "session_child_sync_delivery_ctx", "sync", 1, map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create sync subturn: %v", err)
	}
	engine := New(s)
	runner := engine.runner("session_child_sync_delivery_ctx")
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	runner.publishSubTurnLifecycle(cancelCtx, "turn_child_sync_delivery_ctx", "completed")
	parentMsgs, err := s.ListMessages(ctx, "session_parent_delivery_ctx")
	if err != nil {
		t.Fatalf("list parent messages: %v", err)
	}
	found := false
	for _, msg := range parentMsgs {
		if msg.Role == "system" && msg.Payload["kind"] == "subturn_result" && msg.Payload["delivery_mode"] == "sync" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected sync subturn delivery despite canceled caller context, got %#v", parentMsgs)
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
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	runner.publishSubTurnLifecycle(cancelCtx, "turn_child_orphan", "completed")

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
