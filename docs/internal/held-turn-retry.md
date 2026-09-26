# Held-turn retry admission

`Engine.RetryHeldTurn` uses a durable reservation before submitting follow-on
work. This is native API hardening, not a new web or terminal control.

## Contract

- A conditional SQLite update reserves one unresolved held terminal failure.
  Concurrent callers cannot both submit it. A failed, aborted or cancelled
  original remains terminal; retry creates a separate turn in the same session.
- `retry_admission_token` is an additive `turn_failures` column, separate from
  the foreign-key `resolved_turn_id`. The follow-on carries the token and
  `retry_of_turn_id` in its metadata.
- Retry uses queue intent. It never turns into steering for another active run.
  The original prompt, model, parent relationship, custom metadata and tool
  restrictions are retained; media and parent/tool restrictions pass through
  native validation again. Previous continuation messages, routing/ingress
  fields, retry fields and the terminal media claim are not replayed.
- Only the reserving caller, after synchronous `SubmitPrompt` returns, calls
  `FinishHeldRetry`. A known error with no matching stored turn releases the
  reservation. One durable matching turn resolves it even if submission
  reported a post-insert error. Read/transaction failures do not release it.
- Other callers use `ReconcileHeldRetry`. A visible turn INSERT is insufficient:
  subturn setup can still roll it back. Recovery also requires the durable
  `turn.submitted` event, written after that rollback boundary. Duplicate
  receipts and receipts in another session remain blocked.
- Resolution and original-turn phase changes commit together. Hold/skip writes
  are transactional; pending and resolved failures reject re-hold/upsert.
  Clear and queued/running/completed transitions reject protected failure rows.
  Stale-claim recovery releases a held turn's claim without replaying its prompt.
- Resolution notifications use the persisted summary. They may repeat as
  invalidations; consumers must read stored state rather than count events as
  exactly-once receipts.

## Conservative recovery limit

A crashed reservation without confirmed admission stays `retry_pending`.
Re-entering retry only checks stored evidence. It does not resend, clear the
hold, or allow skip/re-hold to overwrite the reservation.

`turn.submitted` is currently a warning-only audit write. If it fails and the
owner cannot persist resolution before it dies, a follow-on may exist without
that confirmation. Such a reservation can remain held indefinitely, even when
work actually ran. The same applies to a crash before any turn was inserted.
This is a known recovery/availability gap, not full retry acceptance. There is
no automatic replay, token-forced discard or operator recovery UI in this slice.
A dedicated durable post-rollback receipt or another safely fenced recovery
protocol remains work before exposing terminal retry controls.

A review raised this missing-audit case as a blocker to complete recovery. The
implementation deliberately retains it as a blocked state instead of treating
row absence/presence as permission to resend. Tests exercise the production
warning-only audit failure together with failed resolution and database reopen.

## Verification

All tests use disposable databases, never the live instance.

- `make test-held-retry`: store/engine race checks, three repeats. Covers two
  engines/connections; pre-insert, post-insert and resolution failures; held
  stale-claim recovery; metadata filtering; active-run queue isolation; wrong,
  duplicate and foreign tokens; protected clear/upsert/re-hold/status paths;
  schema upgrade/reopen; and a concurrent observer while subturn rollback is
  paused. Missing submission audit plus failed resolution remains held on reopen.
- `make test vet bun-checks`: core suite, vet and hook checks.
- `make test-ux BIN_DIR=/tmp/gi-held-retry-functional-bin`: full isolated build and
  functional suite. 107 passed, 11 existing skips.

The focused race gate is included in required CI. Whole-CI status and deployment
are tracked separately; this document does not claim deployment or new feature
mappings. No terminal rows, controls, shortcuts or clipboard behaviour change.
