# ADR-0054: Native Quick Actions with pinned Classic UI

## Status

Accepted — 2026-09-23. Quick Actions now opens from eligible timeline typing and uses native session selection, workspace visibility and command prefill. Frozen cases **original-003, 005, 006 and 007** pass in all six browser projects. Combined coverage is **50/236 Classic**, **2/42 shared**, with **186/40 unmapped**.

## Source and host boundary

`web/upstream/piclaw-quick-actions-70d33bc93.json` pins the unchanged Classic component, item-building helper and keyboard-shortcut helper from `70d33bc93ab540845bbcf5f80503ca8125c71594`. It also hashes the exact Quick Actions CSS region from Classic `chat.css`, starting at `.timeline-quick-actions-overlay` and ending before `.compose-submit-spinner`. Existing supplied files are unchanged; the three newly imported source files are byte-identical to upstream. Helper tests verify every hash.

`app.ts` mounts the supplied component under the active session key, outside search mode. The native catalogue and typing guard are Gi adapters. `GET /api/quick-actions` is authenticated, private/no-store and read-only. It advertises only:

- Show/Hide workspace and Open explorer;
- `/model` and `/compact`, which have native web execution paths.

The component receives live native agents/sessions. It filters archived/duplicate entries and groups Agents, Workspace and Slash commands. Settings, chat-only toggling, terminal and VNC panes are not advertised because their complete native paths are unavailable. Skill discovery does not yet have a faithful web `/skill:<name>` command path and is not advertised.

## Keyboard and command behaviour

One printable, non-whitespace character from eligible timeline content opens the palette with that query. Its input receives focus; exact-title/prefix selection and wrapping Up/Down navigation use the pinned implementation. Enter runs the selected action and dismisses the palette. Escape and outside pointer dismissal clear the query without running an action.

The host's window-capture guard is installed before the component listener. It rejects repeated or already-consumed events, capability-loading events, and additional interactive targets without rewriting the supplied typeahead helper. IME/composition, modifier and whitespace rejection remain in the pinned helper. Reserved keyboard shortcuts retain their existing exclusion. Noninteractive timeline typing only is affected; the guard does not prevent native input default behaviour.

Capabilities start unready and reset on session switch. A revision gate prevents an old response from opening the gate for a newer session. Failed capability loading yields empty workspace/slash selections; agents remain available. The component's own cancellation guard discards late command responses after session remount. No old response opens a palette or changes a draft.

Slash selection creates a new compose-prefill token with `command + ' '`. The supplied composer replaces text, retains attachment/reference state, focuses the textarea and places the cursor at its end. It does not submit. Content changes consume the host request so reload/session revisits cannot replay stale prefill over newer typing. Draft persistence continues through the existing session repository.

Alt+Enter on an agent opens the native session URL in a separate tab. Startup validates `chat_jid=gi:<id>` against the session API before selecting it. A later in-tab session switch updates an existing URL parameter so reload returns to that selection. The original tab retains its draft and selected session. The upstream link also contains `chat_only=1`; this slice does not implement or credit chat-only layout.

## Acceptance boundaries

New frozen mappings:

- **original-003:** timeline typing/query/focus, enabled group order, initial matching, wraparound navigation, Enter action/dismissal;
- **original-005:** prevented, repeated, composing, whitespace and Ctrl/Meta/Alt events do not open the palette;
- **original-006:** Escape/outside dismissal clears query and executes nothing;
- **original-007:** supported command prefill replaces text, adds a trailing space, restores focus/end cursor and creates no turn.

**original-004 remains unmapped.** Real composer, menu and sidebar exclusions pass, and DOM guard fixtures exercise absent classes, but the complete CodeMirror/dialog/listbox/open-popup scenario outline lacks native surface evidence. **original-008 remains unmapped** because skill commands are not exposed. No synthetic DOM fixture earns frozen surface credit.

Classic compose-004/original-017 still require replacing the editor and clearing media on queue return. That conflicts with the independently verified shared-28 merge/latest-draft contract in ADR-0017; this slice leaves those Classic cases unmapped rather than changing queue recovery or claiming incompatible criteria both pass.

## Evidence

Eight browser tests exercise the four new mapped cases plus native exclusions, capability-gated workspace actions, native session selection, linked tabs, delayed responses and capability failure. Prefill coverage includes retained attachment pills, newer typing/reload durability and no prompt admission. Native API tests verify GET-only behaviour, exact supported commands, auth denial, no-store and no turn creation. Helper tests cover capability filtering/deduplication, pinned source/CSS hashes and report boundaries that keep neighbouring cases unmapped.

The complete browser set passed **564 executions**: **396 main** (three saved 132-execution project batches), 54 reconnect, 66 compaction, 18 Steer/admission, 12 context-fit, 12 meter and six index-configuration. A final strengthened **48-execution Quick Actions** capture run passed; the workspace transition also passed **24 repeated executions**. These reruns are additional checks, not double-counted mapping evidence.

`make check BIN_DIR=/tmp/gi-scheduler-bin` passed Go tests, vet, web build, hook checks and **79 functional tests**. **36 helpers** and the full web race suite ×3 passed. Delegated read-only reviews timed out and supplied no independent review evidence.

Earlier focused runs exposed incorrect fixture selectors, an unwrapped settings payload, missing repeated/interactive event guards, and native session names that also matched `/model`. Selectors/settings/guards were corrected. The prefill assertion now counts the supported slash row while still requiring it to be the exact highlighted result; native matching agents are not removed. Held route handlers are drained before replacing failure gates. The pinned component installs/rebinds keyboard listeners in passive effects, so tests wait for focus/effect boundaries between separate user interactions. Assertions were not relaxed. One full run failed these fixture cases; a subsequent run was interrupted at 324/396 without a report. The complete three-batch run replaces that incomplete evidence.

Reports: `/tmp/gi-quick-main-{chromium-mobile,mixed,webkit-large}.json`, supplementary `test-results/ux-parity/*-results.json`, and combined `matrix.{json,md}`. Logs: `/workspace/tmp/gi-quick-*`. Phone/tablet/desktop captures: `/workspace/tmp/gi-ui-captures/september23/quick-actions-*.png`.

## Minimal terminal adaptation

Keep ordinary typing in the terminal editor; do not copy the browser's global timeline-typeahead trigger. Existing Alt-S/Alt-M/Alt-I selectors and slash entry already cover several native actions. A future combined action chooser should require an explicit shortcut/command, use the existing temporary selector pattern with at most six results, default to nonexecuting selection, and restore draft/cursor/reader on Escape. Distinguish command insertion from execution. Do not add a permanent palette bar, sidebar or idle hint row.

This web slice adds no terminal code or terminal acceptance credit. Combined chooser filtering, native action execution, session ownership and three-size fullscreen/regular preservation require separate tests before implementation is accepted.
