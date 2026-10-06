# Scripting runtimes

Gi runs Joker natively and offers two internal JavaScript runtimes: native Goja and isolated QuickJS under wazero. No external interpreter is required. [Upgrade verification and limits](runtime-upgrade-2026-10-06.md) record the pinned versions and tested scope.

## Selection

The `script` tool's `engine` field accepts:

- `joker`: native embedded go-joker/v42 for `.joke` or `.clj` code.
- `goja`: native JavaScript with synchronous bridge calls.
- `quickjs`: a fresh QuickJS/WASM instance with promise-based bridge calls.
- `js` or `javascript`: the configured JavaScript runtime, defaulting to Goja.

Without an explicit engine, `.joke` and `.clj` files select Joker; other files and inline scripts select the configured JavaScript runtime. Set `"javascriptRuntime": "quickjs"` or `"goja"` in user or project `.gi/settings.json` (Pi paths remain a fallback). Project settings override user settings. Unknown runtime names fail; execution never silently switches engines. Script-backed tools, commands and hooks use the same `ScriptTool` selection.

QuickJS exposes the same host-provided `gi` values and helper names as Goja. Functions return promises, including nested helpers such as `gi.http.request` and `gi.topics.publish`. Await them explicitly:

```javascript
(async () => {
  const state = await gi.getSessionState();
  await gi.setSessionState({ runs: (state.runs || 0) + 1 });
  return gi.sessionId;
})()
```

Goja executes user code natively. In the QuickJS adapter, a Goja object marshals the existing Go bridge bindings; it never executes user JavaScript. User code executes only inside QuickJS. Every run gets a fresh guest instance, cancellation and a five-minute default deadline; host calls share that deadline and are joined before returning. Output/result handling follows the script contract: console text takes precedence, otherwise the completion value is returned. No ambient Node process, require or fetch is supplied. Host bridge capabilities remain privileged and operator-owned; choosing QuickJS does not remove authorised file/network/tool access.

## Native Joker compilation

The native Joker host provides `joker.jit/compile-wasm`. Eligible user functions become WASM modules executed through Joker's wazero runtime, which compiles them to machine code on supported hosts. The Joker runtime itself is never built or hosted as a WASI guest. Unsupported compilation returns an error; it is not credited as compiled execution.

```clojure
(require '[joker.jit :as jit])
(let [kernel (jit/compile-wasm (fn [x y] (+ (* x x) y)))]
  (kernel 6 6))
```

Compilation eligibility and integer/float semantics belong to the pinned Joker release. General bridge scripts can use ordinary native Joker evaluation around explicitly compiled kernels. Native Joker remains a trusted scripting surface, not a capability sandbox.

QuickJS is the only whole runtime hosted in a wazero isolate. It also executes MCP codemode through the existing capability-limited codemode host; internal scripts do not gain codemode's MCP tools automatically. See [MCP codemode](../codemode.md), [bridge](bridge.md), [Joker](joker.md), [namespaces](namespaces.md) and [contract](contract.md).
