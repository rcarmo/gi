# Classic workspace 004, 008, 011: bounded clause review

The installed Piclaw 3.2.4 release, asset `990f0c49a932`, supplies source in
`app.bundle.js.map`. This review compares that source with frozen Classic
requirements, disposable Gi browser assertions, and Gi handlers. The three
current Piclaw UI interactions were **not** directly run here; source
comparison cannot establish current-release behavioural acceptance.

| Frozen ID | Piclaw source boundary | Gi requirement → assertion → code evidence and limit |
|---|---|---|
| `@ux-workspace-004` hidden-file toggle | Installed `components/workspace-explorer.ts:1745–1765` persists `workspaceShowHidden`, calls `setWorkspaceVisibility`, invalidates tree requests, reloads root and expanded subtrees. This module **changed** versus the frozen 70d33bc tree, so the cited current source matters. | `tests/ux/workspace-preview.spec.mjs:38–68` toggles through the visible menu, checks hidden path visibility, `/api/workspace/tree` depth-one root/subtree reads with `show_hidden`, persistence after reload, and unchanged draft/media/no turns. Gi `web/src/components/workspace-explorer.ts:1568–1578` follows that flow. No live Piclaw browser proof; the test does not prove every error path or physical touch. |
| `@ux-workspace-008` preview kinds and metadata | Installed `panes/workspace-preview-pane.ts` is byte-identical to the frozen version (`frozen-to-oracle-sources.json`); the frozen scenario lists Markdown, text, image, binary and available metadata. | `tests/ux/workspace-preview.spec.mjs:21–37` loads native Markdown, escaped text/SVG, image and binary through workspace selection and checks metadata plus large-text truncation; Gi preview source is `web/src/panes/workspace-preview-pane.ts`. The test also keeps the composer draft across reload. It does not establish identical Piclaw rendering pixels, every content type or editor write capability. |
| `@ux-workspace-011` MRU/pinned tab closing | Installed `panes/tab-store.ts` is byte-identical to the frozen version. Its `close(id)` takes the next active tab from MRU and `closeOthers`/`closeAll` preserve pins. | `tests/ux/workspace-tabs.spec.mjs:138–158` uses three native **read-only preview** tabs. It asserts MRU activation after active close, protects a pinned tab during Close Others/All, allows explicit individual close, verifies the context actions and retains the composer draft/media. Gi `web/src/panes/tab-store.ts:80–125` holds state. Addon/terminal/editor tabs and cross-window sync are outside this journey. |

`make test-ux-parity UX_PARITY_PORT=19134
UX_PARITY_ARGS='tests/ux/workspace-preview.spec.mjs tests/ux/workspace-tabs.spec.mjs --grep "@ux-workspace-00[48]|@ux-workspace-011"'`
passed **18/18** across Chromium and WebKit phone/tablet/desktop projects with
a disposable Gi API. These three rows have bounded source-and-native-test
review; they earn no Piclaw 3.2.4 runtime or physical-device acceptance.
