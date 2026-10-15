# gi docs

[Features and Piclaw parity](feature-parity.md) gives the current browser, terminal and integration status, including the latest fixtures-vibes compliance results and open issues. tsnet web access and Iroh chat are planned; physical passkey devices are untested.

## Structure

- `adr/` — architecture decision records (0001–0060)
- `checklists/` — phased implementation checklist by subsystem
- [Implementation notes](implementation/README.md) — feature changes, plans, audits and verification results, grouped by topic; not embedded in the binary
- [Internal reference](internal/README.md) — the shipped `vfs://reference/...` contracts for tools, scripting, hooks, routing, VFS, skills, MCP and runtime behaviour
- `reference/` — the original specification transcript

## Documents

### ADRs

The first eight ADRs set the architecture; 0009–0060 record later decisions on sessions, queues, compaction, search, workspace, terminal rendering and copy/delete behaviour.

- `adr/0001-overall-architecture.md` — runtime, web, inference and turn engine architecture
- `adr/0002-turn-engine-and-recovery.md` — event log, checkpoints and recovery policy
- `adr/0003-state-and-storage.md` — SQLite schema, filesystem versus database, search
- `adr/0004-ui-surface-model.md` — web, TUI and CLI surfaces; Piclaw TypeScript source strategy
- `adr/0005-tools-skills-and-scripting.md` — built-in tools, Joker scripting and hooks
- `adr/0006-sqlite-virtual-filesystem-for-managed-assets.md` — SQLite-backed managed VFS for skills, scripts and templates
- `adr/0007-self-hosted-agent-reference.md` — shipped internal documentation and the `vfs://reference/...` model
- `adr/0008-workspace-hybrid-search.md` — hybrid workspace search using SQLite FTS5, sqlite-vec and gte-go

### Features and tests

* [Feature and parity matrix](feature-parity.md) — current implementation, compliance results, limits and requested integrations.
* [Browser test guide](../tests/ux/README.md) — fixtures-vibes compliance, Gi regressions and runners.
* [Multi-passkey contract](internal/passkey-contract.md) — required enrolment, sign-in and lockout-safety tests; the [per-case review](implementation/audits/passkey-scenario-review.md) records partial and manual gaps.
* [Passkey backend](internal/passkeys.md) — opt-in RP/origin configuration, APIs, storage and browser-test limits.
* [tsnet plan](implementation/plans/peering-tsnet-plan.md) — the existing scaffold and remote-access work.
* [UX audit](implementation/audits/ux-test-audit-2026-09-24.md) and [full web/TUI plan](implementation/plans/full-web-tui-parity-plan.md) — dated September 2026 reviews; their test counts and gaps describe the code at that time.

### Internal reference

- `internal/README.md` — contract and index for the shipped internal docs
- `internal/tools/`, `internal/scripting/`, `internal/hooks/`, `internal/vfs/`, `internal/skills/` — tool contracts, scripting bridges, hooks, managed `vfs://` semantics and skill structure
- `internal/mcp.md`, `internal/codemode.md` — MCP client and codemode engine
- `internal/keychain.md`, `internal/shell-environment.md`, `internal/config-files.md` — secrets, shell environment and configuration lookup
- `internal/routing.md` — routing and route-event behaviour
- `internal/search/` — scoped lexical indexing, plus the planned hybrid/vector design
- `internal/keybindings.md`, `internal/tui-clipboard-media.md` — terminal keybindings and clipboard/media contracts; port details and verification live in [terminal implementation notes](implementation/terminal/README.md)
- `internal/extension-command-semantics.md` — planned extension command registration contract

### Checklists

- `checklists/implementation.md` — phased implementation checklist by subsystem

### Reference

- `reference/chat-transcript-2026-04-22-gi-spec.md` — original specification conversation (verbatim database export)
