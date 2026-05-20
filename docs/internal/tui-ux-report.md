# Gi TUI UX report (draft)

Date: 2026-05-05

## Executive summary

The gi TUI now covers the main interaction loop expected from a Pi-like terminal coding assistant:

- boot with immediate status/context visibility
- prompt submission with streaming output
- runtime model/thinking/tool/session controls
- responsive narrow-terminal layout
- expanding multiline input
- transcript scrollback control
- Markdown transcript rendering with responsive table fallback
- tmux-backed Gherkin regression coverage for core workflows

The biggest remaining UX gaps are:

1. **ANSI/bracketed paste parity** with Pi
2. **screenshot-backed layout parity audit** against Pi/Claude Code
3. final packaging of this report as a **PDF with screenshots**

## Evidence base

### Implemented feature coverage

- `features/tui/assistant_basics.feature`
- `features/tui/keyboard_behavior.feature`
- `features/tui/session_workflows.feature`
- `features/tui/tool_controls.feature`
- `features/tui/plugins.feature`
- `features/tui/settings_and_approvals.feature`

### Current TUI implementation files

- `internal/tui/chat.go`
- `internal/tui/chat_test.go`
- `internal/tui/multiline_input.go`
- `internal/tui/markdown.go`

### Supporting runtime/config files

- `internal/config/config.go`
- `internal/config/config_test.go`
- `internal/inference/inference.go`
- `internal/inference/inference_test.go`
- `internal/turn/engine.go`
- `internal/turn/engine_test.go`

### Recent shipped slices

- `8826864 feat(tui): default gi to opencode zen free model`
- `c867c0b feat(tui): simplify input chrome and cap scrollback`
- `6233992 feat(tui): make layout responsive to terminal size`
- `5195d26 feat(tui): render markdown transcript content`

## Feature audit

### 1. Boot and orientation

Status: **implemented**

The TUI surfaces status, session identity, agent, model, provider, thinking level, message/turn counts, and stable footer hints on first render.

Evidence:

- `internal/tui/chat.go`
- `features/tui/assistant_basics.feature`
- `features/tui/keyboard_behavior.feature`

### 2. Streaming answers

Status: **implemented**

Draft deltas are surfaced during generation and finalized into transcript entries when the assistant reply completes.

Evidence:

- `internal/tui/chat.go`
- `internal/turn/engine.go`
- `internal/turn/engine.go`
- `internal/tui/chat_test.go`

### 3. First-use model readiness

Status: **implemented**

Gi now defaults to `opencode-zen/minimax-m2.5-free`, and when no model is configured it blocks prompt submission with first-use guidance.

Evidence:

- `internal/config/config.go`
- `internal/inference/inference.go`
- `internal/inference/inference_test.go`
- `internal/tui/chat_test.go`

### 4. Runtime controls

Status: **implemented**

Available controls include:

- `/model`
- `/thinking`
- `/compact`
- `/scrollback`
- `/settings`
- `/approvals`
- `/cancel`

Evidence:

- `internal/tui/chat.go`
- `features/tui/assistant_basics.feature`
- `features/tui/settings_and_approvals.feature`

### 5. Tool and skill discoverability

Status: **implemented**

The TUI supports staged discovery and active tool management from inside the terminal.

Evidence:

- `internal/tui/chat.go`
- `features/tui/tool_controls.feature`
- `features/tui/assistant_basics.feature`

### 6. Session and branch workflows

Status: **implemented**

Available flows include:

- `/agents`
- `/tree`
- `/fork`
- `/switch`
- `/send`
- `/where`

Evidence:

- `internal/tui/chat.go`
- `features/tui/session_workflows.feature`

### 7. Plugin/debug visibility

Status: **implemented**

The TUI exposes extension and hook visibility with `/plugins` / `/extensions`.

Evidence:

- `internal/turn/debug_info.go`
- `internal/tui/chat.go`
- `features/tui/plugins.feature`

### 8. Keyboard-only operation

Status: **implemented**

Covered behaviors include:

- blur/focus
- history navigation
- transcript scroll
- resize stability
- quit behavior

Evidence:

- `features/tui/keyboard_behavior.feature`
- `internal/tui/chat.go`

### 9. Input ergonomics

Status: **implemented**

The input now:

- expands as content wraps
- uses simplified horizontal-rule chrome
- no longer shows an `Input:` label

Evidence:

- `internal/tui/multiline_input.go`
- `internal/tui/chat.go`
- `internal/tui/chat_test.go`

### 10. Transcript readability

Status: **implemented**

The transcript now supports:

- folded tool results
- compaction summaries
- Markdown projection for headings/lists/blockquotes/links/code
- responsive table fallback for narrow widths

Evidence:

- `internal/tui/markdown.go`
- `internal/tui/chat.go`
- `internal/tui/chat_test.go`

### 11. Responsive terminal layout

Status: **implemented**

The layout now wraps header/footer metadata and budgets transcript space from actual block sizes.

Evidence:

- `internal/tui/chat.go`
- `internal/tui/chat_test.go`
- `features/tui/keyboard_behavior.feature`

## Current UX gaps

### 1. ANSI/bracketed paste support

Status: **not complete**

Observed state:

- Pi documentation references richer terminal input behavior
- gi’s current terminal stack (`go-tui`) shows Kitty key protocol support
- bracketed-paste (`CSI 200~` / `CSI 201~`) handling is not yet evidenced in the parser path

Conclusion:

- this remains a likely upstream/runtime parser gap rather than a simple TUI widget omission

### 2. Screenshot-backed layout parity audit

Status: **not complete**

Observed state:

- gi now has a responsive, information-dense terminal layout
- a screenshot-based parity comparison against Pi/Claude Code still needs to be recorded

## Recommended next steps

1. Audit or extend `go-tui` for bracketed-paste / ANSI paste handling.
2. Capture screenshots at representative sizes (e.g. 60x18, 100x22, wide desktop).
3. Record Pi/Claude Code comparison notes against those screenshots.
4. Export the final version of this report as a PDF with embedded screenshots.

## Appendix: command inventory covered by current UX stories

- `/help`
- `/tools`
- `/skills`
- `/model`
- `/thinking`
- `/compact`
- `/scrollback`
- `/settings`
- `/approvals`
- `/cancel`
- `/agents`
- `/tree`
- `/plugins`
- `/fork`
- `/switch`
- `/send`
- `/where`
