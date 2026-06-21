# Gi TUI searchable selectors (PiSwift port)

Status: searchable model selector implemented; pattern reusable for further selectors.

## PiSwift reference

`Sources/PiSwiftCodingAgentTui/Modes/Interactive/Components/ModelSelectorComponent.swift`:

- search input at top;
- fuzzy filter over `id provider`;
- up/down navigation with wrap;
- provider badge and current-model checkmark;
- scrolling window;
- enter selects, escape cancels.

## Gi implementation

`internal/tui/chat.go` model menu (`/model` with no args, or the model menu open state):

- `modelMenuAll` holds the full choice list; `modelMenuChoices` holds the filtered view; `modelMenuQuery` holds the live query.
- Typing in the open menu appends to the query (`modelMenuTypeRune`); backspace edits it (`modelMenuBackspace`).
- `filterModelMenuChoices(all, query)` + `fuzzyMatch(query, candidate)` filter by case-insensitive, whitespace-separated substring tokens, so `gpt` matches only gpt models and `claude sonnet` matches `anthropic/claude-sonnet`.
- The menu renders a `search:` line with a live cursor and match count, a `no matching models` line when empty, the current model marker, and the selected-row highlight.
- Selection/cancel reset all menu state (`modelMenuAll`, `modelMenuQuery`).

## Keys

- `/model` opens the searchable selector.
- type to filter; Backspace edits the query.
- Up/Down/PageUp/PageDown/Home/End navigate.
- Enter selects; Esc cancels.
- `Ctrl-L`/`Alt-L` still cycle enabled models without opening the selector.
- `/model <name|index>` remains a textual fallback for tmux/script use.

## Tests

`internal/tui/chat_test.go`:

- `TestModelMenuFuzzyFilter` covers empty/substring/multi-token/no-match filtering.
- `TestModelMenuTypeAndBackspaceFiltersChoices` covers live typing and backspace restoring the full list.

## Reuse

The same machinery (all/filtered/query + `filterModelMenuChoices`) is the basis for future PiSwift-style selectors (session/tree/settings). Each must keep a textual command fallback and live in the bottom overlay area, never as top chrome.
