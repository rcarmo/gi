package store

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
)

func TestActiveSteeringRequiresCurrentRunningClaim(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "gi.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, id := range []string{"A", "B"} {
		if _, err := s.CreateSession(ctx, id, id, map[string]any{"status": "idle"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.CreateTurnWithStatus(ctx, "run", "A", "running", "original", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatal(err)
	}
	enqueue := func(sid, tid string) error {
		_, err := s.EnqueueActiveSteering(ctx, sid, tid, "user", "new", nil, nil, "")
		return err
	}
	for _, tc := range [][2]string{{"A", "run"}, {"B", "run"}, {"A", "missing"}} {
		if err := enqueue(tc[0], tc[1]); !errors.Is(err, ErrQueueConflict) {
			t.Fatal(tc, err)
		}
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "A", "run", "test", "token"); err != nil {
		t.Fatal(err)
	}
	for _, status := range []string{"completed", "failed", "cancelled", "aborted", "cancelling", "queued"} {
		if err := s.UpdateTurnStatusAndPhase(ctx, "run", status, status); err != nil {
			t.Fatal(err)
		}
		if err := enqueue("A", "run"); !errors.Is(err, ErrQueueConflict) {
			t.Fatal(status, err)
		}
		if count, _ := s.SteeringQueueLength(ctx, "A"); count != 0 {
			t.Fatal("conflict left steering", status, count)
		}
		session, _ := s.GetSession(ctx, "A")
		if session.State["status"] != "idle" {
			t.Fatal("conflict resurrected session", session.State)
		}
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, "run", "running", "inference"); err != nil {
		t.Fatal(err)
	}
	if err := enqueue("A", "run"); err != nil {
		t.Fatal(err)
	}
	session, _ := s.GetSession(ctx, "A")
	if session.State["status"] != "running" || session.State["active_turn_id"] != "run" || session.State["model"] != "bootstrap" {
		t.Fatal(session.State)
	}
	for i := 1; i < 10; i++ {
		if err := enqueue("A", "run"); err != nil {
			t.Fatal(err)
		}
	}
	if err := enqueue("A", "run"); err == nil || errors.Is(err, ErrQueueConflict) {
		t.Fatal("capacity must remain a rejection, not fallback", err)
	}
}

func TestActiveSteeringRollbackAndMaintenance(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "gi.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.CreateSession(ctx, "A", "A", map[string]any{"status": "idle"})
	s.CreateTurnWithStatus(ctx, "run", "A", "running", "", nil)
	s.ClaimSessionActiveTurn(ctx, "A", "run", "test", "token")
	if _, err := s.DB().Exec(`create trigger refuse_normalization before update on sessions begin select raise(abort,'injected'); end`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.EnqueueActiveSteering(ctx, "A", "run", "user", "must rollback", nil, nil, ""); err == nil {
		t.Fatal("fault ignored")
	}
	if count, _ := s.SteeringQueueLength(ctx, "A"); count != 0 {
		t.Fatal("enqueue survived failed state commit")
	}
	s.DB().Exec(`drop trigger refuse_normalization`)
	s.ReleaseSessionActiveTurn(ctx, "A", "token")
	s.CreateTurnWithStatus(ctx, "maintenance", "A", "running", "", map[string]any{"operation": "manual_compaction"})
	s.ClaimSessionActiveTurn(ctx, "A", "maintenance", "test", "new")
	if _, err := s.EnqueueActiveSteering(ctx, "A", "maintenance", "user", "no maintenance steering", nil, nil, ""); !errors.Is(err, ErrQueueConflict) {
		t.Fatal(err)
	}
	if _, err := s.EnqueueActiveSteering(ctx, "A", "run", "user", "stale claim", nil, nil, ""); !errors.Is(err, ErrQueueConflict) {
		t.Fatal(err)
	}
}

func TestActiveSteeringCompletionRaceAcrossStoreConnections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gi.db")
	a, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	ctx := context.Background()
	for i := 0; i < 25; i++ {
		sid := fmt.Sprintf("race-%d", i)
		a.CreateSession(ctx, sid, sid, map[string]any{"status": "running"})
		a.CreateTurnWithStatus(ctx, sid, sid, "running", "first", nil)
		a.ClaimSessionActiveTurn(ctx, sid, sid, "worker", sid)
		start := make(chan struct{})
		admitted, completed := make(chan error, 1), make(chan error, 1)
		go func() {
			<-start
			_, err := a.EnqueueActiveSteering(ctx, sid, sid, "user", "raced", nil, nil, "")
			admitted <- err
		}()
		go func() { <-start; completed <- b.UpdateTurnStatusAndPhase(ctx, sid, "completed", "completed") }()
		close(start)
		enqueueErr := <-admitted
		if err := <-completed; err != nil {
			t.Fatal(err)
		}
		if enqueueErr != nil && !errors.Is(enqueueErr, ErrQueueConflict) {
			t.Fatal(enqueueErr)
		}
		n, err := a.SteeringQueueLength(ctx, sid)
		if err != nil {
			t.Fatal(err)
		}
		if enqueueErr == nil && n != 1 || enqueueErr != nil && n != 0 {
			t.Fatal("non-atomic admission", enqueueErr, n)
		}
		// Once completion is committed, no further steering can enter even though
		// the active claim still belongs to the exhausted run.
		if _, err := b.EnqueueActiveSteering(ctx, sid, sid, "user", "too late", nil, nil, ""); !errors.Is(err, ErrQueueConflict) {
			t.Fatal(err)
		}
	}
}
