# Passkey login cancellation boundary

The login gate offers **Cancel passkey prompt** only while the browser credential
request is pending. It stays busy during the native HTTP start and finish, with
separate “Starting” and “Confirming” status text.

Previously the gate exposed Cancel as soon as the user clicked sign in. Cancelling
could abort the browser's start request while the server was still writing its
auth state. An immediate retry had produced a real 409 in earlier CI. The prior
prompt-cancel test correction waited for `navigator.credentials.get`; it did not
cover this earlier window.

## Ownership

`runPasskey` reports `starting`, `prompt` and `finishing`. Prompt ownership begins
after invoking the native credentials API and ends when its promise resolves.
The login gate keeps its existing single-flight controller throughout, and a
synchronous ref makes a stale rendered Cancel button inert after prompt ownership
ends. A cancelled/unmounted caller is checked after the start response and before
opening any browser ceremony. That abort check also applies to the shared helper's
register/reauth callers.

The gate still requires successful finish and an authenticated native status read
before mounting the application. No server route, auth-file writer, ownership
check, cookie policy, retry rule or timeout changed. No automatic auth POST replay
was added. Settings operation cancellation is unchanged and needs its own separate
boundary assessment. Native OS/browser prompt behaviour still needs physical-device
acceptance; these tests use Chromium's virtual authenticator.

## Verification

`make test-passkey-login-boundary` holds the first start either before forwarding
it or after the native server returns. At both boundaries the login action remains
disabled, no Cancel button exists and no credential request/finish occurs. Release
opens a real virtual-authenticator prompt; explicit cancellation restores focus.
An explicit retry succeeds. A held native finish keeps the application gated and
has no prompt-cancel action. Draft, authenticated cookie and exact start/finish
counts are asserted. All six cases failed on the old UI's premature Cancel button,
then passed with the fix, using the existing four-second assertions.

- `make test-passkey-cancel-repeat`: 27 passes (new boundary cases and existing
  Settings two-key/cancel journey, three sizes, three repeats).
- `make test-ux-passkeys`: 195 passes.
- `make test-ux-auth`: 120 passes across Chromium/WebKit and three sizes.
- `make test-ux`: 129 passes, 11 existing skips.
- `make test-passkey-criteria`: eight helpers, 1661 assertions, including late
  aborted starts for login/register/reauth, phase ordering, cancellation without
  finish/replay and surfacing a 409 without invoking credentials.
- `make test vet bun-checks`: passed. Supplied components/ui/panes are unchanged.
- Focused independent review found no blocker.

An initial combined repeat/full-suite command exceeded the outer shell budget at
case 85, without a reported test failure. The full suite passed when run separately;
no Playwright timeout or assertion changed. Evidence is retained in
`/workspace/tmp/gi-passkey-login-boundary`.

## Deployment

Exact source `a1be540836363b85168ce14002f66fd32e1cd76b` passed whole-product CI
[36241463237](https://github.com/rcarmo/gi/actions/runs/36241463237), including all
four release builds, then replaced the receipt-only `d62967f` process on port 8090.
Live PID `74444`, process group/session `74437`; executable SHA-256
`f3e528a037b89ca0c02677b8d194733338c87e11c3eb31e6993a7f437464b218`.

Four blocked non-localhost HTTP send probes and six read-only UI probes passed.
DB counts remain 62 sessions, 51 turns, 146 messages and zero active turns; integrity
and foreign keys are clean. Full SQL dumps including receipts match after removing
only the dispatcher lease. Auth hashes and isolated local-only TUI WIP HEAD/diff
match. No live auth changes or chat writes occurred; passkey mutation acceptance
comes from the disposable tests above. Reload existing tabs for the new bundle.

Deployment evidence: `/workspace/tmp/gi-passkey-deploy-a1be540`. Database dumps and
auth hashes are retained locally and excluded from the archive.
