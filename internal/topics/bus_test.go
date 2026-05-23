package topics

import (
	"context"
	"testing"
	"time"
)

func TestBusSubscribeWildcardAndScopedFilters(t *testing.T) {
	bus := NewBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	allCh, unsubAll := bus.Subscribe(ctx, "turn.*", SubscribeOptions{Buffer: 4})
	defer unsubAll()
	sessionCh, unsubSession := bus.Subscribe(ctx, "turn.*", SubscribeOptions{Buffer: 4, SessionID: "s1"})
	defer unsubSession()

	bus.Publish(Envelope{Topic: "turn.status", SessionID: "s1", AgentID: "agent1", Type: "status"})
	bus.Publish(Envelope{Topic: "turn.status", SessionID: "s2", AgentID: "agent2", Type: "status"})

	select {
	case env := <-allCh:
		if env.Topic != "turn.status" {
			t.Fatalf("unexpected topic: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected wildcard event")
	}
	select {
	case env := <-sessionCh:
		if env.SessionID != "s1" {
			t.Fatalf("unexpected scoped envelope: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected scoped event")
	}
	select {
	case env := <-sessionCh:
		t.Fatalf("unexpected second scoped event: %#v", env)
	default:
	}
}

func TestBusDropsOldestOnPressure(t *testing.T) {
	bus := NewBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := bus.Subscribe(ctx, "turn.*", SubscribeOptions{Buffer: 1})
	defer unsub()

	bus.Publish(Envelope{Topic: "turn.status", Payload: map[string]any{"seq": 1}})
	bus.Publish(Envelope{Topic: "turn.status", Payload: map[string]any{"seq": 2}})

	select {
	case env := <-ch:
		if got := env.Payload["seq"]; got != 2 {
			t.Fatalf("expected latest event after drop-oldest, got %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected buffered event")
	}
}

func TestBusAssignsMonotonicSequences(t *testing.T) {
	bus := NewBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := bus.Subscribe(ctx, "runtime.*", SubscribeOptions{Buffer: 4})
	defer unsub()

	bus.Publish(Envelope{Topic: "runtime.turn"})
	bus.Publish(Envelope{Topic: "runtime.session"})
	first := <-ch
	second := <-ch
	if first.Sequence == 0 || second.Sequence == 0 || second.Sequence <= first.Sequence {
		t.Fatalf("expected monotonic sequences, got first=%#v second=%#v", first, second)
	}
}

func TestBusPreservesExplicitSequenceAndAdvancesNextAutoSequence(t *testing.T) {
	bus := NewBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := bus.Subscribe(ctx, "runtime.*", SubscribeOptions{Buffer: 4})
	defer unsub()

	bus.Publish(Envelope{Topic: "runtime.turn", Sequence: 41})
	bus.Publish(Envelope{Topic: "runtime.session"})
	first := <-ch
	second := <-ch
	if first.Sequence != 41 || second.Sequence != 42 {
		t.Fatalf("expected explicit sequence preserved and auto sequence advanced, got first=%#v second=%#v", first, second)
	}
}

func TestBusAgentFilter(t *testing.T) {
	bus := NewBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := bus.Subscribe(ctx, "extension.*", SubscribeOptions{Buffer: 2, AgentID: "agent1"})
	defer unsub()

	bus.Publish(Envelope{Topic: "extension.loaded", AgentID: "agent2"})
	bus.Publish(Envelope{Topic: "extension.loaded", AgentID: "agent1"})

	select {
	case env := <-ch:
		if env.AgentID != "agent1" {
			t.Fatalf("unexpected agent envelope: %#v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected agent-scoped event")
	}
}

func TestBusConcurrentUnsubscribeAndPublishDoesNotPanic(t *testing.T) {
	bus := NewBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, unsub := bus.Subscribe(ctx, "turn.*", SubscribeOptions{Buffer: 8})

	panicCh := make(chan any, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				panicCh <- r
			}
		}()
		for i := 0; i < 4000; i++ {
			bus.Publish(Envelope{Topic: "turn.status", SessionID: "s1", Payload: map[string]any{"seq": i}})
		}
	}()
	time.Sleep(2 * time.Millisecond)
	unsub()
	<-done
	select {
	case p := <-panicCh:
		t.Fatalf("publish panicked during concurrent unsubscribe: %v", p)
	default:
	}
}
