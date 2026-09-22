# ADR-0031: Opt-in terminal-owned scrollback

## Status

Accepted — 2026-09-22. `gi -tui -tui-mode regular` and `gi-tui -tui-mode regular` use the terminal's main screen. Fullscreen remains the default. The broader Pi convergence goal is unfinished.

## Rendering and input

Regular mode uses go-tui 0.18.2 `WithInlineHeight`, `PrintAboveElement` and a post-render flush. It does not enter the alternate screen or enable mouse capture. The terminal or multiplexer owns history scrolling, text selection and copying. Existing shell output survives startup and exit.

The idle dock has five rows: separator, one-line editor, separator, path, stats. Multiline input and temporary selectors grow the dock; active output has a maximum three-row preview. There is no persistent history panel, toolbar or added status row. Home/End control the editor in regular mode; fullscreen retains its transcript-navigation policy. In-app transcript paging/folding bindings are absent from the regular-mode main keymap. Model/session selectors retain their navigation.

Completed retained output prints once, fully expanded, with the Pi outcome colors from ADR-0030. Native active-claim release and final draft/tool replacement gate printing. Unknown store activity keeps output in the preview. Terminal-owned rows are immutable: they cannot later fold, recolor, update or disappear. Session changes print a labelled boundary and replay that session's retained transcript without clearing previous terminal history.

## Resize and editor corrections

The initial smoke checks missed missing or reordered printed rows after resize and selector growth. Stronger checks compare all twelve numbered native responses, in order and without duplicates, after those operations. On a terminal-size change, a `sys: terminal resized to WIDTHxHEIGHT` history marker re-establishes the inline renderer's invalidated geometry before dynamic dock resizing. These markers consume history lines only when size changes; they add no idle row. Rapid reflow across every terminal implementation has not been established.

Multiline verification also exposed two pre-existing editor defects: Shift+Enter had no explicit key binding despite the handler supporting it, and the rendered line container used horizontal layout. Shift+Enter and Ctrl-J now insert a newline, and editor lines stack vertically. PTY tests use a real modified-Enter byte sequence, verify separate rows and retain them through a resize round trip.

## Evidence

`make test-tui-regular` runs real tmux PTYs at 60×18, 100×22 and 140×36. It verifies:

- main screen and mouse-reporting flags stay off;
- pre-existing shell output, every numbered prompt/response and their order survive printing, resize and selectors;
- tmux-native history selection copies text while a native shell-provider turn is active;
- completion does not exit native copy mode or return its reader to the bottom;
- output bands and the first/last rows of a 40-line tool result reach native scrollback;
- newer text/cursor, multiline layout, Home/End, temporary selectors and session isolation remain intact;
- the idle dock returns to five rows;
- normal exit retains history, and restart reloads persisted messages.

Unit tests cover invalid-mode rejection before store access, final-response/running-tool gates, pruning versus printed offsets, key ownership, session/editor state and multiline rendering. Full Go tests/vet, terminal races repeated three times, all existing three-size outcome/reading/session/model/compaction tests, smoke/Gherkin, 29 helpers and 70/70 functional browser tests pass. CLI entrypoints compile. Web components and frozen scenarios are unchanged; browser mapping remains 34/236 Classic, 2/42 shared (202/40 unmapped), with the prior full matrix at 384/384 for `4071c01`.

Artifacts: `test-results/tui-regular/`. Real 100×22 XTerm pending/completed captures are under `/workspace/tmp/gi-ui-captures/september22/tui-regular-scrollback-*.png` and were attached to the conversation.

## Limits

This is opt-in to avoid changing existing fullscreen workflows. The active three-row preview is application-owned; completed retained output is terminal-owned. Existing scrollback/output retention limits still apply before printing and on session reload. Already-printed text survives Gi's in-memory eviction, but terminal history capacity is configured by the user.

Switching away during active work cannot print unobserved origin completion into the new session; returning reloads persisted history. Failure/uncertain-delivery draft recovery and exactly-once reconciliation across edits/reloads remain separate work. Oversized editors/widgets and very small terminal dimensions need additional bounded-layout tests.

Fullscreen transcript search, marked-prompt jumps, selection/copy/edge autoscroll, light-theme support and stronger reflow/eviction anchors remain open. Folder-reference WIP is still saved separately at `/workspace/tmp/gi-folder-reference-wip.patch`, unmapped and unshipped.
