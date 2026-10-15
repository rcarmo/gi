# Shipped internal reference system

## Goal
The agent running inside gi must be able to discover how gi extends itself.

The internal reference system is the documentation substrate that makes that possible.

## Source of truth
Agent-facing reference contracts are authored under `docs/internal/`. Feature implementation details, plans, audits and verification results belong in `docs/implementation/`; they are not embedded.

This subtree is the canonical source for:
- tool contracts
- scripting runtimes and bridge APIs
- hook surfaces
- `vfs://` semantics
- skill/package structure
- other agent-visible extension points

## Shipped form
`docs/embed.go` embeds the reference Markdown. The store exposes it through the read-only `vfs://reference/...` namespace and the shared VFS resolver. Normal agent write/edit operations cannot change it.

## Read-only contract
The reference namespace is for inspection only.

Normal tool behavior should therefore treat `vfs://reference/...` as:
- readable
- listable
- globbable / walkable
- not writable
- not renamable
- not deletable

## Resolution expectations
Internal tools and runtimes resolve through one shared resolver layer and now support:
- workspace filesystem paths
- mutable `vfs://...` managed assets
- read-only `vfs://reference/...`

through one shared resolver layer.

## Minimum contents to ship
At minimum, the shipped reference should include pages for:
- each built-in tool
- each scripting runtime
- bridge API semantics
- hook lifecycle semantics
- managed VFS rules
- skill/package structure

## Maintenance rule
Changes to agent-visible extension surfaces must update `docs/internal/` in the same change.

## Future enforcement
Current checks:
- ensure every registered tool has a matching `docs/internal/tools/<name>.md`
- ensure scripting runtimes and hooks have matching docs pages
- embed docs tree during build and run a smoke test over `vfs://reference/...`
