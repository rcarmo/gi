# Thinking transcript width and reflow

Thinking (`thought`) blocks retain their original accumulated Markdown in the
transcript marker's `markdown_source`, including after completion. Projection
uses `transcriptBlockContentWidth(kind)`: the current transcript width less both
horizontal message-band paddings, clamped to at least one cell. With today's
one-cell thinking padding, the projection width is `W - 2`, not `W`.

The renderer reprojects the source at the current inner width. It must not reuse
rows projected at an earlier outer width: rich-text wrapping those rows again
creates short orphan-word lines, while no-wrap table/code rows can be clipped.
Assistant source-backed blocks use the same width helper. Existing blocks with
no source retain their legacy body rendering; original paragraph boundaries
cannot reliably be recovered from already wrapped text.

Both fullscreen and regular-mode preview/commit use this renderer. Search and
selection's offscreen row rendering carry the same source and use the requested
viewport width without leaking that temporary width back into the live UI.

## Regression evidence

`thinking_wrap_test.go` renders real go-tui buffers through the transcript row
index. It checks streamed accumulation, widths 28/40/80/100 and repeated resizing,
exact prose line breaks, retained source/status after completion and intact table
borders with both margins. Both tests fail against the original renderer and pass
with this fix. This is buffer-level functional evidence, not a direct Ghostty
visual test or a resolution claim for the separate emoji redraw issue.

The full TUI suite currently fails `TestDurableHistoryDoesNotRecordRejectedRoute`
on this branch's base (`b1355dc6`): its peer-routing rejection assumption conflicts
with the new composer routing behavior. That test and routing code are untouched.

## Composer navigation

Editing new composer input returns the transcript to the bottom and resumes
following new output. Programmatic history/draft restoration preserves a reader's
scroll position. `TestInputChangeReturnsTranscriptToBottom` and
`TestInputRestorePreservesTranscriptPosition` cover this distinction.
