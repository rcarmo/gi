# ADR-0050: Application-owned explicit index refresh

## Status

Accepted — 2026-09-23. Web startup now owns the bounded scheduler from ADR-0049. Authenticated explicit POST reindex requests share its batches. Startup, GET status and GET search remain scan-free. Filesystem/mutation producers, automatic refresh and terminal controls are not enabled.

## Server lifecycle

`Server.New` still starts no index goroutines. An embedding calls `StartWorkspaceIndex(applicationContext)` and defers `CloseWorkspaceIndex()` before closing SQLite. The main binary and deterministic UX server do this explicitly. Start resolves `all`, `notes` and `skills` once; their configurations and workspace identities are shared by GET and POST paths for that server lifetime. Retargeting a workspace symlink cannot silently change query identity while the scheduler still scans the original workspace.

Repeated start is idempotent; nil or already-cancelled contexts are rejected. Invalid index settings leave chat/server startup available, log the index startup error and retain the existing HTTP 400 configuration validation. Unstarted/closed indexing rejects valid POST requests with 503. Repeated/concurrent close calls reject new work and join the same coordinator. Close cannot restart a scheduler.

No refresh runs during construction or start, and no index rows are created by ordinary status/search reads. Read handlers never enqueue work. Changing roots/settings still requires restart.

## Explicit HTTP refresh

POST first validates the native scope and request context, then requests a shared ticket and waits with the HTTP context. A caller disconnect cancels only its wait. The application-owned batch continues within its existing bounds, so another caller can receive the result without a duplicate scan. A request already cancelled before admission creates no work.

Successful responses project the scheduler's verified ready observation. A second refresh starting immediately afterwards cannot replace that response with an unrelated indexing/failed status. GET status independently returns the current durable state. Ready observations are not a guarantee against changes arriving after the observation; durable revisions and future producers govern freshness.

Scan/write failures still return 500, and auth/method/config validation is unchanged. Exhausted contention, ownership loss or newer pending revisions return 409. Shutdown rejects new work with 503; an already-waiting request can receive the existing 500 cancellation error. There is no implicit retry of permanent scan failure. The native worker, revision counters and lease fences remain the storage authority.

## Joined HTTP shutdown

`internal/httpserver.Run` serves and joins the main and optional ACME HTTP listener as one lifetime. Signal cancellation or any listener exit cancels application work, stops admission, calls Shutdown on every listener and waits for every serve loop. A handler wrapper tracks admitted handlers under a shared admission mutex. This closes the gap where `ListenAndServe` returned `ErrServerClosed` before its shutdown goroutine finished draining requests.

The five-second graceful deadline triggers `Server.Close` to cancel remaining network contexts. The runner still joins admitted handlers; a handler ignoring cancellation can delay shutdown beyond five seconds. The timeout cannot permit premature database teardown. Web routes do not use hijacked connections; this runner does not own external hijacked connections.

The main entry point now returns listener errors through `run()` so deferred scheduler, engine, store and PID cleanup runs before the process exits nonzero. ACME/TLS configuration errors after startup also return through that cleanup. The deterministic UX server uses the same runner. TUI startup is unchanged.

## Evidence

- Native startup/GET checks create zero workspace index identities; repeated start/close, unavailable startup, invalid settings and parent cancellation behave as specified.
- Concurrent HTTP requests wait behind a native peer lease. Disconnecting one waiter leaves the shared ticket and another waiter intact; release permits one generation containing the native bytes.
- Concurrent application close terminates active/queued tickets and does not release or fail a different process's lease.
- Startup-resolved workspace identity survives a symlink retarget; an already-cancelled POST creates no work.
- Real HTTP tests hold a handler during shutdown, prove cleanup waits for it, join primary/secondary listeners, and force timeout cancellation without abandoning a held handler. A secondary listener bind failure cancels and joins its peer.
- A built-binary rehearsal verified scan-free startup, explicit native reindex, nonzero bind failure with deferred PID removal, SIGTERM during a lease-blocked POST, clean exit, database integrity, unchanged committed generation and retained peer lease. Its comparison command initially used a double-quoted SQL literal rejected by SQLite; the corrected rehearsal passed.
- A functional test edits indexed bytes and proves GET keeps the committed snapshot until the next explicit refresh. Watcher/mutation invalidation remains disconnected.

`make check BIN_DIR=/tmp/gi-scheduler-bin` passed Go tests, vet, web build, hook checks and **77/77 functional tests**. **32/32 helpers** parse **21 derived proposals** separately from frozen features. HTTP/index lifecycle races passed ×10; full HTTP/web/indexer race suites passed ×3. **24 focused browser executions** passed: 18 workspace and six configured-root cases across Chromium/WebKit at phone/tablet/desktop sizes. Logs are `/workspace/tmp/gi-index-lifecycle-*`.

`UX_LOCAL_BIN` and `UX_LOCAL_PORT` Makefile overrides allow disposable UX binaries on `/tmp` and an unused local port without replacing another harness process. Defaults remain unchanged. No supplied component, browser layout, TUI row or frozen scenario changed. No three-size PTY matrix was rerun for this web lifecycle slice. Frozen coverage remains **45/236 Classic**, **2/42 shared**, **191/40 unmapped**.

## Next integration

Record native mutation/watch events durably before requesting scheduler work, with bounded event delivery, overflow handling and shutdown ordering. Plan restart recovery and stale background requests separately; GET remains read-only until an explicitly reviewed contract changes it. Terminal index actions still require temporary bounded surfaces and independent draft/cursor/reader/idle-row evidence.
