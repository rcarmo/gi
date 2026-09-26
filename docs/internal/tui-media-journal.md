# Durable terminal attachment references

Pending terminal attachment references now survive reopening the same database and session. `/attachments` reports the durable list and any unresolved admission; `/detach` removes references without deleting stored media. No idle attachment panel or row is added.

## Storage and admission boundary

The journal uses the existing SQLite `kv_store` under `tui_pending_media_v1`, keyed by session ID. It contains pending native `MediaRef` values and at most one tokenised admission claim. `tui_media_draft.go` performs transactional read/modify/write operations; frontends never replace a complete stale snapshot. File-backed connections use the existing immediate transaction configuration. Staging validates media/session ownership; pending plus claimed references share the six-item limit.

Normalised no-op reads/updates do not rewrite journal timestamps. Stage/claim/detach/settle transitions commit before the frontend adopts their result. Claiming is atomic, checked again against directed-send/model eligibility, and happens before clearing the draft/history input. Storage failure reports an error and preserves the editor. File ingestion and journal staging are separate writes: if staging fails, the stored file remains, the command reports its media ID and no prompt is sent. A crash between those writes can leave an unreferenced stored file; it cannot fabricate a pending reference or prompt.

Accepted native admission retains `tui_media_claim` in turns, steering queue or messages. The original frontend settles only its exact token and restores rejected references only after native `SubmitPrompt` has returned and all admission tables confirm absence. Restoration prepends the original references to any newer staged references within the same transaction. Old callbacks cannot overwrite a replacement claim.

## Restart and multiple clients

A new frontend reads the session journal on attachment commands and before ordinary submission. Staged references can be sent normally. Confirmed admission consumes a recovered claim. If no admission is found, the claim stays held: an old process might still finish submitting, so absence alone is not safe evidence for replay. No background resend or automatic ref restoration occurs.

`/attachments` rechecks held claims. `/detach unresolved` explicitly discards that claim's references, guarded by the observed token; it does not cancel an old process, remove stored files or undo accepted work. The user must inspect history before deliberately reattaching/resending. A still-live caller cannot discard its own in-flight claim. `/detach all` removes pending refs only and leaves held claims untouched.

The durable pending list is shared by terminal clients selecting that database/session. Atomic claims prevent two clients sending the same staged batch, and concurrent staging preserves updates within the shared limit. Media remains session-bound; directed text cannot implicitly move attachments to another agent. Reopening another session does not inherit the list.

This persists media references and claim state only. Plain-text editor drafts, undo/cursor history and queued-draft recovery across process restart remain separate. Pre-journal builds' in-memory attachment lists cannot be migrated after those processes exit. Browser drafts and live authentication state are unchanged.

## Verification

- `make test-tui-media-journal`: store/TUI media admission tests pass under `-race -count=3`. The gate is added to required core CI.
- Store tests cover reopen, recovered unknown versus admitted claim, explicit discard keeping stored bytes, cross-session rejection,12concurrent clients accepting exactly6stages, rollback on failed journal update, duplicate claims, stale settlement tokens and live rejection restore.
- TUI tests preserve existing accepted/rejected/uncertain admission semantics and add restarted claim blocking, newer staged refs, late callback after discard, and exact draft/cursor/history preservation when claiming fails.
- `make test-tui-pending-media`: six fullscreen/regular real tmux PTYs at60×18,100×22,140×36 pass. Each verifies A/B ownership, rejected retry, exact stored bytes, no reuse, detach retention, staged refs after process reopen, injected crash-after-claim recovery, blocked resend, explicit unresolved discard and unchanged idle footprint/settings.
- `make test vet bun-checks` passes. `make test-ux`:107passes,11existing fixture-dependent skips. No browser code changed.

Existing injected in-memory claim tests now install matching durable crash-boundary journal fixtures; native staging validation is not bypassed in production. No tests removed or timeouts increased. Two delegated design/review attempts timed out without results. Evidence preserves local logs and bounded PTY captures, excluding DB backups.

Whole CI and deployment are separate gates. Tests use disposable stores and sessions; no live attachment, prompt or auth mutation was performed. Physical clipboard/device and broad parity acceptance remain open.
