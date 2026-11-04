# Scoped models

In gi, a **scoped model** is a model in your ordered shortlist of enabled
models. The scope makes switching between a few preferred provider/model
pairs convenient, instead of cycling through the whole catalogue. It is not
a separate kind of AI model, a permission boundary, or an automatic fallback
or routing policy.

## Default shortlist

When neither project nor user settings specify `enabledModels`, gi starts with
this ordered shortlist from its pinned model catalogue:

| Family | Default model reference |
|---|---|
| GPT Sol | `github-copilot/gpt-6.1-sol` |
| GPT Luna | `github-copilot/gpt-6-luna` |
| Claude Opus | `github-copilot/claude-opus-5.5` |

These are the latest entries of each family in the pinned catalogue, not
automatically resolved aliases. Explicit project settings override user
settings, which override this built-in list. `"enabledModels": []` explicitly
disables the shortlist and enables all-model cycling. Models still require
credentials and account availability; unavailable entries are skipped. The
shortlist does not change the configured startup model or grant Copilot access.

## Choose and use a scope

1. Run `/scoped-models` in the terminal to open **Model Configuration**.
2. Type to filter; use Up/Down to move and Enter to toggle a model.
3. Use Alt+Up/Down to reorder enabled models. Their order is the cycling order.
4. Changes take effect in the running TUI immediately. Escape closes the
   selector; it does **not** undo the already-applied in-memory changes.
5. Press Ctrl+S to persist the selection for future runs.

Outside the selector, Ctrl+P cycles forward through available models in the
scope (using the default keybindings). `/model` opens the model picker; when a
scope is available, Ctrl+A in that picker switches between the scoped and all
models views. `/model provider/model` can still select another available
model directly. Selecting a model changes the addressed session, not every
session or the workspace default; explicitly saving a default is separate.

Inside `/scoped-models`, Ctrl+P has a different meaning: toggle all models
from the highlighted model's provider. Ctrl+A enables all models and Ctrl+X
clears the selection; when filtering, those two shortcuts affect only the
search results. Ctrl+C clears a search, or closes the selector if it is empty.

Only available models participate in cycling. With no effective scope (no
effective shortlist, every model enabled, or no available enabled model),
cycling falls back to all available models. A one-model list cannot cycle and
shows an informational notice. A scope does not provide credentials or make
an unavailable provider usable.

## Persistence

Ctrl+S in `/scoped-models` saves the ordered `enabledModels` array in the
workspace's settings file. Gi updates `.gi/settings.json` if that is the
project settings file in use, otherwise `.pi/settings.json`; when neither
exists it creates `.pi/settings.json`. See [configuration lookup](config-files.md).
Enabling every model and saving removes the project `enabledModels` key.
On the next settings load, removing the project key restores the user setting
or built-in default shortlist. To persist all-model cycling regardless of those
defaults, set `"enabledModels": []` explicitly in project settings.

For example, assuming these models are available to your account:

```json
{
  "enabledModels": [
    "github-copilot/gpt-6.1-sol",
    "github-copilot/gpt-6-luna",
    "github-copilot/claude-opus-5.5"
  ]
}
```

This controls the shortlist, not the current model or automatic provider
selection. To select the first model now, use
`/model github-copilot/gpt-6.1-sol`.

The textual commands `/scoped-models list`, `add <model> [model...]`,
`remove <model|index>` (one-based index), and `set <model> [model...]` are also
supported. Unlike unsaved selector edits, these legacy commands persist
immediately through the model-settings writer; they can also adjust the
configured default if it is no longer in the list. Use the selector for
session-only changes.

Gi accepts explicit model references, not Pi's glob patterns or `:thinking`
suffixes. The scope is a TUI switching preference, not a restriction on web,
API, sub-turn or explicit model selection. For selector key details and Pi
compatibility limits, see [TUI selectors](tui-selectors.md#scoped-models).
