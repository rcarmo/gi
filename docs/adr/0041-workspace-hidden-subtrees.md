# ADR-0041: Native hidden-file and subtree queries

## Status

Accepted — 2026-09-22. Classic workspace-004 passes all six browser projects. Coverage is 45/236 Classic and 2/42 shared, with 191/40 unmapped. Reindexing, file mutations and terminal file navigation remain separate work.

## Native tree contract

The host previously fetched the same two-level root tree for every expansion and ignored the depth and hidden-files arguments. `/api/workspace/tree` now accepts workspace-relative `path`, `depth` (0–8) and `show_hidden` (`true`/`false`). The normal explorer requests one level; folder summaries request eight. Native responses remain directory-first/name-sorted and root-relative, with private/no-store caching.

Reads use `os.OpenRoot`. Outside/broken symlinks and unsupported entries are omitted from child listings; direct unavailable/escaping paths fail. Relative in-root symlinks work. The traversal has a 10,000-node total budget and bounded per-directory reads. Exceeding those limits returns HTTP 413, never a successful partial tree. Depth bounds limit cycles from in-root directory symlinks. These limits are not paging or complete large-workspace indexing.

A fetched empty directory returns `children: []`; an unexpanded stub returns `null`. The supplied explorer preserves cached descendants for stubs but clears them for explicit empty snapshots.

No-argument calls retain the old depth-two policy: hidden entries such as `.pi` and `.piclaw` appear, while `.git*` and `node_modules` are excluded. Explicit queries use the requested hidden flag and include ordinary dependency folders; `show_hidden=true` includes dot entries. Hidden visibility is presentation state, not an access-control boundary.

## Host visibility bridge

The shared CSS hides the explorer's own menu in favour of the global menu. That global menu persists `workspaceShowHidden` and emits `piclaw:toggle-hidden-files`, but the pinned explorer does not handle the event. It has a working stateful toggle in its own menu.

`web/src/gi-workspace-visibility.ts` connects the host event to that existing control. It opens the CSS-hidden explorer menu programmatically, waits for its rendered toggle, and invokes it only when its state differs from the requested value. The supplied handler persists state, updates its visibility flag, and reloads root plus expanded subtrees. The bridge does not remount the explorer, replace tree responses or alter expansion state. Its observer/listener are removed with the workspace surface. It relies on the pinned explorer's `Show hidden files`/`Hide hidden files` labels and menu classes; upstream component changes require this adapter to be reviewed.

Gi has no workspace push subscription to reconfigure. `setWorkspaceVisibility` returns the local intent without a backend mutation; each subsequent pull carries the hidden flag. Supplied components, UI helpers and stylesheets are unchanged. Browser tests click the visible global menu normally, without forced clicks; internal forwarding belongs to the production host adapter.

## Evidence

The workspace-004 test creates native root/nested hidden and visible files, expands three levels, then toggles on/off/on through the visible menu. It observes real `depth=1` requests with the correct flag for root and every expanded directory, verifies rows, and reloads to check persisted visibility. Unsent text/files survive and no message is submitted. The existing preview case also passes with one-level subtree loading.

Initial fixture failures found the hidden local menu and the missing event listener. Folder summaries also exposed the supplied depth-eight request, so the bounded server limit accepts eight. The final empty-message assertion normalises the API's `null` to an empty list; no acceptance criterion was removed.

Native tests cover legacy/default policy, explicit flags, sorted root/subtrees, deep paths, empty/stub children, invalid inputs/methods, symlink confinement, node-budget rejection and authentication. The functional test exercises subtree queries and the visible toggle through the browser.

**486/486 browser executions:** 324 main (162 per browser family), 54 reconnect, 66 compaction, 18 Steer/admission, 12 context-fit and 12 context-meter. **74/74 functional**, **31/31 helpers**, full Go tests/vet/build/hook checks and web race tests ×3 pass. Coverage reporter checks keep workspace-005 unmapped. Frozen feature hashes are unchanged.

Desktop/tablet/phone screenshots were attached from `/workspace/tmp/gi-ui-captures/september22/gi-workspace-hidden-*.png`. Logs are under `/workspace/tmp/gi-tree-*`; browser artifacts remain under `test-results/ux-parity/`.

## Terminal adaptation

A temporary, height-bounded file chooser may expose a show-hidden toggle in its own key hints. Keep visibility local to that chooser/session, reload only its current directory, preserve highlighted path where it still exists and restore the draft/cursor/reader on Escape. No persistent tree/sidebar, added idle rows or hidden-file status badge is needed. Retain Pi transcript padding.

Independent three-size acceptance must cover hidden-only/empty folders, nested navigation, inaccessible paths, toggling with selection, resize and dismissal before terminal credit. No terminal implementation changed or PTY suite was rerun in this browser-only slice.
