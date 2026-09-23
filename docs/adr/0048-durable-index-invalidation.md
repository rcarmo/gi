# ADR-0048: Durable index invalidation revisions

## Status

Accepted — 2026-09-23. Internal storage now records scoped invalidation revisions and protects changes arriving during refresh. Watchers, mutation hooks and background scheduling are **not connected**. Manual indexing and query semantics remain as in ADR-0046/0047. Frozen coverage remains 45/236 Classic and 2/42 shared, with 191/40 unmapped.

## Additive v2 migration

`002_scope_invalidation.sql` creates `workspace_index_invalidations`, keyed by workspace/scope, with nonnegative requested and acknowledged revisions and a foreign key to the scope. Existing scopes receive zero/zero counters. The immutable v1 SQL and checksum are unchanged; existing content, memberships, status, generation and timestamps are preserved.

The migrator now validates a contiguous ordered ledger, validates the already-applied objects, applies missing versions and validates the final schema in the existing core transaction. Future versions, ledger gaps, checksum mismatches and conflicting unversioned objects fail closed. A late ledger-write failure after v2 DDL/backfill rolls all v2 work back. Reopen applies neither migration twice.

## Invalidation and publication

`RefreshStore.Invalidate` takes a resolved scope and a bounded list of clean relative paths. Files, removed/renamed directories and ancestors of roots can invalidate a scope. An empty list represents a deliberate whole-scope event such as watcher overflow. Paths are compared against both requested and last committed roots during a configuration transition. Events outside those roots do not create state. The caller must invoke the method for each relevant scope; no implicit fan-out or filesystem access occurs.

A successful invalidation increments the requested revision in a transaction and changes ready→stale. It never changes committed content/count/generation/last-success time, does not acquire the refresh lease, and retains indexing/failed/never-indexed state and failure detail. The counters tell a future scheduler that additional work exists without clearing an error just because another event arrived. Multiple events may accumulate; coalescing belongs to the scheduler.

`Begin` captures the requested revision after acquiring ownership and before scanning. Successful `Commit` acknowledges only that captured revision in the same transaction as content/membership publication. If a newer revision exists, content can publish but the scope remains stale. If an event races with the write transaction, SQLite serialises it either before the revision check or after commit; neither ordering loses the event. Failure, cancellation or rollback does not acknowledge pending revisions.

A shared-document edit by another scope also increments the affected scopes' requested revisions transactionally. Their failed status is retained; otherwise they become stale. This turns the pre-existing shared-edit stale marking into durable work for future scheduling. Reopen preserves counters. Existing ownership/expiry fences remain in force; invalidation cannot let a stale owner publish.

Path matching is conservative. In-root changes to currently unsupported file types may still request refresh, avoiding ambiguous deleted-directory inference. This is an internal operation, not a public unauthenticated event API. It provides no watcher delivery guarantee, background freshness or bounded retry/backoff by itself.

## Evidence

- Initial invalidation marks ready stale and preserves the committed snapshot; events during held scans remain pending after publication until a later refresh.
- Failed writes/failure reporting preserve counters across reopen and keep FTS content unchanged.
- Unrelated scopes/workspaces remain unchanged; renamed/deleted parent paths and old committed roots are matched during configuration changes.
- Concurrent invalidations from two stores retain every revision. Commit/event races in both SQLite serialisation orders leave pending work.
- A forced status-write failure rolls back the earlier revision increment, proving invalidation is atomic.
- A worker held after scanning receives a native file edit plus durable invalidation, publishes its captured old snapshot as stale, then a subsequent worker publishes the new bytes as ready.
- v1→v2 tests retain ledger time, scope state/generation and FTS; late-v2 rollback, concurrent creation, missing-object detection and reopen pass.

A separate SQLite backup of the dev database passed migration and integrity checks. All **30 pre-existing tables plus the v1 ledger row** were unchanged. The comparison script initially assumed all FTS shadow tables had rowid; it was corrected to compare canonical row sets without displaying content. Private backup paths are recorded in `/workspace/tmp/gi-index-invalid-rehearsal.txt`.

Full Go/vet/build/hook checks pass with **76/76 functional**, **32/32 helpers**, search-store/indexer/core-store race tests ×3, **24 focused browser executions** (18 workspace + 6 configured-root), and TUI smoke/Gherkin. The full 498-browser and three-size PTY matrices were not rerun or recredited for this storage-only slice. No supplied UI or frozen features changed. Derived index scenario 018 adds the revision guarantee; there are **18 proposal scenarios**, not frozen mappings.

Logs are `/workspace/tmp/gi-index-invalid-*`. No new screenshot is needed because no UI changed.

## Next integration

Add an application-owned coalescing scheduler with cancellation/drain before database close, bounded retries/backoff and cross-process busy handling. Wire mutation/watch invalidations before ordinary searches can request nonblocking background work. A stale state must remain visible until the captured revision is fully acknowledged; work arriving during a scan schedules another pass. Keep terminal status/actions temporary with unchanged idle rows and independent three-size acceptance.
