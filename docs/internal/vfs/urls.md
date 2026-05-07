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
- virtual read-only chat projection is available under `vfs://chat/...` (sessions/messages/turns as markdown+frontmatter)
- read-only search locators are available under `fts://...` and rendered through the same read path contract

## Namespace classes
- mutable managed namespaces such as `vfs://skills/...`, `vfs://scripts/...`, `vfs://templates/...`
- read-only embedded namespaces such as `vfs://reference/...`
- read-only virtual runtime namespaces such as `vfs://chat/...`
- read-only virtual search namespace via `fts://...`

## Resolver contract
The shared resolver distinguishes between:
- workspace paths
- `vfs://` paths
- `fts://` locators (read-only)

and returns the correct backend without requiring every tool to duplicate that logic.

## `vfs://chat` projection paths
- `vfs://chat/README.md`
- `vfs://chat/sessions/index.md`
- `vfs://chat/sessions/<session-id>/session.md`
- `vfs://chat/sessions/<session-id>/messages/index.md`
- `vfs://chat/sessions/<session-id>/messages/<message-id>.md`
- `vfs://chat/sessions/<session-id>/turns/index.md`
- `vfs://chat/sessions/<session-id>/turns/<turn-id>.md`

Each document is markdown with JSON-compatible frontmatter for model-friendly structured access.

## `fts://` target examples
- `fts://messages?q=steering+queue&limit=20`
- `fts://turns?q=compaction&session=session_1`
- `fts://workspace?q=HookResponse&glob=internal/**/*.go&limit=20`
- `fts://all?q=subturn&limit=20`

## `fts://` workspace namespaces and hints
- `gi` — core/runtime code paths and docs
  - `fts://gi?q=steering+queue`
- `go-joker` (alias: `joker`) — Joker runtime/bridge and `.joke` surfaces
  - `fts://go-joker?q=register+event+hook`
- `tooling` — tool resolver/execution, web tool API, operational scripts
  - `fts://tooling?q=ResolveToolPath`

See `fts://help` or `fts://namespaces` for the full namespace/hint index.

## Deep-dive docs
- `docs/internal/vfs/chat-projection.md`
- `docs/internal/search/fts-namespace.md`
