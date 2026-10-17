# Full TUI alignment with Pi 1.3

Target: released pi-mono `71023aec1546d6d677a65669c5665c0927ce0dd0` (1.3.0). User requested full alignment on 16 October 2026. This is a behavioural alignment, not replacing Go with the TypeScript renderer. No full parity claim until the crosswalk and acceptance are complete.

Branch: review/tui-pi-1.3-alignment, based on merged main76baa1a2. Release-gate repairs are separate in PR58; v0.1.1 is an immutable failed release attempt and no new tag is authorised by passing only a TUI subset.

## Coverage required

| Area | Upstream source surface | Gi surface | State |
|---|---|---|---|
| Editor/input | tui components/editor,input; coding-agent custom-editor | multiline_input, editor_layout, keybindings, app_keys | Audit cursor, selection, paste, undo, Unicode, history, submit |
| Fullscreen/regular terminal | tui-alt-screen,tui,terminal,stdin-buffer | chat, terminal_output, go-tui fork | Audit resize, retained scrollback, redraw, focus, protocol/IME |
| Transcript rendering | markdown,text,truncate; assistant/user/tool messages | markdown*, transcript*, tool rendering | Audit width/theme invalidation, whitespace, outputPad, thinking/tool expansion |
| Overlays | tui overlay handles and interactive selectors | selectors, settings, workspace index, menus | Unhandled wheel slice implemented; ownership/closure/focus audit pending |
| Settings/keymaps/theme | settings-list,keybindings,theme | settings_list/menu, themes, keybindings | Full defaults/options/error crosswalk pending |
| Runtime UX | interactive.ts, footer, queue, compaction, session tools | chat, turn/store adapters | Audit visible state, steering/followup/cancel/retry, active-tool resume |
| Terminal integrations | terminal-image, OSC/status/cursor/clipboard | terminal theme/output, status, clipboard | Source/contract audit and supported-platform matrix pending |
| Acceptance | upstream tui and interactive tests | Go tests and real tmux harnesses | Per-contract evidence and final profiled suite pending |

Existing wheel acceleration, transcript caching, tool timing, same-session reset and parser-error work must be reused, not reimplemented. Historical docs contain stale go-tui0.18.2 restrictions; current runtime-only0.22.1 fork carries fixes, so verify its actual behaviour instead of blindly downgrading.

## First completed slice

Pi fix11f99114 routes wheel input not consumed by a keyboard overlay to the fullscreen transcript. Gi previously swallowed all mouse events for rename/workspace overlays. It now routes wheel up/down through existing transcript scrolling without closing overlays or mutating drafts. Focused overlay and input-region tests pass race ×3 under Go1.27.2 with isolated project test roots. This is development verification, not full pre-release profiling or PTY acceptance.

## Definition of done

Inventory all upstream tui source modules and relevant interactive components; map each user-visible contract to Gi implementation/tests, or document an explicit host/platform incompatibility for user review. Implement gaps without weakening security, cancellation, history or layout assertions. Verify fullscreen and regular modes at60×18,100×22,140×36, including Unicode, resize, streaming, overlays, cancellation and session changes. Profile CPU/alloc_space/alloc_objects during final pre-release tests, analyse and immediately dispose captures. Full alignment remains in progress.

## Acceptance checkpoint — 2026-10-16

At frontend `407850a1`, shared editor/terminal compliance completed all six
Chromium/WebKit layout projects: 178 passed, 14 existing device/capability
skips, no unexpected or flaky results (192 cases, 16.3 minutes). Capability
and skip declarations were unchanged. The earlier timed-out invocation was
not passing evidence; its orphan server was stopped before the completed run.
Tab-close cleanup warnings occurred without test failures.

Gi `e3100c84` moves the unchanged command-dispatch/palette/helper block into
`chat_commands.go`, with palette/help tests in `chat_commands_test.go`.
This is a first same-package extraction for #11, not completion of the wider
refactoring issue. No runtime behaviour was intentionally changed.

Race verification exposed separate defects fixed at `afad9bcc`: theme
shutdown now drains the watcher and fences queued reloads; settings fixtures
stop their watcher before restoring globals; MCP Tools snapshots its timeout
under the server lock. The shutdown regression checks a queued callback after
idempotent stop. Focused theme/settings/MCP/command tests passed race ×3.
After the autocomplete fixture waited for its completed four-tool discovery
(`69667c84`), the full TUI and MCP package race suites passed under Go 1.27.2.
These are diagnostic development gates, not final release profiling.

Real tmux PTY retry-command and no-selected-model admission matrices each
passed six cases: fullscreen/regular at 60×18, 100×22 and 140×36. Retry checks
cover restart/token ownership, foreign-session rejection, unknown admission,
release without resend, duplicate suppression, metadata and settings. The
no-model matrix checks draft/cursor retention and one explicit admission.
Output now belongs beneath `GI_TEST_RUN_ROOT/tests/`, not source test-results.
Full upstream behaviour crosswalk and remaining PTY scenarios are still open.

VNC live owned-RFB acceptance at `3fd8b2b1` passed 18/18: six browser/layout
projects in each of read-only, interactive and empty-policy modes. It checks
actual decoded framebuffer pixels, keyboard/pointer/clipboard gates, menu
launch and tab teardown. Inherited agent-directory overrides are removed from
fixture children. Shared @ux-workspace-015 error/reconnect acceptance and
capability enablement remain outstanding; #47 is not closed by these slices.

Raw completed results, logs, test binaries and fixture state were removed
after analysis. No merge, release tag or deployment is implied by this record.

### Shared VNC and codemode follow-up

On candidate `4a9baf2e` with frontend `407850a1`, shared VNC
@ux-workspace-015 passed all six browser/layout projects against an isolated
fixture with direct connections explicitly enabled. Production defaults stay
unchanged. The published profile now declares @cap-vnc, and the two stale
VNC-related skip entries (workspace-008/015) are removed. Workspace-008 generic
preview coverage passed Chromium 3/3; its existing three WebKit upload skips
remain. After the declarations changed, workspace-015 passed 6/6 again.
Together with the earlier owned live-RFB 18/18, this completes the measured
VNC error/reconnect, empty/readonly/interactive and menu/tab-close slices.

Codemode consumer acceptance passed 42/42 over Chromium/WebKit phone, tablet
and desktop: live nested call and saved-card reload, script failure without
invented nested calls, cross-session isolation, interruption/cancellation,
legacy output without details, in-place reset and busy-reset completion.
These results do not alone establish every requested #30 presentation detail
(script syntax, duration/error previews and truncated-output links), which
still require review. They are development acceptance, not final release
profiling. Completed scratch data is removed after analysis.

### Codemode presentation repair

Frontend `da963f6a` adds literal-text lexical script spans, collapsible source,
nested error previews and the stored full-output path. The path is shown as a
copyable reference, not an unsafe arbitrary-URL link. Vendored Piclaw remains
unchanged. Script tokenisation preserves the original text; it is not a full
JavaScript parser. Theme colours are inherited, with weight/decoration cues.

Gi now stores the executed script in tool-result details. Conversation pages
omit intermediate assistant tool-call rows, so a parent-call lookup alone
cannot reconstruct source on reload. Legacy results without script details
still render output and do not invent source. The focused native
TestCodemodeCallDetails passes. Frontend card units pass 12/12 (77 assertions).

The refreshed consumer suite passes 48/48 on all six browser/layout projects,
including real script persistence on reload and an injected presentation-only
payload for script tokens, nested error text, full-output path and collapse.
Injected markup remains literal text with no img/svg elements. This injection
does not claim sandbox execution of the payload. The remaining 42 cases retain
real fixture turn/lifecycle coverage. Final release profiling and broad browser
refresh remain required for this new frontend pin. Completed/probe scratch was
removed after diagnosis, retaining only these conclusions.

### Multiline navigation and structural checkpoint

Mechanical extractions `c08a3387` and `a2835061` move shell-turn execution
and TUI tool-text formatting, with matching tests, into focused same-package
files. Declaration blocks were checked unchanged before separating the
behavioural fix below. This reduces large-file ownership but does not complete
all suggested #11 boundaries.

The editor viewport PTY exposed an actual gap: Up/Down always recalled history
instead of navigating multiline drafts. They now move across wrapped editor
rows on grapheme boundaries and fall back to history at document boundaries.
A regression covers Unicode, long soft wraps, boundary progress and unchanged
text. Ctrl+A/E remain line-home/end; the old PTY expected document-home/end
and was corrected to navigate rows explicitly, retaining exact-content checks.

Six real PTYs passed (fullscreen/regular × 60×18, 100×22, 140×36): bounded
32-line Unicode editor, cursor insertion, picker restoration, 42×14 resize,
exact stored draft, one explicit submission and idle shrink. Captures now use
GI_TEST_RUN_ROOT/tests/tui-editor-viewport. All native packages and go vet
passed under Go 1.27.2 after the navigation change; editor regression/golden
keys passed race ×3. Agent-directory overrides were unset: an earlier native
probe inherited host credential paths and failed isolation-sensitive inference
tests; that probe is not acceptance evidence. These development checks are not
final release profiling. Full Pi 1.3 crosswalk, broader PTY/browser refresh and
issue #16 command policy remain open. Completed scratch was removed after use.

### Session/model picker primary-screen restoration

Real session/model PTY acceptance now runs explicitly in fullscreen and
regular mode at 60×18, 100×22 and 140×36. Both modes passed: bounded picker
rows, exact cancel snapshots, A/B draft isolation, unavailable-model errors,
selected-model persistence after process restart, unchanged global settings,
no unintended turns and no extra editor rows. The native durable-model-command
regression also preserves a draft across session switches.

The old harness launched from the source directory without an isolated HOME,
and did not wait for fully typed commands before Enter. It now uses the owned
fixture workspace/HOME, removes inherited agent-directory overrides, waits
for input and dismisses completion before executing a supplied command.
Captures belong beneath GI_TEST_RUN_ROOT/tests/tui-sessions/<mode>.

Regular-mode acceptance exposed a real one-row transcript shift on picker
cancellation. ExitAltScreen already restores the saved inline layout; the
subsequent resetInlineHistoryModel printed an empty history line. Removing
that call from picker teardown preserves the exact primary-screen snapshot.
The exact assertion was retained, not relaxed. With this fix, fullscreen
session/model checks were rerun, editor viewport passed all six mode/size
cases, and the full internal/tui race suite passed under Go 1.27.2. This is
not final release profiling or exhaustive Pi 1.3 parity. Completed scratch
and failed-probe captures were deleted after analysis.

### Native profiling and external pending-panel changes

At `a2ee6531`, Go 1.27.2 profiled native acceptance completed 2,340 tests
across 35 packages with no failures. The retained diagnostic run recorded
82 seconds test wall time, 41.9 seconds child CPU and 769 MiB peak child RSS
(including linking); wrapper wall time was 89 seconds. Cumulative CPU,
alloc_space and alloc_objects were inspected before disposal. TUI allocations
were 755 MiB: transcriptRowsAtWidth accounted for 266.6 MiB cumulative (35.3%);
flat NewBuffer 208.4 MiB and segmentStyledRunes 96.2 MiB were substantial.
TUI allocation objects included cloneModel 518,524 and wrapSpans 479,798.
Scripting CPU was dominated by embedded-Joker execution (2.27 s cumulative),
and codemode VM execution accounted for 436.1 MiB cumulative allocations.
These are whole-test workloads, not production latency estimates. No measured
improvement is claimed from the repeated run; no equivalent controlled tuning
comparison was made. Final integrated revision still needs profiling.

Queue-display PTYs then exposed an idle invalidation defect: externally
removed queue items remained painted when footer data did not change. The
existing five-second fallback now includes the read-only pending-panel
projection in change detection, without unconditional redraw. Fullscreen and
regular queue-display proofs pass: external add/remove, media-only display,
session isolation, steering/follow-up order, draft preservation and drain.
The helper now waits for supplied text and dismisses slash completion before
command execution; captures use its owned run directory. The full TUI race
suite passes after this fix. Raw profiles, binaries, disposable logs and
failed probe captures were deleted after analysis.

### Published Pi 1.3 reference and selection checkpoint

The exact published @earendil-works/pi-tui and pi-coding-agent 1.3.0 packages
now drive `make test-tui-pi130-reference` (the old pi110 target is an alias).
Navigation defaults, existing editor scenarios and all four platform hotkey
fixtures match. A new upstream-generated vertical-column scenario verifies
column restoration across short lines. Pi keeps a preferred visual column;
Gi now does so too, resetting it on edits and horizontal/line navigation.
Up at the first visual row moves to line start before entering history;
Down at the last row moves to line end unless already browsing history.
History ownership tests reflect these measured Pi semantics. README now states
Go 1.27.2, go-ai v1.3.0 and the in-progress Pi TUI 1.3.0 target, not old pins.

Real fullscreen selection/copy PTYs passed at 60×18, 100×22 and 140×36:
SGR double-word/triple-line and reverse-word selection, forward/reverse drag,
OSC52 release/Ctrl-C/Ctrl-X copy, opt-out, highlight/Escape/editor preservation,
held-edge autoscroll, resize/output invalidation and no added idle rows.
The harness isolates HOME and agent overrides, navigates with Ctrl+Home and
Ctrl+Down to the first prompt rather than assuming plain Home or page offsets,
and puts captures under GI_TEST_RUN_ROOT/tests/tui-selection. One refreshed
run timed out at the 140×36 double-click check; an unchanged rerun passed all
three sizes. This remains a timing/flakiness limitation, not deterministic
release evidence. No assertion was relaxed to obtain the rerun.

After preferred-column changes, viewport PTYs pass six mode/size cases, full
TUI race tests pass, and all native packages plus vet pass under Go 1.27.2.
Reference regeneration passes its 1.3.0 package/version guards and all three
Bun unit tests. Only these reference surfaces are proven; exhaustive TUI
crosswalk, browser refresh and final integrated release profiling remain open.
Completed reference installs, raw captures, binaries and probe logs were
removed after analysis; source goldens and these conclusions remain.

### Selection timing diagnosis

The earlier 140×36 double-click timeout reproduced before any selection code
change. The harness waited only for prompt/reply text to be visible after
Ctrl+Down; at the largest viewport both were visible before the navigation
render finished. Click coordinates were calculated from that old viewport.
The failure diff showed the startup header disappearing between the recorded
baseline and clicks. The wait now also requires SELECT-01 at the viewport top,
confirming prompt navigation actually finished. No runtime click threshold,
selection assertion or clipboard check changed, and no click retry was added.
The complete three-size suite then passed three consecutive runs (nine PTYs).
This resolves the observed race; it is not a claim about all terminal timing.
Completed probe captures, binaries and logs were removed after diagnosis.

### Offline informational commands (#16)

`/changelog [version]` now reads an explicit allowlist of bundled notes,
currently `unreleased`; it performs no network or model call, accepts no
filesystem path and caps assets at 64 KiB. Unknown released versions report
that no notes are packaged. Candidate notes explicitly do not announce a
release. `/bug [preview]` renders local allowlisted build/OS/architecture/TUI
state only. It does not marshal RuntimeConfig, inspect environment variables,
read credentials/transcripts or include paths, endpoints, model/provider
names or session IDs. Save/upload arguments are rejected; no file or network
effect exists. Users review and copy the preview themselves.

`/trust` remains absent from completion. An explicit invocation reports that
project MCP configuration is ignored and no trust state changed. This avoids
pretending that a trust-enforcement policy exists; issue #16 remains open for
that contract decision. Existing MCP tests verify untrusted project config is
ignored. Diagnostic units check exclusions, path/version rejection, draft
preservation and explicit trust feedback. Full TUI units pass, diagnostic/
command/draft tests pass race ×3, and vet passes under Go 1.27.2. No release
profiling claim is attached to these development tests. Scratch is disposed
after analysis.

### Queue restore and command PTYs

At candidate `5dab97ad`, Alt+Up queue restore passed fullscreen and regular
mode: two queued prompts and the current draft return in order, both queued
turns become cancelled/aborted, queue counters clear and no redelivery occurs.
Queue command acceptance passed six mode/size cases (60×18, 100×22, 140×36):
restart persistence, exact-snapshot move, chronology/metadata preservation,
remove/steer, stale/duplicate rejection, draft preservation, no new turns and
no additional idle rows. Its restart now uses the documented double-Ctrl+C
exit instead of assuming one press exits. Assertions were unchanged.
Both helpers now place disposable captures under GI_TEST_RUN_ROOT/tests.
Completed logs/captures/binaries were removed after analysis. This acceptance
does not establish external-provider admission or complete Pi TUI parity.

### Terminal API path and handoff refresh

The terminal browser protocol checks passed three repetitions (six tests):
actual PTY workspace execution, 91×27 resize, shell-state/session-ID retention
through handoff, disconnect cleanup and cross-origin/client-identity rejection.
On this host /workspace resolves to /srv/piclaw-dev/workspace in shell PWD;
the assertion now compares realpath directory identity and reads an anchored
WORKSPACE output line, not the echoed printf command. This keeps the exact
workspace-execution requirement without requiring one symlink spelling.
It does not replace terminal dock/tab/popout UI acceptance. Completed probe
artifacts and the owned server were removed after analysis.

### Current browser/native refresh at 2ed15afe

With frontend `da963f6a`, session-thinking passed 36/36 and the separate
composer/voice/notifications/startup matrix passed 132/132 across all six
projects (9.6 minutes). The first invocation enabled multiple single-suite
flags, so conditional testMatch selected thinking only; the remaining specs
were then explicitly selected with GI_WEB_REGRESSION_LIST_ALL. No combined
coverage was inferred from the first run. Recognition and notifications use
mocked browser events; physical microphone quality is not established.

Profiled native verification passed 2,342 tests across 35 packages; vet and
nested go-tui tests also exited zero under Go 1.27.2. Test wall 87 seconds,
child CPU 50.5 seconds, peak child RSS 782 MiB including builds; wrapper wall
94 seconds. Cumulative CPU, alloc_space and alloc_objects were inspected:
TUI database/sql retry paths accounted for 26.2% cumulative CPU, fixture
session setup 24.0%, SQLite prepare 23.6%; transcriptRowsAtWidth allocated
278.5 MiB cumulative (37.5%), go-tui.NewBuffer 211.0 MiB flat (28.4%).
Wrapped-spans/height work accounted for 1,269,090 cumulative objects (29.0%);
renderTranscriptRows 1,040,778 (23.8%). Profiles are whole-suite diagnostics,
not production-latency measurements. No controlled tuning improvement is
claimed. Earlier cache-backed/non-equivalent run deltas are not optimisation
evidence. Raw captures, matching scratch, logs and fixtures were removed after
analysis. Remaining broad browser suites, full TUI crosswalk, #11/#16 policy
completion and final integrated release acceptance still block v0.1.2.

### Theme, tool syntax and streamed-table PTYs

Theme negotiation passed 13 controlling-PTY scenarios against the exact
published Pi coding-agent 1.3.0 system-theme generator: OSC background and
palette, fragmented replies, CSI scheme, fallbacks, explicit/paired themes,
disabled queries and light/dark native UI. The Make target now installs its
own exact reference under the owned test root and disposes it, rather than
requiring a global Pi install. Package identity/version is checked.

Tool-syntax PTYs passed fullscreen/regular × widths 60/100/140, including
native read/write, keyword colour, collapse/expand, narrow resize and stored
reopening without reading changed source. Its Make target now passes the
actual project-owned test binary path. Theme/tool-syntax Make targets were
rerun successfully after wiring changes; captures use GI_TEST_RUN_ROOT/tests.

The six table PTYs passed streaming, Unicode grid borders, resize, retained
draft, final cells and stored Markdown/reopening checks. Initial failures
were fixture-auth isolation: a login shell reintroduced host agent-directory
overrides. The table fixture now clears both overrides with test-scoped env
restoration, and the launcher removes them. This is not a renderer assertion
relaxation. Full TUI units and vet pass under Go 1.27.2. Completed captures,
failed probes, reference installs, logs and matching binaries were removed
after analysis. Remaining terminal/runtime crosswalk and integrated release
gates are still open.

### Search reflow, session actions and timing

Session-action PTYs passed six fullscreen/regular × size cases: pin/unpin,
archive/restore, rename save/cancel/validation, root/busy gates, resize,
draft/cursor/history preservation and stable dock geometry. Transcript search
passed three fullscreen sizes with 96 occurrence checks, case-insensitive,
Unicode/cross-wrap matching, precise highlights, next/previous/wrap, hidden
and expanded tool output, live output, reopen, draft/history restoration and
no query submissions. A real bug changed the chosen occurrence on resize:
refresh compared old wrapped coordinates to the new layout. Layout-only
refresh now retains occurrence index; a native regression covers both matches
across multiple widths. Content/expansion changes retain their existing
refresh logic. Startup-aware prompt navigation replaced stale Home assumptions
in the harness; match-count/selection assertions were not relaxed.

Tool timing passed six mode/size cases: live elapsed ticks, frozen completion
and failure, native timestamp agreement within the existing 500 ms contract,
resize/cursor/draft and idle geometry. Full TUI race suite and vet pass under
Go 1.27.2. These helpers use owned tests/ capture paths and isolated HOME/
agent overrides. The reading helper now accepts the actual binary override.

Reading-position acceptance is explicitly failing at "typing moved the history
reader": its assertion requires preserving scroll during edits, while current
ensureInput deliberately calls scrollTranscriptToBottom on user edits (also
assumed by selection/search acceptance). Published Pi 1.3 editor input only
establishes immediate render here, not a directly equivalent Gi fullscreen
historical-viewport contract. No runtime change or assertion relaxation was
made to paper over this conflict. This remains an open release/TUI crosswalk
gate. Completed failure captures, reference installs, logs and binaries were
removed after diagnosis; concise limitation retained here.

### Reading contract resolved; regular-mode retained-history failure

Git history resolves the edit/reading conflict: `b452a2dc` (September 22)
introduced edit-preserved scroll; `e476b704` (September 30) deliberately made
user edits return to latest and added separate restoration-preserves-position
tests. The older reading PTY had not been updated. It now verifies editing
returns to latest, then explicitly re-enters history before asynchronous
completion, resize and submission. All three sizes pass, retaining cursor,
exactly-once storage and idle geometry assertions. No runtime scroll change
was made. The corresponding edit/restoration and regular unit tests pass.

Source-copy PTYs pass six fullscreen/regular × size cases: exact OSC52 source
bytes (including whitespace/trailing newline), newest persisted assistant
selection, clipboard-off fallback, unsent draft and stable geometry. Source
bytes are an exact clipboard contract, not generated-signal equivalence.

Regular terminal acceptance remains failing at 60×18: after width/height
resize and temporary picker cancellation, subsequent local-command output
loses the earlier gate-response text from native scrollback. Added snapshots
and an earlier intact() check isolate this to the first local error output.
The native copy-mode probe now locates the shell marker instead of assuming
no leading blank scrollback, and multiline reset explicitly clears both
logical lines (Ctrl+A/K is line-local). Those harness fixes do not remove the
history-preservation assertion. Speculative picker-width invalidation and
invalid-layout growth changes did not fix the failure and were reverted.
No passing regular-mode acceptance is claimed. This concrete retention bug
still blocks release. Completed probe data and matching binaries were removed
after diagnosis, keeping the failing test and this concise reproduction.

### Regular retention root cause and fix

The retained-history failure was caused by picker code calling Terminal's
EnterAltScreen/ExitAltScreen directly while continuing to resize the inline
App. That restored the primary screen's pixels without restoring its inline
history geometry. Subsequent PrintAbove output used alternate-screen layout
bookkeeping and overwrote primary retained rows. The picker now calls App's
EnterAlternateScreen/ExitAlternateScreen, which save/restore geometry and
pixels together. No empty history print or speculative resize workaround is
needed. No vendored runtime changes were required.

Regular acceptance passes all three sizes, including native scrollback copy,
retention through output/resize/picker/session transitions, full tool output,
no mouse capture, exit and reopen. Regular session/model acceptance passes
all three sizes with exact cancellation snapshots, draft isolation and model
restart persistence. Full TUI race tests and vet pass under Go 1.27.2.
This resolves the concrete retention blocker recorded above. Completed probe
captures, binaries and logs were removed after analysis.

### Issue review and final shared-compliance restart

Merged frontend PR #4 (`237e66dc`) and Gi PR #60 (`f1910265`) repair active
compaction accessibility: active meters no longer advertise the idle Compact
context action; the shared locator recognises the active measurement without
matching Stop response. All 54 shared compaction cases pass. The first full
shared run was interrupted after exposing these failures and is not passing
evidence. Full 1,590-case shared compliance was restarted on the repaired tree;
its result must be recorded separately when complete.

Issues #47 and #30 were closed with merged VNC/codemode acceptance; #25 was
closed after #30 plus integrated MCP/codemode/TUI/turn race suites passed.
#16 is closed as a decision: offline bounded changelog, allowlisted local-only
bug preview and explicit unsupported trust. Project MCP config remains
ignored; a meaningful trust/enforcement system is future scope, not implied
by this closure. Documentation's obsolete OAuth/codemode status is corrected.

For #11, `04abdf54` mechanically extracts the session collection handler into
api_sessions.go and moves its main-session allocation test alongside it.
The declaration block was checked unchanged; all native tests and vet pass.
Together with command/tool-text/shell-turn extractions, this establishes
functional same-package boundaries without speculative subpackages. Broader
shared/functional browser acceptance is still running, so #11 remains open.
No release or deployment is implied. Active run artifacts must not be removed
until their jobs finish and results are analysed.

### Acceptance follow-up in progress

The functional gate's first run stopped after exposing an obsolete radio
selector (`monokai` instead of the current `Monokai Original` accessible name).
The selector now uses the exact current name; persistence and stored theme-ID
assertions are unchanged. Functional acceptance was restarted and its result
is not yet established. The independent 1,590-case shared run continues on
`f1910265` with frontend `237e66dc`; no completed result is claimed. Dependency
symlinks and active run roots must remain until the owners finish. The later
session API extraction is a checked pure move, but exact integrated acceptance
must still include it before release. Do not tag based on these partial runs.

### Browser acceptance repairs after the stopped shared run

The theme/picker matrix passed 126/126 across six projects after the stored
appearance-command receipt (`9900bc08`) and deferred picker focus changes. Web
race checks, native tests and vet also passed at that checkpoint.

Subsequent functional diagnosis found and repaired real host integration gaps:
mobile picker dismissal, duplicate hidden-file event ownership, expanded subtree
refresh, clipboard guard resolution before the overlay resolver, session prefix
matching, clean external editor refresh notifications, and mobile tab/menu overlap.
The shared frontend checkpoint is `a45cb2da`; its units pass 193/193 and the hook
check and build pass. Vendored Piclaw sources remain unchanged.

Functional tests now use native option IDs/ARIA state, native editor tabs and
Markdown preview, revision-conditional saves, and the project-owned fixture DB.
The removed custom read-only tab host is not the current editor contract. The
latest full functional run passed 146 cases with 13 existing capability skips and
one harness failure: a tab-centre click hit its close control. Targeting the tab
label fixed that case; focused acceptance passed 2/2. This is not a final full-run
pass, integrated-main acceptance, or release authorization. Issue #11 stays open.

### Integrated browser and queue-boundary acceptance

Frontend PR #5 merged at `34452428`; Gi PR #61 pins it. Full functional
acceptance on that tree passed 147/147 runnable cases with 13 existing capability
skips. The viewport-specific native editor test targets the tab label rather
than the tab centre, which can legitimately contain its close control.

Issue #11's next mechanical boundary is `a597df8d`: 299 unchanged implementation
lines move from `engine.go` to `queue_launch.go`; three unchanged launch/conflict
regressions move to `queue_launch_test.go`. Function/test bodies were compared
against their pre-move source. No public renames or new sub-packages were added.
The package composition roots deliberately remain; smaller files are not a
reason to introduce circular ownership or hide queue invariants.

Post-extraction native pre-release verification passed 35 packages / 2,344 tests,
vet and the Bun hook check. CPU, allocation-space and allocation-object profiles
were analysed, then raw captures and matching test binaries were disposed of by
the profiling helper. Test workload: Go 1.27.2, wall 92s, CPU 54.1s, peak RSS 769 MB
including build. These are observations, not a controlled speedup comparison.
Hotspots remain interpreter calls/compilation, bounded auth-state decoding and
transcript row projection; preserving validation and terminal semantics takes
priority over speculative caching in this structural change.

Post-extraction PTYs passed regular transcripts (three sizes), session actions
(six fullscreen/regular cases), queue commands, queue restoration and selection
(three sizes). The broad 1,590-case shared suite and another profiled functional /
race / Pi1.3 reference sequence are running. They are not yet passing evidence;
issue #11 and v0.1.2 remain gated on their results.

The post-extraction profiled functional rerun also passed 147 cases with 13
existing skips. Race checks passed for TUI (50.867s), turn (55.196s), web (19.448s)
and MCP (5.058s); Pi1.3 navigation/editor/hotkey goldens passed. Functional
profiling observed 249s wall, 100s CPU and 1,368 MB peak RSS. Node exit heap samples
are live allocations, not allocation history; repository Go allocation profiles
were analysed separately. No measured performance improvement is claimed.

Extended PTYs passed editor viewport, sessions/model restoration, search and
reading. Tables initially could not launch because the standalone script needs
`GI_TABLE_PTY_BIN`, not `GI_TUI_BIN`; after building the matching test binary and
providing the documented variable, all six table cases and six source-copy cases
passed. This was a harness invocation failure, not a masked product result.
Completed raw profiles, failed probe fixtures and matching test binaries are
removed after analysis; active shared-suite data is retained until its owner
finishes. The shared suite still has no final verdict.

The remaining targeted TUI checks passed: six syntax PTYs, six tool-timing PTYs
and 13 theme PTYs against the pinned Pi1.3 package. The successful completion
marker is `LAST_TUI_PASS`. Temporary reference installs and matching binaries
are disposable; no global Pi installation was changed. Shared-browser acceptance
is still running, so this does not close #11 or authorize a release.

Cross-compilation passed for Linux/amd64, Linux/arm64, Darwin/amd64,
Darwin/arm64 and Windows/amd64 through `check-cross-build`. Windows/arm64 was
built separately to cover the sixth release matrix target and also passed.
These are build checks, not runtime tests on those operating systems. Disposable
cross-build outputs were removed at completion. The shared-browser suite remains
active and has no final verdict.

The integrated web-adapter gate exposed stale Bun transpiler entries resolving
Piclaw3.3.0 imports after the pin moved to3.3.4. Repeating the unchanged tests
with `BUN_RUNTIME_TRANSPILER_CACHE_PATH=0` passed 193/193 frontend units and the
host adapter checks. The Makefile/test-environment helper now disable that cache
explicitly; the gate passed again without a command-local override. No source
assertions or vendored modules were relaxed. This is a verification-environment
repair; the active shared browser run is untouched.

### Full shared-browser verdict

The six-project matrix completed: **1,519 passed, 71 capability skips, zero
browser failures** across 1,590 cases. The upstream compliance gate, generated
from those results without rerunning tests, reports **Gate: OK**: 310 catalogue
scenarios, 236 passed, zero failed, three skipped and 71 with no suite test. The
last category is explicitly not passing coverage or a blanket parity claim.

The first reporting invocation failed because its relocated temporary source
could not resolve AJV. `scripts/fixtures-report.ts` now validates and relocates
both AJV package imports to the pinned dependency tree, just as it already did
for local imports. Running the unchanged upstream gate on the completed results
then succeeded. Only import paths changed; schema and acceptance logic did not.
The original profiled make invocation therefore remains a failed invocation;
its successful browser results and successful report regeneration are recorded
separately. Profile analysis/disposal is still running before final cleanup.

### v0.1.2 release-CI failure and next patch repairs

Tag v0.1.2 at57dc2a8c is immutable. CI38976489907 exposed additional release
surfaces absent from the earlier local gate selection: obsolete helper filenames,
an unrelated removed pane-host fixture prerequisite, pinned picker geometry /
touch dismissal, and intermittent settings-to-composer focus restoration. The
run was cancelled after diagnosis; no successful v0.1.2 release is claimed.

Repairs are staged on review/release-gates-v0.1.3 and frontend
review/release-picker-contracts (5870d21a). Current verification: picker geometry
24/24, compose surface36/36, tool-terminal12/12, and all release helper unit
targets pass. Tool-terminal diagnosis also corrected project-owned queue release
paths and obsolete native session selectors; assertions remain intact. Focus
restoration retries are bounded and act only when no control owns focus and no
modal is open. Native close-button Enter bypasses session row activation.

These are focused development results, not final integrated pre-release evidence.
Frontend/Gi PRs, independent review, profiling and applicable final gates remain
before a new patch tag. Preserve both failed v0.1.1 and v0.1.2 tags.

Frontend PR6 merged58a97a42; Gi integration pin c3bff8f2. Integrated pre-release
profiles and checks pass: 35 packages /2,344 tests (99s wall,52.6s CPU,791 MB RSS
including build), vet, functional147 passed /13 existing skips, picker24/24 and
compose36/36 across six projects. CPU/allocation summaries were analysed and
completed captures disposed by the profiler. Node exit heaps describe live
samples, not allocation history; no speedup claim follows from these runs.
The full1,590-case matrix is running under patch13-shared; no final verdict yet.

### Additional release-panel audit (patch13)

Running the CI model/session/voice targets directly against the existing merged
frontend binary found further stale selectors and missing model-picker semantics.
Voice passed30/30. Session acceptance was updated from removed wrapper IDs to
native option IDs and listbox active-descendant state; all12 cases pass, including
first-frame search/navigation and backend pin persistence. No assertions skipped.

Model-panel audit remains red:36 failures /6 passes across six projects. Removed
`.current-model` markup is a stale test assumption, but trigger `aria-haspopup`,
listbox busy state and loading/switching guards must be reviewed and restored
where required before release; do not merely delete these assertions. The active
1,590-case shared run is untouched and not yet complete. Passing its overlapping
cases will not substitute for these explicit release-CI contracts.

Isolated model repair checkpoint: frontend864db3c adds busy-state/option guards,
clear-search refocus, trigger popup semantics and response focus ownership.
Units193/hook/build pass, but the focused phone contract still fails and this
checkpoint is NOT integrated. Its test-only consumer clone is at
`/workspace/tmp/gi/runs/model-contract/worktree` with server
`/workspace/tmp/gi/runs/model-contract/server`; the clone's model spec was restored
to source after diagnosis (no weakened focus assertions retained). Failure details
in the current diagnosis: old `.current-model` selector; trigger `aria-expanded`
missing; option instance IDs change across reopen; current-model selection is a
native no-op while pending-response tests click it. Separate genuine contracts
from stale assumptions, preserving loading/disabled/no-write/focus assertions.
The active shared checkout remains on merged frontend58a97a42 and was not rebuilt
with this candidate. Session acceptance change365cd6bf is local pending publish.

### Isolated model contract repair verified

Frontend model changes through a6e5c7f (unit-only coverage checkpoint befbffc)
passed42/42 across all six projects; frontend193units /1217assertions and hook
check pass. The consumer test now uses native command POST admission rather than
Settings-only PATCH, explicit accepted/rejected responses, native selected-option
and active-descendant semantics, and a count derived from the supplied catalogue.
Loading/no-extra-write/disabled/focus/Tab assertions remain. Stable list IDs,
trigger ARIA, busy keyboard guards, inline errors and scroll-region Tab exclusion
are actual host repairs. No final integrated acceptance claimed: active shared
run still uses58a97a42 and must not be modified until it finishes.

FrontendPR7 merged922fd997. The isolated merged-pin rerun passes model42/42
with Node CPU/exit-heap samples, and native35packages/2344tests plus vet. Native
Go CPU/alloc_space/alloc_objects were analysed/disposed by the helper; workload
wall207s, CPU74s, peakRSS773MB including build (concurrent browser workload; not
an equivalent-workload speedup comparison). Node captures analysed separately;
exit heaps are live samples, not allocation histories. Raw model browser profiles
and disposable run data were removed after analysis. The source checkout running
shared acceptance remains on58a97a42, so final integration of922fd997 is pending
that safe boundary; no release authorization is implied.

Patch13 shared run recorded a Chromium-desktop shared-messages-tool failure:
after navigation, expected third reply absent and snapshot stayed `Loading Gi…`.
Trace analysis finds `net::ERR_NETWORK_CHANGED` loading app.bundle.js plus
marked/KaTeX/mermaid/CSS; several subsequent API requests have status -1. This
is network-loading evidence, not proof of a persistence defect or a passing
matrix. Active run continues unchanged. An isolated exact-assertion repeat of
both shared-messages-tool cases (three repetitions) is running on922fd997;
retry-only evidence will not replace a clean final integrated run.

Merged model922fd997 isolated profiled functional acceptance passes147 cases
with13 existing skips. Race passes web19.897s,turn54.955s,TUI50.163s,MCP6.074s.
Functional workload278s wall,106sCPU,1368MBpeakRSS; captures analysed/disposed,
no controlled improvement claim. This supplements merged-model42/42 and native
2344/vet. The old-pin shared run remains active with one documented network load
failure. Its full rerun frequency limit is four hours; do not start another full
matrix until that boundary, even after active-job cleanup.
