package turn

import (
	"context"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestPeeringToolStatus(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e := New(s)
	tool, ok := e.tools.Get("peering")
	if !ok {
		t.Fatal("missing peering tool")
	}
	out, err := tool.Executor(context.Background(), ToolRuntime{Engine: e, Store: s, WorkspaceRoot: t.TempDir()}, goai.ToolCall{Name: "peering", Arguments: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"\"backend\": \"tsnet\"", "\"state\": \"disabled\""} {
		if !strings.Contains(out, want) {
			t.Fatalf("peering output missing %q: %s", want, out)
		}
	}
}
