# ADR-0055: Whole-message copy and truthful clipboard failure

## Status

Accepted — 2026-09-23. Whole-message copy now has native browser evidence for source Markdown, rich clipboard data and fallback/failure handling. A missing Clipboard API no longer produces false success when legacy copy also fails. This is a partial implementation of combined copy/delete scenarios: frozen original-024 and shared-37 remain unmapped until deletion/cascade criteria pass. Coverage remains **50/236 Classic**, **2/42 shared**, **186/40 unmapped**.

## Narrow compatibility fix

The supplied `post-runtime-safety.ts` helper awaited `clipboard?.writeText?.(value)` and returned true even when neither the clipboard object nor its method existed. With `execCommand('copy')` unavailable/false, the post button therefore displayed Copied without a clipboard write.

`web/src/gi-clipboard-safety.ts` requires a callable `writeText`, invokes it with its receiver intact, awaits its result and returns false for missing methods, throwing getters or rejected writes. It re-exports all other helpers unchanged. This does not guarantee an OS clipboard accepted bytes merely because an API resolved; it correctly distinguishes an absent API from a successful call.

`build.js` uses `Bun.build` for the app bundle and a single exact-importer resolver rule: only the Post component's `./post-runtime-safety.js` import resolves through the compatibility adapter. The adapter itself imports the original helper module, avoiding recursive resolution. Vendor builds, browser/ESM target, external CodeMirror path, linked sourcemap and embedded output names are unchanged. Supplied components, UI utilities and panes remain unedited. A bundled-browser unavailable-API test verifies the override is actually active, in addition to direct helper tests.

## Copy contract and evidence boundary

The supplied copy control keeps its existing priority: rich legacy copy, rich asynchronous Clipboard API, then plain-text fallback. The copied plain text is the stored message content with CR/CRLF normalised to LF and trailing whitespace trimmed, as specified by the supplied Markdown payload builder. It retains Markdown syntax and fenced code. Rich `text/html` contains rendered formatting and the supplied clipboard stylesheet; it is not used as the plain-text payload.

Tests submit real prompts, wait for native completion, read the stored assistant message and click that post's actual copy button. A passive document copy-event listener reads the payload while the event is valid. It does not replace clipboard handlers or APIs in success tests. The event must be trusted and contain the exact native Markdown plus rich HTML. The newer composer text, caret, attachment pill and stored messages remain unchanged; reload retains the draft and media.

Failure tests deliberately remove or reject clipboard mechanisms. They prove honest error feedback, not OS clipboard availability. Rich-copy denial followed by an unmodified native legacy plain-copy attempt produces a trusted event with exact Markdown and no HTML. Denying every mechanism yields Copy failed for whole messages and code, then returns to the idle label. A held asynchronous rejection after a session switch cannot change the destination post/draft or replay on return.

These tests observe the browser's real copy event and payload. They do not read a desktop OS paste buffer or claim cross-application paste-format compatibility.

## Validation

- **30 message-copy executions** across Chromium/WebKit at phone/tablet/desktop sizes cover success, absent APIs, rich-to-plain fallback, denial/reset and late session-switch failure.
- The combined focused message-copy/rendering suite passed **54 executions**, preserving existing source-code/SVG copy assertions.
- Full browser validation passed **594 executions**: 426 main in three saved 142-execution project batches, 54 reconnect, 66 compaction, 18 Steer/admission, 12 fit, 12 meter and six index config.
- Go tests, vet, web build, hook checks and **80 functional tests** passed through `make check BIN_DIR=/tmp/gi-scheduler-bin`.
- **38 helper tests** passed, including absence/denial/receiver handling and supplied source hashes. No native Go runtime code changed; no new race or terminal matrix is credited.

An earlier clipboard observer incorrectly read the event's data in a microtask after the browser invalidated it; reading synchronously during the event corrected the fixture. The unavailable-API failure was a genuine production bug fixed by the adapter. A code-button selector was corrected to the supplied class.

A broader run exposed the existing Quick Actions failure fixture replacing route handlers during reload: its trace showed HTTP 200 rather than the intended rejection. The test now keeps one handler, drains held calls, switches to rejection mode and asserts at least two aborted requests before checking conservative UI. **18 repeated executions** and the rerun main matrix passed; its zero-unsupported-actions assertions are unchanged. Delegated review timed out and supplies no review evidence.

Logs are `/workspace/tmp/gi-message-copy-*`. Saved main reports are `/tmp/gi-message-copy-main-{chromium-mobile,mixed,webkit-large}.json`; the combined report is `test-results/ux-parity/matrix.{json,md}`. Captures are `/workspace/tmp/gi-ui-captures/september23/message-copy-{phone,tablet,desktop}.png`.

## Terminal adaptation

Existing `/copy` selects the latest assistant's stored content and applies the configured native/OSC52 policy; fullscreen drag-copy selects rendered rows. Those are distinct contracts: rendered selection inserts visual-row breaks, and source-copy must not acquire those breaks. This browser slice changes neither path.

Any future per-message source-copy action should use an explicit shortcut or bounded temporary selector and existing transient feedback. Missing helpers, clipboard opt-out and denied writes must not announce success. Preserve editor/cursor/reader and add no permanent copy button, status row or panel. Exact source whitespace, session ownership, failure feedback and three-size fullscreen/regular evidence are required before terminal source-copy parity credit.
