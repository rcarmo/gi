# Internal reference

This subtree is the **canonical internal documentation surface** for gi runtime features that the agent can use or extend.

It is written for:
- the agent running inside gi
- humans extending gi
- future packaging into a read-only embedded reference tree such as `vfs://reference/...`

## Contract

If a change adds or materially changes any of the following, it must update this subtree in the same change:
- built-in tools
- scripting runtimes or bridge APIs
- hook surfaces
- managed `vfs://` behavior
- skill/package structure
- other agent-visible extension points

## What belongs here

- **tool contracts**: purpose, parameters, outputs, side effects, examples
- **scripting docs**: engine behavior, bridge globals, file/state access, examples
- **hook docs**: lifecycle timing, guarantees, mutation rules, examples
- **VFS docs**: `vfs://` URL semantics, path resolution, read-only namespaces, export/sync model
- **skill/package docs**: structure, conventions, discovery, packaging

## Stability policy

Paths under `docs/internal/` should remain stable once referenced by prompts, tools, or future `vfs://reference/...` URLs.

## Current structure

- `tools/` — built-in tool contracts
- `scripting/` — scripting runtimes and bridge docs
- `hooks/` — hook and lifecycle docs
- `vfs/` — managed VFS and `vfs://` URL docs (including `vfs://chat` projection)
- `skills/` — skill/package structure docs
- `search/` — hybrid workspace search and indexing design (including `fts://` namespace docs)
- `routing.md` — routing policy, route resolution, and model-routing observability
- `runtime-target-state.md` — database-backed target state for the core runtime refactor (sessions, turns, steering, hooks, events, IPC, multi-channel bindings)
- `repo-structure-refactor.md` — interim repository-structure reassessment and functional regrouping plan for the runtime refactor
- `subturn-runtime.md` — concrete sub-turn runtime contract, limits, store APIs, and current implementation status

The `search/` subtree is the canonical reference for the current hybrid workspace search direction: SQLite metadata + FTS5 + vec + local embeddings, plus the runtime-facing read-only `fts://` locator contract.

`routing.md` is the canonical in-repo reference for routing decisions and route-event persistence.

`runtime-target-state.md` is the canonical schema/state target for the runtime refactor.

`repo-structure-refactor.md` is the working note for the interim structure-tidying phase that reassesses package/file grouping before deeper runtime changes continue.

Current interim structure status:
- `internal/store` session identity/allocation code has been regrouped by responsibility (`session_identity.go`, `session_aliases.go`, `session_main.go`, `session_resolution.go`, `session_channel_bindings.go`)
- `internal/turn` routing/session/helper/shell/value logic has been split out of `engine.go`
- `internal/tui` session-reference helpers have been split out of `chat.go`
- follow-up audit fixes tightened identity-read query scope and removed repeated transcript DB reloads from the TUI render path
- routed/default session allocation now prefers relational identity state over `sessions.scope_json`, including main-session preference, store-boundary `identity_links` normalization, first-pass multi-channel binding reuse / explicit continuation semantics, and fallback-aware route-context/source-agent derivation so canceled callers do not silently regress route preparation to stale scope snapshots

## Current status

This subtree is being bootstrapped. When internal surfaces are implemented before full docs exist, add at least a minimal placeholder page and update it as the implementation stabilizes.

Current search status:
- architecture chosen: **vec + FTS**
- ADR written
- internal design written
- Go package scaffold added under `internal/search/`
- runtime read-only `fts://` namespace is implemented for model/tool retrieval workflows
- full `internal/search` backend wiring and advanced index pipeline work remains pending

Current topic/runtime publication status:
- the in-memory topic bus is live and now carries bridged turn/session notices plus runtime-critical steering and subturn lifecycle topics
- connectivity events are bridged into the topic bus under `connectivity.*`, and the bridge subscription now follows engine lifecycle cancellation instead of living on a `context.Background()` subscription
- inbound queue and dispatcher lifecycle now publish onto the canonical topic bus under `runtime.inbound_work` and `runtime.dispatcher`
- hook invocation lifecycle plus higher-level hook decision notices now publish under `runtime.hook`, and hook invocation audit persistence, turn-failure persistence (including durable `turn.failure_marked` postmortem markers), turn finalization persistence (including shell-path finalization), explicit cancel-request bookkeeping (`CancelTurn(...)` active/queued branches plus shell-run cancel cleanup), deferred agent-end hook emission, compaction restore cleanup, detached submit/launch cleanup, queued-turn/continuation coordination (including staged continuation submission/launch plus the newer queued-session persistence updates for ordinary queued submit, same-session prompt steering, active/idle/queued continue handoff, queued-turn launch, launch-conflict steering fallback, and cleanup-time re-normalization), active-turn heartbeat refresh, final steering checkpoint / steering-reject bookkeeping, post-turn coordination cleanup, workspace extension loading, subturn result/orphan parent-notice delivery, routing/fork system notices, direct ingress session-key resolution + downstream submit/continue routing, startup interrupted-turn recovery, launch-conflict steering fallback, finalize-time turn identity/model recovery, post-claim inbound-work retry/failure/completion bookkeeping, web inbound-dispatcher lease release, dispatcher nil-context normalization, and `cmd/gi` web-runtime startup/shutdown wiring now use explicit caller context or engine-owned/detached lifecycle context where they must outlive request cancellation without becoming process-lifetime work; `CancelTurn(...)` also now enforces the caller-supplied session boundary instead of trusting turn id alone and keeps post-cancel session state aligned with both queued depth and any still-active claim, while same-server dispatcher startup is now single-shot so repeated calls cannot spawn competing same-owner loops
- core runtime turn/session lifecycle checkpoints now also publish under `runtime.turn` and `runtime.session` (now including generic shared state notices plus explicit setup, queue-submission, hold-resolution, and terminal/idle transitions, with explicit checkpoints kept non-duplicated, completed paths reported consistently, completed exits carrying explicit completion metadata such as `iterations` / `completion_kind` across both turn and session runtime notices, interrupted-turn recovery publishing both explicit `turn_recovered` checkpoints plus generic recovery `turn_state` / `session_state` notices while also restarting queued follow-up work automatically when stale-claim recovery leaves runnable queued work behind, active cancellation publishing explicit `turn_cancelling` checkpoints plus richer session-state metadata, queued cancellation now publishing explicit terminal turn/session checkpoints when it leaves the session idle while also continuing later queued work when the session remains queued, queued turn creation publishing explicit `turn_submitted` checkpoints, and held-failure review/skip/retry resolution publishing explicit `turn_failure_held` / `turn_failure_resolved` checkpoints rather than remaining DB-audit-only; those hold-resolution notices now publish the post-update phase and normalized held payload fields so the live topic payload matches persisted `turn_failures` state)
- core runtime tool lifecycle checkpoints now also publish under `runtime.tool` (currently started, finished, failed, and skipped)
- core runtime routing/allocation decision notices now also publish under `runtime.routing` (currently persisted route-decision and incoming-route notices)
- the TUI now has a first topic-native consumption slice for canonical topic families on the active session, using `turn.status` / `turn.response` / `turn.draft` / `turn.thought` plus `runtime.tool` / `runtime.hook` / `runtime.turn` / `runtime.session` / `runtime.routing` / `runtime.inbound_work` and `session.compaction` / `session.steering` / `turn.subturn` notices for status and transcript updates; its `runtime.hook` rendering covers invocation errors/timeouts plus abort/deny/modify/respond decisions, its `runtime.inbound_work` rendering covers queue/retry/failure/completion plus manual requeue/discard notices and retry attempt counts, `turn.draft`, `turn.thought`, `turn.status` running, `runtime.tool` started, `runtime.turn` waiting-on-tools, and `runtime.session` running notices mark the UI as actively running, `runtime.session` queued notices now render a visible queued state, idle/completed/terminal turn/session notices (`turn.status` idle, `runtime.turn`, `runtime.session`) clear stale draft/running UI state, the newer explicit cancel/recovery checkpoints now flow through the same canonical topic families rather than staying DB-audit-only, terminal system messages now broadcast live for all terminal outcomes (not only completed/failed) and therefore flow through the same `turn.response`/system-message bridge the TUI already consumes, both the running-entry and cleanup paths are now centralized in helpers to reduce drift, and routing, tool, hook, inbound-work, compaction, steering, and sub-turn rendering are centralized too so the remaining legacy fallback status/draft/thought plus completion/error cleanup paths still mirror the same entry/reset semantics when topic-native mode is inactive while legacy routing broadcasts render through fallback; when a live topic subscription is active, overlapping legacy broadcast handling acts as fallback rather than the primary source for those families
- dedicated topic SSE streaming is now live via `/sse/topics`
- script-facing topic APIs now exist in both script bridges: JS exposes `gi.topics.publish(...)` plus polling-style `gi.topics.subscribe(...)`, `gi.topics.read(...)`, and `gi.topics.unsubscribe(...)`, while embedded Joker exposes `gi-topic-publish`, `gi-topic-subscribe`, `gi-topic-read`, and `gi-topic-unsubscribe`; closed topic subscriptions are now dropped eagerly on read, while explicit unsubscribe stays idempotent, and raw/websocket close operations follow the same idempotent cleanup contract

Current hook/runtime interception status:
- provider-level hook parity now reaches the inference layer: `before_provider_request` supports both context mutation and send-time raw request replacement, and `after_provider_response` observes real provider status/headers when available
- process hooks now run as mounted persistent subprocess sessions per registered handler instead of spawn-per-invocation JSON-RPC calls

Current session/runtime identity status:
- canonical session identity lookup is store-backed via relational tables rather than runtime/table scans
- alias resolution, main-session preference, and allocation resolve-or-create now flow through explicit store APIs
- multi-channel binding support currently covers explicit continuation plus bound reuse; broader automatic linking/fan-out policy remains intentionally undocumented until it is implemented
- direct/IPC ingress now has a normalized engine-facing envelope (`DirectInput` / `DirectOrigin`) so non-web/TUI callers can reuse the same submit/route/continue runtime paths instead of creating separate execution flows
- system/internal-origin processing now has explicit engine entrypoints on top of that envelope, and same-session direct/system follow-ups reuse steering instead of spawning competing turns; direct/session-key ingress audit metadata is now aligned across turn metadata, persisted user-message payloads, and `turn.started` audit rows (including `ingress_session_key` when provided), and direct/IPC/system prompt + peer-message + continue actions now also survive already-canceled caller contexts long enough to reach those hardened runtime paths instead of failing before session-key resolution or downstream submit/continue
- the first durable inbound queue layer is now live via `inbound_work_queue` plus engine enqueue/claim/process/drain helpers; the guarded web runtime now exposes enqueue/list/drain/requeue/discard endpoints, eligibility-aware queue introspection, and a small configurable background dispatcher with bounded retry/backoff state plus store-backed lease ownership, it publishes inbound/dispatcher lifecycle notices onto the canonical topic bus instead of keeping this runtime surface web-local, and once an item is already claimed its retry/failure/completion bookkeeping now survives transient caller cancellation instead of stranding claimed rows
- web consumers now also have a dedicated `/sse/topics` endpoint for canonical topic-bus streaming by topic pattern and optional session/agent scoping
