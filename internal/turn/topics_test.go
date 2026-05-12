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

func TestSteeringBroadcastEventsMapToSessionSteeringTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "session.steering", topics.SubscribeOptions{Buffer: 16})
	defer unsub()

	types := []string{"steering_enqueued", "steering_dequeued", "steering_continue_staged", "steering_continued", "steering_injected"}
	for _, typ := range types {
		engine.broadcast("session_steering_topic", map[string]any{"type": typ, "turn_id": "turn_1"})
		select {
		case env := <-ch:
			if env.Topic != "session.steering" {
				t.Fatalf("unexpected topic for %s: %#v", typ, env)
			}
			if gotType, _ := env.Payload["type"].(string); gotType != typ {
				t.Fatalf("unexpected payload for %s: %#v", typ, env.Payload)
			}
		case <-time.After(time.Second):
			t.Fatalf("expected session.steering event for %s", typ)
		}
	}
}

func TestEmitHookPublishesRuntimeHookTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "runtime.hook", topics.SubscribeOptions{Buffer: 8})
	defer unsub()
	if _, err := engine.RegisterHook(HookToolCall, "test", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		return HookResponse{Action: "modify"}, nil
	}); err != nil {
		t.Fatalf("register hook: %v", err)
	}
	if _, err := engine.emitHook(context.Background(), HookRequest{Name: HookToolCall, SessionID: "session_hook_topic", TurnID: "turn_hook_topic", AgentID: "agent"}); err != nil {
		t.Fatalf("emit hook: %v", err)
	}
	select {
	case env := <-ch:
		if env.Topic != "runtime.hook" {
			t.Fatalf("unexpected runtime hook topic: %#v", env)
		}
		if env.Payload["hook"] != HookToolCall || env.Payload["action"] != "modify" || env.Payload["source"] != "test" {
			t.Fatalf("unexpected runtime hook payload: %#v", env.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("expected runtime.hook topic event")
	}
}

func TestEmitHookPublishesRuntimeHookErrorTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "runtime.hook", topics.SubscribeOptions{Buffer: 8})
	defer unsub()
	if _, err := engine.RegisterHook(HookToolCall, "broken", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		return HookResponse{}, context.DeadlineExceeded
	}); err != nil {
		t.Fatalf("register hook: %v", err)
	}
	engine.runtimeCfg.Hooks.OnError = "continue"
	if _, err := engine.emitHook(context.Background(), HookRequest{Name: HookToolCall, SessionID: "session_hook_topic", TurnID: "turn_hook_topic", AgentID: "agent"}); err != nil {
		t.Fatalf("emit hook with continue policy: %v", err)
	}
	select {
	case env := <-ch:
		if env.Topic != "runtime.hook" {
			t.Fatalf("unexpected runtime hook topic: %#v", env)
		}
		if env.Payload["hook"] != HookToolCall || env.Payload["source"] != "broken" || env.Payload["error"] == "" {
			t.Fatalf("unexpected runtime hook error payload: %#v", env.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("expected runtime.hook error topic event")
	}
}

func TestSubTurnBroadcastEventsMapToTurnSubTurnTopic(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	engine := New(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := engine.Topics().Subscribe(ctx, "turn.subturn", topics.SubscribeOptions{Buffer: 8})
	defer unsub()

	for _, typ := range []string{"subturn_created", "subturn_status", "subturn_result_ready", "subturn_result_delivered", "subturn_orphaned", "subturn_cancel_requested"} {
		engine.broadcast("session_subturn_topic", map[string]any{"type": typ, "parent_turn_id": "turn_parent", "child_turn_id": "turn_child"})
		select {
		case env := <-ch:
			if env.Topic != "turn.subturn" {
				t.Fatalf("unexpected subturn topic for %s: %#v", typ, env)
			}
			if gotType, _ := env.Payload["type"].(string); gotType != typ {
				t.Fatalf("unexpected payload type for %s: %#v", typ, env.Payload)
			}
		case <-time.After(time.Second):
			t.Fatalf("expected turn.subturn event for %s", typ)
		}
	}
}
