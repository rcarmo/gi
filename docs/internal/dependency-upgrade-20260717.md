# Dependency upgrade impact assessment — 2026-07-17

## Upgrades applied

| Module | From | To | Kind |
|---|---|---|---|
| github.com/rcarmo/go-ai | v0.79.3 | v0.80.10 | minor |
| github.com/grindlemire/go-tui | v0.17.0 | v0.18.2 | minor |
| github.com/dop251/goja | 20260607… | 20260701… | pseudo |
| github.com/yuin/goldmark | v1.8.2 | v1.8.4 | patch |
| golang.org/x/crypto | v0.53.0 | v0.54.0 | minor |
| golang.org/x/net | v0.56.0 | v0.57.0 | minor |
| modernc.org/sqlite | v1.52.0 | v1.54.0 | minor |
| (transitive) klauspost/compress, x/sync, x/sys, x/term, x/text, modernc.org/libc | — | — | bumped |

## go-ai 0.79.3 → 0.80.10 — runtime impact

**Verdict: no breakage. Additive with two removals that we do not use.**

### Removed (potential breaking) — none affect us
- `EstimateTokens(ctx *Context) int` — removed upstream. We do **not** call it; our
  `internal/compaction.EstimateTokens(text string)` is a separate local function.
- `ResolveCloudflareBaseURL(model *Model)` — removed upstream. Not referenced anywhere
  in `internal/` or `cmd/`.

### Added (available for future adoption, not yet used — YAGNI)
- `ModelRuntime.Refresh(ctx, allowNetwork) ModelRuntimeRefreshResult`, plus
  `GetModel` / `GetModels` — a first-class model-runtime refresh surface. Could later
  back our `/model` selector and provider/model listing in `internal/inference`.
- `StaticModelProvider` (implements `ID` / `StaticModels` / `RefreshModels`) and
  `DeferredToolPlan.HasDeferred()`.
- `ProviderEnv map[string]string`, `ModelCostTier` (cost metadata).
- Richer provider error surface: `providerErr.ProviderErrorBody()` /
  `ProviderErrorStatus()` and `OAuthResponseError.Error()`. Could improve how we
  surface upstream provider/oauth failures in the TUI/web error paths later.

### Packages we import (all still compile clean)
- root `github.com/rcarmo/go-ai`, `oauth`, and providers
  `anthropic` / `openai` / `openaicodex` / `openairesponses`.

## go-tui 0.17.0 → 0.18.2 — TUI impact
- Builds clean; **all `internal/tui` tests pass**. No API changes required in our
  chat TUI (`gotui.OnStop`, `KeyMap`, refs, layout primitives unchanged for our use).

## Validation
- `go build ./...` ✅
- `go vet ./...` ✅
- `go test ./...` ✅

## Required runtime fix following validation

Validation exposed that Gi executed tool commands with `sh -lc`. The `-l` made every
shell tool source the user's login profile, which is both non-deterministic and allowed
profile diagnostics/errors to contaminate tool output. In this workspace it sourced
Swiftly's bash-specific `env.sh` under POSIX `sh` and emitted `[[: not found`.

All Gi shell execution paths now use deterministic non-login POSIX shells (`sh -c`):
`ExecuteShell`, `ExecuteRTK`, the shell turn runtime, local TUI shell execution, and the
web tool executor. Runtime event command metadata was updated to match. The regression
test `TestExecuteShellDoesNotSourceLoginProfile` proves `.profile` output is not mixed
into tool output.

## Shell output drain correction — 2026-09-21

The web parity run reproduced empty output from the shell responder. `cmd.Wait()`
closed its stdout/stderr pipes before the readers drained buffered bytes. The runner
now waits for both readers before reaping the child; context cancellation still kills
the process group to unblock reads. A slow-reader regression test failed with 128 of
26,013 bytes before the fix and preserves the entire response afterward.

## Pooled SQLite connection correction — 2026-09-21

The multi-session browser matrix reproduced `SQLITE_BUSY` while setting up a turn
and recording its terminal failure, leaving accepted rows marked `running/setup`.
The startup `PRAGMA` calls configured only one physical connection. A regression
holding four `database/sql` connections found the second lacked the required settings.

`store.Open` now supplies connection-local busy timeout, foreign-key, synchronous
and temp-store settings through modernc's `_pragma` DSN options. `_txlock=immediate`
reserves the writer before read-then-write coordination transactions, avoiding WAL
snapshot-upgrade failures. WAL and the existing append-only event contract are unchanged.
The full Go suite and 24-case browser parity matrix passed after this correction.

## Follow-ups (optional, YAGNI)
- Adopt `ModelRuntime.Refresh` to back live model listing/refresh in `internal/inference`.
- Use `ProviderErrorBody/Status` to surface structured provider errors in the TUI.
