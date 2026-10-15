# go-ai and Pi 1.1.0 baseline refresh

Gi pins go-ai `e8fe159110e2` as
`v1.1.1-0.20261009233122-e8fe159110e2`. This follows the published native
1.1.0 baseline and retains the Codex 4 MiB WebSocket limit. The dependency
change since Gi's previous `6244052633a0` pin is README-only; no provider API
adaptation, credential change or additional dependency upgrade was needed.

## TUI reference

`make test-tui-pi110-reference` installs exactly published pi-tui and
pi-coding-agent 1.1.0 in isolated project-owned scratch, regenerates goldens
there and compares them without rewriting the checked-in fixtures. Navigation,
editor operations and Linux/macOS/WSL/Windows hotkeys match. Generators now
reject other package versions or identities before importing them. The
installed Pi harness and managed settings/authentication remain untouched.

This revalidates the existing bounded TUI port, not every Pi feature.
[Remaining 1.1.0 work](../terminal/tui-pi-1.1.0.md#remaining-110-work) includes
output padding, MCP lifecycle comparison, shell/codemode presentation and
new tool-loadout semantics. OSC 7501 and recorded tool duration are already
implemented. No physical-terminal cross-platform acceptance is claimed.

## Verification

- Three reference-helper tests pass; exact published golden comparisons pass.
- `make test-upstream-alignment PROFILING=1`: 114 focused tests pass three
  race-enabled repetitions across TUI, inference, configuration and turns.
  This includes navigation, editor/hotkeys, recorded durations, OSC 7501,
  native transport precedence, a large provider WebSocket event and retries.
- Full native normal run: 2,275 tests across 35 packages; 2,274 passed and one
  failed because the separate, unfinished Piclaw migration had removed the
  committed 3.3.0 viewer-policy reference file from the working submodule.
  Restoring that exact committed reference made the unchanged failed check
  pass in a focused rerun. No second full run or fresh browser acceptance is
  claimed.
- Explicit `make vet` and a CGO-disabled `make build` pass; the isolated
  binary's `--version` check passes and that binary is disposed. The frontend submodule pin and its unfinished
  migration are excluded from this change. README names the committed
  Piclaw 3.3.0/fixtures `92425ad` baseline, not an unaccepted 3.3.2 candidate.

## Profiling and disposal

CPU, alloc_space and alloc_objects were inspected. Repeated race checks are
103 seconds wall/44.9 seconds process CPU including compilation and PTY
checks. Inference allocation sampling attributes about 40 MiB and 281,268
objects to registry initialisation, and 7.44 MiB to the large-message fixture.
TUI allocation sampling is dominated by test setup and registry cloning;
full-suite codemode VM calls account for about 388 MiB/6.8 million allocated
objects in that package view. These are existing workloads, not a measured
regression or improvement from a documentation-only dependency revision.
No pooling, ownership changes or assertion relaxation was justified.

An initial race invocation hit the tool's execution timeout after three
packages passed; its available captures were analysed and disposed before
the successful unchanged rerun. Used profiles, matching test binaries,
isolated reference packages and disposable logs are removed after analysis.
Active previews and the unfinished frontend source are preserved.
