# ADR-0044: Rooted index scanning and deterministic line chunks

## Status

Accepted — 2026-09-22. The internal scanner can produce complete snapshots for the scoped refresh API. No background worker, settings integration, native search/status/reindex route or terminal indexing control is connected. Frozen coverage remains 45/236 Classic and 2/42 shared, with 191/40 unmapped.

## Chunking

`internal/search/chunking/lines.go` implements `utf8-lines-8k-128-v1`: contiguous non-overlapping source slices capped at 8,192 bytes and 128 newline characters. It prefers complete lines, preserves CRLF and final newlines, and splits a long physical line only at a UTF-8 boundary. Empty files produce one empty chunk at line 1. Byte offsets are end-exclusive; line numbers count preceding newline characters, matching the refresh-store validation contract.

Invalid UTF-8 or NUL content is rejected by the chunker. The scanner classifies those files as binary and excludes them. Headings, token estimates and language inference are not implemented. Markdown uses the same literal line splitting; this does not fulfil the heading-aware or semantic chunking design in ADR-0008. Changing the splitting policy requires a new chunker version.

## Complete inventory

`internal/search/indexer.ScanScope` requires a resolved scope configured with the line-chunker version. Every selected root must exist as a real directory; a missing root is an error, never an empty inventory. Default `all` therefore requires both `notes` and `.pi/skills`. Optional default-root semantics need an explicit later policy. Callers can construct scopes containing only their required roots.

Traversal uses `os.OpenRoot`, bounded per-directory reads and sorted names. Overlapping configured roots have already been collapsed by `ScopeConfig`. The scanner excludes `.git`, `node_modules`, `.cache` and `generated` directories, and does not blanket-exclude dot directories: `.pi/skills` remains eligible. It excludes symlinks, special entries, unsupported extensions, oversized files and binary content. An explicitly selected root cannot be a symlink, have a symlink ancestor or pass through an excluded directory. Invalid UTF-8/oversized path names fail the scan.

Limits are 10,000 traversed entries, 64 levels below a selected root, 10,000 text documents, 1 MiB per candidate file, 32 MiB total candidate reads and 20,000 emitted chunks. Binary candidates count against the read budget. Final verification rereads at most the same bounded candidate-byte set. Entry/depth/aggregate limits and traversal/read errors fail the entire scan; there is no successful partial inventory. Stable files over 1 MiB are explicit exclusions, not truncated indexed documents.

On Linux/macOS, final-component opens use `O_NOFOLLOW` and `O_NONBLOCK`. This prevents a symlink replacement from being followed or a FIFO replacement from blocking before its type is checked. Handles are restatted and matched to observed regular-file identity before reading. Other platforms have no unverified blocking fallback; Windows remains outside the requested scope. Rooted access confines symlinks, not mount boundaries or arbitrary filesystem latency.

## Change detection and failure

The scanner records path identity, mode, size and modification time for visited entries and directories. It verifies those observations after inventory and rehashes read candidates. Rehashing catches same-size content changes with restored timestamps. It also checks that the workspace path still resolves to the opened directory. Tests cover rename/replacement of the workspace and selected-root ancestor.

This is bounded change detection, **not an atomic filesystem snapshot**. Mutations after an entry's final verification, or unobserved adversarial metadata changes, need later invalidation/refresh. Context cancellation is checked between traversal/read operations; it cannot interrupt every stalled filesystem syscall. No filesystem or database write occurs during scanning.

Any failure returns a zero, incomplete snapshot. The refresh API rejects it; tests explicitly verify that scan failures cannot clear committed memberships/FTS. A successful empty inventory can remove its own old memberships, retaining content owned by another scope. The caller remains responsible for lease renewal and for reporting failure through `Refresh.Fail` with a usable cleanup context.

## Verification

- Chunk tests cover deterministic reconstruction, CRLF/newline preservation, long Unicode lines, source byte/line locations, empty text and rejected input. A 10-second fuzz run passed **2,383,106 executions**.
- Scanner tests cover default notes/skills/all scopes, overlapping roots, eligibility/exclusions, in-scope hidden text, oversized/binary files, bounds, missing/unreadable roots, cancellation, changes during scanning and workspace/ancestor replacement.
- Linux/macOS-specific tests verify final-component symlink rejection and non-blocking FIFO replacement. Permission rejection is exercised for non-root test processes.
- Native scan→commit tests check unchanged chunk identities, same-size/mtime content changes, shared membership retention, final orphan deletion and FTS integrity. Failed scans preserve prior content/count/generation and report failure through the store.
- Full Go tests/vet/build/hook checks, **74/74 functional tests** and **32/32 helpers** pass. Scanner/chunker/search-store race suites pass ×3, including a final rerun after replacement/empty-scope tests.
- Pure-Go application and scanner-test cross-builds pass for Linux and macOS on amd64/arm64. Cross-compilation is not native macOS execution.

An interrupted `make check` did not produce a complete result. A subsequent run hit workspace-disk exhaustion while Playwright wrote artifacts and a build copied its binary. Four disposable dependency-upgrade binaries were removed, preserving database backups/screenshots; the unchanged full check then passed. Cross-build outputs were moved to `/tmp` with available space. The workspace disk remains tight and needs operational attention beyond this slice.

Logs are `/workspace/tmp/gi-index-scan-{check3,helpers,race2,fuzz}.log`; cross-build directory is recorded in `/workspace/tmp/gi-index-scan-build-directory.txt`. No UI code changed, no screenshot was generated, and full browser/PTY matrices were not rerun or credited. Frozen/derived Gherkin counts remain unchanged.

## Next integration

Wire a lease-renewing worker around Begin→ScanScope→Commit/Fail, with bounded cleanup and cancellation, two-worker exclusion and safe restart recovery. Then add settings/root policy, invalidation/background refresh and bounded lexical query/status/reindex surfaces. Browser mapping requires complete frozen criteria; terminal actions require independent three-size tests and no permanent idle rows. The unscoped full-rebuild prototype stays stashed.
