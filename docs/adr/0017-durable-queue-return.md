# ADR-0017: Persist queue-return recovery before removal

## Status

Accepted — 2026-09-22

## Contract conflict

The frozen shared case “Return a queued item to the latest editor draft” requires preserving concurrent text/media/references and persisting recovery before DELETE. Classic `@ux-original-017` and `@ux-compose-004` specify replacing the draft and clearing its media. Gi implements the safer shared behaviour. Those incompatible Classic cases stay unmapped; neither feature source is edited.

The shared corpus has no source ID tags. The existing catalogue assigns ordinal `@shared-28` to this case; a test pins its name to detect mapping drift. Shared pass counts are now reported separately from Classic in the JSON/Markdown matrix.

## Implementation

The visible Return action is available only for a durable queue row. While an operation is pending, overlapping queue controls are disabled. It captures the row's session ID and queued turn ID, parses its serialised text/file-folder/message references, and fetches attachment bytes through session-scoped media endpoints. Media metadata is checked for origin ownership; unreadable or oversized attachments fail before removal. The 10 MiB per-file cap is retained.

After asynchronous media retrieval, the draft repository merges with the latest origin-session draft. It preserves current files/media/references and prepends recovered text. Distinct queue IDs with identical text remain distinct; only the durable queue ID makes a retry idempotent.

The merge and `queueReturns[turnID] = {state: prepared, recoveredAt}` are written in the same IndexedDB session record. UI publication happens immediately after merging, so typing during persistence builds on the merged draft. Before DELETE, the caller awaits that write and all edits queued while it was pending. A persistence error prevents DELETE and retains the recovered content in memory. Retry persists the existing merge rather than prepending it again.

The existing queued-only DELETE atomically cancels an unclaimed queued turn. Success marks the recovery record `removed`. Failed or ambiguous removal leaves the record and draft intact and reports an incomplete return. Reload retains both; retry uses the same durable ID. A turn consumed before DELETE produces a conflict: recovered data is retained with a warning to check whether the original ran before submitting again. No automatic resend or blind deletion occurs.

The current selected chat receives updated editor state only if it is still the origin; another session's draft/focus remains untouched. Return to the editor focuses it with the cursor at the end. Origin recovery survives a switch, reload and A→B→A through the existing repository/mount boundary.

## Evidence

- `@shared-28` across six projects: held real attachment response, concurrent newer text/media, storage quota fault preventing DELETE, failed DELETE, reload/retry without duplicate text/media, persisted recovery inspected before DELETE, real media and file/folder/message refs, focus/cursor restoration.
- Additional native browser regressions: consumption before DELETE retains data and warns across reload; completion after a session switch restores only the origin.
- Repository/helper tests: storage gating, idempotent same-ID retry, distinct same-text IDs, media origin validation, byte recovery and download errors.
- Report test: shared rows stay separate; all six projects are required for a pass. Classic assertions/hashes remain unchanged.

Validation: 192/192 browser executions; Classic 15/236 pass, 221 unmapped; shared 1/42 pass, 41 unmapped. Existing functional suite 70/70, Go tests/vet, hook checks and 23 source/helper tests pass.

## Limits and terminal adaptation

Browser recovery is local to one profile and assumes one editing tab per session. It is not an atomic transaction with the server; crash or network failure between persistence and deletion leaves an intentional retryable recovery record. Cross-tab merge, retention/garbage collection of recovery markers and attachment re-upload deduplication are not implemented. On later send, restored files use the existing upload path; return/retry itself does not upload duplicate media.

Steer remains open. Classic replacement semantics are not offered as an alternate unsafe action. Terminal queue return should apply the same sequence using a native local recovery store: persist origin draft/media refs and durable ID, then request queued-only cancellation, then mark acknowledgement. Use the existing editor and a bounded on-demand queue list, with incomplete-return warnings in the existing footer/transcript. No new idle rows or permanent recovery panel are needed. Terminal durable return/media staging is not implemented in this slice.
