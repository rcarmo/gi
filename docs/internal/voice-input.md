# Browser voice input

Gi exposes browser-native voice input when a secure browser provides `SpeechRecognition` or `webkitSpeechRecognition`. The control appends recognition text to the captured draft; it never sends a message automatically. Real microphone capture, native permission prompts and recognition-service availability are unverified.

## Capability and privacy boundaries

The host-only adapter follows the speech support and microphone icon in pinned Classic `0afe5366ced9`. No supplied component, UI utility or pane source changed. No audio endpoint, recording store or background permission probe was added. Browser recognition may use an external service; the button tooltip says so. Constructing and starting recognition requires a click, keyboard activation or non-mouse pointer gesture.

In insecure contexts or desktop browsers without a recognition API, the control is absent. iPhone/iPad standalone mode, or iOS without an in-page recognition API, exposes keyboard-dictation guidance and focuses the draft without constructing a recognizer. These capability decisions are tested with controlled runtimes; physical-device acceptance is separate.

Mouse/keyboard activation toggles recognition. Non-mouse pointer down starts push-to-talk; release requests stop and suppresses the following synthetic click. Release before permission confirmation waits for the browser start callback before stopping. Pointer cancellation aborts. The controller bounds pending start at15seconds, stop completion at5seconds and total run at90seconds. Errors/timeouts retain draft text and allow explicit retry; there is no automatic restart.

## Draft and lifecycle ownership

`gi-voice-input-state.ts` owns one recognizer/run token. Each callback verifies that the current session, draft text and surface eligibility still match its captured owner. Web Speech results are cumulative: final/interim text is rebuilt from the result list rather than repeatedly appended. The latest visible interim transcript is retained if recognition ends or is cancelled.

Manual typing, submission, Escape, search, modal appearance, hidden/detached surface, page hiding and unmount retire the run. Session changes cannot redirect a late callback into another draft. Media/reference state remains in the existing composer. Observers and global release listeners exist only during active recognition; unmount removes handlers and aborts the native object. Existing compaction disables voice input. Native speech playback remains a separate feature.

## Verification

- Eight helper tests cover capability decisions, cumulative results, immutable ownership, stale callbacks, early release, timeouts, errors/retry, disposal and guarded patch anchors.
- `make test-ux-voice-input`:30 Chromium/WebKit × viewport cases pass using a controlled recognition engine. Covers gesture-only start, transcript replacement, manual-edit precedence, retained attachments, no prompt admission, denial recovery, modal/search/pagehide exclusion, touch release/click suppression, unsupported/iOS fallback, session switch and reload persistence.
- Existing compose surface30, slash72 and phone draft56 cases pass. The first draft run passed55 but WebKit raised an internal error during the last reload of upload-cancellation recovery; the unchanged full rerun passed. A separate build attempt hit `text file busy`; rerun used an isolated binary path without killing any unknown process.
- Go, vet, hook checks and14pixel helper tests pass. Stock functional suite:107passes and11fixture-dependent skips.

Initial adapter render failures referenced state identifiers absent from the older supplied composer; they were corrected before the completed browser run. A session-switch test initially assumed focus behaviour not promised by that path; it now verifies session/draft ownership while existing focus suites remain unchanged. A narrow delegate review timed out; no review approval is claimed.

Pixel run `run-1790379403189-394956`:72captures completed without fixture/page errors; all18cross-host full frames differ and12/36repeat pairs are unstable. Compose-region changed pixels: phone950/1472, tablet968/1528, desktop994/1785 (light/dark). Capability-dependent controls and rendering instability prevent attributing those counts solely to voice input. Official renderer flags, exact RGBA criteria and masks remain unchanged.

The pinned feature corpus covers speech playback, not this voice-capture event contract; no new formal feature mapping is claimed. Notifications, real audio/prompt/device acceptance and Visual remain open. TUI adaptation is terminal/OS dictation as ordinary paste/input; no terminal microphone indicator or idle row is added. CI must pass before deployment; live Gi remains on the previously verified contrast build during this work.
