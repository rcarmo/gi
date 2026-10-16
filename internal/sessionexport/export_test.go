package sessionexport

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/store"
)

func exportFixture(t *testing.T) (*store.Store, string) {
	t.Helper()
	ctx := context.Background()
	s, err := store.Open(filepath.Join(t.TempDir(), "gi.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	const sid = "session_export"
	if _, err := s.CreateSession(ctx, sid, "Export <demo>", map[string]any{}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_a", sid, "completed", "list files", map[string]any{}); err != nil {
		t.Fatal(err)
	}
	add := func(id, role, content string, payload map[string]any) {
		t.Helper()
		if err := s.AddMessage(ctx, id, sid, role, content, payload); err != nil {
			t.Fatal(err)
		}
	}
	// New-style tool calls carry full arguments.
	add("m1", "user", "list files", map[string]any{"turn_id": "turn_a"})
	add("m2", "assistant", "[tool_call: shell]", map[string]any{"kind": "tool_calls", "model": "github-copilot/gpt-5-mini", "turn_id": "turn_a",
		"tool_calls": []any{map[string]any{"id": "call_1", "name": "shell", "arguments": map[string]any{"command": "ls -la"}}}})
	add("m3", "tool_result", "a.go\nb.go", map[string]any{"tool_call_id": "call_1", "is_error": false, "turn_id": "turn_a"})
	// Legacy tool calls: arguments come from the tool.started preview.
	add("m4", "assistant", "[tool_call: read]", map[string]any{"kind": "tool_calls", "model": "github-copilot/gpt-5-mini", "turn_id": "turn_a"})
	if err := s.AppendTurnEvent(ctx, "turn_a", sid, "tool.started", map[string]any{"tool": "read", "tool_call_id": "call_2", "preview": "a.go"}); err != nil {
		t.Fatal(err)
	}
	add("m5", "tool_result", "package a", map[string]any{"tool_call_id": "call_2", "is_error": true, "turn_id": "turn_a"})
	add("m6", "assistant", "Two files: <a.go> & b.go", map[string]any{"kind": "chat", "model": "github-copilot/gpt-5-mini", "turn_id": "turn_a"})
	add("m7", "system", "compaction completed", map[string]any{})
	return s, sid
}

// The JSONL export follows Pi's session format v3: a header, then a linear
// id/parentId chain of message entries with Pi's roles and content blocks.
func TestJSONLFollowsPiSessionFormat(t *testing.T) {
	s, sid := exportFixture(t)
	data, err := JSONL(context.Background(), s, sid, "/work")
	if err != nil {
		t.Fatal(err)
	}
	var lines []map[string]any
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		var m map[string]any
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			t.Fatalf("invalid JSON line %q: %v", sc.Text(), err)
		}
		lines = append(lines, m)
	}
	if h := lines[0]; h["type"] != "session" || h["version"] != float64(3) || h["cwd"] != "/work" || h["id"] != sid {
		t.Fatalf("header %v", h)
	}
	if lines[1]["type"] != "session_info" || lines[1]["name"] != "Export <demo>" || lines[1]["parentId"] != nil {
		t.Fatalf("name entry %v", lines[1])
	}
	for i := 2; i < len(lines); i++ {
		if lines[i]["parentId"] != lines[i-1]["id"] {
			t.Fatalf("entry %d parentId %v, want %v", i, lines[i]["parentId"], lines[i-1]["id"])
		}
	}
	msg := func(i int) map[string]any { return lines[i]["message"].(map[string]any) }
	if m := msg(2); m["role"] != "user" || m["content"] != "list files" {
		t.Fatalf("user %v", m)
	}
	// The first assistant message's model is a model change, as in Pi.
	if e := lines[3]; e["type"] != "model_change" || e["provider"] != "github-copilot" || e["modelId"] != "gpt-5-mini" {
		t.Fatalf("model change %v", e)
	}
	// gi's shell tool is Pi's bash.
	call := msg(4)["content"].([]any)[0].(map[string]any)
	if msg(4)["stopReason"] != "toolUse" || msg(4)["provider"] != "github-copilot" || msg(4)["model"] != "gpt-5-mini" ||
		call["type"] != "toolCall" || call["id"] != "call_1" || call["name"] != "bash" || call["arguments"].(map[string]any)["command"] != "ls -la" {
		t.Fatalf("assistant tool call %v", msg(4))
	}
	if m := msg(5); m["role"] != "toolResult" || m["toolCallId"] != "call_1" || m["toolName"] != "bash" || m["isError"] != false {
		t.Fatalf("tool result %v", m)
	}
	// Legacy calls: the preview under the key Pi's tool takes.
	legacy := msg(6)["content"].([]any)[0].(map[string]any)
	if legacy["id"] != "call_2" || legacy["name"] != "read" || legacy["arguments"].(map[string]any)["path"] != "a.go" {
		t.Fatalf("legacy call %v", legacy)
	}
	if m := msg(7); m["toolName"] != "read" || m["isError"] != true {
		t.Fatalf("legacy result %v", m)
	}
	if m := msg(8); m["stopReason"] != "stop" || m["content"].([]any)[0].(map[string]any)["text"] != "Two files: <a.go> & b.go" {
		t.Fatalf("assistant text %v", m)
	}
	if e := lines[9]; e["type"] != "custom" || e["customType"] != "gi.system" {
		t.Fatalf("system notice %v", e)
	}
}

// Responses recorded with their usage, stop reason, thinking and thinking
// level export them as Pi's assistant messages and thinking_level_change
// entries.
func TestJSONLExportsResponseRecords(t *testing.T) {
	ctx := context.Background()
	s, sid := exportFixture(t)
	usage := map[string]any{"input": 12, "output": 7, "cacheRead": 3, "cacheWrite": 0, "totalTokens": 22, "cost": map[string]any{"input": 0.1, "output": 0.2, "cacheRead": 0, "cacheWrite": 0, "total": 0.3}}
	add := func(id, model, level, stop string) {
		payload := map[string]any{"kind": "chat", "model": model, "usage": usage, "stop_reason": stop, "thinking_level": level,
			"thinking_blocks": []any{map[string]any{"type": "thinking", "thinking": "Plan.", "thinkingSignature": "sig"}}}
		if err := s.AddMessage(ctx, id, sid, "assistant", "Answer "+id, payload); err != nil {
			t.Fatal(err)
		}
	}
	add("r1", "anthropic/claude-sonnet-4-5", "high", "stop")
	add("r2", "anthropic/claude-sonnet-4-5", "high", "length")
	add("r3", "anthropic/claude-sonnet-4-5", "low", "stop")
	_, entries, err := Document(ctx, s, sid, "/work")
	if err != nil {
		t.Fatal(err)
	}
	var kinds []string
	for _, e := range entries[len(entries)-6:] {
		kind, _ := e["type"].(string)
		if m, ok := e["message"].(map[string]any); ok {
			kind = fmt.Sprint(m["role"], ":", m["stopReason"])
		}
		kinds = append(kinds, kind)
	}
	if got := strings.Join(kinds, ","); got != "model_change,thinking_level_change,assistant:stop,assistant:length,thinking_level_change,assistant:stop" {
		t.Fatalf("entries %s", got)
	}
	m := entries[len(entries)-1]["message"].(map[string]any)
	content := m["content"].([]any)
	if first := content[0].(map[string]any); first["type"] != "thinking" || first["thinking"] != "Plan." || first["thinkingSignature"] != "sig" {
		t.Fatalf("thinking %v", content)
	}
	if fmt.Sprint(m["usage"]) != fmt.Sprint(usage) {
		t.Fatalf("usage %v", m["usage"])
	}
	if e := entries[len(entries)-2]; e["thinkingLevel"] != "low" {
		t.Fatalf("thinking change %v", e)
	}
}

// Export picks Pi's HTML page by default and JSONL for .jsonl paths,
// resolved against cwd. The page carries the session as base64 data, so no
// message text reaches the markup.
func TestExportPathsAndHTML(t *testing.T) {
	s, sid := exportFixture(t)
	cwd := t.TempDir()
	path, err := Export(context.Background(), s, sid, cwd, "", Theme{})
	if err != nil || path != filepath.Join(cwd, "gi-session-"+sid+".html") {
		t.Fatalf("default path %q %v", path, err)
	}
	page, _ := os.ReadFile(path)
	m := regexp.MustCompile(`id="session-data"[^>]*>([A-Za-z0-9+/=]+)<`).FindSubmatch(page)
	if m == nil || strings.Contains(string(page), "Two files") {
		t.Fatal("HTML export does not carry the session as data")
	}
	decoded, _ := base64.StdEncoding.DecodeString(string(m[1]))
	var data SessionData
	if err := json.Unmarshal(decoded, &data); err != nil || data.Header["id"] != sid || !strings.Contains(string(decoded), "Two files: <a.go> & b.go") || data.LeafID != data.Entries[len(data.Entries)-1]["id"] {
		t.Fatalf("session data %s (%v)", decoded, err)
	}
	if !strings.Contains(string(page), "--exportPageBg: rgb(36, 37, 46);") {
		t.Fatal("no default export colours")
	}
	path, err = Export(context.Background(), s, sid, cwd, "out/session.jsonl", Theme{})
	if err != nil || path != filepath.Join(cwd, "out", "session.jsonl") {
		t.Fatalf("jsonl path %q %v", path, err)
	}
	data2, _ := os.ReadFile(path)
	var header map[string]any
	if err := json.Unmarshal([]byte(strings.SplitN(string(data2), "\n", 2)[0]), &header); err != nil || header["type"] != "session" {
		t.Fatalf("jsonl header %.120s (%v)", data2, err)
	}
}

func TestRecordedCallsRetainArgumentsError(t *testing.T) {
	calls := recordedCalls(map[string]any{"tool_calls": []any{map[string]any{"id": "bad", "name": "shell", "arguments": map[string]any{}, "argumentsError": "Expected object"}}})
	if len(calls) != 1 || calls[0].ArgumentsError != "Expected object" {
		t.Fatalf("parse error lost: %+v", calls)
	}
	entry := (item{role: "assistant", calls: calls}).entry()
	msg := entry["message"].(map[string]any)
	blocks := msg["content"].([]any)
	if blocks[0].(map[string]any)["argumentsError"] != "Expected object" {
		t.Fatalf("export lost marker: %+v", entry)
	}
}
