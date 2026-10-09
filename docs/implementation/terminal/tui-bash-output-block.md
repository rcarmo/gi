# Gi live bash output block (PiSwift port)

Status: implemented for the local `!!command` shell shortcut.

## Goal

Match PiSwift's bash execution component behavior in Gi's existing transcript-block model while preserving the no-top-chrome Pi layout and the single bottom status row.

PiSwift reference:

- `Sources/PiSwiftCodingAgentTui/Modes/Interactive/Components/BashExecutionComponent.swift`
- `Sources/PiSwiftCodingAgentTui/Modes/Interactive/Components/ToolExecutionComponent.swift`

PiSwift behavior ported:

- bordered bash block with a `$ command` header;
- trailing-line preview when collapsed;
- full output when expanded;
- skipped-line hint with an expand affordance;
- a full-output file reference when output is truncated.

## Gi implementation

`internal/tui/chat.go`:

- `localShellShortcutLines` now runs the local command and emits a `bash` transcript block via `bashBlockLines` instead of a flat `local$` block.
- `bashBlockLines(command, output, status, runErr, startedAt, endedAt)`:
  - normalizes CRLF/CR to LF;
  - retains the full output in the block body, capped at 500 lines;
  - when capped, writes the complete output to `<workspace>/.gi-run/bash-output/bash-<ns>.txt` and adds a truncation footer;
  - appends an error line when the command failed;
  - encodes a `transcriptBlockMeta` with kind `bash`, the `$ command` title, status, start/end timestamps, and an optional footer.
- `transcriptRenderableBlock` gained `PreviewLimit`, `PreviewTail`, and `Footer` fields.
- The block builder sets `PreviewLimit = bashPreviewLines (10)` and `PreviewTail = true` for bash blocks, so collapsed bash blocks show the last 10 lines.
- `renderTranscriptBlock` honors `PreviewLimit`/`PreviewTail`, renders a `… N more line(s) · F8 expand` hint when collapsed, an `F8 collapse` hint when expanded, and the block footer line.
- `bash` has a dedicated palette color; completed bash blocks do not render an `[ok]` status label.

## Interaction

- `!!command` runs locally and renders a bash block.
- `F6`/`F7` select transcript blocks; `F8` expands/collapses the selected block; clicking a block toggles it.
- Collapsed bash output shows the trailing preview; expanded shows full output; truncated output references a full-output file.

## Tests

`internal/tui/chat_test.go`:

- `TestLocalShellShortcutRunsLocally` asserts the bash block meta and body.
- `TestBashBlockPreviewTailAndExpand` asserts tail preview window, full-body retention, and expandability.

## Layout constraints preserved

- No top chrome; transcript-first.
- Single bottom status row unchanged.
- The bash block lives in the transcript, not in permanent chrome.
