package turn

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestSidePromptContextNoToolsNoPersistenceAndSessionIsolation(t *testing.T) {
	db := openTestStore(t)
	defer db.Close()
	ctx := context.Background()
	db.CreateSession(ctx, "side-a", "side-a", nil)
	db.CreateSession(ctx, "side-b", "side-b", nil)
	db.AddMessage(ctx, "history-a", "side-a", "user", "history alpha", nil)
	db.AddMessage(ctx, "history-b", "side-b", "user", "other session secret", nil)
	cfg := config.RuntimeConfig{DefaultProvider: "side-fixture", DefaultModel: "model", DefaultThinkingLevel: "off", Transport: "sse"}
	e := NewWithRuntimeConfig(db, cfg, "")
	defer e.Close()
	goai.RegisterModel(&goai.Model{Provider: "side-fixture", ID: "model", Api: goai.ApiOpenAICompletions, ContextWindow: 32000})
	old := streamWithToolsWithHooks
	defer func() { streamWithToolsWithHooks = old }()
	streamWithToolsWithHooks = func(ctx context.Context, model string, conv *goai.Context, emit func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
		if len(conv.Tools) != 0 || model != "side-fixture/model" || hooks.Transport != goai.TransportSSE || hooks.MaxTokens != 1024 || hooks.CacheRetention != goai.CacheRetentionNone {
			t.Fatal(model, hooks, conv.Tools)
		}
		var transcript strings.Builder
		for _, m := range conv.Messages {
			for _, c := range m.Content {
				transcript.WriteString(c.Text)
			}
		}
		if !strings.Contains(transcript.String(), "history alpha") || !strings.Contains(transcript.String(), "side question") || strings.Contains(transcript.String(), "other session secret") {
			t.Fatal(transcript.String())
		}
		emit(map[string]any{"type": "thinking_delta", "delta": "thought"})
		emit(map[string]any{"type": "text_delta", "delta": "answer"})
		return &inference.StreamResult{Text: "answer", Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop}}, nil
	}
	result, err := e.RunSidePrompt(ctx, "side-a", "side question", "", nil)
	if err != nil || result.Result != "answer" || result.Thinking != "thought" {
		t.Fatal(result, err)
	}
	messages, _ := db.ListMessages(ctx, "side-a")
	turns, _ := db.ListTurns(ctx, "side-a")
	if len(messages) != 1 || len(turns) != 0 {
		t.Fatal("side request persisted", len(messages), len(turns))
	}
}

func TestSidePromptCancellationErrorsAndOutputBounds(t *testing.T) {
	db := openTestStore(t)
	defer db.Close()
	db.CreateSession(context.Background(), "s", "s", nil)
	e := NewWithRuntimeConfig(db, config.RuntimeConfig{DefaultProvider: "side-fixture", DefaultModel: "model", DefaultThinkingLevel: "off"}, "")
	defer e.Close()
	goai.RegisterModel(&goai.Model{Provider: "side-fixture", ID: "model", Api: goai.ApiOpenAICompletions, ContextWindow: 32000})
	old := streamWithToolsWithHooks
	defer func() { streamWithToolsWithHooks = old }()
	streamWithToolsWithHooks = func(ctx context.Context, _ string, _ *goai.Context, emit func(map[string]any), _ *inference.StreamHooks) (*inference.StreamResult, error) {
		emit(map[string]any{"type": "text_delta", "delta": strings.Repeat("x", 512*1024+1)})
		return &inference.StreamResult{}, ctx.Err()
	}
	if _, err := e.RunSidePrompt(context.Background(), "s", "q", "", nil); err == nil || !strings.Contains(err.Error(), "output exceeds") {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := e.RunSidePrompt(ctx, "s", "q", "", nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := e.RunSidePrompt(context.Background(), "s", "", "", nil); err == nil {
		t.Fatal("empty prompt")
	}
	if _, err := e.RunSidePrompt(context.Background(), "missing", "q", "", nil); err == nil {
		t.Fatal("unknown session")
	}
}

func TestSidePromptSnapshotSurvivesMainAppendAndAbortAndEngineCloseJoins(t *testing.T) {
	db := openTestStore(t)
	defer db.Close()
	ctx := context.Background()
	if _, err := db.CreateSession(ctx, "s", "s", nil); err != nil {
		t.Fatal(err)
	}
	if err := db.AddMessage(ctx, "before", "s", "user", "before-side", nil); err != nil {
		t.Fatal(err)
	}
	e := NewWithRuntimeConfig(db, config.RuntimeConfig{DefaultProvider: "side-fixture", DefaultModel: "model", DefaultThinkingLevel: "off"}, "")
	defer e.Close()
	goai.RegisterModel(&goai.Model{Provider: "side-fixture", ID: "model", Api: goai.ApiOpenAICompletions, ContextWindow: 32000})
	old := streamWithToolsWithHooks
	defer func() { streamWithToolsWithHooks = old }()
	entered := make(chan struct{})
	exited := make(chan struct{})
	release := make(chan struct{})
	streamWithToolsWithHooks = func(ctx context.Context, _ string, conv *goai.Context, _ func(map[string]any), _ *inference.StreamHooks) (*inference.StreamResult, error) {
		defer close(exited)
		close(entered)
		<-ctx.Done()
		for _, m := range conv.Messages {
			for _, b := range m.Content {
				if strings.Contains(b.Text, "after-side") {
					t.Error("snapshot changed")
				}
			}
		}
		<-release
		return nil, ctx.Err()
	}
	sideDone := make(chan error, 1)
	go func() { _, err := e.RunSidePrompt(ctx, "s", "question", "", nil); sideDone <- err }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("side inference not started")
	}
	if err := db.AddMessage(ctx, "after", "s", "user", "after-side", nil); err != nil {
		t.Fatal(err)
	}
	// Stopping the main session never falls back to cancelling side inference.
	_ = e.CancelActiveTurn(ctx, "s", "")
	select {
	case <-sideDone:
		t.Fatal("main stop cancelled side inference")
	default:
	}
	closeDone := make(chan struct{})
	go func() { e.Close(); close(closeDone) }()
	select {
	case <-closeDone:
		t.Fatal("engine closed without joining side inference")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	select {
	case err := <-sideDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("side request not cancelled")
	}
	select {
	case <-closeDone:
	case <-time.After(time.Second):
		t.Fatal("close did not join")
	}
	select {
	case <-exited:
	default:
		t.Fatal("inference still running")
	}
	if _, err := e.RunSidePrompt(ctx, "s", "question", "", nil); err == nil {
		t.Fatal("admitted after close")
	}
}

func TestSidePromptCompactedContextModelChoiceAndFit(t *testing.T) {
	db := openTestStore(t)
	defer db.Close()
	ctx := context.Background()
	if _, err := db.CreateSession(ctx, "s", "s", map[string]any{"selected_model": "side-choice/reasoner", "thinking_model": "side-choice/reasoner", "thinking_level": "high"}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"covered", "recent"} {
		if err := db.AddMessage(ctx, id, "s", "user", id, nil); err != nil {
			t.Fatal(err)
		}
	}
	snapshot, err := db.ContextSnapshot(ctx, "s")
	if err != nil {
		t.Fatal(err)
	}
	boundary, err := store.PrepareContextBoundary(snapshot, 1)
	if err != nil {
		t.Fatal(err)
	}
	// Seed an already committed checkpoint; compaction transaction tests live
	// in store. This test exercises its read-only side-request projection.
	covered, err := json.Marshal(boundary.Covered)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB().ExecContext(ctx, `insert into context_checkpoints(session_id,version,summary,covered_json,created_at) values('s',1,'sealed summary',?,strftime('%Y-%m-%dT%H:%M:%fZ','now'))`, string(covered)); err != nil {
		t.Fatal(err)
	}
	e := NewWithRuntimeConfig(db, config.RuntimeConfig{DefaultProvider: "ignored", DefaultModel: "model"}, "")
	defer e.Close()
	high := "high"
	goai.RegisterModel(&goai.Model{Provider: "side-choice", ID: "reasoner", Api: goai.ApiOpenAICompletions, ContextWindow: 8000, Reasoning: true, ThinkingLevelMap: map[goai.ModelThinkingLevel]*string{"high": &high}})
	old := streamWithToolsWithHooks
	defer func() { streamWithToolsWithHooks = old }()
	calls := 0
	streamWithToolsWithHooks = func(_ context.Context, model string, conv *goai.Context, _ func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
		calls++
		if model != "side-choice/reasoner" || hooks.Thinking != "high" || conv.SystemPrompt != "custom system" || len(conv.Messages) != 3 {
			t.Fatal(model, hooks, conv)
		}
		if !strings.Contains(conv.Messages[0].Content[0].Text, "sealed summary") || conv.Messages[1].Content[0].Text != "recent" {
			t.Fatal(conv.Messages)
		}
		return &inference.StreamResult{Text: "final only", Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "thinking", Thinking: "final thought"}}}}, nil
	}
	result, err := e.RunSidePrompt(ctx, "s", "q", "custom system", nil)
	if err != nil || result.Result != "final only" || result.Thinking != "final thought" {
		t.Fatal(result, err)
	}
	if _, err := e.RunSidePrompt(ctx, "s", strings.Repeat("q", 64*1024), "", nil); err == nil || !strings.Contains(err.Error(), "does not fit") {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("called provider on oversized context", calls)
	}
}

func TestSidePromptIncompleteAndFailedFinalResponse(t *testing.T) {
	db := openTestStore(t)
	defer db.Close()
	db.CreateSession(context.Background(), "s", "s", nil)
	e := NewWithRuntimeConfig(db, config.RuntimeConfig{DefaultProvider: "side-fixture", DefaultModel: "model"}, "")
	defer e.Close()
	goai.RegisterModel(&goai.Model{Provider: "side-fixture", ID: "model", Api: goai.ApiOpenAICompletions, ContextWindow: 32000})
	old := streamWithToolsWithHooks
	defer func() { streamWithToolsWithHooks = old }()
	for _, message := range []*goai.Message{nil, {Role: goai.RoleAssistant, StopReason: goai.StopReasonError}, {Role: goai.RoleAssistant, StopReason: goai.StopReasonAborted}} {
		streamWithToolsWithHooks = func(context.Context, string, *goai.Context, func(map[string]any), *inference.StreamHooks) (*inference.StreamResult, error) {
			return &inference.StreamResult{Text: "partial", Message: message}, nil
		}
		if _, err := e.RunSidePrompt(context.Background(), "s", "q", "", nil); err == nil {
			t.Fatal("incomplete response accepted", message)
		}
	}
}

func TestSidePromptMediaBoundBeforeProjection(t *testing.T) {
	db := openTestStore(t)
	defer db.Close()
	ctx := context.Background()
	db.CreateSession(ctx, "s", "s", nil)
	media, err := db.CreateMedia(ctx, "s", "large.png", "image/png", []byte("fixture"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB().ExecContext(ctx, `update media set original_size=? where id=?`, 2*1024*1024, media.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.AddMessage(ctx, "m", "s", "user", "image", map[string]any{"media": []any{map[string]any{"media_id": media.ID}}}); err != nil {
		t.Fatal(err)
	}
	e := NewWithRuntimeConfig(db, config.RuntimeConfig{DefaultProvider: "side-fixture", DefaultModel: "model"}, "")
	defer e.Close()
	if _, err := e.RunSidePrompt(ctx, "s", "q", "", nil); err == nil || !strings.Contains(err.Error(), "context exceeds limit") {
		t.Fatal(err)
	}
}
