# Bounded terminal editor viewport

Long terminal drafts now scroll inside a bounded, cursor-following editor window. The full draft stays in memory and is submitted unchanged. Empty and short drafts retain their existing footprint; no counter, scrollbar or permanent hint is added.

## Height and wrapping

The installed pi-tui editor uses `max(5, floor(terminalRows * 0.3))` visible lines and adjusts a scroll offset to keep the cursor in view. Gi uses that cap, then clamps to space left by its existing footer, widgets, separators, picker and transcript/history reserve. Fullscreen reserves four transcript rows; regular mode reserves at least one terminal-owned history row. Very small terminals can have fewer than five editor rows.

`editor_viewport.go` shares the budget between fullscreen and regular layouts. Model-menu sizing uses a bounded editor estimate so a large draft cannot crowd out all picker entries. The actual editor budget is recomputed after menu height is known. Search uses its own editor instance and scroll offset with the same bounded height.

`multiline_input.go` wraps grapheme clusters by display-cell width instead of rune count. Rendered rows disable a second word-wrap pass. The cursor remains rune-indexed for existing editing commands; if its index falls inside a combining cluster, the marker is displayed before the intact cluster. A two-cell grapheme in a one-column viewport uses a replacement display cell, without changing the draft. CRLF remains one visible line break.

The viewport neither truncates storage nor imposes a new paste-size limit. The [layout-cache follow-up](tui-editor-layout-cache.md) reuses unchanged redraw layouts. Changes to text, cursor, width or focus still lay out the whole in-memory draft; large-input editing latency and richer grapheme-aware editing remain separate concerns. Extremely short terminals with more fixed widgets than available rows are not covered by a universal layout guarantee.

## Ownership

Home/end editing, insertion, undo/yank, submission and queued/session drafts retain their existing paths. Scrolling changes presentation state only. Opening or closing a selector recomputes the visible budget, and the cursor is restored in view. Regular-mode history stays terminal-owned; resize notices and earlier dock output may remain in scrollback. No renderer swap, mouse capture or history clear is introduced.

## Verification

- `make test vet bun-checks` passes. New unit cases cover1/2/9/60/100/140columns, long Unicode and explicit-newline drafts, cursor positions inside combining clusters, home/end, smaller budgets, empty shrink, independent editor/search offsets, picker bounds and regular history reserve. Rendering preserves text, cursor, undo and yank values.
- `make test-tui-editor-viewport`: six fullscreen/regular tmux PTYs at60×18,100×22,140×36 pass. Each enters32long Unicode lines, edits/restores head and tail, opens/closes Alt-M, resizes to42×14 and back, verifies zero navigation submissions, then explicitly submits once and compares exact stored user-message bytes. The editor returns to one idle row.
- Model picker, search and regular-mode regression targets pass another12PTY configurations, for18total. Existing draft/cursor/session/selection/scrollback and idle-dock assertions remain.
- `make test-ux`:107passes,11existing fixture-dependent skips. Browser code is unchanged.
- Initial PTY failure counted old regular-mode separators retained in terminal scrollback; it now measures the active dock's last two separators. No test removed or timeout increased. A read-only delegated review timed out without a result.

Artifacts are under `test-results/tui-editor-viewport`, `tui-model-picker`, `tui-search` and `tui-regular`. Physical/emulator-wide acceptance, broader Markdown/link/selection limits, light theme and stable reflow anchors remain open. Whole CI and deployment are separate gates; live sessions were not used for testing.
