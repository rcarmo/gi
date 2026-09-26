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

## Reopening before acknowledgement

New sends carry the persisted draft-capture token as `client_request_id` for
both idle and queue modes. A guarded build adapter changes only that payload
expression; supplied composer source remains untouched. Older idle captures may
have a different request ID and are not inferred to have succeeded.

On startup, the draft repository asks a read-only recovery helper to check up
to six pending captures with two workers and one3-second network deadline.
Session turn snapshots are shared across captures in the same session. Only an
exact same-session token match plus the matching post-rollback submission event
confirms a capture. Confirmed text/attachments are retired before unknown
captures merge into newer draft content. Cleanup persistence is awaited; storage
failure rejects repository loading and the host surfaces a recovery error.

Actual HTTP proxy tests hold the single native response after admission, close
the tab, and reopen the same browser storage. They prove newer text survives,
accepted attachments are not restored, repeated reload does not duplicate, and
there was one native POST/turn. A failed receipt lookup instead restores unknown
text once with a warning and makes no automatic POST. An initial Playwright
`route.fetch` version was discarded: closing the paused page could release the
original POST after forwarding its clone, creating two native turns. The real
proxy forwards once, and the exact-one assertion remains unchanged.

### Limits

This recovers the live page's same-session lost HTTP reply, not general exactly-
once delivery. Routed turns admitted in another session are not inferred from
source-session data. Explicit peer-target submissions currently do not carry
this request token into the target turn. They may remain unknown after a lost
reply. A review identified that gap; no full routed-delivery acceptance is
claimed. Closing/reloading with a new same-session capture now attempts the bounded
startup check above. Legacy/mismatched captures, captures beyond the six-check
budget, failed/late receipts and cross-tab concurrent draft writers remain
outside confirmed recovery; they retain the existing unknown-delivery fallback.
The cap bounds request count/concurrency, not the size of the native turn-history
response. A dedicated bounded receipt endpoint is still future work. Steering receipts
without a follow-on turn also remain unknown. No automatic replay is added.

## Human-expectation acceptance

All tests use disposable state and a real non-loopback HTTP origin, without
secure-context or crypto overrides. Both Chromium and WebKit are covered.

`make test-web-basic-send` now runs18checks:

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
- Closing before the accepted reply and reopening preserves newer text without
  restoring accepted attachments; failed startup reads keep unknown text once.

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

## Deployment

Exact `003b9e669682b3aafda9621ec976fbb12ac4ddb8` passed whole product
[CI36233054266](https://github.com/rcarmo/gi/actions/runs/36233054266), including
all14basicHTTP journeys,18control/model checks and four platform builds. It
contains the delivery fix from `acb069e`, whose own CI36232554338 failed the
prompt-cancel case above; that failed run was not used as deployment evidence.
The later cancellation test correction `5fa51a6` is excluded from this release.

Detached-source Makefile deployment runs on8090, PID1403292, PGID/SID1403282.
Four blocked-send HTTP probes (Chromium/WebKit×Return/Send) reach the exact
native endpoint and preserve unknown-delivery feedback/drafts without native
writes. Six read-only UI probes pass. DB62sessions/51turns/146messages,
integrity/FKs, normalised SQL excluding runtime leases and absent auth remain
unchanged. Existing tabs need reload. Lost-success/model/Stop/reconnect acceptance
uses the isolated native fixtures; no live chat mutation is claimed.

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
that assertion was added and passed. No runtime production code changed in that model-acceptance follow-up; the
two-model provider fixture is enabled only under `GI_UX_BASIC_HTTP`.

Closed-page follow-up:18HTTP journeys,18controls/model cases,17helpers/68assertions,
core/vet/hooks and125functional tests (11existing skips) pass. Adapter tests
reject changed/duplicate source anchors; helper tests cover exact token/session,
partial/failed proof, cached reads/concurrency cap and persistence failure.
Two delegated review attempts timed out and supply no independent approval.
Initial helper-name and proxy-harness mistakes are retained as failed evidence;
no assertions, timeouts or security gates were relaxed. This follow-up awaits
its own whole product CI and deployment; live remains003b9e6.
