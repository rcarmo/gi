# PiSwift TUI port status

Date: 2026-06-21

This tracks how much of PiSwift's TUI look/feel and component model has been ported into Gi, preserving Gi's SQLite/WAL runtime truth and the no-top-chrome Pi layout.

Reference comparison: `piswift-tui-comparison-20260621.md`.

## Ported in this effort

### Component-style transcript model
- transcript blocks with metadata (kind/status/timestamps/footer), append/replace/delete/select/toggle lifecycle, and expand/collapse.
- `F6`/`F7` select blocks, `F8` toggles, click toggles.
- Tests: lifecycle, selection, toggle, runtime-event updates, dedup error blocks.

### Live bash output block (PiSwift `BashExecutionComponent`)
- `!!command` renders a bordered `$ command` block.
- collapsed = trailing preview (last 10 lines); expanded = full output.
- skipped-line hint with `F8 expand`; `F8 collapse` when expanded.
- full output written to `<workspace>/.gi-run/bash-output/bash-*.txt` with a truncation footer beyond 500 lines.
- Tests: meta/body, tail preview window, full-body retention/expand.

### Expandable bottom footer (PiSwift `FooterComponent`)
- the bottom band can grow beyond one line.
- path/branch row; stats row (counts + token usage + cache read/write + cost + context) with model/thinking right-aligned; optional transient notice row.
- Tests: usage parts (tokens, cache, cost) and transient notice row.

### Searchable selectors (PiSwift `ModelSelectorComponent`)
- `/model` opens a searchable model selector (type to filter, substring/multi-token, match count, navigation, current marker).
- `/sessions` opens a searchable session resume selector reusing the same machinery (kind + label→value map).
- textual fallbacks retained: `/model <name|index>`, `/resume <index|session_id>`.
- Tests: fuzzy/substring filter, live typing/backspace, session open/filter/switch.

### TUI extension slots (PiSwift hook UI context)
- `extension.status` topic → keyed footer status segments.
- `extension.widget` topic → keyed multi-line widget between transcript and editor.
- `extension.tool_render` topic → per-tool body render mode (full/compact/hidden).
- Tests prove the slots add only bottom-band/widget rows or change existing tool blocks, never writing transcript rows, so they cannot create top chrome.

## Captured evidence

`artifacts/tui-piswift-port-20260621/`:

- `01-startup.txt` — no top chrome, transcript-first, expandable footer.
- `02-bash-collapsed.txt` — bash block tail preview + skipped-line hint.
- `03-bash-expanded.txt` — full bash output after `F8`.
- `04-model-search.txt` — searchable model selector filtered by `gpt`.
- `05-session-selector.txt` — searchable session selector.

## Not yet ported (deferred)

- editor-replacement extension slot.
- inline terminal image rendering / Ctrl-V image paste (Gi uses `/attach` + media store).
- theme/OAuth/settings selector components.
- configurable keybinding map (Gi shortcuts remain mostly fixed).

## Constraints preserved

- No top chrome; transcript starts at row 0.
- Extension slots only add to the bottom band, the widget area, or the transcript.
- SQLite/WAL runtime truth and existing API/SSE/topic/tool contracts unchanged.
