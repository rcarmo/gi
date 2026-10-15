# Pi-compatible fullscreen scrolling and redraws

## Source comparison

Compared installed `@earendil-works/pi-tui` **1.0.0**, specifically
`dist/tui-alt-screen.js` (`doRender`, wheel routing, page shortcuts) and
`dist/wheel-scroll.js`. Pi compares screen rows, moves to column one, erases the
whole changed line, then writes all of it inside synchronized output. It does
not use terminal scrolling regions to move fullscreen transcript rows.

Gi previously used go-tui's per-cell diff. Scrolling through compound Unicode
can therefore resume a write in a row that the terminal has laid out differently,
leaving glyph fragments. The new path adopts Pi's redraw granularity; it is not
a promise that every terminal's Unicode width rules match Gi's.

## Contract

- Fullscreen compares complete rows (glyphs, cluster tails, styles, links and
  continuation cells). Each changed row is erased from column zero and repainted
  contiguously from the start. No mid-row cell-diff cursor jumps are used.
- `ESC[K` at column one is equivalent to Pi's `ESC[2K` whole-line erase. Style
  and OSC8 transitions still use go-tui's serializer. Blank tails stay erased;
  styled spaces are retained. Unchanged rows produce no text output.
- Startup, resize and resume force all rows. This path does not erase terminal
  scrollback, clear the whole screen on every wheel event, or replay output in
  a post-render hook. Existing synchronized updates and final cursor placement
  remain intact.
- Regular mode still uses terminal-owned scrollback and the existing inline dock
  renderer, with no mouse capture. The row option is set once on the app and is
  ignored in inline mode, so regular-to-fullscreen switches also use it.
- Auto wheel uses Pi's 5 ms burst threshold, 200 ms gesture gap, 100 ms reference,
  smoothed velocity, fractional carry and six-line cap. Local macOS with **none**
  of `SSH_CONNECTION`, `SSH_CLIENT`, `SSH_TTY` present moves one line per event
  because the OS has already accelerated it. SSH and other platforms use the
  accelerator. Numeric overrides remain fixed; changing the setting resets the
  gesture. Event gaps retain sub-millisecond precision.
- PgUp/PgDn keep four viewport rows of overlap, with a minimum one-row movement.
  Existing follow-at-bottom, input-edit jump-to-bottom and selector ownership
  remain unchanged. This is not an implementation of all Pi nested scroll views,
  scrollbar options or half-page keybindings.

## Dependency patch

The runtime-only go-tui v0.22.1 snapshot at `third_party/go-tui` is selected by
`go.mod`. It adds opt-in `WithRowRedraw`, `RenderRows` and `Buffer.RowDiff`.
See its `README.gi.md`, retained upstream LICENSE and `gi-row-redraw.patch` for
provenance and the entire behavioral patch. Default library cell diffs and inline
rendering are unchanged. No module-cache files were modified. The local replacement
should be removed when equivalent support is available upstream.

## Verification

- `make test-pi-scroll`: actual ANSI UTF-8, unsplit flags/ZWJ/combining clusters,
  row-clear-before-paint, no intra-row cursor jumps, unchanged rows, shrink/blank,
  forced repaint, wide continuations, backgrounds, OSC8, resize, wheel and paging.
- `make test-go-tui-runtime`: retained upstream runtime/internal tests and the
  new application-path synchronized-row regression.
- `make test-tui-table-scroll`: **516** real tmux frame comparisons at four widths,
  scrolling both ways through Unicode tables using `RenderRows`.
- `make test-tui-tool-syntax-pty`: six live fullscreen/regular scenarios checking
  source colours, expand/collapse, resizing and persisted reopen.
- `make test-tui-regular`: three PTYs checking normal screen, native scrollback,
  ordering, dock/editor/selector/session/resize and reopen.
- `make test-tui-selection`: three PTYs checking selection, OSC52, edge scrolling,
  resize and new-output invalidation. The harness restores Home after a draft
  edit before reverse-selection comparison, respecting the existing input-edit
  jump-to-bottom contract rather than comparing two different viewports.
- Full Go suite and vet pass on the final run. An earlier full run hit a transient
  turn-engine fixture error (`no such table: media`); no turn-engine code changed.
  Browser tests are skipped per ChromeOS host guidance.

**Ghostty visual acceptance remains pending.** ANSI-byte tests establish the
redraw method; tmux tests establish tmux rendering, not Ghostty rendering. The
prior Ghostty probe remains available but is not required to implement this
redraw change. Neither the live binary nor the user's terminal was restarted.

## Jump to latest message

The cue matches Pi's fullscreen `scrollToEndIndicator` (pi-tui `TuiAltScreen`, as configured by pi-coding-agent's `tui-renderer`). It is implemented in `internal/tui/jump_to_latest.go`.

- **When:** fullscreen only, while the transcript is not following its end.
- **What:** ` ↓ Jump to latest message · End ` (`End` is Pi's default key for `tui.altScreen.bottom`), drawn in the theme's `text` colour on `selectedBg`.
- **Where:** centred on the transcript's last visible row and cut to the transcript width. It is composited over the laid-out frame through go-tui's `WithPreFlushHook` and does not take a layout row.
- **Clicking it** (a left press) scrolls to the bottom and resumes following. A click on the cue is handled before text selection starts. `End` does the same.
