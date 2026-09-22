# ADR-0028: Refresh the selected view after message acceptance

## Status

Accepted — 2026-09-22. Classic compose-007, 009, 010 and 011 pass all six browser projects. Compose-008 remains unmapped.

## Acknowledgement and view ownership

The supplied composer captures and persists the outgoing draft before clearing it. Uploads and submission run in the background. Acceptance removes the captured recovery token without clearing newer text, attachments or cursor selection. The native media adapter serialises each uploaded ID with its destination session; the composer retains the matching filename in the attachment reference block.

`App.handlePost` keeps its captured-session guard and calls the existing `refreshAfterConnection` dispatcher. That dispatcher refreshes the current timeline or search, subject to connection readiness. Session-list refresh remains independent. Timeline paging and SSE own near-bottom following and visible-message anchors; acknowledgement adds no separate scroll action.

Previously, `handlePost` called `loadPosts` directly. That loader correctly refuses to populate an active search, so a delayed acknowledgement did not refresh search. The new search regression fails against the old callback and passes with the shared dispatcher. The old scroll helper already checked the near-bottom threshold; it was not an unconditional jump.

Supplied composer, timeline, pane and UI helper files are unchanged.

## Native evidence

`tests/ux/drafts.spec.mjs` adds six cases:

- **compose-007:** hold a real accepted HTTP response while native execution and SSE continue; release acknowledgement, observe another page read, verify the stored ID appears exactly once and survives reload.
- **compose-009:** upload three distinct files without interception; compare returned IDs with outgoing media references and filename blocks, persisted filenames and downloaded bytes.
- **compose-010:** type a new multiline draft, add an unsent attachment and move the cursor while an earlier acknowledgement waits. Release it; verify text, cursor, attachment, pending-token cleanup and reload preservation. Only the captured message is stored.
- **compose-011:** build history through native turns. Scroll away while acknowledgement waits, retain a visible-message anchor within one CSS pixel after release and a later arrival, then verify newest-edge following and exact stored-ID order without duplicates.
- **Gi search guard:** release acknowledgement while a no-match search is open; observe a search request, no main-page request, unchanged query and restored editor draft on Escape.
- **Gi origin guard:** release an accepted origin response after selecting a child; child draft and empty timeline remain intact, and returning to the origin shows the stored post once.

No fabricated SSE, synthetic post payloads, forced clicks, SQL-seeded browser history or paid provider is used. Holding acknowledgement introduces test-controlled delay; there is no one-second delivery guarantee.

History setup waits for native idle after completion because the completed row precedes release of the active engine claim. Uploads use response observation: routing multipart requests through Playwright lost file bytes under WebKit. Scroll measurements wait for wheel/entry animation to settle. The context-fit fixture now polls accepted child-model state instead of racing the click's HTTP mutation.

Final validation: **384/384 browser executions** (228 main, 54 reconnect, 66 compaction, 12 Steer, 12 context-fit, 12 meter), **70/70 functional**, **29/29 helpers**, Go tests/vet and hook checks. Main Chromium and WebKit results were combined from two successful 114-case runs after a renewal interrupted the first full run. The independent review delegate timed out and supplied no review evidence.

Coverage is **34/236 Classic**, **2/42 shared**, with **202/40 unmapped**. Compose-008 requires text plus file, folder and message references, including references-only submission. Folder clicks in the supplied workspace explorer expand/select the directory without invoking `onFileSelect`; a complete folder-reference selection path is not established. Partial file/message evidence earns no mapping for that case.

## Minimal terminal adaptation

No terminal code changes in this slice. Browser evidence does not verify terminal acceptance.

Use the existing transcript and editor for acceptance: reconcile by durable message ID, clear only the captured editor draft, preserve newer text/cursor, and retain the reader's anchor unless already following the newest edge. Do not add a delivery panel, spinner row, toast stack or acknowledgement badge. A failure may use the existing bounded notice/error surface; it must retain recovery data and must not automatically resend an uncertain submission.

The current `ChatApp.submitWithMetadata` clears input once and scopes asynchronous callbacks to the origin session, but accepted routing can still request scrolling. Terminal attachment/reference recovery and accepted-message anchor behaviour need their own tests. Before terminal credit, verify delayed acceptance, A→B→A, failure/uncertain delivery, resize and loaded-history reading at 60×18, 100×22 and 140×36, with unchanged idle row count. This extends the existing pi-style editor/transcript contract without a permanent control.
