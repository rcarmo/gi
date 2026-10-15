# Shared27, 29, 30: queue identity and matching-run Steer

Gi's mounted `web/src/app.ts` scopes queue mutations to the selected session,
uses durable queued-turn IDs and revision guards, and refreshes after failures.
Steer requires a connected matching active turn. The Classic idle-Steer
backend policy differs; `classic-queue-review.md` keeps that conflict open.

| ID | Tagged native browser assertion | Limit |
|---|---|---|
| `27` | Two composer follow-ups retain FIFO IDs, text, media and references once, without affecting another session. | `queue.spec.mjs`: six projects passed. |
| `29` | Adjacent reorder affects the captured group; a queued row consumed by the runner before a held remove returns 409, with authoritative reconciliation and preserved unrelated queue/draft. | `queue.spec.mjs`: six projects passed. |
| `30` | Disconnected and idle Steer are disabled; stale/foreign run requests fail, retry and duplicate prevention hold, and successful delivery consumes the exact queue ID against the active run. | `queue-steer.spec.mjs`: six projects passed after a test-only locator correction. |

Focused 27/29 run passed **12/12**. The first Shared30 run failed all six
projects at a stale `Model picker` menu-item selector, before reaching the
remaining Steer assertions. The mounted picker exposes the `Models` listbox
and an `option`; changing only that test locator yielded **6/6** on an
unchanged Gi runtime. It does not prove deployed Gi, current Piclaw backend,
physical input or Classic019 idle-Steer parity. No production code or frozen
Gherkin changed.
