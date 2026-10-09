package turn

import (
	"context"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/compaction"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	goai "github.com/rcarmo/go-ai"
)

func TestConfiguredTransportReachesMainTurnAndCompaction(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	defer s.Close()
	inference.Init()
	goai.RegisterModel(&goai.Model{Provider: "transport-fixture", ID: "model", Api: goai.ApiOpenAICompletions, ContextWindow: 64000, MaxTokens: 4096, Input: []string{"text"}})
	if _, err := s.CreateSession(ctx, "transport-session", "Transport", nil); err != nil {
		t.Fatal(err)
	}
	seen := make(chan goai.Transport, 4)
	withStreamWithToolsHookStub(t, func(_ context.Context, _ string, _ *goai.Context, _ func(map[string]any), h *inference.StreamHooks) (*inference.StreamResult, error) {
		seen <- h.Transport
		return &inference.StreamResult{Text: "done", Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}}, nil
	})
	e := NewWithRuntimeConfig(s, config.RuntimeConfig{DefaultProvider: "transport-fixture", DefaultModel: "model", DefaultThinkingLevel: "off", Transport: "sse"}, "")
	defer e.Close()
	turn, err := e.SubmitPrompt(ctx, RunInput{SessionID: "transport-session", Prompt: "hello", Model: "transport-fixture/model"})
	if err != nil {
		t.Fatal(err)
	}
	check := func() {
		t.Helper()
		select {
		case got := <-seen:
			if got != goai.TransportSSE {
				t.Fatalf("transport=%q want sse", got)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("provider was not invoked")
		}
	}
	check()
	_, err = e.runner("transport-session").compactionSummarizer(turn.TurnID, "transport-fixture/model")(ctx, compaction.SummaryRequest{Prompt: "Summarise", MaxTokens: 1024})
	if err != nil {
		t.Fatal(err)
	}
	check()
}
