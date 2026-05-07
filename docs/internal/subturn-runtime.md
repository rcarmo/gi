# Sub-turn runtime contract

## Status
Partially implemented and active in runtime/store paths.

## Purpose
Define how a child turn (sub-turn) is linked to a parent turn so orchestration can be controlled and audited without relying on ad-hoc metadata scans.

---

## Canonical model

A sub-turn is represented as:

- parent turn id (`parent_turn_id`)
- parent session id (`parent_session_id`)
- child turn id (`child_turn_id`)
- child session id (`child_session_id`)
- delivery mode (`delivery_mode`, currently `sync`)
- runtime status (`running|completed|failed|aborted|cancelled`)
- lineage depth (`depth`)
- metadata (JSON)

This is persisted in SQLite table `subturns`.

---

## Persistence decision

Current decision: **ephemeral execution + durable audit trail in SQLite**.

That means:

- execution itself remains in the normal turn runtime path
- parent/child linkage and lifecycle state are persisted durably in `subturns`

This supports crash recovery, observability, and policy enforcement (depth/concurrency limits).

---

## Runtime behavior (implemented)

### Link creation
When `SubmitPrompt(...)` is called with `ParentTurnID`:

1. parent turn is resolved
2. child depth is computed from parent (`parent.metadata.subturn_depth + 1`)
3. limits are enforced:
   - max depth (default `8`)
   - max concurrent running children for parent (default `4`)
4. child turn is created
5. `subturns` link row is created

### Status synchronization
When child turn status transitions, `subturns.status` is synchronized via store update paths.

### Metadata fields
Runtime annotates child/subturn metadata with:

- `subturn_depth`
- `subturn_parent_turn_id`
- `subturn_max_depth`
- `subturn_max_concurrency`

---

## Limits and policy

### Depth
- default max depth: `8`
- overflow error: `subturn depth limit exceeded: depth=<d> max=<m>`

### Concurrency per parent
- default max running children: `4`
- overflow error: `subturn concurrency limit exceeded: running=<n> max=<m>`

### Optional caller overrides
`RunInput.Metadata` may provide:

- `subturn_max_depth`
- `subturn_max_concurrency`

Invalid/non-positive values fall back to runtime defaults.

---

## Store APIs

- `CreateSubTurn(...)`
- `GetSubTurnByChild(...)`
- `ListSubTurnsByParent(...)`
- `UpdateSubTurnStatusByChild(...)`
- `CountRunningSubTurnsByParent(...)`

---

## Current gaps (next steps)

- explicit async result delivery mode semantics
- orphan-result handling contract when parent ends first
- sub-turn tool inheritance/restriction policy
- sub-turn lifecycle publication on canonical topic bus
- cancellation propagation policy for parent abort/timeout

---

## Test coverage (current)

- store lifecycle create/list/update
- running-child count transitions
- engine link creation on parent-driven submit
- depth overflow rejection
- per-parent concurrency overflow rejection
