package turn

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func TestBusyResetWaitsForCompleteToolBatch(t *testing.T) {
	ctx := context.Background()
	s, err := store.Open(filepath.Join(t.TempDir(), "reset-batch.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e := New(s)
	defer e.Close()
	const sid, tid = "reset-session", "reset-turn"
	if _, err := s.CreateSession(ctx, sid, "s", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, tid, sid, "prompt", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, sid, tid, "test", "claim"); err != nil || !ok {
		t.Fatalf("claim: %t %v", ok, err)
	}
	executed := []string{}
	e.RegisterTool(tools.RegisteredTool{Name: "reset_first", Description: "fixture", Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) {
		executed = append(executed, "first")
		queued, err := e.ResetSession(ctx, sid, "after reset")
		if err != nil || !queued {
			t.Fatalf("reset admission: %t %v", queued, err)
		}
		return "first result", nil
	}})
	e.RegisterTool(tools.RegisteredTool{Name: "reset_second", Description: "fixture", Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) {
		executed = append(executed, "second")
		return "second result", nil
	}})
	outcome := e.runner(sid).executeToolCallsPhase(ctx, s, tid, sid, "fixture", "agent", 1, &goai.Context{}, []goai.ToolCall{{ID: "a", Name: "reset_first", Arguments: map[string]any{}}, {ID: "b", Name: "reset_second", Arguments: map[string]any{}}}, nil, "", 0, &goai.Usage{})
	if !outcome.terminated || !reflect.DeepEqual(executed, []string{"first", "second"}) {
		t.Fatalf("batch reset: %+v executed=%v", outcome, executed)
	}
	snapshot, err := s.ContextSnapshot(ctx, sid)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Messages) != 0 {
		t.Fatalf("old tool history remained after reset: %+v", snapshot.Messages)
	}
	messages, err := s.ListMessages(ctx, sid)
	if err != nil {
		t.Fatal(err)
	}
	results := 0
	for _, m := range messages {
		if m.Role == "tool_result" {
			results++
		}
	}
	if results != 2 {
		t.Fatalf("visible tool history lost: %+v", messages)
	}
}

func TestFinalAnswerResetRetainsAnswerButResetsProviderHead(t *testing.T) {
	ctx := context.Background()
	s, err := store.Open(filepath.Join(t.TempDir(), "reset-final.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e := New(s)
	defer e.Close()
	if _, err := s.CreateSession(ctx, "s", "s", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, "t", "s", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "s", "t", "test", "claim"); err != nil || !ok {
		t.Fatalf("claim %t %v", ok, err)
	}
	if _, err := e.ResetSession(ctx, "s", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.AddMessage(ctx, "answer", "s", "assistant", "final answer", nil); err != nil {
		t.Fatal(err)
	}
	e.runner("s").finishTurnOK(s, "t", "s", "agent", "fixture", 1)
	snapshot, err := s.ContextSnapshot(ctx, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Messages) != 0 {
		t.Fatalf("old answer entered new head: %+v", snapshot.Messages)
	}
	page, err := s.PageConversationMessages(ctx, "s", "", "", 50)
	if err != nil || len(page.Messages) != 1 || page.Messages[0].Content != "final answer" {
		t.Fatalf("answer lost: %+v %v", page, err)
	}
}
