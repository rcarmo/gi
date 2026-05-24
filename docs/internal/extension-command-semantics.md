# Extension command registration semantics

Status: design contract for the next Gi TUI/runtime slices.

This note defines how Pi-like extension commands should work in Gi without changing the current SQLite-backed runtime contract or the existing startup extension loader.

## Goals

- Let Joker, JavaScript, and process extensions expose TUI commands with predictable names.
- Keep command handling separate from model prompts unless a command explicitly submits work.
- Keep command registration observable through `/plugins`, topics, and future reload output.
- Preserve existing script/tool/hook contracts; extension commands are a thin command-dispatch surface, not a replacement for tools.

## Command names

Extensions register commands by name, without the leading slash.

Recommended shape:

```json
{
  "name": "foo",
  "description": "Run foo",
  "usage": "/foo [args]",
  "source": "extension-id-or-path",
  "engine": "joker|js|process"
}
```

Rules:

- Names are case-insensitive for lookup and stored in lowercase.
- Names may contain letters, numbers, `_`, and `-`.
- Names must not include `/` or whitespace.
- Built-in TUI commands take precedence over extension commands.
- If two extensions register the same command, the first loaded command wins and later registrations are reported as conflicts.
- Extensions that want namespacing should prefer `prefix-command`, for example `jira-open`.

## Invocation semantics

When the TUI sees `/name ...` and no built-in command matches:

1. Look up an extension command named `name`.
2. Pass the raw argument string and parsed argv to the handler.
3. The handler returns one of these outcomes:
   - `message`: append terminal-safe lines to the transcript.
   - `submit`: submit a prompt/turn through the existing runtime queue.
   - `error`: append an error line and publish an extension error topic.
   - `noop`: no transcript change beyond optional status text.

Command handlers must not mutate session state by editing JSON directly. They should call existing Gi bridge APIs (`gi.state`, `gi.runtime`, `gi.topics`) once those namespaces are documented and exposed for this surface.

## Engine-specific registration

### JavaScript

Future JS bridge shape:

```js
gi.commands.register({
  name: "demo",
  description: "Demo command",
  usage: "/demo [text]"
}, (ctx) => {
  return { type: "message", lines: [`demo: ${ctx.args}`] };
});
```

### Joker

Future Joker bridge shape:

```clojure
(gi-command-register
  {:name "demo"
   :description "Demo command"
   :usage "/demo [text]"}
  (fn [ctx]
    {:type "message" :lines [(str "demo: " (:args ctx))]}))
```

### Process extensions

Process extensions should use the same mounted JSON-RPC process model as process hooks. Registration is part of the process hello/capabilities response, and invocation uses a stable method name:

- `command.invoke`

Request params should include:

```json
{
  "name": "demo",
  "args": "raw text",
  "argv": ["raw", "text"],
  "session_id": "...",
  "agent_id": "..."
}
```

Response shape should mirror the in-process result object:

```json
{
  "type": "message|submit|error|noop",
  "lines": ["optional transcript lines"],
  "prompt": "optional prompt for submit",
  "status": "optional status text"
}
```

## Lifecycle and reload

- Registrations are process-local and loaded during extension startup.
- `/reload` may refresh config/skill discovery first, but command registration cleanup requires a safe extension reload lifecycle before live handler replacement.
- Until safe reload exists, command changes require process restart and should be reported by `/reload` as unchanged mounted extension state.

## Observability

Registration and conflicts should publish topic notices under `extension.command` with payloads like:

```json
{
  "type": "registered|conflict|invoked|failed",
  "command": "demo",
  "engine": "joker",
  "source": ".gi/extensions/demo.joke"
}
```

`/plugins` should eventually include registered commands under each loaded extension.

## Security and safety

- Extension commands run with the same trust boundary as the extension that registered them.
- Command output must be transcript-safe text; rich TUI widgets are out of scope for this first surface.
- Commands that execute shell/process work must use existing tool/runtime approval paths where practical instead of bypassing them.

## Current implementation status

Defined here only. Gi currently supports startup extension loading, tool/hook registration, topics, `/plugins`, and `/skill:name`; extension command dispatch is the next implementation slice after this contract.
