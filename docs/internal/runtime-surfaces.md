# Runtime surfaces living index

## Status

Living index for the top runtime surfaces ported during the core runtime refactor. Each surface points to the canonical contract/audit document and the owning implementation area.

## 1. Session identity and allocation

Canonical docs:

- `runtime-target-state.md` — relational identity schema and current implementation status
- `routing.md` — route/session allocation semantics and multi-channel binding policy
- `runtime-refactor-adr.md` — coordinated runtime model

Owners:

- `internal/session`
- `internal/store/session_identity*.go`
- `internal/store/session_aliases.go`
- `internal/store/session_channel_bindings.go`
- `internal/routing/routedsession`

Runtime contract:

- canonical identity is relational, not session-row JSON
- aliases and channel bindings resolve through store APIs
- explicit continuation/bindings are allowed; automatic cross-channel merging is not

## 2. Turn coordination

Canonical docs:

- `runtime-refactor-adr.md`
- `runtime-target-state.md`
- `picoclaw-parity-status.md`

Owners:

- `internal/turn/engine.go`
- `internal/store/turn_coord.go`
- `internal/store/turns_extra.go`
- `internal/store/turn_failures*.go`

Runtime contract:

- one active coordinating worker per session
- queued/running/cancelling/terminal states are durable
- cleanup, cancellation, and crash recovery reconcile to store truth

## 3. Steering

Canonical docs:

- `runtime-target-state.md`
- `topic-system.md`

Owners:

- `internal/turn/engine.go`
- `internal/store/schema.go` (`steering_queue`)
- store steering APIs in `internal/store`

Runtime contract:

- same-session input during active turns becomes durable steering
- SQLite is the ordering-critical steering queue; no in-memory fast path
- steering is consumed at explicit checkpoints and publishes `session.steering`

## 4. Hooks

Canonical docs:

- `hooks/lifecycle.md`
- `agentic-loop-hooks.md`
- `runtime-target-state.md`

Owners:

- `internal/turn/engine.go`
- `internal/inference`
- `internal/store/audit/hook_invocations.go`
- `internal/scripting`
- `internal/tools/script.go`

Runtime contract:

- one JSON-safe logical hook request/response contract
- canonical actions: continue, modify, respond, deny, abort_turn, hard_abort
- mounted process hooks share the same contract as in-process/script hooks
- invocation audit rows and `runtime.hook` topics are emitted

## 5. Topic/event layer

Canonical docs:

- `topic-system.md`
- `runtime-refactor-adr.md`
- `runtime-target-state.md`

Owners:

- `internal/topics`
- `internal/turn/engine.go`
- `internal/web/sse.go`
- `internal/tui/chat.go`
- `internal/tools/script.go`

Runtime contract:

- bounded in-memory firehose with monotonic bus-wide sequence numbers
- drop-oldest backpressure, no sticky/replay state
- SSE exposes sequence/gap metadata; durable replay belongs to SQLite APIs
- JS and Joker script bridges expose publish/subscribe/read/unsubscribe

## Additional surfaces

### Sub-turns

Canonical doc: `subturn-runtime.md`.

Owner areas: `internal/turn`, `internal/store/subturns.go`.

### Direct/IPC/system ingress

Canonical docs: `runtime-target-state.md`, `routing.md`, `topic-system.md`.

Owner areas: `internal/turn`, `internal/store/queue`, `internal/web`.

### Multi-channel policy

Canonical docs: `routing.md`, `runtime-target-state.md`.

Owner areas: `internal/session`, `internal/store/session_channel_bindings.go`, `internal/routing/routedsession`.

### TUI runtime consumption

Canonical docs: `tui-stack-evaluation.md`, `tui-pi-parity-plan.md`, `topic-system.md`.

Owner area: `internal/tui`.
