# Basic web send on HTTP hosts

## Incident

The deployed browser composer could not submit on ordinary non-localhost HTTP.
`createDraftRepository.begin()` called `crypto.randomUUID()` before the network
request and outside the send error handler. That API is secure-context-only;
on an HTTP hostname or LAN address it was undefined. The exception stopped
submission before POST and gave no useful UI error.

A read-only live probe reproduced the difference in the same Chromium browser:
loopback HTTP reached `/api/sessions/<id>/prompt`, while `http://gi-send.test`
(mapped to loopback without marking it trustworthy) threw
`crypto.randomUUID is not a function` and made no POST. Both had
`crypto.getRandomValues`. All writes were blocked by the probe.

Earlier localhost browser tests and read-only deployment checks missed this.
Read-only UI acceptance is not proof of sending. The older functional checks
also used post counts that could be satisfied by previous messages; the new
basic journey checks unique prompt and admitted turn identities instead.

## Fix

`web/src/gi-random-id.ts` creates UUIDv4 identifiers with16cryptographically random
bytes from `crypto.getRandomValues`, with the correct version/variant bits.
No `Math.random`, global API shim or secure-context override is used. Missing
randomness fails closed rather than emitting weak IDs.

Gi draft capture, restoration and prefill use the helper. The immutable
composer's queue fallback is adapted at build time by
`scripts/patch-compose-random-id.mjs`, which requires exactly one known anchor.
Supplied component bytes remain unchanged. Authentication, passkeys, secure
cookies, origin guards and notification capability gates are untouched.

## Deployment

Exact `f7c969323d4c21347cb131d602d8f71aa8a97e2a` passed product
[CI36230683412](https://github.com/rcarmo/gi/actions/runs/36230683412), including
the required HTTP-send job and all four builds. Detached-source deployment runs
on8090, PID1276518, PGID/SID1276508. Uncommitted autosave work is excluded.

Four deployed non-loopback HTTP checks (Chromium/WebKit × Return/Send) confirm
insecure context, absent `randomUUID`, the exact POST path/prompt, retained text
and visible unknown-delivery feedback when the probe blocks the request. No
native writes occurred. Six existing read-only UI probes also pass. DB remains
62sessions/51turns/146messages, integrity/FKs clean, normalised SQL excluding
runtime leases and absent auth unchanged. Existing browser tabs must reload to
load the fixed asset.

Separately, a disposable instance using the actual configured
`github-copilot/gpt-5-mini` credentials admitted an HTTP-origin browser send202,
completed with `WEB_SEND_PROVIDER_OK`, and displayed it before/after reload with
no browser errors. Its temporary credential copy was removed after shutdown.
No provider request was made in a live chat. An additional real-HTTP probe in
both browser engines passed New session→send/response→return to parent draft→
reload→return to child history. Initial probe selector/response-shape mistakes
were corrected before the accepted run. Full live-chat mutating acceptance is
not claimed; the deployed transport, isolated end-to-end and provider evidence
are distinct.

## Acceptance

`make test-web-basic-send` starts a disposable native instance and uses a
non-loopback HTTP reverse proxy for both Chromium and WebKit. No browser
`crypto` or `isSecureContext` substitution is used. It asserts real insecure
context and absent `randomUUID` on that origin, then checks:

- Return and Send each create exactly one new native turn with the exact unique
  Unicode prompt in the selected session.
- The turn completes; the matching user message and deterministic native
  assistant response appear exactly once.
- Reload preserves both, and a second message succeeds independently.
- Admission503 and aborted transport retain text, show truthful error/unknown
  delivery feedback, and make no native admission. Explicit retry then admits
  exactly once and shows the matching response.

The response model is the deterministic `test-model`; this proves the
browser→native admission→response→reload path, not external-provider availability.
The proxy forwards only to the isolated test server, never the live database.
Tests require a non-loopback IPv4 interface rather than silently skipping HTTP.

`make test-web-http-helpers`:10tests/35assertions pass, including draft capture
without `randomUUID`, bit layout, uniqueness sample, and strict adapter anchors.
`make test vet bun-checks` passes. Full isolated functional suite:113passed,
11existing skips. A focused read-only review found no blocker. Initial new test
failures were incorrect helper API assumptions and a null empty-store turn
array; fixed before the accepted runs, with no weakened journey assertions.

A required `Basic web send on HTTP and localhost` CI job gates all platform builds.
Deployment requires whole product CI for the exact commit, not a cleanup
workflow or a read-only probe. Live acceptance and external-provider health are
recorded separately. TUI/autosave work remains paused and excluded from this fix.
