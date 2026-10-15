## Resume removed

Gi's Resume queue control and Stop-created queue holds have been removed. They were invented Gi behaviour, not part of the installed Piclaw 3.2.4 UX or workflow. Earlier tests validating those holds were not evidence of Piclaw parity.

Stop cancels the addressed active run. Normal cleanup can then process queued work without another user action. The removal includes the banner, button, client API, HTTP endpoint, activity field, engine/store resume methods and all hold checks. Database opening records a one-time handoff for sessions with an obsolete `web_queue_holds` row and drops that table; it preserves queued turns, order, metadata and messages. Engine startup uses normal continuation for those sessions and clears the handoff marker only after successful admission. Recovery tests cover both an already-idle held queue and a crash with a cancelling claim, followed by automatic processing of the queued successor.

## Installed oracle

`make test-piclaw-stop-queue` executes two independent probes:

* The shipped browser bundle sends `/abort` with `mode: steer`, retains the editor draft and shows no Resume control. All six Chromium/WebKit phone/tablet/desktop journeys passed. The bundle source map is pinned by its recorded provenance hash.
* Installed abort/control and finalization functions run with an in-memory database and disposable collaborators. Abort does not clear the pending item. Terminal finalization stores the deferred follow-up and calls internal `resumeChat` automatically. If storage fails, the item remains queued and that dispatch does not happen. Internal `resumeChat` is not a user-facing Resume action.

The first UI probe incorrectly assumed the operator run-abort endpoint and a pending disabled button. Source inspection corrected it to the actual composer `/abort` command; the wire request uses `mode`, not `followup_mode`. Those failed attempts are retained. The method probe isolates abort and successful terminal persistence/finalization; it does not simulate every provider cancellation outcome.

## Native verification

* Store/turn/web race gates passed three times, including stale Stop, FIFO continuation, migration and crash recovery.
* HTTP controls: 36/36, including two queued items after Stop, preserved drafts/media across reload, and lost Stop acknowledgement without replay.
* Reconnect: 72/72 after replacing the obsolete held-queue assertion with automatic completion; the 12 Stop-specific cases also passed after backend deletion. The earlier six failures are retained.
* Idle and ended-run Steer: 6/6 each after removing hold checks.
* Go tests/vet/build/hooks and functional gate: 144 passed, 11 skipped. Support inventory: 210 tests, 7,795 assertions.

Obsolete Resume-specific tests were replaced with oracle-backed Stop/continuation and migration tests. No timeout or SSE exception was added. The separate Escape support test initially used an incomplete adapter chain and failed; the corrected composition passed. The focused independent review timed out; there is no independent approval for this removal.

The previous `@shared-36` mapping relied on the invented hold/Resume workflow. Its tag was withdrawn from the reconnect test and catalogue; Shared36 is unmapped pending a fresh clause-by-clause Piclaw journey. The matrix now has 29 Shared mapped and 13 unmapped cases. [Current audit ledger](web-ux-oracle-matrix-2026-09-28.md). No deployment, restart or live-chat write occurred. Physical-device behaviour, the strict WebKit reload failure, unconsumed steering restoration and full queue handoff remain outside this verification. The whole-web audit and requested branch/worktree consolidation are unfinished.
