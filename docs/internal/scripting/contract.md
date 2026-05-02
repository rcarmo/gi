# Script bridge contract (JS + Clojure/Joker)

This document is the canonical machine-friendly contract for gi script execution and bridging APIs.

- Tool: `script`
- Engines: `js` (Goja) and `joker` (embedded Joker/Clojure)
- `quickjs` is intentionally out of scope for current releases; `js`/`joker` are the supported engines.
- Engine selection: explicit `engine`, then file extension, otherwise inline defaults to `js`

---

## 1) Tool input/output contract

### Input (`script` tool)
- `script` (string) — inline script source
- `path` (string) — workspace-relative path or `vfs://namespace/path`
- `engine` ("js" | "joker") — optional
- `session_id` (string) — optional

Exactly one of `script` or `path` is required.

### Output
- `result` (string)
- `error` (string, optional)

The JSON Schema and TypeScript shapes are available here:
- [`contract.schema.json`](./contract.schema.json)
- [`contract.types.ts`](./contract.types.ts)

---

## 2) Goja (`js`) bridge shape

A global object `gi` is injected in script context.

### Core
- `gi.sessionId: string`
- `gi.config: object`
- `gi.runtimeConfig: object`
- `gi.sessionState: object`

### Helpers
- `gi.getSessionState(): object`
- `gi.setSessionState(patch: object): void`
- `gi.getSessionInfo(): object`
- `gi.getRuntimeConfig(): object`
- `gi.listTurns(limit?: number): object[]`
- `gi.listMessages(limit?: number): object[]`

### File helpers
- `gi.readFile(path: string): string`
- `gi.writeFile(path: string, content: string): void`
- `gi.listDir(path: string): object[]`

### Event hooks
- `gi.registerEventHook(spec: EventHookSpec): void`
- `gi.emitEvent(name: string, payload?: object): void`
- `gi.clearEventHooks(): void`

### Raw sockets
- `gi.net.openRawSocket(spec: RawSocketSpec): string`
- `gi.net.writeRawSocket(payload: RawSocketPayload): number`
- `gi.net.readRawSocket(payload: RawSocketPayload): string`
- `gi.net.closeRawSocket(socketId: string): void`

### WebSocket
- `gi.websocket.open(spec: WebSocketSpec): string`
- `gi.websocket.write(socketId: string, payload: string): void`
- `gi.websocket.read(socketId: string, timeout_ms?: number): string`
- `gi.websocket.close(socketId: string): void`

### HTTP
- `gi.http.request(req: HTTPRequestSpec): HTTPResponse`

### Logging
- `gi.log(level: string, message: string): void`
- console methods exist: `console.log/error/warn/debug`

---

## 3) Clojure/Joker bridge shape

Joker executes with a preloaded preamble that injects helper functions in the current namespace.

### Session/state/config
- `gi-get-session-state` -> map
- `gi-set-session-state!` (patch map) -> nil
- `gi-get-session-info` -> map
- `gi-get-runtime-config` -> map
- `gi-list-turns` -> list
- `gi-list-messages` (optional `{:limit n}`) -> list

### Events
- `gi-register-event-hook` (spec map) -> nil
- `gi-emit-event` (name string, payload map) -> nil
- `gi-clear-event-hooks` -> nil

### Raw sockets
- `gi-open-raw-socket` (spec map) -> string socket-id
- `gi-write-raw-socket` (payload map) -> number
- `gi-read-raw-socket` (payload map) -> string
- `gi-close-raw-socket` (socket-id) -> nil

### WebSocket
- `gi-open-websocket` (spec map) -> string socket-id
- `gi-write-websocket` (socket-id, payload-string) -> nil
- `gi-read-websocket` (socket-id, timeout-ms) -> string
- `gi-close-websocket` (socket-id) -> nil

### HTTP
- `gi-http-request` (request map) -> response map:
  - `{:status_code <int>, :status <string>, :headers <map>, :body <string>, :url <string>}`

---

## 4) Notes

- Path resolution for file helpers uses the shared resolver.
- `vfs://reference/...` paths remain read-only for write operations.
- `session_id` is required at runtime for session/state-sensitive scripts.
