# ADR 0004: UI Surface Model

- Status: Draft
- Date: 2026-04-22

## Decision

Web UI is canonical.

Gi must provide:
- **web UI** with functionality equivalent to Piclaw, implemented as plain JS with a small vendored **Preact** layer
- **TUI** with shared interaction concepts and minimalistic rendering
- **CLI** for batch, maintenance, and operations

## Shared concepts

All surfaces should understand shared concepts for:
- status/progress
- intents
- session operations
- streaming output
- slash commands
- message search
- schedules

The web asset pipeline should follow the Vibes/Piclaw pattern:
- Bun allowed only at build time
- embedded static assets in the Go binary
- vendored browser dependencies checked into the repo/build outputs

## Web-specific requirements

- pane management
- workspace browser
- editor panes openable by agent
- interactive widgets
- inline charting
- file pills to open referenced files
- live SSE-driven updates

## TUI requirements

- good scrollback
- minimal formatting
- expandable input field
- status/progress parity where practical
- forms support
- image previews in capable terminals

## Non-requirement

Adaptive Cards are not part of Gi v1.
