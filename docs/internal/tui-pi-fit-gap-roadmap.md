# TUI Pi fit/gap roadmap

## Status

This is the closure/acceptance note for the Pi-like `gi -tui` iteration. The iteration made the TUI progressively feel closer to Pi while preserving Gi's SQLite-backed runtime architecture, current Go TUI stack, runtime/tool/SSE/topic contracts, and tmux-friendly rendering.

It intentionally does **not** require pixel-perfect cloning. Parity means the same workflow is discoverable, test-backed, deterministic in terminal captures, and adapted to Gi-specific runtime semantics.

## Current `gi -tui` command surface

Implemented commands in `internal/tui/chat.go`:

- `/help`
- `/commands [query]` / `/palette [query]`
- `/session`
- `/new`
- `/name <name>`
- `/resume [index|session_id]`
- `/clone [@agentN]`
- `/copy [--osc52|--native|--auto|--fallback]`
- `/reload`
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
- `!cmd` and `!!cmd` shell shortcuts

Still intentionally absent or deferred:

- `/export` and `/share` rich session export/share affordances;
- `/hotkeys` and `/changelog` dedicated Pi-style informational commands;
- process-extension command dispatch (JS/Joker extension command dispatch is now implemented);
- richer interactive pickers/widgets beyond the textual `/commands` fallback.

## Current keyboard/editor surface

Implemented in `internal/tui/multiline_input.go` and `internal/tui/chat.go`:

- Enter submits;
- Alt+Enter uses the same submit path and is documented as the queue-friendly submit chord;
- Shift+Enter inserts newline;
- Escape blurs input;
- Backspace/Delete delete characters;
- Left/Right move by character;
- Alt+Left/Alt+Right move by word;
- Ctrl+W and Alt+Backspace delete word backward;
- Alt+Delete deletes word forward;
- Ctrl+A/Ctrl+E move to input start/end;
- Ctrl+U/Ctrl+K delete to input start/end;
- Ctrl+Z/Ctrl+Y provide minimal one-step undo/yank;
- Tab completes workspace-relative paths;
- textual `@path` completion uses the same Tab fallback;
- Alt+Up restores the most recent locally queued draft;
- Up/Down, F2/F3, Ctrl+P/Ctrl+N history paths are covered;
- PgUp/PgDn and Home/End transcript scrolling are covered;
- Ctrl+L/Alt+L cycle enabled models;
- Ctrl+T/Alt+T cycle thinking levels;
- focus/blur, mouse focus, resize, and quit behavior have tests/features.

Deferred/adapted editor items:

- configurable keybindings are not implemented;
- Ctrl+O tool collapse is not implemented;
- Ctrl+T is adapted for thinking-level cycling rather than a Pi-style thinking collapse UI;
- ordinary text paste remains terminal-rune behavior;
- explicit bracketed paste and image paste are deferred per `tui-paste-analysis.md` and `tui-clipboard-media.md`.

## Current session/runtime UX surface

Implemented:

- session/agent context summary through `/where` and persistent context area;
- detailed `/session` command with queue/steering counts and active turn state;
- `/new`, `/name`, `/resume`, `/clone`, `/copy`, and `/reload` command/session workflow affordances;
- session tree/debug output through `/tree`, with compact narrow-terminal formatting;
- peer session creation via `/fork`;
- session switching via `/switch`;
- peer message sending via `/send`;
- topic-native status rendering for runtime turn/tool/hook/routing/session/inbound/dispatcher events;
- durable same-session steering in the runtime underneath the TUI;
- visible queued/steering counts in the context summary;
- visible transcript/status feedback when a message is queued during an active turn;
- Alt+Up local queued-draft restore.

Adapted/deferred:

- Pi-style follow-up vs steering split is documented as an adapted gap: Gi currently routes active-session submissions through existing steering/queue semantics;
- rich export/share UX is deferred.

## Current transcript/layout surface

Implemented:

- Pi-like hierarchy: status, context, transcript, input, footer;
- compact status icon labels for idle/queued/running/tool/hook/error/compaction states;
- responsive footer hints;
- grouped `/help` with keys, editor bindings, runtime controls, discovery commands, and session workflows;
- textual `/commands [query]` palette fallback;
- grouped `/settings` with runtime, model, editor, session, discovery, compaction, and peering sections;
- denser `/tree`, `/resume`, `/settings`, and model-selection output for narrow terminals;
- editor bindings for word movement, word deletion, line deletion, minimal undo/yank, `!`/`!!` shell shortcuts, Tab path completion, and textual `@path` completion;
- terminal-safe Markdown rendering for headings, lists, blockquotes, links, code blocks, and responsive table fallback;
- folded multi-line tool results;
- code block line counts;
- compact runtime compaction summaries;
- tmux/Gherkin-friendly text captures under `features/tui/`.

Remaining polish:

- richer collapse/expand affordances for tools/thinking if key support exists;
- richer visual pickers/palettes only if they can be added without replacing the current stack.

## Skills/extensions surface

Implemented/adapted:

- `/skills [query]` lists discovered skills, command hints, source paths, and metadata warnings;
- `/skill:name [args]` loads discovered `SKILL.md` text;
- skill metadata warnings cover missing/empty `Name` and `Description` fields while retaining fallback behavior;
- `/reload` refreshes config and discovery safely and reports extension discovered/mounted counts;
- extension command registration and TUI dispatch are implemented for JS/Joker extensions and documented in `extension-command-semantics.md`;
- `gi.state`, `gi.topics`, and `gi.runtime` extension-author semantics are documented in `scripting/namespaces.md`.

Deferred:

- process extension command dispatch;
- live extension handler unload/reload. `/reload` reports when restart is needed.

## Clipboard/media surface

Implemented/adapted:

- `/copy` is a transcript-safe fallback by default that prints the last assistant message;
- tests lock that default `/copy` does not emit OSC 52 escape sequences;
- OSC 52 is available opt-in via `/copy --osc52` or persisted `tuiClipboardMode=osc52`;
- native clipboard helpers are available opt-in via `/copy --native` / `--auto` using dependency-light helper detection;
- ordinary terminal paste remains unchanged;
- bracketed paste remains parser/editor-dependent;
- image paste waits for a shared media ingestion contract.

See `tui-clipboard-media.md` for the detailed boundary.

## Validation

The final closure validation passed:

```bash
go test ./...
go vet ./...
```

Targeted validation was also run throughout the slices, primarily:

```bash
go test ./internal/tui ./internal/turn ./internal/web
go test ./internal/skills ./internal/tui ./internal/turn ./internal/web
```

## Acceptance criteria closure

This iteration is accepted because:

- current runtime/store/tool/SSE/topic contracts were preserved;
- each implementation slice added or updated focused unit tests, docs, or Gherkin captures;
- transcript output remains deterministic and tmux-friendly;
- missing Pi behavior is explicitly documented as adapted/deferred rather than silently implied;
- the current Go TUI stack remains in place, with no stack migration required.
