package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

func renderedText(c *chatTUI) string {
	var out []string
	for _, r := range c.transcriptRowsAtWidth(100) {
		out = append(out, strings.TrimRight(r.text, " "))
	}
	return strings.Join(out, "\n")
}

// A codemode call renders like Pi's: the script, nested calls as rows of
// the same block (never separate tool blocks), and the output without the
// "Script completed" header.
func TestCodemodeBlockLive(t *testing.T) {
	c := &chatTUI{transcriptExpanded: map[string]bool{}}
	now := time.Now()
	code := "const r = await tools.greet({ name: \"bob\" });\nreturn r;"
	c.renderToolEvent(map[string]any{"type": "tool_started", "tool": "codemode", "tool_call_id": "cm1", "turn_id": "t1", "arguments": map[string]any{"code": code}}, now)
	c.renderToolEvent(map[string]any{"type": "tool_started", "tool": "greet", "tool_call_id": "cm1/1", "turn_id": "t1", "parent_tool_call_id": "cm1", "arguments": map[string]any{"name": "bob"}}, now)
	live := renderedText(c)
	if !strings.Contains(live, `… greet {"name":"bob"}`) {
		t.Fatalf("running call row missing:\n%s", live)
	}
	c.renderToolEvent(map[string]any{"type": "tool_finished", "tool": "greet", "tool_call_id": "cm1/1", "turn_id": "t1", "parent_tool_call_id": "cm1", "duration_ms": 250}, now.Add(9*time.Second))
	c.renderToolEvent(map[string]any{"type": "tool_started", "tool": "mcp__x__y", "tool_call_id": "cm1/2", "turn_id": "t1", "parent_tool_call_id": "cm1"}, now)
	c.renderToolEvent(map[string]any{"type": "tool_failed", "tool": "mcp__x__y", "tool_call_id": "cm1/2", "turn_id": "t1", "parent_tool_call_id": "cm1", "error": "nope"}, now.Add(time.Second))
	output := "Script completed\nWall time 0.3 seconds\nOutput:\nhello bob"
	c.renderToolEvent(map[string]any{"type": "tool_finished", "tool": "codemode", "tool_call_id": "cm1", "turn_id": "t1", "output": output}, now.Add(2*time.Second))
	text := renderedText(c)
	for _, want := range []string{"codemode", "const r = await tools.greet", `✓ greet {"name":"bob"} 250ms`, "✗ mcp__x__y", "hello bob"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "Script completed") || strings.Count(strings.Join(c.transcript, "\n"), "⟦gi:block:") != 1 {
		t.Fatalf("header shown or extra blocks:\n%s", text)
	}
}

// A reloaded codemode result renders the same rows from its stored details.
func TestCodemodeBlockFromStoredResult(t *testing.T) {
	c := &chatTUI{transcriptExpanded: map[string]bool{}}
	m := store.Message{ID: "m1", Role: "tool_result", Content: "Script failed\nWall time 1.0 seconds\nOutput:\npartial\nScript error:\nError: boom",
		Payload: map[string]any{"tool_name": "codemode", "is_error": true, "details": map[string]any{"calls": []any{
			map[string]any{"id": "a/1", "name": "read", "args": `{"path":"x"}`, "status": "ok", "durationMs": 1500.0},
		}, "fullOutputPath": "vfs://codemode-output/s/x.txt"}}}
	c.transcript = c.renderToolResultWithArguments(m, map[string]any{"code": "await tools.read({path:'x'})"})
	text := renderedText(c)
	for _, want := range []string{"await tools.read", `✓ read {"path":"x"} 1.5s`, "partial", "Error: boom"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "Script failed") {
		t.Fatalf("header shown:\n%s", text)
	}
}
