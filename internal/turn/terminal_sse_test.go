package turn

import (
	"context"
	"testing"
	"time"
)

func TestTerminalSSEPreservesTurnIdentity(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	defer e.Close()
	ctx := context.Background()
	session, err := s.CreateSession(ctx, "sse-identity", "identity", map[string]any{"model": "test-model"})
	if err != nil {
		t.Fatal(err)
	}
	rec, err := s.CreateTurn(ctx, "turn-identity", session.ID, "identity", nil)
	if err != nil {
		t.Fatal(err)
	}
	events := e.Subscribe(session.ID)
	defer e.Unsubscribe(session.ID, events)
	r := &sessionRunner{engine: e, store: s}
	r.broadcastPost(session.ID, rec.ID, "message-identity", "answer", "sse-identity")
	r.finishTurn(s, rec.ID, session.ID, "sse-identity", "test-model", "completed", "", "")
	seen := map[string]bool{}
	deadline := time.After(2 * time.Second)
	for len(seen) < 3 {
		select {
		case event := <-events:
			kind, _ := event["type"].(string)
			if kind != "new_post" && kind != "agent_response" && kind != "agent_status" {
				continue
			}
			if event["turn_id"] != rec.ID || event["chat_jid"] != "gi:"+session.ID {
				t.Fatalf("wrong terminal identity: %#v", event)
			}
			if kind == "agent_status" && event["status"] != "idle" {
				t.Fatalf("wrong status: %#v", event)
			}
			seen[kind] = true
		case <-deadline:
			t.Fatalf("missing SSE kinds: %#v", seen)
		}
	}
}
