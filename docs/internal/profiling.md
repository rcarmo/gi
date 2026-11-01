# Profiling gi

Set `GI_PPROF=127.0.0.1:6060` to serve Go's `net/http/pprof` from a running gi (TUI or `-web`). The server listens on its own address, never on the web UI's handler. It also enables mutex and block profiling.

```sh
GI_PPROF=127.0.0.1:6060 bin/gi
go tool pprof -top -cum bin/gi "http://127.0.0.1:6060/debug/pprof/profile?seconds=20"   # CPU
go tool pprof -top bin/gi http://127.0.0.1:6060/debug/pprof/allocs                       # allocations
curl -s "http://127.0.0.1:6060/debug/pprof/goroutine?debug=1" | head                     # goroutines
```

## Idle TUI CPU (2026-10-02)

An idle TUI on a long real session used about 21% CPU. Each forced frame lays out and re-wraps the whole transcript, and frames were being forced while nothing changed:

- **Cursor blink:** a 500 ms timer redrew the whole UI, although the editor's cursor is steady (as in Pi) and never used the blink state. The timer is removed.
- **1 s timer:** it re-rendered unconditionally. It now re-renders only during live activity (a running turn, compaction, the activity indicator) or when the footer data changed.
- **Block markers:** parsed on every render and every 80 ms check (base64 + JSON). Parsed markers are now memoized, and lines are filtered by prefix first.

After the change, the same session idles at about 3%. What remains is go-tui's 60 fps polling loop and gi's 80/120 ms timers.

Rule: a timer must call `MarkDirty` only when its state changed, because a frame costs a full transcript layout.

## Frame cost (#34)

These are measured with `make bench-tui-transcript-frame`: a 600-block session in a 100×40 window, one fullscreen frame of transcript layout and render into a reused buffer.

| Step | Time per frame | Allocations |
|---|---|---|
| All blocks laid out (before) | ~94 ms | 81 MB |
| Viewport window (spacers for blocks off screen) | ~3.9 ms | 3.3 MB |
| + go-tui wrap cache per element (layout, height and draw shared one wrap) | ~2.1 ms | 1.6 MB |
| + rendered block elements reused across frames (keyed by content hash) | ~1.8 ms | 1.3 MB |
| + block list and keys memoized while transcript lines are unchanged | ~0.86 ms | 0.49 MB |
| + ASCII fast paths (`RuneWidth`, `stringWidth`), direct narrow `Fill` | ~0.54 ms | 0.49 MB |
| + cached cluster segmentation per wrapped line; blocks indexed, not copied | ~0.29 ms | 41 KB |
| + heights kept apart from a bounded LRU of rendered blocks | **~0.25 ms** | **41 KB** |

Details:

- **Windowing** (`internal/tui/transcript_window.go`): only blocks that intersect the viewport, plus 8 rows of margin, are laid out. Spacers keep the content height, scroll offsets and stick-to-bottom identical. `TestTranscriptWindowMatchesFullLayout` checks that the rows match a full layout.
- **Block cache:** keyed by a hash of the block's content, its spacing context, the width and the theme. Every block's height is kept (up to 65,536, a few bytes each), so offscreen blocks are measured once. Rendered elements and their click targets are kept only for the 256 most recently shown blocks (an LRU), which covers the viewport and its margins; memory follows the screen, not the session's length (`TestTranscriptBlockCacheIsBounded`). Running blocks are never cached.
- **Block memo:** transcript lines are immutable strings, so pointer equality detects changes without hashing text.
- **go-tui** (`third_party/go-tui`): `textWrapCache` (`text_wrap_cache.go`) keeps an element's last wrap and its clusters, reset when the text changes. Callers must not modify the returned lines.

## Idle wake-ups

- **go-tui's `Run` loop** blocks while nothing is dirty and no events are queued. `MarkDirty` signals a wake channel, so a frame is never missed.
- **gi's 80 ms, 80 ms and 120 ms timers** are merged into one 80 ms tick (Pi's spinner cadence). The draft save runs every second tick.
- **The 1 s check** reads the session for the footer only after a topic event invalidated it, or every 5 s as a fallback.
- **The running-block answer** is reused while transcript lines are unchanged.

On the long real session, idle CPU went from about 21% to about 1%.

## Pre-release and diagnostic runs (scripts/testprofile)

Set `PROFILING=1` for pre-release verification or a targeted diagnostic run.
Ordinary development tests do not require profiling. Inspect CPU, `alloc_space`
and `alloc_objects`, investigate material outliers and remeasure fixes with the
same workload. The Makefile wrapper includes prerequisites, startup and cleanup.
`GI_TEST_PROFILE_ACTIVE` is an internal recursion marker.

- **`PROFILING=1 make test`**, with or without `TEST_RUN`: `testprofile go` runs each package with
  tests on its own (`go test -json -cpuprofile -memprofile`; one package at a
  time, as the Makefile throttling requires), then prints:
  - packages, tests, failures, wall and CPU time (build included), peak RSS
    (RUSAGE_CHILDREN: the largest process, often the linker; linking
    internal/web's test binary peaks near 1.3 GB, its tests near 100 MB)
    and the package that reached it, compared with the previous passing run
    of the same package set (▲ above +25%);
  - the slowest packages (wall incl. build · reported test time · CPU) and
    the slowest tests, with their change;
  - packages and tests that got markedly slower (test time +25% and +2 s per
    package, +50% and +1 s per test);
  - hot spots in the slowest packages: flat CPU and allocated bytes charged
    to the nearest gi function (`pprof -show=github.com/rcarmo/gi`);
  - disk used by the Go build cache, `GOTMPDIR` and the kept profiles.
  Profiled runs are not cached by `go test`, so every test executes. Test
  binaries are deleted after profiling (the profiles carry their symbols).
  Focused/repeated/benchmark runs have separate baselines keyed by package set, filter and flags. Specialised Go targets use `testprofile gotest` to preserve their race, count and benchmark options while collecting CPU/allocation profiles.
- **Script suites and fixtures**: `testprofile run -name …` measures wall time, CPU time and peak RSS of the complete lifecycle and compares them with the previous passing run. `make fixtures-vibes-report` regenerates only the report against saved results under the same profiling wrapper. Script timing/RSS is not allocation profiling; investigate runtime allocations through `GI_PPROF` or browser allocation tooling when needed. After each run, inspect the printed analysis and saved report before publication.

### Fixtures-vibes worker profiles

The disposable `fixtures_vibes` build writes per-worker `cpu.pprof` and cumulative `mem.pprof` files when `GI_FIXTURE_PROFILE_DIR` is set. The Makefile passes `FIXTURES_RUNTIME_PROFILE_DIR` (default `test-results/fixtures-runtime-profile`) to both full and focused targets. Collection starts during fixture model registration and ends after graceful runtime teardown. During pre-release or diagnostic profiling, inspect CPU plus `alloc_space` and `alloc_objects`; a killed worker without a heap profile fails the profiling check even if its browser tests passed. `make test-fixtures-profile-lifecycle` verifies startup, SIGTERM shutdown and retained profile analysis before a long run.

On 2026-10-05 this check completed in 2.7 seconds including build, with 10 ms sampled CPU, 11.9 MiB sampled allocations and about 70,500 objects. Joker native-variable metadata led object counts (46%); main startup and model/config setup led allocation bytes. This short startup check is separate from compliance workload measurements.

Scratch lives beneath the resolved project temporary root, in
`runs/profiling/<worktree>`. Raw profiles and matching test binaries are disposed
of after automatic analysis by default. Use `PROFILE_KEEP=1` only while manual
analysis needs the captures; delete them and disposable logs immediately after
comparison. Keep concise measurements and limitations, without raw archives.

## Findings acted on (#38)

- New stores copy a migrated schema template instead of running
  `initSchema` (`internal/store/schema_template.go`): an empty in-memory
  store is restored from it, and a database file not created yet is written
  from it beside the target and linked into place (an existing file, or one
  another process created first, is migrated as before).
  `TestSchemaTemplateMatchesInitSchema` compares the schema, migration
  records and `user_version` with `initSchema`'s. `initSchema` had been
  28–41% of CPU in internal/turn, internal/store and internal/web.
- `validateWorkspaceSchema` normalised the whole migration text once per
  schema object; it now does so once.
- Tailscale (`tsnet`, some 200 packages) is linked only into gi:
  `internal/peering` starts it through `peering.StartBackend`, which
  `internal/peering/tsnetbackend` sets and only `cmd/gi` imports. Test
  binaries no longer link it.
- The 1.3 GB peak RSS was the linker building internal/web's test binary,
  not its tests (about 100 MB); without Tailscale the peak is about 740 MB.
- Result (warm cache): 1962 tests in 58 s wall and 30 s CPU, from 2m37s and
  3m6s.

## Tool-checkpoint recovery (2026-11-01)

`PROFILE_KEEP=1 PROFILING=1 make test-tool-batch RACE=` passed 48 reported tests
(`-count=3`) in `internal/store` and `internal/turn`, using Go 1.27.3.
CPU, `alloc_space` and `alloc_objects` were inspected for both packages.
The seven hard-kill cases exercise recovery, usage and interrupted projections;
profiles cover the surviving test processes, with stubbed model calls. Killed
children cannot flush final profiles.

The model-context assertions repeatedly converted the same JSON buffer to a
string. Reusing one string per request removed that fixture allocation without
changing its assertions:

| Turn-package measurement | Before | After |
|---|---:|---:|
| Sampled allocated space | 67.13 MiB | 57.65 MiB (-14.1%) |
| Assertion callback, flat allocated space | 6.81 MiB | 0 sampled |
| Sampled allocated objects | 465,419 | 469,674 (+0.9%) |
| Sampled CPU | 580 ms | 450 ms |
| Build-inclusive peak RSS | 683 MiB | 695 MiB (+1.8%) |

These short sampled runs do not establish a latency or production-memory
improvement. SQLite/syscalls led CPU; test regexp compilation led allocation
space. `AdmitTurnEffect`'s sampled 32,768 boxed SQL arguments accounted for
0.5 MiB (1.8% of store-package allocations); its ownership and transaction
checks were preserved. Sampled object counts and short CPU samples vary;
build-inclusive RSS includes compilation. Raw captures and disposable logs were
deleted after analysis. Subsequent full Go tests, focused tool and ownership
race checks, vet, Bun checks and 204 web-adapter tests passed.

## Post-repair auth reads (2026-11-01)

The baseline was main `53e273dc`, Go 1.27.3, two CPUs, with sequential
`PROFILING=1 PROFILE_KEEP=1 make test RACE=`. All 2,955 reported tests in
35 packages passed: 110 s package wall time, 60.0 s child CPU and 795 MB
build-inclusive peak RSS. CPU, `alloc_space` and `alloc_objects` were inspected.

The largest inspected costs were:

- SQLite preparation: 2.25 s (32.9%) of turn-package CPU; parser work was
  1.15 s (33.1%) in store. `Store.Open` accounted for 1.28 s in store and
  0.74 s in turn. These tests repeatedly create independent databases and
  execute transactional state transitions. Schema-template reuse already
  applies. Statement caching is a candidate for a separate fixed-workload
  experiment, not a reason to remove transaction, cancellation or ownership
  checks. No statement-cache benefit was measured in this pass.
- Turn allocations: 735.12 MiB. The large-image boundary fixture accounted
  for 173.37 MiB cumulatively; Wasm memory initialization/growth also appeared.
  These totals include deliberately oversized image inputs and required
  isolation. They do not establish a production image-memory regression.
- TUI allocations: 703.53 MiB, including 199.16 MiB in `go-tui.NewBuffer`,
  103.67 MiB in styled-rune segmentation and 70.18 MiB directly in
  `transcriptRowsAtWidth`. The nested go-tui frames were inspected; this pass
  makes no renderer change or claim about steady-state frame performance.
- Scripting: 2.20 s CPU, 258.97 MiB and 5.10 million sampled objects.
  Joker evaluation beneath `runAndCacheEmbeddedJokerProgram` accounted for
  about 165.66 MiB and 4.69 million objects. That path both parses and
  evaluates on cache misses; it is not solely parsing overhead. The existing
  bounded parsed-program cache and per-call bridge setup remain unchanged.
  Reusing mutable evaluation state without an isolation benchmark is not an
  accepted optimization. Warm/cold scripting and steady-state rendering
  remain separate diagnostic workloads, not verified production improvements.
- Auth: 177.77 MiB, including 68.69 MiB in `io.ReadAll` buffer growth.
  `readState` already knows the opened file's size. It now preallocates a
  bounded buffer with EOF-probe slack, while retaining the read limit,
  identity checks, JSON decoding and unknown-field preservation. Size is only
  a hint: reads are neither truncated to the stat size nor cached across calls.

The concurrent-login test has a timing-dependent reader count, so its aggregate
allocation totals cannot prove a per-read improvement. `make bench-auth-state`
uses fixed synthetic state, a 2 KiB unknown extension, 5,000 reads per case,
three repetitions, and the same profiling options before and after the change.
Medians (sample size three; no statistical significance claim):

| Sessions | ns/read before → after | B/read before → after | allocs/read |
|---|---:|---:|---:|
| 1 | 13,492 → 11,860 (-12.1%) | 9,610 → 7,493 (-22.0%) | 32 → 27 |
| 32 | 59,079 → 54,061 (-8.5%) | 53,595 → 40,659 (-24.1%) | 74 → 64 |
| 128 | 191,631 → 180,374 (-5.9%) | 195,354 → 147,307 (-24.6%) | 176 → 162 |

Across the equivalent benchmark runs, sampled CPU was 4.42 → 4.02 s (-9.0%),
allocated space 3,763.77 → 2,808.97 MiB (-25.4%), and objects
4,222,090 → 3,667,060 (-13.1%). Sampling contributes noise. Build-inclusive
peak RSS rose 206 → 209 MB (+1.5%); that small secondary change does not
outweigh the repeatable per-read allocation reduction. This is a bounded
read-path improvement, not a measured end-to-end login latency improvement.

Verification: `make test-auth-state` passed the complete auth package ten times
with the race detector; fresh reads, the exact 1 MiB boundary, oversized,
truncated and empty state were added to the regression coverage. Existing
symlink, corruption, cross-process, revoke and crash-lock tests remain intact.
A second profiled full Go run passed 2,962 reported tests in 35 packages, and
`make vet RACE=` passed. Inherited `PI_CODING_AGENT_DIR` and
`GI_CODING_AGENT_DIR` were cleared for verification. The second suite includes
new tests and variable concurrent-reader work, so its aggregate changes are
not an equivalent-workload performance claim. Raw captures, disposable logs
and owned scratch were removed after analysis. No release or deployment.

## Joker bridge setup: cold and steady state (2026-11-01)

On main `225ab4de`, the earlier scripting hotspot was reproduced with a fixed
bridge: 32 messages of 512 characters, a counter of 7, and a script returning
counter plus message count (39). Go 1.27.3, two CPUs, sequential Make targets:

- `PROFILING=1 PROFILE_KEEP=1 make bench-joker-first-call RACE=` runs one call
  in each of three fresh test processes. Repeating within one process would
  not measure first-use library setup.
- `PROFILING=1 PROFILE_KEEP=1 make bench-joker-steady RACE=` warms libraries
  and the script, then measures 500 calls per case, three repetitions. The
  hit case retains bridge metadata; the miss case changes only the session
  ID to exercise the existing full-script cache key. The script and result
  are unchanged. Profiling totals include calibration/priming; per-operation
  benchmark counters exclude setup.

This was not predominantly cold initialization. Before tuning, warmed hits
still took about 98 ms and allocated 12 MiB each. CPU profiles charged
251.32 s (71.2% of 353.13 s) cumulatively to go-joker's `findFnVarName`, whose
fallback scans namespace mappings during interpreted calls. A temporary
expression-labelled probe attributed 5.23 of 5.34 labelled CPU seconds to the
four plain `ns` forms. They repeatedly interpreted the core-Var referral loop.
Interpreted recursive bridge-key conversion added further avoidable work.
The diagnostic labels were removed, not shipped.

Two bounded changes avoid that repeated interpretation without modifying the
dependency or retaining mutable state across executions:

- Bridge namespace entry uses `EnsureSymbolIsNamespace`, `ReferAll` and
  `SetCurrentNamespace` on every call. It preserves the core-Var refresh,
  retains local/private bindings and updates `*loaded-libs*` through persistent
  `Conj`, not a mutation of an old snapshot. Entry is restricted to `user`,
  `gi.state`, `gi.topics` and `gi.runtime`.
- A native walker transforms the existing JSON reader's objects. Scalar
  objects, including numeric representations and nil, are unchanged. Maps
  use persistent `Assoc`, preserving the old promotion to HashMap for wide
  objects. The JSON reader's `:keywords?` shortcut was investigated but not
  retained: values alone were insufficient evidence because concrete map
  representation matters too. Nil must be handled before Map: Joker's Nil
  implements that interface. The original `walk` alias remains available.

Medians of the same three-sample workloads (no significance claim):

| Case | Time/call before → after | B/call before → after | allocs/call before → after |
|---|---:|---:|---:|
| Fresh first call | 102.56 → 11.10 ms (-89.2%) | 14,167,552 → 2,271,880 (-84.0%) | 367,985 → 44,083 |
| Warm parsed hit | 97.52 → 1.412 ms (-98.6%) | 12,228,897 → 581,605 (-95.2%) | 324,836 → 7,161 |
| Warm parsed miss | 120.89 → 12.77 ms (-89.4%) | 13,724,271 → 1,926,579 (-86.0%) | 366,070 → 42,631 |

For the equivalent steady benchmark processes, sampled CPU was
353.13 → 23.43 s (-93.4%), allocated space 37,413.54 → 3,657.97 MiB (-90.2%),
and sampled objects 1,041,413,131 → 76,143,725 (-92.7%). Build-inclusive peak
RSS was 496 → 497 MB (+0.2%); it is not isolated runtime memory. Fresh-call
CPU profiles summed only 30 ms after tuning, too sparse for fine attribution;
the first-call timing/allocation comparison comes from fixed-count benchmark
counters, not those few samples.

Residual misses spend most time parsing and evaluating macro-expanded preamble
forms (16.59 s cumulative in `runAndCacheEmbeddedJokerProgram` in the final
steady profile). The existing cache remains bounded at 128 programs. A
metadata-independent parsed template would need its own freshness, callback,
context and namespace-isolation tests before adoption; no benefit from such a
change is claimed here. This pass fixes measured bridge overhead, not arbitrary
user-program performance, renderer costs or SQLite preparation.

Verification passed: 2,967 reported tests across all 35 Go packages under
profiling; selected-runtime race checks across scripting/tools/config/inference
at count 3; both compiler and interpreter Joker/Wasm race checks at count 3;
and vet. Differential tests compare the native walker with the previous
interpreted reference, including recursive types, wide maps, nested nil,
numbers and unusual keys. Namespace mappings are compared with plain `ns`,
including refresh after removing a core reference. Persistent loaded-library
snapshots, changing bridge state/session IDs, and invalid namespace entries
are also checked. Inherited PI/GI agent directories were cleared. Full-suite
profiles were inspected, but added reference tests make suite-wide totals an
invalid matched performance comparison. Captures, diagnostic/probe artifacts,
matching generated helpers, disposable logs and owned scratch were removed
after analysis. No release or deployment.
