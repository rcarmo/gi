package connectivity

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
)

type subscription struct {
	id      string
	pattern string
	ch      chan EventEnvelope
}

// EventBus is an in-memory pub/sub bus used to decouple transports from the
// agent loop. It is deliberately bounded: slow subscribers drop events rather
// than blocking route delivery.
type EventBus struct {
	mu   sync.RWMutex
	next uint64
	subs map[string]subscription
}

func NewEventBus() *EventBus { return &EventBus{subs: make(map[string]subscription)} }

func (b *EventBus) Emit(_ context.Context, event EventEnvelope) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, sub := range b.subs {
		if !topicMatches(sub.pattern, event.Topic) {
			continue
		}
		select {
		case sub.ch <- event:
		default:
		}
	}
	return nil
}

func (b *EventBus) Subscribe(ctx context.Context, pattern string, buffer int) (<-chan EventEnvelope, func()) {
	if strings.TrimSpace(pattern) == "" {
		pattern = "*"
	}
	if buffer <= 0 {
		buffer = 64
	}
	b.mu.Lock()
	b.next++
	id := storeID("sub", b.next)
	ch := make(chan EventEnvelope, buffer)
	b.subs[id] = subscription{id: id, pattern: pattern, ch: ch}
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		if sub, ok := b.subs[id]; ok {
			delete(b.subs, id)
			close(sub.ch)
		}
		b.mu.Unlock()
	}
	go func() {
		<-ctx.Done()
		unsubscribe()
	}()
	return ch, unsubscribe
}

func topicMatches(pattern, topic string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" || pattern == "*" || pattern == topic {
		return true
	}
	if strings.HasSuffix(pattern, ".*") {
		return strings.HasPrefix(topic, strings.TrimSuffix(pattern, "*"))
	}
	matched, err := filepath.Match(pattern, topic)
	return err == nil && matched
}
