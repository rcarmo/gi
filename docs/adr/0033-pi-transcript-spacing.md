# ADR-0033: Restore Pi's blank separators and message padding

## Status

Accepted — 2026-09-22. Corrects the overly compressed transcript introduced with ADR-0030's outcome bands. Minimal footprint constrains permanent controls; it does not remove readable message spacing.

## Reference and layout

The installed Pi 0.85.1 interactive renderers define:

- `user-message.js`: `Box(outputPad, 1)` around the complete message, with `outputPad` defaulting to one column.
- `assistant-message.js`: a leading `Spacer(1)` before visible content, then `Markdown(text, outputPad, 0)` on the terminal background.
- `tool-execution.js`: a leading `Spacer(1)` and default `Box(1, 1)` or `Text(..., 1, 1)` with an outcome background. Custom self-rendered tool shells are a separate contract.

Gi now applies those defaults at message-block boundaries. User messages have a blank colored row above and below their content and one column of horizontal padding. Assistant messages have a blank uncolored separator and horizontal padding. Tool/bash/error blocks have an external blank separator plus colored top/bottom padding and horizontal padding. The external spacer never inherits the outcome background. Blocks remain borderless.

The editor separators, footer and idle row count are unchanged. Longer transcripts scroll through the existing rendered-height path.

## Grouping and navigation

Gi's Markdown projection already indents continuation rows by the speaker-prefix width. The block builder now groups those rows into their user/assistant message before applying padding. Paragraphs, lists and code blocks keep their existing internal line breaks; they do not acquire a separator around every projected line and ordinary multiline messages do not collapse as tools.

The user-prompt marker in the rendered search index moves from the block's first row to its first content row, after top padding. Prompt jumps, search row counts and anchors therefore use the corrected layout. Both fullscreen rendering and regular-mode printed history use the same padded block renderer.

This keeps Gi's existing speaker labels and Markdown projection. Full Pi typography, mixed-inline reflow and other previously documented rendering limits are not established by this slice.

## Evidence

Rendered-buffer tests at 60, 100 and 140 columns check exact heights, blank top/bottom rows, background coverage at both edges and the middle, and the absence of box glyphs. Grouping tests verify paragraph/list/code continuation rows remain inside their parent message. Search tests compare the rendered row count and prompt positions with the scroll-region layout.

`make test-tui-outcomes` now checks real terminal captures for the user padding, external tool separator, tool padding and assistant separator, as well as unchanged editor/footer row positions. Native outcome/search/reading/regular-scrollback/session/model/compaction acceptance passes at 60×18, 100×22 and 140×36. Full Go tests/vet, terminal race suite ×3, terminal smoke/Gherkin, 29 helpers and 70/70 functional browser tests pass.

Fresh real XTerm screenshots at all three sizes are under `/workspace/tmp/gi-ui-captures/september22/tui-pi-padding-*.png` and attached to the conversation. Text/ANSI spacing evidence is in `test-results/tui-outcomes/*-message-spacing.*`.

No web source or frozen browser criterion changed. Browser coverage stays 34/236 Classic and 2/42 shared, with 202/40 unmapped. The latest complete browser parity matrix remains the 384/384 run for `4071c01`; only the 70-test functional browser suite was rerun in this terminal slice.

## Unfinished work kept separate

Fullscreen selection/copy/edge-autoscroll is not shipped here. Its first three-size matrix passed, but the expanded tool-click case still fails. That work is saved in `/workspace/tmp/gi-selection-wip.patch` and `/workspace/tmp/gi-selection-wip-new-files.tar.gz`; it needs manual patch application, updated padded hit-region handling and complete regression verification.

Folder references remain saved in `/workspace/tmp/gi-folder-reference-wip.patch`. Light theme, stronger reflow/eviction anchors, per-occurrence/cross-wrap search and the broader web feature families remain open.
