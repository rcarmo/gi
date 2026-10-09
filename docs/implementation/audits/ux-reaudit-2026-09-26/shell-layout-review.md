# Classic shell/menu/layout 001–009: bounded native trace

Frozen source is `tests/ux/features/classic/compose/hamburger-layout-scale.feature`.
The pinned 3.2.4 source manifest marks `runtime/web/src/components/tab-strip.ts`
identical and `runtime/web/src/components/timeline-menu.ts` changed. Gi's
`web/src/components/timeline-menu.ts` supplies the menu; `web/src/app.ts`
mounts it. Focused native browser tests for `002`, `003`, `005`, `007` passed
**24/24** across six Chromium/WebKit viewport projects. This is not a direct
Piclaw UI run, physical-device check, or complete shell parity result.

| ID | Gi trace and assertion | Qualification / gap |
|---|---|---|
| `001` | Menu offers enabled New file, Refresh tree, Reindex workspace when open; `gi-workspace-visibility.ts` forwards refresh/reindex to the explorer's native controls. | **Partial native gap**: no tagged three-action journey; the hamburger emits `new-file` but its visibility adapter accepts only `refresh`/`reindex`. The explorer has its own New file control, not proof that this hamburger action works. |
| `002` | Tagged `shell.spec.mjs` checks real hidden-file rows, stored `workspaceShowHidden`, one event per toggle, refresh query, keyboard/pointer, and reload. | 6/6 native fixture projects, source-backed; Piclaw UI unprobed. |
| `003` | Tagged test checks disabled four workspace items in hidden workspace, nonactivation by pointer/keyboard, no mutation requests or draft change. | 6/6; `chatOnlyMode` is hard-coded false in Gi's mounted `TimelineMenu`, so **actual chat-only mode** is not tested. Only workspace-hidden interpretation is covered. |
| `004` | Menu conditionally offers Terminal/VNC on supplied callbacks. | Gi mounted `TimelineMenu` supplies neither callback. `WorkspaceExplorer` receives no-op callbacks, not working terminal/VNC actions. Native capability gap; no tagged journey. |
| `005` | Tagged test compares compose/input geometry to chat-column width through workspace transitions, resize and reload, preserving draft/media. | 6/6 native fixture projects; no pixel oracle. |
| `006` | `timeline-menu.ts` computes a safe-area top offset for its portal; `timeline-menu.css` fixes the portal and styles trigger. | No tagged viewport/safe-area bounds assertion. Source and emulated viewport alone cannot prove physical notch/safe-area behavior. |
| `007` | `tab-strip.ts` stops propagation on close; tagged read-only-tab test closes background tabs by pointer and keyboard without activating or re-reading, preserves draft/media. | 6/6 native read-only fixture; broader editor tab behaviour not implied. |
| `008` | `TimelineMenu` has stored display-scale input and `shell.spec.mjs` has a Gi-native standalone/capability-emulated persistence test. | Test is **untagged** and checks Gi PWA preference/viewport logic, not full frozen Classic control semantics. Scale is hidden in normal browser mode; no clause credit beyond this bounded source trace. |
| `009` | Gi read-only Markdown preview uses `.gi-workspace-tab .workspace-preview-text code { font-family: var(--font-family-mono, monospace); }`. | No tagged rendered inline-code/computed-font test; the frozen editor Markdown extension is not Gi's read-only preview. Do not equate a CSS selector with the editor-preview clause. |

`@ux-shell-001` and `004` require decisions/implementation rather than
retroactive credit from adjacent controls. The installed Piclaw menu/safe-area
and Markdown editor interactions were not browser-exercised in this review.
