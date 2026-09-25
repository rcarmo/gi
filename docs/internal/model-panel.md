# Native model-panel adaptation

Gi's model panel now matches the pinned Classic outer bounds in phone, tablet and desktop captures. Exact pixels still fail. The adapter exposes native model selection and a Models-settings handoff; pinning, pricing and thinking-level mutation are not implemented.

## Layout and ownership

`patch-model-panel.mjs` follows the existing guarded compose adapters. It adds the reference search/count header, current-first sections, separate display name/key/context badges, mobile row/footer layout and a Models-settings action. Supplied component, pane and UI sources remain byte-identical. `gi-model-panel.css` adapts the pinned `0afe5366ced9` catalogue rules in a separate host stylesheet.

Filtering preserves option identity and the order of non-current models. The existing component owns selection, context-fit blocking, mutation tokens, stale responses and model-state reconciliation. Unknown context/reasoning values produce no metadata badges. The old cycle button is replaced by the reference's Settings action; native model-cycle commands remain available.

The real search field receives initial focus. Arrows change the highlighted enabled entry while focus stays in the search field; Enter accepts that entry. Keyboard navigation from list rows retains native button focus. Existing menu/menuitem roles are retained; the reference's listbox/combobox contract needs a separate accessibility acceptance pass.

The Models-settings action closes the panel and dispatches `piclaw:open-settings` with a validated section name and the model trigger as the return-focus target. Unknown sections use General. Existing shortcuts still open General; requests cannot replace an already open modal. Closing Settings restores the trigger without changing session, draft, media or references.

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

CI must pass before deployment. The preceding compose-surface commit `75da410` passed CI36180760610; it was not deployed while these model-panel changes were unverified in the working tree. No terminal chrome was added.
