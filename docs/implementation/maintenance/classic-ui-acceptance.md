# Classic UI acceptance, 2026-10-05

Gi consumes the shared frontend at fixtures `4259e825964d2b60dd6c6f7242e2b1daf4280a7d` through `ui/classic`. Handoff `5ab99156` was merged without rebasing; macOS PR53/54 fixes were retained. Gi's frontend imports, Makefile targets, embed assets and tests use the canonical path. The frontend owner confirmed this handoff; the future v0.2.0 suite repin is separate.

## Results

| Check | Result | Profile |
|---|---|---|
| Plan/widget fixtures, including extra014 | 66/66 across six Chromium/WebKit projects | 290.8s wall, 299.3s CPU, 1,358MB peak RSS, including builds/browser processes |
| Frontend units | 189/189 | Adapter lifecycle 0.34s wall, 0.38s CPU, 133MB peak RSS |
| Gi functional browser/API suite | 142 passed, 11 skipped | 171.9s wall, 83.0s CPU, 403MB peak RSS |
| Static CORS/auth, Plan/widget handler checks | 15/15 | CPU/allocation profiles under `~/.cache/gi-test-profile` |
| Full Go suite | 2,122 passed, zero failures, 37 packages | 90.2s Go wall, 80.6s CPU, 748MB peak RSS including build |
| Vet and hook TDZ checks | Passed | Static analysis, no behavioural test run |

The focused shared run used `make fixtures-vibes-focused FIXTURES_SPEC_ARGS='suite/specs/plan.spec.ts suite/specs/widgets.spec.ts'`. It did not overwrite the archived full4 raw report in `/workspace/tmp/fv-full4-archive`. No full shared browser gate was rerun at this pin. The profile now claims `@cap-plan-sidebar` and `@cap-widgets`; eight Plan skips are removed. Forty-two issue-backed core entries and two Windows host skips remain.

The first functional run failed only because the KaTeX check read Gi's root test dependency (0.18.10) instead of the shared frontend's installed package (0.19.0). The test now reads the frontend package and retains the local CSS/font/renderer checks. The focused check and full functional rerun pass.

## Security and allocation findings

The shared frontend's anchored build patches remove `allow-same-origin` and require the iframe source for every bridge message. Gi grants CORS only to existing public embedded JavaScript on GET/HEAD, including precompressed variants. Authenticated routes grant no CORS; cross-session artifact lookup remains authenticated and isolated.

The static-handler profile showed whole compressed assets copied into fresh buffers per request. Serving seekable embedded files directly reduced sampled allocations attributed to the handler from 12,172.7KiB to 7,661.1KiB in the same 15-test group (about 37%). The fallback still reads non-seekable files. Test time fell from 0.169s to 0.109s; these short sampled runs are diagnostic, not a stable throughput benchmark. The second run cleared an oversized build cache, so its 66s build wall time cannot measure runtime performance.

The full Go run was slower than the earlier 2,109-test run (wall +33%, CPU +115%) with rebuilding, additional tests and changed binaries. No individual package/test crossed the profiler's regression thresholds. The largest allocation hot spots remain transcript rendering (~187MB) and effective-thinking configuration (~159MB across turn tests); these are follow-up optimisation candidates, not changes made in this integration.

Script-suite profiles measure CPU, wall time and peak RSS, not allocation counts. Runtime allocations require Go pprof or browser allocation tools. Focused filters and full suites have separate comparison keys; the first full functional comparison after removing legacy nested timing includes an incompatible earlier baseline and is not a performance regression measurement.

## Local logs

* `/workspace/tmp/gi-classic-plan-widgets-profiled.log`
* `/workspace/tmp/gi-classic-adapters-profiled.log`
* `/workspace/tmp/gi-classic-functional-verified.log`
* `/workspace/tmp/gi-classic-static-optimized.log`
* `/workspace/tmp/gi-classic-go-profiled.log`
* `/workspace/tmp/gi-classic-final-vet.log`
