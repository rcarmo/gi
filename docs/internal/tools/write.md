# Tool: `write`

## Status
Implemented and wired via `/api/tools/execute`.

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

For VFS writes, `write` persists into the managed namespace using metadata-safe storage semantics (compressed content in DB, logical plaintext API).

## Path semantics
- workspace paths resolve against configured `workspace_root`
- `vfs://skills/...`, `vfs://scripts/...`, etc. resolve into managed namespaces
- `vfs://reference/...` is read-only and must fail on write
