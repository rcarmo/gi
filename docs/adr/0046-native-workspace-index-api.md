# ADR-0046: Explicit scoped indexing and lexical query API

## Status

Accepted — 2026-09-22. The verified scoped worker is reachable through authenticated native endpoints and the supplied web Reindex action. Frozen coverage stays 45/236 Classic and 2/42 shared (191/40 unmapped). Workspace-005 still requires its complete visible creation/upload/menu contract; this slice adds no mapping.

## Native endpoints

- `GET /api/workspace/index?scope=all|notes|skills` reports durable state, roots, count, timestamps, generation, configuration hash and `required_roots: true`.
- `POST /api/workspace/index?scope=…` explicitly waits for Begin→ScanScope→Commit/Fail. Competing/lost ownership returns 409; scan/write failures return 500 and retain the committed snapshot. Request cancellation propagates to the worker, whose failure cleanup uses its own bounded context.
- `GET /api/workspace/search?scope=…&q=…&limit=…&offset=…` reads committed FTS/chunks only. Default limit is 10; accepted bounds are 1–50 and offset 0–10,000. Query text must be nonblank valid UTF-8, at most 1,024 bytes, with no NUL. `refresh` is rejected: explicit POST must precede a refreshed query.

All endpoints use the existing instance authentication boundary and private/no-store caching. There is no new per-user/session ACL. Standard scopes follow the Piclaw-derived configuration: notes, `.pi/skills`, and their union. Every configured root is required; missing directories fail rather than creating an empty successful index. Extra roots/settings and optional-root policy are not yet exposed.

Scope membership is explicit. Refreshing `all` does not manufacture a committed generation for `notes` or `skills`; refresh those scopes before querying them separately. Queries filter workspace identity, scope membership, configuration hash and chunker version. Changed configuration cannot expose an old broader index. Same-configuration stale/failed/indexing scopes retain their last committed hits.

## Lexical queries

`internal/search/store/query.go` joins native chunk IDs to scoped documents and FTS. Hits contain document/chunk identity, path, source byte/line ranges, a snippet capped at 512 characters and bm25 rank, with stable path/chunk tie-breaks. Results are chunks, not deduplicated files.

Recognised malformed FTS expressions (syntax errors or unterminated strings) use an AND literal-substring fallback capped at 16 terms. It is parameterised and retains the same ownership/configuration filters. SQLite `lower()` only folds ASCII in that fallback; Unicode literal spelling is preserved. FTS itself uses the schema's Unicode tokenizer. Database/cancellation errors are not swallowed; unknown FTS column errors also propagate rather than masking a missing schema column. Fallback snippets show the bounded start of content, without occurrence highlighting.

There is no vector search, embedding generation or semantic ranking. These filesystem-index routes remain separate from chat search and the pre-existing `fts://workspace` documentation registry.

## Supplied UI integration

The API adapter replaces hard-coded index status/no-op reindex calls. `gi-workspace-visibility.ts` also forwards the global Refresh/Reindex actions to the explorer's existing stateful menu controls, following the hidden-files adapter. It relies on the pinned menu labels/classes and preserves the mounted explorer. Supplied components, UI helpers, panes and stylesheets are unchanged.

The supplied explorer shows native missing/failed/indexing states and hides its indicator when ready. The host does not add a search field or permanent status chrome. Explicit retry is required after failure. Ordinary status/query reads never acquire a lease or launch refresh; no startup watcher, scheduling or automatic invalidation is enabled. Thus `ready` describes the last successful refresh, not a guarantee that later filesystem edits have been observed.

## Verification

Native query tests cover real scoped FTS, source locations, ranking/pagination, Unicode, malformed-expression fallback, query limits, configuration/workspace isolation, cancellation and schema-error propagation. Endpoint tests cover authentication, methods, valid/invalid scopes, never-indexed reads, native worker publication, missing-root failure, held ownership conflict, retained results, retry and server recreation without implicit work.

The six-project browser case writes actual notes and skills, invokes the visible Reindex control, queries stored lexical hits, renames the required notes root to cause a real scanner error, then checks durable failure/reload and retained hits. Restoring the root and explicitly retrying replaces changed content. Draft text/files survive failure, reload and session switches, with no chat submission. Refresh finds a newly written file. No response or lifecycle evidence is fabricated.

Initial fixture failures used the wrong error-label casing and attempted session selection behind the mobile workspace drawer; the final fixture uses the exact supplied label and normally closes the toggle before selecting sessions. No forced clicks were added.

**492/492 browser executions:** 330 main, 54 reconnect, 66 compaction, 18 Steer/admission, 12 context-fit and 12 context-meter. **75/75 functional**, **32/32 helpers**, full Go/vet/build/hook checks and search-store/indexer/web race tests ×3 pass. The first Chromium run hit ENOSPC while writing artifacts/queue gates. Disposable Gi build binaries and test artifacts were moved to `/tmp` (locations recorded in `/workspace/tmp/gi-index-api-artifact-locations.txt`); databases/source/screenshots were retained. One WebKit main run failed at reload with an internal browser error, and one reconnect search case timed out; both complete suites passed unchanged on rerun. Failed runs are retained as limitations, not counted as passes.

Desktop/tablet/phone failure screenshots were attached from `/workspace/tmp/gi-ui-captures/september22/gi-index-failure-*.png`. Logs are `/workspace/tmp/gi-index-api-*`. The ignored local `test-results` symlink points to the relocated disposable artifacts; it is not tracked in Git. Workspace storage remains constrained.

## Terminal and remaining lifecycle

A terminal reindex/status action should run explicitly, use existing transient status or a temporary bounded detail view, preserve draft/cursor/reader on cancellation/dismissal and add no idle rows. No terminal control or PTY acceptance is added here. The earlier worker's native cancellation/lease tests are not terminal interaction evidence.

Next work: application-owned scheduling, invalidation/coalescing, background freshness, settings/root policy, direct query consumers and independent terminal acceptance at 60×18, 100×22 and 140×36. Keep the unscoped full-rebuild prototype stashed; native endpoints use only the migrated scoped implementation.
