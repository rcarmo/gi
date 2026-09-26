# Native model-panel adaptation

Gi's model panel now matches the pinned Classic outer bounds in phone, tablet and desktop captures. Exact pixels still fail. The adapter exposes native model selection and a Models-settings handoff; pinning, pricing and thinking-level mutation are not implemented.

## Layout and ownership

`patch-model-panel.mjs` follows the existing guarded compose adapters. It adds the reference search/count header, current-first sections, separate display name/key/context badges, mobile row/footer layout and a Models-settings action. Supplied component, pane and UI sources remain byte-identical. `gi-model-panel.css` adapts the pinned `0afe5366ced9` catalogue rules in a separate host stylesheet.

Filtering preserves option identity and the order of non-current models. The existing component owns selection, context-fit blocking, mutation tokens, stale responses and model-state reconciliation. Unknown context/reasoning values produce no metadata badges. The old cycle button is replaced by the reference's Settings action; native model-cycle commands remain available.

The real search field receives initial focus. Arrows change the highlighted enabled entry while focus stays in the search field; Enter accepts that entry. Keyboard navigation from programmatically focused list rows retains native button behaviour. The guarded `patch-model-accessibility.mjs` adapter adds the reference combobox/listbox contract; options are excluded from sequential Tab navigation.

The Models-settings action closes the panel and dispatches `piclaw:open-settings` with a validated section name and the model trigger as the return-focus target. Unknown sections use General. Existing shortcuts still open General; requests cannot replace an already open modal. Closing Settings restores the trigger without changing session, draft, media or references.

## Model accessibility

The search combobox and trigger control one listbox with a per-composer ID. Option IDs encode model labels and remain stable across filtering and reopening. `aria-selected` follows the accepted current model; `aria-activedescendant` follows the enabled keyboard highlight. Empty, loading, all-blocked and switching states have no active descendant. Blocked options remain visible with `aria-disabled` and an accessible context-fit description. The listbox reports loading with `aria-busy`.

Mouse selection retains search focus while the native PATCH is pending. Rejection restores an enabled highlight and leaves the panel usable. Accepted selection returns focus to the connected trigger only when the panel still owned focus at completion, the document has no focused control and Settings does not own the keyboard. A delayed response cannot take focus from the composer or Settings.

Refreshing previously left cached rows available to keyboard Enter even though they were not rendered. `patch-model-picker.mjs` now includes `loadingModels` in its disabled-entry projection. The injected hooks follow the loading-state declaration to avoid a temporal-dead-zone error. Supplied sources, native mutation/context guards and stale-session response handling are unchanged.

Accessibility verification on 2026-09-26:

- Eight model helper tests and 14 pixel helper tests pass; `make test vet bun-checks` passes.
- `make test-ux-model-panel`: 36 passes across Chromium/WebKit and three viewports. Covers unique/stable IDs, current versus highlighted option, Tab order, empty/loading/switching states, pointer focus, rejection recovery, accepted-selection return focus and delayed-response focus ownership.
- `make test-ux-context-fit`: 36 passes, including blocked-option descriptions and no descendant for an all-blocked filter. An older broad `.active` session assertion now targets `[role="menuitem"].active` because status badges also use `.active`.
- `make test-ux-picker-geometry`: 24 passes. Models/Quick Actions: 174 passes. `make test-ux`: 107 passes, with the existing 11 fixture-dependent skips.
- A read-only delegated review timed out; no independent review result is available for this slice.

The initial expanded tests caught missing post-selection return focus. An intermediate loading guard also exposed an injected-hook ordering error; both were repaired before the final passing runs. One concurrent Go/build run failed while the bundler replaced embedded chunks; the sequential core run passed. Evidence retains those failures. No tests were removed or timeouts increased.

This verifies browser DOM and keyboard behaviour. Screen-reader announcements, physical-device interaction and exact full-frame pixels have not been accepted. The existing pixel failures below are historical measurements; this semantic slice has no new pixel capture claim. The deployment attempt and rollback below supersede the original hold on `dcc8c16`. No terminal code or idle chrome changed.

## Overflow Tab-order regression and rollback

`497e264` passed whole CI36210237405 and was deployed from an isolated checkout on 2026-09-26. The read-only live probe failed before completing its first browser/viewport: Chromium made the overflowing listbox an implicit Tab stop. The small journey catalogue did not overflow and missed this case. The first probe also counted the native option in the read-only thinking selector; model option assertions now scope to the model listbox.

The accessibility adapter now sets `tabIndex="-1"` on the listbox itself as well as its options. Six new browser cases supply 43 catalogue entries and read-only thinking. They check actual overflow, search → Settings and reverse Tab order, clear-button order with a query, empty results and no writes. The panel suite passes 42 cases, geometry passes 24, functional tests pass 107 with 11 existing skips, and Go/vet/hooks/model/pixel helpers pass. No CSS, supplied sources, tests or timeout limits were relaxed.

The failed live attempt was rolled back to exact `dcc8c16` on port 8090, PID681390, PGID/SID681291. Normalised before/after SQL excluding runtime leases is identical (SHA256 `300ce6794c6fe9b8fb93c826fcfecf093676a6cf369f4be4e11dd624a829653d`); DB counts stay 62 sessions, 51 turns, 146 messages, integrity/FKs clean, auth file absent. Probes block all non-read requests and request no microphone/notification permission. Deployment acceptance for the accessibility slice is incomplete; the overflow guard needs green CI before another deployment.

## Explicit capability differences

- No model-pin operation exists. The pin column has an inert spacing slot, not a fake star action. There is no persistence or Alt+Enter pin claim.
- Native model responses do not expose pricing/variant metadata; none is invented.
- A reasoning model's current thinking level is a disabled, labelled read-only selector. Gi does not send a thinking mutation or fabricate a list of supported levels.
- Incompatible models stay visible and disabled. The reference's hide/show-incompatible and compact-context footer actions are not ported in this slice.
- Gi retains its 44px mobile Close button. The pinned reference has no equivalent model-close button; Gi's header is taller. Physical touch and native-prompt acceptance are untested.

## Evidence

Final capture run: `run-1790368026776-6747`. All 72 captures completed with no fixture/page errors. All 18 cross-host full frames differ; 9/36 same-host pairs are unstable. The preceding run had 13 unstable pairs. No masks, tolerance or font-raster flag changes were adopted.

| Model panel | Reference and Gi bounds | Exact changed panel pixels, light / dark |
|---|---|---|
| Phone 390×844 | x8 y8, 374×828 | 40,800 / 49,762 |
| Tablet 820×1180 | x52 y798, 680×263 | 4,013 / 6,602 |
| Desktop 1440×900 | x281 y518, 680×263 | 6,818 / 6,627 |

The original desktop/light model-panel difference was 134,939 pixels. Header/row/footer structure now accounts for far less of that difference, but unstable repeats prevent an exact parity claim. Session-panel structure is unchanged.

Verification:

- `make test-ux-model-panel`: 12 Chromium/WebKit × viewport cases for search/count/clear, initial focus, native metadata, Settings handoff, fallback section, return focus, modal ownership, draft/session retention and zero writes during browsing.
- `make test-ux-picker-geometry`: 24 passes, including touch dismissal and the 44px close target.
- Model/Quick Actions suite: 174 passes. Settings/shell suite: 216 passes. Context-fit suite: 36 passes, including blocked selections, pending-response truthfulness, exactly one mutation and draft/media/reference retention.
- Go, vet, hook checks, six model helpers and 14 pixel helpers pass. Stock functional suite: 107 passes, 11 fixture-dependent skips.
- A bounded read-only delegate review found no concrete bug in ordering, mutation ownership or Settings section/opener handling.

Sparse-row and active-row tests now inspect the reference name/key structure and catalogue class. Shared31/shared32 search tests exercise the real filtering field, clear it before selection, and retain the pending-write current-row check. Frozen feature files and mapping counts are unchanged. Early failed assertions about first-row DOM focus or hidden filtered rows were harness assumptions, not relaxed product criteria.

The compose-surface commit `75da410` passed CI36180760610; deployment was held while panel changes were unverified. The stack through `8f97f08` is now [deployed and read-only verified](session-panel.md#verification-and-release-state) after CI36193525146. Exact pixel and capability gaps above remain open. No terminal chrome was added.
