# Native Joker v42.12.2

Gi pins `github.com/rcarmo/go-joker/v42 v42.12.2`, tag commit `d277f315659c55b468b525e81a4a2b3cf83b1946`, released on 2026-10-06. Go module provenance and sums were checked. go-ai stays at v1.0.4, the frontend pin stays unchanged, and the historical third_party Joker reference data is untouched.

## Emitted-function execution

Joker stays native inside Gi. The release adds `JOKER_WASM_ENGINE=auto|interpreter|compiler` (`native` aliases compiler) and `(jit/wasm-engine)` to report actual execution mode. Configure the environment before starting Gi; the shared Joker WASM runtime initialises once. `compiler` errors if native compilation is unavailable, while default `auto` may use the interpreter on unsupported/restricted hosts. Numeric eligibility, overflow and type contracts are unchanged.

Gi's new numeric-loop test explicitly calls `jit/compile-wasm`, checks the independent sum `4950` for integers 0–99, and checks script/Go engine-report agreement. Existing kernel and unsupported-shape tests also assert the requested engine. Fresh native test processes run compiler and interpreter modes separately, with three race-enabled repetitions in each mode. This covers generated functions, with no whole-Joker WASM/WASI guest.

The upstream release also provides `joker compile --native --run` standalone packaging. It bundles the native runtime and source, then compiles eligible emitted WASM functions at execution time. Gi does not add a standalone-packaging tool or substitute that workflow for its native embedded bridge. See [startup selection](joker.md#wasm-execution-selection).

## Verification

On Linux amd64 / Go 1.27.1:

- `make test-joker-wasm-engines`: kernel, unsupported-shape recovery and loop/actual-engine tests pass three times under race detection for each explicit mode.
- Nested upstream `core/wasm` tests pass three times: same independent module result, actual compiler/interpreter selection, native alias and invalid-mode validation. Upstream `std/jit` tests pass, including native subprocess loop/engine checks. These tests use the selected release without changing the module cache.
- Full Gi regression under explicit `JOKER_WASM_ENGINE=compiler`: 2,173 tests across 35 packages, no failures. Vet and whitespace checks pass.
- Native `CGO_ENABLED=0` Gi build succeeds; binary metadata records Joker v42.12.2, go-ai v1.0.4 and Go 1.27.1.

No browser/full-fixture rerun, live provider test, standalone CLI packaging acceptance or other-platform native execution was performed. Explicit compiler mode was verified on this supported host; unsupported-architecture behaviour follows upstream's architecture checks and was not executed locally.

## Profile findings and disposal

CPU, cumulative allocation bytes and objects were inspected for compiler/interpreter repetitions and the full regression. Race-enabled kernel workloads sampled about 9.3 seconds compiler CPU and 9.5 seconds interpreter CPU; total allocations were about 138.5 MiB/3.70M objects and 128.5 MiB/3.60M objects. These workloads include repeated bridge preamble parsing/evaluation and startup. Core evaluation dominates; they are functional profiling workloads, not a kernel throughput comparison or an upgrade speedup measurement. Native compilation remains process-shared; Gi does not create a new Joker WASM runtime per script.

The short nested upstream captures contain zero sampled CPU, including parent-only captures for subprocess work. Their allocation profiles mostly contain setup. They provide functional tests, not complete per-subprocess CPU profiling evidence. The longer in-process Gi native-host workloads supply sampled CPU/allocation coverage.

Existing costs outside this dependency change include Joker bridge preamble evaluation, QuickJS fresh-instance setup and transcript projection when invalidated. No correctness, numerical, cancellation or security bound was weakened to reduce them. Raw profiles, matching test binaries, temporary native build probes and disposable logs were removed immediately after analysis; only these concise conclusions remain.
