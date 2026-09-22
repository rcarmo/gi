package turn

import (
	"context"
	"errors"
	"testing"

	"github.com/rcarmo/gi/internal/store"
)

func TestQueueCancelledBeforeLaunchClaimCannotRun(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	e := New(s)
	defer e.Close()
	if _, err := s.CreateSession(ctx, "launch-cancel", "Queue", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "queued", "launch-cancel", "queued", "must not run", nil); err != nil {
		t.Fatal(err)
	}
	e.beforeLaunchClaimHook = func(ctx context.Context, sessionID, turnID string) {
		if err := s.CancelQueuedTurn(ctx, sessionID, turnID); err != nil {
			t.Fatal(err)
		}
	}
	runner := e.runner("launch-cancel")
	runner.mu.Lock()
	launched, err := e.launchTurnLocked(ctx, runner, "launch-cancel", "queued")
	runner.mu.Unlock()
	if launched || !errors.Is(err, store.ErrQueueConflict) {
		t.Fatalf("cancelled turn launched: %v %v", launched, err)
	}
	current, _ := s.GetTurn(ctx, "queued")
	if current.Status != "cancelled" {
		t.Fatal("cancel overwritten")
	}
	events, _ := s.ListTurnEvents(ctx, "queued")
	if len(events) != 0 {
		t.Fatal("cancelled turn emitted runtime events")
	}
}

func TestExplicitQueueDoesNotSteerActiveTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	e := New(s)
	defer e.Close()
	session, err := s.CreateSession(ctx, "queue-origin", "Queue", map[string]any{"status": "running", "active_turn_id": "active", "model": "bootstrap"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "active", session.ID, "running", "active", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, session.ID, "active", "test", "token"); err != nil {
		t.Fatal(err)
	}
	result, err := e.SubmitPrompt(ctx, RunInput{SessionID: session.ID, Prompt: "follow-up", Intent: "queue", Model: "bootstrap"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Queued || result.TurnID == "active" || result.Status != "queued" {
		t.Fatalf("queue became steer: %+v", result)
	}
	queue, err := s.ListQueuedTurns(ctx, session.ID)
	if err != nil || len(queue) != 1 || queue[0].ID != result.TurnID {
		t.Fatalf("missing durable queue: %+v %v", queue, err)
	}
	state, _ := s.GetSession(ctx, session.ID)
	if state.State["status"] != "running" || state.State["active_turn_id"] != "active" {
		t.Fatalf("active state replaced: %+v", state.State)
	}
	if err := e.CancelQueuedTurn(ctx, session.ID, "active"); !errors.Is(err, store.ErrQueueConflict) {
		t.Fatalf("stale cancel accepted: %v", err)
	}
	if err := e.CancelQueuedTurn(ctx, session.ID, result.TurnID); err != nil {
		t.Fatal(err)
	}
	if err := e.CancelQueuedTurn(ctx, session.ID, result.TurnID); err != nil {
		t.Fatal("cancel retry not idempotent", err)
	}
	events, err := s.ListTurnEvents(ctx, result.TurnID)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, event := range events {
		if event.Type == "turn.cancelled" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("duplicate cancellation events: %d", count)
	}
}
