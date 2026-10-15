# Classic023: reconnect and run-bound Stop

The frozen clause combines SSE recovery with a Stop request for the selected
active turn. Gi's mounted `web/src/app.ts` advances connection and activity
revisions on disconnect, reloads posts or the active search plus selected
status/queue on reconnect, and issues `cancelSessionRun` with the captured
session and `activity.turn_id`. It invalidates activity and refreshes after the
request. A cancellation HTTP response is not treated as the final turn state.

`tests/ux/reconnect.spec.mjs` starts a disposable Gi server, disconnects its
real SSE stream through a proxy, admits a queued turn while offline, and checks
fresh `/messages`, `/activity`, `/queue` requests after reconnect. The tagged
`@ux-original-023` case then checks an exact active-turn Stop request, queued
turn and draft preservation, and subsequent authoritative cancellation state.
Focused run: **6/6** across Chromium/WebKit phone, tablet and desktop.

The pinned Piclaw 3.2.4 connection-lifecycle source is byte-identical in the
manifest, but its full backend/Stop UI was not exercised. The separate
`@ux-reconnect-004` clean-state version-drift contradiction remains in
`reconnect-version-review.md`; this Stop result does not settle reload policy.
No live Gi writes or production code changed.
