package scripting

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/rcarmo/gi/internal/codemode"
)

// QuickJSRunner executes user JavaScript in a fresh QuickJS/WASM isolate.
// It uses the same host bridge as Goja; bridge calls return promises and must
// be awaited. Goja only marshals existing Go host bindings, never user code.
type QuickJSRunner struct{}

func NewQuickJSRunner() *QuickJSRunner { return &QuickJSRunner{} }
func (r *QuickJSRunner) Name() string  { return "quickjs" }
func (r *QuickJSRunner) Execute(ctx context.Context, script string, bridge *Bridge) (string, error) {
	// The bridge closures and QuickJS VM share the same deadline, so a host
	// call cannot outlive the isolate's default execution timeout.
	ctx, cancel := context.WithTimeout(ctx, codemode.DefaultTimeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if bridge == nil {
		return "", fmt.Errorf("quickjs: bridge is required")
	}
	engine, err := codemode.Default()
	if err != nil {
		return "", err
	}
	vm := goja.New()
	object, err := buildJSBridge(ctx, vm, bridge)
	if err != nil {
		return "", fmt.Errorf("build bridge: %w", err)
	}
	// Snapshot JSON values and capture only host-callable bindings. All namespace
	// paths come from the host object, never from an untrusted script lookup.
	functions := map[string]goja.Callable{}
	paths := []string{}
	var snapshot func(*goja.Object, string) map[string]any
	snapshot = func(o *goja.Object, prefix string) map[string]any {
		out := map[string]any{}
		for _, key := range o.Keys() {
			v := o.Get(key)
			path := key
			if prefix != "" {
				path = prefix + "." + key
			}
			if fn, ok := goja.AssertFunction(v); ok {
				functions[path] = fn
				paths = append(paths, path)
				continue
			}
			if child, ok := v.(*goja.Object); ok && child.ClassName() == "Object" {
				out[key] = snapshot(child, path)
			} else {
				out[key] = v.Export()
			}
		}
		return out
	}
	initial, err := json.Marshal(snapshot(object, ""))
	if err != nil {
		return "", fmt.Errorf("quickjs bridge values: %w", err)
	}
	pathJSON, _ := json.Marshal(paths)
	initialJSON, _ := json.Marshal(string(initial))
	sourceJSON, _ := json.Marshal(script)
	var mu sync.Mutex
	result := engine.Execute(ctx, fmt.Sprintf(`const gi = JSON.parse(%s);
 for(const path of %s){const parts=path.split('.');let o=gi;for(const part of parts.slice(0,-1)){o=o[part]??=(Object.create(null));}o[parts.at(-1)]=(...args)=>__giBridge(path,args);}
 return await eval(%s);`, initialJSON, pathJSON, sourceJSON), codemode.Options{Globals: []codemode.Global{{Name: "__giBridge", Spread: true, Execute: func(callCtx context.Context, args json.RawMessage) (json.RawMessage, error) {
		if err := callCtx.Err(); err != nil {
			return nil, err
		}
		var a []json.RawMessage
		if err := json.Unmarshal(args, &a); err != nil || len(a) != 2 {
			return nil, fmt.Errorf("invalid bridge call")
		}
		var path string
		var argv []any
		if err := json.Unmarshal(a[0], &path); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(a[1], &argv); err != nil {
			return nil, err
		}
		mu.Lock()
		defer mu.Unlock()
		fn, ok := functions[path]
		if !ok {
			return nil, fmt.Errorf("unknown bridge function")
		}
		values := make([]goja.Value, len(argv))
		for i, x := range argv {
			values[i] = vm.ToValue(x)
		}
		v, err := fn(goja.Undefined(), values...)
		if err != nil {
			return nil, err
		}
		if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
			return json.RawMessage("null"), nil
		}
		return json.Marshal(v.Export())
	}}}})
	var output []string
	for _, item := range result.Output {
		if item.Type == "text" {
			output = append(output, item.Text)
		}
	}
	if !result.OK {
		if ctx.Err() != nil {
			return strings.Join(output, "\n"), ctx.Err()
		}
		return strings.Join(output, "\n"), result.Error
	}
	if len(output) > 0 {
		return strings.Join(output, "\n"), nil
	}
	if len(result.Value) == 0 || string(result.Value) == "null" {
		return "", nil
	}
	var text string
	if json.Unmarshal(result.Value, &text) == nil {
		return text, nil
	}
	return string(result.Value), nil
}
func (r *QuickJSRunner) ExecuteFile(ctx context.Context, path string, b *Bridge) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return r.Execute(ctx, string(raw), b)
}
