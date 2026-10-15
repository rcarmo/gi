# Classic canonical 001–029: clause review

Baseline: 24-file frozen Classic corpus from Piclaw `70d33bc`; current oracle is
shipped Piclaw 3.2.4. This is a **qualified source/test review**, not a new full
browser run, physical-device acceptance, or permission to rewrite frozen hashes.
Installed 3.2.4 UI probes cover slash/fixture-listed skill prefill, model-picker
keys, queued return, SVG, first-send request and a bounded Plan add-on
remote-update/Refresh slice. A separate disposable UI/backend-function deletion
fixture covers reply prompts, cascade, cancellation and an unseen-reply orphan.
All other current-oracle rows remain provisional; none of these probes is a
production HTTP/auth/live or full combined-case acceptance.
`oracle-deltas.md`, `manual-findings.json`, and `picker-keyboard-gap.md` contain
source locations and the focused discrepancies.

| ID | Gi test / code trace | Review disposition and missing proof |
|---|---|---|
| 001 | `tests/ux/classic.spec.mjs:5-58,64-99`; native menu | Bounded pointer/keyboard open and Escape/outside dismissal; shared test adds Tab, focus and no underlying send. Fresh boot and pixels separate. |
| 002 | `classic.spec.mjs:38-62` | Toggle twice, retain draft/session, no prompt; not editor/workspace-file parity. |
| 003 | `quick-actions.spec.mjs:20-34` | Real typing, groups, filtered navigation and action; exact-title-before-prefix ranking lacks a competing-result assertion in the tagged case. |
| 004 | `quick-actions.spec.mjs:65-84,149-218`; `web/src/ui/app-browser-events.ts:6-16` | Six outline surfaces are not individually proved by real editable controls. Artificial contenteditable/CodeMirror DOM cannot earn editor-input credit; Shared8 likewise stays unmapped. |
| 005 | `quick-actions.spec.mjs:53-64` | Guard branches for composing/repeated/prevented/modifiers are tested synthetically with positive real-key readiness; not a physical IME pass. |
| 006 | `quick-actions.spec.mjs:36-42,237-275` | Escape/outside dismissal, no action, fresh query and draft/focus in shared variants. |
| 007 | `quick-actions.spec.mjs:44-51` | **Current-oracle conflict.** Gi expects `/model `; shipped Piclaw 3.2.4 (all six isolated projects) produces `/model`, replaces draft, focuses and does not send. Historical Gherkin's trailing-space clause even appears unsupported by its own `70d33bc` source. Versioned oracle Gherkin records actual behaviour. |
| 008 | `skills.spec.mjs:4-47`; `tests/ux/oracle/piclaw-basic-probe.mjs` | **Disputed mapping:** tagged Gi body preserves an existing draft. Shipped Piclaw 3.2.4 selects a fixture-listed `/skill:proof` in the Slash commands group and replaces the draft with exactly `/skill:proof` in six projects, no POST. This isolates the UI prefill path; the fixture does not prove real skill discovery, loading or execution. |
| 009–012 | Installed `piclaw-addon-plan-sidebar` 0.1.25 source; no Gi Plan route/tool/sidebar located | A shipped-Classic/add-on disposable browser probe in Chromium/WebKit covers only the 010 remote-update/explicit-Refresh slice: dirty text survives the event, then Refresh replaces it without confirmation. Save, Submit and tool clauses remain source-only; no Gi implementation/mapping. See `plan-oracle-gap.md` for Shared18–22 strengthening. |
| 013 | `session.spec.mjs:9-76` | Native picker grouping, pointer/keyboard, search focus, Escape and exact trigger focus exercised; seeded sessions, not fresh boot. |
| 014 | `session.spec.mjs:453-518` | Keyboard selection, session-scoped timeline/queue reads, held old response rejection and per-session drafts exercised. |
| 015 | `session.spec.mjs:570-649` | Pin/rename/archive/restore success/error/capability gates, reload and draft/routing ownership exercised. |
| 016 | `queue.spec.mjs:43-74` | Duplicate-text queue entries reconciled by native IDs across held ack and external insertion; no text-only identity. |
| 017 | Shared28 `queue-return.spec.mjs:26` is not this contract | **Unmapped policy conflict:** Classic/Piclaw replaces draft and clears media; Shared28/Gi retains latest concurrent draft/media transactionally before DELETE. Do not swap policies silently. |
| 018 | `queue.spec.mjs:126+` | Native queue reorder/removal and 409 refresh are tested; backend identity and unrelated-group invariants need their store tests, not the screenshot. |
| 019 | Shared30 steering work is not full Classic policy | **Unmapped:** Classic allows backend to send immediately if the stream ended; Gi shared contract Steer requires matching active run. Requires policy review. |
| 020 | `models.spec.mjs:20-45` | Native captured-chat PATCH, accepted context, rejected model, keyboard selection and no draft submission exercised. |
| 021 | `models.spec.mjs:46+`, `web/src/gi-model-picker.ts`, `scripts/patch-model-picker.mjs` | **Locally corrected model half:** former tag tested stale model replies; a new tagged native browser journey now covers search, Arrow/Page, plain caret vs Control/Meta+Home/End, one Enter model PATCH, Escape focus and no prompt across six projects. Existing Shared34 six-project test skips disabled entries. The frozen session-picker half remains separate; a six-project isolated shipped Piclaw 3.2.4 keyboard probe covers model search focus, Page and modified Home/End, and Escape focus without model mutation. See `picker-keyboard-gap.md`. |
| 022 | `models.spec.mjs:78+` | Sparse metadata/unknown context and stale catalogue test is bounded evidence; no inferred measured usage. |
| 023 | `reconnect.spec.mjs:77+` | Native reconnect refresh and captured Stop evidence; keep generic/TUI cancellation distinct. |
| 024 | `shared-copy-delete.spec.mjs:47+` covers narrower Shared37 | **Unmapped Gi:** joined shipped-UI/installed-backend-function disposable fixtures in Chromium/WebKit confirm/cancel a visible parent+three-reply cascade but have no combined copy journey or production HTTP/auth/live acceptance. An unseen stored reply is orphaned on direct deletion. Gi still lacks a native reply graph and cascade controls; see `message-deletion-review.md`. |
| 025 | `internal/tools/messages.go`, `docs/internal/message-retrieval.md` | **Unmapped:** current-session-only numeric-ID tool lacks Classic all-chat/family-owned authorization, despite complete Shared38 web mapping. |
| 026 | `drafts.spec.mjs:542+` | Isolated native upload failure/retry preserves byte identity and blocks submission until media ID exists; browser physical file picker separate. |
| 027 | `tool-activity.spec.mjs:7-35` | Status/timer and stale call identity shown. This is not Shared40 expandable output pane; pending terminal-provenance WIP500e02f is not deployed. |
| 028 | `speech-contract.spec.mjs:31+` | Browser speech ownership/copy path bounded by fixture; physical audio playback and assistive-technology acceptance separate. |
| 029 | `rendering.spec.mjs:54-61` | **Version-stale.** Frozen source-only fence conflicts with shipped Piclaw 3.2.4 safe data-URL image/default sanitizer; unsafe fence returns inert source. No current SVG credit from old test. |

## Cascade order

1. Keep frozen historical IDs/report stable while adding an explicit **current
   oracle** matrix. Do not advertise historical mappings as current-release parity.
2. Review and CI the local Classic021 Gi-owned picker correction without a
   supplied-component edit. Six focused model and six Shared34 browser projects
   passed; an isolated six-project Piclaw picker-key probe passed. Verify the
   session-picker half and native mutation separately before full acceptance.
3. Resolve destructive prefill/queued-return policies with the user. Versioned
   Piclaw oracle behaviour is proven; Gi's no-loss semantics are deliberate and
   should not be removed merely to raise historical counts.
4. Keep Plan, cascade deletion, all-chat/family message reads, SVG safety and
   tool-pane output scoped as unimplemented/unmapped until native journeys pass.
