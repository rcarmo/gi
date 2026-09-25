# Browser-owner session proof

Gi distinguishes ordinary API authentication from authority to manage the owner's
credentials. This prerequisite implements session provenance and recent TOTP
proof. WebAuthn verification and passkey management are not implemented yet.

## Session records

`POST /api/auth/session` verifies the TOTP code and issues the existing private
HttpOnly, SameSite Strict cookie. Its persisted session now also records
`purpose: "browser-owner"`, `auth_factor: "totp"` and `authenticated_at`.
`POST /api/auth/totp/verify` and native `VerifyLogin` retain ordinary bearer-token
semantics, without owner-management purpose. Copying such a token into the
`gi_session` cookie does not elevate it.

Existing sessions with missing purpose remain usable for ordinary application
access. They cannot use owner-proof routes: sign in through the browser endpoint
to create a session with provenance. Unknown purposes and factors fail closed.
An older binary that rewrites session objects may discard these new fields;
returning to this version then requires a new browser login for owner proof.
No authority is inferred from `created_at` or the token's delivery mechanism.

## Routes

| Route | Behaviour |
|---|---|
| `GET /api/auth/session/proof` | Returns advisory `reauth_required`, `authenticated_at`, `fresh_until` and `expires_at` for the current browser-owner session. It does not modify storage. |
| `POST /api/auth/session/reauth/totp` | Accepts `{"code":"123456"}`; validates the existing browser-owner session and TOTP under the auth writer lock, then refreshes only that session's proof. |

Both routes require exactly one `gi_session` cookie, direct TLS or a loopback
peer and Host, and the existing same-origin checks. Any explicit Authorization
header or `auth_token` query parameter is rejected, even with a valid cookie.
Forwarded headers do not establish authority. Responses are `private, no-store`.

Reauthentication accepts one JSON object of at most 1024 bytes; unknown fields and
trailing JSON are rejected. It creates no new session, sets no cookie, changes no
expiry and does not refresh another browser's proof. Failure leaves auth state
unchanged. HTTP statuses distinguish missing/expired/legacy sessions (401),
transport/origin or explicit credential rejection (403), state conflicts (409),
and storage failures (500). Storage errors return a generic message.

Recent proof lasts strictly less than five minutes and never outlives the login
session. Missing, future or pre-session timestamps are not fresh. TOTP must still
be enabled and configured; unsupported factor strings do not grant freshness.
Only the server supplies proof timestamps. Loading Settings or reloading a page
cannot refresh them.

## Future credential mutations

The proof response is advisory. A future add/rename/remove operation must invoke
`requireBrowserOwnerSession` with `fresh=true` against the same `State` snapshot
being changed inside `updateState`. This rechecks revocation, expiry, purpose and
factor proof at write time. It must also check the current RP/origin/login policy
and last-usable-factor rule in that transaction. None of those future WebAuthn
checks can be replaced by a successful proof-status request.

Tests cover independent sessions, exact five-minute boundaries, legacy/bearer
rejection, disabled factors, replay of revoked session tokens, restart recovery,
failed-write bytes and concurrent reauthentication/revocation. Browser tests
sign in normally in separate contexts, then invoke the proof APIs using their
real cookies while preserving drafts and attachments. They do not simulate a
passkey ceremony or establish any of the additive passkey scenario mappings.

The terminal does not acquire owner proof or gain new login rows. Passkey
management will belong in authenticated browser Settings.
