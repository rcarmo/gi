package store

import (
	"context"
	"testing"
)

func TestStoreSessionTurnMessageFlow(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	session, err := s.CreateSession(ctx, "session_1", "Test", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if session.State["model"] != "bootstrap" {
		t.Fatalf("unexpected session state: %#v", session.State)
	}

	turn, err := s.CreateTurn(ctx, "turn_1", session.ID, "hello", map[string]any{"intent": "prompt"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if turn.Status != "running" {
		t.Fatalf("unexpected turn status: %s", turn.Status)
	}

	if err := s.AppendTurnEvent(ctx, turn.ID, session.ID, "turn.started", map[string]any{"phase": "turn", "checkpoint": true}); err != nil {
		t.Fatalf("append turn event: %v", err)
	}
	if err := s.AddMessage(ctx, "msg_1", session.ID, "user", "hello", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("add message: %v", err)
	}

	events, err := s.ListTurnEvents(ctx, turn.ID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	if len(events) != 1 || events[0].Payload["phase"] != "turn" {
		t.Fatalf("unexpected events: %#v", events)
	}

	msgs, err := s.ListMessages(ctx, session.ID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Content != "hello" {
		t.Fatalf("unexpected messages: %#v", msgs)
	}
}
