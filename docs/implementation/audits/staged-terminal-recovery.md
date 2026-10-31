# Staged terminal recovery

A recovered `turn.terminal_staged` decision retains its message identity and
committed usage without entering prompt setup, inference or tool dispatch.
[ADR 0063](../../adr/0063-child-before-parent-settlement.md) defines claim and
child-settlement checks. Issues #86/#87/#90/#92 require further work.

## Profiling, 31 October 2026

`make bench-durable-recovery PROFILING=1 PROFILE_KEEP=1 RACE= RECOVERY_BENCH_TIME=2s`
compared the initial classifier with a SQLite existence query. Both used one
file-backed WAL turn with 1,000 tool-progress events containing 1 KiB each,
Go 1.27.3 and two execution CPUs. Setup was outside the benchmark timer.

| Per classification | Initial journal decode | Existence query |
| --- | ---: | ---: |
| Time | 3,508,547 ns | 23,093 ns |
| Allocated bytes | 4,149,107 | 2,360 |
| Allocations | 21,838 | 71 |

Time fell by 99.34%, allocated bytes by 99.94% and allocations by 99.67%.
Cumulative CPU, `alloc_space` and `alloc_objects` profiles identified
`ListTurnEvents` and JSON decoding as the avoidable work. The existence query
removes journal materialisation; `GetTurn` is the largest remaining application
allocation site at the much lower per-classification cost. This lookup only
classifies recovery; takeover rechecks the checkpoint and claim transactionally.

These are single profiled benchmark samples with different calibrated iteration
counts (688 and 101,128). Aggregate profile totals are not equivalent-workload
comparisons; the table uses measured per-operation values. Whole-run peak RSS
was 679/681 MB including separate builds and does not measure per-request memory.
Function-level profile analysis used embedded Go symbols; the profiling helper
had already removed matching test binaries. No source/disassembly analysis was
performed. Raw profiles, disposable reports, logs and fixtures were removed
after comparison.

## Contract checks

Regressions cover transaction rollback, competing stores, renewed heartbeats,
replaced and released claims, token reuse, malformed decisions and live local
hosts. Engine tests preserve history identities and accounting, honour an
explicit stale-turn abort, retain a parent behind an unjoined foreign child,
and reopen a WAL after killing the checkpoint owner. The recovered paths make
zero inference calls.

`make test-durable-recovery RACE=-race` passed three repetitions of the focused
store/turn contracts, including existing Escape-restore and lost-heartbeat
cancellation checks. Full `make test RACE=`, `make vet RACE=`, `make build-web`,
`make bun-checks` and `git diff --check` passed. A fresh isolated `make test-ux`
instance passed 147 functional scenarios and skipped 13; this was not the
six-project shared fixtures matrix. No live instance was deployed or restarted.

Generation/tool recovery, durable retry sleep, restart parking and
checkpoint-preserving `Close` are outside this increment. Ordinary `Close`
retains abort semantics.
