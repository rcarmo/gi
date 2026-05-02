# Tool: `read`

## Status
Implemented and wired via `/api/tools/execute`.

## Purpose
Read text content from a workspace file or managed VFS asset.

## Input
```json
{
  "path": "relative/path.txt"
}
```

### Fields
- `path` — workspace-relative path (`.`, `docs/foo.txt`, etc.) or `vfs://namespace/path`

## Output
Success:
```json
{
  "result": "file contents as text",
  "error": ""
}
```

Error:
```json
{
  "result": "",
  "error": "..."
}
```

## Path semantics
- workspace paths resolve against configured `workspace_root`
- `vfs://` paths resolve through the shared resolver:
  - `vfs://skills/...`
  - `vfs://scripts/...`
  - `vfs://reference/...` (read-only)

The resolver enforces safety checks:
- empty path rejection
- workspace traversal rejection
- `vfs` read/write semantics (including read-only reference namespace)

## Failure modes
- missing file/path
- path traversal rejection (`path escapes workspace`)
- missing or invalid `vfs://` path
- binary content handling is currently text-oriented (`result` returns raw bytes as UTF-8 string)
