# ADR-0042: Versioned scoped workspace index schema

## Status

Accepted — 2026-09-22. This is a storage prerequisite for the [Piclaw/Tau/Vibes-derived indexing design](../internal/search/indexing-lineage-20260922.md). No new frozen web mapping or terminal index capability is implemented. Coverage remains 45/236 Classic and 2/42 shared, with 191/40 unmapped.

## Startup migration

`internal/search/store/migrations/001_scoped_workspace_index.sql` promotes the tested candidate DDL into an embedded v1 migration. `internal/search/store/migration.go` runs inside `internal/store.initSchema`'s existing transaction, after the core additive upgrades and indexes. The caller commits only after both core and search schema work succeed.

A dedicated `workspace_index_migrations` ledger stores version, SHA-256 of the embedded SQL and application time. Reopen verifies the version/checksum, required columns, object types and normalised object definitions, including all FTS triggers. Missing, changed or future schemas fail startup rather than being silently recreated. This is schema validation, not a full FTS-content integrity scan on every startup.

Unversioned candidate/partial tables with the new names are rejected. Operators must inspect/back up those databases before repair; the migrator does not guess whether their data is safe to adopt. File-backed opens already use immediate transactions and a busy timeout; a two-connection test verifies concurrent migration attempts produce a single ledger record.

## Preserve old data and ownership

The migration is additive. Sessions, messages, turns/events, media, VFS and KV data are untouched. Old scaffold tables (`workspace_documents`, `workspace_chunks`, `workspace_chunks_fts`, `workspace_index_meta`) remain unchanged, including any existing rows. Their unscoped path identities cannot reliably establish a filesystem workspace or overlapping-scope membership.

The new scoped tables start empty. No root is invented, no old `ready` flag is relabelled, and no scanner is started. A later explicit scoped refresh can populate the new index. Retiring the old derived tables requires a separate migration and caller audit; this migration does not remove them.

The schema separates workspace identity, scope/configuration/status, documents, overlapping memberships, source-located chunks and a workspace writer-lease record. External-content FTS uses chunk IDs and triggers for insert/update/delete/rename and cascades. There is no vector table or embedding capability; chunker versions are separate from future embedding metadata.

## Tests and rehearsal

- Fresh creation, concurrent connections and reopen retain a single stable ledger row and existing indexed data.
- A collision at the final trigger rolls back earlier FTS/tables/triggers and the ledger. The core integration test also proves legacy ALTER/backfill work rolls back with that late failure.
- Future version, checksum change, missing table/trigger, no-op replacement trigger and unversioned candidate tables are rejected.
- Synthetic current/legacy databases preserve all columns and values in runtime and old search tables across migration and reopen. Foreign-key checks remain clean.
- The prior candidate test now executes the production migration before exercising chunk/FTS identity, scope membership, rename, rollback, cascade, rebuild, integrity and lease-token SQL. It still does not prove a runtime scanner or lease protocol.
- A SQLite backup of the local dev database was migrated separately. All **19 pre-existing tables** matched their pre-migration content hashes, and `integrity_check` returned `ok`. Backup and comparison files remain private under `/workspace/tmp/gi-index-{migration-backup.db,before.db,before.sql}`; no database contents were posted.

Full Go tests/vet/build/hook checks pass, with **74/74 functional tests**, **32/32 helpers**, and store/search-store race tests ×3. The focused six-project workspace browser run passes **12/12**. Session, compaction, reading, outcomes, regular, search and selection TUI suites pass at **60×18, 100×22 and 140×36**; smoke and Gherkin also pass. No UI layout or terminal code changed. The full 486-execution browser matrix from the prior UI slice was not rerun or recredited here.

Logs are `/workspace/tmp/gi-index-migration-*`. Frozen Gherkin remains unchanged; derived index scenarios remain proposed until their entire runtime behaviour is verified.

## Remaining work

Implement configured root/scope resolution and fingerprints; bounded incremental scan/chunking; atomic scoped replacement/cleanup; fenced worker acquisition, renewal, commit and crash recovery; invalidation/background refresh; native lexical query/status/reindex; then web and explicit compact terminal actions. The stored lease row alone does not enforce a running worker protocol. The provisional global full-rebuild implementation remains stashed and must not be restored as the final indexing strategy.
