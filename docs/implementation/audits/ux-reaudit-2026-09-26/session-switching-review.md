# Classic session switching 001–006: bounded clause review

Shipped Piclaw 3.2.4 asset `990f0c49a932` supplies the current source oracle.
`ui/compose-session-switcher.ts`, `ui/app-browser-events.ts` and
`ui/chat-swipe-navigation.ts` are byte-identical to the frozen 70d33bc source
(`frozen-to-oracle-sources.json`). A mounted installed-Classic fixture now
checks search, Escape dismissal, filtered keyboard selection and visible
post replacement against a disposable two-chat catalogue. The other rows
are source→Gi native-assertion→Gi-handler traces; the fixture does not
establish whole-feature or physical-input acceptance.

| Frozen ID | Gi clause evidence and handler | Boundary |
|---|---|---|
| `@ux-session-002` picker groups | `session.spec.mjs:169–208` creates current, pinned, active, tree, other and archived native sessions and checks group order, current marker, active archive-action exclusion and retained draft. Focused Gi run passed 6/6. Installed 3.2.4 mounted picker with six disposable entries in mixed source order also rendered Current, Pinned, Active, This session tree, Other sessions, Archived in that order, with one selected current row and a pressed Pin control on the pinned row; 6/6 Chromium/WebKit phone/tablet/desktop. | Classic's section headings use `role=presentation` and the current row uses `aria-selected`; Gi exposes named groups and `aria-current`. No assistive-technology equivalence was tested. Piclaw reads pin JIDs from `piclaw:session-picker-preferences:v1`, whereas the Gi fixture pins via its session API. The installed fixture does not test real parent graph, branch actions, live mutation/refresh, persistence after reload, or physical input. Group display only; no whole-clause parity. |
| `@ux-session-001` selected timeline | `tests/ux/session.spec.mjs:556–600` creates two native sessions with different completed messages, holds the old session's message response, selects research in the real picker and checks selection, correct timeline/draft and rejection of the late old read. The focused native journey passed 6/6. Installed Piclaw 3.2.4 selects a disposable research chat by Tab and replaces the main fixture post with the research fixture post in 6/6 mounted browser cases. | Installed fixture does not hold a superseded timeline read, use stored backend messages or test switching back. Its textarea still contains `main fixture draft` after selection and a 300 ms settle. After typing `research fixture draft` and switching back, the main post returns but the textarea still contains `research fixture draft` (6/6). Gi's native test switches to the research draft and restores the main draft on return. The installed fixture does not test draft reload, stored per-chat drafts or physical picker acceptance. |
| `@ux-session-003` search | `session.spec.mjs:405–452` searches identifier and model metadata, navigates a filtered list with keyboard and preserves drafts while selecting. Gi `compose-session-switcher.ts:112–129` matches metadata and `compose-box.ts` renders it. Installed Piclaw's mounted picker filters a disposable two-chat list for the research name and restores both on query reset; Home, End and ArrowDown retain the sole filtered selection and Tab activates it, changing the fixture URL to `web:research`. | Gi's tagged native search/arrow/Enter journey passed 6/6 separately. Its untagged Tab-focus journey passed 6/6: Tab moves to Pin and then the row without selecting, while installed Classic uses Tab to activate the selected row. Model metadata and physical IME were not compared in the installed probe. |
| `@ux-session-004` archive/restore | `session.spec.mjs:678–746` holds Gi's archive PATCH acknowledgement to reject optimistic success, verifies accepted archive/restore and catalogue refresh against native state, then aborts another PATCH and checks an error and unchanged state/draft. Focused native run passed 6/6. Installed 3.2.4 mounted archived-row Restore 6/6 sends one `POST /agent/branch-restore` with the fixture JID; while held, the current URL and draft stay in place; after success, the URL selects the restored fixture chat. | Classic's picker offers Prune for an eligible inactive branch and Restore for an archived row, but no Archive action. Gi offers Archive/Restore. The installed response is disposable; it does not verify real branch persistence, catalogue refresh, failure display, cross-device sync or physical input. Do not equate Prune with Archive or grant whole-clause parity. |
| `@ux-session-005` swipe eligibility | `session.spec.mjs:213–296` checks archived exclusion, active-first/chat-ID carousel order, selected text and control exclusions, native session/draft changes. Gi `web/src/ui/chat-swipe-navigation.ts` filters archived candidates and sorts active-first. | Playwright touch events in viewport projects are emulated; not a physical phone or tablet. |
| `@ux-session-006` dismissal | `session.spec.mjs:79–122` opens Gi's native picker, enters a query, presses Escape and checks popup dismissal, query reset on reopen, trigger focus, unchanged selected session/draft and history. Installed Piclaw 3.2.4 mounted picker independently filters a read-only two-chat list, dismisses on Escape, restores trigger focus asynchronously, clears the query on reopen and keeps the URL chat unchanged. | Installed fixture does not test a saved draft, backend mutation, typeahead state, outside-click or assistive technology. Gi and installed evidence are separate. |

`make test-ux-parity UX_PARITY_PORT=19134
UX_PARITY_ARGS='tests/ux/session.spec.mjs --grep "@ux-session-00[1-6]"'`
passed **36/36** in Chromium/WebKit phone/tablet/desktop projects with a
disposable Gi server. Passing tags alone did not establish the clauses; the
assertions and source above bound the six findings.
`make test-piclaw-session-picker-dismiss` passed **6/6** installed Piclaw
3.2.4 mounted Classic cases across Chromium/WebKit phone/tablet/desktop. It checks
search, Escape and single-result keyboard selection, including Tab activation.
Piclaw's fixture post replacement and shared textarea across a two-chat
round trip are observed. `make test-piclaw-session-groups` passed **6/6**
separate installed Piclaw 3.2.4 cases with read-only, mixed-order catalogue
entries and local pinned preference. Gi's focused `@ux-session-002` passed
**6/6**; focused `001`, `003` and untagged Tab-focus runs also passed **6/6**
each. The earlier focused `003/006` run passed **12/12**. Superseded reads,
draft reload, grouped roster actions, backend archive/restore, real sessions
and physical input still need separate comparison. The earlier inventory
passed 218 helper tests and 8,004 assertions without whole-clause parity credit.
