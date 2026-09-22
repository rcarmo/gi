# ADR-0034: Fullscreen transcript selection and clipboard copying

## Status

Accepted — 2026-09-22. Fullscreen primary-button drag selection, copy and held-edge scrolling work on the padded Pi-style transcript. Regular mode continues to use terminal-owned selection.

## Interaction and footprint

Press and drag across rendered transcript cells to select text. Holding the pointer at the top or bottom viewport edge advances one row every 80 ms after an actual drag begins. Release copies through the configured clipboard mode. Ctrl-C or Ctrl-X repeats the copy while a selection exists; Escape clears selection before affecting search or compaction. Clicking without dragging still toggles a tool block.

Selected cells reverse foreground/background. Instructions and four-second copy outcomes replace the existing upper editor separator; no permanent controls or additional idle rows are added. The editor text, cursor and history are untouched.

Rows and columns come from the same rendered-cell snapshot used by fullscreen search. Blank separators and padded rows remain selectable. Wide and combining glyphs are copied as complete cells; trailing row spaces are trimmed and visual rows are joined with newlines. Copies contain plain displayed text, not terminal styling. Soft-wrapped rows retain visual newlines; reconstructing the original unwrapped source is outside this implementation.

## Clipboard policy and ownership

The existing `tuiClipboardMode` setting remains authoritative:

- `off`: keep selection and show an opt-in hint, without a clipboard write.
- `osc52`: emit the existing OSC 52 sequence, rejecting content above 64 KiB before writing.
- `native`/`auto`: use the existing native helper selection and two-second timeout. Run outside the UI loop, with one outstanding native write per frontend so clipboard operations cannot complete out of order. A busy copy can be retried.

A selection snapshots the session generation, retained transcript, expansion/search state, terminal dimensions and viewport dimensions. Output, resize, search changes, picker activation or session switching invalidate that selection instead of copying text that moved underneath it. Already-dispatched external clipboard writes cannot be retracted; their completion notices are fenced by selection and session ownership. Clipboard failure retains a current selection for retry. OSC 52 acceptance by a terminal is not an acknowledgement from the desktop clipboard.

## Click and scroll repairs

The saved WIP's tool-click regression had two causes: a stationary press at an edge could trigger autoscroll, and changing from follow mode to selection used a stale model scroll offset. The corrected handler snapshots the actual post-layout scroll offset and starts edge scrolling only after pointer movement. Tool hit identity comes from the rendered-row index, including its padded content region, and survives replacement of the element tree during selection rendering. A plain click restores the prior follow mode.

## Evidence

`make test-tui-selection` runs real tmux PTYs at 60×18, 100×22 and 140×36. Each isolated tmux server enables clipboard reception, so the test reads actual OSC 52 data rather than mocking the renderer. Native shell turns provide history. Tests cover forward/reverse drags, release/Ctrl-C/Ctrl-X copy, highlight/Escape/editor preservation, held-edge selection of offscreen output, resize and new-output invalidation, normal tool-click expansion, clipboard opt-out, picker cancellation and unchanged idle row count.

Unit tests cover padded wide/combining cells, reverse selections, partial-wide-cell boundaries, stationary edge presses, click hit identity, search/session/resize/output invalidation, clipboard error/size handling, one pending native write and stale completion notices. Full Go tests/vet, terminal race suite repeated three times, all outcome/search/reading/regular/session/model/compaction/smoke/Gherkin suites, 29 helpers and 70/70 functional browser tests pass.

Fresh real 100×22 XTerm drag and clipboard-off screenshots are attached, under `/workspace/tmp/gi-ui-captures/september22/`. Text/ANSI evidence is in `test-results/tui-selection/`. Native helper integration on every OS is not established by the OSC 52 PTY tests; the existing helper-selection and execution tests remain separate evidence.

## Remaining parity

Double/triple-click word/line selection, rectangular selection, clickable OSC 8 link precedence and source-unwrapped copying are not implemented. Selection cancels on changed output/reflow rather than tracking stable text through mutation. Stronger reader anchors, light theme, per-occurrence/cross-wrap search, oversized-editor limits and the remaining web families are still open.

No web source or frozen criterion changed. Browser coverage stays 34/236 Classic and 2/42 shared, with 202/40 unmapped; the latest complete browser matrix remains 384/384 at `4071c01`. Folder-reference work remains saved in `/workspace/tmp/gi-folder-reference-wip.patch` and is not shipped here. ADR-0033's unshipped selection warning is superseded by this verified slice.
