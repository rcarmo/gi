# Gi implementation checklist

Status: Draft
Date: 2026-04-22

This checklist is organized by **subsystem** and grouped by **phase**.

---

## Phase 1 — minimal vertical slice

### Turn engine
- [x] Define append-only turn event model
- [x] Define turn state reconstruction from events
- [x] Persist turn start/progress/end events
- [x] Serialize turns per session
- [x] Allow concurrent turns across sessions
- [ ] Implement queue/reorder/cancel state model
- [ ] Implement cancellation state transitions (`running` → `cancelling` → `cancelled`)

### Database/state model
- [x] Create SQLite schema baseline
- [x] Add sessions table
- [ ] Add forks/ancestry table(s)
- [x] Add messages/content block tables
- [x] Add turn events/checkpoints tables
- [ ] Add schedules/background tasks tables
- [ ] Add attachments metadata tables
- [ ] Add settings/assets mirror tables
- [x] Enable WAL and concurrency-safe pragmas

### Web UI shell
- [x] Boot plain-JS web app from embedded assets
- [x] Implement session list/open/create
- [x] Implement basic chat timeline rendering
- [x] Implement compose box with Piclaw-style progress/status area
- [ ] Implement streaming output rendering
- [x] Implement one prompt → one turn flow
- [x] Implement clean control return after turn

### Tools
- [x] Implement `shell` minimal path with streaming output and cancellation
- [ ] Implement `read`
- [ ] Implement `write`
- [ ] Implement `edit` baseline
- [ ] Implement `messages` read/search baseline
- [ ] Implement `exit`
- [ ] Implement `script` baseline with Joker execution

### Integration slice
- [x] Start web service
- [x] Open/create session
- [x] Send one prompt
- [x] Run one shell tool call
- [ ] Stream progress
- [x] Persist all messages/events in SQLite
- [x] Return control cleanly

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
- [ ] Workspace browser
- [ ] Editor panes openable by agent
- [ ] Diff view
- [ ] Split panes/tabs
- [ ] File pills in timeline
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
- [ ] Preserve settings semantics
- [ ] Preserve model/provider naming
- [ ] Preserve prompt template semantics
- [ ] Preserve structured message/content model
- [ ] Preserve intents/queue/steer conventions
- [ ] Preserve keychain env injection semantics

### Testing
- [ ] Unit tests
- [ ] Fake-server provider tests
- [ ] DB integration tests
- [ ] Golden tests for rendering/message projection
- [ ] E2E web smoke tests
- [ ] E2E CLI smoke tests
- [ ] E2E TUI smoke tests
- [ ] `go vet`
- [ ] Fuzzing

---

## Explicit investigation items

- [ ] Investigate exact Piclaw transcript/message export path for future import tools
- [ ] Investigate Piclaw artifact pill/open-editor behavior for parity
- [ ] Investigate Piclaw compaction and rollover implementation details for faithful compatibility
- [ ] Investigate `go-tui` capabilities vs required parity surface
- [ ] Investigate image processing library options for thumbnails/previews
- [ ] Investigate simple packaged-skill download/extract flow from GitHub URLs
- [ ] Investigate embedded web asset pipeline with plain JS + Bun bundling only at build time
