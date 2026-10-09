# Classic thought/draft disclosure 001–005: bounded clause review

The installed Piclaw 3.2.4 `components/status.ts` is byte-identical to the
frozen 70d33bc source (`frozen-to-oracle-sources.json`). A mounted installed
Classic UI and event-translator probe now checks long synthetic Thoughts and
Draft disclosures against a disposable timeline. The native Gi test uses a
disposable streaming fixture; its tags are separate browser evidence.

| Frozen ID | Native assertion and Gi handler | Limit |
|---|---|---|
| `@ux-thoughts-001` | `tests/ux/thoughts.spec.mjs:91+` checks `data-expanded=false`, rendered overflow for a wrapped paragraph, and disclosure remeasurement on viewport/container resize and later streamed text. Gi `web/src/components/status.ts` renders disclosure; `web/src/gi-preview-overflow.ts` measures clipping. The installed probe independently shows the newest nine of 14 distinct source lines in each collapsed panel. Gi's native streamed fixture now asserts exactly lines 04–12 of its 12 initial source lines in both collapsed panels. | Different fixture text and transport; installed fixture does not test wrapped-line measurement, viewport remeasurement or physical pixel identity. |
| `@ux-thoughts-002` | `thoughts.spec.mjs:20–61` keeps a collapsed Thoughts/Draft panel while accepted streamed text grows and checks the new text is present without opening it. | Fixture-supplied text is not provider thought semantics. |
| `@ux-thoughts-003` | The native test activates separate Thoughts and Draft disclosure controls and checks each panel's expansion state without discarding text. Installed UI independently keeps the other panel collapsed while expanding either one, and preserves earlier lines on expansion. | The installed fixture supplies bounded thought-fetch and visibility responses; live callback persistence is not tested. |
| `@ux-thoughts-004` | The native test presses unmodified Escape to collapse an expanded status panel; editable focus and modified Escape stay with their controls. The installed probe checks the same three keyboard conditions on mounted panels. | No hardware keyboard or assistive-technology pass. |
| `@ux-thoughts-005` | Expanding/collapsing after streaming keeps full thought/draft text and scroll behaviour, with the composer draft unchanged. Gi preview state is in `web/src/ui/app-agent-previews.ts`. | No full Piclaw current-UI comparison. |

An earlier broad `make test-ux-thoughts` run passed 47/48 but Chromium tablet
`@gi-preview-001` stayed on `Loading Gi…` during startup, before thought
assertions. The focused `@ux-thoughts-00[1-5]` run was aborted after 28/30,
so it has no result. An unchanged rerun via `make test-ux-steer` with
`GI_UX_THOUGHTS=1` and the five-tag grep passed **30/30** across six
Chromium/WebKit viewport projects. A later focused rerun with the explicit
nine-line source-tail assertion passed **30/30**. The full Thoughts file then
passed **48/48** across the six projects on 29 September 2026; the earlier
startup failure is not credited as a pass. The mounted installed Piclaw 3.2.4
probe, run with `ORACLE_DISCLOSURE_ONLY=1` and explicit `ORACLE_BROWSER` and
`ORACLE_VIEWPORT` values for each Chromium/WebKit phone/tablet/desktop case,
passed **6/6**, with no fixture request failures. Its synthetic thought
fetch and visibility callbacks do not establish live provider thought content,
saved disclosure state, reduced-motion behaviour or physical accessibility.
