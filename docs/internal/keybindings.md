# Keybindings

Gi's TUI targets Pi 1.1.0 default keybindings (pi-coding-agent
`docs/keybindings.md`, `core/keybindings.js` and pi-tui's `keybindings.js`).
The current bounded alignment and remaining gaps are in
[Pi 1.1.0 alignment](../implementation/terminal/tui-pi-1.1.0.md). gi-only actions use keys
that Pi leaves free. `/hotkeys` prints Pi's own `/hotkeys` text, followed by a
table of gi's extra keys.

Code:

- `internal/tui/keybindings.go`: Pi's defaults that depend on the platform, and
  Pi's key names.
- `internal/tui/multiline_input.go`: the `tui.editor.*` and `tui.input.*` keys.
- `internal/tui/chat.go` (`KeyMap`) and `internal/tui/app_keys.go`: the `app.*`
  keys and the fullscreen `tui.altScreen.*` keys.
- `internal/tui/hotkeys.go`: `/hotkeys`.
- `third_party/go-tui/parse.go`: legacy terminals send Ctrl+\, Ctrl+] and
  Ctrl+- as the bytes 0x1c, 0x1d and 0x1f. go-tui now reads them as pi-tui
  does; it used to drop them.

## Pi's defaults

| Action | Keys |
|---|---|
| `app.interrupt` | Escape |
| `app.clear` | Ctrl+C: the first press clears the editor; a second press within 500 ms exits. With a transcript selection, Ctrl+C copies the selection instead (gi). |
| `app.exit` | Ctrl+D when the editor is empty; otherwise Ctrl+D deletes the next character (`tui.editor.deleteCharForward`) |
| `app.suspend` | Ctrl+Z (none on Windows) |
| `app.editor.external` | Ctrl+G opens the draft as `prompt.md` in `externalEditor`, then `$VISUAL`, then `$EDITOR`, falling back to nano (notepad on Windows) |
| `app.clipboard.pasteImage` | Ctrl+V attaches a clipboard image (gi's `/paste-image`) |
| `app.model.select` | Ctrl+L |
| `app.model.cycleForward` / `cycleBackward` | Ctrl+P / Shift+Ctrl+P |
| `app.thinking.cycle` | Shift+Tab |
| `app.thinking.toggle` | Ctrl+T hides thinking blocks behind an italic "Thinking..." label and saves `hideThinkingBlock` |
| `app.tools.expand` | Ctrl+O |
| `app.message.copy` | Ctrl+X copies the selection, or else the last assistant message |
| `app.message.followUp` / `dequeue` | Alt+Enter / Alt+Up |
| `tui.editor.*` | Pi's cursor, word, line, kill ring (Ctrl+W, Alt+Backspace, Alt+D, Alt+Delete, Ctrl+U, Ctrl+K), Ctrl+Y yank, Alt+Y yank-pop, Ctrl+- undo, Ctrl+] and Ctrl+Alt+] character jump |
| `tui.altScreen.*` | PageUp/PageDown, Ctrl+Home/Ctrl+End (top/bottom), Ctrl+Shift+Up/Down and Ctrl+Up/Down (previous/next prompt), Ctrl+Shift+F (search) |

### Windows and WSL

Pi switches some keys on Windows and WSL (`useWindowsKeybindings`: `win32`, or
Linux with `WSL_DISTRO_NAME` or `WSL_INTEROP` set), and gi does the same:

- Undo is Ctrl+Z on Windows and Alt+Z on WSL.
- Cycling models backwards is Alt+P.
- Follow-up is Ctrl+Q, and restoring queued messages is Alt+Q.
- Pasting an image is Alt+V.
- Transcript search is Ctrl+F.
- Previous and next prompt are Ctrl+Up and Ctrl+Down only.

Windows has no suspend key. On macOS, key names show Alt as Option, as Pi's
`formatKeyText` does.

## Editor semantics (pi-tui `Editor`)

- **Word moves and word kills** stop where Pi's `findWordBackward` and
  `findWordForward` stop: after whitespace, at a word, at a run of punctuation,
  or at a paste marker (treated as one unit). At the start or end of a line, a
  word move crosses the line break.
- **Kills** stay within the cursor's line. At a line edge, a kill removes the
  line break. Killed text goes to the kill ring. Consecutive kills accumulate:
  backward kills prepend, forward kills append.
- **Yank-pop** works only right after a yank. It replaces the yanked text with
  the previous kill and rotates the ring.
- **Undo** is a stack. A typed word is one undo unit, and each typed whitespace
  character is its own. `SetText` is undoable, as Pi's `setText` is. Submitting
  clears the history (`Reset`), and so does switching session drafts.
- **Character jump**: the next printable key is the target. Pressing the jump
  key again cancels the jump. When the target is not found, the cursor stays
  where it is.

Golden tests from Pi's own code:

- `scripts/golden-editor-keys.mjs` → `TestEditorKeysMatchPi`
- `scripts/golden-hotkeys.mjs` → `TestHotkeysMatchPi`, which checks Linux,
  macOS, WSL and Windows

`TestAppKeysMatchPiDefaults` covers the app keys.

## Differences

- gi has no `keybindings.json`. Pi's defaults are fixed.
- Word boundaries: Pi uses `Intl.Segmenter`. gi treats letters, digits, marks
  and `_` as word characters, and other non-space characters as punctuation.
  This matches Pi wherever the segmenter splits words at Pi's punctuation
  characters, which the golden tests cover.
- `hideThinkingBlock` is saved to the project's settings, as gi saves its other
  TUI settings. Pi saves it to the global settings.
- History recall (Up/Down) takes an undo snapshot at every step. Pi takes one
  when browsing starts.
- Home/End and Ctrl+A/Ctrl+E move to the editor's line start/end. In fullscreen,
  Ctrl+Home/Ctrl+End navigate the transcript without moving the editor cursor.
  In regular mode, Ctrl+Home/Ctrl+End do not move the editor cursor; terminal
  scrollback stays terminal-owned.
- The external editor's notice says "gi" where Pi's says "Pi".
- gi-only keys:
  - F6, F7 and F8 select and expand transcript blocks.
  - Ctrl+R searches prompt history.
  - F2 and F3 recall history.
  - Alt+L and Alt+T cycle models and thinking backwards.
  - Alt+M opens the model selector.
  - Alt+S opens sessions.
  - Alt+C compacts the session.
  - Alt+I opens the workspace index.
