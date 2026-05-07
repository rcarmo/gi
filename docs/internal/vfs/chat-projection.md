# `vfs://chat` projection contract (read-only chat tree)

## Status
Implemented as a virtual read-only namespace.

## Goal
Expose chat/session/turn/message runtime artifacts as browsable markdown files with frontmatter so models can inspect raw source state without SQL.

---

## Namespace semantics

- namespace: `vfs://chat/...`
- class: virtual + read-only
- writes/deletes are rejected (`read-only` error)
- available through all shared read paths (`read`, web tool execute read, script `readFile`)

---

## Tree layout

### Root

- `vfs://chat/README.md`
- `vfs://chat/sessions/index.md`
- `vfs://chat/sessions/<session-id>/...`

### Per session

- `vfs://chat/sessions/<session-id>/session.md`
- `vfs://chat/sessions/<session-id>/messages/index.md`
- `vfs://chat/sessions/<session-id>/messages/<message-id>.md`
- `vfs://chat/sessions/<session-id>/turns/index.md`
- `vfs://chat/sessions/<session-id>/turns/<turn-id>.md`

---

## Document formats

All leaf records (`session.md`, message docs, turn docs) are emitted as:

1. frontmatter-like metadata block
2. markdown body

This keeps docs both:

- human-readable
- model-parseable

### Session document (`session.md`)

Frontmatter includes:

- `kind: "chat/session"`
- `session_id`
- `title`
- `parent_session_id`
- `created_at`, `updated_at`
- `aliases`
- `scope`
- `state`

### Message document (`messages/<id>.md`)

Frontmatter includes:

- `kind: "chat/message"`
- `session_id`
- `message_id`
- `role`
- `created_at`
- `payload`

Body is the raw message content.

### Turn document (`turns/<id>.md`)

Frontmatter includes:

- `kind: "chat/turn"`
- `session_id`
- `turn_id`
- `status`, `phase`
- `created_at`, `updated_at`
- `claimed_by`, `claimed_at`, `started_at`, `finished_at`
- `metadata`

Body includes the prompt text.

---

## Index documents

- `sessions/index.md` lists all sessions and links to each subtree
- `messages/index.md` lists session messages + links to each message doc
- `turns/index.md` lists session turns + links to each turn doc

Index pages are intended as traversal aids for agents.

---

## Recommended usage patterns

1. Start at `vfs://chat/sessions/index.md` for discovery.
2. Follow specific session links.
3. Read message/turn leaf documents for raw source details.
4. Pair with `fts://messages` / `fts://turns` for fast retrieval + deterministic source links.

---

## System prompt hinting guidance (for later)

Recommended hints:

- “Use `vfs://chat/sessions/index.md` to discover sessions.”
- “Use `vfs://chat/sessions/<id>/messages/<msg-id>.md` for exact message source with payload frontmatter.”
- “Use `vfs://chat/sessions/<id>/turns/<turn-id>.md` for exact turn lifecycle/metadata source.”
- “Prefer `fts://...` first, then follow returned `vfs://chat/...` links.”

---

## Related docs

- `docs/internal/search/fts-namespace.md`
- `docs/internal/vfs/urls.md`
- `docs/internal/reference-system.md`
