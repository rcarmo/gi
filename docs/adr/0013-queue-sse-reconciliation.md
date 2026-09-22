# ADR-0013: Reconcile queue state across SSE delivery and reconnect

## Status

Accepted — 2026-09-22

## Context

The web stream forwarded legacy engine broadcasts, while queue lifecycle changes were published to the topic bus. External queue changes therefore depended on the ten-second safety poll. Disconnect only changed the connection indicator; assistant previews and running state stayed visible. Callbacks from an old EventSource could still invoke the current handlers.

Classic `@ux-original-016` requires accepted queue display and authoritative reconciliation. `@ux-reconnect-001` requires clearing transient agent state while preserving the user's composer draft. Other reconnect scenarios also require context usage, search-view preservation or version-drift UI; those remain separately unmapped.

## Decision

Keep legacy SSE events and add an invalidation event, `queue_changed`, for the selected session's `runtime.turn`, `runtime.session` and `session.queue` topic events. Successful HTTP queue mutations publish `session.queue`. Subscribe before sending `connected`, so readiness does not precede subscription. Invalidation causes a coalesced refresh of persisted queue/status/timeline data; notifications do not become synthetic queue records.

A queued composer submission gets a browser-generated `client_request_id`. The server stores it in turn metadata and queue GET returns that metadata. An optimistic placeholder appears immediately and its mutation controls are disabled until it has a durable identity. The local token is correlation metadata, not an idempotency key or authorisation credential.

When a server queue record carries the same client token, remove the local placeholder. Merely hiding it is insufficient: it would reappear after the durable record was consumed or cancelled. Do not deduplicate by text; two identical prompts can be legitimate distinct entries. Rejected submissions remove their placeholder and retain the existing origin-owned draft recovery path.

Disconnect/stale notifications clear agent status, assistant draft/plan/thought previews, pending request/turn refs and running state. Composer text/media/references remain untouched. A connection revision rejects status/queue reads that started before disconnect; periodic refreshes cannot reintroduce transient state while the transport is down. Reconnect and wake refresh persisted status, queue and timeline immediately.

Move the native SSE transport to `web/src/gi-sse-client.ts`, re-exporting the existing `SSEClient` API. Each source's callbacks check that the source is still current. Error closes and invalidates the source before reconnect scheduling; explicit disconnect cancels timers and clears connecting state. The component hook also gates callbacks by the captured selection generation, including the interval before an old effect is cleaned up. Repeated A→B→A visits have different generations.

## Verification

- `@ux-original-016`: real queued prompt with held HTTP acknowledgement, SSE-first reconciliation, two same-text entries with distinct IDs, external append and external removal within four seconds (below the safety-poll interval).
- Rejected-send regression: pending placeholder disappears and captured text merges ahead of the newer origin draft.
- `@ux-reconnect-001`: the isolated shell gate emits real stdout. A byte-for-byte local SSE proxy closes actual sockets and rejects reconnection temporarily. The draft/status panels and stop control disappear; user input remains. An independent API client changes the queue during the outage, and reconnect restores authoritative rows/running state. No fabricated timeline/SSE payload is injected.
- `TestLegacySSEQueueInvalidationIsReadyAndSessionScoped`: real streaming HTTP, readiness ordering, foreign-session exclusion and lifecycle invalidation; three race-detector runs.
- `sse-client.test.ts`: replaced/closed transport callbacks cannot deliver data, change status or resurrect reconnect timers; reconnect while connecting creates a fresh source.

Full matrix: 138/138 executions, 12/236 Classic IDs passing, 224 unmapped. All 42 shared-contract cases remain unmapped. Existing functional web tests: 70/70. Go tests/vet, Bun hook checks and 15 source/helper tests pass.

## Limits and terminal adaptation

No durable event replay cursor, context-usage reconstruction, search-view reconnect contract, manual version-drift notice or queue return/Steer implementation is added. Queue invalidation is best effort; the existing ten-second poll remains a recovery path. Client tokens do not prevent duplicate submission after uncertain network delivery.

The TUI consumes the native topic bus rather than browser SSE. Its existing generation guards already reject old-session events. Future queue controls should refresh an on-demand bounded list from persisted records after topic invalidation, retaining the existing nonzero footer count. Transport warnings belong in the existing transient footer notice; no persistent connection row, sidebar or header is needed. This web slice adds no new terminal controls or terminal parity credit.
