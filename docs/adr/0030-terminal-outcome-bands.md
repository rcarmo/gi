# ADR-0030: Pi outcome bands and fullscreen transcript navigation

## Status

Accepted — 2026-09-22. Implements the user's requested Pi-style output striping and a first fullscreen scrollback slice. Regular-mode terminal-owned scrollback and remaining fullscreen capabilities are open work.

## Source and scope

Reference: installed `@earendil-works/pi-tui` / `pi-coding-agent` 0.85.1 in the PiClaw 3.2.1 baseline release. Read `README.md`, `docs/tui.md`, `docs/themes.md`, `docs/keybindings.md` and `docs/terminal-setup.md`, plus built-in dark/light theme JSON and interactive user/tool renderers.

Pi bands represent content and outcomes. They are not alternating row-number stripes. Its user-message renderer uses `userMessageBg`; its tool renderer selects `toolPendingBg` while partial, then `toolErrorBg` for errors or `toolSuccessBg` for success. The dark theme JSON SHA-256 is `103a5aecb74a2dab5cc903c9741845ee6158658ce2ff6e5445948784116eaef8`.

Gi uses these exact dark RGB values:

| Content | Background |
|---|---|
| User message | `#343541` |
| Pending/unknown tool result | `#282832` |
| Successful tool/bash result | `#283228` |
| Failed/cancelled tool/bash result | `#3c2828` |

Tool output is neutral `#d4d4d4`; error headings use `#f48771`. Status text remains visible, so color is not the sole outcome signal. Assistant output retains the terminal background. Bands fill each rendered block's width, including padding and empty trailing cells. Tool/error/bash blocks are flat; they add no box-border rows or permanent widgets. Existing diff colors remain intact.

## Fullscreen navigation

Gi's go-tui frontend uses the alternate screen. Rendered content dimensions now determine page and bottom bounds, including expanded tool output and wrapped plain text. Following schedules bottom alignment after layout, when the final height is known. Fixed single-line height no longer clips plain output.

- PgUp/PgDn navigate transcript pages.
- Home/End navigate transcript top/bottom even while the editor is focused.
- Ctrl-Home/Ctrl-End and Ctrl-A/Ctrl-E retain editor line navigation.
- Ctrl-O toggles expandable tool/bash output together; F6/F7/F8 and pointer selection remain available.
- Mouse wheel over the editor/footer falls back to transcript scrolling; an open selector retains ownership.
- Typing and native completion retain the reading policy from ADR-0029.

The input component's focused Home/End handlers run before the framework's preemptive handlers. Explicit callbacks route those two keys to the transcript; generic preemption alone was insufficient.

## Evidence

`make test-tui-outcomes` uses real tmux PTYs, native shell-provider execution, a process gate and local shell commands. At 60×18, 100×22 and 140×36 it checks captured RGB ANSI backgrounds for user/pending/success/error states, long wrapped output, Ctrl-O expansion/collapse, focused-editor Home/End, PageUp, editor Ctrl-Home/End, no unintended messages and identical idle separator positions.

Unit tests inspect rendered buffer cells at the left/middle/right edges, check absence of box rows, distinguish outcome colors, verify actual rendered scroll bounds, global tool expansion and wheel fallback without editor mutation. Full Go tests/vet and the terminal race suite repeated three times pass. Three-size reading/session/model/compaction acceptance, terminal smoke/Gherkin, 29 helpers and 70/70 functional browser tests pass. The local settings Gherkin now navigates Home/End to inspect both ends of the longer wrapped settings output; its assertions remain intact. Frozen browser features are unchanged.

New real XTerm screenshots of pending and completed bands were captured at 100×22 under `/workspace/tmp/gi-ui-captures/september22/`. PTY text/ANSI evidence lives in `test-results/tui-outcomes/`.

## Remaining work

Pi regular mode leaves scrollback and text selection to the terminal; fullscreen mode adds rendered-transcript search, marked-message navigation, pointer selection/copy with edge autoscroll and a temporary jump-to-latest affordance. Gi does not yet match those capabilities. Native terminal scrollback needs an inline/regular rendering path and cannot be replaced by this internal scrolling change.

Current bands use Pi's dark theme only. Automatic light-theme selection, full mixed-inline-span reflow, anchors across scrollback eviction or above-viewport edits, and full output beyond existing retention limits still need implementation and evidence. No browser mapping credit is added: 34/236 Classic, 2/42 shared, 202/40 unmapped; the latest complete web matrix remains the 384/384 run for `4071c01`.

The uncommitted folder-reference slice was saved separately as `/workspace/tmp/gi-folder-reference-wip.patch` before this work. It has only two Chromium-phone acceptance passes and is not mapped or shipped with the terminal changes.
