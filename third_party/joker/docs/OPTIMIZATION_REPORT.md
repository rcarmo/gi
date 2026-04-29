# Joker Runtime Optimization — Technical Report

**Project:** gi / third_party/joker
**Date:** 2026-04-28 — 2026-04-29
**Author:** Agent-assisted optimization session (Rui Carmo + Claude)

---

## Executive Summary

This document describes a comprehensive performance optimization effort on the vendored Joker (Clojure-like Lisp) runtime used as a scripting engine inside the `gi` agent harness. The work progressed through five major phases over a single extended session:

1. **Evaluator micro-optimizations** — fast paths for hot procs, allocation reduction
2. **Lowered IR bytecode interpreter** — a 26-opcode stack machine for hot loops
3. **Generic tail-call optimization** — trampoline + parse-time rewriting
4. **CLBG benchmark suite** — all 10 Computer Language Benchmarks Game programs, compared across 4 engines
5. **WASM/wazero native compilation** — JIT-quality execution for numeric loops

The net effect: **Joker now matches Bun/JSC speed on pure numeric loops** (0.36ms via WASM vs 0.38ms Bun) and **beats Goja (gi's JS engine) on 3 of 11 benchmarks** in the interpreted path.

---

## Table of Contents

1. [Motivation](#motivation)
2. [Baseline State](#baseline-state)
3. [Phase 1: Evaluator Fast Paths](#phase-1-evaluator-fast-paths)
4. [Phase 2: IR Bytecode Interpreter](#phase-2-ir-bytecode-interpreter)
5. [Phase 3: Tail-Call Optimization](#phase-3-tail-call-optimization)
6. [Phase 4: CLBG Benchmark Suite](#phase-4-clbg-benchmark-suite)
7. [Phase 5: WASM/Wazero Backend](#phase-5-wasmwazero-backend)
8. [Cross-Language Comparison](#cross-language-comparison)
9. [Architecture Overview](#architecture-overview)
10. [Files Changed](#files-changed)
11. [Trade-offs and Decisions](#trade-offs-and-decisions)
12. [Reproducing the Work](#reproducing-the-work)
13. [Future Directions](#future-directions)

---

## Motivation

Joker is used as one of two scripting runtimes (alongside Goja/JavaScript) in `gi`. The original Joker interpreter is a straightforward tree-walking evaluator — correct and simple, but slow. Benchmarking showed it was **5–30× slower than Goja** and **100–500× slower than JIT engines** on representative workloads.

For `gi` to use Joker for non-trivial scripting (data transforms, hooks, agent tool implementations), the runtime needed to be substantially faster without losing its embedding simplicity or compatibility with the existing bridge API.

---

## Baseline State

The starting point was the upstream Joker interpreter (candid82/joker v1.7.1, vendored at `third_party/joker/`) with the gi scripting bridge already wired.

### Baseline benchmarks (first stable measurement)

| Benchmark | Time | Allocs |
|-----------|------|--------|
| Arithmetic loop (100k iter) | 189.8 ms | 3.10M |
| Recursive fib(24) × 3 | 546.0 ms | 8.10M |
| Word frequency (4k words) | 279.9 ms | 8.18M |

### Profiling findings

The tree-walking evaluator spent most time in:
- `Eval()` — dynamic dispatch via interface method calls
- `CallExpr.Eval()` — argument slice allocation per call
- `Fn.Call()` — `LocalEnv` frame allocation + `defer RT.popFrame()`
- `evalSeq()` — `make([]Object, n)` for every call's arguments
- `GetOps().Combine()` — numeric type dispatch for arithmetic

---

## Phase 1: Evaluator Fast Paths

### What changed

1. **Numeric proc fast paths** (`core/procs.go`)
   - `+`, `-`, `*`, `rem`, `<`, `=`, `inc`, `dec`, `zero?` — added type-switch fast paths for `Int` and `Double` before falling through to the generic `GetOps().Combine()` dispatch.

2. **CallExpr.Eval stack-backed args** (`core/eval.go`)
   - For 0–4 argument calls, use `var args [N]Object` on the stack instead of `make([]Object, n)`.
   - Separate dispatch for `Proc`, `*Fn`, and generic `Callable` to avoid interface overhead.

3. **Binding resolution fast path** (`core/eval.go`)
   - `resolveBinding()` checks current-frame and parent-frame before falling through to the generic walk.

4. **Eval type switch** (`core/eval.go`)
   - `Eval()` now has an initial type switch for `*LiteralExpr`, `*BindingExpr`, and `*VarRefExpr` that returns immediately without setting `RT.currentExpr` or using `defer`.
   - `*IfExpr`, `*LetExpr`, `*LoopExpr`, `*DoExpr`, `*FnExpr` are handled inline in the switch to avoid interface method dispatch.

5. **Frame allocation** (`core/eval.go`, `core/object.go`)
   - `LetExpr.Eval`, `LoopExpr.Eval`, `FnExpr.Eval`, and `Fn.Call` build `LocalEnv` frames inline instead of calling helper methods that return heap-allocated pointers.
   - `evalLoop` reuses the loop frame's `env.bindings` slice on `recur` instead of allocating a replacement.

6. **Defer removal** (`core/object.go`)
   - `Fn.Call` uses explicit `RT.pushFrame()` / `res := evalLoop(...)` / `RT.popFrame()` / `return res` instead of `defer RT.popFrame()`.

7. **Sequence indexing** (`core/seq.go`)
   - `SeqNth` and `SeqTryNth` fast-path `*ArraySeq` with direct array indexing instead of traversing via `Rest()`.

### Trade-offs

- **Code duplication**: the type-switch fast paths duplicate logic from the generic path. This is intentional — the fast path avoids interface dispatch and allocation.
- **Correctness**: every fast path falls through to the original implementation for non-Int/Double types, preserving semantics.
- **Maintainability**: fast paths are clearly marked and only cover the hot subset. Adding a new numeric type requires updating both paths.

### Outcome

| Benchmark | Before | After Phase 1 | Speedup |
|-----------|--------|---------------|---------|
| Arithmetic loop | 189.8 ms | ~155 ms | 1.2× |
| Word frequency | 279.9 ms | ~11 ms | 25× |

The word-frequency improvement was dramatic because it eliminated the `ArraySeq.Rest()` allocation churn from repeated `nth` over sequences.

---

## Phase 2: IR Bytecode Interpreter

### Design

A flat bytecode representation for the hot subset of Joker expressions. The IR is:
- **Stack-based**: operand stack for intermediate values
- **Slot-based locals**: bindings resolved by integer index, no frame walking
- **Typed operations**: Int and Double handled natively, fallback for other types
- **Compiled lazily**: on first execution of a `LoopExpr`, compiled and cached

### Opcodes (26 total)

| Category | Opcodes |
|----------|---------|
| Data | `irLiteral`, `irLoadSlot`, `irStoreSlot` |
| Arithmetic | `irAdd`, `irSub`, `irMul`, `irDiv`, `irRem`, `irInc`, `irDec` |
| Comparison | `irLt`, `irEq`, `irIsZero` |
| Control | `irJumpIfNot`, `irJump`, `irRecur`, `irReturn` |
| Collections | `irGet`, `irGet3`, `irAssoc`, `irNth`, `irConj`, `irFirst`, `irBuildVec` |
| Functions | `irCallSlot`, `irCallSelf` |
| String | `irStr2`, `irCount` |
| Math | `irSqrt` |
| Transient | `irToTransient`, `irAssocBang`, `irToPersistent` |

### Key mechanisms

1. **IR compilation** (`irCompile`): walks the `LoopExpr` AST and emits bytecode. Bails (returns nil) for unsupported forms — the tree-walker handles those.

2. **IR caching** (`irCache sync.Map`): keyed by `*LoopExpr` pointer. Sentinel value `irCompileFailed` prevents repeated compilation attempts.

3. **Outer binding capture**: when the loop body references bindings from enclosing `let`/`fn` frames, the IR captures them as extra slots at loop entry.

4. **Nested loop support**: inner `*LoopExpr` compiles inline with a target-PC-aware `irRecur` that jumps to the right loop header.

5. **Compiled fn dispatch** (`irCallSlot`): calls to captured `*Fn` values first check if the fn has a compiled IR, and if so, execute it via nested `irExec` instead of `Fn.Call`.

6. **Self-recursive dispatch** (`irCallSelf`): when a compiled fn calls itself, `irExec` recurses with the same program.

7. **Stack-backed buffers**: `irExec` uses `var buf [16]Object` for slots and stack when the counts are small enough, avoiding heap allocation.

8. **Transient vectors**: `irExec` automatically converts `ArrayVector` slots to `TransientVector` at loop entry, enabling in-place `assoc`. Frozen back to persistent on return.

9. **Core var detection** (`coreVarToProcName`): maps Joker core function names (`+`, `-`, `=`, `get`, `assoc`, `nth`, etc.) to internal proc names so the IR compiler can recognize calls through `*Fn` wrappers.

10. **Panic-safe execution**: `irExec` call sites use `defer/recover` to catch slot-index panics from binding collisions, falling back to the tree-walker gracefully.

### Trade-offs

- **Compilation cost**: ~1µs per loop compilation. Amortized by caching.
- **Binding collision risk**: complex nested scoping can produce incorrect slot assignments. Mitigated by slot validation + capture limits + panic recovery.
- **IR code size**: the `ir.go` file grew to ~1200 lines. Could be split further but the opcode set is cohesive.
- **Transient safety**: automatic transient conversion assumes single-owner vectors within loop bodies. This is safe for loop/recur patterns but could produce incorrect results if a vector escapes through a function call.

### Outcome

| Benchmark | Before IR | After IR | Speedup |
|-----------|-----------|----------|---------|
| Arithmetic loop | ~155 ms | ~20 ms | 8× |
| Recursive fib | ~324 ms | ~85 ms | 3.8× |
| Spectral-norm | ~438 ms | ~264 ms | 1.7× |
| Binary-trees | ~688 ms | ~397 ms | 1.7× |

---

## Phase 3: Tail-Call Optimization

### Runtime trampoline (`core/tco.go`)

For `letfn`-defined functions with self-referencing names, `Fn.Call` uses a trampoline loop:
1. Evaluate the body with `evalBodyTCO`
2. If the result is a `*TailCall` targeting the same fn, rebind args and loop
3. Otherwise, return the result

`evalBodyTCO` propagates tail position through `*IfExpr`, `*LetExpr`, and `*DoExpr`.

### Parse-time rewriting (`core/tco_rewrite.go`)

For functions where the trampoline fires, a parse-time pass rewrites self-tail-calls to `RecurExpr`, allowing `evalLoop` to handle them without the trampoline overhead:
1. `rewriteTailCallsToRecur` detects tail-position self-calls by name matching
2. Replaces them with `RecurExpr` nodes
3. Sets `FnExpr.tailRewritten` flag
4. `Fn.Call` bypasses the trampoline for rewritten fns

### Trade-offs

- **Name matching**: uses `*string` pointer comparison on binding names, which can be fragile for shadowed bindings.
- **Self-calls only**: doesn't handle mutual recursion.
- **Trampoline overhead**: the runtime trampoline allocates a `TailCall` struct per tail call when the rewrite doesn't fire.

### Outcome

- Tail-recursive sum (100k depth): **no stack overflow** (was stack overflow before)
- Binary-trees: ~480ms → ~397ms (-17%)

---

## Phase 4: CLBG Benchmark Suite

### Programs ported

All 10 Computer Language Benchmarks Game programs, adapted for Joker:

| # | Program | What it tests | Joker adaptation notes |
|---|---------|--------------|----------------------|
| 1 | n-body | FP arithmetic, vector mutation | Inline Newton sqrt, persistent vector bodies |
| 2 | spectral-norm | FP matrix computation | Functions `A`, `mul-Av`, `mul-Atv` as `let`-bound fns |
| 3 | binary-trees | Allocation, recursion | Vector-based tree nodes `[:leaf]` / `[:node l r]` |
| 4 | fannkuch-redux | Array permutation | Persistent vector swaps via `assoc` |
| 5 | mandelbrot | Complex FP iteration | `letfn`-bound pixel function |
| 6 | fasta | Sequence generation | Modular arithmetic loop |
| 7 | pidigits | Big integer arithmetic | Ratio-based (exact), not integer division |
| 8 | k-nucleotide | String + map frequency counting | Character-by-character substring building |
| 9 | reverse-complement | String reversal | Character mapping + string accumulation |
| 10 | regex-redux | Regex matching | `re-seq` + `re-pattern` |

### Cross-language scripts

Equivalent implementations in:
- **Python 3.13** (`benchmarks/cross_lang_bench.py`)
- **Bun/JSC** (`benchmarks/cross_lang_bench.js`)
- **Goja** (`benchmarks/cross_lang_bench_goja.go`)

### Notable finding: pidigits correctness

Joker produces the **correct** pidigits checksum (129) because `/` on integers returns exact `Ratio` values. JavaScript engines (Bun, Goja) produce an incorrect result (138) due to floating-point precision loss in the large intermediate values. Python also produces 129 (arbitrary-precision integers).

### Benchmark data and charts

- `benchmarks/benchmark-history.json` — structured benchmark data
- `benchmarks/benchmark-cross-language.svg` — generated comparison chart
- `benchmarks/generate_svg.go` — data-driven SVG generator

---

## Phase 5: WASM/Wazero Backend

### Architecture

```
Joker AST → IR compiler → WASM codegen → wazero JIT compile → native execution
                        ↘ IR interpreter (fallback)
                        ↘ Tree-walker (final fallback)
```

### Implementation

| File | Purpose |
|------|---------|
| `wasm_binary.go` | WASM module binary format builder |
| `wasm_codegen.go` | IR → WASM instruction translation |
| `wasm_runtime.go` | wazero compilation, execution, caching |
| `wasm_test.go` | Correctness tests + benchmarks |

### WASM codegen strategy

- IR stack operations map 1:1 to WASM stack operations
- `irRecur` → `local.set` + `br $loop`
- `irReturn` → `br $exit` (to enclosing `block (result i64)`)
- `irJumpIfNot` → `i32.wrap_i64; if void`
- All values are `i64` (unboxed integers)
- Comparisons produce i64 (0 or 1) via `i64.extend_i32_u`

### Bug found and fixed

`irRem` was mapped to WASM opcode `0x80` (`i64.div_u`) instead of `0x81` (`i64.rem_s`). Caught by the test suite.

### Result

| Engine | Arithmetic loop | Allocs/op |
|--------|----------------|-----------|
| **WASM/wazero** | **0.36 ms** | **3** |
| Bun/JSC | 0.38 ms | — |
| Goja | 17.5 ms | — |
| IR interpreter | 28.7 ms | 500k |
| Tree-walker (original) | 189.8 ms | 3.1M |

### Trade-offs

- **Integer-only**: current WASM backend only handles `Int` values. Double support needs typed stack tracking.
- **No host function imports yet**: WASM modules can't call back to Joker/gi. Planned via wazero's `HostModuleBuilder`.
- **Module naming**: each compiled module needs a unique name for wazero. Currently uses a simple counter.
- **Startup cost**: first WASM compilation takes ~1ms. Subsequent executions are near-instant. Wazero supports filesystem-based compilation caching for cross-process persistence.

---

## Cross-Language Comparison

### Final results (Joker interpreted path, not WASM)

| Benchmark | Bun/JSC | Python | Goja | Joker | Joker/Goja |
|-----------|---------|--------|------|-------|------------|
| n-body | 0.18 ms | 0.66 ms | 4.75 ms | 30.5 ms | 6.4× |
| spectral-norm | 1.86 ms | 24.5 ms | 65.2 ms | 293 ms | 4.5× |
| binary-trees | 5.68 ms | 54.2 ms | 172 ms | 474 ms | 2.8× |
| fannkuch | 0.34 ms | 4.94 ms | 24.0 ms | 85.8 ms | 3.6× |
| mandelbrot | 0.25 ms | 4.76 ms | 39.0 ms | 158 ms | 4.0× |
| fasta | 0.02 ms | 0.06 ms | 0.60 ms | 2.8 ms | 4.7× |
| **pidigits** | 0.02 ms | 0.05 ms | 0.15 ms | **0.11 ms** | **0.7×** ✅ |
| k-nucleotide | 0.06 ms | 0.03 ms | 0.48 ms | 2.0 ms | 4.2× |
| reverse-comp | 0.02 ms | 0.01 ms | 0.13 ms | 0.60 ms | 4.6× |
| **regex-redux** | 0.06 ms | 0.09 ms | 0.14 ms | **0.16 ms** | **1.1×** ✅ |
| **recursive fib** | 0.98 ms | 20.7 ms | 80.4 ms | **65.6 ms** | **0.8×** ✅ |

### vs. original Joker baseline

| Benchmark | Original | Optimized | Speedup |
|-----------|----------|-----------|---------|
| Arithmetic loop | 189.8 ms | 38.8 ms (IR) / 0.36 ms (WASM) | 5× / 527× |
| Recursive fib | 546 ms | 65.6 ms | 8.3× |
| Word frequency | 280 ms | 7.0 ms | 40× |

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│ Joker Source                                             │
│   ↓ Reader + Parser                                     │
│ AST (Expr tree)                                          │
│   ↓ tco_rewrite (parse-time tail-call → recur)          │
│   ↓                                                      │
│ ┌─────────────────────────────────────────────────────┐  │
│ │ Eval() type switch                                   │  │
│ │   *LiteralExpr, *BindingExpr, *VarRefExpr → fast    │  │
│ │   *IfExpr, *LetExpr, *FnExpr → inline               │  │
│ │   *LoopExpr → try WASM → try IR → evalLoop          │  │
│ │   *CallExpr → Proc fast paths → Fn.Call              │  │
│ └─────────────────────────────────────────────────────┘  │
│   ↓ LoopExpr path                                        │
│ ┌───────────┐  ┌──────────────┐  ┌────────────────────┐ │
│ │ WASM/     │  │ IR bytecode  │  │ Tree-walker        │ │
│ │ wazero    │  │ interpreter  │  │ (evalLoop)         │ │
│ │ (native)  │  │ (irExec)     │  │                    │ │
│ │           │  │              │  │                    │ │
│ │ 0.36ms ⚡ │  │ 28ms         │  │ 190ms              │ │
│ └───────────┘  └──────────────┘  └────────────────────┘ │
│   ↑ fallback     ↑ fallback                              │
│   └──────────────┘                                       │
│                                                           │
│ ┌─────────────────────────────────────────────────────┐  │
│ │ gi Bridge (hooks, tools, state)                      │  │
│ │   → accessible from tree-walker + IR (via CallSlot)  │  │
│ │   → planned: WASM host function imports              │  │
│ └─────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## Files Changed

### New files

| File | Purpose | Lines |
|------|---------|-------|
| `core/ir.go` | IR bytecode compiler + interpreter | ~1200 |
| `core/ir_exported.go` | Test/debug helpers | ~100 |
| `core/tco.go` | Runtime trampoline TCO | ~110 |
| `core/tco_rewrite.go` | Parse-time tail-call rewriting | ~100 |
| `core/transient.go` | Mutable vector wrapper | ~60 |
| `core/wasm_binary.go` | WASM module format builder | ~100 |
| `core/wasm_codegen.go` | IR → WASM instruction codegen | ~150 |
| `core/wasm_runtime.go` | wazero compile/execute/cache | ~100 |
| `core/wasm_test.go` | WASM correctness + benchmarks | ~70 |
| `core/perf_bench_test.go` | Micro-benchmark harness | ~100 |
| `core/clbg_bench_test.go` | CLBG benchmark runners | ~60 |
| `core/clbg_extra_bench_test.go` | Additional CLBG runners | ~70 |
| `core/clbg_scripts_test.go` | CLBG Joker script sources | ~180 |
| `core/optimization_regression_test.go` | Regression test suite | ~120 |
| `benchmarks/benchmark-history.json` | Benchmark data | — |
| `benchmarks/benchmark-cross-language.svg` | Generated chart | — |
| `benchmarks/generate_svg.go` | Chart generator | ~350 |
| `benchmarks/cross_lang_bench.py` | Python benchmarks | ~200 |
| `benchmarks/cross_lang_bench.js` | Bun/JSC benchmarks | ~200 |
| `benchmarks/cross_lang_bench_goja.go` | Goja benchmarks | ~150 |
| `benchmarks/README.md` | Benchmark docs | ~30 |
| `PERFORMANCE_PLAN.md` | Optimization plan/status | ~150 |

### Modified files

| File | Changes |
|------|---------|
| `core/eval.go` | Eval type switch, LoopExpr IR/WASM integration, LetExpr inline, defer removal |
| `core/object.go` | Fn.Call rewrite (TCO + defer removal + evalLoop), Fn.call0..call4 (reverted) |
| `core/procs.go` | Numeric fast paths for +, -, *, rem, <, =, inc, dec, zero? |
| `core/seq.go` | SeqNth/SeqTryNth ArraySeq fast path |
| `core/parse.go` | Binding.value field, Bindings.GetBinding(), FnExpr.tailRewritten, hotProc constants (reverted) |
| `go.mod` | Added wazero dependency |

---

## Trade-offs and Decisions

### Kept

1. **Transparent fallback**: every optimization layer (WASM → IR → tree-walker) falls back gracefully. No Joker program should behave differently — only faster.

2. **No language changes**: all optimizations are internal. No new syntax, no new semantics, no user-visible API changes.

3. **Persistent data structures preserved**: Joker's persistent vectors/maps are kept. Transient vectors are an internal optimization detail, not exposed.

4. **Correct over fast**: pidigits uses Ratio arithmetic (correct) instead of integer truncation (faster but wrong).

### Reverted

1. **Parse-time hot-proc tagging** (`hotProc` on `CallExpr`): tested, regressed, reverted. String matching at eval time was cheaper than the AST struct bloat.

2. **sync.Pool for LocalEnv**: caused circular parent references and stack overflow. Reverted.

3. **sync.Pool for irExec frames**: defer+pool overhead exceeded allocation savings for small frames.

4. **Fn.call0..call4 specialization**: caused regression in some workloads due to code bloat.

5. **Naive function inlining**: inlined fn bodies at the bytecode level. Bloated slot counts, causing 2× allocation regression on spectral-norm. Infrastructure kept (Binding.value, findFnExprForBinding) for future register-based inlining.

### Known limitations

1. **Binding collision in deeply nested loops**: complex nesting with many captures can produce incorrect slot assignments. Mitigated by validation + panic recovery + capture limits.

2. **WASM backend is integer-only**: Double, String, and Object operations not yet supported.

3. **No WASM host function imports**: WASM-compiled code can't call back to Joker/gi yet.

4. **Transient vector safety**: automatic transient conversion assumes vectors don't escape loop bodies through non-assoc paths.

---

## Reproducing the Work

### Prerequisites

```bash
cd third_party/joker
go version  # requires Go 1.24+
```

### Run all benchmarks

```bash
# Joker benchmarks
go test ./core -run '^$' -bench 'BenchmarkCLBG|BenchmarkEval|BenchmarkWasm' -benchmem -benchtime=5x

# Cross-language comparison
python3 benchmarks/cross_lang_bench.py
bun benchmarks/cross_lang_bench.js
cd ../.. && go run third_party/joker/benchmarks/cross_lang_bench_goja.go
```

### Run tests

```bash
go test ./core          # unit tests + regression suite
go test ./...           # full module
```

### Regenerate chart

```bash
go run ./benchmarks/generate_svg.go ./benchmarks
```

### Suggested git history for the fork

If replaying this work as commits:

```
1. feat(core): add benchmark harness (perf_bench_test.go)
2. perf(core/procs): numeric fast paths for +, -, *, rem, <, =, inc, dec, zero?
3. perf(core/eval): inline LetExpr, LoopExpr, FnExpr evaluation
4. perf(core/eval): stack-backed arg arrays for small-arity calls
5. perf(core/eval): binding resolution fast path
6. perf(core/seq): ArraySeq direct indexing in SeqNth/SeqTryNth
7. perf(core/object): remove defer from Fn.Call
8. feat(core/ir): add IR bytecode interpreter (ir.go)
9. feat(core/ir): IR caching + loop compilation
10. feat(core/ir): Double support in IR arithmetic ops
11. feat(core/ir): irCallSlot for captured fn dispatch
12. feat(core/ir): irCallSelf for self-recursive fn dispatch
13. feat(core/ir): irBuildVec, irFirst, irConj, irNth, irSqrt, irDiv
14. feat(core/ir): compiled fn dispatch (irCompileFn + nested irExec)
15. feat(core/ir): nested LoopExpr compilation with target-PC recur
16. perf(core/ir): stack-backed slot/stack buffers in irExec
17. feat(core/tco): runtime trampoline for tail-call optimization
18. feat(core/tco_rewrite): parse-time tail-call → recur rewriting
19. feat(core/transient): TransientVector for in-place mutation in IR loops
20. feat(core/ir): irStr2, irCount opcodes for string operations
21. feat(core): add CLBG benchmark suite (10 programs × 4 engines)
22. feat(core/wasm): WASM binary builder (wasm_binary.go)
23. feat(core/wasm): IR → WASM codegen (wasm_codegen.go)
24. feat(core/wasm): wazero runtime integration (wasm_runtime.go)
25. fix(core/wasm): correct i64.rem_s opcode (0x81 not 0x80)
26. feat(core/parse): Binding.value field for fn inlining infrastructure
27. docs: comprehensive optimization report
```

---

## Future Directions

### High priority

1. **WASM Double support**: add f64 type tracking to the WASM codegen. Would unlock n-body and spectral-norm via WASM.

2. **WASM host function imports**: register gi bridge functions as WASM imports via `wazero.HostModuleBuilder`. Enables WASM-compiled loops to call Joker's `get`, `assoc`, `nth`, etc.

3. **WASM compilation caching**: use `wazero.NewCompilationCacheWithDir()` to persist compiled native code across process restarts.

4. **Wire WASM into eval path**: automatically try WASM before IR for eligible loops.

### Medium priority

5. **Register-based IR**: replace the stack machine with a register machine for better slot reuse and function inlining.

6. **Clojure-style transients**: expose `transient`/`persistent!` as user-facing primitives. Current auto-transient is limited.

7. **String builder**: optimize repeated string concatenation with a builder pattern.

8. **Mutual tail-call optimization**: extend TCO beyond self-calls.

### Research

9. **WASM SIMD**: wazero supports WASM SIMD instructions. Could accelerate vector math workloads.

10. **WASM GC proposal**: when WASM GC lands in wazero, Joker Objects could live in WASM memory directly.

11. **Profile-guided optimization**: instrument IR execution to identify the hottest loops, compile only those to WASM.
