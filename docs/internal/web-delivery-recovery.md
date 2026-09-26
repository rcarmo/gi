# Basic HTTP delivery and run-control checks

## Lost acknowledgement

A native prompt keeps running after the browser disconnects. Previously, losing
its successful POST reply restored the accepted prompt above newer draft text,
even when its response was already visible. Pressing Return again could submit
it twice. The new regression reproduced that behaviour in Chromium and WebKit.

Every `sendAgentMessage` now carries a random client request ID, including idle
sends. On a transport `TypeError`/`AbortError`, the adapter performs read-only
reconciliation before reporting unknown delivery. It accepts exactly one
same-session turn with that ID and a matching `turn.submitted` event, written
after native admission's rollback boundary. No text matching or POST retry is
used. Each read has a3-second timeout. Confirmed admission acknowledges only its
captured draft; newer typing survives and no duplicate prompt is restored.

Absent, duplicate, foreign, malformed or unaudited receipts and failed reads
remain unknown. HTTP error responses are not reinterpreted as success. The
existing warning keeps text available and tells the user to check the timeline;
unknown delivery is not permission to resend.

### Limits

This recovers the live page's same-session lost HTTP reply, not general exactly-
once delivery. Routed turns admitted in another session are not inferred from
source-session data. Explicit peer-target submissions currently do not carry
this request token into the target turn. They may remain unknown after a lost
reply. A review identified that gap; no full routed-delivery acceptance is
claimed. Closing/reloading before recovery completes still uses the older
pending-draft warning/recovery path and is separate work. Steering receipts
without a follow-on turn also remain unknown. No automatic replay is added.

## Human-expectation acceptance

All tests use disposable state and a real non-loopback HTTP origin, without
secure-context or crypto overrides. Both Chromium and WebKit are covered.

`make test-web-basic-send` now runs14checks:

- Return and Send create one exact native turn and matching visible response;
  reload preserves both and a second message works.
- HTTP503/pre-admission network failure retain text and show honest feedback;
  explicit retry creates one turn.
- Attachment selection, upload, send and reload preserve exact native file
  bytes and metadata; accepted references leave the composer.
- A held acknowledgement shows Sending while newer typing remains usable.
  Aborting the response after native success performs read-only reconciliation,
  preserves only the newer draft, and never sends a second POST.
- Failed recovery reads remain unknown with one accepted turn, no auto retry.
- New session, child send, parent draft return/reload and child history return
  preserve the correct session and content.

`make test-web-basic-controls` now runs18checks (three journeys × two engines ×
three sizes):

- Stop posts the exact active turn ID to the native activity endpoint, cancels
  that turn without creating another, preserves the next draft through reload,
  and permits the next send to complete.
- Severing real SSE sockets shows Reconnecting, keeps the draft, then restores
  the native completed response after reconnection with one turn and no resend.

- Model selection persists without submitting the draft; reload keeps the
  selection. Subsequent native turns use the selected model, proven by both
  turn metadata and the deterministic provider's actual request-body model.
  A503selection failure keeps the visible label, stored choice and draft;
  reload and the next send still use that prior model. A later successful
  change reaches the other model. Each case creates exactly three turns.

These checks are included in required basic-web CI, gating all platform builds.
The provider is deterministic; live Copilot availability was verified separately
for the preceding HTTP-send fix, not rerun for every case here.

## CI cancellation-test follow-up

Product CI36232554338 failed1of189passkey cases: the login cancellation test
waited100ms after the busy Cancel button appeared. The trace showed an aborted
`/login/start`, then409 `Authentication state changed; retry`, rather than a
cancelled WebAuthn prompt. The button is visible before the server start finishes.
The test now observes the real forwarded `navigator.credentials.get` invocation
before cancelling; no credentials, signals, requests or validation are replaced.
Nine repeat cases and the full189case matrix pass, plus core/vet/hooks and121
functional tests (11existing skips). Review confirmed the boundary is appropriate
for this scenario. Rapid cancellation during the server start and immediate
retry remains a separate UX gap. No physical OS-sheet or race-fix claim is made.
The failed run/trace are retained and no timeout was increased.

## Verification and exclusions

Core/vet/hooks and13helper tests/50assertions pass. Full isolated functional
suite:121passed,11existing skips. Initial attachment-selector and native-event
field mistakes were corrected before the accepted run. The SSE assertion now
uses the actual Reconnecting label rather than an obsolete offline class; its
native completion/no-replay checks were retained. No test removed, timeout
increased, pixel criterion relaxed or supplied component modified.

Live remains the preceding exact HTTP-send release until this follow-up has
whole product CI and a separately recorded deployment. TUI/autosave work stays
paused. Routed/lost-reload delivery and remaining weak post-count tests outside the
chat-flow suite remain on the basic-workflow backlog.

The model follow-up also replaces the original chat-flow suite's broad post
counts and fixed sleeps with unique per-test prompts, exact native admission
IDs, one-new-turn checks, matching user/assistant messages and reload proof.
Avatar and cleared-composer assertions now refer to the new admission rather
than existing history. All18control/model cases and121functional checks pass
(11existing skips), plus core/vet/hooks and13helpers/50assertions. A focused
review found one missing immediate label assertion after failed model selection;
that assertion was added and passed. No runtime production code changed in this
acceptance follow-up; the two-model provider fixture is enabled only under
`GI_UX_BASIC_HTTP`.
