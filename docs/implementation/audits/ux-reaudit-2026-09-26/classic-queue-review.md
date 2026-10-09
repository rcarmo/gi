# Classic canonical queue 016, 018–019

Gi's mounted `web/src/app.ts` obtains queue state for the selected session,
uses revision guards around responses, and routes reorder, remove, return and
Steer through native queue APIs. The copied Classic composer and Gi backend
have different steering policy: Gi requires a connected, matching active turn.
The frozen `@ux-original-019` also permits the backend to send immediately
when the stream ends. Shared30's safer gate cannot earn that Classic clause.

| ID | Focused native evidence | Boundary |
|---|---|---|
| `016` | `queue.spec.mjs` holds an active turn, admits queued items, checks returned rows and client-token reconciliation across HTTP acknowledgement and server refresh. | Six Chromium/WebKit viewport projects passed with disposable Gi. |
| `018` | The same spec checks optimistic order and removal, persistence after reload, HTTP 409 on stale order, failure rollback and successful retry. | Six projects passed; current Piclaw queue runtime unprobed. |
| `019` | `queue-steer.spec.mjs` covers Shared30: an active-turn ID, stale/foreign 409, retry, duplicate prevention and unconsumed-row state. Gi disables Steer without a matching active turn; Classic allows a backend decision after stream end. | **Policy conflict**, not full Classic coverage. No tagged `@ux-original-019` journey. |

The focused Classic016/018 run passed **12/12**. The local `test-ux-steer`
wrapper was attempted twice with a quoted grep expression and reported `No
tests found`; those attempts have no pass result. The corrected
`make test-ux-parity` command produced the 12 passes above. This review does
not reclassify the separate queued-return `017` oracle conflict, test live Gi,
or change frozen requirements.
