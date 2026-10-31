package store

import (
	"context"
	"errors"
	"testing"
)

func TestCancellationMarkAtomicAndCannotReopenTerminal(t *testing.T) {
	s, _ := settlementFixture(t)
	ctx := context.Background()
	if _, err := s.db.Exec(`create trigger reject_cancel before insert on turn_events when new.event_type='turn.cancelling' begin select raise(abort,'reject cancellation');end`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RequestTurnCancellation(ctx, "t", "s"); err == nil {
		t.Fatal("cancellation checkpoint failure hidden")
	}
	turn, err := s.GetTurn(ctx, "t")
	if err != nil || turn.Status != "running" {
		t.Fatal("cancellation mark escaped rollback", turn, err)
	}
	if _, err := s.db.Exec(`drop trigger reject_cancel`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RequestTurnCancellation(ctx, "t", "foreign"); err == nil {
		t.Fatal("foreign cancellation accepted")
	}
	marked, err := s.RequestTurnCancellation(ctx, "t", "s")
	if err != nil || !marked {
		t.Fatal(marked, err)
	}
	marked, err = s.RequestTurnCancellation(ctx, "t", "s")
	if err != nil || marked {
		t.Fatal("duplicate cancellation mark", marked, err)
	}
	decision := sampleTurnSettlement()
	decision.Status = "cancelled"
	decision.Phase = "aborted"
	settled, err := s.SettleRunningTurn(ctx, "t", "s", decision)
	if err != nil || !settled.Applied {
		t.Fatal(settled, err)
	}
	if _, err := s.RequestTurnCancellation(ctx, "t", "s"); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("terminal reopened", err)
	}
	turn, err = s.GetTurn(ctx, "t")
	if err != nil || turn.Status != "cancelled" || turn.FinishedAt == "" {
		t.Fatal("late cancellation altered receipt", turn, err)
	}
	events, err := s.ListTurnEvents(ctx, "t")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, event := range events {
		if event.Type == "turn.cancelling" {
			count++
		}
	}
	if count != 1 {
		t.Fatal("duplicate cancellation checkpoint", count)
	}
}
