# Classic lightbox dismissal 013–016: bounded clause review

Shipped Piclaw 3.2.4 `components/image-modal.ts` is byte-identical to the
frozen 70d33bc component (`frozen-to-oracle-sources.json`). It closes on
Escape and on a click inside the modal wrapper, including the image. Gi uses
the supplied `web/src/components/image-modal.ts` through `BodyPortal` with
native stored-media URLs. A mounted installed Classic 3.2.4 lightbox probe
now checks the same dismissal gestures against a disposable image post.
Gi browser tests and installed-fixture evidence remain separate; neither is
physical-touch acceptance.

| Frozen ID | Gi assertion → code | Limit |
|---|---|---|
| `@ux-timeline-013` | Gi `tests/ux/lightbox.spec.mjs:31–35` opens a real stored image, presses Escape, checks dismissal and timeline, draft and attachment preservation, including after reload. Installed Piclaw mounted fixture also opens a lightbox and Escape dismisses it with a retained unsent draft. | Installed fixture does not test reload or stored attachment ownership. No screen-reader focus-trap claim. |
| `@ux-timeline-014` | Gi `lightbox.spec.mjs:36–40` presses Space, Enter, a letter and an arrow; installed Piclaw fixture checks Space, Enter, x, ArrowDown and ArrowLeft. Both keep the lightbox open and retain the draft. | No IME/hardware keyboard pass or installed backend queue assertion. |
| `@ux-timeline-015` | Gi and installed Classic fixtures each click the backdrop and image, reopening between them; both close the lightbox. The wrapper's click handler receives the image click. | Browser pointer only; installed image is a disposable media response. |
| `@ux-timeline-016` | Gi `lightbox.spec.mjs:78+` uses `hasTouch:true`, taps backdrop and image, and checks four trusted single-touch starts while retaining native media and drafts. Installed Classic fixture also opens by tap, closes from backdrop/image taps and retains its draft. | Installed probe does not count trusted touch-starts. Emulated browser touch is **not** a physical phone/tablet or OS gesture pass. |

The fixture uploads `native-image.png`, reads the native stored bytes back,
and leaves an unsent draft with `keep.txt`. `make test-ux-parity
UX_PARITY_PORT=19134 UX_PARITY_ARGS='tests/ux/lightbox.spec.mjs --grep
"@ux-timeline-01[3-6]"'` passed **24/24** across Chromium/WebKit
phone/tablet/desktop projects. The extended
`make test-piclaw-non-ipad-lightbox` passed **6/6** installed Classic
Chromium/WebKit viewport cases for Escape, non-Escape keys, backdrop/image
click and tap. It uses disposable media, an emulated non-iPad identity and a
browser tap. Piclaw reads `/media/:id`; Gi reads `/api/media/:id/raw`. These
results support fidelity of the existing dismissal behaviour, not a real
Piclaw media store, physical touch, iPad annotator or whole-image workflow.
