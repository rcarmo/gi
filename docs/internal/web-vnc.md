# Web VNC backend

Gi provides Piclaw 3.2.5-compatible VNC metadata, WebSocket-to-TCP relay and handoff endpoints. The browser owns RFB negotiation, decoding, authentication and rendering. The shared Classic app at fixtures4259e82 has no mounted VNC host; issue #47 stays open and `@cap-vnc` is unclaimed until independent viewer/tab/popout acceptance.

## Configuration and target policy

VNC is disabled by default. Set `GI_WEB_VNC_TARGETS` to a JSON array of operator-approved targets before starting Gi:

```json
[{"id":"desktop","label":"Desktop","host":"127.0.0.1","port":5900,"readOnly":false}]
```

Invalid configuration yields an empty target list. Target records require a nonempty id/host and port 1..65535. At most 128 valid targets are loaded. Host/port details are omitted from UI metadata; labels default to IDs. `readOnly` is an advisory UI input setting, not an enforced transport security boundary. Use server-side VNC permissions to enforce read-only access.

`GI_WEB_VNC_ALLOW_DIRECT=true` explicitly enables caller-supplied `host:port` or `[IPv6]:port` destinations. Any other value leaves direct targets disabled. This option grants trusted browser users access to network destinations reachable by the Gi process; keep it disabled when an allowlist is sufficient. The relay is not an SSRF-safe proxy for untrusted callers.

## Routes and protocol

* `GET /vnc/session` returns `{enabled,transport:"websocket",ws_path:"/vnc/ws",renderer:"placeholder",host_policy,direct_connect_enabled,targets}`. The renderer field follows Piclaw's existing metadata contract. `?target=id` adds `{target:{id,label,read_only,direct_connect}}`, or returns 404 for an unknown target. Metadata does not dial TCP.
* `GET /vnc/ws?target=id` upgrades to WebSocket. The server sends `{type:"vnc.connected",target:{id,label}}`; RFB data then uses binary frames in both directions. Text `{type:"ping"}` gets a local `{type:"pong"}`; other frames retain Piclaw's raw-forwarding behaviour. Initial connection failures send `{type:"vnc.error",error}` before closing.
* `POST /vnc/handoff?target=id` (or JSON `{target:id}`) returns `{handoff:{token,expires_at}}` for an attached connection. The next socket uses `?target=id&handoff=token`. Tokens bind to owner/target, are single-use and expire after 15 seconds. Handoff keeps the same TCP connection and closes the previous browser socket. The relay does not replay framebuffer bytes; detached-window state transfer belongs to the frontend.

Every route requires same-origin HTTPS or loopback peer/Host. Enrolled instances require the browser session cookie and reject bearer/query authority. Owner identity is cookie-scoped; cookie validity is checked before inbound frames and outbound delivery. Unenrolled trusted-local instances use Piclaw's local fallback owner. This fallback does not authenticate separate local users. No route grants CORS.

## Bounds and shutdown

The manager admits at most 16 active connections plus pending dials. Each WebSocket reads frames up to 1MiB and has a bounded 64-frame outgoing queue; TCP reads use 32KiB buffers. Slow readers close instead of extending queues. TCP connect timeout is five seconds, initial-data timeout ten seconds, and writes have five-second deadlines. Target dialing runs outside the manager lock. Shutdown rejects new attaches, closes TCP/browser sockets and joins relay workers and pending dials.

Disconnect closes TCP immediately unless a handoff is pending; the token timer then closes an unattached connection at expiry. Revoked cookies fence subsequent I/O and close the upstream on disconnect. Binary protocol content is forwarded unchanged and is never interpreted or executed by the Go backend.

## Verification and performance, 2026-10-05

* Seven focused configuration/bridge/security tests pass: disabled defaults, explicit direct opt-in, binary data, ping, persistent TCP handoff, replay/expiry/wrong-owner rejection, target policy, auth/origin guards, cookie revocation, silent greeting timeout and shutdown.
* Full Go suite: 2,141 tests passed across 37 packages, 79.7s wall, 61.7s CPU, 746MB peak RSS including compilation. Versus the preceding full run: wall -28%, CPU -2%, RSS +1%; no new reported regression.
* Functional regressions: 147 passed, 12 skipped; lifecycle 228.3s wall, 165.7s CPU, 1,341MB peak RSS including compilation. The extra skip is the configured local RFB fixture test, which passes under `make test-vnc-api`.
* `make test-vnc-api`: two real-browser WebSocket protocol tests pass against a disposable local TCP server. Final warm lifecycle 11.9s wall, 8.7s CPU, 183MB peak RSS; browser tests take 1.4s. The first cold lifecycle took 172.8s; this difference reflects compilation/cache state, not relay throughput.
* Vet passes. No physical remote display, full shared VNC scenario acceptance, Windows/macOS runtime or race-detector result is claimed.

Cumulative CPU and `alloc_space`/`alloc_objects` were inspected for the focused run. It is too short for useful CPU samples. Allocated bytes include about 528KiB attributed to the relay reader and about 1MiB to server route setup; fixture/server construction and package initialisation dominate sampled object counts. TCP chunks are copied only when an attached client needs a frame, avoiding allocations while waiting for a handoff. The bounded per-frame copy is retained so queued writes never alias the reusable TCP read buffer. Browser lifecycle timing/RSS does not measure allocations; no live rendering performance claim is made.

Local logs: `/workspace/tmp/gi-vnc-final-focused.log`, `gi-vnc-full-go-profiled.log`, `gi-vnc-functional-all.log`, `gi-vnc-browser-final.log`, `gi-vnc-profile-analysis.log`, `gi-vnc-vet.log`. Profiles/history are in `~/.cache/gi-test-profile`.
