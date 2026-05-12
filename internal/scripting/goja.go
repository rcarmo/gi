package scripting

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dop251/goja"
	"github.com/rcarmo/gi/internal/connectivity"
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
			obj.Set("runtimeConfig", cfg)
		}
		obj.Set("getRuntimeConfig", func(call goja.FunctionCall) goja.Value {
			cfg, err := bridge.Funcs.GetConfig(ctx)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			v, _ := vm.RunString("(" + mustJSONStr(cfg) + ")")
			return v
		})
	}

	// Session state
	if bridge.Funcs.GetSessionState != nil {
		ss, err := bridge.Funcs.GetSessionState(ctx)
		if err == nil && ss != nil {
			obj.Set("sessionState", ss)
		}
		obj.Set("getSessionState", func(call goja.FunctionCall) goja.Value {
			ss, err := bridge.Funcs.GetSessionState(ctx)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			v, _ := vm.RunString("(" + mustJSONStr(ss) + ")")
			return v
		})
	}
	if bridge.Funcs.SetSessionState != nil {
		obj.Set("setSessionState", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				panic(vm.NewGoError(fmt.Errorf("setSessionState requires a patch argument")))
			}
			exported := call.Arguments[0].Export()
			patch, ok := exported.(map[string]any)
			if !ok {
				b, err := json.Marshal(exported)
				if err != nil {
					panic(vm.NewGoError(fmt.Errorf("setSessionState patch must be an object")))
				}
				if err := json.Unmarshal(b, &patch); err != nil {
					panic(vm.NewGoError(fmt.Errorf("setSessionState patch must be an object: %w", err)))
				}
			}
			if err := bridge.Funcs.SetSessionState(ctx, patch); err != nil {
				panic(vm.NewGoError(err))
			}
			return goja.Undefined()
		})
	}
	if bridge.Funcs.GetSessionInfo != nil {
		info, err := bridge.Funcs.GetSessionInfo(ctx)
		if err == nil && info != nil {
			obj.Set("sessionInfo", info)
		}
		obj.Set("getSessionInfo", func(call goja.FunctionCall) goja.Value {
			info, err := bridge.Funcs.GetSessionInfo(ctx)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			v, _ := vm.RunString("(" + mustJSONStr(info) + ")")
			return v
		})
	}

	// Functions
	if bridge.Funcs.ListTurns != nil {
		obj.Set("listTurns", func(call goja.FunctionCall) goja.Value {
			limit := 50
			if len(call.Arguments) > 0 {
				limit = int(call.Arguments[0].ToInteger())
			}
			turns, err := bridge.Funcs.ListTurns(ctx, limit)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			v, _ := vm.RunString("(" + mustJSONStr(turns) + ")")
			return v
		})
	}

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

	if bridge.Funcs.PublishTopic != nil || bridge.Funcs.SubscribeTopic != nil || bridge.Funcs.ReadTopicSubscription != nil || bridge.Funcs.UnsubscribeTopic != nil {
		topicsObj := vm.NewObject()
		if bridge.Funcs.PublishTopic != nil {
			topicsObj.Set("publish", func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) == 0 {
					panic(vm.NewGoError(fmt.Errorf("topics.publish requires an envelope object")))
				}
				exported := call.Arguments[0].Export()
				envelope, ok := exported.(map[string]any)
				if !ok {
					b, err := json.Marshal(exported)
					if err != nil {
						panic(vm.NewGoError(fmt.Errorf("topics.publish envelope must be an object")))
					}
					if err := json.Unmarshal(b, &envelope); err != nil {
						panic(vm.NewGoError(fmt.Errorf("topics.publish envelope must be an object: %w", err)))
					}
				}
				if err := bridge.Funcs.PublishTopic(ctx, envelope); err != nil {
					panic(vm.NewGoError(err))
				}
				return goja.Undefined()
			})
		}
		if bridge.Funcs.SubscribeTopic != nil {
			topicsObj.Set("subscribe", func(call goja.FunctionCall) goja.Value {
				pattern := "*"
				if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) && !goja.IsNull(call.Arguments[0]) {
					pattern = call.Arguments[0].String()
				}
				opts := TopicSubscribeOptions{}
				if len(call.Arguments) > 1 && !goja.IsUndefined(call.Arguments[1]) && !goja.IsNull(call.Arguments[1]) {
					exported := call.Arguments[1].Export()
					if err := mapToStruct(exported, &opts); err != nil {
						panic(vm.NewGoError(err))
					}
				}
				id, err := bridge.Funcs.SubscribeTopic(ctx, pattern, opts)
				if err != nil {
					panic(vm.NewGoError(err))
				}
				return vm.ToValue(id)
			})
		}
		if bridge.Funcs.ReadTopicSubscription != nil {
			topicsObj.Set("read", func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) == 0 {
					panic(vm.NewGoError(fmt.Errorf("topics.read requires a subscription id")))
				}
				limit := 50
				if len(call.Arguments) > 1 {
					limit = int(call.Arguments[1].ToInteger())
				}
				events, err := bridge.Funcs.ReadTopicSubscription(ctx, call.Arguments[0].String(), limit)
				if err != nil {
					panic(vm.NewGoError(err))
				}
				v, _ := vm.RunString("(" + mustJSONStr(events) + ")")
				return v
			})
		}
		if bridge.Funcs.UnsubscribeTopic != nil {
			topicsObj.Set("unsubscribe", func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) == 0 {
					panic(vm.NewGoError(fmt.Errorf("topics.unsubscribe requires a subscription id")))
				}
				if err := bridge.Funcs.UnsubscribeTopic(ctx, call.Arguments[0].String()); err != nil {
					panic(vm.NewGoError(err))
				}
				return goja.Undefined()
			})
		}
		obj.Set("topics", topicsObj)
	}

	// Event hooks
	if bridge.Funcs.RegisterEventHook != nil {
		registerEventHookFn := func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				panic(vm.NewGoError(fmt.Errorf("registerEventHook requires a hook spec")))
			}
			exported := call.Arguments[0].Export()
			var spec EventHookSpec
			if err := mapToStruct(exported, &spec); err != nil {
				panic(vm.NewGoError(err))
			}
			if err := bridge.Funcs.RegisterEventHook(ctx, spec); err != nil {
				panic(vm.NewGoError(err))
			}
			return goja.Undefined()
		}
		obj.Set("registerEventHook", registerEventHookFn)
		obj.Set("on", registerEventHookFn)
	}
	if bridge.Funcs.RegisterTool != nil {
		obj.Set("registerTool", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				panic(vm.NewGoError(fmt.Errorf("registerTool requires a tool spec")))
			}
			var spec ToolSpec
			if err := mapToStruct(call.Arguments[0].Export(), &spec); err != nil {
				panic(vm.NewGoError(err))
			}
			if err := bridge.Funcs.RegisterTool(ctx, spec); err != nil {
				panic(vm.NewGoError(err))
			}
			return goja.Undefined()
		})
	}
	if bridge.Funcs.SetActiveTools != nil {
		obj.Set("setActiveTools", func(call goja.FunctionCall) goja.Value {
			var names []string
			if len(call.Arguments) > 0 {
				if err := mapToStruct(call.Arguments[0].Export(), &names); err != nil {
					panic(vm.NewGoError(err))
				}
			}
			if err := bridge.Funcs.SetActiveTools(ctx, names); err != nil {
				panic(vm.NewGoError(err))
			}
			return goja.Undefined()
		})
	}
	if bridge.Funcs.GetActiveTools != nil {
		obj.Set("getActiveTools", func(call goja.FunctionCall) goja.Value {
			names, err := bridge.Funcs.GetActiveTools(ctx)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			v, _ := vm.RunString("(" + mustJSONStr(names) + ")")
			return v
		})
	}
	if bridge.Funcs.SetModel != nil {
		obj.Set("setModel", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				panic(vm.NewGoError(fmt.Errorf("setModel requires a model id")))
			}
			if err := bridge.Funcs.SetModel(ctx, call.Arguments[0].String()); err != nil {
				panic(vm.NewGoError(err))
			}
			return goja.Undefined()
		})
	}
	if bridge.Funcs.AppendEntry != nil {
		obj.Set("appendEntry", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 2 {
				panic(vm.NewGoError(fmt.Errorf("appendEntry requires type and data")))
			}
			var data map[string]any
			if err := mapToStruct(call.Arguments[1].Export(), &data); err != nil {
				panic(vm.NewGoError(err))
			}
			if err := bridge.Funcs.AppendEntry(ctx, call.Arguments[0].String(), data); err != nil {
				panic(vm.NewGoError(err))
			}
			return goja.Undefined()
		})
	}
	if bridge.Funcs.GetEntries != nil {
		obj.Set("getEntries", func(call goja.FunctionCall) goja.Value {
			entryType := ""
			if len(call.Arguments) > 0 {
				entryType = call.Arguments[0].String()
			}
			entries, err := bridge.Funcs.GetEntries(ctx, entryType)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			v, _ := vm.RunString("(" + mustJSONStr(entries) + ")")
			return v
		})
	}
	if bridge.Funcs.EmitEvent != nil {
		obj.Set("emitEvent", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				panic(vm.NewGoError(fmt.Errorf("emitEvent requires name and optional payload")))
			}
			name := call.Arguments[0].String()
			payload := map[string]any{}
			if len(call.Arguments) > 1 {
				if err := mapToStruct(call.Arguments[1].Export(), &payload); err != nil {
					panic(vm.NewGoError(err))
				}
			}
			if err := bridge.Funcs.EmitEvent(ctx, name, payload); err != nil {
				panic(vm.NewGoError(err))
			}
			return goja.Undefined()
		})
	}
	if bridge.Funcs.ClearEventHooks != nil {
		obj.Set("clearEventHooks", func(call goja.FunctionCall) goja.Value {
			if err := bridge.Funcs.ClearEventHooks(ctx); err != nil {
				panic(vm.NewGoError(err))
			}
			return goja.Undefined()
		})
	}

	// Connectivity route registry
	connectObj := vm.NewObject()
	if bridge.Funcs.RegisterConnectivityRoute != nil {
		connectObj.Set("registerRoute", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				panic(vm.NewGoError(fmt.Errorf("connect.registerRoute requires a route spec")))
			}
			var spec connectivity.RouteSpec
			if err := mapToStruct(call.Arguments[0].Export(), &spec); err != nil {
				panic(vm.NewGoError(err))
			}
			info, err := bridge.Funcs.RegisterConnectivityRoute(ctx, spec)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			v, _ := vm.RunString("(" + mustJSONStr(info) + ")")
			return v
		})
	}
	if bridge.Funcs.UnregisterConnectivityRoute != nil {
		connectObj.Set("unregisterRoute", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				panic(vm.NewGoError(fmt.Errorf("connect.unregisterRoute requires a route id")))
			}
			if err := bridge.Funcs.UnregisterConnectivityRoute(ctx, call.Arguments[0].String()); err != nil {
				panic(vm.NewGoError(err))
			}
			return goja.Undefined()
		})
	}
	if bridge.Funcs.ListConnectivityRoutes != nil {
		connectObj.Set("listRoutes", func(call goja.FunctionCall) goja.Value {
			filter := map[string]any{}
			if len(call.Arguments) > 0 && call.Arguments[0] != goja.Undefined() && call.Arguments[0] != goja.Null() {
				if err := mapToStruct(call.Arguments[0].Export(), &filter); err != nil {
					panic(vm.NewGoError(err))
				}
			}
			routes, err := bridge.Funcs.ListConnectivityRoutes(ctx, filter)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			v, _ := vm.RunString("(" + mustJSONStr(routes) + ")")
			return v
		})
	}
	if bridge.Funcs.EmitConnectivityEvent != nil {
		connectObj.Set("emit", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				panic(vm.NewGoError(fmt.Errorf("connect.emit requires topic")))
			}
			payload := map[string]any{}
			if len(call.Arguments) > 1 && call.Arguments[1] != goja.Undefined() && call.Arguments[1] != goja.Null() {
				if err := mapToStruct(call.Arguments[1].Export(), &payload); err != nil {
					panic(vm.NewGoError(err))
				}
			}
			if err := bridge.Funcs.EmitConnectivityEvent(ctx, call.Arguments[0].String(), payload); err != nil {
				panic(vm.NewGoError(err))
			}
			return goja.Undefined()
		})
	}
	if len(connectObj.Keys()) > 0 {
		obj.Set("connect", connectObj)
	}

	// Network and transport APIs
	netObj := vm.NewObject()
	if bridge.Funcs.OpenRawSocket != nil {
		netObj.Set("openRawSocket", func(call goja.FunctionCall) goja.Value {
			spec := RawSocketSpec{Protocol: "tcp", TimeoutMS: 5000}
			if len(call.Arguments) > 0 {
				exported := call.Arguments[0].Export()
				if err := mapToStruct(exported, &spec); err != nil {
					panic(vm.NewGoError(err))
				}
			}
			id, err := bridge.Funcs.OpenRawSocket(ctx, spec)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			return vm.ToValue(id)
		})
	}
	if bridge.Funcs.WriteRawSocket != nil {
		netObj.Set("writeRawSocket", func(call goja.FunctionCall) goja.Value {
			payload, err := rawSocketPayloadFromArgs(call)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			n, err := bridge.Funcs.WriteRawSocket(ctx, payload)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			return vm.ToValue(n)
		})
	}
	if bridge.Funcs.ReadRawSocket != nil {
		netObj.Set("readRawSocket", func(call goja.FunctionCall) goja.Value {
			payload, err := rawSocketPayloadFromArgs(call)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			msg, err := bridge.Funcs.ReadRawSocket(ctx, payload)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			return vm.ToValue(msg)
		})
	}
	if bridge.Funcs.CloseRawSocket != nil {
		netObj.Set("closeRawSocket", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				panic(vm.NewGoError(fmt.Errorf("closeRawSocket requires socketId")))
			}
			if err := bridge.Funcs.CloseRawSocket(ctx, call.Arguments[0].String()); err != nil {
				panic(vm.NewGoError(err))
			}
			return goja.Undefined()
		})
	}
	if len(netObj.Keys()) > 0 {
		obj.Set("net", netObj)
	}

	wsObj := vm.NewObject()
	if bridge.Funcs.OpenWebSocket != nil {
		wsObj.Set("open", func(call goja.FunctionCall) goja.Value {
			spec := WebSocketSpec{}
			if len(call.Arguments) > 0 {
				if err := mapToStruct(call.Arguments[0].Export(), &spec); err != nil {
					panic(vm.NewGoError(err))
				}
			}
			id, err := bridge.Funcs.OpenWebSocket(ctx, spec)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			return vm.ToValue(id)
		})
	}
	if bridge.Funcs.WriteWebSocket != nil {
		wsObj.Set("write", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 2 {
				panic(vm.NewGoError(fmt.Errorf("websocket.write requires socketId and payload")))
			}
			if err := bridge.Funcs.WriteWebSocket(ctx, call.Arguments[0].String(), call.Arguments[1].String()); err != nil {
				panic(vm.NewGoError(err))
			}
			return goja.Undefined()
		})
	}
	if bridge.Funcs.ReadWebSocket != nil {
		wsObj.Set("read", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				panic(vm.NewGoError(fmt.Errorf("websocket.read requires socketId")))
			}
			timeout := 5000
			if len(call.Arguments) >= 2 {
				timeout = int(call.Arguments[1].ToInteger())
			}
			msg, err := bridge.Funcs.ReadWebSocket(ctx, call.Arguments[0].String(), timeout)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			return vm.ToValue(msg)
		})
	}
	if bridge.Funcs.CloseWebSocket != nil {
		wsObj.Set("close", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				panic(vm.NewGoError(fmt.Errorf("websocket.close requires socketId")))
			}
			if err := bridge.Funcs.CloseWebSocket(ctx, call.Arguments[0].String()); err != nil {
				panic(vm.NewGoError(err))
			}
			return goja.Undefined()
		})
	}
	if len(wsObj.Keys()) > 0 {
		obj.Set("websocket", wsObj)
	}

	// HTTP request API (full header control)
	if bridge.Funcs.DoHTTPRequest != nil {
		httpObj := vm.NewObject()
		httpObj.Set("request", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				panic(vm.NewGoError(fmt.Errorf("http.request requires request spec")))
			}
			spec := HTTPCallSpec{TimeoutMS: 5000, Headers: map[string][]string{}}
			if err := mapToStruct(call.Arguments[0].Export(), &spec); err != nil {
				panic(vm.NewGoError(err))
			}
			resp, err := bridge.Funcs.DoHTTPRequest(ctx, spec)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			v, _ := vm.RunString("(" + mustJSONStr(resp) + ")")
			return v
		})
		obj.Set("http", httpObj)
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

func mapToStruct(src any, dst any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, dst); err != nil {
		return err
	}
	return nil
}

func rawSocketPayloadFromArgs(call goja.FunctionCall) (RawSocketPayload, error) {
	if len(call.Arguments) == 0 {
		return RawSocketPayload{}, fmt.Errorf("socket payload requires an object")
	}
	var payload RawSocketPayload
	if err := mapToStruct(call.Arguments[0].Export(), &payload); err != nil {
		return RawSocketPayload{}, err
	}
	return payload, nil
}
