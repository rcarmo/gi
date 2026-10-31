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
replay effects. Automatic recovery of these checkpoints is work for #92.
Ordinary `Close` still uses abort semantics until checkpoint-preserving close
and recovery are implemented; #86 and #87 stay open.

Async results arriving during a staged parent decision use orphan handling,
with a notice that the parent was ending. They do not imply the parent terminal
transaction has already committed.

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
