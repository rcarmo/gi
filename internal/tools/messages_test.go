package tools

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestMessageRetrievalToolStrictArgumentsAndRuntimeScope(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "messages.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	for _, id := range []string{"A", "B"} {
		if _, err := s.CreateSession(ctx, id, id, nil); err != nil {
			t.Fatal(err)
		}
		if err := s.AddMessage(ctx, "msg-"+id, id, "user", "secret-"+id, nil); err != nil {
			t.Fatal(err)
		}
	}
	rt := ToolRuntime{Store: s, SessionID: "A"}
	tool := MessagesTool()
	if tool.Kind != "read-only" || tool.Source != "builtin" || !json.Valid(tool.Parameters) {
		t.Fatal(tool)
	}
	text, err := tool.Executor(ctx, rt, goai.ToolCall{Arguments: map[string]any{}})
	if err != nil || !strings.Contains(text, "secret-A") || strings.Contains(text, "secret-B") {
		t.Fatal(text, err)
	}
	bad := []string{
		`{"session_id":"B"}`, `{"chat_jid":"*"}`, `{"action":"delete"}`, `{"row_ids":null}`, `{"row_ids":[]}`, `{"row_ids":[1.2]}`, `{"row_ids":["1"]}`, `{"row_ids":[1,1]}`, `{"row_ids":[9007199254740992]}`,
		`{"limit":0}`, `{"limit":101}`, `{"limit":true}`, `{"limit":null}`, `{"limit":1.5}`, `{"content_bytes":2049}`, `{"content_bytes":0}`, `{"after_row":0}`, `{"before_row":0}`, `{"after_row":5,"before_row":4}`, `{"row_ids":[1],"before_row":2}`, `{"context_before":1}`, `{"row_ids":[1],"context_before":11}`, `{"cursor":33}`,
	}
	for _, input := range bad {
		var args map[string]any
		json.Unmarshal([]byte(input), &args)
		if out, err := ExecuteMessages(ctx, rt, goai.ToolCall{Arguments: args}); err == nil {
			t.Fatalf("accepted %s: %s", input, out)
		}
	}
	for _, runtime := range []ToolRuntime{{Store: s}, {SessionID: "A"}, {}} {
		if _, err := ExecuteMessages(ctx, runtime, goai.ToolCall{}); err == nil {
			t.Fatal("missing runtime accepted")
		}
	}
	if _, err := ExecuteMessages(ctx, rt, goai.ToolCall{Arguments: map[string]any{"cursor": strings.Repeat("x", 17000)}}); err == nil {
		t.Fatal("oversized arguments accepted")
	}
}
