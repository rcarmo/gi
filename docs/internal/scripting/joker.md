# Joker runtime

## Status
Implemented.

## Summary
Joker is gi's Clojure scripting runtime.

In gi, Joker is treated as a baked-in runtime surface:
- no external `joker` executable is required at deployment time
- gi vendors generated Joker sources locally
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

## Current file semantics
- script file loading is workspace-relative
- future work: route file operations and script paths through the workspace/VFS resolver so `vfs://` becomes first-class

## Error behavior
Joker errors are returned as textual tool errors prefixed with `joker:`.
