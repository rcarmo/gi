package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"sync"
	"testing"
)

func TestQueueSteerRunOwnershipAtomicityAndRecovery(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "steer.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, id := range []string{"A", "B"} {
		if _, err = s.CreateSession(ctx, id, "test", nil); err != nil {
			t.Fatal(err)
		}
	}
	for _, v := range []struct{ id, session, status string }{{"active", "A", "running"}, {"next", "A", "running"}, {"foreign", "B", "running"}, {"q1", "A", "queued"}, {"q2", "A", "queued"}, {"qB", "B", "queued"}} {
		if _, err = s.CreateTurnWithStatus(ctx, v.id, v.session, v.status, v.id, map[string]any{"media": []any{map[string]any{"media_id": 7, "session_id": v.session}}, "client_request_id": v.id}); err != nil {
			t.Fatal(err)
		}
	}
	reject := func(session, queued, active string) {
		t.Helper()
		if err := s.SteerQueuedTurn(ctx, session, queued, active); !errors.Is(err, ErrQueueConflict) {
			t.Fatalf("%s/%s/%s: %v", session, queued, active, err)
		}
	}
	reject("A", "q1", "")
	reject("A", "q1", "active") // idle despite running row
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "active", "runner", "token"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	reject("A", "q1", "next")
	reject("A", "qB", "active")
	reject("B", "qB", "active")
	reject("A", "active", "active")
	if err := s.UpdateTurnStatus(ctx, "active", "cancelling"); err != nil {
		t.Fatal(err)
	}
	reject("A", "q1", "active")
	if err := s.UpdateTurnStatus(ctx, "active", "running"); err != nil {
		t.Fatal(err)
	}
	// Force failure after the source UPDATE: the original row must roll back.
	if _, err = s.DB().Exec(`create trigger reject_steer before insert on steering_queue begin select raise(abort,'injected failure'); end`); err != nil {
		t.Fatal(err)
	}
	if err = s.SteerQueuedTurn(ctx, "A", "q1", "active"); err == nil {
		t.Fatal("expected failure")
	}
	q, _ := s.GetTurn(ctx, "q1")
	if q.Status != "queued" {
		t.Fatal(q)
	}
	if _, err = s.DB().Exec(`drop trigger reject_steer`); err != nil {
		t.Fatal(err)
	}
	// Two tabs/requests competing for one durable ID admit exactly one.
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- s.SteerQueuedTurn(ctx, "A", "q1", "active") }()
	}
	wg.Wait()
	close(results)
	accepted := 0
	for err := range results {
		if err == nil {
			accepted++
		} else {
			t.Logf("admission: %v", err)
		}
	}
	if accepted != 1 {
		t.Fatalf("accepted %d", accepted)
	}
	reject("A", "q1", "active")
	for _, active := range []string{"", "next", "foreign"} {
		msgs, err := s.DequeueSteeringForTurn(ctx, "A", active)
		if err != sql.ErrNoRows || len(msgs) != 0 {
			t.Fatal(active, msgs, err)
		}
	}
	// Continuations must exclude bound messages even when no ordinary queue exists.
	if err = s.CancelQueuedTurn(ctx, "A", "q2"); err != nil {
		t.Fatal(err)
	}
	if _, _, err = s.StageSteeringContinuation(ctx, "A", "continuation"); err != sql.ErrNoRows {
		t.Fatal(err)
	}
	msgs, err := s.DequeueSteeringForTurn(ctx, "A", "active")
	if err != nil || len(msgs) != 1 {
		t.Fatal(msgs, err)
	}
	if msgs[0].TurnID != "active" || msgs[0].Payload["source_queue_id"] != "q1" || msgs[0].Content != "q1" || msgs[0].Payload["media"] == nil {
		t.Fatal(msgs)
	}
	if err := s.PersistBoundSteering(ctx, "A", "active", msgs[0], map[string]any{"turn_id": "active", "source_queue_id": "q1"}); err != nil {
		t.Fatal(err)
	}
	if err := s.PersistBoundSteering(ctx, "A", "active", msgs[0], nil); !errors.Is(err, ErrQueueConflict) {
		t.Fatal(err)
	}
	q, _ = s.GetTurn(ctx, "q1")
	if q.Status != "cancelled" || q.Phase != "steered" {
		t.Fatal(q)
	}
	reject("A", "q1", "active")
	if _, err = s.CreateTurnWithStatus(ctx, "q3", "A", "queued", "recover me", nil); err != nil {
		t.Fatal(err)
	}
	if err = s.SteerQueuedTurn(ctx, "A", "q3", "active"); err != nil {
		t.Fatal(err)
	}
	if msgs, err := s.DequeueSteeringForTurn(ctx, "A", "active"); err != nil || len(msgs) != 1 {
		t.Fatal(msgs, err)
	}
	if err = s.ReleaseSessionActiveTurn(ctx, "A", "wrong-token"); err != nil {
		t.Fatal(err)
	}
	q, _ = s.GetTurn(ctx, "q3")
	if q.Status != "steering" {
		t.Fatal(q)
	}
	if err = s.ReleaseSessionActiveTurn(ctx, "A", "token"); err != nil {
		t.Fatal(err)
	}
	q, _ = s.GetTurn(ctx, "q3")
	if q.Status != "queued" || q.Phase != "steer_returned" || q.Prompt != "recover me" {
		t.Fatal(q)
	}
	if next, err := s.GetNextQueuedTurn(ctx, "A"); err != sql.ErrNoRows || next != nil {
		t.Fatal("returned Steer must not auto-send", next, err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "next", "runner", "next"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if msgs, err := s.DequeueSteeringForTurn(ctx, "A", "next"); err != sql.ErrNoRows || len(msgs) != 0 {
		t.Fatal(msgs, err)
	}
	var n int
	if err = s.DB().QueryRow(`select count(*) from steering_queue where source_queue_id='q1'`).Scan(&n); err != nil || n != 1 {
		t.Fatal(n, err)
	}
}

func TestQueueSteerFullQueueRetainsOriginal(t *testing.T) {
	ctx := context.Background()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.CreateSession(ctx, "A", "test", nil)
	s.CreateTurn(ctx, "active", "A", "run", nil)
	s.CreateTurnWithStatus(ctx, "q", "A", "queued", "queued", nil)
	s.ClaimSessionActiveTurn(ctx, "A", "active", "runner", "active")
	for i := 0; i < 10; i++ {
		if _, err = s.EnqueueSteering(ctx, "A", "active", "user", "message", nil, nil, ""); err != nil {
			t.Fatal(err)
		}
	}
	if err = s.SteerQueuedTurn(ctx, "A", "q", "active"); !errors.Is(err, ErrQueueConflict) {
		t.Fatal(err)
	}
	q, _ := s.GetTurn(ctx, "q")
	if q.Status != "queued" {
		t.Fatal(q)
	}
}
