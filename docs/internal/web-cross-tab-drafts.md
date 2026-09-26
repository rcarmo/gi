# Cross-tab draft write fencing

Each session's browser draft now has a revision. IndexedDB writes compare that
revision and replace the row within one read/write transaction. A stale page
cannot silently erase another page's pending send, queue-return marker or edits.

This is a fail-closed policy. It does not merge concurrent edits or coordinate
which tab is active. When a session conflicts, this page retains its local edits,
reports the conflict and refuses further persistence or submission for that
session. Copy unsaved text and attachments before a full reload. Other sessions
can still persist. A capture's durable readiness still gates all uploads/POSTs;
queue return still requires durable recovery before deleting a queued turn.

## Writes and startup

The adapter encodes attachment bytes before opening the transaction. It then
reads the current row, compares the expected revision, writes revision + 1 and
resolves only after commit. Missing legacy revisions are zero; malformed or
exhausted revisions cannot write. Each page tracks committed revisions separately
from queued snapshots, so its own consecutive writes do not conflict.

Startup stages receipt reconciliation outside the public draft map. If its
recovery write loses a revision race, startup rejects and exposes no speculative
recovered text. A page whose load failed cannot send; a full reload is required.
Bootstrap retry does not reset this guard or discard local conflict edits.

IndexedDB version 2 prevents older version-1 unconditional writers from bypassing
the fence. The existing adapter closes on versionchange; its subsequent writes
fail. If another connection refuses to close, startup reports the blocked upgrade
and cannot submit. A late successful open after a blocked failure is closed to
avoid leaking an inaccessible connection. Legacy rows and attachment bytes are
preserved. Test introspection opens use the existing version rather than requesting
version 1; explicit migration/fencing tests still create version-1 databases.

## Limits

If unknown recovery commits before another tab confirms delivery, the later
acknowledgement cleanup conflicts. The unknown text and warning remain. This
change does not make confirmation win every ordering, add POST idempotency,
automatically resend, or synchronise open editors. Missing/duplicate/legacy/late
receipts and captures outside the six-capture lookup budget remain unknown.

The persisted token is shared with idle and queued HTTP sends by the guarded
`patch-compose-capture-token.mjs` build adapter; supplied composer source remains
unchanged. Builds must use the normal build pipeline.

## Verification

Two repository regressions failed before the change: stale updates erased a
foreign pending token, and stale unknown recovery restored text after confirmed
cleanup. Both now reject the stale write.

- `make test-web-http-helpers`: 24 tests, 119 assertions. Covers both races,
  same-tab queued writes, per-session freezes, queue-return fences, failed capture
  restoration, legacy/malformed revisions and full-reload recovery.
- `make test-web-basic-send`: 32 Chromium/WebKit HTTP cases, including two real
  pages sharing IndexedDB, no POST from the conflicting page, captured attachment
  byte preservation, stale recovery fencing, legacy migration, open old writers
  and blocked upgrades. Existing exact-token closed-page recovery stays covered.
- `make test-ux`: 139 passes, 11 existing skips.
- `make test-web-basic-controls`: 24 passes.
- `make test-ux-steer UX_LOCAL_ENV= UX_LOCAL_FUNCTIONAL= UX_LOCAL_SPEC=tests/ux/drafts.spec.mjs`:
  168 passes.
- `make test-ux-parity UX_PARITY_ARGS=tests/ux/queue-return.spec.mjs`: 18 passes.
- `make test vet bun-checks`: passed; supplied components/ui/panes unchanged.

Focused final review accepted this bounded scope. An earlier review overlooked
the build token adapter and questioned the retry policy; the adapter was verified,
and reload-only recovery is now explicit and tested. Initial queue-return runs
used the wrong lightweight fixture (`UX steer gate` versus `UX queue gate`);
canonical parity-fixture results above supersede those harness failures. One
outer shell cutoff interrupted the long draft suite; its standalone run passed.
No browser timeout or assertion was weakened. Evidence: `/workspace/tmp/gi-cross-tab`.

Whole-product CI and exact-source deployment are still required. Live `a1be540`
is unchanged while this slice is packaged.
