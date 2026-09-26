# Web Stop preserves pending work

Web Stop cancels the captured active turn and preserves queued work behind an
explicit **Resume queue** action. Generic engine/TUI cancellation and normal
completion retain their existing handoff behaviour.

## Durable boundary

`POST /api/sessions/<id>/activity` still requires the captured `turn_id`. The
engine revalidates local ownership under the runner lock; the store transaction
checks the matching active claim and token, marks the turn cancelling and records
`web_queue_holds(session_id, stop_turn_id, created_at)` if runnable queued or
steering work exists. Queue rows, ordering, media and metadata are unchanged.
An empty Stop does not introduce a hold. Work arriving only after that empty Stop
is a new admission and follows the existing rules.

Active claims, steering staging, recovery handoff and manual compaction respect
the durable fence. Thus cleanup or a restarted engine cannot launch preserved
work behind the user's Stop. An already-held session queues later prompts until
Resume. The additive hold row is deleted with its session. Mixed old/new writers
against one database are unsupported.

Returned, unconsumed Steer rows remain queued but non-runnable. They do not create
or trap a hold. Bound steering restored during cancellation remains visible and
is not automatically resent.

## Resume and uncertainty

Activity GET includes `queue_hold_turn_id`. The app shows Resume queue only when
a hold is present, and enables it only for fresh, connected, idle activity. The
button captures session and stop ID, is single-flight, and does not alter draft
text, attachments or another session's controls.

`POST /api/sessions/<id>/resume-queue` accepts exactly one `stop_turn_id`. It rejects
stale, foreign, repeated or still-active requests with 409. Generic `/continue`
cannot bypass a web hold. No resume POST is automatically replayed; uncertain
responses retain a warning and refresh authoritative state.

Resume keeps the hold while staging/claiming work. Only the exact captured hold
permits a resumed claim; generic claims stay fenced. The hold is retired after
the new active claim, turn status and session state are durable, before starting
the runner. Launch or retirement failure rolls back the claim and retains the
hold. If no runnable/steering work remains, exact Resume clears the empty hold.
A crash before retirement leaves the hold available after recovery. A crash after
retirement uses the normal durable active-turn recovery path. This is an explicit
resume, not exactly-once inference or automatic replay of uncertain requests.

## Verification

- `make test-web-queue-hold`: race ×3 across store, engine and HTTP tests. Covers
  exact ownership, stale/foreign/active resume, duplicate resume, no-work Stop,
  unchanged FIFO/media/metadata, rollback, reopen, crash-before-cleanup, generic
  claim/continue fences, two-store concurrent resume claims, queued and steering
  launch failures, retirement failure, and returned-Steer non-runnable rows.
- `make test-web-basic-controls`: 36 cases across Chromium/WebKit at three sizes,
  including Stop → preserved queue/draft/media → reload → explicit resume, plus
  lost Resume acknowledgement without replay.
- `make test-ux-reconnect`: 66 passes. The captured-Stop test now asserts exact
  unchanged queue before explicit resume, preserves foreign-session activity,
  then replays old native terminal frames without clearing the resumed run.
- `make test-ux`: 139 passes, 11 existing skips.
- `make test vet bun-checks`: passed. Supplied components/ui/panes unchanged.

Review found and corrected two blockers: deleting the hold before durable launch,
and treating returned Steer items as runnable hold blockers. Final scoped review
accepted the corrections. Initial native test harness mistakes and superseded
failing logs remain in `/workspace/tmp/gi-web-queue-hold`; no timeout was increased.

Implementation and local acceptance do not grant automatic Shared36 mapping,
physical-device, accessibility-matrix or exact-pixel credit. Whole-product CI and
exact-source deployment remain separate gates. TUI WIP `2a87a79` stays isolated.
