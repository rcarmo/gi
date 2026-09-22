package inference

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/store"
)

func TestResolveSessionModelRejectsUnavailableAmbiguousAndUnknown(t *testing.T) {
	options := []ModelOption{{ID: "bootstrap", Label: "test/bootstrap", Provider: "test", Enabled: true}, {ID: "absent", Label: "test/absent", Provider: "test", Enabled: true}}
	for _, label := range []string{"", "missing", "test/absent", strings.Repeat("x", 257)} {
		if _, err := ResolveSessionModel(options, label); err == nil {
			t.Fatalf("accepted %q", label)
		}
	}
	for _, label := range []string{"bootstrap", " test/bootstrap "} {
		m, err := ResolveSessionModel(options, label)
		if err != nil || m.ID != "bootstrap" {
			t.Fatalf("valid %q: %+v %v", label, m, err)
		}
	}
	options = append(options, ModelOption{ID: "bootstrap", Provider: "other", Label: "other/bootstrap"})
	if _, err := ResolveSessionModel(options, "bootstrap"); err == nil {
		t.Fatal("ambiguous ID accepted")
	}
	if _, err := ResolveSessionModel([]ModelOption{{ID: "bootstrap", Label: "test/bootstrap", Provider: "test"}}, "bootstrap"); err == nil {
		t.Fatal("disabled shell model accepted")
	}
}

func TestSelectSessionModelPersistsOnlyTargetAndPrefersUserSelection(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "models.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"A", "B"} {
		if _, err := s.CreateSession(ctx, id, id, map[string]any{"model": "test-model", "provider": "test", "thinking_level": "high"}); err != nil {
			t.Fatal(err)
		}
	}
	options := []ModelOption{{ID: "bootstrap", Label: "test/bootstrap", Provider: "test", Enabled: true}}
	if _, err := SelectSessionModel(ctx, s, "missing", options, "bootstrap"); err == nil {
		t.Fatal("missing session success")
	}
	a, err := SelectSessionModel(ctx, s, "A", options, "test/bootstrap")
	if err != nil {
		t.Fatal(err)
	}
	if a.State["selected_model"] != "bootstrap" || a.State["thinking_level"] != "" {
		t.Fatalf("bad state %+v", a.State)
	}
	if err := s.TouchSessionState(ctx, "A", map[string]any{"model": "test-model", "provider": "previous"}); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s, err = store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	a, _ = s.GetSession(ctx, "A")
	b, _ := s.GetSession(ctx, "B")
	choice := SessionModel(a.State, SessionModelChoice{Model: "fallback", Provider: "fallback", Thinking: "low"})
	if choice.Label() != "test/bootstrap" || choice.Thinking != "" {
		t.Fatalf("runtime replaced selected: %+v", choice)
	}
	if b.State["model"] != "test-model" || b.State["thinking_level"] != "high" {
		t.Fatal("other session mutated")
	}
	for _, id := range []string{"A", "B"} {
		turns, _ := s.ListTurns(ctx, id)
		if len(turns) != 0 {
			t.Fatal("selection created turn")
		}
	}
}
