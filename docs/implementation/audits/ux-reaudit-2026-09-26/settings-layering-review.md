# Classic settings layering 001–004: bounded overlay trace

`features/ux/classic/settings/settings-layering.feature` scopes backdrop,
dialog and workspace interaction to Classic, not other skins. The installed
Piclaw 3.2.4 Settings backdrop is now browser-probed with a disposable body
control. Gi's Settings dialog modules have their own implementation.

Gi's `web/src/gi-settings.ts` renders a `BodyPortal` and modal backdrop,
inerts the app while open, traps focus and restores prior state on cleanup.
`internal/web/static/css/gi-settings.css` fixes the portal/backdrop across the
viewport, stacks them at 12000 and uses `rgba(0, 0, 0, 0.5)` for the backdrop.
The four tagged cases in `tests/ux/settings-shell.spec.mjs` each use a real
workspace file and trusted pointer interaction. They check coverage geometry,
computed opacity, dialog center hit-testing and viewport bounds, no workspace
click or file read through the overlay, then restored workspace interaction
after dismissal. Focused run: **24/24** across six Chromium/WebKit viewport
projects.

`make test-piclaw-settings-layering` passed **6/6** mounted Classic cases:
its fixed backdrop covers the viewport at z-index 12000 with computed
`rgba(0, 0, 0, 0.5)`; the dialog center wins hit-testing, and a trusted
pointer to a disposable body button is blocked while Settings is open and
works again after dismissal. The first phone measurement caught the installed
slide-up transform before it settled; the bounds assertion waits 250 ms.
The installed ordinary-open path did not render a `.settings-portal` element,
so it tests backdrop stacking, not portal-parent equivalence with Gi. Backdrop
dismissal used a synthetic click; pointer blocking and restoration used
`page.mouse.click` at the fixture button.

Gi's **24/24** run used a real workspace file and checked that no workspace
click or file read crossed the overlay. The installed fixture has no real
workspace explorer or file read. Neither run establishes physical pointer,
Visual-skin stacking, or whole-clause parity. Frozen Gherkin and production
code were unchanged.
