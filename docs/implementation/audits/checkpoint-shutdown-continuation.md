# Checkpoint-preserving shutdown and continuation

Retry waits and committed wakes have a fenced restart route after runtime
initialisation. Graceful close requeues these verified checkpoints after joining
owned work. General generation/tool continuation still needs implementation;
#83/#86/#92 stay open. See [graceful retry close](graceful-retry-close.md).

## Verified working-tree guards

- `ResumeProviderRetry` rejects an older wait when a newer scheduled wait or a
  later effect/progress/terminal boundary exists. The superseded-wait and first
  eleven boundary cases reproduced failures before the fix. The final fourteen
  boundary cases pass; the complete Go suite and focused wake race ×3 pass.
  See `atomic-provider-retry-checkpoints.md`.
- Shared session coordination does not transfer host ownership between engines.
  A regression reproduced one engine's `Close` cancelling another engine's host
  on the same `Store`. Host registration now records its immutable engine owner;
  close cancellation skips foreign hosts. Close, parent-join and staged-terminal
  recovery tests pass three times with race detection; the complete Go suite
  also passes with both guards. Final repeated retry/ownership race checks,
  full Go/vet/build/Bun and browser147/13 pass. Publication is separate.

## Native shutdown contract

The oracle is go-ai v1.4.0 `durable/harness.go:Close` and
`durable/close_requeue.go:requeueAtClose`: seal admission, cancel/join scheduler
and invocation hosts, then requeue running task checkpoints. Preserve abort
marks and restart counts; graceful close does not manufacture a crash outcome.
Gi separates graceful retry close from user abort. Other generation/tool
phases still use conservative abort settlement until their continuation routes
are verified. Every path retains the host join barrier.

A pending checkpoint must not enter Gi's normal whole-prompt queue launcher.
That launcher reruns setup and cannot implement native continuation safely.
Keep conservative holds until the corresponding continuation route is verified.

## Completed retry route

Versioned lifecycle values and tool-failure budgets now survive retry-wait
recovery. Takeover rechecks the frontier, receipt, token and heartbeat before
rotating ownership and its permanent writer fence. The runtime-ready barrier
prevents early provider dispatch. Current context/tool preparation, fixed
iteration/attempt/deadline, cancellation and real SIGKILL/reopen checks pass.
See [retry continuation](provider-retry-continuation.md) for checks and profiling.
The [post-wake route](provider-retry-post-wake.md) also handles a committed wake
before the next request admission, without duplicating the wake or budgets.

## Next implementation boundary

1. Extend graceful close beyond verified retry checkpoints. Join actual
   hosts/progress/hooks before requeue or claim cleanup; preserve explicit
   cancellation and propagate persistence errors.
2. Add generation/tool boundary continuation and process-kill regressions.
   Never re-execute settled tools. Unknown tool effects require both persisted
   and current `ReplaySafe`; retain prepared arguments without rerunning hooks.
   Count crash interruptions, reset only after committed forward progress, and
   park exhausted/no-progress tasks. Graceful close must not increment counts.
