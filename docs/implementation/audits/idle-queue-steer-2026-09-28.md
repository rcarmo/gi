# Explicit idle queue Steer

> Correction: Stop/Resume holds described below were Gi-only behaviour and have been removed. The current contract and results are in [Stop/Resume removal](stop-resume-removal-2026-09-28.md). Earlier hold tests establish only historical Gi behaviour.

Installed Piclaw 3.2.4 lets a queued item's Steer button send that item when no run is active. Gi previously disabled it. The button now permits an explicit idle action after a fresh activity read. Disconnected, unknown and pending states remain fenced.

## Native admission

`POST /api/sessions/:session/queue/:turn/steer` requires an explicit string `active_turn_id`. A nonempty value binds to the observed run, with the later [released-claim fallback](ended-queue-steer-2026-09-28.md) restricted to that latest terminal run and no Stop hold. An empty string means the caller observed idle; missing/null/nonstring values fail validation.

Idle admission claims the exact queued row, session and observed Stop hold atomically. Foreign, cancelled, already claimed or already completed rows conflict. If a newer active claim or different hold wins the race, the request conflicts rather than steering a different run. The selected row keeps its turn identity and captured metadata.

A paused queue is not resumed wholesale: the selected row may bypass the captured Stop hold, but the hold remains for sibling work. A returned, unconsumed steering row can also be explicitly sent. No item auto-sends merely because the UI became idle.

Before launching execution, the original user prompt is durably stored. Failure rolls launch state back, preserving the original queue phase and leaving the item retryable. The store owns the prompt's turn marker and deduplicates it; setup does not insert it again. A focused review prompted a regression for nil/conflicting caller metadata. The transaction does not mutate the caller's map.

## Evidence

- `make test-piclaw-idle-steer`: six Chromium/WebKit phone/tablet/desktop cases run the installed shipped browser against Piclaw's real queue action method and disposable storage/dispatch collaborators. Idle enablement, storage-failure restoration, retry, media and unchanged draft pass. No live inference or backend durability claim.
- `make test-idle-queue-steer`: store/turn/API race×3 coverage passes. Cases include selected-row metadata, paused siblings, duplicate requests, foreign rows, raced active claim/hold/cancellation, launch rollback, prompt-storage failure and prompt identity/deduplication.
- Native idle browser suite: six cases pass, including Stop, reload, rejected request, explicit retry, unchanged draft and sibling queue preservation.
- Existing active-Steer suite: 18 cases pass. Its old readiness check assumed an enabled button implied an active run; it now waits for the actual run because idle Steer is valid.
- Final `make check`: Go/vet/build/hooks and 144 functional tests pass; 11 skipped. Support inventory: 206 tests / 7,760 assertions pass.

The first broad investigation and review attempts timed out. Focused store review identified the caller-controlled deduplication marker; it was fixed and race/full gates rerun. There is no broad review approval claim.

## Still open

This is not full queue parity. Piclaw automatically falls back to normal processing when a run ends during active Steer admission; Gi still rejects a stale active target and requires an explicit retry. Missing-row idempotency, active prompt timing, Return replacement versus recovery merge, Pi dequeue-all/abort restoration and delivery-order proof remain open. `@gi-ux-005` remains a migration-gap regression, not a newly accepted parity mapping.

No supplied component/UI/pane changes, deployment, restart or live-chat writes. Logs: `/workspace/tmp/gi-idle-steer-*.log`; installed oracle artifacts: `test-results/ux-oracle/idle-steer/`.
