# 0061: Receipt accounting and terminal settlement

Status: accepted

## Decision

SQLite commits each received provider response with its usage, latest context
measurement and next execution boundary. `turn_receipts` uses
`(turn_id, occurrence_id)` as its primary key. Provider occurrence IDs match the
committed request intent; tool occurrence IDs match the execution intent.
Repeated settlement returns the first committed receipt without adding usage.

A direct answer selects `reply_boundary`, which checks steering before finishing.
A tool-call answer selects `tools`. Failed attempts select `error_boundary` for
the retry/failure policy. These checkpoints retain the adopted response and do
not authorise another provider request or an unsafe tool replay. Checkpoint
recovery is tracked separately in #92.

Tool settlement commits the result, terminal event, nested-model usage and next
tool boundary together. Cancelled/aborted tool boundaries retain known usage and
mark additional remote usage as uncertain. A model call without counters marks
its enclosing tool receipt uncertain even when the script catches the error.

`SettleRunningTurn` commits status, phase, finish time, sub-turn status, system
message, failure metadata, derived usage event, finish event and idle session
projection in one transaction. A stored cancellation request wins over success.
An older turn cannot clear a newer turn's session ownership. Runtime publication
starts after commit; a failed transaction publishes no terminal outcome.

## Accounting

Totals are rebuilt from committed receipts. Integer overflow, negative counters
and non-finite/negative costs reject settlement. The metadata total and the
terminal `inference.finished` event use the same receipt set.

`unknown_receipts` includes missing counters and admitted provider/tool work
without a matching receipt. Provider counter objects containing only zero token
counts also carry uncertainty because the adapter cannot distinguish missing
counters from an explicit zero. Known partial counters and costs are retained.
This provides local accounting consistency; remote billing can still be unknown
following interruption. Historical pre-ledger events remain unchanged; Gi does
not invent missing historical counters.

## Verification

Fault-injection tests reject each provider/terminal/tool transaction boundary,
check rollback, retry settlement and reopen the database. Real engine regressions
cover successful provider and nested-model work followed by failure/cancellation,
late unadopted responses and suppressed publication after a failed commit.
