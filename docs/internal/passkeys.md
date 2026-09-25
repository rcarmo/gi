# Multi-passkey authentication

Status: opt-in native backend, passkey login and Settings > Authentication.
Settings includes lockout-safe sign-in policy controls. Initial owner bootstrap
and physical-device validation are not implemented in the browser. The
[26-scenario review](passkey-scenario-review.md) records partial and manual gaps.

The backend uses `github.com/go-webauthn/webauthn` v0.18.2 for WebAuthn verification
with required user verification. Runtime code is Go; tests use Chromium's virtual
authenticator and real `navigator.credentials.create/get`, not fabricated success
responses. Linux/macOS amd64/arm64 and Windows amd64 builds use CGO disabled.

## Configuration and authority

Startup reads `.pi/settings.json`:

```json
{
  "passkeys": {
    "rp_id": "gi.example.com",
    "origins": ["https://gi.example.com"]
  }
}
```

Missing or invalid configuration disables passkey endpoints. The RP must be a
valid domain (not an IP or public suffix); localhost is allowed for isolated
browser development. Each origin must match the RP or its subdomain, contain no
userinfo/path/query/fragment, and use HTTPS except loopback development. Exact
origin strings include any port. Requests never configure the RP through Host,
forwarded headers or body fields. Host selects an already approved origin.

Management requires the [browser-owner session](browser-auth-proof.md) cookie and
recent proof within five minutes. List is permitted without recent proof. Any
Authorization header or `auth_token` query is rejected. Ceremony and mutation
POSTs require an exact browser Origin header, approved transport, JSON content
and bounded bodies. Automation/legacy tokens copied into cookies have no
management authority. The local terminal gains no extra controls or idle rows.

Persisted `login_policy` accepts `either`, `totp-only` or `passkey-only`. Empty
preserves TOTP and allows explicitly configured passkeys; unknown policy fails
closed. Settings > Authentication can change this policy after recent proof and
explicit confirmation. The current policy must accept that proof; for example,
a preceding TOTP login cannot authorise changes after switching to passkey-only.
Stored credentials and existing login sessions are not deleted or extended.

## Policy API

`GET /api/auth/policy` returns the effective policy, revision and booleans for
configured TOTP, usable passkeys and passkey origin configuration. It requires a
browser-owner cookie but not fresh proof. It works even when passkeys are disabled
or unconfigured; reads never renew authentication or modify the store.

`POST /api/auth/policy` accepts `{"policy":"passkey-only","revision":"..."}`.
It requires the same cookie/transport/origin boundary, an explicit Origin header,
strict JSON up to1024 bytes and fresh proof accepted under the current policy.
A legacy empty revision is read as `initial`; each successful change creates a
new opaque revision. A stale revision returns409. The server rechecks authority,
revision and usable target factors inside the same transaction used by credential
removal. An unsafe target returns409 without a write. A passkey from another RP,
a pending/unverified factor or an active session is not a usable sign-in method.

The target must accept verified configured TOTP or a registered credential for
the currently configured RP and origin. A policy/removal race cannot leave
passkey-only with no usable key. A lost response requires explicit refresh before
another attempt; no write is retried automatically. TOTP-only still shows the
configured RP's credential inventory but disables passkey mutations/ceremonies.

## Passkey API

These routes are under `/api/auth/passkeys` and return `private, no-store`.

| Route | Method and body |
|---|---|
| `/api/auth/passkeys` | GET; public metadata for current-RP credentials only. |
| `/register/start` | POST `{"name":"Laptop"}`; recent owner proof required. |
| `/register/finish` | POST `{"ceremony_id":"...","credential":{...}}`; native registration response. |
| `/login/start` | POST `{}`; returns allowed credentials and sets a short-lived HttpOnly ceremony cookie. |
| `/login/finish` | POST `{"ceremony_id":"...","credential":{...}}`; native assertion; issues the ordinary HttpOnly owner session cookie, never a token in JSON. |
| `/reauth/start` | POST `{}`; existing owner session required. |
| `/reauth/finish` | POST `{"ceremony_id":"...","credential":{...}}`; refreshes only that session's proof, without extending its expiry. |
| `/rename` | POST `{"id":"credential-id","name":"Tablet"}`; recent proof. |
| `/remove` | POST `{"id":"credential-id"}`; recent proof plus atomic last-factor check. |

Start responses contain a random `ceremony_id` and `options.publicKey`. The browser
must decode base64url fields before invoking WebAuthn and send the resulting
credential JSON to finish. There is no automatic finish retry. Error responses
omit credential material and underlying storage details.

## Storage and races

The private `.gi/auth.json` transaction stores a stable random user handle and up
to 32 credentials with RP, name, creation/last-used times and full verified library
credential records. Public lists omit keys, attestation and counters. Credential
IDs must be unique; duplicate finish never overwrites an existing record. Names
accept 1-80 Unicode code points with no control characters and must be rendered
as literal text.

Challenges expire after five minutes and bind operation, exact origin, configuration
fingerprint and either the owner session hash or a separate login-cookie hash.
Up to eight owner ceremonies and eight login ceremonies are retained. New login
starts evict only the oldest login ceremony, never an owner's registration or
reauthentication. This bounds storage; it is not a network rate limiter.

A finish consumes the matching challenge in a committed transaction before parsing
and verification. Wrong session bindings cannot consume another browser's challenge.
Invalid/expired proof, a crash after consumption or a later write conflict require a
new explicit ceremony. All owner, policy, current-credential and expiry checks run
again in the verification/commit transaction. Persisted challenges can survive a
restart within their lifetime. Revoking a session while a prompt is open prevents
registration; removing a credential prevents later assertions and invalidates
recent proof based on that credential without silently revoking existing sessions.

Removal counts only other credentials for the current RP, or verified TOTP accepted
by the current policy. The check and removal share the writer lock. Concurrent
removals cannot delete both last usable keys. Changing RP configuration does not
turn an account with stored credentials into an unenrolled, public instance.

The verifier checks signatures, RP/origin, challenge, user handle and user verification.
A reported signature-counter clone warning is rejected; authenticators that legitimately
use zero counters follow the library's handling. Physical synced credentials need their
own device results.

## Browser controls

The login gate uses server capability flags for TOTP/passkey buttons. A successful
finish must be followed by confirmed cookie authentication before the app mounts.
A failed or uncertain result offers explicit status refresh; no automatic ceremony
retry occurs. Missing browser WebAuthn/secure-context support has an explanation.

Authentication is a lazy Settings pane. It lists names, complete distinguishing
IDs, creation and last-used times, with Never used for unused credentials. TOTP or
passkey reauthentication refreshes the current session before changes. Rename and
removal remain inline, with explicit cancellation and focus return. Removal explains
that future sign-ins are blocked while existing sessions remain valid.

Native prompt work is abortable on explicit cancellation, pane change or close;
blur alone does not cancel it. Escape cancels local authentication work/confirmation
before closing Settings. Late completions cannot replace another pane. Reads and
writes keep the last confirmed list, disable mutations until refresh after failure,
and never repeat a consumed ceremony. If registration finish is rejected, the UI
explains that an unregistered local credential may remain in the authenticator.
Lost responses require an authoritative refresh before a new attempt.

## Test scope and remaining work

`make test-ux-passkeys` runs17 integration tests at three Chromium viewport sizes
and is required by CI before build jobs. The tests cover two distinct credentials,
server restart, independent cookie-free sign-ins, passkey-only further enrolment,
rename material preservation, removed-key rejection, last-factor refusal, replay,
wrong session, revoked session, stale proof/reauth, cancellation, expiry, altered
origin, correctly signed wrong-RP assertion, duplicate-ID protection and uncertain
successful registration. Fixtures seed specific policy/error states explicitly.

Thirteen tests drive actual Classic Settings/login controls: two-key enrolment,
independent sign-in after restart, retained drafts/media, rename/remove/cancel,
passkey-only reauth/add/login, failed reads/writes, uncertain finish and unmount
cancellation, policy changes, stale revisions and a policy change during removal
confirmation. Added cases verify one-of-two unnamed credential rename/reload,
blank/whitespace/overlong/control-name rejection and literal HTML, with fresh
passkey sign-in after each example. Cancellation counts zero removal requests,
restores the exact opener and has a positive confirmed-removal control.

A two-browser test ages both proof records, reauthenticates only the first and
renames there. Reloading the second leaves its controls disabled; direct add,
rename and remove requests fail without any auth-state change. Only its own
reauthentication enables a successful write. All fixtures are disposable.

The suite uses HTTP localhost, not the pinned feature Background's HTTPS host.
Synthetic blur tests application ownership only, not OS prompt focus. A separate
six-project auth regression verifies unavailable-state messaging in Chromium and
WebKit. No WebKit passkey ceremony or Visual-skin acceptance is claimed.

Go tests cover config, fresh proof and RP/policy checks, ceremony bounds/consumption,
concurrent removal, native route authority and failed-write preservation. Existing
TOTP/browser suites remain separate. CDP supplies virtual authenticator operations;
WebKit and physical native prompts are not covered by this passkey suite.

The 26 pinned [Settings scenarios](../../tests/ux/features/additions/piclaw-2026-09-24/README.md)
include Visual-skin, device and policy cases not established by the current tests.
Next work is completing the per-scenario gaps, initial-owner bootstrap,
full accessibility and physical/synced-device validation.
Choose a stable production HTTPS hostname/RP before enrolling real credentials.
