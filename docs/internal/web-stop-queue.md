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

## Deployment and Shared36 mapping

Exact source `caa83c795fa7b092bd923bbbb3bbc5309f0f3359` passed whole-product CI
[36247821576](https://github.com/rcarmo/gi/actions/runs/36247821576), including all
four release builds, then replaced `9c01b5d` on port 8090. Live PID `312567`, process
group/session `312558`; binary SHA-256
`e8795f746eff8bc9398f175d7e8c6a077be0fb81f4d48a6128b81eb563231595`.

Four blocked non-localhost HTTP sends and six read-only UI probes passed. Before
and after: 62 sessions, 51 turns, 146 messages, zero active turns; integrity/FKs
clean. Every pre-existing table, including receipts, matches except for the runtime
dispatcher lease. Auth hashes and local-only TUI WIP HEAD/diff match. The additive
hold table is empty. No live chat writes occurred. Deployment evidence is in
`/workspace/tmp/gi-web-queue-deploy-caa83c7`; DB dumps/auth hashes stay local-only.

Independent clause review accepted the browser evidence for Shared36. The one
canonical test in `reconnect.spec.mjs` now carries `@shared-36`; supportive HTTP
and native tests are not tagged a second time. `make test-shared-stop-evidence`
passed 66 reconnect cases and three catalogue/report tests (1954 assertions),
then generated six-project Shared36 pass evidence. The focused report has one
shared pass, 30 mapped-but-not-run and 11 unmapped cases. Mapping coverage is 31/42;
it is not a complete matrix run. Classic mappings remain 101/236.

The report tests reject five-project and duplicate-file evidence and do not grant
Classic original023 or Shared35 credit from Shared36-only input. A duplicate local
variable in the new report fixture was corrected before those tests passed.
CI now runs this evidence gate with `UX_PARITY_ARGS='--grep @shared-36'` and
preserves its report. Adding the entire 66-case reconnect suite after HTTP tests
in mapping CI `36249440286` exhausted the existing ten-minute job budget; its
cancelled run is not approval. The focused six-project gate passed locally in
28.8 seconds plus three report/ledger tests. No timeout or assertion changed.
The full reconnect suite remains available and passed locally as recorded above.
Frozen feature files and manifests are unchanged. This adds no physical-device, accessibility-matrix,
exact-pixel or TUI acceptance. TUI WIP `2a87a79` stays isolated.
