# Native multi-passkey backend

Status: opt-in backend and browser API integration. Settings controls, passkey
login UI, policy controls and physical-device validation are not implemented.
No complete additive passkey scenario is mapped yet.

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
closed. There is no public policy setter yet. Tests modify only disposable
stores to exercise passkey-only and commit-time policy changes. Operators should
not hand-edit production auth state to activate an unfinished UI workflow.

## API

All routes are under `/api/auth/passkeys` and return `private, no-store`.

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

## Test scope and remaining work

`make test-ux-passkeys` runs four integration tests at three Chromium viewport sizes
and is required by CI before build jobs. The tests cover two distinct credentials,
server restart, independent cookie-free sign-ins, passkey-only further enrolment,
rename material preservation, removed-key rejection, last-factor refusal, replay,
wrong session, revoked session, stale proof/reauth, cancellation, expiry, altered
origin, correctly signed wrong-RP assertion, duplicate-ID protection and uncertain
successful registration. Fixtures seed specific policy/error states explicitly.

Go tests cover config, fresh proof and RP/policy checks, ceremony bounds/consumption,
concurrent removal, native route authority and failed-write preservation. Existing
TOTP/browser suites remain separate. CDP supplies virtual authenticator operations;
WebKit and physical native prompts are not covered by this passkey suite.

The 26 pinned [Settings scenarios](../../tests/ux/features/additions/piclaw-2026-09-24/README.md)
include controls, focus, user-facing errors and manual-device cases that these API
tests do not establish. Next work is the Settings pane and login UI, safe policy
configuration, cancellation/uncertain-result presentation and device validation.
Choose a stable production HTTPS hostname/RP before enrolling real credentials.
