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
- The menu renders a `search:` line with a live cursor and match count, a kind-specific empty-state line, the current model marker, and one accented selected-row marker.
- The temporary menu has no box border. It uses two title/search rows plus at most six result rows, reduced to fit the terminal, editor and footer. Closed menus use zero rows.
- Rows are single-line, UTF-8-safe and bounded by terminal cell width. Rendering after resize keeps the selected entry visible.
- Selection/cancel reset all menu state (`modelMenuAll`, `modelMenuQuery`).

## Keys

- `Alt-M` opens the searchable model selector without replacing the unsent draft; `/model` is the command fallback.
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

- **model selector** (`Ctrl+L`, `Alt+M` or `/model`; `Ctrl+P`, `Shift+Ctrl+P` and `Alt+L` cycle);
- **session selector** (`Alt-S` or `/sessions`): a searchable resume picker that lists sessions as `@agent title (id) · status`, filters using full IDs, and switches on Enter. Alt-S preserves the unsent draft while opening. Escape restores the editor without changing its text, cursor or session;
- **thinking-level selector** (`/thinking` with no args): low/medium/high picker that sets the level on Enter.

The menu carries a `kind` (`model`|`session`|`thinking`) and an optional label→value map, so Enter dispatches to the right action (`/model <name>`, `switchSession(id)`, or `/thinking <level>`). `/model <name|index>`, `/resume <index|session_id>`, and `/thinking <level>` remain textual fallbacks for tmux/script use, so existing scripted flows are unaffected.

Session state and buffered events are isolated by session and selection generation. Per-session caches preserve editor text/cursor, undo/yank and history state. `make test-tui-sessions` verifies live draft round trips, exact cancellation footprint and resize behaviour at 60×18, 100×22 and 140×36. See [ADR-0010](../adr/0010-terminal-session-selection.md) for limits and tests.

Model choices use the shared native validator in `internal/inference/session_model.go`. Selection persists only the addressed session; `/model` and cycle keys no longer write workspace defaults. Invalid/unavailable choices keep the old model and show an error in the existing search line. Successful picker selection preserves the editor, adds no transcript message and updates only the existing footer. Startup, switch and footer reads prefer explicit selection over runtime model fields. `/scoped-models` opens Pi's scoped-models selector (below).

The live tmux harness also verifies model selection/error/cancel at all three sizes, byte-identical settings, clean-restart restoration and the next real turn's model. See [ADR-0015](../adr/0015-terminal-session-model-selection.md).

Future Pi selectors (tree/settings/theme) can reuse the same `kind`/values machinery. Each must keep a textual command fallback and live in the bottom overlay area, never as top chrome.

## Scoped models

A scoped model is an entry in an ordered shortlist for the TUI's model picker
and cycling shortcuts. It is a switching preference, not a new model type,
security restriction or automatic fallback policy. See [scoped models](scoped-models.md)
for a user-oriented guide, configuration example and persistence rules.

A bare `/scoped-models` opens Pi's `ScopedModelsSelectorComponent`. The code
is in `internal/tui/scoped_models.go`. Pi's golden is
`scripts/golden-scoped-models.mjs`, checked by `TestScopedModelsMatchPi`.

- The list shows the available models, with the enabled ones first and in
  their order.
- Enter toggles the selected model.
- Ctrl+A enables all models and Ctrl+X clears them; while a search is active,
  both apply only to the search results.
- Ctrl+P toggles every model from the selected model's provider.
- Alt+Up and Alt+Down reorder enabled models.
- Ctrl+S saves the selection as `enabledModels`. When every model is enabled,
  Ctrl+S removes the project setting instead; the next settings load then
  uses the user setting or built-in default shortlist. An explicit
  `enabledModels: []` in project settings persists all-model cycling.
- Typing searches with Pi's fuzzy filter. Ctrl+C clears the search, or cancels
  when the search is empty. Escape cancels.

Changes apply to the running TUI at once, as Pi's session scope does.
Closing the selector does not revert those changes; Ctrl+S is needed only to
persist them for future runs. If none of the enabled models is available, the
TUI uses the all-models scope. In the `/model` picker, Ctrl+A switches between
the scope and all models; an explicit `/model provider/model` is not restricted
to the shortlist.
`Ctrl+P` cycles through the enabled models that are available. When there is
no scope, it cycles through every available model. When only one model is
left, it shows Pi's "Only one model in scope" or "Only one model available".

The subcommands `/scoped-models list|add|remove|set` are kept for scripts.

Differences from Pi:

- Pi saves `enabledModels` to the global settings. gi saves it to the
  project's settings, as gi saves its other TUI settings.
- Pi refreshes the model catalogues while the selector is open and shows a
  status line for the refresh. gi lists the catalogue it has.
- gi's scope does not support Pi's patterns (globs and `:thinking` suffixes).
  Configured entries are resolved to model IDs.
