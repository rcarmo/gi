# ADR: SQLite-backed coordinated runtime model

## Status

Accepted for the current Gi core runtime refactor.

## Context

Gi needed to tighten agentic-loop correctness while preserving its existing web/TUI/tool contracts and keeping SQLite/WAL as canonical shared state. PicoClaw provided useful reference semantics for active-turn ownership, steering, hooks, direct ingress, and lifecycle events, but Gi deliberately does not adopt PicoClaw's JSONL session storage model.

## Decision

Gi's runtime model is a coordinated, DB-backed runtime with a small live topic layer:

1. **Session identity** is relational and canonical.
   - Canonical identity is `(agent_id, channel, account, dimensions, dimension values, canonical scope signature, opaque session key)`.
   - Aliases and channel/account bindings resolve to canonical sessions through store APIs.
   - `sessions.scope_json` is not a runtime source of truth.

2. **Active turns** are store-coordinated.
   - Only one coordinating worker may own the active turn for a session.
   - Queued turns, active claims, cancellation, cleanup, and crash recovery use durable state.
   - In-memory runner state is only a local execution handle/cache and must reconcile to store truth.

3. **Steering** is durable and session-scoped.
   - Same-session input during an active turn becomes steering, not a competing turn.
   - Steering uses SQLite for ordering-critical enqueue/dequeue behavior; there is no in-memory fast path.
   - Steering is observed only at explicit checkpoints.

4. **Sub-turns** use normal turn execution plus durable lineage/audit rows.
   - Parent/child relationships, status, delivery mode, depth, and metadata live in SQLite.
   - Execution remains in the normal turn runtime path.

5. **Hooks** share one JSON-safe logical contract.
   - Go, script, and process hooks use the same request/response DTO semantics.
   - Process hooks run as mounted persistent JSON-RPC subprocess sessions.
   - Provider-level request/response interception is wired through `internal/inference`.

6. **Runtime topics** are the canonical live event layer.
   - The topic bus is a bounded in-memory firehose with monotonic sequences and drop-oldest backpressure.
   - It is not a persistent replay log; durable replay/state reconstruction belongs to SQLite APIs.
   - Runtime families publish under canonical topics (`runtime.turn`, `runtime.session`, `runtime.tool`, `runtime.hook`, `runtime.routing`, `runtime.inbound_work`, `runtime.dispatcher`, `session.steering`, `turn.subturn`, etc.) plus aggregate `runtime` notices where appropriate.

7. **Direct/IPC/system ingress** uses the same runtime path.
   - `DirectInput` / `DirectOrigin` normalize non-web/TUI input.
   - Direct prompt, peer-message, continue, system, and internal input flow into normal routing/session/steering/turn coordination.
   - Durable inbound queue rows provide the first queued ingress surface, with retry/backoff and dispatcher lease ownership.

8. **Multi-channel sessions** are explicit-policy only.
   - Existing bindings and explicit continuation can attach multiple channel/account/chat identities to one logical session.
   - Gi does not automatically merge unrelated channel identities or fan out responses to every bound channel.

## Cut-over boundary

The cut-over boundary is the turn engine/store interface:

- Web/TUI/script callers may keep their existing external contracts.
- Runtime coordination decisions must go through store-backed primitives or documented engine helpers.
- Compatibility wrappers/shims are not retained for internal runtime paths once direct call sites have moved.

## Durable vs in-memory state

Durable in SQLite:

- canonical session identities, aliases, and channel bindings
- turn rows, turn phases/statuses, active-turn claims, queued turns, and failure markers
- steering queues
- subturn lineage and delivery metadata
- hook invocation audit rows
- route/routing event audit rows
- inbound work queue rows, retry state, and dispatcher leases

In memory:

- local goroutine cancellation handles and runner bookkeeping
- live topic subscribers and bounded buffers
- mounted process-hook handles
- transient provider/tool execution data

In-memory state must be reconstructable, disposable, or reconciled from SQLite state after crash/restart.

## Consequences

- SQLite/WAL remains the single source of durable runtime truth.
- Topic consumers get low-latency live updates but must use store APIs for replay or reconciliation.
- Same-session correctness is favored over opportunistic parallelism.
- Multi-channel behavior is deterministic and explicit rather than heuristic.
- Hook/process/script extension points share a single runtime contract.

## Related docs

- `runtime-target-state.md`
- `routing.md`
- `topic-system.md`
- `hooks/lifecycle.md`
- `subturn-runtime.md`
- `picoclaw-parity-status.md`
