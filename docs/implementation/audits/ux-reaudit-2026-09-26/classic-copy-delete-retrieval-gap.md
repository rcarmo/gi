# Classic024–025: combined copy/delete and scoped message retrieval

Both frozen cases join behaviour that narrower Gi tests keep separate.
`@ux-original-024` requires source-Markdown copy, code-copy, reply-aware
cascade confirmation, cancellation and parent/reply removal.
`@ux-original-025` requires bounded explicit row IDs/windows across permitted
single-user scopes and family-owned authorisation.

| ID | Existing Gi evidence | Missing clause |
|---|---|---|
| `024` | `tests/ux/shared-copy-delete.spec.mjs` checks post Markdown and code text copying, then a single idle direct delete. Installed Piclaw copy UI 6/6 and focused Gi Markdown copy 6/6 are recorded in [copy and cascade evidence](../copy-cascade-2026-09-28.md). `message-deletion-review.md` records 017 direct delete at 6/6. | `web/src/app.ts` always calls `deletePost(id,false,originChat)`; `internal/web/message_delete.go` rejects `cascade=true` with HTTP 400. No reply count/prompt/cancel or accepted parent-and-replies removal. The combined frozen case cannot pass from separate copy and direct-delete tests. |
| `025` | The native `messages` tool has durable numeric row IDs, explicit anchors, context and windows, bounded output, missing-row reporting and session isolation. `docs/internal/message-retrieval.md` records the accepted Shared38 current-session web mapping. An [installed in-memory oracle](../message-retrieval-oracle-2026-09-28.md) verifies Piclaw's wider single-user and family-owned reads. | The tool accepts no `session_id` or all-chat scope and has no family-owned authorisation mode. Its strict current-session boundary deliberately excludes the wider Classic request. Shared38 cannot count as Classic025. |

These are **policy/capability gaps** in the combined cases; the narrower native
journeys retain their bounded evidence. `message-deletion-review.md` now
records pinned Piclaw 3.2.4 backend-function and shipped-UI fixture probes.
An isolated in-memory cascade deletes a parent and three direct replies; a
joined shipped-UI/backend-function disposable fixture confirms and deletes
three visible replies with their parent, or cancels and retains all four.
With an unseen stored reply, the same fixture deletes its parent directly and
leaves the reply orphaned. The production Piclaw HTTP router/authentication and live deletion were not
probed. The scoped message-tool runtime was called only against an isolated
in-memory fixture; see the [Classic025 retrieval evidence](../message-retrieval-oracle-2026-09-28.md).
The family/all-chat extension and Gi reply graph need separate identity and
destructive-action decisions. No production code or frozen contract changed.
