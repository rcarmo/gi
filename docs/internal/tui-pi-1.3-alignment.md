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
