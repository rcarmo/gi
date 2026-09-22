# Piclaw web interaction parity

Gi vendors the same frozen Classic Gherkin baseline used for the Tau/Vibes audit:

- Piclaw commit `70d33bc93ab540845bbcf5f80503ca8125c71594`.
- `features/classic/`: 24 byte-identical feature files, 236 unique scenario IDs, 256 expanded cases.
- `features/shared-canonical-ux.feature`: byte-identical shared Tau/Vibes interaction contract, 42 expanded cases; SHA-256 `a08a623880c6f327bc051edc51bb2bbff2959aed86421b5227e61d5a92fc2441`.
- Upstream file hashes and provenance are preserved in `upstream/`.

The sources came from `/workspace/evidence/piclaw-classic-70d33bc93` and `/workspace/tau/tests/ux/features/canonical-ux.feature`. Tests do not depend on those paths after copying. Source-relative links inside frozen features refer to the original Piclaw checkout.

The Classic and shared contracts differ in places, including command-prefill and idle-Steer safety semantics. Both are preserved. Passing a Classic case does not imply passing the stronger shared contract.

## Run

```sh
make ux-parity-inventory
make test-ux-parity
make test-ux-parity UX_PARITY_ARGS='--project=chromium-desktop'
```

The parity target uses `.gi-ux-parity/` and loopback port 19091, seeds deterministic `test-model` configuration, and removes its isolated process/database after the run. It does not touch the developer's sessions. Install browsers with `scripts/run-playwright.sh install chromium webkit` if necessary.

`@cucumber/gherkin` parses the original features and expands examples. `classic.spec.mjs` supplies Gi-native browser steps for mapped IDs, attaches the original steps to each result, and drives visible enabled controls. Native APIs only seed fixtures and verify persistence; they do not substitute for tested user actions. No DOM injection, forced clicks, automatic resubmission, navigation fallback, retries, or inherited Tau/Vibes pass statuses.

`support/catalogue.mjs` rejects modified source hashes. The report always inventories every frozen scenario. A scenario passes only when all expanded cases pass in Chromium and WebKit at 390×844, 820×1180, and 1440×900. Unmapped scenarios are reported as `unmapped`, not skipped or passed. The shared contract is inventoried separately and has no Gi step mappings yet.

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
