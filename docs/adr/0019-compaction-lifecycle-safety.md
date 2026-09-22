# ADR-0019: Commit compaction outcomes before publishing success

## Status

Accepted — 2026-09-22. Native automatic-compaction prerequisite; browser parity is unchanged.

## Problem

Automatic compaction previously set `running/compacting`, then unconditionally restored `running/running` in a defer. That could overwrite a concurrent cancellation. It replaced the in-memory context and announced completion before persisting its summary. A failed or cancelled write could leave subscribers reporting success without a durable summary. Repeated compactions in one turn reused the same summary message ID.

## Implementation

`Store.BeginCompaction` conditionally enters `compacting` only for the session's claimed, running turn. The phase change and `compaction.started` event share a transaction. The event sequence identifies this compaction occurrence.

`Store.FinishCompaction` accepts only the latest unclosed start sequence. It commits one terminal event, an optional summary message and conditional phase restoration in one transaction. Summary IDs include the terminal event sequence. Late/duplicate completion cannot finish another occurrence. A cancellation or missing claim changes the accepted outcome to `cancelled`, with no summary write. Phase restoration never overwrites cancelling/terminal status or touches another claim.

The runtime publishes `compaction_started`, followed by exactly one successfully persisted `compaction_completed`, `compaction_cancelled`, `compaction_suppressed` or `compaction_failed` outcome. Payloads include the session, turn and `started_seq`; pre-compaction token counts are explicitly estimates. These events use the existing engine broadcast path and `session.compaction` topic. The legacy completion-only `compaction` notice remains available to TUI subscribers after durable success.

The context is replaced only after a completed transaction. Hook suppression/error retains the original context and records a terminal outcome. Cancellation also retains it. A completion-write failure rolls back the event and summary together, attempts a failure checkpoint, and prevents the next inference request. Terminal writes use the engine's durable context so user cancellation can still be recorded. If storage remains unavailable, the turn fails; no successful compaction event is published.

The next provider iteration checks cancellation and conditionally changes only the phase of a claimed, running turn. It cannot restore status after a cancellation committed during context hooks.

## Verification

- Store tests: atomic start/event rollback; completion/summary rollback; repeated and stale occurrence rejection; unique summaries across repeated compaction; cancellation wins without status resurrection; eight concurrent completion/cancellation races per run.
- Runtime tests: start failure, cancellation, suppression, hook error, completion-write failure, unchanged original context, and persistence before success notification.
- Engine integration: a real turn and hook checkpoint for completion/cancellation/suppression/write failure. The provider stub receives the compacted context only after successful persistence; cancellation/write failure makes no inference call. A separate test cancels in the before-provider hook and checks that provider entry rejects it.
- Existing direct provider-hook tests now establish the claim required by production. The heartbeat test waits for asynchronous session cleanup after terminal turn status, without retries or weakened assertions.
- Full Go suite, vet, hook checks, targeted race suite repeated three times, and the existing isolated functional browser suite (70/70) pass.

## Boundaries

This change does not enable a manual Compact action, add web compaction status, or earn any new frozen-feature pass. UI cancellation/elapsed labels, reconnect ownership, temporary suppression notices and post-completion usage refresh still need browser acceptance. Screenshot captures supplied earlier show `eba8955`, before this prerequisite.

Stored summaries still accompany the original history on subsequent turns. This change improves the current turn's context and durable audit; it does not implement a persisted history boundary or guarantee token reduction. The default summary remains the existing transcript-excerpt heuristic. External-provider exactly-once execution and crash fault-injection tests are not covered.

TUI layout is unchanged. Completion retains the existing transcript/status rendering. Future active compaction belongs in the existing transient status/footer and stop interaction, with no top chrome, permanent panel or additional idle row. Verify 60×18, 100×22 and 140×36 when those controls change.
