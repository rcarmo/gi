# Classic timeline deletion 017–022: direct deletion versus cascade

`features/ux/classic/timeline/message-deletion.feature` separates a
single post without visible replies from backend-detected or visible reply
cascades. The pinned Piclaw 3.2.4 source map's
`src/ui/app-timeline-actions.ts` counts visible `thread_id` children, prompts
before deleting them, sends `cascade=true`, marks returned IDs removing, and
contains a second prompt/retry branch if direct deletion throws `Replies exist`.
The installed 3.2.4 backend source takes `cascade` from
`src/channels/web/http/dispatch-content.ts`, forwards it through
`src/channels/web/endpoints/channel-endpoint-facade-service.ts`, and in
`src/channels/web/timeline-service.ts:204–225` calls
`deleteMessageByRowId` for a direct deletion or `deleteThreadByRowId` for a
cascade. `src/db/messages.ts:519–567` directly deletes one row without checking
`thread_id` children; the cascade query selects the parent and its direct
children by `chat_jid` and `thread_id`. Neither path emits `Replies exist`.
An isolated installed-backend probe (`PICLAW_DB_IN_MEMORY=1 bun
tests/ux/oracle/piclaw-deletion-backend-probe.ts`) pins the installed version
and Classic source-map hash, uses an in-memory SQLite database, and checks both
paths. Direct deletion of a parent with an unseen persisted reply returned 200
with only the parent ID, leaving the reply row with its original `thread_id`.
Cascade deletion returned 200 with the parent and three direct reply IDs and
removed those four rows. This reproduces the `018`/`019` backend-rejection
premise mismatch **at the backend-function boundary**. It does not run the
HTTP route or prove a user-visible deletion in Piclaw.

`tests/ux/oracle/piclaw-deletion-ui-probe.mjs` pins the shipped Classic assets
and runs separate disposable `/timeline` and `/post/:id` fixtures in Chromium
and WebKit desktop. With three visible replies, cancellation sends no DELETE
and confirmation sends `cascade=true` and removes the returned IDs. When the
fixture deliberately returns HTTP 409 `Replies exist` for a hidden reply, the
UI prompts after `cascade=false`: cancellation leaves the parent visible and
confirmation retries with `cascade=true`. These fixture assertions exercise
the shipped UI branch, not the installed backend. A third probe joins the
shipped UI to installed `getTimelineResponse` and `deletePostResponse` through
**disposable routes** and in-memory SQLite:
`PICLAW_DB_IN_MEMORY=1 ORACLE_BROWSER=chromium bun
tests/ux/oracle/piclaw-deletion-combined-probe.mjs`; repeat with
`ORACLE_BROWSER=webkit`. Its deliberately restricted current view presents a real stored
parent while hiding a real stored reply. In both desktop browsers, clicking
Delete sends `cascade=false`, shows no prompt, removes the parent from the UI
and leaves the reply row orphaned (`thread_id` still points to the parent).
The same joined probe also runs `ORACLE_DELETE_CASE=visible-confirm` and
`visible-cancel` (three stored replies). In both browsers, confirmation prompts
with the exact count, sends `cascade=true` and removes all four stored rows;
cancellation sends no DELETE and retains all four. These are joined
browser/backend-function **disposable** checks for the visible-reply clauses.
They do not exercise the production HTTP router, authentication, or live chat.
The backend's 200 direct-delete response cannot reach the synthetic-409 retry
branch in the hidden-reply setup.

| ID | Bounded Gi evidence | Result |
|---|---|---|
| `017` | `tests/ux/message-delete.spec.mjs` holds native DELETE acknowledgment, checks no premature removal, transient removing class, eventual removal, absent stored/search row and reload absence; draft/file survive. `web/src/app.ts:571–600` performs only `deletePost(id,false,originChat)` and animates after success. Mounted installed Classic 3.2.4 with a disposable post/held DELETE separately requests `cascade=false`, shows no prompt, keeps the post until acknowledgment, then marks/removes it; the fixture's reload stays empty and draft survives. | Gi tagged focused run **6/6** and installed fixture **6/6** Chromium/WebKit phone/tablet/desktop. The installed probe has no real DB/search, production HTTP/auth or backend reply graph; removal delay differs (Piclaw 180 ms, Gi 220 ms). No live deletion acceptance. |
| `018` | Gi always sends `cascade=false`. `internal/web/message_delete.go` rejects `cascade=true` with HTTP 400; no reply-detection retry in Gi `handleDeletePost`. Joined shipped UI/installed-backend-function fixture deletes a stored parent with an unseen reply directly, leaving an orphan; a separate shipped-UI fixture retries only after synthetic 409 `Replies exist`. | Verified Gi gap; frozen Piclaw reply-rejection premise fails in the joined disposable fixture. Production HTTP router/live untested. |
| `019` | Gi has no follow-up prompt after `Replies exist`; it displays a deletion error and leaves the post on failure. The joined shipped UI/installed-backend-function fixture receives 200 and removes the parent without a prompt; the separate shipped UI cancels and preserves the parent only after synthetic 409. | Gi cancellation interaction absent; generic failure preservation is insufficient. No production HTTP-router/live check or Gi tagged journey. |
| `020` | Gi does not count visible `thread_id` replies to build the three-reply confirmation prompt. Shipped Piclaw UI with three stored replies shows the exact `Delete this message and its 3 replies?` dialog in joined Chromium/WebKit desktop fixtures. | Verified Gi prompt gap; joined disposable fixture is bounded, without production HTTP or physical acceptance. |
| `021` | Gi server disallows cascade, UI removes only the acknowledged direct ID. Joined shipped UI/backend-function fixtures send `cascade=true` and remove the stored parent and three replies in Chromium/WebKit desktop. | Verified Gi cascade gap; joined disposable fixture does not establish a production HTTP/auth/live journey. |
| `022` | Gi has no visible-reply prompt or its cancel path; generic failure preservation is insufficient. Joined shipped UI/backend-function fixtures cancel the prompt, send no DELETE, and retain the stored parent and three replies in both desktop browsers. | Verified Gi prompt/cancel gap; joined disposable fixture does not establish production HTTP/auth/live acceptance. |

Gi's `internal/store/message_delete.go` enforces native session/turn constraints
and deletes a single flat row. Its `internal/store/schema.go` messages table has
no `thread_id` column; `web/src/components/timeline.ts` can render a
`data.thread_id` supplied by the Classic side, but this does not give Gi a
persisted reply graph. `tests/ux/shared-copy-delete.spec.mjs` checks busy-turn
rejection and later direct deletion, not cascade. The HTTP/store deletion tests
cover direct deletion and cascade rejection; a focused
`go test ./internal/web ./internal/store -run 'Test.*(DeleteMessage|MessageDelete|DeletePost)' -count=1`
passed. No Gi test constructs and deletes a persisted reply graph. An API
argument named `cascade`, CSS removal animation, and generic failure
preservation cannot satisfy `018`–`022`. Reply-aware Gi deletion needs an
explicit identity/data-model and destructive-action policy decision.

The active Gherkin was parsed by `loadCorpus()` as six scenarios (`017`–`022`),
each with steps; only `017` is mapped. The frozen historical snapshot at
`features/ux/upstream/classic-snapshot/timeline/message-deletion.gherkin` was
not revised. `make test-piclaw-direct-delete` and the focused Gi `017` journey
both passed **6/6** across Chromium/WebKit phone/tablet/desktop. This confirms
fidelity of Gi's already implemented single-post path within the stated
fixture boundaries; no reply-aware feature port, production-code change,
frozen-feature change or live-chat deletion was made.
