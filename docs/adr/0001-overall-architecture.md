# ADR 0001: Overall Architecture for Gi

- Status: Draft
- Date: 2026-04-22

## Context

Gi is a coding agent built on `go-ai`, incorporating lessons learned from Pi, Piclaw, and Vibes. The primary pain points to address are upstream churn, unreliable turn/session handling, brittle compaction/retry behavior, and maintenance cost.

The user requires:
- unified web/TUI/CLI product
- web UI as canonical surface
- Piclaw compatibility for settings, message model, keychain semantics, prompt templates, and core UX conventions
- a simpler, more reliable runtime architecture centered on an append-only event log

## Decision

Gi will use:
- **Go** for the core runtime
- **go-ai** as the model/provider layer
- **plain JS** for the web UI
- **Joker** as the primary scripting/hook language in v1
- **SQLite (WAL)** as the shared state store across web, TUI, and CLI processes

The product architecture is:
- **single binary** with separate modes/process roles
  - long-running **web mode** under supervisor/systemd
  - **CLI** as an operational interface
  - **TUI** as an interactive terminal interface
- **workspace-centric** operation with a single workspace root
- **shared structured message model** compatible with Piclaw
- **event-log-based turn execution** with checkpoints and resumability

## Consequences

### Positive
- simpler operational model
- low dependency footprint
- strong portability via binary + database
- durable observability and resumability
- easier UI parity because web is canonical and TUI/CLI map onto shared concepts

### Negative
- SQLite schema and event model must be designed carefully up front
- Piclaw compatibility increases initial scope
- some assets/state mirrored in DB and filesystem adds complexity

## Notes

See:
- `0002-turn-engine-and-recovery.md`
- `0003-state-and-storage.md`
- `0004-ui-surface-model.md`
- `0005-tools-skills-and-scripting.md`
- `../reference/chat-transcript-2026-04-22-gi-spec.md`
