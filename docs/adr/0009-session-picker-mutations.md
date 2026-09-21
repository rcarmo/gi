# ADR-0009: Persist reversible session-picker mutations

## Status

Accepted — 2026-09-21

## Context

The frozen `@ux-original-015` contract requires capability-gated session actions, distinct pin/rename/archive/restore paths, and errors rather than false success. Gi's web adapters returned `null` for rename, prune and restore. The store already supported display titles and a JSON state document, but no complete mutation API.

## Decision

Expose `PATCH /api/sessions/{id}` with one action per request:

- `{"action":"rename","title":"Display name"}`: trim and validate a single-line title of 1–160 Unicode characters. This does not rename the agent or alter its routing scope.
- `{"action":"pin","pinned":true}`: persist a picker preference in session state. Unlike the pinned Piclaw helper's browser-local preferences, Gi stores this across browser profiles.
- `{"action":"archive"}`: set `archived_at` for an idle child session. Main sessions, queued/running turns and active turn claims are rejected. Repeated archive retains the first timestamp.
- `{"action":"restore"}`: remove `archived_at`. History, identity, parent linkage and pinned preference remain intact.

Use the existing title and JSON state fields, updated in one transaction. File-backed SQLite already reserves the writer at transaction start. Reject unknown actions, unknown request fields, missing pin booleans and invalid names. Return 400 for invalid input, 404 for missing sessions and 409 for state conflicts.

Archive is a reversible picker classification, not deletion, access control or agent shutdown. Direct routing to the session remains possible after archiving. The idle check applies at commit time; it is not a permanent prohibition on later work. Do not imply stronger lifecycle semantics in the UI. There is no permanent-delete control in this slice.

The picker exposes controls only when the row and host callback permit them. It shows an inline rename editor and an archive confirmation. Failed mutations preserve selection and draft and show the server error. Successful metadata operations do not switch chats. Recheck constraints on the server because a row can become busy after rendering.

Guard session-list refreshes with a revision so pre-mutation responses cannot undo acknowledged metadata in the view. Guard picker feedback with a mount/open epoch so late results cannot leak into a later opening. Do not optimistically remove rows or change selection. Selected and runtime-active sessions are distinct concepts.

Forked sessions reset archive/pin metadata and queue count rather than inheriting the source's picker state. They retain their own native identity and copied history.

## Consequences

No schema migration or destructive history rewrite is needed. Existing append-only turn events, WAL and agent/main-session identity contracts remain unchanged. These operations are metadata mutations, not new timeline messages or turn events.

Clients can share pinned state. This intentionally differs from upstream browser-local pinning and is recorded in the component provenance manifest. Native runtime status informs activity; browser selection no longer pretends a session is busy.

Terminal adaptation uses the existing bounded session selector with a temporary row-action submenu, a one-line name prompt and explicit archive confirmation. Escape returns to the selector/editor with its draft intact. Archived entries appear only in an on-demand group/filter; no idle rows, sidebar, header or persistent action bar are added. This design still needs terminal implementation and acceptance evidence.

## Verification

- `TestSessionMetadataMutations`: validation/status errors, root and busy guards, identity/model preservation, idempotence and fork metadata reset; file-backed SQLite and race runs.
- `@ux-original-015`: native browser controls, invalid-name failure, pin/rename/archive/restore/unpin persistence, confirmation cancellation, reload, capability gating, preserved draft/selection and unchanged agent identity.
- Delayed real PATCH error: dismiss the originating picker before delivery, then reopen without stale error/success feedback.
- All existing picker cases run alongside mutation cases across Chromium/WebKit × phone/tablet/desktop. Existing functional web, Go, vet and source-hash checks remain required.
