# Using Gi's terminal UI

Gi is a single Go binary. It normally opens the TUI; `gi -web` opens the web
server instead. Run `gi -h` for actual CLI defaults and paths. No Bun/Node runtime
is required. Implementation status and known differences are recorded in the
[terminal alignment audit](tui-pi-1.5-alignment.md).

These documents ship in the binary. An assistant can read this page with
`read({"path":"vfs://reference/tui.md"})` and browse the
[reference index](README.md). `vfs://` URLs are virtual read-tool paths, not shell
filenames. Relative documentation links resolve under `vfs://reference/`, never
under the user's current workspace.

## Start, navigate and exit

- Use `gi -workspace /path/to/project` to select a workspace. `-db /path/to/gi.db`
  selects a SQLite store; preserve it to retain sessions and history.
- `-tui-mode fullscreen` uses an application-owned transcript with mouse
  selection and search. `-tui-mode regular` leaves retained output/history to
  the terminal and keeps the editor dock inline. Pick a mode with `/settings`.
- Enter sends the draft; Shift+Enter inserts a newline. Ctrl+G opens the external
  editor. Up/Down navigate multiline rows, then input history at the boundaries.
- Ctrl+C clears the draft; a second Ctrl+C within 500 ms exits. Ctrl+D exits
  only with an empty draft. Escape interrupts active work; it is not an implicit
  retry or resend. Ctrl+Z suspends on supported Unix terminals.
- `/help`, `/commands` and `/hotkeys` provide the built-in command/key inventory.
  PageUp/PageDown page an open selector/autocomplete before the transcript.
  In regular mode the terminal owns transcript scrollback.

See [keybindings](keybindings.md) for all platform variants, editor kill/yank/
undo semantics and Gi-only shortcuts. macOS calls Alt “Option”; terminals must
send the corresponding Meta sequence. WSL/Windows use different follow-up,
dequeue, clipboard, undo and reverse-model shortcuts.

## Models and reasoning

`/login` configures providers; `/logout` removes the selected saved credential.
`/model` (Ctrl+L or Alt+M) opens a searchable selector. `/model provider/id`
selects explicitly. Enter selects for the session; Ctrl+S also saves the default.
Models with missing credentials or inadequate measured context are not admitted.
Friendly names appear when they add information to the model ID.

`/scoped-models` manages an ordered cycling shortlist, **not** a model type,
security restriction or automatic failover policy. Enter toggles, Alt+Up/Down
reorder **inside this selector**, Ctrl+S saves; Ctrl+P outside it cycles the scope.
See [scoped models](scoped-models.md) for persistence/defaults and `/model` scope.
`/thinking` selects the reasoning level, Shift+Tab cycles it, Ctrl+T toggles the
visibility of thinking text. Actual levels depend on the selected model.

## Pending messages and safe editing

While a response is running, Enter submits steering and Alt+Enter submits a
follow-up. Windows/WSL use Ctrl+Q for follow-ups. The pending panel distinguishes
Steering and Follow-up. **Alt+Up** (Alt+Q on Windows/WSL) withdraws all eligible
still-pending text to the editor in that order, ahead of the current draft. It
works even after the input loses focus and does not stop the running response.
Edit and submit again only when ready; retrieval never submits automatically.

Gi saves the draft and withdraws pending delivery atomically. If a worker has
already claimed a message, the draft changed elsewhere, an attachment is present
or maintenance work cannot be represented as text, it reports a conflict rather
than dropping data. Already delivered steering is not pending and cannot be
recalled. The durable draft survives session switches/reopen. `/queue` inspects
pending work; explicit remove/steer/reorder operations require current full IDs.
See [queue commands](tui-queue-commands.md), [text journal](tui-text-journal.md)
and [pending attachments](tui-media-journal.md).

Escape has a separate interrupt-and-restore path with stricter run-ownership
guards. Held/interrupted work is not automatically replayed: inspect `/retry`
and [held retry](tui-held-retry.md) before choosing an explicit recovery action.
Never resend just because a response is slow or the UI reconnects.

## Sessions, transcript and attachments

- `/new` starts a session. `/resume` or `/sessions` opens the session picker;
  per-session drafts/model choices are retained. `/name` renames the session.
- `/tree` navigates the session tree; `/fork` selects an earlier user message
  for a branch. See [tree](tui-tree.md) and [session import/export](session-import.md).
- Ctrl+O expands/collapses tool output in fullscreen; regular retained output
  cannot be rewritten in place. Tool bands show pending/success/error colors
  with consistent padding and uncolored separation from the next message.
- In fullscreen, Ctrl+Shift+F searches rendered text (Ctrl+F on Windows/WSL).
  Drag selects; double-click selects words/path tokens; triple-click selects
  a line. Ctrl+X copies the selection, otherwise the latest assistant source.
  `/copy` copies latest assistant source. Clipboard policy is explicit and
  respected; do not assume that printing text copied it to the OS clipboard.
- `/attach`, `/attachments`, `/detach` and `/paste-image` manage media.
  Ctrl+V attaches a clipboard image (Alt+V on Windows/WSL). See
  [clipboard/media](tui-clipboard-media.md) and [paste](tui-paste.md).

## Settings, tools and diagnostics

`/settings` changes supported terminal settings and themes; `/config` displays
runtime settings. [Configuration lookup](config-files.md) explains `.gi` versus
`.pi`, user/project precedence, `settings.json`, `models.json`, authentication,
transport and retry configuration. Secrets belong in credential storage, not
chat, logs or examples. Gi does not implement every Pi setting or extension API.

`/mcp` manages configured MCP servers; project MCP trust is not implemented and
untrusted project configuration is ignored. `/tools` inspects active tools;
`/skills` lists discovered skills. See [MCP](mcp.md), [tools](tools/README.md),
[scripting](scripting/README.md) and [codemode](codemode.md). Gi scripting helpers
are documented APIs, not Pi's TypeScript extension API.

`/compact [instructions]` requests a context summary; `/compact info` inspects
policy. Cache read/write counts and CH% come from provider usage, not from a local
answer cache. See [compaction](compaction.md) and [prompt caching](prompt-caching.md).
`/bug preview` provides bounded local diagnostic information; `/changelog`
displays bundled notes. Neither automatically uploads transcripts or secrets.
