# Gi acceptance and test-path migration

Gi's tracker now distinguishes published backend work from browser acceptance.
Issue bodies #25/#40/#45/#46/#47 were reconciled on 8 October 2026. #27–#29
are closed; #30's remaining work is web rendering. All 44 skip IDs/reasons
were retained; descriptions and the fixtures pin now match `b17ef01`.

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
CSS bundle rather than the retired standalone editor.css path. The full Go
regression suite and vet also pass; used per-package profiles were analysed
and automatically disposed. No CI run or production deployment was made.

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
  VNC's shared scenario cannot construct empty/read-only configuration; those
  cases still need separate checks. An automatically captured/disposed BTW
  rerun passed, verifying capture and cleanup wiring.

Native runtime allocations are dominated by inference catalogue startup
(about 29–35 MiB per worker), with low CPU sample counts in short cases.
Browser allocations concentrate in Preact/editor handling, CodeMirror,
KaTeX and xterm; idle time dominates many CPU samples. Runner costs include
module parsing and source maps. No measured speedup is claimed. Used raw
profiles, disposable copied specs/fixtures/logs and test binaries were deleted
after analysis. The frozen v0.2.0 run, source UI pin and unrelated Joker guidance
were untouched. A read-only migration delegate timed out without findings.
