# ADR-0051: Native write invalidation producers

## Status

Accepted — 2026-09-23. Engine `write`, HTTP `write` and script `gi.writeFile` now use one filesystem writer and record scoped invalidation. They do not request scheduler work. Explicit Reindex still controls publication; GET status/search remain read-only. There is no native `edit` executor to connect in this slice.

## Shared write path

`tools.WriteFile` accepts the caller's captured runtime configuration. It resolves `all`, `notes` and `skills` with the same extra-root, extension, optional-root and chunker policy as indexing. The engine supplies its runtime configuration rather than only a workspace string. HTTP and the shared JS/Joker bridge call the same implementation. VFS writes retain their managed database path and inferred content type; they create no filesystem index revisions.

Filesystem writes validate configuration before mutation. Invalid index settings return `write not attempted: index configuration: ...`; they no longer permit a native filesystem write with untracked index state. Chat startup remains available. Fix settings and restart before retrying. VFS writes remain independent of filesystem index configuration.

The writer opens an `os.Root`, requires a local relative destination, creates parents and writes a regular file. Existing symlink components and nonregular targets are rejected. Linux/macOS opens use `O_NOFOLLOW|O_NONBLOCK`, check file type before truncation, and do not block on a FIFO. This intentionally narrows the old write behaviour: writing through an existing symlink alias is no longer supported. The root confines concurrent path replacement; parent-path rename and hard-link alias attribution still require external/watch reconciliation. This slice does not change shell capabilities.

## Atomic scope fan-out

`RefreshStore.InvalidateScopes` validates one to three distinct resolved scopes for one workspace, with the existing 10,000-path limit. It applies the event to all affected scopes in one SQLite transaction. A trigger/write failure in one scope rolls back the entire event, including earlier scope changes. `Invalidate` remains a one-scope wrapper.

Root matching includes current and last committed roots. Notes changes affect notes/all, skills changes affect skills/all, and configured extra roots affect all only. Unrelated paths leave revision/state untouched; matching is conservative for unsupported extensions. Each event advances requested revisions without altering indexed content, count, generation or last-success time. No refresh lease is acquired.

## Before and after notification

The shared writer first records invalidation, then attempts the filesystem write, then records a second invalidation. Both notifications have five-second contexts. Failure of the first prevents mutation. The second runs after any attempted write, including partial failure, using a bounded context independent of caller cancellation.

A refresh that captures the first revision while the file write is in progress cannot acknowledge the second revision. Both scopes retain pending work until a later explicit refresh. The second notification may conservatively retain staleness even if the file operation failed before changing bytes.

Post-notification failure returns an error that says bytes may have changed and explicit reindex is required. The filesystem write cannot be rolled back by SQLite. A caller must not interpret this error as proof that the original bytes remain or blindly retry a non-idempotent operation. Existing full-content writes remain idempotent only with respect to the supplied bytes.

### Crash boundary

Filesystem writes and SQLite notifications are separate transactions. A process can die after a refresh acknowledges the first revision but before the second notification. This protocol does **not** guarantee crash-atomic freshness. Successful notification persistence/reopen is tested; the uncovered crash window needs a durable mutation-intent/recovery protocol or equivalent reconciliation before automatic freshness is enabled. External edits, shell commands, uploads outside these paths, hard-link aliases and watcher events are also outside this producer's delivery guarantee.

## Evidence

- Native tool and script writes mark only affected scopes stale with two requested revisions, while indexed bytes/count/generation/time remain unchanged until explicit refresh.
- The registered engine executor records notes/all revisions without advancing committed generation. HTTP functional coverage asserts stale status and retained old query hits after a write, then new hits after Reindex.
- Extra-root, skills, unrelated-path and VFS cases retain scope isolation.
- Native trigger failures prove whole-event rollback before mutation and post-write failure with changed bytes plus the durable pre-revision retained.
- A refresh acquired between notifications publishes with a pending later revision; a second connection observes the same durable counters.
- Cancelled callers cannot suppress post-notification; partial-write errors are retained. Invalid configuration, traversal, symlinks, directories and FIFO targets are rejected.

`make check BIN_DIR=/tmp/gi-scheduler-bin` passed Go tests, vet, web build, hook checks and **77 functional tests**. **32 helpers** parse **22 derived proposals** separately from frozen features. Tools/search-store, web and turn race suites passed ×3 in separate runs; Linux tests and Darwin arm64 tools/turn test-binary cross-compilation passed. **24 focused browser executions** passed (18 workspace and six configured-root cases).

The first combined race command timed out after tools/search-store passed. Repeated engine runs exposed a shared-memory database fixture in the new test; it now uses its own on-disk temporary database. An existing cancellation fixture checked claim absence after `FinishedAt` but before `cleanupTurnRun`; it now waits for that cleanup boundary and retains the exact no-active-claim assertion. The full turn race suite then passed ×3. The delegated design review timed out and supplied no evidence. Logs are `/workspace/tmp/gi-index-producer-*`.

No supplied component, frozen feature, schema migration or terminal row changed. No new terminal acceptance or frozen credit: **45/236 Classic**, **2/42 shared**, **191/40 unmapped**.
