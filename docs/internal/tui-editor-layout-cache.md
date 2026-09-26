# Reuse unchanged terminal editor layout

The editor now reuses its current immutable line layout when text, effective width, cursor index, focus and cursor marker are unchanged. This removes repeated whole-draft wrapping during idle redraws without changing visible rows, stored text or key handling.

## Ownership and invalidation

`multilineInput.layoutCache` holds one entry. Its key includes every input used by `layoutLines`; its immutable value contains all wrapped lines and the cursor row. A cache miss constructs a new entry. The menu's copied input snapshot can therefore build another layout without overwriting the composer's cached slice.

Viewport height/scroll, style and blink do not affect wrapping. Each call still computes the visible window and keeps the cursor visible. Returned slices have capacity limited to length so appending an optional help row cannot overwrite another cached row. Callers treat existing rows as read-only. Clearing the draft drops the cache and resets viewport scroll.

Session editor snapshots store text/cursor/undo/history/queued drafts, not the cache. Restoring a different draft changes the key; search owns a separate input instance. No global cache, size limit, text truncation, new controls or extra idle rows are introduced.

This optimises unchanged redraw layout only. Edits, cursor movement, focus changes and resizing still rebuild the full layout. Very large paste/edit latency and retained current-layout memory remain separate concerns.

## Measurement

`make bench-tui-editor-layout` runs the same benchmark before and after the change (linux/amd64, Intel i7-12700, Go test `-benchmem -benchtime=100ms`). These are layout microbenchmarks, not complete terminal-render or input-latency measurements.

| Draft size | Before unchanged layout | Before allocations | Cached lookup | Cached allocations |
|---|---:|---:|---:|---:|
| ~1 KiB | 18.7 µs | 5,680 B /128 | 5.1 ns | 0 B /0 |
| ~100 KiB | 1.93 ms | 574,872 B /11,833 | 5.2 ns | 0 B /0 |
| ~1 MiB | 24.9 ms | 7,192,110 B /121,015 | 5.1 ns | 0 B /0 |

The benchmark generates complete Unicode lines, so actual input bytes are slightly above the labels. Cursor and resize cases still take roughly19ms per1MiB rebuild in the after run; individual timings vary and are not pass thresholds. Unit tests assert zero allocations only for a warmed unchanged layout call, not a UI frame.

## Verification

- `make test vet bun-checks` passes. Tests cover every key change, style/blink/height reuse, fresh-versus-cached equivalence, copied-input independence, clear behaviour and append capacity safety.
- Editor viewport, model picker, session picker, search and regular-mode suites pass24real tmux PTY configurations at60×18,100×22,140×36. They preserve exact long-draft submission bytes, cursor/menu/resize behaviour, per-session drafts, clipboard/scrollback and idle footprint.
- `make test-ux`:107passes,11existing fixture-dependent skips. Browser code is unchanged.
- A delegated read-only cache review timed out without findings. Local inspection confirmed session snapshots exclude the cache; no independent review approval is claimed.

Evidence retains before/after logs and PTY captures. Whole CI and deployment are separate gates. Exact web pixels, physical/emulator-wide acceptance and the broader parity goal remain open.
