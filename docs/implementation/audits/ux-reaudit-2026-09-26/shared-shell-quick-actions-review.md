# Shared1–15: shell and Quick Actions (Shared8 excluded)

The shared feature expands its first six definitions into fifteen cases.
Gi's mounted `TimelineMenu` and `TimelineQuickActions` handle focus, dismissal,
scoped results and keyboard guards. The pinned Piclaw 3.2.4 source manifest
marks `timeline-quick-actions.ts` identical; its UI was not replayed here.

| Shared cases | Native browser assertion | Boundary |
|---|---|---|
| `1`–`2` | Pointer/outside and keyboard/Escape menu paths reach enabled items, close without activating the underlying Send button, and restore usable focus. | `classic.spec.mjs`; trusted touch is a separate Gi test. |
| `3` | Workspace show/hide and narrow drawer backdrop retain current session, draft and references. | `workspace-preview.spec.mjs`; no CRUD inference. |
| `4` | Idle timeline typing focuses and ranks grouped Quick Actions, wraps selection and activates once. | `quick-actions.spec.mjs`; current Piclaw UI unprobed. |
| `5`–`7`, `9`–`11` | Real composer, Settings input/select, button/link, sidebar, modal and picker targets retain keyboard input without opening Quick Actions. | Shared8's real contenteditable editor remains absent. |
| `12` | Consumed, repeated, composing, whitespace and modified events do not open the palette, with a positive eligible-key control after each rejection. | Event metadata fixtures do not validate physical IME. |
| `13`–`15` | Escape, outside pointer and close control restore the opening trigger and retain session/draft/media/references without executing an action. | Touch and reload checks are additional Gi-native assertions. |

The focused shared filter passed **84/84** across six Chromium/WebKit viewport
projects: fourteen shared cases, **not** Shared8. The parity report recorded
14 pass, 19 not-run and nine unmapped in the whole 42-case shared matrix.
Shared8 has only artificial contenteditable selector coverage; Gi's mounted
workspace editor is read-only. The current Piclaw 3.2.4 UI, physical touch and
assistive technology were not tested. No frozen text or production code
changed.
