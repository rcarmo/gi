package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	goai "github.com/rcarmo/go-ai"
)

func TestToolBatchCloseFencesRecordersAndPreservesBudget(t *testing.T) {
	s, oldCtx, _, call := toolBatchFixture(t)
	startToolBatchCall(t, s, oldCtx, call)
	for i := 0; i < generationReadyRecoveryLimit; i++ {
		if err := s.AppendTurnEvent(oldCtx, "t", "s", "tool.batch_recovered", map[string]any{"phase": "tool_batch", "checkpoint": true, "batch_id": "batch-provider", "iteration": 2}); err != nil {
			t.Fatal(err)
		}
	}
	_, oldToken, err := s.GetSessionActiveTurn(oldCtx, "s")
	if err != nil {
		t.Fatal(err)
	}
	state, err := s.ReserveToolBatchClose(oldCtx, "s", "t", oldToken, "closing")
	if err != nil || state == nil {
		t.Fatalf("reserve: %+v %v", state, err)
	}
	if err = s.SaveToolUsage(oldCtx, "s", "t", "physical-1", goai.Usage{Input: 99}, false, false); !errors.Is(err, ErrQueueConflict) {
		t.Fatalf("old recorder: %v", err)
	}
	closing := WithTurnWriter(oldCtx, "s", "t", "closing")
	if err = s.SaveToolUsage(closing, "s", "t", "physical-1", goai.Usage{Input: 11}, true, true); err != nil {
		t.Fatal(err)
	}
	if err = s.CommitToolBatchClose(context.Background(), "s", "t", "closing"); err != nil {
		t.Fatal(err)
	}
	row, err := s.GetTurn(closing, "t")
	if err != nil || row.Status != "queued" || row.Phase != ClosedToolBatchPhase || row.FinishedAt != "" {
		t.Fatal(row, err)
	}
	if next, err := s.GetNextQueuedTurn(closing, "s"); !errors.Is(err, sql.ErrNoRows) || next != nil {
		t.Fatalf("ordinary launch: %+v %v", next, err)
	}
	claims, err := s.ListStaleActiveTurnClaims(closing, time.Now(), "s")
	if err != nil || len(claims) != 1 {
		t.Fatal(claims, err)
	}
	state, err = s.AcquireToolBatchRecovery(closing, claims[0], "resumed", time.Now())
	if err != nil || state == nil || !state.Calls[0].CallbackReturned || !state.Calls[0].ReceivedUsageUnknown || state.Calls[0].ReceivedUsage.Input != 11 {
		t.Fatal(state, err)
	}
	if err = s.SaveToolUsage(closing, "s", "t", "physical-1", goai.Usage{Input: 100}, false, true); !errors.Is(err, ErrQueueConflict) {
		t.Fatalf("closing recorder escaped: %v", err)
	}
	if _, err = s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
		t.Fatal(err)
	}
	claims, err = s.ListStaleActiveTurnClaims(context.Background(), time.Now(), "s")
	if err != nil || len(claims) != 1 {
		t.Fatal(claims, err)
	}
	if cp, err := s.AcquireToolBatchRecovery(context.Background(), claims[0], "crashed", time.Now()); err != nil || cp != nil {
		t.Fatalf("Close reset crash budget: %+v %v", cp, err)
	}
}

func TestToolBatchCloseRejectsChangedOrCancelledFrontier(t *testing.T) {
	for _, mode := range []string{"cancelled", "unknown", "commit-fault", "stale"} {
		t.Run(mode, func(t *testing.T) {
			s, ctx, _, call := toolBatchFixture(t)
			startToolBatchCall(t, s, ctx, call)
			_, old, err := s.GetSessionActiveTurn(ctx, "s")
			if err != nil {
				t.Fatal(err)
			}
			if cp, err := s.ReserveToolBatchClose(ctx, "s", "t", old, "closing"); err != nil || cp == nil {
				t.Fatal(cp, err)
			}
			switch mode {
			case "cancelled":
				if _, err = s.RequestTurnCancellation(ctx, "t", "s"); err != nil {
					t.Fatal(err)
				}
			case "unknown":
				if err = s.AppendTurnEvent(ctx, "t", "s", "unsupported.frontier", map[string]any{}); err != nil {
					t.Fatal(err)
				}
			case "commit-fault":
				if _, err = s.DB().Exec(`create trigger reject_close before insert on turn_events when new.event_type='turn.checkpoint_queued' begin select raise(abort,'reject close');end`); err != nil {
					t.Fatal(err)
				}
			case "stale":
				if _, err = s.DB().Exec(`update session_active_turns set claim_token='replacement'`); err != nil {
					t.Fatal(err)
				}
			}
			if err = s.CommitToolBatchClose(context.Background(), "s", "t", "closing"); err == nil {
				t.Fatal("unproven checkpoint queued")
			}
			row, err := s.GetTurn(ctx, "t")
			if err != nil || row.Status == "queued" || row.FinishedAt != "" {
				t.Fatal(row, err)
			}
			var queued int
			if err = s.DB().QueryRow(`select count(*) from turn_events where event_type='turn.checkpoint_queued'`).Scan(&queued); err != nil || queued != 0 {
				t.Fatal(queued, err)
			}
		})
	}
}
