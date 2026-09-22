# ADR-0036: Admit a distinct prompt after turn completion

## Status

Accepted — 2026-09-22. Fixes the display-idle/admission-claim race recorded in ADR-0035. No new frozen browser mapping.

## Failure and admission rule

A turn can be durably completed and display idle while its final hook and cleanup still own the session's active claim. Previously, `SubmitPrompt` treated any active claim as a steering target. A prompt submitted in that interval returned the completed predecessor's ID with `status: running`, even though that run could no longer process it normally.

The native regression holds a real terminal session-state hook before claim release. Against the previous code, it fails with the second submission pointing to the exhausted turn. With the fix, it returns a distinct durable queued turn, retains the predecessor's claim until cleanup, then executes the new prompt exactly once in the tested run.

`EnqueueActiveSteering` checks the matching session/turn claim, running status and non-maintenance operation in the same SQLite transaction as steering insertion and running-state normalisation. Completion cannot interleave those writes. A terminal, cancelling, missing or replaced claim returns `ErrQueueConflict`; queue capacity and persistence errors remain errors rather than becoming silent new-turn fallbacks.

For ordinary prompt admission, that conflict causes a distinct queued turn to be persisted. Cleanup remains responsible for releasing the predecessor's claim and launching queued work in order. If another Store/engine has already released the claim, admission checks again after its events are durable and attempts queued launch, avoiding a stranded fallback row. The already-persisted launch-conflict fallback also refuses to steer a terminal turn.

A still-running or cancelling manual-compaction operation keeps its existing conflict response. Completed maintenance no longer blocks an ordinary prompt indefinitely. `SteerQueuedTurn` remains strictly run-bound and has no idle/new-turn fallback. The generic `EnqueueSteering` store API remains available for continuation/recovery work; engine live admission uses the new checked method.

## Lifecycle and UI

Display-idle still means the visible turn is terminal, not that cleanup has released every resource. The safe response during that interval is a distinct queued admission. This avoids early claim release and prevents a successor from overlapping predecessor cleanup. Existing queue surfaces show that admission; no browser component or TUI row/control is added.

An ordinary prompt arriving while a genuinely running turn still accepts steering continues to use the existing steering/continuation behavior. A turn completing immediately after successful atomic steering admission can still use the established continuation path. This fix does not promise exactly-once external-provider processing or change recovery semantics for process crashes.

## Evidence

Native tests cover:

- actual completed-shell hook held across display-idle, retained claim, distinct next-turn ID, no steering row and one stored user message per prompt;
- completed/failed/cancelled/aborted/cancelling claims, terminal maintenance and FIFO drainage of three admissions;
- launch-claim conflict with a terminal predecessor retaining the already-persisted prompt;
- stale/foreign/missing claims, live-maintenance rejection, capacity rejection and transactional rollback after session-state failure;
- 25 completion/admission races using two Store connections, followed by guaranteed rejection after completion is committed.

The native browser case in `queue-steer.spec.mjs` holds the same real terminal hook through a file gate, observes display-idle, submits from the composer, verifies a distinct queued HTTP acknowledgement and retained predecessor identity, then releases cleanup. It verifies two durable turns, one user record each, timeline deduplication and newer-draft preservation through reload. No fabricated SSE or seeded lifecycle rows are used.

Final validation: **408/408 browser executions** (246 main, 54 reconnect, 66 compaction, 18 Steer/admission, 12 context-fit, 12 meter), **70/70 functional**, **29 helpers**, full Go tests/vet, targeted store/turn race tests repeated three times, and all existing three-size outcome/search/selection/reading/regular/session/model/compaction plus terminal smoke/Gherkin suites. Main browser results combine two successful 123-case runs. A read-only review delegate timed out and supplied no review evidence.

Paging acceptance previously assumed two rows per turn and failed when safe queued admission legitimately added a native status row. It now compares the exact persisted message-ID suffix before and after reconnect, retaining its no-gap/no-duplicate and viewport-anchor assertions. Frozen feature files are unchanged.

Coverage remains **35/236 Classic**, **2/42 shared**, with **201/40 unmapped**. This is native admission-safety evidence rather than a new feature mapping. ADR-0035's unresolved admission-boundary warning is superseded; other crash/replay, terminal recovery, reflow/theme and web-family gaps remain open.
