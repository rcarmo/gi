# Gi TUI UX report

Date: 2026-05-05
Updated: 2026-05-25

## Executive summary

The gi TUI now covers the main interaction loop expected from a Pi-like terminal coding assistant:

- boot with immediate status/context visibility
- prompt submission with streaming output
- runtime model/thinking/tool/session controls, including model/thinking cycling
- Pi-like session commands (`/session`, `/new`, `/name`, `/resume`, `/clone`, `/reload`)
- queue/steering visibility and queued draft restore
- responsive narrow-terminal layout with compact command output
- expanding multiline input with word movement/deletion, line deletion, undo/yank, Tab completion, and `@path` fallback
- transcript scrollback control
- Markdown transcript rendering with responsive table fallback
- tmux-backed Gherkin regression coverage for core and Pi-like workflows
- documented clipboard/media boundaries, including opt-in OSC 52/native copy support, and extension-author namespace contracts

The biggest remaining UX gaps are now deliberately bounded:

1. **ANSI/bracketed paste and image paste parity** are deferred pending parser/editor support and a shared media ingestion contract.
2. **Rich visual pickers/collapse widgets** remain future work unless they can be added without changing the current Go TUI stack.
3. **Screenshot-backed layout parity audit** and optional PDF packaging remain documentation/reporting follow-ups, not blockers for the current tmux-friendly parity slice.

## Evidence base

### Implemented feature coverage

- `features/tui/assistant_basics.feature`
- `features/tui/keyboard_behavior.feature`
- `features/tui/session_workflows.feature`
- `features/tui/tool_controls.feature`
- `features/tui/plugins.feature`
- `features/tui/settings_and_approvals.feature`
- `features/tui/pi_like_workflows.feature`

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

The original report predated the Pi fit/gap implementation. The current closure status is summarized in `tui-pi-fit-gap-roadmap.md` and `tui-pi-parity-plan.md`; notable additions include session workflow commands, queue feedback, editor bindings, model/settings management, skill discovery/loading, command palette fallback, clipboard/media boundary docs, and expanded tmux/Gherkin coverage.

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

- `internal/turn/engine.go`
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

## Bounded remaining gaps

### 1. ANSI/bracketed paste and media paste

Status: **deferred intentionally**

Observed state:

- ordinary text paste remains terminal-rune behavior;
- `/copy` is locked to transcript-safe fallback behavior by default;
- OSC 52 and native clipboard helpers are available opt-in and documented in `tui-clipboard-media.md`;
- bracketed-paste (`CSI 200~` / `CSI 201~`) still depends on parser/editor support;
- image paste is deferred until Gi has a shared media ingestion contract.

Conclusion:

- this is now a documented boundary rather than an open parity blocker for the current TUI iteration.

### 2. Rich visual pickers/collapse widgets

Status: **future polish**

Observed state:

- Gi now has textual fallbacks for command palette behavior, model selection, skills, and narrow-terminal session maps;
- richer visual pickers/collapse widgets should only be added if they fit the current Go TUI stack without a migration.

### 3. Screenshot-backed layout parity audit

Status: **optional reporting follow-up**

Observed state:

- gi now has a responsive, information-dense terminal layout with tmux/Gherkin captures;
- a screenshot-based comparison against Pi/Claude Code can still be recorded, but is not required for the accepted tmux-friendly parity slice.

## Recommended next steps

1. Define a separate follow-up plan for any deferred rich picker/collapse work.
2. If richer clipboard behavior becomes necessary, add terminal-specific guidance or UI around the existing opt-in OSC 52/native targets.
3. If media paste becomes a hard requirement, design shared web/TUI/API media ingestion first.
4. Optionally capture screenshot comparisons at representative sizes for reporting.

## Appendix: command inventory covered by current UX stories

- `/help`
- `/commands` / `/palette`
- `/session`
- `/new`
- `/name`
- `/resume`
- `/clone`
- `/copy [--osc52|--native|--auto|--fallback]`
- `/reload`
- `/tools`
- `/skills`
- `/skill:name`
- `/model`
- `/scoped-models`
- `/thinking`
- `/compact`
- `/scrollback`
- `/settings`
- `/approvals`
- `/cancel`
- `/agents`
- `/tree`
- `/plugins` / `/extensions`
- `/fork`
- `/switch`
- `/send`
- `/where`
- `!cmd` / `!!cmd`
