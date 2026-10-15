# Classic mobile swipe 001–006: clause review

The installed Piclaw 3.2.4 `ui/chat-swipe-navigation.ts` is byte-identical to
frozen 70d33bc (`frozen-to-oracle-sources.json`). The source checks target
eligibility, text selection, direction, candidate ordering and Safari wheel
gates. A mounted installed-Classic bundle probe now exercises a narrow
text-selection gate with a disposable two-chat roster. Playwright dispatches
synthetic touch/wheel events against disposable Gi sessions. Neither is a
physical-touch run or whole-feature acceptance.

| Frozen ID | Gi assertion → handler | Limit |
|---|---|---|
| `@ux-mobile-001` | `tests/ux/session.spec.mjs:297–331` swipes eligible timeline space, checks adjacent session and wrap, and restores per-session drafts. Gi `web/src/ui/chat-swipe-navigation.ts` handles candidate selection and threshold. Installed Classic probe additionally confirms an eligible swipe with whitespace-only selected text chooses the adjacent fixture chat. | Browser-dispatched touch and synthetic selection, not physical iOS/Android or a real Piclaw session change. |
| **`@ux-mobile-002`** (seven outline examples) | Two tagged cases in `tests/ux/mobile-exclusions.spec.mjs` run against a disposable mounted Gi page. They gesture on the real composer, `.workspace-sidebar`, read-only preview, stored-image attachment preview, Adaptive Card input/Submit and model popup; each keeps the selected session, with an eligible timeline swipe as positive control. `web/src/app.ts:858–878` gates starts to the timeline/status host; `chat-swipe-navigation.ts:59–105` adds target exclusions. | **Partial six-surface native evidence:** both cases passed in all six browser/viewport projects (**12/12**), using simulated iPhone user agent and synthetic touch events. The terminal/dock example cannot run: mounted `web/src/app.ts` has no terminal dock and supplies empty terminal/VNC callbacks. This is not seven-example parity, physical-device acceptance or a full Piclaw oracle comparison. |
| `@ux-mobile-003` | `tests/ux/status-swipes.spec.mjs:16+` checks real streamed draft/status panel link gestures, selected-text and direction guards, excluded input/settings controls, pen-contact guard and actual timeline navigation. Gi `chat-swipe-navigation.ts:89–105` permits designated thinking/status ancestors. Installed Classic probe separately confirms nonblank selection blocks a timeline gesture. | The Classic probe does not mount a status panel; the separate thought variant `@gi-swipe-003` is not a Piclaw or physical touch pass. |
| `@ux-mobile-004` | `tests/ux/session.spec.mjs:770–846` mixes active, pinned, ordinary and archived sessions, checks deduplicated active-first/chat-JID order despite selection/unpinning, and keeps drafts. Gi `chat-swipe-navigation.ts:106+` filters archived and sorts. | Does not test an arbitrarily changing remote roster on a live peer. |
| `@ux-mobile-005` | `session.spec.mjs:332–365` first checks predominantly vertical gesture cancellation, then performs a fresh eligible horizontal swipe as a positive control. Gi `chat-swipe-navigation.ts:177+` gates direction/threshold. | Browser event emulation, not OS gesture arbitration. |
| `@ux-mobile-006` | `session.spec.mjs:366–404` checks no wheel navigation under Chrome/iOS mode and positive desktop-Safari path with draft ownership. Gi `chat-swipe-navigation.ts` gates browser and axis. | Emulated browser mode and wheel events do not establish physical trackpad behaviour. |

`make test-ux-parity UX_PARITY_PORT=19134
UX_PARITY_ARGS='tests/ux/session.spec.mjs --grep "@ux-mobile-00[1456]"'`
passed **24/24**. `make test-ux-status-swipes` passed **12/12**, six of which
carry `@ux-mobile-003`; the other six are Gi thought-panel variants.
`make test-ux-mobile-exclusions` passed **12/12** across six projects with the
isolated card fixture on retry. The preceding run passed all six mounted-control
cases but failed all six attachment-preview cases at the Escape dismissal;
the test now uses the modal's visible Close button. An earlier WebKit-phone
test mistake queried the composer while a read-only preview hid it; that test
now checks selection in preview mode and draft retention on return. The
attachment case uses the attachment pill's `Preview` button; clicking a message
image opens a separate `ImageModal`. Neither test correction changed production
code, frozen Gherkin or timeout.
`make test-piclaw-swipe-selection` passed **4/4** installed Classic
Chromium/WebKit phone/tablet cases: nonblank selection blocks navigation;
whitespace-only selected text permits a synthetic timeline swipe to
`web:other`. Unlike the installed helper's trimmed selection check, Gi's host
capture guard treated whitespace-only selection as blocking. Gi
`web/src/app.ts` now trims before rejecting; the mounted Gi test passed
**6/6** across Chromium/WebKit phone/tablet/desktop, with a retained draft
on the blocked path and a cleared composer after switching. The installed
fixture's navigation target is a disposable HTML page, not a second mounted
Piclaw chat; Gi uses native temporary sessions. The fix is limited to the
selection predicate; it does not claim the missing terminal/dock example,
whole-clause or physical-device acceptance. Selector-only synthetic unmounted
DOM is not a substitute.
