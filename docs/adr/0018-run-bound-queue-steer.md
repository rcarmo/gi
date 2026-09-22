# ADR-0018: Bind queue Steer to an active run

## Status

Accepted — 2026-09-22

## Contract

Gi implements shared `@shared-30`, “Steer only a matching active run”. Idle or unknown activity disables Steer; the server requires the expected active turn ID. Classic `@ux-original-019` permits immediate sending when a stream ends and stays unmapped. Frozen feature files are unchanged.

`POST /api/sessions/{session}/queue/{queuedTurn}/steer` accepts only `{ "active_turn_id": "..." }`. Invalid bodies return 400; missing sessions return 404; stale, foreign, consumed, cancelling or idle targets return 409. The endpoint never creates a fresh user turn as a fallback.

## Admission and delivery

`Store.SteerQueuedTurn` conditionally changes the original unclaimed queue row to `steering` and inserts a steering entry with a dedicated `source_queue_id` column in one transaction. It requires a running turn with the matching session claim and room in the steering queue. Rollback leaves the original row queued; a second request cannot admit the same row twice. The original prompt, metadata, media and FIFO position remain available for recovery.

Only a checkpoint for that turn can claim the entry. Generic steering continuations and other runs cannot dequeue it. Claiming alone does not consume the source. `PersistBoundSteering` atomically persists the user message, acknowledges the steering entry and marks the original turn `cancelled` with phase `steered`. The message's turn identity comes from the consuming checkpoint. The existing provider-safe media projection supplies image blocks and attachment placeholders.

If claim release or restart recovery finds an unconsumed bound entry, it restores the original row as `queued` with phase `steer_returned`. This row is held: automatic scheduling and generic steering continuation exclude it. A new explicit prompt can still start. Return-to-editor, cancellation or an explicit Steer into a new active run are available. A visible notice explains the held row. Retrying after acknowledgement cannot inject the same entry again.

Queue Steer is separate from ordinary prompt steering. Existing implicit steering and its continuation behaviour retain their contracts. A continuation now receives a FIFO position like other queued turns.

## Browser ownership

The queue response includes the native active turn ID. The app captures that ID, origin session and selection generation when activating Steer. Controls require connected, active state and a durable queue row; in-flight queue mutations disable overlapping activation. The request handler checks these conditions again. Failures retain the row or reconcile consumed state without restoring unrelated drafts. Replies from an old selection cannot change the current session's queue, errors or draft.

The supplied stack has no disabled-Steer prop. `RunBoundQueueStack` in `web/src/app.ts` applies the native `disabled` property to its existing button in a layout effect. Components, UI utilities and panes are unchanged. The host callback also guards disabled actions. Appearance and normal keyboard activation are retained.

## Evidence

- `tests/ux/queue-steer.spec.mjs`: shared-30 across Chromium/WebKit at phone/tablet/desktop sizes. Visible controls, failed transport, repeated keyboard activation, stale/foreign HTTP requests, real SSE socket closure, idle recovery/reload, explicit retry into a new run, native persisted messages and provider input echoed by a local fixture.
- Additional six-project regression: a delayed request reaches the server after its target and queued item completed; a session switch retains both origin and target drafts without leaking failure feedback.
- Store/API tests: two concurrent admissions, rollback after source update, full steering queue, wrong claim token, run-scoped dequeue, persistence failure, at-most-once message acknowledgement, held recovery and rejection of automatic continuation.
- Engine test: native checkpoint retains image bytes, rejects repeated injection and uses the consuming turn ID despite conflicting input metadata.

`make test-ux-steer` runs an isolated Go server with a deterministic local OpenAI-compatible provider. Production HTTP/SSE, SQLite and inference checkpoints are used. Credentials and workspace live in a temporary directory; no paid provider or synthetic agent events are used. The regular matrix retains its native shell fixture. Both result files must be passed explicitly to the parity reporter; it never silently reuses another run.

## Limits

Acknowledgement records acceptance into the run context, not exactly-once processing by an external provider. A process can fail after persisting the user message but before the next provider request. The content remains in history and is not automatically resent. Queue Steer does not offer an execution-completion receipt. Native image projection has Go coverage; this browser slice uses text steering. Multi-process/network crash testing and whole-application parity are incomplete.

## Terminal adaptation

Use the same engine method from an on-demand queue selector, capped at six entries like Alt-S. Keep the editor visible, use existing muted/accent selection and arrow/Enter/Escape interaction, and show Steer only as an available action for a captured active run. No persistent queue panel, top bar or idle status rows are needed. Escape returns to the unchanged editor/cursor; async completion belongs to the origin session/generation.

A failed or unconsumed return uses the existing transient notice and held queue row. Never fall back to Enter/send. Verify duplicate activation, idle/cancelling/unknown state, session switches, resize and zero idle-row movement at 60×18, 100×22 and 140×36. These terminal controls are designed but not implemented in this slice.
