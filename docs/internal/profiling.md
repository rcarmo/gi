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

## Every test run (scripts/testprofile)

Every test run is profiled and analysed (AGENTS.md), including focused runs and fixture report checks. The Makefile dispatches test goals through a profiled recursive make, so prerequisites, server startup and cleanup are measured too. `GI_TEST_PROFILE_ACTIVE` is an internal recursion marker, not a user override. `TEST_PROFILE=0` no longer disables profiling.

- **`make test`**, with or without `TEST_RUN`: `testprofile go` runs each package with
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

The disposable `fixtures_vibes` build writes per-worker `cpu.pprof` and cumulative `mem.pprof` files when `GI_FIXTURE_PROFILE_DIR` is set. The Makefile passes `FIXTURES_RUNTIME_PROFILE_DIR` (default `test-results/fixtures-runtime-profile`) to both full and focused targets. Collection starts during fixture model registration and ends after graceful runtime teardown. Inspect CPU plus `alloc_space` and `alloc_objects` after every run; a killed worker without a heap profile fails the profiling check even if its browser tests passed. `make test-fixtures-profile-lifecycle` verifies startup, SIGTERM shutdown and retained profile analysis before a long run.

On 2026-10-05 this check completed in 2.7 seconds including build, with 10 ms sampled CPU, 11.9 MiB sampled allocations and about 70,500 objects. Joker native-variable metadata led object counts (46%); main startup and model/config setup led allocation bytes. This short startup check is separate from compliance workload measurements.

Everything is kept in `~/.cache/gi-test-profile` (`GI_TEST_PROFILE_DIR`,
`TEST_PROFILE_DIR`): `history.jsonl` (totals of every run), `latest-<suite>.json`
(baselines) and the last five Go runs (`go-<time>/report.json` plus each
package's `cpu.pprof` and `mem.pprof`, readable with `go tool pprof`).

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
