package store

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
)

func TestManualCompactionAdmissionAtomicClaimAndEvent(t *testing.T) {
	for _, mode := range []string{"race", "rollback"} {
		t.Run(mode, func(t *testing.T) {
			s, err := Open(filepath.Join(t.TempDir(), "admit.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			ctx := context.Background()
			if _, err = s.CreateSession(ctx, "A", "A", nil); err != nil {
				t.Fatal(err)
			}
			for _, id := range []string{"one", "two"} {
				if err = s.AddMessage(ctx, id, "A", "user", id, nil); err != nil {
					t.Fatal(err)
				}
			}
			snapshot, err := s.ContextSnapshot(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			token := ContextToken(snapshot)
			if mode == "rollback" {
				if _, err = s.DB().Exec(`create trigger reject_submit before insert on turn_events begin select raise(abort,'event failed'); end`); err != nil {
					t.Fatal(err)
				}
				if err = s.AdmitManualCompaction(ctx, "A", "manual", token, "test-model"); err == nil {
					t.Fatal("expected rollback")
				}
				for _, table := range []string{"turns", "session_active_turns", "turn_events"} {
					var n int
					if err = s.DB().QueryRow("select count(*) from " + table).Scan(&n); err != nil || n != 0 {
						t.Fatal(table, n, err)
					}
				}
				session, err := s.GetSession(ctx, "A")
				if err != nil || session.State["active_turn_id"] != nil {
					t.Fatal(session, err)
				}
				return
			}
			var wg sync.WaitGroup
			results := make(chan error, 2)
			for i := 0; i < 2; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					results <- s.AdmitManualCompaction(ctx, "A", fmt.Sprintf("manual%d", i), token, "test-model")
				}(i)
			}
			wg.Wait()
			close(results)
			accepted := 0
			for err := range results {
				if err == nil {
					accepted++
				}
			}
			if accepted != 1 {
				t.Fatal(accepted)
			}
			turns, err := s.ListTurns(ctx, "A")
			if err != nil || len(turns) != 1 || turns[0].Status != "running" || turns[0].Prompt != "" {
				t.Fatal(turns, err)
			}
			active, claim, err := s.GetSessionActiveTurn(ctx, "A")
			if err != nil || active != turns[0].ID || claim != active {
				t.Fatal(active, claim, err)
			}
			events, err := s.ListTurnEvents(ctx, active)
			if err != nil || len(events) != 1 || events[0].Type != "turn.submitted" {
				t.Fatal(events, err)
			}
		})
	}
}
