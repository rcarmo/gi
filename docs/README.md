# gi docs

Start with [features and Piclaw parity](feature-parity.md) for the current browser,
terminal and integration status. It separates shipped behaviour from partial
support, source-test mappings and planned work. Multi-passkey APIs and Settings/login have opt-in support with virtual-authenticator
tests. Physical-device checks, policy controls, tsnet web access, Iroh chat and the
token-saving MCP gateway are not implemented yet.

## Structure

- `adr/` — architecture decision records
- `checklists/` — phased implementation checklists by subsystem
- `internal/` — canonical internal reference docs for tools, scripting, hooks, routing, VFS, and skills
- `reference/` — transcripts and reference material

## Documents

### ADRs
- `adr/0001-overall-architecture.md` — runtime, web, inference, turn engine architecture
- `adr/0002-turn-engine-and-recovery.md` — event log, checkpoints, recovery policy
- `adr/0003-state-and-storage.md` — SQLite schema, filesystem vs DB, search
- `adr/0004-ui-surface-model.md` — web/TUI/CLI surfaces, Piclaw TypeScript source strategy
- `adr/0005-tools-skills-and-scripting.md` — built-in tools, Joker scripting, hooks
- `adr/0006-sqlite-virtual-filesystem-for-managed-assets.md` — SQLite-backed managed VFS for skills, scripts, and templates
- `adr/0007-self-hosted-agent-reference.md` — shipped internal documentation and future `vfs://reference/...` model
- `adr/0008-workspace-hybrid-search.md` — hybrid workspace search using SQLite FTS5 + sqlite-vec + gte-go

### Features and tests
* [Feature and parity matrix](feature-parity.md) -- current implementation, limits and requested integrations.
* [Browser suite guide](../tests/ux/README.md) -- frozen contracts, runners and specialised fixtures.
* [UX audit](internal/ux-test-audit-2026-09-24.md) -- source review and gaps in interaction/visual testing.
* [Full web/TUI plan](internal/full-web-tui-parity-plan.md) -- scoped implementation and verification work.
* [Multi-passkey contract](../tests/ux/features/additions/piclaw-2026-09-24/README.md) -- required enrolment, sign-in and lockout-safety tests; native APIs and Settings/login journeys exist; the complete per-case mapping is outstanding.
* [Passkey backend](internal/passkeys.md) -- opt-in RP/origin config, APIs, storage and browser-test limits.
* [tsnet plan](internal/peering-tsnet-plan.md) -- existing scaffold and remote-access work.

### Internal reference
- `internal/README.md` — contract and index for shipped internal docs
- `internal/tools/` — built-in tool contracts
- `internal/scripting/` — scripting runtimes and bridge docs
- `internal/hooks/` — hook and lifecycle docs
- `internal/vfs/` — managed `vfs://` semantics
- `internal/skills/` — skill/package structure
- `internal/routing.md` — runtime routing/route-event behavior and SSE observability
- `internal/search/` -- implemented scoped lexical indexing/reindexing plus planned hybrid/vector work
- `internal/tui-pi-fit-gap-roadmap.md` — closure status for the Pi-like TUI parity iteration
- `internal/tui-pi-parity-plan.md` — maintained TUI parity plan and implemented slices
- `internal/tui-clipboard-media.md` — TUI clipboard/media support boundaries
- `internal/extension-command-semantics.md` — planned extension command registration contract

### Checklists
- `checklists/implementation.md` — phased implementation checklist by subsystem

### Reference
- `reference/chat-transcript-2026-04-22-gi-spec.md` — original spec conversation (verbatim DB export)
