# Runtime package inventory

## Status

Current ownership map for packages touched by the core runtime refactor.

## `internal/turn`

Files:

- `engine.go`
- `engine_test.go`

Ownership:

- turn orchestration and lifecycle phases
- active-turn coordination calls into store primitives
- steering checkpoints and continuation launch decisions
- hook registry/runtime/process-hook mounting
- sub-turn orchestration call sites
- direct/IPC/system ingress entrypoints
- runtime topic publication helpers
- local runner cancellation handles and provider/tool execution loop

Boundary:

- may coordinate runtime phases, but durable truth belongs in `internal/store` and typed allocation/routing policy belongs in `internal/session` / `internal/routing`.

## `internal/session`

Files:

- `allocator.go`
- `key.go`
- `scope.go`
- `allocator_test.go`

Ownership:

- canonical session allocation inputs
- opaque session key/signature construction
- dimension normalization
- default/routed allocation shape
- identity-link canonicalization helpers used before store resolution

Boundary:

- pure identity/allocation logic; no DB access and no turn orchestration.

## `internal/routing`

Files:

- `agent_id.go`
- `audit.go`
- `classifier.go`
- `features.go`
- `prompt_and_decision.go`
- `route.go`
- `router.go`
- `runtime_flow.go`
- `types.go`
- `routedsession/`
- tests

Ownership:

- route parsing/classification
- model routing features/router
- route decision DTOs and audit publication shape
- runtime flow helpers for route/session preparation
- routed-session resolution orchestration in `routedsession`

Boundary:

- owns routing policy and routed allocation construction; store persistence is called through explicit store/audit APIs and turn execution remains in `internal/turn`.

## `internal/store`

Files/subpackages:

- root store/session/turn files (`store.go`, `schema.go`, `turn_coord.go`, `subturns.go`, `turns_extra.go`, session identity/resolution files)
- `audit/`
- `queue/`
- `cache/`
- `object/`
- `vfs/`
- `internalx/`

Ownership:

- SQLite schema and migrations/current schema creation
- canonical session identity, aliases, main-session and channel-binding persistence
- active-turn claims and turn/session state updates
- steering queue rows
- subturn rows
- turn failure markers
- routing/hook audit persistence under `audit/`
- inbound work queue and dispatcher lease persistence under `queue/`
- storage helpers for cache/object/VFS namespaces

Boundary:

- no runtime policy decisions beyond atomic persistence invariants; callers decide orchestration semantics.

## `internal/web`

Ownership:

- HTTP/SSE/API surface
- guarded runtime endpoints for inbound work queue controls
- topic SSE delivery
- connectivity route auth/adapters
- web runtime dispatcher startup/shutdown wiring
- script bridge host callbacks exposed through web runtime

Boundary:

- must reuse engine/store runtime paths; should not implement separate turn/session coordination logic.

## `internal/tui`

Ownership:

- terminal UI state and rendering
- topic-native active-session consumption
- legacy fallback broadcast handling when no live topic subscription exists
- local session/fork/agent selection helpers

Boundary:

- observes topics and submits through engine APIs; does not own runtime state.

## `internal/tools/script.go`

Ownership:

- script tool execution envelope
- JavaScript/Joker bridge wiring to host callbacks
- session-owned script resources: event hooks, topic subscriptions, sockets, websockets
- script-visible topic APIs and connectivity helpers

Boundary:

- bridge/resource ownership only; runtime decisions stay in `internal/turn`, connectivity dispatch in `internal/connectivity`, and durable state in `internal/store`.

## Cross-package invariants

- SQLite/WAL remains canonical durable runtime state.
- Topic bus is live notification/gap-detection, not replay storage.
- Same-session input serializes through active-turn/steering/queue coordination.
- Direct/IPC/system ingress enters normal route/session/steering/turn paths.
- Cleanup/resource APIs should be idempotent when already-gone is harmless.
