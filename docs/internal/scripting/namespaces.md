# Gi scripting namespaces

Status: current bridge reference plus near-term extension-author contract.

Gi exposes a small host bridge to JavaScript (Goja), Joker, and process-style extensions. The bridge is intentionally data-oriented: SQLite/store state remains the durable source of truth, while in-memory handles such as running turns, mounted hooks, and subscriptions remain process-local.

## Namespace overview

| Namespace | Purpose | JavaScript surface | Joker surface |
| --- | --- | --- | --- |
| `gi.state` | Session-scoped state and transcript/store reads | Current compatibility methods plus planned grouped aliases | `gi-get-session-state`, `gi-set-session-state!`, `gi-get-session-info`, etc. |
| `gi.topics` | Process-local canonical topic bus publish/subscribe/read/unsubscribe | Implemented as `gi.topics.*` | Implemented as `gi-topic-*` helpers |
| `gi.runtime` | Runtime configuration, active session/turn metadata, and future command/runtime helpers | Current compatibility methods plus planned grouped aliases | `gi-get-runtime-config`, turn/message helpers |

The current JavaScript bridge still exposes several top-level compatibility methods (`gi.getSessionState()`, `gi.setSessionState()`, `gi.getRuntimeConfig()`, etc.). Extension authors should treat the grouped namespaces below as the intended documentation model even where a compatibility method is the shipped entrypoint today.

## `gi.state`

Purpose: read and patch session-scoped durable state without bypassing the store APIs.

Current JavaScript compatibility methods:

```js
gi.getSessionState();
gi.setSessionState({ key: "value" });
gi.getSessionInfo();
gi.listMessages(50);
gi.listTurns(20);
```

Current Joker helpers:

```clojure
(gi-get-session-state)
(gi-set-session-state! {:key "value"})
(gi-get-session-info)
(gi-list-messages {:limit 50})
(gi-list-turns {:limit 20})
```

Rules:

- State patches are merged through the runtime bridge; do not edit `sessions.state_json` directly.
- Session identity, channel bindings, routing allocation, and turn coordination tables are not part of `gi.state`; use runtime APIs or tools when available.
- Reads are scoped to the current script/session context.

Planned grouped JavaScript aliases:

```js
gi.state.get();
gi.state.patch({ key: "value" });
gi.state.info();
gi.state.messages({ limit: 50 });
gi.state.turns({ limit: 20 });
```

## `gi.topics`

Purpose: publish and consume process-local canonical topic envelopes for runtime observability and extension-to-extension fan-out.

Current JavaScript surface:

```js
gi.topics.publish({
  topic: "runtime.example",
  type: "notice",
  payload: { ok: true }
});

const sub = gi.topics.subscribe("runtime.*", { after_sequence: 0 });
const events = gi.topics.read(sub, 10);
gi.topics.unsubscribe(sub);
```

Current Joker surface:

```clojure
(gi-topic-publish {:topic "runtime.example" :type "notice" :payload {:ok true}})
(def sub (gi-topic-subscribe "runtime.*" {:after_sequence 0}))
(def events (gi-topic-read sub 10))
(gi-topic-unsubscribe sub)
```

Rules:

- Topic subscriptions are polling handles, not long-lived script callbacks.
- Handles are process-local and must be unsubscribed when no longer needed.
- Topic envelopes use monotonic process-local sequences; sequence gaps indicate missed in-memory events, not durable replay failure.
- Script-facing topic operations are session-bound when a current session exists. Cross-session publish/subscribe attempts should be rejected by the host bridge.
- Durable replay belongs to SQLite/store APIs, not to `gi.topics`.

Recommended topic names:

- `extension.<name>` for extension-owned notices.
- `extension.command` for future extension command registration/invocation notices.
- Existing runtime families such as `runtime.turn`, `runtime.session`, `runtime.tool`, `runtime.hook`, `session.steering`, and `turn.subturn` should be treated as runtime-owned.

## `gi.runtime`

Purpose: inspect runtime configuration and, over time, expose safe runtime operations that are not simple state patches.

Current JavaScript compatibility methods:

```js
gi.getRuntimeConfig();
gi.runtimeConfig;
gi.config;
gi.sessionId;
```

Current Joker helpers/data:

```clojure
(gi-get-runtime-config)
*gi-bridge* ; contains session/config/runtime-config data
```

Current adjacent helper families exposed through the bridge include:

- file helpers: `readFile`, `writeFile`, `listDir` / workspace-relative and `vfs://` paths,
- event hooks: `registerEventHook`, `emitEvent`, `clearEventHooks`,
- networking: raw sockets, websocket, HTTP request helpers,
- logging: `gi.log(...)` / console methods in JavaScript.

Planned grouped JavaScript aliases:

```js
gi.runtime.config();
gi.runtime.sessionId();
gi.runtime.registerHook(spec);
gi.runtime.emitEvent(name, payload);
gi.runtime.log("info", "message");
```

Rules:

- Runtime operations must preserve the existing tool/hook/API contracts.
- Long-running hooks or commands block the agent loop unless they explicitly use a process-backed surface designed for long-lived work.
- Process extension commands should use the mounted JSON-RPC process model described in `../extension-command-semantics.md`.

## Process extensions

Process extensions do not run inside the Goja/Joker bridge. They should receive equivalent context through JSON-RPC request params:

```json
{
  "session_id": "...",
  "agent_id": "...",
  "runtime_config": {},
  "session_state": {},
  "payload": {}
}
```

Process hooks and future process commands should return JSON objects that map back onto the same runtime `HookResponse` or command result contracts documented elsewhere. They should not assume direct access to in-process `gi.*` objects.

## Compatibility note

This document names the stable conceptual namespaces for extension authors. Some aliases are implemented today as top-level JavaScript methods or Joker helper functions rather than nested objects. New code should prefer the documented semantics, and implementation slices should add grouped aliases without removing compatibility names.
