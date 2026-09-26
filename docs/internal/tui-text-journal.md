# Terminal plaintext journal prerequisite

The store exposes a revision-checked draft journal under the
`tui_text_draft_v1` KV namespace. It is not connected to terminal input yet.
Typing, session switching and submitted-work recall retain their current
process-local behaviour. This slice adds no autosave, commands or idle rows.

## Contract

- `LoadTUITextDraft` is read-only and session-scoped. Missing sessions fail;
  invalid UTF-8, cursor positions or corrupt journal state fail closed.
- `SaveTUITextDraft` compares the expected revision in an immediate SQLite
  transaction. Text bytes, whitespace and rune-offset cursor are preserved
  without truncation. Identical saves do not advance the revision. Empty records
  retain their revision, preventing stale revision-zero writes after clearing.
- `ClaimTUITextDraft` moves the editable snapshot into a held claim and clears
  only its editable slot. It must commit before any frontend clears its editor
  or starts submission. Tokens use256random bits, not wall-clock IDs. Only one
  unresolved claim is allowed, but newer draft edits may be saved separately.
- The future caller must carry `tui_text_claim` in native admission metadata.
  `ReconcileTUITextDraft` retires only confirmed receipts in the same session:
  a turn plus its post-rollback `turn.submitted` event, a steering row, or a
  retained message. A turn INSERT alone can still be rolled back by subturn
  setup. No prompt-text matching or session fallback is used.
- `FinishTUITextDraft` is for the original live caller only, after synchronous
  submission returns. Success still requires a stored receipt. An error plus an
  unaudited turn stays held; it is neither restored nor discarded. Only proven
  absence on error permits restoration, and only if no edit occurred since the
  claim. Newer text is never overwritten.
- If newer edits exist on proven rejection, the old snapshot is marked rejected
  and remains separate. `RestoreRejectedTUITextDraft` requires a blank editable
  slot plus exact revision/token; `DiscardRejectedTUITextDraft` drops only that
  rejected snapshot, keeping newer text. Neither operation accepts an unknown
  admission. Discard permits the newer draft to be claimed without resending
  the older one. Revision exhaustion is rejected before creating an unresolvable
  claim.

## Integration limits

This is not restart-persistence acceptance for the TUI. Frontend integration
still needs callback ordering, session-visit ownership, failure visibility and
explicit conflict handling. The private claim token must not become forgeable
through arbitrary prompt metadata. Text and media claims must be coordinated
before exposing combined sends, and directed routing must preserve session
ownership. Missing submission audits may leave a crash-recovered claim held;
there is no forced discard or automatic replay of that uncertainty.

The existing `queuedDrafts` stack recalls already-submitted work. It is not an
unsent queue and is deliberately not copied into this journal. No claim is made
for its restart restoration, automatic resend, physical terminal behaviour or
new web feature mappings.

## Verification

`make test-tui-text-journal` passes store race tests three times and runs in
required CI. Tests cover two file-backed connections racing saves/claims,
clear/retype revision fencing, exact large Unicode/combining-character text and
cursor on reopen, wrong session/token, foreign receipts, absent/provisional/
confirmed admission, steering/message receipts, newer edit retention, explicit
rejected restore/discard, random token replacement, corruption, failed writes,
failed receipt reads and revision exhaustion. No timeouts or assertions were
relaxed. File-backed `_txlock=immediate` is the concurrency acceptance path.

`make test vet bun-checks` passes. The isolated functional suite passes107tests
with11existing skips. Review found token reuse risk, a missing rejected-snapshot
discard path and ambiguous-error settlement; all gained implementation and
regression coverage. No independent final approval or live journal use is
claimed. Deployment is separate from this unwired prerequisite.
