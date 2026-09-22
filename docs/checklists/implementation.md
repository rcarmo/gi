# gi implementation checklist

Status: Active
Date: 2026-04-22
Last updated: 2026-09-21

This checklist is organized by **subsystem** and grouped by **phase**.

---

## Piclaw Classic web parity — 2026-09-21

- [x] Repair legacy startup: atomically create tables, apply additive columns, then build indexes; rollback/idempotence tests, real-copy migration, 70/70 functional and Go/vet/race checks. Dev instance restarted on 8090; two-way backup comparisons preserve 51 turns, 146 messages, 317 events and existing session fields.

- [x] Vendor the frozen Vibes/Tau Classic corpus and shared interaction contract with hash checks.
- [x] Add a Gi matrix covering Chromium/WebKit at phone/tablet/desktop sizes; unmapped cases are not passes.
- [x] Map `@ux-original-001` and `002`, import the upstream TimelineMenu and drawer behavior, and pass all 12 executions.
- [x] Map `@ux-original-014` with real delayed-response/draft-switch checks and test New creating a distinct child (24/24 combined browser runs).
- [x] Fix pooled SQLite configuration exposed by the session matrix; existing functional web suite passes 70/70.
- [x] Map searchable picker/focus (`013`) and native keyboard navigation; combined matrix 36/36 with pinned helper provenance.
- [x] Prevent runner startup before submission-event persistence; reproduced regression and full 70/70 functional suite.
- [ ] Map the remaining 202 frozen scenarios and 40 shared interaction cases; full scope in `docs/internal/full-web-tui-parity-plan.md`.
- [x] Complete capability-gated rename, pin, archive and restore with persisted native metadata and failure-safe picker actions (`015`); 48/48 matrix, 70/70 functional, Go/race/Bun checks.
- [x] Persist per-session browser text/media/reference drafts and unacknowledged submissions; recover on reload without automatic resend (browser-local IndexedDB).
- [x] Capture background-send ownership, merge failed submissions with newer origin drafts and report storage failures (`ux-compose-001`, `002`, `003`, `006`); 102/102 matrix, 70/70 functional.
- [x] Honour explicit queued follow-ups during active turns; add persistent FIFO/reorder and queued-only cancel APIs with conflict guards.
- [x] Wire reorder/removal error recovery and session-owned late responses (`018`); 120/120 matrix, 70/70 functional, Go/race/Bun tests.
- [x] Complete `016` accepted/SSE queue reconciliation, disconnect transient clearing (`ux-reconnect-001`) and connection-generation guards; 138/138 matrix, 70/70 functional, Go/race/Bun checks.
- [ ] Complete remaining reconnect criteria separately; shared return/Steer pass, conflicting Classic replacement/idle-send cases remain unmapped.
- [x] Add authoritative session-local model commands/mutations with catalogue validation, persistence, failed-switch retention and late-response guards (`021`, compaction `008`); 168/168 matrix, 70/70 functional, Go/race/Bun checks.
- [x] Share validated model selection between web and TUI; persist terminal choices per session, restore on startup/switch, preserve editor state and settings bytes.
- [x] Verify bounded model selector/cycle/error paths at 60×18, 100×22 and 140×36 with zero added idle rows; live restart/next-turn checks, TUI smoke/Gherkin and Go/race tests.
- [x] Persist latest provider-request input/cache measurement separately from cumulative turn usage; expose session-scoped context and validate model fit without fabricated zeroes.
- [x] Render truthful context metadata/unknown state with capability-gated compaction; native/helper tests and `ux-context-002` browser acceptance (174/174 matrix; other context IDs remain open).
- [x] Implement shared queue return: merge latest origin text/media/refs, persist recovery before DELETE, retry without duplication and block DELETE on storage failure; Classic replacement conflict documented in ADR-0017.
- [x] Add separately reported shared-contract browser mappings; shared-28 passes six projects (192/192 executions overall), with incompatible Classic return cases still unmapped.
- [x] Complete run-bound atomic queue Steer (shared-30), disabled idle/unknown actions, at-most-once consumption and held recovery; 12/12 local-provider matrix, 192/192 existing matrix, 70/70 functional, Go/vet/race and 23 helper tests. See ADR-0018; Classic idle-send remains unmapped.
- [x] Verify measured context-fit rejection/accepted switch (compaction 006/007) through local-provider browser matrix, preserving drafts, native rejection and per-session usage; 12/12 context-fit, 216/216 combined browser, 70/70 functional, Go/vet and 23 helpers.
- [x] Verify context-meter formatting, tooltip data, clamped fill and boundary colours (context 001/005) through native local-provider measurements; 228/228 combined browser, 70/70 functional, Go/vet and 23 helpers. Three-size terminal footer/live regressions pass without UI changes; unsupported compaction cases stay unmapped.
- [x] Make automatic compaction lifecycle cancellation-safe and durable before UI wiring: atomic occurrence-keyed phase/event/summary updates, terminal suppression/failure outcomes, no cancelled-status overwrite. Go/vet, targeted race ×3, hook checks and 70/70 functional; ADR-0019. No browser parity credit or manual Compact capability added.
- [x] Wire automatic-compaction activity snapshot/Stop API and host status/meter adapters; compaction 001–005/context 004 pass across six projects with draft/media retention, reload, late-session and failed-stop acceptance. 276/276 combined browser, 70/70 functional, Go/vet/race and 24 helpers; ADR-0020. Manual Compact remains disabled.
- [x] Persist ID/fingerprint-owned eligible context checkpoints without deleting timeline; exact native projection/snapshot/prefix guards, atomic summary commit, cancellation/failure preservation and edit/delete fallback. 54/54 compaction, 282/282 combined browser, 70/70 functional, Go/vet/race, 24 helpers and three-size TUI regression; ADR-0021. No new frozen mapping.
- [x] Implement explicit idle-only manual Compact with atomic snapshot-token/claim admission, cancellable maintenance turn/no provider request and host callback preserving drafts/media; context-003, 66/66 compaction, 294/294 combined browser, 70/70 functional, Go/vet/race and 24 helpers. ADR-0022; terminal command remains informational.
- [x] Wire terminal /compact and draft-preserving Alt-C to shared maintenance admission; native scoped lifecycle/Stop, focused Escape fix, /compact info diagnostics. Live 60×18/100×22/140×36 gate/cancel/resize/reopen/zero-idle-row checks, unit/race, existing TUI suites and 70/70 functional pass; ADR-0023.
- [x] Verify reconnect refresh of native timeline/status/queue/context (002), fence pre-disconnect HTTP responses/errors, and surface real server asset-version drift without reload (004); 18/18 reconnect, 312/312 combined browser, 70/70 functional, Go/vet and 26 helpers. ADR-0024; search/pagination/full crash criteria stay open.
- [x] Implement bounded native message search/current-family-all scopes and supplied search composer; preserve draft/media and active view across real reconnect with query/session/connection guards; reconnect-003. 330/330 browser, 70/70 functional, Go/vet/race and 27 helpers; ADR-0025. Hashtag/paging and terminal search remain open.
- [x] Coordinate initial chat activation/SSE-ready refresh (reconnect-005) without pre-subscription snapshots or skipped reconnect; delayed readiness, failed initial read, A→B→A/native reconnect tests. 342/342 browser, 70/70 functional, Go/vet and 28 helpers; ADR-0026. All reconnect IDs mapped; broader crash/paging/hashtag gaps remain.
- [x] Implement stable bounded native message paging, promise-owned older loads and viewport anchoring; reconnect catches up across pages without replacing loaded history; old pages cannot overwrite search. 348/348 browser, 70/70 functional, Go/vet/race and 29 helpers; ADR-0027. No new frozen mapping.
- [x] Verify accepted-message refresh/visibility, native multi-upload ID/name/byte pairing, newer text/cursor, reader anchors and search/origin ownership; compose-007/009/010/011, 384/384 browser, 70/70 functional, Go/vet/hook and 29 helpers; ADR-0028.
- [ ] Establish file/folder/message reference selection and references-only submission for compose-008; folder selection currently only expands the supplied explorer. Terminal acceptance/recovery needs independent three-size, zero-idle-row tests.
- [ ] Complete remaining reconnect ownership and full browser compaction acceptance.
- [x] Implement per-session terminal editor/history state, generation-owned event/submit delivery and cancel-safe forwarding; unit/race coverage.
- [x] Bound the temporary terminal session selector to six results; live tmux verifies cancel/resize/zero added idle rows at 60×18, 100×22 and 140×36 (`make test-tui-sessions`).
- [ ] Complete remaining terminal adaptations in `docs/internal/web-session-parity.md` (mutation submenus/media/queue acceptance still separate).
- [ ] Continue component refreshes from a pinned upstream source where the suite demonstrates a gap.

Evidence and commands: [`../../tests/ux/README.md`](../../tests/ux/README.md). These results cover the mapped scenarios, not full Piclaw compliance.

## Phase 1 — minimal vertical slice

### Turn engine
- [x] Define append-only turn event model
- [x] Define turn state reconstruction from events
- [x] Persist turn start/progress/end events
- [x] Serialize turns per session
- [x] Allow concurrent turns across sessions
- [x] Implement queue/cancel state model
- [x] Implement cancellation state transitions (`running` → `cancelling` → `cancelled`)
- [x] Wire go-ai inference into turn engine
- [x] Load system prompt from workspace AGENTS.md
- [x] Stream inference tokens via go-ai `Stream()`
- [x] Broadcast Piclaw-compatible SSE events during inference
- [ ] Implement queue reorder
- [ ] Add centralized runtime budget config (tool calls per turn, retries, queue depth)

### Database/state model
- [x] Create SQLite schema baseline with JSON fields and expression indexes
- [x] Add sessions table
- [x] Add messages/content block tables
- [x] Add turn events/checkpoints tables
- [x] Add turns table with status tracking
- [x] Enable WAL and concurrency-safe pragmas
- [x] Auto-create default session on startup
- [ ] Add forks/ancestry table(s)
- [ ] Add schedules/background tasks tables
- [ ] Add attachments metadata tables
- [ ] Add settings/assets mirror tables

### Inference
- [x] Integrate go-ai as model/provider layer
- [x] Register builtin models (openai, openairesponses, anthropic)
- [x] Load auth tokens from `~/.pi/agent/auth.json`
- [x] GitHub Copilot token exchange (refresh → session token + enterprise endpoint detection)
- [x] Apply Copilot headers for IDE auth
- [x] Build conversation history from session messages
- [x] Streaming inference via `goai.Stream()`
- [x] Broadcast text deltas during streaming
- [ ] Token/cost tracking per turn
- [ ] Provider failover / retry with exponential backoff
- [ ] Context window overflow detection and compaction

### Web UI
- [x] Piclaw TypeScript source ported verbatim (199 files)
- [x] gi-specific `api.ts` adapter (same function signatures, maps to gi REST endpoints)
- [x] gi-specific `app.ts` entry point (uses Piclaw components verbatim)
- [x] Piclaw CSS stack served from `/css/styles.css`
- [x] Piclaw fonts vendored and embedded
- [x] Vendor bundles: preact-htm, marked, katex, beautiful-mermaid, codemirror
- [x] Load vendor scripts as ESM modules (scoped var declarations)
- [x] IIFE-wrapped app bundle to prevent global shadowing
- [x] Cache busters on all bundle URLs per server restart
- [x] `window.onerror` error boundary with backend log reporting
- [x] Import map for `#editor-vendor/codemirror`
- [x] Piclaw theme/tint runtime (presets, cycling, localStorage, system dark mode)
- [x] Timeline with Piclaw Post component (markdown rendering, avatars, file pills)
- [x] ComposeBox with Piclaw component (history, keyboard, slash commands, model picker)
- [x] AgentStatus with Piclaw component (draft/thought/plan panels)
- [x] WorkspaceExplorer with Piclaw component
- [x] TabStrip with Piclaw component
- [x] SSE streaming endpoint (`/sse/stream`) with Piclaw event model
- [x] Real SSEClient implementation (reconnection, heartbeat, event bindings)
- [x] Route-event-aware SSE subscriptions for `routing_decision`/`routing_incoming`
- [x] Frontend log bridge (`/api/frontend/log`)
- [x] Runtime config API (`/api/runtime/config`)
- [x] Workspace tree/file APIs (`/api/workspace/tree`, `/api/workspace/file`)
- [x] Route-event introspection API (`/api/sessions/{id}/route-events`) in frontend adapter
- [x] SSE-driven real-time timeline updates (wired but not yet consuming events in app.ts)
- [x] Streaming draft display in compose area

### Slash commands

The ComposeBox already has Piclaw's slash command autocomplete UI. The backend needs to handle them.

### System meters HUD

Direct port of Piclaw's `/meters` functionality. On by default until slash commands are implemented.

- [ ] Add `/api/system-metrics` endpoint (CPU, RAM, swap, RSS)
- [ ] Wire `SystemMetersHud` component in `app.ts` (already copied from Piclaw)
- [ ] Enable meters on by default (skip `/meters` toggle until slash commands exist)
- [ ] Poll metrics every 5s from the Go backend
- [ ] SSE `ui_meters` event for real-time updates

#### Session/model commands (Phase 1)
- [ ] `/model` — list available models or switch model
- [ ] `/cycle-model` — cycle to next available model
- [ ] `/thinking` — show or set thinking/effort level
- [ ] `/cycle-thinking` — cycle thinking level
- [ ] `/theme` — set UI theme
- [ ] `/tint` — tint default light/dark UI
- [ ] `/abort` — abort current response
- [ ] `/state` — show current session state
- [ ] `/stats` — show session token and cost stats
- [ ] `/context` — show context window usage
- [ ] `/last` — show last assistant response
- [ ] `/commands` — list available commands

#### Queue/steering commands (Phase 1)
- [ ] `/queue` — queue a follow-up message
- [ ] `/steer` — steer the current response
- [ ] `/abort-retry` — abort retry backoff

#### Session management commands (Phase 2)
- [ ] `/new-session` — start a new session
- [ ] `/session-name` — set or show the session name
- [ ] `/compact` — manually compact the session
- [ ] `/auto-compact` — toggle auto-compaction
- [ ] `/auto-retry` — toggle auto-retry
- [ ] `/fork` — fork from a previous message
- [ ] `/clone` — duplicate current branch into a new session
- [ ] `/tree` — list the session tree

#### Identity commands (Phase 2)
- [ ] `/agent-name` — set or show agent display name
- [ ] `/agent-avatar` — set or show agent avatar URL
- [ ] `/user-name` — set or show user display name
- [ ] `/user-avatar` — set or show user avatar URL

#### Tool commands (Phase 2)
- [ ] `/shell` — run a shell command and return output
- [ ] `/bash` — run a shell command and add output to context
- [ ] `/search` — search notes and skills in workspace
- [ ] `/skill:` — run a workspace skill

#### Auth/admin commands (Phase 3)
- [ ] `/login` — login to an AI model provider
- [ ] `/logout` — logout from a provider
- [ ] `/passkey` — manage passkeys
- [ ] `/totp` — show TOTP enrolment QR code
- [ ] `/restart` — restart the agent
- [ ] `/exit` — exit the process
- [ ] `/export-html` — export session to HTML
- [ ] `/tasks` — list scheduled tasks

### Pi/Piclaw config compatibility
- [x] Load `.piclaw/config.json` (assistant name/avatar, user name/avatar/background)
- [x] Load `.pi/settings.json` (provider, model, thinking level, enabled models)
- [x] Load `AGENTS.md` as system prompt
- [x] Load auth from `~/.pi/agent/auth.json`
- [x] Preserve Pi model/provider naming semantics

### Scripting engines
- [x] Goja (`js`) runtime implemented
- [x] Joker runtime implemented and bridge parity expanded
- [x] QuickJS deliberately out of scope (Goja is the supported JavaScript runtime)

### Testing
- [x] Go unit tests (store, turn, web, config)
- [x] Playwright base UX tests (13 tests)
- [x] Isolated test instance with dedicated DB and workspace (`make test-ux`)
- [x] Hook TDZ checker (`scripts/check-hook-tdz.ts`)
- [x] `go vet` / `go test ./...`
- [ ] CI pipeline (`.github/workflows/`)

### Developer tooling
- [x] Makefile with detached lifecycle management
- [x] `-bind`, `-port`, `-model`, `-log-file`, `-pid-file` CLI flags
- [x] `make help/bootstrap/deps/start/stop/restart/status/logs/run/clean`
- [x] `make test-ux` (isolated Playwright test instance)
- [x] `make build-web` / `make bun-checks` / `make check`
- [x] Bun build pipeline for vendor + app bundles

---

## Phase 2 — context, keychain, skills

### Context and recovery
- [ ] Implement compacted context rebuild from DB
- [ ] Carry forward last N messages on rollover
- [ ] Make N configurable
- [ ] Implement mid-turn compaction as same-turn recovery phase
- [ ] Emit UI hint for every compaction
- [ ] Checkpoint at compaction boundaries
- [ ] Implement provider retry checkpoints
- [ ] Preserve partial output on provider failure
- [ ] Auto-resume after clean restart
- [ ] No auto-resume after crash

### Keychain
- [ ] Port Piclaw encryption model
- [ ] Store encrypted secrets in SQLite
- [ ] Implement env/bootstrap unlock path
- [ ] Implement interactive unlock path
- [ ] Start in degraded mode when locked
- [ ] Emit unlock UI prompt when secret is needed
- [ ] Auto-resume blocked turn/tool after unlock
- [ ] Keep unlocked for process lifetime
- [ ] Support per-process unlock state

### Skills and hooks
- [ ] Define `SKILL.md` + frontmatter loader
- [ ] Implement SQLite managed VFS for skills/scripts/templates/packages
- [ ] Load skills from managed VFS first, then filesystem fallback where appropriate
- [ ] Migrate loose skill/script/template mirrors into the managed VFS
- [ ] Implement Joker-first hooks
- [ ] Add Go hook surface
- [ ] Implement required hook points
- [ ] Add packaged skill import baseline
- [ ] Investigate simple GitHub-sourced skill extraction flow

### Internal reference and self-extension docs
- [x] Create `docs/internal/` as the source tree for shipped internal reference docs
- [x] Add bootstrap docs for tools, scripting, VFS, and skills
- [ ] Document every built-in tool contract under `docs/internal/tools/`
- [ ] Document hook lifecycle and extension contracts under `docs/internal/hooks/`
- [ ] Embed `docs/internal/` as read-only runtime reference content (for example `vfs://reference/...`)
- [ ] Add a doc-maintenance check so new internal surfaces cannot land undocumented

---

## Phase 3 — UI parity and operational surfaces

### Web UI
- [x] Workspace browser (basic tree + file open)
- [x] Tab strip for open files
- [ ] Editor panes with CodeMirror (read-only pane exists, needs full editing)
- [ ] Diff view
- [ ] Split panes
- [x] File pills in timeline
- [ ] Interactive widgets
- [ ] Inline charting
- [ ] Search UI
- [ ] Schedule management commands
- [ ] Messages inspection UI

### TUI
- [ ] Evaluate `go-tui` fit/gaps
- [ ] Basic chat surface
- [ ] Good scrollback
- [ ] Expandable input
- [ ] Status/progress parity subset
- [ ] Forms support
- [ ] Advanced-terminal image preview path
- [ ] Session commands/search/schedules

### CLI
- [ ] Prompt submission
- [ ] Create/list/resume sessions
- [ ] Schedule commands
- [ ] Search/messages inspection
- [ ] Keychain management
- [ ] Direct script execution
- [ ] Maintenance/admin commands
- [ ] Prune state commands

---

## Phase 4 — storage, search, artifacts, observability

### Search and indexing
- [x] Decide hybrid workspace search direction: SQLite FTS5 + sqlite-vec + gte-go
- [x] Add ADR for workspace hybrid search (`adr/0008-workspace-hybrid-search.md`)
- [x] Add internal design doc for hybrid search/indexing (`docs/internal/search/README.md`)
- [x] Scaffold `internal/search/` package layout (service, rank, query, chunking, embed, vector, store, indexer)
- [x] Scaffold workspace search schema definitions (`workspace_documents`, `workspace_chunks`, `workspace_chunks_fts`, `workspace_index_meta`)
- [ ] Wire workspace search schema into main DB initialization
- [ ] Implement real `gte-go` embedder
- [ ] Implement real `sqlite-vec` backend
- [ ] Implement SQLite FTS strategy for workspace chunks
- [ ] Index messages
- [ ] Index notes/memory
- [ ] Index skills/scripts/templates/assets
- [ ] Index attachment filename/metadata/text
- [ ] Expose search via tool + CLI + UI

### Artifacts/attachments
- [ ] Define workspace-path artifact model
- [ ] Define selected DB-blob rules for images/generated outputs
- [ ] Implement attachment import/export/read flows
- [ ] Implement thumbnail generation table and cleanup
- [ ] Investigate image processing library choice
- [ ] Investigate Piclaw-compatible artifact metadata/FTS behavior

### Observability
- [ ] Define structured event schema
- [ ] Add counters tables/materialized aggregates as needed
- [ ] Record provider/tool/script/schedule/session events
- [ ] Add token/cost usage events
- [ ] Add compaction/rotation/recovery events
- [ ] Expose status views in UI/CLI

---

## Phase 5 — compatibility and hardening

### Pi/Piclaw compatibility
- [x] Preserve settings semantics
- [x] Preserve model/provider naming
- [ ] Preserve prompt template semantics
- [x] Preserve structured message/content model (via Piclaw components)
- [ ] Preserve intents/queue/steer conventions
- [ ] Preserve keychain env injection semantics

### Testing
- [x] Unit tests (Go)
- [ ] Fake-server provider tests
- [x] DB integration tests
- [ ] Golden tests for rendering/message projection
- [x] E2E web smoke tests (Playwright)
- [ ] E2E CLI smoke tests
- [x] E2E TUI smoke tests (tmux harness)
- [x] `go vet`
- [ ] Fuzzing

---

## Explicit investigation items

- [ ] Investigate exact Piclaw transcript/message export path for future import tools
- [ ] Investigate Piclaw artifact pill/open-editor behavior for parity
- [ ] Investigate Piclaw compaction and rollover implementation details for faithful compatibility
- [ ] Investigate `go-tui` capabilities vs required parity surface
- [ ] Investigate image processing library options for thumbnails/previews
- [ ] Investigate simple packaged-skill download/extract flow from GitHub URLs
- [x] Investigate embedded web asset pipeline with plain JS + Bun bundling only at build time
- [x] Port Piclaw TypeScript web source verbatim
