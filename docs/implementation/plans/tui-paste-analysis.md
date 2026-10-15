# TUI paste / clipboard analysis

Date: 2026-05-05

## Goal

Assess what “copy/paste using ANSI extended sequences” should mean for gi's TUI, using Pi as the reference where possible.

## Evidence inspected

### Pi docs

- `docs/usage.md`
  - mentions multiline input via `Shift+Enter`
  - mentions image paste with `Ctrl+V` (`Alt+V` on Windows)
- `docs/terminal-setup.md`
  - documents Kitty keyboard protocol and modified Enter forwarding
  - does **not** document bracketed text paste (`CSI 200~` / `CSI 201~`)
- `docs/keybindings.md`
  - documents clipboard/image-related bindings
- `docs/extensions.md`
  - documents `ctx.ui.pasteToEditor()` and `setEditorText()`
- `docs/rpc.md`
  - states `pasteToEditor()` degrades to `setEditorText()` in RPC mode

### Pi installed runtime

Inspected:

- `dist/utils/clipboard.js`
  - copy path prefers native clipboard helpers and OSC 52 fallback
- `dist/utils/clipboard-image.js`
  - image paste is implemented through OS clipboard integrations
- `dist/modes/interactive/components/custom-editor.js`
  - explicit keybinding support exists for `app.clipboard.pasteImage`

Observed:

- strong evidence for **copy to clipboard**
- strong evidence for **image paste from clipboard**
- no clear evidence of dedicated **bracketed text paste parsing** in the terminal input path

### gi terminal stack

Inspected:

- `github.com/grindlemire/go-tui` docs and parser surface
- existing `gi` TUI implementation and key handling

Observed:

- Kitty keyboard protocol support exists for modified keys
- I did **not** find bracketed-paste / ANSI paste sequence handling (`CSI 200~`, `CSI 201~`) in `go-tui`
- gi currently accepts ordinary pasted text only insofar as the terminal emits it as normal rune input

## Conclusion

Pi's practical clipboard model appears to be:

1. **text copy** via native clipboard helpers and OSC 52 fallback
2. **image paste** via OS clipboard integration + explicit keybinding
3. **editor injection** through `pasteToEditor()` / `setEditorText()` APIs
4. **modified keys** via Kitty keyboard protocol

What I could **not** confirm from the available Pi docs/runtime evidence:

- a dedicated bracketed text paste implementation based on ANSI paste delimiters

So for gi, the exact checklist wording “Copy/Paste using ANSI extended sequences” needs to be interpreted carefully:

- **ordinary text paste** is mostly terminal-driven today
- **ANSI/bracketed text paste parity** looks blocked on deeper parser/editor support in `go-tui`
- **clipboard copy** and **image paste** would require separate runtime integrations, not just key parsing

## Recommendation for gi

Short term:

- treat this checklist item as **analysis complete, implementation deferred**
- document that true bracketed-paste parity depends on `go-tui` parser/editor support or a richer editor component

Medium term options:

1. Extend `go-tui` to recognize bracketed paste start/end sequences and surface them as paste events.
2. Add a gi-side editor API that can accept pasted blocks distinctly from per-rune typing.
3. Add clipboard helpers for:
   - copy last assistant message (`OSC 52` + native helpers)
   - optional image paste path if/when gi needs terminal-side image upload parity

## Why this closes the current plan item

The checklist item explicitly included “analyze Pi's approach”. That analysis is now complete, evidence-backed, and it clarifies that the missing piece is not just local TUI polish inside gi.
