# ADR 0057: Integrated message copy and deletion evidence

Status: Accepted (shared web contract)
Date: 2026-09-23

## Decision

Map `@shared-37` only with a single browser scenario that exercises the supplied message and code copy controls and Gi's native idle-only DELETE against persisted messages. Keep `@ux-original-024` unmapped: it requires reply-cascade confirmation and deletion that Gi's flat timeline does not implement.

The main session contains a persisted user Markdown message with a fenced code block and an assistant reply. A child session contains separate messages. The editor holds an unsent draft, a pending file and a reference to the unrelated assistant reply. The test uses keyboard and pointer controls, checks stored source bytes from actual copy events (including rich HTML for message copy), and checks success/error glyphs returning to idle. It denies clipboard mechanisms for the error paths rather than inventing clipboard contents. A real local provider gate keeps a turn running while the visible Delete action calls the native endpoint and receives 409; the rejected row remains. After releasing that gate, the same captured ID is deleted. The unrelated message, reference, file, draft and child-session history stay intact, including after reload.

## Terminal disposition

Browser clipboard success does not prove terminal source-copy fidelity. Keep source-copy and selected-message deletion as separate temporary actions using the existing transcript selection or bounded picker, with no permanent buttons, footer rows or top chrome. Verify stable message IDs, explicit deletion confirmation, idle admission, clipboard policy and failure feedback, and draft/cursor/reader preservation in fullscreen and regular modes at 60×18, 100×22 and 140×36 before terminal credit. See the existing [terminal layout contract](../internal/tui-pi-layout-contract.md) and [ADR 0056](0056-idle-single-message-deletion.md).

## Evidence

`make test-ux-shared-copy-delete` runs `tests/ux/shared-copy-delete.spec.mjs` against the local gated provider. Six projects pass: Chromium and WebKit at phone, tablet and desktop sizes. The combined parity report includes six isolated main-project results, reconnect, context, compaction, Steer and shared-37 results; it reports **51/236 Classic**, **3/42 shared**, **185/39 unmapped**. The frozen feature files and supplied components are unchanged. No terminal deletion or source-copy implementation is included.
