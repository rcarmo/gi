# Web session selection and compact TUI adaptation

## Native write index invalidation (2026-09-23)

[ADR-0051](../adr/0051-native-write-index-invalidation.md) connects engine/HTTP/script regular filesystem writes to atomic scoped invalidation before/after mutation, without automatic refresh. Native tests cover partial errors, cancellation, scope/VFS isolation and pending revisions; shell/external/watch delivery and crash recovery remain gaps. Go/vet/build/hook, 77 functional, 32 helpers, focused24 browser, tools/store/web/turn race ×3 and Darwin arm64 cross-compilation pass. An existing cancellation fixture now waits for claim cleanup before its unchanged absence assertion. No UI/TUI or frozen credit: **45/236 Classic**, **2/42 shared**, **191/40 unmapped**; 22 derived proposals stay separate.

## Application-owned explicit index refresh (2026-09-23)

[ADR-0050](../adr/0050-application-owned-index-refresh.md) wires POST reindex to shared bounded batches, isolates caller disconnects and joins HTTP/index shutdown before SQLite closes. Startup and GET remain scan-free; watcher/mutation delivery and automatic freshness are not enabled. Go/vet/build/hook, 77 functional, 32 helpers, focused 24 browser, lifecycle race ×10 and full HTTP/web/indexer race ×3 pass. Built-process SIGTERM/bind-failure cleanup also passes. No supplied UI/TUI change, no new frozen credit: **45/236 Classic**, **2/42 shared**, **191/40 unmapped**; 21 derived proposals remain separate.

## Bounded internal index scheduler (2026-09-23)

[ADR-0049](../adr/0049-bounded-index-scheduler.md) adds fixed-scope coalescing, shared completion tickets, revision-aware peer/follow-up completion, bounded retries and cancel/join shutdown. It has no production caller yet. Go/vet/build/hook, 76 functional, 32 helpers, store/indexer race ×3 and scheduler race ×10 pass; 20 derived proposals parse separately. Build output moved to `/tmp` after ENOSPC without removing backups/screenshots. No browser/terminal matrix or new frozen credit: **45/236 Classic**, **2/42 shared**, **191/40 unmapped**. Query GET remains read-only; application/watcher wiring is next.

## Durable index invalidation prerequisite (2026-09-23)

[ADR-0048](../adr/0048-durable-index-invalidation.md) adds v2 requested/acknowledged scope revisions so changes during refresh remain pending after publication. Native races, failure/reopen, held-worker tests and copied dev database 30-table/v1-ledger preservation pass. Go/vet/build/hook, 76 functional, 32 helpers, store/indexer race ×3, focused 24 browser and TUI smoke/Gherkin pass. Watcher/mutation delivery and scheduling remain disconnected; 18 derived proposals remain separate. No UI change or new frozen credit: **45/236 Classic**, **2/42 shared**, **191/40 unmapped**.

## Index startup settings and optional roots (2026-09-23)

[ADR-0047](../adr/0047-index-settings-and-optional-roots.md) wires extra roots/extensions and explicit optional roots from `.pi/settings.json`. Native/browser evidence covers initial absence, changed configurations and retained index/drafts on populated-root disappearance with explicit retry. **498 browser**, **76 functional**, **32 helpers**, Go/vet/build/hook and config/search/web race ×3 pass. Meter capability timing is now gated on native claim cleanup without weaker assertions.

No frozen credit: **45/236 Classic**, **2/42 shared**, **191/40 unmapped**. Background freshness, query consumers and terminal controls remain gaps; 17 derived proposals stay separate. Screenshots attached at all three browser sizes.

## Explicit native index API and web controls (2026-09-22)

[ADR-0046](../adr/0046-native-workspace-index-api.md) connects authenticated scoped status/query/POST reindex to the verified worker and supplied Refresh/Reindex controls. Missing-root failure, durable retained results, explicit retry and draft/session/reload preservation pass natively and in all six browser projects. **492/492 browser**, **75/75 functional**, **32 helpers**, Go/vet/build/hook and search-store/indexer/web race ×3 pass. ENOSPC and WebKit reruns are documented.

No new frozen mapping: workspace-005 is still compound/incomplete, and background freshness, settings/optional roots, vectors and terminal controls remain gaps. Coverage stays **45/236 Classic**, **2/42 shared**, **191/40 unmapped**. Failure screenshots attached at desktop/tablet/phone sizes.

## Explicit index worker prerequisite (2026-09-22)

[ADR-0045](../adr/0045-index-refresh-worker.md) connects lease acquisition/renewal to native scan/commit/failure cleanup, without automatic scheduling. Tests cover cancellation, write/renewal failures, two-store takeover and killed-process expiry/recovery. Go/vet/build/hook, 74 functional, 32 helpers and worker/store race ×3 pass. No application index controls or terminal UI are connected, and coverage remains **45/236 Classic**, **2/42 shared**, **191/40 unmapped**.

## Rooted scanner/chunker prerequisite (2026-09-22)

[ADR-0044](../adr/0044-rooted-index-scanning.md) implements internal bounded inventory scanning and lossless UTF-8 line chunks, including skills roots, rehash/change detection, required-root failure and native scan→commit preservation. Go/vet/build/hook, 74 functional, 32 helpers, scanner/chunker/store race ×3, 2.38M chunk fuzz executions and Linux/macOS cross-builds pass. No worker, native search/status/reindex or terminal control is connected. Coverage remains **45/236 Classic**, **2/42 shared**, **191/40 unmapped**.

## Scoped refresh storage APIs (2026-09-22)

[ADR-0043](../adr/0043-scoped-index-refresh-transactions.md) adds deterministic scope configuration, fenced lease/failure/recovery and atomic complete-snapshot publication with unchanged identity preservation and overlapping membership cleanup. Go/vet/build/hook, 74 functional, 32 helpers and store race ×3 pass. Scanner/workers and web query/status/reindex are not connected; terminal controls remain design work. No new frozen credit: **45/236 Classic**, **2/42 shared**, **191/40 unmapped**.

## Workspace-index storage prerequisite (2026-09-22)

[ADR-0042](../adr/0042-versioned-workspace-index-schema.md) installs the scoped schema through a versioned atomic startup migration, preserving legacy search and runtime data. A copied dev database retained all 19 existing tables with clean integrity. Go/vet/build/hook, 74 functional, 32 helpers, store race ×3, focused 12-case workspace browser and all three-size TUI suites pass. There is no scanner, background worker or native status/reindex API yet; frozen coverage remains **45/236 Classic**, **2/42 shared**, **191/40 unmapped**.

## Earlier indexing design gate (2026-09-22)

The provisional whole-workspace rebuild is shelved pending adaptation of Piclaw's configured roots/scopes and incremental/background lifecycle. [Pinned Piclaw/Tau/Vibes comparison](search/indexing-lineage-20260922.md) separates workspace indexing from conversation FTS, adds 15 proposed non-frozen Gherkin scenarios, and tests a candidate SQLite schema. Go/vet, candidate race ×3 and 32 helpers pass; these are design/schema checks, not browser or runtime indexing acceptance. Coverage stays **45/236 Classic**, **2/42 shared**, **191/40 unmapped**. Workspace-005 remains open.

## Latest feature evidence: hidden files and subtrees (2026-09-22)

Workspace-004 passes native root/expanded-subtree reloads with persisted hidden visibility and retained text/files. The host bridges the global menu event to the pinned explorer's existing control; the API honours bounded depth/path/hidden queries. Legacy no-argument trees remain compatible. [ADR-0041](../adr/0041-workspace-hidden-subtrees.md).

Verified **486/486 browser**, **74/74 functional**, **31 helpers**, full Go/vet/build/hook and web race ×3. Coverage **45/236 Classic**, **2/42 shared**, **191/40 unmapped**. Supplied sources are unchanged; reindexing, mutations and terminal file navigation remain unverified. Three-size browser screenshots attached.

## Earlier feature evidence: read-only workspace previews (2026-09-22)

Workspace-008 passes native Markdown/text/image/binary rendering and metadata acceptance. The host registers the supplied preview panes; native bounded text and rooted raw retrieval use existing authentication and active-content safety headers. Supplied UI sources are unchanged. [ADR-0040](../adr/0040-read-only-workspace-previews.md).

Verified **480/480 browser**, **73/73 functional**, **31 helpers**, full Go/vet/build/hook and web race ×3. Coverage **44/236 Classic**, **2/42 shared**, **192/40 unmapped**. Editor/CRUD, expanded format support and compact terminal viewer acceptance remain gaps. Screenshots attached at desktop/tablet/phone sizes.

## Earlier feature evidence: stored images and lightbox (2026-09-22)

Native media IDs/typed blocks now reach timeline and search. Guarded metadata/raw endpoints supply stored bytes to the unchanged image modal. **timeline-013/014/015/016** pass keyboard, pointer and trusted-touch dismissal; additional session/search/reload/404 checks retain drafts and files. [ADR-0039](../adr/0039-native-media-lightbox.md).

Verified **474/474 browser**, **72/72 functional**, **31 helpers**, full Go/vet/build/hook and web race ×3. Coverage **43/236 Classic**, **2/42 shared**, **193/40 unmapped**. No terminal implementation; explicit bounded attachment actions are design work. Raster browser proof uses PNG; annotation, iPad drawing, thumbnail generation and richer preview contracts remain gaps.

## Dependency regression (2026-09-22)

Upgraded go-ai to `upstream-v0.87.0`, Go to 1.26.8, and compatible Go/web dependencies. KaTeX now bundles matching CSS/fonts. **444/444 browser**, **71/71 functional**, **29 helpers**, Go/vet/race ×3 and all three-size TUI suites pass. go-tui stays at 0.18.2 because newer releases lose regular-mode history on resize; gVisor stays at Tailscale's required version. [Upgrade evidence](dependency-upgrade-20260922.md). Coverage remains **39/236 Classic**, **2/42 shared**, **197/40 unmapped**. Lightbox work is separate and unverified.

## Earlier feature evidence: native upload recovery (2026-09-22)

Original-026 passes native upload-ID and error acceptance, including a partially uploaded batch, newer draft merging, session isolation, reload and explicit retry. Tests remove a multipart boundary header to obtain a real parser rejection without rewriting file bytes or fabricating a response. Persistent IDs, filenames and bytes are checked through the actual API. Application/supplied components unchanged. [ADR-0038](../adr/0038-native-upload-failure-recovery.md).

Verified: **444/444 browser**, **70/70 functional**, **29 helpers**, full Go/vet/hook. Coverage **39/236 Classic**, **2/42 shared**, **197/40 unmapped**. Compose-005 lacks separate upload progress; terminal pending media remains design work without new idle chrome. The terminal matrix was not rerun for this test-only change.

## Earlier browser evidence: native Markdown rendering (2026-09-22)

Timeline tables now retain full-width automatic table layout through a host-scoped override; overflowing columns remain reachable by native horizontal scrolling. Stored assistant posts verify code-copy placement/trusted plain-text payloads and SVG fences staying literal source. Mapped: **timeline-023/024, original-029**. Supplied components/stylesheets unchanged. [ADR-0037](../adr/0037-native-markdown-rendering.md).

Verified: **432/432 browser**, **70/70 functional**, **29 helpers**, full Go/vet/hook. Coverage **38/236 Classic**, **2/42 shared**, **198/40 unmapped**. One capability-timing meter failure passed unchanged on rerun and is documented separately. Compound link-preview/copy-delete/speech contracts and terminal source-copy adaptation remain unverified; no terminal code or new idle UI in this slice.

## Earlier native/browser evidence: completed-claim admission (2026-09-22)

Prompt admission now atomically requires a matching running claim for live steering. A terminal predecessor still in cleanup causes a distinct durable queued turn; cleanup retains claim ownership and drains FIFO. Strict queue Steer remains unchanged. A held native completion hook reproduces the old exhausted-turn response and verifies separate HTTP admission, exactly one stored user message per tested prompt and newer draft/reload preservation. [ADR-0036](../adr/0036-completed-claim-admission.md).

Verified: **408/408 browser**, **70/70 functional**, **29 helpers**, full Go/vet, store/turn races ×3 and all existing TUI suites at the three required sizes. Paging now checks exact persisted ID windows, including legitimate queue-status rows. No UI rows/components or frozen mappings added: **35/236 Classic**, **2/42 shared**, **201/40 unmapped**. The earlier display-idle/admission warning is resolved by safe queuing; crash/replay and broader parity remain open.

## Earlier browser evidence: explicit folder references (2026-09-22)

The host adds an explicit selected-folder reference action in the existing workspace header, preserving navigation and supplied components. compose-008 verifies exact multiline/file/folder/message-reference serialisation and references-only submission. Native tests also cover keyboard activation, duplicate controls, durable drafts, session isolation and failed-send recovery. [ADR-0035](../adr/0035-explicit-folder-references.md).

Verified: **402/402 browser**, **70/70 functional**, **29/29 helpers**, Go tests/vet and hook checks. Coverage: **35/236 Classic**, **2/42 shared**, **201/40 unmapped**. Reconnect fixture waits now observe the selected chat's native subscription; selection clears stale connected status. History setup uses explicit queue intent; display-idle can still precede active-claim release and needs separate follow-up. No terminal code or new terminal credit in this slice.

## Latest terminal evidence: fullscreen selection/copy (2026-09-22)

Padded-cell drag selection, clipboard-governed release/Ctrl-C/X copy and held-edge scrolling pass native SGR/OSC52 PTY tests at **60×18, 100×22 and 140×36**. Stationary tool clicks still expand; output/layout/search/session changes invalidate stale selections; notices use the existing separator. All TUI suites, Go/vet/race ×3, 29 helpers and 70/70 functional browser tests pass. [ADR-0034](../adr/0034-fullscreen-transcript-selection.md). This resolves the saved selection WIP; pointer link precedence, mutation-stable anchors and broader parity remain open. No frozen browser mapping change.

## Earlier terminal evidence: Pi message spacing correction (2026-09-22)

User bands now have top/bottom blank padding; tools have an external blank separator and colored padding; assistant messages have a leading blank separator. Markdown continuation rows remain in one message. Exact rendered-cell and real terminal checks pass at **60×18, 100×22 and 140×36**, with unchanged editor/footer rows. All TUI suites, Go/vet/race ×3, 29 helpers and 70/70 functional browser tests pass. [ADR-0033](../adr/0033-pi-transcript-spacing.md). Three fresh screenshots are attached. Fullscreen selection WIP remains separate and unshipped; frozen browser mapping stays 34/236 Classic, 2/42 shared.

## Earlier terminal evidence: rendered search/prompt jumps (2026-09-22)

Fullscreen Ctrl-Shift-F temporarily replaces the editor with a query; matching rendered rows are highlighted, Enter/Shift-Enter navigate and Escape restores draft/cursor/reader state. Ctrl-Shift-Up/Down jump between user prompts. **60×18, 100×22 and 140×36** native PTYs verify Unicode, tool visibility, live arrivals, resize/reopen and no idle-row growth. All TUI suites, Go/vet/race ×3, 29 helpers and 70/70 functional browser tests pass. [ADR-0032](../adr/0032-fullscreen-transcript-search.md) records per-row/retention limits. Fullscreen pointer selection/copy and stronger reflow/eviction anchors remain open. No browser mapping changes.

## Earlier terminal evidence: opt-in native scrollback (2026-09-22)

`-tui-mode regular` prints completed expanded output into native terminal history, with no alternate screen or mouse capture, a five-row idle dock and bounded active preview. **60×18, 100×22 and 140×36** PTYs verify ordered/deduplicated history, native selection/copy during completion, multiline draft/cursor, resize, selectors, session isolation, exit and reopen. All existing TUI suites, Go/vet/race ×3, 29 helpers and 70/70 functional tests pass. [ADR-0031](../adr/0031-regular-terminal-scrollback.md) records the resize-marker workaround and remaining retention/layout limits. Fullscreen search/prompt jumps/selection and light theme remain open. Browser mapping unchanged.

## Earlier terminal evidence: Pi bands and fullscreen navigation (2026-09-22)

Pi dark user/pending/success/error backgrounds now fill flat output bands. Rendered-height paging/bottom, focused Home/End, Ctrl-Home/End editor movement, Ctrl-O expansion and dock-wheel fallback are verified at **60×18, 100×22 and 140×36**. Native ANSI snapshots, rendered-cell checks, all TUI suites, Go/vet/race ×3, 29 helpers and 70/70 functional tests pass. See [ADR-0030](../adr/0030-terminal-outcome-bands.md). Pi regular-mode native scrollback and fullscreen search/prompt jumps/selection remain open; dark-theme styling and internal navigation alone do not complete that request. No new browser mapping; folder-reference WIP remains separate.

## Earlier terminal evidence: reading position (2026-09-22)

Typing and ordinary submission preserve existing transcript follow mode. Native delayed-provider completion, newer draft/cursor, same-row history anchors, resize round trips and explicit newest-edge navigation pass at **60×18, 100×22 and 140×36**, with **zero additional idle rows**. Full Go/vet, terminal races ×3, session/model/compaction/smoke/Gherkin tests, 29 helpers and 70/70 functional browser tests pass. See [ADR-0029](../adr/0029-terminal-reading-position.md). Delayed routed acceptance/recovery, reflow/eviction anchoring and durable transcript-ID reconciliation remain open. Frozen browser counts stay **34/236 Classic**, **2/42 shared**, **202/40 unmapped**; the complete 384/384 matrix below was run for `4071c01`.

## Latest browser evidence: accepted-message refresh (2026-09-22)

Acknowledgements refresh the selected timeline or search through the existing connection/view dispatcher. Native tests verify stored-message visibility/deduplication, three-file ID/name/byte association, newer draft/cursor preservation, near-bottom following, history-anchor retention and origin-session isolation. Supplied components remain unchanged. See [ADR-0028](../adr/0028-accepted-message-refresh.md).

Verified: **384/384 browser** (228 main + 156 specialised), **70/70 functional**, **29/29 helpers**, Go tests/vet and hook checks. The search regression rejects the old callback. Four new mappings: **compose-007/009/010/011**. Coverage: **34/236 Classic**, **2/42 shared**, **202/40 unmapped**. Compose-008 has no verified folder-reference selection path and stays unmapped. The compact terminal adaptation uses existing transcript/editor/notice surfaces with zero extra idle rows; terminal implementation and acceptance are separate work.

## Earlier browser evidence: bounded native timeline pages (2026-09-22)

Timeline now loads a native latest-50 page and older/forward pages using session-scoped timestamp/ID cursors. The host serialises page loads, preserves the loaded history window, drains reconnect catch-up without gaps and rejects stale pages after search/session changes. Actual visible-message anchors preserve reading position across older/live/offline arrivals, with at most one CSS pixel of rounding. ID-only completion events cannot replace full posts. Supplied components are unchanged. See [ADR-0027](../adr/0027-bounded-timeline-pages.md).

Verified: **54/54 reconnect/paging**, **348/348 combined browser**, **70/70 functional**, **29 helpers**, Go/vet/hook checks and native paging race ×3. Native tests cover ties, deleted/foreign/malformed cursors and bounds; the browser uses native turns/wheel input, real socket loss, multi-page catch-up and held-page/search isolation. No new frozen mapping: **30/236 Classic**, **2/42 shared**, 206/40 unmapped.

Unpaged export remains compatible. Backdated inserts, out-of-window edits/deletes and browser loaded-window eviction are not fully reconciled. Terminal paging adaptation is bounded on-demand loading through existing scroll controls, with no extra idle chrome; no terminal code or credit in this slice. Earlier evidence follows.

## Earlier browser evidence: initial refresh ownership (2026-09-22)

Initial activation and native SSE readiness now share one authoritative refresh owner per selection/connection epoch. No timeline/search/activity read is trusted before subscription readiness. Real reconnect, failed initial reads and A→B→A remain refreshable; rendered-selection wrappers also reject old callbacks before the hook rerenders. See [ADR-0026](../adr/0026-initial-refresh-ownership.md).

Verified: **48/48 reconnect/search**, **342/342 combined browser**, **70/70 functional**, **28 helpers**, Go/vet/hook checks. `reconnect-005` adds one mapping: **30/236 Classic** (206 unmapped), **2/42 shared** (40 unmapped). All five frozen reconnect IDs are mapped; this is not active-turn crash or hashtag/pagination acceptance. Exact request counts apply to the idle native test fixture only.

No terminal control or idle row is added. Local terminal subscription/reopen evidence remains separate. Earlier evidence follows.

## Earlier browser evidence: native search view (2026-09-22)

Search now uses a bounded native current/family/all-chat query and the supplied composer's independent search field. Query/session/connection generations protect results; active search blocks normal timeline HTTP refresh and unfiltered SSE appends. Reconnect refreshes search plus activity/queue/context while preserving draft/media. Escape restores the normal timeline and composer. See [ADR-0025](../adr/0025-native-search-view.md).

Verified: **36/36 reconnect/search**, **330/330 combined browser**, **70/70 functional**, **27 helpers**, Go/vet/hook checks and native search races ×3. `reconnect-003` adds one mapping: **29/236 Classic** (207 unmapped), **2/42 shared** (40 unmapped). Scope/literal-input/error/stale-query tests also pass. Native search is trimmed ASCII-case-insensitive substring matching; the UI caps results at 50. All chats means the existing authenticated workspace scope, not a new per-session ACL.

Hashtag navigation, paging, Unicode folding and terminal search remain open. Terminal design is an on-demand six-result selector using existing picker navigation/styles, with editor/cursor restoration and no added idle row; it has no implementation credit yet. Earlier sections retain historical evidence.

## Earlier browser evidence: reconnect and version drift (2026-09-22)

Reconnect `002/004` now pass across all six projects. Native activity/queue/context/timeline reload after real SSE loss; session/connection/request guards reject old timeline replies and late errors. The loaded script version is compared with the native connected envelope; real server restart produces one `New UI available` manual-reload notice without automatic navigation, even with a clean draft. See [ADR-0024](../adr/0024-reconnect-refresh-and-version-drift.md).

Verified: **18/18 reconnect**, **312/312 combined browser**, **70/70 functional**, **26 helpers**, full Go/vet/hook checks. Coverage: **28/236 Classic** (208 unmapped), **2/42 shared** (40 unmapped). Tests preserve drafts/media, reconcile new native work and reject delayed pre-disconnect responses. Search-specific reconnect, initial-refresh deduplication, full pagination and active-turn crash recovery remain open.

Terminal disposition: local Go TUI has no browser-version or SSE transport counterpart. No persistent terminal banner or idle row is added; session generation/reopen evidence remains separately documented below. Earlier sections retain historical evidence and limits.

## Earlier terminal evidence: compact maintenance controls (2026-09-22)

Terminal `/compact` and draft-preserving Alt-C now call the verified shared operation. `/compact info` retains diagnostics. Active elapsed status occupies the existing stats row; native outcome/error notices expire after four seconds. Focused Escape requests run-bound cancellation without clearing editor state. Lifecycle/suppression/failure events no longer render false success or generic compaction hook blocks. See [ADR-0023](../adr/0023-terminal-compaction.md).

Live **60×18, 100×22 and 140×36** native-hook PTY tests pass: empty/busy rejection, active/cancel/success, exact draft/cursor preservation, resize, command/info and checkpoint reopen, with **zero additional idle rows**. Full Go/vet, TUI races ×3, existing session/model PTY, seven-file Gherkin, smoke, hook checks and **70/70 web functional tests** pass.

Browser coverage is unchanged: **26/236 Classic**, **2/42 shared**, 210/40 unmapped. The earlier **294/294** browser matrix was not rerun for this terminal-only slice. Terminal pending-media/queue recovery, broader reconnect/crash and pathological storage-contention responsiveness remain open. Earlier sections record previous limitations.

## Earlier evidence: manual Compact (2026-09-22)

The meter now offers manual Compact only for a fresh eligible idle-session snapshot. Native admission atomically creates an empty-prompt maintenance turn, claim and submission event; it rejects stale/busy/repeated requests. Execution forces the existing checkpoint path without an inference call. Stop/cancellation and failed delivery preserve the original draft/media and checkpoint. Provider usage stays historical until another real request. See [ADR-0022](../adr/0022-manual-compaction.md).

Verified: **66/66 compaction browser**, **294/294 combined**, **70/70 functional**, **24 helpers**, Go/vet/hook checks and targeted race ×3. `context-003` adds one mapping: **Classic 26/236** (210 unmapped), **shared 2/42** (40 unmapped). Tests cover callback availability, success/reload, stale token/no turn, duplicate activation, failed delivery, cancellation/no user prompt, atomic admission and interrupted-operation recovery without replay. No supplied component/UI/pane changes.

The terminal `/compact` still displays information. Its command adaptation is specified, not implemented: call the shared native method, reuse transient status/cancel keys, preserve editor/cursor and add no idle rows. Broader reconnect/crash, summary quality and general hook/tool/media checkpointing remain open. Earlier sections record prior scope and capability limits.

## Earlier evidence: durable context checkpoints (2026-09-22)

Eligible automatic compaction now persists a session-local summary and covered message IDs/fingerprints atomically with its completion. Later turns/reopen project that checkpoint plus uncovered messages. The full timeline stays intact; edits/deletes to covered history invalidate the projection. Exact native-context/version/prefix guards reject races and avoid hiding hook/tool/media context that the text summary cannot safely cover. See [ADR-0021](../adr/0021-durable-context-checkpoints.md).

Verified: **54/54 compaction browser executions**, **282/282 combined**, **70/70 functional**, **24 helpers**, Go/vet/hook checks and targeted race ×3. A new native-provider browser regression checks next-request context and unchanged timeline records after reload. Three-size live terminal session/model regressions pass with zero added idle rows; no terminal UI changed. Native checkpoint reopen has store coverage, not live terminal checkpoint acceptance.

Mappings remain **25/236 Classic**, **2/42 shared** (211/40 unmapped). Manual Compact, general tool/multimodal checkpointing, large-history performance and full crash acceptance remain open. Historical usage is not rewritten to claim a reduction. Earlier evidence below records previous implementation limits.

## Earlier browser evidence: automatic compaction (2026-09-22)

Automatic compaction now renders elapsed/style/title state from a native activity snapshot. Stop targets the captured session/run without clearing text or attachments, is disabled while activity is unknown/stale/pending, and reconciles a completion race without cancelling new work. Suppression shows native detail; completion refreshes measured usage without requiring a reduction. See [ADR-0020](../adr/0020-automatic-compaction-web-status.md).

Verified: **48/48 new browser executions**, **276/276 combined**, **70/70 functional**, **24/24 helpers**, Go/vet/hook checks and targeted races ×3. Mappings: compaction `001–005` and context `004`; coverage **Classic 25/236** (211 unmapped), **shared 2/42** (40 unmapped). Native hook/provider tests include completion, stop, suppression, reload, delayed activity after switching sessions, failed Stop and completion-conflict reconciliation. No supplied component/UI/pane edits.

Manual Compact/context `003`, persisted history boundaries and broader reconnect/crash acceptance remain open. Terminal lifecycle adaptation is documented as a transient update within the existing footer/status and cancellation controls, with zero new idle rows; no terminal implementation or live evidence is added here.

## Native compaction prerequisite (2026-09-22)

Automatic compaction now commits an occurrence-keyed start and terminal outcome, summary and conditional phase restoration before publishing success. Cancellation cannot be overwritten by a running-status restore, stale completions cannot finish a later occurrence, and failed persistence prevents the next provider call. See [ADR-0019](../adr/0019-compaction-lifecycle-safety.md).

Full Go/vet, hook checks, targeted races repeated three times and 70/70 functional tests pass. Browser mapping totals are unchanged: **19/236 Classic**, **2/42 shared** (217/40 unmapped). The earlier **228/228** browser matrix remains the meter slice's evidence, not a new compaction matrix. No manual callback or compaction browser status is enabled; stored summaries still accompany original history on future turns. TUI layout is unchanged. A follow-up review delegate timed out and supplied no evidence.

## Latest browser evidence: context-meter rendering (2026-09-22)

Classic `@ux-context-001/005` pass all six projects using explicit usage returned by the local provider through native inference. Assertions cover K/M formatting, rounded percentages, title/tooltip-data/accessibility values, a full arc at 125% without hiding the percentage, zero input, 75/75.01/90/90.01 colour boundaries, reload, session-local unknown usage and model-change tooltip refresh.

The only production change is an app-level tooltip-data adapter using the existing container ref. The native title remains the source; a scoped observer catches child updates and cleans up with the render lifecycle. No added DOM parent, CSS or supplied component/UI/pane change. Details and review findings: [ADR-0016](../adr/0016-measured-request-context.md).

Current totals: **228/228 browser executions**, **70/70 functional**, **23/23 helper tests**, Go tests/vet and hook checks. Coverage: **Classic 19/236** passing (217 unmapped), **shared 2/42** passing (40 unmapped). `context-003/004` and full compaction remain unmapped.

Terminal adaptation uses the existing inline context segment and on-demand `/context` details. Unit tests verify zero/overflow measurements at widths 60/100/140 with no added or wrapped rows. The live session/model harness passes at **60×18, 100×22 and 140×36**. No terminal UI changed; unknown versus explicit zero remains an intentional current limitation, not web parity credit.

## Earlier evidence: measured model context fit (2026-09-22)

Classic `@ux-compaction-006/007` now pass all six projects using real local-provider usage through the native inference loop. A measured 100-token request blocks an 80-token model in the picker and API, permits 100/200-token models, and updates displayed capacity/percentage without changing the measurement or submitting a turn. Model selection and text/attachment drafts survive reload; unknown usage in another session stays independent. See [ADR-0016 follow-up](../adr/0016-measured-request-context.md).

Current totals: **216/216 browser executions**, **70/70 functional**, **23/23 helper tests**, Go tests/vet and hook checks. Coverage: **Classic 17/236** passing (219 unmapped), **shared 2/42** passing (40 unmapped). Full compaction and the other context-meter scenarios have no added credit. The local provider is deterministic, not a paid-provider or tokenizer-accuracy test. Supplied components/UI/panes and production runtime code are unchanged by this evidence slice.

Terminal context-fit rejection already uses the same validation. Keep the existing transient notice and editor/model state; add no idle rows. No new terminal acceptance is claimed by these browser tests.

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
