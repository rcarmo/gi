# Gi settings

Gi settings reuse Piclaw's modal layout while exposing Gi's own capabilities. The derived scenarios are in `tests/features/settings/gi-settings.feature`; they never increase frozen Classic/shared coverage.

## Implemented sections

| Section | Scope | Native source | Editing |
|---|---|---|---|
| General | Instance, loaded at startup | Authenticated `GET /api/runtime/config`, `GET/PATCH /api/settings/identity` | Active values read-only; explicit saved display-name writes requiring restart |
| Models | Captured session | Authenticated `GET/PATCH /api/sessions/:id/model` | Explicit Apply, native validation, context-fit guard, no global writes |
| Appearance | Browser profile and origin, all sessions | One `gi_browser_appearance_v1` localStorage record | Explicit save/reset, preset and default-theme hex tint; no server/TUI writes |
| Compaction | Effective startup policy + captured session action | Authenticated session compaction/activity endpoints | Policy read-only; explicit Compact/Refresh and matching-turn Stop |

Piclaw scaffold: commit `70d33bc93ab540845bbcf5f80503ca8125c71594`, `runtime/web/src/components/settings-dialog.ts`, `runtime/web/static/classic/css/settings.css`, keyboard `openSettings` bindings. Gi owns its derived dialog and CSS; supplied components remain unchanged. `BodyPortal` and the existing timeline `piclaw:open-settings` event are reused.

The General/Models hierarchy, fixed half-opaque backdrop, sidebar-to-tabs responsive layout and close behaviour follow the scaffold. Gi adds dialog semantics, a focus trap and inert background. Control/Meta/Alt+Comma open one instance. Model reads and writes are fenced by dialog mount and session selection; closed-view results cannot update a new view. The adapter's model mutation revision protects composer state against older safety-poll responses.

## Capability audit

- `internal/config/config.go` loads `.piclaw/config.json` identity and `.pi/settings.json` runtime defaults. The narrow identity endpoint edits only assistant/user names; Piclaw General autosave and other global writes are unsupported.
- `internal/web/session_model.go` validates model changes through `inference.SelectSessionModel`; session choices persist in SQLite. It reports current model, thinking, context usage and available model metadata.
- Provider auth currently loads native credentials; no browser provider sign-in/key storage API is available.
- Compaction displays the engine's effective startup policy and session-bound actions. Piclaw policy writes, remote models, watchdog and backoff settings remain unsupported.
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

The initial three proposals included Appearance; it is implemented below. Identity is implemented below; the provider/compaction write proposal remains unimplemented. Frozen parity remains **64/236 Classic**, **3/42 shared**, **172/39 unmapped**. No supplied component or terminal code changed. Scaffold hashes are recorded in `web/upstream/piclaw-settings-scaffold-70d33bc93.json`.

## Browser appearance contract — 2026-09-23

`gi-settings-009`–`011` replace the appearance proposal; the file now contains 11 executable-scope Gi scenarios and two future proposals. The operator chooses a supplied preset and optional default-theme hex tint. Fields are drafts until Save. Changing preset clears the tint draft. Reset saves default/no tint and follows system light/dark mode; it does not remove the record or revive a legacy chat override.

Storage is one versioned `{version:1, theme, tint}` record at `gi_browser_appearance_v1`. Validation precedes one `setItem`, and rendering follows successful persistence. Invalid/unsupported data is ignored at startup. Denied storage reports an error without changing the visual theme, and the draft remains available for retry. No configuration, server model or legacy Piclaw theme key is written. An explicit Gi record wins over legacy theme selection at startup; absent records retain legacy startup behaviour.

Same-origin tabs receive storage events and apply valid preferences immediately. A clean Appearance form updates; a dirty one preserves fields and announces the external change. Saving that form explicitly overwrites the current preference. Scope is this browser profile/origin across sessions, not accounts, devices or terminals. Clearing browser storage resets this preference. No credentials are stored here.

`build.js` exports `THEME_PRESETS` and `applyThemeState` from the unchanged supplied `ui/theme.ts` during bundling. Gi calls the renderer with `persist:false`, avoiding its per-chat writes and swallowed storage errors. No duplicate palette list or copy of theme algorithms is maintained. Tint accepts only empty, `#RGB` or `#RRGGBB`, normalised to six lowercase digits; arbitrary CSS expressions are rejected.

Evidence: Settings/Models matrix **90/90** (54 Gi settings tests + 36 existing model tests), functional **83/83**, support **43/43**, Go/vet/build/hook pass. Browser tests exercise no-pre-save changes, exact persisted records, unchanged legacy keys, zero server mutations, session/reload scope, tint round-trip, storage denial/retry, malformed startup data, cross-tab dirty drafts and system colour changes. Captures: `test-results/gi-settings-captures/appearance-*.png`. Frozen counts remain **64/236**, shared **3/42**, unmapped **172/39**.

Terminal disposition: browser tint and CSS presets do not alter the TUI. A future terminal theme selector should be an explicit bounded choice using terminal palettes and separate PTY tests, never a persistent sidebar or added idle row.

## Saved instance names — 2026-09-23

`gi-settings-012`/`013` replace the identity proposal. General displays active startup values and separate editable saved names. Save is explicit; a successful response returns the stored revision and `restart_required`. Reload saved names explicitly discards the form draft. The process does not update its startup config or restart itself. The next Gi process loads the names from the usual Piclaw file. There are now 13 Gi scenarios and one provider/compaction proposal.

`GET/PATCH /api/settings/identity` reuses authentication and returns only names/revision/activation state, never unrelated configuration or avatars. PATCH requires JSON, a bounded single request object and Go's cross-origin protection check. Names must be valid UTF-8, nonblank, no more than 128 Unicode characters, and free of control characters; whitespace is trimmed on save. Existing legacy strings remain readable so operators can repair them; all new writes must pass validation.

Storage is `.piclaw/config.json`, preserving unknown/nested keys and number values through `json.RawMessage`, plus existing permission bits. Reads reject malformed/nonobject JSON, invalid object sections, files over 1 MiB, symlinks and nonregular files. Rooted filesystem operations and opened-object checks protect the config-directory/lock path. In-process writers are serialised; Linux/macOS cooperating processes use a nonblocking advisory lock. The revision is a hash of the whole input file. A stale revision or held lock yields 409; the draft remains visible. Unsupported operating systems reject writes rather than proceeding without a lock.

Save writes a unique temporary file, syncs it, checks the revision again, renames atomically and syncs the directory. Pre-rename failures leave the original file intact; a post-rename directory-sync failure reports uncertainty and requires a reload to verify. Existing file ownership/ACL/xattrs are not guaranteed to survive atomic replacement; regular mode bits are preserved. New files use 0600. The `.gi-identity.lock` file intentionally remains. Noncooperating external editors can still race the final revision check/rename: this is not an OS-level compare-and-swap. Do not share this config directory with hostile local writers.

Evidence: native preservation/validation/oversize/symlink/directory/FIFO tests, stale/concurrent revisions, cross-process locking, fresh-process config activation, authenticated routes, cross-origin/method/body rejection, and real browser filesystem failure/retry. Six-project Settings matrix **72/72**; earlier Settings/Models combined **102/102**; functional **83/83**, support **43/43**, Go/vet/build/hook and targeted config/web race ×3 pass. Closing a held save cannot announce success in a new view. Captures: `test-results/gi-settings-captures/identity-*.png`.

Terminal disposition: these names load on the next process start without additional terminal UI. A future explicit identity command must share this storage contract and state the restart requirement; no idle rows or terminal acceptance are added here. Frozen parity remains **64/236 Classic**, **3/42 shared**, **172/39 unmapped**.

## Compaction inspection and session actions — 2026-09-23

`gi-settings-014`–`017` add Compaction to the modal: 17 Gi scenarios plus one future provider/policy-write proposal. `GET /api/sessions/:id/compaction` includes the engine's effective startup policy and `policy_scope:"startup"`, with private/no-store caching. A test intentionally supplies different web-server config to ensure the engine remains authoritative. General config is not reloaded or mutated by this pane.

The pane shows automatic enablement, configured context window/threshold/reserve/recent-token budget and strategy label as read-only values. The current engine uses a local compactor and configured hooks; no remote compaction model, native-provider mode, watchdog, suppression reset or policy-save controls are exposed.

Compact now posts the captured history token to the existing idle-only admission path. Stop turn posts the displayed active compaction's turn ID; for automatic compaction it cancels the entire owning turn, not merely the compaction phase. Admission and cancellation responses are acknowledgements, never completion. The activity endpoint supplies matching occurrence progress and terminal state. History remains visible; drafts/media/model selection are untouched.

A one-second, non-overlapping refresh runs only while the pane is mounted. Polling continues during a delayed action response but actions stay disabled until it settles and state is refreshed. Reads are fenced by a revision; each dialog/session mounts its own lifecycle. Failed reads disable actions and recover on a successful refresh. Action errors remain until Refresh or another explicit action. Native token/run checks handle work arriving between a poll and a click without inventing an atomic browser snapshot.

Evidence: full real-provider compaction suite **96/96**, including **30** Gi-pane browser tests across six projects, plus Settings/Models **108/108**, functional **83/83**, support **43/43**, Go/vet/build/hook and targeted web/turn race ×3. Tests cover native read failure, stale-token 409, Stop failure/retry/cancellation, durable history-preserving completion, closed-view/session fences for held reads and actions, and authoritative completion before a delayed admission response. Captures: `test-results/gi-settings-captures/compaction-*.png`.

Terminal adaptation remains `/compact`, `/compact info`, Alt-C and focused Escape, with their earlier independent PTY evidence. No terminal change or new frozen credit: Classic **64/236**, shared **3/42**, unmapped **172/39**. Editing automatic policy still needs its own validated persistence and restart contract.
