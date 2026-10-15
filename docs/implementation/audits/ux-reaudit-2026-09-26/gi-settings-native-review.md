# Gi-derived Settings 001–029: native evidence, separate from Piclaw parity

`tests/features/settings/gi-settings.feature` now defines 29 numbered
Gi-specific scenarios (`001`–`029`) and one browser OAuth proposal
(`@gi-settings-next-004`). The baseline inventory has 28 definitions in this
file: `001`–`027` and the proposal. `028` and `029` were added later. The
proposal is not a shipped capability. Gi mounts
General, Models, Appearance, Compaction, Providers and Authentication through
`web/src/gi-settings.ts`; five non-General panes lazy-load. These tests use
disposable Gi servers and cannot earn frozen Classic/Shared Piclaw parity.

| IDs | Native test/code trace | Scoped result |
|---|---|---|
| `001`–`003` | Modal portal/focus/backdrop and six-section scopes; held config read, error and Retry with cached General. `gi-settings.spec.mjs` plus `settings-shell.spec.mjs`. | Included in focused **114/114** shell/core run; no full Settings gate. |
| `004`–`007` | Lazy model read, 50-result cap, explicit Apply and thinking, rejected/incompatible model, stale session/view guards. `gi-settings.spec.mjs`, plus `context-fit.spec.mjs` catalogue gate. | `004` was previously corrected locally; model/policy run **24/24** included `004`–`006`; catalogue gate **6/6**; core run included `007`. |
| `008` | `internal/web/settings_auth_test.go` denies unauthenticated native Settings/model reads and writes before malformed payload parsing, and checks no auth secret leakage. | `go test ./internal/web -run '^TestGiSettingsRoutesPreserveAuthentication$' -count=1` passed. No tagged browser journey for this ID; cross-origin writes have separate server tests. |
| `009`–`011` | `gi-settings-appearance.ts` persists a versioned browser-local preference, validates tint/storage errors, resets and handles cross-tab changes without changing server/legacy state. | Included in focused **114/114** core run; not Classic `/theme`/`/tint`. |
| `012`–`013` | Explicit owner display-name save; revision conflict, invalid or unsafe config, failed write and restart-required active/saved distinction. | Included in focused **114/114** core run; isolated config fixture, not live restart. |
| `014`–`017` | Read-only active compaction policy, explicit native admission/Stop, authoritative progress and session/dialog fencing. `compaction.spec.mjs`. | **30/30** across six projects; not Piclaw watchdog/backoff controls. |
| `018`–`019` | Saved automatic policy is separate from active engine policy; restart required, validation/conflict/failed-write and close fences. `gi-settings.spec.mjs`. | **24/24** model/policy run includes both; no live restart. |
| `020`–`023` | Metadata-only provider reads; explicit allowlisted API-key save/consumption, validation/conflict and confirmed removal with closed-view guards. `gi-providers.spec.mjs`. | **36/36** across six projects. Credential file is private plaintext, not encrypted Keychain or browser OAuth. |
| `024`–`027` | Lazy chunks, cache/import failures, responsive header and model settlement/read fencing. `settings-shell.spec.mjs` and `gi-settings.spec.mjs`. | Included in focused **114/114** core run. |
| `028`–`029` (post-baseline) | Explicit supported thinking selection reaches provider reasoning, default omits it, model switch resets model-bound choice without submitting draft. `session-thinking.spec.mjs`. | **12/12** across six projects; these two are outside the 358-definition baseline ledger. |

The first attempt at the separate `004` large-catalogue test omitted
`GI_UX_SETTINGS_CATALOGUE=1` and found no tests. The corrected six-project
run passed. Focused commands overlap IDs and are not additive coverage counts.
Current Piclaw 3.2.4 Settings UI, physical authenticator, deployed Gi, and a
whole-product CI gate were not exercised in this audit. Production files and
frozen Classic/Shared Gherkin were not changed.
