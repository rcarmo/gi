# Historical workspace editor verification

These measurements describe the native backend work on 5 October 2026. Current shared browser acceptance is in [workspace editor acceptance](workspace-editor-acceptance.md); the runtime API contract is in [the internal reference](../../internal/workspace-editor-backend.md).

## Verification and performance, 2026-10-05

* Seven focused handler/watcher tests pass, including complete reads, size/encoding/path rejection, unchanged-save mtime, compatibility assets, authentication, shared resources, shutdown and external-write SSE snapshots.
* Full Go suite: 2,127 tests passed, 37 packages; 83.5s wall, 62.2s CPU, 752MB peak RSS including compilation. Versus the preceding run: wall -11%, CPU -2%, RSS -1%; no reported test regression.
* Functional suite: 144 passed, 11 skipped; complete lifecycle 204.1s wall, 98.5s CPU, 370MB peak RSS. Versus the earlier full run: wall/CPU +19%, RSS -8%, with two new tests and filesystem monitoring. No isolated latency regression is established by that aggregate comparison.
* The two new functional checks exercise complete edit/save/oversize rejection and tool-written external changes arriving over SSE without contents. Their selected run passed in 1.6s; the 133.5s lifecycle included a cold build after cache trimming.
* Vet passes. The MCP wrong-issuer test flagged at 1.5s in an intermediate full profile passed its focused profiled check in 8ms; no reproducible slowdown was found.

Focused sampled allocations are mainly bounded preview/JSON responses and test response buffers. The text/content aliases share one string allocation. Watch registration uses bounded directory reads and constant-time watch-set membership rather than allocating a complete watch-list copy per directory. No stable browser allocation measurement or race-detector result is claimed.

The measurements above are historical. Used raw logs/profiles and matching
test artifacts are disposable; current runs use project-owned scratch as
described in [8 October acceptance](../maintenance/acceptance-2026-10-08.md).
