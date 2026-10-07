# Joker runtime

## Status
Implemented.

## Summary
Joker is gi's Clojure scripting runtime.

In gi, Joker is treated as a baked-in runtime surface:
- no external `joker` executable is required at deployment time
- gi imports the released `github.com/rcarmo/go-joker/v42 v42.12.4` module, including its generated runtime files
- script execution runs in-process inside gi

## Engine name
`joker`

## Source forms
- inline script text
- workspace-relative `.joke` / `.clj` files

## Bridge state
The current bridge state is injected as `*gi-bridge*`.

Current injected keys include:
- `:session-id`
- `:config` when available
- `:runtime-config` (alias of runtime config)
- `:session-state` when available
- `:session-info` when available
- `:turns` and `:messages` when their respective bridge callbacks are configured

## Live helper functions
The current Joker preamble exposes:
- `gi-get-session-state`
- `gi-set-session-state!`
- `gi-get-session-info`
- `gi-get-runtime-config`
- `gi-list-turns`
- `gi-list-messages` (optional `:limit` map arg)

Session-state mutations are applied back through the live bridge after execution.

## Output model
The runtime executes the script inside a wrapper that emits structured JSON containing:
- the script result
- the final session-state view

The final value of the script body becomes the returned textual result.

## Native compilation

Joker itself runs as native Go in Gi. `joker.jit/compile-wasm` compiles eligible user functions to WASM for machine-code execution through Joker's wazero runtime. The whole Joker runtime is never a WASI guest. The numeric-kernel test returns 42 from a compiled function, rejects unsupported string-building code and verifies ordinary native evaluation still works afterwards. Integer/float eligibility and fallbacks outside explicit `compile-wasm` follow the pinned Joker release.

## WASM execution selection

Set `JOKER_WASM_ENGINE` **before starting Gi**:

- `compiler` (or `native`): compile user-function WASM to host machine code; fail if platform or executable-memory support is unavailable. No silent interpreter fallback.
- `interpreter`: interpret the generated function's WASM.
- `auto` (default): prefer compilation, fall back to interpretation when unavailable.

The shared Joker WASM runtime initialises once. Changing the variable after the first JIT call does not switch an existing runtime. `(jit/wasm-engine)` reports the actual selection. This environment setting affects native Joker's generated-function execution only; it does not select goja/QuickJS or move Joker itself into an isolate.

```sh
JOKER_WASM_ENGINE=compiler gi
```

```clojure
(let [sum (jit/compile-wasm
            (fn [n] (loop [i 0 total 0]
                      (if (< i n)
                        (recur (+ i 1) (+ total i)) total))))]
  [(sum 100) (jit/wasm-engine)])
;; => [4950 "compiler"] with the explicit compiler startup setting
```

`make test-joker-wasm-engines` runs Gi's emitted-kernel tests in fresh native test processes for each requested mode, repeating three times with race detection where supported. [v42.12.2 verification](joker-v42.12.2.md) records scope and profile limitations. The upstream standalone `joker compile --native --run` command bundles a native Joker executable and source; Gi does not expose that CLI through its script tool.

## Checked buffer kernels

The `joker.jit` bridge also exposes `numeric-buffer`, `buffer-get`, `buffer-set!` and `compile-wasm` with explicit `{:buffers [argument-indices]}`. This supports bounded f64 kernels and nested loops without importing host functions into the module. [v42.12.4 verification and example](joker-v42.12.4.md) cover native execution, aliasing, bounds traps and pre-trap mutation copy-back. The scalar compiler is unchanged when no buffer options are given.

## Current file semantics
Script loading uses the workspace/VFS resolver; both workspace files and `vfs://` scripts are supported.

## Error behavior
Joker errors are returned as textual tool errors prefixed with `joker:`.
