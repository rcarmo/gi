# Explicit terminal queue commands

`/queue` now inspects the native durable queue. Removal, steering and relative moves require full turn IDs; there is no implicit resend, retry or idle-submit fallback. Responses appear only when requested, with no permanent queue panel or extra idle rows.

## Commands

- `/queue [page]` reads queued turns in native queue order, six rows per page. It shows total count, page count and the current running turn ID when available. IDs are complete; prompt previews are control-sanitised and limited to40display cells.
- `/queue move <turn-id> before|after <target-id>` reorders the last successful `/queue` snapshot for the current session visit. Both IDs must belong to it. Self/no-op moves are rejected. Native exact-order/permutation and claimed-row guards run in the transaction. Mutation success or conflict clears the snapshot; `/queue` must refresh before another move. Restart and A → B → A visits cannot reuse the old snapshot. Timestamps, prompts, media and metadata remain unchanged.
- `/queue remove <turn-id>` calls `CancelQueuedTurn`. Only an unclaimed queued row belonging to this session can be cancelled. Accepted history and stored media remain. Success publishes the same `session.queue` change notice used by the web adapter.
- `/queue steer <queued-id> <active-id>` calls `Engine.SteerQueuedTurn`, preserving runner locking, transactional source/target validation and native checkpoint admission. Stale, foreign, claimed or non-running IDs fail without creating another prompt. Existing structured media and claim metadata remain in the native steering payload.

Inspect output is advisory and can become stale; mutation guards run again at execution. Page values are positive integers, oversized values fail parsing, and out-of-range pages ask for a refresh. There are no mutable display-row indexes.

Fullscreen appends responses to its ordinary transcript. Regular mode prints this explicit response once above the dock, without flushing partial model output or putting it into the delayed transcript-flush batch. This keeps queue results visible while another turn is active and leaves terminal-owned history/selection unchanged.

Commands use the normal command editor path. Typing a command replaces the editor content deliberately; the queue functions themselves do not alter draft/cursor/undo, process-local queued-draft shortcuts, attachments, selected model or settings. No new keyboard binding is claimed.

## Verification

- `make test-terminal-queue`: race-enabled TUI/store/engine queue tests pass three repeats; the target joins required core CI.
- Unit tests cover pagination/preview sanitisation, native durable rereads, event delivery, removal conflicts, active-claim/cross-session rejection, run-bound steering, unchanged media/custom metadata, duplicate/no-idle-fallback semantics and direct-call editor preservation.
- `make test-tui-queue-commands`: six fullscreen/regular tmux PTYs at60×18,100×22,140×36 pass. Disposable SQLite fixture rows survive process reopen; commands use the native store/engine methods. Tests assert guarded removal/steering, retained metadata, no additional turns, preserved final draft/settings and unchanged idle footprint.
- `make test-tui-regular test-tui-pending-media` passes another9PTY configurations; durable attachment and scrollback behaviour remain intact.
- `make test vet bun-checks` passes; stock functional suite107passes,11existing fixture-dependent skips.

Initial fixture code used wrong store method/signature assumptions; these were corrected before passing tests. PTY testing then exposed regular-mode command responses being held behind active work. Direct print-above output fixes that without committing partial transcript content. No tests removed or timeouts increased. A read-only delegated review timed out; no independent result is claimed.

## Relative move verification

The move follow-up passes `make test-terminal-queue` under race detection with three repeats, core/vet/hooks, six queue PTYs, 48 browser queue regressions and107functional checks (11existing skips). PTYs verify before/after order, restart persistence, timestamp/metadata preservation, rejection after external reorder, fresh-snapshot retry, and no added turns or idle rows.

The shared native reorder transaction now rejects a queued row already claimed by a runner. Unit tests cover that boundary, current-visit snapshot ownership, A → B → A, invalid/no-op IDs and unchanged direct-call editor state. A two-connection file-backed contention test verifies exactly one reorder winner and one `ErrQueueConflict` for simultaneous reads of the same expected order.

Read-only review raised a possible deferred-transaction upgrade error. Inspection confirmed file-backed `store.Open` applies `_txlock=immediate`; the contention test passes. The issue was not reproduced on that supported path. General database failures remain reported without automatic retry; no timeout or test criteria were relaxed.

## Open work

This slice adds durable inspection/removal/steering and snapshot-bound relative reorder. Explicit retry controls, persistent queued text drafts and the unresolved Stop/queue policy remain separate. The native Steer checkpoint/recovery path is covered by existing engine tests; the PTY fixture verifies command admission, not external-provider behaviour. Physical terminal/emulator acceptance and the broader web/terminal parity goal remain open. Whole CI and deployment are separate gates; live data was not mutated.
