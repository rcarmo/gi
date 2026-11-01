# Interrupted provider continuation (#92)

Base: main `ec2c786a`, Go 1.27.3, 2026-11-01. This increment continues one unsettled provider admission after a versioned generation-ready frontier. It does not cover every retry/tool boundary, and does not close #83/#86/#92.

## Transaction and continuation

The recovery reader accepts a single `inference.started` and optional matching `inference.request_prepared` suffix. Session, iteration and phase must match the saved frontier. It decodes typed identity fields, not an executable request. A settled response, tool admission, retry event, duplicate/foreign request, unsupported policy or type-invalid header rejects this route. SQLite's JSON expression indexes already reject syntactically invalid JSON; the non-object-header regression uses a valid JSON array.

Claim rotation, old-writer fencing, the interrupted receipt, retained partial row, new ready frontier and recovery event commit in one transaction. Receipt/message/event/phase/session write faults roll back ownership as well as accounting. A crash immediately after this commit can recover the new frontier without duplicating the interrupted receipt.

`interrupted_boundary` requires unknown usage and cannot schedule a provider retry. `inference.interrupted` is bookkeeping, not meaningful progress: it does not reset the five-admission no-progress budget. Usage from a valid committed snapshot remains counted once and uncertain; absent usage is not recorded as a known zero. The generation index remains the same, while both physical request occurrences retain separate receipts. Exact remote billing is not promised.

Recovery derives retained display text from the structured partial, not mutable UI text. The original prefix, thinking/signature data and unadmitted tool prefix remain on the existing unadopted row. Failed output stays outside model context. The resumed host rebuilds current instructions, selected model and tool definitions, reapplies accepted lifecycle policy, and runs fresh context/payload hooks. It never sends the old provider payload or executes a tool-call prefix. Prompt/start/approval hooks are not replayed.

Provider settlement now has a shared transaction helper; normal live settlement keeps its validation, cancellation, idempotence and commit semantics. Conservative partial sealing also uses the interrupted event type and structured display text.

## Regressions and verification

- Store: silent/partial interrupted outcomes; detached policy; exact ownership and stale-writer fences; admitted-only uncertainty; transaction faults; no-progress exhaustion; rejection of settled/tool/retry/foreign/ambiguous/type-invalid boundaries; interruption cannot claim known usage.
- SIGKILL integration: kill and join a provider owner after admission and request preparation, with and without committed progress. Reopen SQLite, change model/instructions/tool definition, then resume after runtime readiness. Assert one fresh request, one original prompt, no old payload or partial contribution in context, no tool/start/approval replay, preserved signature/prefix, and uncertain received usage without double counting.
- Shared Store: a stale-timestamp observer cannot fence, seal, replay or cancel a still-live provider host. Closing the observer leaves the owner's completion intact.
- Chromium/WebKit: both ready and admitted-provider frontiers surface one unavailable-provider error, drain queued follow-up work and retain one original prompt across reload. The admitted case also retains the committed prefix rather than a tampered draft. Successful native continuation is covered by the process tests; the browser harness has no native inference provider.

Sequential Makefile checks passed:

- `make test-provider-interruption PROFILING=1 PROFILE_KEEP=1`: four packages, 90 tests, race detector, three repetitions.
- `make test-durable-recovery`: ownership/Close/parent/cancellation race matrix, three repetitions.
- `make test RACE= PROFILING=1`: 35 packages, 2,788 tests.
- `make vet RACE=`, `make bun-checks RACE=`.
- `make test-generation-ready-web RACE=`: eight functional regressions.
- Fresh isolated `make test-ux RACE= PROFILING=1`: 153 passed, 13 skipped, including web/native builds. A monitored repeat also passed 153/13.

## Profiling and disposal

Inspected CPU, `alloc_space` and `alloc_objects`. The focused race profiles are dominated by checkptr/TSAN CPU and test-selection regexp allocations. Store totals were 107.88 MB and 469,381 sampled objects; acquisition including nested receipt settlement accounted for 0.82 s cumulative CPU, 4 MB and 53,213 objects. The turn package's corresponding acquisition cost was 0.09 s CPU and 2.05 MB. These recovery-heavy tests overrepresent acquisition. Its work is bounded to a verified journal suffix, one receipt transaction and a capped number of restarts; it does not decode the old request into a reusable provider payload.

The non-race full run's material allocations remain SQLite rows, image-bound fixtures, WASM growth and JSON/base64 processing. The extracted settlement body retains the existing receipt/accounting work; no new ordinary-request allocation loop was introduced. No speedup is claimed from different workloads or wrapper baselines. Node browser profiles describe harness CPU/live heap, not Go allocation history.

An earlier browser wrapper reported a 1,389 MB peak. That counter is `RUSAGE_CHILDREN`, not aggregate application memory. A monitored equivalent browser run reported 399 MB there and an owned-server `VmHWM` of 126,784 KiB (about 124 MiB). A fresh native build reproduced 1,402,496 KiB peak; timing the Go build separately reproduced 1,363,740 KiB with a `GOGC=50` probe. The large transient belongs to building/linking the roughly 123 MiB executable, not server heap growth. The GC probe showed no material benefit worth adopting; runtime/build GC defaults remain unchanged. Cold native linking still needs roughly 1.4 GB of build memory; revalidate this constraint on a lower-memory CI host.

Raw Go/Node profiles, matching/probe binaries, temporary fixtures and disposable logs were deleted after analysis. Only these conclusions remain. No release, deployment or live-provider billing validation was performed.

## Remaining work

Unknown outcomes reached through retry boundaries without a fresh generation frontier remain held, as do partial tool batches and ambiguous/legacy states. Per-tool interrupted receipts, persisted/current ReplaySafe decisions with original prepared arguments, explicit parked-checkpoint retry and general checkpoint-preserving Close still need verified routes. Settled responses and tools must never be replayed to fill those gaps.
