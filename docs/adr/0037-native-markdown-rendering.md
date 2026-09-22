# ADR-0037: Native Markdown table and source-copy acceptance

## Status

Accepted — 2026-09-22. Classic timeline-023, timeline-024 and original-029 pass all six browser projects. Coverage is 38/236 Classic and 2/42 shared, with 198/40 unmapped.

## Table adaptation

The frozen table criterion requires `display: table`, full content width and automatic column layout. The supplied content stylesheet uses `display: block`. A scoped style in `web/src/app.ts` overrides timeline tables to `display: table; width: 100%; table-layout: auto` and gives table-bearing `.post-content` containers horizontal overflow. Supplied component and stylesheet files remain unchanged.

Overflow belongs to the content container so a wide table can retain native table layout without being clipped by the supplied post-body boundary. Native wheel acceptance verifies the last column is reachable, the page does not grow horizontally and a reload preserves the draft. Non-table posts keep their existing overflow rules.

## Stored-post and clipboard evidence

`tests/ux/rendering.spec.mjs` creates each post through the native session/prompt API and deterministic shell-provider response, waits for the durable turn, then locates the stored assistant ID in the actual timeline. A prose prefix and blank line separate the shell's response prefix from Markdown fences/tables; no rendered post or SSE payload is invented.

- **timeline-023:** a normal table fills its content container within one CSS pixel, uses table display/automatic layout, and retains both rows and Unicode cell text.
- **timeline-024:** each tested code block has a visible copy button within the top-right bounds, and its normal action places the exact code text into a trusted browser copy event. Highlighted HTML is excluded.
- **original-029:** a model-facing SVG fence remains literal code, creates no matching inline SVG, and copies its exact source through the normal control.
- **Gi overflow guard:** a twelve-column table remains inside the page, horizontal wheel scrolling reaches its final cell, and content/draft survive reload.

Clipboard observation runs in the bubble phase after the supplied copy handler fills `clipboardData`. It replaces neither `execCommand` nor clipboard APIs and asserts `event.isTrusted`. Initial microtask observation read empty data after the event's lifetime; synchronous observation fixed the fixture. This proves browser copy-event payloads, not paste behavior in every operating-system application. Copy failure/retry and unavailable-clipboard paths are not mapped by these tests.

## Verification

**432/432 browser executions**: 270 main, 54 reconnect, 66 compaction, 18 Steer/admission, 12 context-fit and 12 meter. Main results combine successful 135-case Chromium/WebKit batches. **70/70 functional**, **29 helpers**, full Go tests/vet and hook checks pass. No terminal implementation changed, so the existing terminal matrices were not rerun in this browser-only slice.

The first wide-table check sampled wheel motion before the final cell settled; the fixture now polls the real cell bounds. A separate meter run sampled an unavailable compaction capability before cleanup, then observed the UI's available label and failed. The unchanged full meter suite passed on rerun; that capability-snapshot timing remains a fixture follow-up. Frozen criteria/hashes are unchanged.

Screenshots of native desktop and phone views were attached from `/workspace/tmp/gi-ui-captures/september22/web-markdown-*.png`. Browser artifacts and result files remain under `test-results/ux-parity/`.

## Terminal adaptation and remaining contracts

Terminal tables should use existing width-bounded textual projection, without a permanent table pane. Fenced SVG stays source text; terminal image protocols are not required by this source-code contract. Existing selection/copy and native scrollback supply explicit access to retained text without idle buttons. Exact source whitespace across Markdown projection and soft wraps requires independent terminal acceptance; browser evidence earns no terminal credit.

Remote links plus link previews (timeline-025), combined post/code copy and cascade deletion (original-024), speech ownership (original-028/timeline-027/028), and failed clipboard feedback remain unmapped. Three source-rendering passes do not establish those compound contracts.
