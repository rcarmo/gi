# Pi TUI 1.1.0 alignment

Gi now targets Pi TUI 1.1.0 behaviour. The first bounded port separates editor
line navigation from fullscreen transcript navigation; the Go/SQLite runtime
and independently pinned Classic frontend are unchanged.

The reference is the published `@earendil-works/pi-tui@1.1.0` and
`@earendil-works/pi-coding-agent@1.1.0` packages, released 7 October 2026.
`dist/keybindings.js`, the coding-agent `/hotkeys` implementation,
`docs/keybindings.md`, terminal setup and release notes supplied the contracts.
`scripts/golden-pi-navigation.mjs` extracts the selected defaults from exactly
pi-tui 1.1.0; its fixture is retained in `internal/tui/testdata`.

## Navigation

Home/End move to the current editor line's beginning/end, alongside Ctrl+A/E.
Ctrl+Home/End scroll the fullscreen transcript to its top/bottom without moving
the draft cursor. Ctrl+End resumes following output. Regular mode leaves
scrollback to the terminal and does not repurpose Ctrl+Home/End as editor keys.
Selectors retain their own Home/End navigation. The jump-to-latest hint now
names Ctrl+End. Linux/macOS/Windows/WSL hotkey navigation rows were compared
with the 1.1.0 source/defaults and updated.

Verification on Go 1.27.1: 57 focused native checks each pass three race
repetitions, including input line/cursor preservation, fullscreen dispatch,
regular-mode and selector boundaries, selection lifecycle, jump hint and
existing editor-key goldens. The tmux smoke passes: it submits a literal draft
which proves Home/End edit at the intended positions and Ctrl+Home/End retain
the draft cursor. Its isolated quietStartup setting avoids unrelated host
resource listings filling the small viewport. The smoke exits via Ctrl+D so
shutdown completes before analysis/cleanup.

Fresh full native verification passes 2,201 tests across 35 packages with
`RACE=`; explicit vet and CGO-disabled build pass. Shared frontend source,
capabilities and skips are unchanged. No deployment or physical-terminal
cross-platform test ran. The installed harness still uses Pi 1.0.4; it was not
upgraded as part of Gi's reference comparison.

## Remaining 1.1.0 work

Selection invalidation on session switch is already implemented: Gi clears
selection/click history before replacing the transcript; scoped stale-copy
completions are fenced. Focused selection/session tests passed. Other 1.1.0
contracts need separate implementation or verification:

* Recorded tool duration now uses execution measurements for live and resumed
  results; missing legacy measurements stay unknown. See [tool-duration.md](../../internal/tool-duration.md)
  for boundary, projection and verification limits.
* `outputPad` must apply consistently to tool, shell and summary output. Gi
  does not yet expose this setting across those render paths.
* OSC 7501 program status needs support negotiation/explicit opt-out, bounded
  metadata-only messages, and idle/working/blocked/done/error transitions.
  No protocol report should contain prompts or model output.
* `/mcp` must remain usable during enable/reconnect/disable, and sign-in timeout
  must cover the whole cancellable request chain. Gi opens its manager live,
  but action screens currently replace it with status; whole-chain deadline
  and shutdown behaviour require comparison with the 1.1.0 implementation.
* Split ANSI shell chunks, dim `!!` headers, codemode output separation and
  async discovery declarations need comparison. Native tool-loadout `+name`
  and `-name` semantics are also new in 1.1.0 and have not been ported here.

Pi extension APIs, JavaScript components and provider-specific release changes
are not automatically Gi TUI contracts. Existing Gi commands and fixed-key
configuration differences stay explicit; the Go renderer is not replaced.

## Profiles and disposal

Pre-release native CPU, alloc_space and alloc_objects were inspected. Focused
navigation/selection sampling concentrates in rendered transcript/Markdown
layout and test fixtures; full-suite hotspots include existing codemode and
transcript allocations outside this navigation slice. No equivalent-workload
speedup was measured. The tmux lifecycle has process wall/CPU/RSS measurements,
but the production TUI binary does not capture CPU/allocation profiles; native
in-process regression captures supply those views, not the PTY run.

Used native profiles and matching test binaries were disposed by the wrapper.
The downloaded reference packages, failed startup probe, tmux artifacts and
scratch were deleted after comparison/use. Installed toolchains/dependencies,
source, active shared caches, durable runtime state and unrelated Joker edits
were preserved.
