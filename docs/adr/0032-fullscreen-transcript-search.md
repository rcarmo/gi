# ADR-0032: Fullscreen transcript search without extra idle rows

## Status

Accepted — 2026-09-22. A compact adaptation of Pi's fullscreen search and marked-message navigation. Fullscreen pointer selection/copy/edge autoscroll remain open.

## Interaction

Ctrl-Shift-F temporarily replaces the fullscreen editor with a separate search input. The existing upper separator carries the result count and key hints. Enter or Ctrl-G selects the next matching row; Shift-Enter or Ctrl-Shift-G selects the previous row. Navigation wraps. Escape, Ctrl-C or Ctrl-Shift-F closes search and restores the original editor, cursor, undo/yank state, scroll offset and follow mode. Search never calls submission or mutates prompt history.

Ctrl-Shift-Up/Down jump between rendered user-prompt starts without changing the editor. The keys follow Pi's Linux fullscreen defaults from `docs/keybindings.md` in installed Pi 0.85.1. Modified keys require a terminal/multiplexer that forwards them; native PTY tests send the corresponding CSI-u/modified-arrow sequences.

Regular mode retains terminal-owned history and search. This application search is fullscreen-only. No permanent search control, status row or side panel is added.

## Rendered-row index

The index is built from Gi's existing retained transcript renderer, including wrapping, current tool expansion and outcome styles. A temporary buffer extracts displayed cells and prompt starts. Overflow indexing accounts for go-tui's reserved scrollbar column. Tests compare indexed row count with the actual scroll region's rendered content height.

Matching is literal, case-insensitive substring matching within each rendered row. Matching rows receive an underlined background; the selected row reverses the foreground/background pair and becomes bold. Repeated occurrences on the same row count as one result. Collapsed-away tool output is excluded until expanded; metadata markers never become search text. A query spanning a soft-wrap boundary does not match across rows. This is a bounded rendered-transcript search, separate from the web database search.

The cached index invalidates on retained transcript changes, width/viewport changes or tool expansion state. Native arrivals refresh the active query. Session switching closes the search before saving/restoring editor snapshots. Row-position updates synchronise the current scroll element as well as the next-render state, preventing navigation from reading a stale offset.

## Evidence

`make test-tui-search` runs real tmux PTYs at 60×18, 100×22 and 140×36. Twenty-four native shell turns establish history, and a gated native turn delivers output during search. The tests verify:

- literal/case-insensitive and Unicode searches, highlighting, next/previous shortcuts and no-match feedback;
- no query submissions or accidental draft sends;
- query/results across resize, reopen and live output;
- collapsed tool text excluded, expanded text included;
- original draft/cursor and reading offset restored on Escape;
- prompt navigation while the editor is focused;
- no added idle rows.

Unit tests cover saved editor undo/yank state, session ownership, visible versus hidden output, Unicode, rendered cell highlights, wrapped prompt positions, actual rendered scroll height and regular-mode exclusion. Initial live runs exposed focus cycling that blurred a single editor and fixture key sequences sent before the previous frame changed ownership; the final tests await the real view transition.

Full Go tests/vet, terminal race suite repeated three times, all regular/outcome/reading/session/model/compaction/smoke/Gherkin suites, 29 helpers and 70/70 functional browser tests pass. Fresh real 100×22 XTerm search/restored-draft screenshots are attached. Evidence is under `test-results/tui-search/` and `/workspace/tmp/gi-ui-captures/september22/`.

The read-only delegate could not select an approved executable model and supplied no review evidence. Frozen browser criteria and web components are unchanged; browser mappings remain 34/236 Classic and 2/42 shared (202/40 unmapped), with the prior full 384/384 matrix at `4071c01`.

## Remaining differences

This implementation highlights matching rows rather than individual occurrences and has no clickable search arrows. Cross-row search, stronger reader anchors across reflow/eviction/above-viewport edits, long-history performance, light-theme support and fullscreen pointer selection/copy/edge autoscroll need separate work. Existing retention limits bound the searchable transcript. No parity credit is assigned to those paths.
