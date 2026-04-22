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

## Development targets

Detached lifecycle management lives in the `Makefile`.

- `make start` — build and start Gi detached on port `8090`
- `make stop` — stop the detached process
- `make restart` — restart it
- `make status` — show status/listener
- `make logs` — tail the log file
- `make run` — foreground run

Override defaults as needed:

```sh
make start PORT=3000 BIND=0.0.0.0 WORKSPACE=/workspace
```

### Relevant CLI flags

- `-listen :8090` — explicit full listen address, overrides bind/port
- `-bind 0.0.0.0` — bind host/interface
- `-port 8090` — listen port
- `-db .gi-run/gi.db`
- `-workspace /workspace`
- `-log-file .gi-run/gi.log`
- `-pid-file .gi-run/gi.pid`
