## Graceful generation Close (#86)

Close preserves a verified generation-ready frontier or its single unsettled provider admission. The turn can continue after runtime setup without resending its original prompt, replaying start hooks or adopting interrupted output. Tool/legacy phases without a verified continuation still use conservative abort settlement; #86 and #92 need those remaining routes.

Base: `e7296d5b`, Go 1.27.3, 2026-11-01.

## Reservation, progress and publication

`ReserveGenerationClose` verifies the journal and lifecycle policy, rotates the active claim to `generation-closing`, and fences the old writer in one transaction. The turn stays running throughout the join. The Close controller keeps the replacement lease alive while providers, hooks, side requests, maintenance and peering hosts finish. Concurrent callers share the existing completion/error barrier; its grace interval only controls diagnostics.

The owned provider recorder can flush bounded snapshots after cancellation. Its private persistence callback selects the reserved writer under session coordination, with one retry if reservation wins between reading the old token and saving. The callback stays bound to the original host and occurrence. Provider/hook contexts keep their revoked writer, and the callback exposes only partial persistence. Late snapshots cannot authorise requests, tool execution or response adoption. A newer host never supplies this capability to an older recorder.

After all hosts and the closing lease join, `CommitGenerationClose` revalidates ownership and the frontier. It seals an unsettled request with uncertain billing, retains the last structured unadopted prefix, and atomically writes a new ready frontier, queued phase and session state. Cancellation or a changed frontier wins over publication. A reservation/commit fault returns through Close; it does not fabricate an abort or grant whole-prompt retry authority.

`checkpoint_generation_ready` is excluded from ordinary queue selection and direct claim acquisition, including after claim release. The completed reservation can be discovered immediately despite its fresh timestamp. Constructors wait for runtime readiness before adopting it. Recovery restores accepted lifecycle values, current context/model/tool definitions and saved iteration/failure budgets. `generation.close_resumed` neither increments nor resets the crash counter. A crash during the reservation or before a resumed host prepares work uses normal fenced recovery; invalid checkpoints become held failures.

## Regression checks

Store regressions cover ready/silent/partial close, reopen, old-writer fencing, ordinary queue/direct-claim rejection, immediate discovery, crash-budget preservation, cancellation, malformed policies, unknown/tool boundaries, reservation faults and six commit-fault classes. Crashes during an unfinished reservation retain a recoverable frontier without fabricating completed Close.

Engine regressions cover cancellation-ignoring providers/context hooks, concurrent Close callers, shared-Store observers, selected-model refresh, received usage and unadopted output, one original prompt and once-only start/end lifecycle. Commit failure retains its reservation and error for later conservative recovery. Invalid queued policy cannot turn into prompt replay. Existing past-grace provider tests continue to require the final owned snapshot after cancellation; only their final state changes to the dedicated queued checkpoint.

Chromium/WebKit fixtures recover a fresh closed-generation claim, surface one unavailable-provider error and drain queued follow-up work without duplicating the original prompt across reload. Native successful continuation runs in engine tests; the browser fixture has no native inference provider.

Completed sequential Makefile checks:

* `make test-generation-close PROFILING=1 PROFILE_KEEP=1`: 112 tests in two packages, race detector, three repetitions.
* `make test-durable-recovery`: ownership/Close/parent/cancellation race matrix, three repetitions.
* `make test RACE= PROFILING=1`: 35 packages, 2,819 tests.
* `make vet RACE=` and `make bun-checks RACE=`.
* `make test-generation-ready-web RACE=`: ten Chromium/WebKit regressions.
* Fresh `make test-ux RACE= PROFILING=1`: 155 passed, 13 skipped, including native/web builds.

## Profiling and cleanup

Inspected CPU, `alloc_space` and `alloc_objects`. The initial non-race full-suite profile (2,815 tests, before the final crash/invalid-policy regressions) sampled 2.77 s CPU/113.47 MB allocated in store and 6.05 s/686.56 MB in turn. `CommitGenerationClose` used 20 ms cumulative CPU in each package. In turn it accounted for 2.01 MB cumulative allocations; the nested checkpoint reader accounted for 1.52 MB, and reservation for 0.51 MB. Inclusive totals overlap. Recorder persistence symbols fell below the displayed threshold; sampling does not establish zero cost.

The focused race workload sampled 17.49 s CPU/98.44 MB in store and 9.07 s/209.25 MB in turn. Race/checkptr instrumentation led CPU, and test-selection regexp state allocated 51.44 MB and 100.19 MB respectively. Filtered close/checkpoint/persistence functions accounted for 4.40% and 2.98% flat CPU. Their SQL ownership/accounting checks and policy decoding have bounded work per active host. The non-race run's larger allocation sites remain SQLite rows, WASM memory, reflection and existing image/JSON fixtures. No additional material application hotspot required tuning.

The final full-suite wrapper reported 1m55s wall and 779 MB peak RSS including build. Browser acceptance reported 3m26s wall and 443 MB child peak RSS. Node exit heap samples held 55,686 to 76,093 KiB; they measure live browser-runner allocations, not Go allocation history. Cold native-link memory was investigated in the [interruption audit](provider-interruption-continuation.md). These runs use different workloads and instrumentation, so their totals do not provide a speedup comparison.

Raw Go/Node profiles, matching test/probe binaries, temporary fixtures and disposable logs were removed after analysis. No release, deployment or live-provider billing validation was performed.
