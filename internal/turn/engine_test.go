package turn

import (
	"context"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

func TestSubmitPromptQueuesSecondTurn(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	_, err = s.CreateSession(ctx, "session_1", "Test", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)

	first, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_1", Prompt: "one", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit first: %v", err)
	}
	second, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_1", Prompt: "two", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit second: %v", err)
	}
	if first.Queued {
		t.Fatalf("first should not be queued: %#v", first)
	}
	if !second.Queued || second.Status != "queued" {
		t.Fatalf("second should be queued: %#v", second)
	}

	time.Sleep(2500 * time.Millisecond)
	turns, err := s.ListTurns(ctx, "session_1")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected 2 turns, got %d", len(turns))
	}
	if turns[0].Status != "completed" || turns[1].Status != "completed" {
		t.Fatalf("unexpected turn statuses: %#v", turns)
	}
}

func TestCancelQueuedTurn(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	_, _ = s.CreateSession(ctx, "session_1", "Test", map[string]any{"model": "bootstrap"})
	engine := New(s)
	_, _ = engine.SubmitPrompt(ctx, RunInput{SessionID: "session_1", Prompt: "one", Model: "bootstrap"})
	second, _ := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_1", Prompt: "two", Model: "bootstrap"})
	if err := engine.CancelTurn(ctx, "session_1", second.TurnID); err != nil {
		t.Fatalf("cancel queued: %v", err)
	}
	turnRec, err := s.GetTurn(ctx, second.TurnID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if turnRec.Status != "cancelled" {
		t.Fatalf("expected cancelled, got %s", turnRec.Status)
	}
}
