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

Gi's slots mirror PiSwift's `setStatus`/`setWidget`/extension-footer behavior while preserving the no-top-chrome layout:

- Slots may only add content to the **bottom band** (below the editor), the **widget area** (between transcript and editor), or the transcript.
- Slots may never add top chrome or a header region.
- Extension footer status segments render as extra dim footer lines after the stats/notice rows.
- Extension widgets render as a small block between the transcript and the editor, inside the bottom-band height budget.

## Implementation

`internal/tui/chat.go`:

- `chatTUI.extensionStatuses map[string]string` holds keyed extension status segments.
- `setExtensionStatus(key, text)` sets or clears (empty text removes) a segment; text is sanitized to a single line.
- `extensionStatusLines()` returns segments sorted by key.
- `footerLines(width)` appends extension status lines after the path/stats/notice rows, each truncated to width.
- `chatTUI.extensionWidgets map[string][]string` holds keyed multi-line widgets.
- `setExtensionWidget(key, lines)` sets or clears a widget; `extensionWidgetLines()` returns widget lines sorted by key.
- `Render` draws widget lines as a block between the transcript and the editor separator, and includes their height in the reserved bottom-band budget.
- `chatTUI.extensionToolModes map[string]string` holds per-tool render modes.
- `setExtensionToolRender(tool, mode)` chooses how a named tool's block body renders: `full` (default), `compact` (first body line only), or `hidden` (header only). `applyToolRenderMode` is applied when building tool blocks.
- `handleTopicEvent` handles the `extension.status` topic (`{key, text}`), the `extension.widget` topic (`{key, lines}` or `{key, text}`), and the `extension.tool_render` topic (`{tool, mode}`), so extensions can drive all slots through the topic bus.

## Constraints proven by tests

`internal/tui/chat_test.go`:

- `TestExtensionStatusSlotAddsFooterRowsOnly`:
  - setting statuses adds footer rows;
  - statuses are sorted by key;
  - clearing a status via the `extension.status` topic removes its row;
  - setting status writes **no** transcript rows, so it cannot create top chrome.
- `TestExtensionWidgetSlotRendersAboveEditorOnly`:
  - the `extension.widget` topic sets/clears keyed widget lines;
  - text and line-array payloads both work;
  - setting a widget writes **no** transcript rows, so it cannot create top chrome.
- `TestExtensionToolRenderSlotControlsToolBody`:
  - the `extension.tool_render` topic switches a tool block between full/compact/hidden body;
  - the slot only changes how an existing tool block renders, never adding chrome.

## Next slots (deferred)

- editor replacement for bounded prompts/forms.

Each future slot must keep the same constraint: bottom-band/widget/transcript only, never top chrome.
