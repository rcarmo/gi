# Connectivity hooks and route registration

**Date:** 2026-05-02  
**Scope:** Agent/runtime extension points for outbound connectivity, inbound routes, and event streams. Non-UX.

---

## 1. Why this is a separate surface

Pi's extension model exposes process/file/tool/provider hooks, but gi also needs first-class connectivity primitives because our agents are intended to act as automation endpoints, not just CLI assistants.

There are two distinct needs:

1. **Outbound clients** — scripts/tools initiate connections:
   - TCP/UDP sockets
   - Unix sockets / Windows named pipes
   - WebSocket clients
   - SSE clients
   - HTTP requests
   - MQTT clients

2. **Inbound route registration** — scripts/extensions expose handlers that external systems can call:
   - HTTP routes/webhooks
   - WebSocket routes
   - SSE event streams
   - MQTT topic subscriptions
   - Named pipe/socket listeners
   - Internal event bus routes

The existing script bridge already has a partial outbound surface:

- `gi.net.openRawSocket()` / Joker `gi-open-raw-socket`
- WebSocket open/read/write/close
- HTTP request with header control
- Script event hook registration

The missing layer is a durable, managed **connectivity registry** with lifetimes, routing, auth, backpressure, and safe delivery into the agent loop.

---

## 2. Core model

Connectivity should be modeled as **routes** plus **transports**.

```text
Transport endpoint
  └─ route registration
       └─ matcher
            └─ handler
                 ├─ returns immediate response
                 ├─ emits event
                 ├─ appends message
                 ├─ triggers/queues agent turn
                 └─ streams response/events
```

### Route spec

```go
type ConnectivityRouteSpec struct {
    ID          string         `json:"id,omitempty"`
    Name        string         `json:"name"`
    Transport   string         `json:"transport"` // http|websocket|sse|mqtt|tcp|udp|unix|pipe|event
    Direction   string         `json:"direction"` // inbound|outbound|duplex
    Source      string         `json:"source,omitempty"`
    Match       map[string]any `json:"match,omitempty"`
    Auth        map[string]any `json:"auth,omitempty"`
    Options     map[string]any `json:"options,omitempty"`
    Engine      string         `json:"engine,omitempty"` // js|joker
    Script      string         `json:"script,omitempty"`
    Path        string         `json:"path,omitempty"`
    SessionID   string         `json:"session_id,omitempty"`
    AgentID     string         `json:"agent_id,omitempty"`
    Mode        string         `json:"mode,omitempty"` // observe|message|turn|tool|respond
}
```

### Delivery modes

| Mode | Behavior |
|------|----------|
| `observe` | Handler runs; no agent message is created unless handler explicitly does so. |
| `message` | Handler output is stored as a system/custom message only. |
| `turn` | Handler output is submitted as a user/peer prompt and triggers an agent turn. |
| `tool` | Handler is exposed as a callable tool endpoint. |
| `respond` | Handler returns a transport-native response (HTTP response, WS reply, MQTT publish). |

---

## 3. Registry API

### Go engine API

```go
type ConnectivityRegistry interface {
    RegisterRoute(ctx context.Context, spec ConnectivityRouteSpec) (RouteHandle, error)
    UnregisterRoute(ctx context.Context, id string) error
    ListRoutes(ctx context.Context, filter map[string]any) ([]ConnectivityRouteInfo, error)
    Emit(ctx context.Context, event ConnectivityEvent) error
    Subscribe(ctx context.Context, spec ConnectivitySubscriptionSpec) (<-chan ConnectivityEvent, error)
}
```

### Script API — JS

```js
gi.connect.registerRoute({
  name: "github-webhook",
  transport: "http",
  match: { method: "POST", path: "/hooks/github" },
  mode: "turn",
  engine: "js",
  script: `
    const event = gi.event;
    "GitHub webhook: " + JSON.stringify(event.body)
  `,
});

gi.connect.unregisterRoute("github-webhook");
gi.connect.listRoutes();
```

### Script API — Joker

```clojure
(gi-register-route
  {:name "mqtt-power"
   :transport "mqtt"
   :match {:topic "zigbee/+/power"}
   :mode "message"
   :engine "joker"
   :script "(str \"power event: \" *gi-event*)"})
```

---

## 4. Outbound connectivity APIs

### 4.1 Raw sockets

Already partly present. Extend to listener/route registration.

```js
const id = gi.net.openRawSocket({
  protocol: "tcp",
  address: "127.0.0.1:9000",
  timeout_ms: 5000,
});

gi.net.writeRawSocket({ socket_id: id, data: "ping\n" });
const data = gi.net.readRawSocket({ socket_id: id, max_bytes: 4096 });
gi.net.closeRawSocket(id);
```

Add listener routes:

```js
gi.connect.registerRoute({
  name: "local-tcp",
  transport: "tcp",
  direction: "inbound",
  match: { listen: "127.0.0.1:9911" },
  mode: "respond",
  script: `"ok\n"`,
});
```

### 4.2 Unix sockets / named pipes

Normalize under `transport: "pipe"`:

```js
gi.connect.registerRoute({
  name: "build-pipe",
  transport: "pipe",
  direction: "inbound",
  match: {
    path: "/tmp/gi-build.sock"       // Unix
    // or name: "\\\\.\\pipe\\gi-build" // Windows
  },
  mode: "turn",
});
```

Notes:

- Unix domain sockets should unlink on shutdown unless `options.unlink_on_close === false`.
- Windows named pipes need ACL defaults and path validation.
- Pipe routes must be session-scoped by default; global routes require explicit config.

### 4.3 HTTP client

Already present as `gi.http.request()`.

Needed improvements:

- stream response body option
- download-to-file option
- retry policy object
- response truncation metadata
- optional secret header references rather than literal secrets

```js
gi.http.request({
  method: "POST",
  url: "https://api.example.com/events",
  headers: { "content-type": ["application/json"] },
  body: JSON.stringify({ ok: true }),
  timeout_ms: 5000,
  retry: 2,
});
```

### 4.4 HTTP route registration

```js
gi.http.registerRoute({
  name: "webhook",
  method: "POST",
  path: "/webhook/:source",
  auth: { type: "bearer", keychain: "webhook/token" },
  mode: "turn",
  script: `
    "Webhook " + gi.event.params.source + ": " + JSON.stringify(gi.event.body)
  `,
});
```

Response model:

```json
{
  "status": 200,
  "headers": { "content-type": ["application/json"] },
  "body": "{\"ok\":true}"
}
```

### 4.5 WebSocket client

Already partly present.

Add route registration:

```js
gi.websocket.registerRoute({
  name: "ws-control",
  path: "/ws/control",
  mode: "respond",
  script: `
    if (gi.event.message === "ping") "pong";
  `,
});
```

WebSocket events:

| Event | Meaning |
|-------|---------|
| `open` | Client connected |
| `message` | Text/binary message received |
| `close` | Client disconnected |
| `error` | Transport error |

### 4.6 SSE client/server

#### SSE outbound/client

```js
const sub = gi.sse.subscribe({
  url: "https://example.com/events",
  headers: { authorization: ["Bearer ..."] },
  event: "message",
  mode: "message",
});
```

#### SSE inbound/server route

```js
gi.sse.registerRoute({
  name: "agent-events",
  path: "/events/agent/:session",
  source: "turn-events",
});
```

This lets scripts expose a stream backed by gi's internal event bus.

### 4.7 MQTT

MQTT should be a first-class route transport rather than just a raw socket use case.

```js
gi.mqtt.registerClient({
  name: "home",
  broker: "mqtt://192.168.1.10:1883",
  client_id: "gi-agent",
  auth: { username: "gi", password_keychain: "mqtt/home" },
});

gi.mqtt.subscribe({
  client: "home",
  topic: "zigbee2mqtt/+/power",
  qos: 0,
  mode: "message",
  script: `
    "MQTT " + gi.event.topic + ": " + gi.event.payload
  `,
});

gi.mqtt.publish({
  client: "home",
  topic: "gi/status",
  payload: "online",
  retain: true,
});
```

MQTT event shape:

```json
{
  "transport": "mqtt",
  "client": "home",
  "topic": "zigbee2mqtt/socket/power",
  "payload": "123.4",
  "qos": 0,
  "retained": false,
  "timestamp": "..."
}
```

---

## 5. Internal event/SSE abstraction

We should separate **event bus** from **transport delivery**.

```go
type EventBus interface {
    Emit(ctx context.Context, topic string, event map[string]any) error
    Subscribe(ctx context.Context, topicPattern string) (<-chan EventEnvelope, error)
}
```

Transports then become adapters:

```text
HTTP webhook → EventBus.Emit("http.webhook.github", event)
MQTT topic   → EventBus.Emit("mqtt.home.zigbee.power", event)
Turn event   → EventBus.Emit("turn.session.iteration", event)
SSE route    → subscribes to EventBus topic and serializes as text/event-stream
Script hook  → subscribes via route spec and may trigger a turn
```

### Event envelope

```go
type EventEnvelope struct {
    ID        string         `json:"id"`
    Topic     string         `json:"topic"`
    Source    string         `json:"source"`
    Transport string         `json:"transport"`
    Timestamp string         `json:"timestamp"`
    SessionID string         `json:"session_id,omitempty"`
    AgentID   string         `json:"agent_id,omitempty"`
    Payload   map[string]any `json:"payload"`
}
```

---

## 6. Route-to-agent delivery

Inbound connectivity routes need consistent delivery into the agent loop.

### `observe`

Only run handler:

```text
external event → script handler → optional side effects
```

### `message`

Store as non-context or context message depending on `options.context`:

```text
external event → message store → SSE broadcast
```

### `turn`

Submit prompt into the normal router:

```text
external event → render prompt → SubmitPromptRouted / SubmitPeerMessage
```

Options:

```json
{
  "mode": "turn",
  "queue": "steer|follow_up|next_turn|immediate",
  "agent_id": "ops",
  "session_policy": "source|route|new"
}
```

### `respond`

For request/response transports:

```text
external request → script handler → transport-native response
```

### `tool`

Expose registered route as a tool:

```text
LLM tool call → route handler → result to LLM
```

---

## 7. Security model

Connectivity hooks are more dangerous than ordinary tool hooks because they can expose the agent to a network.

### Required controls

1. **Default bind is loopback only** for inbound HTTP/WS/SSE/TCP.
2. **Explicit config to bind non-loopback**.
3. **Route auth required** for non-loopback HTTP/WS/SSE.
4. **Secret references only** for credentials:
   - `keychain: "mqtt/home"`
   - not literal passwords in scripts.
5. **Per-route rate limits**.
6. **Max body/message size**.
7. **Backpressure policy**:
   - drop
   - queue bounded
   - block
8. **Audit log** for inbound events and route registrations.
9. **Session scoping** by default.
10. **Route ownership** / source metadata.

### Example auth block

```json
{
  "type": "hmac",
  "header": "x-signature",
  "keychain": "github/webhook-secret",
  "algorithm": "sha256"
}
```

---

## 8. Lifetime model

Routes need explicit lifetimes:

| Lifetime | Meaning |
|----------|---------|
| `turn` | Automatically removed when current turn finishes. |
| `session` | Removed on session shutdown/reload. |
| `workspace` | Persisted in workspace config. |
| `process` | Lives until process exit/reload, not persisted. |

Default should be `session` for script registrations.

```json
{
  "lifetime": "session",
  "restart": "restore|ignore|fail"
}
```

---

## 9. Implementation plan

### Phase 1 — Registry and event bus

Files:

- `internal/connectivity/types.go`
- `internal/connectivity/registry.go`
- `internal/connectivity/eventbus.go`
- `internal/connectivity/auth.go`

Deliverables:

- Route spec type
- Register/unregister/list APIs
- In-memory event bus
- Audit events in store
- Script bridge methods:
  - JS: `gi.connect.registerRoute`, `unregisterRoute`, `listRoutes`, `emit`
  - Joker: `gi-register-route`, `gi-unregister-route`, `gi-list-routes`, `gi-emit-connectivity-event`

### Phase 2 — HTTP + SSE inbound

Files:

- `internal/connectivity/http.go`
- `internal/connectivity/sse.go`
- web server integration

Deliverables:

- Dynamic HTTP route dispatch under a reserved prefix, e.g. `/api/routes/:routeID/...`
- Webhook handler
- SSE stream route backed by event bus
- Body limits and auth

### Phase 3 — WebSocket inbound + outbound cleanup

Files:

- `internal/connectivity/websocket.go`

Deliverables:

- WS server route registration
- WS event shape
- Client connection pooling with route lifetimes

### Phase 4 — MQTT

Files:

- `internal/connectivity/mqtt.go`

Candidate Go library:

- `github.com/eclipse/paho.mqtt.golang`

Deliverables:

- Named MQTT clients
- Subscribe route registration
- Publish API
- Reconnect policy

### Phase 5 — Named pipes / Unix sockets / TCP listeners

Files:

- `internal/connectivity/socket.go`
- `internal/connectivity/pipe_unix.go`
- `internal/connectivity/pipe_windows.go`

Deliverables:

- Unix socket listener
- Windows named pipe listener
- TCP/UDP listener routes
- Backpressure and body/message framing

---

## 10. Minimal first API surface

To keep this manageable, start with:

```js
gi.connect.registerRoute(spec)
gi.connect.unregisterRoute(id)
gi.connect.listRoutes(filter)
gi.connect.emit(topic, payload)
gi.connect.subscribe(spec)

gi.http.request(spec)
gi.http.registerRoute(spec)

gi.sse.subscribe(spec)
gi.sse.registerRoute(spec)

gi.websocket.open(spec)
gi.websocket.registerRoute(spec)

gi.mqtt.registerClient(spec)
gi.mqtt.subscribe(spec)
gi.mqtt.publish(spec)

gi.net.openRawSocket(spec)
gi.net.registerRoute(spec)
gi.pipe.open(spec)
gi.pipe.registerRoute(spec)
```

The route spec should be common across transports, and transport-specific helpers should compile down to `connect.registerRoute`.

---

## 11. Implemented slice (2026-05-03)

The first connectivity slice is now implemented:

- `internal/connectivity` package with:
  - `RouteSpec`, `RouteInfo`, `EventEnvelope`, `RouteResponse`
  - in-memory route registry
  - bounded in-memory event bus with topic-pattern subscriptions
  - register/unregister/list/deliver/emit APIs
- Engine-owned connectivity registry via `turn.Engine.Connectivity()`.
- Script bridge APIs:
  - JS: `gi.connect.registerRoute(spec)`, `unregisterRoute(id)`, `listRoutes(filter)`, `emit(topic, payload)`
  - Joker: `gi-register-route`, `gi-unregister-route`, `gi-list-routes`, `gi-emit-connectivity-event`
- Script tool callbacks connect route registrations to the engine registry.
- HTTP route dispatch:
  - registered routes are reachable under `/api/connect/routes/{routeID}/...`
  - request body is capped to 1 MiB in this first slice
  - handlers receive method/path/query/headers/body/remote address
  - auth middleware allows unauthenticated loopback calls but requires auth for non-loopback clients unless `options.allow_unauthenticated_external=true`
  - supported auth types: `basic`, `bearer`, `header`, `query`, `hmac` (SHA-256), `totp`
  - `basic` supports literal/env-backed username/password today
  - `totp` uses the conditional enrollment/login flow under `/api/auth/*` and accepts the resulting bearer token on connectivity routes
  - WebAuthn remains a planned follow-up auth type once gi has a browser credential storage layer

### Conditional TOTP enrollment routes

Implemented Piclaw-style conditional enrollment endpoints:

| Route | Method | Purpose |
|-------|--------|---------|
| `/api/auth/status` | `GET` | Reports whether TOTP enrollment is complete and whether enrollment is required. |
| `/api/auth/enroll/start` | `POST` | Starts first-user enrollment; loopback-only and only allowed while no user is enrolled. Returns a base32 secret and `otpauth://` URL. |
| `/api/auth/enroll/verify` | `POST` | Verifies the pending enrollment TOTP code and persists the enrolled user. Loopback-only. |
| `/api/auth/totp/verify` | `POST` | Verifies an enrolled user's TOTP code and returns a short-lived bearer token. |

Example route using TOTP session auth:

```json
{
  "name": "external-webhook",
  "transport": "http",
  "auth": { "type": "totp" },
  "mode": "respond"
}
```

Requests to `/api/connect/routes/{routeID}/...` must then include:

```http
Authorization: Bearer <token-from-/api/auth/totp/verify>
```

### HTTPS and ACME serving

The web server remains HTTP-by-default for local development, but can now be started with static TLS certificates or ACME-managed certificates for external connectivity routes.

Static certificate mode:

```sh
gi -listen :8443 -tls-cert /path/fullchain.pem -tls-key /path/privkey.pem
```

ACME/Let's Encrypt mode:

```sh
gi \
  -listen :https \
  -acme-domains gi.example.com \
  -acme-email admin@example.com \
  -acme-accept-tos
```

ACME stores certificates in SQLite by default (`-acme-cache sqlite`) under the `kv_store` namespace `acme/autocert`. Use `-acme-cache vfs` to store certificates in the SQLite-backed VFS namespace `acme-autocert`, or pass an explicit filesystem directory path if a file cache is deliberately wanted. `-acme-http-listen :http` starts an HTTP listener for HTTP-01 challenges and redirects; set it to an empty string to disable the companion listener when TLS termination/challenge handling is managed elsewhere.

- SSE stream adapter:
  - `/api/connect/sse/{topic-pattern}` streams internal connectivity events
  - `?topic=` can be used instead of a path pattern

Already present before this slice:

- outbound raw TCP/UDP client
- outbound WebSocket client
- outbound HTTP request
- script event hooks

Still missing:

- inbound WebSocket routes
- SSE outbound/client abstraction
- MQTT client/subscription/publish
- named pipes / Unix sockets
- TCP/UDP listener routes
- route auth/rate limits beyond body cap
- durable route lifetimes beyond in-memory session/process lifetime
- route-to-agent `message`/`turn` delivery modes

---

## Recommendation

Implement this as a new `internal/connectivity` package and keep it separate from `internal/turn`.

`turn` should only know about a generic delivery interface:

```go
type InboundDelivery interface {
    Deliver(ctx context.Context, event connectivity.EventEnvelope, mode string) error
}
```

That keeps transports from leaking into the agent loop, and lets HTTP/MQTT/socket events all enter through the same queueing/routing path.
