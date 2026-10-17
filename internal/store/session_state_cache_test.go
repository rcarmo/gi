package store

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestSessionStateCacheOwnershipFreshnessAndFailure(t *testing.T) {
	var c sessionStateDecodeCache
	raw := `{"nested":{"items":[{"value":"original"}]},"number":3}`
	a, err := c.decode("a", raw)
	if err != nil {
		t.Fatal(err)
	}
	a["nested"].(map[string]any)["items"].([]any)[0].(map[string]any)["value"] = "mutated"
	b, err := c.decode("a", raw)
	if err != nil {
		t.Fatal(err)
	}
	if b["nested"].(map[string]any)["items"].([]any)[0].(map[string]any)["value"] != "original" {
		t.Fatal("shared mutable state")
	}
	b, err = c.decode("a", `{"number":4}`)
	if err != nil || b["number"] != float64(4) {
		t.Fatal(b, err)
	}
	for _, invalid := range []string{`{`, `[]`, `42`} {
		if _, err = c.decode("a", invalid); err == nil {
			t.Fatalf("accepted invalid map %q", invalid)
		}
	}
	for _, empty := range []string{"", "null"} {
		b, err = c.decode("a", empty)
		if err != nil || len(b) != 0 {
			t.Fatal(b, err)
		}
	}
}

func TestSessionStateCacheBoundedAndConcurrent(t *testing.T) {
	var c sessionStateDecodeCache
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < 400; i++ {
				if _, err := c.decode(fmt.Sprintf("%d-%d", worker, i), `{"value":"data"}`); err != nil {
					t.Error(err)
				}
			}
		}(worker)
	}
	wg.Wait()
	if len(c.entries) > sessionStateCacheEntries || c.bytes > sessionStateCacheBytes {
		t.Fatal("unbounded cache", len(c.entries), c.bytes)
	}
	huge := `{"value":"` + strings.Repeat("x", sessionStateCacheBytes/16) + `"}`
	if _, err := c.decode("huge", huge); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.entries["huge"]; ok {
		t.Fatal("oversized entry retained")
	}
}

func TestListSessionsCacheObservesDatabaseChanges(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.CreateSession(t.Context(), "a", "session", nil); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"first", "second"} {
		if err = s.TouchSessionState(t.Context(), "a", map[string]any{"value": value}); err != nil {
			t.Fatal(err)
		}
		list, err := s.ListSessions(t.Context())
		if err != nil || len(list) != 1 {
			t.Fatal(list, err)
		}
		if list[0].State["value"] != value {
			t.Fatalf("stale state: %+v", list[0].State)
		}
		list[0].State["value"] = "caller mutation"
	}
	if _, err = s.db.ExecContext(t.Context(), `update sessions set state_json = '{' where id='a'`); err == nil {
		t.Fatal("database accepted malformed state")
	}
	list, err := s.ListSessions(t.Context())
	if err != nil || list[0].State["value"] != "second" {
		t.Fatal("rejected update changed state", list, err)
	}
}
