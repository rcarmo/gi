# Workspace editor backend

Gi serves complete editable snapshots and revision-conditional writes through the shared Classic CodeMirror/Vim UI at fixtures-vibes `bb7786f` (Piclaw v3.3.0). [Current acceptance](../implementation/web/workspace-editor-acceptance.md) covers 19 shared scenarios and both desktop consumer checks. Issue #46 stays open for the shared conflict scenario: it does not approve the adopted UI's reviewed-Overwrite dialog. The earlier lazy-loader revision loss is fixed by the adopted frontend. Editor behaviour is core/`@cap-workspace` in the current catalogue; `@cap-editor` is no longer defined.

## Files and public assets

Authenticated aliases `/workspace/file`, `/workspace/raw` and `/workspace/stat` use the same handlers and guards as their `/api/workspace/` counterparts. `/static/` maps only to embedded public assets, including `/static/dist/`, `/static/css/` and `/static/editor-vendor/`; it never exposes filesystem or API routes. The editor bundle itself must arrive through the shared frontend build.

`GET /workspace/file?path=relative&max=1000000&mode=edit` returns complete UTF-8 text, metadata and `truncated:false`. Both the initial stat and bounded read reject files larger than 256 KiB with HTTP 400 and `File too large to edit`. Binary/NUL/invalid UTF-8 documents are rejected. Missing edit files return 404; legacy preview reads retain their existing error response.

Without edit mode, preview defaults to 20,000 bytes. Piclaw's `max` parameter clamps to 1 KiB..64 KiB; Gi's existing `max_bytes` parameter keeps its 256 KiB validation bound. Edit mode ignores preview limits to prevent a truncated document from replacing a larger file on save.

`GET ...?mode=edit` includes an opaque `revision`; `PUT /workspace/file`
requires `{path, content, expected_revision}`. Missing revisions return 428
`revision_required`; stale revisions return 409 `revision_conflict`. Existing
regular files only; content is bounded to 256 KiB, permissions are retained,
and unchanged contents preserve mtime. Missing files return 404. Save Copy uses
create-only `POST {path:parent,name:basename,content}`; there is no unconditional
Overwrite or automatic retry. Overwrite requires review of the exact snapshot
whose revision authorises the write. See
[revision-safe writes](revision-safe-writes.md).

Filesystem reads and writes use `os.Root`; path traversal and out-of-root symlinks are rejected. The protocol does not provide a transactional compare-and-swap against concurrent external writers.

## External changes

SSE clients share one bounded fsnotify watcher per web server. It starts at the first client and stops/joins when the last disconnects. No background polling runs without clients. Directory registration is limited to depth four, 1,024 watches and 10,000 entries per registration walk; excluded build/cache/dependency directories and directory symlinks are not traversed. Files outside these bounds do not have guaranteed change notifications; frontend stat polling is still required.

Changes are coalesced for 100ms, with at most 256 path entries. Overflow asks clients to refresh `.`. A slow subscriber receives a root invalidation instead of accumulating events. The SSE shape is:

```json
{"updates":[{"path":"folder","root":{"name":"folder","path":"folder","type":"dir","children":[]},"truncated":false,"changed_paths":["folder/file.md"]}]}
```

Snapshots use the rooted tree reader at depth four for `.` and three for other subtrees, with a shared 10,000-node budget. They contain names/metadata, never file contents. A budget failure yields a directory stub with `truncated:true`. The UI uses concrete `changed_paths` to refresh clean editors and `root` to update explorer subtrees. Session streams are authenticated; events apply to the shared workspace, not just the active conversation.

## Agent file-open requests

The web server registers `open_workspace_file` for existing confined workspace files, with `target=tab|popout` and an optional label. It emits a session-scoped `extension_ui_request` with `options.action=open_workspace_file` and waits up to 15 seconds for the browser outcome. The turn remains cancellable.

`POST /agent/respond` accepts `{request_id, chat_jid, outcome}`. The request must belong to the same browser owner, session, path and target, remain unexpired and uncancelled, and receive only one response. Enrolled instances require a valid browser cookie; bearer tokens, cross-origin requests and unsafe non-localhost transport are rejected. Timeout/shutdown cannot turn a late response into a successful open.

## Verification

[Current browser acceptance](../implementation/web/workspace-editor-acceptance.md) records 19 accepted shared scenarios and the remaining conflict-test mismatch. [Historical backend measurements](../implementation/web/workspace-editor-backend-verification.md) retain the 5 October API/watcher verification.
