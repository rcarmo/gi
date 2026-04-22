# ADR 0005: Tools, Skills, and Scripting

- Status: Draft
- Date: 2026-04-22

## Decision

### Built-in v1 tools
- shell
- read
- write
- edit
- messages
- schedule
- exit
- script

### Script runtime

`script` runs **Joker** and has full access to:
- Gi internal state
- go-ai state
- tools
- memory
- scheduling/background tasks
- extension hooks

### Skills

A skill consists of:
- `SKILL.md`
- frontmatter
- associated scripts/assets

Skills can come from:
- workspace discovery
- packaged imports
- database mirrors

### Hooks

v1 supports a smaller hook surface, with Joker first and Go also available.
Required hook points include:
- before provider call
- after provider response
- compaction
- tool start/end
- turn start/end
- schedule/task lifecycle
- error handling

## Roadmap

- MCP/ACP in 1.1
- QuickJS after scripting bridge design is defined
