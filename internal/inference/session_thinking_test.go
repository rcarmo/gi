package inference

import (
	"context"
	"testing"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestSessionThinkingCatalogueValidationAndFailBeforeNetwork(t *testing.T) {
	Init()
	low, high := "low", "high"
	goai.RegisterModel(&goai.Model{ID: "restricted", Provider: "thinking-fixture", Api: goai.ApiOpenAICompletions, ContextWindow: 32000, Reasoning: true, ThinkingLevelMap: map[goai.ModelThinkingLevel]*string{"off": nil, "minimal": nil, "low": &low, "medium": nil, "high": &high}})
	model := "thinking-fixture/restricted"
	levels := ThinkingLevels(model)
	if len(levels) != 2 || levels[0] != "low" || levels[1] != "high" {
		t.Fatal(levels)
	}
	for _, level := range []string{"", "low", "high"} {
		if err := ValidateThinking(model, level); err != nil {
			t.Fatal(level, err)
		}
	}
	for _, level := range []string{"medium", "off", "HIGH", " high", "xhigh", "bogus"} {
		if err := ValidateThinking(model, level); err == nil {
			t.Fatal(level)
		}
		_, err := StreamWithToolsWithHooks(context.Background(), model, &goai.Context{}, nil, &StreamHooks{Thinking: level})
		if err == nil || err.Error() != "thinking level is not supported by this model" {
			t.Fatal(level, err)
		}
	}
	if len(ThinkingLevels("unknown/model")) != 0 {
		t.Fatal("unknown advertised support")
	}
	if len(ThinkingLevels("opencode-zen/minimax-m2.5-free")) != 0 {
		t.Fatal("unsupported custom transport")
	}
	session := &store.Session{State: map[string]any{"thinking_model": model, "thinking_level": "high"}}
	if CapturedSessionThinking(session, model) != "high" {
		t.Fatal("missing capture")
	}
	if CapturedSessionThinking(session, "other/model") != "" {
		t.Fatal("cross-model capture")
	}
	delete(session.State, "thinking_model")
	if CapturedSessionThinking(session, model) != "" {
		t.Fatal("legacy unvalidated captured")
	}
}
