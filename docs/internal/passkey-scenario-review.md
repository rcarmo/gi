# Multi-passkey scenario review

Date: 2026-09-25. Scope: the separately pinned26-scenario Piclaw contract, after
native WebAuthn, Classic Settings/login and sign-in policy controls. This is a
source-to-test review, not an automated per-case pass report. No new mappings
are added by this document. Frozen Classic/shared counts are unchanged.
The [criterion ledger](passkey-criteria.md) records exact source steps/examples,
named assertion anchors and the gaps below; it does not award full mappings.

The current passkey suite has 62 tests in Chromium across three viewport projects
(186 executions). Two narrow-interaction tests explicitly create390px touch-capable
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
| 002 | UI `Settings first passkey waits…`: zero initial keys, recent TOTP proof, native finish held with no success/row/store mutation; release verifies registration, unchanged TOTP/session records, fresh cookie-free TOTP sign-in. Chat messages unchanged, POSTs limited to registration, captured challenge/ceremony/secret absent from navigation/chat. | Candidate: standalone Classic journey; enrolment-token flow is absent rather than simulated. |
| 003 | UI `Settings adds a second distinct key…`: exactly one existing key, native passkey-only policy, TOTP-disabled disposable precondition; fresh passkey login then second distinct key. Existing record/session unchanged, no TOTP controls/chat-link/extra POST, second key signs in. | Candidate: standalone Classic journey. Virtual credential selection and seeded TOTP removal do not establish physical authenticator selection or TOTP-removal UI. |
| 004 | API and UI two-key sign-in after server restart with fresh cookies; only used key's last-used time advances. | Candidate: both passkey examples, virtual-authenticator scope only. |
| 005 | Data model stores credentials, not physical devices. | Manual: one synced credential used on two physical devices, one row retained. |
| 006 | UI `Settings renames one of two unnamed credentials…`: two canonical empty names seeded in the disposable store; identify one row by credential ID, save Tablet, compare full credential records and reload Settings. | Candidate: bounded Classic/localhost evidence; all authentication material and the other unnamed row remain unchanged. |
| 007 | UI `Settings rename validates…`: separate blank, whitespace,81-code-point, control and literal-HTML cases. Exact request/error or saved text, full record equality before sign-in, then cookie-free sign-in with that same authenticator and persisted name. | Candidate: all name examples through Classic; literal markup neither renders nor executes. |
| 008 | UI `Settings confirmed removal rejects…`: confirmation name/ID, zero POST before confirm, held native removal with no optimistic success; session records and survivor unchanged. Client allow-list alteration produces a real removed-key assertion rejected400/no cookie/protected401; surviving key signs in200 with ID/name/RP/public key retained. | Candidate: complete bounded Classic removal journey; altered client options exercise native rejection rather than physical prompt affordances. |
| 009 | UI `Settings removal cancellation sends no request…`: pointer Cancel and Escape preserve both stored records and restore the identified opener. Zero remove requests before/after cancellation and refresh; positive control observes one confirmed removal. | Candidate: bounded Classic cancellation journey; no request-count credit inferred from inventory alone. |
| 010 | UI Escape/app cancel, no finish/auto-repeat, unchanged list and explicit new attempt. | Partial: application AbortSignal cancellation is covered; physical native-prompt cancel behaviour needs manual testing. |
| 011 | Synthetic blur during pending UI ceremony leaves it pending. | Manual: real OS/browser focus transfer and successful completion without duplicate prompt. |
| 012 | UI `Settings rejects a verified duplicate credential…`: real first key and second creation with exclusion list, crafted original none-attestation material/fresh client data returns native409 after verification. No server collision seed, second row or success; original record/sessions unchanged, no replay on refresh/re-entry, old key signs in. | Candidate: bounded adversarial-client Settings path; no physical authenticator duplicate UX, attestation-forgery or full fixture-substitution credit. |
| 013 | UI `Settings … failure retains confirmed state…`: list/rename/remove each fail with503 and pre-delivery network abort. Role alerts, last-confirmed snapshot, no success/empty list, disabled writes and unchanged native state; explicit Refresh then deliberate native retry succeeds. | Candidate: six Classic fault/recovery cases; no post-commit lost-write inference from pre-delivery faults. |
| 014 | UI `Settings … reconciles a lost registration response without replay`: native commit then lost response; leave via pane switch or close Settings. Returning automatically fetches inventory; held GET gates rows with no false empty state. Key appears once; second return and reload preserve exact committed state with no further auth POST. | Candidate: the explicit-return gap is closed for Classic/localhost/CDP. Draft and attachment pill retained; attachment-byte equality not established by this case. |
| 015 | API/policy unavailable-state cases plus `Actual insecure host blocks browser authority…`: real HTTP non-localhost hostname mapped to loopback, secureContextfalse, copied valid cookie refused, no credential/UI writes, protected APIs401/403/full state equality and localhost positive control. | Partial: Gi blocks Settings at the login gate and shows HTTPS/localhost guidance there; the frozen insecure Settings journey cannot run without weakening security. API overrides still do not establish old-browser compatibility. |
| 016 | UI `Classic narrow passkeys avoid horizontal clipping…`:390px,80-code-point name, metadata/button/error bounds, associated input labels and accessible button names. Tab/Enter reaches rename/remove/cancel; trusted tap repeats those actions, without writes. Status/alert roles and held-read outside-focus retention checked. | Partial: Visual skin and physical assistive-technology announcements unverified. Touch path uses programmatic focus for the outside-pane focus guard only; action activation uses real taps. |
| 017 | Native `TestBrowserPasskeyManagementAuthority` covers exact missing-owner list, automation register, expired/revoked rename, foreign-origin remove and account-query list refusal; full auth-file byte equality, no inventory/secret/cookie disclosure and fresh owner-list control. Browser uses a real key for seven query variants and normal refresh. | Partial: the first five example contexts have bounded native evidence; family-shared mode is unsupported, so the final example remains open. |
| 018 | UI `Settings native registration … failure hides proof…`: seven native finish failures (expiry, consumed, foreign session, altered origin/RP, invalid proof, revoked after creation). Retained records unchanged, old key signs in after every failure; actual secrets/challenges/proof blobs absent from DOM/storage/URL/native error. Foreign proof later succeeds in its own session. | Partial: revocation is after real creation but before finish delivery, not while a physical native prompt is open. Altered origin/RP cases prove rejection of tampered proof, not isolated verifier-check ordering. |
| 019 | UI `Settings cancelled reauthentication cannot authorise…`: separate add/rename/remove cases; reload/Refresh preserve stale proof, direct403, button/Escape abort real credential.get without finish or credential mutation; presence restoration cannot complete it. Third explicit ceremony refreshes proof without extending expiry and permits the intended change. Verification focus returns to the same logical action after rendering. | Candidate: all three operations with application cancellation in Classic/CDP; physical OS prompt interaction stays unverified. |
| 020 | UI `Settings removal explains…` and native result tests cover six supported outline cases plus defensive disabled-secret state. Atomic response identifies refusal reason or accepted remaining method; UI explains it after confirmed result, with no optimistic change/session loss. | Partial: pending TOTP addition for an established owner is not implemented. A seeded disabled secret is defensive-state coverage, not that workflow; RP/skin/device substitutions still apply. |
| 021 | UI `Two Settings views concurrently remove different keys…`: both POSTs held, no optimistic removal, then one200/one409; both refresh to the same surviving key, with no automatic retry. Explicit retry returns exact last-factor409; cookies/sessions survive and the remaining key signs in afresh. | Partial: the non-blocking auth writer lock can refuse the concurrent loser with a state-conflict409 before the factor check. The frozen scenario requires refusal by the write-time lockout check; the later explicit retry proves that check separately. |
| 022 | UI `Policy changes reject stale browser revisions…`: removal confirmation predates other browser's passkey-only change, removal refused at commit, refresh reconciles. | Candidate: direct UI policy/removal ordering plus native race tests. |
| 023 | No implemented legacy passkey-delete command workflow established. | Unsupported: explicit command refusal/Settings guidance test. |
| 024 | UI `Removing a passkey preserves sessions until explicit…`: removal retains exact sessions/cookies; one browser logs out despite removed-factor proof while another stays valid, or crosses a shortened native expiry deadline. New login retains draft/pill. Separate Chromium/WebKit logout checks gate transition on native status and cancel without writes. | Candidate: native session-end paths verified. Expiry uses an accelerated two-second fixture deadline, not a twelve-hour wall-clock wait; inherited scope substitutions remain. |
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
