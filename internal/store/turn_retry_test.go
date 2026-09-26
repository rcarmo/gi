package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestHeldRetryReservationSurvivesRestartAndSchemaUpgrade(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "retry.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateTurnWithStatus(ctx, "old", "A", "cancelled", "old prompt", nil); err != nil {
		t.Fatal(err)
	}
	if err = s.HoldTurnFailure(ctx, "old", "review", "held"); err != nil {
		t.Fatal(err)
	}
	// Simulate the pre-migration schema, then let normal Open upgrade it.
	if _, err = s.DB().Exec(`ALTER TABLE turn_failures DROP COLUMN retry_admission_token`); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	f, err := s.GetTurnFailure(ctx, "old")
	if err != nil || f.RetryAdmissionToken != "" || f.HoldState != "review" {
		t.Fatal(f, err)
	}
	if ok, err := s.ReserveHeldRetry(ctx, "old", "token"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	s.Close()
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.ReconcileHeldRetry(ctx, "old", "token", "recover"); !errors.Is(err, ErrRetryPending) {
		t.Fatal("absence replayed", err)
	}
	if _, err = s.ReconcileHeldRetry(ctx, "old", "wrong", "recover"); !errors.Is(err, ErrFailureConflict) {
		t.Fatal("wrong token released", err)
	}
	if _, err = s.CreateTurnWithStatus(ctx, "new", "A", "queued", "old prompt", map[string]any{"retry_of_turn_id": "old", "retry_admission_token": "token"}); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ReconcileHeldRetry(ctx, "old", "token", "recover"); !errors.Is(err, ErrRetryPending) {
		t.Fatal("unconfirmed insertion accepted", err)
	}
	if err = s.AppendTurnEvent(ctx, "new", "A", "turn.submitted", nil); err != nil {
		t.Fatal(err)
	}
	id, err := s.ReconcileHeldRetry(ctx, "old", "token", "recover")
	if err != nil || id != "new" {
		t.Fatal(id, err)
	}
	old, err := s.GetTurn(ctx, "old")
	if err != nil || old.Status != "cancelled" || old.Phase != "aborted" {
		t.Fatal(old, err)
	}
	f, err = s.GetTurnFailure(ctx, "old")
	if err != nil || f.ResolvedTurnID != "new" || f.RetryAdmissionToken != "token" || f.HoldState != "none" {
		t.Fatal(f, err)
	}
	if id, err = s.ReconcileHeldRetry(ctx, "old", "token", "recover"); err != nil || id != "new" {
		t.Fatal("idempotent check", id, err)
	}
	if _, err = s.ReconcileHeldRetry(ctx, "old", "wrong", "recover"); !errors.Is(err, ErrFailureConflict) {
		t.Fatal("wrong token matched resolved state")
	}
	var integrity string
	if err = s.DB().QueryRow(`PRAGMA integrity_check`).Scan(&integrity); err != nil || integrity != "ok" {
		t.Fatal(integrity, err)
	}
}

func TestHeldRetryProtectedFailureCannotBeClearedOrReopened(t *testing.T) {
	for _, state := range []string{"held", "retry_pending", "retried", "skipped"} {
		t.Run(state, func(t *testing.T) {
			ctx := context.Background()
			s, err := Open(filepath.Join(t.TempDir(), "protected.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			if _, err = s.CreateSession(ctx, "A", "A", nil); err != nil {
				t.Fatal(err)
			}
			if _, err = s.CreateTurnWithStatus(ctx, "old", "A", "failed", "old", nil); err != nil {
				t.Fatal(err)
			}
			if err = s.HoldTurnFailure(ctx, "old", "review", "held"); err != nil {
				t.Fatal(err)
			}
			if state == "retry_pending" || state == "retried" {
				if ok, err := s.ReserveHeldRetry(ctx, "old", "token"); err != nil || !ok {
					t.Fatal(ok, err)
				}
			}
			if state == "retried" {
				if _, err = s.CreateTurnWithStatus(ctx, "new", "A", "queued", "old", map[string]any{"retry_of_turn_id": "old", "retry_admission_token": "token"}); err != nil {
					t.Fatal(err)
				}
				if _, err = s.FinishHeldRetry(ctx, "old", "token", "done", false); err != nil {
					t.Fatal(err)
				}
			}
			if state == "skipped" {
				if err = s.ResolveTurnFailure(ctx, "old", "skipped", "done", ""); err != nil {
					t.Fatal(err)
				}
			}
			before, err := s.GetTurnFailure(ctx, "old")
			if err != nil {
				t.Fatal(err)
			}
			if err = s.ClearTurnFailure(ctx, "old"); !errors.Is(err, ErrFailureConflict) {
				t.Fatal("cleared protected failure", err)
			}
			for _, status := range []string{"queued", "running", "completed"} {
				if err = s.UpdateTurnStatusAndPhase(ctx, "old", status, status); !errors.Is(err, ErrFailureConflict) {
					t.Fatal("changed protected status", status, err)
				}
			}
			if state != "held" {
				if err = s.UpsertTurnFailure(ctx, "old", "A", "late", "review", "late"); !errors.Is(err, ErrFailureConflict) {
					t.Fatal("upsert erased resolution", err)
				}
				if err = s.HoldTurnFailure(ctx, "old", "review", "late"); !errors.Is(err, ErrFailureConflict) {
					t.Fatal("rehold erased resolution", err)
				}
			}
			after, err := s.GetTurnFailure(ctx, "old")
			if err != nil || *after != *before {
				t.Fatal("failure changed", before, after, err)
			}
			turn, err := s.GetTurn(ctx, "old")
			if err != nil || turn.Status != "failed" {
				t.Fatal(turn, err)
			}
		})
	}
}

func TestHeldRetryAmbiguousReceiptsRemainHeld(t *testing.T) {
	for _, mode := range []string{"duplicate", "foreign"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			s, err := Open(filepath.Join(t.TempDir(), "ambiguous.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			for _, id := range []string{"A", "B"} {
				if _, err = s.CreateSession(ctx, id, id, nil); err != nil {
					t.Fatal(err)
				}
			}
			if _, err = s.CreateTurnWithStatus(ctx, "old", "A", "failed", "old", nil); err != nil {
				t.Fatal(err)
			}
			if err = s.HoldTurnFailure(ctx, "old", "review", "held"); err != nil {
				t.Fatal(err)
			}
			if ok, err := s.ReserveHeldRetry(ctx, "old", "token"); err != nil || !ok {
				t.Fatal(ok, err)
			}
			session := "A"
			if mode == "foreign" {
				session = "B"
			}
			metadata := map[string]any{"retry_of_turn_id": "old", "retry_admission_token": "token"}
			if _, err = s.CreateTurnWithStatus(ctx, "new", session, "queued", "old", metadata); err != nil {
				t.Fatal(err)
			}
			if mode == "duplicate" {
				if _, err = s.CreateTurnWithStatus(ctx, "other", session, "queued", "old", metadata); err != nil {
					t.Fatal(err)
				}
			}
			if _, err = s.FinishHeldRetry(ctx, "old", "token", "recover", true); !errors.Is(err, ErrFailureConflict) {
				t.Fatal("ambiguous receipt accepted/released", err)
			}
			f, err := s.GetTurnFailure(ctx, "old")
			if err != nil || f.ResolutionState != "retry_pending" || f.HoldState != "review" {
				t.Fatal(f, err)
			}
		})
	}
}
