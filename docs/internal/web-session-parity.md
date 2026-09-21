# Web session selection and compact TUI adaptation

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

Persistent drafts, late failed-send recovery into an unmounted origin, model mutation, queue actions and reconnect ownership still require dedicated acceptance cases. The later mutation slice below covers `015`.

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

## TUI adaptation: design, not implementation credit

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

- Switch A → B → A with distinct unsent multiline drafts and media refs; prove no model submission and no lost draft.
- Deliver buffered A tool/draft/status events after B is selected; prove B stays unchanged.
- Cancel selectors and resize at 60×18, 100×22 and 140×36; bound every row, preserve focus and draft.
- Idle layout adds zero rows compared with the existing contract; only nonzero counts occupy footer segments.
- Queue failures leave the durable item and editor consistent; no duplicate dispatch after retries.

The existing terminal selectors provide a starting point, but browser success does not mark these terminal checks complete.
