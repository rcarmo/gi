# Gi implementation checklist

Status: Active
Date: 2026-04-22
Last updated: 2026-04-22

This checklist is organized by **subsystem** and grouped by **phase**.

---

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
- [x] Gi-specific `api.ts` adapter (same function signatures, maps to Gi REST endpoints)
- [x] Gi-specific `app.ts` entry point (uses Piclaw components verbatim)
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
- [x] Frontend log bridge (`/api/frontend/log`)
- [x] Runtime config API (`/api/runtime/config`)
- [x] Workspace tree/file APIs (`/api/workspace/tree`, `/api/workspace/file`)
- [x] SSE-driven real-time timeline updates (wired but not yet consuming events in app.ts)
- [x] Streaming draft display in compose area

### Slash commands

The ComposeBox already has Piclaw's slash command autocomplete UI. The backend needs to handle them.

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
- [x] `make start/stop/restart/status/logs/run/clean`
- [x] `make test-ux` (isolated Playwright test instance)
- [x] `make build-web` / `make bun-checks`
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
- [ ] Load skills from DB mirror first, then filesystem
- [ ] Mirror skills/scripts/prompt templates into DB
- [ ] Implement Joker-first hooks
- [ ] Add Go hook surface
- [ ] Implement required hook points
- [ ] Add packaged skill import baseline
- [ ] Investigate simple GitHub-sourced skill extraction flow

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
- [ ] Implement SQLite FTS strategy
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
- [ ] E2E TUI smoke tests
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
