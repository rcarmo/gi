# Disabled context-control explanations

The composer context control now explains why manual compaction is unavailable. It uses the native capability reason when fresh, and separate text for a pending request, refreshing state or disconnected client. No extra row, badge or idle panel is added.

## Scope and ownership

The existing composer already calls the native token-guarded manual compaction API. An empty session has insufficient eligible history and correctly disables the control. The render-only fixture's disabled opacity differs from Piclaw's compaction-capable control; that difference does not justify enabling an unsupported action.

`compactionUnavailableReason` gives connection state precedence over pending/freshness state, then active work and native availability. Reasons are written as tooltip text and `aria-description`. The numerical accessible name, measured/unknown usage semantics, colours and arc remain unchanged. The description is removed when the action becomes available; an active compaction retains its existing elapsed status.

The selected-session refresh commits activity and capability together. A delayed initial refresh leaves the previous session's meter absent. Later stale responses cannot replace the current session's reason. The existing selection/connection/revision guards, POST token, server admission checks and duplicate-click lock are unchanged. Tooltip updates never submit prompts, compact automatically or write session state.

## Refinement decisions

This is a bounded explanation fix for the existing control. Inputs are the native compaction snapshot and current activity/connection/pending state. Errors remain in the existing alert; policy editing, new compaction modes, counters, persistent UI and new terminal controls are out of scope. Draft, attachments, references, keyboard ownership and native mutations keep their existing paths. Web tests use disposable sessions only; live data and permissions are untouched.

The TUI already provides explicit `/compact` and Alt-C actions and native error output. It needs no disabled web-style meter or idle tooltip analogue. No terminal code changes in this slice.

## Verification

- `make test-context-control-helpers`: five context/compaction helpers pass, including state precedence, native reasons, unknown capacity and warning thresholds.
- `make test-ux-context-meter`: 18 Chromium/WebKit × viewport checks pass, including accessible disabled reasons, inert clicks, coherent session activation and rejected late responses.
- `make test-ux-compaction`: 96 checks pass. A held manual admission POST shows its pending reason and rejects a second click while preserving draft/media; existing automatic/manual/cancel/failure/Settings ownership tests remain intact.
- The unknown-usage suite and targeted reconnect/Stop regression each pass six cases. The latter verifies the reconnect explanation while disconnected. An older model-picker selector in the unknown-usage test was migrated to the existing listbox/option contract.
- `make test vet bun-checks test-pixel-helpers` passes, including 15 pixel helpers. Stock functional tests: 107 passes, 11 existing fixture-dependent skips.
- A separate required 15-minute CI job runs the meter/compaction matrix, keeping the existing compose job's budget unchanged. Helper checks join the existing core job. No tests removed, timeouts increased or failure tolerances added.

Initial assertions assumed different native wording and a visible meter during an incomplete initial session refresh; they now assert the actual native reason and coherent no-stale-meter behaviour. The delayed-route fixture was changed to intercept one response and await its completion before removal. A delegated review timed out without a result.

The slice is not deployed until whole CI succeeds. Live Gi remains the verified `daa8371` build on 8090. No new pixel result is claimed; exact parity, physical/screen-reader and Visual acceptance remain open.
