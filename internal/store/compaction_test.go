package store

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
)

func compactionStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "compact.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	ctx := context.Background()
	if _, err = s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateTurn(ctx, "run", "A", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "run", "runner", "run"); !ok || err != nil {
		t.Fatal(ok, err)
	}
	return s
}
func TestCompactionCommitAndOccurrenceOwnership(t *testing.T) {
	s := compactionStore(t)
	ctx := context.Background()
	if _, err := s.BeginCompaction(ctx, "foreign", "run", nil); !errors.Is(err, ErrCompactionInactive) {
		t.Fatal(err)
	}
	seq, err := s.BeginCompaction(ctx, "A", "run", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.BeginCompaction(ctx, "A", "run", nil); !errors.Is(err, ErrCompactionInactive) {
		t.Fatal(err)
	}
	if out, err := s.FinishCompaction(ctx, "A", "run", seq, "completed", "first summary", nil); err != nil || out != "completed" {
		t.Fatal(out, err)
	}
	if _, err := s.FinishCompaction(ctx, "A", "run", seq, "completed", "duplicate", nil); !errors.Is(err, ErrCompactionInactive) {
		t.Fatal(err)
	}
	seq2, err := s.BeginCompaction(ctx, "A", "run", nil)
	if err != nil || seq2 <= seq {
		t.Fatal(seq2, err)
	}
	if _, err := s.FinishCompaction(ctx, "A", "run", seq, "completed", "stale", nil); !errors.Is(err, ErrCompactionInactive) {
		t.Fatal(err)
	}
	if out, err := s.FinishCompaction(ctx, "A", "run", seq2, "completed", "second summary", nil); err != nil || out != "completed" {
		t.Fatal(out, err)
	}
	msgs, err := s.ListMessages(ctx, "A")
	if err != nil || len(msgs) != 2 || msgs[0].ID == msgs[1].ID {
		t.Fatal(msgs, err)
	}
	events, err := s.ListTurnEvents(ctx, "run")
	if err != nil || len(events) != 4 {
		t.Fatal(events, err)
	}
	turn, err := s.GetTurn(ctx, "run")
	if err != nil || turn.Status != "running" || turn.Phase != "running" {
		t.Fatal(turn, err)
	}
}
func TestCompactionCancelCannotBeResurrected(t *testing.T) {
	s := compactionStore(t)
	ctx := context.Background()
	seq, err := s.BeginCompaction(ctx, "A", "run", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, "run", "cancelling", "cancelling"); err != nil {
		t.Fatal(err)
	}
	if out, err := s.FinishCompaction(ctx, "A", "run", seq, "completed", "must not persist", nil); err != nil || out != "cancelled" {
		t.Fatal(out, err)
	}
	if err := s.SetClaimedRunningPhase(ctx, "A", "run", "running"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	turn, _ := s.GetTurn(ctx, "run")
	if turn.Status != "cancelling" || turn.Phase != "cancelling" {
		t.Fatal(turn)
	}
	msgs, err := s.ListMessages(ctx, "A")
	if err != nil || len(msgs) != 0 {
		t.Fatal(msgs, err)
	}
	events, _ := s.ListTurnEvents(ctx, "run")
	if len(events) != 2 || events[1].Type != "compaction.cancelled" {
		t.Fatal(events)
	}
}
func TestCompactionSummaryFailureRollsBackCompletion(t *testing.T) {
	s := compactionStore(t)
	ctx := context.Background()
	seq, err := s.BeginCompaction(ctx, "A", "run", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().Exec(`create trigger fail_summary before insert on messages begin select raise(abort,'summary failure'); end`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.FinishCompaction(ctx, "A", "run", seq, "completed", "summary", nil); err == nil {
		t.Fatal("wanted rollback")
	}
	events, _ := s.ListTurnEvents(ctx, "run")
	if len(events) != 1 || events[0].Type != "compaction.started" {
		t.Fatal(events)
	}
	turn, _ := s.GetTurn(ctx, "run")
	if turn.Phase != "compacting" {
		t.Fatal(turn)
	}
	if out, err := s.FinishCompaction(ctx, "A", "run", seq, "failed", "", map[string]any{"detail": "summary failure"}); err != nil || out != "failed" {
		t.Fatal(out, err)
	}
	turn, _ = s.GetTurn(ctx, "run")
	if turn.Phase != "running" {
		t.Fatal(turn)
	}
}
func TestCompactionStartEventFailureRollsBackPhase(t *testing.T) {
	s := compactionStore(t)
	ctx := context.Background()
	if _, err := s.DB().Exec(`create trigger fail_start before insert on turn_events begin select raise(abort,'start failure'); end`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.BeginCompaction(ctx, "A", "run", nil); err == nil {
		t.Fatal("wanted failure")
	}
	turn, _ := s.GetTurn(ctx, "run")
	if turn.Phase != "setup" {
		t.Fatal(turn)
	}
}

func TestCompactionCompletionRacesCancellation(t *testing.T) {
	for i := 0; i < 8; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			s := compactionStore(t)
			ctx := context.Background()
			seq, err := s.BeginCompaction(ctx, "A", "run", nil)
			if err != nil {
				t.Fatal(err)
			}
			gate := make(chan struct{})
			done := make(chan error, 1)
			go func() { <-gate; done <- s.UpdateTurnStatusAndPhase(ctx, "run", "cancelling", "cancelling") }()
			close(gate)
			out, finishErr := s.FinishCompaction(ctx, "A", "run", seq, "completed", "summary", nil)
			if err := <-done; err != nil {
				t.Fatal(err)
			}
			if finishErr != nil {
				t.Fatal(finishErr)
			}
			turn, err := s.GetTurn(ctx, "run")
			if err != nil || turn.Status != "cancelling" || turn.Phase != "cancelling" {
				t.Fatal(turn, err)
			}
			msgs, err := s.ListMessages(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			switch out {
			case "completed":
				if len(msgs) != 1 {
					t.Fatal(msgs)
				}
			case "cancelled":
				if len(msgs) != 0 {
					t.Fatal(msgs)
				}
			default:
				t.Fatal(out)
			}
			events, err := s.ListTurnEvents(ctx, "run")
			if err != nil || len(events) != 2 || events[1].Type != "compaction."+out {
				t.Fatal(events, err)
			}
		})
	}
}
