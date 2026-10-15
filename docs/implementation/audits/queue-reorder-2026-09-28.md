# Queue reorder and removal against Piclaw 3.2.4

Installed Piclaw 3.2.4 and Gi both optimistically move and remove queued follow-ups, then reconcile failed mutations with queue state. The installed UI and Gi were tested in Chromium and WebKit at phone, tablet and desktop sizes. These are separate, bounded runs: Piclaw used shipped browser assets with mocked queue APIs; Gi used the local native test server.

## Installed browser

`make test-piclaw-queue-reorder` runs `tests/ux/oracle/piclaw-queue-reorder-probe.mjs`. Six cases passed. The fixture supplied three rows and an unsent draft. Moving the second row up immediately changed the order and posted `{from_index:1,to_index:0,chat_jid:'web:default'}` to `/agent/queue-reorder`. Removal posted `{row_id:31,chat_jid:'web:default'}` to `/agent/queue-remove`. The draft stayed intact.

The probe then made each mutation return `409`. Reorder and removal refreshed the queue from the fixture and restored its two remaining rows. Each case recorded three queue reads and four mutation requests. The failure paths logged their errors; the browser probe did not require a toast for reorder. Results: `test-results/ux-oracle/queue-reorder/evidence.json` and `/workspace/tmp/gi-queue-reorder-oracle-final.log`.

A first failure probe expected an immediate queue GET and observed only the optimistic order (`/workspace/tmp/gi-queue-reorder-oracle-attempt6.log`). Installed `app-refresh-coordination.ts` coalesces foreground refreshes within 250 ms. Once the fixture waited 400 ms after the preceding mutation, all six cases fetched the queue and reconciled. The premature observation did not show a product defect.

## Gi and limits

`make test-ux-parity UX_PARITY_ARGS='tests/ux/queue.spec.mjs --grep "@ux-original-018"'` passed 6/6 (`/workspace/tmp/gi-queue-reorder-native.log`). That journey verifies optimistic adjacent movement, a real rejected reorder after another queued turn is appended, rejected removal, draft retention, reload, and eventual native order. Gi uses durable turn IDs and an expected/order snapshot in its API, while installed Piclaw uses positional indices; both clients reconcile the tested outcomes. This does not grant a new mapping: `@ux-original-018` was already mapped by the native suite.

`make check ux-parity-inventory` passed 144 functional tests with 11 skips and 214 support tests with 7,977 assertions (`/workspace/tmp/gi-queue-reorder-check-final.log`). The first standard check stopped at 143 passes on a read-aloud test that remained at “Loading Gi…” until its 30-second locator timeout (`/workspace/tmp/gi-queue-reorder-check.log`). An unchanged rerun was interrupted after case 130/155 (`/workspace/tmp/gi-queue-reorder-functional-rerun.log`); the subsequent complete check passed.

The installed fixture did not execute a provider or test backend persistence, concurrent tabs, session switching, or all queue handoffs. The local Gi run did not deploy a binary. Piclaw's strict WebKit reload/SSE diagnostic remains a separate failing gate; this probe performs no reload.
