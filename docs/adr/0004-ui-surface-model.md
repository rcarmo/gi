# ADR 0004: UI Surface Model

- Status: **Active**
- Date: 2026-04-22
- Last updated: 2026-04-22

## Decision

Web UI is canonical.

gi uses **Piclaw's TypeScript web source verbatim** (199 files). Only two files are gi-specific:
- `api.ts` — API adapter implementing Piclaw's function signatures against gi's REST endpoints
- `app.ts` — entry point wiring gi sessions into Piclaw's component tree

The web asset pipeline uses Bun at build time only:
- vendor bundles built from entry files matching Piclaw's vendor manifests
- app bundle built as ESM, IIFE-wrapped to prevent global var shadowing
- all assets embedded in the Go binary via `embed.FS`
- cache busters injected per server restart

## Shared concepts

All surfaces should understand shared concepts for:
- status/progress
- intents
- session operations
- streaming output
- slash commands
- message search
- schedules

## Web-specific requirements

- pane management (TabStrip, WorkspaceExplorer — using Piclaw components)
- workspace browser (using Piclaw WorkspaceExplorer)
- editor panes openable by agent (via CodeMirror vendor bundle)
- interactive widgets (FloatingWidgetPane — Piclaw component)
- inline charting
- file pills to open referenced files (FilePill — Piclaw component)
- live SSE-driven updates (`/sse/stream` with Piclaw event model)

## SSE event model

gi implements a Piclaw-compatible SSE endpoint at `/sse/stream` that broadcasts:
- `connected` — initial connection confirmation
- `heartbeat` — 15-second keep-alive
- `agent_status` — turn status updates
- `agent_draft_delta` — streaming text tokens
- `agent_thought_delta` — streaming thinking tokens
- `new_post` — completed message
- `agent_response` — agent turn completion

## TUI requirements

- good scrollback
- minimal formatting
- expandable input field
- status/progress parity where practical
- forms support
- image previews in capable terminals

## Non-requirement

Adaptive Cards are not part of gi v1.
