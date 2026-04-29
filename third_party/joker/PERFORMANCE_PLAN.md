# Joker performance plan for gi

Status: Active
Date: 2026-04-28

This plan tracks the work needed to drastically improve the performance of the vendored Joker runtime used by gi.

## Goals

1. Build a **repeatable benchmark and profiling harness** inside the fork.
2. Optimize the **current interpreter/runtime hot paths** using benchmark-guided changes.
3. Introduce a **lowered execution path** that can evolve into a compiled/bytecode path.
4. Preserve language compatibility for the subset exercised by gi scripting.

## Constraints

- Keep the fork buildable inside gi without external services.
- Prefer incremental, measurable changes.
- Avoid semantic drift from upstream Joker unless explicitly documented.
- Optimize the subset relevant to gi first: loops, recursion, numeric ops, text processing, JSON, map updates, and native bridge calls.

## Workstreams

### Workstream A — Benchmark + profiling harness

Add benchmark coverage directly in `core/` for:

- recursive fibonacci
- tight arithmetic loop
- word frequency over text
- JSON parse + projection
- realistic gi-style bridge/data transforms later

Deliverables:

- Go benchmark file in `core/`
- helper for parse-once / eval-many benchmarking
- baseline numbers recorded before and after optimizations
- optional CPU/alloc profiling instructions in benchmark comments or this file

### Workstream B — Interpreter/runtime hotspot optimization

Profile and optimize the current evaluator first.

Primary suspects:

- `CallExpr.Eval`
- `Fn.Call`
- numeric procs in `core/procs.go`
- sequence iteration and allocation in loops
- local binding lookup / frame handling
- persistent map update overhead for common patterns

Initial optimization targets:

1. numeric fast paths for `+`, `*`, `rem`, `inc`, `dec`, comparisons
2. reduce common allocations in argument evaluation / call dispatch
3. optimize common lookup/update patterns used by text and JSON workloads

### Workstream C — Lowered execution path

Do not jump directly to a bytecode VM.

Stage the work as:

1. define a lowered internal IR for a hot subset
2. interpret that IR with slot-resolved locals
3. add optional bytecode emission later if the IR pays off

Target subset for first lowering pass:

- literals
- local binding load/store
- `let`
- `if`
- `do`
- `loop/recur`
- direct function calls
- primitive arithmetic/comparison ops

### Workstream D — Validation

For every optimization stage:

- rerun benchmarks
- compare alloc/op and ns/op
- confirm gi scripting tests still pass
- keep changes bisectable

## Milestones

### Milestone 1 — Baseline and first wins

- [x] Add benchmark harness in `core/` (`core/perf_bench_test.go`)
- [x] Capture initial numbers for arithmetic/fib/text benchmarks
- [x] Land first runtime fast paths in arithmetic/comparison ops (`+`, `*`, `rem`, `<`, `=`, `inc`, `dec`)
- [ ] Record deltas against a clean pre-optimization baseline run

Current benchmark checkpoint (`go test ./core -run '^$' -bench 'BenchmarkEval(ArithmeticLoop|RecursiveFib|WordFrequency)$' -benchmem -benchtime=5x` on the gi dev host):

- arithmetic loop: ~172ms, ~77.6 MB, ~2.70M allocs
- recursive fib: ~393ms, ~190.9 MB, ~6.98M allocs
- word frequency: ~10.1ms, ~6.5 MB, ~172k allocs

Additional findings and changes landed so far:

- `evalLoop` no longer allocates a replacement `LocalEnv` frame on every `recur`; it now reuses the existing loop frame by replacing `env.bindings` in place.
- `CallExpr.Eval` / `RecurExpr.Eval` now use fixed-size stack-backed arg arrays for 0–4 argument common cases before falling back to generic allocation.
- `CallExpr.Eval` now also dispatches directly to `Proc.Fn` and `*Fn` in the common paths before falling back to the generic `Callable` interface path.
- binding resolution now fast-paths current-frame and parent-frame lookups before falling back to the generic environment walk.
- `Fn.Call`, `FnExpr.Eval`, `LetExpr.Eval`, and `LoopExpr.Eval` now build child environments inline instead of routing through helper methods that return new frame pointers.
- numeric fast paths now cover `+`, `-`, `*`, `rem`, `<`, `=`, `inc`, `dec`, and `zero?` for common `Int`/`Double` cases.
- `CallExpr.Eval` now additionally fast-paths a hot subset of builtin proc execution directly in the evaluator (`procAdd`, `procSubtract`, `procMultiply`, `procRem`, `procLt`, `procEq`, `procInc`, `procDec`, `procIsZero`) to avoid temporary `[]Object` allocation on extremely hot call sites.
- `SeqNth` / `SeqTryNth` now fast-path `*ArraySeq` directly by index instead of repeatedly traversing via `Rest()`, which removed a large amount of allocation from sequence-indexing-heavy workloads.
- profiling showed the original word-frequency benchmark was dominated by `ArraySeq.Rest` allocation churn triggered indirectly by repeated `nth` over sequences.
- the main remaining hotspots for arithmetic/recursive workloads are still evaluator dispatch and allocation pressure, but the temporary arg-slice allocation component has now been reduced materially.

These are still early checkpoint numbers, but the harness is now in place and later passes should keep splitting work by workload type: arithmetic/recursion vs. sequence/map-heavy text processing.

### Milestone 2 — Call/eval optimization

- [x] Profile arithmetic-loop hotspot shape with `pprof`
- [x] Reduce some call-path allocations for small arg-count call/recur cases
- [ ] Reduce evaluator dispatch overhead further (`Eval`, `IfExpr`, `Fn.Call`)
- [ ] Record deltas against a clean pre-optimization baseline branch/run

### Milestone 3 — Lowered IR prototype

- [ ] Design minimal IR types
- [ ] Lower simple loop/arithmetic subset
- [ ] Add IR evaluator behind an internal switch
- [ ] Benchmark against interpreter

### Milestone 4 — Compiled execution expansion

- [ ] Extend lowered path to more forms used by gi
- [ ] Decide whether to freeze at IR or continue to bytecode VM

## Files of interest

- `core/eval.go`
- `core/object.go`
- `core/procs.go`
- `core/numbers.go`
- `core/map.go`
- `core/vector.go`
- `core/array_vector.go`
- future benchmark files in `core/*_test.go`

## Current execution order

1. Add baseline benchmark file
2. Optimize hot numeric/runtime paths
3. Re-benchmark
4. Start IR design after measured gains from interpreter changes
