# Retry continuation after a committed wake

A retry can recover after its wake commits and before another provider request
is admitted. The original regression returned no recoverable checkpoint in this
window. Recovery now accepts the latest `inference.retry_resumed` boundary with
its receipt, versioned policy and unchanged iteration/attempt/deadline/budgets.
It rotates the stale owner and permanent writer fence in the same transaction
used for sleeping retries. The recovered host skips the already committed wake
and rebuilds current generation inputs.

Any later provider, tool or terminal effect boundary rejects this path. An
admitted request still has unknown remote effects and stays subject to the
conservative hold policy. This adds no provider request replay or exactly-once
billing guarantee. Runtime initialisation also gates the `running` recovery
phase, so an expired/resumed checkpoint cannot dispatch from the constructor.

Restoration previously loaded current settings before a future deadline and
failed the turn immediately if those files were malformed. A separate failing
regression reproduced that error. Restoration now copies the engine config;
full current settings validation still runs after the wake, before generation.
This preserves the sleep and avoids a redundant settings/instruction reload.

## Checks

- The store regression fails before the change and passes afterwards; the wake
  count stays one and the old owner cannot write after takeover.
- Thirteen later provider/tool/terminal boundaries reject recovery without
  rotating ownership. Existing write-fault and stale/foreign-owner tests use the
  same transaction path and continue to pass.
- Engine tests preserve captured prefix, fresh instructions/history/tool
  definitions, same budgets/accounting, single prompt admission and single
  agent-end callback after an already committed wake.
- Real SIGKILL tests now cover both sleeping and already-woken processes. The
  latter blocks generation preparation before a second provider admission;
  reopening resumes once with two provider receipts, one user prompt and one
  wake event.
- Config/store/turn/TUI retry checks pass race detection three times. The final
  full Go suite and vet pass after the deferred-settings fix. Web build and
  Bun checks pass; the final isolated browser suite passes 147 with 13 skips.
  Publication is separate from these local checks.

## Profiling

Go 1.27.3, base `4132effe` plus this change. The profiled retry target passes
187 tests across four packages, each repeated three times. CPU, `alloc_space`
and `alloc_objects` were inspected. Turn tests take 17.018 s with 3.19 s sampled
CPU, 288.77 MB allocated space and 2,276,012 allocated objects. SQL locking and
SQLite execution account for 2.11 s cumulative CPU (66.1%); receipt/admission
transactions retain their ownership and cancellation gates. Regexp compilation
is mainly Go test-name matching; the prior status-only guard optimisation is
retained. Neither retry takeover nor restoration appears in the displayed
sampled allocation rows; that is a sampling limit, not zero allocation.

The earlier suite took about 11.8 s and covered 168 tests. The new SIGKILL wake
case adds a real 1.5-second deadline per repetition (about 4.5 s); new boundary
fixtures add more work. These workloads differ, so aggregate timing/allocation
changes do not establish a throughput regression or improvement. No new material
application hotspot was identified in this increment. Raw profiles, matching
binaries and disposable logs are removed after this analysis.

Graceful checkpoint-preserving close, general generation/tool continuation,
prepared-argument replay safety and progress-based interruption parking still
keep #83/#86/#87/#92 open.
