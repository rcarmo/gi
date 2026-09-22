# ADR-0039: Native media projection and lightbox dismissal

## Status

Accepted — 2026-09-22. Classic timeline-013/014/015/016 pass all six browser projects. Coverage is 43/236 Classic and 2/42 shared, with 193/40 unmapped.

## Host adaptation

Uploaded files already persisted in native media storage and user-message payloads. The timeline adapter omitted their IDs and typed blocks, and the supplied `/api/media/:id` helpers had no native routes. `web/src/gi-message-media.ts` now projects ordered native references into the supplied Post component's parallel `media_ids` and `content_blocks`, shared by timeline and search results. It rejects invalid IDs and explicitly foreign session references, preserves non-media blocks, and derives image/file type and filename from native metadata. Existing media blocks are replaced by the native reference projection.

`internal/web/media_lookup.go` supplies metadata and original bytes at `/api/media/:id` and `/api/media/:id/raw`, behind the same single-user authentication boundary as the existing session routes. GET and HEAD are supported; raw bytes also support ranges. Native storage has no generated thumbnails, so the thumbnail helper uses the original image URL. This avoids a second image pipeline; large images still cost their full download size.

Responses use `private, no-store` and `nosniff`. Raw responses carry a sandbox/default-src-none CSP. Raster image types may render inline; other types, including HTML and SVG documents, receive attachment disposition. The route has no per-session access-control layer beyond Gi's existing instance-wide authentication. Tests check rejection after enrollment and access with a valid bearer token. Browser enrollment/cookie flows are outside these lightbox scenarios.

Supplied components, UI helpers and panes are unchanged. Image annotation, thumbnail generation, download lifecycle and richer previews are not established by this slice.

## Browser evidence

`tests/ux/lightbox.spec.mjs` creates a PNG through canvas, selects it through the composer and submits through native upload/prompt APIs. The tests identify the stored user message, verify image decode dimensions, and compare the retrieval bytes with the original upload. No image, post or lifecycle response is fabricated.

- **timeline-013:** Escape closes the image modal; timeline and unsent draft/files remain available, including after reload.
- **timeline-014:** Space, Enter, a letter and arrow keys leave it open without submitting or changing the draft.
- **timeline-015:** both backdrop and image clicks close the modal.
- **timeline-016:** touch-enabled contexts open/dismiss through native taps on backdrop and image. All four `touchstart` events must be trusted and contain one touch.
- **Gi guard:** session changes and search preserve image projection and origin drafts. A request redirected to a nonexistent native media ID returns a real 404; reload after removing the redirect recovers image rendering without draft loss.

Playwright WebKit on Linux reports `navigator.maxTouchPoints=0` even with touch emulation. The initial fixture stopped at that property check; the corrected fixture checks actual trusted touch events without patching navigator. A separate fixture initially used the nonexistent button label `Exit search`; it now selects the supplied `Close search` control. Neither issue required application changes.

Desktop, tablet and phone screenshots from the passing touch cases were attached as `gi-lightbox-{desktop,tablet,phone}.png`. They are also available under `/workspace/tmp/gi-ui-captures/september22/`.

## Verification

**474/474 browser executions:** 312 main (156 per browser family), 54 reconnect, 66 compaction, 18 Steer/admission, 12 context-fit and 12 meter. **72/72 functional**, **31/31 helpers**, full Go tests/vet/build/hook checks and web race tests ×3 pass.

Native endpoint tests cover metadata, original bytes, range/HEAD, invalid IDs/paths/methods, active-content response headers and authentication. Helper tests cover reference ordering, types, invalid/foreign references and input immutability. The new functional case verifies upload/render/lightbox/reload through the native path. Raster browser acceptance uses PNG; it does not establish support for every stored MIME type. Independent delegated review timed out earlier and supplied no evidence.

## Terminal adaptation

Use an explicit attachment action with a temporary bounded selector showing filename/type/size. Escape dismisses it and restores editor text, cursor and transcript reading position; unrelated keys never submit. An optional open/download action must be explicit and capability-gated. Preserve Pi transcript padding and add no idle rows or image panel. Inline terminal graphics and browser-specific backdrop/touch dismissal are not required equivalents.

The selector, media retrieval failure and explicit opening need independent 60×18, 100×22 and 140×36 acceptance before terminal credit. No terminal code changed and PTY suites were not rerun for this web-only slice. Pending terminal media remains open alongside iPad drawing, annotation and richer browser preview contracts.
