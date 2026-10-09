# Delegation items, 8 October 2026

Implementation baseline: Gi `a2ffbdf28803b4ab2114861f7bdfd9f5196db79d`,
go-ai v1.1.0 (`f44bf37`), fixtures-vibes
`0259e9a4816d3d643159538095fdffc884104598`. Recheck current refs before work;
do not roll back later published fixes. These items are refined proposals,
not assignments. No agents have been dispatched.

## Parallel lanes

| Item | Owner role | Parent | Estimate / risk | Readiness | Dependency |
|---|---|---|---|---|---|
| N1 Recorded execution duration | Gi native | Pi 1.1.0 alignment | M / medium | 9/10 | None |
| F1 Editor revision forwarding | Shared frontend | #46 | S / medium | 10/10 | Consumer acceptance after published pin |
| F2 WebKit terminal reattach | Shared frontend | #45 | M / medium | 9/10 | Separate acceptance after published pin |
| F3 Web codemode cards | Shared frontend | #30 | M / medium | 9/10 | Freeze event/detail contract before coding |
| A1 VNC acceptance reconciliation | Gi acceptance | #47 | S / low | 8/10 | Decide scope of historical preview008 reference |
| C1 Missing command contracts | Gi design | #16 | S / medium | 7/10 | Rui approval before implementation |
| R1 MCP umbrella reconciliation | Gi review | #25 | S / low | 9/10 | Closure waits for F3 and missing acceptance |
| R2 Mechanical split proposal | Gi architecture | #11 | S / low for proposal | 8/10 | Implementation waits for native lanes to settle |

Scores assess clarity/scope/tests/dependencies/risks; they do not measure
completion. N1 and C1/R1/R2 proposal work can run beside one frontend owner.
F1/F2/F3 touch the same shared frontend/build outputs: separate branches and
serial integration, or one owner doing these in order. Do not run independent
writers against the same checkout. Maximum three active implementation lanes.

## Common assignment contract

Use a named feature branch/worktree with exclusive ownership of listed paths.
Do not rebase, deploy, contact another agent, close issues, change capabilities
or remove skips without explicit authorisation. Preserve unrelated Joker edits,
active work and durable runtime state. Shared UI changes belong in
`rcarmo/fixtures-vibes/ui/classic`; Gi owns native APIs, its consumer tests and
submodule adoption. Source/oracle assertions cannot be silently rewritten to
turn a failing run green.

All code items need a red/green regression, bounded failure/cancellation cases,
relevant integration checks, docs and exact commit/pin/counts. Use repository
Make workflows sequentially with portable project-owned scratch. Pre-release
checks inspect CPU, alloc_space and alloc_objects (plus browser allocation/CPU
where supported), then immediately dispose used profiles, matching binaries,
traces, fixture state and logs. WebKit browser results are functional-only.
No performance claim without equivalent workloads. New exported TypeScript
symbols need JSDoc. Publish only verified changes with Rui's Git identity and
the approved push helper; merge authorised branches without rebase.

Return: scope/files, commit and parent, test commands/counts, profile conclusions
and limitations, disposal confirmation, remaining blockers, and next consumer
acceptance command. Consumer acceptance must be independent of isolated mocks.

## N1: Recorded tool execution duration

Outcome: live and resumed shell/tool results show the same measured execution
duration. Clock changes, event delivery delay and reload must not change Took.
Assistant response duration from go-ai is a different measurement.

Paths: `internal/turn` tool execution/publication, persisted message/event
payloads in `internal/store`, `internal/tui/chat.go`, `tool_render.go` and
associated tests. First inspect `store/tool_activity.go` and
`turn/tool_terminal_test.go`; occurrence/call identity and interruption fencing
already exist and must be preserved. Avoid a broad chat.go extraction.

Recommended: persist measured duration at the actual execution boundary, reuse
existing occurrence identity, carry it through final events and tool-result
messages, and prefer that optional field in live/reloaded rendering. Alternative:
reconstruct from persisted execution events only; reject it if it requires
ambiguous call matching or wall-clock subtraction. Old sessions lacking the
measurement display unknown, not invented 0.0s; recorded zero is valid.

Acceptance: success/error/cancellation have truthful timing; missing terminal
boundaries remain unknown; duplicate/late events cannot mutate settled timing;
identical tool names/call IDs across turns cannot cross-match; replay/reload
preserves duration; running Elapsed never becomes final Took prematurely.
Nested codemode rows retain their own durations. No output or tool replay.

Tests: focused turn/store/TUI race repetitions, independent fake clock or
controlled elapsed measurement, live event versus restored-message rendering,
two occurrences, end-without-start, interruption and legacy payloads; PTY Took
visibility if available. Full native regression/vet/CGO-disabled build before
publication. Done when contracts, tests and current parity docs agree. F3 may
consume a new optional field only after its shape is published; F3 need not
wait for timing to render the existing nested-call details.

## F1: Forward editor revision through the lazy loader (#46)

Outcome: clean workspace-update SSE refresh retains the complete loaded
revision, so typing after refresh remains conditionally saveable.

Upstream paths: `scripts/piclaw-web.mjs`, `scripts/patch-editor-revision.mjs`,
editor-loader adapter and mounted/unit tests. Gi consumer files:
`tests/consumer/editor-refresh.spec.ts`, `editor-conflict.spec.ts` and
`playwright.consumer.config.ts`. Keep upstream vendored Piclaw bytes pristine;
prefer the existing validated build adapter. Alternative: extend the shared
PaneInstance typing/proxy contract, with all callers checked for compatibility.

Root cause: orchestration calls setContent(text, mtime, revision), but
LazyEditorInstance forwards only text/mtime. The real editor correctly clears
loadedRevision on a missing third argument; Save is disabled and Save Copy
silently returns. Preserve that fail-closed guard. Do not fetch a fresh
revision at save time or add unconditional retries.

Acceptance: initial load, clean SSE refresh, lazy proxy and transferred editor
retain the snapshot revision; partial/missing revisions lock out writes;
newer typing during fetch is preserved. Review Cancel sends no PUT; approval
writes exactly the reviewed revision, a subsequent remote write yields 409,
and Save Copy is create-only without clearing the source draft. No implicit
Overwrite approval.

Shared workspace018's earlier Reload diagnosis is corrected: Reload passes;
the old scenario never approves the new Review overwrite modal. Record that
contract mismatch separately. Any upstream scenario addition must retain
review/cancel/conflict assertions and provenance; no Gi-only edit of shared
specs or oracle data.

Upstream tests: actual lazy loader plus mounted editor/SSE path, not only direct
StandaloneEditorInstance calls. Handoff includes build/TDZ/unit/contract checks.
Gi acceptance after final pin: `make PROFILING=1 test-editor-conflict-consumer`
(current four workloads: one pass/three failures), then affected shared editor
checks. #46 closure requires its broader editor/popout/Vim gates too; fixing
this regression alone is insufficient.

## F2: WebKit terminal popout reattach (#45)

Outcome: closing/reattaching a terminal popout restores its original terminal
tab and usable existing PTY connection without losing unrelated tabs/drafts.

Upstream scope: terminal-pane, pane-popout/live-transfer/tab orchestration and
focused tests under ui/classic. Preserve backend owner/target handoff tokens,
replay protection and cleanup. No VNC, editor contract or terminal prompt changes.
Recommended: identify the lost tab/transfer state in the close/reattach lifecycle
and fence stale windows. Alternative: explicit host transfer acknowledgement
only if existing lifecycle cannot express completion; document any new API first.

Acceptance: shared `@ux-terminal-006` passes WebKit desktop three consecutive
repetitions with zero retries (current reproduced failure 3/3); Chromium desktop
regression passes; original tab identity, PTY process/connection, command output
and draft survive. Late close from an obsolete window cannot remove a newer
host; shutdown leaves no workers/popouts. Preserve disabled touch-device cases.

Use `fixtures-vibes-focused` with
`--grep @ux-terminal-006 --project webkit-desktop --repeat-each 3`, then affected
terminal matrix after published upstream pin. Consumer profile conclusions and
safe cleanup required. Independent acceptance owns issue/skip decisions.

## F3: Web codemode rendering (#30)

Outcome: running and persisted codemode calls render highlighted script,
nested call states and output with truthful expansion/truncation behaviour.
TUI implementation already exists in `4f73b964`; do not redo it.

Owner: fixtures-vibes Gi adapters/host renderers and browser tests. Read Gi's
`internal/turn/codemode_tool.go`, `internal/tui/codemode_render.go`,
`docs/internal/codemode.md`, the real SSE translator and persisted tool-result
shape before changes. Running nested events carry parent_tool_call_id; settled
result details carry calls/fullOutputPath. Freeze a small shape table first.

Recommended: Gi-specific renderer/adapter over unchanged Piclaw components.
Alternative: generic shared tool-renderer hook if a reusable hook already
exists; no fork of vendored components solely for this item.

Acceptance: script/code highlighting, running/ok/error/cancelled nested rows,
final details replacing live rows, reload without server availability, stale
parent/turn/session fencing, compact default plus expansion, truncation link
through authenticated output lookup, escaped hostile output and safe links.
Missing details degrade visibly without inventing calls or timing. No generated
output hash gate, arbitrary filesystem path URL, global mutable renderer state
or duplicated output. Preserve streaming/thought/finished-tool boundaries.

Tests: mounted live SSE and reload cases using actual Gi fake-tool/codemode
runs, bad/missing details, tool errors, interrupted parent, two sessions, narrow
width and six browser/device projects where applicable. Isolated renderer
fixtures alone cannot close #30. Consumer adoption runs after upstream final
pin; integrate separately from F1/F2.

## A1: VNC issue/skip reconciliation (#47)

Outcome: decide the exact remaining VNC gate using already verified results;
do not rerun the whole VNC matrix merely to refresh counts.

Known passes: live RFB policy/read-only/interactive/empty matrix 18/18, shared
error/retry 6/6 and shell menu/tab-close 12/12, six native checks x3 race.
Historical issue cites preview008, which is general workspace preview:
Chromium passes 3/3; WebKit has a shared Linux upload-bytes limitation.

Recommended: audit feature tags/current scope and propose separating generic
preview from VNC acceptance, with documented source limitation. Alternative:
retain preview008 as a closure dependency and build an independently valid
WebKit file-placement check without weakening assertions. Rui must approve
changing acceptance scope or removing skips. Source skip is not a passing test.

Acceptance: one table maps issue requirements to exact pins/tests/limitations,
read-only is explicitly UI suppression (raw proxy does not police hostile
RFB clients), untested surfaces are named, and proposed capability/skip edits
are minimal reviewed diffs. Do not claim VNC protocol/security guarantees from
banner echo tests. Documentation/report-only assignment initially; no closure.

## C1: Define /changelog, /bug and /trust (#16)

Outcome: give Rui three independent command contracts to approve.
Recommended proposals: /changelog reads a packaged release-note source without
network access; /bug generates a local, opt-in redacted diagnostic report and
never uploads automatically; /trust stays omitted until Gi has a meaningful
project admission/resource-loading policy. Alternatives: open release URLs for
changelog, or retain omissions for all three with documented reasons.

Inspect `internal/tui/chat.go` command catalogue/dispatch, config/resource
loading and debug/export code. Deliver exact command arguments, outputs,
failure cases and UI, data included/excluded, permission/path limits and tests.
A label or blanket "trusted" state without enforcement is unacceptable.
No diagnostics may expose keys/tokens, raw prompts/tool results or home paths by
default. Resource trust must cover actual hooks/extensions/MCP and be revocable.

This item is design-ready, implementation blocked on Rui's decisions. After
approval split commands into separate S/M implementation tickets; test offline
missing notes, denied paths, bounded diagnostic size/redaction and cancellation.
Do not mix command code with the timing or restructuring lane.

## R1: Reconcile MCP/codemode umbrella (#25)

Outcome: update stale checklist claims to match published contracts. #27/#28/#29
are closed; TUI #30 work is done, web #30 is open. Whole-Joker WASI is retired;
native Joker/functions compiled to WASM are the current boundary.

Read issue history, `../plans/mcp-codemode-plan.md`, current scripting docs and #30 tests.
Recommended: append a precise current checklist and archive superseded roadmap
text with dates. Alternative: replace obsolete body sections while preserving
history through links. No code changes or automatic closure.

Acceptance: every checked row has a current source/commit/test reference;
unchecked rows point to bounded work; remaining hostile-script/fake-server,
PTY/web and cancellation coverage is listed without blanket pass claims.
Final closure waits for F3 and any genuinely missing acceptance. This review
can run now without touching native/UI implementation paths.

## R2: Plan one mechanical extraction (#11)

Outcome: propose one behaviour-preserving extraction with exclusive file
ownership and a merge window, rather than refactor several large files during
feature ports. Start with a self-contained helper subsystem, not turn execution.

Inspect current chat.go/engine.go/server.go sizes and dependencies. Recommended:
move a cohesive helper plus matching tests inside its existing package, leaving
public APIs/import boundaries unchanged. Alternative: postpone until N1/F3
interfaces settle; cross-package restructuring needs a separate design.

Deliver source/destination symbol list, collision map with N1/F3/C1, baseline
commands and proof that no behaviour/security/numerical policy changes are
needed. Implement only after approved scope/merge window. Existing #11 requires
native/vet/functional/PTY checks per slice; if a helper is absent report the gap
instead of counting it as pass. No giant reformatting or dependency bump.

## Later Pi 1.1.0 tickets

Keep these separate from N1: outputPad propagation (0/1 setting through tool,
shell and summary rendering); OSC7501 negotiated metadata-only status with
opt-in/opt-out and lifecycle tests; live MCP action menus/whole-chain OAuth
cancellation; split ANSI chunks/codemode output separation; +/- tool-loadout
inheritance. Each needs its own reference contract and regression. None is
implicitly assigned by this packet.

## Refinement record

2026-10-08: inspected all seven open issue bodies plus current implementation
and acceptance notes. Existing issue text includes stale editor-open claims;
F1 reflects published backend/SSE work and the actual remaining loader defect.
No implementation or tests ran during refinement. Scores are proposal readiness;
shared-contract decisions and consumer publication are explicit dependencies.
