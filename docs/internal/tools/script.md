# Tool: `script`

## Status
Implemented.

## Purpose
Execute a script with access to gi live session state, session info, runtime config, and workspace file helpers.

## Supported engines
- `js` — Goja JavaScript runtime compiled into gi
- `joker` — Joker/Clojure runtime baked into gi via vendored generated sources
- `quickjs` is intentionally out of scope in the current track; use `js` or `joker` instead.

## Input
```json
{
  "script": "40 + 2",
  "path": "optional/script.js",
  "engine": "js",
  "session_id": "optional-session-id"
}
```

### Fields
- `script` — inline script source
- `path` — workspace-relative script path or `vfs://namespace/path`
- `engine` — `js` or `joker`; auto-detected from file extension when omitted
- `session_id` — session used to build the bridge context

Either `script` or `path` is required.

## Output
```json
{
  "result": "42",
  "error": ""
}
```

### Fields
- `result` — textual script result or captured console output
- `error` — optional error string

## Engine selection
Current resolution order:
1. explicit `engine`
2. `.joke` / `.clj` -> `joker`
3. `.js` -> `js`
4. inline default -> `js`

## Live bridge highlights
Current bridge capabilities include:
- session introspection via session info/state helpers
- controlled session-state mutation
- runtime config access
- turn listing
- message listing
- workspace file reads/writes/listing

## File access semantics
Script execution input and bridge file helpers resolve using the shared resolver and support both:
- workspace-relative paths
- `vfs://namespace/path`

This includes read/list/write operations through mutable namespaces and read-only `vfs://reference/...`.

## See also
- `../scripting/README.md`
- `../scripting/contract.md`
- `../scripting/joker.md`
- `../scripting/bridge.md`
