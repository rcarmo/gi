# Restart-safe provider retry waits

Committed provider retry waits can continue after a stale owner is fenced. This
increment covers retry sleep only. #92 also needs generation/tool continuation,
interruption counts and parking; #86 still needs checkpoint-preserving `Close`.

## Recovery boundary

The failed-attempt receipt stores version 1 of the accepted once-only agent-start
policy: prompt/message replacements, raw messages/tool definitions and registered
tool origin flags. It also stores the iteration, retry attempt, absolute deadline
and repeated-tool-failure budget. Recovery restores values, rebuilds current
instructions/history/settings and rebinds registered tools to current definitions.
Setup, original prompt admission, agent/turn start hooks and settled tools are
not replayed. Current retry limits are checked again before provider admission.

`AcquireProviderRetryRecovery` verifies the session, current retry frontier,
receipt, phase, claim token and observed heartbeat in one transaction. Claim
rotation, the permanent old-writer fence and recovery admission event commit
together. Rollback tests cover each write. Legacy, incomplete and unsupported
policies enter the existing conservative hold path. Changed ownership or later
effects reject takeover. Unknown provider/tool effects still require review.

Constructors do not dispatch retry continuations. Hosts call
`Engine.ResumeInterruptedWork` after registering tools/hooks and enabling MCP;
HTTP and TUI startup call it at that boundary. Normal prompt admission also opens
the barrier after its MCP wait. This prevents an expired checkpoint dispatching
against a partly initialised registry. A recovered wait retains its original
deadline; explicit user cancellation still settles it as cancelled.

## Checks completed

- Store reopen, lifecycle validation/copy isolation, deadline retention,
  session/claim/heartbeat conflicts, later effects and write-fault rollback.
- Engine continuation with fresh instructions, committed history and replacement
  tool definitions; no setup, prompt, lifecycle or settled-tool replay.
- Same iteration, retry-attempt and repeated-tool-failure budgets; disabled
  current policy prevents a request. Legacy/invalid policy holds and cancellation
  while waiting are covered.
- A live engine is killed with SIGKILL after committing its failed request and
  retry successor. The reopened engine resumes once after runtime initialisation,
  retains the captured prefix and original prompt, and preserves accounting.
- Cancellation during a recovered wait originally skipped `agent_end`. A failing
  regression reproduced it; the loop now owns that final callback before waiting,
  while failed fresh setup still does not acquire an agent lifecycle.
- Final full Go suite and vet pass. Config/store/turn/TUI retry checks pass race
  detection three times after the status-read and agent-end fixes; durable
  ownership/close/parent/staged-recovery checks also pass race×3. Web build and
  Bun checks pass; the final isolated browser suite passes 147 with 13 skips.
  The profiled non-race retry suite passed three times (168 tests). Publication
  is separate from these local checks.

## Profiling and status-read optimisation

Go 1.27.3; base `5b4a6b0a` plus this working increment. The profiled retry target
runs config/store/turn/TUI tests three times. CPU, `alloc_space` and `alloc_objects`
were inspected. Turn tests sampled 3.10 CPU seconds, 263.86 MB and 2,129,962
allocated objects. SQLite execution dominates CPU; its transactional ownership
and receipt writes are retained. Testing's regular-expression matcher accounts
for 46.53 MB of the 47.53 MB regexp allocation cost. TUI catalogue initialisation
is a one-time test-process cost.

Repeated turn metadata decoding was avoidable: refresh loaded a whole turn to
check its session/status, then the context builder read fresh metadata again.
`CheckProviderRetryRunning` reads only session/status for the first check. The
builder's metadata read and all transactional admission/settlement fences remain.

Equivalent guard benchmarks use the same database and turn metadata, three
one-second runs each:

| Guard | Time | Bytes/op | Allocations/op |
| --- | ---: | ---: | ---: |
| Full turn record | 17.251–18.240 µs | 2,721 | 77 |
| Session/status only | 6.230–6.632 µs | 663–664 | 24 |

The guard saves 2,057–2,058 bytes (75.6%) and 53 allocations (68.8%); elapsed time
falls about 64%. These are guard measurements, not end-to-end throughput gains.
The turn-test workload is unchanged in the follow-up profile: sampled CPU
3.10→3.20 s (+3.2%), allocated space 263.86→266.85 MB (+1.1%) and objects
2,129,962→1,886,762 (-11.4%). Sampled aggregate allocations vary; reported test
elapsed stays about 11.8 s. The overall command also rebuilt changed packages,
so its increased build-inclusive CPU/wall time is not application latency.

The narrow status read has a measured per-operation benefit, keeps current
metadata and security checks, and introduces no observed end-to-end latency
increase. The final broad and repeated race suites pass on this change.
Raw captures, matching binaries and disposable profiling logs are deleted after
this analysis.

## Remaining recovery work

A crash after retry wake but before the next provider admission still enters a
conservative hold. General generation/tool checkpoints, safe unknown-tool replay,
interruption/progress accounting and graceful close requeue are not implemented
by this increment. These keep #83/#86/#87/#92 open.
