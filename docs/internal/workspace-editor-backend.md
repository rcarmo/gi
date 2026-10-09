# Workspace editor backend

Gi serves complete editable snapshots and revision-conditional writes. The
shared Classic frontend is pinned to `0259e9a` (Piclaw v3.3.0); editor saving,
Plan and agent-open checks are recorded in
[8 October acceptance](../implementation/maintenance/acceptance-2026-10-08.md). Issue #46 stays open: the lazy
editor loader drops revision metadata on clean SSE refresh, and shared
conflict018 lacks approval for the new reviewed-Overwrite dialog.
`@cap-editor` is unclaimed.

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

## Verification and performance, 2026-10-05

* Seven focused handler/watcher tests pass, including complete reads, size/encoding/path rejection, unchanged-save mtime, compatibility assets, authentication, shared resources, shutdown and external-write SSE snapshots.
* Full Go suite: 2,127 tests passed, 37 packages; 83.5s wall, 62.2s CPU, 752MB peak RSS including compilation. Versus the preceding run: wall -11%, CPU -2%, RSS -1%; no reported test regression.
* Functional suite: 144 passed, 11 skipped; complete lifecycle 204.1s wall, 98.5s CPU, 370MB peak RSS. Versus the earlier full run: wall/CPU +19%, RSS -8%, with two new tests and filesystem monitoring. No isolated latency regression is established by that aggregate comparison.
* The two new functional checks exercise complete edit/save/oversize rejection and tool-written external changes arriving over SSE without contents. Their selected run passed in 1.6s; the 133.5s lifecycle included a cold build after cache trimming.
* Vet passes. The MCP wrong-issuer test flagged at 1.5s in an intermediate full profile passed its focused profiled check in 8ms; no reproducible slowdown was found.

Focused sampled allocations are mainly bounded preview/JSON responses and test response buffers. The text/content aliases share one string allocation. Watch registration uses bounded directory reads and constant-time watch-set membership rather than allocating a complete watch-list copy per directory. No stable browser allocation measurement or race-detector result is claimed.

The measurements above are historical. Used raw logs/profiles and matching
test artifacts are disposable; current runs use project-owned scratch as
described in [8 October acceptance](../implementation/maintenance/acceptance-2026-10-08.md).
