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

## Evidence

- Provider-loop test uses a deterministic inference result stub to verify the real loop records input/cache usage. A separate multiple-iteration test proves the latest request wins over cumulative totals and excludes output. No live paid provider was called.
- Store/inference tests cover append/reopen persistence, session isolation, missing measurement/capacity and shared model-fit rejection.
- Web API test uses explicitly supplied stored measurements and a real catalogue window. Terminal tests verify measured footer usage and unchanged idle row count.
- Bun helper tests exercise supplied-value formatting, unknown markers, threshold boundaries, fill clamping and fit predicates. These are helper evidence, not provider-backed browser measurements.
- `@ux-context-002` runs all six browser/viewport projects against native shell sessions, before/after a real turn, reload and model selection. Unknown values remain unknown and unsupported compaction stays disabled.

Validation: 174/174 browser matrix executions, 70/70 existing functional tests, Go tests/vet, three-run focused race tests and 18 Bun source/helper tests. Live three-size TUI session/model tests, smoke and seven-file Gherkin harnesses pass.

## Remaining scope

Browser coverage is 15/236 Classic IDs (221 unmapped); all 42 shared-contract cases remain unmapped. Context formatting/threshold/fit helpers do not map `ux-context-001/003/005`, compaction `006/007` or original `020` without their full browser acceptance fixtures. Active compaction, current tokenisation, reserved output and post-compaction measurement invalidation need dedicated implementation and evidence.
