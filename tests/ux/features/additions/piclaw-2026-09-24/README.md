# Single-user passkey Settings — additional contract

Copied verbatim from Piclaw commit `b531ea3a8cb38ba8f1e0729f8b6d0439f63cb5f1` (2026-09-24), `tests/e2e/features/shared/single-user-passkey-settings.feature`.

SHA-256: `bd48cab9126778763cab3ddfd8ee04a89bd8c29da62e3bc4188cf24dc3a55b82`.

The 26 scenarios/outlines describe **upstream** implementation and browser evidence, not Gi support. Gi now has an opt-in native WebAuthn store, one-use challenges, registration/assertion verification and multi-key management APIs. Its Chromium virtual-authenticator tests now also drive Settings enrolment/management and passkey login. All 26 complete scenarios remain unmapped pending a per-case audit; Visual skin, physical devices and policy controls are not covered. The upstream `@implemented` and `@browser-verified` tags are retained as source text, not Gi credit. See the [backend contract](../../../../../docs/internal/passkeys.md).

This additive contract does not replace or expand the frozen 236 Classic / 42 shared inventory. Track it separately until a runner and per-case Gi evidence exist. Upstream design and evidence are at the same commit:

- `docs/design/single-user-passkey-settings.md`
- `docs/reviews/single-user-passkey-settings.md`

Implementation prerequisites: persistent per-credential IDs/public keys/counters/names; RP/origin-bound, expiring one-use registration/assertion challenges; five-minute recent-auth proof; atomic lockout-safe remove/recheck; owned rename/delete; duplicate/uncertain-registration recovery; policy-aware login; accessible browser-owned prompt cancellation/focus; no secret-bearing browser storage. Removal must block future credential assertions without silently revoking existing sessions. Physical/synced authenticators remain manual-device evidence.

The browser gate now offers accepted TOTP and configured passkeys, and Settings > Authentication manages credentials with explicit proof, cancellation and error recovery. Tests verify the Classic host with Chromium virtual authenticators; native physical prompt behaviour remains separate. The local TUI does not use HTTP authentication: do not add idle login rows or passkey-management chrome. Any future terminal action should open the authenticated browser Settings, not emulate browser WebAuthn or bypass recent proof.

## Required multi-passkey acceptance

The owner explicitly requested working multi-passkey enrolment on 2026-09-24.
This is a release requirement for the passkey feature; the existing TOTP/storage
work does not satisfy it. The feature file and its upstream tags remain unchanged.

- Drive Settings with two independent browser virtual authenticators, each with
  its own private key. Use real `navigator.credentials.create/get` ceremonies and
  native Go verification; mocked success responses cannot establish enrolment.
- Add the first credential after recent TOTP proof, then add another after recent
  passkey proof under passkey-only policy with no TOTP configured. Adding either
  must preserve the other's credential ID, public key and name.
- Restart the server and sign in from a fresh cookie-free browser with each
  authenticator independently. Only the credential used gets a new last-used
  timestamp. Also restart during a pending ceremony and fail closed or recover
  its exact session binding; never accept a challenge twice.
- Rename one credential, cancel removal, then remove it. A fresh sign-in with the
  removed key must fail; the remaining key must still work. Existing sessions
  retain their normal lifetime. Concurrent removal of the last two usable keys
  must leave one usable credential, with policy rechecked inside the transaction.
- Bind writes to the authenticated browser session and factor proof within five
  minutes. Automation tokens, proof from another browser, stale or revoked
  sessions, wrong origin/RP, expired/replayed challenges and duplicate credential
  IDs must fail without changing credentials.
- Exercise native cancellation, explicit retry, failed finish after local
  credential creation, and a lost successful response. Refresh the authoritative
  list; do not repeat a consumed ceremony or claim local authenticator cleanup.
- Verify accessible list/add/rename/remove controls, long literal-text names,
  truthful errors and focus behaviour at narrow and desktop widths. Keep drafts
  and media unchanged. Gi's Classic host and upstream Visual skin evidence remain
  separately scoped.

Chromium virtual-authenticator evidence must be labelled as such. Actual synced
credentials, platform/security-key prompts, cross-device enrolment and native
prompt focus need separate manual-device results. One synced credential is one
registered credential, regardless of how many devices hold it.

## Origin and remote-access prerequisite

Use a configured, stable RP ID and explicit HTTPS origin allowlist. Never derive
management authority from arbitrary Host or forwarded headers. The intended
production hostname must be settled before owner enrolment: localhost and a
future tsnet hostname generally cannot reuse the same RP-scoped passkey. Tests
use a disposable origin/account; they must not enrol keys or change auth policy
on the operator's live instance. Remote access remains opt-in and cannot bypass
the browser auth or recent-proof requirements.
