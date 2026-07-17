# gi

<img src="docs/icon-256.png" width="128" alt="gi">

A coding agent built on `go-ai`, informed by lessons learned from Pi, Piclaw, and Vibes.

## Status

**Phase 1 functional** — turn engine, inference, web UI, and Playwright tests working.

The web UI uses Piclaw's TypeScript source verbatim with a gi-specific API adapter and entry point. Inference runs through `go-ai` with GitHub Copilot enterprise token exchange.

## Goals

- unified **web / TUI** experience from one binary, with CLI startup/configuration flags
- boringly reliable **turn handling** via append-only event log
- workspace-centric operation with **SQLite-backed state**
- **Piclaw-oriented compatibility** for settings, message model, injected keychain references, prompt templates, and UX conventions (unsupported web APIs remain explicit adapter no-ops)
- **Clojure-first** scripting and skills via Joker
- **go-ai** as the model/provider layer

## Architecture

- `cmd/gi/` — main binary entrypoint for web server or TUI mode (`-tui`)
- `internal/config/` — Pi/Piclaw config loader (settings, auth, AGENTS.md)
- `internal/store/` — SQLite state store (sessions, messages, turns, events)
- `internal/turn/` — append-only turn engine with queue/cancel/streaming
- `internal/inference/` — go-ai inference with provider auth and SSE broadcasting
- `internal/web/` — HTTP server, REST API, SSE streaming, workspace file APIs
- `web/src/` — Piclaw TypeScript web source (verbatim) + gi `api.ts`/`app.ts` adapters
- `docs/` — ADRs, internal shipped-reference source docs, implementation checklist, transcripts
- `scripts/` — build/check scripts (hook TDZ checker)
- `tests/` — Playwright base UX tests

## Internal reference

The repo includes a growing internal documentation subtree under `docs/internal/`.

This is shipped in the binary as the read-only `vfs://reference/...` surface for the agent itself (for example `vfs://reference/README.md` and `vfs://reference/tools/read.md`).

If a change adds or materially changes an internal tool, scripting bridge capability, hook, managed `vfs://` behavior, or skill/package contract, the same change should update `docs/internal/`.

## Development

### Prerequisites

- Go 1.22+
- Bun (build-time only, not runtime)
- Playwright + Chromium (for `make test-ux`)

### Targets

Start with:

```sh
make bootstrap
```

That installs Go/Bun dependencies, installs Playwright Chromium, and builds `gi` on a fresh machine.

| Target | Description |
|---|---|
| `make help` | Show the grouped target list |
| `make bootstrap` | Install dependencies, install Playwright Chromium, and build `gi` |
| `make deps` | Download Go modules and install Bun packages |
| `make start` | Build and start gi detached on port 8090 |
| `make stop` | Stop the detached process |
| `make restart` | Restart it |
| `make status` | Show status/listener |
| `make logs` | Tail the log file |
| `make run` | Foreground run |
| `make build` | Build the main `gi` binary (includes `build-web`) |
| `make build-web` | Bundle web assets via Bun |
| `make test` | Go unit tests |
| `make vet` | Go vet |
| `make bun-checks` | Hook TDZ checker |
| `make check` | Run the standard verification suite |
| `make test-ux` | Playwright tests against isolated instance (artifacts under `test-results/`) |
| `make test-tui-smoke` | tmux-driven TUI smoke test (artifacts under `test-results/tui-smoke/`) |
| `make test-tui-gherkin` | TUI gherkin harness |
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
| `-tls-cert` / `-tls-key` | (none) | TLS certificate/private-key files |
| `-acme-domains` | (none) | Comma-separated domains for ACME HTTPS |
| `-acme-email` | (none) | ACME registration contact |
| `-acme-cache` | `sqlite` | ACME cache backend: sqlite, vfs, or directory |
| `-acme-accept-tos` | `false` | Accept ACME CA terms |
| `-acme-http-listen` | `:http` | ACME HTTP-01/redirect listener; empty disables |
| `-tui` | `false` | Run the terminal UI instead of the web server |
| `-db` | `./gi.db` | SQLite database path |
| `-workspace` | `/workspace` | Workspace root |
| `-log-file` | (none) | Log file path |
| `-pid-file` | (none) | PID file path |

### TUI mode

Run the terminal UI from the same binary:

```sh
gi -tui -db .gi-run/gi.db -workspace /workspace
```

The current TUI uses `go-tui`, supports terminal resize handling through the runtime event loop, and enables mouse clicks so the input can regain focus.

## Web UI

The web UI uses **Piclaw's TypeScript source files verbatim** (199 files). Only two files are gi-specific:

- `web/src/api.ts` — API adapter implementing Piclaw's function signatures against gi's REST endpoints
- `web/src/app.ts` — Entry point wiring gi sessions into Piclaw's component tree

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

gi uses `go-ai` for model inference. Supported providers:

- OpenAI (completions + responses)
- Anthropic
- GitHub Copilot (with automatic enterprise/individual endpoint detection)

Auth is loaded from `~/.pi/agent/auth.json`. The system prompt is loaded from `AGENTS.md` in the workspace root.

## Testing

```sh
make test       # Go unit tests
make test-ux    # Playwright tests against an isolated fresh instance
make vet        # go vet
make bun-checks # hook TDZ checker
make check      # standard verification suite
```

The `test-ux` target creates a completely isolated test environment with its own database, workspace, and config — no state leaks between test runs.

The `test-tui-smoke` target launches `gi -tui` inside tmux, captures the pane, submits input, verifies blur handling, exercises transcript scrolling keys, resizes the terminal, and writes pane captures plus session artifacts under `test-results/tui-smoke/`. Mouse click focus is covered in unit tests.

## License

TBD
