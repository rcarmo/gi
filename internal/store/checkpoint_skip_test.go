package store

import (
	"context"
	"testing"
	"time"
)

func TestCheckpointSkipStagesFailureWithoutRetry(t *testing.T) {
	for _, kind := range []string{"generation", "provider", "tool"} {
		t.Run(kind, func(t *testing.T) {
			s, old := parkedCheckpointFixture(t, kind)
			ctx := context.Background()
			// Parent linkage is not permission to retry, but abandoning an invocation
			// may settle its original failure without reviving its execution.
			if _, err := s.DB().Exec(`update turns set metadata_json=json_set(metadata_json,'$.parent_turn_id','parent') where id='t'`); err != nil {
				t.Fatal(err)
			}
			decision, handled, err := s.StageHeldCheckpointSkip(ctx, "t", "skip-owner", "operator acknowledged")
			if err != nil || !handled || decision == nil || decision.Status != "failed" || decision.Payload["checkpoint_skip"] == nil {
				t.Fatal(decision, handled, err)
			}
			rec, err := s.GetTurn(ctx, "t")
			if err != nil || rec.Status != "running" || rec.Phase != "settling" {
				t.Fatal(rec, err)
			}
			failure, err := s.GetTurnFailure(ctx, "t")
			if err != nil || failure.HoldState != "none" || failure.ResolutionState != "skipped" || failure.ResolutionSummary != "operator acknowledged" {
				t.Fatal(failure, err)
			}
			var retries int
			if err = s.DB().QueryRow(`select count(*) from turn_events where turn_id='t' and event_type='turn.checkpoint_retry_requested'`).Scan(&retries); err != nil || retries != 0 {
				t.Fatal("skip reset retry budget", retries, err)
			}
			if err = s.AdmitTurnEffect(old, "s", "t", "running", "inference.started", map[string]any{"request_id": "stale"}); err == nil {
				t.Fatal("old writer retained authority")
			}
			if kind == "tool" {
				state, err := s.LoadCancelledToolBatchCheckpoint(WithTurnWriter(ctx, "s", "t", "skip-owner"), "s", "t")
				if err != nil || state == nil {
					t.Fatal("skip cleanup cannot read original slots", state, err)
				}
				if dispatch := readToolBatch(t, s); dispatch != nil {
					t.Fatal("skip granted tool replay authority", dispatch)
				}
			} else {
				tx, err := s.DB().BeginTx(ctx, nil)
				if err != nil {
					t.Fatal(err)
				}
				phase := "generation_ready"
				if kind == "provider" {
					phase = "running"
				}
				checkpoint, _, readErr := generationCheckpointTx(ctx, tx, "s", "t", phase)
				_ = tx.Rollback()
				if readErr != nil || checkpoint != nil {
					t.Fatal("skip granted generation dispatch authority", checkpoint, readErr)
				}
			}
			if claim, _, err := s.QueueHeldCheckpointRetry(ctx, "t", "retry-owner", "revive"); claim != nil {
				t.Fatal("skipped child revived", claim, err)
			}
		})
	}
}

func TestCheckpointSkipRollbackRetainsHoldAndPendingProvider(t *testing.T) {
	s, _ := parkedCheckpointFixture(t, "provider")
	ctx := context.Background()
	if _, err := s.DB().Exec(`create trigger reject_skip before insert on session_active_turns begin select raise(abort,'skip claim rejected');end`); err != nil {
		t.Fatal(err)
	}
	if decision, handled, err := s.StageHeldCheckpointSkip(ctx, "t", "skip-owner", "skip"); err == nil || !handled || decision != nil {
		t.Fatal(decision, handled, err)
	}
	rec, err := s.GetTurn(ctx, "t")
	if err != nil || rec.Status != "failed" || rec.Phase != "held_for_retry_or_skip" {
		t.Fatal(rec, err)
	}
	failure, err := s.GetTurnFailure(ctx, "t")
	if err != nil || failure.HoldState != "review" || failure.ResolutionState != "" {
		t.Fatal(failure, err)
	}
	var rows int
	if err = s.DB().QueryRow(`select (select count(*) from turn_events where turn_id='t' and event_type in ('turn.checkpoint_skip_requested','turn.terminal_staged'))+(select count(*) from turn_receipts where turn_id='t' and occurrence_id='pending')`).Scan(&rows); err != nil || rows != 0 {
		t.Fatal("partial skip escaped rollback", rows, err)
	}
}

func TestCheckpointSkipReopensOnlyTerminalCleanup(t *testing.T) {
	for _, kind := range []string{"generation", "provider", "tool"} {
		t.Run(kind, func(t *testing.T) {
			s, _ := parkedCheckpointFixture(t, kind)
			ctx := context.Background()
			if _, handled, err := s.StageHeldCheckpointSkip(ctx, "t", "skip-owner", "skip"); err != nil || !handled {
				t.Fatal(handled, err)
			}
			claim := staleToolBatchClaim(t, s)
			var seq int
			var name, path string
			if err := s.DB().QueryRow(`pragma database_list`).Scan(&seq, &name, &path); err != nil {
				t.Fatal(err)
			}
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
			reopened, err := Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			decision, err := reopened.AcquireStagedTerminalRecovery(ctx, claim, "reopened-owner", time.Now(), false)
			if err != nil || decision == nil || decision.Status != "failed" || decision.Payload["checkpoint_skip"] == nil {
				t.Fatal(decision, err)
			}
			if kind == "tool" {
				state, err := reopened.LoadCancelledToolBatchCheckpoint(WithTurnWriter(ctx, "s", "t", "reopened-owner"), "s", "t")
				if err != nil || state == nil {
					t.Fatal(state, err)
				}
			}
		})
	}
}

func TestCheckpointSkipDoesNotRestageConsumedTerminal(t *testing.T) {
	s, _ := parkedCheckpointFixture(t, "generation")
	ctx := context.Background()
	if err := s.AppendTurnEvent(ctx, "t", "s", "turn.finished", map[string]any{"status": "failed"}); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkTurnFinished(ctx, "t"); err != nil {
		t.Fatal(err)
	}
	if decision, handled, err := s.StageHeldCheckpointSkip(ctx, "t", "skip-owner", "acknowledge"); err != nil || handled || decision != nil {
		t.Fatal(decision, handled, err)
	}
	if err := s.ResolveTurnFailure(ctx, "t", "skipped", "acknowledge", ""); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := s.DB().QueryRow(`select count(*) from turn_events where turn_id='t' and event_type in ('turn.terminal_staged','turn.checkpoint_skip_requested')`).Scan(&count); err != nil || count != 0 {
		t.Fatal("consumed child restaged", count, err)
	}
}
