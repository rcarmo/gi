package turn

import (
	"context"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

func TestLaunchedTurnWaitsForSubmissionEvent(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	defer e.Close()
	ctx := context.Background()
	session, err := s.CreateSession(ctx, "submission-order", "Order", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatal(err)
	}
	turn, err := s.CreateTurnWithStatus(ctx, "submission-order-turn", session.ID, "queued", "hello", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatal(err)
	}
	runner := e.runner(session.ID)
	// Hold the submission lock after launching, reproducing a descheduled
	// SubmitPrompt caller before it appends turn.submitted.
	runner.mu.Lock()
	locked := true
	defer func() {
		if locked {
			runner.mu.Unlock()
		}
	}()
	launched, err := e.launchTurnLocked(ctx, runner, session.ID, turn.ID)
	if err != nil || !launched {
		t.Fatalf("launch: %v, %v", launched, err)
	}
	time.Sleep(50 * time.Millisecond)
	events, err := s.ListTurnEvents(ctx, turn.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("runner emitted events before submission lock release: %+v", events)
	}
	if err := s.AppendTurnEvent(ctx, turn.ID, session.ID, "turn.submitted", map[string]any{}); err != nil {
		t.Fatal(err)
	}
	runner.mu.Unlock()
	locked = false
	waitForCondition(t, 3*time.Second, func() bool {
		record, err := s.GetTurn(ctx, turn.ID)
		return err == nil && record.Status != "running" && record.Status != "queued"
	}, "terminal turn")
	events, err = s.ListTurnEvents(ctx, turn.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) < 2 || events[0].Type != "turn.submitted" {
		t.Fatalf("submission must precede runner events: %+v", events)
	}
	var started *store.TurnEvent
	for i := range events {
		if events[i].Type == "turn.started" {
			started = &events[i]
		}
	}
	if started == nil || started.Seq <= events[0].Seq {
		t.Fatalf("missing or out-of-order turn.started: %+v", events)
	}
}
