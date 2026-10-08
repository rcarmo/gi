# Revision-safe editor and Plan writes

Complete editor and Plan reads return an opaque revision. Browser mutations
must send the revision they loaded; Gi rejects missing or stale preconditions.
This is the Gi backend slice of fixtures-vibes #1. The pinned `b17ef01` UI does
not yet send revisions, so its saves fail closed until the authorised upstream
frontend change is adopted. No unconditional compatibility fallback exists.

## HTTP fields

| Operation | Contract |
| --- | --- |
| GET `/api/workspace/file?path=...&mode=edit` (also `/workspace/file`) | Complete editable UTF-8 text, `truncated:false`, `revision:"file-v1-..."`. Preview reads grant no revision. |
| PUT `/api/workspace/file` | `{path,content,expected_revision}`; success returns `mtime`, size and new `revision`. No-op saves preserve mtime/revision. |
| POST `/api/workspace/file` | `{path,name,content}`, create-only. Existing names return 409 `file_exists`; PUT never creates missing files. |
| GET `/api/sessions/{s}/plan` | Existing `{ok:true,plan:{...}}`, with `plan.revision:"plan-v1-..."`. |
| POST Plan | `{action:"write"|"reset",markdown?,expected_revision}`; success contains new `plan.revision`. |

Missing revisions return 428 `{code:"revision_required",error:...}`. Stale
revisions return 409 `{code:"revision_conflict",error:...,revision:<current>}`
without changing content. The error revision is informational; a client must
review a complete current snapshot before intentional Overwrite. It must not
auto-adopt the value and retry. Drafts without loaded baselines are read-only
until the user resolves them. Save Copy uses POST create-only.

Plan conditional mutations acquire SQLite's write lock before checking the
stored revision. Browser writes and internal agent mutations share that
transaction boundary; every mutation advances the returned revision. File
revisions include complete bytes, mtime and mode. Gi serialises its workspace
mutations and checks current content immediately before the write. External
processes do not participate in this in-process lock; a filesystem change
between the check and write cannot be excluded by this contract.

## Agent file opening

The web engine registers `open_workspace_file({path,target?,label?})`. It
opens an existing regular workspace file only after `os.Root` confinement
checks, emits a private `extension_ui_request` with `kind/method:"custom"`
and `options.action:"open_workspace_file"`, then waits at most 15 seconds.
Only an existing browser SSE listener for that session receives it. Conflicting
browser-cookie owners cause rejection rather than arbitrary delivery.

POST `/agent/respond` requires same-origin HTTPS or localhost and the enrolled
browser cookie. Request ID, owner, session, target and path must match; expired,
cancelled, revoked, foreign and replayed replies reject. Pending requests and
listener queues are bounded. Cancellation and process shutdown release waiters;
there is no durable request backlog or automatic retry. Unenrolled localhost
sessions share the explicit auth-free local-browser owner. Bearer-only clients
do not receive or answer UI requests.

## Verification state

Twelve revision/editor/Plan checks passed three times with race detection;
nine file-open checks passed three times, including actual SSE delivery. Tests cover missing/stale revisions,
concurrent Plan/file saves, agent invalidation, reset, unchanged-save mtime,
create-only copies, path confinement, wrong/revoked browser owners, session
isolation, replay, cancellation, timeout and shutdown. CPU and both allocation
views were analysed and automatically disposed. Full Go regression, vet and
a CGO-disabled build passed; no deployment was made.

Browser agent-open checks against `b17ef01`: missing/outside paths pass;
valid-file and other-chat cases fail because `src/gi-sse-client.ts` does not
subscribe to extension request/status events. The authorised frontend owner
has been given that finding. No editor capability or skip removal is justified
by these backend checks. Conditional-save browser acceptance must use the new
frontend revision once published.
