# ADR-0049: Bounded workspace index scheduler

## Status

Accepted — 2026-09-23. `internal/search/indexer/scheduler.go` provides an internal coordinator around the explicit worker. Production web/TUI startup, HTTP handlers, queries and watchers do not call it yet. Explicit HTTP reindex still uses the request-owned worker from ADR-0046.

## Ownership and bounds

`NewScheduler(ctx, store, configs)` accepts one to three distinct, resolved startup scopes for one workspace. Scope configurations are fixed until the owner creates a replacement scheduler. One goroutine serialises local work across scopes because the durable lease is workspace-wide. Cross-process ownership continues to use ADR-0043 fences.

`Request(scope)` performs no database or filesystem access. It returns a shared `RefreshTicket`; queued or active requests for that scope join the same batch. The queue contains at most one batch per configured scope. Repeated requests do not extend its attempt budget or deadline. A new explicit request after completion starts another batch even when the index is ready.

`ticket.Wait(ctx)` waits for a batch outcome. Cancelling one wait does not cancel the worker or other waiters. The result includes a detached copy of the last observed scope status and an error. Callers must check the error; status accompanying an error may precede the failure. Failure detail and revisions remain in durable storage.

Defaults are four worker attempts, a five-minute context deadline per active batch, five-second status operations, and contention/follow-up delays of 250 ms, 500 ms and one second, capped at two seconds. Queued time is outside the active batch deadline; at most two other scopes can precede a batch. Worker limits and separate failure cleanup still apply. Uninterruptible filesystem/SQLite calls can delay cancellation beyond the context deadline.

## Completion and retries

Every batch attempts an explicit refresh, including a ready scope. After an attempt, completion requires a matching configuration, ready state, equal requested/acknowledged revisions and a generation newer than the batch's initial observation. This prevents an old ready snapshot from satisfying a refresh while a peer owns an unrelated scope's lease.

A newer compatible peer publication can satisfy a batch after contention without a duplicate local scan. Otherwise only lease-busy/lost errors and successful-but-stale publication retry automatically, within the same attempt/time budget. Scan, write and status errors stop the batch. Exhaustion returns `ErrRefreshPending` and preserves pending revisions. No durable error is cleared by the scheduler itself. A later explicit request can retry a failed batch.

Request counters protect the final status-read/completion race. A caller that durably invalidates a path and then requests work either causes a readiness re-read or creates a new batch after the prior one closes. Status reads stay outside the mutex so SQLite cannot block request admission. Rechecks have the same finite count bound as worker attempts; a continuous request storm returns a pending error rather than spinning indefinitely.

Without a local request, an invalidation arriving after the final status read remains durable for a future caller; this coordinator does not poll idle scopes. A watcher/mutation producer must record invalidation before requesting work. Restart recovery also needs an application-owned durable-status inspection; the in-memory queue does not survive process exit.

## Shutdown

The application context cancels active work and retry timers. `Close()` is idempotent, cancels and joins the coordinator, waits for worker renewal/failure cleanup, and terminates queued tickets with `ErrSchedulerClosed`. Requests after cancellation are rejected. The owner must call `Close()` before closing SQLite. Close has no timeout that could abandon work and falsely permit database teardown.

The scheduler neither starts on its own nor scans at construction. Application lifecycle wiring, mutation/watch delivery and read-only query policy need separate integration tests before activation. Search GET remains strictly read-only.

## Evidence

Native tests cover:

- 64 concurrent same-scope requests sharing one ticket/scan, isolated waiter cancellation, detached result slices and a later explicit ready refresh;
- a held scan, real file edit and durable invalidation followed by a second scan that publishes new bytes and acknowledges the newer revision;
- serial notes/skills batches, bounded configured scopes and workspace isolation;
- two-store contention resolved by a newer peer commit, and unrelated peer ownership that cannot satisfy a ready refresh with its old generation;
- native lease loss and successor publication without stale-owner damage or duplicate scan;
- capped retry delays, attempt exhaustion under busy/lost/contention or continuous invalidation, and no automatic retry of permanent scan failure;
- the final-status-read/request race, bounded status-read storms and status-read failure;
- shutdown held inside a cancelled scanner, joined failure cleanup/released lease, queued ticket termination, deadline cancellation, parent cancellation and idempotent close after database teardown.

Search-store/indexer race suites passed ×3; scheduler-specific race tests passed ×10 after adding native takeover, deadline and read-storm cases. `make check BIN_DIR=/tmp/gi-scheduler-bin` passed Go tests, vet, web build, hook checks and **76/76 functional tests**. **32/32 helpers** passed, including parsing **20 derived proposals** separately from the frozen catalogue.

The first standard check failed while writing the binary because the workspace volume ran out of space. Re-running with disposable binaries in `/tmp` passed. Backups/screenshots were retained. Logs are `/workspace/tmp/gi-index-scheduler-{check,helpers,race,stress}.log`.

A delegated design review returned lifecycle recommendations. A later delegated code review timed out and supplied no review evidence. No focused browser or three-size terminal matrix was rerun for this unused internal coordinator. Frozen credit remains **45/236 Classic**, **2/42 shared**, **191/40 unmapped**; no UI, terminal rows or frozen feature criteria changed.
