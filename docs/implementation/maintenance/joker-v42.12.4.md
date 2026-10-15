# Native Joker v42.12.4

Gi pins `github.com/rcarmo/go-joker/v42 v42.12.4`, tag commit `239f297d2f493e8ac7121fbda0298b2c77d17e61`. Module provenance and sums were verified on 2026-10-07. go-ai remains v1.0.4; the frontend and historical Joker reference dataset are unchanged.

## Checked numeric buffers

The native Joker host now exposes the release's checked dense-buffer API through its existing `joker.jit` registration. Only emitted numeric kernels run through wazero; Joker itself remains native and QuickJS remains the only whole-runtime isolate.

```clojure
(let [values (jit/numeric-buffer 4)
      fill! (jit/compile-wasm
              (fn [a]
                (loop [i 0]
                  (if (< i 4)
                    (do (jit/buffer-set! a i (* i 0.25))
                        (recur (+ i 1)))
                    (jit/buffer-get a 3))))
              {:buffers [0]})]
  [(fill! values) (jit/wasm-engine)])
;; [0.75 "compiler"] with JOKER_WASM_ENGINE=compiler before Gi startup
```

These kernels use f64 arithmetic and checked buffer indices. Buffer arguments are declared explicitly; aliases share a linear-memory region. Writes completed before a trap are copied back without replay. Unsupported shapes, captured locals, nonintegral/out-of-bounds indices and integers outside the exact f64 range are rejected according to the upstream contract. Scalar compilation without options is unchanged. See [engine selection](../../internal/scripting/joker.md#wasm-execution-selection).

## Verification and profiles

- Native embedded Gi checks verify fill results, same-buffer aliasing, bounds failure, pre-trap mutation preservation and actual compiler/interpreter selection. Each explicit mode passes three race-enabled repetitions.
- Upstream `TestDenseNumericWASM` passes through Gi's Makefile, covering nested loops, safety/rebinding, aliasing and concurrency; its child-process coverage is functional evidence, not complete per-child profiling.
- Full Go regression under explicit compiler selection passes 2,177 tests across 35 packages; vet and whitespace checks pass.
- A native `CGO_ENABLED=0` build records Go 1.27.1 and Joker v42.12.4 in binary metadata.

The first new fixture used an unresolved catch type in Gi's wrapped namespace. The corrected fixture lets the tool surface the trap, then reads the same buffer in a separate native execution; it verifies preserved effects without replaying the trapped operation.

CPU, cumulative allocated bytes and objects were inspected. Race-enabled embedded scripting sampled 17.54 seconds compiler CPU and 19.43 seconds interpreter CPU; approximately 237.9/234.3 MiB and 6.38/6.36M objects were allocated. Native bridge preamble parsing/evaluation dominates; these mixed startup/bridge/kernel tests provide no throughput or old/new speedup claim. Existing compiler/runtime sharing and the bounded parsed-program cache remain in place. Raw profiles, matching test binaries, failed-probe logs and temporary native build outputs were disposed after analysis.

No whole-Joker WASI machinery, SDL/FFI activation, standalone-packaging tool, browser rerun or other-platform execution was added.
