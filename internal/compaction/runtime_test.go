package compaction

import (
	"context"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	goai "github.com/rcarmo/go-ai"
)

func TestPrepareCompactionKeepsRecentMessages(t *testing.T) {
	messages := []goai.Message{
		goai.UserMessage(strings.Repeat("a", 400)),
		goai.UserMessage(strings.Repeat("b", 400)),
		goai.UserMessage(strings.Repeat("c", 400)),
		goai.UserMessage(strings.Repeat("d", 400)),
	}
	prep := Prepare(messages, EstimateMessagesTokens(messages), 120, 20, 100, "default")
	if prep.MessagesToSummarize == 0 || prep.RecentMessages == 0 {
		t.Fatalf("bad preparation: %#v", prep)
	}
	if !strings.Contains(prep.Transcript, "user") {
		t.Fatalf("expected transcript: %#v", prep)
	}
}

func TestMaybeCompactContextUsesHookSummary(t *testing.T) {
	conv := &goai.Context{Messages: []goai.Message{
		goai.UserMessage(strings.Repeat("older1 ", 80)),
		goai.UserMessage(strings.Repeat("older2 ", 80)),
		goai.UserMessage(strings.Repeat("older3 ", 80)),
		goai.UserMessage(strings.Repeat("older4 ", 80)),
		goai.UserMessage("recent question"),
		goai.UserMessage("recent answer"),
	}}
	MaybeCompactContext(context.Background(), RuntimeRequest{SessionID: "s", TurnID: "t", AgentID: "agent", Model: "bootstrap", Settings: config.CompactionSettings{Enabled: true, ContextWindow: 1000, ThresholdTokens: 50, KeepRecentTokens: 20, ReserveTokens: 10}}, conv, RuntimeOps{BeforeCompact: func(context.Context, map[string]any, []goai.Message) (HookDecision, error) {
		return HookDecision{Payload: map[string]any{"summary": "smart joker summary"}}, nil
	}})
	if len(conv.Messages) >= 6 {
		t.Fatalf("expected compacted context, got %d", len(conv.Messages))
	}
	if !strings.Contains(goai.GetTextContent(&conv.Messages[0]), "smart joker summary") {
		t.Fatalf("missing hook summary: %q", goai.GetTextContent(&conv.Messages[0]))
	}
}
