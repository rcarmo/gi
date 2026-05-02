# ADR 0006: SQLite Virtual Filesystem for Managed Assets

- Status: Draft
- Date: 2026-04-23

## Context

gi needs a **managed virtual filesystem inside SQLite** for agent-managed assets such as:
- skills
- scripts
- prompt templates
- packaged imports
- small associated assets and metadata

This is not the same as the user workspace tree on disk. The workspace remains the place for ordinary project files. The new filesystem is for **agent-managed content** that benefits from:
- transactional updates
- history/versioning hooks
- DB-first lookup
- uniform access from web, TUI, CLI, internal tools, and scripts
- packaging/import/export without depending on host paths

The key UX requirement is that internal surfaces should be able to use **`vfs://...` URLs directly**.

Examples:
- `vfs://skills/foo/SKILL.md`
- `vfs://scripts/lib/util.joke`
- `vfs://templates/reply.md`

ADR 0003 already established SQLite as the shared state store and noted that skills/scripts/templates may be mirrored there. This ADR refines that decision: for managed assets, gi should expose a real application-level virtual filesystem model in SQLite rather than a loose collection of mirrored rows.

## Research: useful Go facilities and libraries

### Standard library features to build around

#### `io/fs`
Go's `io/fs` package is the most useful foundation for the read side.

Why it matters:
- gives gi a standard read-only filesystem contract
- uses slash-separated paths consistently across platforms
- integrates with `fs.WalkDir`, `fs.ReadFile`, `fs.ReadDir`, `fs.Glob`, and `fs.Sub`
- makes it easy to expose DB-backed assets to template loaders, script discovery, and packaging code

The internal VFS should therefore implement at least:
- `fs.FS`
- `fs.ReadFileFS`
- `fs.ReadDirFS`
- `fs.StatFS`

Optional later:
- `fs.GlobFS`
- `fs.SubFS`

#### `path` instead of `filepath`
Inside the virtual filesystem, paths should use **portable slash semantics** and therefore use `path`, not `filepath`, for normalization and joining.

#### `testing/fstest.MapFS`
`fstest.MapFS` is useful as a mental model and for tests/fixtures, but not as the production storage engine.

### Go libraries worth knowing about

#### `github.com/spf13/afero`
Afero is useful as a compatibility layer and for tests. It provides a mature mutable filesystem abstraction and adapters to/from standard library patterns.

Usefulness for gi:
- good reference for API shape
- useful in tests and adapter code
- not ideal as the core persistence model, since gi still needs explicit SQLite schema, transactions, versioning, and metadata control

#### `github.com/go-git/go-billy/v5`
Billy provides another filesystem abstraction with stronger "real filesystem" ergonomics than `io/fs`.

Usefulness for gi:
- worth considering if future Git/package workflows want a mutable FS abstraction
- still better treated as an adapter target than the storage model itself

#### FUSE libraries (`bazil.org/fuse`, `github.com/hanwen/go-fuse/v2`)
These are only relevant if gi later wants to mount the SQLite VFS into the host OS for debugging or inspection.

Usefulness for gi:
- potentially useful for developer tooling
- not needed for the core design
- should not shape the internal schema

#### SQLite VFS hooks
SQLite itself has a VFS layer, and some Go SQLite drivers expose VFS-related hooks. That mechanism is for changing **how SQLite reads its own database file**, not for representing an application-level filesystem stored as rows.

Conclusion:
- SQLite VFS is **not** the right primitive for this feature
- gi should build an application-level VFS schema on top of ordinary SQLite tables

## Decision

gi will implement a **DB-backed managed virtual filesystem** for agent-managed assets, with:
- `vfs://...` URLs as the first-class path form
- SQLite as the canonical store
- `io/fs`-compatible read APIs
- a small gi-native mutable API for writes and edits
- a shared path resolver so internal tools and runtimes can work with both workspace paths and `vfs://` URLs

This VFS will be used for:
- skills
- scripts
- prompt templates
- packaged imports
- other small agent-managed assets
- later, read-only shipped reference content such as `vfs://reference/...`

This VFS will **not** replace the normal workspace filesystem on disk.

## Design direction

### 1. Separate namespaces: workspace FS vs managed VFS

gi should keep two distinct storage surfaces:
- **workspace FS** — host filesystem under the workspace root
- **managed VFS** — SQLite-backed filesystem for agent-managed assets addressed as `vfs://...`

The agent and runtime can search both, but they must remain conceptually separate.

### 2. `vfs://` should be first-class for internal tools and scripts

Anything inside gi that already works on paths should learn to accept:
- normal workspace paths
- `vfs://` URLs

This includes, at minimum:
- `read`
- `write`
- `edit`
- script loaders
- skill loaders
- template loaders
- scripting bridge file helpers

### 3. Use a shared resolver layer

Internal code should not duplicate path-kind logic. gi should have a shared resolver, conceptually like:
- `ResolveRead(name string) -> fs.FS + cleanPath`
- `ResolveMutable(name string) -> backend + cleanPath`

This allows the same call sites to work with either workspace paths or managed `vfs://` URLs.

### 4. Read side should look like `io/fs`

The managed VFS should expose an adapter that satisfies:
- `fs.FS`
- `fs.ReadFileFS`
- `fs.ReadDirFS`
- `fs.StatFS`

That allows direct reuse with:
- script discovery
- skill loaders
- template loaders
- `fs.WalkDir`
- `fs.Glob`
- archive/export helpers
- future busybox-style internal commands

### 5. Write side should be gi-native

`io/fs` is read-only by design. gi should define a small mutable interface for the managed VFS, for example:
- `WriteFile`
- `ReadFile`
- `MkdirAll`
- `Remove`
- `Rename`
- `Stat`

This mutable layer should be transaction-aware and explicit.

### 6. Use inode-like tables, not path-as-primary-key only

A path-only table is simple but becomes awkward for rename, dedupe, metadata, and future versioning.

Prefer an inode-like model:
- `vfs_nodes`
  - id
  - parent_id
  - name
  - kind (`file`, `dir`, later `symlink`)
  - mode
  - size
  - content hash
  - created_at / updated_at
  - revision
- `vfs_blobs`
  - content hash
  - bytes or chunk reference
  - size
  - compression flag
- optional `vfs_chunks`
  - for larger payloads later
- optional `vfs_mounts` / `vfs_packages`
  - provenance, package source, imported bundle metadata

Key properties:
- directories are first-class nodes
- rename updates metadata/parent linkage, not every descendant row
- content-addressed blobs enable dedupe
- future snapshots/versioning remain possible

### 7. Paths should be normalized in a strict, portable way

Managed VFS paths should:
- always use `/`
- reject `..` escapes after normalization
- reject empty path segments except root
- treat root and relative-open behavior consistently
- avoid host-OS-specific semantics
- parse `vfs://` URLs into a normalized managed path form

Use `path.Clean` and additional validation, not `filepath.Clean`.

### 8. Bash uses sync-out / sync-back, not direct live VFS access

Bash and ordinary subprocesses do not understand `io/fs` directly.

So the initial shell strategy is:
- export a selected managed subtree to a temp directory
- run bash or another subprocess there
- optionally sync changes back into the managed VFS

This is an explicit adapter layer, not the primary storage model.

### 9. `go-busybox` is a promising middle layer

`rcarmo/go-busybox` is a good fit for future internal command surfaces because it can potentially operate over `io/fs` plus the mutable managed backend.

That suggests useful execution tiers:
1. native internal tools — direct `vfs://`
2. busybox-style internal commands — direct `vfs://` over `io/fs` and the writable backend
3. real bash — sync out and sync back

### 10. Keep payload strategy simple at first

Initial scope should optimize for the actual target data:
- markdown
- scripts
- JSON/YAML
- small assets

Start with inline blobs in SQLite rows or a simple blob table.
Only add chunking/compression thresholds when real data size justifies it.

### 11. Versioning should be designed in, even if delayed

The schema should leave room for:
- optimistic revision numbers
- import/export provenance
- snapshots or copy-on-write later
- audit events for script/skill changes

Even if full history is deferred, node revisions and change records should be anticipated.

## Consequences

### Positive
- managed assets become transaction-safe and portable
- internal tools and scripts can use `vfs://` directly
- easier packaging/import/export of skills and templates
- read path integrates cleanly with standard Go `io/fs` tools
- future DB search/indexing is simpler because content and metadata are already local
- later read-only shipped docs can live under the same model (`vfs://reference/...`)

### Negative
- adds a second filesystem concept to gi
- requires careful path normalization and resolver semantics
- `io/fs` covers reads well but not writes, so gi must own the mutable API design
- shell compatibility requires export/sync adapters until a richer process integration exists

## Rejected alternatives

### Rejected: plain mirror tables without filesystem semantics
This is too weak for rename, packaging, tree traversal, and future versioning.

### Rejected: path-only key/value storage as the long-term model
This is tempting for v1, but makes rename, metadata, dedupe, and snapshots harder.

### Rejected: using SQLite's own VFS mechanism
Wrong abstraction layer. SQLite VFS controls database file IO, not an application filesystem stored inside tables.

### Rejected: replacing the workspace root entirely with the DB VFS
The real workspace must stay available for ordinary project files and host tool interoperability.

### Rejected: making bash the primary design driver
Internal Go tools and scripts should use `vfs://` directly. Bash compatibility is important, but should be handled by export/sync adapters rather than forcing the whole design around subprocess path semantics.

## Initial implementation sketch

Phase 1 for this feature should:
1. add schema for directories, files, and content blobs
2. implement strict path normalization using `path`
3. define `vfs://` URL parsing and canonicalization rules
4. implement a shared resolver for workspace paths vs `vfs://` URLs
5. implement an `io/fs` adapter for managed reads
6. implement basic writable operations: `WriteFile`, `ReadFile`, `MkdirAll`, `ReadDir`, `Remove`, `Rename`, `Stat`
7. make internal tools like `read`, `write`, and `edit` resolve `vfs://` directly
8. route scripting bridge file helpers through the same resolver
9. store skills/scripts/templates in the managed VFS instead of loose mirrored rows
10. expose export/sync helpers between managed VFS paths and temp workspace directories for bash
11. evaluate `go-busybox` as a direct consumer of the `io/fs` + writable backend model

## Notes

Related documents:
- `0003-state-and-storage.md`
- `0005-tools-skills-and-scripting.md`
- `0007-self-hosted-agent-reference.md`
- `../checklists/implementation.md`
- `../internal/vfs/urls.md`
