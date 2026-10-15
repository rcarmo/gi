# Fullscreen word and line selection

Double-click now selects a rendered word, triple-click selects the rendered line, and dragging after either gesture extends by that unit. The existing selection separator, highlight and clipboard paths are reused. No idle rows or controls are added.

## Reference and boundaries

The installed pi-tui `tui-alt-screen.js` uses a 500ms sequence on the same word bounds and scroll view, cycles character → word → line → character, and joins Unicode word segments through `/` and `-`. Gi applies the same interval, same-word display bounds, sequence cycle and joiners to its fullscreen transcript.

`transcript_selection_words.go` uses the already-pinned `clipperhouse/uax29/v2` Unicode17 word segmenter. The module is now a direct dependency; its version/checksum did not change. Segment byte positions are mapped back to the full rendered grapheme cells. This retains combining marks, CJK width and emoji cells. Locale dictionary behaviour may differ from JavaScript `Intl.Segmenter`, especially scripts without explicit word separators; exact segmentation parity for every language is not established.

Word/line granularity updates both endpoints for forward or reverse drags and held-edge scroll. A stationary double/triple click does not start edge scrolling. Copy-on-release, Ctrl-C/Ctrl-X, clipboard opt-in, OSC52 payload bounds, asynchronous native copy and stale-copy rejection use existing policy. Line copy trims trailing rendered padding, as ordinary selection does.

OSC8 text keeps terminal link ownership. Clickable tool blocks retain their single-click expansion path; they do not acquire word/line multiclick behaviour. Regular mode retains terminal-owned mouse selection, scrollback and copy. Web controls and live runtime state are untouched.

## Sequence ownership

A released first click retains a bounded snapshot for the next click. Matching requires the same session generation, retained transcript, expansion/search state, word bounds and view dimensions. Output, resize validation, session/query changes, Escape, outside/scrollbar press, modified clicks, wheel input and actual drag movement discard the sequence.

While a button is held, the active selection owns validation; the released-click snapshot is installed only at release. This prevents an intervening go-tui render from erasing a valid sequence. Existing render invalidation still clears stale active selections. Multi-click state never becomes durable session/editor state.

## Verification

- `make test vet bun-checks` passes. Unit tests cover word/path/kebab/punctuation/combining/CJK/emoji boundaries, wide-cell halves, sequence interval/cycle, forward/reverse range extension, stationary and moving edge scroll, stale scopes/search/output/resize, modified/wheel/outside gestures, clipboard-off behaviour and editor preservation.
- A press → render validation → release regression covers a race first caught in PTY testing. No test or timeout was removed to obtain a pass.
- `make test-tui-selection test-tui-search test-tui-regular BIN_DIR=/tmp/gi-word-selection-test-bin` passes nine real tmux PTY configurations at60×18,100×22,140×36. Native SGR bytes exercise double/triple clicks and reverse word dragging; actual OSC52 clipboard output is checked. Existing selection/copy/search/soft-wrap/live-arrival/resize/regular-scrollback cases remain.
- `make test-ux`:107passes,11existing fixture-dependent skips. Browser code is unchanged.
- A read-only delegated review timed out; no independent review result is available.

The first unit draft had fixture/API naming and coordinate errors. A later resize check exposed released-click validation coverage; the added validation initially invalidated a held first press during redraw, now repaired and covered. Logs retain these failures.

Artifacts are in `test-results/tui-selection`, `test-results/tui-search` and `test-results/tui-regular`. PTY verification does not establish every emulator, pointer device or locale. Cross-wrap word/line selection, wrapped-link activation, prewrapped Markdown search, stable reflow/eviction anchors, light-theme parity and oversized-editor limits remain open. CI/deployment are separate gates; no live terminal session was mutated.
