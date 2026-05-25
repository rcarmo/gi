# Scripting bridge

## Status
Implemented for path resolution (`readFile`, `writeFile`, `listDir`) with shared resolver and `vfs://` support.

Implemented: event hooks, raw sockets, websocket transport, and HTTP request APIs are now functional in the script bridge for JavaScript (Goja) via `buildBridge` in `ScriptTool`.

Networking and transport callbacks are now exercised in `internal/tools/script_test.go`.

Execution remains implemented for JS via Goja; Joker execution is also active for `.joke`/`.clj` scripts via the Joker runtime. (Shell execution is still unavailable in the scripting bridge.)

## Purpose
The scripting bridge is the controlled host API exposed to script runtimes.

## Core bridge model
Every engine receives a `Bridge` with:
- `SessionID`
- `BridgeFuncs` host callbacks

Current callback families:
- session state
  - `GetSessionState`
  - `SetSessionState`
  - `GetSessionInfo`
- turns and turn events
  - `ListTurns`
  - `ListTurnEvents`
- messages
  - `ListMessages`
  - `AddMessage`
- turn events
  - `ListTurnEvents`
- runtime config
  - `GetConfig`
- files
  - `ReadFile`
  - `WriteFile`
  - `ListDir`
- shell execution
  - `Exec` (currently stubbed / not available)
- event hooks
  - `RegisterEventHook`
  - `EmitEvent`
  - `ClearEventHooks`
- raw sockets
  - `OpenRawSocket`
  - `WriteRawSocket`
  - `ReadRawSocket`
  - `CloseRawSocket`
- websocket transport
  - `OpenWebSocket`
  - `WriteWebSocket`
  - `ReadWebSocket`
  - `CloseWebSocket`
- HTTP request interface
  - `DoHTTPRequest` (method, URL, full header map, timeout, and body)
- logging
  - `Log`

## JavaScript bridge surface
The Goja runtime exposes a global `gi` object with the methods listed below, including active network/event bridges. See `namespaces.md` for the extension-author view of the conceptual `gi.state`, `gi.topics`, and `gi.runtime` namespaces; several shipped APIs still exist as top-level compatibility methods.
- `gi.sessionId`
- `gi.config`
- `gi.sessionState`
- `gi.sessionInfo`
- `gi.runtimeConfig`
- `gi.getSessionState()`
- `gi.setSessionState(patch)`
- `gi.getSessionInfo()`
- `gi.getRuntimeConfig()`
- `gi.listTurns(limit)`
- `gi.listMessages(limit)`
- `gi.readFile(path)`
- `gi.writeFile(path, content)`
- `gi.listDir(path)`
- `gi.registerEventHook({name, source, filter, arguments})`
- `gi.emitEvent(name, payload)`
- `gi.clearEventHooks()`
- `gi.net.openRawSocket({protocol, address, timeout_ms, local_addr}) -> socketId`
- `gi.net.writeRawSocket({socket_id, data, timeout_ms}) -> bytes`
- `gi.net.readRawSocket({socket_id, max_bytes, timeout_ms})`
- `gi.net.closeRawSocket(socketId)`
- `gi.websocket.open({url, headers, subprotocol, timeout_ms}) -> socketId`
- `gi.websocket.write(socketId, payload)`
- `gi.websocket.read(socketId, timeout_ms)`
- `gi.websocket.close(socketId)`
- `gi.http.request({method, url, headers, body, timeout_ms, retry, allow_redirects, skip_tls}) -> {status_code, status, headers, body}`
- `gi.log(level, message)`
- `gi.topics.publish(envelope)`
- `gi.topics.subscribe(pattern, opts)`
- `gi.topics.read(subscriptionId, max)`
- `gi.topics.unsubscribe(subscriptionId)`

It also provides `console.log/error/warn/debug`.

## Joker bridge surface
Joker currently receives a data-oriented binding via `*gi-bridge*` plus helper functions defined in the execution preamble.

Current injected data:
- session id
- config
- runtime-config
- session state
- session info
- turns
- messages (when the tool provides them)

Current helper functions:
- `gi-get-session-state`
- `gi-set-session-state!`
- `gi-get-session-info`
- `gi-get-runtime-config`
- `gi-list-turns`
- `gi-list-messages` (optional `:limit` map arg)
- `gi-topic-publish`
- `gi-topic-subscribe`
- `gi-topic-read`
- `gi-topic-unsubscribe`

The current Joker mutation path applies session-state changes back through the live bridge after execution.

## Path semantics
All file helpers resolve through the shared path resolver and accept both:
- workspace-relative paths
- `vfs://...` URLs

Read-only semantics are enforced for `vfs://reference/...`.

