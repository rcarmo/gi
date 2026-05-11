# Repository structure refactor plan

This note tracks the **interim repository-structure phase** inserted into the core runtime refactor.

The goal is not a cosmetic reorg. The goal is to re-establish clear **functional grouping** after the recent runtime/session/turn/hook work so the next implementation phases land into a cleaner tree instead of compounding historical sprawl.

## Why this phase exists

The runtime refactor has added substantial new logic across:

- `internal/turn`
- `internal/store`
- `internal/session`
- `internal/routing`
- `internal/topics`
- `internal/web`
- `internal/tui`

That work improved correctness, but it also increased the risk that responsibilities drift across packages, helper functions get duplicated, and package/file layout stops matching the conceptual runtime model.

This interim phase pauses feature work long enough to:

- reassess package boundaries
- tighten file grouping inside packages
- make responsibility ownership clearer
- reduce “grab bag” files before more IPC/topic/provider work lands

## Constraints

This phase should be **mechanical and test-backed**, not architecture cosplay.

Keep these constraints:

- preserve external wire contracts unless there is a compelling cleanup payoff and no active consumer breakage
- do **not** add compatibility wrapper packages/files just to preserve the old internal layout; structural slices in this phase are intentional cut-overs
- prefer file regrouping and local helper consolidation before package renames or public import churn
- keep SQLite/WAL and current runtime semantics intact
- avoid broad move-only churn mixed with behavior changes in the same commit
- after each structural slice, re-run focused tests plus broader validation as needed

## Current functional groups

### Runtime core

Owns durable runtime semantics and coordination:

- `internal/turn`
- `internal/store`
- `internal/session`
- `internal/routing`
- `internal/topics`

### User/runtime surfaces

Owns transport/UI entry points and presentation:

- `internal/web`
- `internal/tui`
- `internal/app`
- `cmd/gi`
- `cmd/gi-tui`

### Extension and agent-facing surfaces

Owns model/tool/script/search/package interfaces:

- `internal/tools`
- `internal/skills`
- `internal/scripting`
- `internal/search`

### Infrastructure and integrations

Owns supporting runtime services and external integration glue:

- `internal/config`
- `internal/auth`
- `internal/connectivity`
- `internal/peering`
- `internal/inference`
- `internal/rtk`

## Reassessment questions

Use these questions before moving code:

1. **Does this file/package own runtime state or just consume it?**
2. **Is this helper duplicated because the owning package boundary is unclear?**
3. **Would a file regrouping solve the problem without a package rename?**
4. **Is this API runtime-core, UI-surface, or extension-surface?**
5. **Does the current location make future IPC/topics/provider-hook work harder to place?**

## Expected outputs of this phase

### 1. Package/file inventory

Produce a short inventory of each touched package with:

- what it owns
- what it is allowed to depend on
- what has recently drifted into it opportunistically

### 2. Target grouping rules

Define the intended grouping rules for:

- runtime-core coordination code
- session identity/allocation code
- hook/runtime protocol code
- UI/transport-specific logic
- extension/tool/script/search surfaces

### 3. Low-risk cleanup batches

Land small, mechanical regrouping batches such as:

- moving related files next to each other within a package
- consolidating duplicated identity/session helper logic into store/session ownership points
- splitting overgrown package files into clearer sub-files by responsibility
- moving UI-only glue out of runtime-core packages when ownership is ambiguous

### 4. Follow-up documentation

Update:

- `docs/internal/README.md`
- package-specific internal docs
- any runtime-target notes affected by the regrouping

## Candidate cleanup targets

These are starting hypotheses, not yet committed moves.

### A. Runtime-core ownership tightening

Focus on clearer ownership between:

- `internal/turn`
- `internal/store`
- `internal/session`
- `internal/routing`

Likely tasks:

- remove remaining identity-resolution glue that lives outside store/session ownership
- keep route/session allocation logic flowing through store-backed contracts instead of turn-local fallbacks
- keep topic publication glue clearly separated from core coordination logic where possible

### B. UI/transport separation tightening

Focus on:

- `internal/web`
- `internal/tui`
- `internal/app`

Likely tasks:

- avoid runtime-core helpers leaking into transport code when store-backed identity/runtime APIs should exist instead
- keep session-selection, forking, and display helpers thin wrappers over runtime/store semantics

### C. Extension surface clarity

Focus on:

- `internal/tools`
- `internal/skills`
- `internal/scripting`
- `internal/search`

Likely tasks:

- keep agent-visible extension points documented and grouped consistently
- avoid mixing extension-surface contracts with transport-only or runtime-core-only helpers

## Working approach

1. Inventory current structure and note drift.
2. Pick the smallest useful regrouping slice.
3. Keep the slice mostly mechanical.
4. Test it.
5. Update docs in the same change.
6. Repeat until the tree feels coherent again.

## Landed slices so far

- `internal/store` session identity/allocation code was split into clearer functional files:
  - `session_identity.go`
  - `session_aliases.go`
  - `session_main.go`
  - `session_resolution.go`
  - `session_channel_bindings.go`
- `internal/turn` helper drift was reduced by moving session-identity and routing-helper logic out of `engine.go` into dedicated files:
  - `session_identity_helpers.go`
  - `routing_helpers.go`
- `internal/turn` routing-facing engine methods were split out of `engine.go` into:
  - `routing_runtime.go`
- `internal/turn` generic coercion/sorting helpers and bootstrap shell runtime were split out of `engine.go` into:
  - `value_helpers.go`
  - `shell_runtime.go`
- `internal/tui` session-reference, agent-tree, and fork-target helpers were split out of `chat.go` into:
  - `session_refs.go`
- follow-up audit fixes after the regrouping tightened a few concrete issues without reintroducing grab-bag files:
  - `GetSessionIdentity(...)` now uses targeted per-session detail hydration instead of list-style full-table detail scans
  - recent route/allocation helper paths now reuse caller contexts instead of detached background lookups where that was unintentional
  - store-backed allocation now owns `identity_links` collapse, main-session promotion, and channel-binding attachment/reuse instead of scattering those semantics back into turn/web/TUI callers
  - TUI transcript rendering now uses maintained in-memory transcript state instead of reloading SQLite on each render path

## Done criteria for the interim phase

This phase is “done enough” when:

- the main runtime packages have a clearer ownership map
- recent runtime refactor files are grouped by responsibility rather than chronology
- the next priority phases have obvious landing zones in the tree
- docs reflect the regrouped structure
- no broad import churn or wire-contract regressions were introduced just for tidiness
