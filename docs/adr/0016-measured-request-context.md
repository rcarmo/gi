# ADR-0016: Keep provider-request context separate from cumulative usage

## Status

Accepted — 2026-09-22

## Context

Gi persisted `inference.finished` billing totals accumulated across a turn's provider iterations. The TUI used those totals as context occupancy. The web returned no context usage. Summing successive requests or including their outputs can greatly inflate a context meter.

## Decision

After a successful provider result with reported usage, append `context.measured` to the turn event log. Record request input, cache-read/cache-write input, model and iteration. Tokens equal `Input + CacheRead + CacheWrite`: go-ai reports uncached input and cache components separately. Output is excluded. Empty usage objects and negative counters do not create measurements; the deterministic shell responder has no provider measurement.

Keep cumulative billing/cost events unchanged. `LatestContextMeasurement` reads only explicit measurement events for the addressed session, in append order, and adds turn ID and observation timestamp. Historical billing totals are never backfilled into occupancy. The value is the latest observed provider request, not a live tokenizer count or a prediction of the next request after tools, edits or compaction.

The session model API exposes `context_usage` with nullable `tokens`, `contextWindow` and `percent`, a source marker and measurement provenance. Capacity comes from the configured provider catalogue. Missing capacity yields no percentage; missing measurements remain null even when capacity is known. Errors remain errors instead of becoming zero usage.

The web renders compact K/M counts with `?` markers for unknown fields, an accessible label and an explicit latest-request tooltip. Unknown usage uses a neutral indicator. Known thresholds are green through 75%, amber above 75% through 90%, red above 90%. Only the arc clamps at 100%; the label retains the supplied percentage. Compaction is enabled only when a real host callback is supplied; Gi does not supply one in this slice.

The model picker blocks a known smaller context window before mutation. Shared native model validation also checks the latest measurement, covering direct web/TUI commands and changes between catalogue load and selection. Unknown usage/capacity does not imply either overflow or compatibility. Measured fit does not guarantee the next request will fit; tokenisers, reserved output, tool/schema additions and post-measurement context changes are not modelled.

The TUI keeps cumulative input/output/cost statistics but takes its context segment from the separate measurement. It adds no row, panel or persistent notice. Unknown measurements omit the context segment; the existing footer layout stays intact.

## Browser context-meter follow-up (2026-09-22)

`make test-ux-context-meter` now covers frozen `@ux-context-001/005` in all six projects. The local provider supplies explicit request usage through the production parser/persistence/API/SSE path. Browser assertions use fixed expected strings and arc lengths, independently of the production formatter: 85K and 1.5M formatting, fractional percentage rounding, a 125% label with a full arc, explicit zero input with nonzero output, 75/75.01/90/90.01 colour boundaries, reload and session-local unknown state.

The app copies the supplied meter's native title into `data-tooltip`, the missing contract attribute. It uses the existing container ref and a composer-scoped observer for child-owned title updates; observer cleanup follows renders/unmount. A missing title removes the attribute. No wrapper, CSS, component, UI utility or pane change is needed. The context-fit suite additionally checks tooltip refresh after model mutation. A focused delegate review identified the discarded wrapper's ancestry risk and empty-attribute handling before the final matrix.

Validation: 228/228 browser executions (192 base + 12 Steer + 12 context-fit + 12 meter), 70/70 functional, Go tests/vet, hook checks and 23 helpers. Classic 19/236 passing (217 unmapped); shared 2/42 passing (40 unmapped). Compaction callback/active-label cases `context-003/004` remain unmapped because their full behaviour is unavailable.

Terminal adaptation retains the existing inline `ctx used/window percent` footer segment and on-demand `/context` detail. Browser hover text and pie colours do not require a permanent terminal widget. Unit tests now cover measured/zero/overflow usage at widths 60, 100 and 140, preserving row count and cell width; percentages above 100 stay visible. The live session/model harness also passes at 60×18, 100×22 and 140×36. Unknown and explicit zero still share the terminal's omitted inline segment, so this is no claim of full web/terminal meter equivalence. No terminal UI changed.

## Browser context-fit follow-up (2026-09-22)

`make test-ux-context-fit` now covers frozen `@ux-compaction-006` and `007` across all six browser/viewport projects. The isolated provider reports 100 input tokens through its real streaming response; production inference persists the measurement. Registry entries have capacities 80, 100 and 200. The picker blocks the 80-token entry without a request, and a direct native PATCH rejects it with 400 while retaining the model. Equal/larger entries follow normal selection, update context capacity/percentage and survive reload. Text and attachment drafts stay intact. A different session with unknown usage can select the smaller entry.

The test verifies the measurement's original turn/model/iteration, unchanged request token count after model selection and no extra submitted turn. No SQL usage seed, synthetic SSE, paid-provider call or component change is used. The local fixture is deterministic; this is not a live-tokenizer or paid-provider accuracy test.

Validation: 12/12 context-fit executions, 12/12 Steer, 192/192 existing matrix, 70/70 functional, Go tests/vet, hook checks and 23 helper tests. Combined coverage: Classic 17/236 passing (219 unmapped), shared 2/42 passing (40 unmapped). Full compaction, suppression/cancellation and non-fit context feature IDs still need their own evidence.

Terminal model selection already calls the same native fit check. Keep rejections in the existing transient notice and leave editor state/model unchanged; no idle rows or wider footer. This browser evidence adds no new live terminal acceptance credit.

## Original measurement evidence

- Provider-loop test uses a deterministic inference result stub to verify the real loop records input/cache usage. A separate multiple-iteration test proves the latest request wins over cumulative totals and excludes output. No live paid provider was called.
- Store/inference tests cover append/reopen persistence, session isolation, missing measurement/capacity and shared model-fit rejection.
- Web API test uses explicitly supplied stored measurements and a real catalogue window. Terminal tests verify measured footer usage and unchanged idle row count.
- Bun helper tests exercise supplied-value formatting, unknown markers, threshold boundaries, fill clamping and fit predicates. These are helper evidence, not provider-backed browser measurements.
- `@ux-context-002` runs all six browser/viewport projects against native shell sessions, before/after a real turn, reload and model selection. Unknown values remain unknown and unsupported compaction stays disabled.

Validation: 174/174 browser matrix executions, 70/70 existing functional tests, Go tests/vet, three-run focused race tests and 18 Bun source/helper tests. Live three-size TUI session/model tests, smoke and seven-file Gherkin harnesses pass.

## Remaining scope

Browser coverage is 15/236 Classic IDs (221 unmapped); all 42 shared-contract cases remain unmapped. Context formatting/threshold/fit helpers do not map `ux-context-001/003/005`, compaction `006/007` or original `020` without their full browser acceptance fixtures. Active compaction, current tokenisation, reserved output and post-compaction measurement invalidation need dedicated implementation and evidence.
