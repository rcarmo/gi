# Gi TUI Pi-identical layout contract

Status: active layout contract for the Pi/PiClaw UX convergence track.

## Goal

Gi's steady-state terminal layout must match Pi's row structure, not merely approximate it. Gi may display Gi-specific values, but the physical layout and chrome budget should be identical.

## Vertical row order

From top to bottom:

1. Transcript area starts at row 0. No top status/header/context chrome.
2. Bottom separator line spans the full content width.
3. Editor/input row. Empty editor renders as a blank line, with only a cursor when focused/blinking.
4. Bottom separator line spans the full content width.
5. Path/session row: current workspace path plus branch/session marker.
6. Single status/notification row.

No other permanent rows are allowed in steady state.

## Bottom status line

The final row is the only physical line for transient notifications/status.

It must:

- remain one line;
- never wrap;
- left-align current transient notification or compact counters;
- right-align model and thinking where width allows;
- include queue/steering counters only when non-zero;
- truncate segments rather than adding another line.

Durable events can still be written to the transcript when they are part of history, but short-lived running/tool/error/queued indicators should prefer the single status row.

## Path/session row

The path row should mirror Pi's footer path row:

```text
/path/to/workspace (branch)
```

Gi uses the workspace path and reads `.git/HEAD` directly for a best-effort branch label. If no branch is available, it displays just the workspace path.

## Editor row

The empty editor row is blank, matching Pi's empty editor band. Gi must not display placeholder text like `Send a message…` in the empty steady-state layout.

## Regression states

Capture and compare at 60x18, 100x22, and 140x36:

- empty startup;
- existing transcript;
- focused empty editor;
- prompt being typed;
- running turn;
- queued follow-up;
- tool/error notification;
- scrolled transcript.
