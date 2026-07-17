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
- `go test ./...` — all packages green **except pre-existing environmental failures**
  unrelated to the upgrade (verified identical on the pre-upgrade baseline via
  `git stash`): shell-backed tests in `internal/turn` and `internal/web` fail because
  the container's login shell sources `~/.local/share/swiftly/env.sh`, which uses
  bash-only `[[` under `sh` and pollutes shell-tool stdout
  (`sh: env.sh: [[: not found`). This is a workspace environment issue, not a code or
  dependency regression.

## Follow-ups (optional)
- Adopt `ModelRuntime.Refresh` to back live model listing/refresh in `internal/inference`.
- Use `ProviderErrorBody/Status` to surface structured provider errors in the TUI.
- Fix the container `swiftly/env.sh` (`[[` → `[`) to unblock shell-backed tests locally.
