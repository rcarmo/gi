## Graceful close of retry checkpoints

`Engine.Close` preserves a valid retry checkpoint without recording a user abort.
This applies to a sleeping retry and to its committed wake before another effect
admission. Other generation/tool phases still use conservative abort settlement;
general checkpoint routing keeps #83/#86/#92 open.

A closing engine seals launch and control admission, then reserves each eligible
retry under the session lock. The reservation verifies the receipt, lifecycle,
budgets, frontier and current claim. It rotates the claim and permanent writer
fence before cancelling the old host. A separate closing heartbeat retains the
reservation while cancellation-ignoring work finishes.

Close joins main hosts, hook callbacks, side requests, maintenance writers and
peering.
Only then does the Store commit `queued` with `checkpoint_retry_wait` or
`checkpoint_retry_ready`. The original receipt, absolute deadline, iteration,
attempt and lifecycle policy survive. Shutdown emits no agent-end callback or
terminal failure. Runtime-ready recovery rotates the fence again and continues
without setup, prompt, settled-tool or duplicate-wake replay.

Checkpoint queue selection and ordinary claim acquisition reject these phases,
even after a privileged claim release. The retained closed-checkpoint reservation
makes them eligible for runtime-ready recovery immediately, without waiting for
the crash lease timeout. Unknown, legacy and invalid checkpoints cannot use this
route. An already recorded abort wins. A privileged or other-engine abort after
reservation is preserved and causes the close commit to report a conflict; it
cannot be overwritten with queued state.

Reservation/commit failures roll back atomically and reach every concurrent or
later Close caller. Failed prefixes keep their claim for conservative recovery.
No shutdown step changes interruption counts. General progress-based crash
counting and parking belong to #92.

## Verification

The original engine regression produced a cancelled terminal row. The new tests
cover sleep/wake Close and reopening, one prompt and one wake, deferred agent-end,
a callback ignoring cancellation, concurrent Close, another engine sharing the
Store, explicit user abort, and failed reservation/commit propagation. Store tests
cover legacy/later/foreign/aborting states; claim, fence, turn, session and event
write faults; abort after reservation; and whole-prompt claim rejection after
release. The race checks also wait for the actual host before counting end hooks;
a terminal row alone does not establish that the deferred callback has finished.

The expanded config/store/turn/TUI target passes race detection three times. The
final full Go suite, vet, web build and Bun checks pass. The isolated browser
suite passes 147 with 13 skips. Publication is separate from these checks.

## Profiling

Go 1.27.3, base `fa4c8f38` plus this change. The profiled retry target passes 209
tests in four packages, each repeated three times. CPU, `alloc_space` and
`alloc_objects` were inspected. Turn tests took 18.90 s with 3.28 s sampled CPU,
309.78 MB allocated space and 2,411,852 allocated objects. SQL locking/SQLite
accounted for 2.40 s cumulative CPU (73.2%). These transactions enforce durable
admission and cancellation; the existing status-only guard optimisation remains.
The closing helpers occupied 20 ms of sampled CPU (0.61%). Checkpoint acquisition
accounted for 1 MB and 36,864 sampled objects. Test-name regexp matching accounted
for 65.04 MB (21.0%), outside the application path.

The previous 187-test target sampled 3.19 s CPU and allocated 288.77 MB/2,276,012
objects. This target adds 22 tests, including a 500 ms retry deadline, close/join
fixtures and failure transactions per repetition. The workloads differ; their
aggregate changes do not establish a throughput regression or improvement. No
new material application hotspot was found. Raw profiles, matching binaries,
disposable fixtures and logs are removed after this analysis.
