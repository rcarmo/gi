package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

func TestLocalBashOutputSanitizedBeforePreview(t *testing.T) {
	c := &chatTUI{}
	lines := c.bashBlockLines("printf", "\x1b[31m  red\x1b[0m\n\x1b]0;title\x07next", "ok", nil, time.Now(), time.Now())
	if got := strings.Join(lines, "\n"); strings.ContainsRune(got, '\x1b') || !strings.Contains(got, "│   red\n│ next") {
		t.Fatalf("bash block: %q", got)
	}
	if got := toolOutputBodyLines("\x1b[31mred\x1b[0m"); len(got) != 1 || got[0] != "red" {
		t.Fatalf("live output: %#v", got)
	}
}

func TestStoredSingleLineToolOutputPreservesSourceSpacing(t *testing.T) {
	c := &chatTUI{}
	lines := c.renderToolResultLines(store.Message{ID: "one", Role: "tool_result", Content: "    if ready {  return 42 }\x1b[0m", Payload: map[string]any{"tool_name": "shell"}})
	if len(lines) != 2 || lines[1] != "│     if ready {  return 42 }" {
		t.Fatalf("single-line source spacing: %#v", lines)
	}
	if got := toolOutputBodyLines("    if ready {  return 42 }\x1b[0m"); len(got) != 1 || got[0] != "    if ready {  return 42 }" {
		t.Fatalf("live source spacing: %#v", got)
	}
}
