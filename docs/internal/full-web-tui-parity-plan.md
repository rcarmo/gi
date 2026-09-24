# Full web parity and compact terminal adaptations

Scope confirmed 2026-09-21: finish the whole imported corpus, not just the currently mapped session flows. Frozen acceptance criteria are unchanged.

## Completion gates

- Classic: 236 scenario IDs / 256 expanded cases from 24 pinned files. Every applicable browser case must pass Chromium and WebKit at phone, tablet and desktop sizes. Unsupported features remain gaps; do not mark them passed or silently omit them.
- Shared contract: 42 cases, inventoried and evidenced separately. Similar Classic tests do not automatically satisfy the shared contract.
- Native behaviour: browser interaction drives actions; APIs seed real records or verify persistence. Keep real shell execution/cancellation, late real-response tests, SQLite/WAL, append-only event sequences and one main session per agent.
- Terminal: implement suitable functional equivalents separately, preserving transcript/editor/footer and zero new idle rows. Browser-only properties (touch, PWA installation, CSS layering) need a documented terminal disposition, not a fictitious terminal pass.
- Delivery: small tested commits, source provenance, truthful capability/error paths and current evidence reports. Full completion requires no unexplained unmapped cases.

Latest Settings acceptance (2026-09-24): Settings002 independently verifies uncached immediate shell and first General resolution, with held native reads,503/retry and exact draft/media retention.216Settings browser/98standard functional/96helpers with1,175 assertions/Go-vet-hook/review pass. Coverage **96/236 Classic + 30/42 shared**, **140/12** unmapped. Test-only; no restart or terminal behaviour change. Details below.

Earlier outcome acceptance (2026-09-24): timeline026 passes native recovered final-response markers after the timestamp, on one metadata row, across six projects.54browser+1seededfunctional/98standardfunctional/96helpers with1,170 assertions, Go/vet/hook/recovery race x3 and sixTUI timing regressions pass. Coverage **95/236 Classic + 30/42 shared**, **141/12** unmapped. Timeout semantics and terminal outcome display are separate. Details below.

Earlier link acceptance (2026-09-24): timeline025 passes stored resource/preview navigation with isolated tabs, null opener/referrer, no image fetch and retained draft/media through reload/search across six projects.60browser,1seeded functional plus98standard functional,95helpers/1,159 assertions and Go/vet/hook pass. Coverage **94/236 Classic + 30/42 shared**, **142/12** unmapped. No URL discovery or terminal change. Details below.

Earlier terminal tool adaptation (2026-09-24): existing inline tool headers now use turn+call identity; retained completed/error/skipped blocks reject late starts and duplicate terminal events, and absent start timing stays unknown. Six timing PTYs plus18media/session/model regressions,98functional/6browser/92helpers, Go/vet/hook and targeted TUI race x3 pass. Zero new idle rows or controls; coverage remains **93/236 Classic + 30/42 shared**, **143/12** unmapped. Details below.

Earlier tool-status slice (2026-09-24): original027 passes native identity/name/preview, one-second elapsed time and terminal timing/state across six projects. SQLite snapshot ownership and invalidation-only SSE protect newer calls.60tool-queue-context+96compaction+48thought browser,98functional,92helpers/1,129 assertions, Go/vet/hook/targeted race x3 and six terminal media PTYs pass. Coverage **93/236 Classic + 30/42 shared**, **143/12** unmapped. Shared40 pane lifecycle stays open; no terminal UI change. Details below.

Earlier speech acceptance (2026-09-24): timeline027 verifies native empty/whitespace assistant rows and synthesis capability; original028 and shared42 independently combine trusted code clipboard bytes with playback ownership and stale callbacks.90browser/97functional/91helpers with1,115 assertions/Go-vet-hook/review pass. Coverage **92/236 Classic + 30/42 shared**, **144/12** unmapped. Test-only; no restart, physical-audio or terminal change. Details below.

Earlier speech slice (2026-09-24): native assistant Read aloud/Stop follows the frozen Piclaw control through a guarded build adapter. Timeline028 ownership transfer passes six projects, including synchronous cancel and stale callbacks.72browser/97functional/91helpers with1,103 assertions/Go-vet-hook pass. Coverage **90/236 Classic + 29/42 shared**, **146/13** unmapped. Browser speech boundary is controlled; physical audio and combined copy/speech cases are unverified. No terminal change. Details below.

Earlier Classic acceptance (2026-09-24): original008 independently passes the startup-loaded skill catalogue, description filtering, Slash commands-only grouping and shared compose-prefill path across six projects. Classic/shared scenarios execute separately; the reporter remains unchanged.18browser/96functional/87helpers with1,069 assertions/Go-vet-hook pass. Coverage **89/236 Classic + 29/42 shared**, **147/13** unmapped. Test-only; no restart or terminal change. Candidate gaps below.

Earlier shared acceptance (2026-09-24): shared17 passes with startup-loaded `/skill:<name>` entries, draft-preserving insertion, captured-session native expansion and recoverable unknown/stale commands. Final12 browser cases across six projects, session126/QuickActions138 regressions, functional96/helpers87 (1,062 assertions), Go/vet/hook and targeted race x3 pass. Coverage **88/236 Classic + 29/42 shared**, **148/13** unmapped. Terminal invocation is still load-only. Details and limitations below.

Earlier browser repairs (2026-09-24): read-only Markdown inline code uses the configured mono stack; horizontal table scrollers keep their gestures instead of switching Safari sessions. 210 combined browser +30 swipe regressions, 95 functional on clean unchanged rerun, 86 helpers/1,047 assertions and Go/vet/hook pass. Shell009 precondition and shared41 SVG conflict remain gaps; no mapping/TUI change. Details below.

Earlier shared acceptance (2026-09-24): shared39 now passes the attach-file route across all six browser projects: visible upload, cancel retaining draft, explicit retry reusing the native media ID, one media item/user message, original file unlink and reload/raw-byte/render durability. Multipart uploads atomically reuse exact session/name/MIME/bytes. 198 browser/94 functional/85 support tests with 1,039 assertions, Go/vet/hook, store+web race ×3 and concurrent process/store proof pass. Coverage **88/236 Classic + 28/42 shared**, **148/14** unmapped. No general message idempotency or new TUI credit; details below.

Earlier browser repair (2026-09-24): session-owned Cancel uploads aborts captured media batches before message dispatch and restores exact draft/media bytes through existing recovery. Other sessions, newer typing and already-dispatched sends remain independent. 30 final focused/204 regression browser cases, 94 functional and 85 support tests/1,034 assertions pass; Go/vet/hook and bounded review pass. Shared39 remains open; coverage **88/236 + 27/42** unchanged. Details below.

Earlier terminal adaptation (2026-09-24): `/attach` and `/paste-image` stage up to six process-local media references per session for the next ordinary prompt. `/attachments` and `/detach` provide explicit review/removal with no idle row. Native pre/post-admission failure tests, six PTYs, existing TUI regressions, race ×3, 93 functional and 82 support tests pass. No browser coverage change; restart durability and queued-draft media recovery remain open. Details below.

Earlier browser acceptance (2026-09-24): read-only tabs use the supplied tab store's MRU and pin semantics. Workspace-011 passes all six projects with pinned-before-MRU, bulk-close protection and exact draft/media retention; 186 combined browser, 93 functional and 82 support tests/1,004 assertions pass alongside Go/vet/hook checks. Coverage is **88/236 Classic + 27/42 shared**, leaving **148/15** unmapped. No editing, dirty/save, durable pin or terminal credit.

Earlier terminal selector fix (2026-09-24): Alt-S uses the existing bounded temporary screen in regular mode; resize cannot leave selector rows in history. Captured-generation acceptance keeps failed switches open for retry, and closes only after success. Six session-picker PTYs plus six model-picker regressions, existing TUI suites,92functional/75helpers/Go-vet-build-hook/TUI race ×3 pass. Browser coverage remains **87/236 Classic**, **27/42 shared**; no idle rows added.

Earlier read-only tab slice (2026-09-24): the formerly empty workspace tab host now mounts native bounded previews with loading, retry and stale-read/disposal guards. Shell007 background close preserves active content; pointer/keyboard/mobile/focus/draft checks pass: **18/18** focused, **168/168** regressions, **92/92** functional, **75/75** helpers, Go/vet/build/hook. Coverage **87/236 Classic**, **27/42 shared**, unmapped **149/15**. No editing, dirty-tab, pin/pop-out or terminal credit.

Earlier Settings acceptance (2026-09-24): Settings001 native menu/shortcut/header/navigation and dialog002 cached reopen within1s pass independently across six browser projects, including a held fresh read and exact draft/media retention. **12/12** focused, **234/234** regressions, **91/91** functional, **72/72** helpers, Go/vet/build/hook. Coverage **86/236 Classic**, **27/42 shared**, unmapped **150/15**. Test-only; terminal settings evidence separate.

Earlier host fix (2026-09-24): standalone display-scale control visibility now shares the supplied navigator/display-mode detection. Capability-gated native control, storage, viewport and cross-tab tests pass: **168/168** browser regressions, **91/91** functional, **72/72** helpers, Go/vet/build/hook and review. Shell008 remains unmapped because the frozen wording omits the standalone precondition. Coverage unchanged **84/236 Classic**, **27/42 shared**; no physical-PWA or terminal credit.

Earlier native shell acceptance (2026-09-24): Classic shell002/003/005 independently verify hidden-setting events/persistence, disabled workspace controls while hidden, and composer available-column geometry. **18/18** focused, **162/162** regressions, **90/90** functional, **68/68** helpers, Go/vet/build/hook and review pass. Coverage **84/236 Classic**, **27/42 shared**, unmapped **152/15**. Test-only; terminal evidence remains separate.

Earlier terminal slice (2026-09-24): Alt-M metadata search, enabled-only navigation, unavailable retry and captured-session generation checks pass in six fullscreen/regular PTYs at60×18,100×22,140×36. Regular mode uses a temporary screen to keep resized selector fragments out of scrollback. Draft/cursor, multiline editor, history and idle rows are retained. Existing TUI suites, Go/vet/build/hook/race ×3,90functional/68helpers pass. Browser coverage unchanged **81/236 Classic**, **27/42 shared**; deployment below.

Earlier reliability slice (2026-09-24): native terminal SSE carries turn identity and stale terminal frames cannot clear a newer run's controls/previews. Reconnect **66/66**, functional **90/90**, helpers **68/68**, Go/vet/build/hook, race ×3 and existing three-size TUI suites pass. Coverage unchanged: **81/236 Classic**, **27/42 shared**. Shared36 remains unmapped because cancellation advances the queue; deployment below.

Earlier model-filter acceptance (2026-09-24): shared34 adds native metadata search, enabled-only navigation and prefix-first typeahead with actual button focus. Models **42/42**, fresh-project picker/Quick Actions/Classic regressions **294/294**, functional **89/89**, helpers **67/67**, Go/vet/build/hook pass. Classic **81/236**, shared **27/42**, unmapped **155/15**. Supplied sources remain unchanged; terminal Alt-M acceptance is separate. Deployment below.

Earlier session-typeahead acceptance (2026-09-24): shared33 adds non-search incremental matching to native session entries, preferring enabled prefixes over pinned substrings and focusing the actual button for Enter. Search/grouping, Arrow/Home/End/Page keys, cancellation and drafts pass. Sessions **126/126**, isolated model/Quick Actions/menu regression **204/204**, functional **88/88**, helpers **64/64**, Go/vet/build/hook pass. Classic **81/236**, shared **26/42**, unmapped **155/16**. Deployment below; terminal Alt-S evidence separate. Shared16 has an explicit frozen command-prefill conflict.

Earlier shared ranking acceptance (2026-09-24): shared4 verifies native idle typing, initial-query and full groups, exact/prefix/multiple-result fallback ranking, filtered arrow wrap and highlighted workspace activation once with exact durable drafts/references intact. Browser **258/258**, functional **87/87**, helpers **62/62**, Go/vet/build/hook pass. Classic **81/236**, shared **25/42**, unmapped **155/17**. Test-only, no restart; shared16 and terminal activation acceptance remain separate.

Earlier close-control acceptance (2026-09-24): shared15 adds a compact accessible Close button to the transient palette header. Pointer, Tab/Enter/Space and trusted touch preserve draft/media/references and restore Conversation focus. Native action-row buttons own their activation keys; layout-effect keyboard setup fixes rapid reopen. Browser **252/252**, functional **87/87**, helpers **62/62**, Go/vet/build/hook pass. Classic **81/236**, shared **24/42**, unmapped **155/18**. Upstream source/CSS bytes unchanged; deployment below, no idle or terminal rows.

Earlier shared dismissal acceptance (2026-09-24): shared13/14 verify Escape and outside-click focus restoration to the native Conversation region, with exact stored text/media bytes/file/message references preserved. Guarded adapter cancels late focus frames; trusted-click consumption keeps Alt+Enter programmatic links intact. Browser **234/234**, functional **87/87**, helpers **61/61**, Go/vet/build/hook pass. Classic **81/236**, shared **23/42**, unmapped **155/19**. Deployment below; shared15 close-control and terminal acceptance remain separate.

Earlier native target-key acceptance (2026-09-24): shared7/9/10 reproduce and fix swallowed timeline-control and Settings input keydowns. Palette exclusions are non-consuming predicates; modal targets receive keys while background popups suspend input. Composing Escape, Tab wrapping and pointer isolation pass. Browser **216/216 + 108/108**, functional **86/86**, helpers **58/58**, Go/vet/build/hook pass. Classic **81/236**, shared **21/42**, unmapped **155/21**. Deployment below; terminal input ownership remains separate.

Earlier shared typing acceptance (2026-09-24): shared5/6/11/12 verify normal native composer Unicode editing, Settings input filtering, session-picker search and prevented/repeated/IME/modified timeline keys. Positive real-key palette activation precedes exclusions and follows every rejected event; native snapshots and unsent media remain intact. Combined **192/192**, functional **85/85**, helpers **55/55**, Go/vet/build/hook pass. Classic **81/236**, shared **18/42**, unmapped **155/24**. Test-only; terminal input ownership acceptance remains separate.

Earlier shared menu acceptance (2026-09-24): shared1/2 reproduced and fixed outside-click draft submission and Escape focus loss. A guarded build adapter consumes outside gestures through their click and restores the connected trigger; supplied component bytes stay unchanged. Combined menu/workspace/Quick Actions/Settings **192/192**, functional **85/85**, helpers **55/55**, Go/vet/build/hook pass. Classic **81/236**, shared **14/42**, unmapped **155/28**. Runtime deployment recorded below; terminal UI unchanged.

Earlier shared workspace acceptance (2026-09-24): shared3 verifies native menu/tree visibility and hide, exact persisted composer content/session preservation, and real pointer blocking over textarea/Send beneath the narrow drawer backdrop. Workspace **24/24**, functional **85/85**, helpers **51/51**, Go/vet/build/hook pass. Classic **81/236**, shared **12/42**, unmapped **155/30**. Test-only; no Plan or terminal feature credit. Shared39 stays unmapped because upload has no exposed Cancel control.

Earlier shared capability acceptance (2026-09-23): shared26 verifies native supported session actions, absent unsupported deletion for root/running/unknown-count selections, native405/409 errors and recoverable rename400/retry. Pin/archive/restore/unpin and child creation persist without losing drafts/media. Sessions **120/120**, functional **85/85**, helpers **51/51**, Go/vet/build/hook pass. Classic **81/236**, shared **11/42**, unmapped **155/31**. Test-only; terminal mutation submenus still require independent implementation and acceptance.

Earlier shared FIFO acceptance (2026-09-23): shared27 verifies two native composer follow-ups with exact text/file/folder/message references, one uploaded media record per follow-up and downloaded source bytes. Held POST acknowledgement/SSE reconciliation, distinct submission tokens, reload, eventual FIFO history and other-session isolation pass. Queue **48/48**, functional **85/85**, helpers **51/51**, Go/vet/build/hook pass. Classic **81/236**, shared **10/42**, unmapped **155/32**. Test-only; compact terminal queue/media implementation and acceptance remain separate.

Earlier coherent-session acceptance (2026-09-23): shared25 verifies keyboard selection with five held native reads from the previous session. Exact timeline IDs, queued IDs/text, model/context and draft/media survive each late response, reload and return to the origin. Context/shared **36/36**, sessions **114/114**, functional **85/85**, helpers **51/51**, Go/vet/build/hook pass. Classic **81/236**, shared **9/42**, unmapped **155/33**. Test-only; compact Alt-S and terminal queue/media acceptance stay separate.

Earlier shared model acceptance (2026-09-23): shared31/32 verify pointer/keyboard opening, native incremental typeahead, held server-confirmed model/context switch and persisted session-only selection with text/media/file/exact-message refs unchanged. Six-project context/shared **30/30**, functional **85/85**, helpers **51/51**, Go/vet/build/hook pass. Classic **81/236**, shared **8/42**, unmapped **155/34**. Test-only; existing terminal Alt-M acceptance remains separate.

Earlier shared picker acceptance (2026-09-23): shared23/24 pointer and keyboard opening verify native insertion/first-frame search focus, both trigger buttons, composer anchor/resize and Escape focus return with no draft/session/turn changes. Dedicated **12/12**, full sessions **114/114**, functional **85/85**, support **51/51**, Go/vet/build/hook pass. Classic **81/236**, shared **6/42**, unmapped **155/36**. Test-only; existing Alt-S terminal evidence stays separate.

Earlier shared queue acceptance (2026-09-23): shared29 verifies actual adjacent reorder payload/persisted row equality, native consumed-active DELETE409 reconciliation, unchanged second-session queue and two drafts, reload and eventual FIFO history. Dedicated **6/6**, full queue **42/42**, functional **85/85**, helpers **51/51**, Go/vet/build/hook pass. Classic unchanged **81/236**, shared **4/42**, unmapped **155/38**. Test-only; terminal queue actions remain design-only.

Earlier wrapped-preview slice (2026-09-23): Thoughts001 passes real native single-paragraph overflow/disclosure across six projects. Guarded build-time renderer adapter keeps supplied source bytes/counts intact; generic Show more follows actual collapsed DOM clipping, ResizeObserver/stream/font/resize with noRO ancestor mutation fallback. Preview **48/48**, status **12/12**, functional **85/85**, support **51/51**, Go/vet/build/hook pass. Classic **81/236**, shared **3/42**, unmapped **155/39**. No terminal controls/idle rows.

Earlier streamed preview slice (2026-09-23): native buffers now use supplied string normalization; component-owned disclosures work for multiline streams, keyed session/turn resets with tagged-delta fencing and activity-read restoration. Gi-only status region is bounded to40% with scrolling. Frozen Thoughts002–005 pass; Thoughts001 stays unmapped for the supplied long-single-line disclosure limitation. Preview **36/36**, status **12/12**, reconnect **60/60**, functional **85/85**, support **49/49**, Go/vet/build/hook pass. Classic **80/236**, shared **3/42**, unmapped **156/39**; no TUI rows.

Earlier status passthrough slice (2026-09-23): frozen `@ux-mobile-003` native streaming draft links plus Gi thought links navigate through the supplied thinking/status eligibility rule. Conversation-host boundary excludes composer/queue/search and detached modals; pointerdown/touchstart/wheel share the boundary. Native panel **12/12**, full sessions **102/102**, Steer **18/18**, functional **85/85**, support **48/48**, Go/vet/build/hook pass. Classic **76/236**, shared **3/42**, unmapped **160/39**. No new terminal interaction or idle rows; mobile002 and preview expansion remain gaps.

Earlier rapid-swipe correction (2026-09-23): host native swipe attachment now uses layout effect so next-frame reverse gestures see the committed selected chat; supplied helper and selection/draft logic unchanged. Derived Gi swipe001/002 reproduce pre-fix failure and pass **6/6**, session suite **102/102**, functional **84/84**, helpers **48/48**, Go/vet/build/hook on full rerun. Initial Go steering fixture/repeat flake is recorded below. Frozen counts unchanged **75/236 + 3/42**, unmapped **161/39**; browser-only fix, no terminal rows.

Earlier mobile ordering acceptance (2026-09-23): frozen `@ux-mobile-004` passes native pinned/active/ordinary overlap, active-first/JID tie order, archived omission, bidirectional navigation/wrap and pin-independent order across all six projects. Explicit duplicate candidates are tested directly through the supplied resolver; persisted API catalogue IDs are unique. Dedicated **6/6**, swipe regression **30/30**, full session suite **96/96**, functional **83/83**, helpers **47/47**, Go/vet/build/hook pass. Classic **75/236**, shared **3/42**, unmapped **161/39**. Test-only change; terminal Alt-S evidence stays separate. See the ordering evidence section below for the activation timing limit.

Earlier Models freshness slice (2026-09-23): local model request settlement invalidates only mounted consumers of its captured session; authoritative GETs reject older generations and preserve newer filter/selection drafts. Lost-write/read-failure recovery retains action errors; explicit Refresh supports external changes. Browser **234/234**, functional **83/83**, helpers **46/46**, Go/vet/build/hook/HTTP graph pass. Keyed session remount verified in review. Frozen counts unchanged **74/236 + 3/42**, unmapped **162/39**; no TUI chrome or extra credit.

Earlier Settings header slice (2026-09-23): frozen `@ux-settings-004` header filtering/focus and pinned dialog-width 860/720 compact/narrow classes verified; query/selection/native catalogue remain stable across resize, per-visit Apply locking survives late responses. Shell **90/90**, combined shell/Gi/models **216/216**, functional **83/83**, support **44/44**, Go/vet/build/hook pass. Classic **74/236**, shared **3/42**, unmapped **162/39**. Existing bounded Alt-M is the terminal adaptation, no new rows or inferred credit. The model commit-after-reentry gap found in this review is closed by the later freshness slice above.

Earlier lazy Settings slice (2026-09-23): non-General panes now use on-demand hashed modules, bounded loading/error UI and per-document code cache; General is static and native data remains freshly scoped. Bootstrap/hashed-app graph prevents duplicate app initialisation. Frozen `@ux-settings-dialog-005` and `@ux-settings-003` pass real chunk-request/held-import/cache browser proof. Stable bootstrap caching is explicitly disabled; embedded HTTP graph and six-browser missing-chunk/explicit-reload recovery pass. Shell **72/72**, shell/Gi/model regression **186/186**, compaction **96/96**, Providers **36/36**, functional **83/83**, helpers **44/44**, Go/vet/build/hook pass. Classic **73/236**, shared **3/42**, unmapped **163/39**. Cached Settings data remains a gap; no terminal code, idle rows or inferred credit.

Earlier frozen Settings shell slice (2026-09-23): `@ux-settings-layering-001`–`004` and `@ux-settings-dialog-001/003/004` pass six-project browser acceptance. Real pointer clicks cannot reach a seeded workspace file beneath the body portal; dismissal restores native preview reads. Fixed geometry, exact half-opaque backdrop, single rapid-shortcut dialog, immediate loading/General-first timing and typed numeric fields are verified. Modal focus/inert/Escape setup now runs in the layout effect to close a first-paint Escape race; rapid-open tablet stress **10/10**, frozen **42/42**, shell+Gi Settings **132/132**, functional **83/83**, helpers **43/43**, Go/vet/build/hook pass. Classic **71/236**, shared **3/42**, unmapped **165/39**. Cached reopen (`dialog-002`) and lazy imports (`dialog-005`) remain gaps. Layering is browser-only; no terminal rows or inferred terminal credit.

Earlier Gi Providers slice (2026-09-23): explicit OpenAI/Anthropic API-key save/confirmed removal uses a private atomic home-scoped credential writer shared by native logout. Metadata-only OAuth/custom rows, per-process HMAC revision, authenticated JSON/same-origin/TLS-or-loopback-peer+Host writes; stored never means remotely verified. Twenty-three Gi scenarios plus browser-OAuth proposal. Isolated native provider **36/36**, Settings/Models **126/126**, functional **83/83**, helpers **43/43**, Go/vet/build/hook/inference-web race ×3 and Darwin compile pass. Frozen **64/236 Classic**, **3/42 shared**, **172/39 unmapped** unchanged; zero new terminal rows. See [Gi settings](gi-settings-plan.md#provider-api-key-setup--2026-09-23).

Earlier Gi automatic-policy save slice (2026-09-23): saved enablement/token budgets use a revision-checked atomic Pi settings path shared by all native model/TUI preference writers. The active engine stays unchanged until manual restart; conflicts/failures preserve the form. Nineteen Gi scenarios and one provider proposal, no frozen credit. Settings/Models **126/126**, compaction **96/96**, functional **83/83**, helpers **43/43**, Go/vet/build/hook/race ×3, Darwin config compile and three-size existing TUI compaction acceptance pass. Frozen **64/236 Classic**, **3/42 shared**, **172/39 unmapped** unchanged; zero new terminal rows. See [Gi settings](gi-settings-plan.md#saved-automatic-compaction-policy--2026-09-23).

Earlier Gi Compaction settings slice (2026-09-23): read-only effective engine startup policy plus token-bound Compact and matching-turn Stop use existing authenticated native APIs. Authoritative refresh continues during held POSTs, without enabling duplicate actions or leaking old-session responses. Seventeen Gi scenarios plus one future proposal; native compaction **96/96**, Settings/Models **108/108**, functional **83/83**, helpers **43/43**, Go/vet/build/hook/web-turn race ×3 pass. Frozen **64/236 Classic**, **3/42 shared**, **172/39 unmapped** unchanged. Terminal `/compact`/Alt-C retains separate earlier evidence and zero new idle rows. See [Gi settings](gi-settings-plan.md#compaction-inspection-and-session-actions--2026-09-23).

Earlier Gi instance identity slice (2026-09-23): General explicitly saves only display names through a revision-checked native endpoint. Active names stay unchanged until manual restart; invalid input, conflicts and storage failures preserve the draft. Unknown configuration keys/avatars survive atomic replacement. Thirteen Gi scenarios and one future proposal are separate from frozen parity. Settings **72/72**, Settings/Models **102/102**, reviewed identity **18/18**, functional **83/83**, support **43/43**, Go/vet/build/hook and config/web race ×3 pass. Frozen **64/236 Classic**, **3/42 shared**, **172/39 unmapped** unchanged. No new terminal controls or idle rows. See [Gi settings](gi-settings-plan.md#saved-instance-names--2026-09-23).

Earlier Gi browser Appearance slice (2026-09-23): explicit preset/tint save/reset uses one browser-local record and the unchanged supplied renderer. Storage failures leave the displayed palette unchanged; tabs synchronise without discarding dirty fields. Eleven Gi scenarios plus two future proposals remain outside frozen parity. Settings/models **90/90**, functional **83/83**, support **43/43**, Go/vet/build/hook pass. Classic **64/236**, shared **3/42**, unmapped **172/39** unchanged. Browser theme preferences never change the TUI or add idle rows. See [Gi settings](gi-settings-plan.md#browser-appearance-contract--2026-09-23).

Earlier Gi-specific settings slice (2026-09-23): derived General/Models settings use Piclaw's body-portal layout and Gi's existing authenticated APIs. General labels startup instance values read-only; Models filters a bounded native catalogue and explicitly applies to the captured session after server confirmation, with context-fit and stale-response guards. Eight Gi scenarios and three future proposals are separate from frozen parity. Settings **36/36**, native catalogue/context **24/24**, model/session regression **156/156**, functional **82/82**, support **40/40**, Go/vet/build/hook and auth/model race ×3 pass. Frozen Classic **64/236**, shared **3/42**, unmapped **172/39** unchanged. Terminal Alt-M/`/model` remain separate and add no idle rows. See [Gi settings](gi-settings-plan.md).

Earlier static PWA manifest slice (2026-09-23): `@ux-pwa-001` serves a linked `/manifest.json` with Gi's configured assistant name and any/maskable 192/512 PNG records. The exact declared `/static/icon-*.png` URLs now resolve to embedded PNG bytes, with GET/HEAD headers checked in Go and six Chromium/WebKit phone/tablet/desktop browser projects; swipe regression **30/30**, functional **81/81**, support **39/39**. Classic **64/236** (**172 unmapped**), shared **3/42** (**39 unmapped**). Browser-controlled home-screen installation has no terminal analogue or TUI chrome. Avatar-version and touch/favicons (`@ux-pwa-002`–`006`) have no frozen credit from this static manifest.

Earlier Safari wheel-exclusion slice (2026-09-23): `@ux-mobile-006` wires the existing Safari detector into Gi's supplied native timeline swipe listener. A horizontal wheel event cannot switch sessions in Chrome or iOS modes; the same listener/delta selects the adjacent persisted session on desktop Safari as a positive control. Six Chromium/WebKit phone/tablet/desktop projects pass, picker regression **90/90** after explicitly waiting for filtered picker selection to settle, and swipe-family proof **24/24**. Classic **63/236** (**173 unmapped**), shared **3/42** (**39 unmapped**). Browser-only wheel gesture, no new TUI chrome; explicit terminal Alt-S remains separate. `@ux-mobile-002/003/004` still require native browser evidence.

Earlier vertical swipe-cancellation slice (2026-09-23): `@ux-mobile-005` verifies the native iOS timeline listener cancels a contact after a primarily vertical first move (>16 px), even when later movement exceeds horizontal distance, then permits a fresh eligible horizontal contact to select the adjacent persisted non-archived session. Six browser projects pass and picker regression **84/84**; Classic **62/236** (**174 unmapped**), shared **3/42** (**39 unmapped**). Touch is browser-only: terminal Alt-S remains separately verified, explicit and bounded, with no idle UI rows. `@ux-mobile-002/003/004/006` remain unmapped pending their own browser evidence.

Earlier mobile swipe-wrap slice (2026-09-23): `@ux-mobile-001` verifies an eligible horizontal finger gesture on the supplied timeline selects the next native non-archived session and wraps from the catalogue's last entry to its first. It uses the full persisted catalogue because six Playwright projects share one test server; the original draft returns when its session is reselected. Six-project proof plus **78/78** picker regression pass. Classic **61/236** (**175 unmapped**), shared **3/42** (**39 unmapped**). Touch remains browser-only; terminal Alt-S keeps its separately verified six-result, zero-idle-row path.

Earlier reconnect/Stop slice (2026-09-23): `@ux-original-023` uses a real gated turn and SSE outage. After reconnect, the selected timeline, activity and queue are fetched from the native server; the visible Stop posts the captured active turn ID and waits for cancellation to appear in authoritative turn state. The older completed run cannot retake the selected control, and the draft survives. Six browser projects pass; reconnect regression **60/60**. Classic **60/236** (**176 unmapped**), shared **3/42** (**39 unmapped**). Terminal `/cancel` is an explicit, no-idle-row action with separate evidence; browser reconnection earns no terminal credit.

Earlier sparse model-metadata slice (2026-09-23): `@ux-original-022` verifies a real enabled synthetic model without context-window or reasoning metadata. The supplied picker renders only its reported name, the selected chat's context meter reports `?` usage with unavailable provenance, and a delayed model catalogue from the superseded chat cannot replace the target's model or draft. Six browser projects pass with **36/36** model regression; Classic **59/236** (**177 unmapped**), shared **3/42** (**39 unmapped**). Terminal model labels and the measured footer remain separately verified; no new idle rows or terminal credit from browser evidence.

Earlier model-picker selection slice (2026-09-23): `@ux-original-020` uses the unchanged supplied picker and native session-scoped PATCH. A captured model choice returns its current model and context fields; accepted selection persists across reload, keyboard selection follows the same path, unavailable models report a 400 and leave the draft, attachments and session model untouched. The six-project test and **30/30** model regression pass. Classic **58/236** (**178 unmapped**), shared **3/42** (**39 unmapped**). Existing terminal Alt-M and `/model` are separately verified with no new idle rows; this browser result adds no terminal credit.

Earlier touch swipe slice (2026-09-23): `@ux-session-005` attaches the supplied iOS timeline swipe listener from Gi's adapter, using the persisted catalogue for archived exclusion and active-first/JID order. Native interactive controls and selected timeline text cannot start a swipe; session changes keep drafts isolated. Six Chromium/WebKit phone/tablet/desktop projects pass, picker regression **72/72** and frozen session family **36/36**. Classic **57/236** (**179 unmapped**), shared **3/42** (**39 unmapped**). Touch is a browser-only gesture: terminal Alt-S remains the bounded Pi-style six-result selector, triggered explicitly, with no idle rows or swipe emulation.

Earlier native grouping picker slice (2026-09-23): `@ux-session-002` uses persisted session metadata to place one entry in each supplied Current, Pinned, Active, This session tree, Other sessions and Archived group. The current row alone has `aria-current=true`; a live gated turn supplies active state, while native pin/archive mutations supply their groups. A six-project pass and **66/66** picker regression leave the selected session and draft unchanged. Classic **56/236** (**180 unmapped**), shared **3/42** (**39 unmapped**). Terminal Alt-S remains its separately verified bounded six-result selector; no browser result grants new terminal credit or adds idle rows.

Earlier archive/restore picker slice (2026-09-23): `@ux-session-004` uses unchanged supplied picker controls and native child-session PATCH callbacks. A held accepted archive response cannot regroup or announce success early; restore regroups from the authoritative catalogue. A rejected archive transport reports the failure without moving the row, changing the selected chat or losing its draft. Six browser projects pass; picker regression **60/60**, Classic **55/236** (**181 unmapped**), shared **3/42** (**39 unmapped**). Terminal mutation submenus still require their own compact implementation and three-size evidence.

Earlier selected-session timeline slice (2026-09-23): `@ux-session-001` switches through the visible supplied picker between two native persisted histories. A held real response from the superseded chat cannot replace the selected timeline or draft; returning restores origin history/draft. Six browser projects pass, picker suite **54/54**, functional **80/80**, helpers **39**, Go/vet/build/hook checks pass. Classic becomes **54/236** (**182 unmapped**); shared stays **3/42** (**39 unmapped**). Existing terminal Alt-S selection/generation tests remain separate evidence; this browser case grants no new terminal credit.

Earlier session-picker metadata search slice (2026-09-23): `@ux-session-003` passes six projects against native session IDs, a persisted session-local model change (refreshed by the existing bounded poll) and two-term handle search; ArrowUp/Down wraps only matching rows. Draft stays in its origin until Enter switches, then returns intact. Picker regression **48/48**, functional **80/80**, helpers **39**, Go/vet/build/hook pass. Classic rises to **53/236** (**183 unmapped**), shared stays **3/42** (**39 unmapped**). Existing compact Alt-S has separate terminal acceptance; browser search grants no new terminal credit.

Earlier session-picker dismissal slice (2026-09-23): `@ux-session-006` uses the unchanged supplied picker. Native Escape clears search/typeahead and restores each initiating trigger without switching sessions; repeated reopen, stored history and unsent draft pass in six Chromium/WebKit phone/tablet/desktop runs. Classic becomes **52/236** passing (**184 unmapped**), shared stays **3/42** (**39 unmapped**). Existing Alt-S terminal dismissal has separate acceptance; this browser case adds no terminal credit.

Earlier terminal latest-assistant source-copy slice (2026-09-23): existing `/copy` sends exact stored assistant bytes under native/OSC 52 opt-in, including leading/trailing whitespace and Unicode. Six native fullscreen/regular PTYs at 60×18/100×22/140×36 verify OSC 52 byte payloads, fallback, draft retention and zero idle-row growth; Go/vet and existing TUI regressions pass. [ADR-0058](../adr/0058-terminal-latest-assistant-source-copy.md). The current-message picker and terminal deletion are still design-only: ordinary transcript blocks lack stable message IDs, and regular scrollback is terminal-owned. Browser mappings remain **51/236 Classic**, **3/42 shared**.

Earlier integrated copy/delete shared-contract slice (2026-09-23): `@shared-37` passes six Chromium/WebKit phone/tablet/desktop runs with stored Markdown/code clipboard bytes, success/error glyph reset, a real busy-run native 409 and later captured-ID removal. Another session, an unrelated message reference and the draft/media stay intact after reload. Classic stays **51/236** passing (**185 unmapped**); shared rises to **3/42** passing (**39 unmapped**). Original-024 still requires reply-cascade confirmation; terminal source-copy and deletion remain two design-only actions with no added idle rows or terminal credit. See [ADR-0057](../adr/0057-integrated-message-copy-delete.md).

Earlier single-message deletion slice (2026-09-23): native idle-only transactional deletion and success-owned web animation verify `@ux-timeline-017` across six browser projects with delayed timeline/search responses, failure, session origin, draft/media and reload checks. Store/web tests verify checkpoint invalidation, rollback and audit/media retention. **450/450** main browser executions (six isolated 75-case project runs), **80/80** functional, **39** helpers, Go/vet/race and hook checks pass; separate reconnect/context/compaction/Steer results complete the aggregation. [ADR-0056](../adr/0056-idle-single-message-deletion.md). This adds **one Classic mapping**: **51/236 Classic**, **2/42 shared**, **185/40 unmapped** after the full matrix. Reply detection/cascade (`timeline-018`–`022`) and Classic combined copy/delete (`original-024`) stay open; shared-37 is verified separately above. Terminal selected-message action is design-only with no idle-row growth or terminal credit.

Earlier clipboard slice (2026-09-23): native whole-message Markdown/rich copy and failure paths verified; absent Clipboard API false success fixed through a narrow build adapter. **594 browser**, **80 functional**, **38 helpers**, Go/vet/build/hook pass. Combined copy/delete remains incomplete, so coverage stays **50/236 Classic**, **2/42 shared**, **186/40 unmapped**. [ADR-0055](../adr/0055-message-copy-clipboard-safety.md). Terminal source-copy and rendered-row selection remain distinct; no new terminal credit.

Earlier Quick Actions slice (2026-09-23): pinned component/helper/styles plus native capability catalogue and host typing/prefill guards verify original-003/005/006/007. **564 browser**, **79 functional**, **36 helpers**, Go/vet/build/hook and web race×3 pass; additional 48 capture and 24 transition stress checks pass. Coverage is **50/236 Classic**, **2/42 shared**, **186/40 unmapped**. [ADR-0054](../adr/0054-native-quick-actions.md). Full exclusion surfaces (004), skill commands (008) and incompatible Classic queue-return replacement remain gaps. Combined terminal chooser adaptation is design-only with zero idle rows.

Earlier compose slice (2026-09-23): compose-005 upload/sending separation passes with real multipart bytes/progress, concurrent/session/draft/error recovery and untouched supplied components. **516 browser**, **78 functional**, **34 helpers**, Go/vet/build/hook pass. Coverage is **46/236 Classic**, **2/42 shared**, **190/40 unmapped**. [ADR-0053](../adr/0053-compose-upload-and-send-progress.md). Alt-I indexing has separate three-size fullscreen/regular evidence in [ADR-0052](../adr/0052-compact-terminal-index-actions.md); terminal media progress is design-only and automatic freshness remains disabled.

Earlier hidden/subtree slice: workspace-004 persisted toggle/root and expanded reloads, bounded native path/depth/hidden queries and host event bridge without component edits. **486/486 browser**, **74/74 functional**, **31 helpers**, Go/vet/build/hook/web race ×3; **45/236 Classic**, **2/42 shared**, **191/40 unmapped**. [ADR-0041](../adr/0041-workspace-hidden-subtrees.md). Reindex/CRUD and a temporary terminal file chooser remain separate work.

Earlier workspace preview slice: workspace-008 Markdown/escaped text/raster/binary metadata, bounded UTF-8 reads and rooted raw retrieval, using registered supplied read-only panes. **480/480 browser**, **73/73 functional**, **31 helpers**, Go/vet/build/hook/web race ×3; **44/236 Classic**, **2/42 shared**, **192/40 unmapped**. [ADR-0040](../adr/0040-read-only-workspace-previews.md). Editor/CRUD and a temporary compact terminal viewer remain separate work.

Earlier lightbox slice: stored-image projection in timeline/search, authenticated media metadata/raw retrieval, Escape/non-Escape/pointer/trusted-touch acceptance (timeline-013–016) and session/search/reload/404 draft guards. **474/474 browser**, **72/72 functional**, **31 helpers**, Go/vet/build/hook/web race ×3; **43/236 Classic**, **2/42 shared**, **193/40 unmapped**. [ADR-0039](../adr/0039-native-media-lightbox.md). Supplied components unchanged. Terminal temporary attachment selector/open actions remain design work without idle UI growth.

Earlier upload slice: native original-026 success IDs and parser failure, no premature submission, partial-batch origin recovery and explicit retry. **444/444 browser**, **70/70 functional**, **29 helpers**, Go/vet/hook; **39/236 Classic**, **2/42 shared**, **197/40 unmapped**. [ADR-0038](../adr/0038-native-upload-failure-recovery.md). No application changes. Compose-005 separate progress and terminal pending-media implementation/three-size acceptance remain open.

Earlier Markdown slice: native stored tables/code/SVG source, host-only table display/overflow adaptation, trusted code-copy payload evidence and wheel-reachable wide columns. timeline-023/024 and original-029 map. **432/432 browser**, **70/70 functional**, **29 helpers**, Go/vet/hook; **38/236 Classic**, **2/42 shared**, **198/40 unmapped**. [ADR-0037](../adr/0037-native-markdown-rendering.md). Terminal adaptation stays in existing textual projection/copy/scrollback with no idle UI; independent source-whitespace acceptance remains open.

Earlier admission-safety slice: terminal claims no longer accept fresh prompts as steering. Atomic running-claim admission plus distinct queued fallback preserves cleanup ownership and FIFO; strict queue Steer unchanged. Native held-hook, two-store race, rollback/capacity/maintenance and six-project browser evidence pass. **408/408 browser**, all TUI suites, **70/70 functional**, Go/vet/race ×3 and **29 helpers**. [ADR-0036](../adr/0036-completed-claim-admission.md); mapping stays **35/236 Classic**, **2/42 shared**, **201/40 unmapped**. No extra UI footprint.

Earlier folder-reference slice: compose-008 maps exact multiline/file/folder/message references and references-only native submission. Explicit host header action preserves folder navigation, durable session drafts and failure recovery. **402/402 browser**, **70/70 functional**, **29 helpers**, Go/vet/hook checks. Coverage **35/236 Classic**, **2/42 shared**, **201/40 unmapped**. [ADR-0035](../adr/0035-explicit-folder-references.md) records native-subscription fixture fixes and the still-open display-idle/admission-claim boundary. Terminal adaptation stays editor-based, with no added idle chrome or new acceptance credit.

Earlier selection slice: padded fullscreen drag/highlight/release or Ctrl-C/X copy, held-edge scrolling, tool-click preservation and existing clipboard policy. No idle-row growth; selection invalidates on output/layout/search/session changes and native copy completions are fenced/serialized. Three-size SGR/OSC52 tests, all TUI suites, Go/vet/race ×3, 29 helpers and 70/70 functional pass. [ADR-0034](../adr/0034-fullscreen-transcript-selection.md). The earlier failing saved tool-click WIP is now resolved. Link precedence, mutation-stable anchors, light theme and remaining web cases stay open; browser coverage unchanged.

Earlier spacing correction: Pi user top/bottom padding, assistant leading blank separator, tool external separator plus colored padding, all at message boundaries with Markdown continuation grouping. No editor/footer row growth. Three-size native spacing/outcome/search/reading/regular/session/model/compaction checks, Go/vet/race ×3, smoke/Gherkin, 29 helpers and 70/70 functional pass. [ADR-0033](../adr/0033-pi-transcript-spacing.md). Selection WIP is saved separately and still fails its expanded tool-click case; no selection completion credit. Browser coverage unchanged.

Earlier fullscreen-search slice: Ctrl-Shift-F temporary query editor, matching-row highlight/next/previous, Escape restores draft/cursor/reader state, Ctrl-Shift-Up/Down user-prompt jumps. Three-size Unicode/collapsed-expanded-tool/live-output/resize/reopen/zero-idle-row checks, all TUI suites, Go/vet/race ×3, 29 helpers and 70/70 functional pass. [ADR-0032](../adr/0032-fullscreen-transcript-search.md). Search is per rendered row, not cross-wrap or individual-occurrence matching. Fullscreen pointer selection/copy, light theme, reflow/eviction and broader parity gaps remain; browser counts unchanged.

Earlier regular-mode slice: opt-in `-tui-mode regular`, main-screen terminal-owned scrollback, no mouse capture, five idle dock rows and bounded active preview. Three-size native ordered-history/copy-selection/multiline/draft/cursor/resize/selector/session/exit/reopen tests pass; all existing TUI suites, Go/vet/race ×3, 29 helpers and 70/70 functional tests pass. [ADR-0031](../adr/0031-regular-terminal-scrollback.md). Fullscreen search/prompt jumps/selection, light theme and deeper reflow/retention limits remain open. Browser mappings unchanged at 34/236 and 2/42.

Earlier Pi-output slice: exact dark user/tool outcome bands, flat blocks, rendered-height scroll bounds, focused-editor Home/End, Ctrl-O tool expansion and dock-wheel fallback. Three-size PTY RGB/navigation/idle-row checks, all existing TUI suites, Go/vet/race ×3, 29 helpers and 70/70 functional tests pass. [ADR-0030](../adr/0030-terminal-outcome-bands.md) records the remaining regular-mode native scrollback, fullscreen search/prompt-jump/selection, light-theme and reflow/eviction work. Browser mappings remain 34/236 Classic, 2/42 shared; folder-reference WIP is saved separately and unmapped.

Earlier terminal reading slice: ordinary editing/submission preserves reader follow mode, with native delayed-provider completion, newer draft/cursor, same-row anchors, resize round trips and zero added idle rows at **60×18, 100×22 and 140×36**. Go/vet, terminal race ×3, reading/session/model/compaction/smoke/Gherkin suites, 29 helpers and 70/70 functional browser tests pass. See [ADR-0029](../adr/0029-terminal-reading-position.md). Routed acceptance/recovery, durable transcript IDs and reflow/eviction anchors still need work; browser mappings stay **34/236**, shared **2/42**, **202/40 unmapped**.

Earlier accepted-message slice: **384/384 browser**, **70/70 functional**, **29 helpers**, Go/vet/hook checks. Native acknowledgement now refreshes the selected search/timeline; uploads, newer typing/cursor, reading anchors and origin isolation are verified. Classic **34/236**, shared **2/42**, **202/40 unmapped**. Compose-007/009/010/011 map; compose-008 lacks a verified folder-reference path. [ADR-0028](../adr/0028-accepted-message-refresh.md) specifies a zero-idle-row terminal adaptation; no terminal credit from browser tests.

Earlier paging slice: **348/348 browser**, **70/70 functional**, **29 helpers**, Go/vet/hook and native races. Stable bounded message cursors, host-owned reverse-scroll paging/anchors, gap-free reconnect catch-up and stale-page/search isolation are verified; see [ADR-0027](../adr/0027-bounded-timeline-pages.md). Mapping unchanged: Classic **30/236**, shared **2/42** (206/40 unmapped). Out-of-window mutation reconciliation, loaded-window limits and terminal persistent paging remain open.

Earlier initial-refresh slice: **342/342 browser**, **70/70 functional**, **28 helpers**, Go/vet/hook checks. Classic **30/236**, shared **2/42** (206/40 unmapped). `reconnect-005` verifies readiness/activation ownership, delayed subscription, retry after failure and selection/reconnect epochs. All reconnect IDs mapped; broader crash/hashtag/paging and terminal search/queue/media still open. See [ADR-0026](../adr/0026-initial-refresh-ownership.md). No terminal UI added.

Earlier search slice: **330/330 browser**, **70/70 functional**, **27 helpers**, Go/vet/hook/race checks. Classic **29/236**, shared **2/42** (207/40 unmapped). Native current/family/all-chat search and search-view reconnect protection map `reconnect-003`. See [ADR-0025](../adr/0025-native-search-view.md). Hashtag navigation, result paging, terminal search and broader crash work remain open.

Earlier reconnect slice: **312/312 browser executions**, **70/70 functional**, **26 helpers**, Go/vet/hook checks. Classic **28/236**, shared **2/42** (208/40 unmapped). Reconnect `002/004` verify authoritative native refresh, delayed-response rejection and manual-only version drift after real socket loss/server restart; see [ADR-0024](../adr/0024-reconnect-refresh-and-version-drift.md). Search reconnect, initial refresh deduplication and active-turn crash work remain open. No terminal chrome is added.

Earlier terminal slice: `/compact`, `/compact info`, Alt-C and focused Escape now use native maintenance/status/cancellation. New live compaction PTY tests pass at **60×18, 100×22 and 140×36** with preserved draft/cursor, resize/reopen and **zero added idle rows**. Existing TUI session/model, smoke/Gherkin, Go/vet/race and 70/70 functional pass. See [ADR-0023](../adr/0023-terminal-compaction.md). Browser mapping is unchanged; broader reconnect/crash and terminal queue/media work remain open.

Earlier manual-Compact evidence: **294/294 browser executions**, **70/70 functional**, **24 helpers**, Go/vet/hook/race checks. Classic **26/236**, shared **2/42** (210/40 unmapped). `context-003` verifies callback availability, no-provider maintenance admission, cancellation and draft/media preservation. See [ADR-0022](../adr/0022-manual-compaction.md). Terminal `/compact` wiring and broader reconnect/crash acceptance remain open.

Earlier durable-checkpoint evidence: **282/282 browser executions**, **70/70 functional**, **24 helpers**, Go/vet/hook and race checks; three-size TUI regression passes. Eligible ID/fingerprint checkpoints now affect later requests without deleting history; exact native projection guards preserve hook/tool/media ambiguity. See [ADR-0021](../adr/0021-durable-context-checkpoints.md). Mapping totals remain Classic **25/236**, shared **2/42** (211/40 unmapped). Manual Compact and broader crash/reconnect acceptance are still open.

Earlier automatic-compaction evidence: **276/276 browser executions**, **70/70 functional**, **24/24 helpers**, Go/vet/hook and targeted race checks. Classic **25/236** passing (211 unmapped), shared **2/42** (40 unmapped). Compaction `001–005` and context `004` now have native lifecycle/Stop/suppression/usage evidence with reload and late-response ownership. See [ADR-0020](../adr/0020-automatic-compaction-web-status.md). Manual callback/history-boundary and broader reconnect work remain open; TUI adaptation is design-only.

Earlier meter evidence: **228/228 browser executions** (192 existing + 12 Steer + 12 context-fit + 12 meter), **70/70 functional**, Go tests/vet, hook checks and **23/23 helpers**. Classic **19/236** passing, 217 unmapped; shared **2/42** passing, 40 unmapped. `context-001/005` now verify formatting, tooltip data, arc clamping and exact colour boundaries. `compaction-006/007` have native local-provider browser evidence; full compaction is still open. Terminal measured/overflow footer unit tests and the live three-size harness pass without new UI/idle rows. The legacy dev schema migration is repaired and port 8090 is running (`1c1111b`).

Earlier evidence: fifteen Classic IDs (original `001`, `002`, `013`, `014`, `015`, `016`, `018`, `021`; compose `001`, `002`, `003`, `006`; compaction `008`; reconnect `001`; context `002`), 192/192 browser executions: fifteen Classic mappings, one shared mapping (`shared-28`) and sixteen Gi-only regressions; 70/70 functional tests. Shared queue return now persists recovery before DELETE and preserves concurrent drafts; incompatible Classic replacement cases remain unmapped. See [ADR-0017](../adr/0017-durable-queue-return.md). Latest provider-request measurement is separate from cumulative billing; browser context-fit scenarios 006/007 are now mapped; other compaction/context scenarios remain open. See [ADR-0016](../adr/0016-measured-request-context.md). Authoritative session-local model commands and stale-response isolation are verified in [ADR-0014](../adr/0014-session-model-selection.md). Queue SSE reconciliation and disconnect cleanup are verified in [ADR-0013](../adr/0013-queue-sse-reconciliation.md). Durable queue reorder/cancel and explicit queue admission are verified in [ADR-0012](../adr/0012-queued-followup-order.md). Browser-local IndexedDB drafts and captured-send recovery are verified; see [ADR-0011](../adr/0011-browser-draft-recovery.md). Native metadata mutations now persist and report failures; archive is reversible, not permanent deletion or agent shutdown. The remaining 206 Classic IDs and 40 shared cases are open. Run-bound queue Steer (`shared-30`) adds 12/12 executions using an isolated local provider: **204/204** combined, **2/42** shared passing; 70/70 functional, Go/vet/race and 23 helpers pass. See [ADR-0018](../adr/0018-run-bound-queue-steer.md). The terminal session slices now have separate unit/race and live tmux evidence: draft/history isolation, generation-owned event delivery, a temporary six-result picker and shared validated session-local model selection at 60×18, 100×22 and 140×36. Model selection preserves settings bytes and survives process restart; see [ADR-0015](../adr/0015-terminal-session-model-selection.md). Broader terminal proposals below remain open; see [ADR-0010](../adr/0010-terminal-session-selection.md).

Native automatic-compaction lifecycle safety is now implemented as a prerequisite: transactional occurrence-keyed start/finish/summary, cancel-safe phase restoration, and no provider call after failed persistence. Go/vet/race and 70/70 functional checks pass; see [ADR-0019](../adr/0019-compaction-lifecycle-safety.md). Manual compaction, browser status/stop/suppression/reconnect acceptance and durable history-boundary semantics are still open. No new mapping credit.

## Delivery order and dependencies

| Workstream | Frozen sources (under `tests/ux/features/classic/`) | Dependency / implementation work | Compact terminal equivalent |
|---|---|---|---|
| Session/agent lifecycle | `canonical/canonical-ux.feature`, `sessions/session-switching.feature` | Complete capability-gated rename/pin/archive/restore/delete, directed selection and failure paths. Persist drafts and scope model/queue/late errors/reconnect data to the origin. Native session identity, ancestry and mutation APIs come before UI success states. | Existing session selector with bounded search/results; per-session editor state; no sidebar or top header. |
| Composer, model and queue | `compose/compose-stability.feature`, `compose/compaction-model-switch.feature`, `compose/context-meter-tooltip.feature`, `compose/instant-visibility.feature`, remaining canonical cases | Real model mutations, queue actions, draft recovery and authoritative usage/compaction state. Preserve interrupted text, attachment ownership and IME/input semantics. | On-demand model/queue selectors, existing footer segments, transient notices. Unknown measurements remain unavailable. |
| Streaming/recovery and thoughts | `compose/sse-reconnection.feature`, `compose/thoughts-panel.feature`, `canonical/core-interactions.feature` | Cursor/replay ownership, delayed response guards, retry consistency and stream lifecycle. Folded tools/thoughts must use real event state. | Origin-owned buffered events, folded transcript blocks and explicit expansion; no permanent activity panel. |
| Workspace/files/editor | `canonical/workspace-flows.feature`, `editor/editor-stability.feature` | Real VFS/file/editor operations, save/conflict/failure paths, attachment references and focus. Keep preview/read-only capabilities explicit. | Path completion, bounded file selector and external editor round-trip; no file sidebar. |
| Timeline/media | `timeline/rendering.feature`, `timeline/message-deletion.feature`, `timeline/annotation-highlights.feature`, `timeline/lightbox-dismissal.feature` | Native persisted messages/media, stable streaming render, deletion/annotation semantics and viewer dismissal. Preserve the 10 MiB media limit. | Folded transcript actions and explicit media/file opening; avoid image-like terminal chrome or unbounded metadata. |
| Settings/capabilities | `canonical/core-settings.feature`, `settings/settings-dialog.feature`, `settings/settings-layering.feature`, `compose/theme-tint.feature` | Replace empty/no-op adapters with native settings persistence and truthful capabilities; test save errors, layering and changed runtime effects. | Temporary settings lists/submenus; theme-derived colours; cancel returns editor focus. |
| Authentication and operations | `canonical/core-auth.feature`, `panes/terminal.feature` | Authentication/session lifecycle, actual operational endpoints, process I/O and cancellation. No simulated terminal pane or placeholder OAuth success. | Existing shell/command paths, bounded interactive prompts and explicit login flow; separate terminal evidence. |
| Responsive/PWA/navigation | `compose/hamburger-layout-scale.feature`, `mobile/pwa-manifest.feature`, `mobile/swipe-independence.feature`, remaining canonical navigation cases | Actual responsive controls, manifest/offline/install semantics and independent touch areas. Browser-specific tests remain necessary. | Resize-safe selectors at 60×18, 100×22 and 140×36; document browser-only properties as such. |
| Shared contract | `../shared-canonical-ux.feature` | Give every expanded case a stable mapping and its own report. Includes Quick actions, Plan and session/queue/model/workspace interactions beyond the first Classic mappings. | Map user intent to existing commands, temporary selectors or requested transcript output; no permanent Plan/Quick actions panel. |

## Evidence workflow

1. Select a bounded feature family and inspect every frozen assertion before implementation.
2. Identify the native API/state contract and missing capabilities. Refresh pinned Piclaw components when authorised; keep Gi behaviour in adapters where possible and record component deviations.
3. Add real-state fixtures, native pointer/keyboard actions, failure tests and delayed-response tests where ownership matters. No forced clicks, retries, fabricated timeline data or alternative submission fallbacks.
4. Run the six-project parity matrix plus existing functional tests. Record scenario IDs separately from browser execution counts; preserve failing evidence until diagnosed.
5. Derive terminal acceptance cases from verified user flows. Test draft retention, event isolation, cancellation/focus, long Unicode text and resizing against the existing layout contract.
6. Update the catalogue, implementation checklist and this scope report after each tested commit. Do not mark the full goal complete at a feature-family milestone.

Commands: `make test`, `make vet`, `make bun-checks`, `bun test tests/ux/support/`, `make test-ux`, `make test-ux-parity`. Corpus reports live in `test-results/ux-parity/`; pinned sources and checksums live under `tests/ux/upstream/` and `web/upstream/`.

## Immediate next slice

The session mutation slice (`015`) is verified; see [ADR-0009](../adr/0009-session-picker-mutations.md). The terminal draft/event isolation and bounded selector slice is also verified. Browser-local durable drafts and origin-owned send recovery are now verified. Durable queue reorder/cancel is also verified. Queue SSE reconciliation and disconnect cleanup are verified. Session-local model mutations and command isolation are also verified. Latest provider-request measurement and unknown-context browser behaviour are verified. Shared queue return is verified separately from incompatible Classic replacement semantics. Shared run-bound Steer is also verified; Classic idle-send remains unmapped. Next: full browser compaction and remaining context-meter acceptance, remaining reconnect ownership, upload progress, pending terminal media refs, mutation submenus and durable terminal draft/queue acceptance. Remaining lifecycle cases still need their own evidence; see [web-session-parity.md](web-session-parity.md).

## Mobile ordering evidence — 2026-09-23

`@ux-mobile-004` uses real isolated native sessions: two simultaneously gated active turns, an active pinned session, a pinned idle session, an ordinary session and an archived child. The test calculates expected active-first/JID order from the persisted catalogue, then swipes forwards and backwards from five positions including the final catalogue entry. It checks draft restoration and repeats navigation after unpinning. Neither selected-session state nor catalogue responses are fabricated. Browser gestures dispatch touch events into the actual timeline listener under the same iOS capability override as existing swipe acceptance; this is browser-automation evidence, not a physical-device gesture claim.

The native list contains one record per session ID; pinned/active are overlapping metadata on those records. `tests/ux/support/chat-swipe.test.ts` adds explicit duplicate candidate rows to the actual supplied resolver and verifies deduplication, archived omission, active-first/JID ties, current-independent order and wrap in both directions. The browser test verifies the delivered native carousel and no extra stop at overlapping membership. Helper evidence alone does not earn the mapping.

Initial mixed browser execution was **26/30** because the new reverse gesture ran immediately after localStorage selection changed, before asynchronous catalogue activation. The test now waits for the visible picker to contain the native catalogue and mark the correct current row before each subsequent gesture. No implementation or frozen assertions changed. This acceptance did not cover rapid gestures during activation. The later derived regression below identifies and fixes passive listener replacement timing without adding frozen mapping credit. A wrong placeholder selector in the first focused probe was also corrected to the supplied searchbox's accessible name.

Final dedicated **6/6**, mixed swipe **30/30**, complete session regression **96/96**, functional **83/83**, support **47/47**, Go/vet/build/hook passed (recorded target exit 0). The combined parity report is **75/236 Classic**, **3/42 shared**, **161/39 unmapped**. This slice changes tests, mapping and evidence only; the deployed runtime needs no restart. Browser touch has no TUI counterpart: Alt-S retains its bounded searchable selector, arrows/Enter/Escape and zero additional idle rows, with its earlier terminal acceptance kept separate.

Workspace-005 reassessment: the header menu remains deliberately unmapped. Besides its hidden CSS entry point, `createWorkspaceFile` currently calls an unsupported POST and swallows failure, while `uploadWorkspaceFile` returns null. Restoring the header without native capability gating would expose unusable create/upload actions. Leave supplied components untouched and derive mutation contracts before enabling those controls; the replacement timeline menu does not satisfy the frozen header requirement.

## Rapid swipe listener ownership — 2026-09-23

`handleSwitchChat` already invalidates selection generations synchronously and retains the full agent catalogue for navigation. The native swipe attachment captured its old current chat and callback until a passive effect ran after the new selection rendered. A reverse contact on the next frame could therefore remain in B. Switching only that host attachment to `useLayoutEffect` replaces the closure in the commit phase. Supplied `ui/chat-swipe-navigation.ts`, sorting, exclusion rules, draft persistence, SSE gates and selection guards are unchanged.

Derived `tests/features/sessions/gi-swipe.feature` defines next-frame fresh-contact ownership and late target-timeline isolation. Its browser regression failed before the fix and passes all six projects afterwards. The final test drives four A→B→A cycles with one animation frame between contacts, then holds a fetched native B timeline, edits B's draft, reverses to A and releases stale responses. A's timeline marker, draft and uploaded attachment survive; no B marker or additional turn appears. A separate functional test checks immediate reverse and draft preservation. These are synthetic native-listener gestures, not physical-device testing. Two reversals in the same JavaScript task are outside the contract.

Validation: dedicated **6/6**, complete session suite **102/102** (including selected-text swipe guards and Safari/iOS wheel exclusions), functional **84/84**, support **48/48**, Go/vet/build/hook pass on the final runs. Review confirmed the commit-phase fix; its wheel/selection concern is covered by the full session suite. The first Go run failed `TestProcessSystemDirectWhileActiveSteersSameSession` with depth0; ten immediate repeats failed with depth2 and database-closed logs. That fixture inspects a queue while its live runner may consume it and closes a shared in-memory DB without joining cleanup. The full Go rerun passed, but fixture stabilisation is still open. No production turn code was changed to silence it.

Refinement: the operator expects a fresh reverse contact to use the displayed session, without waiting for unrelated network reads. The minimal scope is host listener timing with no new UI, persistence, permissions, dependencies or mutation endpoints. Existing native response fences remain authoritative. No new frozen mapping, terminal gesture, selector row, footer row or notification is introduced. Alt-S remains the independently verified bounded terminal adaptation. Closure requires red/green browser proof, the normal release gates, push, restart and read-only live health verification.

Deployment: `bb92551` pushed; `make restart BIN_DIR=/tmp/gi-scheduler-bin WORKSPACE=/workspace` succeeded on **8090**, PID **3925255**. Live Chromium phone read-only smoke performed four next-frame A/B transitions using existing sessions and restored its local draft. No settings/session writes or page errors; **61** sessions available. SQLite integrity `ok`, foreign-key check empty. Evidence `/workspace/tmp/gi-rapid-live.log`. Isolated test mutations were not repeated against operator data.

## Steering fixture stabilisation — 2026-09-23

The flaky `TestProcessSystemDirectWhileActiveSteersSameSession` now uses an isolated temporary SQLite database and an existing setup hook to hold the bootstrap shell runner before it can consume steering. The same-turn, running/not-queued, exactly-one-row, system-role and ingress-source assertions are unchanged. Deferred cleanup releases the gate, takes the runner lock and waits for both an empty current-run pointer and no active claim before closing the engine/store; the lock joins final queue/session normalisation after claim release. No production scheduling code changed.

Verification: **100** repeated runs, **50** race-detector repeated runs, full Go/vet/build/hook, **48/48** support and **84/84** functional tests pass. Focused independent review found no cleanup or weakened-coverage issue. Logs: `/workspace/tmp/gi-steering-fixture-{100,race,gates}.log`. This test-only slice needs no runtime restart and earns no additional browser or terminal credit; coverage remains **75/236 Classic**, **3/42 shared**.

## Native status-panel swipe passthrough — 2026-09-23

Gi rendered `AgentStatus` next to the timeline but attached the supplied swipe helper only to the timeline. A genuine streamed Markdown link inside `.agent-thinking` / `.agent-status-panel` passed the helper's target rule but never reached its listener. The frozen mobile003 probe failed before the host change. The listener now uses the existing conversation container, with a boundary listener registered first for `pointerdown`, `touchstart` and `wheel`. Only timeline descendants and direct-child status panels can start navigation. Direction, platform, stylus and interactive-target logic remain in the unchanged supplied helper. Layout-effect replacement preserves the earlier rapid-reversal fix.

The boundary listener runs in the bubble phase, after target handlers, and does not prevent defaults. Excluded input touch/wheel callbacks still run. `stopImmediatePropagation` prevents those events reaching the swipe helper or further ancestors; this explicit propagation boundary is not general document-event transparency. The status-panel direct-child check matches current markup; future wrapper changes must update the boundary and browser tests. Review identified an initially ungated pen pointerdown; it is now gated and covered by a composer pen event followed by a valid finger contact on the status link.

`make test-ux-status-swipes` starts the existing isolated local provider fixture. It emits real thought/draft Markdown through provider parsing, inference, native SSE and the supplied renderer. The composer submits the gated turn so its accepted-send catalogue refresh runs normally. All six Chromium/WebKit viewport projects prove draft and thought links, vertical cancellation, selected-text suppression, composer/Settings exclusion, target callback/default retention, adjacency and draft restoration. No DOM panels, API snapshots or SSE messages are fabricated. Draft/thought share the thinking/status ancestry; no independent interactive intent control exists in this Gi adapter, and none is claimed. The functional test independently swipes the native completed-status surface. Synthetic browser touch events are not physical-device evidence.

Validation: red→green frozen probe, panel **12/12**, session **102/102** including rapid gestures/selection/wheel, Steer **18/18**, functional **85/85**, support **48/48**, Go/vet/build/hook all passed with exit 0. Final report **76/236 Classic**, **3/42 shared**, **160/39 unmapped**. Steer evidence was regenerated after an exploratory fixture run replaced its result file. Captures: `test-results/status-swipe-captures/chromium-{phone,tablet,desktop}.png`.

An initial long-text fixture could not expose expand buttons: Gi sends zero preview line counts and uses a no-op `onPanelToggle`. This is a separate native adapter gap. Current passthrough evidence uses real rendered links and does not award truncation/expansion credit. Terminal adaptation remains the existing bounded Alt-S selector; status text consumes no additional terminal controls or idle rows, and browser passthrough adds no TUI acceptance.

Deployment: `28996cf` pushed and Makefile restart succeeded on **8090**, PID **4005246**. Live Chromium phone read-only smoke swiped the native status surface, then performed four rapid timeline transitions and restored its local draft. No session/settings writes or page errors; **61** sessions remain available. SQLite integrity `ok`, foreign-key check empty. Log `/workspace/tmp/gi-panels-live.log`; live provider turns were not started. Streaming-link evidence and captures come from isolated instances.

## Native streamed preview disclosure — 2026-09-23

Gi formerly supplied `totalLines: 0` for every draft/thought preview, suppressing disclosure controls even for multiline streams. It now passes the accumulated text to the unchanged component's normalizer. `AgentStatus` already owns its expansion set and Escape handling; the no-op toggle callback was unnecessary, not the source of the bug. The host keys the component by session/native turn, restores current turn identity from accepted activity reads, adopts a tagged delta only when no turn is known, and rejects deltas tagged for another current turn. Completion/disconnect/session changes clear preview ownership through existing lifecycle paths. Native untagged diagnostic content retains selected-session semantics; this is not replay/recovery of complete previews across reconnection.

Expanded content previously had no scroll container inside Gi's fixed-height shell. A Gi-owned stylesheet bounds the direct-child status region to40% of the conversation height, enables vertical scrolling, and prevents flex-shrinking of its panels. No supplied component, utility or stylesheet was edited. The full supplied Markdown stays rendered and copyable; collapse uses the existing CSS limit, expansion removes it.

The deterministic local provider streams twelve explicit lines each for thought/draft, then four more while collapsed, then another four while expanded. Real provider parsing and SSE drive the native panel. Six-project acceptance checks independent draft/thought disclosure, continued updates in either state, click/close/show-less/keyboard scope, unmodified Escape versus Shift+Escape and composer Escape, preserved text, scrolling, viewport resize and composer visibility. Gi preview001 verifies session switch, new-turn and reload resets using actual native response/subscription readiness. No DOM state, stream message or native payload is fabricated.

Review identified two lifecycle gaps, now corrected: activity GET restoration of turn identity and tagged-delta ownership. It also identified a remaining supplied limitation: long single paragraphs use hard-newline metadata even when CSS wraps/clips them, so a disclosure button can remain absent. Consequently **Thoughts001 is unmapped**. The explicit-newline disclosure case has only Gi-derived evidence. **Thoughts002–005** satisfy their frozen update, toggle, Escape and text-preservation criteria across six projects. Further single-line disclosure work needs a narrow renderer adaptation and native browser evidence, not inflated line metadata.

Final gates: preview **36/36**, status-swipe **12/12**, reconnect **60/60**, functional **85/85**, helpers **49/49**, Go/vet/build/hook pass with exit0. An initial parallel Go/embed run raced hashed-asset pruning; gates were rerun serially. A lifecycle probe released a chunk before the returned session subscribed; it now waits for the native activity read and connected indicator before release. Report **80/236 Classic**, **3/42 shared**, **156/39 unmapped**. Captures `test-results/thought-captures/chromium-{phone,tablet,desktop}.png`.

Terminal disposition: retain the existing streamed transcript and independently tested Ctrl-O disclosure; no separate thought/draft sidebar, scroll region, footer row or idle status is introduced. Terminal reset/reflow acceptance remains separate from the browser cases. Scope is native text metadata, turn-scoped host mounting, scroll bounds and tests; no persistence/schema/auth/provider credential change, automatic retry, export or new network service.

Deployment: pushed `965e819`; Makefile restart on **8090**, PID **4083186**, succeeded. Read-only live smoke confirmed native status swipe, four rapid transitions, restored local draft, no page errors or session/settings writes. `/css/gi-status.css` returned200; **61** sessions available, SQLite integrity `ok`, foreign-key check empty. Log `/workspace/tmp/gi-preview-live.log`. Expanded-stream evidence comes from isolated fixture captures; no operator prompt or credential/settings mutation was made.

## Wrapped-preview disclosure — 2026-09-23

Thoughts001 is now mapped. `scripts/patch-status-preview.mjs` applies exact guarded substitutions only while Bun loads the supplied status renderer: a top-level Gi measurement hook, draft/thought body refs, and a generic Show more fallback. The build fails if an anchor is absent or duplicated; support tests also reject double application. Supplied component files, Markdown rendering, line metadata and component-owned expansion/Escape behaviour are unchanged. The existing omitted-line label remains preferred for explicit multiline content.

`gi-preview-overflow.ts` measures collapsed `scrollHeight > clientHeight + 1` against the actual CSS box, so padding/wrapping/viewport changes require no fabricated line estimate. Measurements coalesce into animation frames. Expanded panels retain their last overflow result until collapse so a resize never strands them without a collapse action. Each panel body and its immediate Markdown children are observed; source updates explicitly remeasure even when the clipped body height stays constant. Window resize and font completion also remeasure. Without ResizeObserver, class/style changes on the bounded ancestor chain plus transition completion cover host layout changes. Cleanup disconnects observers, cancels queued frames and removes listeners; no idle polling or extra network calls.

The local native provider emits a paragraph without internal newlines, first short enough for desktop width, then longer while the collapsed box is unchanged. Six-project browser tests verify disclosure appears/disappears with actual overflow, content bytes remain intact, expanded collapse/Escape remains available, viewport and container-only changes work, and the noRO fallback follows the same contract. The existing multiline/session/turn/reload matrix remains green. A first text assertion needed `trimEnd()` for the Markdown renderer's trailing paragraph newline; a container-only probe initially hit the desktop minimum width and was corrected to resize the actual box. Final **48/48** preview, **12/12** status-swipe, **85/85** functional, **51/51** support, full Go/vet/build/hook passed. Independent review caught the missing container-only fallback; the final test covers it. Full report **81/236 Classic**, **3/42 shared**, **155/39 unmapped**.

Refinement: the operator must be able to disclose visibly clipped streamed text. Scope is measured browser presentation only; no persisted preferences, provider credentials, endpoint/schema changes or altered source text. Failure of guarded upstream anchors stops the build. The Gi adapter lives outside supplied directories. Terminal uses existing transcript disclosure and has no browser layout analogue, new panel or idle rows. Captures: `test-results/thought-captures/wrapped-chromium-{phone,tablet,desktop}.png`.

Deployment: `bdf731e` pushed; Makefile restart on **8090**, PID **4118891**, succeeded. Read-only Chromium phone smoke confirmed status swipe, four rapid transitions, restored local draft, no page errors or settings/session writes. **61** sessions available; SQLite integrity `ok`, foreign-key check empty. Log `/workspace/tmp/gi-overflow-live.log`. Actual wrapped-stream overflow/captures were verified only against isolated fixture instances; no operator turn was created.

## Shared29 durable queue identity — 2026-09-23

The shared scenario has its own browser test and mapping; Classic018 evidence is not reused as automatic credit. Two native sessions run independent gated turns. A visible adjacent move in A sends the exact expected/order ID arrays, changes only the persisted FIFO order, preserves every queue-record field, and survives reload. B's queued records are equal before/after; both session drafts remain unchanged.

The test holds the user's DELETE before backend admission, releases A's original run and lets the selected queued ID become active. Continuing the unchanged DELETE produces native409. The UI reconciles the consumed ID away, retains the remaining queue, surfaces the failure and does not cancel the active target. Releasing that target proves the reordered FIFO execution in persisted user history. This satisfies the frozen “remains or reconciles to authoritative consumed state” alternative with a real rejection. Existing Classic tests separately cover transport failure retaining a queued row; they are not described as native rejection.

Focused independent review confirmed the consumed-ID branch and suggested stronger content checks; the final case compares complete records modulo order. Dedicated six-project **6/6**, complete queue **42/42**, functional **85/85**, support **51/51**, Go/vet/build/hook passed. Logs `/workspace/tmp/gi-shared29-{final,queue,gates}.log`. Initial probe passed the session object instead of its ID to the picker; corrected without changing runtime behaviour. The support report requires all six projects and keeps shared evidence separate. Combined totals **81/236 Classic**, **4/42 shared**, **155/38 unmapped**.

Terminal adaptation remains the planned bounded explicit queue selector/actions: select a durable row, move by one adjacent position or confirm cancellation, then refresh native state; no idle queue stack or permanent sidebar. Drafts and origin IDs must survive rejection before it earns terminal acceptance. No terminal implementation or runtime code changed in this test-only slice; no restart is required.

## Shared23/24 picker opening — 2026-09-23

The supplied picker already focuses search in a layout effect. Dedicated shared tests now observe real popup insertion and its first animation frame; both record the search input as active, and the first frame contains exactly one visible popup. Pointer clicks and Enter on both native session trigger buttons are exercised. The popup remains absolutely positioned against `.compose-input-main`, aligned left and six pixels above it, inside the viewport before/after resizing. Tests search by persisted `gi:` identifier, keep main/research records visible as applicable, and confirm Escape restores the exact trigger without changing session, draft or either session's turn count.

These are browser DOM/frame observations supported by the actual layout-effect source, not compositor tracing or a claim that CSS Anchor Positioning/Popover APIs are used. The frozen phrase “native composer target” refers to Gi's real component anchor. No synthetic popup, application state, focus handler or geometry replacement is used. Independent review accepted the behaviour evidence with those limits. The report requires six project results for each shared ID; five results remain partial-matrix.

Dedicated **12/12**, complete session suite **114/114**, functional **85/85**, support **51/51**, Go/vet/build/hook pass. First full run was **112/114** because mobile001 read an assistant message before native active-claim cleanup and assumed that session already sorted last among idle sessions. The fixture now waits for the authoritative session idle status before its unchanged tail-order assertion. An initial probe also corrected the native trigger wrapper selector. No production component was altered. Logs `/workspace/tmp/gi-sharedpicker-{final,matrix,gates}.log`; totals **81/236 Classic**, **6/42 shared**, **155/36 unmapped**.

Terminal disposition: existing Alt-S opens its bounded searchable selector, restores editor/cursor state on Escape and adds no idle row. It has no CSS anchoring or browser first-paint analogue; its earlier independent PTY evidence remains the terminal acceptance. This browser test-only slice requires no restart and awards no additional terminal credit.

## Shared31/32 authoritative model selection — 2026-09-23

Dedicated shared cases use the isolated context-fit provider/registry: current `ux-local/gate` has32K context, target `ux-local/large` has200, and an actual provider request records100 tokens with native measurement provenance. Both registry capability records and visible context labels are checked. Native incremental typeahead highlights `large`; pointer click or Enter activates that real enabled entry. The outline requires search but not a filter field; no synthetic search input or filtered catalogue is introduced. Broader grouped-filter/typeahead scenarios remain unmapped.

The actual accepted PATCH200 response is held. Gi displays disabled “Switching…”, retains its prior current row and32K context, and only after delivery shows the selected model and100/200 (50%) context. Reload preserves selection in main only. Another session's complete model snapshot remains equal. Unsent text, uploaded media, a real workspace file reference and the exact native message-reference ID are preserved through mutation, reload and session switches. No additional turn is submitted.

Six-project context/shared matrix **30/30**, functional **85/85**, support **51/51**, Go/vet/build/hook passed. Focused review accepted typeahead-as-search for this outline and requested stronger visible capability/exact-reference checks, now explicit. Initial fixture corrections followed actual native contracts: false `reasoning` is omitted from JSON, model titles use `ctx`, the workspace toggle closes its drawer, and timestamps create message-reference pills. Pending state is “Switching…” rather than the old model label; the previous confirmed row/context is still checked. These were assertion/selector corrections, not runtime changes. Logs `/workspace/tmp/gi-sharedmodel-{final,gates}.log`; totals **81/236 Classic**, **8/42 shared**, **155/34 unmapped**.

Terminal adaptation is the already implemented bounded Alt-M/session-local model selector and compact context footer. Pointer UI, DOM reload and browser reference pills add no terminal credit; pending-media terminal references remain separate work. Test-only commit, no runtime restart or new idle chrome.

## Shared25 coherent session switching — 2026-09-23

`tests/ux/context-fit.spec.mjs` independently verifies the shared scenario with two native sessions. Each has a completed provider measurement, distinct history IDs, an active gated provider request, one durable queued turn and a saved composer draft with media. Main uses a 32K context model; research uses 200 tokens. Both retain their measured 100-token provenance.

All main timeline/model/queue/activity/compaction responses are fetched from the native server and held during selection. The native picker accepts research through Enter, identifier search, ArrowDown and Enter. Every research surface must show its own positive marker before old responses are released. Each delivery waits for request completion and two animation frames before repeating the exact view assertions. The timeline is explicitly a single initial page: no `after` cursor, `has_more=false`, one held request. The test also preserves newer research typing, reloads the draft/media and returns to both sessions. Final native message, model and queue snapshots remain equal; neither session gains an extra turn. No DOM or backend state is fabricated.

Independent review found no remaining gap against this scenario. `make test-ux-context-fit UX_LOCAL_BIN=/tmp/gi-coherent-bin` passed **36/36**; canonical session regression passed **114/114**; functional **85/85**, support **51/51**, Go/vet/build/hook passed. Logs: `/workspace/tmp/gi-coherent-{matrix,session,standard}.log`. Initial fixture corrections used the corpus `name` property, top-level compaction `reason`, and the paged `messages?limit=50` API. An initial session regression used the wrong local-provider fixture and rejected six `test/bootstrap` selections with400; the unchanged tests passed with `make test-ux-parity`. Report guards require all six projects and keep shared26 unsupported. Coverage is **81/236 Classic**, **9/42 shared**, **155/33 unmapped**.

Terminal disposition: retain the explicit six-result Alt-S selector, Enter confirmation and Escape cancellation. Selection should refresh transcript, model/context in the existing footer, and the session-owned editor under one generation; delayed events must fail that generation check. No permanent session panel, queue list or additional idle row is appropriate. Existing terminal selection/generation evidence stays separate. Integrating durable queue recovery and pending-media drafts still needs terminal implementation and independent fullscreen/regular PTY tests at60×18,100×22 and140×36, including resize and draft/cursor restoration. This browser-only slice adds no terminal credit or runtime changes.

## Shared27 exactly-once FIFO follow-ups — 2026-09-23

`tests/ux/queue.spec.mjs` sends two independent follow-ups through the native composer while the captured main turn is running. Both include multiline text, a real workspace file and folder, a native message reference and an uploaded text file. Active status, queue ownership and the Stop control are checked before each submission. The first actual202 response is held while SSE replaces the optimistic row with its durable ID. Two POSTs carry distinct client request tokens; the native queue contains exactly those two IDs in FIFO order.

The test compares complete serialized prompts, stored media references and downloaded source bytes. Each turn has one media reference and the session has exactly two uploaded media records. Native media resolution adds a validated creation timestamp. Queue records and a newer composer draft survive reload. After releasing the active turn, the two user messages appear exactly once in FIFO order, with their original turn IDs and media metadata retained. Research has its own history and unsent media draft; its messages, turns, model, queue and uploaded-media snapshots remain unchanged. Its local draft survives switching and reload.

Both full queue runs passed **48/48** across Chromium/WebKit at phone/tablet/desktop sizes; the final run includes explicit pre-submission active-turn assertions requested by review. Functional **85/85**, support **51/51**, Go/vet/build/hook passed. Logs: `/workspace/tmp/gi-fifo-{matrix,final}.log`. The report uses the separately extracted shared27 results from `shared-fifo-queue-results.json` to avoid counting queue regressions twice. Totals: **81/236 Classic**, **10/42 shared**, **155/32 unmapped**. Initial fixture corrections scoped attachment pills to the editor instead of queue rows, used the explorer caret to reopen an already-selected folder, accepted the native creation timestamp and compared media inventory independently of its newest-first order. No runtime or supplied-component changes; no restart.

Terminal adaptation stays on demand: use a short queue selector in the existing transient interaction area, show native FIFO identity, and offer explicit return/remove/adjacent-move actions only after selection. Escape restores the editor and cursor without mutation. A pending action keeps its captured session and durable queue ID; failure must retain recoverable input and support explicit retry. No always-visible queue list or additional idle row. Pending-media drafts and durable terminal queue actions still need implementation and fullscreen/regular PTY acceptance at60×18,100×22 and140×36. This browser result adds no terminal credit.

## Shared26 session mutation capabilities — 2026-09-23

`tests/ux/session.spec.mjs` independently verifies the shared capability scenario with a native root, an idle child and a running child. The native catalogue has no message-count fields. Session DELETE returns405 with `Allow: GET, PATCH` and leaves each record unchanged. Neither row deletion nor current-session deletion appears as each session becomes current; no browser DELETE request occurs. Gi supplies no deletion callback. If deletion is implemented later, it will need explicit known-count and activity guards before that callback is exposed.

Visible actions exercise actual native pin/unpin, rename, archive/restore and fork writes. A blank rename returns400; the same form retains its input, displays the error, and accepts a corrected retry without changing selection or losing composer text/media. Reload preserves archived state and the draft. New creates a distinct native child and returning restores the origin draft. Identity, ancestry and turn counts are checked.

Review requested native evidence for hidden non-delete actions. Root and busy archive requests return409 without changing records; archived rename/pin/unpin also return409. Restore-on-idle and repeated archive are native idempotent operations; their redundant controls are hidden and native state, including the original archive timestamp, remains unchanged apart from `updated_at`. The supplied popup places edit forms, alerts and global actions outside its inner ARIA menu; selectors cover the complete `.compose-session-popup`.

Final six-project sessions **120/120**, functional **85/85**, support **51/51**, Go/vet/build/hook passed. Logs: `/workspace/tmp/gi-capabilities-{final-matrix,gates}.log`. `shared-capabilities-results.json` extracts only shared26 from the full session run for nonduplicated aggregation. Totals **81/236 Classic**, **11/42 shared**, **155/31 unmapped**. Test-only, no runtime or supplied-component edits and no restart.

Terminal disposition: preserve Alt-S's six-result selector and default Enter-to-switch behaviour. A future explicit action submenu should temporarily replace the selector content, offer only native supported actions for its captured session, and keep rename input/errors recoverable. Escape backs out to selection and then restores editor/cursor/reading position. Archive requires explicit confirmation; deletion stays absent. No persistent action bar, session panel or new idle row. Implementation and fullscreen/regular PTY checks at60×18,100×22 and140×36 remain separate from this browser evidence.

## Shared3 workspace visibility and drawer isolation — 2026-09-24

`tests/ux/workspace-preview.spec.mjs` independently verifies the native workspace menu, actual tree file visibility and hide action. The composer starts with multiline text, a real workspace file reference, a native message ID and an unsent attachment. Its complete IndexedDB draft record, including attachment bytes, remains equal through show/hide, backdrop dismissal and reload. The current session and native message/turn history remain unchanged.

At phone/tablet sizes, the test first confirms that the selected coordinates hit the textarea or enabled Send button with the workspace closed. After opening, `elementFromPoint` must resolve to the backdrop at those same coordinates. Real pointer clicks dismiss the drawer without reaching any composer click handler, submitting content, opening pickers or focusing the textarea. Phone session-picker coordinates fall under the sidebar itself and are outside this backdrop test. Desktop verifies the sidebar with no displayed backdrop. Gi has no native Plan opener; its absence satisfies this negative activation criterion but adds no Plan functionality or acceptance.

Focused review found no necessary gap. Full six-project workspace suite **24/24**, functional **85/85**, helpers **51/51**, Go/vet/build/hook passed. Logs `/workspace/tmp/gi-workspace-shared-{matrix,gates}.log`; extracted shared3 evidence is `shared-workspace-results.json`. A duplicate helper variable name in the first report-guard run was corrected before all gates were rerun. Totals **81/236 Classic**, **12/42 shared**, **155/30 unmapped**. No runtime or supplied-source change; no restart.

Terminal disposition: browser drawers/backdrops have no terminal analogue. A future file chooser should occupy only a temporary, bounded interaction area, with keyboard navigation and Escape restoring editor text, cursor and reading position. Do not add a permanent tree or a Plan row. Existing Alt-I index actions retain their separate evidence; terminal workspace selection/reference handling still needs independent acceptance.

### Shared39 upload cancellation gap

The frozen attachment scenario requires cancelling a selected upload, retaining its draft, explicitly retrying one durable media item, and surviving reload/source removal. `web/src/api.ts` handles XHR aborts, and `ComposeTransfer` in `web/src/app.ts` shows native progress, but neither exposes an upload Cancel action. Existing original026 failure/retry and compose005 progress tests cannot satisfy that step. Shared39 remains unmapped. Terminal `/attach` and `/paste-image` currently store or submit media directly; they do not provide a pending-media draft/retry flow. Keep cancellation, retry ownership and terminal acceptance as separate implementation work.

## Shared1/2 menu dismissal safety — 2026-09-24

New independent cases in `tests/ux/classic.spec.mjs` reproduced two bugs: clicking outside the menu over Send submitted the draft, and Escape left focus on a removed menu item. The supplied menu closed on `mousedown` without consuming the later click and had no focus return. `scripts/patch-timeline-menu.mjs` now replaces only those two effects, under exact single-occurrence anchors. The supplied `web/src/components/timeline-menu.ts` stays unchanged. The new layout effect runs after the portal render so first-frame refs exist.

`web/src/gi-menu-dismissal.ts` consumes outside pointer/touch event propagation while the menu is open. Mouse-down default focus is prevented; touch defaults remain enabled so real taps and scrolling/cancellation retain browser semantics. The outside click is consumed before closing and restoring the connected, enabled trigger. Escape is consumed with the same focus return, except during IME composition. Inside-menu and trigger actions remain unchanged. Listeners are removed on closure/unmount; there is no delayed click guard that could consume a later independent gesture. Cancelled pointer sequences leave the menu open for a new dismissal.

Shared cases verify real Tab traversal through every enabled menu item, pointer/keyboard opening, a dismissal click over enabled Send, zero underlying composer activation/submission, draft/media retention and trigger focus. Separate trusted-touch browser tests prove the first tap dismisses and the next gesture works. Synthetic pointercancel evidence is labelled separately; no physical-device claim. Helper tests cover unchanged render content, rejected adapter drift/double application, event cleanup, IME and disconnected triggers. The numeric display-scale control is hidden outside PWA display modes; normal-browser focus acceptance uses visible menu items.

Six-project menu/workspace/Quick Actions/Settings regression **192/192**, functional **85/85**, support **55/55**, Go/vet/build/hook passed; implementation review found no blocker. Logs `/workspace/tmp/gi-shell-{red,matrix,gates,support-final}.log`. Red tests: unintended native submit and lost focus. Final report: **81/236 Classic**, **14/42 shared**, **155/28 unmapped**. Native HTTP routes, turn scheduling, credentials and storage are unchanged.

Terminal disposition: keep the existing bounded Alt-S/Alt-M/Alt-I interactions and focused Escape cancellation. This browser portal/pointer repair needs no terminal menu bar, backdrop or additional idle row. Terminal mutation/file/queue menus still need their own implementation and PTY acceptance; no terminal credit is added here.

Deployment: pushed `8580f57`, restarted8090 using `/tmp/gi-scheduler-bin`. Read-only live checks ran Chromium and WebKit at390×844,820×1180 and1440×900. Escape returned trigger focus; mouse and trusted-touch dismissal over Send retained the draft; the next editor gesture worked. Mutating API requests were blocked and none were attempted; no page errors. `/api/sessions` returned200 with61 sessions; SQLite integrity `ok`, foreign-key check empty. Evidence `/workspace/tmp/gi-shell-live.{mjs,log}`. Worktree clean after deterministic rebuild.

## Shared5/6/11/12 native typing ownership — 2026-09-24

Independent cases in `tests/ux/quick-actions.spec.mjs` seed real history and a second native session, retain unsent media, and establish a full native message/turn/model snapshot. Each proves the palette can open from existing history before checking exclusion. Composer typing uses sequential `é文`, cursor checks, Backspace and `ß`; Settings uses the actual Models filter, appending `boot` then `strap` and checking the result without applying it. Session search types a native destination ID, checks the exact filtered row, then dismisses with Escape without switching. Shared11 exercises the session-picker alternative; model-picker search and unsupported contenteditable surfaces gain no credit.

Shared12 dispatches prevented, repeated, composing, whitespace and Control/Meta/Alt-modified key metadata on an existing native history node. After two animation frames, Quick Actions must remain closed with draft/media, selection and complete native snapshots unchanged. A real printable-key open/close after every rejected event verifies the listener remains usable. These event-metadata fixtures do not test a physical IME or add fabricated DOM controls. No mutation request is attempted during the guard checks.

Focused four-case run passed; six-project Quick Actions/menu/Settings shell regression **192/192**, functional **85/85**, support **55/55**, Go/vet/build/hook passed. Focused review found no material gap. Logs `/workspace/tmp/gi-keys-{matrix,gates}.log`; `shared-keyguards-results.json` extracts the four independent cases without duplicating regression evidence. All four report guards reject five-project partial matrices. Coverage **81/236 Classic**, **18/42 shared**, **155/24 unmapped**. Test-only, no runtime changes or restart.

Terminal disposition: ordinary typing belongs to the focused editor or selector search field. Continue explicit Alt-S/Alt-M/Alt-I activation instead of adopting printable-key global palette capture. Commands, navigation and Escape must remain owned by the currently focused terminal interaction, with draft/cursor restoration after cancellation and no extra idle rows. Existing terminal acceptance is separate; browser IME/modifier results add no PTY evidence. Quick Actions dismissal/activation and the remaining surface rows still need their own tests or implementation.

## Shared7/9/10 target keyboard delivery — 2026-09-24

Independent shared tests showed that timeline Copy/link targets and Settings inputs never received their own `keydown` events. The palette remained closed, but Gi's global capture guard and Settings' blanket propagation stop swallowed the event before the target. Sidebar delivery already worked.

`web/src/gi-quick-actions.ts` now provides a non-consuming predicate covering readiness, consumed/repeated keys and interactive targets, plus one Settings ownership predicate. The application no longer registers a global swallowing key guard. `scripts/patch-popup-keys.mjs` inserts the predicates into the supplied palette and composer popup handlers under exact checked anchors; supplied files and render templates stay unchanged. Background palette/composer/menu handlers ignore Settings keys and pointer gestures. Settings captures only Escape and trapped Tab; ordinary keys reach the target and stop bubbling at the dialog. Composing Escape reaches the input. No DOM controls or native backend responses are fabricated.

The shared tests observe trusted `q` keydown delivery at native buttons/links, the focused workspace tree and a real modal field, then verify no palette or backend mutation and preserved draft/media/state. Separate stacked-popup tests cover model/session/palette/menu behind Settings, actual modal clicks, printable and navigation keys, composing Escape, two-way Tab wrapping, inner-then-outer Escape and unchanged background highlights. A new functional regression checks the target key and modal ownership paths. The hidden contenteditable surface remains unmapped.

Final serial six-project suites passed **216/216** (Quick Actions/menu/Settings shell) and **108/108** (Gi Settings), with **86/86** functional, **58/58** support and Go/vet/build/hook checks. Logs `/workspace/tmp/gi-surfaces-{red,modal-final,batches,gates,support-final}.log`; shared-only report extraction is `shared-surfaces-results.json`. Review requested removal of every global swallowing predicate, pointer suspension, composing Escape and Tab acceptance; all are included. An early matrix timed out at305/324, a later snapshot raced model-catalogue readiness, and a subsequent WebKit reload failed before input checks with internal resource-load errors. Final serial reruns use unchanged acceptance assertions; no partial run receives credit. The focus-wrap fixture names the actual last control, Reload saved names. Coverage **81/236 Classic**, **21/42 shared**, **155/21 unmapped**.

Terminal disposition: focused editor/selector controls must receive ordinary keys before global command handling. Keep explicit Alt-S/Alt-M/Alt-I selectors, scoped Escape cancellation and zero new idle rows. Suspended terminal interactions must not consume another foreground interaction's keys. This browser repair grants no new terminal acceptance; existing PTY evidence and pending terminal media/queue/mutation work remain separate.

Deployment: pushed `4ae11da`, restarted8090 using `/tmp/gi-scheduler-bin`. Live Chromium/WebKit at390×844,820×1180 and1440×900 verified timeline-link key delivery, Settings field/navigation keys, suspended model selection, composing Escape, two-way Tab wrap, nested Escape and retained draft. Mutating API routes were blocked and no writes were attempted; no page errors. `/api/sessions` returned200 with61 sessions; SQLite integrity `ok`, foreign-key check empty. Evidence `/workspace/tmp/gi-surfaces-live.{mjs,log}`; deterministic rebuild left a clean worktree.

## Shared13/14 Quick Actions dismissal — 2026-09-24

The existing conversation host is now a focusable, labelled Conversation region. It adds no visible control or layout row and gives keyboard users a native timeline-typing target. The palette captures the opening active element, with Conversation as the fallback for pointer-only timeline typing. `web/src/gi-quick-actions-focus.ts` restores a connected, visible, enabled, non-inert opener on dismissal, cancels and fences its scheduled focus frame on close/unmount, and never focuses beneath Settings. Action activation/unmount does not restore opener focus.

The exact-anchor build adapter consolidates key ownership and dismissal changes in `scripts/patch-popup-keys.mjs`; supplied component render/source bytes stay unchanged. The existing menu dismissal helper accepts a nullable trigger and consumes trusted outside gestures through their click. Quick Actions' overlay is pointer-transparent: Send is the actual outside hit target. Escape ignores composing events and Settings retains ownership while open. Programmatic clicks pass through: an initial regression caught Alt+Enter's temporary native link being swallowed, now fixed and verified independently with the original linked-tab test.

Independent shared tests create native history, file/message references and a local media draft. The entire IndexedDB draft including attachment bytes remains equal after Escape/outside dismissal, reopen and reload. Focus returns to the opening Conversation region; no underlying composer click, prompt or session change occurs. Trusted-touch dismissal and the next independent editor tap pass; composing Escape stays with the palette. Unit fixtures cover late-frame fencing, cleanup, removed/inert openers, modal exclusion and source-drift rejection. Unit callback events model trust explicitly because real `Event.isTrusted` cannot be redefined; browser tests use actual trusted mouse/touch input.

Six-project Quick Actions/menu/Settings shell **234/234**, functional **87/87**, support **61/61**, Go/vet/build/hook passed. Final review found no concrete blocker. Logs `/workspace/tmp/gi-dismiss-{red,linkfix,matrix-final,gates-final,support-final}.log`. The first matrix timed out while exposing the Alt+Enter regression; no incomplete evidence is counted. Shared-only results: `shared-dismiss-results.json`. Totals **81/236 Classic**, **23/42 shared**, **155/19 unmapped**. Shared15 remains unmapped because the supplied palette has no close control.

Terminal disposition: retain explicit bounded selectors with Escape returning to the captured editor/cursor/reading position. Do not introduce a global printable-key palette or a permanent menu/focus row. Terminal action execution and cancellation continue to need separate PTY evidence; this browser focus repair adds none.

Deployment: pushed `5353392`, restarted8090 from `/tmp/gi-scheduler-bin`. Read-only live Chromium/WebKit at390×844,820×1180 and1440×900 passed Escape, mouse and trusted-touch dismissal over Send, Conversation focus restoration, retained draft and next editor tap. No blocked mutation attempts or page errors;61 sessions/HTTP200, integrity `ok`, foreign-key check empty. The first smoke typed before palette readiness on its second viewport; the final probe waits for the native catalogue and passive listener setup. Evidence `/workspace/tmp/gi-dismiss-live.mjs` and `gi-dismiss-live-final.log`. Worktree clean after deterministic rebuild.

## Shared15 explicit Quick Actions close control — 2026-09-24

Refined scope: provide an accessible dismissal control for the existing transient palette, preserve captured focus/drafts, and keep native action activation unchanged. No new idle UI, terminal selector, backend route or saved preference. The feature's frozen close-control scenario supplies the acceptance contract; source provenance and six-project browser evidence remain delivery gates.

The guarded adapter adds one `type=button` Close quick actions control to the search header. `bindQuickActionsFocus` gives it the same one-shot dismissal, frame cancellation and opener restoration as Escape/outside clicks. Cleanup only removes listeners; it cannot steal focus after activation. Enter/Space on any focused native palette button now belongs to that button, preventing Close or a tab-focused action from executing an unrelated highlighted result. Search-field navigation keeps its supplied behaviour.

The existing header now uses three columns; the Close target is28×28px. On phones, keyboard hint keycaps remain while decorative hint text is hidden, leaving over100px of input width without wrapping or horizontal overflow. Styles live in the separate `gi-quick-actions-controls.css`. The first attempt appended to `gi-quick-actions.css`; provenance checks caught that this file holds pinned upstream CSS, so its exact bytes were restored. Source and CSS hash gates pass unchanged.

Shared15 verifies pointer and real Tab/Enter/Space dismissal with exact stored text, attachment bytes, workspace/message references and native session snapshots unchanged. Separate tests cover trusted touch, subsequent editor gestures and focused action-row activation once. An initial full matrix exposed passive-listener stale state on immediate reopen; the adapter now installs palette keyboard listeners in a layout effect, with no test delay added. Geometry assertions cover all six projects. Unit fixtures verify one-shot close, modal exclusion, cleanup and exact single-button insertion alongside unchanged supplied render content.

Final full six-project Quick Actions/menu/Settings shell **252/252**, functional **87/87**, support **62/62**, Go/vet/build/hook pass. Logs `/workspace/tmp/gi-close-{red,layout,final,support-final}.log`; shared-only report file `shared-close-results.json`. Screenshots are attached from the final run. Coverage **81/236 Classic**, **24/42 shared**, **155/18 unmapped**. Review found the grid-column and focused-row risks; both have direct regression evidence.

Terminal disposition: retain Escape cancellation on explicit selectors; a terminal Close button would add clutter without a new capability. No idle rows, terminal key changes or new terminal acceptance credit.

Deployment: pushed `d0b209e`, restarted8090 with `/tmp/gi-scheduler-bin`. Read-only Chromium/WebKit at390×844,820×1180 and1440×900 passed Close pointer/touch/Enter/Space plus Escape, immediate reopen,28px target and retained focus/draft. No mutation attempts or page errors; HTTP200/61 sessions, SQLite integrity `ok`, foreign-key check empty. Logs `/workspace/tmp/gi-close-live.{mjs,log}`. Final phone/desktop captures attached; deterministic rebuild left a clean worktree.

## Shared4 native Quick Actions ranking and activation — 2026-09-24

An independent case in `tests/ux/quick-actions.spec.mjs` opens the palette by typing `o` in the idle native Conversation region. The initial query, focus and single palette are checked before any filter editing. Matching session titles/JIDs, workspace actions and supported slash commands are compared in native group order against actual API catalogues. Clearing the query verifies the full catalogue separately.

Three native sessions use competing visible titles: substring first in API order, prefix second, exact third. Exact and prefix queries must select the stronger title despite its later position. A `gi:` query matches multiple native JIDs but no title, forcing the first-result fallback. ArrowDown/ArrowUp wrap within the filtered three-row list without changing sessions or drafts. Enter activates highlighted Open explorer once, observed as one real closed-to-open workspace transition with a native tree file visible. Generic action activation satisfies shared4; shared16's command insertion and failed activation criteria remain unmapped.

The captured composer contains text, unsent media, a workspace reference and the exact native message reference. IndexedDB must finish persisting the complete draft, including media bytes, before setup reload. The first matrix exposed this fixture write race; waiting for the durable record replaces an implicit rendering-based assumption. Full draft records remain equal through navigation, action activation and reload, with no mutation requests or extra turns. Initial fixture inspection also corrected top-level native Quick Actions settings fields rather than changing API behaviour. Review requested grouping evidence under the first typed character and multiple-result fallback, both now explicit; one review attempt timed out without findings.

Final six-project Quick Actions/menu/Settings shell **258/258**, functional **87/87**, support **62/62**, Go/vet/build/hook passed. Logs `/workspace/tmp/gi-rank-{final,support-final}.log`; `shared-ranking-results.json` contains only shared4 evidence. Coverage **81/236 Classic**, **25/42 shared**, **155/17 unmapped**. Test-only, no runtime changes or restart.

Terminal disposition: keep explicit, bounded selectors with stable native ordering, Enter activation and Escape draft/cursor restoration. A global printable-key palette would steal normal editor input and is not a suitable terminal adaptation. Existing Alt-S/Alt-M evidence remains separate; no new idle rows or terminal acceptance credit.

## Shared33 native session typeahead — 2026-09-24

The session picker lacked incremental typeahead when focus left its search field. `web/src/gi-session-typeahead.ts` uses the supplied prefix-first matcher and700ms buffer, filters disabled entries, and returns the original full-list index. It ignores native editing targets, modifiers, composition, repeats and consumed events. A guarded insertion in `patchComposePopupKeys` updates the highlight and synchronously focuses the existing entry button so native Enter activates that same entry once. Navigation resets the buffer; open/Escape already reset it in supplied code. Component source/render bytes stay unchanged.

The independent shared33 case creates12 native sessions with similar names and a pinned substring match. Native identifier, display-name and model metadata queries retain grouped order. Typing from the pinned row chooses a prefix first, then additional characters select the other prefix without changing the search filter. Home/End, ArrowUp/Down and PageUp/Down traverse enabled filtered entries; Escape restores trigger focus, current selection and draft/media. One Enter produces one observed native localStorage selection and no session write, followed by origin return/reload. Unsupported Delete remains absent. Helper tests explicitly cover disabled-index remapping and native editing guards. A functional test covers prefix focus and actual Enter selection.

Sessions **126/126**, focused shared **6/6**, model/Quick Actions/menu regression **204/204** (six fresh isolated34-test projects), functional **88/88**, support **64/64**, Go/vet/build/hook passed. Review found no blocking issue. Initial test ordering was corrected to scope native records by unique test IDs; two long regression runs hit WebKit internal reload errors in unrelated fixtures before assertions. Final fresh-project runs retain the same tests and assertions. Logs `/workspace/tmp/gi-typeahead-{session,standard,support-final,regression-*}.log`; shared-only report file `shared-typeahead-results.json`. Coverage **81/236 Classic**, **26/42 shared**, **155/16 unmapped**.

Terminal disposition: preserve Alt-S's explicit bounded selector, focused search editing, native enabled-entry ordering and Escape restoration. Do not add another selector or idle row. Browser typeahead receives no inferred PTY credit; existing terminal tests and outstanding terminal mutation/media/queue work stay separate.

### Shared16 frozen prefill conflict

`tests/ux/features/classic/canonical/canonical-ux.feature` original007 requires: “the composer text is replaced with the command prefill”. Shared16 requires: “command insertion preserves the existing composer draft and does not submit it”. The same ordinary Quick Actions activation cannot satisfy both for a nonempty draft. Both frozen files remain unchanged and current replacement behaviour retains its existing Classic mapping. Shared16 remains unmapped, including its recoverable failed-activation requirements. Any future preservation/recovery mechanism or explicit safety deviation needs its own agreed contract and independent evidence; changing labels or testing an empty draft would not resolve this conflict.

Deployment: pushed `b39b182`, restarted8090 using `/tmp/gi-scheduler-bin`. Read-only Chromium/WebKit at390×844,820×1180 and1440×900 verified typeahead over existing native session names, matching actual button focus, unchanged search text, Escape trigger restoration and retained current session/draft. No mutation attempts or page errors; HTTP200/61sessions, integrity `ok`, foreign-key check empty. Evidence `/workspace/tmp/gi-typeahead-live.{mjs,log}`. No live session was created, renamed or selected by Enter; deterministic rebuild left the worktree clean.

## Shared34 native model search and keyboard ownership — 2026-09-24

The supplied picker had no search field and arrow navigation counted blocked models. `scripts/patch-model-picker.mjs` now adapts it after the existing popup-key adapter. Every replacement requires one exact anchor; supplied component bytes remain unchanged. Search uses native identifier, name and context metadata, with case-insensitive AND terms and catalogue order preserved. There is one additional field only while the popup is open. Existing initial popup focus and non-search typeahead remain available; no idle controls or CSS were added.

`web/src/gi-model-picker.ts` computes navigation against enabled results and returns visible-list indices. Arrow keys wrap; Home/End and Page keys use the existing session navigation policy. Search retains text editing, including Home/End, spaces and Backspace. Prefix-first typeahead reuses the session helper and focuses the actual button synchronously. Native buttons own Enter/Space, search Enter selects the highlighted enabled result, and repeated Enter cannot dispatch a second request. Consumed, composing and modified keys are ignored. Empty/all-disabled results cannot select a model. Escape restores the trigger without a mutation; reopening clears the filter.

The independent shared34 fixture registers similar model names, a prefix/substring competitor, twelve catalogue additions and a context-blocked entry in the real local-provider registry. Assertions cover identifier/name/context filtering, authoritative order, empty and disabled-only results, editing, incremental matching, every navigation key, single activation, native model persistence, other-session isolation and exact IndexedDB text/media/reference preservation through reload. Existing shared31/32 typeahead remains valid.

Validation: **42/42** context/model tests across Chromium/WebKit phone/tablet/desktop; **294/294** unchanged session/Quick Actions/Classic regressions across six fresh isolated projects; **89/89** functional; **67/67** helper tests with **812** assertions; Go/vet/build/hook and diff checks pass. The report requires all six projects for shared34 and leaves shared35 unmapped. Aggregate coverage is **81/236 Classic**, **27/42 shared**, with **155/15** unmapped.

Initial setup used an obsolete browser-cache path; the installed browsers are under `/home/agent/.cache/ms-playwright`. A combined regression attempt hit a session-filter ArrowDown timing failure and exceeded the tool timeout; the final fresh-project runs passed without assertion changes. A draft probe initially used the wrong IndexedDB schema; it now reads `gi-session-drafts` and waits for the exact committed record. Two read-only delegated reviews timed out, so there is no independent-review claim. Local review and tests cover the adapter composition, enabled-index mapping, native button ownership, modifiers, repeats and composition. Evidence: `/workspace/tmp/gi-model-filter-{browser-final,regression-*,functional,support-final,standard}.log` and `shared-model-filter-results.json`.

Terminal adaptation: retain the explicit Alt-M on-demand selector and native `/model` command. Metadata filtering should live inside that bounded selector, with no idle search row; navigation must skip unavailable choices, preserve captured session identity and leave editor/cursor intact on Escape. Terminal typing belongs to selector search rather than globally intercepted prefix shortcuts. This browser slice changes no TUI code and earns no new PTY credit. Filtering and enabled navigation still need independent regular/fullscreen checks at 60×18, 100×22 and 140×36 before terminal acceptance.

Deployment: pushed `9b745d1` and restarted port8090 using `/tmp/gi-scheduler-bin`. Read-only Chromium/WebKit at390×844,820×1180 and1440×900 verified native catalogue filtering, empty-result Enter, native editing, Home/End enabled-button focus, filter reset, Escape trigger restoration and retained current model/session/draft. Zero API mutation attempts or page errors; HTTP200/61sessions, integrity `ok`, no foreign-key errors, deterministic build/clean tree. The live probe waits for the authoritative session model rather than capturing the initial default-model placeholder. Evidence `/workspace/tmp/gi-model-filter-live.{mjs,log}`. No live model or session was changed.

## Terminal SSE occurrence identity and shared36 gap — 2026-09-24

Investigation of shared35 found that Gi emits measured provider usage or unavailable context, not local estimates. Existing model-race coverage alone cannot establish its full capability contract. No shared35 mapping was added.

A native shared36 probe found two independent behaviours. First, terminal SSE notifications lacked turn IDs, and replaying a recorded old idle frame removed the newer run's Stop control while its authoritative activity read was held. Second, cancellation calls `cleanupTurnRun`, which starts the next queued turn. That means the frozen “composer and queue are preserved” requirement does not currently have unchanged-queue evidence. Changing cancellation/queue policy is outside this bounded event fix; shared36 remains unmapped. The regression is Gi-only, not a shared36 acceptance tag.

`internal/turn/engine.go` now includes `turn_id` on idle status, response completion and final assistant post notifications (including native shell posts). `web/src/gi-turn-event.ts` recognises mismatched terminal events. They still request authoritative refresh and can carry persisted message content, but do not invalidate the current run's controls or clear its transient status/draft/thought buffers. Untagged legacy terminal frames follow the same refresh-only rule while a current run is known. Matching current-turn events retain existing behaviour. Session selection, replaced-source guards and HTTP revisions remain unchanged; no supplied component was edited.

The browser test records real SSE wire frames through the existing disconnect proxy. It verifies captured cancellation, same runtime asset epoch across reconnect, native FIFO queue advancement without tail loss, independent-session activity, a stale cancellation409 and exact repeated old idle/completion replay. Activity GETs are held while the newer Stop control and a fresh native preview delta remain visible. Reconnect intentionally drops pre-readiness transient text, so preview assertions use the provider's `.more` gate after the new readiness frame. This is not synthetic DOM/SSE state. An engine test verifies event identities and a functional test checks the native shell post wire payload.

Validation: focused **6/6**, full reconnect **66/66**, functional **90/90**, helper **68/68** with825 assertions, Go/vet/build/hook, turn/web/TUI race ×3. Existing fullscreen outcomes, regular-mode history/editor, session/model selectors, smoke and Gherkin suites passed at60×18,100×22,140×36. Initial repeated race testing exposed an unclosed test-only memory store; the test now closes it before reruns. The TUI build used a temporary binary path because `bin/gi` points to the running dev binary; its symlink was restored after testing. A bounded delegated review reported no blocking issue, with the documented assumptions that native status values are running/cancelling/idle and preview state belongs to a known current turn.

No frozen mapping changed: **81/236 Classic**, **27/42 shared**, **155/15** unmapped. Evidence `/workspace/tmp/gi-terminal-sse-{standard,functional,browser,reconnect,race,tui}.log`; red stale-frame evidence `/workspace/tmp/gi-shared-stop-red.log`.

Terminal adaptation: keep `/cancel` and focused Escape on the existing run controls; do not add an idle row or persistent queue pane. Terminal cancellation and queue actions must capture the session and turn identity, report conflicts rather than retargeting a newer run, and document whether queued work advances. The existing TUI generation guards and regression suites remain separate evidence; this browser fix earns no new terminal queue acceptance.

Deployment: pushed `73613d4`, restarted8090 using `/tmp/gi-scheduler-bin` (PID695304). Reused the read-only six-browser-size model/shell probe against the deployed bundle: filtering, navigation, focus and draft retained, no mutation attempts/page errors, HTTP200/61sessions. Integrity `ok`, foreign-key check empty, deterministic build/clean tree, `bin/gi` symlink restored. Evidence `/workspace/tmp/gi-terminal-sse-live.log`; no live turns were started or cancelled. Stale terminal replay proof remains on isolated native instances, not production.

## Compact terminal Alt-M adaptation — 2026-09-24

The model selector now filters on authoritative label/provider/ID/display name, formatted context capacity and advertised reasoning support. Case-insensitive AND terms retain native catalogue order and model identity. It remains a title, one search/error line and at most six result rows; there are no extra idle rows or new global printable-key bindings. The pi-tui SelectList reference informed bounded rows, focus-owned input, Enter selection, Escape cancellation and dim descriptions. No browser typeahead behaviour is imposed on terminal text editing: printable characters belong to the existing selector query.

`internal/tui/model_picker.go` snapshots native model metadata and latest measured context only while the picker is open or an unavailable row is explicitly retried. Known unusable credentials, missing catalogue entries, too-small context and failed context reads are dimmed with an inline unavailable marker/reason. Arrow wrapping, Home/End and Page navigation skip those rows without changing underlying IDs. An unavailable-only result has no selection. Enter retries its metadata; recovery highlights the row but does not silently apply it. A second Enter is required. Enabled acceptance still calls `chooseSessionModel`, so context/credential changes after opening are revalidated. The captured session generation rejects both A→B and A→B→A retargeting. No global model settings are written.

The initial regular-mode PTY check found that inline selector resize could place menu fragments in scrollback. Alt-M now temporarily uses the alternate screen while retaining go-tui's inline renderer, following the existing Alt-I pattern. It grows the temporary region to the bounded menu, restores the saved dock height before returning to the main screen, and preserves terminal-owned history. Resize geometry is re-established with the existing resize notice. Closing clears modal visibility before dispatching the restoration resize and restores the editor's input ownership. Exit cleanup also restores the main screen. Fullscreen keeps its existing inline selector geometry and colours.

Independent native acceptance is in `tests/tui-model-picker/main.go` plus `scripts/test-tui-model-picker.mjs`, exposed by `make test-tui-model-picker`. The isolated process registers a deterministic catalogue and a clearly marked stored context-measurement fixture; it never calls an external provider. Six PTYs verify metadata queries, enabled-only navigation/wrapping/paging, all-disabled/empty results, resize, native persisted selection, other-session/global-settings isolation, no prompt submission, exact draft insertion point, Unicode multiline editor, dock rows, and retained regular-mode history. Unit tests add stale context revalidation, generation mismatch, read failure and in-place retry recovery. The older selector harness now recognises the disabled marker without relaxing its row-count, draft or rejection assertions.

Validation: six new PTYs pass; existing session/model, regular, outcome, smoke and Gherkin suites pass at three sizes; Go/vet/build/hook, TUI race ×3,90functional and68helpers pass. Review identified cached negative metadata preventing recovery; the explicit retry path fixes it. Follow-up review found no serious issue; its screen-restoration ordering nit was fixed and all six new PTYs rerun. Evidence `/workspace/tmp/gi-tui-model-{pty-final,standard,race,functional,regression}.log` and `test-results/tui-model-picker/` captures. Hardware-cursor coordinates are not acceptance criteria for a renderer that hides that cursor; insertion of a marker into the preserved draft verifies the actual logical cursor instead.

Browser mappings remain **81/236 Classic**, **27/42 shared**, **155/15** unmapped. This is separate terminal adaptation evidence, not an additional frozen web mapping. Queue actions, pending-media recovery and mutation submenus remain open.

Deployment: pushed `5ae2196`, restarted8090 using `/tmp/gi-scheduler-bin` (PID736156), no web bundle changes. Read-only six-browser-size shell/model smoke passed with retained focus/drafts, zero writes/errors, HTTP200/61sessions. Integrity `ok`, foreign-key check empty, tree clean and `bin/gi` restored to the live binary. Evidence `/workspace/tmp/gi-tui-model-live.log`; final successful text/ANSI captures and summary attached as `gi-tui-model-picker-evidence.tar.gz`. No live session was mutated; terminal proof used isolated databases.

## Native Classic shell002/003/005 evidence — 2026-09-24

`tests/ux/shell.spec.mjs` adds three independent frozen cases without changing runtime code or supplied sources:

- shell002 creates a native hidden file, exercises pointer and keyboard menu toggles, and checks exactly one `piclaw:toggle-hidden-files` event per activation, exact `workspaceShowHidden` persistence, native tree reload/visibility and persistence after reload.
- shell003 checks all four workspace actions are disabled while the workspace is hidden, rejects real pointer attempts and native focus, skips them during Tab traversal, and observes no workspace actions, writes or setting changes. It repeats after a visible→hidden transition. The frozen title mentions chat-only, but its actual Given is “the workspace is not visible”; no URL-driven chat-only mode support is claimed.
- shell005 measures the compose wrapper and its inner input against the available native chat-column content box. It checks left alignment, width and document overflow with the workspace hidden/visible, narrow drawers, desktop sidebar, resized viewport and reload. This does not infer mobile safe-area or display-scale behaviour.

Every case snapshots exact committed IndexedDB text/media bytes and verifies the current session, attachment pill and absence of submitted turns. Six browser projects passed **18/18** focused and **162/162** shell/Classic/workspace/Settings regressions. Functional **90/90**, helper **68/68** with839 assertions, Go/vet/build/hook and diff checks passed. Independent review found no acceptance gap for these three cases. Initial fixture mistakes assumed a `workspace-open` class and empty-array rather than nullable turn results; assertions now use the actual native class and empty-result contract. No runtime change was needed.

Report guards keep five-project evidence partial, require all six projects and leave shell001/004/006/008 unmapped. Aggregate coverage is **84/236 Classic**, **27/42 shared**, **152/15** unmapped. Evidence `/workspace/tmp/gi-shell-{browser,regression,functional,standard,support-final}.log`; `native-shell-results.json` is the focused report input. Test-only commit needs no restart; read-only live sessions200/61, integrity `ok`, foreign-key check empty.

Terminal disposition: full-width editor and zero additional idle controls remain covered by existing three-size PTYs, not these browser results. A future hidden-file option belongs inside an explicit file selector or scoped search, not a permanent toolbar. Disabled capabilities must stay non-actionable in bounded selectors and mutation submenus. There is no implied terminal workspace tree, browser scale, mobile safe area or new terminal credit here.

## Standalone display-scale visibility fallback — 2026-09-24

The supplied viewport writer recognises `navigator.standalone`, but the supplied control CSS only recognises display-mode media queries. A capability fixture therefore applied a saved scale while hiding its control. `web/src/gi-display-scale.ts` wraps the unchanged scale synchronisation and maintains one host class using the supplied standalone detector. `gi-display-scale.css` reveals the existing control for that class. Ordinary tabs still hide it; desktop standalone can show the stored preference without receiving mobile viewport scaling. No supplied component, utility or CSS bytes changed, and no new menu item or idle control was added.

`tests/ux/shell.spec.mjs` uses six browser/size projects, emulating only standalone/touch capability. Native inputs, localStorage, viewport writes, real cross-tab storage events and draft persistence run unchanged. Tests cover the saved90% initial value, Enter and blur commits,85/110/100/80/95% values, reload, ordinary and non-touch gates, cross-tab synchronisation and exact draft/media preservation without prompt submission. These are not physical installation, device safe-area or pixel-scale measurements.

Shell008 remains **unmapped**. Its frozen Given only says the workspace menu is open, whereas the shipped implementation and this test require standalone capability to show the control. Review identified that missing precondition; the executable scenario is now Gi-only and neither the frozen source nor catalogue mapping was changed. A future contract decision must resolve that discrepancy explicitly.

Review also found a throwing `matchMedia` startup path. The host now catches unsupported queries like the supplied detector; helper tests cover modern/legacy listeners, native bounds/migration values, fallback and cleanup. Follow-up review found no remaining issue. Full browser regression **168/168**, functional **91/91**, support **72/72** with890 assertions, Go/vet/build/hook and diff checks passed. The first workspace regression captured a completed turn before `finished_at`; it now waits for the native finished timestamp and idle activity before taking the unchanged-state baseline. A parallel Go build collided with bundle regeneration; final Go/functional checks ran serially against the finished bundle. Logs `/workspace/tmp/gi-scale-{regression-final,functional-final,standard-final}.log`.

Terminal disposition: display scale belongs to the terminal emulator's font/zoom settings, not another Gi selector or footer row. Keep width-aware truncation and resize restoration in the existing TUI. This browser host fix adds no terminal controls or acceptance credit. Coverage remains **84/236 Classic**, **27/42 shared**, **152/15** unmapped.

Deployment: pushed `a3f6c8a`, restarted8090 using `/tmp/gi-scheduler-bin`. Final guarded live smoke (`/workspace/tmp/gi-scale-live.{mjs,log}`) selected an existing session and blocked API mutations. Both browsers at three sizes passed saved scale/control/viewport application, ordinary-mode hiding, Escape/focus and retained draft, with zero write attempts/errors. Integrity `ok`, foreign-key check empty, deterministic build/clean tree.

Live-probe caveat: the earlier fullscreen capability probe (`/workspace/tmp/gi-scale-capability.mjs`) opened the live app without selecting an existing session or blocking writes. Bootstrap created `session_1790227382181629026` (`@web`,2026-09-24T05:23:02.181Z); read-only inspection confirms zero messages and zero turns. Session count is now62 rather than61. This was an unintended live mutation, not acceptance setup, and has been disclosed. No deletion or corrective live mutation was attempted. Future capability probes must use an isolated instance or block mutations and select an existing session before navigation. The final deployment check, unlike that preliminary probe, was guarded and read-only at the API layer.

## Native Settings open and cached reopen — 2026-09-24

Two independent acceptance cases in `tests/ux/settings-shell.spec.mjs` cover Settings001 and dialog002. They open the native menu action, assert the visible header and navigation, General first/current, and active config values obtained from the real API. Supported pane traversal verifies only navigation and headings. After returning to General and closing, the test holds a real200 refresh response and reopens with Control+comma. The dialog and exact cached General values must appear within1s, without the cold General loading message, while the fresh response is still held. Releasing it preserves the values. This proves a cache-backed reopen, not absence of network requests, offline operation or every pane's data caching.

Exact IndexedDB text/media bytes, session identity, attachment pill and zero submitted turns are checked through dismissal and reload; no settings/session mutations occur. The first fixture compared `innerText` to `textContent`, producing false whitespace mismatches; the final equality uses the same native text representation. Source criteria were not changed.

Validation: **12/12** final focused tests, **234/234** Settings/shell regressions, **91/91** functional, **72/72** helper tests with898 assertions, Go/vet/build/hook and diff checks pass. An initial file review timed out. A subsequent bounded criterion review judged the coverage sufficient subject to explicit header/nav assertions; those were added and all twelve cases rerun. Five-project report inputs remain partial; Settings002 stays unmapped. Aggregate **86/236 Classic**, **27/42 shared**, **150/15** unmapped. Evidence `/workspace/tmp/gi-settings-cache-{browser-final,regression,functional,standard,support-final}.log` and `settings-cache-results.json`.

Terminal adaptation: preserve explicit on-demand settings/model selectors and existing shortcut ownership, with Escape returning to the editor/cursor and no idle panel. Browser Settings cache timing adds no terminal rows, network-cache requirement or inferred PTY acceptance. Test-only change; no deployment restart required. Read-only live check: HTTP200/62sessions, integrity `ok`, foreign-key check empty.

## Read-only workspace tabs and shell007 — 2026-09-24

The workspace callback previously created tab labels with an empty host. `WorkspaceTab` now reads the native rooted preview API with a20,000-byte limit and mounts the existing supplied preview extension in `view` mode. `gi-workspace-tab-lifecycle.ts` owns each request/instance: switch, refresh or close disposes once and invalidates late success/error responses. A real404 removes stale content and offers explicit Retry. Markdown uses the existing sanitising renderer; plain text remains escaped. Preview metadata retains truncation and native file details. This is read-only viewing, not a file editor.

A guarded build adapter changes only the existing workspace open action's visible wording to “Open read-only tab”; source files remain unchanged. The host suppresses the tab context menu rather than exposing unsupported editor, pin, pop-out, comparison or specialised-viewer actions. Close buttons remain visible and pointer-enabled without hover. On drawer layouts, opening a tab closes the drawer and gives the preview the chat column until the final tab closes. The meters overlay is hidden while a read-only tab is open so it cannot intercept tab clicks. The idle shell is unchanged. Deferred last-close focus checks its epoch, active target and Settings ownership before returning to the composer.

The independent shell007 case opens three native files, closes background tabs by pointer and keyboard, and observes unchanged active content with no new preview read. It verifies active/last close, no unsupported context menu and exact committed draft/media/reference retention through reload. Gi-only cases hold a real native response across close, verify it cannot remount, remove/restore an isolated source through native tools for404/retry proof, check escaped text and focused Ctrl+W, and exercise a same-frame Settings focus race. The latter uses synthetic key delivery only to align timing; close activation evidence uses native keyboard/pointer input. File selection still deliberately adds references; the retained reference set is captured before close assertions.

Validation: **18/18** final focused tests across six browser projects, **168/168** workspace/shell/Settings regressions, **92/92** functional, **75/75** helpers with925 assertions, Go/vet/build/hook and diff checks. A full file review timed out; a bounded implementation/contract review found no blocker, conditional on error loading reset and safe rendering, both verified in code/tests. Five-project report inputs remain partial; workspace009/010 and shell009 stay unmapped. Coverage **87/236 Classic**, **27/42 shared**, **149/15** unmapped. Logs `/workspace/tmp/gi-tabs-{browser-final,regression-final,functional-final,standard-final,support-final}.log`; screenshots under `test-results/ux-parity/artifacts/workspace-tabs--ux-shell-0-41946-y-tabs-never-activates-them-*/`.

Terminal adaptation: do not add a persistent tab strip or file tree. Future file viewing belongs to an explicit bounded read-only viewer or external pager, with captured path, explicit retry and Escape restoring editor/cursor and reading position. Existing Alt-I index and terminal copy/search tests do not prove a terminal file viewer. No new terminal rows or acceptance are claimed here.

Deployment: pushed `f0db954`, restarted8090 using `/tmp/gi-scheduler-bin` (PID882686). Guarded live smoke selected an existing session before navigation, blocked API mutations and opened existing `AGENTS.md` read-only. All six browser/size combinations rendered native content, closed the preview and retained the draft/session with zero writes/errors. HTTP200/62sessions, integrity `ok`, foreign-key check empty, deterministic build/clean tree. Evidence `/workspace/tmp/gi-tabs-live.{mjs,log}`. Phone/desktop screenshots from isolated fixtures attached; no live file/session mutation or turn submission.

## Alt-S temporary screen and session-owned acceptance — 2026-09-24

The new regular-mode PTY case reproduced stale session-selector rows in terminal history after resizing. Alt-S now reuses Alt-M's temporary alternate-screen lifecycle while retaining the inline renderer. On resize the host clears only that temporary visible screen with cursor-home/ClearToEnd, never `ESC[3J` or main scrollback. This also removes stale reflowed rows above a full-height model selector. Escape and successful selection restore the saved dock height and main screen; exit cleanup follows the same lifecycle. The selector remains title/search plus at most six results, with unchanged idle rows.

Session selection captures the same generation token as model selection. A stale A→B→A picker cannot act on a newer session view. Acceptance performs the native `switchSession` read while the selector is still owned, keeping it open with a retryable error on failure and closing only after success. An initial preflight-read approach was removed after review identified a second-read handoff gap. The origin editor/subscription remains intact on failed native lookup; reopening clears old selector errors. Thinking selection uses the common close cleanup but otherwise retains its existing behaviour.

`scripts/test-tui-session-picker.mjs` and `make test-tui-session-picker` cover60×18,100×22,140×36 in fullscreen and regular modes. Fixtures create ten sessions through `/fork`, build history through native turns, and verify Arrow/Home/End/Page navigation, resize of the selected tail row, empty search, cancellation, A/B Unicode drafts and insertion points, multiline layout, no new turns/settings writes, retained regular history and exit with the picker open. Raw ANSI captures assert no scrollback erase; terminal flags return to main-screen/no-mouse after exit. Unit tests cover stale generations and native database-failure/retry ownership.

All six session PTYs and six model PTY regressions pass. Existing session/model, regular, outcomes, smoke and Gherkin suites pass; Go/vet/build/hook,92functional/75helpers and TUI race ×3 pass. Follow-up bounded review cleared the handoff finding. Browser feature mappings remain **87/236 Classic**, **27/42 shared**, **149/15** unmapped. Evidence `/workspace/tmp/gi-session-picker-{red,pty-final,model-regression,regression,standard-final,race-final,functional}.log`; text/ANSI captures under `test-results/tui-session-picker/`.

This is independent terminal selector acceptance, not browser credit or terminal mutation-submenu support. No toolbar, persistent tab strip, queue pane or idle status row was added.

Deployment: pushed `4797dad`, restarted8090 using `/tmp/gi-scheduler-bin` (PID915887). The guarded existing-session/file live probe passed both browsers at three sizes with zero API write attempts/errors and retained preview-close/draft behaviour. HTTP200/62sessions, integrity `ok`, foreign-key check empty, deterministic build/clean tree. Evidence `/workspace/tmp/gi-session-picker-live.log`; terminal text/ANSI snapshots and summary attached as `gi-session-picker-evidence.tar.gz`. Live TUI sessions were not touched; terminal proof uses isolated PTYs/databases.

## Read-only Piclaw pane host conformance (2026-09-24)

The [named subset contract](pane-host-subset.md) pins Piclaw bfc34e4eb rather than
implying general compatibility from copied interfaces. Native bounded text,
mtime/full-size metadata, capability rejection, resize and captured pane-close
callbacks now have independent conformance fixtures. Refresh remounts and stale
callbacks cannot dispose a newer occurrence. Asynchronous reads do not steal
focus. Supplied pane bytes and frozen features are unchanged.

Validation: 24 focused and 174 combined workspace/shell/settings cases across all
six browser projects, 92 functional cases, 80 support tests/963 assertions,
Go tests/vet and hook checks. The interrupted regression run was discarded;
the fresh full run passed. Independent review raised arbitrary host callback
failure/reentrancy assumptions; the contract now states that these internal
Preact/shell callbacks are trusted, distinct from guarded extension hooks.

Coverage remains **87/236 Classic + 27/42 shared**, **149/15** unmapped. No editor,
dirty/save, retained instance, dock/pop-out or general-extension claim. Terminal
preview adaptation is design-only: temporary bounded surface, Escape restoration,
no idle rows/tab strip/dock; six independent PTY checks are required before credit.

Deployment: pushed `34d163f`, rebuilt/restarted the Gi dev instance on port 8090
(PID 988953). The guarded read-only live check passed Chromium/WebKit at
390x844, 820x1180 and 1440x900: preview open/close, draft and selected-session
preservation, zero API mutations and zero page errors. `/api/sessions` returned
200 with 62 sessions; read-only SQLite integrity returned `ok`, foreign-key
check returned no rows. Logs: `/workspace/tmp/gi-pane-live.log`; bundled results:
`/workspace/tmp/gi-pane-host-evidence.tar.gz`.

## Read-only tab MRU and pinning (2026-09-24)

Gi now subscribes to the supplied Piclaw tab store rather than maintaining a
second tab array with insertion-order fallback. Opening/activating records MRU;
closing the active tab chooses the most recently used remaining tab. Close
Others and Close All preserve pinned tabs, while an explicit single-tab close
can remove a pin. Pin state is page-local, not persisted or session-owned.
Every store change publishes tabs and active ID together. Last-close focus
restoration is fenced by epoch, remaining tabs and modal ownership.

The supplied tab-strip/store bytes are unchanged. An opt-in exact-anchor build
adapter exposes only Close, Close Others, Close All and Pin/Unpin, suppressing
standalone-viewer routes (tested with CSV), and positions the requested menu
inside the viewport. Touch targets are 44px. The outside-click/Escape listener
now installs before paint; a fast Escape previously left the menu open.
Settings blocks Gi's background tab shortcuts. Browser-reserved shortcuts stay
native rather than being globally swallowed. Settings dismissal already restores
the connected opener or composer; the same-frame last-close test now explicitly
checks composer focus after Escape.

Frozen workspace-011 first failed the insertion-order fallback. Final coverage
pins A before A→B→close B, checks A instead of insertion-tail C, then tests both
bulk closes with the pin intact and explicit pin closure. Exact IndexedDB
text/media bytes, selected session, empty turns and reference preservation remain
asserted. Synthetic edge-position stress supplements, but does not replace,
trusted context-menu interaction. A focused independent review raised these
coverage gaps; the strengthened fixtures pass, and a bounded follow-up review
accepted the described resolution and explicit browser shortcut policy.

Validation: 186/186 combined pane/workspace/shell/settings cases across Chromium
and WebKit at phone/tablet/desktop sizes; 93/93 functional cases; 82/82 support
tests with 1,004 assertions; Go tests/vet and hook checks. Helper syntax/optional-
callback expectations and a functional hamburger locator were corrected before
these final runs. Five-project evidence remains partial; the report maps only
workspace-011 after the full matrix. Logs: `/workspace/tmp/gi-tab-mru-*.log`;
native result snapshot: `test-results/ux-parity/workspace-mru-results.json`.

Workspace-005 stays open: its hidden header menu also requires unavailable create
and upload actions. Editing/dirty-save/retained pane instances/pop-out remain
separate unsupported work, not inferred from pinning.

Terminal disposition: MRU/session navigation can reuse temporary ≤6-result
pickers; a pin is a navigation preference, not justification for persistent tab
chrome. File-preview pin/MRU support is not implemented in the TUI. Any future
adaptation must preserve logical cursor, Unicode/multiline drafts, history and
session ownership in both modes at all three PTY sizes, with no extra idle rows.
This browser slice changes no terminal code and claims no terminal acceptance.

Deployed `2dd52fc` on port 8090 (PID 1047425). Six guarded live browser/size
checks opened existing `AGENTS.md`, pinned it, kept it through Close All and
explicitly closed it with the draft/session intact. No API mutation attempts or
page errors; session API 200/62 sessions, SQLite integrity `ok`, no foreign-key
violations. Logs: `/workspace/tmp/gi-tab-mru-{deploy,live}.log`. Terminal sessions
were untouched. Supplied files and frozen feature source remain unchanged.

## Terminal pending attachments (2026-09-24)

`/attach <path>` and `/paste-image` used to store a file without including it in
the next prompt. They now stage native references in a map owned by the source
session. The optional inline prompt uses the same path. `/attachments` lists up
to six pending refs; `/detach <media:id|all>` removes pending references, not
stored files. Six slots include in-flight refs, so staging while admission is
pending cannot overflow the limit on rejection. File reads are bounded at 10 MiB
and reject non-regular files before reading.

This is a process-local terminal adaptation, not a persistent attachment tray.
The editor/footer layout is unchanged. Commands print ordinary transcript
feedback; no added idle row, tab strip, pane or modal is needed. Session switches
preserve staged refs and the existing Unicode/multiline draft and logical cursor.
No-model and directed `@agent` sends with pending refs are rejected before clearing
the editor. Slash commands and local `!!` shell commands do not consume refs.

An ordinary submission moves refs into one captured claim, blocking a second
ordinary submission from that source until settlement. Claimed media uses native
same-session `SubmitPrompt`, not implicit peer routing. This avoids creating or
switching a peer session before media ownership is checked. No-media prompts keep
the existing routed path; queue/steering policy is unchanged. Pending attachments
are restored after a confirmed rejected admission, but submitted prompt text is
not restored over newer input. Retry is explicit: review refs, then re-enter or
recall the desired prompt. Already accepted queued refs belong to the native turn;
the existing text-only queued-draft restoration is not extended by this slice.

Claim completion runs on the UI loop even if another session is selected, and
only the exact claim can settle. A rejected submission is checked against native
turn/steering/message metadata with a source-session token. The native call
returns after its SQLite write or rollback; this is not HTTP or eventual-consistency
reasoning. A durable match consumes refs even if launch/queue-count sync returned
an error. A failed authority read holds the claim and blocks resend;
`/attachments` explicitly retries a bounded one-second check. It does not
silently assume rejection or delete media. Ordinary input/cursor and another
session's refs are untouched by background settlements, including A→B→A.

The regular renderer exposed an existing rejection bug: the thinking indicator
remained active after admission failed, preventing final transcript output.
The error path now finishes that indicator before reporting failure. Independent
PTY failure/retry checks cover both terminal renderers.

Validation: six PTYs at 60x18, 100x22 and 140x36 cover A/B pending refs, Unicode
cursor insertion, multiline resize, real SQLite-triggered pre-admission rejection,
explicit retry with exact native file bytes/metadata, next prompt without reuse,
other-session non-submission, detach preserving stored bytes and unchanged settings
and five-row idle dock. The harness waits for resize settlement and measures the
current dock, not historic separators in terminal-owned scrollback. Separate Go
tests inject a real post-INSERT queue-count failure, check no reattachment, hold
unknown reads until authority returns, exercise slot reservation and clipboard
staging. Go tests/vet/hook checks, TUI race ×3, existing sessions/regular/outcomes/
smoke/Gherkin suites, 93 functional browser cases and 82 helpers/1,004 assertions
pass. Diagnostics have no Go/mjs validator; compiler/tests are the evidence.

A file-based delegated review timed out. A bounded inline review questioned
read-after-error ordering; source inspection and native post-INSERT fault proof
resolved that issue, and a follow-up reviewer found no demonstrated duplicate-send
path within this model. This is not crash/restart, remote delivery or cross-process
idempotency. Staged refs are lost on process exit, while their stored media bytes
remain. Unknown DB reads deliberately trade availability for no automatic resend.

Evidence: `/workspace/tmp/gi-tui-media-{pty-final,standard-final,race-final,regression,functional,support}.log`
and `test-results/tui-pending-media/` captures/native records. Coverage remains
**88/236 Classic + 27/42 shared**, **148/15** unmapped. Terminal mutation submenus,
restart-durable media drafts and queue recovery still need their own acceptance.

Deployed `ef65ccd` on port 8090, PID 1108465. The existing six-size Chromium/WebKit
read-only smoke passed pin/bulk-close/preview/draft preservation with zero API
mutations or page errors. `/api/sessions` returned 200 and 62 sessions; SQLite
integrity `ok`, foreign-key check empty; binary symlink restored and tree clean.
No live terminal session was manipulated. Successful six-PTY captures, native
media records and logs are attached as `/workspace/tmp/gi-tui-media-evidence.tar.gz`.

## Session-owned upload cancellation (2026-09-24)

The transient upload status now has a Cancel uploads button. It aborts a snapshot
of captured upload batches belonging to the visible session, not another
session's transport and not a message already being submitted. The button exists
only during upload work; idle composer geometry is unchanged.

Each captured media submission registers its own AbortController. The guarded
build adapter checks its signal before the sequential upload loop and again
before constructing/dispatching the message, passes it to each native XHR, and
unregisters before message dispatch and in final cleanup. Supplied composer source
bytes stay unchanged. Cancellation uses the existing captured-draft failure path:
restore exact file bytes/references and merge newer typing/media in the origin,
even if the user switched sessions. Retry is explicit; the browser does not
submit a cancelled message or resume its remaining files automatically.

The API adapter checks cancellation before and after bounded multipart encoding,
removes abort listeners and progress handlers on settlement, ignores late XHR
callbacks, and cleans up synchronous open/header/send errors. Encoding itself
cannot be interrupted; cancellation fences the later network start. The short gap
between sequential files remains batch-owned. Once message dispatch can start,
cancellation is no longer offered rather than falsely implying recall.

Verification: 12 dedicated cancellation cases (two tests × six projects), 30 final
focused transport cases, 204 draft/Classic/shell regressions, 94 functional tests,
85 support tests/1,034 assertions, Go/vet and hook checks pass. Native responses are
held after the real server stores multipart bytes: Cancel produces a real XHR
abort, no message POST, no second-file upload, exact IndexedDB recovery/reload and
one explicit retry with the expected native media refs/bytes. Concurrent session
uploads, newer origin files/text and already-dispatched sends remain independent.
Helper tests cover early abort, listener disposal, stale completion, synchronous
XHR failure and new work created during cancellation callbacks. A missing test
locator and functional BASE_URL constant were corrected before the passing runs.

A file-based delegated review timed out; a smaller inline lifecycle review found
no blocker within this stated scope. Optional oxlint was unavailable and mjs has
no diagnostics validator; builds/tests remain the validation evidence.

This is not storage rollback. If the server accepted a file before cancellation,
its unreferenced stored media may remain, as in the pre-existing partial-failure
path. Shared39's full source-removal/durability/no-duplication contract is therefore
not claimed. Its missing Cancel control is repaired, but it stays unmapped;
coverage remains **88/236 Classic + 27/42 shared**, **148/15** unmapped. Do not
reinterpret cancellation as changing active-turn Stop or queued-follow-up policy.

Terminal disposition: process-local `/detach` removes staged refs before
admission, while in-flight native admission remains held until settled. There is
no browser XHR progress/cancel analogue to reproduce as permanent terminal UI.
This web slice adds no terminal code, idle rows or terminal acceptance credit.
Logs: `/workspace/tmp/gi-upload-cancel-{browser-final,regression,functional-final,helpers-final,standard-final}.log`.

Deployed `318f23c` on port 8090, PID 1146944. Guarded six-size Chromium/WebKit
read-only smoke passed preview/pin/bulk-close/draft checks with zero API mutation
attempts/page errors; session API 200/62, SQLite integrity `ok`, FK check empty.
Upload cancellation was exercised only in isolated fixtures, not live sessions.
Results/logs attached as `/workspace/tmp/gi-upload-cancel-evidence.tar.gz`.

## Shared39: durable attachment retry (2026-09-24)

Multipart browser uploads now use a separate create-or-reuse store operation.
Within the same session, an identical filename, MIME type and original byte
sequence returns the existing native media ID. Different files, names, types or
sessions remain distinct. JSON API and tool/TUI `CreateMedia` callers still create
new rows; only the reserved web-upload hash field is stripped from caller metadata.
There is no schema migration or retroactive enrolment of old uploads.

The lookup and insert share a SQLite write transaction. File stores use immediate
transactions; a no-op write reserves deferred in-memory connections before reading.
A server-only hash narrows candidates, then exact original bytes decide reuse,
with bounded gzip decompression. Response metadata is read before Commit, so a
post-commit caller cancellation cannot turn success into a second database-query
error. Corrupt compressed candidates fail closed without replacing identity or
changing bytes already referenced by messages. Explicit repair is required;
no silent auto-healing or duplication. Multipart temporary files are removed.

Shared39 is mapped through its disjunctive attach-file path, not inferred from
other tests. The six-project case selects a real local PNG, sees queued filename
and upload progress, holds an actual native 201 response, cancels with no prompt
or message, restores the draft, unlinks the source file, reloads, and retries from
IndexedDB. The server returns the original media ID, stores exactly one media
row, and receives exactly one successful message with that ID/session. The image
renders at its native width after reload; raw bytes still match the removed
source. Another session's draft, media and messages remain untouched. Additional
drop/paste DOM-handler tests pass but are not physical OS clipboard evidence.

Earlier partial-upload failure coverage now requires reuse of the accepted
prefix rather than expecting duplicate stored rows. That is a deliberate native
behaviour improvement; frozen source assertions and supplied component/pane bytes
are unchanged. Cancellation without a later retry may still leave unreferenced
media. Reuse avoids another row on retry, but is not garbage collection, general
message-send idempotency or crash-safe queue recovery. The shared wording does
not require those additional guarantees.

Validation: 198/198 combined draft/lightbox browser cases, 94/94 functional cases,
85 helpers/1,039 assertions, Go tests/vet/hook checks, full store+web race ×3.
Native tests cover 16 simultaneous writes from two Store instances, reopen, four
independent child processes under race, distinct-key isolation, cancelled/failed
transactions, compressed exact bytes, reserved metadata and corrupt candidates.
All repeated process/store tests pass. Initial browser failures were incorrect
image/drop/paste targets and an evidence-loader name; corrected fixtures retain
the assertions, and the clean full rerun passed.

The implementation review found the post-commit read hazard, fixed above. The
reviewer's suggested corruption fallback was deliberately rejected in favour of
explicit failure; a follow-up accepted that documented policy. A separate
criterion review accepted the attach-only tag because the trigger wording is OR,
while requiring limits on physical clipboard and broader delivery claims.

Coverage is **88/236 Classic + 28/42 shared**, leaving **148/14** unmapped. Terminal
pending refs keep their independent admission/detach behaviour; multipart retry
reuse adds no terminal rows, functionality or acceptance credit. Evidence logs:
`/workspace/tmp/gi-media-reuse-{regression-final,functional,standard,race-full,process,support}.log`;
`test-results/ux-parity/media-retry-results.json` includes six-project shared39.

Deployed `7a79e11` to port 8090 (PID 1202304). Guarded six-size Chromium/WebKit
read-only preview/pin/draft smoke passed with zero API mutation attempts/page
errors. Session API returned 200/62 sessions; SQLite integrity `ok`, FK check
empty, working tree clean. No live media upload/deduplication was exercised.
Results attached as `/workspace/tmp/gi-media-reuse-evidence.tar.gz`.

## Preview fonts and nested horizontal gestures (2026-09-24)

Read-only Markdown tabs used the browser's generic monospace default rather than
the configured editor code-font stack. A host-only CSS rule now applies
`--font-family-mono` to their code elements. Six browser projects compare computed
fonts and equal i/W glyph widths, change Appearance through native controls,
reload/reopen, and preserve exact session draft/media state. The supplied pane and
CSS sources remain unchanged. This proves preview typography, not an editable
Markdown document: independent scope review kept shell009 unmapped because its
Given requires an editor.

The broader regression run exposed an existing Safari interaction bug: the
session-swipe listener intercepted horizontal wheel events on wide Markdown
tables. The host now yields gestures to nested `overflow-x: auto/scroll` elements
whose content actually exceeds their width. The guard runs before the supplied
swipe listener, calls only `stopImmediatePropagation`, and leaves native scrolling
to the browser. It covers the full gesture, including at scroller edges; wheel
input over a table must not change sessions just because the last column is
visible. Event targets are already normalised to Element; shadow-root scrollers
are not part of this shipped renderer contract. Supplied navigation code is intact.

Verification: 24 rendering cases now pass, including WebKit's three previously
failing native wheel checks; the final column is reachable and repeated edge
wheel input preserves the selected session/draft. Thirty ordinary swipe cases
still pass, including desktop Safari's positive wheel control, Chrome/iOS
exclusions, vertical cancellation, target exclusions and rapid reverse swipes.
The full workspace/preview/shell/settings/rendering run passed 210/210 cases.
Functional coverage checks the native user table with Safari detection enabled;
95/95 passed on a clean rerun. The preceding run hit the unchanged session-
typeahead active-class assertion after native focus moved correctly; that timing
failure was recorded, not weakened. Earlier wrong preview/functional selectors
were corrected before these results. Go/vet/hook and 86 helpers/1,047 assertions
pass. A bounded inline review found no major issue and confirmed listener ordering
is part of the host contract. Diagnostics lacked optional oxlint/mjs validation.

There is a separate frozen-policy conflict: Classic original029 requires fenced
SVG to remain source-only, while shared41 requires safe SVG rendered inline. Both
cannot be satisfied by silently changing the same renderer. Shared41 remains
unmapped pending explicit policy reconciliation. Shell009 also remains unmapped;
coverage stays **88/236 Classic + 28/42 shared**, **148/14** unmapped.

Terminal disposition: terminal font selection belongs to the emulator; no web
font setting or persistent preview pane should be copied into the TUI. Existing
terminal reader/selection/picker gestures retain their separately tested ownership.
This slice changes no terminal code or idle rows and grants no terminal credit.
Logs: `/workspace/tmp/gi-preview-font-{regression-final,standard-final,support-final,functional-rerun}.log`
and `/workspace/tmp/gi-scroll-gesture-{browser,swipe}.log`; screenshots are under
`test-results/ux-parity/artifacts/workspace-tabs-Gi-read-onl-*/preview-code-font.png`.

Deployed `bda0e24` on port 8090 (PID 1259900). Guarded read-only six-size
Chromium/WebKit preview/pin/draft smoke passed with zero API mutations/page
errors; session API 200/62 sessions, SQLite integrity `ok`, FK check empty. This
live smoke is not substituted for isolated wheel/appearance evidence. Captures
and logs attached as `/workspace/tmp/gi-preview-gesture-evidence.tar.gz`.

## Startup-loaded browser skills (2026-09-24)

The browser catalogue uses the native startup discovery order (`.gi` before `.pi`, case-insensitive name deduplication). Command names use 1-64 ASCII letters, digits, underscores or hyphens, starting with a letter or digit. Eligible skills appear once as `/skill:<name>` under Slash commands and can be filtered by name or description. All sessions share the instance catalogue; there are no per-agent overrides or runtime discovery updates.

The server records each eligible file's relative path and SHA-256 at startup. Invocation rereads at most100KiB of regular UTF-8 content through `os.Root`, then checks the hash before admitting a turn. Workspace escapes fail; links resolving inside the root are allowed. Linux/macOS non-blocking opens and descriptor checks reject a swapped FIFO without hanging the request. Other platforms fail closed. Unknown, missing, invalid or changed files return400 before admission. Restoring the original bytes permits explicit retry; loading a changed or new skill requires restarting Gi.

A guarded build adapter prepends the selected skill command and a space to the existing draft without submitting, replacing media, or changing ordinary slash-command semantics. Settings and unmount checks fence late focus. The captured native session, model, media and submission token follow the existing prompt path; explicit cross-agent skill requests fail. Expanded skill content and arguments reach the native engine, with original command/name/hash stored as turn metadata. Skill content remains user-level workspace input, not a system instruction or plugin execution environment.

Verification:

* `make test-ux-skills BIN_DIR=/tmp/gi-media-bin`:12/12 across six projects. The tagged shared17 case combines canonical listing, both filters, exact prefix/draft/media retention, held POST across a session switch, native assistant marker/media bytes, unknown/stale draft recovery, reload and explicit retry. A separate keyboard/focus test preserves the trailing space and later Settings ownership. Earlier18-case runs split the tagged success/recovery steps and produced a strict-report partial matrix; merging them preserves the assertions and earns one full pass per project without changing the reporter.
* Fresh session126/126 and QuickActions138/138 regressions pass. The earlier combined run had263 passes and one WebKit held-route error in the rapid-swipe fixture; assertions and swipe sources were unchanged for the successful fresh reruns.
* Functional96/96, helpers87/87 with1,062 assertions, `make test vet bun-checks`, and `make test-web-skills` (targeted race x3) pass. The pre-boundary full web race x3 also passed; the final targeted run covers the added FIFO test. Bounded file-based review found no concrete blocker after an earlier review timed out.

Coverage is88/236 Classic and29/42 shared, leaving148/13 unmapped. Shared16's ordinary command-prefill conflict and shared36's cancellation/queue policy are unchanged.

Terminal adaptation: keep discovery inside existing slash completion, with at most six results and no idle panel, row or status badge. A future invocation path should insert the command into the existing editor without submitting, retain draft/cursor/media per captured session, and use the same bounded/hash-checked expansion at explicit submission. Escape should restore the editor without output, and errors should use the existing transient status. Fullscreen and regular mode require independent60x18,100x22,140x36 native execution, failure, resize and zero-idle-growth tests. The current terminal `/skill:` loads text only; this browser slice changes no terminal behaviour and earns no terminal invocation credit.

Deployment: `6e7f05c` pushed to main; `make restart BIN_DIR=/tmp/gi-scheduler-bin BIND=0.0.0.0 PORT=8090` serves PID1355992. Mutation-blocked Chromium/WebKit checks at390x844,820x1180,1440x900 verified27 unique real loaded skills, canonical Slash commands listing, insertion/focus and exact unsent draft retention. All six checks recorded zero API writes and page errors; no live skill was executed. SQLite integrity is `ok`; before/after counts are unchanged at62sessions,51turns,146messages. Browser captures, native/parity reports and logs are in `/workspace/tmp/gi-skills-evidence.tar.gz`. The full goal remains open at148 Classic and13 shared gaps.

## Classic skill binding and candidate gaps (2026-09-24)

Classic original008 has its own execution of the complete native skill test, with its exact frozen steps attached. The shared17 invocation/recovery proof runs separately. A helper test caught the attempted dual-tag title: the unchanged reporter recognises only its first ID, so the final18-case matrix has one Classic case, one shared case and one keyboard supplement per browser project. No criteria, supplied components or runtime code changed. `make test-ux-skills BIN_DIR=/tmp/gi-media-bin`, `make test-ux BIN_DIR=/tmp/gi-media-bin`, `make ux-parity-inventory` and `make test vet bun-checks` pass. The final focused report grants one pass in each corpus; cumulative mapped counts are89/236 Classic and29/42 shared.

Shared35 was inspected but not mapped. `internal/inference/context_usage.go` emits `provider_request` or `unavailable`, with no local estimate source. Existing model/context tests cover several neighbouring assertions, but they cannot demonstrate the frozen local-estimate labelling clause. Independent review rejected a silent pass based on the absence of estimates. Implement or explicitly resolve that capability contract before claiming the entire scenario; keep unknown values unavailable in both browser and terminal meanwhile.

The pinned `web/src/components/post.ts` and current host have no speech-synthesis/read-aloud implementation. Timeline027/028, original028 and shared42 therefore need capability detection, accessible per-message state, cancellation/ownership and stale-callback tests before mapping. Browser synthesis does not imply a terminal speech control. Retain terminal `/copy` and native accessibility without adding an idle toolbar; any future speech action must be explicitly invoked and separately tested.

## Browser read-aloud ownership (2026-09-24)

`gi-post-speech.ts` and the guarded `patch-post-speech.mjs` restore the speaker/stop action and Markdown-to-text rules from Piclaw `70d33bc93` (`runtime/web/src/components/post-speech.ts` and Post speech markup). Supplied Post, utilities and styles remain byte-for-byte unchanged. The action uses the existing post toolbar geometry, accessible labels and `aria-pressed`; a Gi stylesheet colours active playback. Only assistant posts with nonempty speakable content and callable browser synthesis APIs expose it. Fenced code is omitted and text is capped at1,600 characters, matching the frozen helper.

One controller owns one utterance per page. Each mounted post occurrence has an opaque owner; token checks prevent cancelled utterances' late `onend`/`onerror` callbacks from clearing a replacement, even when the same post is restarted. Ownership is cleared before calling cancel, covering synchronous callbacks. Session selection, owner unmount, source/speakable-text change, page hide and hidden-document transitions stop playback. Browser exceptions clear the active control without changing stored messages or drafts. No provider, server audio API, voice configuration or credential is added. Voice availability, audibility and any platform speech-service processing depend on the user's browser/OS.

`tests/ux/speech.spec.mjs` controls only the speech-synthesis boundary and drives native stored messages, supplied rendering, actual buttons and session switches. It verifies unavailable/available APIs, assistant-only controls, bounded text with omitted code, keyboard/pointer stop/start, transfer ordering, synchronous cancellation, old callbacks, same-owner restart, errors, pagehide/visibility cleanup and exact draft/media/session isolation. Native empty-assistant capability coverage is not yet present, so timeline027 stays unmapped. Message-copy and code-copy regressions pass independently; original028 and shared42 still need a single integrated copy/speech scenario.

The final speech/copy/rendering matrix passes72/72 across all six projects. Functional97/97, helpers91/91 with1,103 assertions and `make test vet bun-checks` pass. Early failures were a missing injected hook import and an incorrect fixture selector; the final adapter uses a local handler and native post IDs. The source review found no token-fence defect and flagged missing visibility coverage and overly broad content-change wording; visibility now has a browser check, and source content joins the cleanup dependency. Physical audio has not been tested.

Terminal disposition: retain existing `/copy`, transcript rendering and terminal/OS accessibility. Browser speech has no automatic terminal equivalent, and this slice adds no TUI command, picker or idle row. A future explicit speech integration would need its own audio backend, failure/cancellation contract and separate three-size fullscreen/regular checks.

Deployment: `557ebf7` pushed to main and restarted through `make restart BIN_DIR=/tmp/gi-scheduler-bin BIND=0.0.0.0 PORT=8090`, PID1407593. Mutation-blocked live Chromium/WebKit at390x844,820x1180,1440x900 used existing stored messages and the controlled speech boundary to verify transfer, stale callbacks, stop and draft preservation. All six reported zero API writes/page errors. Physical audio was not exercised. SQLite integrity is `ok`; before/after62sessions,51turns,146messages are unchanged. Reports, logs and captures are attached as `/workspace/tmp/gi-speech-evidence.tar.gz`.

## Integrated clipboard and speech acceptance (2026-09-24)

Timeline027 now covers the missing native empty-assistant precondition. The isolated Go UX server seeds two valid assistant records (empty and whitespace-only) through `Store.AddMessage` when `GI_UX_SPEECH=1`; production startup and storage are unchanged. The test reads them back through the production HTTP projection and verifies rendered assistant posts expose no speech control. Ordinary submitted messages prove the converse: no control without browser APIs, one control per speakable assistant with APIs, and no user-post control.

Separate original028 and shared42 executions compare trusted browser clipboard events with code extracted from the exact stored assistant Markdown, including a literal HTML-looking string and trailing newline. HTML clipboard data is empty. Each execution starts one post, transfers to another with the exact speak/cancel/speak ordering, fires the old completion/error callbacks, then copies again while the newer speech remains active. Stop, unchanged persisted messages, other-session turns, exact draft/attachment retention and reload complete the checks. The OS speech boundary is controlled; audible output is unverified.

`make test-ux-speech-contract UX_LOCAL_BIN=/tmp/gi-media-bin/gi-ux-speech` passes90/90 across six projects, including prior speech/copy/rendering regressions. Functional97/97, support91/91 with1,115 assertions and `make test vet bun-checks` pass. An initial incorrect code-copy selector timed out; final tests use the supplied `.post-code-copy-btn` and exact stored bytes. Independent review found no mapping blocker. The strict report has separate five-project partial/six-project pass checks for all three IDs; shared41 stays unmapped. Source components, criteria and production code are unchanged.

Cumulative coverage is92/236 Classic and30/42 shared, leaving144/12 unmapped. Terminal `/copy` has its own earlier native byte tests; no terminal speech feature or idle UI is inferred from browser playback. Continue using explicit copy and terminal/OS accessibility until an independently tested audio backend exists.

## Latest native tool status (2026-09-24)

`SessionActivity` now reads the selected active/latest turn and its latest tool occurrence in one read-only SQLite transaction. The start sequence and call ID pair with the first subsequent matching finish/failure, so repeated tool names cannot exchange results. Legacy shell events use a stable `turn_id:start_seq` identity; they contain one shell call per turn. A tool without a terminal event loses its running state when its claim is gone, but duration stays unknown because claim loss has no authoritative end timestamp.

Native tool-start persistence retains at most200 Unicode characters of command/path/query preview; arbitrary arguments and output are excluded from the projection. Preview text is rendered as text, never HTML. Committed lifecycle events trigger `tool_activity_changed`; the browser uses these only to refresh its selected activity through existing selection/connection/revision guards. Delayed old reads cannot replace newer occurrences. The summary uses the current status area and replaces the generic Working status, with a running spinner, one-second elapsed label, completed/failed marker and frozen end duration. Compaction hides it while active. There is no expandable output pane or tool history list.

Original027 passes across all six projects using held real shell execution, reload while running, completion, a shell exit7 failure, different call IDs, a held old activity read, session switching and exact draft/attachment retention. The store test covers equal tool names with distinct IDs, legacy IDs, bounded preview, no arbitrary-argument leak, no cross-session result, new-turn reset and claim-loss unknown duration. Targeted store/web races run three times. A review correctly identified non-authoritative interrupted timing, which was removed; its nullable-column concern did not apply because the code already used `sql.NullString`, and the unused column was removed.

Final results:60/60 tool/queue/context,96/96 compaction,48/48 thoughts,98/98 functional,92/92 helpers with1,129 assertions, Go/vet/hook and six fullscreen/regular pending-media PTYs pass. Queue shared27 initially failed because its old assertion removed `created_at` only from stored media; it now compares timestamps explicitly and other fields symmetrically, retaining the byte/identity/FIFO assertions. No supplied component or frozen feature source changed.

Coverage is93/236 Classic and30/42 shared, leaving143/12 unmapped. Shared40 requires expandable tool panes, keyboard/focus handling and full persisted-pane reconstruction, which this latest-call summary does not provide. Terminal adaptation stays within existing tool blocks: show identity/preview while active and freeze an authoritative duration on completion; do not add a persistent status panel or idle row. The current terminal renderer is unchanged; the six PTYs are regression checks, not new terminal timing acceptance.

Deployment: `1ee70f1` pushed and restarted through `make restart BIN_DIR=/tmp/gi-scheduler-bin BIND=0.0.0.0 PORT=8090`, PID1551354. Mutation-blocked live Chromium/WebKit at390x844,820x1180,1440x900 verified an existing persisted completed tool's identity, duration, absent spinner and retained draft. Six checks reported zero API writes/page errors. SQLite integrity is `ok`;62sessions,51turns,146messages unchanged. Reports, regression logs and captures are attached as `/workspace/tmp/gi-tools-evidence.tar.gz`. No live shell invocation was needed.

## Terminal tool identity and inline timing (2026-09-24)

The terminal already displays elapsed/final timing in its existing flat tool headers. Fullscreen redraws active timing; regular mode withholds mutable tool blocks and prints the final block into terminal-owned scrollback. No extra timer, toolbar, picker, status panel, idle row or colour change was needed.

`toolRuntimeBlockKey` now includes the turn ID as well as the provider call ID, with length encoding to avoid delimiter ambiguity. Providers may reuse call IDs across turns without overwriting earlier tool output. Once a retained block has an end timestamp, late start and duplicate/conflicting terminal events leave its body and timestamps unchanged. An end received without a start displays no duration instead of a fabricated zero. This assumes call IDs identify one call within a turn; events do not stream output after a terminal result. Transcript eviction removes the identity map, and restart replay has no new deduplication guarantee.

Unit tests inject cross-turn ID reuse, repeated tool names, late starts, duplicate/conflicting ends and missing-start events; targeted TUI race tests run three times. The six native PTYs at60x18,100x22,140x36 use a held real shell and exit7 failure in fullscreen and regular modes. Header durations match stored runtime-event intervals within500ms, stay frozen after completion/failure, and retain the Unicode draft/cursor across resize. Regular scrollback contains no running tool block; its dock remains five rows. The initial test compared absolute screen rows, then counted a transcript separator as an extra dock band; the final measurement isolates the current editor/footer dock while fullscreen still checks absolute rows.

`make test-tui-tool-timing BIN_DIR=/tmp/gi-media-bin`, eighteen pending-media/session/model-picker PTYs, `make test-terminal-tool-identity`, Go/vet/hook,92helpers with1,129 assertions,98functional and six browser original027 regressions pass. Bounded review found no blocker within the retained-occurrence contract. Browser mappings stay93/236 Classic and30/42 shared; shared40's expandable/reconstructed pane lifecycle is unchanged.

Deployment: `05defbb` pushed and restarted through the standard Make target on8090, PID1577123. Guarded Chromium/WebKit at three sizes verified the existing completed-tool summary/draft with zero API writes/page errors. SQLite integrity is `ok`;62sessions,51turns,146messages unchanged. Six terminal-mode/size native timing captures and logs are attached in `/workspace/tmp/gi-tui-tools-evidence.tar.gz`.

## Stored remote resource cards (2026-09-24)

Gi already passed stored `resource_link` blocks through its media projection, but discarded `link_previews`. A shared adapter now validates resource-link destinations and projects previews into the supplied Post cards in both timeline and search. It accepts at most8 entries of each kind,2048-character credential-free HTTP(S) URLs,200-character titles and1000-character descriptions. Executable/data/file/relative schemes, literal control characters, backslashes and credentials are rejected. Preview sites derive from the URL; remote images and arbitrary extra fields are omitted. No URL discovery, metadata scraper, server-side fetch or image request was added. Other existing content-block/media behaviour is unchanged and has no new general safety guarantee.

`make test-ux-links UX_LOCAL_BIN=/tmp/gi-media-bin/gi-ux-links` uses native stored records in an isolated Go server. Timeline025 exercises real pointer and keyboard navigation, HTTP and HTTPS, `_blank`, `noopener noreferrer`, null popup opener/referrer and controlled destination pages. It checks no automatic remote request, removal of unsafe resource/preview entries, retained source text and draft/media, reload, and a new popup from search results. Only destination network responses are local fixtures. No live database fixture is inserted.

Final results:60browser cases across six projects (links plus rendering/lightbox),1seeded functional case,98standard functional cases,95helpers with1,159 assertions, Go/vet/hook pass. The standard run skips the one seed-dependent case; the link target runs it separately. Functional artifacts use a separate directory so Playwright cannot delete parity JSON before reporting. Early test errors used the wrong Search accessible name and asserted the draft while the search editor was still active; final tests use native controls and close search before checking the draft. Review requested pointer, HTTP and restored-view navigation tests, now included. Its media-reordering concern described pre-existing projection semantics, not a change in this slice.

Coverage is94/236 Classic and30/42 shared, leaving142/12 unmapped. Terminal disposition: retain ordinary Markdown links and existing opt-in link/copy interactions; browser metadata cards need no terminal panel or new idle row. No automatic URL opening, fetching or browser-derived terminal acceptance is introduced.

Deployment: `eddfcbb` pushed and restarted via the standard Make target on8090, PID1618912. Mutation-blocked six-browser-size existing-message/status/draft smoke passed with zero API writes/page errors. Remote-card navigation remains the isolated fixture proof; no link records were added live. SQLite integrity is `ok`;62sessions,51turns,146messages unchanged. Reports and logs are attached as `/workspace/tmp/gi-links-evidence.tar.gz`.

## Native recovery outcome placement (2026-09-24)

Final inference and successful shell responses consult persisted `turn.recovered` events for their exact session/turn. Only `requeue_interrupted_turn` and `requeue_after_compaction_checkpoint` add a recovery marker; interrupted tool checkpoints held for review, cancellation recovery and terminal-claim release do not. Attempt count is native requeue count plus the initial execution, not provider retry count. The marker is persisted with the response and sent in its live post payload. Existing history is not backfilled, recovery policy is unchanged, and no timeout is inferred from generic failure or cancellation.

A guarded Post build adapter moves the existing timestamp ahead of recovery/timeout chips without changing supplied source bytes or inventing chip markup. Timeline026 verifies actual DOM order and same-row geometry at all six browser sizes. The fixture creates a stale compaction claim in an isolated database and lets the real startup recovery and shell completion paths produce the marker; no marked reply is seeded. A native inference test separately verifies both persisted and live post blocks. Browser checks retain exact draft/media through reload/search and verify ordinary completion has no inherited marker.

Final checks:54browser cases (outcome, speech and copy),1seeded functional,98standard functional,96helpers/1,170 assertions, Go/vet/hook, targeted recovery race x3 and six terminal timing regressions pass. The standard functional run skips the2seed-dependent cases, each covered by its separate Make target. Review requested live-event proof, now covered; the shell path was already the browser fixture, while the native test covers inference. Exact fixture block equality is a regression assertion, not a general restriction against future response content blocks.

Coverage is95/236 Classic and30/42 shared, leaving141/12 unmapped. A future terminal recovery indication should reuse the existing final response/system status line and its Pi colours without a badge row or persistent panel. This slice adds no terminal display; the six timing PTYs are regressions only.

Deployment: `6ff9f72` pushed and restarted through the standard Make target on8090, PID1647034. Mutation-blocked Chromium/WebKit at three sizes verified existing-message/status/draft rendering with zero API writes/page errors. Native recovered-chip geometry remains isolated-fixture evidence; no recovery was induced live. SQLite integrity is `ok`;62sessions,51turns,146messages unchanged. Reports and logs are attached in `/workspace/tmp/gi-outcomes-evidence.tar.gz`.

## Uncached Settings shell acceptance (2026-09-24)

Settings002 now has an independent test of the unchanged loading behaviour. A real200 runtime snapshot is held before delivery; the dialog, header, navigation and General loading state must appear within1s without values. Releasing that native snapshot fills General with the instance's actual names, workspace and model, with no section switch. After a fresh document clears the General cache, a separate503 transport-failure fixture verifies an explicit Retry recovers the native values. Navigation is not required to be disabled by this frozen scenario and its behaviour is unchanged.

All six projects retain the exact Unicode draft and attachment bytes from IndexedDB, unchanged session record and empty turn list, with no settings/session mutation request. The complete Settings-shell/Gi Settings run passes216/216;98standard functional cases,96helpers with1,175 assertions and Go/vet/hook pass. The2seed-dependent functional cases remain skipped in the standard runner and were not counted here. Early fixture errors referenced incorrect helper names and the wrong persisted draft shape; final checks use the existing helpers and record format. Independent review found no criterion blocker.

Coverage is96/236 Classic and30/42 shared, leaving140/12 unmapped. This test-only slice changes no runtime file and needs no restart; live PID1647034 retains62sessions,51turns,146messages and SQLite integrity `ok`. Terminal settings/selector behaviour has its own earlier acceptance; no browser loading-shell result grants a new terminal control or idle row. Reports and logs are attached as `/workspace/tmp/gi-settings-loading-evidence.tar.gz`.
