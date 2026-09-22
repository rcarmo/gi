# ADR-0043: Scoped index refresh transactions and ownership

## Status

Accepted — 2026-09-22. Internal configuration and storage APIs are implemented on the v1 schema. No scanner, worker, native query/status/reindex route or terminal control is connected. Frozen coverage remains 45/236 Classic and 2/42 shared, with 191/40 unmapped. The derived index scenarios still require end-to-end acceptance.

## Scope configuration

`internal/search/store/config.go` resolves an existing filesystem workspace to an absolute symlink-resolved identity. Scope roots must be clean relative paths; overlapping roots within one scope collapse to their parent. Arrays are private and accessors return copies. A deterministic fingerprint covers sorted roots/extensions, chunker version and per-document byte policy.

Default scopes follow Piclaw: `notes`, `.pi/skills`, and their union for `all`. Extra roots extend only `all`. Extensions include the pinned Piclaw defaults plus Go/Python/Rust text sources; callers can supply extra extensions or construct a custom scope. There is no environment/settings wiring in this slice. Root existence, symlink confinement during scanning and exclusion policy belong to the forthcoming scanner; configuration validation alone does not read or authorise files.

## Fenced ownership

`RefreshStore.Begin` acquires the workspace-wide lease in a database transaction with a random 256-bit owner token. Scope overlap cannot bypass ownership. Lease duration is 1 second–5 minutes; the future worker must renew it while scanning. SQLite's shared wall clock supplies expiry comparisons. Wall-clock jumps can cause early expiry or delayed takeover; monotonic cross-process time is not supplied by SQLite. Token fencing still prevents a replaced owner from changing successor data.

Renewal requires a matching unexpired lease and cannot resurrect expiry. Commit and failure reporting verify owner and expiry both before writes and immediately before releasing the lease and committing. File-backed core stores already use immediate transactions, so a competing writer cannot take over midway. An expired worker may continue computing, but its old handle cannot renew, fail, or publish.

`Status` projects an orphaned `indexing` row as stale when no valid lease remains. The next successful acquisition persists stale recovery for interrupted scopes, then marks its selected scope indexing. Last committed configuration/count/timestamp/generation remain intact until publication. Reopen does not trigger scanning or eagerly mutate status. These methods assume one v1-migrated Gi database and its standard foreign-key/transaction configuration.

## Complete snapshot commit

A trusted caller supplies `CompleteSnapshot`. Incomplete snapshots are rejected before writes. The completeness flag is a caller contract; this API cannot prove that a filesystem traversal succeeded. The scanner must never set it after skipped IO errors or inventory limits.

Commit validates root/extension membership, unique clean paths, UTF-8 content, sequential chunk indices, source byte/line ranges, exact source slices and code-point boundaries. Limits are 10,000 documents, 1 MiB per document, 32 MiB total source text, 20,000 chunks and 64 MiB aggregate chunk text/headings. Metadata lengths are bounded. Invalid snapshots leave the old committed generation untouched and retain ownership for explicit retry or failure reporting.

The transaction hashes source bytes rather than relying on mtime/size. Same-content/chunker/language documents retain document and chunk IDs; an unchanged document's mtime can advance. Changed documents retain their document ID, replace chunks and update FTS through the migrated triggers. A chunker version is the caller's deterministic chunking contract; changing chunk boundaries requires a new version.

Changed shared documents mark other owning scopes stale; unrelated scopes remain unchanged. Publication replaces only the selected scope's memberships. Documents/chunks/FTS are removed only when no scope owns them. The scope's new configuration, roots, count, timestamp and incremented generation commit atomically with content and membership cleanup. Two calls using one handle cannot publish twice because the first releases its unique token.

`Fail` requires live ownership and an error, stores bounded UTF-8 failure detail, retains last committed data/config/count/timestamp/generation and releases the lease. Cancellation callers need a fresh bounded cleanup context to report failure; an abandoned or expired handle instead becomes stale through recovery. There is no automatic retry.

## Verification

Tests use real v1-migrated SQLite stores and distinct connections. They cover deterministic configuration/root deduplication, skills eligibility, unchanged identities, same-size/mtime content replacement, overlapping-scope cleanup, unrelated scope isolation, root/chunker changes, malformed/oversized snapshots, Unicode offsets/lines, cancellation, failure and reopen.

A real SQLite insert trigger injects a mid-write error; rollback retains prior FTS results. Another trigger expires the lease after chunk insertion; the final fence rolls back the transaction. Two-store acquisition yields one owner; expiry/takeover rejects all old-owner actions. A two-goroutine commit test yields one generation. FTS external-content integrity passes after updates and cleanup.

Full Go tests/vet/build/hook checks and **74/74 functional tests** pass. **32/32 helpers** pass; store and search-store race suites pass ×3, with search-store rerun after the final Unicode validation guard. Logs are `/workspace/tmp/gi-index-refresh-*`. Delegated read-only review timed out and supplied no evidence.

No runtime caller, migration, web component or terminal layout changed, so the full browser parity and PTY matrices were not rerun or credited. No screenshot is needed for these internal APIs.

## Next integration gate

Implement the bounded rooted scanner and deterministic chunker, with explicit optional-root policy, complete-inventory proof, exclusion rules and file-change detection. Add refresh scheduling, lease renewal/cleanup, invalidation and lexical query execution before exposing controls. Connect the existing web status/reindex surfaces through host/API adapters and separately verify temporary terminal status/actions at all three sizes without extra idle rows. The provisional unscoped rebuild remains stashed.
