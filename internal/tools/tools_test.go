package tools_test

import (
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/turn"
)

func executeToolsTool(args map[string]any) (string, error) {
	return turn.New(nil).ExecuteToolsMeta(args)
}

func TestExecuteToolsToolListAll(t *testing.T) {
	result, err := executeToolsTool(map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "tool(s):") {
		t.Fatalf("expected tool count header, got: %q", result)
	}
	for _, name := range []string{"tools", "skills", "compact", "rtk", "read", "write", "shell"} {
		if !strings.Contains(result, "- "+name+":") {
			t.Fatalf("expected tool %q in listing, got: %q", name, result)
		}
	}
}
