package store

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
)

func TestQueuedTurnOrderPersistsWithoutRewritingHistory(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "queue.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { s.Close() }()
	for _, id := range []string{"A", "B"} {
		if _, err := s.CreateSession(ctx, id, id, nil); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"one", "two", "three"} {
		if _, err := s.CreateTurnWithStatus(ctx, id, "A", "queued", id, nil); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.CreateTurnWithStatus(ctx, "foreign", "B", "queued", "foreign", nil); err != nil {
		t.Fatal(err)
	}
	original, _ := s.GetTurn(ctx, "one")
	if err := s.ReorderQueuedTurns(ctx, "A", []string{"one", "two", "three"}, []string{"three", "one", "two"}); err != nil {
		t.Fatal(err)
	}
	for _, order := range [][]string{{"one", "one", "two"}, {"foreign", "one", "two"}, {"one"}} {
		if err := s.ReorderQueuedTurns(ctx, "A", []string{"three", "one", "two"}, order); !errors.Is(err, ErrQueueConflict) {
			t.Fatalf("invalid permutation accepted: %v", err)
		}
	}
	if err := s.ReorderQueuedTurns(ctx, "A", []string{"one", "two", "three"}, []string{"three", "one", "two"}); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("stale snapshot accepted")
	}
	s.Close()
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	next, err := s.GetNextQueuedTurn(ctx, "A")
	if err != nil || next.ID != "three" {
		t.Fatalf("order not persisted: %+v %v", next, err)
	}
	after, _ := s.GetTurn(ctx, "one")
	if after.CreatedAt != original.CreatedAt || after.UpdatedAt != original.UpdatedAt {
		t.Fatal("reorder rewrote event chronology")
	}
	if _, err := s.CreateTurnWithStatus(ctx, "four", "A", "queued", "four", nil); err != nil {
		t.Fatal(err)
	}
	queue, err := s.ListQueuedTurns(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []string{"three", "one", "two", "four"} {
		if queue[i].ID != want {
			t.Fatalf("order: %+v", queue)
		}
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "A", "three", "test", "claim"); err != nil {
		t.Fatal(err)
	}
	if err := s.ReorderQueuedTurns(ctx, "A", []string{"three", "one", "two", "four"}, []string{"one", "three", "two", "four"}); !errors.Is(err, ErrQueueConflict) {
		t.Fatalf("claimed reorder accepted: %v", err)
	}
	if err := s.CancelQueuedTurn(ctx, "A", "three"); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("claimed turn cancelled")
	}
	if err := s.CancelQueuedTurn(ctx, "B", "two"); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("foreign turn cancelled")
	}
	if err := s.CancelQueuedTurn(ctx, "A", "two"); err != nil {
		t.Fatal(err)
	}
}

func TestQueuedTurnReorderConcurrentSnapshotsSerialize(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "concurrent.db")
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
	if _, err = a.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"one", "two", "three"} {
		if _, err = a.CreateTurnWithStatus(ctx, id, "A", "queued", id, nil); err != nil {
			t.Fatal(err)
		}
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i, s := range []*Store{a, b} {
		wg.Add(1)
		go func(i int, s *Store) {
			defer wg.Done()
			<-start
			order := []string{"three", "one", "two"}
			if i == 1 {
				order = []string{"two", "three", "one"}
			}
			results <- s.ReorderQueuedTurns(ctx, "A", []string{"one", "two", "three"}, order)
		}(i, s)
	}
	close(start)
	wg.Wait()
	close(results)
	wins, conflicts := 0, 0
	for err := range results {
		if err == nil {
			wins++
		} else if errors.Is(err, ErrQueueConflict) {
			conflicts++
		} else {
			t.Fatal("unexpected concurrency error", err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatal(wins, conflicts)
	}
}
