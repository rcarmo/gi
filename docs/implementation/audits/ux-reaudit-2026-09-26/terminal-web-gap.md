# Classic terminal pane 001–017: mounted web capability gap

The frozen `tests/ux/features/classic/panes/terminal.feature` spans a live
standalone terminal, command input/output, tab close/popout/theme, dock and
splitter beneath an editable file, plus zen-mode chrome/exit behaviours.
The pinned Piclaw 3.2.4 manifest marks `runtime/web/src/panes/terminal-pane.ts`
changed in Gi and `runtime/web/src/components/tab-strip.ts` identical. Those
files are **not** proof that the mounted Gi web entry exposes a terminal.

`web/src/app.ts` supplies no-op `onOpenTerminalTab`/`onOpenVncTab` callbacks to
the workspace explorer and does not supply dock, zen or popout callbacks to
its `TabStrip`. Its `WorkspaceTab` mounts only read-only/preview-capable panes
through `gi-workspace-tab-lifecycle.ts` in `view` mode, explicitly refusing
terminal/editor/mixed-capability panes. No native terminal websocket route was
found in `internal/web/server.go`; no directly tagged `@ux-terminal-001`–`017`
Gi web journey exists. Thus:

- `001`–`007`: no mounted terminal surface to test shell output, IME, close,
  popout or theme coherence.
- `008`–`012`: no editable editor/terminal dock, keyboard/toggle/splitter or
  zen-dock state to exercise.
- `013`–`017`: no mounted zen-mode chrome or hover/escape/tab-strip affordances.

All seventeen are **native mounted-web capability gaps**, not test passes.
The separate Gi TUI keyboard/PTY review is a different product surface and
cannot substitute for a Classic web terminal. Piclaw 3.2.4 terminal runtime
and physical keyboard/touch acceptance were not probed; no frozen Gherkin,
production code or live instance was changed here.
