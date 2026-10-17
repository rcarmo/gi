package store

import "sync"

// Cache only decoding, never database reads or response freshness. Each listing
// compares the state read in its transaction and returns an owned deep copy.
const sessionStateCacheBytes = 32 << 20
const sessionStateCacheEntries = 256

type cachedSessionState struct {
	raw   string
	value map[string]any
}
type sessionStateDecodeCache struct {
	mu      sync.Mutex
	entries map[string]cachedSessionState
	bytes   int
}

func (c *sessionStateDecodeCache) decode(id, raw string) (map[string]any, error) {
	c.mu.Lock()
	entry, ok := c.entries[id]
	c.mu.Unlock()
	if ok && entry.raw == raw {
		return cloneSessionState(entry.value).(map[string]any), nil
	}
	value, err := unmarshalJSONMap(raw)
	if err != nil {
		return nil, err
	}
	// The raw text and decoded strings are both retained. Conservatively bound
	// each entry using a multiplier; tiny structural JSON has much higher overhead.
	cost := len(raw)*16 + 256
	if cost <= sessionStateCacheBytes {
		c.mu.Lock()
		if old, exists := c.entries[id]; exists {
			c.bytes -= len(old.raw)*16 + 256
			delete(c.entries, id)
		}
		if c.bytes+cost > sessionStateCacheBytes || len(c.entries) >= sessionStateCacheEntries {
			c.entries = nil
			c.bytes = 0
		}
		if c.entries == nil {
			c.entries = make(map[string]cachedSessionState)
		}
		c.entries[id] = cachedSessionState{raw: raw, value: value}
		c.bytes += cost
		c.mu.Unlock()
	}
	return cloneSessionState(value).(map[string]any), nil
}

func cloneSessionState(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, item := range v {
			out[k] = cloneSessionState(item)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = cloneSessionState(item)
		}
		return out
	default:
		return value // JSON scalar values, including strings, are immutable.
	}
}
