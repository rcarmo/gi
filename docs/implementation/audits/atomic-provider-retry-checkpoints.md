# Atomic provider retry checkpoints

Status: verified checkpoint increment; #90 and #92 remain open.

A failed provider receipt can carry a `ProviderRetryWait` successor. One SQLite
transaction records its non-adopted history, known/unknown accounting, measured
context, settled/retry events, deadline and `retry_wait` phase. Failure at any
write rolls everything back. A duplicate returns the original committed wait,
not a caller's replacement deadline.

The checkpoint requires the current provider admission, an active claim and no
later provider/tool admission. Cancellation retains observed accounting without
scheduling a retry. Terminal staging cannot be reopened. Wake-up checks the
stored deadline and claim incarnation, then commits its phase and event together.
The host waits until that absolute deadline rather than adding another delay
after settlement and hooks.

Classification uses `goai.IsRetryableAssistantError` on provider-boundary error
diagnostics, never assistant content. Local categories affect presentation only.
Cancellation and hook aborts still suppress retries. Streamed progress remains a
guard; returned partial text/tool prefixes, pending/deferred messages and
incomplete results also suppress retry when an adapter emitted no deltas.
Explicit zero delays and up to 100 retries are honoured; existing delay caps
remain. This does not yet provide fresh committed preparation or restart-safe
sleep/generation/tool continuation.

## Regressions

- Faults at receipt, settled event, retry event and phase writes roll back together.
- Reopen and duplicate settlement retain the deadline and known accounting.
- Future deadlines, cancelled/revoked/released owners and wake faults cannot advance.
- Zero-delay exhaustion performs 101 distinct requests/receipts and 100 waits/wakes.
- Returned partials without deltas never replay and retain known accounting.
- Cancellation publishes its committed partial before stopping the retry controller.
- A side-prompt test now waits for the held main stream's partial row before its
  before/after persistence comparison. Admission alone was not a stable barrier:
  legitimate main progress could insert that row during the side request. The
  original no-side-persistence and single-turn assertions remain.

## Profiling, 31 October 2026

`make test-provider-retry PROFILING=1 PROFILE_KEEP=1 RACE=` captured cumulative
CPU, `alloc_space` and `alloc_objects` for three repetitions, including real
zero-delay 100-retry hosts. Go 1.27.3, two execution CPUs. Before/after comparisons
below use the same `internal/turn` cases and dispatch counts; build totals and
other packages are excluded.

The largest avoidable allocation source was `sanitizedProviderIntent`: JSON
round-tripping every context/tool schema accounted for 45.23 MB and 506,572
objects. Already-owned JSON now takes a fast path when it has no credential
keys. A key-only scan handles escapes with the standard decoder and preserves
Unicode lowercase semantics. Secret-bearing envelopes still use recursive
scrubbing with exact JSON numbers. No intent is cached across attempts.
Security/ownership/large-number regressions pass; differential fuzzing against
the decoded reference passed 33,294 executions in five seconds.

| Runtime profile | Initial | JSON/key-scan tuning |
| --- | ---: | ---: |
| Sampled allocated space | 111.13 MB | 84.00 MB |
| Sampled allocated objects | 1,157,712 | 518,742 |
| Sampled CPU | 1.67 s | 1.59 s |
| Test process duration | 1.94 s | 1.88 s |
| Intent snapshot allocation | 45.23 MB | 23.73 MB |

Allocation space fell 24.4% and objects 55.2%; intent allocation fell 47.5%.
Single CPU/elapsed samples are too short to claim a throughput improvement.
Remaining intent work is the required fresh JSON snapshot and persistence.
Ledger rebuilding accounted for 0.52 s cumulative CPU across this stress run.
Reusing its decode scratch was investigated and measured but showed no reliable
overall improvement, so was reverted. The authoritative, ordered receipt sum and
numerical checks remain unchanged. The 100-retry host takes about 0.5 s without
network delay; no major latency blocker was found. Recheck the same workload
when fresh preparation/reopen support lands rather than extending optimisation
indefinitely. Raw profiles, matching binaries and disposable profiling logs are
deleted after comparison.

## Verification

Post-change race checks passed three repetitions across config/store/turn/TUI,
including intent security and differential-fuzz seeds. The full Go suite, vet,
web build, Bun checks and a fresh isolated browser suite passed (147 passed,
13 skipped). The side-prompt synchronisation regression also passed ten
repetitions with race detection. The 101-attempt stress fixture uses a bounded
20-second guard under race instrumentation; ordinary runs take about 0.5 s.
No issue is closed by this increment.
