# Reordered inline prose in screenshot 3139

The retained manual-validation binary reproduces the screenshot's reordered inline path and surrounding prose. Fresh main and the separate `/workspace/projects/gi/gi` binary pass the same six cases. The screenshot alone does not identify its exact executable.

## Reproduction

`features/gi/tui/inline-prose-layout.feature` describes an assistant sentence containing `internal/tui`, followed by two shell results. `scripts/test-tui-inline-prose.mjs` loads it into a disposable SQLite database and opens the real TUI in tmux at widths 60, 100 and 140, in fullscreen and regular modes.

The check reads rendered rows in order, requires every word and the inline path exactly once, verifies multi-row wrapping at width 60 and single-row fit at width 140, requires the two tool blocks immediately afterward in stored order, checks for absent decorative output boxes, resizes with an unsent draft, and verifies that rendering creates no messages. It does not submit prompts or use provider credentials.

- Retained `/workspace/tmp/gi-autosave-validation/bin/gi-integrated`: **0/6 passed**. At width 100 the displayed sentence became `I added these tests under changes untouched. and the suite passed. I left the other uncommitted`, losing the inline path and reordering continuation text. At width 140 it stopped after `internal/tui`.
- `/workspace/projects/gi/gi`: **6/6 passed**. Its embedded VCS metadata reports `dc809ee`, modified; therefore its source contents cannot be inferred from that commit alone.
- Fresh `/workspace/projects/gi-main/bin/gi`, built from main `2f305b7`: **6/6 passed**, including resize and draft checks.

The synthetic text is an analogue of the screenshot, not a transcription or a replay of the user's private session. Evidence is in `/workspace/tmp/gi-tui-3139/{integrated,old,main}/`, including plain and ANSI captures and JSON outcomes.

## Existing fix and new regression coverage

Consolidation commit `41f87fc` already brought in the relevant rendering changes: styled inline text is rendered as one rich-text element, rather than independently laid-out fragments, with code spacing preserved. It also removes decorative boxes around assistant and tool output. This follow-up adds the missing multi-line ordering test and integrates it into `make test-tui-gherkin`; it does not add another speculative renderer patch.

Commands:

```sh
make test-tui-inline-prose
GI_TUI_BIN=/workspace/tmp/gi-autosave-validation/bin/gi-integrated \
  GI_TUI_PROSE_OUTPUT=/workspace/tmp/gi-tui-3139/integrated \
  make test-tui-inline-prose-binary
```

An independent review found missing explicit wrapping and tool-order assertions. Both were added; the strengthened six-case run passes and is retained in `/workspace/tmp/gi-tui-3139/reviewed-main/`.

The old manual-validation executable has not been replaced or restarted. Passing this stored-transcript test does not establish streaming, scrolling, viewport-edge clipping, all terminal emulators, or the broader Piclaw web UX contract. The earlier stray `bytes).` fragment also remains unexplained.
