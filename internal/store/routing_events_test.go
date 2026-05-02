package store

import (
	"context"
	"testing"
)

func TestStoreRouteEventsPersistAndList(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	source, err := s.CreateSession(ctx, "session_source", "@agent", map[string]any{"status": "idle", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	target, err := s.CreateSession(ctx, "session_target", "@agent1", map[string]any{"status": "idle", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create target: %v", err)
	}

	_, err = s.CreateTurn(ctx, "turn_1", source.ID, "seed prompt", map[string]any{"intent": "test"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	evtID, err := s.RecordRouteEvent(ctx, RouteEvent{
		TurnID:         "turn_1",
		SourceSession:  source.ID,
		TargetSession:  target.ID,
		SourceAgentID:  "agent",
		TargetAgentID:  "agent1",
		Mode:           "prompt",
		MatchedBy:      "default",
		RoutingPolicy:  "default",
		RequestedAgent: "agent1",
		Metadata:       map[string]any{"reason": "test"},
	})
	if err != nil {
		t.Fatalf("record route event: %v", err)
	}
	if evtID <= 0 {
		t.Fatalf("expected positive route event id")
	}

	events, err := s.ListRouteEvents(ctx, source.ID)
	if err != nil {
		t.Fatalf("list route events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 route event, got %d", len(events))
	}
	if events[0].SourceSession != source.ID || events[0].TargetSession != target.ID {
		t.Fatalf("unexpected route event: %#v", events[0])
	}

	sourceEvent, err := s.GetRouteEvent(ctx, evtID)
	if err != nil {
		t.Fatalf("get route event: %v", err)
	}
	if sourceEvent.TargetAgentID != "agent1" {
		t.Fatalf("unexpected route event target agent: %#v", sourceEvent)
	}
}
