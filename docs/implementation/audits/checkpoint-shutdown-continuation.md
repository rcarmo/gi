# Checkpoint-preserving shutdown and continuation

Retry waits, committed wakes and proven generation/tool frontiers have fenced
restart routes after runtime initialisation. Graceful close requeues verified
checkpoints after joining owned work. Unknown frontiers retain review holds.
See [focused findings acceptance](durable-findings-acceptance.md) for the completed
#83/#86/#92 checks and [graceful retry close](graceful-retry-close.md) for that route.

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
Gi separates graceful close from user abort. Proven retry, generation and tool
frontiers retain their checkpoints. Unsupported frontiers park a review hold
after joining owned hosts; they do not fabricate cancellation or a root terminal
receipt. Explicit user cancellation keeps its terminal settlement path.

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

## Completed generation and tool routes

Generation/tool continuation and process-kill regressions now preserve settled
effects, original prepared arguments and received accounting. A started tool can
run again only when both persisted and current `ReplaySafe` permit it. Recovery
counts no-progress interruptions and parks exhausted checkpoints; graceful Close
leaves that crash budget unchanged.

Explicit retry resumes a proven root checkpoint on its original turn. Parent
consumption guards keep parent-linked retry blocked. Explicit skip stages failed
terminal cleanup, settles outstanding wire calls without executor replay and
joins child final hooks. See [held-turn retry and skip](../../internal/held-turn-retry.md)
for admission, rollback and conservative limits.
