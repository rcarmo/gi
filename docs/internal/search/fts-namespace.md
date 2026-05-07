# `fts://` namespace contract (read-only virtual search)

## Status
Implemented and wired through the shared tool/script read path.

## Goal
Expose searchable runtime/workspace data through URL-like locators so models can discover information without SQL.

`fts://` is intentionally **read-only** and is treated as a virtual namespace resolved by the same path layer used for workspace and `vfs://` paths.

---

## Read-only and resolver semantics

- `fts://...` resolves to virtual content via `internal/tools/path_resolver.go`
- writes to `fts://...` are rejected (`read-only` error)
- any component using shared read-path resolution can consume it:
  - builtin `read` tool
  - `/api/tools/execute` `read`
  - script bridge `readFile(...)`

---

## URL grammar

General form:

```text
fts://<target>?q=<query>&limit=<n>&session=<session-id>&glob=<glob>&ns=<namespace>
```

Common query parameters:

- `q` / `query` — search query text (required for data targets)
- `limit` — max results (default `20`, bounded)
- `session` — optional session filter for message/turn targets
- `glob` — optional workspace glob filter
- `ns` / `namespace` — optional workspace namespace selector

---

## Built-in targets

### `fts://help`
Human-readable help page with target list + namespace hints.

### `fts://namespaces`
Index of workspace namespace profiles and hint queries.

### `fts://messages?...`
Searches message content and returns markdown with pointers into `vfs://chat/...` message docs.

### `fts://turns?...`
Searches turn prompt/metadata text and returns pointers into `vfs://chat/...` turn docs.

### `fts://workspace?...`
Searches workspace files (text files only), optional `glob`, optional `ns` namespace profile.

### `fts://all?...`
Unified sectioned result document combining `messages`, `turns`, and `workspace`.

---

## Workspace namespace profiles

These are semantic filters over workspace paths/content, used by:

- direct target form: `fts://<namespace>?q=...`
- generic target form: `fts://workspace?ns=<namespace>&q=...`
- unified form: `fts://all?ns=<namespace>&q=...`

### `gi` (alias: `core`)
Scope: runtime/core code, docs, scripts.

Typical hint queries:

- `fts://gi?q=steering+queue`
- `fts://gi?q=subturn+parent_turn_id`
- `fts://workspace?ns=gi&q=hook+approve_tool`

### `go-joker` (alias: `joker`)
Scope: Joker runtime/bridge and `.joke` surfaces.

Typical hint queries:

- `fts://go-joker?q=register+event+hook`
- `fts://go-joker?q=setSessionState`
- `fts://workspace?ns=go-joker&q=tool_call`

### `tooling` (aliases: `tools`, `ops`)
Scope: tool resolver/execution paths, web tool API, operational scripts.

Typical hint queries:

- `fts://tooling?q=ResolveToolPath`
- `fts://tooling?q=api/tools/execute`
- `fts://workspace?ns=tooling&q=vfs://chat`

---

## Output contract

All `fts://` targets return markdown text designed for model consumption.

Common characteristics:

- frontmatter-like metadata block at top
- stable section headings (`#`, `##`)
- source pointers where possible (especially `vfs://chat/...`)
- bounded excerpts rather than full file dumps

`messages` and `turns` always include canonical chat-tree source links:

- `vfs://chat/sessions/<session-id>/messages/<message-id>.md`
- `vfs://chat/sessions/<session-id>/turns/<turn-id>.md`

---

## Error contract

Examples:

- unknown target → `unknown fts target: <target>`
- missing query for data targets → `fts://<target> requires q=... or query=...`
- unknown namespace → `unknown workspace namespace: <ns>`

---

## System prompt hinting guidance (for later)

Recommended instruction snippets:

1. “Before using SQL for discovery, try `fts://messages`, `fts://turns`, or `fts://all`.”
2. “For runtime code search, prefer namespace shortcuts (`fts://gi`, `fts://go-joker`, `fts://tooling`).”
3. “When investigating chat behavior, follow links into `vfs://chat/...` for raw source documents.”
4. “Use `fts://help` / `fts://namespaces` to discover available namespace hints.”

---

## Related docs

- `docs/internal/vfs/chat-projection.md`
- `docs/internal/vfs/urls.md`
- `docs/internal/reference-system.md`
