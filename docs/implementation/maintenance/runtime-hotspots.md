# Runtime hotspot tuning (Go 1.27.1)

Three repeated-work paths were tuned on 2026-10-06. Provider-backed MCP authentication (#29) is verified separately in [MCP](../../internal/mcp.md#provider-authentication).

## Equivalent workloads

`env -u PI_CODING_AGENT_DIR -u GI_CODING_AGENT_DIR make RACE= bench-hotspot-targets` runs the same three benchmarks before and after the changes, at two Go processors and one-second bench duration. The baseline uses the pre-tuning functions from `f4ff8ad`; the tuned run uses the current implementations. Both use Go 1.27.1, identical fixtures and benchmark code. CPU, cumulative allocation bytes and allocation objects were inspected for both runs. Raw profiles, matching test binaries and disposable logs were removed after analysis.

| Workload | Baseline ns/op | Tuned ns/op | Baseline B/op | Tuned B/op | Allocations/op before → after |
|---|---:|---:|---:|---:|---:|
| Runtime options: five enabled IDs, authenticated OpenAI catalogue | 103,279 | 48,510 | 165,023 | 43,541 | 1,676 → 386 |
| List 100 sessions with stored identities/dimensions | 2,622,673 | 487,330 | 388,810 | 233,296 | 8,533 → 5,379 |
| Repeat unchanged transcript projection at width 100 | 2,930,067 | 200 | 6,652,494 | 64 | 44,347 → 4 |

Adaptive benchmarks execute different iteration counts; raw aggregate CPU/heap totals are therefore not normalised improvement estimates. The per-operation table describes these workloads. Cold startup, changing/oversized transcripts, browser latency and general inference latency are not measured by these benchmarks. The full suite includes new invalidation/oversize tests and is not an equivalent performance baseline.

## Catalogue options

`ListRuntimeOptions` snapshots each requested provider's catalogue once per call and looks up enabled/default IDs in that snapshot. Previously each `GetModel` cloned the provider's entire catalogue, followed by another list for authenticated providers. The benchmark cuts allocation bytes by 74% and objects by 77%.

There is no cross-request catalogue or credential cache. New model registrations, catalogue refreshes and external credential rotation/logout are visible on the next request. Provider `ModifyModels` still runs after enabled/default model resolution, preserving Copilot filtering and the existing explicit-model behaviour. Registry/credential-change tests cover invalidation without a new upstream registry-generation API. Catalogue cloning remains a cost for large lists; changes to the upstream model registry are outside this patch.

## Session listing

`Store.ListSessions` reads sessions and required identity fields in one joined query, then dimensions in one batched query, inside a read transaction. Previous code made two additional queries per session. The transaction provides one consistent snapshot across the two reads; aliases, state, ordering and normalisation are unchanged. Missing/incomplete identities still return `sql.ErrNoRows`; cancellation and malformed data errors propagate. Tests compare batched records with individual `GetSession` results and exercise identity failures/cancellation. Allocation bytes fall 40%; remaining SQLite decoding and state/JSON allocation is uncached.

## Transcript projection

Search/selection projection retains one width/generation snapshot, bounded to 262,144 display cells. It reuses the existing exact transcript-line and UI-state invalidation: changed streaming content, theme generation, output width, assistant identity, tool expansion/render modes and selection rebuild it. Running blocks bypass the memo; oversized projections are rendered without retention. Viewport scrolling does not change the retained projection.

The benchmark measures repeated unchanged content after warming. First render and content/width/theme changes still pay projection cost. Tests compare cached output with fresh renderer output after width, content, expansion, selection, theme and identity changes; existing Unicode/grapheme, soft-wrap, prompt-jump, highlighting and session-isolation tests remain in place. Numerical rendering contracts are unchanged.

## Verification

The full Go suite passed 2,167 tests across 37 packages and vet. Twenty-nine targeted identity/catalogue/transcript tests ran three times under the race detector; provider-auth tests also passed three race-enabled repetitions. The first repeated catalogue fixture reused a registry ID across repetitions; unique IDs fixed that test fixture. An independent delegated read-only review timed out, so acceptance uses the tests, equivalent profile comparison and direct source review. Browser fixtures, frontend pins and live runtime state were not changed.
