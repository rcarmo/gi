# Tool: `edit`

## Status
Planned.

## Purpose
Edit an existing text file by applying exact replacements to current file content.

## Planned semantics
- read current content
- validate that each replacement matches exactly once unless explicitly relaxed in a future version
- write back updated content
- reject path escapes
- later support both workspace paths and `vfs://` URLs

## Why it matters
`edit` is the safest high-signal mutation primitive for agent-driven file changes and should remain preferred over whole-file rewrites when making localized changes.
