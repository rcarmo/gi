# ADR-0024: Fence reconnect refreshes and report UI version drift

## Status

Accepted — 2026-09-22. Classic reconnect `002` and `004` mapped; search-view protection and initial-refresh deduplication remain open.

## Refresh ownership

Timeline HTTP replies now require the current session generation, connection revision and timeline request generation. Disconnect/reconnect, session selection and an accepted new-post event invalidate earlier requests. Full reloads and older-page requests use the same generation guard, so an earlier response cannot replace or merge into a newer view. Existing server pagination semantics are unchanged; this slice does not establish full pagination acceptance.

Timeline errors use those same guards. Status/queue/context refresh errors also check connection and activity revision, preventing a failed pre-disconnect request from overwriting the healthy reconnect state. The normal reconnect path reloads native activity, queue, model/context capability and timeline. User draft/media state is not cleared.

The independent review identified the original timeline connection-scope gap and the stale error path before implementation. Tests hold real native HTTP responses while severing SSE sockets; offline emulation and fabricated SSE payloads are not used.

## Asset version warning

The server already advertises `app_asset_version` in its native connected envelope and stamps the loaded app script URL with the same value. The SSE adapter now forwards that envelope through its existing source/selection guard. The app compares it with the loaded script version, remembers observed differences and displays one persistent `New UI available` notice with a manual reload instruction. It never calls reload or navigates automatically, including when the draft/editor are clean. Repeated announcements for the same version are ignored.

The existing server version is generated per process start, not from a content hash. A server restart can therefore warn even if the asset bytes are unchanged. No version-generation policy is changed here. The warning is page-global and remains until manual reload; it is not an origin-session error or a request to discard unsaved work.

## Evidence

`make test-ux-reconnect` builds the existing local-provider Go fixture. Each browser case owns a loopback instance and temporary persistent state directory. A byte-for-byte SSE proxy closes real sockets; the version test stops and restarts the real server with the same database. No paid provider, SQL-seeded response state, forced click or retry is used.

- `reconnect-002`: native active run and queue before loss; offline completion, queue replacement and new run; reconnect refreshes activity, queue, context and timeline. An old held timeline response cannot remove the newer post. Stop targets the newly observed run. Draft/media survive.
- `reconnect-004`: actual server restart changes the advertised version; one manual-reload warning appears. Neither dirty nor clean state causes automatic navigation. Repeated reconnects retain one warning. Manual reload loads the new version and clears it.
- Gi-only regression: a held pre-disconnect activity request fails after reconnect; it cannot replace healthy state with an error or disturb the draft.
- Helper checks cover request-generation invalidation, version baseline/deduplication and superseded SSE sources.

Validation: 18/18 reconnect executions across Chromium/WebKit phone/tablet/desktop, 312/312 combined browser executions, 70/70 functional tests, 26/26 helper tests, full Go tests/vet and hook checks. Coverage: Classic 28/236 passing (208 unmapped); shared 2/42 passing (40 unmapped).

## Terminal disposition and remaining work

The local Go TUI has neither browser assets nor an SSE transport. It needs no permanent version banner, reconnect panel or additional idle row. Its existing session-generation subscription guards and explicit process-reopen tests remain separate evidence. Any future remote terminal transport should use bounded transient notices and the same origin/run guards.

Search/hashtag reconnect protection (`003`) is unimplemented because the corresponding view is not wired. Initial activation refresh deduplication (`005`), full pagination, crash recovery during active provider work and cross-process/multi-tab draft reconciliation remain open. A controlled idle server restart is not evidence of active-turn crash recovery.
