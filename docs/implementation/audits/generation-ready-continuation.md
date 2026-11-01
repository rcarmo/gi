# Generation-ready continuation (#92)

Base: main `7879c51e`, Go 1.27.3, 2026-11-01. This increment recovers a prepared generation before its provider admission. Mid-request and mid-tool checkpoints remain held. #83/#86/#92 remain open.

## Frontier and ownership

`SaveGenerationReady` atomically writes a version-one lifecycle/iteration checkpoint and the `generation_ready` phase. It runs after iteration preparation and verifies that every provider/tool admission since the previous verified frontier has a receipt. The suffix check avoids rescanning the whole journal on each iteration. The checkpoint contains accepted lifecycle overrides and budgets, with no old provider request or executable callback.

`AcquireGenerationReadyRecovery` compares the stale claim token and timestamp, validates policy and journal suffix, and rotates the claim plus recovery fence in one transaction. A changed/fresh claim, cancellation, held/resolved failure, unsupported policy or any unknown later event rejects recovery. Concurrent Store connections cannot both acquire it. Claim/fence/event write faults roll back the complete acquisition.

Recovery waits for `ResumeInterruptedWork` and bypasses prompt setup, start/approval hooks and completed tools. It rebuilds history, current generation settings and current tool definitions, then reapplies accepted lifecycle policy. Generation/context/request hooks receive fresh inputs. Iteration, maximum-iteration and repeated-tool-failure limits survive; the current maximum cannot enlarge the saved budget. Sleeping/woken provider retries retain their existing dedicated route.

At most five automatic generation-ready acquisitions may follow the same receipt frontier. A rewritten ready checkpoint does not reset this count; an atomic provider/tool receipt does. Exhaustion falls back to the existing conservative hold. Explicit parked-checkpoint retry and broader interruption accounting remain unfinished.

## Regressions

- Store: disk reopen and detached policy; stale-writer fencing; two competing Store connections; unsupported/missing policy; unknown later event/effect; cancellation; claim refresh/replacement; claim/fence/event faults; refusal to hide unsettled or legacy admissions; failed-checkpoint rollback; no-progress limits and receipt-based reset.
- Killed process: complete a tool using hook-prepared arguments, commit its receipt, prepare iteration two, block before provider admission, kill and join the process, then reopen SQLite. Recovery uses fresh instructions/current tool definitions, preserves one original prompt and one tool result, performs one new provider request and no repeated tool/start/approval hooks, and accounts exactly two provider iterations.
- Chromium/WebKit: seed an unavailable generation-ready frontier, trigger the runtime-ready recovery scan through queued admission, surface one provider failure, retain one original prompt across reload, and complete the queued follow-up. The browser harness has no native inference provider; successful continuation is covered by the killed-process Go integration test. The functional fixture checks the surfaced failure independently of earlier scenarios' current harness model selection.

The first broad turn run exposed a pre-existing shared-memory SQLite deadlock in `TestSubTurnSyncVsAsyncResultDelivery`. An unchanged-main run reproduced it. In-memory stores now use one pooled connection per Store, retaining isolation and cancellation; file-backed WAL pooling is unchanged. The regression passes three race-enabled repetitions, and a Store test checks waiting-writer cancellation plus unchanged WAL pooling. This does not serialise independent Store instances sharing a named memory database.

## Verification and profiling

Sequential Makefile verification:

- `make test-generation-ready PROFILING=1 PROFILE_KEEP=1`: 56 top-level tests in store/turn, race detector, three repetitions; passed. The final unprofiled race run also passes with the added provider-dispatch write-fault regression.
- `make test RACE= PROFILING=1 PROFILE_KEEP=1`: 35 packages, 2,756 tests; passed.
- `make vet RACE=`: passed.
- `make test-generation-ready-web RACE=`: six Chromium/WebKit conversation regressions passed.
- `make test-ux RACE= PROFILING=1 GI_TEST_RUN_ID=generation-ready-final-20261101`: fresh isolated web/Go build, 151 browser tests passed, 13 skipped.
- `make bun-checks RACE=`: passed.

CPU, `alloc_space` and `alloc_objects` were analysed for the focused and full runs. Race CPU is dominated by checkptr/TSAN instrumentation; regexp test selection leads its allocated bytes. The non-race full turn profile sampled 5.76 s CPU and 675.18 MiB/2,714,296 allocations. `SaveGenerationReady` accounts cumulatively for 0.05 s (0.87%), about 0.50 MiB and 1,489 sampled objects; the settled-effect suffix check accounts for 0.01 s CPU. Allocation sampling is coarse. Existing image-bound fixtures, WASM growth and SQLite rows dominate broader allocations; no material new frontier outlier was found. Different test counts and older wrapper baselines do not establish a speedup.

The failed early phase guard omitted initial `setup`; it now accepts a prepared initial generation. Browser fixture probes also corrected schema fields, the singular `/prompt` endpoint and harness-model assumptions. One final browser attempt was interrupted at test 130/164; it supplies no final acceptance result. The subsequent fresh run passed all 151 applicable tests. Browser CPU/live-heap profiles contain Node/Playwright harness work (zero repository self frames), rather than Go allocation history; the Go package profiles supply that history. Raw captures, matching binaries, failed fixtures and disposable logs were removed after analysis; only these conclusions remain. No release, deployment or live-provider billing validation was performed.

## Remaining scope

Unknown provider outcomes, partial tool batches, per-call interrupted outcomes, persisted/current ReplaySafe decisions, explicit parked-checkpoint retry and general checkpoint-preserving graceful Close still need continuation routes. Unstarted/ambiguous states must not be promoted into whole-prompt replay. Existing `original_call`, `prepared_call`, schema and `replay_safe` tool intents are already persisted; consume them without replacing original prepared arguments or rerunning settled effects.
