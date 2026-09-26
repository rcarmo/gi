package store

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
)

func atomicHeldRetryFixture(t *testing.T) (*Store, string, HeldRetryAdmission, map[string]any) {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "atomic-retry.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	if _, err = s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateTurnWithStatus(ctx, "old", "A", "failed", "old prompt", nil); err != nil {
		t.Fatal(err)
	}
	if err = s.HoldTurnFailure(ctx, "old", "review", "held"); err != nil {
		t.Fatal(err)
	}
	retry := HeldRetryAdmission{OriginalTurnID: "old", Token: "atomic-token", Summary: "explicit retry"}
	if ok, err := s.ReserveAtomicHeldRetry(ctx, retry.OriginalTurnID, retry.Token); err != nil || !ok {
		t.Fatal(ok, err)
	}
	return s, path, retry, map[string]any{"retry_of_turn_id": retry.OriginalTurnID, "retry_admission_token": retry.Token, "intent": "queue"}
}

func TestHeldRetryAtomicReleaseFencesOldOwnerAfterRestart(t *testing.T) {
	s, path, retry, metadata := atomicHeldRetryFixture(t)
	ctx := context.Background()
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, args := range [][3]string{{"B", "old", retry.Token}, {"A", "old", "wrong"}, {"A", "foreign", retry.Token}} {
		if err = s.ReleaseUnadmittedHeldRetry(ctx, args[0], args[1], args[2]); !errors.Is(err, ErrFailureConflict) {
			t.Fatal("wrong ownership released", args, err)
		}
	}
	if _, err = s.ReconcileHeldRetry(ctx, "old", retry.Token, "check only"); !errors.Is(err, ErrRetryPending) {
		t.Fatal("check implicitly released", err)
	}
	if err = s.ReleaseUnadmittedHeldRetry(ctx, "A", "old", retry.Token); err != nil {
		t.Fatal(err)
	}
	if err = s.AdmitHeldRetry(ctx, retry, "late", "A", "old prompt", metadata, nil); !errors.Is(err, ErrFailureConflict) {
		t.Fatal("stale owner admitted", err)
	}
	if ok, err := s.ReserveAtomicHeldRetry(ctx, "old", "new-token"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if err = s.AdmitHeldRetry(ctx, retry, "still-late", "A", "old prompt", metadata, nil); !errors.Is(err, ErrFailureConflict) {
		t.Fatal("stale owner reused new reservation", err)
	}
	retry.Token = "new-token"
	metadata["retry_admission_token"] = retry.Token
	if err = s.AdmitHeldRetry(ctx, retry, "new", "A", "old prompt", metadata, nil); err != nil {
		t.Fatal(err)
	}
	if err = s.ReleaseUnadmittedHeldRetry(ctx, "A", "old", retry.Token); !errors.Is(err, ErrFailureConflict) {
		t.Fatal("released committed admission", err)
	}
	rows, err := s.ListTurns(ctx, "A")
	if err != nil || len(rows) != 2 {
		t.Fatal(rows, err)
	}
}

func TestHeldRetryAtomicReleaseAndAdmissionHaveOneWinner(t *testing.T) {
	for i := 0; i < 8; i++ {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			s, path, retry, metadata := atomicHeldRetryFixture(t)
			ctx := context.Background()
			other, err := Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer other.Close()
			start := make(chan struct{})
			var wg sync.WaitGroup
			var admitted, released error
			wg.Add(2)
			go func() {
				defer wg.Done()
				<-start
				admitted = s.AdmitHeldRetry(ctx, retry, "new", "A", "old prompt", metadata, nil)
			}()
			go func() {
				defer wg.Done()
				<-start
				released = other.ReleaseUnadmittedHeldRetry(ctx, "A", "old", retry.Token)
			}()
			close(start)
			wg.Wait()
			if (admitted == nil) == (released == nil) {
				t.Fatal("expected one winner", admitted, released)
			}
			loser := admitted
			if admitted == nil {
				loser = released
			}
			if !errors.Is(loser, ErrFailureConflict) {
				t.Fatal("unexpected conflict", loser)
			}
			f, err := s.GetTurnFailure(ctx, "old")
			if err != nil {
				t.Fatal(err)
			}
			turns, err := s.ListTurns(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			if admitted == nil {
				if len(turns) != 2 || f.ResolvedTurnID != "new" || f.ResolutionState != "retried" {
					t.Fatal(turns, f)
				}
			} else {
				if len(turns) != 1 || f.ResolutionState != "" || f.RetryAdmissionToken != "" {
					t.Fatal(turns, f)
				}
				if err = s.AdmitHeldRetry(ctx, retry, "late", "A", "old prompt", metadata, nil); !errors.Is(err, ErrFailureConflict) {
					t.Fatal("released token admitted", err)
				}
			}
		})
	}
}

func TestHeldRetryAtomicSubturnTransactionAndRestart(t *testing.T) {
	for _, fail := range []bool{true, false} {
		t.Run(map[bool]string{true: "rollback", false: "commit"}[fail], func(t *testing.T) {
			s, path, retry, metadata := atomicHeldRetryFixture(t)
			ctx := context.Background()
			if _, err := s.CreateTurnWithStatus(ctx, "parent", "A", "completed", "parent", nil); err != nil {
				t.Fatal(err)
			}
			sub := &RetrySubTurn{ParentTurnID: "parent", ParentSessionID: "A", DeliveryMode: "async", Depth: 1, Metadata: map[string]any{"effective_tools": []string{"read"}}}
			if fail {
				if _, err := s.DB().Exec(`CREATE TRIGGER reject_subturn BEFORE INSERT ON subturns BEGIN SELECT RAISE(ABORT,'injected subturn failure'); END`); err != nil {
					t.Fatal(err)
				}
			}
			err := s.AdmitHeldRetry(ctx, retry, "child", "A", "old prompt", metadata, sub)
			if fail && err == nil {
				t.Fatal("expected rollback")
			}
			if !fail && err != nil {
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
			f, err := s.GetTurnFailure(ctx, "old")
			if err != nil {
				t.Fatal(err)
			}
			var children int
			if err = s.DB().QueryRow(`select count(*) from subturns where child_turn_id='child'`).Scan(&children); err != nil {
				t.Fatal(err)
			}
			if fail {
				if f.ResolutionState != "retry_pending" || children != 0 {
					t.Fatal(f, children)
				}
				turns, err := s.ListTurns(ctx, "A")
				if err != nil || len(turns) != 2 {
					t.Fatal("provisional turn survived", turns, err)
				}
				if err = s.ReleaseUnadmittedHeldRetry(ctx, "A", "old", retry.Token); err != nil {
					t.Fatal(err)
				}
			} else {
				if f.ResolutionState != "retried" || f.ResolvedTurnID != "child" || children != 1 {
					t.Fatal(f, children)
				}
				if id, err := s.ReconcileHeldRetry(ctx, "old", retry.Token, "recover"); err != nil || id != "child" {
					t.Fatal(id, err)
				}
				if count, err := s.CountQueuedTurns(ctx, "A"); err != nil || count != 1 {
					t.Fatal(count, err)
				}
			}
		})
	}
}

func TestHeldRetryAtomicVersionUpgradeDoesNotReleaseLegacyReservation(t *testing.T) {
	s, path, retry, _ := atomicHeldRetryFixture(t)
	ctx := context.Background()
	// Version-zero is the previous release's schema and reservation contract.
	if _, err := s.DB().Exec(`ALTER TABLE turn_failures DROP COLUMN retry_admission_version`); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	f, err := s.GetTurnFailure(ctx, "old")
	if err != nil || f.RetryAdmissionVersion != 0 || f.RetryAdmissionToken != retry.Token {
		t.Fatal(f, err)
	}
	if err = s.ReleaseUnadmittedHeldRetry(ctx, "A", "old", retry.Token); !errors.Is(err, ErrRetryPending) {
		t.Fatal("legacy reservation released", err)
	}
}
