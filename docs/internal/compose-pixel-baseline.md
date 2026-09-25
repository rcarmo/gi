# Compose and panel pixel baseline

The cross-UI pixel gate fails. The original baseline captured 72 images with no route or page errors, but all 18 Gi/Piclaw comparisons differed and 13 of 36 same-host comparisons were unstable. The latest [session-panel run](session-panel.md) also captured 72 images; all 18 cross-host pairs differ and 15/36 same-host pairs are unstable. Geometry tests and successful screenshot capture do not establish pixel parity.

## Run

```sh
make deps
make test-pixel-helpers
PICLAW_PIXEL_ROOT=/path/to/piclaw-3.2.2/app/runtime make pixel-baseline
# Recompute reports from saved images without launching a browser:
PIXEL_RUN_DIR=test-results/compose-pixels/run-<id> make pixel-compare
```

The runner needs headed Playwright Chromium, `xvfb-run`, Fontconfig and the pinned reference files. `PICLAW_PIXEL_ROOT` is a runtime directory, not a server URL. The release installer can be extracted into a test-only directory; do not install or start it. The installed 3.2.2 tree was removed during a host upgrade to 3.2.3, so local evidence now uses an extracted published 3.2.2 release. All eight hashes match the existing manifest. The running Piclaw installation and its dirty source checkout were untouched.

`PIXEL_OUTPUT` defaults to `test-results/compose-pixels`; each invocation creates a unique `run-<timestamp>-<pid>` directory, printed at exit. Old failure screenshots and subset runs cannot contaminate a later run. `PIXEL_VIEWPORTS`, `PIXEL_THEMES` and `PIXEL_SCENARIOS` accept comma-separated subsets. Reports disclose reduced coverage. The full matrix uses 390×844, 820×1180 and 1440×900; light/dark themes; compose, model picker and session picker; two captures per host. No cases are skipped. Phone/tablet labels describe viewport sizes, not physical-device or touch acceptance.

A completed run exits 1 for pixel differences and 2 for incomplete capture evidence. Zero requires exact full-frame equality for every repeat and cross-host pair in the selected matrix. Make reports a failed recipe as exit 2 even when the script exits 1. The pixel gate is intentionally not a green required CI job while these failures exist. Its helper tests run in required CI.

## Inputs and capture

- `tests/ux/fixtures/compose-pixel-reference.json` pins the published Piclaw 3.2.2 Classic HTML, bundled CSS/JS, editor dependency and ancillary assets. Source commit: `0afe5366ced9bca8246abd99a0feb1875a6ffbcc`; asset version: `15958f2c3dc9`.
- `compose-pixel-state.json` contains one session, one model, an empty timeline, a fixed draft and zero context usage. Host adapters translate that state into each native API shape. This does not exercise native backend behaviour.
- Each capture gets a fresh headed browser and context, en-US, UTC, DPR1, a fixed clock, blocked service workers, software rendering and sRGB. The report records the browser binary hash, launch arguments, Fontconfig inventory, requested asset hashes and loaded font families. System fonts remain an environment dependency; cross-machine reproducibility is unverified.
- Asset preflight fails on missing or changed reference files. Candidate assets are hashed before capture, checked when served and checked again on finalisation. External requests, undeclared routes/writes, wrong session scopes, page errors and network failures reject evidence.
- Only the first aborted Gi `/sse/stream` fetch is retained in `streamAborts`; capture requires a live replacement stream. Further aborts, Piclaw aborts and topic-stream aborts fail. The allowed Piclaw visibility/presence writes terminate in fixture responses.
- Fonts and images settle, input loses focus, the pointer moves away, and two animation frames precede capture. A fixed warmup screenshot is saved before each final screenshot. Screenshots disable animations and hide carets equally for both hosts. There are no masks, injected layout rules, resizing or screenshot selection based on pixel results.

The raw originals, warmups, DOM geometry/style metadata, request traces, image hashes, full-frame diffs and overlays are retained. Compose/panel reports use the union of both elements' viewport bounds, preserving displacement. These regions supplement the mandatory full-frame result; they cannot turn a failing frame into a pass. Missing finalisation, missing images, changed image hashes and missing/duplicate repeats fail closed. Comparison can resume from saved finalised evidence after interruption.

## Baseline differences before the surface adaptation

The [compose surface adaptation](compose-surface.md) now matches the six compose/textarea rectangles and reduces compose-region differences to 1,514–2,438 pixels. The table below retains the pre-adaptation measurements; panel structure and exact pixels still fail.

Latest desktop/light captures:

| Surface | Piclaw | Gi |
|---|---|---|
| Compose box | 900×139 at y=761 | 900×126 at y=774 |
| Textarea | 80px high, 2px top padding | 70px high, 32px right padding |
| Model panel | 680×263 at y=518 | 680×147.5 at y=636.5 |
| Session panel | 878×306 at y=475 | 878×300 at y=484 |

The reference model panel has a search/count header, separate model metadata, reasoning controls and a settings action. Gi renders its older title/list/Next-model structure. Composer metadata order and session-panel controls also differ. These need host-only adaptations; supplied component/pane files and CSS must remain unchanged.

Same-host differences range from 2 to 5,048 pixels in the final run, chiefly at antialiased borders. Identical DOM geometry and computed styles did not establish the cause. Earlier all-zero desktop repeats did not generalise to the full matrix. No antialias tolerance or automatic waiver applies. The previous full run had 20 unstable pairs; the final run still fails despite the lower count.

Next work: isolate repeat instability, close remaining compose control/type differences and model/session capability gaps without losing existing keyboard, session and draft ownership. The [session-panel adaptation](session-panel.md) matches all six outer rectangles and retains native mutation semantics; exact pixels and unsupported reference actions remain open. Then expand model counts, filtering, loading/error/disabled states, media and focus states. Frozen Visual criteria, physical devices, native authentication prompts, feature mappings and TUI acceptance remain separate.

## Regression and skip accounting

Go tests, vet, hook checks and 14 pixel-helper tests pass. The isolated stock functional run passes 107 of 118 cases and skips 11:

- Six auth cases require `GI_UX_SERVER_BIN` and the disposable auth fixture (verification focus, owner inventory, last-key fallback, logout, setup draft preservation, bound setup authority).
- Five existing specialised cases require their links, thread-scroll, chart, recovery-placeholder or card-rejection fixture flags.

These skips are recorded separately from passes. The earlier 113-pass/five-skip workspace run used the auth fixture; its count does not apply to this stock run. Pixel captures are render-only and do not close those native integration checks.
