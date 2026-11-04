# Tool: `script`

Execute a script with access to gi's live session state, session info, runtime
config, turn history and workspace or managed VFS files.

## Engines and input

- `goja`: native JavaScript; synchronous bridge helpers.
- `quickjs`: isolated JavaScript; await bridge helpers, including nested calls.
- `js`: configured `javascriptRuntime`, defaulting to Goja.
- `joker`: native Joker/Clojure (also auto-selected for `.joke`/`.clj` files).

Supply inline `script` or a script `path` (workspace-relative or `vfs://...`).
An explicit `engine` wins; other files and inline scripts use the configured
JavaScript runtime. Unknown engine names fail without fallback. The host
supplies the current session to session-sensitive helpers.

```json
{
  "engine": "goja",
  "script": "console.log(JSON.stringify(gi.listDir('vfs://reference/')));"
}
```

## JavaScript bridge names

Helpers are properties of the global **`gi`** object, not bare globals:
`gi.getSessionInfo()`, `gi.getSessionState()`, `gi.setSessionState(patch)`,
`gi.getRuntimeConfig()`, `gi.listTurns(limit)`, `gi.listTurnEvents(turnID)`,
`gi.listMessages(limit)`, `gi.readFile(path)`, `gi.writeFile(path, content)`,
`gi.listDir(path)`, and `gi.log(level, message)`. Use `console.log()` for
captured tool output. Bare `listDir()` and `log()` are not provided.

QuickJS example:

```javascript
(async () => {
  const entries = await gi.listDir("vfs://reference/");
  console.log(JSON.stringify(entries));
})()
```

## File access

`vfs://reference/...` is a read-only virtual reference tree, not an OS
filesystem directory. Access it using the `read` tool or `gi.readFile` and
`gi.listDir`; shell `grep`, `find` or `ls` cannot resolve a `vfs://` URL.
Mutable namespaces support writes through the same shared path resolver.

## Failure diagnosis

Inspect turn status with `gi.listTurns()`, then use
`gi.listTurnEvents(turnID)` to inspect the recorded `inference.failed`,
`turn.failure_marked` and `tool.failed` events. Events are ordered by sequence;
only turns in the current session are accepted. An existing turn with no
events returns `[]`; missing/foreign turn IDs fail. Goja and QuickJS expose
this helper; there is no corresponding Joker helper yet.

A tool error need not be the error that terminated a turn. In particular,
provider HTTP 400 validation errors can arise from request serialization or
replayed history after tools have finished. A successful tool retry or a
model switch is not proof that the provider error was fixed. Inspect only
relevant events; do not dump credentials, full runtime config or raw request
payloads into chat.

## Output

`result` is the textual completion value or captured console output; `error`
is an optional execution error. Console output takes precedence over the
completion value. Partial console output may accompany an error.

See the [runtime overview](../scripting/README.md),
[contract](../scripting/contract.md), [Joker](../scripting/joker.md) and
[bridge](../scripting/bridge.md) for the full API.
