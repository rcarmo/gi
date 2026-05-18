package turn

import (
	"context"
	"github.com/rcarmo/gi/internal/compaction"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestPrepareCompactionKeepsRecentMessages(t *testing.T) {
	messages := []goai.Message{
		goai.UserMessage(strings.Repeat("a", 400)),
		goai.UserMessage(strings.Repeat("b", 400)),
		goai.UserMessage(strings.Repeat("c", 400)),
		goai.UserMessage(strings.Repeat("d", 400)),
	}
	prep := compaction.Prepare(messages, compaction.EstimateMessagesTokens(messages), 120, 20, 100, "default")
	if prep.MessagesToSummarize == 0 || prep.RecentMessages == 0 {
		t.Fatalf("bad preparation: %#v", prep)
	}
	if !strings.Contains(prep.Transcript, "user") {
		t.Fatalf("expected transcript: %#v", prep)
	}
}

func TestMaybeCompactContextUsesHookSummary(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	cfg := config.RuntimeConfig{WorkspaceRoot: t.TempDir(), DefaultModel: "bootstrap", MaxIterations: 64, Compaction: config.CompactionSettings{Enabled: true, ContextWindow: 1000, ThresholdTokens: 50, KeepRecentTokens: 20, ReserveTokens: 10}, Agents: config.AgentsConfig{List: []config.AgentConfig{{ID: "agent", Default: true, Model: "bootstrap"}}}}
	e := NewWithRuntimeConfig(s, cfg, "")
	_, err = e.RegisterHook(HookSessionBeforeCompact, "test", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		return HookResponse{Payload: map[string]any{"summary": "smart joker summary"}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	r := &sessionRunner{store: s, engine: e}
	conv := &goai.Context{Messages: []goai.Message{
		goai.UserMessage(strings.Repeat("older1 ", 80)),
		goai.UserMessage(strings.Repeat("older2 ", 80)),
		goai.UserMessage(strings.Repeat("older3 ", 80)),
		goai.UserMessage(strings.Repeat("older4 ", 80)),
		goai.UserMessage("recent question"),
		goai.UserMessage("recent answer"),
	}}
	r.maybeCompactContext(context.Background(), "missing-session", "turn_test", "agent", "bootstrap", conv)
	if len(conv.Messages) >= 6 {
		t.Fatalf("expected compacted context, got %d", len(conv.Messages))
	}
	if !strings.Contains(goai.GetTextContent(&conv.Messages[0]), "smart joker summary") {
		t.Fatalf("missing hook summary: %q", goai.GetTextContent(&conv.Messages[0]))
	}
}
