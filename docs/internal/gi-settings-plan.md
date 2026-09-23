# Gi settings

Gi settings reuse Piclaw's modal layout while exposing Gi's own capabilities. The derived scenarios are in `tests/features/settings/gi-settings.feature`; they never increase frozen Classic/shared coverage. Independent whole-criterion frozen acceptance lives in `tests/ux/settings-shell.spec.mjs`.

## Implemented sections

| Section | Scope | Native source | Editing |
|---|---|---|---|
| General | Instance, loaded at startup | Authenticated `GET /api/runtime/config`, `GET/PATCH /api/settings/identity` | Active values read-only; explicit saved display-name writes requiring restart |
| Models | Captured session | Authenticated `GET/PATCH /api/sessions/:id/model` | Explicit Apply, native validation, context-fit guard, no global writes |
| Appearance | Browser profile and origin, all sessions | One `gi_browser_appearance_v1` localStorage record | Explicit save/reset, preset and default-theme hex tint; no server/TUI writes |
| Compaction | Active startup policy, saved next-start policy, captured session action | Authenticated policy and session compaction/activity endpoints | Explicit policy save with restart requirement; Compact/Refresh and matching-turn Stop |
| Providers | Service user's native credential file, shared by Gi processes | Authenticated `GET/PATCH/DELETE /api/settings/providers` | OpenAI/Anthropic API-key save and confirmed removal; OAuth/custom entries read-only |

Piclaw scaffold: commit `70d33bc93ab540845bbcf5f80503ca8125c71594`, `runtime/web/src/components/settings-dialog.ts`, `runtime/web/static/classic/css/settings.css`, keyboard `openSettings` bindings. Gi owns its derived dialog and CSS; supplied components remain unchanged. `BodyPortal` and the existing timeline `piclaw:open-settings` event are reused.

The General/Models hierarchy, fixed half-opaque backdrop, sidebar-to-tabs responsive layout and close behaviour follow the scaffold. Gi adds dialog semantics, a focus trap and inert background. Control/Meta/Alt+Comma open one instance. Model reads and writes are fenced by dialog mount and session selection; closed-view results cannot update a new view. The adapter's model mutation revision protects composer state against older safety-poll responses.

## Capability audit

- `internal/config/config.go` loads `.piclaw/config.json` identity and `.pi/settings.json` runtime defaults. The narrow identity endpoint edits only assistant/user names; Piclaw General autosave and other global writes are unsupported.
- `internal/web/session_model.go` validates model changes through `inference.SelectSessionModel`; session choices persist in SQLite. It reports current model, thinking, context usage and available model metadata.
- Providers supports explicit OpenAI/Anthropic API-key changes to the native credential store. OAuth/token/custom entries remain read-only; browser OAuth login and refresh management are not implemented.
- Compaction displays active/saved policies separately and supports validated enablement/token-budget saves for the next restart. Remote models, watchdog and backoff settings remain unsupported.
- Workspace indexing is explicitly configured at startup; the existing index actions are distinct from settings writes.
- Budget, scheduled tasks, recordings, environment overrides, keychain CRUD, add-on management and browser shortcut editing lack the Piclaw backend contracts. No placeholder enabled navigation entries.
- Appearance has a Gi-owned browser storage contract and uses the supplied theme catalogue/renderer through build-time exports. Editor preferences still need separate scope and acceptance.

## Frozen-spec relationship

`settings-dialog.feature`, `settings-layering.feature` and canonical `core-settings.feature` inform the shell and interaction tests. Frozen assumptions about General autosave, pane imports, provider authentication, broad model filters and services do not automatically hold for Gi. Keep frozen source files unchanged; map a row only when its entire criterion independently passes.

Independent shell acceptance now covers `@ux-settings-layering-001`–`004` and `@ux-settings-dialog-001/003/004`: actual workspace pointer blocking/recovery, viewport/portal/dimming, rapid keyboard open, immediate loading shell/General-first response within two seconds, and typed numeric input. Frozen six-project matrix **42/42**, combined shell/Gi regression **132/132**, rapid-open tablet repeat **10/10**. A first-paint Escape race was fixed by attaching modal focus/inert/keyboard handling in `useLayoutEffect`. Cached reopen (`dialog-002`) remains unmapped because not all Settings data is cached; lazy-pane imports (`dialog-005`) remain unmapped because panes are statically imported. Overall coverage is **71/236 Classic**, **3/42 shared**, **165/39 unmapped**; browser layering adds no terminal claim.

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

## Saved automatic-compaction policy — 2026-09-23

`gi-settings-018`/`019` add explicit next-start policy saves: 19 Gi scenarios and one provider proposal. The active engine policy remains read-only above a separate saved form. `GET/PATCH /api/settings/compaction` returns active policy, saved policy/revision and `restart_required`. No API updates the running engine or restarts it. A late save response cannot populate or announce success in a reopened view.

Editable fields are enabled, context window, reserve, keep-recent and trigger threshold. All numeric fields must be positive integers; context is capped at 16,777,216 tokens, reserve must be below context, and keep-recent ≤ threshold ≤ context minus reserve. These are write constraints; legacy policy values remain visible for explicit repair. The current loader's zero/default rules are unchanged, including the distinction between an absent settings file (automatic compaction enabled) and an existing file without a compaction section (disabled). Strategy remains a preserved label, not an algorithm chooser; unknown section/root fields are retained.

Persistence uses `.pi/settings.json` with whole-file SHA-256 revision checks, rooted regular-file checks, a 1 MiB bound, temporary write/fsync/rename/directory sync, and `.gi-settings.lock`. All five native model/TUI preference writers now share that same mutex/advisory lock and atomic merge, avoiding native lost updates across fields. Invalid, empty or unreadable files are no longer silently replaced by these older writers. New directories/files use 0700/0600; existing file mode bits are preserved. Unknown numbers retain JSON representation via RawMessage. Existing identity writes keep their separate Piclaw file/lock through a shared directory-opening helper.

Locking supports Linux/macOS; unsupported platforms reject mutation. Noncooperating external editors still have a final-check/rename race, and hostile local config-directory writers are outside the contract. Atomic replacement does not guarantee ownership/ACL/xattr preservation. Post-rename directory-sync errors require Reload to verify the saved file. A conflict or failed write preserves form fields; Reload saved policy explicitly discards them. JSON-only, same-origin browser-write protection and existing authentication apply before mutation; response bodies exclude unrelated config values.

Evidence: Settings/Models **126/126**, real-provider compaction **96/96**, functional **83/83**, support **43/43**, Go/vet/build/hook and config/web race ×3. Tests include concurrent model/policy merge, all native writers respecting another process's lock, unchanged originals on failure, unknown-field/mode preservation, default/load equivalence, unsafe-file rejection and fresh-process policy activation. Existing TUI compaction acceptance passes at **60×18, 100×22, 140×36**, preserving draft/cursor/resize/reopen/idle rows. Darwin arm64 config test compilation passes. An initial full Go run hit a direct-steering SQLite lock error; that test passed ten repeats and the subsequent full suite passed.

Captures: `test-results/gi-settings-captures/policy-*.png`. No new terminal control, provider credential operation or frozen parity credit. Frozen counts remain **64/236 Classic**, **3/42 shared**, **172/39 unmapped**.

## Provider API-key setup — 2026-09-23

`gi-settings-020`–`023` add Providers: 23 Gi scenarios plus a browser-OAuth proposal. The service user owns `~/.pi/agent/auth.json`; it is shared across that user's Gi instances/workspaces, not session-scoped or an encrypted keychain. The UI allows explicit OpenAI/Anthropic `api_key` setup and confirmed removal. Unknown/OAuth/token entries are read-only, including allowlisted provider IDs whose existing entry contains access/refresh/token credentials. Removal does not revoke credentials at the provider or cancel requests already using them.

Metadata includes provider identifier/name, stored/kind/material flags and edit capability. Stored never means provider-verified. Reads do not contact providers. No key/access/refresh values, credential entry extras or file path are returned. Revisions are per-process HMACs with a random key rather than externally reusable credential hashes; restart requires fresh metadata. PATCH/DELETE require existing instance authentication, JSON, same-origin browser protection, and either direct TLS or both a loopback peer and loopback Host. Forwarded headers are not trusted. HTTPS termination at a reverse proxy needs a separate trusted transport design; it does not silently enable writes here. Gi's existing unenrolled-instance authentication semantics are unchanged.

Keys must be nonempty valid UTF-8, at most 4096 bytes, without whitespace/control characters. Password inputs are transient DOM state; success, failure, filter/refresh and dialog teardown clear them. Browser local/session storage never receives secrets. Errors are generic and never echo provider bodies or request keys. Save success means the file was persisted, not that a provider accepted the key; the next native request reads it. Model selection and drafts stay unchanged.

Credential writes use a dedicated mutex, nonblocking `.gi-auth.lock` on Linux/macOS, whole-store revision comparison, rooted directory/opened-file checks, 1 MiB bound, RawMessage preservation, temporary 0600 file, fsync/rename/directory sync. Parent dirs created by Gi use 0700; existing directory ownership/modes are not changed. Both Settings writes and native TUI `/logout` use this path. RawMessage preserves unrelated/custom fields; unsupported/unsafe stores are rejected without replacement. As with other settings writes, noncooperating external tools can race final check/rename, ownership/ACL/xattrs are not preserved by atomic replacement, and post-rename sync failure requires a metadata reload. Browser OAuth/refresh is not implemented.

Evidence: isolated fixture overrides HOME and routes an OpenAI model to a local HTTP provider that rejects requests without the exact saved fixture key. Native turn completion proves consumption; no operator credentials or remote provider quota are used. Six-project provider matrix **36/36**, Settings/Models **126/126**, functional **83/83**, support **43/43**, Go/vet/build/hook, inference/web race ×3 and Darwin inference test compilation pass. Coverage includes stale revisions, cancellation/confirmation, real filesystem failure and retry, transport-disabled forms, no response/storage/console key disclosure, held reads/writes, native logout preservation, concurrent saves, symlink/FIFO/corrupt rejection, cross-process locking and fresh-process credential consumption.

Terminal disposition: existing `/login` inspection and `/logout` retain their explicit no-idle-row interaction; logout now shares the safe writer. No new terminal credential editor or OAuth claim. Frozen parity remains **64/236 Classic**, **3/42 shared**, **172/39 unmapped** because Piclaw's full provider setup scenario includes unsupported OAuth/custom flows. Captures: `test-results/gi-settings-captures/providers-*.png`.
