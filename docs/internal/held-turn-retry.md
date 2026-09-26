# Held-turn retry admission

`Engine.RetryHeldTurn` uses a durable reservation before submitting follow-on
work. This is native API hardening, not a new web or terminal control.

## Contract

- A conditional SQLite update reserves one unresolved held terminal failure.
  Concurrent callers cannot both submit it. A failed, aborted or cancelled
  original remains terminal; retry creates a separate turn in the same session.
- `retry_admission_token` is an additive `turn_failures` column, separate from
  the foreign-key `resolved_turn_id`. The follow-on carries the token and
  `retry_of_turn_id` in its metadata. `retry_admission_version=1` marks the new
  atomic protocol; existing reservations migrate to version zero.
- Retry uses queue intent. It never turns into steering for another active run.
  The original prompt, model, parent relationship, custom metadata and tool
  restrictions are retained; media and parent/tool restrictions pass through
  native validation again. Previous continuation messages, routing/ingress
  fields, retry fields and the terminal media claim are not replayed.
- The engine validates prompt/media/parent/tool inputs, then `AdmitHeldRetry`
  checks the reservation token/version/session and terminal original status
  inside a writer transaction. It commits the queued turn, optional subturn
  link, failure resolution, original phase and queue count together. No
  provisional queued row is visible to another runner. An insertion, subturn,
  resolution or queue-count failure rolls everything back.
- Only the reserving caller, after synchronous submission returns, calls
  `FinishHeldRetry`. A known error with no matching stored turn releases the
  reservation. Read/transaction failures do not release it. A committed
  version-one admission is already resolved, even if later launch/audit work
  fails. Repeating retry after a lost successful reply returns that same ID.
- Public `SubmitPrompt` strips caller-supplied `retry_*` metadata. Only the
  internal retry path passes a separate admission object to the store.
- Legacy version-zero callers use `ReconcileHeldRetry`: a visible turn INSERT
  is insufficient because old subturn setup could still roll it back. That
  path still requires the durable `turn.submitted` event. Duplicate receipts,
  foreign-session receipts and inconsistent version-one pending rows remain
  blocked.
- Resolution and original-turn phase changes commit together. Hold/skip writes
  are transactional; pending and resolved failures reject re-hold/upsert.
  Clear and queued/running/completed transitions reject protected failure rows.
  Stale-claim recovery releases a held turn's claim without replaying its prompt.
- Resolution notifications use the persisted summary. They may repeat as
  invalidations; consumers must read stored state rather than count events as
  exactly-once receipts.

## Explicit pre-admission recovery

A crashed reservation without confirmed admission stays `retry_pending`.
Re-entering retry only checks stored evidence; it does not resend or clear it.
For version one, `ReleaseUnadmittedHeldRetry(sessionID, turnID, token)` is a
separate explicit store action. It clears only the matching pending reservation
with no receipt. Admission and release serialize on the file-backed SQLite
writer lock. If release wins, the old caller's token can no longer admit; if
admission wins, release fails without changing committed work. A new explicit
retry may reserve a new token. Release itself never submits or deletes a turn.

Version zero remains conservative: a warning-only `turn.submitted` write can
be lost along with the old caller's resolution. Such a reservation can remain
held indefinitely; the new release method refuses it because an old writer
cannot be fenced. Migration never relabels old reservations as version one.
Mixed old/new binaries writing one database are unsupported for the new fence;
stop old writers before upgrade. There is no token-forced legacy discard.

The new atomic path no longer relies on `turn.submitted` for correctness. Tests
fail that audit and resolution independently, reopen the database, and recover
the same committed ID without replay. No new web or terminal recovery controls
are exposed yet; session guards, full-ID command design and PTY acceptance
remain separate work.

## Verification

All tests use disposable databases, never the live instance.

- `make test-held-retry`: store/engine race checks, three repeats. Covers two
  engines/connections; pre-insert, post-insert and resolution failures; held
  stale-claim recovery; metadata filtering; active-run queue isolation; wrong,
  duplicate and foreign tokens; protected clear/upsert/re-hold/status paths;
  schema upgrade/reopen; and a concurrent observer while subturn setup is
  paused. Atomic tests cover full rollback, committed subturn recovery without
  an audit, eight two-connection admission/release races, stale-token fencing
  after release/re-reservation, legacy migration refusal, a paused engine owner,
  forged public metadata and lost-success recovery.
- `make test vet bun-checks`: core suite, vet and hook checks.
- `make test-ux BIN_DIR=/tmp/gi-retry-atomic-functional-bin`: full isolated build and
  functional suite. 107 passed, 11 existing skips.

A focused independent review found no blocker on the file-backed path. The
race acceptance uses `_txlock=immediate` file-backed stores, not shared-memory
SQLite contention or physical terminal acceptance.

The focused race gate is included in required CI. Whole-CI status and deployment
are tracked separately; this document does not claim deployment or new feature
mappings. No terminal rows, controls, shortcuts or clipboard behaviour change.
