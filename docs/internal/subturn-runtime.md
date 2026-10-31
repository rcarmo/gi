# Sub-turn runtime contract

## Status
Implemented for current sync/async delivery modes, parent-terminal cancellation propagation, and restricted tool inheritance/runtime filtering (with broader future policy refinements still possible).

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

The rows support observability and depth/concurrency enforcement. Interrupted checkpoint recovery is tracked in #92; linkage alone does not authorise execution or replay.

---

## Runtime behavior (implemented)

### Link creation
When `SubmitPrompt(...)` is called with `ParentTurnID`:

1. parent turn is resolved
2. child depth is computed from parent (`parent.metadata.subturn_depth + 1`)
3. limits are enforced:
   - max depth (default `8`)
   - max concurrent running children for parent (default `4`)
4. delivery mode is normalized (`sync` by default, explicit `async` supported)
5. child turn is created
6. `subturns` link row is created unless the live parent is settling or cancelling; a rejected link rolls back the new child turn
7. lifecycle event `subturn_created` is published on `turn.subturn`

### Status synchronization
When child turn status transitions, `subturns.status` is synchronized via store update paths.

### Result delivery modes

#### `sync` (default)
- child completion publishes `subturn_result_delivered` on topic `turn.subturn`
- if parent and child sessions differ, a parent-session system message is appended with
  - `kind: subturn_result`
  - `delivery_mode: sync`
  - parent/child ids and summary

#### `async`
- child completion publishes `subturn_result_ready` on topic `turn.subturn` when the parent is still running without a staged terminal decision
- no parent-session result message is appended automatically in the non-orphan case
- caller/automation can later fetch child artifacts and decide when/how to surface them

##### Orphan async completion handling
If async child completion occurs after the parent is terminal, cancelling or has staged its terminal decision:

- runtime marks subturn metadata with orphan markers:
  - `orphaned: true`
  - `orphaned_at`
  - `orphan_reason`
- publishes `subturn_orphaned` on `turn.subturn`
- appends parent-session system message with:
  - `kind: subturn_orphan_result`
  - `delivery_mode: async`
  - parent/child ids and summary

### Metadata fields
Runtime annotates child/subturn metadata with:

- `subturn_depth`
- `subturn_parent_turn_id`
- `subturn_max_depth`
- `subturn_max_concurrency`
- `subturn_critical`
- `effective_tools`
- `subturn_tools_restricted`

### Tool inheritance / restriction

#### Default inheritance
- child turns inherit the parent turn's persisted `effective_tools`
- if parent metadata has no `effective_tools`, runtime falls back to the engine's current active tool set

#### Explicit restriction
Child creation may provide either:
- `subturn_tools`
- `subturn_allowed_tools`

Behavior:
- explicit lists must be a subset of the parent turn's `effective_tools`
- unknown tool names are rejected
- the resolved child turn tool set is persisted in `effective_tools`
- `subturn_tools_restricted=true` marks that explicit restriction was applied

#### Runtime enforcement
- model-visible tool definitions for a child turn are filtered from `effective_tools`
- tool execution also re-checks the same persisted set, so a provider cannot invoke a tool outside the child turn's allowed scope merely by hallucinating the name

### Parent-terminal cancellation propagation

#### Graceful parent finish (`completed`)
- non-critical child subturns receive cancellation requests
- critical child subturns are allowed to continue

#### Non-timeout failure (`failed`)
- non-critical child subturns receive cancellation requests
- critical child subturns are allowed to continue

#### Hard abort (`cancelled`/`aborted`)
- child and descendant subturns receive cancellation requests regardless of criticality

#### Timeout failure (`failure_kind` contains `timeout`/`deadline`)
- child and descendant subturns receive cancellation requests regardless of criticality

#### Settlement order

The parent commits `turn.terminal_staged` and phase `settling` before requesting child cancellation. This checkpoint retains its proposed outcome and system message; it does not publish terminal state or release the claim. New child admission is fenced while the parent is settling or cancelling.

Before terminal commit, the coordinator joins required child hosts through their hook/cleanup boundary and checks committed terminal receipts. Critical children may continue after a completed or non-timeout failed parent. Hard abort and timeout require all children. An abort received during a graceful join upgrades the decision and joins critical children too. Cancellation marks and their checkpoints commit atomically; a late cancellation cannot reopen a terminal child.

Missing child receipts, ownership cycles or failed settlement leave the parent checkpoint and claim intact without a live worker. A required foreign child with an active claim also lacks joined-host proof, even if its terminal receipt exists. Recovery or review must resolve these before new work passes that ownership. The runtime does not force foreign children terminal or replay their effects.

Stale-claim recovery can resume a committed terminal decision under a fresh, atomically fenced claim. It preserves the staged message identity and usage, bypasses prompt setup and effect dispatch, and repeats the required-child join before terminal commit. Fresh heartbeats and live local hosts prevent takeover; superseded recovery writers cannot settle after the successor releases ownership. General generation/tool recovery and restart parking require further #92 work. Main hosts stop and join their heartbeat writers before claim cleanup.

#### Cancellation audit markers
When a parent-triggered child cancellation is requested, subturn metadata is patched with:

- `cancel_requested_by_parent`
- `cancel_requested_at`
- `cancel_requested_parent_turn`
- `cancel_requested_parent_status`
- `cancel_requested_failure_kind`
- `cancel_reason`

Runtime also emits `subturn_cancel_requested` on `turn.subturn`.

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
- `StageTurnTerminal(...)`
- `HasTurnTerminalReceipt(...)`
- `RequestTurnCancellation(...)`

---

## Current gaps (next steps)

- broader future policy refinements (for example agent-config-driven presets or capability classes) if needed

---

## Test coverage (current)

- store lifecycle create/list/update
- running-child count transitions
- engine link creation on parent-driven submit
- depth overflow rejection
- per-parent concurrency overflow rejection
- delivery mode validation (`sync`/`async`, invalid-mode rejection)
- sync vs async result-delivery behavior in parent session history
- async orphan completion handling (durable metadata marker + parent notice message)
- graceful parent finish cancellation propagation
- hard-abort descendant cancellation propagation
- timeout-driven cancellation propagation for critical child subturns
- required child receipt/cleanup ordering before parent commit
- abort during graceful join upgrades critical-child cancellation
- staged decision rollback/reopen and child-admission fencing
- missing foreign receipts retain parent ownership; cycles fail closed
- late cancellation cannot reopen terminal state during observer cleanup
- child turn effective tool inheritance
- explicit child turn restricted tool-set enforcement
