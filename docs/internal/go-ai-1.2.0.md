# go-ai 1.2.0 alignment

Gi's upgrade branch pins `github.com/rcarmo/go-ai v1.2.0`, release commit `a927454af26e172462b0ed0d0acc3be9978dde38`. That release targets upstream pi `v1.2.0`, revision `06a466935536f8c9bcacb2d5efc0fc021be1bdff`. The local go-ai development checkout has started 1.3 work; it is not the upgrade baseline.

## Runtime boundaries and gaps

Gi uses go-ai's provider APIs, model catalogue and OAuth support. Gi owns its turn loop, tool hooks, compaction, SQLite store and web projections. Updating the module does not automatically adopt go-ai/durable's runtime contracts.

| Area | Current result | Remaining verification or implementation |
| --- | --- | --- |
| Provider API upgrade | Inference, store and turn packages compile and pass their tests with v1.2.0 | Full suite and provider request/replay regressions; no paid-provider acceptance yet |
| Session append atomicity | Message insert plus session timestamp, and event insert plus turn timestamp, now share transactions | Pre-release CPU/allocations and concurrent-write verification |
| Failed append semantics | SQLite trigger tests prove a failed parent update leaves no inserted row; retry succeeds once | Cancellation admission/commit boundary needs an explicit contract audit |
| Tool cancellation/reload | Existing terminal events can be projected for TUI display without entering provider history | Web conversation paging does not expose cancelled codemode results; preserve occurrence identity and pagination when fixing it |
| Successful tool result | Tool result messages and lifecycle events are separate writes | Audit crash gaps and failed write ordering before choosing a unified append model |
| Branch/session history | Gi has its own messages, turns, events, forks and compaction checkpoints | Compare upstream branch/reset/replay and tool-result identity contracts, including reused call IDs |
| Images and model calls | v1.2.0 adds explicit image resize zero-value flags and other durable model/image contracts | Map exposed Gi features to the tagged AI/durable/codemode crosswalks; do not infer support from unused APIs |
| Web frontend | F3 candidate is preserved in the original worktree; four checks passed 24/24 across six browser/layout projects | Cancellation reload fails; integrate candidate only after fixing this gate and testing the final pins |

No storage replacement has been selected. Backward compatibility is not required. A replacement must preserve provider transcript correctness, cancellation, occurrence identity and working web live/reload behaviour.

## Development verification

Go 1.27.1, project-owned test environment, inherited `PI_CODING_AGENT_DIR` and `GI_CODING_AGENT_DIR` unset:

- `go test ./internal/inference ./internal/store ./internal/turn -count=1` passes after the upgrade and transactional append changes.
- Atomic append failure and normal store flow regressions pass with `-race -count=3`.
- These are development checks. Full pre-release CPU, alloc_space and alloc_objects analysis and final web acceptance have not run.

Protected dirty F3 and Joker edits in the original checkout were not changed by this slice.
