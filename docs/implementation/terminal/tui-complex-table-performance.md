# Complex Markdown table rendering performance

## Scope and measured cause

Profiled native TUI Markdown projection and transcript layout/drawing with 40/200
rows, six columns, nested inline styles, code, escaped pipes, CJK, Arabic,
decomposed accents, VS15/VS16, flags and ZWJ/modifier emoji. Baseline: main
`78a35073`. About 95% of the initial mixed-workload CPU sample was spent measuring
rich-text widths through `IntrinsicSize` and the Unicode cluster scanner.
Block-level windowing did not help a *single* long table: its visible message
still contains many layout children, and layout repeatedly measures them.

This change:

- Measures intrinsic text width once per text assignment. Padding, border and
  explicit dimensions are not cached, so their changes remain visible.
- Reuses unwrapped rich-text clusters. Table rows are intentionally not
  word-wrapped by go-tui; they formerly bypassed the existing wrap cache.
  The cache invalidates for a changed base style (including message background),
  text assignment or a different wrap mode/width. It remains element-owned and
  single-width, not an unbounded global string cache.
- Avoids Unicode category checks for impossible low-code-point extenders;
  uses the equivalent combined `M` mark category, explicit join controls and
  emoji modifiers. Wide-range lookup is binary search; the ranges are now sorted.
  Zero-width classification keeps its *different* format-control semantics.
- Parses a table-bearing chat message once and projects the AST once rather than
  discarding an initial prefixed projection and reparsing twice more.
- Preallocates table glyphs, computes styles per segment rather than per glyph,
  and reuses a table's row divider. Stored/model-visible source is unchanged.

The local go-tui snapshot/replacement is already used for Pi-style row output.
`third_party/go-tui/README.gi.md` and `gi-unicode-perf.patch` record the additional
patch provenance. No module-cache files or Piclaw components were modified.

## Measurements

Linux ARM64, Go 1.27.1, Makefile defaults: CPU_SET=0-1, CPU_PROCS=2,
CPU_NICE=10; two 300 ms benchmark runs. Rounded means, not latency percentiles:

| Fixture / operation | Before | After | Improvement |
|---|---:|---:|---:|
| 200 rows, cached scroll + layout/draw | 213.0 ms | 1.86 ms | ~115× |
| 200 rows, cold layout/draw | 243.5 ms | 17.5 ms | ~14× |
| 200 rows, alternating 60/120-cell widths | 230.0 ms | 2.89 ms | ~80× |
| 200 rows, projection at 120 cells | 26.8 ms | 9.81 ms | ~2.7× |
| 40 rows, cached scroll + layout/draw | 44.3 ms | 0.583 ms | ~76× |

The 200-row scroll benchmark allocates about 232 KB / 202 allocations per frame,
down from 985 KB / 4,004. At 120 cells, projection allocates about 5.1 MB instead
of 17.25 MB. This is measured renderer work; ANSI serialization, terminal painting,
provider/storage traffic and real UI event-to-paint latency are **not included**.

A separate 500 ms width microbenchmark uses the frozen original implementation:
mixed Unicode is ~2.6× faster (9.86 µs → 3.86 µs), box-drawing borders ~2.2×
(14.7 µs → 6.66 µs). Both allocate zero bytes. Pure ASCII remains approximately
unchanged (~90–94 ns), retaining its existing fast path.

Eight increasing streaming snapshots ending at 200 rows take ~127 ms *in total*
and allocate ~62.5 MB across all eight projections and first layouts. No baseline
speedup is claimed for that added benchmark. Streaming still reparses changed
source; layout still traverses all children of an intersecting large message.
The final targeted warm-scroll profile is dominated by layout traversal, not
repeated Unicode segmentation. Row-level virtualization or incremental parsing
would be separate architectural changes, not justified as part of this patch.

Artifacts are under `test-results/tui-table-perf/`. Reproduce through Make:

```sh
make bench-tui-complex-tables BENCH_ARGS=-count=2
make profile-tui-complex-tables TABLE_BENCH='BenchmarkComplexMarkdownTable/rows=200/scroll' TABLE_BENCH_TIME=2s
make profile-tui-complex-tables-report
make bench-unicode-perf TABLE_BENCH_TIME=500ms BENCH_ARGS=-count=2
```

Use `GI_GOCACHE` for an isolated **on-disk** cache when other sessions build.
An initial full test run lost cache entries to concurrent cache trimming; the
rerun with a private cache passed. Bun builds/harnesses reuse the installed
packages; no npm or new dependencies are required.

## User input boundaries and Pi presentation

Compared installed Pi 1.0.0's `UserMessageComponent`: Markdown uses
`userMessageText`, a distinct `userMessageBg`, outputPad (default 1), and vertical
padding 1. Gi already has the corresponding light/dark full-width user band;
assistant output is unbanded and tool calls use their own status backgrounds.

A table-source metadata branch incorrectly set the transcript loop index to
`j` rather than `j-1`. The loop then incremented again and skipped the following
message's first row. That dropped its user header, broke the coloured band, and
also could skip a subsequent tool or assistant boundary. The corrected parser
consumes exactly its own rows. It does not invent speaker prefixes inside tables,
change user text, or paint assistant/tool cells with the user background.

## Acceptance

- `make test-unicode-perf`: every code point plus invalid boundary runes, 4,000
  deterministic valid/malformed randomized strings, byte/string cluster agreement,
  text setters/options, width/padding/border changes and wrap/base-style invalidation.
  This preserves go-tui's existing terminal profile, **not full UAX #29** support.
- `make test-tui-complex-tables`: projection equivalence, styled/Unicode widths,
  table→user→tool→user-table→assistant boundaries, both themes, and full-buffer
  cell equality against unwindowed rendering (including backgrounds and links).
  The boundary regression fails in all six theme/width cases on the old code.
- `make test-tui-complex-tables-pty`: 148 actual tmux screen comparisons, both
  themes at 60/120 cells, scrolling both ways across table/user/tool boundaries.
- `make test-tui-table-scroll`: all 516 existing four-width terminal frames pass.
- `make test-tui-tables`: Pi oracle comparison and six live native SSE/store/resize/
  reopen scenarios in regular/fullscreen modes pass.
- `make test-tui-selection`: three native selection/copy scenarios pass.
- `make test-tui-tool-syntax-pty`: six highlighting/resize/reopen scenarios pass.
- Full Go suite, dependency tests, vet, web build and hook checks pass. Browser
  suites are skipped per the ChromeOS host guidance. No live instance restarted.

The absolute-path tool resolver problem encountered during profiling is fixed
in the separate `fix/tool-absolute-paths` branch (`ffbafe19`), not in this renderer
commit. That change covers confined absolute paths, paginated reads, missing-file
errors, symlink escapes and native write/index behavior. Its full Go suite, vet,
Bun web build and hook checks pass. Both branches are now integrated into main:
renderer merge `9b464fe2`, resolver merge `623262fc`. On the merged code, Unicode
regressions, the TUI/tool suites, all 148 complex-table PTY frames and vet pass.
The full Go suite passed on rerun; its first run failed in
`TestQueueSteerNativeCheckpointKeepsMediaAndAtMostOnce` with `no such table: turns`.
No turn-engine code was changed to address that intermittent failure.
