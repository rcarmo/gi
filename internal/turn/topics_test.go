package turn

import (
	"context"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/topics"
)

func TestBroadcastPublishesNormalizedTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "turn.*", topics.SubscribeOptions{Buffer: 4})
	defer unsub()

	engine.broadcast("session_1", map[string]any{"type": "agent_status", "agent_id": "agent1", "status": "running"})

	select {
	case env := <-ch:
		if env.Topic != "turn.status" {
			t.Fatalf("unexpected topic: %#v", env)
		}
		if env.SessionID != "session_1" || env.AgentID != "agent1" {
			t.Fatalf("unexpected scoped envelope: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected topic event")
	}
}

func TestConnectivityBusBridgesIntoTopics(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "connectivity.*", topics.SubscribeOptions{Buffer: 4})
	defer unsub()

	if err := engine.Connectivity().Emit(context.Background(), "route.http.demo", map[string]any{"ok": true, "session_id": "s1"}); err != nil {
		t.Fatalf("emit connectivity event: %v", err)
	}

	select {
	case env := <-ch:
		if env.Topic != "connectivity.route.http.demo" {
			t.Fatalf("unexpected connectivity topic: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected bridged connectivity topic")
	}
}

func TestRecordExtensionPublishesExtensionTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "extension.*", topics.SubscribeOptions{Buffer: 4})
	defer unsub()

	engine.recordExtension(ExtensionInfo{Engine: "joker", Path: ".gi/extensions/demo.joke", Status: "loaded"})

	select {
	case env := <-ch:
		if env.Topic != "extension.loaded" {
			t.Fatalf("unexpected extension topic: %#v", env)
		}
		if got := env.Payload["path"]; got != ".gi/extensions/demo.joke" {
			t.Fatalf("unexpected payload: %#v", env.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("expected extension lifecycle topic")
	}
}
