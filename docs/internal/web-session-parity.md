# Web session selection and compact TUI adaptation

## Legacy dev startup repair (2026-09-22)

The dev database predates `turns.phase`. Schema initialisation now runs table creation, additive columns/backfill and index creation in one transaction, in that order. New phases derive from existing turn status; later opens leave them unchanged. Unrelated ALTER failures are not treated as duplicate-column success. A late index error rolls back all schema changes.

Validation: synthetic legacy/reopen/rollback tests, the real database copied before migration, Go tests/vet, three repeated migration race runs and 70/70 functional tests. The dev instance now serves the app and `/api/runtime/config` on port 8090. Two-way comparisons against the restricted backup at `/workspace/tmp/gi-legacy-migration/pre-upgrade.db` find no changes to existing session fields, 51 turns, 146 messages or 317 events; integrity/foreign-key checks pass. No parity mapping changes. The attempted review delegate failed its workspace-path check and supplied no evidence.

## Latest verified slice: run-bound queue Steer (2026-09-22)

Queue Steer requires a captured matching active run. Atomic native admission consumes each queued ID at most once; idle, stale, cancelling and foreign targets reject without immediate-send fallback. Unknown/idle browser controls are disabled. Unconsumed entries return as held queue rows, survive reload and can be explicitly retried into a new run. They never auto-send. See [ADR-0018](../adr/0018-run-bound-queue-steer.md) for transaction/recovery semantics and limits.

Current evidence: **204/204 browser executions** (192 existing shell-fixture executions plus 12 local-provider Steer executions), **70/70 functional tests**, **23/23 helpers**, Go tests/vet and targeted race tests repeated three times. Coverage: **Classic 15/236** passing, 221 unmapped; **shared 2/42** passing, 40 unmapped. Classic idle-Steer/send and replacement-return conflicts remain separate and unmapped. No frozen features or supplied component/UI/pane files changed in this slice.

The browser tests verify actual provider-request delivery, duplicate keyboard activation, failed admission, idle recovery/reload/retry, socket-disconnect disabling, stale/foreign API rejection and session-switch feedback isolation. Native image projection and failed-checkpoint recovery have Go tests. An independent review delegate timed out and contributed no evidence.

Terminal adaptation: use the shared engine from a temporary queue selector capped at six rows, with existing muted/accent styling and arrow/Enter/Escape interaction. Capture run/session/generation; preserve editor/cursor and use existing transient notices. Add no top chrome or idle rows. Verify 60×18, 100×22 and 140×36 before claiming implementation. Terminal queue controls and durable media recovery are still unimplemented.

The following sections retain the evidence recorded for earlier slices.

## Implemented browser slice

Session selection now advances a generation token before rendering. Timeline, session-list and model/queue/status fetches capture that token and reject superseded responses, including a switch away and back to the same ID. SSE events must name the selected chat before they can update its timeline or status.

The composer is keyed by session ID. Gi retains draft text, browser File references and file/message references in a page-local map. Switching restores that session's draft and clears outgoing streaming/model/context/queue state before fetching the incoming state. The old component's reference callbacks cannot clear the new session's references. This cache does not survive page reloads; it does not upload attachments or persist drafts to SQLite.

The New action uses the existing native fork endpoint. Reposting a main session for `web` returns its existing main session, so it cannot create a distinct chat. Forking creates a child with the store's parent relationship and allocated agent identity; it can include inherited transcript history.

`getAgentQueueState` now projects actual queued turn IDs and text for the requested session. Reorder, steer and durable queue recovery are not implemented by this slice. The model reader uses selected-session state; model mutation still needs its own parity work. Context usage is cleared on selection and stays unavailable until a native source exists.

### Component change

`compose-box.ts` has a small, explicit Gi host bridge: initial `draftValue`/`draftMediaFiles` and `onContentChange`/`onDraftMediaChange` callbacks. The text contract follows Piclaw `70d33bc93`'s `draftValue`/`onContentChange`; the media callback is Gi-specific. The component is no longer byte-identical to its older vendored version. No CSS, keybinding or selection rules changed in this patch. A full upstream composer refresh is still needed for the searchable picker contract.

### Evidence

- `@ux-original-014`: native main/research sessions, real persisted histories, keyboard selection, correct message/turn request scopes, a held real main-session response released after selection, and draft restoration in both directions.
- Additional browser regression: New creates a distinct child with the expected stored parent ID.
- Combined matrix: 24/24 executions across Chromium/WebKit × phone/tablet/desktop (001, 002, 014 plus child creation).
- Selection-scope unit test covers A → B → A response invalidation.
- Existing functional web suite: 70/70 after fixing the SQLite connection pool configuration.

The later slices below cover `015` mutations and persistent draft/failed-send recovery. Model mutation, durable queue actions and reconnect ownership still require dedicated acceptance cases.

## Searchable picker slice (`013`)

The picker now focuses its search input on mount, groups native sessions using the pinned Piclaw helper, preserves matching descendants' ancestors, and restores the exact opening trigger on Escape. Empty-compose `@` opens it without inserting text; cancelling returns focus to the session pill. Search uses case-insensitive terms over handles, JIDs, state and available model metadata. The Gi adapter resolves nested ancestry to the actual root before grouping.

Native Tab moves focus without switching chats. Home/End edit the search input; arrows/page keys navigate enabled entries. Enter in search selects the highlighted result, while Enter on a focused button activates that button. Empty-result Enter does nothing. A layout-effect keyboard binding and functional index updates avoid stale handlers on rapid key presses.

Provenance: `web/upstream/piclaw-session-picker-70d33bc93.json` records the unchanged helper hash, pinned upstream composer hash and local deviations. Live pin/archive/activity metadata and mutation paths are not delivered here. Helper-only grouping tests for those states are not browser mutation evidence.

Validation: 36/36 matrix executions (four frozen IDs plus two Gi regressions), 70/70 existing functional web tests, Go tests/vet, Bun hook checks and eight source/helper tests. Coverage is **4/236 frozen IDs**, with 232 unmapped; the 42-case shared contract remains unmapped.

The functional suite also exposed a real admission race: the runner goroutine could emit `turn.started` before the submitting caller appended `turn.submitted`. Launch now waits for the caller to release the session coordination mutex before running. `TestLaunchedTurnWaitsForSubmissionEvent` failed before the fix and passed ten runs afterwards. Event rows remain append-only and sequence-ordered. The browser event-order test now selects its own submitted prompt and polls for completion instead of relying on the most recently updated session.

The broader delivery scope is [full-web-tui-parity-plan.md](full-web-tui-parity-plan.md); passing this slice does not complete that scope.

## Native session mutations (`015`)

The picker now calls native `PATCH /api/sessions/{id}` paths for display-name changes, database-backed pins, reversible archive and restore. Controls depend on row state and supplied callbacks. An inline name form and archive confirmation replace no-op success paths. Main sessions cannot be archived; queued/running turns or an active turn claim also block archive. Renaming changes display text, never agent identity or routing scope; browser acceptance also verifies mention autocomplete still inserts the native `@agent_id` after a rename. Archive retains history and is a picker classification, not agent shutdown or a routing prohibition. No permanent deletion is offered.

Failed requests show the server error without changing the selected chat or draft. Successful metadata changes also leave selection unchanged. A picker-open epoch prevents late feedback from appearing after dismissal/reopening; session-list revisions reject pre-mutation list responses. Native runtime status now determines activity rather than treating the selected chat as running. Forks reset archive/pin metadata and queue count.

Evidence: `@ux-original-015` exercises native pin, rename, validation failure, archive confirmation/cancellation, persistence after reload, restore and unpin in all six browser/viewport projects. A separate test delays a real failed PATCH response until the originating picker has closed. `TestSessionMetadataMutations` covers invalid payloads, missing sessions, root/busy conflicts, identity/model preservation, idempotence and fork reset against file-backed SQLite; it also passes under the race detector.

Current totals: **48/48 browser executions**, **5/236 frozen IDs passing**, 231 unmapped; **42 shared-contract cases still unmapped**. The existing functional web suite passes 70/70; Go tests/vet, nine Bun source/helper tests and hook checks pass. See [ADR-0009](../adr/0009-session-picker-mutations.md) for API and lifecycle limits.

Terminal mutation adaptation: add an on-demand row-action submenu to the bounded session selector. Use a temporary one-line rename input and explicit archive confirmation; Escape restores the previous selector/editor and draft. Archived rows live in a requested group/filter. Do not add a permanent action bar, badge row, sidebar or header. These terminal changes remain unimplemented.

## Browser-local durable drafts and captured sends

IndexedDB now retains session text, media bytes, file references and message references across reload. Captured sends are journalled before network I/O; storage failure prevents an unprotected send and retains the draft. Media upload and prompt submission use the captured session ID. Failed sends merge into the origin session's newer draft, including after A→B→A; another selected chat is unchanged.

A pending send recovered after reload shows unknown delivery and is never automatically retried. Acknowledgement removes only its capture. Cleanup-storage failure warns that recovery may contain already-delivered text. Browser-profile-local persistence has no cross-tab merge or cross-device synchronisation. Writes are asynchronous; an abrupt process crash can lose edits that have not committed. See [ADR-0011](../adr/0011-browser-draft-recovery.md).

WebKit required explicit attachment byte/metadata records instead of stored File objects. Native file selection also needed the Gi tree adapter's `{root}` response wrapper. Post timestamp reference links now reserve space for copy/delete controls. These changes support real draft-reference tests; full workspace/editor and upload-progress parity remain open.

Verified `@ux-compose-001`, `002`, `003` and `006`, plus reload, storage failure, A→B→A and acknowledgement-cleanup regressions. Current result: **102/102 browser executions**, **9/236 Classic IDs passing**, 227 unmapped; all 42 shared-contract cases remain unmapped. Existing functional tests pass 70/70; Go tests/vet, Bun checks and 14 source/helper tests pass.

Terminal follow-up: persist origin-owned drafts/recovery in a native local store, stage media refs explicitly, and persist queue recovery before backend deletion. Use existing footer notices and bounded on-demand controls. The current terminal draft cache is process-local; this browser slice adds no terminal rows or terminal implementation credit.

## Durable queue reorder and cancellation (`018`)

Explicit queue intent now creates a durable follow-up behind the active turn rather than steering that turn. A persisted queue position controls FIFO execution and reorder without changing admission timestamps or append-only events. Session-scoped queue APIs reject stale snapshots and foreign/claimed/running IDs; queued-only cancellation cannot stop a turn that has already started.

The web updates order/removal optimistically, refreshes after success/failure and reports errors. Queue revisions reject old polling responses; selection generations reject responses from superseded visits. Duplicate/no-op queue rendering was removed. Steer has no supplied callback and is hidden until implemented.

`@ux-original-018` passes all six projects using real queued prompts and a parity-only gated shell responder. Tests verify reload order, actual reordered execution, concurrent-append conflict, removal failure and stale/cross-session responses. Go/store/API tests and three-run focused race tests pass. Current totals: **120/120 matrix executions**, **10/236 Classic IDs**, 226 unmapped, and all 42 shared cases still unmapped. Existing functional web suite: 70/70. See [ADR-0012](../adr/0012-queued-followup-order.md).

`016` SSE/local optimistic reconciliation, return-to-editor recovery and atomic Steer remain open. Terminal adaptation is an on-demand bounded queue list operating on durable IDs, with existing nonzero footer counts and error notices. It adds no idle rows; terminal queue mutations are not yet implemented.

## Queue SSE reconciliation and disconnect cleanup

The stream now carries selected-session queue invalidations from the native topic bus, with subscriptions installed before the connected notification. Reconnect and wake refresh persisted queue/status/timeline data. Local queued placeholders correlate with durable records through `client_request_id`; reconciliation removes the placeholder once the server row is observed. Identical prompt text remains distinct. Failure removes the placeholder while recovering the origin draft.

Disconnect clears transient assistant status/draft/plan/thought and pending/run refs without changing the user's draft. Connection revisions reject pre-disconnect reads. Source-instance and selection-generation checks suppress callbacks from replaced EventSources; explicit disconnect clears reconnect timers and connecting state.

`016` and `ux-reconnect-001` pass native browser acceptance. A held real acknowledgement tests SSE-first reconciliation; a byte-for-byte proxy closes actual SSE sockets, verifies removal of a real streamed stdout preview and restores queue state changed by an independent API client during the outage. No synthetic SSE/timeline data is injected.

Current totals: **138/138 matrix executions**, **12/236 Classic IDs**, 224 unmapped; all 42 shared cases remain unmapped. Existing functional suite: 70/70. Go/vet, focused three-run SSE race tests, Bun hooks and 15 source/helper tests pass. [ADR-0013](../adr/0013-queue-sse-reconciliation.md) records the contract and limits.

Other reconnect cases, context/search/version-drift behaviour and queue return/Steer remain open. Terminal adaptation uses topic invalidations to refresh a bounded on-demand queue list; the existing footer handles nonzero counts and transient warnings. No additional idle rows are needed, and this slice adds no terminal implementation credit.

## Authoritative session-local model selection

Native model GET/PATCH and `/model` commands now validate the configured/provider catalogue and persist only the target session. New prompts use its selected model; old runtime metadata cannot replace the separately stored selection. Picker changes do not submit the user's draft or change global defaults. Errors retain the previous selection and composer contents.

Host model revisions and component mount guards reject late polls/catalogues/mutations, including A→B→A. Keyboard Enter activates the focused option; Tab traverses controls, and list changes no longer steal focus. `pagehide` explicitly closes SSE before navigation.

`021` and compaction `008` are verified, with additional pointer/keyboard/reload/failure tests. Current totals: **168/168 matrix executions**, **14/236 Classic IDs**, 222 unmapped, all 42 shared cases still unmapped, and **70/70** existing functional tests. Native API/command tests and focused three-run race checks pass. See [ADR-0014](../adr/0014-session-model-selection.md).

Measured context usage, context-fit blocking and richer picker contracts remain open, so `020` and compaction `006/007` are not counted. Terminal adaptation should reuse its bounded model selector and existing footer, with session-local persistence and shared validation. The later terminal model slice below closes the global-write gap.

## Latest measured provider-request context

The provider loop now appends `context.measured` events separately from cumulative billing totals. Context input includes uncached input plus cache-read/write tokens and excludes output. Session model responses expose only measured values and catalogue capacity; unknown fields remain null. The shell responder has no measurement and therefore shows `?`, including after reload and model changes.

The web labels the value as the latest measured provider request, keeps unknown usage neutral and disables compaction without a supplied callback. Model-fit helpers and shared native selection reject known overflow. The terminal uses the measurement in its existing footer context segment instead of cumulative billing totals, with no additional rows.

`ux-context-002` now passes all six projects. Native/provider-loop tests and supplied-value helper tests cover measurement separation, persistence, scope, formatting/thresholds and model-fit predicates. Those tests do not establish provider-backed browser fit/compaction parity; related IDs remain unmapped. [ADR-0016](../adr/0016-measured-request-context.md) records measurement age, tokenisation and reserved-output limits.

Current evidence: **174/174 browser executions**, **15/236 Classic IDs**, 221 unmapped, all 42 shared cases unmapped, **70/70** functional tests, Go/vet/race checks and 18 source/helper tests. The three-size terminal harness, smoke and seven-file Gherkin suite pass.

## Durable queue return (shared contract)

Return now recovers the queued row's text, media bytes and file/folder/message references into the latest origin-session draft. It persists the merged draft and a durable-ID recovery record before queued-only DELETE. Storage failure prevents deletion; failed removal/reload/retry reuses the recovery record without duplicating the returned content. A consumption race keeps the recovered draft and warns that the original turn may already have run. Another selected chat keeps its draft/focus.

Shared `@shared-28` is verified in all six projects, including pre-DELETE inspection of committed IndexedDB state, concurrent edits, quota failure, real failed removal, reload, retry and cursor restoration. Separate regressions cover consumption and session-switch races. Classic `017` and compose `004` explicitly require replacement/media clearing and remain unmapped. Sources are unchanged; the reporter now lists shared results separately. See [ADR-0017](../adr/0017-durable-queue-return.md).

Current totals: **192/192 browser executions**, Classic **15/236** (221 unmapped), shared **1/42** (41 unmapped), **70/70** existing functional tests, Go/vet, hook checks and 23 source/helper tests. No terminal implementation credit is added here.

Terminal adaptation needs native recovery storage, pending-media refs and queued-ID idempotency before removal, using the existing editor and bounded queue view. Incomplete returns use existing notices; no idle rows or permanent recovery panel. Browser multi-tab merging, recovery-marker retention and duplicate-upload avoidance on later send remain open.

## Terminal session slice: implemented and separately tested

`Alt-S` opens the existing session selector without replacing unsent input. It uses at most six results plus two temporary title/search rows, no box border, and no added idle rows. Filtering keeps full session IDs; display truncation is UTF-8-safe and terminal-cell bounded. Up/Down wrap, Enter selects and Escape restores the editor. Resizing keeps the selected row visible.

Session switches preserve editor text, rune cursor, undo/yank, history/search position and the existing local queued-draft list. The target is validated before origin state changes. Streamed assistant text, extension slots, transient blocks and usage values are cleared or reloaded separately. Captured session/generation wrappers reject buffered events, including an A→B→A revisit. Prompt/peer completions validate ownership at UI application time; old forwarders stop even when their downstream channel is full.

Evidence: `make test-tui-sessions` passed native tmux interactions at 60×18, 100×22 and 140×36, including exact before/after cancellation screenshots, both session drafts, zero submitted turns and open-picker resizing. Unit/race tests cover buffered topic events, event types, failed switches, model fallback and extension-question cancellation. Existing TUI smoke and seven-file Gherkin harnesses passed. Artifacts: `test-results/tui-sessions/`; design: [ADR-0010](../adr/0010-terminal-session-selection.md).

These caches are process-local. Pending media refs, durable queue mutation/retry semantics and terminal session-mutation submenus are not implemented. The terminal slice did not change browser coverage. The later browser draft slice brings coverage to 9/236 Classic IDs with 227 unmapped, plus 42 unmapped shared cases.

## Terminal session-local model slice

Web and terminal now share catalogue validation and session persistence through `internal/inference/session_model.go`. `/model <name|index>`, Alt-M and cycle keys update only the selected session. Startup/switch/footer resolution prefers explicit selection over runtime model metadata. Invalid models or failed storage retain the previous selection; no prompt turn is created by model actions. `/scoped-models` keeps its separate workspace configuration role.

The existing temporary model selector remains bounded to six results. Alt-M preserves the unsent editor; errors reuse its search/help line; successful selection closes it without transcript noise or a new idle footer row. The live tmux harness passes 60×18, 100×22 and 140×36, exact cancel/row-position checks, model/draft A→B→A restoration, settings-byte comparisons, clean restart and actual next-turn model verification. Unit/race, existing TUI smoke/Gherkin, Go/vet and browser regression suites pass. See [ADR-0015](../adr/0015-terminal-session-model-selection.md).

Browser coverage is unchanged: 168/168 executions, 14/236 Classic IDs, 222 unmapped and all 42 shared cases unmapped. Pending terminal media, queue recovery and measured model context remain separate gaps.

## Remaining TUI adaptation design

The terminal keeps the current transcript → separator → editor → separator → path/footer layout. Every additional control is temporary or folded. No session sidebar, permanent toolbar, top banner or second transcript is added.

| Web flow | Terminal adaptation | Footprint and interaction |
|---|---|---|
| Session picker | Extend existing `/sessions` selector and `/switch <id>` fallback | Search row and at most six result rows above the editor; clamp to available height. Up/Down navigate, Enter selects, Esc cancels and restores the editor. No persistent picker rows. |
| Switch chat | Save origin draft and restore target draft, refs and history | No new steady-state rows. Buffered events carry session/generation ownership and are discarded after a switch. |
| Child session | Existing `/fork` semantics with parent/child identity | One brief footer notice; `/tree` is requested transcript output, never an always-visible tree. |
| Queue | Existing nonzero queue count; `/queue` opens a bounded list | No empty queue chrome. Each action uses durable item and session IDs; no implicit resend when returning a draft. |
| Model/context | Existing footer model plus `/model` selector | Display only measured/advertised values; unknown context stays unavailable. No extra permanent status band. |
| Workspace | `/attach`, path completion and on-demand file selection | No terminal file sidebar. Browser drawers have no literal terminal counterpart. |
| Errors | Recoverable inline/footer notice near the active editor | Failed switches retain the original draft/session; Esc dismisses temporary UI, never silently discards input. |

Use the pi-tui SelectList/SettingsList interaction pattern: muted secondary labels, one accented selection marker, bounded visible rows, theme-derived colours and width-safe rendering. Pi's component interface requires every rendered line to fit the supplied width. Gi remains on go-tui; this is an interaction/layout adaptation, not an SDK migration.

Reference reviewed: the installed pi-coding-agent `docs/tui.md`, especially Selection Dialog, Persistent Status Indicator, Widgets Above/Below Editor, line-width and focus rules; Gi's `tui-pi-layout-contract.md` and `tui-selectors.md`.

### Terminal acceptance still to implement

- Add a pending-media draft collection and verify per-session media refs. Existing `/attach`/`/paste-image` commands store or submit media immediately; they do not stage composer attachments.
- Prove durable queue failure/retry consistency and no duplicate dispatch. Preserved local queued-draft text alone does not satisfy this contract.
- Add capability-gated terminal mutation submenus and archive/restore filtering with the same bounded footprint.
- Complete wider feature-family terminal adaptations with separate evidence; the verified session slice does not close them.
