package topics

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Envelope is the normalized internal runtime event shape used for extension,
// TUI, SSE, and connectivity fan-out.
type Envelope struct {
	Topic     string         `json:"topic"`
	SessionID string         `json:"session_id,omitempty"`
	AgentID   string         `json:"agent_id,omitempty"`
	Source    string         `json:"source,omitempty"`
	Type      string         `json:"type,omitempty"`
	Sequence  uint64         `json:"sequence,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// SubscribeOptions narrows subscriptions to session/agent scoped events and
// controls the buffer size for the delivered stream.
type SubscribeOptions struct {
	Buffer    int
	SessionID string
	AgentID   string
}

type subscriber struct {
	id      uint64
	pattern string
	opts    SubscribeOptions
	ch      chan Envelope
}

// Bus is an in-memory bounded topic bus. Slow subscribers never block the
// publisher; the oldest buffered event is dropped to keep the latest state.
type Bus struct {
	mu       sync.RWMutex
	next     uint64
	sequence uint64
	subs     map[uint64]subscriber
}

func NewBus() *Bus { return &Bus{subs: make(map[uint64]subscriber)} }

func (b *Bus) Publish(env Envelope) {
	if strings.TrimSpace(env.Topic) == "" {
		return
	}
	if env.Timestamp.IsZero() {
		env.Timestamp = time.Now().UTC()
	}
	b.mu.Lock()
	if env.Sequence == 0 {
		b.sequence++
		env.Sequence = b.sequence
	} else if env.Sequence > b.sequence {
		b.sequence = env.Sequence
	}
	subs := make([]subscriber, 0, len(b.subs))
	for _, sub := range b.subs {
		subs = append(subs, sub)
	}
	b.mu.Unlock()
	for _, sub := range subs {
		if !topicMatches(sub.pattern, env.Topic) {
			continue
		}
		if sub.opts.SessionID != "" && sub.opts.SessionID != env.SessionID {
			continue
		}
		if sub.opts.AgentID != "" && sub.opts.AgentID != env.AgentID {
			continue
		}
		deliverDropOldest(sub.ch, env)
	}
}

func (b *Bus) Subscribe(ctx context.Context, pattern string, opts SubscribeOptions) (<-chan Envelope, func()) {
	if strings.TrimSpace(pattern) == "" {
		pattern = "*"
	}
	if opts.Buffer <= 0 {
		opts.Buffer = 64
	}
	b.mu.Lock()
	b.next++
	id := b.next
	ch := make(chan Envelope, opts.Buffer)
	b.subs[id] = subscriber{id: id, pattern: pattern, opts: opts, ch: ch}
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

func deliverDropOldest(ch chan Envelope, env Envelope) {
	select {
	case ch <- env:
		return
	default:
	}
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- env:
	default:
	}
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
