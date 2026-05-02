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

`script` is an engine-abstracted scripting surface with **Joker first** and other runtimes allowed behind the same bridge contract.

#### Hard architectural constraint
Serious scripting and hook execution must run **in-process** and operate against the **live session/turn context**.

A subprocess model with only serialized snapshots is not sufficient for the final design.

#### Minimum live bridge
At minimum, the in-process scripting bridge must expose callable host APIs for:
- `getSessionState()`
- `setSessionState(patch)` or equivalent controlled session-state mutation
- `listMessages()`
- `addMessage()`
- `readFile()`
- `writeFile()`
- `listDir()`
- later, `vfs://` access through the shared resolver

Scripts must be able to:
- introspect the current session state
- mutate the current session state
- do so through controlled bridge APIs rather than raw Go object references

#### Session mutation semantics
Two mutation scopes matter:
1. **persistent session state** — mutations commit to the canonical session record in storage
2. **live turn/session context** — mutations can affect the current running turn and subsequent hook/tool/provider behavior in the same turn

The target design should support both, with the bridge as the only mutation path.

#### Runtime status note
Any current Joker subprocess + snapshot integration is bootstrap-only and must not be treated as the target extension architecture.

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
Hook-capable scripting runtimes must use the same in-process live bridge described above.

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
- Goja is the in-process JavaScript runtime for v1; QuickJS is out of scope in this track.
- Bring Joker to full in-process live-bridge parity for hooks/extensions
- Route scripting file access through the shared workspace + `vfs://` resolver
