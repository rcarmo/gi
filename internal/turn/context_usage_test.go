package turn

import (
	"context"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	goai "github.com/rcarmo/go-ai"
)

func TestProviderLoopRecordsContextFromActualResultUsage(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	defer e.Close()
	ctx := context.Background()
	session, err := s.CreateSession(ctx, "loop-context", "Loop", nil)
	if err != nil {
		t.Fatal(err)
	}
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}, StopReason: "stop"}, Usage: &goai.Usage{Input: 100, CacheRead: 25, CacheWrite: 5, Output: 20, TotalTokens: 150}}, nil
	})
	result, err := e.SubmitPrompt(ctx, RunInput{SessionID: session.ID, Prompt: "measure", Model: "provider/model"})
	if err != nil {
		t.Fatal(err)
	}
	waitForCondition(t, 3*time.Second, func() bool {
		turn, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && turn.Status == "completed"
	}, "provider loop completion")
	measured, err := s.LatestContextMeasurement(ctx, session.ID)
	if err != nil || measured == nil || measured.Tokens != 130 {
		t.Fatalf("provider input missing: %+v %v", measured, err)
	}
}

func TestProviderRequestContextMeasurementIsNotCumulativeTurnUsage(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	e := New(s)
	defer e.Close()
	if _, err := s.CreateSession(ctx, "context-origin", "Context", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, "context-turn", "context-origin", "hello", map[string]any{"model": "provider/model"}); err != nil {
		t.Fatal(err)
	}
	runner := e.runner("context-origin")
	runner.persistContextMeasurement(s, "context-turn", "context-origin", "provider/model", &goai.Usage{Input: 100, CacheRead: 20, CacheWrite: 30, Output: 400, TotalTokens: 550}, 1)
	runner.persistContextMeasurement(s, "context-turn", "context-origin", "provider/model", &goai.Usage{Input: 140, CacheRead: 40, CacheWrite: 10, Output: 700, TotalTokens: 890}, 2)
	runner.persistUsage(s, "context-turn", "context-origin", &goai.Usage{Input: 240, CacheRead: 60, CacheWrite: 40, Output: 1100, TotalTokens: 1440}, 2)
	measurement, err := s.LatestContextMeasurement(ctx, "context-origin")
	if err != nil {
		t.Fatal(err)
	}
	if measurement == nil || measurement.Tokens != 190 || measurement.Iteration != 2 || measurement.Model != "provider/model" {
		t.Fatalf("wrong latest measurement %+v", measurement)
	}
	usage, err := inference.SessionContextUsage(ctx, s, "context-origin", 1000)
	if err != nil {
		t.Fatal(err)
	}
	if usage["tokens"] != 190 || usage["percent"] != float64(19) {
		t.Fatalf("inflated context: %+v", usage)
	}
	runner.persistContextMeasurement(s, "context-turn", "context-origin", "provider/model", nil, 3)
	runner.persistContextMeasurement(s, "context-turn", "context-origin", "provider/model", &goai.Usage{}, 3)
	runner.persistContextMeasurement(s, "context-turn", "context-origin", "provider/model", &goai.Usage{Input: -1}, 3)
	measurement, _ = s.LatestContextMeasurement(ctx, "context-origin")
	if measurement.Iteration != 2 {
		t.Fatal("invalid usage replaced measurement")
	}
}
