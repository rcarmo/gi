# Compact message-reference labels

Superseded. `019619c` removed these display labels, and `f106199` (2026-10-04) made the reference itself the numeric row ID: chips show `msg:<row id>` and submissions carry `message:<row id>`, as in Piclaw. The text below records the 2026-09-28 design.

The long-ID report affected the composer and queued-message reference pills. Installed Piclaw renders `msg:<numeric id>` with the full reference in the tooltip. Gi stores canonical string IDs such as `msg_<nanoseconds>`; replacing those identifiers would break drafts, links and API consumers.

Conversation paging/search now exposes the existing unique `message_rows.row_id` as optional `display_row_id`. The supplied compose component receives a display-label map through a guarded build adapter. Loaded references display `msg:<row number>`. References outside loaded history use an explicit middle ellipsis rather than inventing a numeric ID. The full canonical ID remains in the tooltip.

Storage IDs, cursor keys, message links, draft references, removal callbacks and submitted `message:<canonical id>` text are unchanged. Raw pagination/search and exports do not acquire a display ID. No schema change or production migration is needed. The existing row identity is not a hash and survives database reopen.

## Verification

- Installed Piclaw 3.2.4 source probe verifies composer/queue pill prefix, full tooltip and submitted-reference contract. This is a source oracle, not whole-journey Piclaw parity.
- Store race×3: stable row labels after reopen, session-scoped paging/search, distinct rows and raw-view compatibility.
- Browser functional tests in Chromium/WebKit: compact label, full tooltip, draft reload and canonical submitted prompt.
- Six browser/viewport queue tests: compact and unknown-reference labels, reload, native queued prompt preservation, return to composer, removal and empty queue.
- Final `make check`: Go/vet/build/hooks and 144 functional tests passed; 11 skipped.
- Final support inventory: 206 tests / 7,760 assertions passed. No new mappings.
- Focused SQL/identity/adapter review found no blocker. `message_rows.message_id` is uniqueness-enforced by the existing schema.

Protected component/UI/pane source and historical contracts remain unchanged. No deployment, restart, live-chat write or canonical-ID migration. Evidence log: `/workspace/tmp/gi-reference-final-all.log`.
