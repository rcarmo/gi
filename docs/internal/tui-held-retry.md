# Explicit terminal held retry

Held failures use on-demand command output, not a permanent panel or status row.
The commands reuse the queue response path: fullscreen appends to the transcript;
regular mode prints once above the dock without flushing partial model output.

## Commands

- `/retry [page]` lists up to six held failures in the current session. IDs are
  complete, summaries are control-sanitised and limited to40display cells. A
  next-page hint appears only when another row exists. Results are advisory.
- `/retry check <turn-id>` reads current native failure state. It shows a
  confirmed admitted ID or the version-one reservation token. It does not
  reconcile, submit, clear, skip or release anything.
- `/retry run <turn-id>` explicitly invokes the native atomic retry in this
  session. It retries the stored prompt, not the composer or queued draft.
  Admission creates separate queued work when another turn is active, never
  steering. A repeated successful request returns the same admitted ID. Pending
  unknown admission reports `no resend` and points to `check`.
- `/retry release <turn-id> <token>` releases only an unadmitted version-one
  reservation. It never submits work. The exact session, turn and token must
  match; the transaction fences a paused old owner. A later retry requires a
  separate explicit `run`. Wrong/stale tokens, committed admissions and legacy
  version-zero reservations fail without clearing the hold.

All mutation IDs are complete store IDs, never row indexes or prefix matches.
Foreign and missing turn IDs have the same session-scoped error. The engine
wrappers recheck ownership before calling native recovery. No new keyboard
binding, clipboard action, draft limit, idle row or automatic retry is added.

Typing these commands deliberately replaces the editor text through the normal
command-entry path. The command handlers themselves do not alter draft/cursor,
pending media, process-local queued drafts or selected settings. A successful
retry invalidates the local queue snapshot. The original turn retains its
history; the follow-on revalidates media and tool restrictions. For transaction,
upgrade and legacy limits see [native held retry](held-turn-retry.md).

## Refinement and scope

This adapts native failure recovery to the existing terminal command model:
explicit inspection, full IDs, bounded output, and no idle chrome. New skip/hold
controls, interactive confirmation panels, browser feature mappings, automatic
legacy recovery and mixed-version database writers are outside this slice.
Acceptance is native/session fault coverage plus both terminal modes at three
sizes, existing queue/regular regressions and the isolated functional suite.
Whole CI, deployment and physical terminal acceptance remain separate gates.

## Deployment

Exact `5e595d79def539e8b880cb09610a650990cb87d2` passed whole CI
[36224693623](https://github.com/rcarmo/gi/actions/runs/36224693623), including
required terminal PTYs and four platform builds. Detached-source Makefile build
and restart runs on8090, PID1103331, PGID/SID1103224. Newer no-model guard and
unwired plaintext journal are excluded.

Six read-only Chromium/WebKit probes at390/820/1440 pass existing composer,
model/session picker, context, theme, focus and draft checks with zero HTTP
writes/errors and no permissions requested. SQLite remains62sessions/51turns/
146messages; integrity/FKs clean. Full normalised dumps differ only by the two
expected additive retry columns, excluding the runtime lease. Auth remains
absent. No live terminal mutation or physical/Visual acceptance is claimed.

## Verification

- `make test-held-retry test vet bun-checks` passes. The focused race gate runs
  store, engine and TUI checks three times, including read-only pagination/check,
  foreign/missing/full-ID guards, wrong-token/legacy release refusal, native
  queue admission, duplicate-ID recovery and draft/media preservation.
- `make test-tui-retry-commands BIN_DIR=/tmp/gi-tui-retry-bin` passes six tmux
  PTYs: fullscreen/regular at60×18,100×22,140×36. Disposable fixtures cover eight
  held failures, pagination, restart, production-length turn IDs/tokens, foreign
  rejection, read-only checks, pending no-resend, release with no admission,
  legacy refusal and exactly one queued retry despite repeated run. Final
  Unicode draft/settings and the idle dock remain unchanged; regular mode keeps
  terminal-owned history/selection and no mouse capture.
- `make test-tui-queue-commands test-tui-regular` passes another nine PTYs.
- `make test-ux BIN_DIR=/tmp/gi-tui-retry-functional-bin`:107passed,11existing skips.
- Read-only review found no safety blocker; its stale fallback help finding was
  fixed, and foreign/missing errors were aligned. A required terminal-retry CI
  job runs the six PTYs and retains captures without changing existing budgets.

Initial fixture failures used a non-existent session column and noncanonical
claim timestamps; both were corrected, not bypassed. A60-column regular capture
then exposed clipped retry guidance; admission and repeat guidance now occupy
separate short lines. A repeated race run also reproduced a native shutdown
panic: queue handoff passed nil coordination context to SQLite after engine
cancellation. Handoff and retry/release now return `context.Canceled` when no
live coordination context exists; a regression test covers that boundary. No
coverage was removed and no timeouts or pixel criteria were relaxed.

Evidence: `test-results/tui-retry-commands/summary.json`, per-mode captures and
local gate logs. No live terminal mutations were used. This is not physical
emulator, exact-pixel, screen-reader or full web/TUI parity acceptance.
