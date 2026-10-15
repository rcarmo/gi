# Remaining lane decisions, 8 October 2026

N1 is published at `2a5e247`; A1, C1, R1 and R2 below are bounded review/design
deliverables. No acceptance scope, capability, skip or command has changed.
Frontend implementation F1/F2/F3 still requires an assigned shared owner and
upstream publication/adoption. Fixtures pin is `0259e9a`; frozen results retain
their failed gate. This review uses existing evidence, not new browser tests.

## A1 — VNC acceptance (#47)

| Requirement | Source/evidence | Result and limit |
|---|---|---|
| Allowlisted native bridge, owner/handoff/auth cleanup | backend `79fd83b`; current native VNC race checks in `c262b2e` | Six checks x3; no raw-RFB read-only policing |
| Readonly input suppression with interactive positive control | `tests/functional/24-vnc-viewer.spec.ts`, `make test-vnc-viewer`, acceptance-2026-10-08.md | 18/18 across six browser/device projects; pixels and counted key/pointer/clipboard frames |
| Retry/error/empty behaviour | shared workspace015 and live empty fixture at fixtures0259e9a | Shared error/retry6/6; live matrix includes empty state |
| Open menu and tab close | shared workspace-menu001/tab-close001, c262b2e | 12/12 across projects |
| Historical workspace008 dependency | generic upload/preview scenario, unchanged assertions at fixtures0259e9a | Chromium3/3; WebKit Linux upload-bytes limitation, not a pass |

Recommend removing workspace008 from the *VNC closure dependency*, while
retaining it in general preview acceptance with its WebKit limitation. It tests
generic file upload/preview rather than the VNC transport/policy. Alternative:
keep #47 open until an independently valid WebKit file-placement workload proves
the same preview assertions. Rui must approve this dependency decision first.

Proposed initial diff: change only #47's required-fixture wording to
workspace015 plus live RFB/menu/tab-close evidence; add a separate generic
preview dependency note. Capability/skip diff: **none proposed for automatic
application**. Inspect the current VNC feature/profile gates before proposing
any skip deletion. Keep authenticated output, owner isolation and hostile-client
limits explicit. Physical remote desktop quality/latency, malicious raw clients
and exhaustive real-server interoperability are unverified.

## C1 — Missing command contracts (#16)

### /changelog [version]

Recommended: display packaged release notes for the exact installed version,
or an explicitly named bundled version; no network fetch or arbitrary path.
No arguments selects installed version. Invalid/unknown version returns a plain
"No bundled notes for <version>" message without changing session/draft.
Packaged source: a version-indexed release-note asset generated at release,
not an inferred Git diff or shell command. Maximum displayed source 64 KiB;
use existing transcript expansion/scrolling. No persistence to model context.

Alternative: show a release URL, with browser opening only after confirmation.
Decision needed: package a note asset, link only, or keep command omitted.
Tests after approval: installed/explicit/missing/malformed version, offline use,
bounds, draft/cursor preservation and no model/tool invocation.

### /bug [preview|save]

Recommended default `/bug` equals `/bug preview`: generate a local report
preview containing Gi version/build, Go/runtime/OS/architecture, frontend pin
when known, named feature flags and coarse error categories. Exclude prompts,
conversation/tool output, credentials, provider tokens, endpoint query strings,
paths/usernames/hostnames, database content and environment enumeration.
Redaction is an allowlist of fields, not regex scrubbing of a full dump.

`/bug save` first previews the exact bounded payload (maximum 64 KiB), then
requires confirmation of an application-owned create-only report destination.
Cancel writes nothing. Denied destination/symlink/conflict/size failure reports
an error without fallback, overwrite or upload. No browser/network action,
issue creation, automatic attachment, session export or command execution.
Decision needed: preview-only first slice, or confirmed local save too.
Tests: secret sentinel exclusion, bounded fields, invalid options, no network,
create-only confinement/symlink rejection, cancellation and draft preservation.

### /trust

Recommend keeping it omitted in this iteration. Gi's MCP config reader accepts
`projectTrusted`, but current callers pass false; hooks/extensions/resource
loading need a common admission policy before a meaningful trust toggle.
A label-only trust command could imply enforcement that does not exist.

If Rui chooses implementation, first define workspace identity (resolved root,
symlink/rename/replacement handling), the exact resources gated, global versus
project precedence, explicit permission categories, revocation and persisted
consent version. `/trust status` is read-only; grant/revoke require reviewed
resource lists and confirmation. No automatic startup trust, inferred trust
from Git origin, blanket enabling of MCP/tools or disabling existing user policy.
Revocation cancels owned pending resource admission and prevents new loads;
already-running external processes require a separate safe stop contract.
Tests must cover each actual loader, sibling roots/symlinks, replaced workspaces,
restart, revoke/timeout/cancel and unchanged global resources. No implementation
until the enforcement surface and failure behaviour are approved.

## R1 — MCP/codemode umbrella (#25)

The 1 October plan's "Nothing here is shipped" banner is historical. Current
contracts are `../../internal/mcp.md`, `../../internal/codemode.md`, `../../internal/codemode-scripts.md`, scripting docs and
published tests. Reconcile the umbrella as follows:

| Row | Current state | Evidence / remaining gate |
|---|---|---|
| Config/core/tool exposure | Implemented | `internal/mcp/config.go`, manager/result tests; #26 config precedence closed; project configs still require trust |
| OAuth/discovery/sign-in UI | Implemented bounded workflows | #28 closed; `71b2a88`, `477a7ab`, OAuth/CLI/TUI tests; whole-login deadline/live action refinement stays separate Pi1.1 work |
| Provider-backed HTTP auth | Implemented | #29 closed, `50f62bf`; provider_auth tests; no new live-provider acceptance in this review |
| Codemode engine/model declarations/toggles | Implemented | internal/turn/codemode_tool.go and codemode tests; #27 models API closed; JS QuickJS/Goja, native Joker and supported kernel compilation per current scripting docs |
| Whole-Joker WASI runtime | Retired | `0bee630`; remove as future completion requirement, preserve dated history |
| TUI codemode cards | Implemented | `4f73b964`, codemode_render_test.go; N1 adds measured nested events |
| Web codemode cards | Open | F3/#30, real live/reload acceptance required |
| Aggregate hostile/cancellation/browser/PTY acceptance | Partially evidenced | Link each concrete test; do not infer full acceptance from adapters or closed subissues |

Proposed issue update: split #30's checked TUI row from unchecked web row, mark
#27/#28/#29 by their actual published evidence, archive whole-WASI text and keep
a named remaining-acceptance checklist. No umbrella closure. GitHub readback confirms #25 open and #26/#27/#28/#29 closed. Historical
checkbox provenance still needs row-by-row source links before issue-body
replacement; no new tests ran here.

## R2 — Mechanical extraction proposal (#11)

Recommended first slice after an approved merge window: move only
`toolInvocationText` and `toolOutputBodyLines` from `internal/tui/chat.go` into
`internal/tui/tool_text.go`, same package and unchanged signatures/bodies.
`toolInvocationBody` and `toolResultBody` stay in chat.go because they use chat
state/storage. Add focused helper tests for precedence, empty/non-map arguments,
JSON fallback, whitespace, line-ending normalisation and literal output. Existing
rendering tests continue to exercise callers. No imports/public APIs/wire schema,
TUI layout, security or numerical behaviour changes.

Collision map: N1 modified surrounding tool event code but is now published;
F3 consumes these semantics through frontend adapters and owns no Go source;
C1 future dispatch work touches chat.go, so integrate this extraction before
command implementation or pick a separate merge window. Keep timing helper,
store projection and runtime engine untouched. Baseline sizes inspected:
chat.go 6,050 lines, engine.go 5,829, web/server.go 1,241 before this proposal.

Verification: `make test TEST_PKGS=./internal/tui`, focused race repetitions,
`make vet`, `make test-ux`, `make test-tui-smoke` and affected regular/fullscreen
PTY targets. Use documented isolated env/scratch and pre-release profiling;
report missing targets/host coverage rather than treating them as passes.
Alternative: postpone extraction until command contracts settle. This proposal
needs scope/merge-window approval; no mechanical changes have been made.

## Follow-ups and handoff

Separate future tickets: adopted scenario-retirement review, outputPad,
OSC7501 metadata negotiation, live MCP action/whole-login cancellation,
ANSI chunk/codemode output boundaries and tool-loadout +/- inheritance.
No frozen acceptance rerun or automatic exclusions. F1/F2/F3 share build paths:
serial integration through exclusive feature branches. N1 source/verification
is in tool-duration.md; this review adds no runtime or frontend changes.
