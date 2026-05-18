package tools_test

import (
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/turn"
)

func TestExecuteToolsToolDetailByName(t *testing.T) {
	result, err := executeToolsTool(map[string]any{"name": "shell"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, `"name": "shell"`) || !strings.Contains(result, `"command"`) {
		t.Fatalf("unexpected detail: %q", result)
	}
}

func TestExecuteToolsToolActivation(t *testing.T) {
	e := turn.New(nil)
	result, err := e.ExecuteToolsMeta(map[string]any{"activate": []any{"read"}})
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !strings.Contains(result, "read") || !strings.Contains(result, "tools") {
		t.Fatalf("activation result: %q", result)
	}
	active := strings.Join(e.ActiveTools(), ",")
	if strings.Contains(active, "write") || !strings.Contains(active, "read") || !strings.Contains(active, "tools") {
		t.Fatalf("active=%s", active)
	}
}
