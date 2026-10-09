# gi

<img src="docs/icon-256.png" width="128" alt="gi">

A coding agent built on `go-ai`, informed by lessons learned from Pi, Piclaw, and Vibes.

## Why

In short, I needed to run `pi` and `piclaw` on RISC-V, and `bun` didn't run there. Then I realized that having a single binary with _everything_ I needed was a much nicer way to deploy an agent overall, and that Go and `go-joker` were a much nicer (and less resource-intensive) thing to use, especially considering that the Rust compiler is slow as molasses and kept filling the disk on all my sandboxes.

So this was the result. If you like `pi`, you should feel very much at home. If you use `piclaw` or love the Vibes web UI (like I do), the same should be true--even if in either case there will be differences. The goal was not to achieve full parity, but to be a tight, efficient bundle of core features that I can rely upon.

## Status

Gi runs a terminal UI (the default) and a web UI (`gi -web`) from one pure-Go binary, with embedded web assets and SQLite-backed sessions, messages and turns. Bun is needed to build the browser assets; there is no Node or Bun runtime dependency.

The terminal follows Pi 1.0.1 and the browser follows Piclaw 3.2.5. Gi reuses pinned Piclaw web components and implements its own backend, adapters and host UI, so the applications are not interchangeable. Browser behaviour is measured against the shared [fixtures-vibes](https://github.com/rcarmo/fixtures-vibes) suite; the [feature and parity matrix][parity] has the latest results, the known differences and the open issues.

## Features

* Streaming chat, session-local models and thinking levels, durable follow-up queues and run-bound steering share one Go turn engine. Reconnect refreshes state from the server. Automatic and manual compaction write model-generated summaries, as Pi does, and keep the visible conversation.
* The terminal has Pi's selectors and commands: `/model`, `/scoped-models`, `/settings`, `/tree`, `/fork`, `/clone`, `/resume`, `/compact`, `/login`, `/logout`, `/export`, `/import`, `/share`, custom themes and Pi's default keybindings. It offers a fullscreen transcript or native scrollback (`-tui-mode regular`).
* The browser has durable drafts and attachments, session selection and management, scoped search, Markdown/code rendering, image lightboxes, read-only workspace tabs and numeric message references (`msg:42`). Settings has General, Models, Appearance, Keyboard, Compaction, Providers, Keychain, Environment and Authentication sections.
* MCP servers (stdio and Streamable HTTP, with OAuth) use Pi's `mcp.json`, direct and deferred tools, `tool_search` and a QuickJS codemode tool; `gi mcp` and `/mcp` manage them. See the [MCP reference](docs/internal/mcp.md).
* An encrypted keychain in Piclaw's format supplies secrets to shell commands by name; Settings also stores shell environment overrides. See the [keychain](docs/internal/keychain.md) and [shell environment](docs/internal/shell-environment.md) contracts.
* Single-user TOTP sign-in uses an HttpOnly browser cookie and transactional auth storage. Tools and extensions run through Go, embedded Joker/JavaScript or explicitly configured subprocesses, with workspace files and managed `vfs://` references.

**Passkeys are opt-in, with login and Settings enrolment controls.** Settings can add, list, rename and remove credentials after recent authentication. Browser tests verify two independent keys after restart, further enrolment without TOTP, cancellation and lockout-safe removal. Physical-device and synced-key checks are still outstanding. See the [backend contract](docs/internal/passkeys.md) for configuration and limits.

tsnet web access and Iroh inter-instance chat are planned. The tsnet manager scaffold has no HTTP listener wiring, and Gi has no Iroh transport. Both must keep Gi's runtime pure Go and add no idle terminal rows.

## Goals

- unified **web / TUI** experience from one binary, with CLI startup/configuration flags
- boringly reliable **turn handling** via append-only event log
- workspace-centric operation with **SQLite-backed state**
- **Piclaw-oriented compatibility** for settings, message models, injected keychain references and UX conventions, with unsupported APIs and workflow gaps documented explicitly
- **Clojure-first** scripting and skills via Joker
- **go-ai** as the model/provider layer

## Architecture

- `cmd/gi/` — main binary: terminal UI by default, web server with `-web`, `gi mcp` subcommands
- `internal/config/` — Pi/Piclaw configuration loader (settings, auth, context files)
- `internal/store/` — SQLite state store (sessions, messages, turns, events)
- `internal/turn/` — append-only turn engine with queue, cancel and streaming
- `internal/inference/` — go-ai inference, provider auth and SSE broadcasting
- `internal/tui/` — terminal UI on `go-tui`
- `internal/web/` — HTTP server, REST API, SSE streaming, Settings and workspace file APIs
- `internal/mcp/`, `internal/codemode/` — MCP client and the QuickJS-on-wazero codemode engine
- `internal/keychain/`, `internal/environment/`, `internal/shellenv/` — encrypted secrets, environment overrides and shell resolution
- `references/fixtures-vibes/ui/classic` — the web front-end (pinned Piclaw components plus Gi API, host, auth and Settings adapters), maintained in rcarmo/fixtures-vibes
- `docs/` — ADRs, the shipped internal reference, the implementation checklist and transcripts
- `scripts/` — build and check scripts, including the build-time patches applied to Piclaw components
- `tests/` — functional, browser regression, fixtures-vibes profile and terminal acceptance tests

## Terminal rendering

`gi -tui-mode regular` keeps completed output in native terminal scrollback with a five-row idle editor/footer; the terminal owns wheel, selection and copy. `fullscreen` remains the default and supports in-app paging and tool-output folding. Regular mode prints retained output fully expanded after native completion and uses a temporary three-row preview while active. See [ADR-0031](docs/adr/0031-regular-terminal-scrollback.md) for retention and resize limits.

Fullscreen `Ctrl+Shift+F` searches rendered transcript rows without submitting a prompt. Enter/Shift+Enter move between matches; Escape restores the draft and reading position. `Ctrl+Shift+Up/Down` jump between user prompts. See [ADR-0032](docs/adr/0032-fullscreen-transcript-search.md) for rendered-row and retention limits.

In fullscreen mode, drag to select transcript text and hold at a viewport edge to scroll. Release or Ctrl-C/Ctrl-X copies through `tuiClipboardMode`; Escape clears selection. Clipboard-off is respected. Feedback occupies the existing separator row. See [ADR-0034](docs/adr/0034-fullscreen-transcript-selection.md).

## Internal reference

`docs/internal/` is shipped in the binary as the read-only `vfs://reference/...` tree for the agent (for example `vfs://reference/README.md` and `vfs://reference/tools/read.md`).

A change that adds or materially changes an internal tool, scripting bridge capability, hook, managed `vfs://` behaviour, or skill/package contract updates `docs/internal/` in the same commit.

## Development

### Prerequisites

* Go 1.27.1 or newer, as declared in `go.mod`
* Bun (build-time only, not runtime)
* Playwright with Chromium for functional tests; Chromium and WebKit for the six-project fixtures-vibes matrix
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
| `make check` | Go tests, vet, web build, hook checks, adapter tests and functional browser tests |
| `make test-ux` | Functional browser/API tests against an isolated instance |
| `make test-ux-journey` | Empty-store startup/new-chat/Return and retry journeys; Chromium/WebKit at three sizes; required CI gate |
| `make test-ux-picker-geometry` | Pinned Classic composer/picker bounds, responsive transitions and dismissal; required browser CI step |
| `make test-ux-slash` | Native command catalogue and composer/Quick Actions keyboard ownership; required browser CI step |
| `make test-ux-workspace-tabs` | Read-only preview/conversation transitions, keyboard/touch tabs and lifecycle; required browser CI step |
| `make fixtures-vibes` | Shared fixtures-vibes compliance across six browser/viewport projects, with its report gate |
| `make test-ux-parity` | Alias for `make fixtures-vibes` |
| `make test-web-adapters` | Gi adapter tests and frozen provenance |
| `make test-web-regression` | Gi-specific browser race and recovery probes |
| `make test-ux-auth` | Isolated TOTP/browser-auth regression suite |
| `make test-ux-passkeys` | Real Chromium WebAuthn API and Settings/login journeys at three sizes; virtual authenticators, no physical-device claim |
| `make check-cross-build` | Optional local pure-Go builds for Linux/macOS amd64/arm64 and Windows amd64; release CI also builds Windows amd64/arm64 (built, not tested) |
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
| `-web` | `false` | Run the web UI server instead of the terminal UI (required for the web-only flags above and `-pid-file`) |
| `-tui` | `true` | Terminal UI (the default; accepted for compatibility) |
| `-tui-mode` | `tuiMode` setting, else `fullscreen` | Terminal-owned scrollback (`regular`) or in-app transcript (`fullscreen`) |
| `-db` | `$XDG_STATE_HOME/gi/gi.db` (else `~/.local/state/gi/gi.db`) | SQLite database path |
| `-workspace` | current directory | Workspace root |
| `-log-file` | (none) | Log file path |
| `-pid-file` | (none) | PID file path |

`gi --version` prints the version. `gi mcp add|remove|list|login|logout` manages MCP servers without starting a UI.

### TUI mode

`gi` starts the terminal UI by default; pass `-web` for the web UI server:

```sh
gi -db .gi-run/gi.db -workspace /workspace            # terminal UI
gi -web -bind 0.0.0.0 -port 8090 -workspace /workspace # web UI
```

Web-only flags (`-listen`, `-bind`, `-port`, TLS/ACME, `-pid-file`) without `-web` are rejected rather than silently opening the TUI.

The current TUI uses `go-tui`, supports terminal resize handling through the runtime event loop, and enables mouse clicks so the input can regain focus.

## Web UI

The web UI reuses pinned Piclaw component sources and is maintained in rcarmo/fixtures-vibes (`ui/classic`), which Gi consumes through its `references/fixtures-vibes` submodule. It supplies `web/src/api.ts`, `web/src/app.ts`, auth/Settings modules and CSS overrides. Build-time patches in `references/fixtures-vibes/ui/classic/scripts/patch-*.mjs` change selected bundled behaviour without editing the supplied components; each patch fails the build if its anchor text changes.

Workspace tabs are read-only previews with [retained conversation return and keyboard/touch navigation](docs/implementation/web/workspace-tab-transitions.md). Editable documents, dirty-buffer workflows, popouts and docking are not implemented. Composer padding and picker outer bounds follow a [pinned Classic reference](docs/implementation/web/picker-geometry.md), with a documented narrow-desktop containment correction. Session-strip/catalogue structure and full visual styling still differ from Piclaw. The reproduced startup/new-chat focus and loading-retry failures are fixed and covered by [first-Return journeys](docs/implementation/web/startup-return-journeys.md); broader keyboard and visual parity remains open. The [UX audit][audit] records those gaps and the limits of existing tests.

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

Credentials are read from `auth.json` (see [Configuration files](#configuration-files)). Context files (`AGENTS.md`/`CLAUDE.md` from the agent directory and the workspace's ancestors, as in Pi) become the system prompt's project instructions ([system prompt](docs/internal/system-prompt.md)).

## Configuration files

gi reads Pi's configuration unchanged and lets its own directories override it. For each file the first existing location wins; the two are never merged:

- **User level:** `$GI_CODING_AGENT_DIR` (default `~/.gi/agent`), then `$PI_CODING_AGENT_DIR` (default `~/.pi/agent`, as in Pi): `settings.json`, `auth.json`, `models-store.json`, `mcp.json`.
- **Project level:** `<workspace>/.gi/`, then `<workspace>/.pi/`: `settings.json`, `mcp.json`.
- **Directories** (`skills/`, `tools/`, `extensions/`): `.gi/` is scanned before `.pi/`; the first item of a name wins. Skills are also loaded from the user agent directories, before the project's, as in Pi.

A file is written where it was read (gi and Pi refresh `auth.json` in place, holding Pi's lock). When neither location has it, it is created in Pi's location, so it stays shared with Pi. `.piclaw/config.json` (assistant and user identity) is Piclaw's file and is read only there. [Details](docs/internal/config-files.md).

Model lists: gi lists the models of go-ai's catalogue plus those `models-store.json` adds, and filters Copilot to the account's `availableModelIds`, which a Copilot token refresh updates. To refresh `models-store.json` without Pi, set `"modelCatalogUrl": "https://pi.dev"` in the user `settings.json` (Pi's catalogue service); `PI_OFFLINE=1` disables startup refreshes.

## Testing

```sh
make test       # Go unit tests
make test-ux    # functional browser/API tests against an isolated fresh instance
make vet        # go vet
make bun-checks # hook TDZ checker
make check      # standard verification suite
```

The `test-ux` target creates a fresh database, workspace and configuration for each run. Gi-specific race and recovery probes live in `tests/web-regression/`; see [the browser suite guide][ux].

Browser compliance uses the shared [fixtures-vibes](https://github.com/rcarmo/fixtures-vibes) suite, checked out at `references/fixtures-vibes` and pinned to `6e49ae1` (Piclaw 3.2.5 scenarios). `make fixtures-vibes` runs it on Chromium and WebKit at phone, tablet and desktop sizes, then applies its report gate. `tests/fixtures-vibes/profile.json` declares the capabilities Gi claims; `tests/fixtures-vibes/skips.json` lists each absent capability and each known defect with its issue. Release-tag CI runs this gate separately from the native and Gi-specific browser checks. Gi keeps no copy of the shared Gherkin.

The `test-tui-smoke` target launches `gi` inside tmux, captures the pane, submits input, verifies blur handling, exercises transcript scrolling keys, resizes the terminal, and writes pane captures plus session artifacts under `test-results/tui-smoke/`. Mouse click focus is covered in unit tests.

The smoke workspace is isolated from `SMOKE_LOWER` (default `/workspace`) with a kernel overlay (`sudo -n`), falling back to `fuse-overlayfs`, a `--reflink=auto` copy (`SMOKE_COPY_LOWER=1`), or an empty scratch workspace where overlayfs, root or the lower directory are unavailable. The chosen mode is written to `test-results/tui-smoke/workspace-mode.txt`.

## Documentation

See the [documentation index][docs], [feature and parity matrix][parity], and [implementation checklist][checklist]. The matrix covers browser behaviour, compact terminal adaptations and planned integrations separately.

## License

MIT. See [LICENSE](LICENSE).

[parity]: docs/feature-parity.md
[audit]: docs/implementation/audits/ux-test-audit-2026-09-24.md
[ux]: tests/ux/README.md
[docs]: docs/README.md
[checklist]: docs/checklists/implementation.md
