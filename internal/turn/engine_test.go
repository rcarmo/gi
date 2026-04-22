package turn

import (
	"context"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/store"
)

func TestRunPromptPersistsTurnAndMessages(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	_, err = s.CreateSession(ctx, "session_1", "Test", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	engine := New(s)
	summary, err := engine.RunPrompt(ctx, RunInput{SessionID: "session_1", Prompt: "hello world", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("run prompt: %v", err)
	}
	if summary.Status != "completed" {
		t.Fatalf("unexpected status: %#v", summary)
	}
	if !strings.Contains(summary.Assistant, "Gi received: hello world") {
		t.Fatalf("unexpected assistant output: %q", summary.Assistant)
	}
	if len(summary.Events) < 3 {
		t.Fatalf("expected events, got %#v", summary.Events)
	}

	msgs, err := s.ListMessages(ctx, "session_1")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
}
