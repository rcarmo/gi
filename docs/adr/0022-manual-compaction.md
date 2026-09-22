# ADR-0022: Explicit manual compaction without a chat submission

## Status

Accepted — 2026-09-22. Browser meter action implemented; terminal command wiring remains separate.

## Admission and execution

`GET /api/sessions/{id}/compaction` returns availability, a reason and a context conflict token. Availability requires an idle session with no pending queue/steering work, at least two projected messages and a prefix eligible for durable text compaction. The token hashes checkpoint version and all current provider-visible message fingerprints. It is not a credential.

`POST` accepts only `{ "token": "..." }`. The engine validates capability/token under its session runner lock. Store admission repeats idle/token validation and atomically writes an empty-prompt running maintenance turn, active claim, submission event and session state. Concurrent activations admit at most one operation. Invalid bodies return 400, absent sessions 404, and busy/stale requests 409.

Manual execution forces the existing native compaction path past automatic enable/threshold checks. It rechecks the context token before starting and uses the same hook, cancellation, occurrence and durable checkpoint transactions as automatic compaction. It never invokes the inference provider or persists `/compact` as user text. The existing default summary is a transcript-excerpt heuristic; registered hooks may supply a summary. Token reduction and summary quality are not guaranteed.

A manual operation cannot accept prompt steering or queued submissions while active. Internal operation metadata is stripped from ordinary submissions. Queued maintenance markers are excluded from automatic scheduling; recovery aborts interrupted manual operations instead of replaying them. The admission transaction closes the previously identified create-before-claim gap. Full process-crash fault injection remains untested.

Cancellation preserves the previous checkpoint and uses the existing run-bound Stop path. A history/version conflict before completion fails without a partial checkpoint. Latest measured provider usage remains unchanged until a later real inference request supplies new measurement.

## Browser control

The app attaches a callback to the existing context meter only while its native capability snapshot is fresh, idle and available. The supplied component remains unchanged. The host intercepts the meter before its legacy `/compact` submission handler, so unsent text, files and references are not cleared or uploaded. Available tooltips include `Compact context`; unavailable controls are disabled and say `Context usage`. Active compaction retains its elapsed/status branch.

Admission errors remain visible for the origin selection and never trigger automatic resubmission. Pending clicks are blocked. Session changes invalidate capability and pending-operation ownership. Completion refreshes capability/activity and preserves the draft. The existing usage tests now check the appropriate capability suffix; their token/percentage expectations are unchanged.

## Evidence

- Frozen `context-003` across all six projects: disabled unavailable meter, available native callback, success, repeated activation rejected, no user prompt or provider-usage update, timeline/draft/media/reload preservation.
- Additional six-project regression: failed delivery retains draft/media, stale token produces no turn, cancellation preserves history/checkpoint and never submits the draft.
- Native tests: concurrent admission, atomic rollback at submission event, stale/busy/foreign/malformed rejection, no inference call, cancellation, history conflict, and interrupted-operation recovery without replay.
- Validation: 66/66 compaction browser executions; 294/294 combined browser executions; 70/70 functional; 24/24 helpers; Go tests/vet, hook checks and targeted races repeated three times.

Coverage: Classic 26/236 passing, 210 unmapped; shared 2/42 passing, 40 unmapped. No other feature receives new credit.

## Terminal adaptation and remaining work

Use the same native availability/token/admission method for an explicit `/compact` command or on-demand action. Show progress in the existing transient status region and reuse cancellation keys; preserve editor/cursor and avoid any new idle rows or permanent control. Test success, failure, cancellation, busy rejection, restart and resize at 60×18, 100×22 and 140×36 before live terminal acceptance. This slice does not replace the current informational terminal `/compact` output.

Broader socket-reconnect and crash acceptance, general hook/tool/multimodal checkpointing, large-history performance and summary fidelity remain open. Browser screenshots delivered earlier predate this change.
