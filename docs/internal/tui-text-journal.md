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

## Paired text and media prerequisite

`LoadTUIComposerDraft` reads both journals in one transaction without settling
claims. `ClaimTUIComposerDraft` moves the current text and staged media into
claims under one random token, after checking text revision and media ownership.
A failure writing either journal rolls back both. Text-only use does not create
an empty media row. Newer drafts and attachments may be staged while a claim is
held; the six-reference limit includes claimed attachments.

Paired claims carry reciprocal ownership markers (`TUITextClaim.Media` and
`TUIMediaClaim.Text`). Single-journal settle, restore, discard or detach cannot
clear one half; legacy media loading does not reconcile a paired claim. This
protects the supported APIs within the upgraded binary, not arbitrary SQL or
concurrent older binaries that ignore the new fields.

`ReconcileTUIComposerDraft` requires both tokens on the same same-session receipt.
Tokens on separate turns, partial receipts, foreign receipts and unaudited turn
INSERTs remain held. `FinishTUIComposerDraft` is only for the live submitting
caller after its synchronous return. It retires confirmed admissions together,
restores a proven rejection only if no subsequent text edit occurred, and keeps
ambiguous/partial admission held. Newer pending media is preserved on restore.

`ResolveRejectedTUIComposerDraft` explicitly restores or discards only a proven
rejected pair with an exact token/revision. Restore requires a blank editor;
discard removes the old snapshot/references, not newer text or stored media
bytes. An inconsistent pair fails closed. This API is also unwired: there is no
new terminal command or automatic restore.

## Integration limits

This is not restart-persistence acceptance for the TUI. Frontend integration
still needs callback ordering, session-visit ownership, failure visibility and
explicit conflict handling. The private claim token must not become forgeable
through arbitrary prompt metadata. Paired claim/settlement storage is now
available, but the frontend must use it before exposing combined sends;
directed routing must preserve session ownership. Missing submission audits may leave a crash-recovered claim held;
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
claimed for the original text-only slice. Deployment is separate from these
unwired prerequisites.

The paired follow-up expands `make test-tui-text-journal` with race×3 claim
contention, restart, claim/settlement write rollback, cross-session/ref validation,
partial/split/foreign/provisional/confirmed receipts, six-slot accounting,
independent-API refusal, newer-edit retention and explicit rejected-pair handling.
Media race regressions, core/vet/hooks and107functional tests (11existing skips)
pass. A focused read-only review found no blocker in pair atomicity, old-API
split prevention, receipt ownership, newer-edit safety or capacity handling;
earlier timeout/file-access failures supply no evidence.
