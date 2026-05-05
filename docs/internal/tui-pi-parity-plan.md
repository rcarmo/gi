# TUI Pi UX parity plan

Goal: make `gi -tui` feel like the Pi interactive assistant while preserving gi's Go runtime, tool contracts, hooks, path policy, and session model.

## Target UX slices

1. **Shell of the assistant**
   - persistent status/header with session, agent, model, provider, thinking level, message/turn counts
   - readable transcript with user/assistant/system/draft lines
   - focused input with history, blur/focus, scroll, resize handling
   - inline `/help` that lists keyboard shortcuts and commands

2. **Session and agent workflows**
   - `/agents`, `/where`, `/fork`, `/switch`, `/send`
   - route/fork state reflected in the header
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
   - consistent status messages for thinking/tool execution/errors
   - keyboard help footer similar to Pi
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

- `/help` command with command and keybinding reference.
- `/tools [query]` command backed by the engine tool registry.
- `/tools active`, `/tools activate ...`, and `/tools reset` for active tool visibility/control.
- `/skills [query]` command backed by workspace skill discovery.
- `/model`, `/thinking`, `/compact`, and `/cancel` runtime controls.
- `/agents`, `/where`, `/tree`, `/plugins`, `/fork`, `/switch`, and `/send` session/agent workflows/debug views.
- Keyboard coverage for blur/focus, F2/F3 history hints, scroll, resize, and quit.
- Gherkin/tmux features covering boot, help, tools discovery/activation, runtime controls, session workflows, prompt submission, persistence, and keyboard behavior.
- CI-friendly tmux/Gherkin artifacts: pane captures, report markdown, per-feature SQLite dumps, session/message extracts, and failure summaries.
- TUI status rendering for thinking deltas, tool completion/failure, generic errors, and context compaction broadcasts.
- Richer transcript line rendering for user/assistant/system roles, folded tool results, and compaction summaries.
- `/settings`/`/config` runtime view for provider/model/thinking, compaction, peering, active tools, and scrollback limit.
- `/approvals` visibility for the current approval-gate state; today this reports that gi has no approval gates configured.
- Expanding multiline input with simplified horizontal-rule chrome instead of a visible `Input:` label.
- Responsive layout that wraps status/context/footer metadata and budgets transcript height from actual block sizes on narrow terminals.
- Terminal-safe Markdown transcript rendering for headings, lists, blockquotes, links, code blocks, and responsive table fallbacks.
