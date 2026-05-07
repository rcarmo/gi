# Managed VFS

Managed VFS is the SQLite-backed asset store behind `vfs://` paths.

## Scope
- Store arbitrary binary/text blobs by `(namespace, path)`.
- Offer workspace-like traversal via normalized path resolution.
- Enforce namespace semantics (`reference` read-only by default).
- Share path handling across tools and script bridge.

## URL format

Examples:
- `vfs://skills/example.md`
- `vfs://scripts/lib/helpers.joke`
- `vfs://reference/tools/script.md`

## Resolution semantics
- Tool and bridge paths are resolved through `internal/tools/path_resolver.go`.
- Workspace paths resolve under configured `workspace_root`.
- `vfs://` prefixes are parsed by `store.ParseVFSPath`.
- Relative traversal (`../`) outside namespace is rejected.

## Namespace rules
- Writable namespaces: managed mutable namespaces (e.g. `skills`, `scripts`, `templates`).
- Read-only namespaces:
  - `vfs://reference/...` (embedded/internal reference)
  - `vfs://chat/...` (virtual chat/session projection)
  - `fts://...` (virtual search locators resolved through read path)
- `path escapes workspace` and `invalid vfs path` errors are returned for invalid inputs.

## Runtime behavior
- Reads are returned as logical plaintext bytes where appropriate by APIs.
- Writes and updates persist via `Store.SaveVFSFile` with transparent compression.
- Directory/list operations use `Store.ListVFSChildren`.

## Script/tool integration
- `script` tool path inputs accept virtual paths.
- Bridge file helpers (`readFile`, `writeFile`, `listDir`) share the same resolver.
- `read`, `write`, `shell` tools are exposed via `/api/tools` and executed via `/api/tools/execute` with workspace or virtual paths.

## Extended docs
- URL/resolver details: `docs/internal/vfs/urls.md`
- Chat tree projection: `docs/internal/vfs/chat-projection.md`
- Search locators + namespace hints: `docs/internal/search/fts-namespace.md`
