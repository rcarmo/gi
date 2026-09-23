# Tool: `write`

## Status
Implemented in the native engine, `/api/tools/execute` and script file-write bridge.

## Purpose
Write text content to a workspace file or managed VFS asset.

## Input
```json
{
  "path": "relative/path.txt",
  "content": "file contents"
}
```

### Fields
- `path` — workspace-relative destination path or `vfs://namespace/path`
- `content` — full text content to write

## Behavior
- create parent directories when writing workspace files (if missing)
- overwrite destination content
- reject path traversal outside the workspace root
- enforce read-only `vfs://reference/...` rejection via shared resolver
- reject existing symlink components and nonregular filesystem destinations
- record atomic affected-scope index invalidations before and after attempted filesystem mutation; no automatic refresh
- fail before mutation if index configuration or pre-notification is invalid
- report that bytes may have changed if post-notification fails; explicit reindex is required

For VFS writes, `write` persists into the managed namespace using metadata-safe storage semantics (compressed content in DB, logical plaintext API). VFS writes do not invalidate the filesystem index.

[ADR-0051](../../adr/0051-native-write-index-invalidation.md) defines the bounded notification protocol and crash limits. Caller cancellation does not skip post-notification after an attempted filesystem write. Index content and last-success metadata remain unchanged until explicit refresh. The filesystem/SQLite gap is not crash-atomic; shell/external edits and hard-link aliases still need reconciliation.

## Path semantics
- workspace paths resolve against configured `workspace_root`
- `vfs://skills/...`, `vfs://scripts/...`, etc. resolve into managed namespaces
- `vfs://reference/...` is read-only and must fail on write
