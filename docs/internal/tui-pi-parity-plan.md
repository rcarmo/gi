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

## Implemented slices

- `/help` command with command and keybinding reference.
- `/tools [query]` command backed by the engine tool registry.
- `/tools active`, `/tools activate ...`, and `/tools reset` for active tool visibility/control.
- `/skills [query]` command backed by workspace skill discovery.
- `/model`, `/thinking`, `/compact`, and `/cancel` runtime controls.
- `/agents`, `/where`, `/tree`, `/fork`, `/switch`, and `/send` session/agent workflows.
- Keyboard coverage for blur/focus, F2/F3 history hints, scroll, resize, and quit.
- Gherkin/tmux features covering boot, help, tools discovery/activation, runtime controls, session workflows, prompt submission, persistence, and keyboard behavior.
- TUI status rendering for thinking deltas, tool completion/failure, generic errors, and context compaction broadcasts.
