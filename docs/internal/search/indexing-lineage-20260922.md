# Indexing lineage and implementation gate — 2026-09-22

Piclaw supplies the closest working workspace-index lifecycle. Tau and Vibes supply useful database consistency patterns for conversation search. Gi must keep these domains distinct when adapting their schemas and acceptance criteria.

The provisional Gi full-workspace lexical rebuild is **stashed and not deployed**. It was tested locally, but its single global status row, whole-index replacement and exclusion of all dot directories do not implement Piclaw's roots/scopes or skills indexing. Runtime scoped explicit reindex is now implemented by [ADR-0046](../../adr/0046-native-workspace-index-api.md); automatic freshness remains a gap; explicit startup/optional-root policy is now implemented by [ADR-0047](../../adr/0047-index-settings-and-optional-roots.md). The comparison below records the design at `80ab58c`. [ADR-0042](../../adr/0042-versioned-workspace-index-schema.md) subsequently promotes the scoped candidate into a versioned, additive startup migration; no worker/query/status implementation is enabled and the derived scenarios remain proposals.

## Pinned sources

| Project | Revision inspected | Relevant source and tests |
|---|---|---|
| Piclaw | `bfc34e4ebfe9b0ce0deefa6202a9d9a4780a5322` | `runtime/src/workspace-search.ts`, `workspace-index-process.ts`, `db/connection.ts`, `channels/web/handlers/workspace.ts`, `channels/web/workspace/service.ts`; `runtime/test/workspace-search.test.ts`, `workspace-index-process.test.ts`, `extensions/extensions-workspace-search.test.ts` |
| Tau | `62d28201ec1d16f3a8097f159a88f2bfdacd932f` | `src/tau_web/sqlite/migrations.py`, `repositories.py` (`SearchRepository`); `tests/web/test_sqlite_runtime_repositories.py::test_search_repository_updates_filters_and_removes_stale_rows`, `test_sqlite_migrations.py` |
| Vibes | `16894139ba502c5555834f9925bac92ca4ddd34e` | `internal/db/migrations.go`, `queries.go` (`SearchInteractions`), `db_fuzz_test.go`; `tests/features/workspace.feature` |
| Gi baseline | `6aff14c` | `docs/adr/0008-workspace-hybrid-search.md`, `internal/search/{types,service}.go`, `internal/search/store/schema.go`, `docs/internal/search/fts-namespace.md` |

These revisions were read locally. Their full upstream test suites were not rerun during this comparison. Tau's checkout is divergent from its remote; the revision above, not an assumed latest release, is the evidence boundary. Frozen Classic provenance remains `70d33bc93ab540845bbcf5f80503ca8125c71594`; no frozen feature is edited by this work.

## Piclaw: file inventory plus lexical FTS

Piclaw uses three tables:

- `workspace_files(path PRIMARY KEY, root, mtime, size)` supplies incremental file inventory.
- `workspace_fts` indexes content with FTS5; path, root, mtime and size are unindexed metadata.
- `workspace_index_status(scope PRIMARY KEY, roots, last_indexed_at, last_error, indexed_file_count, updated_at, state)` persists the lifecycle.

Default scopes include `notes`, `.pi/skills`, and configured additional roots; `notes`, `skills` and `all` queries retain different path filters. Extension filtering and a clamped size cap limit eligible files. The walker skips `.git`, `node_modules`, `.cache` and `generated`; it does not blanket-exclude dot directories, which would break `.pi/skills`.

A refresh skips files with unchanged rounded mtime/size, replaces changed text and inventory rows, and removes unseen paths only in scanned roots. Search runs FTS MATCH with bm25/snippets and bounded limit/offset. Malformed/unavailable FTS takes a content-LIKE fallback; non-ASCII and punctuation handling need dedicated Gi tests rather than assuming identical tokenizer results.

Ordinary search uses the current index and requests a background refresh when it is missing/stale. Explicit refresh blocks. File mutation/watch paths mark affected scopes stale. A child-process launcher avoids competing local children and checks recent `indexing` status before launching. Status distinguishes `never_indexed`, `indexing`, `ready`, `stale` and `failed`; failure retains the last successful timestamp/count.

Important limits in the inspected implementation:

- File-table and FTS changes are separate statements, not an atomic whole-scan transaction.
- Traversal/read errors can be skipped; an incomplete inventory must not become destructive cleanup in Gi.
- mtime/size alone cannot detect same-size content edits with unchanged timestamps.
- Background coordination is not a database-fenced cross-process lease.
- Overlapping root/scope ownership is represented by paths/root metadata, not explicit membership rows.

The stronger behaviours proposed for Gi below are intentional differences, not upstream pass claims.

## Tau: explicit search-document ownership

Tau's `search_fts` holds text plus unindexed `entity_type`, `entity_id` and `session_id`, tokenized with `unicode61 remove_diacritics 2`. `SearchRepository.upsert` deletes/reinserts one entity in a writer transaction. Search uses MATCH, optional session filtering, bm25 rank and a rowid tie-break. Remove and missing-session purge have explicit methods. Repository tests cover replacement, scope filtering, remove and orphan cleanup.

This is conversation/extension-content search. No native filesystem workspace refresh implementation or workspace-index status route was found in the inspected Tau sources. Its transferable requirement is transactional identity and cleanup, not an inferred file-index lifecycle.

## Vibes: external-content FTS with triggers

Vibes' `interactions_fts` points to canonical `interactions` rows by ID, using `porter unicode61`. Insert/delete/update triggers derive indexed content from `data.$.content`. Search joins back to interactions and applies rank, limit and offset. The schema maintains FTS alongside ordinary database writes rather than requiring each caller to remember a separate update.

Its workspace Gherkin covers file CRUD/traversal. No native workspace reindex lifecycle was found in the inspected Go sources. The portable pattern is stable FTS identity plus transactional trigger maintenance. Porter stemming and Tau/Piclaw tokenizers differ; Gi should specify its own Unicode search contract explicitly.

## Gi schema candidate

[`workspace-index-candidate.sql`](workspace-index-candidate.sql) executes on Gi's existing pure-Go SQLite driver in `TestCandidateWorkspaceSchemaFTSAndMembership`. The historical SQL file is not loaded directly by application initialisation; its versioned copy now is, through ADR-0042. The semantics test runs the production migration. It uses new table names to avoid silently reinterpreting the old scaffold:

- workspace identity separates databases reopened against different roots;
- scope status includes configuration hash, roots, committed generation and last-success fields;
- documents are unique within a workspace, with hash/mtime/size and a separate chunker version;
- memberships model overlapping scopes without duplicate documents or cross-scope deletion;
- chunks retain stable IDs and byte/line source locations;
- an external-content FTS view joins chunks to document path/language; triggers maintain inserts, changes, rename and cascade deletion;
- a workspace-level lease stores owner token and expiry for a future fenced worker.

The candidate test executes insertion, content update, rename, overlapping-scope cleanup, foreign-workspace rejection, rollback, cascade deletion, FTS integrity/rebuild, reopen and lease-token SQL. This proves SQLite mechanics only. It does not prove a scanner, watcher, background worker, lease-expiry clock policy, safe query parser, ranking, migration or multi-process runtime.

Vector storage is deliberately absent from this lexical schema experiment. ADR-0008 remains a draft hybrid target. Chunker version must not be written into an `embedding_version` column; embedding model, dimension, normalisation and version need a separate keyed representation before vector integration. Gi's existing `fts://workspace` documentation-bundle registry is also distinct from a live filesystem index.

## Derived Gherkin and test ownership

[`features/search/workspace-index.feature`](../../../features/search/workspace-index.feature) now has 23 `@index-derived-*` scenarios, all marked `@proposal`; ADR-0047–0051 add root policy, durable invalidation, scheduling, application ownership and native writes; ADR-0052 adds independently verified explicit terminal actions. They cover configuration, incremental identity, scoped cleanup, nonblocking/background and explicit refresh, status, atomic failure, bounded scanning, fenced ownership, lexical query fallback, FTS integrity, configuration/version changes, draft isolation and a compact terminal surface.

Tags identify Piclaw-derived expectations, Tau/Vibes schema patterns, and Gi strengthening. The parser test verifies all scenarios parse and that the frozen 236 Classic/42 shared catalogue is unchanged. **It is not a Gherkin step runner and confers no behaviour pass.** The schema test supports the SQL portions of derived-011/012 and lease-token storage, not their complete acceptance.

Workspace-005 remains unmapped: its visible creation/upload controls and menu contract still need complete evidence. Runtime indexing tests from the shelved prototype cannot earn those mappings.

## Implementation and migration sequence

1. **Startup settings implemented by ADR-0047:** extra roots/extensions and exact optional roots feed deterministic fingerprints. Default strictness/skills eligibility and overlapping-root deduplication remain. Absent optional roots with committed memberships fail atomically rather than erasing old content.
2. **Schema installed by ADR-0042:** explicit versioned SQLite migration with collision/version/object checks, preserving chat/session/media and old scaffold rows. New scoped tables start empty; the forthcoming explicit refresh must rebuild derived content rather than guess old ownership. Rollback/reopen/history preservation is tested.
3. **Storage API implemented by ADR-0043:** stable unchanged chunk/document IDs, transactional FTS maintenance, scope membership cleanup and rollback on incomplete/invalid snapshots. **Scanner/chunker added by ADR-0044:** bounded rooted traversal, required roots, deterministic UTF-8 lines, rehash/change detection and scan→commit tests. This is observed consistency, not an atomic filesystem snapshot; optional-root policy and worker integration remain open.
4. **Storage protocol implemented by ADR-0043:** per-workspace token/expiry fences, renewal, failure reporting and stale projection/takeover recovery, tested with two stores. **Explicit worker added by ADR-0045:** periodic renewal during scanning, joined heartbeat before commit, bounded separate failure cleanup and killed-process recovery acceptance. Application scheduling, invalidation and public controls are not yet wired.
5. Adapt Piclaw's incremental metadata fast path, with explicit full/hash verification for metadata-collision cases. **ADR-0048 supplies durable invalidation capture/ack revisions**, including events during commit; **ADR-0049 supplies an internal bounded coalescing scheduler** with revision-aware completion, contention retries and cancel/join shutdown. **ADR-0050 connects application ownership and explicit POST batches**, with cancellation-isolated waits and joined HTTP/index shutdown. **ADR-0051 connects engine/HTTP/script filesystem write invalidation**, with atomic fan-out and bounded pre/post notifications. Shell/external/watch delivery, crash-window/restart reconciliation and automatic background requests still need wiring.
6. **Native/web surface implemented by ADR-0046:** authenticated scoped status/query reads and explicit POST reindex with host Reindex/Refresh actions. Reads do not refresh; tool/TUI query consumers and background scheduling remain open. Semantic capability remains absent pending independent vector/embedding integration.
7. Run step-owned Go/native/browser acceptance, migration/race tests and separate three-size TUI tests; only then integrate UI mappings that satisfy complete frozen criteria.

[ADR-0052](../../adr/0052-compact-terminal-index-actions.md) implements explicit Alt-I status/reindex with five temporary rows, preserving draft/cursor/reader and existing Pi padding. Native three-size fullscreen/regular tests pass. No persistent indexing badge, tree sidebar or additional idle row is added; automatic freshness and terminal search-result navigation remain separate.

## Local verification

Full Go tests and vet pass. The candidate SQL test passes under the race detector three times, including whole-workspace cascade cleanup and FTS integrity. **32/32 helper tests** pass; the derived Gherkin parser test checks 15 unique proposals and unchanged frozen totals (236 scenario definitions, 256 expanded Classic cases, 42 shared cases). No application/browser/TUI code or database migration is deployed by this design change, so no browser or PTY rerun is credited. Logs are `/workspace/tmp/gi-index-lineage-{go,helpers,race}.log`.

## Preserved work

The uncommitted full-rebuild prototype is saved in the Git stash named `Provisional lexical full rebuild before Piclaw Tau Vibes index contract review`. Its API/UI wiring and failure tests may be reused after the lifecycle design above is implemented; the global destructive rebuild and blanket dot-directory exclusion should not be restored as the default strategy.
