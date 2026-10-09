# Compose surface adaptation

Gi now matches the pinned Classic compose-box and textarea bounds in the six viewport/theme captures. Exact pixels still differ. Model/session panel structure and same-host raster instability are separate open work.

## Host-owned changes

`patch-compose-surface.mjs` runs after the existing guarded compose adapters. It moves the session control from the old two-button footer into the reference's single top pill, adds the resize separator, and delegates height changes to `gi-compose-surface.ts`. Every source anchor must occur once. The supplied component, UI utility, pane and CSS sources are unchanged; `gi-compose-surface.css` contains the host overrides.

The textarea reads `piclaw_compose_height`, clamps manual height to the reference viewport bounds, and preserves the raw preference through viewport changes. Automatic growth stops at 40% of viewport height or 300px; manual height stops at 50% or 520px. Phone CSS additionally caps automatic layout at 30% or 200px, matching the reference. Long input scrolls without changing its value. The old `flex-basis: 0` discarded explicit automatic height; the host override uses `flex: 0 0 auto`.

The resize handle supports mouse and touch drag. ArrowUp/ArrowDown change height by ten pixels; Home and double-click restore automatic sizing. Mouse/touch release commits the preference. Escape, touch cancellation, window blur and unmount cancel a drag and restore body cursor/selection styles. Modal dialogs and hidden/detached textareas exclude surface interaction. Draft text, media, references, session selection and submission remain owned by the existing component and host draft controller. No new terminal rows or TUI controls were added.

The shared23/shared24 tests formerly required two trigger buttons, a Gi layout assumption absent from the frozen criteria. They now require the reference's single top pill and retain insertion-time focus, first-frame focus, keyboard/pointer opening, anchored bounds, filtering, Escape focus return and draft/session retention checks.

## Pixel evidence

Final run: `run-1790364442422-4076546`, using the unchanged pinned 3.2.2 reference and capture settings from [the pixel baseline](compose-pixel-baseline.md). All 72 captures completed with no fixture/page failures. All 18 cross-host frames differ; 13/36 repeat pairs are unstable. No masks, tolerance or typography-rendering flag changes were accepted.

| Viewport | Compose y / height, both hosts | Textarea y / height, both hosts | Compose-region changed pixels, light / dark |
|---|---|---|---|
| Phone 390×844 | 708 / 136 | 731 / 80 | 1,514 / 2,025 |
| Tablet 820×1180 | 1041 / 139 | 1067 / 80 | 1,633 / 2,193 |
| Desktop 1440×900 | 761 / 139 | 787 / 80 | 1,658 / 2,438 |

The baseline desktop/light compose-region difference was 25,316 pixels. Remaining differences include controls and typography; full-frame comparison also includes unrelated shell differences. Matching dimensions does not close Visual or pixel acceptance. Diagnostic font-raster flag experiments improved four phone repeats but were not adopted or generalised.

## Verification

- `make test-ux-compose-surface`: 18 passes across Chromium/WebKit × phone/tablet/desktop. Covers persisted sizing, drag commit/cancel, keyboard reset, viewport caps, long-draft scrolling, picker focus, synthetic touch cancellation, modal exclusion and window blur. Physical touch/IME acceptance is untested.
- `make test-ux-picker-geometry test-ux-journey test-ux-slash`: 24, 42 and 72 passes.
- `make test-ux-parity UX_PARITY_ARGS='tests/ux/drafts.spec.mjs --project=chromium-phone --project=webkit-phone'`: 56 passes, including media/reference persistence and failed-send recovery.
- Full model/session matrix: 162 passes on the final rerun. An earlier run passed 161 and failed the existing session-search active-row assertion on Chromium desktop. Five isolated repeats passed before the full rerun. The intermittent failure's cause is unresolved.
- Go tests, vet, hook checks, two compose-helper tests and 14 pixel-helper tests pass. Stock functional suite: 107 passes, 11 fixture-dependent skips, as listed in the pixel baseline.

Initial touch dispatch without a touch-capable browser context, wrong journey-fixture calls and a wrong model-test fixture were harness errors. They provide no product-defect evidence. Interrupted runs are not counted as completed acceptance. The required CI journey job now includes compose-surface checks. Deployment requires green CI; no new feature mappings or passkey/physical/Visual acceptance credit is assigned.
