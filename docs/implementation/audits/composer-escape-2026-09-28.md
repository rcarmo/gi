## Composer Escape

The guarded build adapter adds the final Escape action present in installed Piclaw 3.2.4: blur the editor after search and autocomplete have had first refusal. Supplied component bytes are unchanged. No command is sent and the draft or attachments are not cleared.

The installed-browser probe passed six Chromium/WebKit phone/tablet/desktop cases: plain Escape blurs, the first autocomplete Escape dismisses suggestions while retaining focus, and the second blurs. It records the source-map provenance hash. The native pre-fix desktop run failed both focus checks; the updated native suite passed 12/12, including IME, consumed keys, picker ownership, draft/media retention and restoration of a committed draft. The support test verifies adapter composition and drift rejection.

One Chromium native run exposed a separate loss of the final typed character when reloading before its asynchronous IndexedDB write committed. That failure remains open and is retained in the evidence archive. The Escape reload assertion now explicitly waits for the stored draft before checking restoration; it does not establish immediate-reload durability. The first support run used the command adapter without its popup prerequisite and failed; the corrected chain passed. The initial oracle used the wrong reference metadata field and failed before opening a browser.

These browser tests observe focus, not physical on-screen keyboard dismissal. No deployment or restart occurred.
