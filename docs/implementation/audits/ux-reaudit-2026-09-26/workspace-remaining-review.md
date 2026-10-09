# Classic workspace flows: 15 remaining clauses

This extends `workspace-bounded-review.md` (`004`, `008`, `011`; focused
18/18). The pinned Piclaw 3.2.4 manifest marks `tab-store.ts` and
`tab-strip.ts` identical but `workspace-explorer.ts`, `terminal-pane.ts`,
and `vnc-pane.ts` changed in Gi. Copied explorer/pane handlers cannot be
credited without following the mounted native API and read-only host:
`web/src/app.ts` mounts `TabStrip` and `WorkspaceTab` with read-only preview
capabilities and passes no-op terminal/VNC callbacks.

| ID | Clause-to-native trace | Qualified finding |
|---|---|---|
| `001` | Explorer attempts `createWorkspaceFile(folder,name,'')` and runs untitled-name selection/refresh logic. Native `api.ts` expects `(path,content)` and catches errors to `null`; `/api/workspace/file` is GET-only. Hamburger `new-file` dispatch lacks an adapter consumer. | Native create flow gap; neither create target nor collision/selection aftermath can be credited. |
| `002` | Explorer has rename/remap/dispatch code, but native `renameWorkspaceFile` returns `null`; no mounted `app.ts` listener for `workspace-file-renamed`. | Native rename mutation and tab-sync gap; source-only optimistic UI is insufficient. |
| `003` | Explorer asks for filename confirmation, but `deleteWorkspaceFile` returns `null`; handler treats that as success and can clear selection. | Native deletion gap; a confirmation/optimistic UI is not file deletion or failure handling. |
| `005` | Explorer menu has Refresh/Reindex/hidden/create/upload; no dedicated file-search input in inspected explorer, while Gi has a separate workspace-search endpoint. | Bounded source/UI shape only, no tagged rendered-menu/browser assertion or Piclaw runtime probe; do not infer index refresh from label alone. |
| `006` | Explorer selection/preview logic exists; double-click enters rename mode but native rename endpoint is absent. | Partial native selection/preview path, no tagged single/double-click journey or successful rename. |
| `007` | Explorer has upload progress, conflict prompt and overwrite retry code, but native `uploadWorkspaceFile` returns `null` (call shape also differs). | Native upload gap; no successful path or index refresh from it. |
| `009` | Explorer gates `Open in tab` by specialized handler and `Open in editor` by text/non-directory/256 KiB. Mounted `openEditor` opens a read-only preview tab, not an editable editor. | Partial tab gating; editor action semantics unsupported, no tagged clause test. |
| `010` | Copied `TabStrip`/`TabStore` have dirty and compare affordances, but mounted read-only preview has no editing/saving path to make a tab dirty. | Native dirty-tab/compare capability gap. |
| `012` | `TabStore.rename` updates ID/path/label/MRU/active, and copied `use-editor-state.ts` handles workspace rename events. Mounted `app.ts` does not call that composition or listen for the rename event; native rename API is a no-op. | Source-only unmounted rename-sync contract, no tagged browser assertion. |
| `013` | Copied tab menu gates dock/popout/standalone routes by callbacks; mounted read-only `TabStrip` supplies close/pin, not popout/dock/standalone callbacks. | Native context-route capability gap. |
| `014`–`015` | Copied terminal/VNC panes contain error/read-only handling, but mounted Gi preview host accepts only tab-placed `readonly`/`preview` panes; terminal/VNC callbacks are no-op. | Native mounted terminal/VNC journey gaps; copied pane code is not runtime coverage. |
| `016`–`018` | Native `WorkspaceTab` mounts `PaneContext` in `view` mode with a bounded read; it has no generic editable save, unchanged-save or conflict monitor UI. | Native editable editor/save/conflict gaps. Separate `editor-gap-review.md` reaches the same capability boundary for its five frozen editor clauses. |

Previously reviewed `004`, `008`, `011` are not rescored. The Piclaw 3.2.4
workspace mutation/backend UI and physical-device interactions were not
executed. These are clause findings, not a green browser suite: the fifteen
remaining IDs have no directly tagged Gi journey. No production code or
frozen Gherkin changed.
