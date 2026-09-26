package inference

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestSessionThinkingSelectUsesEffectiveDefaultAndLegacyModel(t *testing.T) {
	Init()
	goai.RegisterModel(&goai.Model{ID: "legacy", Provider: "thinking-select", Api: goai.ApiOpenAICompletions, ContextWindow: 32000, Reasoning: true})
	model := "thinking-select/legacy"
	options := []ModelOption{{ID: "legacy", Provider: "thinking-select", Label: model, Reasoning: true, Authenticated: true}}
	ctx := context.Background()
	s, err := store.Open(filepath.Join(t.TempDir(), "thinking.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, state := range []map[string]any{nil, {"model": "legacy", "provider": "thinking-select", "thinking_level": "arbitrary"}} {
		id := store.NowID("session")
		session, err := s.CreateSession(ctx, id, id, state)
		if err != nil {
			t.Fatal(err)
		}
		if CapturedSessionThinking(session, model) != "" {
			t.Fatal("unvalidated legacy effective")
		}
		token := store.SessionThinkingToken(id, session.State)
		got, err := SelectThinking(ctx, s, id, token, model, "high", options, SessionModelChoice{Model: "legacy", Provider: "thinking-select"})
		if err != nil {
			t.Fatal(err)
		}
		if CapturedSessionThinking(got, model) != "high" {
			t.Fatal(got.State)
		}
		if _, err = SelectThinking(ctx, s, id, token, model, "low", options, SessionModelChoice{Model: "legacy", Provider: "thinking-select"}); !errors.Is(err, store.ErrContextChanged) {
			t.Fatal(err)
		}
	}
}
