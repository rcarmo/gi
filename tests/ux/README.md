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

The isolated parity server uses `tests/ux/shell/sh` to gate only `UX queue gate:<token>` test prompts until a release file exists, then runs the normal shell responder. Gates time out after 60 seconds; test cleanup releases them. No production delay hook or fabricated queue records are used. See [ADR-0012](../../docs/adr/0012-queued-followup-order.md). `016`, `017`, `019`, compose `004` and shared queue contracts still need their own evidence.
