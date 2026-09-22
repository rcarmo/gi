# ADR-0015: Share session-local model selection with the terminal

## Status

Accepted — 2026-09-22

## Context

The web validated model selections and persisted `selected_model`/`selected_provider`, but the TUI's `/model` accepted arbitrary strings and wrote workspace defaults. Startup and footer restoration also read runtime `model`, so an old turn's state could obscure an explicit user selection. The terminal already had a bounded selector; no additional pane was needed.

## Decision

Extract model validation, persistence and state resolution to `internal/inference/session_model.go`. Both web and terminal call `SelectSessionModel`. Unknown, ambiguous, unconfigured synthetic or unavailable/uncredentialled models fail before mutation. The deterministic test catalogue retains its explicit native shell model exceptions. Actual provider availability is established only by a later inference request.

`/model <name|index>`, Ctrl-L/Alt-L cycling and the model selector write only the current session. They do not create turns, modify another session or call `PersistModelSelection`. `/scoped-models` remains an explicit workspace catalogue/configuration command and retains its existing global persistence semantics.

Alt-M opens the existing model selector while preserving unsent text, cursor, undo/yank and history. `/model` remains the textual fallback. Catalogue rows use canonical provider/model labels, preventing an ID shared by multiple providers from becoming an ambiguous picker action. The initial terminal configuration supplies the catalogue's default-provider context; selecting another provider does not reinterpret unqualified enabled-model IDs.

A failed selection keeps the picker and old model. Its error replaces the existing search/help line instead of adding a row. Successful selection closes the picker, restores editor focus and updates the existing footer model segment. No transcript message is added by picker success. A stale idle-status string is updated with the model so it cannot become a new transient footer row; real running/error notices are retained.

Binding at startup and switching sessions restore `selected_model`/`selected_provider` before runtime fields. Footer/session summaries use the same resolver. Non-reasoning selections clear thinking state, including explicit empty-string restoration. A previous turn can finish with its captured model without replacing the durable selection.

## Evidence

- Shared inference tests cover labels/IDs, ambiguity, disabled/unknown/unavailable models, missing sessions, reopen persistence, session isolation and runtime-versus-selected metadata.
- Terminal tests cover text/index/menu/cycle selection, rejected or failed storage operations, startup restoration, footer values, editor state and byte-identical settings files.
- `make test-tui-sessions` uses actual tmux interaction at 60×18, 100×22 and 140×36: Alt-M open/filter/error/cancel/select, at most six rows, unchanged separator positions, preserved drafts, no transcript noise or model turns. Clean restart restores the main session's choice; an invalid command creates no work and a real subsequent turn uses the selected model. Workspace settings remain byte-identical.
- Existing TUI smoke and seven-file Gherkin suites pass. The local TUI unknown-model scenario now rejects `test-alt` and accepts configured `test/test-model`; frozen browser criteria are unchanged.
- Go tests/vet and three-run focused inference/TUI/web race tests pass. Browser regression: 168/168 parity executions and 70/70 functional tests. Browser coverage stays 14/236 Classic IDs, 222 unmapped, with all 42 shared cases unmapped.

## Limits

Terminal drafts remain process-local. Pending media staging, durable queue recovery/mutations, measured context compatibility, richer model filtering and interactive OAuth need separate work. No terminal sidebar, header, persistent picker, extra context row or other idle chrome is added.
