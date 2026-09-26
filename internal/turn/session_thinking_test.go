package turn

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestSessionThinkingCapturedAtAdmissionNotRuntimeOrCallerMetadata(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	defer e.Close()
	inference.Init()
	low, high := "low", "high"
	model := "thinking-test/reasoner"
	goai.RegisterModel(&goai.Model{ID: "reasoner", Provider: "thinking-test", Api: goai.ApiOpenAICompletions, ContextWindow: 32000, MaxTokens: 1024, Reasoning: true, ThinkingLevelMap: map[goai.ModelThinkingLevel]*string{"off": nil, "minimal": nil, "low": &low, "medium": nil, "high": &high}})
	_, err := s.CreateSession(ctx, "A", "A", map[string]any{"selected_model": model, "thinking_model": model, "thinking_level": "low"})
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	original := streamWithToolsWithHooks
	defer func() { streamWithToolsWithHooks = original }()
	var mu sync.Mutex
	var levels []string
	streamWithToolsWithHooks = func(ctx context.Context, _ string, _ *goai.Context, _ func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
		mu.Lock()
		levels = append(levels, hooks.Thinking)
		n := len(levels)
		mu.Unlock()
		if n == 1 {
			close(started)
			select {
			case <-release:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		msg := goai.Message{Role: "assistant", Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}
		return &inference.StreamResult{Message: &msg, Text: "done"}, nil
	}
	first, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "one", Model: model, Metadata: map[string]any{"selected_thinking_level": "high", "selected_thinking_model": model}})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("not started")
	}
	set := func(level string) {
		t.Helper()
		row, err := s.GetSession(ctx, "A")
		if err != nil {
			t.Fatal(err)
		}
		if err = s.SelectSessionThinking(ctx, "A", store.SessionThinkingToken("A", row.State), model, level); err != nil {
			t.Fatal(err)
		}
	}
	set("high")
	second, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "two", Intent: "queue", Model: model, Metadata: map[string]any{"selected_thinking_level": "low"}})
	if err != nil || !second.Queued {
		t.Fatal(second, err)
	}
	set("")
	one, _ := s.GetTurn(ctx, first.TurnID)
	two, _ := s.GetTurn(ctx, second.TurnID)
	if one.Metadata["selected_thinking_level"] != "low" || two.Metadata["selected_thinking_level"] != "high" {
		t.Fatal(one.Metadata, two.Metadata)
	}
	once.Do(func() { close(release) })
	waitForCondition(t, 3*time.Second, func() bool { row, err := s.GetTurn(ctx, second.TurnID); return err == nil && row.Status == "completed" }, "queued completion")
	waitForCondition(t, 3*time.Second, func() bool { _, _, err := s.GetSessionActiveTurn(ctx, "A"); return errors.Is(err, sql.ErrNoRows) }, "cleanup")
	mu.Lock()
	defer mu.Unlock()
	if len(levels) != 2 || levels[0] != "low" || levels[1] != "high" {
		t.Fatal(levels)
	}
}
