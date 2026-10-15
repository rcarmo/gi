# Classic session picker 013–015: bounded native paths

The frozen Classic picker clauses require focus and dismissal, chat-scoped
selection with stale-response guards, and actions supplied for each session.
Gi's mounted `web/src/app.ts` advances `selection` before session-switch
refreshes, invalidates timeline state, and wires rename, pin, archive and
restore through `handleSessionMutation`. The picker uses the copied
`compose-session-switcher.ts` (byte-identical in the pinned 3.2.4 manifest).

| ID | Focused Gi browser evidence | Limit |
|---|---|---|
| `013` | `tests/ux/session.spec.mjs` opens by pointer and keyboard, checks focused search and grouping, then Escape closes and restores trigger focus without changing the selected session. | Six viewport/browser projects passed. |
| `014` | Picker selects another native session while an earlier messages response is held; selection and scoped refresh reject the superseded response. | Six passed; fixture is not a live Piclaw backend. |
| `015` | Visible native pin/rename/archive/restore actions reach their own API paths and handle invalid rename or cancelled archive without reporting success. | Six passed. It does not require every skin to expose deletion or child creation. |

The focused `make test-ux-parity` filter passed **18/18** across Chromium/WebKit
phone, tablet and desktop. The separate `@ux-session-001`–`006` review is in
`session-switching-review.md`. Neither run establishes current Piclaw 3.2.4
browser behaviour, physical keyboard acceptance or broader session-tree
mutation parity. No frozen text or production code changed.
