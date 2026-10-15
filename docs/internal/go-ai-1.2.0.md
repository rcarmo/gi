# go-ai 1.2.0 alignment

Gi's upgrade branch pins `github.com/rcarmo/go-ai v1.2.0`, release commit `a927454af26e172462b0ed0d0acc3be9978dde38`. That release targets upstream pi `v1.2.0`, revision `06a466935536f8c9bcacb2d5efc0fc021be1bdff`. The local go-ai development checkout has started 1.3 work; it is not the upgrade baseline.

## Runtime boundaries and gaps

Gi uses go-ai's provider APIs, model catalogue and OAuth support. Gi owns its turn loop, tool hooks, compaction, SQLite store and web projections. Updating the module does not automatically adopt go-ai/durable's runtime contracts.

| Area | Current result | Remaining verification or implementation |
| --- | --- | --- |
| Provider API upgrade | Inference, store and turn packages compile and pass their tests with v1.2.0 | Full suite and provider request/replay regressions; no paid-provider acceptance yet |
| Session append atomicity | Message insert plus session timestamp, and event insert plus turn timestamp, now share transactions | Pre-release CPU/allocations and concurrent-write verification |
| Failed append semantics | SQLite trigger tests prove a failed parent update leaves no inserted row; retry succeeds once | Cancellation admission/commit boundary needs an explicit contract audit |
| Tool cancellation/reload | Cancelled/aborted codemode events join the conversation relation before cursor filtering; raw/provider history unchanged. Store paging regressions and six-project browser checks pass | Pre-release profiling and committed frontend-pin verification remain |
| Normal successful/failed tool result | Replay result, terminal boundary and parent timestamps now commit in one transaction; publication and provider continuation follow commit. Injected failures roll back both rows and successful settlement survives reopen | Hook-supplied results remain separate; audit hook path and commit-time cancellation semantics |
| Branch/session history | Gi has its own messages, turns, events, forks and compaction checkpoints | Compare upstream branch/reset/replay and tool-result identity contracts, including reused call IDs |
| Images and model calls | v1.2.0 adds explicit image resize zero-value flags and other durable model/image contracts | Map exposed Gi features to the tagged AI/durable/codemode crosswalks; do not infer support from unused APIs |
| Web frontend | Candidate F3 wiring with v1.2.0 passes 30/30 checks across Chromium/WebKit phone/tablet/desktop, zero retries, including cancellation reload as cancelled | Candidate assets are uncommitted; final-pin verification and profiling remain |

No storage replacement has been selected. Backward compatibility is not required. A replacement must preserve provider transcript correctness, cancellation, occurrence identity and working web live/reload behaviour.

## Development verification

Go 1.27.1, project-owned test environment, inherited `PI_CODING_AGENT_DIR` and `GI_CODING_AGENT_DIR` unset:

- `go test ./internal/inference ./internal/store ./internal/turn -count=1` passes after the upgrade and transactional append changes.
- Atomic append failure and normal store flow regressions pass with `-race -count=3`.
- Full store suite passes `-race -count=3` after the terminal conversation projection.
- Candidate consumer matrix: five checks × six projects = 30 passes, zero retries. Covers live nested settlement/reload, failure, two sessions, cancellation/reload and absent-details compatibility injection. The absent-details check modifies real saved HTTP rows and is not live execution evidence.
- Projection regressions cover equal timestamps, backward/forward paging, reused tool-call IDs with distinct occurrences, first-terminal-wins, session fencing, raw-history exclusion and persisted-result deduplication.
- Atomic settlement tests inject failures at event insert, message insert, turn update and session update. Both rows roll back; cancelled admission fails; a subsequent successful pair survives reopen. Focused race tests pass ×3 and full store/turn tests pass.
- Targeted profile: Go 1.27.1, candidate based on 3a2b4847, 30 repetitions of settlement/reopen, append rollback and conversation paging tests. CPU sampled 1.52 seconds over 3.93 seconds; database/sql locking was 57.24% cumulative and store.Open 38.16%. Allocation history: 124.30 MB / 764,902 objects; regexp compilation was 63.23% cumulative alloc_space and 50.89% alloc_objects. This workload repeatedly opens databases and validates schemas; it does not isolate steady-state tool settlement. No equivalent-workload speedup measured; regex attribution needs deeper caller analysis before tuning. Captures and matching binary were deleted immediately after analysis.
- Full pre-release profiling across runtime/frontend workloads and committed-pin web acceptance have not run.

Protected dirty F3 and Joker edits in the original checkout were not changed by this slice.
