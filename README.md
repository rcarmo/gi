# gi

<img src="docs/icon-256.png" width="128" alt="gi">

A coding agent built on `go-ai`, informed by lessons learned from Pi, Piclaw, and Vibes.

## Status

Gi runs a web UI and a terminal UI from one pure-Go binary, with embedded web assets and SQLite-backed sessions, messages and turns. Bun is needed to build the browser assets, but there is no Node or Bun runtime dependency.

Piclaw parity is partial. Gi reuses pinned Piclaw components and implements its own backend, adapters and host UI. Composer/picker layout and several keyboard journeys still need work; sharing component source does not make the applications interchangeable. See the dated [feature and parity matrix][parity] for what works, what differs and what is planned.

## Features

* Streaming chat, session-local models, durable follow-up queues and run-bound steering use the same Go turn engine. Reconnect refreshes authoritative state; automatic and manual compaction retain the visible conversation.
* The browser has durable drafts and attachments, session selection and management, scoped search, Markdown/code rendering, image lightboxes and read-only workspace tabs. Settings covers models, appearance, instance identity, compaction and OpenAI/Anthropic API keys; it does not yet cover every Piclaw pane.
* The terminal offers fullscreen transcript navigation or native-scrollback mode, draft-preserving session/model selectors, compact session actions and file attachments. The regular-mode idle editor/footer uses five rows; selectors and actions are temporary.
* Single-user TOTP sign-in uses an HttpOnly browser cookie and transactional auth storage. Tools and extensions run through Go, embedded Joker/JavaScript or explicitly configured subprocesses, with workspace files and managed `vfs://` references.

**Passkeys are opt-in, with login and Settings enrolment controls.** Settings can add, list, rename and remove credentials after recent authentication. Browser tests verify two independent keys after restart, further enrolment without TOTP, cancellation and lockout-safe removal. Physical-device and synced-key checks are still outstanding. See the [backend contract](docs/internal/passkeys.md) for configuration and limits. tsnet web access, Iroh inter-instance chat and a token-saving MCP gateway are planned; the tsnet manager scaffold has no HTTP listener wiring. These integrations keep Gi's runtime pure Go and add no idle terminal rows.

## Goals

- unified **web / TUI** experience from one binary, with CLI startup/configuration flags
- boringly reliable **turn handling** via append-only event log
- workspace-centric operation with **SQLite-backed state**
- **Piclaw-oriented compatibility** for settings, message models, injected keychain references and UX conventions, with unsupported APIs and workflow gaps documented explicitly
- **Clojure-first** scripting and skills via Joker
- **go-ai** as the model/provider layer

## Architecture

- `cmd/gi/` — main binary entrypoint for web server or TUI mode (`-tui`)
- `internal/config/` — Pi/Piclaw config loader (settings, auth, AGENTS.md)
- `internal/store/` — SQLite state store (sessions, messages, turns, events)
- `internal/turn/` — append-only turn engine with queue/cancel/streaming
- `internal/inference/` — go-ai inference with provider auth and SSE broadcasting
- `internal/web/` — HTTP server, REST API, SSE streaming, workspace file APIs
- `web/src/` -- pinned Piclaw components plus Gi API, host, auth and Settings adapters
- `docs/` — ADRs, internal shipped-reference source docs, implementation checklist, transcripts
- `scripts/` — build/check scripts (hook TDZ checker)
- `tests/` -- Go-backed functional, browser parity and terminal acceptance tests

## Terminal rendering

`gi -tui -tui-mode regular` keeps completed output in native terminal scrollback with a five-row idle editor/footer; the terminal owns wheel, selection and copy. `fullscreen` remains the default and supports in-app paging and tool-output folding. Regular mode prints retained output fully expanded after native completion and uses a temporary three-row preview while active. See [ADR-0031](docs/adr/0031-regular-terminal-scrollback.md) for retention and resize limits.

Fullscreen `Ctrl+Shift+F` searches rendered transcript rows without submitting a prompt. Enter/Shift+Enter move between matches; Escape restores the draft and reading position. `Ctrl+Shift+Up/Down` jump between user prompts. See [ADR-0032](docs/adr/0032-fullscreen-transcript-search.md) for rendered-row and retention limits.

In fullscreen mode, drag to select transcript text and hold at a viewport edge to scroll. Release or Ctrl-C/Ctrl-X copies through `tuiClipboardMode`; Escape clears selection. Clipboard-off is respected. Feedback occupies the existing separator row. See [ADR-0034](docs/adr/0034-fullscreen-transcript-selection.md).

## Internal reference

The repo includes a growing internal documentation subtree under `docs/internal/`.

This is shipped in the binary as the read-only `vfs://reference/...` surface for the agent itself (for example `vfs://reference/README.md` and `vfs://reference/tools/read.md`).

If a change adds or materially changes an internal tool, scripting bridge capability, hook, managed `vfs://` behavior, or skill/package contract, the same change should update `docs/internal/`.

## Development

### Prerequisites

* Go 1.26.8 or newer, as declared in `go.mod`
* Bun (build-time only, not runtime)
* Playwright + Chromium for functional tests; Chromium and WebKit for the six-project parity matrix
* tmux for the terminal acceptance scripts

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
| `make check` | Go tests, vet, web build, hook checks and functional browser tests; excludes the parity matrix |
| `make test-ux` | Functional browser/API tests against an isolated instance |
| `make test-ux-journey` | Empty-store startup/new-chat/Return and retry journeys; Chromium/WebKit at three sizes; required CI gate |
| `make ux-parity-inventory` | Check frozen feature hashes and generate the scenario inventory |
| `make test-ux-parity` | Default browser parity suite in Chromium/WebKit at three viewport sizes; specialised suites have separate targets/flags |
| `make test-ux-auth` | Isolated TOTP/browser-auth regression suite |
| `make test-ux-passkeys` | Real Chromium WebAuthn API and Settings/login journeys at three sizes; virtual authenticators, no physical-device claim |
| `make check-cross-build` | Optional local pure-Go builds for Linux/macOS amd64/arm64 and Windows amd64; Windows is excluded from CI |
| `make test-tui-smoke` | tmux-driven TUI smoke test (artifacts under `test-results/tui-smoke/`) |
| `make test-tui-gherkin` | TUI gherkin harness |
| `make test-tui-regular` | Three-size native scrollback, selection/copy, draft/resize/session/exit/reopen checks |
| `make test-tui-search` | Three-size fullscreen search, prompt-jump, draft/cursor and live-output checks |
| `make test-tui-selection` | Three-size native drag/copy/edge-scroll/clipboard-policy and stale-selection checks |
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
| `-tui-mode` | `fullscreen` | Terminal-owned scrollback (`regular`) or in-app transcript (`fullscreen`) |
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

The web UI reuses pinned Piclaw component sources. Gi supplies `web/src/api.ts`, `web/src/app.ts`, auth/Settings modules and CSS overrides; narrowly guarded build adapters also change selected bundled behaviour without editing supplied components. Component provenance and runtime parity are separate checks.

Workspace tabs are read-only previews: editable documents, dirty-buffer workflows, popouts and docking are not implemented. Current composer/picker styling differs from Piclaw. The reproduced startup/new-chat focus and loading-retry failures are fixed and covered by [first-Return journeys](docs/internal/startup-return-journeys.md); broader keyboard and visual parity remains open. The [UX audit][audit] records those gaps and the limits of existing tests.

### Authentication and exposure

TOTP browser sign-in is available after owner enrolment through the loopback-only API. Browser cookies require direct TLS or a loopback peer and Host, with same-origin checks. Configured passkeys can sign in and be managed under Settings > Authentication. Settings also provides revision-checked TOTP-only, passkey-only or either sign-in policy, refusing changes that leave no usable factor. Settings can explicitly sign out the current browser without clearing local drafts or other sessions. [Initial owner setup in Settings](docs/internal/browser-bootstrap.md) uses a short-lived manual TOTP key and creates the owner/session atomically on loopback, with explicit recovery after uncertain responses. Family accounts and broader session-management UI are not implemented.

**The application permits access before enrolment.** Keep an unconfigured instance on loopback or a protected network. The CLI defaults to loopback, but `make start` defaults to `BIND=0.0.0.0`; use `make start BIND=127.0.0.1` for local development. TLS support alone does not enrol an owner or enable authentication.

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
make test-ux    # functional browser/API tests against an isolated fresh instance
make vet        # go vet
make bun-checks # hook TDZ checker
make check      # standard verification suite
```

The `test-ux` target creates a fresh database, workspace and configuration for each run. It excludes `tests/ux/`, which has a separate runner and specialised fixture targets; see [the browser suite guide][ux].

The frozen catalogue contains 236 Classic scenario IDs (256 expanded cases) and 42 shared cases. Mapped scenarios and successful test executions are tracked separately in the [parity matrix][parity]. CI gates Linux/macOS builds on native tests, the isolated Chromium passkey suite and six-project Chromium/WebKit startup/Return journeys. The complete browser UX matrix is not yet a CI gate.

The `test-tui-smoke` target launches `gi -tui` inside tmux, captures the pane, submits input, verifies blur handling, exercises transcript scrolling keys, resizes the terminal, and writes pane captures plus session artifacts under `test-results/tui-smoke/`. Mouse click focus is covered in unit tests.

## Documentation

See the [documentation index][docs], [feature and parity matrix][parity], and [implementation checklist][checklist]. The matrix covers browser behaviour, compact terminal adaptations and planned integrations separately.

## License

TBD

[parity]: docs/feature-parity.md
[audit]: docs/internal/ux-test-audit-2026-09-24.md
[ux]: tests/ux/README.md
[docs]: docs/README.md
[checklist]: docs/checklists/implementation.md
