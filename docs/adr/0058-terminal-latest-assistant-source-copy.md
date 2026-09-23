# ADR 0058: Exact source for terminal latest-assistant copy

Status: Accepted (bounded terminal adaptation)
Date: 2026-09-23

## Decision

Keep Pi-style `/copy` as the latest-assistant action; it has no new idle button, row, selector or permanent status. Send the exact persisted assistant message content to the selected native clipboard helper or OSC 52 transport. The earlier path trimmed leading and trailing whitespace. Skip empty and whitespace-only assistant messages as before. Do not change fullscreen rendered-cell selection, which uses visual-row breaks.

`/copy --osc52` and `/copy --native` retain the existing opt-in policy. Missing or denied helpers use the existing transcript fallback, never a success claim. The status line reports bytes sent. The fallback is a readable transcript projection and is not byte-exact clipboard output.

## Evidence

`TestCopyLastAssistantSourceRetainsWhitespaceAndModeFailure` checks latest-message selection, leading/trailing spaces and blank lines, Unicode, OSC 52 payload bytes, native helper payload, byte count and missing-helper fallback. `make test-tui-source-copy` launches six real tmux PTYs at 60×18, 100×22 and 140×36 in fullscreen and regular modes, seeds durable store rows, reopens Gi and decodes the emitted OSC 52 sequence from the pane byte stream. The six cases compare exact source bytes, check no new messages or idle rows, and retain an unsent Unicode draft after fallback. tmux may not import application OSC 52 into its own paste buffer; the byte-stream assertion checks the emitted sequence directly.

## Remaining work

Ordinary user and assistant transcript blocks currently have position/text keys; regular-mode history is terminal-owned. Neither supports a safe per-message delete target across reflow, eviction and session switching. A selected-message source-copy or confirmed delete action needs durable message-ID mapping, explicit focus/ownership and three-size fullscreen/regular acceptance. Browser `@ux-timeline-017` and `@shared-37` remain independent of this terminal subset. This change adds no feature-file or terminal per-message deletion credit.
