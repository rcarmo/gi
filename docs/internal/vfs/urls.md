# `vfs://` URL semantics

## Status
Implemented for shared resolver-backed access in tools and scripts.

## Goal
`vfs://...` should be a first-class internal path form for agent-managed assets stored in SQLite.

Examples:
- `vfs://skills/example/SKILL.md`
- `vfs://scripts/lib/helpers.joke`
- `vfs://reference/tools/script.md`

## Implemented rules
- internal Go tools and scripts accept `vfs://` via the shared resolver
- script bridge file helpers (`readFile`, `writeFile`, `listDir`) resolve through the same layer
- VFS content is available through DB-backed media/VFS APIs
- VFS writes/overwrites are persisted through native DB ops with compression
- shell tool execution currently runs a host command (`sh -lc`); no direct VFS mount is required

## Namespace classes
- mutable managed namespaces such as `vfs://skills/...`, `vfs://scripts/...`, `vfs://templates/...`
- read-only embedded namespaces such as `vfs://reference/...`

## Resolver contract
The shared resolver distinguishes between:
- workspace paths
- `vfs://` paths

and return the correct backend without requiring every tool to duplicate that logic.
