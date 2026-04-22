# Gi

A coding agent built on top of `go-ai`, informed by lessons learned from Pi, Piclaw, and Vibes.

## Status

Early architecture/spec drafting.

## Goals

- unified **web / TUI / CLI** experience
- boringly reliable **turn handling**
- workspace-centric operation with **SQLite-backed state**
- **Piclaw compatibility** for settings, message model, keychain, prompt templates, and UX conventions where required
- **Joker-first** scripting and skills

## Initial repo layout

- `cmd/gi/` — main binary entrypoint
- `internal/` — core runtime packages
- `docs/` — ADRs, implementation checklist, transcripts
- `web/` — plain-JS web assets/build inputs

See `docs/README.md`.
