# WebKit reload oracle control

The strict installed-Piclaw lifecycle oracle still fails on reload with:

- `Load request cancelled` for the old `/sse/stream` and `/agent/push/presence` requests.
- A page error ending in `/agent/push/presence due to access control checks.`

Serving the fixture over native HTTP without Playwright routing produced the same failure. That experiment was reverted; the strict fixture and its error policy are unchanged.

`make diagnose-webkit-unload` now runs a minimal control with no Piclaw/Gi assets, route interception, credentials or application backend. A same-origin page opens an EventSource, posts presence on visibility/pageshow, and uses sendBeacon on pagehide/beforeunload. It reloads and waits for the replacement EventSource.

Observed control result:

- Chromium: no page/network errors; replacement stream connected.
- WebKit: the same access-control page error and cancelled old-stream/presence requests; the native server received presence requests and the replacement stream connected.

This isolates the symptom to the browser/navigation interaction, not Piclaw-specific code or the route-only fixture. It does not establish that all production WebKit reloads are harmless, nor does it make the strict Piclaw lifecycle gate pass. No error waiver, product workaround or parity mapping was added.

The diagnostic records observations rather than enforcing a expected browser bug, so it remains usable when the browser version changes. Evidence: `test-results/ux-oracle/webkit-unload-control.json`, `/workspace/tmp/gi-webkit-reload-{baseline,native-presence,native}.log`, and `/workspace/tmp/gi-webkit-control.log`.

No production traffic, deployment or restart.
