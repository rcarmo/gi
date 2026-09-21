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

Next: searchable picker/focus (`013`), real session mutations (`015`), durable drafts and associated queue/model/reconnect race cases. Backend records alone do not establish browser parity.
