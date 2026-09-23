# ADR 0056: Idle single-message deletion

Status: Accepted (web subset)
Date: 2026-09-23

## Scope

The frozen `@ux-timeline-017` case deletes one message with no visible replies. Gi stores a flat message timeline and does not store a reply graph. The native DELETE endpoint accepts a single user or assistant message only when its session has no running, cancelling or queued turn or active claim. `cascade=true` is rejected. The backend does not implement the reply-detection or cascade flows in `@ux-timeline-018`–`022`; the combined copy/delete cases `@ux-original-024` and `@shared-37` remain unmapped.

## Native operation

`DeleteMessage` uses one SQLite write transaction to check message ownership and idle state, delete exactly one row, clear the context summary and covered IDs, and advance the checkpoint version. A concurrent compaction snapshot cannot commit a summary containing deleted text. Turn and event audit records and stored media stay intact. Wrong-session, missing, protected, busy and cancelled operations leave the row intact. The HTTP adapter returns 404, 409 or 400 for missing, busy/protected or unsupported cascade requests respectively, and requires the installed auth guard.

## Browser operation

The supplied Post and Timeline components receive Gi's `onDeletePost` and removal-ID props unchanged. The host captures the post's origin session (including a cross-session search result), waits for native acknowledgement, then applies the existing `.removing` transition and removes the row after 220 ms. Failed requests leave it visible. The page fences pending timeline/search responses and filters acknowledged IDs from later reads. Switching sessions does not redirect a pending deletion or put its error in the new session. The current draft and pending file remain untouched. Reload reads authoritative storage.

## Terminal disposition

Do not add a permanent delete button, row or footer indicator. A future terminal implementation can expose a bounded action on the currently selected transcript message, with an explicit confirmation identifying that message, an idle check and a transient result in the existing notice space. Stable message-ID selection across fullscreen reflow, regular scrollback and session switches must be proved before wiring it. Both modes must preserve draft/cursor, reading anchor and Pi transcript spacing at 60×18, 100×22 and 140×36. No terminal deletion credit is assigned by the browser tests.

## Evidence and limits

The six-project Chromium/WebKit phone/tablet/desktop matrix covers native success/removal delay/reload, failure and session ownership, a held pre-delete timeline page, and cross-session search origin with a held response. Six isolated 75-case browser runs pass (450/450); combined with separate reconnect, context, compaction and Steer results, the report shows 51/236 Classic and 2/42 shared passing. Store tests cover checkpoint conflict and rollback, wrong session, active/queued/protected rejection, reopen/search/page absence and audit/media retention. Go/vet/race checks, 80 functional tests and 39 helpers pass. The frozen feature source and supplied components are unchanged. Inference and reply-graph relationships outside the flat timeline remain for later work.
