<!-- RUI-PROFILE-LIFECYCLE-20261005 -->
## Current profiling and cleanup rule — supersedes older text below

Rui's explicit rule: **profile and tune during pre-release tests; remove profiling data immediately after analysis/use.** Ordinary development tests do not require profiling on every run. Targeted diagnostic profiling is optional when useful and follows the same disposal rule.

- During pre-release verification, capture CPU and heap/allocation behaviour, analyse hotspots and tune avoidable allocations/repeated work. Compare equivalent workloads without weakening correctness/security/numerical contracts. For Go, inspect CPU, alloc_space and alloc_objects; coverage alone is not profiling.
- Keep captures and matching artifacts only while the current analysis needs them. Once used, immediately delete raw profiles, traces, matching test binaries, temporary fixtures and disposable logs, including failed/probe artifacts after diagnosis. Retain only concise conclusions and important measurements/limitations. Do not keep indefinite raw archives or copy them into exports/reports/evidence to evade cleanup.
- Remove all completed disposable cache/build/test/run data promptly. Never delete files still in use: finish or safely pause the owning job and clean at a safe boundary. Preserve source, installed toolchains, durable datasets/checkpoints and intentional release assets. Minimise disk usage; no random exports or redundant snapshots.
- This rule overrides **every conflicting older paragraph in this file and linked local guidance**, including “profile every test”, “unprofiled tests prohibited”, “retain all raw evidence”, “never delete profiles” and cleanup exclusions based only on an evidence/profiles directory name. Update helper/CI cleanup behaviour accordingly; do not weaken pre-release analysis.
- Existing no-agent-contact and execution-pause rules remain unchanged. This policy grants no unsolicited coordination or automatic job restart.
<!-- /RUI-PROFILE-LIFECYCLE-20261005 -->



# gi

## Go toolchain and project scratch

Gi uses Go **1.27.1** (`GOTOOLCHAIN=go1.27.1` in the Makefile); CI reads the root `go.mod`. Vendored modules keep their own minimum Go directives.

The canonical scratch project is `gi`. The repository-vendored `scripts/project-tmp.sh` resolves an absolute `PROJECT_TMP_BASE` to `<base>/gi`, or accepts a project-named `PROJECT_TMP_ROOT`; both must agree. Invalid/unusable overrides fail. Otherwise CI chooses `RUNNER_TEMP`, original inherited `TMPDIR`, then platform temp, even if `/workspace/tmp` exists; local use chooses writable `/workspace/tmp`, then platform temp. The resolved root propagates to children before changing their temp variables, preventing nested `/gi` suffixes.

Verification uses `cache/go-build`, `cache/go-mod`, `cache/go-race-1.27.1.ok`, `build/<worktree>`, and isolated `runs/tests/<worktree>/<run-id>/tmp/go`. Bun/npm/XDG and new Playwright downloads use `cache/<tool>`. Existing installed browsers may be selected explicitly with `PLAYWRIGHT_BROWSERS_PATH`; they are not relocated during active jobs. Profiling scratch uses `runs/profiling/<worktree>` and is removed automatically after CPU/alloc_space/alloc_objects analysis. `PROFILING=1` is required for pre-release runs; ordinary tests default to no captures. `PROFILE_KEEP=1` retains raw data only for current manual analysis and requires removal afterwards. Durable `.gi-run` state is not build scratch. Automatic parse-time cache trimming is disabled because other worktrees may have active builds.

Browser/TUI helpers source `scripts/project-test-env.sh`; direct commands and CI use `eval "$(make -s test-env RACE=)"` before installs or subprocesses. Per-run instance, fixture, output and temporary paths live beneath `GI_TEST_RUN_ROOT`. Helpers reject symlink, unowned and outside-run mutation targets. `make test-project-paths` checks override agreement, workspace/generic/CI selection and non-nesting. `playwright.fixtures.config.ts` redirects shared fixture results without editing the pinned suite. Completed successful profiled runs dispose their owned scratch; failed/manual runs are kept only until diagnosis. `clean` removes only the selected owned run after its jobs have finished, never durable state, another project or active/shared caches. This section supersedes older home-cache/race-probe defaults below.

You are a coding agent working on the gi project: a Go coding agent with a Pi-style terminal UI and a Piclaw-compatible web UI.

## Repository layout

```
cmd/gi/              main binary entrypoint (TUI by default, web server via `-web`, `gi mcp`)
cmd/gi-tui/          compatibility wrapper for TUI mode
internal/
  tui/               terminal UI implementation (go-tui)
  config/            Pi/Piclaw config loader (.gi, then .pi)
  store/             SQLite WAL state store
  turn/              append-only turn engine with queue/cancel
  inference/         go-ai streaming inference with auth
  mcp/, codemode/    MCP client and QuickJS-on-wazero codemode
  keychain/, environment/, shellenv/   secrets, env overrides, shell resolution
  web/               HTTP server, REST API, SSE, Settings, metrics, workspace
tests/functional/    Playwright functional test suite
tests/web-regression/  Gi-only browser regressions
tests/fixtures-vibes/  fixtures-vibes profile, skips and seed script
references/fixtures-vibes/  shared browser compliance suite and the web front-end (ui/classic; git submodule)
scripts/             build/check scripts
docs/
  adr/               architecture decision records
  checklists/        phased implementation checklist
  internal/          shipped internal reference source tree for tools/scripting/hooks/VFS
  reference/         spec transcripts
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
- No approximations — if it's in gi it matches Piclaw exactly; if it's not ready it simply isn't in gi yet
- **The web front-end is not edited here.** Its sources, build, embedded assets and front-end unit tests live in rcarmo/fixtures-vibes `ui/classic` (owned by the fixtures-vibes front-end owner) and reach Gi only through `references/fixtures-vibes`; Go embeds it via `github.com/rcarmo/gi/references/fixtures-vibes/ui/classic` (`giui.Static`, `giui.ThemeCatalogue`). Change it upstream, then bump the submodule.
- Future Piclaw updates should drop in with zero diff on gi's side

### Go-native runtime
- Core runtime is **pure Go** — no CGO, no Node/Bun at runtime
- Bun is allowed **only at build time** for web asset bundling
- Web assets are **embedded in the Go binary** via `embed.FS`
- Use `go-ai` for model/provider abstraction, `go-tui` for the terminal UI
- The main `gi` binary starts the TUI by default; `-web` runs the web UI server

### Configuration compatibility
- Read existing Pi/Piclaw files without modification; `.gi/` locations take precedence over `.pi/` ones (see `docs/internal/config-files.md`):
  - `.piclaw/config.json` — assistant/user identity and avatar
  - `.pi/settings.json` — provider, model, thinking level
  - `~/.pi/agent/auth.json` — provider auth tokens
  - `~/.pi/agent/mcp.json`, `.pi/mcp.json` — MCP servers
  - `AGENTS.md` / `CLAUDE.md` — context files for the system prompt
- Preserve Pi model/provider naming semantics exactly

## Workflow: spec → code → test → ship

### 1. Spec / plan

- Features start in `docs/checklists/implementation.md` — the phased implementation checklist organized by subsystem
- Architecture decisions are recorded in `docs/adr/` — create a new ADR for significant design choices
- The original spec conversation is preserved verbatim in `docs/reference/`
- Agent-facing runtime/tooling/scripting contracts live in `docs/internal/` and ship as the read-only `vfs://reference/...` tree. Feature implementation notes, plans, audits and verification findings live in `docs/implementation/`, grouped by topic; keep them out of the embedded reference.
- For new feature areas, add checklist items first, then implement

### 2. Implement

- Read relevant files before editing — never edit blind
- Use the Makefile for all operations — not raw `go build` or `bun run`
- Push as fixes land — do not batch unrelated changes into large commits
- Commit messages should explain what changed and why
- If a change adds or materially changes an internal tool, scripting bridge capability, hook, managed `vfs://` behavior, or skill/package contract, update `docs/internal/` in the same change

### 3. Test

Run tests **one at a time, through the Makefile only** (see *CPU throttling* below). Never launch several test suites in one command, and never run `go test`, `bun`, or test scripts directly — that bypasses the throttle.

**Every user-visible feature must have corresponding functional tests.**

Pre-release test runs must be profiled, analysed and tuned. Use `scripts/testprofile` and per-package Go CPU/allocation capture for those runs; ordinary focused development tests may run without profiling. Script wall/CPU/RSS totals do not measure allocations. Inspect application hotspots separately from harness overhead and compare equivalent workloads. Keep concise conclusions in `docs/internal/profiling.md` or the normal report; delete raw captures, test binaries, disposable logs and temporary history immediately after use. Do not disable filesystem isolation or correctness checks.

**For now, web (Playwright) tests are not run on the ChromeOS (Crostini) laptop** used for development: `make test-ux` and the other browser targets are skipped there. Verify web-facing changes on that laptop with Go tests (for example `internal/web` handler tests), and run browser acceptance on another host. This is temporary and specific to that laptop; other machines run the web suites as usual.

Before committing, one at a time:
```sh
make test           # Go unit tests (TEST_RUN=... for a focused run)
make vet            # Go vet
make build-web      # Bun web bundle
make test-ux        # Playwright functional tests (21 files, about 130 tests)
make bun-checks     # Hook TDZ checker
```

Run browser targets with `PI_CODING_AGENT_DIR` and `GI_CODING_AGENT_DIR` unset (`env -u PI_CODING_AGENT_DIR -u GI_CODING_AGENT_DIR make ...`), so the host agent's settings do not leak into test instances.

#### Shared browser compliance (fixtures-vibes)

`make fixtures-vibes` runs the shared suite pinned at `references/fixtures-vibes` across six Chromium/WebKit projects and applies its report gate (about 75 minutes). Run focused scenarios first; run the full suite at most once every four hours. Do not rebuild bundles or `bin/gi-fixtures-vibes` while it runs.

- Make Gi match Piclaw rather than loosening a scenario. Report suspected suite bugs to the fixtures-vibes coordinator with evidence.
- Every scenario Gi does not pass is listed in `tests/fixtures-vibes/skips.json` with a reason: `capability-absent` (with the capability), `known-defect` or `not-implemented` (with a Gi issue). `intentional-divergence` needs the owner's approval.
- Remove a skip in the same change that makes its scenario pass; the gate fails on listed scenarios that pass.
- Claim a capability in `tests/fixtures-vibes/profile.json` only when Gi implements it.

### 4. Ship

- Verify `make test-ux` passes on a **fresh isolated instance**
- Update `docs/checklists/implementation.md` to mark completed items
- Push to `main` on `github.com/rcarmo/gi`
- Restart the dev instance: `make restart BIND=0.0.0.0 PORT=8090`

## Makefile reference

### Bootstrap
| Target | Description |
|---|---|
| `make help` | Show the grouped target list |
| `make deps` | Download Go modules and install Bun packages |
| `make bootstrap` | Install repo dependencies, install Playwright Chromium, and build `gi` |

### Dev instance lifecycle
| Target | Description |
|---|---|
| `make start` | Build and start gi detached (default `0.0.0.0:8090`) |
| `make stop` | Stop the detached process |
| `make restart` | Rebuild and restart |
| `make status` | Show PID and listener |
| `make logs` | Tail the log file |
| `make run` | Foreground run (no detach) |

### Build
| Target | Description |
|---|---|
| `make build` | Build web assets + the main `gi` binary |
| `make build-web` | Bun vendor + app bundle only |

### Test
| Target | Description |
|---|---|
| `make test` | Go unit tests (`go test ./...`) |
| `make vet` | Go vet |
| `make bun-checks` | Hook TDZ checker |
| `make check` | Standard verification suite (`test`, `vet`, `build-web`, `bun-checks`, `test-ux`) |
| `make test-ux` | Start isolated instance → run Playwright → stop and clean up (artifacts under `test-results/`) |
| `make test-tui-smoke` | Run the tmux-based TUI smoke harness (startup/resize artifacts under `test-results/tui-smoke/`) |
| `make test-tui-gherkin` | Run the TUI gherkin harness |
| `make fixtures-vibes` | Shared browser compliance (six projects, report gate) |
| `make test-web-regression` | Gi-only browser race/recovery regressions |
| `make test-web-adapters` | Bun adapter tests under `tests/ux/support/` |

### Isolated test instance
| Target | Description |
|---|---|
| `make test-instance-start` | Build binary, create fresh DB/workspace, start on port 19090 |
| `make test-instance-stop` | Stop and remove `.gi-test/` directory |

### Cleanup
| Target | Description |
|---|---|
| `make clean` | Remove `.gi-run/`, `bin/`, `.gi-test/`, `.gi-tui-test/`, `test-results/` |

### CPU throttling

All CPU limiting lives in the Makefile, so every build, test and dev-server run is throttled the same way and throttling is reproducible:

- every recipe runs under `nice -n $(CPU_NICE)` and `taskset -c $(CPU_SET)` (defaults: `10`, `0-1`)
- `GOMAXPROCS` and Go build parallelism (`-p`) follow `CPU_PROCS` (default `2`)
- make is `.NOTPARALLEL`; `test*` targets run one Go package at a time (`-p=1`)
- Go cache/compiler scratch use the resolved project hierarchy; no parse-time cache trimming.
- `-race` uses the project-owned cached ThreadSanitizer probe.

Tune per invocation instead of bypassing make, e.g. `make test CPU_SET=0-3 CPU_PROCS=4` or `make build CPU_NICE=0`.

### Overrides
```sh
make start PORT=3000 BIND=127.0.0.1 MODEL=github-copilot/gpt-5-mini WORKSPACE=/workspace
```

## Functional test suite

### Organization

Tests live in `tests/functional/` and are numbered by feature area:

| File | Area |
|---|---|
| `01-app-shell.spec.ts` | Page load, JS errors, CSS, bundles, appearance, favicon, cache busters |
| `02-config-and-session.spec.ts` | Runtime config, Pi settings, session auto-create |
| `03-chat-flow.spec.ts` | Send/receive, content rendering, persistence, avatars |
| `04-sse-and-streaming.spec.ts` | SSE connection, event persistence, turn completion |
| `05-system-meters.spec.ts` | Metrics API, CPU/RAM/swap, poll interval, HUD |
| `06-workspace.spec.ts` | Tree API, file read, path traversal, workspace toggle |
| `07-compose-interaction.spec.ts` | Keyboard behaviour, focus, sequential messages |
| `08-turn-lifecycle.spec.ts` | Turn events, checkpoints, metadata, prompt match |
| `09-frontend-logging.spec.ts` | Log endpoint, error handler |
| `10-scripting.spec.ts` – `21-conversation-projection.spec.ts` | Scripting, media, remote links, outcomes, recovery controls and placeholders, card rejection, auth gate, HTTP send and delivery, provider retry, conversation projection |

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
- IDs: `prefix_<unix_nano>` strings; `message_rows` assigns each message a numeric row ID, which message references (`msg:42`) and the messages tool use
- Core tables: `sessions`, `messages`, `turns`, `turn_events`; also `kv_store` (settings), `keychain_entries`, `media`, `vfs_files` and the workspace index tables

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
- Auth from `auth.json` in `~/.gi/agent`, else `~/.pi/agent`
- GitHub Copilot: token exchange (refresh → session token + enterprise endpoint detection)
- Pi's structured system prompt, with context files (`AGENTS.md`/`CLAUDE.md`) as project instructions
- Token/cost tracked per turn in event payloads
- Conversation history built from session messages

### Error handling
- Frontend errors captured by `window.onerror` → POST to `/api/frontend/log`
- Inference errors surface as system messages, not silent failures
- Turn failures preserve partial output, set status to `failed`
- Cache busters prevent stale bundles from causing ghost errors
