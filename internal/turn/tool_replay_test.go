package turn

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestForkToolReplayAfterReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "replay.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateSession(ctx, "source", "source", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, "turn", "source", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	calls := []map[string]any{{"id": "call", "name": "shell", "arguments": map[string]any{"command": "exit 1"}}}
	if err := s.AddMessage(ctx, "calls", "source", "assistant", "[tool_call: shell]", map[string]any{"kind": "tool_calls", "display_text": "Checking", "tool_calls": calls}); err != nil {
		t.Fatal(err)
	}
	if err := s.SettleToolResult(ctx, "turn", "source", "tool.failed", map[string]any{"tool": "shell", "tool_call_id": "call", "occurrence_id": "o"}, "result", "exit status 1", map[string]any{"kind": "tool_result", "tool_name": "shell", "tool_call_id": "call", "occurrence_id": "o", "turn_id": "turn", "is_error": true, "details": map[string]any{"exit": 1}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CloneSessionBefore(ctx, "source", "fork", "fork", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	runner := &sessionRunner{}
	for _, id := range []string{"source", "fork"} {
		snapshot, err := s.ContextSnapshot(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		messages := runner.projectContextSnapshot(ctx, id, snapshot)
		if len(messages) != 2 {
			t.Fatalf("%s replay: %+v", id, messages)
		}
		restored := goai.GetToolCalls(&messages[0])
		if len(restored) != 1 || restored[0].ID != "call" || restored[0].Arguments["command"] != "exit 1" {
			t.Fatalf("%s calls: %+v", id, restored)
		}
		if messages[1].Role != goai.RoleToolResult || messages[1].ToolCallID != "call" || !messages[1].IsError || messages[1].Details == nil {
			t.Fatalf("%s result: %+v", id, messages[1])
		}
	}
	fork, err := s.ListMessages(ctx, "fork")
	if err != nil {
		t.Fatal(err)
	}
	if fork[0].ID == "calls" || fork[1].ID == "result" {
		t.Fatal("fork reused row IDs")
	}
	terminal, err := s.ListTerminalToolResults(ctx, "fork")
	if err != nil || len(terminal) != 0 {
		t.Fatalf("source terminal leaked: %+v %v", terminal, err)
	}
}

func TestCompactionCannotSplitReplayedToolExchange(t *testing.T) {
	r := &sessionRunner{}
	snapshot := store.ContextSnapshot{Messages: []store.Message{
		{ID: "call", Role: "assistant", Payload: map[string]any{"kind": "tool_calls", "tool_calls": []map[string]any{{"id": "c", "name": "shell", "arguments": map[string]any{}}}}},
		{ID: "result", Role: "tool_result", Payload: map[string]any{"kind": "tool_result", "tool_call_id": "c", "tool_name": "shell"}, Content: "done"},
	}}
	conv := &goai.Context{Messages: r.projectContextSnapshot(context.Background(), "s", snapshot)}
	split, err := r.compactionBoundary(context.Background(), "s", conv, snapshot, map[string]any{"messages_to_summarize": 1})
	if err != nil || split != nil {
		t.Fatalf("split accepted: %+v %v", split, err)
	}
	complete, err := r.compactionBoundary(context.Background(), "s", conv, snapshot, map[string]any{"messages_to_summarize": 2})
	if err != nil || complete == nil {
		t.Fatalf("complete exchange rejected: %+v %v", complete, err)
	}
}
