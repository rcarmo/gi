# Internal reference

This subtree is the **canonical internal documentation surface** for gi runtime features that the agent can use or extend.

It is written for:
- the agent running inside gi
- humans extending gi
- future packaging into a read-only embedded reference tree such as `vfs://reference/...`

## Contract

If a change adds or materially changes any of the following, it must update this subtree in the same change:
- built-in tools
- scripting runtimes or bridge APIs
- hook surfaces
- managed `vfs://` behavior
- skill/package structure
- other agent-visible extension points

## What belongs here

- **tool contracts**: purpose, parameters, outputs, side effects, examples
- **scripting docs**: engine behavior, bridge globals, file/state access, examples
- **hook docs**: lifecycle timing, guarantees, mutation rules, examples
- **VFS docs**: `vfs://` URL semantics, path resolution, read-only namespaces, export/sync model
- **skill/package docs**: structure, conventions, discovery, packaging

## Stability policy

Paths under `docs/internal/` should remain stable once referenced by prompts, tools, or future `vfs://reference/...` URLs.

## Current structure

- `tools/` — built-in tool contracts
- `scripting/` — scripting runtimes and bridge docs
- `hooks/` — hook and lifecycle docs
- `vfs/` — managed VFS and `vfs://` URL docs (including `vfs://chat` projection)
- `skills/` — skill/package structure docs
- `search/` — hybrid workspace search and indexing design (including `fts://` namespace docs)
- `routing.md` — routing policy, route resolution, and model-routing observability
- `runtime-target-state.md` — database-backed target state for the core runtime refactor (sessions, turns, steering, hooks, events, IPC, multi-channel bindings)

The `search/` subtree is the canonical reference for the current hybrid workspace search direction: SQLite metadata + FTS5 + vec + local embeddings, plus the runtime-facing read-only `fts://` locator contract.

`routing.md` is the canonical in-repo reference for routing decisions and route-event persistence.

`runtime-target-state.md` is the canonical schema/state target for the runtime refactor.

## Current status

This subtree is being bootstrapped. When internal surfaces are implemented before full docs exist, add at least a minimal placeholder page and update it as the implementation stabilizes.

Current search status:
- architecture chosen: **vec + FTS**
- ADR written
- internal design written
- Go package scaffold added under `internal/search/`
- runtime read-only `fts://` namespace is implemented for model/tool retrieval workflows
- full `internal/search` backend wiring and advanced index pipeline work remains pending
