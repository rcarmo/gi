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
- routed/default session allocation now prefers relational identity state over `sessions.scope_json`, including main-session preference, store-boundary `identity_links` normalization, and first-pass multi-channel binding reuse / explicit continuation semantics

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
- connectivity events are bridged into the topic bus under `connectivity.*`
- inbound queue and dispatcher lifecycle now publish onto the canonical topic bus under `runtime.inbound_work` and `runtime.dispatcher`
- hook invocation lifecycle plus higher-level hook decision notices now publish under `runtime.hook`
- core runtime turn/session lifecycle checkpoints now also publish under `runtime.turn` and `runtime.session` (now including generic shared state notices plus explicit setup and terminal/idle transitions, with explicit checkpoints kept non-duplicated and completed paths reported consistently)
- core runtime tool lifecycle checkpoints now also publish under `runtime.tool` (currently started, finished, failed, and skipped)
- core runtime routing/allocation decision notices now also publish under `runtime.routing` (currently persisted route-decision and incoming-route notices)
- the TUI now has a first topic-native consumption slice for canonical topic families on the active session, using `turn.status` / `turn.response` / `turn.thought` plus `runtime.tool` / `runtime.hook` / `runtime.turn` / `runtime.session` / `runtime.routing` and `session.compaction` / `session.steering` / `turn.subturn` notices for status and transcript updates; its `runtime.hook` rendering covers abort/deny/modify/respond decisions, and when a live topic subscription is active, overlapping legacy broadcast handling acts as fallback rather than the primary source for those families
- dedicated topic SSE streaming is now live via `/sse/topics`
- script-facing topic APIs now exist in both script bridges: JS exposes `gi.topics.publish(...)` plus polling-style `gi.topics.subscribe(...)`, `gi.topics.read(...)`, and `gi.topics.unsubscribe(...)`, while embedded Joker exposes `gi-topic-publish`, `gi-topic-subscribe`, `gi-topic-read`, and `gi-topic-unsubscribe`

Current hook/runtime interception status:
- provider-level hook parity now reaches the inference layer: `before_provider_request` supports both context mutation and send-time raw request replacement, and `after_provider_response` observes real provider status/headers when available
- process hooks now run as mounted persistent subprocess sessions per registered handler instead of spawn-per-invocation JSON-RPC calls

Current session/runtime identity status:
- canonical session identity lookup is store-backed via relational tables rather than runtime/table scans
- alias resolution, main-session preference, and allocation resolve-or-create now flow through explicit store APIs
- multi-channel binding support currently covers explicit continuation plus bound reuse; broader automatic linking/fan-out policy remains intentionally undocumented until it is implemented
- direct/IPC ingress now has a normalized engine-facing envelope (`DirectInput` / `DirectOrigin`) so non-web/TUI callers can reuse the same submit/route/continue runtime paths instead of creating separate execution flows
- system/internal-origin processing now has explicit engine entrypoints on top of that envelope, and same-session direct/system follow-ups reuse steering instead of spawning competing turns
- the first durable inbound queue layer is now live via `inbound_work_queue` plus engine enqueue/claim/process/drain helpers; the guarded web runtime now exposes enqueue/list/drain/requeue/discard endpoints, eligibility-aware queue introspection, and a small configurable background dispatcher with bounded retry/backoff state plus store-backed lease ownership, and it now publishes inbound/dispatcher lifecycle notices onto the canonical topic bus instead of keeping this runtime surface web-local
- web consumers now also have a dedicated `/sse/topics` endpoint for canonical topic-bus streaming by topic pattern and optional session/agent scoping
