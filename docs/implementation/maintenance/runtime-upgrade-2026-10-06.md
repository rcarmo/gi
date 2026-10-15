# Runtime upgrade and selection

Gi uses `go-ai v1.0.4` and `go-joker/v42 v42.12.1` under Go 1.27.1. The released module provenance is:

- go-ai: tag `v1.0.4`, runtime commit `65e25029c969d154560d0c26eec032d2605d80da`.
- go-joker: latest stable release checked on 2026-10-06, tag `v42.12.1`, commit `6942169c8118927c02c87b8efd10110d6eb9eef5`. The release uses the `/v42` module path; all active Gi imports migrated from the old unsuffixed pseudo-version.

The historical `third_party/joker` tree is a reference-search dataset, not the active module dependency. It and unrelated policy edits were preserved. The frontend pin did not change.

## Native Joker; QuickJS-only runtime isolates

Joker executes natively inside Gi. `joker.jit/compile-wasm` is registered and preloaded so agent-created numeric kernels can emit WASM for execution through Joker's wazero machine-code compiler on supported hosts. Explicit unsupported compilation fails. The test verifies a compiled integer kernel returns 42, unsupported string-building code is rejected and native evaluation still works afterwards. Ordinary Joker scripts keep the native bridge.

The old whole-interpreter WASI test, guest, bootstrap overlay and Makefile targets were removed at Rui's direction. Its earlier execution does not qualify the native user-function compilation path. QuickJS is the only whole runtime in a wazero isolate, both for internal JavaScript and existing MCP codemode. Goja remains the native JS choice. Internal QuickJS scripts await bridge functions; MCP codemode retains its separate tool-only capabilities.

`javascriptRuntime` chooses the default JS engine (`goja` or `quickjs`), with project settings over user settings. An explicit `engine` overrides it; `js` follows the configured default, and `.joke`/`.clj` select native Joker automatically. Unknown runtime names fail without fallback. See [selection and examples](../../internal/scripting/README.md) and the [bridge contract](../../internal/scripting/contract.md).

## Verification

- Native `CGO_ENABLED=0` Gi build succeeds; `go version -m` reports Go 1.27.1 and the two exact module versions.
- Full Go regression: 2,172 tests across 35 packages, zero failures; vet and whitespace checks pass. Two retired feasibility packages account for the package-count reduction.
- Four nested upstream Joker core tests pass, covering retained user-function WASM bytes and native integer result/range contracts. An initial probe filter selected zero tests; it was corrected and is excluded from execution evidence.
- Sixteen selected runtime tests repeated three times with the race detector pass. These cover native embedded Joker and compiled kernels, QuickJS bridge/console/result semantics, file/payload routing, explicit/default selection, settings precedence, isolate separation, cancellation joined with started host calls, unsupported-engine rejection, and provider catalogue filtering. MCP codemode regressions pass in the full suite.
- An old catalogue fixture reused its registered ID across repetitions; unique IDs now preserve the same first-registration/duplicate/filter assertions. The host-cancellation test now synchronises on a started call instead of expiring during cold QuickJS compilation. A float invocation of an integer-specialised Joker kernel failed; tests use the integer signature, without widening eligibility or claiming float equivalence.

Live providers, platform-wide release builds and general script-to-WASM eligibility are not verified here. The JavaScript engines are explicit alternatives, not syntax-identical substitutes: QuickJS bridge functions are asynchronous. Selecting an isolate does not revoke operator-granted bridge privileges.

## Profiles and disposal

CPU, cumulative `alloc_space` and `alloc_objects` were inspected for the final repeated runtime gate and full regression. Repeated race workload: about 54 seconds wall and 34.7 seconds sampled scripting CPU; fresh Joker bridge preamble parsing/evaluation dominates (~394.8 MiB cumulative core evaluation allocations, 11.6M objects). QuickJS module compilation is one-time and shared through `codemode.Default`; valid executions use fresh instances. The bridge adapter snapshots host values once and serialises marshalled host calls rather than re-compiling QuickJS or executing user code in Goja. Full-suite wall time is about 84 seconds with approximately 61 seconds process-tree CPU; variable OAuth/web fixture waits affect suite timings. No equivalent old/new throughput improvement is asserted.

The implementation keeps the existing parsed-program bound (128 entries) and QuickJS deadline/output/memory bounds. Cancellation waits for host calls to return; host callbacks must honour context. The four-test upstream core CPU capture contains zero samples because execution lasts only milliseconds; it supplies functional execution evidence, not a passing CPU-hotspot qualification. Its heap captures mostly contain runtime/string initialisation. The much longer Gi native-Joker/race workloads supply sampled runtime CPU/allocation evidence. Raw profiles, probe logs, temporary Joker module copies/overlays, matching test binaries and temporary build probes were deleted after analysis. Only these concise results remain.
