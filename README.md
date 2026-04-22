# Gi

<img src="docs/icon-256.png" width="128" alt="Gi">

A coding agent built on `go-ai`, informed by lessons learned from Pi, Piclaw, and Vibes.

## Status

**Phase 1 functional** — turn engine, inference, web UI, and Playwright tests working.

The web UI uses Piclaw's TypeScript source verbatim with a Gi-specific API adapter and entry point. Inference runs through `go-ai` with GitHub Copilot enterprise token exchange.

## Goals

- unified **web / TUI / CLI** experience
- boringly reliable **turn handling** via append-only event log
- workspace-centric operation with **SQLite-backed state**
- **Piclaw compatibility** for settings, message model, keychain, prompt templates, and UX conventions
- **Joker-first** scripting and skills
- **go-ai** as the model/provider layer

## Architecture

- `cmd/gi/` — main binary entrypoint
- `internal/config/` — Pi/Piclaw config loader (settings, auth, AGENTS.md)
- `internal/store/` — SQLite state store (sessions, messages, turns, events)
- `internal/turn/` — append-only turn engine with queue/cancel/streaming
- `internal/inference/` — go-ai inference with provider auth and SSE broadcasting
- `internal/web/` — HTTP server, REST API, SSE streaming, workspace file APIs
- `web/src/` — Piclaw TypeScript web source (verbatim) + Gi `api.ts`/`app.ts` adapters
- `docs/` — ADRs, implementation checklist, transcripts
- `scripts/` — build/check scripts (hook TDZ checker)
- `tests/` — Playwright base UX tests

## Development

### Prerequisites

- Go 1.22+
- Bun (build-time only, not runtime)
- Playwright + Chromium (for `make test-ux`)

### Targets

| Target | Description |
|---|---|
| `make start` | Build and start Gi detached on port 8090 |
| `make stop` | Stop the detached process |
| `make restart` | Restart it |
| `make status` | Show status/listener |
| `make logs` | Tail the log file |
| `make run` | Foreground run |
| `make build` | Build binary (includes `build-web`) |
| `make build-web` | Bundle web assets via Bun |
| `make test` | Go unit tests |
| `make test-ux` | Playwright tests against isolated instance |
| `make vet` | Go vet |
| `make bun-checks` | Hook TDZ checker |
| `make clean` | Remove build/run artifacts |

### Override defaults

```sh
make start PORT=3000 BIND=0.0.0.0 MODEL=github-copilot/gpt-5-mini WORKSPACE=/workspace
```

### CLI flags

| Flag | Default | Description |
|---|---|---|
| `-listen` | (none) | Full listen address, overrides bind/port |
| `-bind` | `127.0.0.1` | Bind host/interface |
| `-port` | `8081` | HTTP port |
| `-model` | (from settings) | Override default model |
| `-db` | `./gi.db` | SQLite database path |
| `-workspace` | `/workspace` | Workspace root |
| `-log-file` | (none) | Log file path |
| `-pid-file` | (none) | PID file path |

## Web UI

The web UI uses **Piclaw's TypeScript source files verbatim** (199 files). Only two files are Gi-specific:

- `web/src/api.ts` — API adapter implementing Piclaw's function signatures against Gi's REST endpoints
- `web/src/app.ts` — Entry point wiring Gi sessions into Piclaw's component tree

All Piclaw components, theme runtime, CSS, icons, and vendor libraries work without modification.

### Vendored libraries

| Library | Path |
|---|---|
| Preact + HTM | `/js/vendor/preact-htm.js` |
| Marked | `/js/marked.min.js` |
| KaTeX | `/js/vendor/katex.min.js` |
| Beautiful Mermaid | `/js/vendor/beautiful-mermaid.js` |
| CodeMirror | `/editor-vendor/codemirror.js` |

### SSE streaming

The server provides a Piclaw-compatible SSE endpoint at `/sse/stream?chat_jid=...` that broadcasts:
- `connected`, `heartbeat`
- `agent_status`, `agent_draft_delta`, `agent_thought_delta`
- `new_post`, `agent_response`

## Inference

Gi uses `go-ai` for model inference. Supported providers:

- OpenAI (completions + responses)
- Anthropic
- GitHub Copilot (with automatic enterprise/individual endpoint detection)

Auth is loaded from `~/.pi/agent/auth.json`. The system prompt is loaded from `AGENTS.md` in the workspace root.

## Testing

```sh
make test       # Go unit tests
make test-ux    # 13 Playwright tests against isolated fresh instance
make vet        # go vet
make bun-checks # hook TDZ checker
```

The `test-ux` target creates a completely isolated test environment with its own database, workspace, and config — no state leaks between test runs.

## License

TBD
