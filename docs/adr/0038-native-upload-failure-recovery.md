# ADR-0038: Native upload rejection and explicit retry

## Status

Accepted — 2026-09-22. Classic original-026 passes all six browser projects. Coverage is 39/236 Classic and 2/42 shared, with 197/40 unmapped. Application code and supplied components are unchanged.

## Upload and submission evidence

`tests/ux/drafts.spec.mjs` observes actual upload responses and prompt requests. To induce a native rejection, the fixture holds an upload and removes the multipart boundary from its Content-Type header. The body remains untouched; the production parser returns HTTP 400 with `no multipart boundary param in Content-Type`. The test does not fulfil a fabricated error response or rewrite multipart bytes.

Two cases cover rejection of the first file and rejection after a successful first upload. While the request is held, the composer is empty and accepts a newer draft and attachment. No prompt, queue or Steer request has been dispatched. On failure:

- the error is visible in the originating session;
- captured text/files merge with newer text/files and persist in IndexedDB with no pending send;
- the partial-batch case switches to another session before the failure, preserving that session's draft and suppressing the foreign error;
- native message and turn lists remain empty, and the rejected upload creates no media record;
- a successful prefix retains its exact filename and bytes;
- reload retains the merged draft/files and does not retry automatically.

An explicit retry uploads the retained batch and submits one prompt. Each response supplies a positive integer ID, each native media record has the expected filename and bytes, and prompt media references match the returned `{media_id, session_id}` pairs. The persisted turn retains those pairs and enriched filenames. Exactly one user message and one turn exist. The unrelated session has no messages or uploads.

The successful prefix from a failed batch remains stored. Retry uploads new copies and attaches only their new IDs. The frozen original-026 criterion explicitly excludes retry deduplication and source deletion; this slice implements neither.

## Verification

**444/444 browser executions**: 282 main (141 Chromium + 141 WebKit), 54 reconnect, 66 compaction, 18 Steer/admission, 12 context-fit and 12 context-meter. The two upload cases also passed a separate 12-execution capture run. **70/70 functional**, **29 helpers**, full Go tests/vet and hook checks pass. The reporter requires all six projects and keeps compose-005 unmapped.

Initial fixture assertions assumed empty turns were `[]` (the endpoint returns `null`) and persisted media metadata equalled the submitted pair exactly (the engine enriches it). The assertions now normalise only the empty-list representation and compare persisted identity pairs and filenames. No application defects were demonstrated.

Native error-state screenshots are attached from `/workspace/tmp/gi-ui-captures/september22/upload-error-{desktop,phone}.png`. Test screenshots are also attached to the Playwright cases. Independent delegated review timed out and supplied no evidence.

## Progress-state gap

Compose-005 requires separate upload progress and submission state. The supplied composer uploads files sequentially before submission but exposes no separate progress state. Original-026 exercises success IDs and upload errors; compose-005 remains unmapped. No supplied component changes were made to bridge that gap.

## Terminal adaptation

Pending media is still design work. Keep selected files in session-owned draft state, separate from persisted media IDs. An explicit attach action may open a temporary bounded chooser; it must preserve editor text/cursor on dismissal. Display transient `Uploading 1/2` and failure notices in the existing status area, with no new permanent rows. Keep Pi's transcript separators and padding unchanged.

Capture destination and selected files before asynchronous work. Persist recovery data before upload, reject submission unless every selected file has a valid native ID, and merge failure recovery only into the origin draft. Retry requires an explicit action; reconnect or reopen must not send. Partial success may retain unused stored media, matching the current web contract.

Before terminal credit, verify first-file failure, partial success, newer typing/cursor, switch-away failure, reload/reopen, ID/filename/byte pairing and explicit retry at 60×18, 100×22 and 140×36 in the applicable modes. Idle row counts must remain unchanged. No terminal implementation or PTY suite changed in this browser-only slice.
