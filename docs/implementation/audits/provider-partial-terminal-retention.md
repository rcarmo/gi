# Provider partial and terminal retention (#87)

Verified on 2026-11-01 against main `810f9be0` plus this repair, using Go 1.27.3. PR #102 supplies bounded partial checkpoints and failed receipts; PR #113 supplies the engine-wide close/join barrier. The final audit found that an empty structured terminal response still replaced the committed prefix and discarded its received usage.

## Receipt contract

A failed attempt keeps one timeline row. Its `provider_message` contains the provider's terminal response; `provider_partial` retains the bounded, detached committed prefix. Both stay available in raw history/export and are removed from conversation messages and search. The existing best-effort encoding path handles unencodable terminal diagnostics.

History text comes from the structured snapshots, never the mutable UI row. A missing or shorter terminal text preserves the committed prefix. A longer terminal text uses that extension; differing fragments remain visible together. Thinking blocks fall back to the checkpoint when absent from the terminal response. Neither failed contribution is adopted into model context or authorises a tool call.

Valid terminal usage takes precedence. Missing, all-zero or invalid terminal usage falls back to valid committed received usage, counted once and marked uncertain. A final bill cannot be inferred from a failed stream.

## Acceptance matrix

- `TestProviderPartialEmptyTerminalPreservesCommittedPrefixOnReopen`: error, user abort, final usage, zero/invalid usage, differing, shorter and longer terminal text; disk reopen, one row, diagnostics, signatures, tool prefix, thinking, accounting and context exclusion. Tampered UI text cannot replace the provider prefix.
- Existing partial tests: detached snapshots, bounded/coalesced writes, sealing, stale-owner fences, failed-transaction rollback, non-assistant rejection and unencodable diagnostics.
- Killed-owner subprocess test: process death precedes SQLite reopen and conversion of the committed partial; no provider or tool replay.
- Producer/engine tests: channel closure precedes terminal return; `Close` joins cancellation-ignoring providers, tools, hooks and child work. Close/retry barriers retain the shared-Store fencing tests from PR #113.
- Go web tests and Chromium/WebKit functional tests: retained text survives reload; conversation messages/search hide raw partial signatures and tool arguments; raw export retains them.

## Verification and profiling

All commands ran sequentially through the Makefile:

- `make test-provider-partials PROFILING=1 PROFILE_KEEP=1`: race detector, three repetitions, 31 top-level tests across store/turn/inference/web; passed.
- `make test RACE= PROFILING=1 PROFILE_KEEP=1`: 35 packages, 2,728 tests; passed.
- `make test TEST_PKGS=./internal/turn TEST_RUN='TestClose|TestShutdown|TestProviderRetryClose|TestSidePrompt|TestAfterTurn'`: race detector; passed.
- `make vet RACE=`, `make build-web RACE=`, `make bun-checks RACE=`: passed.
- `make test-ux RACE= PROFILING=1`: 149 passed, 13 skipped. The first run exposed a missing `created_at` in the new SQLite browser fixture; the corrected fresh-instance run passed.

Inspected CPU, `alloc_space` and `alloc_objects` for the four focused packages and the full Go run. The focused race run is dominated by instrumentation and regexp test selection; inference CPU has only 40 ms of samples. Use the non-race full run for application attribution: turn CPU sampled 5.98 s, allocated space 686.02 MiB and allocated objects 3,114,533. Failure settlement accounts cumulatively for 0.40 s (6.69%), 5.50 MiB (0.80%) and 67,886 objects (2.18%); reading the committed partial accounts for 0.01 s CPU. The largest allocation paths are existing image-bound test fixtures (173.37 MiB cumulative), WASM memory growth and SQLite rows; these do not identify a material repair-specific outlier. No performance improvement is claimed, and the wrapper's older baselines cover different workloads.

Browser profiles contain Playwright/Node harness work, mostly idle time, filesystem reads and snapshot bookkeeping; they do not measure Go allocations. Go allocation history comes from the package profiles above. Raw captures, matching test binaries, temporary browser fixtures and disposable logs were removed after analysis. No release, deployment or live-provider billing validation was performed.

General generation/tool checkpoint continuation remains in #86/#92. It is independent of retaining failed output and joining its producers.
