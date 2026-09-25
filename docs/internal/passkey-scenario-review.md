# Multi-passkey scenario review

Date: 2026-09-25. Scope: the separately pinned26-scenario Piclaw contract, after
native WebAuthn, Classic Settings/login and sign-in policy controls. This is a
source-to-test review, not an automated per-case pass report. No new mappings
are added by this document. Frozen Classic/shared counts are unchanged.

The current passkey suite has 42 tests in Chromium across three viewport projects
(126 executions). Two narrow-interaction tests explicitly create390px touch-capable
contexts in every project; those six executions repeat390px coverage rather than
establishing tablet/desktop geometry. CDP virtual authenticators sign real browser ceremonies on
HTTP localhost, not the frozen Background's `https://piclaw.test` origin.
Physical/native focus, synced credentials and the Visual skin are not verified.
The auth regression runs Chromium and WebKit but does not establish WebKit
passkey ceremonies. Candidate rows below are bounded automated evidence, not
full-contract passes or formal mappings.

Abbreviations: **UI** = `tests/ux/passkeys.spec.mjs`, **Auth UI** =
`tests/ux/auth.spec.mjs`, **Native** = `internal/auth/passkeys_test.go`,
**Policy** = `internal/auth/login_policy_test.go`, **HTTP** =
`internal/web/auth_passkeys_test.go` / `auth_policy_test.go`.
The full source contract and hash are linked in the [addition README][addition].

"Candidate" means source review found coverage for the substantive automated
journey, pending a dedicated criterion mapping. "Partial" means a named step or
outline example lacks direct evidence. A physical-device gap stays manual.

| ID suffix | Current test evidence | Remaining gap / classification |
|---|---|---|
| 001 | UI `Classic Settings shows native credential metadata…`: named used/unused credentials, full IDs and creation/use dates matched to native records, Never used label and enabled controls. Actual cookie, TOTP secret, user handle, public keys and token hashes absent from DOM/storage/URL and inventory JSON. | Partial: Classic verified; Visual skin absent. No synthetic enrolment-token canary is created by this flow. |
| 002 | UI first Settings enrolment after TOTP login; native verification and persisted list. | Partial: explicitly assert retained TOTP login and no enrolment token in URLs/chat. |
| 003 | UI passkey-only reauth and further enrolment with no TOTP; distinct virtual credential material. | Partial: narrowly map one-existing-key to second-key journey and no chat-link prompt. |
| 004 | API and UI two-key sign-in after server restart with fresh cookies; only used key's last-used time advances. | Candidate: both passkey examples, virtual-authenticator scope only. |
| 005 | Data model stores credentials, not physical devices. | Manual: one synced credential used on two physical devices, one row retained. |
| 006 | UI `Settings renames one of two unnamed credentials…`: two canonical empty names seeded in the disposable store; identify one row by credential ID, save Tablet, compare full credential records and reload Settings. | Candidate: bounded Classic/localhost evidence; all authentication material and the other unnamed row remain unchanged. |
| 007 | UI `Settings rename validates…`: separate blank, whitespace,81-code-point, control and literal-HTML cases. Exact request/error or saved text, full record equality before sign-in, then cookie-free sign-in with that same authenticator and persisted name. | Candidate: all name examples through Classic; literal markup neither renders nor executes. |
| 008 | UI confirmation/remove, native removed-key rejection and remaining-key sign-in. | Partial: confirmation identifier and zero removal request before confirmation. |
| 009 | UI `Settings removal cancellation sends no request…`: pointer Cancel and Escape preserve both stored records and restore the identified opener. Zero remove requests before/after cancellation and refresh; positive control observes one confirmed removal. | Candidate: bounded Classic cancellation journey; no request-count credit inferred from inventory alone. |
| 010 | UI Escape/app cancel, no finish/auto-repeat, unchanged list and explicit new attempt. | Partial: application AbortSignal cancellation is covered; physical native-prompt cancel behaviour needs manual testing. |
| 011 | Synthetic blur during pending UI ceremony leaves it pending. | Manual: real OS/browser focus transfer and successful completion without duplicate prompt. |
| 012 | Browser creation exclusions and independently enforced duplicate finish409; stored keys unchanged. | Partial: full duplicate path via Settings, rather than native API integration only. |
| 013 | UI `Settings … failure retains confirmed state…`: list/rename/remove each fail with503 and pre-delivery network abort. Role alerts, last-confirmed snapshot, no success/empty list, disabled writes and unchanged native state; explicit Refresh then deliberate native retry succeeds. | Candidate: six Classic fault/recovery cases; no post-commit lost-write inference from pre-delivery faults. |
| 014 | Lost successful finish, UI uncertain-result explanation, explicit list refresh adds one row, no second finish. | Candidate: authoritative reconciliation without blind replay. |
| 015 | UI `Settings unavailable…`: constructor/container/create/get overrides after native login and enrolment, plus native TOTP-only policy. Exact explanation, disabled controls, zero credential calls/auth writes and full state equality through refresh/re-entry; restored APIs/policy then real assertion and creation as positive controls. Auth UI separately covers unenrolled/unconfigured states. | Partial: actual insecure non-localhost origin remains untested; capability overrides do not establish old-browser compatibility or all outline examples in one mapped contract. |
| 016 | UI `Classic narrow passkeys avoid horizontal clipping…`:390px,80-code-point name, metadata/button/error bounds, associated input labels and accessible button names. Tab/Enter reaches rename/remove/cancel; trusted tap repeats those actions, without writes. Status/alert roles and held-read outside-focus retention checked. | Partial: Visual skin and physical assistive-technology announcements unverified. Touch path uses programmatic focus for the outside-pane focus guard only; action activation uses real taps. |
| 017 | HTTP no-cookie, bearer/query, foreign origin, other-account field and transport denial; session binding. | Partial: all operations/examples, including explicit family-mode denial, plus no inventory leakage. |
| 018 | UI `Settings native registration … failure hides proof…`: seven native finish failures (expiry, consumed, foreign session, altered origin/RP, invalid proof, revoked after creation). Retained records unchanged, old key signs in after every failure; actual secrets/challenges/proof blobs absent from DOM/storage/URL/native error. Foreign proof later succeeds in its own session. | Partial: revocation is after real creation but before finish delivery, not while a physical native prompt is open. Altered origin/RP cases prove rejection of tampered proof, not isolated verifier-check ordering. |
| 019 | UI `Settings cancelled reauthentication cannot authorise…`: separate add/rename/remove cases; reload/Refresh preserve stale proof, direct403, button/Escape abort real credential.get without finish or credential mutation; presence restoration cannot complete it. Third explicit ceremony refreshes proof without extending expiry and permits the intended change. Verification focus returns to the same logical action after rendering. | Candidate: all three operations with application cancellation in Classic/CDP; physical OS prompt interaction stays unverified. |
| 020 | Native current-RP/policy/factor removal matrix; UI last-key refusal. | Partial: active-session-only and pending-unverified-TOTP examples, exact UI explanation for each case. |
| 021 | UI `Two Settings views concurrently remove different keys…`: both POSTs held, no optimistic removal, then one200/one409; both refresh to the same surviving key, with no automatic retry. Explicit retry returns exact last-factor409; cookies/sessions survive and the remaining key signs in afresh. | Partial: the non-blocking auth writer lock can refuse the concurrent loser with a state-conflict409 before the factor check. The frozen scenario requires refusal by the write-time lockout check; the later explicit retry proves that check separately. |
| 022 | UI `Policy changes reject stale browser revisions…`: removal confirmation predates other browser's passkey-only change, removal refused at commit, refresh reconciles. | Candidate: direct UI policy/removal ordering plus native race tests. |
| 023 | No implemented legacy passkey-delete command workflow established. | Unsupported: explicit command refusal/Settings guidance test. |
| 024 | UI removal explains future sign-ins versus existing sessions; native session validity retained. | Partial: normal expiry and explicit logout lifecycle after removal. |
| 025 | UI failed finish says not registered, local credential may remain and Gi did not remove it; native authenticator/store checks. | Candidate: local-orphan explanation without false cleanup. |
| 026 | UI `Settings proof in one browser leaves another stale…`: two different owner cookies aged beyond five minutes; first reauthenticates and renames, second reloads with Add/Rename/Remove disabled. Direct register-start/rename/remove return recent-proof403 with whole auth-state equality. Second succeeds only after its own proof. | Candidate: direct isolated-browser and native write-boundary evidence, without production authentication changes. |

The first unnamed-credential run failed because a missing `name` property was
serialised as `""` on the next Go write. The fixture now uses canonical empty
names; full-record equality assertions remain. This is not a runtime repair.
No frozen feature bytes, mappings or terminal behaviour changed.

Initial owner bootstrap and production RP/domain selection are separate deployment
work. The operator's live instance has not been enrolled or had its policy changed.
Terminal adaptation stays browser-based: no WebAuthn emulation, credential secret
entry or extra idle terminal rows.

[addition]: ../../tests/ux/features/additions/piclaw-2026-09-24/README.md
