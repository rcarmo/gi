# Focused durable findings acceptance

The ten findings in [#83][tracker] are covered by the merged repairs and contract
regressions below. The implementation baseline is main `3c1fcc3b`, including the
parent-skip repair in PR #125. This review reuses the completed verification of
that tree; it does not rerun those suites or certify a release.

Gi keeps one pure-Go SQLite store for intent, receipts, continuation and UI
projections. It implements the bounded contracts behind its existing schema;
it does not install a second writable go-ai durable store. [ADR 0061][ledger],
[ADR 0062][partials] and [ADR 0063][children] record the transaction and lifecycle
decisions. The comparison oracle is go-ai v1.4.0's implementation and named tests.

## Finding coverage

| Finding | Implemented behaviour | Contract regressions |
| --- | --- | --- |
| F1 / #84 | Direct provider messages and legacy envelopes survive reopen/fork with signed blocks and metadata intact. | `internal/turn/provider_replay_test.go` |
| F2 / #85 | Committed claimed-phase intent gates effect entry; prepared arguments remain separate from original provenance. | `internal/store/effect_intent_test.go`, `internal/turn/effect_intent_test.go` |
| F3 / #86 | Shared Close completion fences admission and joins owned providers, tools, hooks and writers. Required child receipts and final hooks precede parent cleanup. | `internal/turn/hook_hosts_test.go`, `child_join_test.go`, `interrupted_close_test.go`, `generation_close_test.go`, `checkpoint_skip_test.go` |
| F4 / #87 | Bounded committed prefixes and received terminals/counters remain separate; unadopted raw bodies stay out of model context and conversation views. | `internal/store/provider_partial_test.go`, `internal/web/replay_privacy_test.go` |
| F5 / #88 | Occurrence-keyed response, usage, successor and terminal projections settle atomically before publication. | `internal/store/provider_settlement_test.go`, `turn_settlement_test.go` |
| F6 / #89 | Role, stop reason, calls, framing and finite non-negative usage are checked before adoption or tool dispatch. | `internal/inference/terminal_test.go`, `internal/turn/terminal_result_test.go` |
| F7 / #90 | Durable retry waits/wakes preserve lifecycle, attempts and deadlines, and rebuild current committed settings/context before a new request. | `internal/turn/provider_retry_process_test.go`, `provider_preparation_test.go` |
| F8 / #91 | One reducer excludes aborted/error/deferred contributions and preserves batch ordering and compaction boundaries. | `internal/turn/context_contributions_test.go`, `context_checkpoint_test.go` |
| F9 / #92 | Original generation/provider/tool checkpoints resume without whole-prompt retry. Saved and current replay safety gate tool re-entry; progress-based crash limits park exhausted work. Explicit retry and failed-terminal skip have atomic admission. | `internal/turn/tool_batch_process_test.go`, `generation_ready_process_test.go`, `internal/store/checkpoint_retry_test.go`, `checkpoint_skip_test.go` |
| F10 / #93 | General and hook tool results share UTF-8-safe output limits, committed progress and separate truncation diagnostics. | `internal/tools/tool_output_test.go`, `internal/turn/tool_output_test.go` |

The process regressions cover killed owners around intent, physical effects,
receipts and retry sleep, including changed registry/safety and parked retries.
The shutdown regressions include cancellation-ignoring callbacks, shared Close
callers, commit/recorder failures, interrupted setup/compaction, and explicit
cancellation. Skip regressions cover partial cleanup, reopen, stale writers,
live-session hosts, original wire IDs and a child's blocked final hook.

## Verification retained

Sequential Makefile verification used Go 1.27.3 with inherited agent-directory
overrides cleared. The final implementation passed full Go, vet, build-web and
Bun checks. Fresh isolated browser acceptance passed 155 tests with 13 skips.

| Final verification | Result |
| --- | --- |
| Full Go | store 11.208s; turn 36.698s; web 7.562s; TUI 21.090s; remaining packages passed |
| Checkpoint skip/retry/cancellation, race x3 | store 20.107s; turn 6.228s |
| Owned recovery/Close, race x3 | store 20.646s; turn 81.430s |
| Held retry/TUI, race x3 | store 16.033s; turn 13.504s; TUI 2.270s |

These durations identify the completed runs, not a performance comparison.
Disposable logs and generated browser/build output were removed or restored
after use. Earlier measured optimisations and profiling disposal are recorded in
the linked retry/partial audits. Rui paused profiling for the remaining repairs;
no new CPU or heap captures were taken. No release or deployment was performed.

## Conservative limits

Unknown, legacy or malformed execution frontiers remain held. Parent-linked
checkpoint retry stays blocked because reversing parent consumption is not
proven safe. A proven parent-linked checkpoint can be explicitly skipped through
failed-terminal cleanup; committed child terminals are acknowledged without
revival or duplicate lifecycle attempts. Unsafe tools receive interrupted
outcomes. Missing remote counters remain uncertain; no exactly-once external
billing guarantee is made.

The tracker explicitly excluded full upstream adherence. Generic dependency
graphs, exhaustive migration/history semantics, no-model input settlement,
failed-run inbox policy, reset/head concurrency, all hook selection phases,
telemetry/observer overflow budgets, nested/codemode recovery and every deferred
host interface remain outside this focused verification. Fresh retry settings
and bounded callback ownership are covered here; exhaustive prompt environment
and context-overflow equivalence need their own comparison before a full
adherence claim. Deferred provider terminals are rejected without tool dispatch.

[tracker]: https://github.com/rcarmo/gi/issues/83
[ledger]: ../../adr/0061-receipt-accounting-and-terminal-settlement.md
[partials]: ../../adr/0062-provider-partial-receipts.md
[children]: ../../adr/0063-child-before-parent-settlement.md
