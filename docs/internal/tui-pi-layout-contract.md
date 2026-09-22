# Gi TUI Pi-identical layout contract

Status: active layout contract for the Pi/PiClaw UX convergence track.

## Pi outcome bands and fullscreen navigation (2026-09-22; verified)

User messages use Pi's neutral `#343541` band; pending/success/error tools use `#282832`/`#283228`/`#3c2828`, with neutral output text and explicit error labels. Flat bands replace tool box rows. PgUp/PgDn and Home/End navigate rendered transcript rows even while editing; Ctrl-Home/End retain editor movement, Ctrl-O expands/collapses tool output, and wheel input over the fixed dock falls back to the transcript. Three-size native PTY and rendered-cell tests pass without extra idle rows. See [ADR-0030](../adr/0030-terminal-outcome-bands.md).

Pi regular-mode native scrollback, fullscreen transcript search, prompt jumps, text selection/copy/edge-autoscroll, light theme and reflow/eviction anchoring remain open. Internal fullscreen scrolling does not provide native terminal scrollback.

## Terminal reading position (2026-09-22; verified)

Editing and ordinary running/idle submission preserve the existing transcript follow mode. PageUp or mouse history navigation stays in place; explicit navigation back to the newest edge resumes following. Native delayed-provider completion, newer text/cursor, same-row history anchors, resize round trips and zero additional idle rows pass at 60×18, 100×22 and 140×36 (`make test-tui-reading`). [ADR-0029](../adr/0029-terminal-reading-position.md) records evidence and limits. Routed acceptance, recovery, reflow and scrollback eviction still need separate work.

## Accepted-message adaptation (2026-09-22; broader design)

Acceptance should reconcile durable IDs in the existing transcript, clear only the captured editor draft, preserve newer text/cursor and retain history-reading position unless already following the newest edge. Add zero idle rows and no permanent delivery controls. Use the existing bounded error/notice surface for failure; preserve recovery data and never automatically resend uncertain delivery. Verify delayed acceptance, A→B→A, resize and history anchors at 60×18, 100×22 and 140×36 before terminal credit. Current submit callbacks are session-scoped but still request scrolling after accepted routing; media/reference recovery also needs implementation. Browser evidence and the open terminal checks are recorded in [ADR-0028](../adr/0028-accepted-message-refresh.md).

## Manual compaction (2026-09-22)

`/compact` runs native maintenance; `/compact info` shows diagnostics. Alt-C preserves editor/cursor and invokes the same operation. Escape stops active compaction through the focused editor callback. Progress uses the existing stats row (`Compacting m:ss`), with four-second outcomes in the existing optional notice row. No idle row or permanent widget is added. `make test-tui-compaction` verifies 60×18, 100×22 and 140×36 using native hook gates; see [ADR-0023](../adr/0023-terminal-compaction.md).

## Goal

Gi's steady-state terminal layout must match Pi's row structure, not merely approximate it. Gi may display Gi-specific values, but the physical layout and chrome budget should be identical.

## Vertical row order

From top to bottom:

1. Transcript area starts at row 0. No top status/header/context chrome.
2. Bottom separator line spans the full content width.
3. Editor/input row. Empty editor renders as a blank line, with only a cursor when focused/blinking.
4. Bottom separator line spans the full content width.
5. Path/session row: current workspace path plus branch/session marker.
6. Stats row: counts (`m/t`, optional `q/s`) plus token usage and context on the left, model/thinking on the right.
7. Optional transient notification row, shown only while a short-lived status is active.

The bottom band may grow to several lines (PiSwift-style footer), but it must never add top chrome and must stay below the editor.

## Bottom status band

The bottom band carries path/branch, stats, and transient notifications. Updated policy (status row may expand):

- the path row is always present;
- the stats row carries counts plus token/context usage on the left and model/thinking on the right;
- a transient notification row appears only while a short-lived status is active;
- each row truncates rather than wrapping;
- queue/steering counters appear only when non-zero;
- model/thinking are right-aligned where width allows.

Durable events can still be written to the transcript when they are part of history, but short-lived running/tool/error/queued indicators should prefer the bottom band.

## Path/session row

The path row should mirror Pi's footer path row:

```text
/path/to/workspace (branch)
```

Gi uses the workspace path and reads `.git/HEAD` directly for a best-effort branch label. If no branch is available, it displays just the workspace path.

## Editor row

The empty editor row is blank, matching Pi's empty editor band. Gi must not display placeholder text like `Send a message…` in the empty steady-state layout.

## Regression states

Capture and compare at 60x18, 100x22, and 140x36:

- empty startup;
- existing transcript;
- focused empty editor;
- prompt being typed;
- running turn;
- queued follow-up;
- tool/error notification;
- scrolled transcript.
