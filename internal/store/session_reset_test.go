package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestResetContextRetainsDisplayHistoryAndReopens(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "reset.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateSession(ctx, "s", "session", nil); err != nil {
		t.Fatal(err)
	}
	if err := s.AddMessage(ctx, "old", "s", "user", "old history", nil); err != nil {
		t.Fatal(err)
	}
	if err := s.ResetContext(ctx, "s", "reset-marker", "fresh"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddMessage(ctx, "later", "s", "assistant", "answer", nil); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	snap, err := s.ContextSnapshot(ctx, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Messages) != 2 || snap.Messages[0].Content != "fresh" || snap.Messages[1].Content != "answer" || snap.Summary != "" {
		t.Fatalf("reset context: %+v", snap)
	}
	page, err := s.PageConversationMessages(ctx, "s", "", "", 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Messages) != 3 {
		t.Fatalf("visible history lost or marker exposed: %+v", page.Messages)
	}
	if _, err := s.CloneSessionBefore(ctx, "s", "fork", "fork", "", ""); err != nil {
		t.Fatal(err)
	}
	fork, err := s.ContextSnapshot(ctx, "fork")
	if err != nil {
		t.Fatal(err)
	}
	if len(fork.Messages) != 2 || fork.Messages[0].Content != "fresh" {
		t.Fatalf("fork reset head lost: %+v", fork)
	}
}

func TestResetContextRollback(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "reset-fail.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "s", "s", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`create trigger reject_reset before update on sessions begin select raise(abort,'reset failed'); end`); err != nil {
		t.Fatal(err)
	}
	if err := s.ResetContext(ctx, "s", "marker", "fresh"); err == nil {
		t.Fatal("expected failure")
	}
	messages, err := s.ListMessages(ctx, "s")
	if err != nil || len(messages) != 0 {
		t.Fatalf("partial reset: %+v %v", messages, err)
	}
}

func TestResetRejectsStaleCompactionBoundary(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "reset-compaction.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.CreateSession(ctx, "s", "s", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, "t", "s", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "s", "t", "test", "claim"); err != nil || !ok {
		t.Fatalf("claim: %t %v", ok, err)
	}
	if err := s.AddMessage(ctx, "old", "s", "user", "pre-reset secret", nil); err != nil {
		t.Fatal(err)
	}
	snapshot, err := s.ContextSnapshot(ctx, "s")
	if err != nil {
		t.Fatal(err)
	}
	boundary, err := PrepareContextBoundary(snapshot, 1)
	if err != nil {
		t.Fatal(err)
	}
	seq, err := s.BeginCompaction(ctx, "s", "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResetContext(ctx, "s", "reset-marker", "fresh context"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.FinishCompactionWithBoundary(ctx, "s", "t", seq, "completed", "obsolete summary", nil, boundary); !errors.Is(err, ErrContextChanged) {
		t.Fatalf("stale compaction accepted after reset: %v", err)
	}
	after, err := s.ContextSnapshot(ctx, "s")
	if err != nil {
		t.Fatal(err)
	}
	if after.Version != 1 || after.Summary != "" || len(after.Messages) != 1 || after.Messages[0].Content != "fresh context" {
		t.Fatalf("old context resurrected: %+v", after)
	}
}

func TestResetHeadUsesInsertionBoundaryForBackdatedMessages(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "reset-backdated.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.CreateSession(ctx, "s", "s", nil); err != nil {
		t.Fatal(err)
	}
	if err := s.AddMessage(ctx, "old", "s", "user", "old", nil); err != nil {
		t.Fatal(err)
	}
	if err := s.ResetContext(ctx, "s", "marker", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.AddImportedMessages(ctx, "s", []ImportedMessage{{ID: "backdated", Role: "user", Content: "new import", CreatedAt: "2000-01-01T00:00:00Z"}}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := s.ContextSnapshot(ctx, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Messages) != 1 || snapshot.Messages[0].ID != "backdated" {
		t.Fatalf("backdated new row excluded: %+v", snapshot)
	}
}
