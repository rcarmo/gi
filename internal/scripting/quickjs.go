package scripting

import (
	"context"
	"fmt"
)

// QuickJSRunner will execute JavaScript scripts via QuickJS.
// This is a placeholder for v1.1 — the interface is defined so
// the bridge design accounts for both engines from the start.
type QuickJSRunner struct{}

func NewQuickJSRunner() *QuickJSRunner {
	return &QuickJSRunner{}
}

func (r *QuickJSRunner) Name() string { return "quickjs" }

func (r *QuickJSRunner) Execute(ctx context.Context, script string, bridge *Bridge) (string, error) {
	return "", fmt.Errorf("quickjs: not yet implemented (planned for v1.1)")
}

func (r *QuickJSRunner) ExecuteFile(ctx context.Context, path string, bridge *Bridge) (string, error) {
	return "", fmt.Errorf("quickjs: not yet implemented (planned for v1.1)")
}
