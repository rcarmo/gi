# TUI Pi UX parity plan

Goal: make `gi -tui` feel like the Pi interactive assistant while preserving gi's Go runtime, tool contracts, hooks, path policy, and session model.

## Target UX slices

1. **Shell of the assistant**
   - no top chrome; transcript starts at row 0
   - Pi-identical bottom band: separator, editor row, separator, path/branch row, single status row
   - readable transcript with user/assistant/system/draft lines
   - focused input with history, blur/focus, scroll, resize handling
   - compact `/help` plus `/commands` for progressive disclosure

2. **Session and agent workflows**
   - `/agents`, `/where`, `/fork`, `/switch`, `/send`
   - route/fork state available via `/session` and `/where`, not permanent header chrome
   - branch/session history remains visible after switching

3. **Tool and skill affordances**
   - `/tools [query]` compact tool discovery with metadata
   - `/skills [query]` discovery and skill loading guidance
   - visible active-tool/compact/rtk hints without flooding the transcript

4. **Runtime controls**
   - `/model`, `/thinking`, `/compact`, `/cancel`
   - surface compaction thresholds and current context estimate
   - expose hook/plugin state enough to debug Joker extensions

5. **Parity polish**
   - consistent single-line bottom status messages for thinking/tool execution/errors
   - bottom footer row structure identical to Pi
   - tmux-safe rendering for scripted tests

## Scripted UX testing

Use Gherkin `.feature` files for readable UX requirements, but drive the terminal through `tmux` rather than Playwright.

- Features live under `features/tui/*.feature`.
- Runner: `scripts/test-tui-gherkin.sh`.
- Artifacts: `test-results/tui-gherkin/`.
- Each scenario starts a fresh workspace/database and a detached tmux session.
- Steps send keys, capture panes, and inspect SQLite state.

This gives us the same behavior-level regression style as Piclaw's UX features, adapted to a TUI.

### Screenshot/testing backend note

`rcarmo/go-te` was considered for screenshotting-style terminal tests. For the current parity phase, tmux pane captures are the primary artifact because they are text-diffable, CI-lightweight, and already exercise real terminal input/rendering. `go-te` remains a possible future enhancement if we need image-level terminal regression evidence, but it is not required for the current Gherkin evidence loop.

## Implemented slices

- Compact `/help` command with progressive-disclosure pointers to `/commands`, `/model`, `/session`, `/where`, `/attach`, and shell shortcuts.
- `/tools [query]` command backed by the engine tool registry.
- `/tools active`, `/tools activate ...`, and `/tools reset` for active tool visibility/control.
- `/skills [query]` command backed by workspace skill discovery, including command invocation hints, source paths, and metadata warnings.
- `/skill:name [args]` command resolution that loads discovered `SKILL.md` files from the TUI.
- `/model`, `/thinking`, `/compact`, and `/cancel` runtime controls, with richer `/model` listing and index selection.
- `/scoped-models [list|add|remove|set]` enabled-model management backed by existing settings persistence.
- Keyboard cycling for models (`Ctrl+L`/`Alt+L`) and thinking levels (`Ctrl+T`/`Alt+T`).
- `/commands [query]` / `/palette [query]` textual command palette fallback.
- `/session`, `/new`, `/name`, `/resume`, `/clone`, `/copy [--osc52|--native|--auto|--fallback]`, and `/reload` command/session workflow affordances.
- `/agents`, `/where`, `/tree`, `/plugins`, `/fork`, `/switch`, and `/send` session/agent workflows/debug views, with denser narrow-terminal output for `/tree` and `/resume`.
- Keyboard coverage for blur/focus, F2/F3 history hints, queue restore, editor word movement/deletion, line deletion, undo/yank, Tab path completion, scroll, resize, and quit.
- Gherkin/tmux features covering boot, help, tools discovery/activation, runtime controls, session workflows, prompt submission, persistence, and keyboard behavior.
- CI-friendly tmux/Gherkin artifacts: pane captures, report markdown, per-feature SQLite dumps, session/message extracts, and failure summaries.
- TUI status rendering for thinking deltas, tool completion/failure, generic errors, context compaction broadcasts, and transient runtime notices in the single bottom status line.
- Richer transcript line rendering for user/assistant/system roles, folded multi-line tool results, code-block line counts, and compact runtime compaction summaries.
- Pi-style transcript tooling/runtime blocks with colored borders, timestamps/elapsed time, live started→finished/failed state updates for tools, structured hook/routing/dispatcher/compaction/sub-turn blocks, and mouse/F6/F7/F8 expansion/navigation.
- `/settings`/`/config` grouped runtime view covering runtime, model, editor, session, discovery, compaction, and peering settings.
- `/approvals` visibility for the current approval-gate state; today this reports that gi has no approval gates configured.
- Expanding multiline input with Pi-style empty editor row instead of placeholder text or an `Input:` label.
- Pi-identical steady-state layout that budgets transcript height from the fixed bottom band: separator, editor row, separator, path/branch row, single status row.
- Terminal-safe Markdown transcript rendering for headings, lists, blockquotes, links, code blocks, and responsive table fallbacks.
- Simplified `/help` output with detailed command discovery delegated to `/commands` and other focused commands.
- Shell shortcuts: `!cmd` asks the model to run/summarize a command, while `!!cmd` runs locally and prints bounded output.
- Textual `@path` file reference fallback and Tab path completion.
- Queue UX: the bottom status line shows queued/steering counts when non-zero, active-turn submissions produce visible queued feedback, and `Alt+Up` restores locally queued drafts.
- Clipboard/media boundary documented: `/copy` remains a transcript fallback by default, OSC 52/native clipboard helpers are opt-in, bracketed paste remains parser/editor-dependent, and image paste waits for a shared media ingestion contract.
- Extension author docs now cover extension command semantics plus `gi.state`, `gi.topics`, and `gi.runtime` namespace expectations.
