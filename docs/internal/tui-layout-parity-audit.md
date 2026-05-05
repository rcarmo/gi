# TUI layout parity audit

Date: 2026-05-05

## Goal

Compare gi's current TUI layout against Pi (and, where possible, the intended Pi/Claude-style information hierarchy) using real tmux captures and screenshot artifacts.

## Evidence captured

Artifacts directory:

- `artifacts/tui-audit/`
- `artifacts/tui-audit-rendered/`

Generated screenshots:

- `artifacts/tui-audit-rendered/gi-wide.png`
- `artifacts/tui-audit-rendered/gi-narrow.png`
- `artifacts/tui-audit-rendered/gi-markdown.png`

Generated PDF report:

- `artifacts/tui-audit-rendered/tui-feature-report.pdf`

Reference capture:

- `artifacts/tui-audit/pi-wide.txt`

## What was compared

### Pi reference

A real `pi` session was executed in tmux at `100x22` and captured.

Observed Pi layout characteristics:

- strong top-of-screen status/update region
- context line near top with workspace/model/provider information
- large central interaction region
- persistent footer/editor region
- more startup chrome and warning panels than gi

### Gi current layout

Gi was executed in tmux at:

- `100x22` (`gi-wide`)
- `60x18` (`gi-narrow`)
- `100x22` with seeded Markdown transcript (`gi-markdown`)

Observed gi layout characteristics:

- strong top status line
- immediate context/session metadata block under status
- large central transcript region
- simplified horizontal-rule input chrome
- persistent footer hint region
- responsive wrapping of context/footer metadata on narrow widths

## Parity judgement

### Matches achieved

Gi now matches the **broad Pi-style terminal information hierarchy**:

1. **status first**
2. **session/model/provider/context summary next**
3. **main transcript in the center**
4. **input/editor at the bottom**
5. **persistent keyboard/help hints in the footer**

This was the main intent behind the checklist item (“where thinking, model, sessions, etc. are”).

### Remaining differences

- Pi has additional startup chrome, notices, and richer built-in shell/editor affordances.
- Pi's exact visual spacing and density are not cloned pixel-for-pixel.
- Claude Code was not available locally for direct side-by-side terminal capture, so this audit uses Pi as the practical reference implementation.

## Conclusion

This checklist item is satisfied at the current fidelity level because:

- gi was executed in tmux
- reference and target captures were taken
- screenshot artifacts were generated
- the resulting audit shows gi now places status, model/session context, transcript, and input/footer in a Pi-like arrangement suitable for terminal use

The remaining delta is polish/parity tuning, not missing core layout structure.
