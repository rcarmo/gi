# ADR-0012: Durable queued follow-ups and conflict-safe controls

## Status

Accepted — 2026-09-22

## Context

Explicit `queue` prompts became steering input when a session had an active turn. The web reorder adapter returned null; removal swallowed failures and used a generic turn-cancel endpoint. A stale queued-row click could therefore target a turn that had begun running.

Classic `@ux-original-018` requires optimistic reorder/removal and authoritative refresh after failure. Classic return-to-editor and Steer differ from the stronger shared contract. Those criteria stay separately unmapped.

## Decision

Preserve explicit `queue` intent as a durable queued turn while another turn is active. Ordinary prompt and steer admission retain their existing behaviour. Queueing keeps the active turn/session state intact and appends the existing admission events. The runner consumes queued turns after the active turn ends.

Add `turns.queue_position` with default zero for existing rows. New turns created through `CreateTurnWithStatus` use the session's maximum position plus one. Queue reads and runner selection order by position, then original creation timestamp and ID. Existing queues preserve FIFO ties. Reorder updates only positions; admission timestamps and append-only event rows are unchanged. Steering-continuation staging still occurs only when no queued turns exist.

Native API:

- `GET /api/sessions/{id}/queue` returns authoritative queued turn records.
- `PATCH /api/sessions/{id}/queue` accepts `{expected: [ids], order: [ids]}`. In one write transaction, require the exact current queue order and a complete permutation. Stale, duplicate, foreign, missing or consumed IDs return 409.
- `DELETE /api/sessions/{id}/queue/{turnID}` cancels only queued, unclaimed turns. Wrong-session IDs return 404; started/claimed turns return 409. Repeating an acknowledged cancellation succeeds without duplicate cancellation events.

Queued cancellation checks under the runner mutex and performs a conditional SQL update excluding active claims. Launch rereads status after acquiring its claim, preventing a cancellation committed immediately before claiming from being overwritten. General running-turn cancellation remains available through its existing endpoint.

The web performs optimistic reorder/hide, disables overlapping controls and reconciles from the server after either outcome. Failures restore the row/order and report an error. Selection generation rejects late responses from another visit; queue revisions reject polls captured before or during a mutation. The duplicated/no-op queue stack is removed. Steer is hidden when no implementation callback is supplied; no Steer parity is claimed here.

## Evidence

`@ux-original-018` uses native browser submissions while a real shell turn is gated. Tests verify optimistic order before the request completes, reload persistence, a real stale-snapshot 409 after concurrent append, failed removal restoring its row, successful cancellation and subsequent execution in the reordered sequence. Separate tests hold an old poll and a mutation response to check stale-row and cross-session isolation.

The parity harness prepends `tests/ux/shell/sh` only to its isolated server process. The wrapper waits for a unique release file only for `UX queue gate:<token>` prompts, then executes the normal `/bin/sh` responder. Tests create no synthetic timeline/queue records. Production shell execution has no test delay hook. Gates are released in test cleanup and have a 60-second safety bound.

Store/engine/API tests cover reopen persistence, unchanged timestamps, invalid permutations, claim races, explicit queue versus steering, queued-only cancellation and duplicate cancellation. Focused race tests run three times. The full browser matrix passes 120/120 executions (10 Classic IDs and 10 Gi regressions); existing functional web tests pass 70/70.

## Limits and terminal adaptation

Return-to-editor, recovery-before-DELETE, Steer consumption and SSE/local optimistic queue deduplication need separate cases. Therefore `016`, `017`, `019`, compose `004` and shared queue scenarios are not passed by this slice.

Terminal queue actions should extend the existing on-demand `/queue` view with a bounded list: select a durable ID, move up/down or cancel, and refresh on conflict. Keep the existing nonzero footer count, with no empty queue row or permanent action bar. Errors use the existing footer/transcript notice. Local queued-draft restoration is not backend removal and must not silently consume a durable queued item. These terminal controls remain unimplemented.
