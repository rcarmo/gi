# ADR-0035: Explicit workspace folder references

## Status

Accepted — 2026-09-22. Classic compose-008 now passes all six browser projects. Web coverage is 35/236 Classic and 2/42 shared; 201/40 cases remain unmapped.

## Host adaptation

The supplied workspace explorer uses directory clicks for navigation and expansion. It does not invoke its file-selection callback for directories. Gi now adds an explicit `+ folder` button, labelled `Reference selected folder`, to the existing workspace header. It appears only when a directory is selected and disables when that path is already in the current draft.

`useWorkspaceFolderReference` in `web/src/app.ts` observes the explorer's selected-row attributes and tree changes. The host owns the button and removes it and its observer on hide/session changes. The click reads the current selected directory, checks captured-session ownership, deduplicates the path and updates the existing durable draft. File clicks retain their existing behaviour. No supplied explorer, composer, pane or UI helper file changes.

The existing composer serialises the captured draft: trimmed multiline text, a `Files:` block containing file/folder paths, then a `Referenced messages:` block. References alone produce a non-empty submission. Folder references are textual paths, not uploads, directory snapshots or automatic recursive reads.

## Native acceptance

Three cases in `tests/ux/drafts.spec.mjs` run on Chromium/WebKit at phone/tablet/desktop sizes:

- **compose-008:** create native workspace entries and history, select a file, explicitly reference a folder through the keyboard-accessible action, and reference a stored message. Assert exact outgoing and stored multiline content and then exact references-only submission with no text/media. Acceptance clears the captured references.
- **Gi durability/isolation:** navigation alone does not attach; repeated attach is disabled; file selection hides the folder action; reload and session switches preserve separate drafts; removal re-enables the action; repeated open/close does not duplicate controls.
- **Gi failure recovery:** hold and fail a native submission while newer text is entered and another chat is selected. Recovery merges into the origin, retains its folder path across reload and does not alter the other chat or create a message.

Full validation: **402/402 browser executions** (246 main, 54 reconnect, 66 compaction, 12 Steer, 12 context-fit, 12 meter), **70/70 functional**, **29/29 helpers**, full Go tests/vet and hook checks. Main results combine two successful 123-case browser-family runs. Frozen feature text and hashes remain unchanged. The read-only review delegate timed out and supplied no review evidence.

## Regression fixes and unresolved boundary

The broader run exposed two existing timing assumptions:

1. Completed status and display-idle can precede native active-claim release. Rapid history-building prompts could become steering for the prior turn. The acceptance fixture now uses explicit queue intent for distinct setup turns and derives its post-count baseline from persisted messages, including legitimate native queue-status messages. Production admission semantics are unchanged. The display-idle/admission-ready distinction remains an open engine/API boundary and is tracked in the checklist.
2. After reload or rapid selection, old request counts could satisfy fixture waits before the current chat's native SSE subscription was ready. The fixture now observes real `connected` events for the selected chat before counting/disconnecting. Gi also marks the connection disconnected synchronously on selection, instead of displaying the old chat's connection state. Initial activation passes 18/18 repeated executions and the final reconnect matrix passes 54/54. Disconnect still destroys the real proxy socket; the temporary EOF experiment and debug logging were removed.

An earlier focused run also hit a WebKit internal reload error before folder interaction. The unchanged folder cases passed subsequent full runs. No forced clicks, fabricated SSE, altered frozen assertions or paid-provider dependency were used.

## Terminal adaptation

No terminal code or terminal acceptance credit in this slice. Keep folder references inside the existing editor through path completion or an explicit command; do not add an idle reference panel. The current `completeInputPath` already retains an `@` prefix and adds a trailing slash to directories, but does not establish the browser's captured `Files:`/message-reference serialization or durable failed-send recovery. Those require independent native tests at 60×18, 100×22 and 140×36 before terminal parity credit.

The earlier saved folder-reference patch is superseded by this implementation. Hashtags, further workspace mutations/editor contracts, crash recovery and the remaining terminal differences are still open.
