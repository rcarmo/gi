# Multi-passkey scenario review

Date: 2026-09-25. Scope: the separately pinned26-scenario Piclaw contract, after
native WebAuthn, Classic Settings/login and sign-in policy controls. This is a
source-to-test review, not an automated per-case pass report. No new mappings
are added by this document. Frozen Classic/shared counts are unchanged.

The current passkey suite has nine tests in Chromium at three viewport sizes.
CDP virtual authenticators sign real browser ceremonies; physical/native focus,
synced credentials and the Visual skin are not verified. The auth regression
runs Chromium and WebKit but does not establish WebKit passkey ceremonies.

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
| 001 | UI `Settings enrolls two passkeys…`: list/actions and credential identity; API metadata projection. | Partial: Visual skin, explicit displayed dates/identifier/Never used and DOM secrecy assertions. |
| 002 | UI first Settings enrolment after TOTP login; native verification and persisted list. | Partial: explicitly assert retained TOTP login and no enrolment token in URLs/chat. |
| 003 | UI passkey-only reauth and further enrolment with no TOTP; distinct virtual credential material. | Partial: narrowly map one-existing-key to second-key journey and no chat-link prompt. |
| 004 | API and UI two-key sign-in after server restart with fresh cookies; only used key's last-used time advances. | Candidate: both passkey examples, virtual-authenticator scope only. |
| 005 | Data model stores credentials, not physical devices. | Manual: one synced credential used on two physical devices, one row retained. |
| 006 | Native rename preserves material; UI rename of named key. | Partial: previously unnamed keys and explicit reload-persistence UI assertion. |
| 007 | Native blank/long/control names rejected, literal markup accepted without changing material. | Partial: all validation examples through UI and subsequent sign-in with the key. |
| 008 | UI confirmation/remove, native removed-key rejection and remaining-key sign-in. | Partial: confirmation identifier and zero removal request before confirmation. |
| 009 | UI cancel preserves rows and restores invoking control focus. | Partial: explicit zero-request assertion on cancellation. |
| 010 | UI Escape/app cancel, no finish/auto-repeat, unchanged list and explicit new attempt. | Partial: application AbortSignal cancellation is covered; physical native-prompt cancel behaviour needs manual testing. |
| 011 | Synthetic blur during pending UI ceremony leaves it pending. | Manual: real OS/browser focus transfer and successful completion without duplicate prompt. |
| 012 | Browser creation exclusions and independently enforced duplicate finish409; stored keys unchanged. | Partial: full duplicate path via Settings, rather than native API integration only. |
| 013 | UI list/rename/remove faults retain confirmed rows and show alerts; Refresh is available. | Partial: explicit accessible retry/no-success assertions for every outline example. |
| 014 | Lost successful finish, UI uncertain-result explanation, explicit list refresh adds one row, no second finish. | Candidate: authoritative reconciliation without blind replay. |
| 015 | Native origin/config/policy gates; UI TOTP-only Add disabled; Auth UI unconfigured/unenrolled explanations. | Partial: all unavailable-browser/insecure-origin outline explanations plus zero credential API calls. |
| 016 | Three viewports, roles/labels, overflow check and keyboard cancellation. | Partial: Visual skin,80-character visible names, complete keyboard/touch traversal and live-announcement/focus checks. |
| 017 | HTTP no-cookie, bearer/query, foreign origin, other-account field and transport denial; session binding. | Partial: all operations/examples, including explicit family-mode denial, plus no inventory leakage. |
| 018 | Native/browser expired/consumed/other-session/revoked/origin/RP/signature failures. | Partial: explicit rendered-failure secret/challenge non-disclosure assertions. |
| 019 | Stale add/remove native checks; UI TOTP/passkey reauthentication and current-policy proof rejection. | Partial: stale rename and cancel-reauth/no-write evidence for each operation. |
| 020 | Native current-RP/policy/factor removal matrix; UI last-key refusal. | Partial: active-session-only and pending-unverified-TOTP examples, exact UI explanation for each case. |
| 021 | Native simultaneous removal has one winner; one key remains and sessions survive. | Partial: two Settings views concurrently remove and refresh to the same final inventory. |
| 022 | UI `Policy changes reject stale browser revisions…`: removal confirmation predates other browser's passkey-only change, removal refused at commit, refresh reconciles. | Candidate: direct UI policy/removal ordering plus native race tests. |
| 023 | No implemented legacy passkey-delete command workflow established. | Unsupported: explicit command refusal/Settings guidance test. |
| 024 | UI removal explains future sign-ins versus existing sessions; native session validity retained. | Partial: normal expiry and explicit logout lifecycle after removal. |
| 025 | UI failed finish says not registered, local credential may remain and Gi did not remove it; native authenticator/store checks. | Candidate: local-orphan explanation without false cleanup. |
| 026 | Auth UI proof refresh affects only one browser; other browser stays stale; native binding tests. | Partial: directly reject add/rename/remove from the other stale browser after proof refresh. |

Initial owner bootstrap and production RP/domain selection are separate deployment
work. The operator's live instance has not been enrolled or had its policy changed.
Terminal adaptation stays browser-based: no WebAuthn emulation, credential secret
entry or extra idle terminal rows.

[addition]: ../../tests/ux/features/additions/piclaw-2026-09-24/README.md
