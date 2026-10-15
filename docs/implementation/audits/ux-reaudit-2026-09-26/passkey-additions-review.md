# Additive passkeys: 26 clause findings

The 26 frozen [additive scenarios](../../../../tests/ux/features/additions/piclaw-2026-09-24/piclaw-single-user-passkey-settings.feature) have source-to-assertion findings, but no full-scenario mappings. The [criterion ledger](../../../internal/passkey-criteria.json) records each step, example and evidence anchor; the [scenario review](../passkey-scenario-review.md) names the tests and outstanding clauses. `make test-passkey-criteria` passed 8 tests and 1,661 assertions in this review. That checks ledger integrity and lifecycle helpers, not the browser journeys. The earlier 189-execution passkey browser result is reported in the scenario review; no new browser run or current Piclaw/deployed-Gi test took place here.

## Mounted route and test scope

`web/src/gi-settings.ts` mounts the Authentication pane; `web/src/gi-settings-authentication.ts` reads the current owner's inventory and drives registration, reauthentication, rename, removal, policy and refresh. `web/src/gi-passkeys.ts` owns browser ceremonies. `internal/web/server.go` registers the `/api/auth/passkeys` list, register, rename and remove endpoints; `internal/web/auth_passkeys.go` applies cookie, origin, proof and mutation checks, with credential/factor state in `internal/auth/passkeys.go`. `tests/ux/passkeys.spec.mjs`, `tests/ux/auth.spec.mjs` and the native Go tests cited by the ledger exercise those paths. The ledger's evidence anchors and actual assertions were spot-checked against these mounted paths. Test declarations are not themselves scenario passes.

All browser ceremony evidence uses Classic Settings, HTTP localhost and Chromium CDP virtual authenticators. The frozen Background requires `https://piclaw.test` and RP `piclaw.test`. Visual Settings, real native prompts, synced/physical devices and production RP choice lack acceptance here. A *candidate* has substantive bounded automated assertions; it is neither a formal mapping nor an expanded-case pass. In particular, the feature's `@browser-verified` tag does not override these limits.

| ID | Bounded assertion and remaining clause | Finding |
|---|---|---|
| 001 | Inventory metadata, controls and secret absence checked in Classic; Visual outline row and a nonempty enrolment-token canary untested. | Partial |
| 002 | Held finish gates first-key success; fresh TOTP sign-in and no chat/URL enrolment checked. | Candidate |
| 003 | Distinct second key and unchanged first checked under passkey-only policy; TOTP removal seeded and authenticator selection virtual. | Candidate |
| 004 | Both keys sign in with fresh cookies after a disposable server restart; only used key's last-used time advances. | Candidate |
| 005 | Credential-per-row data model does not test one synced key across two real devices. | Manual |
| 006 | One of two unnamed records renamed by ID; reload and full credential-record equality checked. | Candidate |
| 007 | Blank, whitespace, long, control and literal-markup inputs checked; same credential remains usable. | Candidate |
| 008 | Named confirmation and held remove prevent optimistic success; removed key rejected, surviving key signs in. | Candidate |
| 009 | Cancel and Escape send no removal POST, preserve both records and restore the opener's focus. | Candidate |
| 010 | App-owned abort has no finish/replay; Add returns after explicit Refresh. Native prompt Cancel and immediate retry availability untested. | Partial |
| 011 | Synthetic blur leaves a ceremony pending; actual prompt focus transfer followed by successful single completion untested. | Manual |
| 012 | Exclusion list and crafted verified duplicate reaching 409 preserve first record; physical duplicate selection untested. | Candidate |
| 013 | List, rename and remove server/network faults preserve confirmed state; explicit Refresh precedes retry. | Candidate |
| 014 | Lost post-commit finish reconciles on pane return or Settings reopen through held inventory GET, with no registration replay. | Candidate |
| 015 | Insecure non-localhost host is stopped at login, before the frozen Settings journey; unavailable-API overrides are not old-browser evidence, and the unconfigured example lacks direct WebAuthn-call instrumentation. | Partial |
| 016 | Classic 390px keyboard/touch and semantic alert coverage; Visual and actual assistive-technology announcements untested. | Partial |
| 017 | Owner authority and first five refusal examples have native/API checks; family-shared mode example unsupported. | Partial |
| 018 | Seven invalid finishes reject without leaking proof and retain old credentials; revocation follows credential creation, not physical prompt opening. | Partial |
| 019 | Three stale-proof writes, cancel and later explicit proof tested; physical OS cancellation untested. | Candidate |
| 020 | Six last-factor examples and UI reasons covered; pending TOTP enrolment on an existing owner unsupported. | Partial |
| 021 | Concurrent loser may receive writer-lock conflict before factor guard; a later explicit retry tests last-factor refusal. | Partial |
| 022 | Policy change between confirmation and commit rejects removal; UI refreshes the reason. | Candidate |
| 023 | No mounted legacy slash-delete command or Settings guidance for it. | Unsupported |
| 024 | Removal leaves sessions usable until logout or fixture-accelerated expiry; no twelve-hour wall-clock wait. | Candidate |
| 025 | Failed finish explains possible local orphan and no server registration; direct fixture has zero existing rows, so it does not independently prove preservation of existing keys. | Candidate |
| 026 | Fresh proof in one browser leaves the second session's direct writes forbidden until its own proof. | Candidate |

The ledger's `formalMapping` values remain false. The 56 expanded passkey cases are not counted as passed. Current Piclaw runtime, deployed Gi and physical-device acceptance have not been compared. The 26 inventory findings record scope and gaps without altering the frozen feature or Classic/shared mapped-ID counts.
