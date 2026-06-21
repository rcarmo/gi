# Gi TUI extension slots (PiSwift port, phase 1)

Status: first backend-safe slot implemented — extension footer status segments.

## PiSwift reference

PiSwift's interactive hook UI context exposes many TUI slots:

- `setStatus(key, text)`
- `setWidget(key, content)`
- `setFooter(factory)`
- `setEditorComponent(factory)`
- overlays and custom renderers.

`FooterComponent` renders extension statuses from `footerData.getExtensionStatuses()` as extra footer lines.

## Gi slot contract

Gi's first slot mirrors PiSwift's `setStatus`/extension-status footer behavior while preserving the no-top-chrome layout:

- Slots may only add content to the **bottom band** (below the editor) or the transcript.
- Slots may never add top chrome or a header region.
- Extension footer status segments render as extra dim footer lines after the stats/notice rows.

## Implementation

`internal/tui/chat.go`:

- `chatTUI.extensionStatuses map[string]string` holds keyed extension status segments.
- `setExtensionStatus(key, text)` sets or clears (empty text removes) a segment; text is sanitized to a single line.
- `extensionStatusLines()` returns segments sorted by key.
- `footerLines(width)` appends extension status lines after the path/stats/notice rows, each truncated to width.
- `handleTopicEvent` handles the `extension.status` topic with payload `{key, text}`, so extensions can drive the slot through the topic bus.

## Constraints proven by tests

`internal/tui/chat_test.go`:

- `TestExtensionStatusSlotAddsFooterRowsOnly`:
  - setting statuses adds footer rows;
  - statuses are sorted by key;
  - clearing a status via the `extension.status` topic removes its row;
  - setting status writes **no** transcript rows, so it cannot create top chrome.

## Next slots (deferred)

- widget row above the editor (still inside the bottom band budget);
- editor replacement for bounded prompts/forms;
- custom tool result renderer.

Each future slot must keep the same constraint: bottom-band/transcript only, never top chrome.
