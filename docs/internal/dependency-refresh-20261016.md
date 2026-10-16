# Dependency refresh — 16 October 2026

Branch: `maintenance/dependency-refresh-20261016`, based on merged Gi `00f7b4e4`. No deployment.

## Updated

- Go toolchain and module minimum to 1.27.2; Makefile race-probe key/docs match and CI reads go.mod. Separate commit cf969417.
- All reachable runtime Go package dependencies via `go get -u ./...` followed by tidy. Includes goja October14 revision, go-git5.19.3, tailscale1.104.1, x/crypto0.58, x/net0.61, x/term0.47 and associated transitives. Direct module update audit reports no newer release.
- Local go-tui fork module dependencies to current versions; upstream latest remains0.22.1. Local runtime changes preserved. Nested Go minimum now1.26.
- Pi codemode prelude/declarations/tool text assets1.0.1→1.3.0. quickjs-wasi remains current3.6.2. Upstream1.3 removed MCP schema wrappers; scripts/codemode-mcp-compat.js explicitly retains Gi's1.0.1 MIT compatibility declarations and the vendoring script splices the guarded output adapter. Description fixture changes reflect the new upstream prose; declaration assertions remain.
- Fixed legacy Adaptive Cards SDK URL to serve the relocated pinned common asset. Updated asset integrity expectation for rebuilt frontend and CSP fixture path to viewers3.3.4. GET/HEAD/CSP checks preserved.

## Current / exceptions

- go-ai1.3.0, go-joker/v42 v42.12.4, go-tui0.22.1, frontend npm packages already current at audit. Frontend pin remains7a25917 (merged upstream as b0ad91a); no redundant frontend version-only commit.
- gVisor retains Tailscale1.104.1's required revision a6f909f08a72. Latest5b88ae5f7c64 failed normal Go package loading with mixed stack/bridge and refs/refs_template packages. Toolchain update alone does not solve that build-layout incompatibility.
- Do not blindly update all unused transitive graph nodes or tool-only modules from dependency go.mod files. Runtime/package updates follow the actual build graph; go list -m -u all also reports graph-only packages not compiled into Gi.
- third_party/joker is retired source, not the runtime module (Gi uses go-joker/v42). Experimental module edits reverted. Its standalone build lacks generated TYPE/linter data; no passing standalone profile claim. Original dirty Joker/main work preserved.
- QuickJS engine consolidation is separate architecture work, not part of this dependency refresh.

## Verification

- Go1.27.2 full root suite:2334 tests/35 packages, exit0; profiled103s wall/53.8s CPU/~788MB peak. CPU/alloc_space/alloc_objects analysed by helper, raw profiles/matching binaries disposed. No equivalent-workload speedup claim.
- Vet and CGO-disabled build pass. Changed codemode/scripting/static-asset/CSP regressions race×3 pass.
- Nested go-tui tests pass. Nested profiling executed; captures too short/missing samples must not be described as useful allocation evidence. Root TUI tests exercise the actual fork during the full profiled workload.
- Rebuilt fixture backend:42/42 Chromium/WebKit phone/tablet/desktop consumer checks, zero retries. Chromium21 CPU+21 sampled-heap captures analysed: CPU mainly program/idle, allocations mainly anonymous/V8 setup; WebKit functional-only. Captures and disposable fixture run removed after exit.
- Independent read-only review found no concrete blockers in vendor compatibility, static alias or manifests.
- Initial probes: inherited auth-directory test failures corrected by unsetting overrides; gVisor latest rejected; retired Joker generation failure not counted as pass. Temporary npm sources removed after vendoring. One Bun invocation without a local manifest resolved through /workspace; avoid repeating this by requiring an explicit manifest in the project-owned source-install root.

No merge/deployment is implied by publishing the review branch. Historical broad editor acceptance remains separately open.
