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

Picker search/focus (`013`), mutations (`015`), persistent drafts, late failed-send recovery into an unmounted origin, model mutation, queue actions and reconnect ownership still require dedicated acceptance cases. They are not included in the three passing frozen IDs.

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
