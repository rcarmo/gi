# Classic auth 001, 002, 004: scope and clause review

The frozen Classic auth Gherkin includes family-shared and single-user modes.
Gi currently exposes **single-user** browser authentication only:
`internal/web/auth.go:46–72` returns `mode: "single-user"`, and
`web/src/gi-auth-policy.ts` rejects another mode. `docs/feature-parity.md:50`
also lists family mode as absent. An installed Piclaw 3.2.4 login-bundle
fixture now covers policy/form display and request shape; it does not
establish family-account authorization or real authentication.

| Frozen ID | Native requirement → assertion → code | Disposition |
|---|---|---|
| `@ux-auth-001` family username + TOTP | Piclaw's frozen scenario requires a visible required username, trimmed username in verification and a network-error state. Gi's native policy parser rejects family mode; the server never advertises it. No `@ux-auth-001` browser test exists. | **Verified native family-mode gap.** Do not borrow the `002` code-only login test. |
| `@ux-auth-002` single-user TOTP | `tests/ux/auth.spec.mjs:133–169` loads the real native policy, verifies code-only sign-in and no username requirement, invalid-code/network errors, disabled pending controls, authenticated cookie, session-expiry re-gate and draft retention. Gi `web/src/gi-auth.ts`, `gi-auth-policy.ts`, `internal/web/auth.go`, `auth_session.go` own the route. | **Bounded native journey.** It does not exercise a Piclaw login backend, account-family scope or physical authenticator. |
| `@ux-auth-004` failed policy load and Retry | `auth.spec.mjs:170+` fails the first `/api/auth/status`, checks credentials stay hidden and Retry appears, then returns valid single-user policy and checks the code field with no auth write. Gi `gi-auth.ts` fetches no-store, validates policy, and gates the app. | **Bounded native journey.** Other malformed policy fields and false-success response checks run in separate untagged native tests; a tag alone cannot import them. |

`make test-piclaw-login-policy` passed **6/6** Chromium/WebKit phone, tablet
and desktop browser cases using installed 3.2.4 login assets and disposable
`/auth/options` and `/auth/verify`: family username normalisation, single-user
code-only request, passkey-only form hiding, and failed-options Retry. It did
not verify a real code, passkey ceremony, authenticated session or family
account. A focused native Gi run of `@ux-auth-002|004` plus the untagged
passkey-only mounted-gap check passed **18/18**. The family mode and
passkey-only form differences remain. The full `make test-ux-auth` previously
stopped after 105/120 tests; that run was not green.
