package topics

import (
	"context"
	"sync"
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

func TestBusSubscribeWithCanceledContextReturnsClosedChannel(t *testing.T) {
	bus := NewBus()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ch, unsub := bus.Subscribe(ctx, "runtime.*", SubscribeOptions{Buffer: 1})
	defer unsub()
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected canceled subscription channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("expected canceled subscription channel to close immediately")
	}
	bus.Publish(Envelope{Topic: "runtime.turn"})
	if got := bus.LastSequence(); got != 1 {
		t.Fatalf("expected publish after canceled subscribe to still advance sequence, got %d", got)
	}
}

func TestBusConcurrentPublishSequencesAreUniqueAndMonotonic(t *testing.T) {
	bus := NewBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, unsub := bus.Subscribe(ctx, "runtime.*", SubscribeOptions{Buffer: 512})
	defer unsub()

	const publishers = 8
	const perPublisher = 32
	var wg sync.WaitGroup
	for p := 0; p < publishers; p++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < perPublisher; i++ {
				bus.Publish(Envelope{Topic: "runtime.turn", Payload: map[string]any{"publisher": id, "seq": i}})
			}
		}(p)
	}
	wg.Wait()

	seen := map[uint64]bool{}
	last := uint64(0)
	for i := 0; i < publishers*perPublisher; i++ {
		select {
		case env := <-ch:
			if env.Sequence == 0 {
				t.Fatalf("expected non-zero sequence: %#v", env)
			}
			if seen[env.Sequence] {
				t.Fatalf("duplicate sequence %d", env.Sequence)
			}
			seen[env.Sequence] = true
			if env.Sequence <= last {
				t.Fatalf("expected subscriber delivery to preserve publish sequence order, got last=%d current=%d", last, env.Sequence)
			}
			last = env.Sequence
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out after %d events", i)
		}
	}
	if got := bus.LastSequence(); got != publishers*perPublisher {
		t.Fatalf("expected last sequence %d, got %d", publishers*perPublisher, got)
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
