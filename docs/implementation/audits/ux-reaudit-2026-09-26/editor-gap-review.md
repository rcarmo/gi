# Classic editor-stability 001–005: native capability gap

The frozen `features/ux/classic/editor/editor-stability.feature` describes
an **editable** Classic workspace. Gi currently mounts read-only preview tabs
(`web/src/gi-workspace-tab.ts:7–26`, `web/src/app.ts:979+,1085+`). Its tab store
has generic `dirty` and pin fields, but the Gi app does not mount an editor
buffer, file-save handler, Markdown preview splitter or editable zen-mode
journey. `docs/feature-parity.md:47` and `docs/checklists/implementation.md:203`
explicitly bound the shipped native tab capability to read-only previews.

| Frozen ID | Missing native acceptance | Existing evidence that cannot substitute |
|---|---|---|
| `@ux-editor-001` | Two **editable** file tabs switch without editor loading/flicker. | `tests/ux/workspace-tabs.spec.mjs` opens read-only previews and fences late reads; no editor buffer. |
| `@ux-editor-002` | Dirty editor tab closes only after a browser confirmation, and dismissing it keeps the tab. | Tab-store `dirty` state and pinned bulk-close tests do not create an unsaved editor document or dismiss a confirmation. |
| `@ux-editor-003` | Primary pointer-down on an inactive editor tab activates it before mouse-up and changes editor content. | Read-only tab clicks do not assert pointer-down timing or editable content ownership. |
| `@ux-editor-004` | Markdown editor preview retains rendered content and persisted height during splitter drag. | Native bounded Markdown **file preview** is not a split editor preview. |
| `@ux-editor-005` | Zen mode hides workspace/chat but keeps the editor document visible. | Switching between conversation and read-only tabs is not editable zen mode. |

All five rows have **no directly tagged Gi browser test** and no native editor
capability. They remain unmapped rather than earning parity through CSS class
names, generic `tabStore` functions or read-only tab coverage. This is a
capability/acceptance gap, not evidence that the frozen Classic editor
requirement is wrong.

`make test-piclaw-editor-tabs` passed **6/6** using the installed 3.2.4
Classic UI and disposable file-read routes across Chromium/WebKit phone,
tablet and desktop. It opened two editable tabs, activated the inactive tab
on primary mouse-down before mouse-up, changed the visible editor content,
then made a browser-local edit. Dismissing the native dirty-close confirmation
kept both tabs and the unsaved text; the composer draft was unchanged and no
workspace save request occurred. The existing typeahead probe separately
establishes that a real CodeMirror editor owns its typing. Markdown split
preview, zen mode, production file saves and physical pointer behaviour
have not been tested. None of these installed-UI observations supplies Gi's
missing editable editor.
