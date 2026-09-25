# Piclaw web interaction parity

> **Audit warning (2026-09-24):** the historical passing totals below do not prove current Piclaw visual or interaction parity. See [the full UX test audit](../../docs/internal/ux-test-audit-2026-09-24.md) and its [per-file inventory](../../docs/internal/ux-test-audit-2026-09-24.csv). CI runs no Playwright suites, `make check` excludes this parity runner, and some specialised cases need manual environment flags. The current source maps 101 Classic / 30 shared IDs, with Classic 008's skill-prefill claim disputed. Composer/new-chat keyboard paths, picker/reference geometry and workspace transitions require revalidation before broader parity claims.

Gi vendors the same frozen Classic Gherkin baseline used for the Tau/Vibes audit:

- Piclaw commit `70d33bc93ab540845bbcf5f80503ca8125c71594`.
- `features/classic/`: 24 byte-identical feature files, 236 unique scenario IDs, 256 expanded cases.
- `features/shared-canonical-ux.feature`: byte-identical shared Tau/Vibes interaction contract, 42 expanded cases; SHA-256 `a08a623880c6f327bc051edc51bb2bbff2959aed86421b5227e61d5a92fc2441`.
- Upstream file hashes and provenance are preserved in `upstream/`.

The sources came from `/workspace/evidence/piclaw-classic-70d33bc93` and `/workspace/tau/tests/ux/features/canonical-ux.feature`. Tests do not depend on those paths after copying. Source-relative links inside frozen features refer to the original Piclaw checkout.

The Classic and shared contracts differ in places, including command-prefill and idle-Steer safety semantics. Both are preserved. Passing a Classic case does not imply passing the stronger shared contract.

## Current status (2026-09-25)

The [feature and parity matrix](../../docs/feature-parity.md) separates shipped
behaviour, known gaps and planned integrations. Source mappings are 101/236
Classic IDs and 30/42 shared cases; they are not a full-suite pass. Classic008 is
disputed. All 26 separately pinned passkey Settings scenarios/outlines are unmapped;
`make test-ux-passkeys` verifies native APIs and Settings/login journeys using real
Chromium WebAuthn and virtual authenticators. Full per-case mapping, Visual-skin
and physical-device evidence are outstanding.
Workspace-collapse motion and native auth persistence repairs have shipped.

The dated sections below retain historical run totals. Do not add them together
or interpret an old "unmapped" statement as the current inventory. The isolated
passkey suite and startup/Return journeys now gate CI builds; the complete browser
UX suite still does not.
Specialised fixture suites need separate targets/flags.

## Run

```sh
make ux-parity-inventory
make test-ux-parity
make test-ux-parity UX_PARITY_ARGS='--project=chromium-desktop'
```

The parity target uses `.gi-ux-parity/` and loopback port 19091, seeds deterministic `test-model` configuration, and removes its isolated process/database after the run. It does not touch the developer's sessions. Install browsers with `scripts/run-playwright.sh install chromium webkit` if necessary.

`@cucumber/gherkin` parses the original features and expands examples. `classic.spec.mjs` supplies Gi-native browser steps for mapped IDs, attaches the original steps to each result, and drives visible enabled controls. Native APIs only seed fixtures and verify persistence; they do not substitute for tested user actions. No DOM injection, forced clicks, automatic resubmission, navigation fallback, retries, or inherited Tau/Vibes pass statuses.

`support/catalogue.mjs` rejects modified source hashes. The report always inventories every frozen scenario. A scenario passes only when all expanded cases pass in Chromium and WebKit at 390×844, 820×1180, and 1440×900. Unmapped scenarios are reported as `unmapped`, not skipped or passed. The shared contract is inventoried separately; 30 of its 42 cases currently have source mappings.

## Startup/Return CI gate — 2026-09-25

`make test-ux-journey` runs seven untagged user journeys (42 executions) in
Chromium/WebKit at all three configured sizes. Each case starts its own native
process and empty database; the page creates its first session without API or
localStorage seeding. It uses deterministic local inference with production
HTTP/SSE/admission/storage. `GI_UX_JOURNEY=1` selects only this suite and writes
`test-results/ux-parity/journey-results.json`.

The journeys cover first Return, exact session/turn and message identity,
Shift+Enter, new-chat focus, parent draft retention, held/rejected admissions,
explicit startup retry, lost create replies, partial bootstrap success and
Settings focus during delayed fork completion. The host focus/retry repairs and
limits are in [startup-return-journeys.md](../../docs/internal/startup-return-journeys.md).
The dedicated CI job requires both browsers and gates Linux/macOS builds; no new
Classic/shared mapping is assigned by these tests. This is a bounded journey gate,
not a full browser or visual-parity result.

## Pinned picker geometry — 2026-09-25

`make test-ux-picker-geometry` adds24 executions to the required browser CI job.
It pins current Classic CSS provenance and observed bounds separately from the
frozen feature files. Tests cover composer padding, responsive session/model
picker bounds,639/640 transitions, long session-list scrolling, touch/keyboard
Close, Escape and retained focus/query/draft. Half the executions explicitly use
390px touch contexts, repeated across projects.

The reference model panel overflows at640px; Gi caps it to the anchor as an
explicit containment correction. Shared23/24 now use fixed-mobile geometry while
retaining desktop anchor and interaction assertions. Tests that switch sessions
from a mobile model picker dismiss the covering panel first. No mappings are
added. See [picker-geometry.md](../../docs/internal/picker-geometry.md) for source
hashes, measurements and remaining structural/visual gaps.

## Direct slash ownership — 2026-09-25

`make test-ux-slash` runs72 Chromium/WebKit cases from fresh native fixtures and
is required in the browser CI job. The composer now reads Gi's native catalogue,
without the unserved Piclaw endpoint or56-command fallback. Cases cover Tab,
bare-fragment Enter, literal arguments, Escape/newline, synthetic repeat/consumed/
IME boundaries with native positive controls, failed/delayed catalogue reads and
Settings/search/Quick Actions exclusion. See [compose-command-ownership.md](../../docs/internal/compose-command-ownership.md).
Physical IME, modifier-command semantics and shared16/Classic008 policy conflicts
are separate. The suite carries no new frozen tags or mappings.

## Read-only workspace transitions — 2026-09-25

`make test-ux-workspace-tabs` runs72 cases in the required browser CI job and saves
`workspace-tabs-results.json`. New journeys cover roving tab navigation, keyboard
context menus, hide/reopen without file re-read, retained draft/media state,
hidden-read/Settings focus isolation and trusted touch background-close. Six
executions repeat390px touch contexts. Existing MRU/pin/late-read/font checks stay.
See [workspace-tab-transitions.md](../../docs/internal/workspace-tab-transitions.md).
The slice fixes read-only host interactions; editor/docking and full Piclaw
transition equivalence remain open. Frozen mappings are unchanged.

## First slice: 2026-09-21

- `@ux-original-001`: menu open/dismiss, pointer and keyboard/Escape/outside-click checks.
- `@ux-original-002`: workspace visibility toggle, preserved draft/session, no prompt submission or stored messages.
- Validation: 12/12 browser executions passed. Coverage: 2/236 scenarios; 234 unmapped.

Results live in `test-results/ux-parity/{results.json,matrix.json,matrix.md,artifacts/}`. Keep matrices tied to their run; inventory-only output contains no browser pass evidence.

## Session slice: 2026-09-21

`@ux-original-014` now covers native session selection, preserved page-local drafts and rejection of a delayed real response from the previous selection. A separate Gi regression checks New allocates a distinct child session. Combined matrix: 24/24 executions, 3/236 frozen IDs passing, 233 unmapped. The extra child test does not inflate the frozen scenario count.

The full existing web suite also passed 70/70 after the pooled-SQLite correction. See [`../../docs/internal/web-session-parity.md`](../../docs/internal/web-session-parity.md) for implementation limits and minimal-footprint TUI adaptations.

## Searchable picker slice: 2026-09-21

`@ux-original-013` adds pointer and keyboard opening, mounted search focus, native ancestry grouping, handle/JID filtering, empty results and Escape restoring the exact trigger without changing the session or draft. A separate regression covers native Tab/text editing and keyboard selection. Combined matrix: 36/36 executions, **4/236 frozen IDs passing**, 232 unmapped. Source and helper tests preserve pinned hashes and exercise grouping precedence separately from live browser evidence.

Scope is full parity, not a picker-only milestone. See [`../../docs/internal/full-web-tui-parity-plan.md`](../../docs/internal/full-web-tui-parity-plan.md) for every feature family and terminal adaptation constraints. The subsequent `015` slice implements native rename/pin/archive/restore, capability-gated controls, validation errors, reload persistence and delayed mutation-error isolation. Combined matrix: **48/48 executions**, **5/236 frozen IDs passing**, 231 unmapped. The shared contract remains unmapped. See [ADR-0009](../../docs/adr/0009-session-picker-mutations.md) for reversible archive and display-name semantics. Next: terminal session draft/event isolation, durable browser drafts and queue/model/reconnect cases. Backend records alone do not establish browser parity.

## Durable browser drafts: 2026-09-22

`drafts.spec.mjs` maps compose `001`, `002`, `003` and `006`: capture-and-clear, failure merging, empty-send rejection and captured destination after delayed upload. Gi regressions cover reload byte/reference persistence, A→B→A recovery, unacknowledged sends, quota failures and acknowledgement cleanup. The full matrix passes **102/102**, covering **9/236 Classic IDs**, with 227 unmapped. Shared cases remain unmapped. Tests use native controls and real records/responses; route aborts and IndexedDB quota faults exercise failure paths without fabricated timeline data.

See [ADR-0011](../../docs/adr/0011-browser-draft-recovery.md) for browser-local scope, uncertain delivery, asynchronous-write and cross-tab limits. Queue return/DELETE recovery and upload progress contracts remain separate.

## Durable queue controls: 2026-09-22

`queue.spec.mjs` maps `@ux-original-018`: native explicit queue admission during a busy turn, optimistic reorder/removal, durable reload order, real stale-snapshot conflict, removal-error recovery and execution in the chosen order. Additional tests hold old polls and mutation replies across session switches. Full matrix: **120/120**, **10/236 Classic IDs**, 226 unmapped; the shared contract remains unmapped.

The isolated parity server uses `tests/ux/shell/sh` to gate only `UX queue gate:<token>` test prompts until a release file exists, then runs the normal shell responder. Gates time out after 60 seconds; test cleanup releases them. No production delay hook or fabricated queue records are used. See [ADR-0012](../../docs/adr/0012-queued-followup-order.md). `017`, `019`, compose `004` and shared queue contracts still need their own evidence.

## Queue SSE and reconnect: 2026-09-22

`016` and `ux-reconnect-001` now have six-project browser evidence: an actual acknowledgement is held while SSE reconciliation arrives; same-text prompts retain distinct durable IDs; external queue mutations refresh promptly. A local byte-for-byte SSE proxy severs real sockets and blocks reconnection, allowing assertions against a real stdout preview, running state and preserved user draft. Independent API queue changes are reconciled on reconnect. A rejected-send regression checks placeholder removal and draft recovery.

Full matrix: **138/138**, **12/236 Classic IDs**, 224 unmapped, with all shared cases unmapped. See [ADR-0013](../../docs/adr/0013-queue-sse-reconciliation.md). Context usage, search preservation and version-drift reconnect criteria remain open; these tests do not map them.

## Session-local model selection: 2026-09-22

`models.spec.mjs` maps `021` and compaction `008`: late model responses stay with their origin, and `/model` resolves through native validation without creating a prompt turn. Additional tests verify pointer/keyboard mutation, reload persistence, unchanged draft/media, rejected choices and A→B→A catalogue isolation. The parity test catalogue includes the native `test-model`/`bootstrap` shell models plus an unusable entry; production provider configuration is untouched.

Full matrix: **168/168**, **14/236 Classic IDs**, 222 unmapped; shared cases remain unmapped. Native model selection has no measured context usage, so `020`, compaction `006/007` and shared model/context criteria remain open. [ADR-0014](../../docs/adr/0014-session-model-selection.md) records selection/runtime metadata separation and terminal gaps.

## Latest provider-request context: 2026-09-22

`context.spec.mjs` maps `ux-context-002`: native shell sessions have no provider token measurement and display unknown markers before/after a turn, reload and model change. Unsupported compaction is disabled. Matrix: **174/174**, **15/236 Classic IDs**, 221 unmapped; shared cases remain unmapped.

Native tests cover provider-loop recording, latest input/cache values versus cumulative turn totals, storage/reopen, session isolation, model-fit validation and unchanged TUI footer row count. `context-usage.test.ts` uses supplied numbers for formatting, thresholds and fit predicates; it does not map measured browser scenarios. [ADR-0016](../../docs/adr/0016-measured-request-context.md) lists the remaining evidence gaps.

## Index startup settings: 2026-09-23

`make test-ux-index-config` runs `workspace-index-config.spec.mjs` against a local server seeded with a real `.pi/settings.json`. Six projects verify extra root/extension indexing, optional initial absence, retained content/status on populated-root disappearance, reload/draft safety and retry. Results: `test-results/ux-parity/index-config-results.json`. [ADR-0047](../../docs/adr/0047-index-settings-and-optional-roots.md).

Latest **498 browser** (330 main + 168 specialised), **76 functional**, **32 helpers**, Go/vet/build/hook/config-search-web race ×3. Coverage **45/236 Classic**, **2/42 shared**, **191/40 unmapped**; 17 derived index proposals are not frozen mappings. Meter tests now wait on native compaction claim cleanup before sampling capability, retaining exact assertions; two full meter passes after the repair. Background freshness and terminal consumers remain open.

## Explicit scoped indexing: 2026-09-22

`workspace-preview.spec.mjs` adds a Gi-only native Reindex/lexical-query case: stored notes/skills, real missing-root scan failure, persisted status and hits, explicit retry after changed bytes, retained drafts/files across reload/session switches, and native Refresh. No new frozen mappings; workspace-005 still has unsatisfied compound menu criteria. [ADR-0046](../../docs/adr/0046-native-workspace-index-api.md).

Latest **492/492 browser** (330 main + 162 specialised), **75/75 functional**, **32 helpers**, Go/vet/build/hook and search-store/indexer/web race ×3. Coverage **45/236 Classic**, **2/42 shared**, **191/40 unmapped**. ENOSPC artifact relocation and unchanged WebKit reload/reconnect reruns are documented. Background freshness, optional roots and terminal controls remain unverified.

## Hidden files and native subtrees: 2026-09-22

`workspace-preview.spec.mjs` maps workspace-004 with native nested files, on/off/on toggle through the visible global menu, observed root/all-expanded requests, reload persistence and retained text/files. A production host bridge invokes the supplied explorer's own stateful toggle; tests use normal user clicks and real responses. [ADR-0041](../../docs/adr/0041-workspace-hidden-subtrees.md).

Latest **486/486 browser** (324 main + 162 specialised), **74/74 functional**, **31 helpers**, Go/vet/build/hook/web race ×3. Coverage **45/236 Classic**, **2/42 shared**, **191/40 unmapped**. Reindexing, CRUD and terminal chooser tests remain open. Native tree tests cover bounds, empty/stub snapshots, legacy policy, confinement and auth.

## Read-only workspace previews: 2026-09-22

`workspace-preview.spec.mjs` maps workspace-008 through files created with native tools: Markdown, escaped text, decoded PNG, binary download message and kind/extension/type/size/mtime/path. SVG script source remains text; large previews truncate at the requested limit, and composer text survives selections/reload. The fixture shell explicitly changes to the runtime workspace. No mocked preview responses. [ADR-0040](../../docs/adr/0040-read-only-workspace-previews.md).

Latest **480/480 browser** (318 main + 162 specialised), **73/73 functional**, **31 helpers**, Go/vet/build/hook/web race ×3. Coverage **44/236 Classic**, **2/42 shared**, **192/40 unmapped**. Editor, mutations, unsupported format and terminal viewer cases remain open.

## Native image lightbox: 2026-09-22

`lightbox.spec.mjs` maps timeline-013/014/015/016 with a real composer-uploaded PNG: Escape closes, other keys retain the modal, image/backdrop clicks close, and touch-enabled contexts deliver trusted taps. Additional session/search/reload/native-404 recovery preserves unsent text/files. [ADR-0039](../../docs/adr/0039-native-media-lightbox.md). No fabricated media/post responses or navigator overrides; touch evidence asserts the delivered events because Linux WebKit reports zero maxTouchPoints.

Latest **474/474 browser** (312 main + 162 specialised), **72/72 functional**, **31 helpers**, Go/vet/build/hook/web race ×3. Coverage **43/236 Classic**, **2/42 shared**, **193/40 unmapped**. PNG acceptance earns no annotation/iPad drawing/all-format/terminal credit. Screenshots attached from native touch cases at desktop/tablet/phone sizes.

## Native upload failure recovery: 2026-09-22

`drafts.spec.mjs` maps original-026: native IDs on successful uploads, native multipart parser errors and no submission on rejection. An additional partial-batch case checks stored successful bytes, newer draft merging, session isolation, reload and explicit retry. The fixture changes only the boundary header, never the multipart bytes or HTTP response. Persisted turns retain the retry IDs/filenames and exactly one user message. [ADR-0038](../../docs/adr/0038-native-upload-failure-recovery.md).

Latest: **444/444 browser** (282 main + 162 specialised), **70/70 functional**, **29 helpers**, Go/vet/hook. Main browser-family batches each pass 141 cases; a separate 12-case screenshot run also passes. Coverage **39/236 Classic**, **2/42 shared**, **197/40 unmapped**. Compose-005 progress-state and terminal pending media stay open. No application changes or new terminal evidence.

## Native Markdown rendering: 2026-09-22

`rendering.spec.mjs` maps timeline-023/024 and original-029 through stored assistant posts: full-width automatic tables, visible top-right code-copy controls with trusted plain-text copy events, and source-only SVG fences. A Gi-only case checks native horizontal wheel reachability for a wide table and draft/reload preservation. No clipboard stubs or fabricated posts. [ADR-0037](../../docs/adr/0037-native-markdown-rendering.md).

Latest: **432/432 browser** (270 main + 162 specialised), **70/70 functional**, **29 helpers**, Go/vet/hook. Main browser-family batches each pass 135 cases. Existing six result files apply. Coverage **38/236 Classic**, **2/42 shared**, **198/40 unmapped**. Clipboard-event payload observation does not prove every OS paste target; combined deletion/speech/link-preview contracts stay unmapped. Meter capability snapshot timing failed once and passed unchanged on rerun.

## Completed-claim admission: 2026-09-22

`queue-steer.spec.mjs` adds a Gi-only native completion-hook gate: the first turn is terminal/display-idle while its claim remains owned by cleanup. A composer prompt must receive its own queued ID, drain after release, preserve newer typing and appear once after reload. Existing strict run-bound Steer assertions are unchanged. Paging compares the exact persisted ID window, including native queue-status messages. [ADR-0036](../../docs/adr/0036-completed-claim-admission.md).

Latest: **408/408 browser** (246 main + 162 specialised), **70/70 functional**, **29 helpers**, Go/vet/race ×3 and all TUI suites. The main Chromium/WebKit batches each passed 123 cases; existing six result files/targets still apply. Frozen mapping stays **35/236 Classic**, **2/42 shared**, **201/40 unmapped**. No synthetic SSE or backend lifecycle seeding for the new browser regression.

## Explicit folder references: 2026-09-22

`drafts.spec.mjs` now maps compose-008 through native file/folder/message selection, exact multiline serialisation and references-only submission. Separate cases verify persistence, isolation, deduplication/cleanup and failure recovery. The host action leaves supplied component files and directory navigation unchanged. [ADR-0035](../../docs/adr/0035-explicit-folder-references.md).

Latest: **402/402 browser** (246 main + 156 specialised), **70/70 functional**, **29 helpers**, Go/vet/hook. Main Chromium/WebKit batches each passed 123 cases, combined in `results.json`; existing six result files/targets apply. Coverage **35/236 Classic**, **2/42 shared**, **201/40 unmapped**. Initial-readiness setup observes real `connected` events; no synthetic SSE. Fast history setup uses queue intent and persisted post counts because completed/display-idle can precede admission-claim release. Earlier failures and this unresolved backend boundary are recorded in ADR-0035.

## Accepted-message refresh: 2026-09-22

`drafts.spec.mjs` holds real native acknowledgements while SSE and execution continue. Six cases verify acceptance-driven refresh/deduplication, upload ID/name/byte pairing, newer draft/cursor preservation, history anchors and near-bottom following, search ownership and origin-session isolation. Upload responses are observed without multipart interception; history setup awaits native idle after completion. The search test rejects the previous callback. See [ADR-0028](../../docs/adr/0028-accepted-message-refresh.md).

Latest: **384/384 browser** (228 main + 156 specialised), **70/70 functional**, **29 helpers**, Go/vet/hook checks. Main Chromium/WebKit batches each passed 114 cases, with their successful suites combined in `results.json`. Existing six result files/targets below still apply. Coverage: **34/236 Classic**, **2/42 shared**, **202/40 unmapped**. Compose-007/009/010/011 map; compose-008 still lacks a verified folder-reference path. No terminal credit in this slice.

## Bounded native timeline paging: 2026-09-22

The Gi-only paging regression in `make test-ux-reconnect` creates history through native turns and checks latest-50 loading, wheel-driven older pages, actual visible-message anchors, more than one forward page while offline, retained drafts and late older-page/search isolation. Responses stay bounded and IDs remain ordered/deduplicated. Subpixel anchor tolerance is one CSS pixel after the supplied entry animation settles.

Latest: **54/54 reconnect/paging + 294/294 existing = 348/348 browser**, **70/70 functional**, **29 helpers**, Go/vet/hook and native paging race ×3. No new frozen mapping: Classic **30/236**, shared **2/42** (206/40 unmapped). Same six result files/targets below. No forced clicks, fake SSE or SQL-seeded browser timeline. See [ADR-0027](../../docs/adr/0027-bounded-timeline-pages.md).

## Initial refresh ownership: 2026-09-22

`make test-ux-reconnect` now includes `reconnect-005`: hold the real native SSE readiness, verify activation does not issue authoritative reads before readiness and no duplicate initial set follows, then test A→B→A and real reconnect. A separate case verifies retry after initial activity failure without draft loss.

Latest totals: **48/48 reconnect/search + 294/294 existing = 342/342 browser**, **70/70 functional**, **28 helpers**, Go/vet/hook checks. Classic **30/236**, shared **2/42** (206/40 unmapped); all reconnect IDs mapped. Exact idle-fixture request counts do not promise exactly-once event delivery. Same six explicit matrix result files apply. See [ADR-0026](../../docs/adr/0026-initial-refresh-ownership.md).

## Native search view: 2026-09-22

`make test-ux-reconnect` now also maps `reconnect-003` through the active-search alternative. Tests open the supplied search field, use native current/family/all queries, hold real old replies and sever real SSE sockets. Reconnect reruns search/activity/queue/context with zero main-timeline requests, preserving draft/media. Additional tests cover literal `%/_`, scope isolation, stale-query rejection, failed search recovery and no prompt submissions.

Latest: **36/36 reconnect/search + 294/294 existing = 330/330 browser**, **70/70 functional**, **27 helpers**, Go/vet/hook and native search race ×3. Classic **29/236**, shared **2/42**, 207/40 unmapped. The same six explicit result files below apply. Hashtag navigation and pagination receive no new credit. See [ADR-0025](../../docs/adr/0025-native-search-view.md).

## Reconnect and version drift: 2026-09-22

`make test-ux-reconnect` maps `reconnect-002/004` and tests late-error isolation. Each case owns a real loopback Go instance; a byte-for-byte proxy severs SSE sockets, and restart reuses the native database while changing the server's advertised asset version. No fabricated lifecycle data or offline-emulation evidence.

Latest: **18/18 reconnect + 294/294 existing browser = 312/312**, **70/70 functional**, **26 helpers**, Go/vet/hook checks. Classic **28/236**, shared **2/42** (208/40 unmapped). Search-specific reconnect and active-turn crash recovery remain unverified. See [ADR-0024](../../docs/adr/0024-reconnect-refresh-and-version-drift.md).

```sh
make test-ux-reconnect
# After running the other five matrix targets:
bun scripts/ux-parity-report.mjs test-results/ux-parity/results.json test-results/ux-parity/steer-results.json test-results/ux-parity/context-fit-results.json test-results/ux-parity/context-meter-results.json test-results/ux-parity/compaction-results.json test-results/ux-parity/reconnect-results.json
```

## Manual Compact: 2026-09-22

`make test-ux-compaction` now includes `context-003` and failed-delivery/stale/cancel regressions for native manual compaction. Availability and expected context tokens come from the server; native turns build history. The meter action preserves draft text/media and creates no chat prompt or inference request. Usage stays unchanged until another real provider request.

Latest totals: **66/66 compaction + 228/228 other browser = 294/294**, **70/70 functional**, **24 helpers**, Go/vet/hook and native race ×3. Classic **26/236**, shared **2/42** (210/40 unmapped). Same five matrix targets/result files below. No forced controls, retries, synthetic lifecycle messages or SQL-seeded history. See [ADR-0022](../../docs/adr/0022-manual-compaction.md).

## Durable context checkpoints: 2026-09-22

The Gi-only compaction regression now checks the next real provider request after compaction/reload: it contains the persisted summary and new request, not covered old history. All existing timeline records remain byte/ID-identical. Native store/engine tests cover reopening, repeated checkpoints, edit/delete fallback, history/version races, prefix validation and hook/tool/media eligibility limits.

Latest: **54/54 compaction + 228/228 existing browser executions = 282/282**, **70/70 functional**, **24 helpers**, Go/vet/race, plus the three-size terminal session/model regression. No additional frozen mapping: Classic **25/236**, shared **2/42**, 211/40 unmapped. Same five explicit result files/targets below. See [ADR-0021](../../docs/adr/0021-durable-context-checkpoints.md).

## Automatic compaction lifecycle: 2026-09-22

`make test-ux-compaction` maps `compaction-001–005` and `context-004`: a native hook gate pauses real automatic compaction after history created by provider turns. The suite verifies elapsed/style/title state, persisted completion, Stop retaining draft/media, suppression detail and refreshed usage (including increased reported tokens). Extra regressions cover reload and delayed activity after switching sessions, failed Stop, and cancellation arriving after completion. No synthetic events, SQL history seeds, forced clicks, retries or paid providers.

Latest: **48/48 compaction + 228/228 existing browser executions**, **70/70 functional**, **24 helpers**, Go/vet/hook and targeted race checks. **25/236 Classic**, **2/42 shared** passes, 211/40 unmapped. Manual Compact/context `003` and durable history replacement remain unavailable. See [ADR-0020](../../docs/adr/0020-automatic-compaction-web-status.md).

```sh
make test-ux-compaction
# After also running the base, steer, context-fit and context-meter targets:
bun scripts/ux-parity-report.mjs test-results/ux-parity/results.json test-results/ux-parity/steer-results.json test-results/ux-parity/context-fit-results.json test-results/ux-parity/context-meter-results.json test-results/ux-parity/compaction-results.json
```

## Context-meter formatting and colours: 2026-09-22

`make test-ux-context-meter` maps `@ux-context-001/005`. Its local provider supplies explicit usage through the native stream parser and persistence; tests use fixed expected strings/arc lengths, not the production formatter. Covered: K/M labels, percentage rounding, title/tooltip-data/accessibility, 125% overflow with full arc, zero input with nonzero output, exact 75/90 warning thresholds, reload/session ownership. The context-fit suite also checks tooltip refresh after a model switch.

```sh
make test-ux-parity
make test-ux-steer
make test-ux-context-fit
make test-ux-context-meter
bun scripts/ux-parity-report.mjs test-results/ux-parity/results.json test-results/ux-parity/steer-results.json test-results/ux-parity/context-fit-results.json test-results/ux-parity/context-meter-results.json
```

Latest totals: **228/228 browser executions** (192 + 12 + 12 + 12), **70/70 functional**, **23/23 helpers**, Go/vet/hook checks. Classic **19/236** and shared **2/42** passing, 217/40 unmapped. No forced interactions, retries, paid providers, SQL usage seeds or synthetic agent events. Compaction `context-003/004` stays unmapped. See [ADR-0016](../../docs/adr/0016-measured-request-context.md).

## Measured model context fit: 2026-09-22

`make test-ux-context-fit` maps `@ux-compaction-006/007` using the existing local-provider fixture with opt-in registry capacities 80/100/200. A real streamed result reports 100 prompt tokens; production inference persists the measurement. Tests verify blocked pointer activation sends no mutation, direct PATCH rejects the undersized model, accepted switches refresh capacity/percentage, equal capacity is permitted, and reload/session switches retain model/text/media ownership. Unknown usage stays unknown in another session.

Run all three matrices explicitly before combining evidence:

```sh
make test-ux-parity
make test-ux-steer
make test-ux-context-fit
bun scripts/ux-parity-report.mjs test-results/ux-parity/results.json test-results/ux-parity/steer-results.json test-results/ux-parity/context-fit-results.json
```

Latest: **216/216 browser executions** (192 + 12 + 12), **70/70 functional**, **23/23 helpers**, Go tests/vet and hook checks. Classic **17/236** and shared **2/42** passing; 219/40 unmapped. Individual reports mark absent fixture mappings not-run. No provider costs, SQL-seeded usage, synthetic agent events, forced clicks or retries. Full compaction and other context-meter IDs are unverified. See [ADR-0016](../../docs/adr/0016-measured-request-context.md).

## Shared run-bound queue Steer: 2026-09-22

`queue-steer.spec.mjs` maps `@shared-30`; the catalogue test pins its title. Run `make test-ux-steer` for the isolated Go server and local streaming provider. Native API/SSE, SQLite and inference checkpoints are unchanged. The fixture uses temporary credentials/workspace, binds to loopback and closes on exit. No paid provider, forced clicks, retries or fabricated timeline events are used.

The six-project suite verifies idle/unknown disabled controls (real SSE disconnection), run/session ownership, failed admission, duplicate activation, actual second-request delivery, held recovery/reload, retry into a new run and stale replies after switching sessions. Native Go tests cover atomic rollback, media projection, persistence failure and at-most-once acknowledgement.

Current validation: **12/12 Steer + 192/192 existing browser executions**, **70/70 functional**, **23/23 helper tests**, Go/vet and targeted race tests repeated three times. Combine explicitly:

```sh
make test-ux-parity
make test-ux-steer
bun scripts/ux-parity-report.mjs test-results/ux-parity/results.json test-results/ux-parity/steer-results.json
```

The combined report has **15/236 Classic** and **2/42 shared** passes (221 and 40 unmapped). Individual runs report the other mapping as not-run. Classic idle-Steer/immediate-send (`019`) stays unmapped. See [ADR-0018](../../docs/adr/0018-run-bound-queue-steer.md).

## Shared durable queue return: 2026-09-22

`queue-return.spec.mjs` maps `@shared-28` (the immutable shared corpus's ordinal ID), with a name assertion preventing mapping drift. Return merges the latest origin draft/media/refs and persists recovery before DELETE. Tests hold real media responses, inspect committed state at DELETE, inject quota/transport failures, reload/retry and verify no duplicate recovery. Separate regressions cover already-consumed items and session switches.

The report now includes separate shared rows/counts. Full matrix: **192/192** executions; Classic **15/236** passing (221 unmapped), shared **1/42** passing (41 unmapped). Classic `017` and compose `004` prescribe replacement/media clearing and remain unmapped; their frozen assertions were not weakened. See [ADR-0017](../../docs/adr/0017-durable-queue-return.md).

## Compose transport progress — 2026-09-23

`@ux-compose-005` now has native upload/prompt phase evidence in `drafts.spec.mjs` across all six projects. XHR provides real byte progress; separate per-session operations drive host status and a send-button ring without altering the supplied composer. Tests preserve exact multipart bytes, newer typing/caret, origin session, failure recovery and concurrent operations. Browser interception holds real responses; it never fabricates success or progress. Direct upload coverage separately records real computable progress events.

Current complete run: **516 browser executions** (348 main, 54 reconnect, 66 compaction, 18 Steer/admission, 12 fit, 12 meter, 6 index config), **78 functional**, **34 helpers**. Combined mapping is **46/236 Classic**, **2/42 shared**, **190/40 unmapped**. See [ADR-0053](../../docs/adr/0053-compose-upload-and-send-progress.md). Terminal media progress remains design-only.

Use `UX_LOCAL_BIN=/tmp/gi-scheduler-bin/gi-ux-steer` for isolated local/reconnect binaries when workspace disk space is low. Reconnect forwards that path as `GI_UX_SERVER_BIN`; defaults are unchanged.

## Native Quick Actions — 2026-09-23

`quick-actions.spec.mjs` maps original-003/005/006/007 to pinned Classic typing/navigation, excluded key events, dismissal and native command prefill. Eight tests across six projects add 48 browser executions. The native catalogue exposes only workspace visibility, `/model` and `/compact`; session actions use real stored sessions. Full excluded surfaces (004) and loaded skill commands (008) remain unmapped. Shared queue merge evidence does not satisfy incompatible Classic replace-and-clear criteria.

The complete main suite now contains **396 executions**. Three disjoint project batches passed and their saved reports combine with **168 supplementary executions** for **564 browser checks**, **50/236 Classic**, **2/42 shared**, **186/40 unmapped**. Final capture48 and transition-stress24 checks are additional, not duplicated in the combined report. Go/vet/build/hook, 79 functional, 36 helpers and web race×3 pass. See [ADR-0054](../../docs/adr/0054-native-quick-actions.md).

When splitting a matrix, pass each project's report exactly once to `ux-parity-report.mjs`. Interrupted or repeated reports cannot supply complete-matrix credit. The current main reports are `/tmp/gi-quick-main-{chromium-mobile,mixed,webkit-large}.json`.

## Whole-message clipboard safety — 2026-09-23

`message-copy.spec.mjs` adds five tests × six projects for native stored-source Markdown/rich copy, unavailable APIs, rich-to-plain fallback, denial/reset and late session-switch failure. Success observes trusted browser copy events synchronously; failure injection does not count as OS clipboard evidence. The bundled Post helper override is exercised directly without editing supplied sources.

**594 browser executions** pass (426 main + 168 supplementary), plus **80 functional** and **38 helpers**. Coverage stays **50/236 Classic**, **2/42 shared**, **186/40 unmapped**: original-024/shared-37 also require native deletion/cascade behaviour. [ADR-0055](../../docs/adr/0055-message-copy-clipboard-safety.md) records the fix and terminal boundaries. Main reports: `/tmp/gi-message-copy-main-{chromium-mobile,mixed,webkit-large}.json`.
