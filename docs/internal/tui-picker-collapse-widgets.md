# TUI picker and collapse widget audit

Status: initial audit complete. Gi keeps textual picker fallbacks as the primary implementation and treats richer modal/collapse widgets as incremental, opt-in rendering helpers only when they remain deterministic under tmux/script captures.

## Current Go TUI capability

Gi currently uses `github.com/grindlemire/go-tui` with `WithLegacyKeyboard()` and a small set of app/root/input key maps.

Observed capabilities that are safe today:

- deterministic static layout composition through existing panes/components;
- focused multiline input with rune/key handling;
- app/root-level preempt key bindings (`Ctrl+L`, `Alt+L`, `Ctrl+T`, `Alt+T`, history recall, scroll keys);
- transcript rendering as plain lines, which is stable under tmux, `script(1)`, and test snapshots;
- textual command output that can act as a picker when entries are numbered.

Observed constraints:

- no current first-class modal/list-select abstraction is used by Gi;
- keyboard handling is already dense, so new global picker keys risk conflicts with input editing, history recall, scroll, and copy/paste behavior;
- terminal capture workflows need all interactive surfaces to have a textual command equivalent;
- collapsing already-rendered transcript regions requires stable message grouping/state metadata, not only a local visual toggle, otherwise transcript and durable history drift.

## Picker policy

For the deferred parity work, Gi will prefer textual pickers first:

- list choices as numbered transcript lines;
- accept `/command <index>` or `/command <name>`;
- keep the current command usable in non-interactive captures;
- optionally add a richer modal later only if it wraps the same deterministic state helpers.

This means existing model selection already satisfies the first safe model-picker prototype:

```text
/model
/model 2
/model provider/model-name
Ctrl+L      # cycle forward
Alt+L       # cycle backward
```

The model list is deterministic, keyboard-friendly, and has textual fallbacks. It should remain the baseline even if a future modal/list overlay is added.

The command picker baseline is similarly textual:

```text
/commands
/commands media
/palette media
```

It includes built-in commands and registered extension commands, and therefore remains stable in tmux captures.

## Collapse widget policy

Tool-result/thinking collapse remains deferred until a small transcript grouping model exists. Requirements before implementation:

- render helpers can identify a transcript entry as thinking/tool/result/status without parsing arbitrary display strings;
- toggles do not conflict with existing keybindings;
- collapsed state is UI-local and never rewrites stored messages/events;
- textual fallbacks can show the same data with `/turns`, `/session`, `/commands`, or future `/view` commands.

## Current decision

- Keep the existing `/model` numbered picker as the Phase 4 model-picker prototype.
- Keep `/commands [query]` / `/palette [query]` as the Phase 4 command-picker prototype.
- Do not add modal overlays yet; they would add key-handling risk without improving tmux capture semantics.
- Defer collapse toggles until transcript grouping helpers exist.
