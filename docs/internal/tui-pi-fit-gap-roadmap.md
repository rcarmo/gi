# TUI Pi fit/gap roadmap

## Status

This is the working acceptance note for making `gi -tui` progressively feel closer to Pi while preserving Gi's SQLite-backed runtime architecture and current Go TUI stack.

It intentionally does **not** require pixel-perfect cloning. Parity means the same workflow is discoverable, test-backed, and tmux-friendly, with Gi-specific runtime semantics preserved.

## Baseline inventory

### Current `gi -tui` command surface

Implemented commands in `internal/tui/chat.go`:

- `/help`
- `/tools [query|active|activate|reset]`
- `/skills [query]`
- `/skill:name [args]`
- `/model [name|index]`
- `/scoped-models [list|add|remove|set]`
- `/thinking [level]`
- `/compact`
- `/scrollback [n]`
- `/settings` / `/config`
- `/approvals`
- `/cancel`
- `/agents`
- `/plugins` / `/extensions`
- `/tree`
- `/fork [@agentN]`
- `/switch @agent|session_id`
- `/send @agent message`
- `/where`

Not yet implemented or not at Pi parity:

- `/session`
- `/new`
- `/name <name>`
- `/resume`
- `/clone`
- `/copy`
- `/reload`
- `/scoped-models`
- `/skill:name [args]`
- `/export`
- `/share`
- `/hotkeys`
- `/changelog`

### Current keyboard/editor surface

Implemented in `internal/tui/multiline_input.go` and `internal/tui/chat.go`:

- Enter submits
- Shift+Enter inserts newline
- Escape blurs input
- Backspace/Delete delete characters
- Left/Right move by character
- Home/End move to input start/end
- Up/Down history behavior is covered in TUI tests
- PgUp/PgDn and Home/End transcript scrolling are covered in TUI behavior/tests
- F2/F3 history hints are documented/tested in the TUI feature set
- Focus/blur, mouse focus, resize, and quit behavior have tests/features

Not yet implemented or not at Pi parity:

- configurable keybindings
- Alt+Enter follow-up submission (currently documented gap: `go-tui` event support and runtime follow-up-vs-steering split need a separate implementation slice)
- Alt+Up queued-message restore
- Ctrl+L model selector
- model/thinking cycling shortcuts
- Ctrl+O tool collapse / Ctrl+T thinking collapse parity
- bracketed paste / clipboard image paste (ordinary terminal paste remains unchanged; explicit bracketed paste remains deferred per `tui-paste-analysis.md`)

### Current session/runtime UX surface

Implemented:

- session/agent context summary through `/where`
- session tree/debug output through `/tree`
- peer session creation via `/fork`
- session switching via `/switch`
- peer message sending via `/send`
- topic-native status rendering for runtime turn/tool/hook/routing/session/inbound/dispatcher events
- durable same-session steering in the runtime underneath the TUI

Not yet at Pi parity:

- explicit `/session` detail command
- create a fresh session through `/new`
- rename session through `/name`
- resume recent sessions through `/resume`
- clone active branch through `/clone`
- copy/export/share session affordances
- visible queued/steering count in header/context (implemented after baseline)
- dequeue queued messages to editor

### Current transcript/layout surface

Implemented:

- Pi-like hierarchy: status, context, transcript, input, footer
- compact status icon labels for idle/queued/running/tool/hook/error/compaction states
- responsive footer hints
- editor bindings for word movement, word deletion, line deletion, minimal undo/yank, `!`/`!!` shell shortcuts, Tab path completion, and textual `@path` completion
- terminal-safe Markdown rendering for headings, lists, blockquotes, links, code blocks, and responsive table fallback
- folded multi-line tool results
- code block line counts
- compact runtime compaction summaries
- tmux/Gherkin-friendly text captures under `features/tui/`

Remaining polish:

- denser `/tree`, `/resume`, `/settings`, and model-selection output for narrow terminals
- optional command palette-like flow if it can be built without replacing the stack
- richer collapse/expand affordances for tools/thinking if key support exists

## Implementation map

The plan sidebar groups gaps into these implementation tracks:

1. **Command/session workflow parity** — add Pi-like session commands first because they are visible and mostly store-backed.
2. **Message queue and steering UX parity** — expose the durable steering state that Gi already has. Current implementation surfaces queued turn and steering depth in the context area; Pi-style Alt+Enter follow-up submission remains deferred until `go-tui` key support and a clear runtime follow-up queue split are implemented.
3. **Editor affordance parity** — add only keybindings that `go-tui` can represent cleanly.
4. **Model/settings parity** — improve selection/listing before attempting interactive pickers.
5. **Skills/extensions parity** — expose existing skills/script primitives through Pi-like command surfaces. `/skill:name [args]`, richer `/skills` discovery output, skill metadata warnings, reload discovery reporting, and extension command semantics/namespace docs are now covered. Extension command dispatch remains a documented future implementation because only the contract has landed.
6. **Clipboard/media parity** — start with `/copy` fallback; defer image paste until a media ingestion contract exists.
7. **Interactive polish** — keep `/help`, footer hints, tmux captures, and docs aligned as behavior changes.

## Acceptance criteria

A slice is acceptable when:

- it preserves current runtime/store/tool/SSE/topic contracts;
- it has focused unit tests for command output, editor behavior, or rendering helpers;
- it passes `go test ./internal/tui ./internal/turn ./internal/web` when touching TUI/runtime boundaries;
- it keeps transcript output deterministic enough for tmux pane captures;
- docs are refreshed when user-visible behavior changes;
- missing Pi behavior is documented as adapted/deferred rather than silently implied.

## Current validation baseline

Before starting this roadmap, this command passed:

```bash
go test ./internal/tui ./internal/turn ./internal/web
```
