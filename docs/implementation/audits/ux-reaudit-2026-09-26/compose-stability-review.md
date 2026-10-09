# Classic compose-stability 001–006: clause audit

Shipped Piclaw 3.2.4 asset `990f0c49a932` provides the current source oracle
through `app.bundle.js.map` → `components/compose-box.ts:2470–2664`. This is a
source-to-assertion-to-Gi-code trace. The separate shipped-UI fixture directly
probed only the queued return's text/focus/removal request; it did not exercise
Piclaw media upload, failure or model backend. Tagged Gi tests run against a
**disposable native API**. This review does not claim physical-device or live
Piclaw backend acceptance.

| Frozen ID | Current oracle source | Gi assertion → code and limit |
|---|---|---|
| `@ux-compose-001` | Piclaw captures text, media, file/folder/message refs and `submissionChatJid` before clearing; the async task runs after the clear (`compose-box.ts:2532–2539,2575–2585`). | `tests/ux/drafts.spec.mjs:89–119` gates a real upload, asserts the editor and pills clear, types a new draft, checks persisted newer text/media, completed captured turn and captured references/media without newer text. Gi `compose-box.ts:1313–1387` captures and clears; `gi-drafts.ts` persists the pending recovery record. No Piclaw backend result is inferred. |
| `@ux-compose-002` | On submission error, `restoreDraft` merges captured text before newer text unless present, merges captured media/refs, and reports error (`compose-box.ts:2555–2573,2649–2656`). | `drafts.spec.mjs:121–151` aborts the real native prompt request after a newer draft and session switch. Asserts merged text/media/two message refs, untouched other session, alert, reload and no delivered turn. Gi `compose-box.ts:1370–1386`, `gi-drafts.ts:20+` use the captured pending record and merge. A pre-admission transport abort cannot prove lost-after-admission reconciliation. |
| `@ux-compose-003` | Piclaw returns without send for whitespace when there is no media or reference (`compose-box.ts:2504–2512`). | `drafts.spec.mjs:189–198` presses Enter on whitespace, checks unchanged input/stored draft, zero prompt POST and zero turns. Gi `compose-box.ts:1313–1337` checks the same emptiness boundary. Other empty-with-reference branches need their own acceptance. |
| `@ux-compose-004` | Piclaw queue-return replaces editor text, clears media and POSTs queue removal; `oracle-deltas.md` has the isolated current-UI probe. | **Accepted safety deviation:** ADR-0017 chooses Shared28's revisioned no-loss merge and durable recovery before DELETE. No `@ux-compose-004` Gi tagged test or Classic replacement credit. The Piclaw fixture did not test media or failure; its text replacement does not override the Gi safety contract. |
| `@ux-compose-005` | Piclaw `uploadFileBatch` reports progress, clears `uploadProgress` before `sendMessage`, while tracked submit state ends in `finally` (`compose-box.ts:2600–2609,2628–2632,2656–2661`). | `drafts.spec.mjs:643–673` holds real media upload then prompt separately. Asserts visible upload progress then distinct sending state, button ownership, per-session visibility, same file bytes/media ID, preserved newer draft/caret and reload cleanup. Gi `gi-compose-transfer.ts` scopes upload/send state. A route-held response alone does not certify all real network progress events; the adjacent untagged native-upload test checks those separately. |
| `@ux-compose-006` | Piclaw passes captured `submissionChatJid` to `sendMessage` after upload (`compose-box.ts:2538,2631`). | `drafts.spec.mjs:200–216` holds native media upload, switches to a child, then checks completed main turn, empty child turns/media and retained child draft. Gi `compose-box.ts:1336–1350,1414–1454` captures chat and sends through its owned origin. Exact user/assistant messages are checked in separate first-send journeys, not this tag. |

Local gate: `make test-ux-parity UX_PARITY_PORT=19134
UX_PARITY_ARGS='tests/ux/drafts.spec.mjs --grep "@ux-compose-00[12356]"'`
passed **30/30** on an unchanged second run. The first run passed 29 and timed
out on WebKit phone `@ux-compose-001`: after `page.reload()`, the page showed
only `Loading Gi…` and `.post-time` never appeared. The second run did not
change code, fixtures, assertion or timeout; the original trace was later
overwritten and the loading-only cause remains unknown. A separate bounded
WebKit-phone repeat (four passes, one failure) hit an internal `page.reload()`
navigation error before compose actions; see `loading-stall-investigation.md`.
Its preserved trace cannot establish the earlier `Loading Gi…` cause. The five
tagged native cases bound the test evidence above, not physical-device
acceptance. `@ux-compose-004` remains a source/UI conflict with Gi's deliberate
no-loss policy.
