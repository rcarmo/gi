# Joker runtime

## Status
Implemented.

## Summary
Joker is gi's Clojure scripting runtime.

In gi, Joker is treated as a baked-in runtime surface:
- no external `joker` executable is required at deployment time
- gi imports the released `github.com/rcarmo/go-joker/v42 v42.12.1` module, including its generated runtime files
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

## Current file semantics
Script loading uses the workspace/VFS resolver; both workspace files and `vfs://` scripts are supported.

## Error behavior
Joker errors are returned as textual tool errors prefixed with `joker:`.
