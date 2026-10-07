# TUI Pi fit/gap roadmap

## Status

This is the closure/acceptance note for the Pi-like `gi -tui` iteration. The iteration made the TUI progressively feel closer to Pi while preserving Gi's SQLite-backed runtime architecture, current Go TUI stack, runtime/tool/SSE/topic contracts, and tmux-friendly rendering.

It intentionally does **not** require pixel-perfect cloning. Parity means the same workflow is discoverable, test-backed, deterministic in terminal captures, and adapted to Gi-specific runtime semantics.

## Current `gi -tui` command surface

The slash catalogue (`piCommands` + `giCommands` in `internal/tui/chat.go`)
lists Pi's built-ins first, in Pi's `BUILTIN_SLASH_COMMANDS` order with Pi's
argument hints and descriptions where gi behaves the same, then gi's own
commands (the way Pi lists extension commands after its built-ins). Aligned
behaviour: `/resume` with no argument opens the session selector (`/sessions`
remains as an alias), `/name` with no argument shows the current name,
`/logout` with no argument lists stored providers, and `/quit` exits.

`/export [path]` writes Pi's HTML page or Pi session JSONL (loadable by Pi;
`docs/internal/session-import.md`),
`/compact [instructions]` passes the focus to the before-compact hook and the
summary, and `/fork` opens Pi's "Fork from Message" selector (history before
the chosen message goes to a new session for the same agent; the message
returns to the editor). `/clone` copies all history for the same agent. Both
copies keep the source channel/account and get distinct chat identities.
gi's former `/fork @agentN` peer session is now `/spawn [@agentN]` (the old
form still works).

Pi built-ins not in gi yet (omitted, not approximated): `/bug`,
`/changelog`, `/trust`. `/import` reads Pi session files and `/share`
uploads one as a secret gist (docs/internal/session-import.md). `/tree`, `/settings`, `/scoped-models` and
`/login` open Pi's interactive UIs; `/tree` navigates gi's branch sessions
(docs/internal/tui-tree.md).

Implemented commands:

- `/help`
- `/commands [query]` / `/palette [query]`
- `/session`
- `/new`
- `/name [name]`
- `/resume [index|session_id]` (no argument: selector)
- `/quit` / `/exit`
- `/clone [@agentN]`
- `/copy [--osc52|--native|--auto|--fallback]`
- `/reload`
- `/tools [query|active|activate|reset]`
- `/skills [query]`
- `/skill:name [args]`
- `/model [provider/model]`
- `/scoped-models [list|add|remove|set]`
- `/thinking [level]`
- `/compact`
- `/scrollback [n]`
- `/settings` / `/config`
- `/approvals`
- `/abort` — abort the running turn like Escape; also finalizes a turn left `running` by a crash or killed process (session looks busy, `/compact` unavailable) instead of replaying it on the next start
- `/cancel`
- `/agents`
- `/plugins` / `/extensions`
- `/tree` (Pi's session tree navigator)
- `/fork` (Pi selector) · `/spawn [@agentN]` (peer session)
- `/export [path]`
- `/switch @agent|session_id`
- `/send @agent message`
- `/where`
- `!cmd` and `!!cmd` shell shortcuts

Still intentionally absent or deferred:

- `/changelog` dedicated informational command (`/hotkeys` is implemented);
- process-extension command dispatch (JS/Joker extension command dispatch is implemented);
- additional rich forms/widgets beyond the searchable model/session/thinking selectors.

## Current keyboard/editor surface

Implemented in `internal/tui/multiline_input.go` and `internal/tui/chat.go`:

- Enter submits;
- Alt+Enter uses the same submit path and is documented as the queue-friendly submit chord;
- Shift+Enter inserts newline;
- Escape blurs input;
- Backspace/Delete delete characters;
- Pi's default `tui.editor.*` and `app.*` keys, including the kill ring,
  yank-pop, undo stack and character jump (see `keybindings.md`);
- Tab completes workspace-relative paths;
- textual `@path` completion uses the same Tab fallback;
- Up/Down and F2/F3 history paths are covered;
- PgUp/PgDn and Home/End transcript scrolling are covered;
- Alt+L/Alt+T cycle models/thinking backwards (gi);
- focus/blur, mouse focus, resize, and quit behavior have tests/features.

Deferred/adapted editor items:

- configurable keybindings are not implemented;
- Ctrl+O tool collapse is not implemented (transcript blocks use F6/F7 selection and F8 expand/collapse);
- Ctrl+T is adapted for thinking-level cycling rather than a Pi-style thinking collapse UI;
- ordinary text paste remains terminal-rune behavior;
- `/paste-image [prompt]` provides explicit command-driven clipboard image ingestion;
- parser-level bracketed paste, raw Ctrl-V image payloads, and inline terminal image protocols remain deferred.

## Current session/runtime UX surface

Implemented:

- session/agent context summary through `/where` and persistent context area;
- detailed `/session` command with queue/steering counts and active turn state;
- `/new`, `/name`, `/resume`, `/clone`, `/copy`, and `/reload` command/session workflow affordances;
- Pi's session tree navigator through `/tree` (branches, labels, branch summaries);
- peer session creation via `/spawn` (legacy `/fork @agentN` still works);
- session switching via `/switch`;
- peer message sending via `/send`;
- topic-native status rendering for runtime turn/tool/hook/routing/session/inbound/dispatcher events;
- durable same-session steering in the runtime underneath the TUI;
- visible queued/steering counts in the single bottom status line;
- visible transcript/status feedback when a message is queued during an active turn;
- Alt+Up local queued-draft restore.

Adapted/deferred:

- Pi-style follow-up vs steering split is documented as an adapted gap: Gi currently routes active-session submissions through existing steering/queue semantics;

## Current transcript/layout surface

Implemented:

- Pi-identical steady-state row order: transcript, separator, editor row, separator, path/branch row, single status row;
- no top chrome or permanent context block;
- transient runtime notices routed to the single bottom status line;
- compact `/help` with detailed discovery delegated to `/commands`, `/session`, `/where`, and `/settings`;
- textual `/commands [query]` palette plus searchable model/session/thinking selectors;
- grouped `/settings` with runtime, model, editor, session, discovery, compaction, and peering sections;
- denser `/resume`, `/settings`, and model-selection output for narrow terminals;
- editor bindings for word movement, word deletion, line deletion, minimal undo/yank, `!`/`!!` shell shortcuts, Tab path completion, and textual `@path` completion;
- terminal-safe Markdown rendering for headings, lists, blockquotes, links, code blocks, and responsive table fallback;
- folded multi-line tool results;
- code block line counts;
- compact runtime compaction summaries;
- tmux/Gherkin-friendly text captures under `features/tui/`.

Remaining polish:

- dedicated thinking-collapse behavior beyond transcript block expansion;
- richer form widgets only if they can be added without replacing the current stack.

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
