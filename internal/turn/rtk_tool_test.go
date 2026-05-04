package turn

import (
	"context"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestRTKToolFilterOnly(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e := New(s)
	tool, ok := e.tools.Get("rtk")
	if !ok {
		t.Fatal("missing rtk tool")
	}
	out, err := tool.Executor(context.Background(), ToolRuntime{Engine: e, Store: s, WorkspaceRoot: t.TempDir()}, goai.ToolCall{Name: "rtk", Arguments: map[string]any{"command": "go test ./...", "filter_only": true, "output": "ok a\n--- FAIL: TestX\nFAIL\n"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "go-test") || !strings.Contains(out, "FAIL") || strings.Contains(out, "ok a") {
		t.Fatalf("unexpected rtk output: %s", out)
	}
}
