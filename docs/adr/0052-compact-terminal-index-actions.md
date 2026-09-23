# ADR-0052: Compact terminal index actions

## Status

Accepted — 2026-09-23. Alt-I opens an explicit five-row workspace index surface in fullscreen and regular terminal modes. It uses the native scoped status and application-owned bounded scheduler. No automatic indexing, query-triggered refresh or persistent terminal row is added.

## Interaction

The five rows contain scope/key hints, state, a single bounded detail line, Status and Reindex. Left/Right cycles `all`, `notes`, `skills`; Up/Down chooses an action. Enter invokes the selected action. Status is selected on opening and after a scope change. Opening reads status asynchronously without scanning. Reindex waits for the native scheduler result, including pending-revision and failure outcomes; errors stay within the detail row. Controls and long Unicode text are sanitised and truncated by terminal-cell width.

Escape dismisses immediately, restores draft/cursor and reader position, and does not cancel a shared refresh. Closing while busy and reopening shows a new status read after the outstanding operation returns. Ctrl-C exits the TUI and drains index work. No index action creates a chat message, steering item or command-history entry. While the panel owns input, ordinary keys and pointer events cannot mutate or submit the draft. A busy operation rejects repeated admission and scope changes.

The existing editor remains unchanged in fullscreen mode while the five rows temporarily reduce the transcript viewport. Regular mode temporarily uses the terminal's alternate screen; only five rows contain panel content. Escape restores the main-screen history and editor. Ordinary regular mode still owns scrollback natively, uses no mouse capture and has the same idle dock.

## Lifetime and isolation

TUI initialisation resolves fixed startup scopes and creates an idle scheduler. Status reads have a five-second deadline. Refresh uses ADR-0049's bounded batches and cross-process fences. One outstanding result is allowed, carried on a bounded channel. Each result carries panel epoch and session generation; results for closed/reopened/changed-session surfaces are discarded. The UI state is updated only on the UI event loop.

Cleanup cancels the TUI index context, joins the scheduler and result goroutines before engine/store teardown. Panel dismissal does not close the scheduler. Invalid configuration appears as a bounded error and does not prevent TUI startup. Help/hotkeys documents Alt-I without adding idle hints.

The editor's keymap is suspended while the index or existing selector surface owns input. A focused editor otherwise receives focused-stop handlers before parent preemptive bindings in go-tui; that let a remounted editor swallow Escape after returning from the index panel. The editor now has no bindings/focusability during modal ownership and resumes unchanged afterwards.

## Regular-mode renderer constraint

Two initial implementations failed stronger native-history acceptance. Growing the inline panel left temporary rows behind after resize. The library's application-level alternate-screen API invoked `RenderFull`, whose `Terminal.Clear` emits `ESC[3J`; tmux discarded main-screen history even while the alternate screen was active.

The accepted implementation switches the terminal screen directly but keeps go-tui's inline renderer. A same-size resize dispatch forces redraw without changing the library's normal-screen mode or clearing scrollback. The original dock height remains unchanged, including multiline editors. Transcript flushing pauses while the panel is visible. On return after a resize, the existing `sys: terminal resized to WxH` marker re-establishes history geometry before the editor can grow; no new idle row is reserved. Raw ANSI acceptance rejects `ESC[3J` in regular mode and requires all prior responses to remain ordered through resize, multiline editing and a subsequent completed turn.

## Evidence

`make test-tui-index BIN_DIR=/tmp/gi-scheduler-bin` exercises the native binary with tmux/SQLite at **60×18, 100×22 and 140×36 in both modes**. It proves:

- opening/default Status creates no index identities or scans;
- explicit Reindex publishes a native generation and matching file bytes;
- required-root failure preserves generation, and retry succeeds;
- scope selection, repeated opening, held-peer busy dismissal and late-error isolation;
- draft/cursor/reader restoration, resize, multiline dock restoration and unchanged idle separators;
- no accidental chat submission, plus a subsequent intentional turn after dismissal;
- all 13 native responses retained in regular scrollback, no panel text in history, normal-screen/mouse flags restored and no scrollback-clear escape;
- session picker opening/dismissal after the panel in both modes.

Unit tests cover native status/refresh/error, cancelled shutdown, peer-lease protection, bounded/sanitised rows, no prompt admission and panel/session-generation result isolation. TUI/indexer race suites passed ×3. Existing three-size sessions/model/regular/search suites, TUI smoke/Gherkin, Go/vet/build/hook, **77 functional tests** and **32 helpers** passed. The derived parser now checks **23 proposals** separately from frozen features. Delegated review timed out and provided no independent review evidence.

Artifacts are `test-results/tui-index/`; logs are `/workspace/tmp/gi-tui-index-*`. PNGs under `/workspace/tmp/gi-ui-captures/september23/tui-index-{60x18,100x22,140x36}.png` render the saved native tmux ANSI cells, with labelled fullscreen/regular captures. These are capture renderings, not browser parity tests.

Frozen web credit remains **45/236 Classic**, **2/42 shared**, **191/40 unmapped**. This verifies the explicit terminal index adaptation. Crash/restart reconciliation, external/shell/watch delivery, automatic freshness, terminal search-result navigation and vector search remain separate gaps.
