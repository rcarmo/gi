# Native writer fencing

Native main/manual hosts bind their context to one session, turn and claim
incarnation. Provider/tool intent admission checks the current token in the
same write as its checkpoint. Provider partials/receipts, tool receipts,
compaction completion and owned progress/retry events reject superseded
writers, including context copies which suppress cancellation. Recovery fences
retain revocation after the replacement releases its claim.

Owned cancellation settlement deliberately suppresses cancellation while
retaining that binding. This preserves known committed accounting and permits
an unrecovered lost-claim host to retire its abandoned turn. Unbound Store
calls retain their privileged Go API behaviour; this binding does not replace
script caller permissions or callback-lifetime checks.

A staged terminal decision cannot reopen inference, tool admission, compaction
or the running phase. Bootstrap shell status/finalisation updates also retain
the native writer binding. Bootstrap's older multi-write terminal/child path
still needs the common settlement protocol under #86.

## Verification

Store regressions exercise active and released replacement claims, detached
context copies, invalid scopes, cancelled receipt writes and known accounting.
Two-store engine tests hold real provider/tool hosts across takeover and check
that late responses, results and progress are rejected without dispatching
another request, finishing the replacement or releasing its claim. A rejected
bootstrap finalisation retains its claim and keeps queued work behind it.

The concurrent-session test used shared-cache memory SQLite and reported
`SQLITE_LOCKED`/deadlock during concurrent read/write transactions. It now uses the
production file-backed WAL path with the same assertions, and joins the engine
before closing storage. Shared-memory SQLite retains its table-locking limits.

## Profiling, 31 October 2026

`make bench-turn-writer PROFILING=1 PROFILE_KEEP=1 RACE=` captured cumulative
CPU, `alloc_space` and `alloc_objects` for repeated bounded progress writes on
one WAL turn, Go 1.27.3 and two execution CPUs. The initial implementation
allocated an explicit transaction and two Go query results for every update.
A conditional insert now checks ownership, revocation and journal sequence in
one atomic SQLite statement.

| Scoped progress write | Initial transaction | Conditional insert |
| --- | ---: | ---: |
| Time | 141,139 ns | 157,003 ns |
| Allocated bytes | 3,223 | 1,168 |
| Allocations | 86 | 31 |

Allocated bytes fell by 63.8% and allocations by 64.0%. Measured latency rose
by 15.9 us (11.2%); at the existing ten-updates/second pacing limit that adds
less than 0.2 ms/second per reporter. The allocation reduction justifies this
small absolute latency cost. These are single profiled samples with calibrated
iteration counts; aggregate profile totals include fixture setup and are not
per-operation comparisons. SQLite execution dominates the remaining CPU.

The unbound control fell from 1,944 B/58 allocations to 1,120 B/28 allocations;
time rose from 129,788 ns to 142,161 ns. Whole-run peak RSS was 291/279 MB,
including builds. Raw captures, disposable logs, reports and fixtures are
removed after comparison. Post-change validation passed the scoped contract
and settlement regressions three times with `-race`, full `make test`, `make
vet`, `make build-web`, `make bun-checks` and isolated `make test-ux` (147 passed,
13 skipped). No recovery acceptance gate or issue is closed by this increment.

This increment does not enable retry-sleep or general generation/tool recovery.
The streamed-progress retry guard and ordinary `Close` abort semantics remain.
