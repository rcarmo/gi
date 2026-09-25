# Browser-bound initial owner setup

Gi has a browser-bound TOTP setup API for the first owner. Settings controls are
not implemented yet. The existing `/api/auth/enroll/start` and `/verify` API
remains available with its existing behaviour.

## HTTP boundary

All three new routes require POST, `application/json` and one JSON object of at
most 1024 bytes. The actual peer and Host must both be loopback, including for
TLS requests. `Origin` must match the request's scheme and Host exactly; foreign
Fetch Metadata, Authorization headers (including empty ones) and any nonempty
query string are refused. Forwarded headers cannot establish loopback authority.
Remote HTTPS does not enable this setup API.

| Route | Body | Result |
|---|---|---|
| `/api/auth/setup/start` | `{}` | Returns `username: "admin"`, the TOTP `secret`, `url` and `expires_in_seconds: 600`; sets a setup cookie. |
| `/api/auth/setup/finish` | `{"code":"123456"}` | Verifies the code and atomically creates the owner and browser-owner session; returns only `ok:true`. |
| `/api/auth/setup/cancel` | `{}` | Discards this binding and clears the setup cookie. It cannot delete a configured owner. |

The 32-byte random setup token is hex-encoded in `gi_setup`, a host-only,
HttpOnly, SameSite Strict cookie scoped to `/api/auth/setup`. It is Secure on TLS
and expires after ten minutes. Finish and cancel require exactly one 64-character
setup cookie; start permits no cookie or one previous binding. Duplicate and
malformed cookies are rejected. Setup tokens are never returned in JSON.

Responses use `Cache-Control: private, no-store`. Only a successful start returns
the secret and `otpauth:` URI. They are inputs for the owner's authenticator and
must not enter chat, logs, persistent browser storage or terminal output.

## Binding and lifetime

`Manager.browserPending` stores at most eight pending setups per process, indexed
by token hash. The pending secret is memory-only until verified. Starting again
with an existing binding discards that binding first; cancellation or expiry
frees its slot. Restart discards all pending setups. An unknown token cannot
consume another browser's setup.

Every finish that reaches `FinishBrowserSetup` consumes its binding before code
verification or storage work. An incorrect six-digit code, expired binding,
conflicting owner or write failure requires explicit new setup. HTTP admission
failures, including malformed code bodies, do not consume the pending setup.
Cancel is idempotent for a well-shaped binding.

Start refuses existing usernames, configured or disabled TOTP secrets, enabled
TOTP, passkeys or sessions. Finish repeats that check against the current file
under the auth writer lock. Concurrent browser finishes and legacy enrolment
therefore converge on at most one owner. The winner has username `admin`, policy
`either`, verified TOTP and one twelve-hour browser-owner session with fresh TOTP
proof. Owner and session are stored together in `.gi/auth.json`; only the session
hash is persisted. Unrelated stored extension fields retain the normal state-file
preservation rules.

Finish clears the setup cookie after a parsed attempt. Only success issues the
host-only HttpOnly Strict `gi_session` cookie at `/` (Secure on TLS). Errors do
not issue a session cookie or return tokens, secrets or storage paths.

| Failure | HTTP status |
|---|---|
| Method, media type or malformed body | 405, 415 or 400 |
| Invalid origin, peer, Host or explicit credentials/query | 403 |
| Missing, duplicate or malformed setup cookie | 401 |
| Invalid code, expired/consumed/unknown binding | 400 |
| Owner already exists or writer contention | 409 |
| Eight live pending setups | 429 |
| Storage failure | 500, generic error |

A write can commit before durability confirmation fails or its HTTP response is
lost. The caller must read `/api/auth/status` before starting again. If the owner
now exists without a usable browser cookie, normal TOTP sign-in uses the verified
secret. Never automatically repeat finish or assume an error means no owner was
created. This slice does not implement that Settings recovery UI.

## Verification and limits

`browser_bootstrap_test.go` and `auth_bootstrap_test.go` cover binding isolation,
capacity, cancellation, expiry, restart, partial/corrupt state, write failure,
owner/session issuance, stale finishes, concurrent browser and legacy enrolment,
origin/transport/body guards and cookie scope. Refused finishes preserve existing
auth-file bytes and return no authority.

`tests/ux/auth.spec.mjs` exercises browser-origin fetches in Chromium and WebKit
at three viewport sizes. It verifies HttpOnly cookie behaviour, fresh proof,
other-browser rejection, persistent owner authority after restart, cancellation,
replay and draft retention. The fixture starts unenrolled and uses only disposable
stores. No Settings setup journey, QR scanner, physical authenticator or passkey
bootstrap is accepted by these API tests. The frozen passkey criteria and formal
mapping counts are unchanged.

This interface trusts the local machine boundary. It does not defend against a
hostile local process capable of forging HTTP headers or accessing the auth
store. Keep an unenrolled instance on loopback or a protected network; broader
application access before enrolment is unchanged.

The TUI gains no login row, secret display or setup state. Initial owner setup
belongs in the browser; any later terminal entry point should be a transient
browser hand-off within the existing status/command model.
