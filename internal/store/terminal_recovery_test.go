package store

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func stagedRecoveryFixture(t *testing.T) (*Store, string, ActiveTurnClaim) {
	t.Helper()
	s, path := settlementFixture(t)
	ctx := context.Background()
	if claimed, err := s.ClaimSessionActiveTurn(ctx, "s", "t", "dead", "old"); err != nil || !claimed {
		t.Fatal(claimed, err)
	}
	in := sampleTurnSettlement()
	in.ClaimToken = "old"
	if _, _, err := s.StageTurnTerminal(ctx, "t", "s", in); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z' where session_id='s'`); err != nil {
		t.Fatal(err)
	}
	claims, err := s.ListStaleActiveTurnClaims(ctx, time.Now().Add(-time.Second), "s")
	if err != nil || len(claims) != 1 {
		t.Fatal(claims, err)
	}
	return s, path, claims[0]
}
func TestStagedTerminalRecoveryCASRollbackAndFencing(t *testing.T) {
	s, _, claim := stagedRecoveryFixture(t)
	ctx := context.Background()
	cutoff := time.Now().Add(-time.Second)
	if _, err := s.db.Exec(`create trigger reject_recovery before insert on turn_events when new.event_type='turn.recovery_admitted' begin select raise(abort,'reject recovery');end`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AcquireStagedTerminalRecovery(ctx, claim, "new", cutoff, true); err == nil {
		t.Fatal("recovery checkpoint failure hidden")
	}
	_, token, err := s.GetSessionActiveTurn(ctx, "s")
	if err != nil || token != "old" {
		t.Fatal("partial claim transfer", token, err)
	}
	turn, err := s.GetTurn(ctx, "t")
	if err != nil || turn.Status != "running" {
		t.Fatal("partial abort", turn, err)
	}
	if _, err := s.db.Exec(`drop trigger reject_recovery`); err != nil {
		t.Fatal(err)
	}
	forged := claim
	forged.ClaimToken = "foreign"
	if _, err := s.AcquireStagedTerminalRecovery(ctx, forged, "new", cutoff, false); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("foreign token accepted", err)
	}
	stale := claim
	stale.UpdatedAt = "1999-01-01T00:00:00Z"
	if _, err := s.AcquireStagedTerminalRecovery(ctx, stale, "new", cutoff, false); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("stale snapshot accepted", err)
	}
	decision, err := s.AcquireStagedTerminalRecovery(ctx, claim, "new", cutoff, false)
	if err != nil || decision == nil || decision.MessageID != "terminal" || decision.ClaimToken != "new" {
		t.Fatal(decision, err)
	}
	old := sampleTurnSettlement()
	old.ClaimToken = "old"
	if _, _, err := s.StageTurnTerminal(ctx, "t", "s", old); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("old incarnation staged", err)
	}
	if _, err := s.SettleRunningTurn(ctx, "t", "s", old); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("old incarnation settled", err)
	}
	if err := s.ReleaseSessionActiveTurn(ctx, "s", "new"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SettleRunningTurn(ctx, "t", "s", old); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("released recovery resurrected old writer", err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "s", "t", "reused", "new"); err == nil || ok {
		t.Fatal("recovery token reused", ok, err)
	}
	decision.ChildrenStatus = "failed"
	if settled, err := s.SettleRunningTurn(ctx, "t", "s", *decision); err != nil || !settled.Applied {
		t.Fatal(settled, err)
	}
}
func TestStagedTerminalRecoveryRejectsRefreshedAndCompetingOwners(t *testing.T) {
	t.Run("refreshed", func(t *testing.T) {
		s, _, claim := stagedRecoveryFixture(t)
		ctx := context.Background()
		if err := s.TouchSessionActiveTurn(ctx, "s", "old"); err != nil {
			t.Fatal(err)
		}
		if _, err := s.AcquireStagedTerminalRecovery(ctx, claim, "new", time.Now().Add(-time.Second), false); !errors.Is(err, ErrQueueConflict) {
			t.Fatal("fresh owner stolen", err)
		}
	})
	t.Run("competing", func(t *testing.T) {
		s, path, claim := stagedRecoveryFixture(t)
		other, err := Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer other.Close()
		results := make(chan error, 2)
		var wg sync.WaitGroup
		for index, db := range []*Store{s, other} {
			wg.Add(1)
			go func(index int, db *Store) {
				defer wg.Done()
				_, err := db.AcquireStagedTerminalRecovery(context.Background(), claim, []string{"one", "two"}[index], time.Now().Add(-time.Second), false)
				results <- err
			}(index, db)
		}
		wg.Wait()
		close(results)
		won, lost := 0, 0
		for err := range results {
			if err == nil {
				won++
			} else if errors.Is(err, ErrQueueConflict) {
				lost++
			} else {
				t.Fatal(err)
			}
		}
		if won != 1 || lost != 1 {
			t.Fatal("claim adopted twice", won, lost)
		}
	})
}

func TestStagedTerminalRecoveryRejectsIncompleteCheckpoint(t *testing.T) {
	for _, patch := range []string{`{"decision":null}`, `{"decision":{"status":"running","phase":"running"}}`, `{"decision":{"status":"failed","phase":"tools"}}`, `{"decision":{"status":"failed","phase":"failed","message":"missing identity"}}`, `{"decision":{"status":"failed","phase":"failed","message_payload":{"turn_id":"foreign"}}}`} {
		t.Run(patch, func(t *testing.T) {
			s, _, claim := stagedRecoveryFixture(t)
			if _, err := s.db.Exec(`update turn_events set payload_json=? where turn_id='t' and event_type='turn.terminal_staged'`, patch); err != nil {
				t.Fatal(err)
			}
			if _, err := s.AcquireStagedTerminalRecovery(context.Background(), claim, "new", time.Now().Add(-time.Second), false); err == nil {
				t.Fatal("invalid checkpoint admitted")
			}
			_, token, err := s.GetSessionActiveTurn(context.Background(), "s")
			if err != nil || token != "old" {
				t.Fatal("invalid checkpoint changed ownership", token, err)
			}
		})
	}
}
