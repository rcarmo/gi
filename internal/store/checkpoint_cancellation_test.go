package store

import (
	"context"
	"errors"
	"testing"
	"time"

	goai "github.com/rcarmo/go-ai"
)

func closedCancellationFixture(t *testing.T, kind string) (*Store, context.Context) {
	t.Helper()
	var s *Store
	var ctx context.Context
	if kind == "tool" {
		var call goai.ToolCall
		s, ctx, _, call = toolBatchFixture(t)
		startToolBatchCall(t, s, ctx, call)
	} else if kind == "generation" {
		s, ctx, _ = generationReadyFixture(t)
		if err := s.SaveGenerationReady(ctx, "s", "t", readyCheckpoint()); err != nil {
			t.Fatal(err)
		}
	} else {
		s, ctx, _ = retryReceiptFixture(t)
		receipt := retryReceipt()
		receipt.RetryWait.Lifecycle = &ProviderLifecyclePolicy{Version: 1}
		if kind == "retry-ready" {
			receipt.RetryWait.RetryAt = time.Now().Add(-time.Second)
		}
		if _, err := s.SettleProviderAttempt(ctx, "t", "s", "request", receipt); err != nil {
			t.Fatal(err)
		}
		if kind == "retry-ready" {
			if err := s.ResumeProviderRetry(ctx, "s", "t", "request"); err != nil {
				t.Fatal(err)
			}
		}
	}
	_, old, err := s.GetSessionActiveTurn(ctx, "s")
	if err != nil {
		t.Fatal(err)
	}
	switch kind {
	case "tool":
		if cp, err := s.ReserveToolBatchClose(ctx, "s", "t", old, "closed"); err != nil || cp == nil {
			t.Fatal(cp, err)
		}
		err = s.CommitToolBatchClose(ctx, "s", "t", "closed")
	case "generation":
		if cp, err := s.ReserveGenerationClose(ctx, "s", "t", old, "closed"); err != nil || cp == nil {
			t.Fatal(cp, err)
		}
		err = s.CommitGenerationClose(ctx, "s", "t", "closed")
	default:
		if cp, err := s.ReserveProviderRetryClose(ctx, "s", "t", old, "closed"); err != nil || cp == nil {
			t.Fatal(cp, err)
		}
		err = s.CommitProviderRetryClose(ctx, "s", "t", "closed")
	}
	if err != nil {
		t.Fatal(err)
	}
	return s, ctx
}

func TestCheckpointCancellationAtomicStageAndReopen(t *testing.T) {
	for _, kind := range []string{"tool", "generation", "retry-wait", "retry-ready"} {
		t.Run(kind, func(t *testing.T) {
			s, ctx := closedCancellationFixture(t, kind)
			before, err := s.GetTurn(ctx, "t")
			if err != nil {
				t.Fatal(err)
			}
			decision, err := s.StageClosedCheckpointCancellation(ctx, "s", "t", "cancel")
			if err != nil || decision == nil || decision.Payload["checkpoint_cancellation"] != before.Phase {
				t.Fatal(decision, err)
			}
			row, err := s.GetTurn(ctx, "t")
			if err != nil || row.Status != "cancelling" || row.Phase != "settling" || row.FinishedAt != "" {
				t.Fatal(row, err)
			}
			if _, err = s.StageClosedCheckpointCancellation(ctx, "s", "t", "duplicate"); !errors.Is(err, ErrQueueConflict) {
				t.Fatalf("duplicate took ownership: %v", err)
			}
			if kind == "tool" {
				if cp := readToolBatch(t, s); cp != nil {
					t.Fatal("cancellation accepted by executor recovery reader")
				}
				if cp, err := s.LoadCancelledToolBatchCheckpoint(ctx, "s", "t"); err != nil || cp == nil {
					t.Fatal(cp, err)
				}
			}
			var seq int
			var name, path string
			if err = s.DB().QueryRow(`pragma database_list`).Scan(&seq, &name, &path); err != nil {
				t.Fatal(err)
			}
			if _, err = s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
				t.Fatal(err)
			}
			if err = s.Close(); err != nil {
				t.Fatal(err)
			}
			s, err = Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			claims, err := s.ListStaleActiveTurnClaims(context.Background(), time.Now(), "s")
			if err != nil || len(claims) != 1 {
				t.Fatal(claims, err)
			}
			decision, err = s.AcquireStagedTerminalRecovery(context.Background(), claims[0], "recovered", time.Now(), false)
			if err != nil || decision == nil || decision.Status != "cancelled" || decision.Payload["checkpoint_cancellation"] != before.Phase {
				t.Fatal(decision, err)
			}
			if kind == "tool" {
				if cp, err := s.LoadCancelledToolBatchCheckpoint(context.Background(), "s", "t"); err != nil || cp == nil {
					t.Fatal(cp, err)
				}
			}
		})
	}
}

func TestCheckpointCancellationFailureRollsBackReservation(t *testing.T) {
	s, ctx := closedCancellationFixture(t, "tool")
	if _, err := s.DB().Exec(`create trigger reject_cancel before insert on turn_events when new.event_type='turn.terminal_staged' begin select raise(abort,'cancel rejected');end`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.StageClosedCheckpointCancellation(ctx, "s", "t", "cancel"); err == nil {
		t.Fatal("failed stage acknowledged")
	}
	row, err := s.GetTurn(ctx, "t")
	_, token, claimErr := s.GetSessionActiveTurn(ctx, "s")
	if err != nil || claimErr != nil || row.Status != "queued" || row.Phase != ClosedToolBatchPhase || token != "closed" {
		t.Fatal(row, token, err, claimErr)
	}
}

func TestCheckpointCancellationAgentEndClaimOnceAndFenced(t *testing.T) {
	s, ctx, _ := generationReadyFixture(t)
	if run, err := s.ClaimAgentEnd(ctx, "s", "t"); err != nil || !run {
		t.Fatal(run, err)
	}
	if run, err := s.ClaimAgentEnd(ctx, "s", "t"); err != nil || run {
		t.Fatal(run, err)
	}
	if _, err := s.CreateSession(ctx, "foreign", "foreign", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ClaimAgentEnd(ctx, "foreign", "t"); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("foreign lifecycle claim", err)
	}
	if _, err := s.DB().Exec(`update session_active_turns set claim_token='new';insert into turn_recovery_fences(turn_id,session_id,claim_token) values('t','s','new')`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ClaimAgentEnd(ctx, "s", "t"); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("stale lifecycle claim", err)
	}
}

func TestCheckpointCancellationRacesWithExecutorAdoption(t *testing.T) {
	s, ctx := closedCancellationFixture(t, "tool")
	claims, err := s.ListStaleActiveTurnClaims(ctx, time.Now(), "s")
	if err != nil || len(claims) != 1 {
		t.Fatal(claims, err)
	}
	start, cancelDone, adoptDone := make(chan struct{}), make(chan error, 1), make(chan error, 1)
	go func() {
		<-start
		_, err := s.StageClosedCheckpointCancellation(ctx, "s", "t", "cancel")
		cancelDone <- err
	}()
	go func() {
		<-start
		cp, err := s.AcquireToolBatchRecovery(ctx, claims[0], "executor", time.Now())
		if cp == nil && err == nil {
			err = ErrQueueConflict
		}
		adoptDone <- err
	}()
	close(start)
	cancelErr, adoptErr := <-cancelDone, <-adoptDone
	if (cancelErr == nil) == (adoptErr == nil) {
		t.Fatal("both/neither admitted", cancelErr, adoptErr)
	}
	if cancelErr != nil && !errors.Is(cancelErr, ErrQueueConflict) {
		t.Fatal(cancelErr)
	}
	if adoptErr != nil && !errors.Is(adoptErr, ErrQueueConflict) {
		t.Fatal(adoptErr)
	}
	row, err := s.GetTurn(ctx, "t")
	if err != nil {
		t.Fatal(err)
	}
	if cancelErr == nil && (row.Status != "cancelling" || row.Phase != "settling") {
		t.Fatal(row)
	}
	if adoptErr == nil && (row.Status != "running" || row.Phase != "waiting_on_tools") {
		t.Fatal(row)
	}
}
