# ADR-0025: Native scoped search with reconnect ownership

## Status

Accepted — 2026-09-22. Search view and reconnect-003 implemented; hashtag navigation and result paging remain open.

## Native search

`GET /api/sessions/{id}/search` accepts `q`, `scope`, `limit` and `offset`. Scope is `current`, `root` (the ancestor root and its descendants) or `all`. Session existence and parameters are validated before execution. Queries are trimmed, nonempty and at most 512 bytes; limit is 1–100 (default 50), offset 0–10000. Invalid requests return 400, unknown sessions 404, unsupported methods 405. Database failures return a generic `Search unavailable` error.

Search uses a bound literal substring with SQLite `instr(lower(content), lower(query))`. `%`, `_` and SQL syntax in the query are ordinary text, not wildcard/operator input. Case folding is SQLite's ASCII behaviour; full Unicode folding and leading/trailing-space literal search are not supported. Results sort newest first with message ID as a stable tie-breaker. Recursive family queries use `UNION`, terminating malformed ancestry cycles; cycles without a root produce no family results.

The endpoint sits under the existing authenticated session route. `all` intentionally searches the authenticated workspace database, matching the supplied “All chats” option and existing session inventory. This server has no per-session principal ACL; this is not a multitenant isolation feature. Search rows retain their true origin session IDs.

## Search view

The app wires the supplied composer's independent search field and scope selector. Normal composer content/media remain mounted and unchanged. Opening/closing search advances a mutable view generation before rendering; each query/scope update also advances it. Replies require current session, view, request and connection generations.

While search is active, the main timeline loader returns without making a request. Native posts cannot append unfiltered rows to the result view. Reconnect, SSE invalidation and periodic refresh rerun the active query while separately refreshing activity, queue, models and context. A pre-search timeline response or older query cannot replace the result view. Search errors stay local to the current query/session, and reconnect can recover them. Escape/Close returns to the main timeline and unchanged composer draft/media.

The UI explicitly says “up to 50 results”. It does not claim pagination or complete history export. Selecting another session closes the current search view. Query/scope persistence across page reload is not implemented. Hashtag callbacks remain unwired.

## Evidence

- Frozen reconnect-003: real SSE disconnect with active search; native work completes while offline; reconnect updates search results/activity/queue/context without a main-timeline request. Old held timeline data cannot replace search. Escape restores draft/media and the normal measured meter.
- Gi-only six-project tests: current/family/all scopes, literal percent/underscore, stale query response rejection, zero prompt submissions, failed query recovery and unchanged composer draft.
- Native API tests: branch isolation, all-chat scope, limits/offset validation, unknown sessions, literal SQL/wildcard input, and cyclic ancestry termination.
- Helper tests: view/query/scope generations invalidate earlier state before Preact renders.

Validation: 36/36 reconnect/search executions; 330/330 combined browser executions; 70/70 functional; 27/27 helpers; full Go tests/vet, hook checks and native search races repeated three times. Coverage: Classic 29/236 passing (207 unmapped); shared 2/42 passing (40 unmapped). The reconnect feature's search alternative is covered; hashtag navigation receives no credit.

A focused review led to periodic search refresh and generic server errors. Its “all chats” scope concern was checked against the existing workspace authentication model and recorded above.

## Terminal adaptation and limits

Use the same store query behind an on-demand terminal search selector, capped at six results with origin labels, arrow/Enter selection and Escape returning to the original editor/cursor. Reuse the session-picker visual style and status line; no persistent search row, sidebar or extra idle chrome. Current/branch-family/all scope changes should stay inside that temporary view. Validate 60×18, 100×22 and 140×36, cancellation, resize and stale-session results before acceptance. This slice adds no terminal UI.

Full search pagination, hashtag navigation, large-history performance, Unicode folding, per-principal ACLs and active-turn crash recovery remain open. Indexed FTS is not introduced for this bounded first implementation.
