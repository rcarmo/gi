# TUI layout parity audit

Date: 2026-05-05
Updated: 2026-05-31

## Goal

Compare Gi's current TUI layout against Pi using real tmux captures and keep the accepted row contract synchronized with implementation.

## Current contract

The current accepted target is stricter than the original May 2026 audit: Gi's **steady-state** layout must match Pi's physical row structure.

From top to bottom:

1. transcript starts at row 0;
2. no top chrome, header, status block, or permanent context block;
3. bottom separator;
4. editor/input row;
5. bottom separator;
6. path/branch row;
7. one physical status/notification row.

The detailed contract is `tui-pi-layout-contract.md`; status-line semantics are in `tui-status-line-semantics.md`.

## Evidence captured

Historical artifacts:

- `artifacts/tui-audit/`
- `artifacts/tui-audit-rendered/`
- `artifacts/tui-compare-current-20260527/`

These show the evolution from broad Pi-like hierarchy to the current no-top-chrome, bottom-band-first layout. Regenerate current comparison artifacts after any layout-sensitive change.

## Pi reference characteristics

Observed Pi steady-state layout characteristics:

- transcript/content area starts at the top;
- editor lives in a bottom band;
- path/branch appears below the editor separator;
- final footer/status row contains metrics/model/thinking;
- transient state is shown in the bottom area rather than as permanent top chrome.

## Gi current layout

Gi now follows the same steady-state row contract:

- no top status/header/context lines;
- transcript-first rendering;
- empty editor row is blank, with only cursor when focused/blinking;
- path row uses workspace path plus best-effort git branch;
- final status row is single-line and non-wrapping;
- transient hook/tool/inbound/dispatcher notices prefer the final status row when they are not durable transcript history.

## Remaining differences

- Pi has cost/context-window metrics; Gi currently has message/turn and queue/steering counters plus model/thinking.
- Pi supports richer terminal image paste/drag-drop; Gi currently uses `/attach <path> [prompt]`.
- Pi has richer extension-provided UI surfaces; Gi currently keeps TUI widgets textual and deterministic.

## Conclusion

The old broad-hierarchy target is superseded. The active target is Pi-identical steady-state row order plus a single bottom status/notification line. Future audit work should compare against that stricter contract.
