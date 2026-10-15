# Complete link targets across terminal wraps

Long parenthesised HTTP(S) targets now retain OSC8 metadata when Markdown projection would previously split the URL into separate strings. The terminal remains responsible for opening links. Gi launches no browser process and adds no controls or idle rows.

## Projection and safety

Markdown renders an explicit link as `label (target)`. `wrapParagraph` now leaves a complete safe `(target)` token intact until `transcriptLinkSpans` attaches the validated target and go-tui performs display-cell wrapping. Each resulting link segment has the same full target. Ordinary trailing sentence punctuation (`.,;:!?`) is preserved but remains outside the clickable span.

The existing validator is unchanged: HTTP(S) with a hostname only, no credentials, backslashes, controls (including percent-decoded controls), malformed escaping or targets over2048bytes. Partial/wrapped fragments are never reconstructed into URLs. Visible text still includes the target; selection copies displayed text only, including existing visual line breaks, without hidden metadata or escapes.

This is a narrow change to projection order. Plain complete links, search rendering, OSC8/tool-click precedence, wide-cell hit testing and clipboard policy retain their existing paths. Regular mode emits the same targets into terminal-owned history with mouse capture off. Browser UI and live session/auth state are untouched.

## Scope limits

Supported tokens begin with the parenthesised target and may end with the punctuation above. Nested parentheses, whitespace in targets, adjoining non-punctuation text, complex inline-code/table projection and arbitrary autolinks are not newly supported. Long first tokens were already retained by the projector; this repairs long tokens reached after preceding prose. It does not overhaul Markdown width/reflow behaviour.

Complete OSC8 emission is verified, not physical click behaviour in every terminal. Search preserves target metadata but still has the documented prewrapped-Markdown boundary limitations; this slice does not claim all repeated substrings spanning those boundaries are found.

## Verification

- `make test-terminal-links`: race-enabled link/selection/search tests pass three repeats. New cases cover30/60/100/140column rendering, long Unicode targets, trailing punctuation exclusion, search metadata, wrapped-link click ownership, visible-slice copy, and unsafe/incomplete token rejection.
- `make test vet bun-checks` passes. `make test-ux`:107passes,11existing fixture-dependent skips; browser code is unchanged.
- `make test-tui-links`: six fullscreen/regular tmux PTYs pass at60×18,100×22,140×36. Captured raw output contains the complete long target. Fullscreen search and resize preserve it; draft/cursor and idle footprint remain unchanged. Terminal activation stays host-owned.
- `make test-tui-search test-tui-selection`: six additional PTYs pass, preserving prior search/multiclick/clipboard/scroll/resize contracts.

Read-only delegate review found that sentence punctuation defeated the first implementation's whole-token guard. The guard and tests now include punctuation without adding it to the link. An initial PTY probe remained at an older selection anchor; it now explicitly navigates to the new output. Another expected complete cross-wrap occurrence counts beyond the current Markdown-search scope; it now tests a visible suffix and independently asserts the exact OSC8 target. No tests removed, timeouts increased or unsafe URLs admitted.

Evidence is under `test-results/tui-links`, `test-results/tui-search` and `test-results/tui-selection`. Whole CI and deployment remain separate gates. Exact web pixels, physical-device/Visual acceptance and remaining terminal gaps stay open.
