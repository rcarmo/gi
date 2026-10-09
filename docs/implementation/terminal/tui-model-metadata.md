# Inline model context in Alt-M

Alt-M shows a known model context window on its existing row when the full model key and suffix fit: `provider/model · 32K ctx`. Narrow rows keep the key without the suffix. Unavailable rows retain their existing reason; unknown context has no label.

`model_picker.go` captures the context label alongside the existing search snapshot. Rendering uses `selectorText` and terminal-cell widths, including the selection marker and numeric prefix. This is advisory catalogue metadata. Selection still revalidates credentials and measured context through the native session-model path.

## Terminal adaptation

The installed pi-tui `SelectList` uses bounded single-line entries, optional descriptions, an accent selection marker, arrows, Enter and Escape. Its `docs/tui.md` guidance keeps default views compact, calculates visible column widths and leaves regular-mode scrollback to the terminal. Gi retains its existing Go selector and keyboard model: six visible results, search terms, enabled-only navigation, explicit acceptance and cancellation. No extra heading, help line, footer, border, permanent panel or idle row is added.

Gi's existing regular-mode selector uses a temporary alternate screen; this change does not alter that lifecycle. It does not implement model pinning, pricing, thinking mutations, a theme switch or a new Settings surface. Context and reasoning remain searchable; only context receives visible inline text in this slice.

## Verification

On 2026-09-26:

- `make test vet` passes. Unit tests cover full key priority, exact-width suffix admission, omitted unknown/blocked metadata, Unicode labels and zero/narrow/wide widths. The shared renderer does not leak model metadata into session selectors.
- `make test-tui-model-picker` passes in six real tmux PTYs: fullscreen/regular at 60×18, 100×22 and 140×36. Captures include the inline context, blocked-reason priority, metadata filtering, resize, native model persistence, draft/cursor restoration, unchanged global settings and other-session state, and unchanged idle rows. Regular-mode checks retain prior scrollback and exclude the temporary picker from history.
- `make test-tui-session-picker BIN_DIR=/tmp/gi-tui-regression-bin` also passes all six fullscreen/regular PTYs. The shared selector renderer preserves session navigation, Unicode drafts/cursors, temporary-screen history and idle rows.
- An initial test compared whole editor structs, including function fields, and failed. It now compares the relevant draft, cursor, undo, yank and focus fields. The existing open/close lifecycle intentionally blurs and restores the editor.

Text/ANSI captures are in `test-results/tui-model-picker`. They verify these renderer/PTY configurations, not every terminal emulator or physical-device acceptance. Broader queue/media persistence, search/reflow/link/selection and light-theme gaps remain in the parity ledger.
