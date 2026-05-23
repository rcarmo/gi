# Internal topic system design

Date: 2026-05-05

## Status

Implemented so far:

- `internal/topics/` in-memory bounded topic bus
- engine-owned `Topics()` accessor
- normalized publication of turn/session broadcast events into the topic bus
- steering lifecycle publication under `session.steering`
- subturn lifecycle publication under `turn.subturn`
- extension lifecycle publication (`extension.loaded`, `extension.failed`)
- bridge from the existing connectivity event bus into topic topics under `connectivity.*`, now bound to engine lifecycle context instead of a process-lifetime `context.Background()` subscription
- inbound queue + dispatcher lifecycle publication under `runtime.inbound_work` and `runtime.dispatcher`
- hook invocation lifecycle publication plus higher-level hook decision notices under `runtime.hook` (including tool-call/approval/result abort, deny, modify, and direct-respond outcomes); hook invocation audit persistence, turn-failure persistence, turn-finalization persistence (including shell-path finalization), deferred agent-end hook emission, compaction-restore cleanup, detached submit/launch cleanup, queued-turn/continuation coordination (including the newer queued-session persistence work for ordinary queued submit, same-session prompt steering, queued cancel with active-claim awareness, staged steering continuations, active/idle/queued continue handoff normalization, queued-turn launch queue-count normalization, and launch-conflict steering fallback normalization), active-turn heartbeat refresh, final steering checkpoint / steering-reject bookkeeping, post-turn coordination cleanup, workspace-extension loading, startup interrupted-turn recovery, launch-conflict steering fallback, finalize-time turn identity/model recovery, and post-claim inbound-work bookkeeping now use explicit caller context or engine-owned background context instead of raw `context.Background()` so they survive request cancellation without becoming process-lifetime work
- core runtime turn/session lifecycle checkpoints under `runtime.turn` and `runtime.session` (now including generic `turn_state` / `session_state` notices from the shared state-hook path, plus explicit setup, queue-submission, hold-resolution, and terminal/idle checkpoints; explicit checkpoints stay singular rather than being duplicated by the generic state path, completed paths use `turn_completed` / `turn_completed`-reasoned idle consistently, completed exits now carry explicit completion metadata such as `iterations` / `completion_kind` across both turn and session runtime notices, interrupted-turn recovery now also publishes explicit `turn_recovered` checkpoints plus generic recovery `turn_state` / `session_state` notices while restarting queued follow-up work automatically when stale-claim recovery leaves runnable queued work behind, active cancellation now publishes explicit `turn_cancelling` checkpoints plus richer session-state metadata, queued cancellation now publishes explicit terminal turn/session checkpoints when it leaves a session idle while also continuing later queued work when queued follow-up remains and preserving `running` when another active claim still exists, queued turn creation plus staged steering continuation submission now publish explicit `turn_submitted` checkpoints while also persisting `queue_count` / queued session state, same-session steering and continue paths now normalize `session_state` back to the active-turn model/claim truth rather than leaving stale queued state behind during early returns or external-claim handoff, and held-failure review/skip/retry resolution now publish explicit `turn_failure_held` / `turn_failure_resolved` checkpoints instead of remaining visible only through stored turn events; those hold-resolution notices now carry the post-update phase and normalized held payload fields so they stay aligned with persisted `turn_failures` state; still not a full mirror of every turn row)
- core runtime tool lifecycle checkpoints under `runtime.tool` (currently started, finished, failed, and skipped)
- core runtime routing/allocation decision notices under `runtime.routing` (currently route-decision and route-incoming checkpoints tied to persisted `routing_events` rows)
- dedicated SSE topic-stream endpoint at `/sse/topics` with topic-pattern plus session/agent scoping

Still pending:

- fuller publication coverage for turn/tool/hook/routing/session lifecycle families beyond the currently bridged/runtime-critical slices; `runtime.turn` / `runtime.session` now include shared state-hook notices plus explicit setup/queue/hold/terminal checkpoints, with non-duplicated publication semantics and consistent completed-vs-terminal naming, but still do not mirror every stored event row; `runtime.tool` currently covers only started/finished/failed/skipped rather than every approval/hook sub-phase; `runtime.hook` now includes invocation rows plus high-level tool-related decisions but not every hook family-specific semantic summary yet; and `runtime.routing` currently covers persisted decision/incoming notices rather than every route-resolution branch before persistence
- broader TUI/topic-native coverage beyond the first adopted runtime/session/steering/subturn families and status paths
- deeper convergence between connectivity and topics

## Goal

Define an internal topic system that can:

- propagate information from extensions to SSE or the TUI
- propagate information between extensions
- remain transport-neutral so future web/TUI/runtime consumers can subscribe without custom wiring for every event type

## Problem

Today gi has multiple event-like mechanisms, but no single shared topic abstraction:

- turn-engine session broadcasts for TUI/web status updates
- connectivity registry event bus for connectivity routes/SSE streams
- hook callbacks for synchronous extension interception

This leaves a gap:

- extensions cannot publish durable, named runtime topics for other extensions
- TUI/web consumers cannot subscribe to a stable internal topic namespace independent of turn/session broadcast internals
- SSE integration and extension-to-extension messaging are coupled to whichever subsystem emitted the event first

## Design principles

1. **Transport-neutral core**
   - topics are internal runtime messages first
   - SSE/TUI/websocket are delivery adapters, not the source abstraction

2. **Session-aware but not session-bound**
   - some topics are global
   - some are session-scoped or agent-scoped

3. **Typed envelope, JSON-safe payload**
   - payloads should be safe to serialize to SSE/web/TUI observers
   - avoid exposing arbitrary Go pointers or runtime-only closures

4. **Bounded and observable**
   - subscriptions should be bounded
   - slow subscribers should drop or coalesce rather than block the runtime

5. **Extension-friendly**
   - extensions need simple publish/subscribe APIs
   - subscriptions must be easy to clean up on unload/reload

## Proposed model

### Topic names

Use dotted topic namespaces:

- `turn.status`
- `turn.draft`
- `turn.tool.start`
- `turn.tool.end`
- `session.changed`
- `session.compaction`
- `extension.loaded`
- `extension.error`
- `ui.notice`
- `connectivity.event`

Optional scoped suffixes:

- `session.<sessionID>.turn.status`
- `agent.<agentID>.notice`
- `chat.<sessionID>.draft`

### Envelope

```go
type TopicEnvelope struct {
    Topic      string         `json:"topic"`
    SessionID  string         `json:"session_id,omitempty"`
    AgentID    string         `json:"agent_id,omitempty"`
    Source     string         `json:"source,omitempty"`      // turn|extension|connectivity|web|tui
    Type       string         `json:"type,omitempty"`        // status|delta|notice|result|error
    Payload    map[string]any `json:"payload,omitempty"`
    Timestamp  string         `json:"timestamp"`
}
```

This is intentionally close to SSE/event-bus payloads already present in gi.

## Core interfaces

```go
type TopicBus interface {
    Publish(TopicEnvelope)
    Subscribe(pattern string, opts SubscribeOptions) (<-chan TopicEnvelope, func())
}

type SubscribeOptions struct {
    Buffer    int
    SessionID string
    AgentID   string
}
```

### Pattern matching

Support simple forms first:

- exact topic: `turn.status`
- prefix wildcard: `turn.*`
- all: `*`

No regex at first.

## Relationship to existing systems

### 1. turn.Engine broadcast

Current session broadcast can become an adapter layer:

- engine emits session-local UI events as it does today
- engine also publishes normalized `TopicEnvelope`s into the topic bus
- TUI/web can continue consuming session broadcast events during transition

### 2. connectivity event bus

Connectivity already has a bounded internal event bus.

Recommendation:

- do **not** duplicate the implementation long-term
- evolve toward a shared internal topic bus with connectivity as one publisher/subscriber family
- keep the connectivity route/topic naming but bridge it into the common bus

### 3. hooks

Hooks stay synchronous mutation/interception surfaces.

They are **not** the same thing as topics.

However, hook callbacks or extension handlers should be able to publish topic messages for observation or fan-out.

## Extension-facing API

### JavaScript

```js
gi.topics.publish({
  topic: "ui.notice",
  session_id: gi.sessionId,
  source: "script",
  type: "notice",
  // optional: sequence may be supplied by trusted bridge callers and advances the bus watermark
  payload: { text: "Index refresh complete" }
})

const sub = gi.topics.subscribe("turn.*", { session_id: gi.sessionId, buffer: 32, after_sequence: lastSeenSequence })
const events = gi.topics.read(sub, 10)
// events include: topic, session_id, agent_id, source, type, sequence, payload, timestamp
gi.topics.unsubscribe(sub)
```

### Joker

- `gi-topic-publish` — publish a canonical topic envelope; optional `sequence` follows the same watermark rules as JS
- `gi-topic-subscribe` — open a polling subscription handle; options include `session_id`, `agent_id`, `buffer`, and `after_sequence`
- `gi-topic-read` — read buffered envelopes from a subscription handle, including the bus-wide `sequence`
- `gi-topic-unsubscribe` — close a polling subscription handle

Like the JS bridge, the current Joker topic API uses polling handles instead of long-lived callbacks.

Script-facing topic APIs are session-bound by the host bridge. When a script runs with a current session, omitted `session_id` values are filled with that session, and explicit `session_id` overrides must match it. Cross-session publish/subscribe attempts are rejected for both JavaScript and Joker. Topic read/unsubscribe handles are also owned by the creating session.

## TUI / SSE integration

### TUI

The TUI can subscribe to:

- session-scoped turn topics
- extension notices
- future cross-extension/internal topics

A first topic-native slice is now live: the TUI subscribes to canonical topic families for the active session and uses `turn.status`, `turn.response`, `turn.draft`, `turn.thought`, `runtime.tool`, `runtime.hook`, `runtime.turn`, `runtime.session`, `runtime.routing`, `runtime.inbound_work`, `session.compaction`, `session.steering`, and `turn.subturn` notices for status and transcript updates. Its `runtime.hook` rendering now covers invocation errors/timeouts plus abort/deny/modify/respond decisions, its `runtime.inbound_work` rendering covers queue/retry/failure/completion plus manual requeue/discard notices including retry attempt counts, `turn.draft`, `turn.thought`, `turn.status` running, `runtime.tool` started, `runtime.turn` waiting-on-tools, and `runtime.session` running notices now mark the UI as actively running, `runtime.session` queued notices now render a visible queued state, idle/completed/terminal turn/session notices (`turn.status` idle, `runtime.turn`, `runtime.session`) clear stale draft/running UI state, and the newer explicit cancel/recovery checkpoints now arrive through the same canonical runtime topic families instead of being visible only through stored turn events. Terminal system messages now broadcast live for all terminal outcomes so they arrive through the same `turn.response` / `system_message` bridge the TUI already consumes. Both the active-running entry path and the cleanup path are now centralized in TUI helpers instead of duplicated across handlers, reducing drift between equivalent canonical lifecycle edges; the remaining legacy fallback status/draft/thought plus completion/error cleanup paths now mirror the same running-state entry/reset semantics when topic-native mode is inactive, and routing, tool, hook, inbound-work, compaction, steering, and sub-turn rendering are now centralized across canonical and fallback paths. Where those topic-native paths have a live active subscription, the overlapping legacy broadcast handlers are treated as fallback compatibility paths rather than the primary source of truth.

This avoids hardcoding every new event type into one custom engine broadcast path.

### SSE

A reserved SSE endpoint now streams canonical topic envelopes directly:

- `/sse/topics?topic=turn.*&session_id=...`
- `/sse/topics?topic=runtime&last_event_id=<sequence>` for aggregate runtime reconnects

The endpoint emits the topic envelope `sequence` as SSE `id:` and returns a `connected` event containing `last_sequence`. On reconnect, clients may send `Last-Event-ID` (or `last_event_id`) and the connected event reports `last_event_id` plus `missed_sequence_count` when the process-local watermark has advanced. This is gap detection, not replay; topic persistence/replay remains a later opt-in.

This coexists with the older chat-turn SSE feed while newer runtime families converge onto the topic bus first instead of requiring bespoke SSE adapters.

## Delivery semantics

### Default

- bounded buffered channels
- publish serializes sequence assignment and delivery under the bus lock so each subscriber observes monotonically ordered envelopes
- every envelope has a bus-wide `sequence`; explicit sequence inputs advance the watermark, automatic envelopes increment it
- slow subscribers use **drop oldest** backpressure, keeping the newest state-like events available without unbounded memory growth
- topic SSE exposes the envelope `sequence` as SSE `id:` and includes the current `last_sequence`, optional `last_event_id`, and optional `missed_sequence_count` in the `connected` event

### Sticky/coalesced topics decision

The shipped bus is currently a bounded **firehose** with bus-wide monotonic sequences and drop-oldest backpressure. It does not retain sticky state and it does not coalesce by key inside the bus. Producers that need durable or resumable state must continue to write canonical SQLite rows first and publish topic envelopes as live notifications of that state.

This keeps the bus simple and prevents topic delivery from becoming a second source of truth. Reconnects can detect sequence gaps via SSE `Last-Event-ID` / `missed_sequence_count`, but replay comes from durable store APIs, not from topic memory.

### Coalescing candidates

Some topics may gain producer-side coalesced summary publication later, but not bus-level mutation semantics:

- `turn.draft`
- `turn.status`
- `session.changed`

## Persistence

First phase: **in-memory only**, with a process-local monotonic sequence watermark for reconnect/audit correlation.

Optional later persistence:

- selected topic envelopes mirrored into turn/session event tables
- replay of the latest sticky topic state for new subscribers

Not every topic should be persisted.

## Initial implementation phases

### Phase 1 — shared internal bus

Status: **implemented**

- added `internal/topics/`
- implemented in-memory bounded bus
- added publish/subscribe tests, including concurrent publish ordering, canceled-subscription cleanup, monotonic sequence delivery, and drop-oldest behavior
- added topic envelope type with monotonic sequence numbers

### Phase 2 — engine + extension bridge

Status: **implemented for the current canonical slices**

- turn/session broadcast events are published into the topic bus
- steering lifecycle notices are published under `session.steering`
- subturn lifecycle notices are published under `turn.subturn`
- extension lifecycle events are published into the topic bus
- script bridge APIs for publish/subscribe/read/unsubscribe are live in both JavaScript and Joker, with session-bound handles and `after_sequence` support
- existing engine broadcasts remain intact as compatibility/fallback paths where consumers still depend on them

### Phase 3 — SSE/TUI adapters

Status: **implemented for the first topic-native consumer set**

- topic-stream SSE endpoint is live at `/sse/topics`
- TUI subscribes to the canonical topic families for active-session runtime status/transcript updates while retaining legacy fallback paths when no live topic subscription exists

### Phase 4 — convergence with connectivity

- bridge or merge connectivity event bus into topic bus
- unify topic naming and subscription behavior

## Why this closes the current plan item

The checklist item asked for a **design** for an internal topic system usable by extensions, SSE, and TUI, including extension-to-extension propagation. This document defines:

- goals
- constraints
- envelope shape
- publish/subscribe interfaces
- relationship to existing engine/connectivity/hook systems
- phased implementation path

That is enough to guide later implementation without forcing premature runtime rewrites today.
