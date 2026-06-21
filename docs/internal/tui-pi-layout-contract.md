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
6. Stats row: counts (`m/t`, optional `q/s`) plus token usage and context on the left, model/thinking on the right.
7. Optional transient notification row, shown only while a short-lived status is active.

The bottom band may grow to several lines (PiSwift-style footer), but it must never add top chrome and must stay below the editor.

## Bottom status band

The bottom band carries path/branch, stats, and transient notifications. Updated policy (status row may expand):

- the path row is always present;
- the stats row carries counts plus token/context usage on the left and model/thinking on the right;
- a transient notification row appears only while a short-lived status is active;
- each row truncates rather than wrapping;
- queue/steering counters appear only when non-zero;
- model/thinking are right-aligned where width allows.

Durable events can still be written to the transcript when they are part of history, but short-lived running/tool/error/queued indicators should prefer the bottom band.

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
