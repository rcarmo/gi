# ADR 0001: Overall Architecture for gi

- Status: **Active**
- Date: 2026-04-22
- Last updated: 2026-04-22

## Context

gi is a coding agent built on `go-ai`, incorporating lessons learned from Pi, Piclaw, and Vibes. The primary pain points to address are upstream churn, unreliable turn/session handling, brittle compaction/retry behavior, and maintenance cost.

## Decision

gi uses:
- **Go** for the core runtime
- **go-ai** as the model/provider layer (with streaming inference)
- **Piclaw TypeScript source** for the web UI (verbatim, with gi-specific API/entry adapters only)
- **SQLite (WAL)** as the shared state store across web, TUI, and CLI processes
- **Bun** for build-time web asset bundling only (not runtime)

### Runtime architecture
- **single binary** with `-bind`, `-port`, `-model`, `-workspace` flags
- long-running **web mode** under supervisor/systemd
- **CLI** and **TUI** as separate processes sharing the SQLite database
- **workspace-centric** operation with a single workspace root

### Web UI architecture
- Piclaw's 199 TypeScript source files copied verbatim
- Only `api.ts` (API adapter) and `app.ts` (entry point) are gi-specific
- Vendor libraries (preact, marked, katex, mermaid, codemirror) built at compile time
- All assets embedded in the Go binary via `embed.FS`
- SSE streaming for real-time events (Piclaw-compatible event model)
- Cache busters on all bundle URLs per server restart

### Inference architecture
- go-ai `Stream()` for token-by-token streaming
- Auth loaded from `~/.pi/agent/auth.json`
- GitHub Copilot: automatic token exchange with enterprise endpoint detection
- System prompt loaded from workspace `AGENTS.md`
- Conversation history built from session messages

### Turn engine architecture
- Append-only event log with checkpoints at tool-call and provider boundaries
- Serialized turns per session, concurrent across sessions
- Queue/cancel state model
- SSE broadcasting of Piclaw-compatible events during inference

## Consequences

### Positive
- near-zero UI maintenance: Piclaw updates can be dropped in with no diff on gi's side
- low dependency footprint (Go + SQLite + embedded assets)
- strong portability via binary + database
- durable observability and resumability via event log
- real streaming UX from day one via SSE

### Negative
- Piclaw UI code is large (~199 files) and not all features are wired yet
- some Piclaw components expect APIs that gi stubs gracefully but doesn't implement
- Bun required at build time for web asset compilation

## Notes

See:
- `0002-turn-engine-and-recovery.md`
- `0003-state-and-storage.md`
- `0004-ui-surface-model.md`
- `0005-tools-skills-and-scripting.md`
- `../checklists/implementation.md`
- `../reference/chat-transcript-2026-04-22-gi-spec.md`
