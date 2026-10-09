# Classic028: code copy and speech ownership

`web/src/components/post.ts` supplies a code-block copy control and the
read-aloud button. `web/src/gi-post-speech.ts` holds one active owner, cancels
prior playback and fences callbacks from an older utterance. Gi speech text
omits fenced code and has a 1600-character cap.

The tagged `@ux-original-028` case in `tests/ux/speech-contract.spec.mjs`
checks a trusted plain-text code-copy event without HTML, then transfers
speech ownership between assistant posts while preserving native draft and
stored messages. Focused `test-ux-steer` run: **6/6** across Chromium/WebKit
phone, tablet and desktop.

`make test-piclaw-speech-ui` runs the installed 3.2.4 Classic UI in a
disposable timeline at the same six browser/viewport combinations: **6/6**.
Two speakable assistant posts expose Read aloud; user and empty assistant
posts do not, and an unsupported browser exposes no speech buttons. Switching
posts cancels the first, and late `onend`/`onerror` callbacks do not clear
the second post's active button. The first utterance omits fenced code, code copy supplies plain
source text without HTML, and the unsent draft remains. Piclaw calls
`cancel` before the first `speak`, as well as during transfer; Gi's existing
browser test expects `speak`, `cancel`, `speak` and tests the same observable
ownership transfer. Installed Classic switches its button label and active
class but does not emit `aria-pressed`; Gi does. Both probes stub the browser
speech and clipboard boundaries. Audible output and physical/assistive-device behaviour remain
untested.

`timeline-rendering-review.md` contains the narrower `@ux-timeline-024`,
`027` and `028` evidence. The shipped-UI probe adds current-oracle coverage
for the combined speech-transfer/code-copy actions. Production routing,
audible speech, assistive devices and a live session were not tested. No
production code or frozen Gherkin changed.
