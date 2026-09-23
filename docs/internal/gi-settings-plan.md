# Gi settings

Gi settings reuse Piclaw's modal layout while exposing Gi's own capabilities. The derived scenarios are in `tests/features/settings/gi-settings.feature`; they never increase frozen Classic/shared coverage.

## Implemented sections

| Section | Scope | Native source | Editing |
|---|---|---|---|
| General | Instance, loaded at startup | Authenticated `GET /api/runtime/config` | Read-only identity, workspace and model/thinking defaults; file/restart guidance |
| Models | Captured session | Authenticated `GET/PATCH /api/sessions/:id/model` | Explicit Apply, native validation, context-fit guard, no global writes |
| Appearance | Browser profile and origin, all sessions | One `gi_browser_appearance_v1` localStorage record | Explicit save/reset, preset and default-theme hex tint; no server/TUI writes |

Piclaw scaffold: commit `70d33bc93ab540845bbcf5f80503ca8125c71594`, `runtime/web/src/components/settings-dialog.ts`, `runtime/web/static/classic/css/settings.css`, keyboard `openSettings` bindings. Gi owns its derived dialog and CSS; supplied components remain unchanged. `BodyPortal` and the existing timeline `piclaw:open-settings` event are reused.

The General/Models hierarchy, fixed half-opaque backdrop, sidebar-to-tabs responsive layout and close behaviour follow the scaffold. Gi adds dialog semantics, a focus trap and inert background. Control/Meta/Alt+Comma open one instance. Model reads and writes are fenced by dialog mount and session selection; closed-view results cannot update a new view. The adapter's model mutation revision protects composer state against older safety-poll responses.

## Capability audit

- `internal/config/config.go` loads `.piclaw/config.json` identity and `.pi/settings.json` runtime defaults. There is no General HTTP write contract. Do not route Piclaw General autosave to a fabricated endpoint.
- `internal/web/session_model.go` validates model changes through `inference.SelectSessionModel`; session choices persist in SQLite. It reports current model, thinking, context usage and available model metadata.
- Provider auth currently loads native credentials; no browser provider sign-in/key storage API is available.
- Manual compaction is session-bound. Piclaw policy, watchdog and backoff settings are not equivalent to Gi's startup configuration.
- Workspace indexing is explicitly configured at startup; the existing index actions are distinct from settings writes.
- Budget, scheduled tasks, recordings, environment overrides, keychain CRUD, add-on management and browser shortcut editing lack the Piclaw backend contracts. No placeholder enabled navigation entries.
- Appearance has a Gi-owned browser storage contract and uses the supplied theme catalogue/renderer through build-time exports. Editor preferences still need separate scope and acceptance.

## Frozen-spec relationship

`settings-dialog.feature`, `settings-layering.feature` and canonical `core-settings.feature` inform the shell and interaction tests. Frozen assumptions about General autosave, pane imports, provider authentication, broad model filters and services do not automatically hold for Gi. Keep every existing frozen row unchanged and unmapped unless its entire criterion independently passes.

## Refinement decisions

The user is the instance operator using the current chat. Success is a useful settings entry point with truthful instance/session scope and native persisted model changes. MVP is General inspection plus Models editing; credentials and global writes are excluded. Go remains the runtime; no new backend service or config migration. Failures appear inside the modal, preserve draft choices and allow retry. Unrelated session/composer state stays intact. Reads use existing authenticated routes; secrets never enter the General view. Choices are filtered and capped at 50 visible rows. No export or settings import is required. Later global writes need revision checks, atomic preservation of unrelated configuration keys and explicit restart semantics.

Browser acceptance: Chromium/WebKit at 390×844, 820×1180, 1440×900; held native responses, real rejection, persisted state/reload, draft/media preservation, keyboard/focus and overlays. Functional suite, Go/vet and web checks remain release gates. New Gi feature parser checks prevent accidental frozen-corpus credit.

Terminal adaptation: keep existing `/model`, Alt-M and context footer. Instance inspection, if added, belongs in an explicit command/short modal with no idle rows; no permanent settings panel or terminal imitation of the web sidebar. No new terminal acceptance is claimed by this browser slice.

## Acceptance evidence — 2026-09-23

| Gi cases | Evidence |
|---|---|
| 001–003 | `gi-settings.spec.mjs`: real menu/shortcut entry, one body portal, open workspace, half-opaque backdrop, inert/focus trap, close/reopen, held GET and retry |
| 004 | `gi-settings.spec.mjs`: delayed native model catalogue and filter; `context-fit.spec.mjs` with `GI_UX_SETTINGS_CATALOGUE=1`: 60 native fixture models, 50 rendered choices, explicit filtering to model 59 |
| 005–006 | Native accepted/rejected PATCH, held response, retained draft/media, reload/other-session/default isolation; real provider measurement blocks an undersized model, permits exact fit and unknown usage |
| 007 | Held native GET and accepted PATCH across close → session switch → reopen; target state and drafts remain isolated |
| 008 | `TestGiSettingsRoutesPreserveAuthentication`: enrolled native auth denies both reads and malformed PATCH before parsing, without leaking fixture credentials |

Final settings matrix: **36/36** (six Chromium/WebKit viewport projects). Native large-catalogue/context suite: **24/24**, including existing compaction checks. Existing model/session plus earlier settings matrix: **156/156**. Fresh functional suite: **82/82**. Support tests: **40/40**. Go tests, vet, web build/hook checks and targeted auth/model race tests ×3 pass. Captures are in `test-results/gi-settings-captures/`.

The initial three proposals included Appearance; it is implemented below. Identity and provider/compaction write proposals remain unimplemented. Frozen parity remains **64/236 Classic**, **3/42 shared**, **172/39 unmapped**. No supplied component or terminal code changed. Scaffold hashes are recorded in `web/upstream/piclaw-settings-scaffold-70d33bc93.json`.

## Browser appearance contract — 2026-09-23

`gi-settings-009`–`011` replace the appearance proposal; the file now contains 11 executable-scope Gi scenarios and two future proposals. The operator chooses a supplied preset and optional default-theme hex tint. Fields are drafts until Save. Changing preset clears the tint draft. Reset saves default/no tint and follows system light/dark mode; it does not remove the record or revive a legacy chat override.

Storage is one versioned `{version:1, theme, tint}` record at `gi_browser_appearance_v1`. Validation precedes one `setItem`, and rendering follows successful persistence. Invalid/unsupported data is ignored at startup. Denied storage reports an error without changing the visual theme, and the draft remains available for retry. No configuration, server model or legacy Piclaw theme key is written. An explicit Gi record wins over legacy theme selection at startup; absent records retain legacy startup behaviour.

Same-origin tabs receive storage events and apply valid preferences immediately. A clean Appearance form updates; a dirty one preserves fields and announces the external change. Saving that form explicitly overwrites the current preference. Scope is this browser profile/origin across sessions, not accounts, devices or terminals. Clearing browser storage resets this preference. No credentials are stored here.

`build.js` exports `THEME_PRESETS` and `applyThemeState` from the unchanged supplied `ui/theme.ts` during bundling. Gi calls the renderer with `persist:false`, avoiding its per-chat writes and swallowed storage errors. No duplicate palette list or copy of theme algorithms is maintained. Tint accepts only empty, `#RGB` or `#RRGGBB`, normalised to six lowercase digits; arbitrary CSS expressions are rejected.

Evidence: Settings/Models matrix **90/90** (54 Gi settings tests + 36 existing model tests), functional **83/83**, support **43/43**, Go/vet/build/hook pass. Browser tests exercise no-pre-save changes, exact persisted records, unchanged legacy keys, zero server mutations, session/reload scope, tint round-trip, storage denial/retry, malformed startup data, cross-tab dirty drafts and system colour changes. Captures: `test-results/gi-settings-captures/appearance-*.png`. Frozen counts remain **64/236**, shared **3/42**, unmapped **172/39**.

Terminal disposition: browser tint and CSS presets do not alter the TUI. A future terminal theme selector should be an explicit bounded choice using terminal palettes and separate PTY tests, never a persistent sidebar or added idle row.
