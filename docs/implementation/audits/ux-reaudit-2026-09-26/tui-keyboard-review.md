# Native TUI keyboard scenario: qualified clause review

`features/tui/keyboard_behavior.feature:4–27` is a Gi-native scenario, not a
Classic web oracle requirement. It combines input blur/focus, a single prompt,
PageUp/End, two resizes and Ctrl-D exit. The canonical tmux Gherkin step runner
checks SQLite message count/content, screen substrings and tmux session state.
The screen checks prove text remains present; they **do not** prove viewport
movement, history position or cursor preservation.

| Clause | Assertion and native path | Limit |
|---|---|---|
| Escape blurs; typing/Enter does not send; Tab restores input focus | `scripts/test-tui-gherkin.sh` sends real tmux keys. The feature checks zero messages before Tab and one exact user message plus the local provider reply afterward. `internal/tui/chat.go:1602–1620` handles Escape/Tab by input state. | No physical keyboard, IME or accessibility screen reader. |
| PageUp/End history | `internal/tui/chat.go:1623–1627` binds transcript paging and newest-edge navigation. The frozen Gherkin only checks the reply is still visible. The **separate** native PTY test `scripts/test-tui-reading.mjs` asserts the upper history anchor, draft text/cursor, delayed provider completion and explicit PageDown following at 60×18, 100×22 and 140×36. | That stronger reading test does not rewrite the Gherkin assertion or cover every key in the combined scenario. |
| Resize to 100×22 then 60×18 | Gherkin checks a live tmux session and visible reply. `test-tui-reading.mjs` verifies a resize round trip leaves the history anchor and cursor intact and adds no idle rows at its three sizes. | The reading test is a separate process and journey, not a tagged execution of the combined feature. |
| Ctrl-D exit | The feature requires tmux session exit; `internal/tui/chat.go:1597` handles Ctrl-D on empty input. | Does not exercise a nonempty editor's Ctrl-D behaviour. |

A focused invocation of `scripts/test-tui-gherkin.sh` with `FEATURE_DIR`
containing only `keyboard_behavior.feature` and separate audit-only
`ARTIFACT_DIR`/`TEST_DIR` passed the one scenario with isolated SQLite/tmux
state. The report and captures are under
`/workspace/tmp/gi-ux-reaudit-20260926/tui-keyboard-evidence`. `make test-tui-reading` passed at 60×18, 100×22 and 140×36.
The earlier whole seven-feature `make test-tui-gherkin` invocation was aborted
while on `assistant_basics.feature`; its result is unknown and it left an
orphaned audit-only tmux session, which was closed. Do not count that run as a
pass. TUI autosave WIP `2a87a790` remains isolated and unpublished.
