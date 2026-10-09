# Classic SSE reconnect 001–005: installed-source correction

The frozen `tests/ux/features/classic/compose/sse-reconnection.feature` requires
a manual-reload warning for a first changed UI asset version, **even with clean
editors and composer**. The installed Piclaw 3.2.4 Classic source agrees. The
shipped `/opt/piclaw/current/app/runtime/web/static/classic/dist/app.bundle.js.map`
SHA-256 is `c53092cfb75415f30b4c6b16b758cab68fed9bbc546953e5916a91b382acb32b`,
pinned by `tests/ux/oracle/piclaw-3.2.4-reference.json`. Its embedded
`src/ui/app-connection-lifecycle.ts` records the changed version and explicitly
disables auto-reload because it caused cascading reloads. It shows `New UI
available` with manual reload instructions; repeats for the same version return
early. The shipped `src/ui/app-shell-state.ts` recognises the Classic bundle URL
`/static/classic/dist/app.bundle.js?v=990f0c49a932`.

An **unshipped, divergent** checkout at `/workspace/projects/piclaw` and the
copy at `web/src/ui/app-connection-lifecycle.ts` still schedule clean-state
reload after 350 ms. The two passing tests in
`tests/ux/support/version-drift-lifecycle.test.ts` exercise that copied helper;
they cannot overturn the installed release or credit Gi's mounted entry point.
The earlier source-conflict finding was a provenance error. A first disposable
probe produced no warning because the pin-sync GET fixture returned an invalid
shape and its error toast displaced the warning. After that fixture was given a
valid scope, revision and pin arrays, `tests/ux/oracle/piclaw-version-drift-probe.mjs`
confirmed SSE delivery, one manual warning on version drift, one warning after a
repeat and zero automatic navigations in Chromium and WebKit. It checked the
shipped asset and source-map hash before each run. These are shipped Classic
browser-asset observations against disposable API/SSE fixtures. No current
Piclaw backend version-change or deployed Gi run was made.

Gi's mounted browser entry point is `web/src/app.ts`. Its selected-session SSE
`connected` handler uses `createAssetVersionGuard(loadedAssetVersion(document))`
to set `newUIVersion`; the rendered `gi-version-warning` says `Reload manually`.
It does not call the auto-reload helper. The separate composition chain
`app-main-orchestration-composition.ts` → `app-main-lifecycle-composition.ts` →
`app-agent-status-lifecycle.ts` → `app-sse-events.ts` does pass `serverVersionContext`
to `handleUiVersionDriftEvent`, and that handler returns before reconnection
resync on drift. No call from the mounted `web/src/app.ts` into that
composition chain was found in the Gi entry point; its presence and copied
oracle logic do **not** prove that clean-state auto-reload happens in Gi.

| Frozen ID | Gi assertion / code | Boundary |
|---|---|---|
| `001` | `tests/ux/queue.spec.mjs` severs SSE, checks that transient agent display clears while the user's draft survives; `web/src/app.ts` disconnect handler clears previews, pending/run state. | Focused real SSE-proxy/browser run: 6/6; current Piclaw UI not run. |
| `002` | `tests/ux/reconnect.spec.mjs` holds a stale response across reconnect, checks authoritative timeline, queue and activity; `activationRefresh` and `refreshAfterConnection` reload selected state. | Focused reconnect tags `002`–`005`: 30/30 across six projects; Gi fixture, not deployed acceptance. |
| `003` | Tagged search test checks no general messages refresh over search, but fresh activity/queue/model data and retained draft/file; search-aware refresh in `web/src/app.ts`. | Same focused run; no Piclaw search-reconnect browser probe. |
| `004` | Existing dirty-draft test checks manual warning, no navigation, unchanged asset URL, then manual reload. **Clean-composer** test waits beyond 350 ms, checks no navigation on two notices, and then manually reloads. `createAssetVersionGuard` checks deduplication. | Clean Gi case 6/6; focused tags 30/30. Installed Classic source agrees with the frozen clause; shipped-asset Chromium/WebKit probes confirm delivered warning, deduplication and zero auto-navigation. Current Piclaw backend, live Gi and physical acceptance remain unverified. The divergent copied helper is unmounted. |
| `005` | Tagged test checks one initial refresh per selected session and a new refresh after actual disconnect; `createActivationRefreshGate` unit test covers readiness ordering. | Same focused run; not an exactly-once transport guarantee. |

A broader unfiltered reconnect command reached 63/72 before its external
260-second timeout; it has **no pass result**. The focused runs do not cover
all additional Gi reconnect journeys. No production code, frozen Gherkin,
installed Piclaw or live Gi state was changed. The shipped-asset probe uses a
mocked backend and does not establish real-server version-change handling. The
copied helper's auto-reload path must not be presented as installed 3.2.4
policy.
