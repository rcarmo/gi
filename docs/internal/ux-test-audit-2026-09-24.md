# Gi UX test audit — 2026-09-24

## Verdict

The test suite does **not** establish that Gi has Piclaw's overall interaction or visual parity. It establishes many useful, bounded backend/UI invariants. The previous summary counts obscured that distinction.

The user's reports—Return-to-start, misaligned pickers, wrong workspace-tab behaviour and an unlike composer—are credible gaps in the acceptance process. The picker and composer differences are directly observed. The exact Return-to-start journey and intended workspace-tab transition still need a faithful reproduction; this review does not mark them fixed.

**No product fixes or new parity credit result from this audit.** Auth persistence commit `56079fe` is paused and not deployed: native macOS/Windows jobs in CI `36063764465` failed. Live Gi remains the TOTP slice `c96ecd4` (deployment record `93f9f4a`), port 8090, PID 2264426.

## Scope and method

Full source review, divided into bounded independent reviews and reconciled against the actual repository inventory:

| Layer | Files | What was reviewed |
|---|---:|---|
| Browser UX | 39 | Every `tests/ux/*.spec.mjs` |
| Functional browser/API | 17 | Every `tests/functional/*.spec.ts` |
| UX helper/contract tests | 47 | Every `tests/ux/support/*.test.*` |
| TUI unit tests | 17 | Every `internal/tui/*_test.go` |
| Terminal acceptance/capture scripts | 20 | All 16 PTY `.mjs` tests, two shell tests and two capture/report scripts |
| Gherkin contracts | 37 | 24 frozen Classic, shared, three Gi-derived browser, additive passkey, seven TUI and workspace-index proposal |
| Runner and harness | separately inventoried | Makefile, CI/config/report/helper code, Go fixture servers, shell/provider boundaries and pane-host fixture |

The companion `ux-test-audit-2026-09-24.csv` names every reviewed file and its evidence boundary. JSON/data fixtures were inspected as fixtures, not counted as executable tests. This was **not a fresh execution of all suites**. Earlier passing runs retain their narrow meaning; they are not proof of current Piclaw parity.

Read-only Chromium comparisons used the deployed Gi on 8090 and running Piclaw on 8080 at widths 390, 820 and 1440, height 900. All non-read HTTP requests were blocked, including attempted submissions and Piclaw presence/visibility posts. Existing sessions and host settings were left intact. Probes captured screenshots, DOM geometry, style properties and focus state. Production history and model catalogues differ, so screenshot differences are evidence to inspect, not a pixel-diff oracle. The local Piclaw source reference `origin/main` was `0afe5366ce`; the running bundle advertised `15958f2c3dc9`. Do not assume they are the same build.

## Ranked findings

### 1. Release checks do not exercise the browser acceptance suite — critical

- `.github/workflows/ci.yml` runs Go tests, shell cancellation/steering checks, vet, builds and the newly added native auth-state tests. It runs **no Playwright suites** and no Bun UX-support suite.
- `Makefile` target `check` invokes `test-ux`, whose config explicitly excludes `tests/ux/**`. It does not mean all UX tests ran.
- `playwright.ux.config.mjs` chooses one specialised suite by environment or a fallback list. Several fallback files skip without their seeding flags.
- Provider Settings has no named Make target. `GI_UX_SETTINGS_CATALOGUE` and `GI_UX_MODEL_PICKER` tests in `context-fit.spec.mjs` also require manual flag combinations.

**Consequence:** green CI does not catch broken Return, misplaced menus, stale component/CSS combinations or unlike Settings/composer layouts.

### 2. The comparison target and acceptance oracle are wrong for overall parity — critical

The suite pins an older feature/component baseline, then often checks Gi's implementation against itself. It does not compare the rendered composer or picker layout with current Piclaw. `page.screenshot()` saves evidence, but there are no reviewed `toHaveScreenshot`/image-difference assertions establishing reference equivalence.

Direct observations:

- Desktop composer side padding: Gi **16px**, running Piclaw **0px**. This is not attributable merely to different model text or history.
- Gi renders the older pair of footer session controls. Running Piclaw has a different session strip/recent-session presentation.
- Session popup at width 1440: Gi **440px** wide; Piclaw **878px**, spanning its composer. At width 820: Gi **440px**; Piclaw **716px**.
- At width 390: Gi session popup remains an absolute anchored panel (**358px**); Piclaw uses a fixed nearly full-viewport panel (**374×884**, inset 8px).
- Desktop model popup widths: Gi **846px**, Piclaw **680px**. At width 390 Piclaw's model popup is also fixed/full-viewport; Gi's is anchored above the input. Heights are catalogue-dependent and are not treated as visual defects by themselves.

`tests/ux/session.spec.mjs:919–940` actively asserts the old absolute-positioned, six-pixel-gap geometry. It can pass while reproducing precisely the layout the user dislikes. `web/src/components/compose-box.ts`, `internal/web/static/css/chat.css:854–875`, and `responsive.css:36` make the baseline mismatch concrete.

**Required change:** establish one explicit current Piclaw reference for component, host markup, CSS and responsive behaviour together. Preserve the frozen feature files as historical contracts, but do not equate those mappings with current visual parity. Do not overwrite supplied component files ad hoc to chase a screenshot.

### 3. First-chat and keyboard-entry journeys are not proved — high

Most detailed specs create sessions via the API and seed `gi_session_id` before loading the page. That is appropriate for isolating queue/draft races, but it bypasses the startup journey.

- `tests/functional/helpers.ts:26–35` selects the first matching textarea, fills it and presses Enter. It does not capture the resulting session/turn ID.
- `03-chat-flow.spec.ts:21–34` waits for a count of posts; later tests check cleared input. A pre-existing/unrelated post can satisfy loose count checks.
- `02-config-and-session.spec.ts:37–60` proves some session exists, not that this page created the expected fresh session.
- `session.spec.mjs:750–761,1008–1014` verifies child creation/selection/draft state, but not the complete create-child → focused composer → Return → exactly one first turn journey.

Live findings, carefully scoped:

- Inside an established Gi composer, Enter attempted the expected prompt POST (blocked by the probe). This is **not** a successful-send or fresh-chat proof.
- Enter from the conversation did not focus the composer in either tested running UI. The user's exact Return-to-start failure remains open, not disproved.
- Escape from the conversation focused Piclaw's composer, but not Gi's.
- `/` from the conversation left Gi's composer empty and opened Quick Actions; running Piclaw also changed the composer to `/`. These competing effects need an explicit shortcut-ownership contract, not a blind port.
- A held-bootstrap exploratory probe found no visible composer before readiness. That timeout is not evidence of a submit defect and is excluded from pass/fail claims.

### 4. Slash commands and Quick Actions have an incomplete interaction contract — high

There is good palette filtering, focus-return and dismissal evidence. It does not substitute for direct composer slash-menu behaviour.

Missing integrated cases include: typing `/` in the textarea, argument-aware Enter versus Tab, Escape with retained draft, focus return, IME, pending picker ownership, and a single gesture producing exactly one command/send. Direct `/model` execution and Quick Actions prefill tests cover only parts of this.

**Questionable mapped credit:** frozen Classic `@ux-original-007` requires replacement of composer text; `@ux-original-008` says skills use that same prefill path. `tests/ux/skills.spec.mjs:5–12` runs one body under both Classic 008 and shared 17 and expects `/skill:proof existing request β`. This proves the shared draft-preserving behaviour, but does not justify the Classic replacement-path claim. Mark Classic 008 disputed pending criterion adjudication; do not silently alter the frozen source or implement destructive replacement just to satisfy a tag.

### 5. Workspace tests validate Gi's adaptation, not the complete Piclaw workflow — high

The current host deliberately offers **Open read-only tab**, not Piclaw's editable surface. `workspace-tabs.spec.mjs` asserts that label and tests the local lifecycle. That is useful evidence for the adaptation, not proof that the workspace tab works like Piclaw.

- `web/src/app.ts:1020–1044` creates its own workspace toggle and read-only editor region; narrow layouts hide the conversation via `gi-workspace-tab.css`.
- Opening/closing the workspace drawer, activating a file tab, returning to conversation and retaining keyboard focus need to be compared as one actual user journey. The user's “wrong way” report is not resolved by current MRU/pinning tests.
- Tabs lack trusted touch coverage despite claims about mobile affordances.
- Context-menu edge clamp in `workspace-tabs.spec.mjs:102–105` uses synthetic `contextmenu` coordinates; ordinary right-click coverage elsewhere does not make that specific edge test trusted.
- Tab-strip Arrow/Home/End activation and close-context focus are incomplete.

The existing frozen mapping correctly leaves most editor/dock/CRUD criteria unmapped. Keep those limitations explicit; do not present read-only adaptation as complete workspace parity.

### 6. Functional smoke tests contain false-positive assertions — high

| Location | Problem |
|---|---|
| `05-system-meters.spec.ts:56–65` | `count >= 0` always passes; HUD presence is not proved. |
| `04-sse-and-streaming.spec.ts:27–63` | One test never inspects its event collection; another polls turn completion rather than observing SSE. |
| `08-turn-lifecycle.spec.ts:12–21,63–89` | Uses first session/latest turn instead of captured submission identity; header promises queue/cancel without exercising them. |
| `helpers.ts:48–52` | Searches all sessions by message substring; fixed repeated prompts can identify the wrong session. |
| `07-compose-interaction.spec.ts:12–21` | Shift+Enter assertion checks remaining text rather than exact newline and zero admissions. |
| `09-frontend-logging.spec.ts:43–46` | Presence of `window.onerror` is not proof that browser errors reach the logging endpoint. |

These are smoke/API tests and should be labelled as such or strengthened. Their passing totals should not be bundled into interaction-parity claims.

### 7. Settings is better tested than the smoke suite, but not complete — high

Strong evidence exists for modal isolation, lazy-module failure, session-owned model mutations, stale replies, dirty-form protection and saved values. The six browser projects **do** supply phone/tablet/desktop coverage; a reviewer suggestion that these tests were single-viewport was rejected.

Remaining test-level gaps:

- Close-button and backdrop dismissal do not consistently assert final focus as rigorously as Escape.
- `@gi-settings-008` unauthenticated runtime/model policy has no dedicated browser mapping. Native security tests must remain separate evidence.
- Identity length/control-character rejection, restart activation and compaction-policy bounds are not all exercised through the browser; native tests cover some of these, so they are not globally “untested”.
- Models overflow/refine and richer picker navigation cases exist under manual flags but are not in a dependable all-suite runner.
- Provider key inference consumption is mostly driven by API prompts, not a subsequent real composer send. Useful backend proof, incomplete full user journey.
- Appearance tests verify stored values and selected style properties, not overall compose/Settings visual equivalence to current Piclaw.

### 8. Coverage accounting cannot validate the meaning of the tests — medium

`scripts/ux-parity-report.mjs:13–34` trusts supplied Playwright-format JSON and tags in spec titles. It validates result counts/projects, not whether all Gherkin steps were asserted. Synthetic JSON in report unit tests is legitimate testing of the reporter, **not evidence that production reports were fabricated**. The weakness is absent provenance/semantic enforcement.

`tests/ux/README.md` mixes old milestone counts with present-tense statements, including obsolete “no shared mappings”. Historical `40 unmapped` meant 42 minus 2, not a denominator bug. The required correction is clear dating/status, not rewriting history.

Current source mapping is 101 Classic / 30 shared; that is a **mapping count**, not current overall UX parity. Classic 008 is disputed above. The 26 additive passkey scenarios remain explicitly unsupported despite retained upstream `@implemented` tags.

### 9. TUI evidence is substantial but has its own limits — medium

Strong PTY scripts cover actual draft/cursor/session/media/clipboard bytes and regular-mode history at the required sizes. They should not be discarded because unit tests use helper state.

Gaps found across the entire terminal layer:

- `capture-tui-audit.sh` and `render-tui-audit-report.mjs` generate presentation from captured text; their HTML screenshots are not terminal screenshot comparison oracles.
- Smoke/Gherkin scripts are mostly fixed-size string/DB-count checks.
- Several older PTY scripts use the default tmux server, risking interference; newer ones use isolated sockets.
- Trailing-whitespace trimming can conceal bottom blank-row growth; separator/footprint counts are a proxy, not a full-buffer layout proof.
- Unit picker/modal tests often inspect lists, widths or helper line counts rather than composed selected/disabled/error styling.
- PTYs cover some live resize behaviour missing from unit tests. The audit does **not** claim resize is wholly untested.
- OSC52/OSC8 emission proves terminal protocol bytes, not physical host clipboard acceptance or all terminal-emulator behaviour.

## Evidence worth retaining

The strongest browser tests use native routes with deterministic providers and exact identities/bytes: upload cancellation/retry and durable media reuse, draft restoration, run-bound queue steering, stale SSE reconciliation, bounded timeline pages, compaction ownership and authentication. Seeded fixtures and controlled OS speech are valid when explicitly scoped. Missing visual or startup coverage does not invalidate those narrow invariants.

## Repair order and acceptance gate

1. Add a small **user-journey suite** that runs in CI: clean boot, new chat, first Return submit, exact captured session/turn, Shift+Enter no-send, pending/failure/retry.
2. Compare/pin **composer and both pickers together** against the current reference: DOM/CSS provenance, responsive structure, alignment, controls, focus and keyboard ownership. Normalise theme/font/catalogue/content before visual baselines.
3. Add direct slash-menu and Quick Actions integration with no unintended submit, retained drafts, IME and modal exclusion. Resolve Classic/shared skill prefill conflict explicitly.
4. Reproduce workspace toggle/file-tab/conversation transitions with real pointer, keyboard and touch. Compare to reference before preserving another Gi-specific assertion.
5. Complete Settings focus/negative-path journeys and wire provider/catalogue flags into named targets.
6. Replace vacuous functional assertions with exact event/request/identity checks; classify API/helper/PTY/manual evidence separately.
7. Add an explicit full-suite manifest/runner with required flags, per-suite result paths, skip accounting, revision/build identity and no silent exclusions. CI smoke must not be labelled the full matrix.
8. Add visual/geometry oracles and full terminal buffer-footprint assertions. Preserve protocol/device caveats and manual-device gaps.

Every fix must first reproduce the defect under a user-level assertion. Passing an assertion copied from the current implementation is not sufficient. No new aggregate parity claim until these basic journeys and the reference comparison are green.
