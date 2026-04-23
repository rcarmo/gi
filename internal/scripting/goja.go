package scripting

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dop251/goja"
)

// GojaRunner executes JavaScript scripts using goja — a pure Go
// ECMAScript 5.1+ engine compiled into the binary. No CGO, no
// external binary, no subprocess.
type GojaRunner struct{}

func NewGojaRunner() *GojaRunner {
	return &GojaRunner{}
}

func (r *GojaRunner) Name() string { return "js" }

func (r *GojaRunner) Execute(ctx context.Context, script string, bridge *Bridge) (string, error) {
	vm := goja.New()

	// Inject the bridge state as a global JS object
	bridgeObj, err := buildJSBridge(ctx, vm, bridge)
	if err != nil {
		return "", fmt.Errorf("build bridge: %w", err)
	}
	vm.Set("gi", bridgeObj)

	// Capture console.log output
	var output []string
	console := vm.NewObject()
	console.Set("log", func(call goja.FunctionCall) goja.Value {
		parts := make([]string, len(call.Arguments))
		for i, arg := range call.Arguments {
			parts[i] = arg.String()
		}
		output = append(output, strings.Join(parts, " "))
		return goja.Undefined()
	})
	console.Set("error", console.Get("log"))
	console.Set("warn", console.Get("log"))
	console.Set("debug", console.Get("log"))
	vm.Set("console", console)

	// Context cancellation check via interrupt
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			vm.Interrupt("context cancelled")
		case <-done:
		}
	}()
	defer close(done)

	// Execute
	val, err := vm.RunString(script)
	if err != nil {
		if len(output) > 0 {
			return strings.Join(output, "\n"), err
		}
		return "", fmt.Errorf("js: %w", err)
	}

	// If there was console output, return that; otherwise return the result
	if len(output) > 0 {
		return strings.Join(output, "\n"), nil
	}

	if val == nil || val == goja.Undefined() || val == goja.Null() {
		return "", nil
	}

	// Try to export as a Go value and JSON-encode it
	exported := val.Export()
	if s, ok := exported.(string); ok {
		return s, nil
	}
	b, err := json.Marshal(exported)
	if err != nil {
		return val.String(), nil
	}
	return string(b), nil
}

func (r *GojaRunner) ExecuteFile(ctx context.Context, path string, bridge *Bridge) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read script %s: %w", path, err)
	}
	return r.Execute(ctx, string(content), bridge)
}

func buildJSBridge(ctx context.Context, vm *goja.Runtime, bridge *Bridge) (*goja.Object, error) {
	obj := vm.NewObject()

	obj.Set("sessionId", bridge.SessionID)

	// Config
	if bridge.Funcs.GetConfig != nil {
		cfg, err := bridge.Funcs.GetConfig(ctx)
		if err == nil && cfg != nil {
			obj.Set("config", cfg)
		}
	}

	// Session state
	if bridge.Funcs.GetSessionState != nil {
		ss, err := bridge.Funcs.GetSessionState(ctx)
		if err == nil && ss != nil {
			obj.Set("sessionState", ss)
		}
	}

	// Functions
	if bridge.Funcs.ListMessages != nil {
		obj.Set("listMessages", func(call goja.FunctionCall) goja.Value {
			limit := 50
			if len(call.Arguments) > 0 {
				limit = int(call.Arguments[0].ToInteger())
			}
			msgs, err := bridge.Funcs.ListMessages(ctx, limit)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			v, _ := vm.RunString("(" + mustJSONStr(msgs) + ")")
			return v
		})
	}

	if bridge.Funcs.ReadFile != nil {
		obj.Set("readFile", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				panic(vm.NewGoError(fmt.Errorf("readFile requires a path argument")))
			}
			content, err := bridge.Funcs.ReadFile(ctx, call.Arguments[0].String())
			if err != nil {
				panic(vm.NewGoError(err))
			}
			return vm.ToValue(content)
		})
	}

	if bridge.Funcs.WriteFile != nil {
		obj.Set("writeFile", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 2 {
				panic(vm.NewGoError(fmt.Errorf("writeFile requires path and content")))
			}
			err := bridge.Funcs.WriteFile(ctx, call.Arguments[0].String(), call.Arguments[1].String())
			if err != nil {
				panic(vm.NewGoError(err))
			}
			return goja.Undefined()
		})
	}

	if bridge.Funcs.ListDir != nil {
		obj.Set("listDir", func(call goja.FunctionCall) goja.Value {
			path := "."
			if len(call.Arguments) > 0 {
				path = call.Arguments[0].String()
			}
			entries, err := bridge.Funcs.ListDir(ctx, path)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			v, _ := vm.RunString("(" + mustJSONStr(entries) + ")")
			return v
		})
	}

	if bridge.Funcs.Log != nil {
		obj.Set("log", func(call goja.FunctionCall) goja.Value {
			level := "info"
			msg := ""
			if len(call.Arguments) >= 2 {
				level = call.Arguments[0].String()
				msg = call.Arguments[1].String()
			} else if len(call.Arguments) == 1 {
				msg = call.Arguments[0].String()
			}
			bridge.Funcs.Log(ctx, level, msg)
			return goja.Undefined()
		})
	}

	return obj, nil
}

func mustJSONStr(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
