# ADR-0027: Page native timeline history without moving the reader

## Status

Accepted — 2026-09-22. Gi-specific paging foundation; no additional frozen feature mapping.

## Native page API

The session messages endpoint now honours `limit`, `before` and `after`. A paged request returns chronological `messages`, `has_more`, and opaque `before`/`after` cursors. The initial page is the latest 50 rows. Backward pages return the nearest older rows; forward pages return the next rows after the supplied cursor. Limits are 1–100, and requesting both directions is invalid.

Cursors contain session ID, creation time and message ID, ordered as the tuple `(created_at, id)`. They are conflict/navigation data, not authentication tokens. Parameterised tuple comparisons handle equal timestamps and remain usable if the cursor row is deleted. Foreign-session, malformed, oversized and invalid-limit requests reject with 400; an unknown paged session returns 404. The existing authenticated session route remains the boundary. Unpaged requests preserve the original full-history export response for existing consumers.

## Browser window and scroll

The host keeps a loaded-history window and opaque cursors. Older requests return their real promise and share an in-flight slot; refresh invalidations during a request coalesce into a follow-up refresh. Live posts merge by durable message ID. ID-only `agent_response` completion notifications invalidate state but cannot overwrite full post rows.

Refresh loads forward from the last accepted cursor rather than replacing the window with the newest page. Reconnect drains as many bounded pages as required, keeping already loaded history and filling the intervening gap. Selection, connection, request and search-view ownership reject obsolete responses. Search entry cancels the page's ownership; returning to the main timeline starts a new latest-page window.

The supplied timeline's reverse prefetch formula assumes positive scrolling, while its CSS uses negative `scrollTop`. The host disables that component's load-more path and listens on the existing timeline element for native scroll near the history edge. Supplied component files remain unchanged. Native wheel/touch/scrollbar interaction determines paging; there is no added permanent control.

Visible message anchors preserve position across prepend, live append and reconnect. Anchor measurements exclude CSS entry transforms and persist across consecutive programmatic updates, avoiding cumulative subpixel drift. User wheel/touch/pointer/key interaction or resize resets the anchor. Newest-edge readers stay at the newest edge; users reading history are not forced down. Browser assertions permit at most one CSS pixel of rounding after layout settles, not a relaxed multi-row jump.

## Evidence

- Native tests: 125 tied-timestamp rows, stable ordering, deleted cursor, backward/forward multi-page traversal, foreign/invalid cursors and API compatibility/limits.
- Six-project browser regression: 61 native shell turns build history; reload initially loads only 50 rows; wheel loads older pages; live append preserves an actual visible-message anchor. More than one forward page arrives offline and reconnect retains all loaded IDs without duplicates or gaps. A held older page cannot merge into an active search; exhausting history terminates correctly. Draft text is retained.
- Helpers verify deterministic merge/order/deduplication.
- Validation: 54/54 reconnect/paging executions, 348/348 combined browser executions, 70/70 functional tests, 29/29 helpers, full Go tests/vet, hook checks and native paging races repeated three times.

Initial runs exposed completion notifications replacing full post payloads and WebKit anchor drift; both were fixed before the final matrix. The SSE proxy now waits for a real tracked socket before severing it during rapid session activation. A review delegate timed out and supplied no evidence.

Coverage stays Classic 30/236, shared 2/42 (206/40 unmapped). These paging checks do not award unrelated message-action or attachment-rendering contracts.

## Limits and terminal adaptation

Chronological cursors assume ordinary append-only insertion times for reconnect catch-up. Backdated inserts and edits/deletions outside a fetched page do not automatically reconcile into the loaded browser window. The full export endpoint remains available; reload starts a fresh latest view. Long-history performance and bounded eviction of loaded browser rows remain open.

The local terminal already has transcript scrolling/scrollback limits. A future persistent-history load should be on-demand through those controls, with a bounded page and anchor retention, no idle row or panel. It must preserve editor/cursor and pass 60×18, 100×22 and 140×36. This browser slice adds no terminal implementation or acceptance credit.
