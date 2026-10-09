# ADR 0060: Pi-style changed-row fullscreen redraws

Status: Accepted

## Context

Gi's fullscreen output uses go-tui cell diffs, whereas Pi clears and repaints
complete changed rows. Compound Unicode scrolling leaves fragments in the user's
Ghostty terminal; previous tmux tests cannot establish Ghostty acceptance. The
user requested the same scrolling method as Pi. go-tui v0.22.1 has no public hook
to replace the frame serializer before output.

## Decision

Add an opt-in row renderer in a provenance-documented runtime snapshot of go-tui,
selected by a local Go module replacement. Keep synchronized output, styles,
links and cursor placement. Enable it for Gi fullscreen rendering only; inline
regular mode retains its original path. Match Pi's wheel acceleration platform
policy and four-row paging overlap.

## Consequences

Changed rows send more bytes than cell diffs but remain bounded by the viewport;
unchanged rows emit nothing. The copy is a maintenance cost: the patch must be
rebased when upgrading go-tui and removed when upstream offers equivalent support.
This avoids reflection, cache edits, global screen clears and duplicate repaint
hooks. It does not resolve terminal Unicode-width disagreements by itself or
establish that the Ghostty report is fixed. Regression evidence and limitations
are recorded in `docs/implementation/terminal/tui-pi-scrolling.md`.
