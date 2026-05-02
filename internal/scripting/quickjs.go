package scripting

import (
	"context"
	"fmt"
)

// QuickJSRunner is retained as an explicit non-implemented placeholder.
// It is intentionally out of scope while Goja is the supported in-process
// JavaScript runtime for current releases.
// The interface is kept to avoid future API churn.
type QuickJSRunner struct{}

func NewQuickJSRunner() *QuickJSRunner {
	return &QuickJSRunner{}
}

func (r *QuickJSRunner) Name() string { return "quickjs" }

func (r *QuickJSRunner) Execute(ctx context.Context, script string, bridge *Bridge) (string, error) {
	return "", fmt.Errorf("quickjs: not in scope (Goja is supported JavaScript engine)")
}

func (r *QuickJSRunner) ExecuteFile(ctx context.Context, path string, bridge *Bridge) (string, error) {
	return "", fmt.Errorf("quickjs: not in scope (Goja is supported JavaScript engine)")
}
