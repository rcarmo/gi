# Gi

You are a coding agent working on the Gi project — a Go-based coding agent with a Piclaw-compatible web UI.

## Repository layout

```
cmd/gi/              web server binary
cmd/gi-tui/          terminal UI binary (go-tui)
internal/
  config/            Pi/Piclaw config loader
  store/             SQLite WAL state store
  turn/              append-only turn engine with queue/cancel
  inference/         go-ai streaming inference with auth
  web/               HTTP server, REST API, SSE, metrics, workspace
    static/          embedded assets (CSS, JS, fonts, icons)
web/src/             Piclaw TypeScript source (verbatim) + Gi adapters
  api.ts             Gi API adapter (same signatures as Piclaw)
  app.ts             Gi entry point (wires Piclaw components to Gi backend)
  components/        Piclaw components (DO NOT MODIFY)
  ui/                Piclaw UI utilities (DO NOT MODIFY)
  panes/             Piclaw pane system (DO NOT MODIFY)
  vendor/            vendor entry files for build
  styles/            Piclaw CSS source
tests/functional/    Playwright functional test suite
scripts/             build/check scripts (hook TDZ checker)
docs/
  adr/               architecture decision records
  checklists/        phased implementation checklist
  reference/         spec transcripts
build.js             Bun web asset build script
Makefile             canonical build/test/run interface
```

## Design principles

### Reliability above all
- Turn execution must be **boringly reliable**
- Auto-recover from provider failures, tool failures, context overflow, and session issues
- **Always hand control back cleanly after every turn** — UI responsive, no hidden tasks, partial output preserved, conversation state consistent
- A turn is complete whenever control returns with consistent state — success, partial success, or surfaced failure

### Simplicity
- Prefer the **simplest possible implementation** that works correctly
- Use an **append-only event log** with minimal turn state — not a complex state machine
- Avoid abstractions until they're needed
- One workspace root, one SQLite database, one binary

### Piclaw UX parity
- Parts that are ported must be **100% identical** to Piclaw — same DOM, same classes, same behavior, same visual output
- No approximations — if it's in Gi it matches Piclaw exactly; if it's not ready it simply isn't in Gi yet
- **Never modify Piclaw component files** — adapt only `api.ts` or `app.ts`
- Future Piclaw updates should drop in with zero diff on Gi's side

### Go-native runtime
- Core runtime is **pure Go** — no CGO, no Node/Bun at runtime
- Bun is allowed **only at build time** for web asset bundling
- Web assets are **embedded in the Go binary** via `embed.FS`
- Use `go-ai` for model/provider abstraction, `go-tui` for the terminal UI

### Configuration compatibility
- Read existing Pi/Piclaw files without modification:
  - `.piclaw/config.json` — assistant/user identity and avatar
  - `.pi/settings.json` — provider, model, thinking level
  - `~/.pi/agent/auth.json` — provider auth tokens
  - `AGENTS.md` — system prompt
- Preserve Pi model/provider naming semantics exactly

## Workflow: spec → code → test → ship

### 1. Spec / plan

- Features start in `docs/checklists/implementation.md` — the phased implementation checklist organized by subsystem
- Architecture decisions are recorded in `docs/adr/` — create a new ADR for significant design choices
- The original spec conversation is preserved verbatim in `docs/reference/`
- For new feature areas, add checklist items first, then implement

### 2. Implement

- Read relevant files before editing — never edit blind
- Use the Makefile for all operations — not raw `go build` or `bun run`
- Push as fixes land — do not batch unrelated changes into large commits
- Commit messages should explain what changed and why

### 3. Test

**Every user-visible feature must have corresponding functional tests.**

Before committing:
```sh
go test ./...       # Go unit tests
go vet ./...        # Go vet
make build-web      # Bun web bundle
make test-ux        # Playwright functional tests (55+ tests)
make bun-checks     # Hook TDZ checker
```

### 4. Ship

- Verify `make test-ux` passes on a **fresh isolated instance**
- Update `docs/checklists/implementation.md` to mark completed items
- Push to `main` on `github.com/rcarmo/gi`
- Restart the dev instance: `make restart BIND=0.0.0.0 PORT=8090`

## Makefile reference

### Dev instance lifecycle
| Target | Description |
|---|---|
| `make start` | Build and start Gi detached (default `0.0.0.0:8090`) |
| `make stop` | Stop the detached process |
| `make restart` | Rebuild and restart |
| `make status` | Show PID and listener |
| `make logs` | Tail the log file |
| `make run` | Foreground run (no detach) |

### Build
| Target | Description |
|---|---|
| `make build` | Build web assets + both binaries (`gi` and `gi-tui`) |
| `make build-web` | Bun vendor + app bundle only |

### Test
| Target | Description |
|---|---|
| `make test` | Go unit tests (`go test ./...`) |
| `make vet` | Go vet |
| `make test-ux` | Start isolated instance → run Playwright → stop and clean up |
| `make bun-checks` | Hook TDZ checker |

### Isolated test instance
| Target | Description |
|---|---|
| `make test-instance-start` | Build binary, create fresh DB/workspace, start on port 19090 |
| `make test-instance-stop` | Stop and remove `.gi-test/` directory |

### Cleanup
| Target | Description |
|---|---|
| `make clean` | Remove `.gi-run/`, `bin/`, `.gi-test/`, `test-results/` |

### Overrides
```sh
make start PORT=3000 BIND=127.0.0.1 MODEL=github-copilot/gpt-5-mini WORKSPACE=/workspace
```

## Functional test suite

### Organization

Tests live in `tests/functional/` and are numbered by feature area:

| File | Area | Tests |
|---|---|---|
| `01-app-shell.spec.ts` | Page load, JS errors, CSS, bundles, theme, favicon, cache busters | 10 |
| `02-config-and-session.spec.ts` | Runtime config, Pi settings, session auto-create | 6 |
| `03-chat-flow.spec.ts` | Send/receive, content rendering, persistence, avatars | 9 |
| `04-sse-and-streaming.spec.ts` | SSE connection, event persistence, turn completion | 4 |
| `05-system-meters.spec.ts` | Metrics API, CPU/RAM/swap, poll interval, HUD | 7 |
| `06-workspace.spec.ts` | Tree API, file read, path traversal, workspace toggle | 6 |
| `07-compose-interaction.spec.ts` | Keyboard behavior, focus, sequential messages | 4 |
| `08-turn-lifecycle.spec.ts` | Turn events, checkpoints, metadata, prompt match | 5 |
| `09-frontend-logging.spec.ts` | Log endpoint, error handler | 4 |

### Rules
- **Add tests when adding features** — no feature ships without functional test coverage
- When adding a new feature area, create a new numbered spec file
- When extending an existing area, add tests to the existing file
- Shared helpers go in `tests/functional/helpers.ts`
- **Never remove existing tests** without explicit approval
- Tests must not depend on external services, live credentials, or prior state

### Test instance isolation
- Fresh database, workspace, and config for every run
- Port 19090 — does not conflict with the dev instance on 8090
- Seeded with minimal `.piclaw/config.json` and `.pi/settings.json`
- Cleaned up after every run

## Technical reference

### Database schema
- SQLite WAL mode with JSON fields and expression indexes
- `_json` suffix columns for JSON payloads
- `json_extract()` expression indexes for queried paths
- ISO 8601 text timestamps
- Foreign keys with cascade delete
- IDs: `prefix_<unix_nano>` strings
- Tables: `sessions`, `messages`, `turns`, `turn_events`

### SSE event model (`/sse/stream`)
- `connected`, `heartbeat` — connection lifecycle
- `agent_status` — turn status updates
- `agent_draft_delta` — streaming text tokens
- `agent_thought_delta` — streaming thinking tokens
- `new_post` — completed message
- `agent_response` — turn completion

### Build pipeline
- Vendor bundles built first (preact, marked, katex, mermaid, codemirror)
- App bundle built as ESM, IIFE-wrapped to prevent global var shadowing
- Vendor scripts loaded as `type="module"` to scope declarations
- Cache busters on all URLs in `index.html` generated per server restart
- CSS served from `/css/styles.css` with `@import` partials

### Inference
- `go-ai` with streaming via `goai.Stream()`
- Auth from `~/.pi/agent/auth.json`
- GitHub Copilot: token exchange (refresh → session token + enterprise endpoint detection)
- System prompt from `AGENTS.md`
- Token/cost tracked per turn in event payloads
- Conversation history built from session messages

### Error handling
- Frontend errors captured by `window.onerror` → POST to `/api/frontend/log`
- Inference errors surface as system messages, not silent failures
- Turn failures preserve partial output, set status to `failed`
- Cache busters prevent stale bundles from causing ghost errors
