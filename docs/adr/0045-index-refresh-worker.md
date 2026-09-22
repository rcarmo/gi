# ADR-0045: Explicit lease-renewing index worker

## Status

Accepted — 2026-09-22. An internal worker now connects the rooted scanner to fenced refresh storage. No automatic scheduling, application-start indexing, native query/status/reindex routes or web/TUI controls are enabled. Frozen coverage remains 45/236 Classic and 2/42 shared, with 191/40 unmapped.

## Execution contract

`internal/search/indexer.Worker.Run` accepts the caller's context and a resolved scope. It acquires the workspace-wide lease before invoking `ScanScope`, renews periodically while scanning, joins renewal before publication, then calls the existing complete-snapshot commit. It performs one attempt only. Busy ownership returns without scanning or writing another worker's status.

Defaults are a 30-second lease, renewal every 10 seconds, a five-minute total run context, five-second acquisition/renew/commit contexts and a two-second cleanup context. Timings are private; public configuration is not part of this slice. The context bounds work cooperatively. Filesystem calls and SQLite waits may not respond immediately to cancellation; `Run` waits for in-flight work rather than abandoning a goroutine that could act later. Callers must wait for `Run` before closing the store.

Renewal failure cancels the scan with its cause. After the scanner returns, the worker stops and joins the renewal goroutine before any commit or failure report. It renews once more immediately before the bounded commit. This avoids a concurrent heartbeat misinterpreting the successful lease release as ownership loss, or fighting the commit's SQLite writer lock. The store's entry and final token/expiry fences remain authoritative.

A successful commit returns success without a later cancellation check. Cancellation before publication prevents commit; cancellation after a successful durable commit must not encourage a duplicate retry by reporting an uncommitted failure.

## Failure and shutdown

Scan, validation, renewal or commit errors invoke `Refresh.Fail` using a separate two-second cleanup context, since the caller may already be cancelled. Cleanup retains the prior committed index/count/timestamp/generation. Its ownership fence prevents an old worker from changing a successor's state. Cleanup errors are joined with the original error rather than discarded.

If cleanup cannot persist or ownership has expired, the lease is allowed to expire and the stored `indexing` state projects as stale. The next acquisition persists interrupted-scope recovery. This is bounded cleanup and lease-based recovery, not a detached retry. No process-wide service manager or background queue has been added.

The scanner requires every configured root and uses literal versioned line chunks. Optional-root policy, mutation invalidation and settings remain unimplemented. The worker neither writes chat drafts/messages nor makes model/provider calls.

## Evidence

Tests run the production scanner and refresh store against migrated SQLite databases. Controlled blocking scanner seams make timing deterministic without replacing HTTP/SSE lifecycle evidence.

- A one-second test lease remains valid beyond its original expiry while a scan is held; the token stays unchanged and a second store cannot acquire it.
- Successful native scans commit exactly one generation and leave no lease. Renewal is joined before return.
- Caller cancellation, run deadline, missing-root scan failure, incomplete snapshots and injected SQLite chunk-insert failure retain the old index and persist failed status.
- Cancellation immediately after scanning never publishes the complete candidate. An explicit later retry succeeds once.
- A real renewal UPDATE rejected by SQLite cancels scanning, records its error and releases ownership without losing content.
- Forced expiry/takeover cancels the old scanner; its cleanup cannot alter the successor's indexing status or generation. The successor commits normally.
- Expiry between scanning and final renewal prevents publication. A rejected cleanup write returns both the original and cleanup errors and leaves the prior snapshot readable.
- A child test process acquires a real lease and blocks in scanning. The parent kills it without Go cleanup, waits for natural expiry, verifies stale status and retained content, and successfully refreshes with a new worker.

Full Go tests/vet/build/hook checks, **74/74 functional tests** and **32/32 helpers** pass. Worker/scanner/store race suites pass ×3, including the killed-child recovery and native renewal-error cases. Logs are `/workspace/tmp/gi-index-worker-*`. A delegated read-only review timed out and supplied no evidence.

No runtime application caller or UI changed, so full browser-parity and PTY matrices were not rerun or credited. The derived worker scenarios have substantial native evidence, but remain proposed until their complete public lifecycle is verified. No screenshot is generated for this internal change.

## Next integration

Add explicit scope/settings and optional-root policy, lexical query execution, status projection and application-owned scheduling/invalidation. Connect native web reindex/status through the host adapter only after endpoint/auth/failure tests. A terminal action should use existing transient status or a temporary bounded detail surface, restore draft/cursor/reader on Escape, and add no idle rows. Its three-size acceptance remains independent of these native worker tests.
