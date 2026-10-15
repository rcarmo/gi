# UX re-audit: maintenance checkpoint, not a completed audit

**2026-09-27 correction:** this audit failed to capture the chat lifecycle.
The user screenshot and read-only Gi probes disprove chat UX acceptance.
The tool-footer mapping is withdrawn; historical counts below are superseded
for current acceptance. See [chat UX recovery](../chat-ux-recovery-2026-09-27.md).

User direction: audit all Gherkin for correct basic UX interactions, then cascade
corrections to tests and code. Use Piclaw's actual behaviour and source as the oracle.
Do not make the requirements merely describe current Gi code. Record the Piclaw
revision and environment for each comparison, observed interaction and source
provenance, expected test assertion, Gi handler and any deliberate safety deviation.
Conflicting historical contracts are preserved byte-for-byte; the active 3.2.4 edits and evidence limits are recorded in `gherkin-alignment.md`.

## Baseline and inventory

Initial clean source baseline: `5dc6bf77968a1c6da1c14292d2a894085012f7b1`.
Its 37 tracked feature files, 358 scenario definitions and 421 expanded cases parsed
successfully with Cucumber. This baseline is a historical snapshot; the separate oracle-only
feature, corrected Gi settings scenarios and active Piclaw 3.2.4 edits came later. The historical Classic/shared hashes are verified under `tests/ux/upstream/`. A current staged-feature parse on `a6d6873` plus the active working tree has 42 files, 380 definitions and 444 expanded cases, including five Gi-specific deviation clauses and the two-example active reconnect scenario. The previous 443-case parse is a pre-outline checkpoint. Includes
native TUI/search, additive Gi settings/session specs, passkey additions and all
Classic/shared contracts—not only the web mapping catalogue.

`scenario-inventory.json` preserves the initial 358-definition inventory, including Background, steps, outline rows
and expanded count. Test candidates are lexical links, NOT verified coverage.
The test index cannot resolve every dynamic tag/title and test factory; a missing
candidate is not proof that no test exists. The 14 qualified findings and 58
group-reviewed Classic/Shared definitions remain short of complete current-oracle
review. These artifacts do not supersede prior accepted evidence.

Four broad and two smaller delegate attempts timed out without usable completed
reviews. An empty partial JSON was excluded. Their work earns no review credit.
The lead read the whole compact scenario corpus; clause-by-clause test/code/oracle
verification remains unfinished. The focused passes below add evidence only for
their stated interactions; they add no general parity credit.

## Work after maintenance clearance

Maintenance hold was explicitly removed; no restart was requested. Piclaw's
installed release changed during the audit from 3.2.3 to **3.2.4**; the current
oracle is asset version `990f0c49a932` with the source-map hash pinned in
`tests/ux/oracle/piclaw-3.2.4-reference.json`. Relevant source-map modules are
byte-identical to the removed 3.2.3 release. `make test-piclaw-oracle-basic`
passed against 3.2.4 with hash-checked UI assets, the production-default SVG
flag substitution and isolated API fixtures. `make test-piclaw-oracle-matrix`
passed six cases—including first selected-chat Return request, fixture-listed
skill prefill and model-picker keys—across six Chromium/WebKit viewport projects. See
`oracle-deltas.md` and
`tests/ux/features/oracle/piclaw-3.2.4-basic-interactions.feature`. The older
3.2.3 reference is historical. These probes are **not Gi parity tests**.

`scenario-inventory.json` contains all 358 baseline definitions. Each has a
qualified clause finding or source→test→Gi-code review; none awaits initial
clause audit. Findings include unsupported controls and unresolved policy
conflicts, so this count is not a parity pass. `scenario-status.csv` is the
initial status snapshot; its audit statuses were not regenerated after later
tranche reviews. Historical Classic canonical 001–029 and Shared 1–42 also
have qualified **group** reviews in `classic-canonical-review.md` and
`shared-contract-review.md`. `compose-stability-review.md` traces five
additional frozen compose clauses through current Piclaw 3.2.4 source, tagged
Gi assertions and code; their Piclaw upload/failure backend was not run.
`workspace-bounded-review.md` adds hidden-tree, preview and tab-store traces
for three more clauses; their current Piclaw UI interactions were not run.
`recovery-card-review.md` adds three source-backed, tagged native-display
journeys. The Piclaw card/recovery UI itself was not exercised. These reviews
do not convert unprobed clauses to current-oracle acceptance.
`expanded-crosswalk.json` enumerates all 421 cases,
including the 56-case pinned passkey criteria ledger. It finds 117 direct-tagged
browser-test candidates; 248 have neither such a candidate nor passkey-ledger
entry. **A direct tag is not behavioural acceptance**, and zero direct candidates
does not establish absence of another test. The crosswalk is a
triage aid, not a completed UX review.

The additive `@gi-settings-004` text was corrected to reflect advertised thinking
support, with `@gi-settings-028` and `@gi-settings-029` scenario clauses. A new
six-project native browser test for model-switch reset passed; the earlier supported
thinking provider-request test is tagged. The broader thinking run was interrupted
by the command harness after progress through 32/42 tests; its result is unknown.

A test-only commit strengthened three functional smoke assertions: visible
meters, session/turn-specific SSE frames after composer submission, and an
uncaught browser-error POST. Their focused tests passed, and the
full functional suite passed 139 cases with 11 existing skips. The separate basic
HTTP run had 31/32 passes, but Chromium initially stayed on Loading Gi because
static requests failed with `net::ERR_NETWORK_CHANGED` before app boot.
An unchanged rerun was interrupted at 21/32 by `make: wait: No child processes`.
Neither run counts as a green gate. Traces and logs are under the local audit
workspace; no assertion, timeout or production code was changed for that failure.

`tests/features/sessions/gi-basic-send.feature` adds two derived Gi-native
requirements without touching the frozen corpus. Their existing exact
session/turn/reload and rejected-admission journeys are now tagged; 12 focused
browser cases passed. The Piclaw oracle's first-send fixture proves only one
selected-chat POST and composer clearing, not a native reply.

The installed Piclaw Plan Sidebar add-on 0.1.25 has four source-reviewed,
parse-checked versioned Gherkin cases at
`tests/ux/features/oracle/piclaw-3.2.4-plan-sidebar.feature` (commit `3ef30aa`).
A shipped-Classic/add-on browser fixture in Chromium and WebKit desktop
exercised only same-chat remote-update warning followed by explicit Refresh:
unsaved text survived the remote event, then Refresh replaced it without a
discard confirmation. Save, Submit and Plan tool clauses remain source-only.
None is Gi parity; see `plan-oracle-gap.md`. Gi's implementation target keeps
Shared revision CAS and dirty-refresh confirmation as a no-loss deviation
from the installed add-on.

Five tagged compose-stability cases (`@ux-compose-001`, `002`, `003`, `005`,
`006`) passed 30/30 in the canonical disposable six-project browser fixture.
The first unchanged run passed 29 and stalled on WebKit phone after reload at
`Loading Gi…`; the unchanged rerun passed. Its trace was overwritten; cause
unknown. `loading-stall-investigation.md` records a later exact-case 4/5
WebKit-phone repeat with a separate `page.reload()` internal navigation error
and retained trace. No assertion or timeout was relaxed. The sixth clause,
queued return `004`, remains uncredited: accepted ADR-0017 chooses Gi's no-loss
recovery over Classic draft replacement.

`workspace-index-derived-review.md` traces the 23 derived index proposals to
mounted scanner, store, worker, scheduler, HTTP and TUI assertions. Accepted
Gi ADR-0046/0050/0051 make reads and native writes explicit-refresh only.
A pinned Piclaw 3.2.4 in-memory search-function probe returned committed hits
and intercepted a background request on cold/stale searches; it launched no
worker. Gi's unimplemented `:33` trigger needs a deliberate policy change;
`:199` and `:211` have bounded explicit-only Gi findings.
There is no complete derived-index acceptance.
`tui-remaining-review.md` traces the final nine native TUI clauses against a
fresh seven-feature disposable tmux run. Screen substring/SQLite checks are not
physical-terminal, accessibility or Piclaw parity acceptance.
`passkey-additions-review.md` traces all 26 additive passkey clauses through
the pinned criterion ledger, bounded assertions and mounted Gi paths. Its 15
candidate, eight partial, two manual and one unsupported findings add no
formal scenario mappings, Visual/physical-device acceptance or current Piclaw
runtime comparison.
`gi-settings-native-review.md` traces the 27 Gi-specific baseline Settings
definitions (including one unimplemented OAuth proposal) through scoped native
browser and auth tests. Post-baseline thinking cases 028–029 remain outside the
358-definition ledger; none grants Piclaw Classic/Shared parity.
`classic-core-settings-review.md` classifies all 32 Classic Settings clauses
against Gi's six mounted sections. Three directly tagged shell cases passed
within a 24/24 run; the remaining clauses have qualified partial or missing
section findings, not Settings parity.
Shared40 is inventoried as a native accessible tool-pane gap: Gi's status
row and the unapproved tool-terminal WIP do not provide the persisted pane
and focus/reduced-motion journey in `shared-contract-review.md`.
`shared-reconnect-content-review.md` traces Shared36–39/42 through five
separate six-project native browser runs; Classic cascade/all-chat and
physical speech limits remain distinct.
`shared-model-review.md` traces Shared31–35 through scoped picker/model/thinking
browser tests (12+6+6+6); current Piclaw backend and physical keys unprobed.
`shared-queue-review.md` traces Shared27/29 (12/12) and Shared30 (6/6 after
a test-only stale model-picker locator fix). Classic idle-Steer remains a
separate policy conflict.
`shared-session-review.md` traces Shared23–26 through picker focus, coherent
cross-session refresh and native mutations (18+6 focused passes); current
Piclaw UI remains unprobed.
Shared Plan definitions at lines 117, 131, 141 and 150 are inventoried
against `plan-oracle-gap.md`: Gi has no mounted Plan workflow. Shared20's
confirm-before-discard behavior is the Gi target despite the installed add-on's
unconfirmed explicit Refresh; it earns no parity credit.
`shared-shell-quick-actions-review.md` traces six shared definitions/first
fifteen cases: 14 cases passed across six projects (84/84); Shared8 still
lacks a real contenteditable editor journey.
`classic-copy-delete-retrieval-gap.md` records Classic024's missing cascade
path and Classic025's all-chat/family scope gap. Direct-delete and Shared38
current-session acceptance do not cover those combined contracts.
`classic-copy-speech-review.md` traces Classic028's combined code-copy and
speech-owner journey (6/6 with stubbed synthesis); audible output remains
unprobed.
`classic-tool-status-review.md` traces Classic027 through a gated native
tool-activity browser journey (6/6); full pane, reduced-motion and WIP
terminal provenance acceptance remain separate.
`classic-upload-review.md` traces Classic026 through a native multipart
failure, retained draft and explicit retry (6/6); Piclaw backend and media
cleanup/deduplication remain unprobed.
`classic-reconnect-stop-review.md` traces Classic023 through a real-SSE
Gi reconnect and exact-turn Stop journey (6/6). Shipped Piclaw 3.2.4 Classic
source and frozen `@ux-reconnect-004` both require manual reload; shipped-asset
browser warning passed Chromium/WebKit fixture probes, while current-backend
acceptance remains unverified.
`classic-model-review.md` traces Classic020/022 through focused Gi model
selection and sparse-metadata tests (12/12); Piclaw backend switching remains
unprobed.
`classic-queue-review.md` traces Classic016/018 through focused native queue
tests (12/12) and records Classic019's idle-Steer policy conflict with Gi's
matching-active-turn requirement.
`classic-picker-review.md` traces Classic013–015 through focused Gi session
picker selection/actions (18/18); current Piclaw browser and physical keyboard
acceptance were not run.
Classic009–012 are now inventoried against `plan-oracle-gap.md`: the
installed Piclaw Plan add-on is source-reviewed and Gi has no mounted Plan
store/tool/sidebar. Piclaw browser execution and Shared revision policy remain
open.
`classic-shell-quick-actions-review.md` traces Classic001–006: five tagged
menu/Quick Actions cases passed 30/30; the six-example excluded-target outline
004 still lacks a real-control journey for every surface.
`terminal-web-gap.md` marks seventeen Classic terminal/dock/zen scenarios as
mounted Gi web capability gaps: the read-only host rejects terminal panes and
terminal/VNC callbacks are no-op. Gi TUI evidence is not web-terminal credit.
`workspace-remaining-review.md` classifies fifteen additional Classic workspace
clauses against Gi's native GET-only file API and read-only preview host;
copied CRUD/editor/terminal/VNC handlers cannot earn mounted capability credit.
`auth-remaining-review.md` classifies eleven additional Classic auth clauses:
passkey-policy and single-user logout analogies remain partial, copied OOBE is
unmounted, and invitation/family/ambient-passkey flows are native gaps.
`core-interactions-gap.md` distinguishes five additional Classic BTW, card,
widget and notification clauses: mounted gaps, source-only evidence, and an
untagged 10-helper/24-browser notification run. It does not rescore the three
previously reviewed card/recovery clauses.
`annotation-highlights-gap.md` records twelve Classic image/text annotation
clauses: eleven native capability gaps and one partial non-iPad lightbox
journey. Metadata badges and search matches do not count as saved annotations.
`settings-layering-review.md` traces four Classic overlay clauses through
Gi's mounted portal/CSS and pointer/geometry browser assertions (24/24), not
current Piclaw UI or physical-device acceptance.
`settings-dialog-review.md` traces five Classic dialog shell clauses through
Gi cache/lazy loading and focused native tests (30/30), not all Settings panes
or current Piclaw UI acceptance.
`timeline-rendering-review.md` traces six table, code-copy, isolated-link,
outcome and speech clauses through focused native tests (12+6+6+12 passes);
Piclaw UI and physical speech playback were not checked.
`accepted-visibility-review.md` traces five accepted-send/refresh clauses
through Gi native upload, reference serialization, newer-draft preservation
and timeline assertions (30/30); Piclaw backend acceptance was not run.
`message-deletion-review.md` separates one direct-delete native journey
(6/6) from five unsupported reply/cascade clauses; Gi's server rejects
cascade deletion. An isolated installed-Piclaw 3.2.4 backend-function probe
returned 200 for direct deletion of a parent with an unseen reply, leaving an
orphan; cascade removed a parent and three replies. Frozen `018`/`019` require
a `Replies exist` direct-delete rejection, which this backend did not emit.
Independent shipped-UI disposable-fixture checks passed Chromium/WebKit
desktop visible-reply prompt/cancel/cascade and synthetic `Replies exist`
retry/cancel branches. A joined shipped-UI/installed-backend-function
in-memory fixture in both browsers deletes a parent with an unseen stored
reply without prompting and leaves that reply orphaned. With three visible
stored replies, joined confirmation deletes all four rows; cancellation keeps
them. The production Piclaw HTTP router/authentication and live acceptance
were not exercised.
`shell-layout-review.md` traces nine Classic menu/layout clauses: four tagged
native cases passed 24/24, while hamburger New file dispatch and terminal/VNC
callbacks remain native gaps; safe-area, scale and inline-code checks are bounded.
`reconnect-version-review.md` corrects an earlier source-provenance error:
the installed 3.2.4 Classic source map disables auto-reload; a divergent
checkout and copied unmounted Gi helper still schedule it after 350 ms. The
frozen clean-state manual-reload clause agrees with the shipped source.
Filtered Gi runs passed 6/6 and 30/30. A disposable shipped-asset browser
probe confirmed SSE delivery, one manual warning and no auto-navigation in
Chromium and WebKit. The broader reconnect run timed out at 63/72 and is not
a pass; current Piclaw backend and deployed Gi were not tested.
`theme-command-gap.md` records fifteen Classic composer `/theme`/`/tint`
scenarios unsupported by Gi's native command path. An installed 3.2.4
parser-only probe passed nine bounded command cases. A shipped-UI
Chromium/WebKit desktop fixture sent `/theme ristretto` and `/tint #e11d48`,
observed local theme/tint state and a timeline response; backend persistence,
reload and pixel colours were not tested. Browser-local Gi Appearance
Settings cannot earn the command/legacy-storage credit.
`context-compaction-review.md` traces thirteen context/compaction/model
clauses through current source and focused Gi native tests; the broad
compaction run was interrupted at 91/102 and is not a pass.
`thoughts-panel-review.md` traces five frozen disclosure clauses through the
installed status component and native streamed tests. The first focused run
was aborted at 28/30; an unchanged rerun passed 30/30. Piclaw UI was not run.
`lightbox-dismissal-review.md` traces four stored-image lightbox clauses. Gi
browser tests passed 24/24, including trusted emulated touch; Piclaw UI and
physical-device acceptance were not run.
`gi-preview-swipe-review.md` covers six Gi-only preview/swipe clauses:
three preview tags passed 18/18 on a scoped rerun, rapid reverse swipe 6/6
and status-panel variants 12/12. A broad thoughts run failed 1/48 during
`Loading Gi…`; its cause remains unknown. These are not Piclaw oracle claims.
`auth-basic-review.md` distinguishes the missing family username/TOTP branch
from two single-user native auth journeys (12/12 focused browser cases); the
full auth run was interrupted at 105/120 and has no green result.
`editor-gap-review.md` identifies five frozen editable-editor journeys with
no Gi editor or tagged native acceptance. Read-only preview tabs are not an
editor. `pwa-manifest-review.md` records six icon clauses. Gi's static manifest and
Apple/favicon routes passed 12 browser cases; avatar PNG/versioned paths in
Piclaw have no Gi counterpart or acceptance test.
`mobile-swipe-review.md` qualifies six mobile definitions. Four session-tagged
IDs passed 24/24; status-panel gestures passed 12/12 (six mobile, six Gi
variants). Two tagged excluded-control journeys passed 12/12 across six
projects for mounted composer, sidebar, read-only preview, attachment modal,
card controls and picker. The terminal/dock example is unmounted; simulated
touch grants no physical-device or complete seven-example credit.
Six session-switching clauses (`@ux-session-001`–`006`) passed 36/36 in
six disposable Gi browser projects. `session-switching-review.md` traces the
current installed Piclaw source, Gi assertions and code; Piclaw UI/physical
touch acceptance is separate.
Three recovery/card clauses (`@ux-extra-003`, `012`, `013`) passed 18/18 in
disposable Gi browser projects. Their current Piclaw UI was not probed.
Three tagged workspace clauses (`@ux-workspace-004`, `008`, `011`) passed 18/18
in the disposable six-project Gi browser fixture. The explorer module changed
between frozen source and 3.2.4; the installed version was checked separately.
A prior run of the seven `features/tui/*.feature` files passed the canonical
tmux Gherkin runner (10 scenarios). The latest whole-suite invocation was
aborted on `assistant_basics.feature` and is **not** a pass. A separate focused
`keyboard_behavior.feature` run passed, as did `make test-tui-reading` at three
terminal sizes. `tui-keyboard-review.md` separates screen-text checks from the
stronger history-anchor/cursor checks. A first run with `BIN_DIR` overridden failed at launch because the
shell runner hard-codes `bin/gi`; it is not a product failure. The target is not in
whole-product CI and screen-text assertions do not establish physical/pixel parity.
The 23 proposal-tagged `features/search/workspace-index.feature` cases are parsed
by a contract test and supported by selected Go tests, not run as Gherkin steps.

## Initial concrete issues to verify against Piclaw

- The initial `tests/features/settings/gi-settings.feature:40` said thinking was read-only;
  `web/src/gi-settings-models.ts:270-305` exposes supported Apply-thinking controls.
  The tagged Settings tests use an unsupported model, so that contradiction can hide.
- `features/search/workspace-index.feature:33` requests automatic background refresh;
  later clauses at199/211 require read-only GET and explicit refresh. Pinned Piclaw
  3.2.4 search-function probe intercepted the default cold/stale background request;
  Gi `internal/web/workspace_index.go:138-141` rejects GET refresh. Decide whether
  to adopt Piclaw's trigger while keeping bounded asynchronous indexing and safe
  read semantics. No worker completion or production HTTP was tested.
- Active Classic `@ux-compose-004` and Shared28 describe Piclaw's queued-item
  draft replacement. Gi's ADR-0017 no-loss recovery has its own `@gi-ux-004`
  clause and earns no Piclaw mapping credit.
- Active Classic007 and Shared16/17 describe Piclaw's exact-command prefill,
  replacing the existing draft. Gi's trailing-space and draft-preserving skill
  paths are `@gi-ux-001` and `@gi-ux-002`; the conflicting Piclaw mappings were
  removed. The authenticated skill backend journey still needs separate evidence.
- Active Classic029 and Shared41 describe Piclaw's sanitized fenced SVG image
  with escaped-source fallback. Gi's source-only path is `@gi-ux-003`.
- `features/tui/keyboard_behavior.feature:4` bundles focus, history, scrolling,
  resizing and quit, but scroll assertions only require unchanged text to remain
  present. It cannot by itself prove scrolling or cursor/viewport preservation.
- Many Classic requirements describe callbacks, supplied props or “can display”
  rather than a complete human journey. Strengthen with actual trigger, exact
  native identity/result, focus/caret, failure/retry and preservation assertions.

These are source-backed audit leads, not an exhaustive defect list. Historical
Classic/shared snapshots and supplied component bytes remain unchanged. Active
Piclaw 3.2.4 contracts were corrected separately; see `gherkin-alignment.md`. A Gi-owned
production helper now handles the model-picker keyboard contract.
Additive Gi Gherkin and functional tests were corrected on the audit branch.
A verified Classic021 finding (`picker-keyboard-gap.md`) showed its former
tagged test checked stale model responses rather than the picker keyboard
contract. Audit-branch commit `f5ad830` corrects the Gi-owned model helper
and tagged test. Six focused Classic021 model journeys, six Shared34 browser
cases, 42 model-panel cases and eight helper tests passed. The shipped 3.2.4 picker-key UI
fixture also passed in six projects. Native oracle model mutation, frozen
session-picker clauses, CI/review and deployment are separate. No blanket
mapping credit was added.

## Resume

1. Establish the Piclaw oracle revision/source and isolated runnable environment.
2. Compare each of all358 scenario definitions and421 examples against actual
   pointer/keyboard/touch, focus, draft/media, session, async-error/reload behaviour.
3. Trace each clause to test assertions (not tags), identify mock/DOM/source-only
   substitutions and exact native-code paths. Keep absent evidence distinct from bugs.
4. Produce per-scenario dispositions and a prioritized requirement→test→code plan.
5. Implement approved corrections with Make gates and fresh evidence; no deployed
   changes without exact-revision whole-product CI/all four builds.

Shared40 WIP is separately checkpointed at500e02f on
`wip/maintenance-tool-terminal-20260926`; it must not be deployed automatically.
The last known deployed Gi revision was 05e287f; port 8090 was absent after
the Piclaw upgrade. Do not restart or redeploy Gi for this audit. TUI WIP2a87a79
remains local-only and untouched.
Maintenance hold is cleared. Continue from this branch without redoing completed
Shared35/36/38 deployment or touching live data.

Inventory scripts currently retain absolute local audit paths; they are checkpoint
reproduction aids, not shipped product tooling. Generated candidates need manual
validation, including dynamically generated test names and non-web harnesses.
