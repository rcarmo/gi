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

The same machinery (all/filtered/query + `filterModelMenuChoices`) now backs three selectors:

- **model selector** (`/model`, or `Ctrl-L` cycle fallback);
- **session selector** (`/sessions`): a searchable resume picker that lists sessions as `@agent title (id) · status`, filters by typed query, and switches to the chosen session on Enter;
- **thinking-level selector** (`/thinking` with no args): low/medium/high picker that sets the level on Enter.

The menu carries a `kind` (`model`|`session`|`thinking`) and an optional label→value map, so Enter dispatches to the right action (`/model <name>`, `switchSession(id)`, or `/thinking <level>`). `/model <name|index>`, `/resume <index|session_id>`, and `/thinking <level>` remain textual fallbacks for tmux/script use, so existing scripted flows are unaffected.

Future PiSwift-style selectors (tree/settings/theme) can reuse the same `kind`/values machinery. Each must keep a textual command fallback and live in the bottom overlay area, never as top chrome.
