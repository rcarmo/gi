# 0063: Child-before-parent settlement

Status: accepted; incremental #86 repair

## Decision

The coordinator commits `turn.terminal_staged` and phase `settling` before
requesting child cancellation. The checkpoint retains the proposed status,
message identity, system message, payload and failure kind. It leaves terminal
status, finish time and the active claim unchanged. Child link insertion checks
the parent's live phase in the same SQLite write, including held-retry admission.

Completed and non-timeout failed parents cancel and join non-critical children;
critical children can continue independently. Hard abort and timeout require all
children and descendants. A cancellation arriving during a graceful join upgrades
the decision and includes critical children before parent terminal commit.

Each launched main/manual host has a completion fence in the shared session
coordinator. The fence spans final hooks and cleanup, including queue handoff;
it is independent of the session's current-turn pointer. A queued child can
settle through its cancellation checkpoint without starting a provider host.

`SettleRunningTurn.ChildrenStatus` checks that the joined child policy still
matches the effective terminal decision. A concurrent abort forces another
child-policy pass before terminal publication. Cancellation marks and their
checkpoints commit atomically against settlement, so late cancellation cannot
reopen terminal state. Queued cancellation also checks that no owner claimed the
turn before the update.

Missing foreign receipts, ownership cycles and failed settlement retain the
parent claim and checkpoint, clear its in-process worker pointer and record a
cleanup diagnostic. The coordinator does not manufacture child completion or
replay effects. Staged-terminal recovery uses the retained decision as described
below. Ordinary `Close` still uses abort semantics until checkpoint-preserving
close is implemented; #86 and #87 stay open.

Async results arriving during a staged parent decision use orphan handling,
with a notice that the parent was ending. They do not imply the parent terminal
transaction has already committed.

## Staged terminal recovery

Recovery adopts a staged decision only after comparing the stale claim's token
and heartbeat with the current SQLite row. Claim rotation, token-history
registration, a persistent `turn_recovery_fences` writer fence and
`turn.recovery_admitted` commit together. Fresh heartbeats, replaced claims and
live local hosts prevent takeover. A failed transaction retains the old claim.

The recovered main host bypasses prompt setup, routing and provider/tool
dispatch. It preserves the staged message identity, payload and committed usage,
then applies the existing child policy and terminal transaction. Explicit abort
still wins. Missing or malformed decisions retain ownership without replay.
Runtime terminal writers supply their claim incarnation; a superseded recovery
writer cannot stage or settle even after the successor releases its claim.
Legacy lost-claim cancellation can still retire an unrecovered abandoned turn.

An unowned child with a terminal receipt and an active claim has no joined host
proof. Its parent retains the checkpoint until that owner releases the claim.
Each main host now stops and joins its heartbeat writer before claim cleanup.

Generation/tool continuation, replay-safety pairs, retry sleep and
progress-based restart parking still require #92/#90 implementation. Existing
whole-turn holds cover those interruptions.

## Verification

Focused store and real-engine regressions pass twice with race detection:

- staged-decision rollback, wrong-session rejection, SQLite reopen and child
  admission fencing;
- required host/receipt/claim cleanup before parent commit, including a queued
  child in the parent's session;
- completed/failed/aborted parent policy and abort during a graceful join;
- late cancellation during terminal observer execution;
- retained ownership after failed staging or an unresolved foreign child;
- rejection of ownership cycles.

Cross-turn commit assertions use the event table's insertion identity; `seq`
is local to each turn. Concurrent engine fixtures use the production file-backed
WAL/immediate-transaction path. Initial shared-memory probes returned SQLite
table-lock/deadlock errors; those probes did not establish concurrent durability
acceptance. Shared-memory transactions keep their existing table-locking limits.

Full store, turn, inference and web packages, vet and diff checks pass. These are
ordinary development checks; no release, deployment or full browser-matrix run.
