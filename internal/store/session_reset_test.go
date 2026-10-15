package store

import (
	"context"
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
