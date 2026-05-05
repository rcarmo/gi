# Gi TUI UX user stories

This document tracks the essential user stories for the current gi TUI and maps them to implemented commands, tests, and remaining gaps.

## Scope

Covers:

- out-of-box boot and first-use model readiness
- prompt entry, streaming, cancellation, and history
- model/thinking/runtime controls
- tool/skill discovery and activation
- session and branch workflows
- debug/visibility surfaces
- layout/resize/scroll behavior

## Core stories

### 1. Start the TUI and understand where I am

**As a user**, I want to open gi in a terminal and immediately understand the current session, model, provider, status, and available hints.

- Evidence:
  - `internal/tui/chat.go` header/context/footer rendering
  - `features/tui/assistant_basics.feature`
  - `features/tui/keyboard_behavior.feature`

### 2. Send a prompt and see the model stream its answer

**As a user**, I want prompt submission to feel live, with draft output appearing as it streams and a final assistant reply persisted in the transcript.

- Evidence:
  - `internal/tui/chat.go` `agent_draft_delta` + `new_post` handling
  - `internal/turn/agent_loop.go`
  - `internal/turn/engine.go` bootstrap streaming fallback
  - `internal/tui/chat_test.go`

### 3. Be blocked cleanly when no model is selected on first use

**As a user**, I want the TUI to explain what to do before my first prompt if no model is configured.

- Evidence:
  - `internal/tui/chat.go` first-use prompt guard
  - `internal/tui/chat_test.go`

### 4. Change models and keep my selection

**As a user**, I want `/model` to switch models and persist my choice so the next run keeps it.

- Evidence:
  - `internal/tui/chat.go` `/model`
  - `internal/config/config.go` model persistence
  - `internal/config/config_test.go`

### 5. Change thinking level at runtime

**As a user**, I want `/thinking` to change the current/default thinking level without leaving the TUI.

- Evidence:
  - `internal/tui/chat.go`
  - `features/tui/assistant_basics.feature`

### 6. Inspect compaction and run-time settings

**As a user**, I want `/compact`, `/settings`, `/scrollback`, and `/approvals` to explain how the runtime is configured.

- Evidence:
  - `internal/tui/chat.go`
  - `features/tui/settings_and_approvals.feature`

### 7. Discover tools and skills from inside the TUI

**As a user**, I want to query available tools/skills and activate or reset tools without leaving the terminal.

- Evidence:
  - `internal/tui/chat.go`
  - `features/tui/tool_controls.feature`
  - `features/tui/assistant_basics.feature`

### 8. Move around sessions and branches safely

**As a user**, I want to inspect agents, fork sessions, switch branches, send messages to peers, and see a compact tree of current sessions.

- Evidence:
  - `internal/tui/chat.go`
  - `features/tui/session_workflows.feature`

### 9. Understand what extensions and hooks are loaded

**As a user**, I want `/plugins` to show active extensions and registered hooks for debugging.

- Evidence:
  - `internal/turn/debug_info.go`
  - `internal/tui/chat.go`
  - `features/tui/plugins.feature`

### 10. Use the TUI entirely from the keyboard

**As a user**, I want blur/focus, history navigation, transcript scrolling, resize, and quit behavior to be predictable.

- Evidence:
  - `internal/tui/chat.go`
  - `features/tui/keyboard_behavior.feature`

### 11. Work comfortably with long prompts

**As a user**, I want the editor to grow as I type and to stay readable in narrow terminals.

- Evidence:
  - `internal/tui/multiline_input.go`
  - `internal/tui/chat.go`
  - `internal/tui/chat_test.go`

### 12. Read structured output instead of a plain text wall

**As a user**, I want transcript entries to preserve useful structure for tool results, compaction summaries, Markdown, and tables.

- Evidence:
  - `internal/tui/chat.go`
  - `internal/tui/markdown.go`
  - `internal/tui/chat_test.go`

## Slash-command user stories

One user story per current slash command family:

- `/help` — explain keybindings and commands
- `/tools` — discover, inspect, activate, reset tools
- `/skills` — discover skills
- `/model` — show or change model
- `/thinking` — show or change thinking level
- `/compact` — inspect compaction settings
- `/scrollback` — inspect or change scrollback limit
- `/settings` / `/config` — inspect runtime configuration
- `/approvals` — inspect approval-gate state
- `/cancel` — cancel running/queued turn or explain no-op
- `/agents` — inspect agent/session roster
- `/tree` — inspect branch/session hierarchy
- `/plugins` / `/extensions` — inspect extension/hook state
- `/fork` — create/switch to peer session
- `/switch` — switch to another session/agent
- `/send` — deliver a message to another agent session
- `/where` — inspect current session summary

## Current gaps

### 1. ANSI/bracketed paste parity

Desired story:

**As a user**, I want reliable terminal paste behavior using ANSI/bracketed-paste sequences, similar to Pi.

Current state:

- ordinary pasted text should arrive as terminal rune input
- explicit bracketed-paste sequence handling is not yet evidenced in `go-tui`
- this remains blocked on deeper terminal/parser support or extra handling in gi

### 2. Layout parity audit against Pi/Claude Code screenshots

Desired story:

**As a user**, I want the placement of thinking/model/session/status information to feel close to Pi/Claude Code.

Current state:

- gi now has responsive header/context/footer layout
- a screenshot-based audit and parity judgement still remain to be documented

## Suggested report follow-up

The UX report should summarize:

- implemented user stories
- command coverage
- feature/test evidence
- screenshots at multiple terminal sizes
- remaining blockers and recommendations
