# Gi acceptance and test-path migration

Gi's tracker distinguishes published backend work from browser acceptance.
The current consumer pin is fixtures-vibes `0259e9a4816d3d643159538095fdffc884104598`,
with Gi backend `78f8f5396a673b4ca2a23d758124c0544dec3668`. The shared editor/Plan
revision work is adopted; conflict scenario workspace018 still fails. #45/#46/#47
stay open. BTW alone was accepted and #40 closed, leaving 43 skip entries.
This adoption removes no additional skips and changes no production capability.
The shared specs/features and Gi skip IDs/reasons are unchanged.

## Owned scratch and profiles

`make test-env` resolves the project root before redirecting child temp paths.
CI selects RUNNER_TEMP, inherited original TMPDIR, then system temp; local
runs prefer writable `/workspace/tmp`. Absolute BASE/ROOT overrides must
agree and end in the canonical `gi` hierarchy. Helpers reject symlinks,
traversal, unowned directories and mutations outside their selected run.

`GI_TEST_RUN_ROOT` defaults to `runs/tests/<worktree>/<run-id>` beneath that
root. Compiler temp, test state, fixtures, Playwright caches/logs/results and
TUI artifacts use this invocation directory. Downloads use `cache/<tool>`;
new Playwright browsers use `cache/ms-playwright`. These tests explicitly used
the existing installed browsers at `/home/agent/.cache/ms-playwright`; installed
tools were not moved. Durable `.gi-run` state and active/shared caches are
preserved by `clean`.

Ordinary Go tests do not capture profiles. `PROFILING=1` captures and analyses
CPU, `alloc_space` and `alloc_objects`, then deletes raw captures. Unique
capture directories prevent collisions; no cross-run pruning occurs.
`PROFILE_KEEP=1` permits current manual investigation and requires removal
immediately after diagnosis. Failed run scratch is available until diagnosis;
successful profiled runs are analysed and disposed by their owner wrapper.

The Gi fixture configuration redirects outputs outside the pinned checkout.
Profiled specs are disposable copies with unchanged assertions; only imports
change to add Chromium CDP CPU and sampled-allocation captures. Node
runner/worker and native fixture processes capture separately. WebKit browser
CPU/allocation capture is unavailable in this helper. Process timing alone
is not browser profiling. The report adapter changes upstream path anchors
only and preserves gate logic.

CI resolves scratch before installs, disables setup-go's home-cache shortcut,
uses profiled pre-release runs and uploads concise measurements instead of raw
captures. The workflow YAML parses; every direct-install job has scratch setup.
CI itself was not run. Some historical TUI Make targets reference absent
`tests/tui` helpers; these were not executed or accepted.

Verification: resolver workspace/BASE/ROOT agreement, CI precedence, generic
system fallback, inherited-root non-nesting, traversal/symlink/outside-run
guards pass. Six testprofile tests pass, including no ordinary captures,
flag preservation and automatic raw disposal. A focused profiled native run
passed 27 tests across testprofile/web; scoped vet, shell/JS syntax and diff
checks pass. The editor asset test now checks the v3.3.0 lazy editor bundle and
CSS bundle rather than the retired standalone editor.css path. Correction:
the migration and initial revision-backend full runs failed
`TestMCPManagerManagesServers`; chained vet/build commands did not run. The
project scratch path hard-wrapped within a path component. Its display assertion
now removes layout whitespace while checking the full path. A fresh profiled
run passes 2,198 tests across 35 packages (`RACE=`); explicit `make vet` and
`make CGO_ENABLED=0 build` also pass. Focused revision (12) and file-open (9)
checks each pass three race repetitions; all three MCP-manager tests pass.
Used per-package profiles were analysed and disposed. No CI or deployment ran.

## Browser checks on c9e5142 plus this migration

All probes use the real Gi fixture binary and shared `b17ef01` assertions.
Temporary profiles enable capabilities only for verification; the production
profile and skips remain unchanged. VNC direct access is fixture-only;
production allowlisting is unchanged.

- Chromium desktop: BTW, editor, terminal and VNC files yielded 31 passes,
  two failures and one skip (34 scenarios). Browser CPU/heap sampling,
  runner/worker captures and native CPU/allocation profiles were analysed.
- Failures: workspace018's Save Copy did not appear in the tree;
  workspace019 lost text typed during a held save. Both require diagnosis;
  issue #46 stays open. Editor003 contains an explicit source skip for a
  renderer-dependent assertion. Missing agent file-open/request-response work
  also prevents editor acceptance.
- Terminal003 initially failed because long isolated workspace paths filled
  the shell prompt and wrapped command echo. A short PS1 in isolated fixture
  HOME fixes the test environment; three captured repetitions pass. Production
  shells are unchanged. Four earlier terminal probes passed.
- Five earlier editor probes passed, including save/confirmation, splitter,
  zen and Vim. One cleanup close-control attempt timed out while the scenario
  itself passed; this is recorded without treating it as full acceptance.
- VNC error/retry state and BTW result/retry/injection/error state passed.
  VNC's shared scenario cannot construct empty/read-only configuration; the
  live fixture checks below now cover those cases. An automatically captured/disposed BTW
  rerun passed, verifying capture and cleanup wiring.

Native runtime allocations are dominated by inference catalogue startup
(about 29–35 MiB per worker), with low CPU sample counts in short cases.
Browser allocations concentrate in Preact/editor handling, CodeMirror,
KaTeX and xterm; idle time dominates many CPU samples. Runner costs include
module parsing and source maps. No measured speedup is claimed. Used raw
profiles, disposable copied specs/fixtures/logs and test binaries were deleted
after analysis. The frozen v0.2.0 run, source UI pin and unrelated Joker guidance
were untouched. A read-only migration delegate timed out without findings.

## Current consumer checks

On `0259e9a` with the backend above, agent-open006/007/008 pass 6/6 on
Chromium/WebKit desktop, including actual tool-driven SSE and cross-chat
isolation. Plan edit/save and dirty remote-update scenarios (shared/original
009/010) pass 8/8; held-save newer typing (workspace019) passes 2/2. There
are no retries. Twenty-nine selected frontend revision/open/SSE tests pass
127 assertions; hook-TDZ checks pass.

Workspace018 fails after Reload: the tab close button stays named "Unsaved
changes" where the shared assertion expects a clean close control. Cleanup
then exceeds its timeout; the combined run produced no final aggregate report.
No count from that interrupted run is used as passing acceptance. Overwrite
and Save Copy in that shared scenario have not passed. The independent HTTP
adapter checks pass 4/4 workloads (Chromium desktop/WebKit phone), including
conditional writes, pending typing, create-only copy, reviewed overwrite,
missing-revision lockout and conditional Plan reset. Those fixtures do not
exercise Gi's live persistence. #46 stays open.

The earlier preview008 exit2 was an instrumentation-copy defect: a literal
`import('../png')` still resolved against scratch. Import relocation now handles
literal dynamic imports as well as static imports; its regression test passes.
Unchanged preview008 assertions pass 3/3 Chromium viewports. WebKit retains
its shared source skip for the Linux upload-bytes limitation. #47's historical
reference to preview008 cannot support a six-project preview acceptance claim.

## VNC and terminal

`make test-vnc-viewer` passes 18/18 workloads: configured read-only, interactive
positive control and empty policy, each on Chromium/WebKit phone/tablet/desktop.
The local RFB 3.8 fixture sends a 2x2 raw colour framebuffer. The canvas check
measures decoded channels against its independent RGB values with a one-unit
8-bit tolerance; it does not compare generated hashes. Read-only input actions
produce no pointer/key frames and clipboard sending is disabled; identical
interactive actions and clipboard sending produce RFB input frames. This tests
UI suppression. The backend remains a binary proxy and does not filter a
malicious custom client's RFB input on read-only targets.

On `0259e9a`, shared VNC015 error/retry passes 6/6 and shell004/007 menu/tab-close
checks pass 12/12. Six native VNC checks pass three race repetitions. Production
allowlisting is unchanged; direct-connect enablement is confined to isolated
retry fixtures. Terminal #45 stays open: WebKit popout restoration failed in
the earlier matrix and in three consecutive focused repetitions. No terminal
frontend changes are included.

## Profiling limits and disposal

Pre-release CPU and both Go allocation views were reviewed. Live VNC workers
allocated 51.6-67.0 MiB / 480k-783k objects across the three policy processes;
model-catalogue cloning accounts for 18.5-38.0 MiB and HTTP/context handling
for much of the remainder. Native CPU samples are short and mostly HTTP/runtime
work. Browser samples concentrate in framework/layout/CodeMirror and idle time;
Node costs are mostly module loading, source maps and runner work. The full
native suite identifies existing codemode argument conversion and transcript
rendering allocations outside this bounded adoption. No equivalent-workload
speedup was measured and no production tuning is claimed.

Chromium browser CPU/sampled-allocation captures were inspected, including the
inline profile attachments in VNC JSON reports. WebKit is functional-only.
Failed probes (missing browser lookup, clipboard locator, import relocation,
conflict timeout and MCP path assertion), matching binaries, fixture state,
traces and used captures were disposed at safe boundaries. The interrupted
terminal matrix was also analysed and disposed after confirming its processes
had stopped. Frozen v0.2.0 output, installed browsers/toolchains, active shared
caches, source and unrelated Joker guidance were preserved. The bounded review
delegate timed out and supplied no review findings.
