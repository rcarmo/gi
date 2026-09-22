# ADR-0021: Persist compaction coverage without deleting history

## Status

Accepted — 2026-09-22. Automatic compaction checkpoint; manual Compact remains disabled.

## Context projection

`context_checkpoints` stores a session-local version, summary and covered message IDs/fingerprints. Timeline messages are never deleted or rewritten by compaction. Provider context reads use a virtual summary followed by uncovered user/assistant messages. Stored compaction audit messages are excluded from inference, including older run-local summaries, to avoid adding summaries to full history indefinitely.

A fingerprint covers role, content, full payload and creation time. If any covered message is missing, changed or duplicated in coverage, projection falls back to the full current non-compaction history. The previous checkpoint version remains available to fence later writes. Invalid checkpoint JSON surfaces an error rather than silently dropping history. New messages remain visible by ID even when backdated before covered rows.

Completion commits the checkpoint, compaction summary message, terminal event and phase restoration in the existing transaction. The original history snapshot and checkpoint version must still match. Cancellation does not write a checkpoint. A summary/checkpoint failure rolls back all completion changes. A conflicting history edit causes a failed compaction/turn; it is never automatically retried or sent from a stale context.

## Eligibility

Before the asynchronous compaction hook, the engine captures the native context snapshot. On completion, the current provider message list must exactly equal the native projection. Hook replacement/prepending, live tool exchanges or other unmatched context stays run-local and does not advance durable coverage. Media-bearing older messages are also excluded from durable compaction because the summariser receives a text transcript, not their bytes.

The new covered IDs must equal the previous coverage plus the exact prefix actually summarised from the current native projection. Commit reconstructs that prefix rather than accepting an arbitrary subset. A delegate review identified the need for this extra check; a forged middle-message coverage test verifies rejection. Coverage on read is an ID set, not a timestamp prefix: backdated inserts intentionally remain uncovered. Direct database tampering with a self-consistent coverage record is outside this integrity model.

Repeated compaction folds the previous virtual summary into the new summary and extends coverage. Later turns and process reopen load the checkpoint. No cursor tied to wall-clock ordering is used, and the original timeline remains available to the user.

`durable_context` on completion/audit records distinguishes committed checkpoints from run-local compaction. Latest provider-request usage stays historical until another real request supplies usage; no fabricated post-compaction reduction is reported.

## Evidence

- Store tests: reopen, session isolation, repeated checkpoints, preserved timeline bytes/IDs, backdated new messages, covered edits/deletes, history/version conflicts, cancellation, atomic summary-write rollback and non-prefix rejection.
- Engine tests: a second native turn receives the committed summary; hook/tool projection changes and older media prevent durable coverage.
- Browser regression: real provider turns create history, automatic compaction commits, reload retains the draft/timeline, and the next provider response echoes the actual summary/new request without covered old history. Every original timeline record remains unchanged.
- Validation: 54/54 compaction browser executions; 282/282 combined browser executions; 70/70 functional; 24/24 helpers; Go tests/vet, hook checks and targeted races repeated three times. Live terminal session/model regression passes at 60×18, 100×22 and 140×36 without additional idle rows.

Classic coverage remains 25/236, shared 2/42 (211/40 unmapped). The new browser regression is Gi-specific and earns no extra frozen-case pass.

## Limits and terminal use

This is a conservative eligible-context checkpoint, not a universal tool/multimodal compaction scheme. Hook-authored summaries are trusted at the existing hook boundary. The default summariser remains the transcript-excerpt heuristic; summary fidelity and token reduction are not guaranteed. Storage reads currently scan/fingerprint session inference history for integrity; large-history performance still needs measurement. Full process-crash fault injection is not covered.

The Go-native TUI automatically benefits from the same provider context assembly. No terminal layout or permanent meter changes. Future manual `/compact` and cancellation UI should use the existing transient notice/keys, preserve drafts/cursor and remain bounded at the three terminal sizes. Live terminal checkpoint/reopen acceptance is separate from the session/model regression above.
