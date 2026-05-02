# ADR 0007: Self-Hosted Agent Reference and Shipped Internal Documentation

- Status: Draft
- Date: 2026-04-23

## Context

gi must be able to explain its own extension surfaces to the agent running inside it.

That means the project needs **shippable internal documentation** for:
- built-in tools
- scripting runtimes and bridge APIs
- hooks and lifecycle events
- managed `vfs://` semantics
- skill/package structure
- extension and import conventions
- operational constraints and testing expectations

Today, some of this knowledge exists only in:
- code
- ADRs
- the implementation checklist
- repo-local `AGENTS.md`

That is not sufficient for a self-extensible system. The agent needs a stable documentation subtree that can be:
- checked into the repo
- reviewed like code
- embedded into the final binary
- exposed later as a read-only `vfs://reference/...` tree

## Decision

gi will maintain a **first-class internal reference subtree** under `docs/internal/`.

This subtree is the canonical author-facing and agent-facing documentation for gi extension surfaces.

The long-term shipped form of this subtree is a **read-only embedded reference filesystem**, likely exposed as:
- `vfs://reference/...`

In the repo, the source of truth remains normal Markdown files under `docs/internal/`.

## Scope of the internal reference

The internal reference must document, at minimum:

### 1. Built-in tool contracts
For each tool:
- purpose
- input schema
- output shape
- side effects
- workspace/vfs path semantics
- error behavior
- examples

### 2. Scripting runtimes
For each runtime (Joker first, others later):
- engine name
- invocation model
- available globals / bridge objects
- file access semantics
- state access semantics
- logging/output behavior
- limitations and safety notes
- examples

### 3. Hook surfaces
For every supported hook point:
- when it fires
- inputs
- outputs / mutations allowed
- ordering guarantees
- retry / failure behavior
- examples

### 4. Managed VFS semantics
- `vfs://` URL format
- read/write/edit rules
- resolver behavior versus workspace paths
- read-only namespaces (for example `vfs://reference/...`)
- export/sync behavior for shell commands

### 5. Skill and package structure
- required files
- frontmatter rules
- associated assets/scripts
- packaged import conventions
- DB/VFS storage expectations

### 6. Documentation maintenance contract
- new internal surfaces must not ship undocumented
- changes to tool schemas, hooks, scripting bridge, or VFS semantics must update `docs/internal/`
- examples should be kept runnable or trivially adaptable

## Documentation layout

The repo should grow a stable subtree like:

```text
docs/internal/
  README.md                 # index and documentation contract
  tools/
    README.md
    read.md
    write.md
    edit.md
    script.md
    ...
  scripting/
    README.md
    joker.md
    bridge.md
  hooks/
    README.md
    lifecycle.md
  vfs/
    README.md
    urls.md
  skills/
    README.md
```

This structure is intentionally boring: stable paths matter more than novelty.

## Design principles

### Docs are part of the runtime surface
If the agent can use it or extend it, it must be documented in `docs/internal/`.

### Markdown-first in repo, embedded at build/runtime
Authoring remains simple Markdown in git. Packaging/embedding happens later.

### Read-only shipped reference
The runtime-facing reference should be immutable from normal agent actions. It is reference material, not mutable workspace content.

### Examples over prose
Every tool, runtime, and hook should include compact examples.

### Stable paths
Once referenced by prompts, tools, or future `vfs://reference/...` URLs, document paths should remain stable.

## Consequences

### Positive
- the agent can discover and explain gi-specific behavior from shipped docs
- new contributors get a clearer extension map
- documentation becomes reviewable alongside code changes
- future embedded reference VFS has a clean source tree

### Negative
- feature work now carries an explicit docs tax
- duplicate high-level information may exist across ADRs and internal reference docs
- doc drift becomes a real maintenance concern if not enforced

## Enforcement policy

The project should adopt the following rule:

> If a change adds or materially changes an internal tool, scripting bridge capability, hook, managed VFS behavior, or skill/package contract, the same change must update `docs/internal/`.

This policy should be reflected in:
- `AGENTS.md`
- `docs/README.md`
- the implementation checklist

A future CI/doc check may enforce presence of internal docs for registered tools/hooks.

## Initial implementation plan

1. create `docs/internal/` with an index and maintenance contract
2. document current built-in tools first
3. document current scripting bridge and Joker runtime
4. document planned `vfs://` rules and shipped read-only reference namespace
5. update repo instructions so internal docs are mandatory for new internal surfaces
6. later: embed `docs/internal/` as read-only `vfs://reference/...`

## Notes

Related documents:
- `0005-tools-skills-and-scripting.md`
- `0006-sqlite-virtual-filesystem-for-managed-assets.md`
- `../checklists/implementation.md`
