# ADR-0053: Separate upload progress and message sending

## Status

Accepted — 2026-09-23. Frozen `@ux-compose-005` passes on Chromium/WebKit at phone, tablet and desktop sizes. Coverage is **46/236 Classic**, **2/42 shared**, with **190/40 unmapped**. Supplied components, panes, UI modules and frozen feature files are unchanged.

## Native transport state

`gi-compose-transfer.ts` tracks transient upload and message-request operations by destination session and unique operation token. Concurrent completion removes only its own operation; another upload or send stays visible. Tokens are not persisted, and reload never restores activity as executable work. Draft recovery remains the existing durable repository's responsibility.

`uploadMedia` validates the destination and existing 10 MiB file bound, then materialises the browser-generated multipart body as an ArrayBuffer and sends it with XMLHttpRequest. Browser-generated boundaries and the matching content type remain intact. XHR upload events supply actual loaded/total bytes; no clock or simulated percentage is used. Before a computable total arrives, progress is indeterminate. At 100%, the upload still waits for a successful native response and media identifier; the label explicitly says “awaiting server”. The percentage covers the multipart HTTP body, not just file payload bytes.

Materialisation adds bounded temporary memory for the file plus multipart body. The composer uploads a batch sequentially. The UI reports currently active requests and their combined byte totals, not an invented whole-batch percentage. Concurrent submissions can show upload and sending rows simultaneously for different operations.

WebKit's file-backed XHR FormData body lost file bytes when Playwright interception forwarded it: the intercepted body had only multipart headers. Direct native uploads preserved bytes. Materialising the browser-generated body resolved forwarding without synthesising test responses or multipart content. Native byte checks now pass both with and without routing in both engines.

`sendAgentMessage` starts a separate operation immediately around the native prompt request. Upload activity ends before the single submission's prompt request begins. Sending ends on HTTP/JSON outcome, not on provider completion. Both paths end their token in `finally`, including parser rejection, network failure and malformed responses.

## Host presentation

`ComposeTransfer` in `app.ts` renders session-owned upload and sending status outside the supplied composer. It hides in search mode and occupies no idle height. Changing sessions cannot display the prior session's transfer; returning while it remains active restores the appropriate status.

A host DOM adapter decorates only `data-gi-sending` and `aria-busy` on the supplied `.compose-send-stack .send-btn`. A scoped CSS ring gives the button a distinct sending state; reduced-motion settings disable animation. Existing icon, label, disabled state and Send/Stop/Compact click behaviour remain owned by the supplied component. Newer text, attachments and cursor remain editable; a pending send does not disable another submission. A child-list observer reapplies host attributes if draft restoration remounts the button and removes owned attributes on cleanup.

Upload failure never creates the prompt request. Native multipart errors retain the existing draft recovery/error path. A successful HTTP response does not imply provider completion, and a network-uncertain prompt outcome retains the existing “check timeline before resending” behaviour.

## Evidence

The compose-005 browser test holds real multipart and prompt requests separately. It checks upload status without a sending button, then no upload status when the prompt route is reached, then distinct busy button/sending state. It switches to a child session during each phase, returns to the origin, checks newer typing/caret preservation, verifies native media ID/bytes and confirms no transient activity after completion/reload.

Additional tests exercise:

- un-intercepted native upload bytes and real computable XHR progress events;
- an upload overlapping a second message request, two concurrent sends and independent completion;
- native Go multipart rejection with progress cleanup and retained retry draft;
- existing multiple-file ID/name/byte pairing, partial failure, session/reload recovery and explicit retry.

All **138 draft/upload executions** passed. Full browser suites passed **516 executions**: 348 main, 54 reconnect, 66 compaction, 18 Steer/admission, 12 context-fit, 12 meter and six index-configuration. The first main-suite command timed out after 251 executions without a test failure; the subsequent complete 348-execution run passed. A final six-project compose-005 capture rerun also passed. Combined native reports map only this new frozen ID; neighbouring compose-004 remains a gap.

`make check BIN_DIR=/tmp/gi-scheduler-bin` passed Go tests, vet, web build, hook checks and **78 functional tests**. **34 helper tests** passed, including independent operation/session lifetime, unknown byte totals and unchanged frozen hashes. The delegated review timed out and supplied no evidence. Reconnect's harness now honours `UX_LOCAL_BIN` through `GI_UX_SERVER_BIN`, avoiding workspace disk exhaustion and replacement of another harness binary.

Logs are `/workspace/tmp/gi-compose-progress-*`; combined results are in `test-results/ux-parity/matrix.{json,md}`. Native main JSON was retained at `/tmp/gi-compose-progress-main.json` before capture reruns. Upload/sending screenshots for all three browser sizes are under `/workspace/tmp/gi-ui-captures/september23/compose-{upload,sending}-{phone,tablet,desktop}.png`.

## Terminal adaptation

Use an existing transient status slot to distinguish pending attachment transfer from submission, scoped to the captured session/draft and hidden at idle. Report bytes only if the native terminal transfer exposes actual byte events; otherwise use an indeterminate label. Never disable newer typing or add a permanent attachment/progress row. Pending-media admission, byte pairing, cancellation/failure/retry, reader/cursor restoration and three-size PTY evidence are still required. This browser slice adds no terminal implementation or acceptance credit.
