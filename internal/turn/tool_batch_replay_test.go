package turn

import (
	"context"
	"testing"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func TestPreparedToolReplayUsesBothPoliciesAndSavedArguments(t *testing.T) {
	for _, mode := range []string{"safe", "saved_default", "current_default", "missing_tool", "result_hook_attempted", "callback_returned", "schema_changed"} {
		t.Run(mode, func(t *testing.T) {
			saved := store.ToolBatchCall{ReplaySafe: true, Prepared: goai.ToolCall{Name: "probe", Arguments: map[string]any{"n": "already prepared"}}}
			current := tools.RegisteredTool{Name: "probe", ReplaySafe: true, Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) { return "", nil }}
			want := false
			switch mode {
			case "safe":
				want = true
			case "saved_default":
				saved.ReplaySafe = false
			case "current_default":
				current.ReplaySafe = false
			case "missing_tool":
				current.Executor = nil
			case "result_hook_attempted":
				saved.ResultHookStarted = true
			case "callback_returned":
				saved.CallbackReturned = true
				want = true
			case "schema_changed":
				current.Parameters = []byte(`{"required":["replacement"]}`)
				want = true
			}
			if got := preparedToolReplaySafe(saved, current); got != want {
				t.Fatalf("replay=%v want %v", got, want)
			}
			if saved.Prepared.Arguments["n"] != "already prepared" {
				t.Fatal("prepared argument changed")
			}
		})
	}
}
