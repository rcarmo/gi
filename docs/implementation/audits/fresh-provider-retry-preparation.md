# Fresh provider retry preparation

A provider retry now rebuilds its inputs from committed sources before host
entry. Restart recovery still parks interrupted work. #90 is closed; #92 tracks
remaining checkpoint continuation.

## Generation inputs

`providerPreparation` belongs to one turn owner. Before a retry it reads the
current session model and thinking selection, turn tool policy, context history
and compaction boundary, registered tool definitions, and file-backed runtime
settings. Instruction files and prompt sections are rebuilt. Settings produced
by `config.Load` reload; configurations constructed without `Load` stay static.
Malformed or unreadable settings stop preparation. Global retry settings apply
below project settings, including explicit false and zero values.

The failed receipt uses the choices captured for that actual request. Retry
eligibility uses current policy before committing the receipt/deadline decision;
wake preparation checks policy again before admitting another request. A lower
limit or disabled policy prevents host entry. The committed deadline still
controls the existing sleep.

Context and provider-request hooks run for the new preparation. Agent/turn
lifecycle hooks, queued steering admission and completed tools are not replayed.
Once-only agent-start policy is retained separately from the failed request.
Its message replacement remains a turn policy; registered tool overrides rebind
to current definitions and omit withdrawn registrations. The resolved model,
thinking and transport carry into later iterations and receipt metadata.
Compaction uses that preparation's choices and preserves the explicit runtime
auto-compaction override.

Prompt section plans are owner-local. A session-global plan could be overwritten
by a superseded builder. Session prompt-policy records now check the running turn
and claim incarnation in the same statement as the write. Provider admission and
retry wake retain their existing claim fences.

## Checks

Focused checks pass for current messages/instructions, poisoned old closures,
model/thinking/transport changes, replaced and withdrawn tools, once-only
lifecycle hooks, a subsequent tool phase, disabled/zero/increased retry budgets,
malformed settings before decision or wake, global/project layering and revoked
prompt-policy writers. Final config/store/turn/TUI race checks pass three
repetitions. Full Go tests, vet, web build, Bun checks and the isolated browser
suite pass (147 passed, 13 skipped). Lifecycle snapshot tests also cover exact
large numbers, empty overrides and registrations added after capture.

## Profiling experiment

Base: main `80ed6217` plus this increment. The workload repeats
`TestProviderRetryFreshCommittedPreparationAndSubsequentToolPhase`: one failed
request, one fresh request returning a tool call, one tool execution and a final
request. `make test-provider-retry-preparation PROFILING=1 PROFILE_KEEP=1
RETRY_PREPARATION_COUNT=1000 RACE=` captured CPU and allocation histories.

Cumulative CPU was led by SQLite statement preparation (5.75 s, 32.9%), hook
emission/auditing (1.50 s, 8.6%) and configuration loading (0.96 s, 5.5%). The
full-policy workload took 19.440 s for 1,000 turns, with 17.48 s sampled CPU.

A narrow failure-decision settings reader avoided instruction/skill discovery;
full preparation still reloaded before host entry. Equivalent 1,000-turn results:

| Measurement | Full reload | Narrow reader | Change |
| --- | ---: | ---: | ---: |
| Test elapsed | 19.440 s | 25.037 s | +28.8% |
| Sampled CPU | 17.48 s | 19.91 s | +13.9% |
| Allocated space | 5,468.01 MB | 5,256.77 MB | -3.9% |
| Allocated objects | 17,737,860 | 16,864,045 | -4.9% |
| Config allocated space | 452.71 MB | 293.34 MB | -35.2% |
| Config allocated objects | 2,717,303 | 1,830,789 | -32.6% |

Shorter 200-turn runs varied substantially. Their CPU/elapsed results did not
establish a consistent gain. The longer comparison did not justify retaining
the reader: it was reverted. Full typed settings validation is retained, with
no new throughput or memory improvement attributed to this increment. SQLite
and hook-audit work still dominate this local workload; required ownership and
receipt writes were preserved. A further optimisation needs a separate,
bounded comparison against the full-reload path, owned by this issue-repair task.
The measured full-policy cost was about 19 ms per local fixture turn; this
experiment does not establish live-provider latency.

Raw captures, matching binaries, fixtures and disposable logs are removed after
analysis. Restart-safe retry sleep and general checkpoint continuation still
need durable continuation preparation and interruption/replay-safety gates.
