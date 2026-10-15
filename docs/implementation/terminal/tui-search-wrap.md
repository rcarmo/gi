# Search across renderer soft wraps

Fullscreen transcript search now finds a literal occurrence spanning soft-wrapped rows of one visible text element. Each occurrence still occupies one search result; all of its visible row segments are highlighted. No new search field, counter, footer or idle row is added.

## Admitted content

`transcript_search_wrap.go` inspects the same laid-out go-tui text leaves used to build rendered search rows. It reads their source text, wrap/alignment setting and content rectangle. A candidate run must reproduce a single source paragraph from the visible grapheme cells before it enters the index. The source supplies a space omitted at a word wrap; trailing buffer padding never becomes content. It does not infer continuity from a full-width row.

The index keeps a byte-to-display-cell map for the verified run, then applies case folding with the original cell mapping. A match inside a wide or combining grapheme highlights the complete cell. Wrapped and same-row occurrences share non-overlapping matching over the verified paragraph, so crossing a row does not restart the match count. Highlighting retains styles and hyperlink metadata outside the selected range. Rendered cells remain the source for selection/copy.

Each leaf is an independent boundary. Explicit newlines reject the whole leaf from cross-wrap joining, retaining its existing row-local search. Separate messages, headers, footers and collapsed tool rows are never concatenated. Multi-paragraph text, prewrapped Markdown, and inline-code layouts split into separate elements still need source-break metadata for complete cross-wrap coverage. Unverified/clipped projections fail closed to row-local search.

## Interaction and footprint

The existing `Ctrl-Shift-F`, Enter, Shift-Enter and Escape actions are unchanged. The input replaces the composer temporarily; closing restores draft, cursor and saved reading/follow state. Regular mode continues to use terminal-owned scrollback/search. No idle chrome, theme or renderer change is introduced.

Resize/live-output invalidation rebuilds the same cache. It retains the previous row/start selection rule; mutation-stable reflow/eviction anchors are not implemented by this slice. Wrapped links can participate in search if their text leaf validates, but general wrapped-link pointer activation remains a separate gap.

## Verification

- `make test vet bun-checks` passes. Unit tests cover widths 18/24/60/100/140, mid-word and word-boundary wraps, wide/combining Unicode, non-overlap across rows, multiple highlighted segments, hard/message boundaries, hidden versus expanded tool output, projection mismatch rejection, and preserved editor/follow state. Existing exact-row and hyperlink-style tests remain.
- `make test-tui-search BIN_DIR=/tmp/gi-search-wrap-test-bin`: three real PTYs at 60×18, 100×22 and 140×36 pass. A long native prompt and reply each create one cross-wrap result; next/previous, original 96 occurrences, no-match, live arrivals, resize, collapsed tools, cursor/draft/history restoration and zero query submissions all pass.
- `make test-tui-selection` passes at the same three sizes, covering native SGR drag, clipboard release/keyboard paths, edge scroll, Escape, resize/new-output invalidation and zero idle rows.
- `make test-ux`: 107 passes, 11 existing fixture-dependent skips. No browser code changed.

An early PTY fixture put the long message at the top of history, shifting an unrelated fixed-count prompt-jump check. Moving it to the middle preserves both original navigation assertions and new cross-wrap coverage. Initial compile errors from the expanded match struct were repaired before the passing runs. No tests removed or timeouts increased. A delegated renderer review timed out; no independent result is claimed.

Evidence is under `test-results/tui-search` and `test-results/tui-selection`, with text and ANSI captures. This verifies those PTY configurations, not every emulator or physical device. Whole CI and deployment remain separate gates; the broader parity goal remains open.
