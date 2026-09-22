# ADR-0026: Give initial activation and SSE readiness one refresh owner

## Status

Accepted — 2026-09-22. Frozen reconnect-005 mapped. No server or terminal behaviour changes.

## Readiness and ownership

The app now treats native SSE readiness as a prerequisite for authoritative timeline/search/activity/queue/model/context reads. It does not reuse pre-subscription HTTP snapshots, which could miss changes before the stream became active.

`createActivationRefreshGate` tracks selection generation, connected state and whether the current readiness epoch has claimed its initial refresh. Chat activation and the first connected callback can occur in either order; only one claims that refresh. The normal activation effect still refreshes session-list metadata. A non-connected state opens a new epoch, so reconnect always fetches current native state. A→B→A uses distinct selection generations.

Periodic refresh, explicit query changes, mutation completion and native lifecycle invalidation keep their existing refresh paths. The gate does not cache successful responses or promise exactly-once event delivery. Failed initial HTTP reads remain retryable on reconnect or the existing periodic refresh; no automatic prompt resend is introduced.

The app wraps event/status/wake callbacks in the selection captured by the rendering frame. This closes the interval between synchronous selection change and SSE-hook rerender: an old callback cannot bind itself to the new selection's readiness epoch. Existing source identity, hook selection and HTTP response guards remain in place.

## Evidence

`reconnect.spec.mjs` can hold the real native SSE connection before readiness. The new case verifies that the rendered shell makes no authoritative reads while held, then issues one initial messages/activity/queue/model/compaction set. It switches A→B→A and severs/reconnects the actual transport to verify fresh activation and reconnect. A separate case fails the initial activity request, then confirms real reconnect clears the error and preserves the draft.

Helper tests cover both activation/readiness orders, duplicate connected notifications, disconnect epochs and selection generations. The exact request-count assertions apply only to the controlled idle fixture; the frozen contract makes no fixed timing or exactly-once event-delivery promise.

Validation: 48/48 reconnect/search executions, 342/342 combined browser executions, 70/70 functional tests, 28/28 helpers, full Go tests/vet and hook checks. Classic 30/236 passing (206 unmapped); shared 2/42 passing (40 unmapped). All five reconnect feature IDs are mapped, each against its tested behaviour. Search reconnect uses the search alternative; hashtag navigation has no new credit.

A focused review identified the pre-render callback ownership concern, which was addressed before the six-project run. No supplied component/UI/pane files or frozen feature assertions changed.

## Terminal disposition and limits

The local Go TUI does not have browser activation/SSE readiness. Its existing generation-owned subscription and reopen semantics remain separate; no extra idle row, banner, selector or other terminal UI is added.

This completes the frozen reconnect family, not full web/TUI parity. Active-turn crash recovery, initial offline timeout UX, hashtag navigation, native timeline pagination, broader feature families and terminal search/queue/media adaptations remain open.
