# ADR 0003: State and Storage

- Status: Draft
- Date: 2026-04-22

## Decision

Gi uses **SQLite WAL** as the shared state database for web, TUI, and CLI.

SQLite stores:
- sessions
- forks/ancestry
- messages/content blocks
- attachments metadata
- selected attachment blobs
- keychain secrets (encrypted)
- schedules/background tasks
- turn events/checkpoints
- memory indices
- search index metadata
- tool/script run records
- provider usage/cost data
- observability counters and structured events
- mirrored settings/skills/scripts/prompt templates/assets

## Filesystem vs DB

Filesystem remains canonical for:
- workspace project files
- general referenced artifacts
- caches

DB stores/mirrors:
- settings/config mirror
- skills/scripts/prompt templates mirror
- selected generated outputs and image blobs
- thumbnails in a dedicated cleanup table

The DB must generally be queried before the filesystem for mirrored assets.

## Search

SQLite FTS backs search across:
- messages
- files/indexed content
- notes/memory
- skills/scripts
- tasks/schedules where relevant
- attachments by filename, metadata, and extracted text
