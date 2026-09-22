# Gi TUI Pi-identical layout contract

Status: active layout contract for the Pi/PiClaw UX convergence track.

## Structured references (2026-09-22; design)

Keep file/folder/message references in the existing editor and explicit commands; add no permanent attachment panel. Existing `@` path completion preserves the prefix and appends `/` to directories, but structured `Files:`/message-reference capture and durable failed-send recovery need separate three-size terminal acceptance. Browser compose-008 is now verified; it earns no terminal credit. See [ADR-0035](../adr/0035-explicit-folder-references.md).

## Fullscreen selection/copy (2026-09-22; verified)

Drag selects rendered cells; held edges scroll after movement; release/Ctrl-C/Ctrl-X copy via the existing clipboard setting. Escape clears selection. A plain tool click retains expansion behavior. Feedback replaces the upper editor separator, adding no idle rows. Clipboard-off, padded wide/combining text, layout/output/session invalidation and late clipboard replies are guarded. Three-size native SGR/OSC52 PTYs and unit/race tests pass; [ADR-0034](../adr/0034-fullscreen-transcript-selection.md). Copies retain visual-row newlines. Link precedence, word/rectangular selection and mutation-stable selection anchors are not implemented.

## Transcript spacing (2026-09-22; corrected and verified)

Match Pi's message-level blank rows: user bands have one top/bottom padded row; assistant messages have a leading uncolored separator; tool output has a leading uncolored separator plus top/bottom padding inside its outcome band. Horizontal padding is one column. Group Markdown continuation rows before padding; keep tool blocks borderless. Minimal footprint restricts permanent chrome, not these readable message boundaries. The editor/footer rows remain unchanged. Three-size buffer/PTY/search/scrollback tests and screenshots verify the correction; see [ADR-0033](../adr/0033-pi-transcript-spacing.md).

## Fullscreen rendered search and prompt jumps (2026-09-22; verified)

Ctrl-Shift-F temporarily replaces the editor with an independent query input; the existing separator shows counts/hints. Enter/Shift-Enter (or Ctrl-G/Ctrl-Shift-G) navigate matching rows; Escape restores editor text/cursor/undo/yank and reading mode. Ctrl-Shift-Up/Down jump between rendered user prompts. Three-size live search/highlight/Unicode/tool-expansion/resize/reopen/native-arrival checks add no idle rows. This is literal case-insensitive per-rendered-row search with whole-row highlighting; cross-wrap matching, individual-occurrence highlighting and clickable controls are not implemented. [ADR-0032](../adr/0032-fullscreen-transcript-search.md).

## Regular-mode native scrollback (2026-09-22; verified)

`gi -tui -tui-mode regular` is opt-in; fullscreen remains the default. Completed retained output prints once, expanded and outcome-colored, into terminal-owned scrollback. No alternate screen or mouse capture. Five idle dock rows hold the existing editor/separators/footer; active preview is temporary and at most three rows. Terminal/multiplexer wheel, selection and copy remain native. Home/End edit the draft. Three-size ordered-history/selection/draft/cursor/multiline/resize/selector/session/exit/reopen tests pass. Resize history markers re-establish go-tui geometry before dock growth. See [ADR-0031](../adr/0031-regular-terminal-scrollback.md) for exact bounds and retention limits.

## Pi outcome bands and fullscreen navigation (2026-09-22; verified)

User messages use Pi's neutral `#343541` band; pending/success/error tools use `#282832`/`#283228`/`#3c2828`, with neutral output text and explicit error labels. Flat bands replace tool box rows. PgUp/PgDn and Home/End navigate rendered transcript rows even while editing; Ctrl-Home/End retain editor movement, Ctrl-O expands/collapses tool output, and wheel input over the fixed dock falls back to the transcript. Three-size native PTY and rendered-cell tests pass without extra idle rows. See [ADR-0030](../adr/0030-terminal-outcome-bands.md).

Light theme and stronger reflow/eviction anchoring remain open; search, prompt jumps and fullscreen selection/copy/edge scrolling are implemented above with explicit limits. Regular mode above provides terminal-owned history separately from fullscreen navigation.

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
