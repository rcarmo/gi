# Classic026: attachment upload before message submission

Gi's mounted `components/compose-box.ts` captures the draft, uploads selected
media before constructing the message request, and pairs returned IDs with
source filenames. An upload rejection restores the captured text/files through
`onDraftFailed`; it cannot reach `sendAgentMessage` as a successful attachment.
The pinned Piclaw 3.2.4 compose-box source is changed in Gi, so the browser
journey establishes only the native branch, not source identity.

`tests/ux/drafts.spec.mjs` passes real multipart bytes through the native
server, removes the boundary from one upload to elicit HTTP 400, and checks
that no prompt or queued turn was submitted. It retains a newer unsent draft,
restores both files, survives reload without automatic retry, then explicitly
retries and checks the durable media IDs and session-scoped request. The tagged
`@ux-original-026` no-successful-prefix variant passed **6/6** Chromium/WebKit
viewport projects. A separate untagged partial-batch variant checks a prior
successful upload and cross-session recovery; it was not part of this focused
run.

This does not establish server deduplication, cleanup of previously uploaded
bytes on failure, current Piclaw backend acceptance or live Gi delivery. No
production code or frozen Gherkin changed.
