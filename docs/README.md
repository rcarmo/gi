# gi docs

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

### Internal reference
- `internal/README.md` — contract and index for shipped internal docs
- `internal/tools/` — built-in tool contracts
- `internal/scripting/` — scripting runtimes and bridge docs
- `internal/hooks/` — hook and lifecycle docs
- `internal/vfs/` — managed `vfs://` semantics
- `internal/skills/` — skill/package structure
- `internal/routing.md` — runtime routing/route-event behavior and SSE observability
- `internal/search/` — hybrid workspace search/indexing design and planned package layout
- `internal/tui-pi-fit-gap-roadmap.md` — closure status for the Pi-like TUI parity iteration
- `internal/tui-pi-parity-plan.md` — maintained TUI parity plan and implemented slices
- `internal/tui-clipboard-media.md` — TUI clipboard/media support boundaries
- `internal/extension-command-semantics.md` — planned extension command registration contract

### Checklists
- `checklists/implementation.md` — phased implementation checklist by subsystem

### Reference
- `reference/chat-transcript-2026-04-22-gi-spec.md` — original spec conversation (verbatim DB export)
