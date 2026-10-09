## Steer after the observed run ends

> Correction: Stop/Resume holds described below were Gi-only behaviour and have been removed. The current contract and results are in [Stop/Resume removal](stop-resume-removal-2026-09-28.md). Earlier hold tests establish only historical Gi behaviour.

A queued Steer request can now launch its selected row after the observed run has ended and released its claim. The request keeps the original `active_turn_id`; it does not retry as an unrestricted idle action.

The installed Piclaw 3.2.4 `WebAgentControlPlaneService.handleAgentQueueSteer` falls back to `processChat` if `queueStreamingMessage` reports that streaming ended. The existing independent method probe exercises this `ended-before-queue` path with in-memory collaborators. `make test-pi-piclaw-queue-oracle` passed all eight method cases. Replacing only Gi's handler with its pre-fix version made the new completion and retry regressions fail three times with `queue changed; refresh and retry`.

## Admission and rollback

`session_last_run` records the latest claim, including claims created through other admission paths. Insert/update triggers on `session_active_turns` update it in the claim transaction. It survives release and database reopening; a failed claim cannot overwrite it. The migration seeds currently active claims and leaves unknown pre-upgrade idle history ineligible. Timestamps do not determine run order.

The fallback's claim statement checks all of the following atomically:

* The observed turn is the latest claimed run in the same session and has a terminal status. No active claim may exist.
* No web Stop hold exists. Unlike explicit idle Steer, the fallback cannot bypass a hold.
* The selected turn is queued, unclaimed, belongs to the session, and is not manual compaction or returned steering. A prior `steering_queue` occurrence also excludes it.

The normal selected-row launch path retains turn identity, metadata, media and prompt persistence. It writes the prompt once before starting the provider. Failed preparation or prompt storage restores the queued row and previous latest-run marker only while the selected claim still owns that marker. Duplicate requests cannot create a second turn or user message.

## Verification

| Gate | Result |
| --- | --- |
| Installed Pi/Piclaw method oracle | 8 cases passed, including ended-before-queue |
| Pre-fix Gi handler, new regression | Completion and retry cases failed on all three runs, as expected |
| `make test-idle-queue-steer` | Store, turn and web race tests passed three times |
| `make test-ux-ended-steer` | 6/6 Chromium/WebKit phone/tablet/desktop cases |
| Existing active Steer browser suite | 18/18 |
| Existing idle Steer browser suite | 6/6 |
| `make check ux-parity-inventory` | Go tests/vet/build checks; 144 functional passed, 11 skipped; 209 support tests, 7,786 assertions |

Native tests cover replacement runs (including replacements which have also ended), Stop holds, cancelled/claimed/foreign/returned rows, manual compaction, unknown history, simultaneous claims, migration/reopen, rollback ownership, storage failures and explicit retry. The browser fixture seeds a disposable observed claim, holds the real Steer POST, releases that claim, then exercises the production HTTP/launch/provider paths. It verifies the original request body, one selected turn and user message, provider receipt of the instruction, duplicate rejection, draft preservation and reload.

The initial browser command found no tests because its new environment flag was missing from the Playwright selector. The corrected selector passed all six cases; the failed log is retained. No test timeout or browser-error assertion was relaxed. The independent design-review delegate timed out and supplied no approval.

## Remaining gap

If the observed claim is still present during cleanup, this fallback rejects. If automatic queue handoff has already claimed or completed the selected item, it also rejects without replaying it. Reserving a row while cleanup owns the run needs a separate handoff design; this patch does not implement that reservation or establish complete end-of-turn parity.

The browser fixture has a controlled claim transition; it does not prove provider/tool timing, physical-device behaviour or pixel equality. No production data, running executable or live chat was changed. No deployment or restart occurred. Parity mappings are unchanged: Classic 97 mapped/144 unmapped and Shared 30 mapped/12 unmapped. Strict WebKit reload, Pi queue restoration and the rest of the web audit are still open.
