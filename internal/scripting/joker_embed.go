package scripting

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	core "github.com/candid82/joker/core"
	_ "github.com/candid82/joker/std/base64"
	_ "github.com/candid82/joker/std/bolt"
	_ "github.com/candid82/joker/std/crypto"
	_ "github.com/candid82/joker/std/csv"
	_ "github.com/candid82/joker/std/filepath"
	_ "github.com/candid82/joker/std/git"
	_ "github.com/candid82/joker/std/hex"
	_ "github.com/candid82/joker/std/html"
	_ "github.com/candid82/joker/std/http"
	_ "github.com/candid82/joker/std/io"
	_ "github.com/candid82/joker/std/json"
	_ "github.com/candid82/joker/std/markdown"
	_ "github.com/candid82/joker/std/math"
	_ "github.com/candid82/joker/std/os"
	_ "github.com/candid82/joker/std/runtime"
	_ "github.com/candid82/joker/std/strconv"
	_ "github.com/candid82/joker/std/string"
	_ "github.com/candid82/joker/std/time"
	_ "github.com/candid82/joker/std/url"
	_ "github.com/candid82/joker/std/uuid"
	_ "github.com/candid82/joker/std/yaml"
	"github.com/rcarmo/gi/internal/connectivity"
)

var embeddedJokerMu sync.Mutex

// ExecuteEmbeddedJoker runs a Joker script in-process and applies any
// controlled session-state mutations back through the live bridge.
func ExecuteEmbeddedJoker(ctx context.Context, script string, bridge *Bridge) (string, error) {
	bridgeState := buildBridgeState(ctx, bridge)
	bridgeJSON := mustJSON(bridgeState)

	embeddedJokerMu.Lock()
	defer embeddedJokerMu.Unlock()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	core.Stdout = &stdout
	core.Stderr = &stderr
	core.Stdin = bytes.NewBuffer(nil)
	core.GLOBAL_ENV.InitEnv(core.Stdin, core.Stdout, core.Stderr, nil)
	core.GLOBAL_ENV.SetClassPath("")
	core.GLOBAL_ENV.SetCurrentNamespace(core.GLOBAL_ENV.EnsureSymbolIsNamespace(core.MakeSymbol("user")))
	installJokerBridgeProcedures(ctx, bridge)

	fullScript := buildEmbeddedPreamble(bridgeJSON) + "\n(println (json/write-string {\"result\" (do\n" + script + "\n) \"session_state\" @*gi-session-state*}))"
	r := core.NewReader(bufio.NewReader(strings.NewReader(fullScript)), "<gi-joker>")

	if err := runEmbeddedJokerReader(r); err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("joker: %s", strings.TrimSpace(stderr.String()))
		}
		return "", fmt.Errorf("joker: %w", err)
	}
	if stderr.Len() > 0 {
		return "", fmt.Errorf("joker: %s", strings.TrimSpace(stderr.String()))
	}

	var payload struct {
		Result       any            `json:"result"`
		SessionState map[string]any `json:"session_state"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		return strings.TrimSpace(stdout.String()), nil
	}

	if bridge.Funcs.SetSessionState != nil {
		original, _ := bridgeState["session-state"].(map[string]any)
		patch := diffMap(original, payload.SessionState)
		if len(patch) > 0 {
			if err := bridge.Funcs.SetSessionState(ctx, patch); err != nil {
				return "", err
			}
		}
	}

	return stringifyScriptResult(payload.Result), nil
}

func runEmbeddedJokerReader(reader *core.Reader) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%v", v)
			}
		}
	}()

	parseContext := &core.ParseContext{GlobalEnv: core.GLOBAL_ENV}
	for {
		obj, readErr := core.TryRead(reader)
		if readErr != nil {
			if readErr == io.EOF {
				return nil
			}
			return readErr
		}
		expr, parseErr := core.TryParse(obj, parseContext)
		if parseErr != nil {
			return parseErr
		}
		if _, evalErr := core.TryEval(expr); evalErr != nil {
			return evalErr
		}
	}
}

func diffMap(before, after map[string]any) map[string]any {
	if after == nil {
		return map[string]any{}
	}
	if before == nil {
		before = map[string]any{}
	}
	patch := map[string]any{}
	for k, v := range after {
		if !valuesEqual(before[k], v) {
			patch[k] = v
		}
	}
	return patch
}

func valuesEqual(a, b any) bool {
	ab, errA := json.Marshal(a)
	bb, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return fmt.Sprint(a) == fmt.Sprint(b)
	}
	return bytes.Equal(ab, bb)
}

func stringifyScriptResult(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
}

func installJokerBridgeProcedures(ctx context.Context, bridge *Bridge) {
	userNS := core.GLOBAL_ENV.EnsureSymbolIsNamespace(core.MakeSymbol("user"))

	meta := core.EmptyArrayMap()
	register := func(name string, fn func(args []core.Object) core.Object) {
		userNS.InternVar(name, core.Proc{Fn: core.ProcFn(fn), Name: name, Package: "scripting"}, meta)
	}

	register("__gi-register-event-hook", func(args []core.Object) core.Object {
		if bridge.Funcs.RegisterEventHook == nil {
			panic(core.RT.NewError("register event hook is not available"))
		}
		var payload string
		var err error
		payload, err = readStringArg(args, 0, "register-event-hook")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		var spec EventHookSpec
		if err := json.Unmarshal([]byte(payload), &spec); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		if err := bridge.Funcs.RegisterEventHook(ctx, spec); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.NIL
	})

	register("__gi-emit-event", func(args []core.Object) core.Object {
		if bridge.Funcs.EmitEvent == nil {
			panic(core.RT.NewError("emit event is not available"))
		}
		if len(args) == 0 {
			panic(core.RT.NewError("emit event requires a name"))
		}
		name, err := readStringArg(args, 0, "emit-event")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		payload := map[string]any{}
		if len(args) > 1 {
			rawPayload, err := readStringArg(args, 1, "emit-event")
			if err != nil {
				panic(core.RT.NewError(err.Error()))
			}
			if len(rawPayload) > 0 {
				if err := json.Unmarshal([]byte(rawPayload), &payload); err != nil {
					panic(core.RT.NewError(err.Error()))
				}
			}
		}
		if err := bridge.Funcs.EmitEvent(ctx, name, payload); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.NIL
	})

	register("__gi-clear-event-hooks", func(args []core.Object) core.Object {
		if bridge.Funcs.ClearEventHooks == nil {
			panic(core.RT.NewError("clear event hooks is not available"))
		}
		if err := bridge.Funcs.ClearEventHooks(ctx); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.NIL
	})

	register("__gi-register-tool", func(args []core.Object) core.Object {
		if bridge.Funcs.RegisterTool == nil {
			panic(core.RT.NewError("register tool is not available"))
		}
		payload, err := readStringArg(args, 0, "register-tool")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		var spec ToolSpec
		if err := json.Unmarshal([]byte(payload), &spec); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		if err := bridge.Funcs.RegisterTool(ctx, spec); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.NIL
	})

	register("__gi-set-active-tools", func(args []core.Object) core.Object {
		if bridge.Funcs.SetActiveTools == nil {
			panic(core.RT.NewError("set active tools is not available"))
		}
		payload, err := readStringArg(args, 0, "set-active-tools")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		var names []string
		if err := json.Unmarshal([]byte(payload), &names); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		if err := bridge.Funcs.SetActiveTools(ctx, names); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.NIL
	})

	register("__gi-register-route", func(args []core.Object) core.Object {
		if bridge.Funcs.RegisterConnectivityRoute == nil {
			panic(core.RT.NewError("register route is not available"))
		}
		payload, err := readStringArg(args, 0, "register-route")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		var spec connectivity.RouteSpec
		if err := json.Unmarshal([]byte(payload), &spec); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		info, err := bridge.Funcs.RegisterConnectivityRoute(ctx, spec)
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		b, _ := json.Marshal(info)
		return core.MakeString(string(b))
	})

	register("__gi-unregister-route", func(args []core.Object) core.Object {
		if bridge.Funcs.UnregisterConnectivityRoute == nil {
			panic(core.RT.NewError("unregister route is not available"))
		}
		id, err := readStringArg(args, 0, "unregister-route")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		if err := bridge.Funcs.UnregisterConnectivityRoute(ctx, id); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.NIL
	})

	register("__gi-list-routes", func(args []core.Object) core.Object {
		if bridge.Funcs.ListConnectivityRoutes == nil {
			panic(core.RT.NewError("list routes is not available"))
		}
		filter := map[string]any{}
		if len(args) > 0 {
			payload, err := readStringArg(args, 0, "list-routes")
			if err != nil {
				panic(core.RT.NewError(err.Error()))
			}
			if strings.TrimSpace(payload) != "" && payload != "{}" {
				if err := json.Unmarshal([]byte(payload), &filter); err != nil {
					panic(core.RT.NewError(err.Error()))
				}
			}
		}
		routes, err := bridge.Funcs.ListConnectivityRoutes(ctx, filter)
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		b, _ := json.Marshal(routes)
		return core.MakeString(string(b))
	})

	register("__gi-emit-connectivity-event", func(args []core.Object) core.Object {
		if bridge.Funcs.EmitConnectivityEvent == nil {
			panic(core.RT.NewError("emit connectivity event is not available"))
		}
		topic, err := readStringArg(args, 0, "emit-connectivity-event")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		payload := map[string]any{}
		if len(args) > 1 {
			raw, err := readStringArg(args, 1, "emit-connectivity-event")
			if err != nil {
				panic(core.RT.NewError(err.Error()))
			}
			if strings.TrimSpace(raw) != "" && raw != "{}" {
				if err := json.Unmarshal([]byte(raw), &payload); err != nil {
					panic(core.RT.NewError(err.Error()))
				}
			}
		}
		if err := bridge.Funcs.EmitConnectivityEvent(ctx, topic, payload); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.NIL
	})

	register("__gi-publish-topic", func(args []core.Object) core.Object {
		if bridge.Funcs.PublishTopic == nil {
			panic(core.RT.NewError("publish topic is not available"))
		}
		payload, err := readStringArg(args, 0, "publish-topic")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		env := map[string]any{}
		if strings.TrimSpace(payload) != "" && payload != "{}" {
			if err := json.Unmarshal([]byte(payload), &env); err != nil {
				panic(core.RT.NewError(err.Error()))
			}
		}
		if err := bridge.Funcs.PublishTopic(ctx, env); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.NIL
	})

	register("__gi-subscribe-topic", func(args []core.Object) core.Object {
		if bridge.Funcs.SubscribeTopic == nil {
			panic(core.RT.NewError("subscribe topic is not available"))
		}
		pattern, err := readStringArg(args, 0, "subscribe-topic")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		opts := TopicSubscribeOptions{}
		if len(args) > 1 {
			raw, err := readStringArg(args, 1, "subscribe-topic")
			if err != nil {
				panic(core.RT.NewError(err.Error()))
			}
			if strings.TrimSpace(raw) != "" && raw != "{}" {
				if err := json.Unmarshal([]byte(raw), &opts); err != nil {
					panic(core.RT.NewError(err.Error()))
				}
			}
		}
		id, err := bridge.Funcs.SubscribeTopic(ctx, pattern, opts)
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.MakeString(id)
	})

	register("__gi-read-topic-subscription", func(args []core.Object) core.Object {
		if bridge.Funcs.ReadTopicSubscription == nil {
			panic(core.RT.NewError("read topic subscription is not available"))
		}
		id, err := readStringArg(args, 0, "read-topic-subscription")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		limit := 10
		if len(args) > 1 {
			if raw, err := readStringArg(args, 1, "read-topic-subscription"); err == nil && strings.TrimSpace(raw) != "" {
				if parsed, parseErr := strconv.Atoi(raw); parseErr == nil && parsed > 0 {
					limit = parsed
				}
			}
		}
		events, err := bridge.Funcs.ReadTopicSubscription(ctx, id, limit)
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		b, _ := json.Marshal(events)
		return core.MakeString(string(b))
	})

	register("__gi-unsubscribe-topic", func(args []core.Object) core.Object {
		if bridge.Funcs.UnsubscribeTopic == nil {
			panic(core.RT.NewError("unsubscribe topic is not available"))
		}
		id, err := readStringArg(args, 0, "unsubscribe-topic")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		if err := bridge.Funcs.UnsubscribeTopic(ctx, id); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.NIL
	})

	register("__gi-list-messages", func(args []core.Object) core.Object {
		if bridge.Funcs.ListMessages == nil {
			panic(core.RT.NewError("list messages is not available"))
		}
		limit := 50
		if len(args) > 0 {
			payload, err := readStringArg(args, 0, "list-messages")
			if err != nil {
				panic(core.RT.NewError(err.Error()))
			}
			if strings.TrimSpace(payload) != "" && payload != "{}" {
				var spec struct {
					Limit int `json:"limit"`
				}
				if err := json.Unmarshal([]byte(payload), &spec); err != nil {
					panic(core.RT.NewError(err.Error()))
				}
				if spec.Limit > 0 {
					limit = spec.Limit
				}
			}
		}
		msgs, err := bridge.Funcs.ListMessages(ctx, limit)
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		b, _ := json.Marshal(msgs)
		return core.MakeString(string(b))
	})

	register("__gi-open-raw-socket", func(args []core.Object) core.Object {
		if bridge.Funcs.OpenRawSocket == nil {
			panic(core.RT.NewError("open raw socket is not available"))
		}
		payload, err := readStringArg(args, 0, "open-raw-socket")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		spec := RawSocketSpec{}
		if err := json.Unmarshal([]byte(payload), &spec); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		id, err := bridge.Funcs.OpenRawSocket(ctx, spec)
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.MakeString(id)
	})

	register("__gi-write-raw-socket", func(args []core.Object) core.Object {
		if bridge.Funcs.WriteRawSocket == nil {
			panic(core.RT.NewError("write raw socket is not available"))
		}
		payload, err := readStringArg(args, 0, "write-raw-socket")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		var spec RawSocketPayload
		if err := json.Unmarshal([]byte(payload), &spec); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		n, err := bridge.Funcs.WriteRawSocket(ctx, spec)
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.MakeInt(n)
	})

	register("__gi-read-raw-socket", func(args []core.Object) core.Object {
		if bridge.Funcs.ReadRawSocket == nil {
			panic(core.RT.NewError("read raw socket is not available"))
		}
		payload, err := readStringArg(args, 0, "read-raw-socket")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		var spec RawSocketPayload
		if err := json.Unmarshal([]byte(payload), &spec); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		msg, err := bridge.Funcs.ReadRawSocket(ctx, spec)
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.MakeString(msg)
	})

	register("__gi-close-raw-socket", func(args []core.Object) core.Object {
		if bridge.Funcs.CloseRawSocket == nil {
			panic(core.RT.NewError("close raw socket is not available"))
		}
		socketID, err := readStringArg(args, 0, "close-raw-socket")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		if err := bridge.Funcs.CloseRawSocket(ctx, socketID); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.NIL
	})

	register("__gi-open-websocket", func(args []core.Object) core.Object {
		if bridge.Funcs.OpenWebSocket == nil {
			panic(core.RT.NewError("open websocket is not available"))
		}
		payload, err := readStringArg(args, 0, "open-websocket")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		var spec WebSocketSpec
		if err := json.Unmarshal([]byte(payload), &spec); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		id, err := bridge.Funcs.OpenWebSocket(ctx, spec)
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.MakeString(id)
	})

	register("__gi-write-websocket", func(args []core.Object) core.Object {
		if bridge.Funcs.WriteWebSocket == nil {
			panic(core.RT.NewError("write websocket is not available"))
		}
		socketID, err := readStringArg(args, 0, "write-websocket")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		payload, err := readStringArg(args, 1, "write-websocket")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		if err := bridge.Funcs.WriteWebSocket(ctx, socketID, payload); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.NIL
	})

	register("__gi-read-websocket", func(args []core.Object) core.Object {
		if bridge.Funcs.ReadWebSocket == nil {
			panic(core.RT.NewError("read websocket is not available"))
		}
		socketID, err := readStringArg(args, 0, "read-websocket")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		timeoutMS := 0
		if len(args) > 1 {
			if timeout, ok := args[1].(core.Int); ok {
				timeoutMS = timeout.I
			} else if timeout, ok := args[1].(core.String); ok {
				if parsed, parseErr := strconv.Atoi(timeout.S); parseErr == nil {
					timeoutMS = parsed
				}
			}
		}
		msg, err := bridge.Funcs.ReadWebSocket(ctx, socketID, timeoutMS)
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.MakeString(msg)
	})

	register("__gi-close-websocket", func(args []core.Object) core.Object {
		if bridge.Funcs.CloseWebSocket == nil {
			panic(core.RT.NewError("close websocket is not available"))
		}
		socketID, err := readStringArg(args, 0, "close-websocket")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		if err := bridge.Funcs.CloseWebSocket(ctx, socketID); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return core.NIL
	})

	register("__gi-http-request", func(args []core.Object) core.Object {
		if bridge.Funcs.DoHTTPRequest == nil {
			panic(core.RT.NewError("http request is not available"))
		}
		payload, err := readStringArg(args, 0, "http-request")
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		var req HTTPCallSpec
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		resp, err := bridge.Funcs.DoHTTPRequest(ctx, req)
		if err != nil {
			panic(core.RT.NewError(err.Error()))
		}
		return httpResponseToJokerMap(resp)
	})
}

func readStringArg(args []core.Object, index int, fnName string) (string, error) {
	if len(args) <= index {
		return "", fmt.Errorf("%s requires argument %d", fnName, index+1)
	}
	if s, ok := args[index].(core.String); ok {
		return s.S, nil
	}
	return "", fmt.Errorf("%s argument %d must be a string", fnName, index+1)
}

func httpResponseToJokerMap(resp HTTPResponse) core.Map {
	statusCode := core.MakeInt(resp.StatusCode)
	status := core.MakeString(resp.Status)
	headers := core.EmptyArrayMap()
	for key, values := range resp.Headers {
		valueList := core.EmptyArrayVector()
		for _, value := range values {
			valueList.Append(core.MakeString(value))
		}
		headers.Set(core.MakeString(key), valueList)
	}
	result := core.EmptyArrayMap()
	result.Set(core.MakeKeyword("status_code"), statusCode)
	result.Set(core.MakeKeyword("status"), status)
	result.Set(core.MakeKeyword("headers"), headers)
	result.Set(core.MakeKeyword("body"), core.MakeString(resp.Body))
	result.Set(core.MakeKeyword("url"), core.MakeString(resp.URL))
	return result
}

func buildEmbeddedPreamble(bridgeJSON string) string {
	return fmt.Sprintf("(require '[joker.json :as json] '[joker.walk :as walk])\n(def ^:private *gi-bridge* (walk/keywordize-keys (json/read-string %q)))\n(def ^:private *gi-session-state* (atom (or (:session-state *gi-bridge*) {})))\n(defn gi-get-session-state [] @*gi-session-state*)\n(defn gi-set-session-state! [patch] (swap! *gi-session-state* merge patch))\n(defn gi-get-session-info [] (:session-info *gi-bridge*))\n(defn gi-get-runtime-config [] (:runtime-config *gi-bridge*))\n(defn gi-list-turns [] (:turns *gi-bridge*))\n(defn gi-list-messages\n  ([]\n   (json/read-string (__gi-list-messages \"{}\")))\n  ([opts]\n   (json/read-string (__gi-list-messages (json/write-string opts))))\n)\n(defn gi-register-event-hook [spec] (__gi-register-event-hook (json/write-string spec)))\n(defn gi-emit-event [name payload] (__gi-emit-event name (if payload (json/write-string payload) \"{}\")))\n(defn gi-clear-event-hooks [] (__gi-clear-event-hooks))\n(defn gi-register-tool [spec] (__gi-register-tool (json/write-string spec)))\n(defn gi-set-active-tools [names] (__gi-set-active-tools (json/write-string names)))\n(defn gi-register-route [spec] (json/read-string (__gi-register-route (json/write-string spec))))\n(defn gi-unregister-route [id] (__gi-unregister-route (str id)))\n(defn gi-list-routes ([] (json/read-string (__gi-list-routes \"{}\"))) ([filter] (json/read-string (__gi-list-routes (json/write-string filter)))))\n(defn gi-emit-connectivity-event [topic payload] (__gi-emit-connectivity-event (str topic) (if payload (json/write-string payload) \"{}\")))\n(defn gi-topic-publish [envelope] (__gi-publish-topic (if envelope (json/write-string envelope) \"{}\")))\n(defn gi-topic-subscribe ([pattern] (__gi-subscribe-topic (str pattern) \"{}\")) ([pattern opts] (__gi-subscribe-topic (str pattern) (if opts (json/write-string opts) \"{}\"))))\n(defn gi-topic-read ([subscription-id] (json/read-string (__gi-read-topic-subscription (str subscription-id) \"10\"))) ([subscription-id limit] (json/read-string (__gi-read-topic-subscription (str subscription-id) (str (or limit 10))))))\n(defn gi-topic-unsubscribe [subscription-id] (__gi-unsubscribe-topic (str subscription-id)))\n(defn gi-open-raw-socket [spec] (__gi-open-raw-socket (if spec (json/write-string spec) \"{}\")))\n(defn gi-write-raw-socket [payload] (__gi-write-raw-socket (if payload (json/write-string payload) \"{}\")))\n(defn gi-read-raw-socket [payload] (__gi-read-raw-socket (if payload (json/write-string payload) \"{}\")))\n(defn gi-close-raw-socket [socket-id] (__gi-close-raw-socket (str socket-id)))\n(defn gi-open-websocket [spec] (__gi-open-websocket (if spec (json/write-string spec) \"{}\")))\n(defn gi-write-websocket [socket-id payload] (__gi-write-websocket (str socket-id) payload))\n(defn gi-read-websocket [socket-id timeout-ms] (__gi-read-websocket (str socket-id) (or timeout-ms 0)))\n(defn gi-close-websocket [socket-id] (__gi-close-websocket (str socket-id)))\n(defn gi-http-request [request] (__gi-http-request (if request (json/write-string request) \"{}\")))", bridgeJSON)
}
