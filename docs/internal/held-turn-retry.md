# Held-turn retry admission

`Engine.RetryHeldTurn` uses a durable reservation before submitting follow-on
work. The native API is also used by [explicit terminal retry commands](tui-held-retry.md);
there is no new browser recovery control.

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

## Automatic provider retry continuation

A committed retry wait with versioned lifecycle policy can resume after a stale
owner is fenced. It keeps its iteration, retry attempt, absolute deadline and
tool-failure budget. Current context, settings and tool definitions are rebuilt;
setup, the original prompt, start hooks and settled tools are not replayed.
A committed wake can also resume if no later provider/tool/terminal boundary
exists; recovery does not repeat the wake. Legacy/invalid policy and unknown
effects stay held for review. Restoring a future wait does not reload settings;
current full settings validation runs after the deadline, before generation.

Graceful engine close preserves valid retry waits and committed wakes in dedicated
queued checkpoint phases. It fences the old writer before cancellation and joins
owned work before requeue. Reopening waits for runtime-ready recovery; the normal
prompt queue cannot claim these checkpoints. Shutdown leaves the original
receipt, lifecycle, deadline and budgets intact, without an agent-end callback
or a fabricated crash failure. Generation-ready and single pending-provider
frontiers also support graceful Close; unproven tool/legacy phases still use
conservative abort settlement.

Embedding hosts call `Engine.ResumeInterruptedWork(ctx)` after registering
runtime tools/hooks and enabling MCP. HTTP and TUI startup do this explicitly;
normal prompt admission also opens the barrier after waiting for MCP. The
constructor does not dispatch effectful continuations. Cancellation while waiting
still ends the lifecycle once. See the
[retry continuation audit](../implementation/audits/provider-retry-continuation.md).

## Generation-ready continuation

A version-one `generation.ready` frontier can resume before provider admission.
The writer commits it after context preparation and verifies that all preceding
provider/tool admissions have atomic receipts. Recovery rotates the stale claim
and fences the old writer in the same transaction. Unsupported policies, later
unrecognised events, cancellation and held/resolved failures reject the route.

The resumed host bypasses prompt setup, start hooks and completed tools. It
rebuilds context and tool definitions from current committed sources, reapplies
accepted lifecycle policy, and retains iteration and repeated-tool-failure
budgets. The current iteration limit cannot expand the saved limit. Current
generation/request hooks run during fresh preparation; the old prepared request
is never dispatched again.

At most five automatic admissions are allowed without meaningful provider/tool
receipt progress. Rewriting preparation or recording an interruption does not
reset this limit. Exhaustion uses the existing conservative hold; explicit
parked-checkpoint retry requires further implementation. See the
[generation-ready audit](../implementation/audits/generation-ready-continuation.md).

## Graceful generation Close

Close reserves a verified generation-ready or single pending-provider frontier
and revokes the host's general writer before cancellation. Its owned recorder can
still flush bounded provider snapshots through a private persistence callback;
the provider and hooks retain their revoked contexts. Close keeps the replacement
lease alive until main, hook, side, maintenance, peering and lease hosts have
joined. It then seals any uncertain request and publishes
`checkpoint_generation_ready` with the original lifecycle and budgets.

The ordinary prompt queue cannot select or claim this phase, including after
claim release. Runtime-ready recovery can adopt a completed Close immediately,
without waiting for stale-owner expiry. `generation.close_resumed` preserves the
existing crash counter: a graceful restart neither consumes nor resets it.
Invalid checkpoints stay held. Persistence failures return through the shared
Close barrier and retain the reservation for conservative recovery.

See the [generation Close audit](../implementation/audits/graceful-generation-close.md).

## Interrupted admitted generation

One unsettled provider admission immediately after a versioned generation-ready
frontier can resume. Recovery atomically fences the dead owner, commits an
`interrupted_boundary` receipt, retains the structured committed prefix on its
existing unadopted row, and writes a new ready frontier. `inference.interrupted`
is accounting, not evidence of generation progress. Missing usage remains
unknown; received usage is retained without claiming a final bill.

The resumed host builds a fresh request with current configuration. It never
resends the saved provider payload or executes a partial tool-call prefix.
Settled responses, tool admissions, retry boundaries without a new generation
frontier, and ambiguous/legacy states still require their own recovery routes.
See the [provider interruption audit](../implementation/audits/provider-interruption-continuation.md).

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
the same committed ID without replay. Terminal commands now expose explicit
check/run/release with session guards and full IDs; see their separate PTY
acceptance. Browser controls and broader recovery workflows remain separate.

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

A later product CI run36225954462 failed during shutdown: subturn lifecycle
publication passed a nil coordination context to `database/sql`, panicked while
its mutex was held, then blocked cleanup/Close until the10-minute test timeout.
Lifecycle publication now returns when both contexts are cancelled; a direct
regression asserts no stored subturn mutation. Race×3/core/vet/hooks and107
functional tests (11existing skips) pass. The failed trace is retained; no
budget or timeout was increased. This fix awaits its own CI/deployment.

The focused race gate is included in required CI. Whole-CI status and deployment
are tracked separately; this document does not claim deployment or new feature
mappings. Native admission itself changes no terminal rows, shortcuts or
clipboard behaviour; terminal command acceptance is recorded separately.
