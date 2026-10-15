# Classic recovery and card rejection: bounded clause review

Current oracle: shipped Piclaw 3.2.4 asset `990f0c49a932`, inspected via its
`app.bundle.js.map`. These three Classic scenarios have native Gi browser
journeys. The rejected-card UI has a disposable installed-browser probe.
A second installed-browser comparison now exposes recovery-display differences;
see [recovery display](classic-recovery-display-review.md). All Gi tests below
used disposable session/fixture data.

| Frozen ID | Installed Piclaw source | Gi assertion → handler; remaining boundary |
|---|---|---|
| `@ux-extra-003` rejected card action | `make test-piclaw-card-rejection` passed 6/6 Chromium/WebKit viewport cases: installed UI activated a fixture `Action.Submit`, received a synthetic HTTP 503, showed its error notice without a success receipt, and retained the card answer and unsent composer draft. `ui/adaptive-card-renderer.ts:435–445` clears prior notice, marks the card busy, reports a rejected `onAction` as an error notice and removes busy state; installed `components/post.ts:1988–2009` awaits `Action.Submit`. | `tests/ux/card-rejection.spec.mjs:6+` activates a fixture card by mouse and Enter. It checks rejection notice (no success), retained card input and composer draft through reload, unchanged stored messages, no turn/write and no page error. Gi `web/src/ui/adaptive-card-renderer.ts:311–320` and `web/src/components/post.ts:972–995` own that path. Neither fixture proves a successful server/card submission or hardware assistive technology. |
| `@ux-extra-012` hidden recovery control | Installed browser probe across six viewports hides valid controls **and** two legacy-shaped blocks with partial primary-failure fields that Gi shows. Installed `components/post.ts:170+` accepts blocks without typed handoff keys before validating primary-failure fields. The probe did not exercise recovery execution. | `tests/ux/recovery-controls.spec.mjs:5+` checks validated control rows are absent while malformed lookalikes stay visible, then reloads/searches without mutating native messages or losing draft/media. Gi `web/src/gi-recovery-control.ts` rejects those partial primary-failure fields in its display-only guard; `scripts/patch-post-recovery-control.mjs` applies the guarded hide after hooks. Partial `primary_failure_*` values are invalid typed recovery metadata under the frozen clause; Gi keeps those rows visible, while installed Piclaw hides them. Do not relax the guard to match that installed edge. The mapped fixture journey does not earn installed parity credit. |
| `@ux-extra-013` silent informational recovery placeholder | Installed browser probe across six viewports hides the empty info row. A Classic-shaped follow-up 6/6 shows a media row when `media_ids` is supplied; it still hides file/message/attachment reference lists after extraction and resource/annotation-only blocks. Installed `components/post.ts:2022–2031` checks text, `media_ids`, cards and submissions; its predicate does not check those extracted extras. | `tests/ux/recovery-placeholders.spec.mjs:5+` checks empty info rows are hidden while prose, attachments, cards, submissions, references, resources and warnings remain visible through Gi's projection. It also verifies draft preservation, search, unchanged stored messages and no writes. Gi `web/src/gi-recovery-placeholder.ts` and the guarded Post patch implement a display-only rule. The fixture-shape mismatch bars whole-clause Piclaw acceptance. |

`make test-ux-steer` with `GI_UX_CARD_REJECTION=1
GI_UX_RECOVERY_CONTROLS=1 GI_UX_RECOVERY_PLACEHOLDERS=1`, these three specs
and `--grep "@ux-extra-00[3]|@ux-extra-01[23]"` passed **18/18** in six
Chromium/WebKit viewport projects. The installed rejected-card probe covers
one synthetic HTTP error. The installed recovery-display probe saw eight
consistent differences in each of six viewports; a focused Classic-shaped
probe resolves the missing-media-ID case but retains five visible-extra
boundaries and the two mixed legacy-control differences. No physical-device, live recovery execution or successful
card-submission assertion was run. These rows remain bounded fixture journeys,
not full feature acceptance.
