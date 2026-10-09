# Current Piclaw oracle versus pinned Gherkin

Current oracle: installed `piclaw-3.2.4-linux-x64-baseline`
(`/opt/piclaw/current`), asset version `990f0c49a932`, source map SHA-256
`c53092cfb75415f30b4c6b16b758cab68fed9bbc546953e5916a91b382acb32b`.
The release manifest names no source commit. Source excerpts below come from
its own `app.bundle.js.map` `sourcesContent`, not the modified local Piclaw
checkout. The four relevant compose, Quick Actions, Markdown and SVG source
files have identical hashes to the earlier shipped 3.2.3 map. Comparing the
234 extracted web modules against the pinned `70d33bc` tree finds 166 identical,
50 changed and 18 new modules. This is source-drift triage, not behaviour credit;
see `frozen-to-oracle-sources.json`. The 3.2.3 release
was removed after the initial probe; its versioned reference is retained only as
historical evidence. The older frozen Classic corpus was cut at `70d33bc93ab...`
(`tests/ux/upstream/PROVENANCE.txt`). The shipped 3.2.2 pixel reference was
`0afe5366...`, not an acceptance claim for 3.2.4.

`make test-piclaw-oracle-matrix` serves byte-hash-checked 3.2.4 UI assets across
Chromium/WebKit phone, tablet and desktop viewports, with fixture-only
`web:default` responses and no real chat/auth writes. All six projects passed. It reproduces
the server's normal `sanitizeSvgFences=true` HTML substitution from
`runtime/src/channels/web/http/static.ts:107` and the default in
`runtime/src/core/config-web.ts:293`. Evidence:
`test-results/ux-oracle/piclaw-3.2.4/{browser}-{viewport}/browser-probe.json`
and screenshots.
It does not exercise Piclaw's live DB, production authorization, physical touch,
passkeys, keyboard assistive technology or Gi's runtime.

| Contract | Piclaw 3.2.4 observation and source | Existing Gherkin/test conflict | Resolution for this audit |
|---|---|---|---|
| Quick Actions slash prefill | Clicking `/model` replaces `ORACLE EXISTING DRAFT` with exactly `/model`, focuses textarea, sends no prompt. Shipped `components/timeline-quick-actions.ts:365`, `app.ts:126-130`, `components/compose-box.ts:1446-1465`. The older `70d33bc` source also forwards `item.commandName` directly; the trailing-space clause appears unsupported even at its source revision. | Frozen Classic `@ux-original-007` requires replacement **with a trailing space**; `tests/ux/quick-actions.spec.mjs:40-51` asserts `/model `. Shared16/17 require **preserving** the previous draft. Mounted Gi `/model` prefill also replaces existing text; only the canonical `/skill:<name>` build adapter preserves it. | Versioned oracle-only scenario records actual `/model` replacement. Do not flip Gi's durable draft handling just to earn parity: replacing user text is destructive and needs an explicit policy decision, a migration story and regression gates. No new Classic/Shared credit. |
| Fixture-listed skill Quick Action | The isolated catalogue lists `/skill:proof`; selecting it in the Slash commands group replaces `ORACLE SKILL DRAFT` with exactly `/skill:proof`, focuses the composer and sends no new prompt in six Chromium/WebKit viewport projects. The shipped `timeline-quick-actions.ts` dispatches a slash command through the same compose prefill path as `/model`. | Classic `@ux-original-008` calls for the same replacement path; Gi `tests/ux/skills.spec.mjs:4-47` instead preserves the draft as Shared17 requires. | Add a versioned oracle-only skill-prefill scenario. The fixture does **not** prove a skill was actually loaded, discovered or executed. The built Gi skill adapter prepends `/skill:proof ` while preserving existing text. Keep this as a bounded Shared17 safety difference; the Piclaw UI probe alone does not establish loaded-skill execution or Gi general-command parity. |
| Model-picker keyboard | With twelve fixture-listed matches, shipped Piclaw 3.2.4 keeps search focus while Page keys change the active result and Control/Meta+Home/End choose the first/last. Escape closes and returns focus to the trigger in six browser/viewport projects. The first PageDown may start from a still-active current model outside the filtered list, so its initial destination is timing-dependent in the probe. `components/model-picker.ts:75-105` and `ui/model-catalogue.ts:697-715` clamp arrows and page by seven. | Classic021's old Gi tag checked stale model responses, and Gi's model helper reused the session picker's wrapping eight-row rule. | The Gi-owned helper and tagged browser journey now use clamped seven-row model navigation and modified Home/End from a focused search field. Six native model journeys, six Shared34 disabled-entry journeys and 42 model-panel regressions passed locally. Fixture navigation does not prove Piclaw model mutation or the frozen session-picker half; CI/review and deployment remain separate. |
| Return queued item | Clicking Edit in compose replaces `NEWER UNSENT TEXT` with `QUEUED ORACLE TEXT`, focuses textarea and POSTs `/agent/queue-remove` with `row_id:101,chat_jid:web:default`. Shipped `components/compose-box.ts:929-985`. | Classic `@ux-original-017` and `@ux-compose-004` replace/clear media, while Shared28 requires merging latest concurrent draft/media before DELETE; Gi's transactional recovery test `tests/ux/queue-return.spec.mjs:26` follows Shared28. | Keep both pinned reference and safety contract visible. ADR-0017 accepts the Gi no-loss recovery path before DELETE; Classic replacement stays unmapped. The oracle fixture does not establish media/failure semantics. Changing the accepted safety policy would need explicit approval. |
| Fenced SVG | Safe SVG renders as `<img class="model-svg-image" src="data:image/svg+xml;base64,..." alt="Oracle SVG">`; source remains in disclosure for copy. A fence with script, `onload` and external image URL stays escaped code, with **no preview, script execution or external request**. Shipped `markdown.ts:684-688`, `utils/svg-images.ts:74-225`; config default noted above. Raw model geometry is not an inline privileged SVG DOM child. | Frozen Classic `@ux-original-029` and `tests/ux/rendering.spec.mjs:54-61` require source-only; Shared41 asks for inline SVG and stripping unsafe elements before rendering. At `70d33bc` `markdown.ts` had no `renderSvgFences`; `f8bfbd25e` later added the safe image path. | Versioned oracle scenario states safe-image plus unsafe-source fallback. Neither frozen scenario can be claimed from the other. A full SVG security suite needs malformed/oversize/external-resource/theme cases before Shared41 credit. |

## Other source-backed re-audit leads

- Classic timeline `018`/`019` expect direct deletion to fail with `Replies exist`
  when a stored reply is outside the loaded view. Pinned Piclaw 3.2.4
  `deletePostResponse` instead returns 200 for a direct parent deletion.
  `tests/ux/oracle/piclaw-deletion-combined-probe.mjs` connects shipped Classic
  assets to installed backend functions through disposable HTTP fixtures and
  an in-memory database. Chromium and WebKit desktop both sent `cascade=false`
  without prompting, removed the parent from the UI and left the stored reply
  orphaned. The joined fixture also confirms a visible-reply cascade removing
  all four stored rows, or cancels without deletion. Separate synthetic-409 UI
  fixtures exercise the dormant hidden-reply retry branch. See
  `message-deletion-review.md`. The production Piclaw HTTP router,
  authentication and live store were not exercised. Gi `018`–`022` remain
  unmapped because Gi has no reply graph or cascade confirmation.
- Installed Piclaw 3.2.4 theme/tint parser checks passed nine exact command
  cases. A shipped-Classic UI fixture in Chromium/WebKit desktop sent
  `/theme ristretto` and `/tint #e11d48` through the composer, applied
  fixture `ui_theme` events to root attributes and localStorage, and displayed
  a tint response in the fixture timeline (`theme-command-gap.md`). The
  installed server handler, persistence, reload and pixel colours were not
  exercised; Gi has no `/theme` or `/tint` composer-command path, leaving all
  fifteen Classic command scenarios unmapped.
- Installed Piclaw 3.2.4 with Plan Sidebar add-on 0.1.25 preserves dirty
  Markdown on a same-chat `plan.changes` event, then explicit Refresh replaces
  it without confirmation. A shipped-Classic/add-on disposable browser probe
  saw the warning, one refresh GET and no write in Chromium and WebKit desktop
  (`plan-oracle-gap.md`). Gi has no Plan feature. Frozen Shared20's
  confirm-before-discard requirement is a proposed no-loss deviation from this
  observed oracle behavior and needs an explicit policy decision before
  implementation; no Shared Plan parity credit follows from this probe.
- `tests/features/settings/gi-settings.feature:40` said thinking was read-only;
  `web/src/gi-settings-models.ts:117` and `tests/ux/session-thinking.spec.mjs:18-35`
  expose/apply native supported choices. The additive Gi feature is corrected with
  explicit choice and model-switch reset clauses; the six-project focused test
  for `@gi-settings-029` passed. This does **not** alter frozen Classic/shared files.
- `features/search/workspace-index.feature:33` asks a default background refresh,
  whereas lines `199,211` require GET to stay read-only. Pinned Piclaw 3.2.4
  `searchWorkspace()` with isolated in-memory SQLite returned a committed hit
  and requested background refresh for cold/stale notes scopes; ready and blank
  searches requested none. The requester was intercepted before child launch
  (`piclaw-workspace-search-trigger-probe.ts`). Gi
  `internal/web/workspace_index.go:138-141` rejects `GET ?refresh=...` and does
  not request background work. This is a policy conflict across native search
  and web contracts; worker completion and production HTTP are untested.
  Do not silently make GET synchronously mutating or weaken transactional-index
  safeguards.
- A six-project Playwright matrix uses viewport sizes, not physical touch, iPad,
  hardware passkey or actual reduced-motion preference by default
  (`playwright.ux.config.mjs:3-10`). Individual `test.use({hasTouch:true})`
  scenarios can prove emulated trusted touch; they still cannot prove physical
  iPad/OS behaviour. Presence of a tag alone is not physical proof.
- A current audit found three functional false positives: the HUD test accepted
  zero elements (`tests/functional/05-system-meters.spec.ts:56-68`), the old
  "sending triggers SSE" test polled only turn completion
  (`04-sse-and-streaming.spec.ts:47-71`), and frontend logging only checked
  `typeof window.onerror` (`09-frontend-logging.spec.ts:43-46`). These tests now
  assert visible meters, exact session/turn SSE frames and a delivered uncaught
  browser-error log. Focused tests passed, then the 139-pass functional suite
  (11 existing skips) passed. No product code changed for these corrections.

The full 358-definition/421-case audit is still underway. These findings are
verified **slices**, not an exhaustive result or a full parity report.
