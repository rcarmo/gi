# TUI stack evaluation

## Status

Decision: **do not switch TUI stacks now**.

This evaluates PicoClaw/Pi-style TUI/launcher reuse after the core runtime refactor stabilized. It builds on the existing layout, paste, and UX parity notes:

- `tui-layout-parity-audit.md`
- `tui-paste-analysis.md`
- `tui-pi-parity-plan.md`
- `tui-ux-report.md`

## Evaluation criteria

### Terminal paste support quality

Current evidence:

- Pi has strong clipboard/image integration and editor injection APIs.
- No confirmed evidence was found for dedicated bracketed text paste parsing (`CSI 200~` / `CSI 201~`) in the inspected Pi runtime path.
- Gi's current Go TUI stack accepts ordinary terminal-emitted text paste, but true bracketed-paste parity would require parser/editor support in `go-tui` or a replacement editor component.

Decision:

- Paste support alone does not justify a stack switch.
- If bracketed paste becomes mandatory, first evaluate adding paste events to the existing Go TUI/editor path.

### Keyboard protocol support

Current evidence:

- Gi already has coverage for key handling, scroll, resize, blur/focus, history hints, and command workflows through tmux/Gherkin tests.
- Kitty/modified-key behavior remains a terminal/editor capability concern, not a runtime coordination blocker.

Decision:

- Current keyboard support is sufficient for the refactored runtime.
- Keep improving targeted key handling in-place rather than replacing the whole stack.

### Multiline/editor behavior

Current evidence:

- Gi has an expanding multiline input area and simplified Pi-like footer/input chrome.
- Terminal-safe Markdown rendering, responsive layout, and transcript/footer/status hierarchy are already implemented and tested.
- Pi has richer shell/editor affordances, but those are polish gaps rather than runtime blockers.

Decision:

- Keep the current editor path.
- Treat richer editor affordances as incremental UX work, not a stack migration prerequisite.

### Integration cost with Gi runtime/store model

Current evidence:

- Gi's TUI now consumes canonical runtime topics for active-session status/transcript updates and preserves fallback behavior when no live topic subscription exists.
- It submits through engine APIs and does not own runtime state.
- Replacing the TUI stack would need to preserve SQLite-backed session identity, active-turn coordination, steering, topics, direct ingress behavior, command workflows, and tmux/Gherkin coverage.

Decision:

- Switching stacks would have high integration risk and limited runtime correctness upside.
- The current TUI is stable enough under the refactored runtime; continue incremental improvements.

## Final decision

Do **not** switch to PicoClaw/Pi's TUI/launcher stack at this stage.

The current Gi TUI should remain in place because it:

- is already topic-native for runtime-critical families,
- preserves Gi's SQLite-backed runtime model,
- has tmux/Gherkin regression coverage,
- matches the broad Pi information hierarchy,
- can receive targeted paste/editor/key improvements without a rewrite.

## Future revisit triggers

Reconsider a stack migration only if at least one of these becomes true:

- bracketed paste/image paste becomes a hard requirement and cannot be added cleanly to the current Go TUI/editor stack,
- keyboard protocol support blocks core workflows across common terminals,
- multiline/editor behavior cannot meet user requirements with incremental changes,
- a replacement stack can demonstrate lower integration risk while preserving Gi's topic/store/runtime contracts.
