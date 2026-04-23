# Gi

You are a coding agent working on the Gi project — a Go-based coding agent with a Piclaw-compatible web UI.

## Repository layout

- `cmd/gi/` — web server binary
- `cmd/gi-tui/` — terminal UI binary
- `internal/` — Go packages (config, store, turn, inference, web)
- `web/src/` — Piclaw TypeScript source (verbatim) + Gi-specific `api.ts` and `app.ts`
- `tests/functional/` — Playwright functional test suite
- `scripts/` — build/check scripts
- `docs/` — ADRs, checklists, architecture diagrams

## Critical rules

### Functional test coverage is mandatory

**Every user-visible feature must have corresponding functional tests in `tests/functional/`.**

When implementing a new feature:
1. Add functional tests that verify the feature works end-to-end
2. Tests must run against the isolated test instance (`make test-ux`)
3. Tests must pass before the feature is considered done
4. Never remove existing tests without explicit approval

When modifying an existing feature:
1. Update the corresponding functional tests to match the new behavior
2. Run `make test-ux` and verify all tests pass before committing

### Test organization

Tests live in `tests/functional/` and are numbered by feature area:
- `01-app-shell.spec.ts` — page load, JS errors, CSS, bundles
- `02-config-and-session.spec.ts` — runtime config, session management
- `03-chat-flow.spec.ts` — send/receive messages, content rendering
- `04-sse-and-streaming.spec.ts` — SSE connection, real-time events
- `05-system-meters.spec.ts` — metrics API and HUD
- `06-workspace.spec.ts` — file tree, file read, workspace browser
- `07-compose-interaction.spec.ts` — keyboard behavior, history
- `08-turn-lifecycle.spec.ts` — turn events, checkpoints, metadata
- `09-frontend-logging.spec.ts` — browser-to-backend log bridge

When adding a new feature area, create a new numbered spec file.
When extending an existing area, add tests to the existing file.

Shared test helpers go in `tests/functional/helpers.ts`.

### Test instance isolation

The test instance (`make test-ux`) is completely isolated:
- Fresh database at `.gi-test/gi.db`
- Isolated workspace at `.gi-test/workspace/`
- Seeded Pi/Piclaw config files
- Port 19090 (not the dev port 8090)
- Cleaned up after every run

Tests must not depend on external services, live API credentials, or state from previous runs.

## Build and verify

Before committing any change:
```sh
go test ./...       # Go unit tests
go vet ./...        # Go vet
make build-web      # Bun web bundle
make test-ux        # Playwright functional tests (55+ tests)
make bun-checks     # Hook TDZ checker
```

## Web UI rules

- The web UI uses **Piclaw's TypeScript source verbatim** (199 files in `web/src/`)
- Only **two files** are Gi-specific: `web/src/api.ts` and `web/src/app.ts`
- **Never modify** Piclaw component files — adapt `api.ts` or `app.ts` instead
- Vendor bundles are built at compile time by `build.js`
- All assets are embedded in the Go binary via `embed.FS`

## Inference

- Uses `go-ai` for model/provider abstraction
- Auth loaded from `~/.pi/agent/auth.json`
- GitHub Copilot requires token exchange (refresh → session token + enterprise endpoint)
- System prompt loaded from workspace `AGENTS.md`
- Streaming via `goai.Stream()` with SSE broadcast

## Development

- `make start` — build and start on port 8090
- `make restart` — rebuild and restart
- `make test-ux` — run functional tests against isolated instance
- `make logs` — tail server log
- `make status` — check if running

## Implementation checklist

See `docs/checklists/implementation.md` for the full phased checklist.
Update it as items are completed.
