# gi implementation checklist

Status: Active
Date: 2026-04-22
Last updated: 2026-09-25

- [x] Deploy contrasta25140f afterCI36197537050 (ninejobs); separate-session Makefile launch PID304534 PGID/SID304361 on8090. Six guarded Chromium/WebKit probes pass contrast/padding and prior panel behaviours,zero writes/errors;DB62/51/146+integrity/FKs and normalizedSQL unchanged except leases,auth absent. Prior process shutdown cause unproven; new launch alive across subsequent tools, restartdurability not claimed. Raster diagnostic flags still leave12/36unstable and were not adopted; official16/36remains.

- [x] Deploy compose/model/session panel stack through8f97f08 after CI36193525146 succeeds (nine jobs, release skipped), including authlistener9cec426. Port8090 restored PID202980. Six read-only Chromium/WebKit×viewport probes pass resize, Models-settings handoff, session metadata/search and draft/focus retention; zero HTTP writes/page errors. DB62/51/146 integrity/FKs OK, normalizedSQL unchanged except runtimeleases SHA256300ce679…653d; authfile absent before/after. Pixel/physical/Visual and unsupported capability gaps remain open.

- [x] Repair isolated journey-server reserve/release port race exposed by CI36162988558 (downloaded fixture log: bind address already in use). Native fixture binds port0 and retains listener, publishing its address in a private readiness file. No timeout increase;42startup/72slash/24geometry and Go/vet pass. Workspace0feba2c and port repair8ed87ca deployed after CI36164916153 passed; six read-only Chromium/WebKit transition probes retained drafts/session/focus with no writes or page errors. This deployment record predates the current host upgrade; it is not a new live acceptance run.
- [x] Add pinned compose/panel capture and exact-RGBA reporting harness (`make pixel-baseline`, `make pixel-compare`), with fail-closed helper tests. Full phone/tablet/desktop × light/dark × compose/models/sessions matrix captured 72 images; all 18 cross-host comparisons fail and 13/36 same-host pairs are unstable in the final run (20 in the preceding run). No masks, tolerance waiver, feature-mapping or Visual credit. See [pixel baseline](../internal/compose-pixel-baseline.md). Host upgrade removed installed 3.2.2 assets; an extracted published reference matches all eight pinned hashes.
- [x] Implement/test host-only compose surface: reference top session pill and resize separator, exact six-case compose/textarea bounds, persisted bounded height and modal/cancel cleanup. 18surface/24geometry/42startup/72slash/56phone-drafts/162model-session pass; Go/vet/hooks/helpers and107functional (11explicit skips) pass. Pixel gate still fails:18/18 crossframes,13/36 repeats; compose region1,514–2,438px. Earlier session-search active-row failure remains unresolved despite five isolated repeats and full rerun passing. See [surface evidence](../internal/compose-surface.md). 75da410 passed CI36180760610; deployment held while the next panel slice is unverified. No mapping/Visual/TUI credit.
- [x] Adapt/test supported native model-panel structure: search/count header, current-first metadata rows, responsive footer and Models-settings handoff. 12panel/24geometry/174models-quick-actions/216settings/36context-fit, Go/vet/hooks/helpers and107functional (11skips) pass. 72captures complete; all18 crossframes fail,9/36 repeats unstable. Panel bounds match all six viewport/theme cases; explicit pin/pricing/thinking mutation/accessibility/mobile-close differences remain. See [model-panel evidence](../internal/model-panel.md). Await CI/deployment; no mapping/Visual/TUI credit.
- [x] Adapt/test session-panel header, compact metadata/badges and leading native pins. Confirmed query-highlight effect race reproduced3/3, fixed before paint;12new+126session+24geometry+72slash+107functional(11skips) pass. All six session bounds match;72captures complete but18crossframes and15/36repeats fail. Preserve native rename/archive/restore and explicit unsupported reference actions. Archive-followup GET stall unresolved despite5isolated/full passes. See [session panel](../internal/session-panel.md); await CI/deploy, no mappings/Visual/TUI credit.
- [x] Remove auth-fixture reserve/release handoff via native owned listener and private readiness origin, preserving restart origin and10sbudget. 9cec426 passes189passkey/120auth/42startup andCI36189360384/36190769031. OriginalCI36186190456 readiness cause unproven (server log absent); CI now retains passkey diagnostics. See [auth fixture](../internal/auth-fixture-listener.md). No live auth/RP mutation.
- [x] Implement/test pinned maximum-contrast foreground and0/8px session-pill padding via guarded host adapter. Four helpers4126assertions,30compose/216Settings/107functional(11skips),Go/vet/hooks/14pixelhelpers pass. Final72captures complete;compose diff930–1735px,all18crossframes fail,16/36repeats unstable. Missing voice-input/notifications remain explicit capability gaps (56px footer difference), no placeholders/masks. See [contrast evidence](../internal/compose-contrast.md); awaitCI/deploy, no Visual/mapping/TUI credit.
- [ ] Raise composer/panel acceptance to user-requested pixel parity: matched-state pinned Piclaw reference renders, typography/assets/controls together, repeated exact RGBA comparisons and preserved diff evidence. Existing outer-bound tests alone do not satisfy this requirement.

- [x] Add read-only workspace Return/Show controls retaining mounted preview, tabs and exact draft/media; hidden panes suppress shortcuts and clear contextmenus. Roving Arrow/Home/End tab focus and ShiftF10 contextmenu restorefocus; WebKit touchclose fixed via pointerdefault preservation without backgroundactivation.72tab/54preview-shell/42startup/24geometry/72slash/113functional (fiveexisting skips), Go-vet-hook/113helpers3020 pass. Initial12transition and3WebKit failures reproduced; review-hiddenmenu fix verified. Firstfunctional run sessiontypeahead classflake retained (unchanged rerun green, not fixed). RequiredCI targetadded; suppliedsource/storeunchanged, no new mappings/editor/docking/TUIrows.

- [x] Deploy `cec9147` native slash catalogue/key ownership after CI36157033453: nine jobs green, required browser gate includes72slash+42startup+24geometry plus189passkey separately. Port8090 PID3756871; six guarded live checks show29native commands/skills, Tab/Escape draft retention, palette/Settings exclusion, zero writes/obsoleteendpointcalls/errors. DB62/51/146, integrity/FKs OK, fullSQL equal except runtimelease (SHA256300ce679…53d); authfile absentbefore/after. No live command submit, enrolment or TUI change.

- [x] Replace direct /agent/commands+56-command fallback with validated native catalogue through guarded build adapter; clear stale matches, ignore retired effects, suppress delayed completion behind Settings/current search, retain separate failure notice. Repeated/consumed/legacyIME Enter red18 cases fixed; modern composing already safe.72slash/174QuickActions-model/18skill/42startup/24geometry/112functional (fiveexisting skips), Go-vet-hook and112helpers/3,013assertions pass. Review fixes verified. Required browserCI includes slash; suppliedsource unchanged. Synthetic IME bounded, modifier semantics/shared16/Classic008 conflicts remain separate; no new mappings or TUIrows.

- [x] Deploy `d6f447c` pinned picker geometry after CI36150466902: nine jobs green, including24geometry+42startup and189passkey cases. Port8090 PID3667793;20 guarded live checks (two engines × five widths × two pickers) passed bounds/focus/draft/no-write assertions. DB62/51/146, integrity/FKs OK, fullSQL equal except runtimelease (SHA256300ce679…53d); auth file absent before/after. Reference640px modeloverflow corrected by anchorcap, other structural/visual differences explicit. No production enrolment or TUI change.

- [x] Pin Piclaw Classic0afe5366 CSS (installed3.2.2, liveasset15958f2c3dc9) and read-only measurements at390/639/640/820/1440. Host CSS fixes composer side padding/session width/model680px/fixed mobile8px panels; explicit anchor cap corrects reference640px overflow. Guarded adapter adds mobile Close only; supplied bytes unchanged.24geometry/42startup/126session/36model/30shell/6filter/111functional (five existing skips)/109helpers2988/Go-vet-hook/review pass. Old mobileclick-through tests now dismiss first, stale-response assertions retained; initial combined regression timeout excluded. Geometry runs in required browserCI; no new mappings/full structuralvisual acceptance/TUIrows.

- [x] Deploy `a2863b9` startup/new-chat focus and retry repair after CI36143979120: nine jobs passed, including the new42-case Chromium/WebKit journey gate, existing passkey gate and four Linux/macOS builds. Port8090 PID3593567; six guarded existing-session live checks verified startup/reload focus, exact Shift+Enter, retained drafts and Settings isolation with zero writes/errors. Initial live-probe stale main-ID assumption corrected using read-only session discovery. Auth file absent before/after; DB62/51/146, integrity/FKs OK, full SQL equal except runtime lease (SHA256300ce679…53d). No live prompt, enrolment, supplied-component edit or TUI rows.

- [x] Repair reproduced startup/new-chat lost focus and initial-session failure stuck at Loading in the Gi host, without editing supplied components. Add explicit loading retry, cancelled-effect fencing and one-shot post-commit focus that respects other controls/modals. Seven empty-store journeys cover captured session/turn IDs, rendered messages, Return/Shift+Enter, held/failed admission, lost create reply, partial bootstrap failure, child/parent draft separation and delayed-fork Settings focus. 42 journey/126 session/216 Settings/168 draft/110 functional cases (five existing skips), full Go/vet/hooks and107helpers/2,975 assertions pass. Final review clears stale draft-load callback fix. Combined regression timed out and a follow-up was interrupted; completed split runs supply acceptance. Required Chromium/WebKit CI job added before Linux/macOS builds; no formal mappings, visual parity or TUI changes claimed.

- [x] Deploy `020919a` first-owner Settings setup after CI36134257925: eight Linux/macOS jobs passed. Port8090 PID3454805; six guarded Chromium/WebKit checks at390/820/1440 verified the setup entry point, no secret/code or sign-out control, disabled passkey creation, retained drafts and zero writes/errors. Setup was never started on the live instance. Auth file absent before/after; database62 sessions/51 turns/146 messages, integrity/FKs OK and full SQL unchanged except runtime lease (SHA256300ce679…53d). Production RP remains undecided; no TUI change.

- [x] Deploy `af4ab28` bootstrap API prerequisite after CI36128019772: eight Linux/macOS jobs passed; port8090 PID3352646. Six guarded Chromium/WebKit live checks passed with no writes/errors and retained drafts. Auth file absent before/after; live still unenrolled/passkeys disabled. DB62 sessions/51 turns/146 messages, integrity/FKs OK, complete SQL unchanged except runtime lease (SHA256300ce679…53d). No live setup call, production RP decision or TUI change.

- [x] Add browser-bound bootstrap backend prerequisite: loopback Host/peer plus exact Origin, ten-minute hashed setup bindings (maximum eight), single-use finish, atomic default-owner/session creation, cancellation, stale and concurrent browser/legacy rejection. Legacy API unchanged. Native auth and race×3, 54 Chromium/WebKit auth cases, 186 passkey cases, 109 functional cases (five existing skips), full Go/vet/hooks and 106 helpers/2,958 assertions pass. Focused review found no blocker after an initial delegate timeout. No Settings setup controls, new formal mappings, production enrolment or TUI rows.

- [x] Add first-owner Settings controls using the browser-bound API and native setup_available status: manual TOTP key held only in mounted component, conservative ten-minute display lifetime, immediate secret/code clearing on finish/cancel, native status confirmation and explicit recovery for lost/false/failed replies. Ready cancel revokes pending binding even during parent refresh; active cancel detaches immediately (WebKit abort race fixed); pane exits send no background writes. Shared-tab replacement/owner-created-elsewhere, draft/pill retention and first usable CDP passkey/TOTP login verified. 120 auth/189 passkey/216 Settings/110 functional cases (five existing skips), Go/vet/hooks/auth race×3 and107helpers/2,975assertions passed. Focused rereview cleared prior cancel bug. No production enrolment/RP decision or terminal rows; Visual/physical/QR checks not established.

- [x] Verify real HTTPgi-insecure.test mappedonlytoloopback with ChromiumsecureContextfalse, validcopiedownercookie refused, no authforms/Settings/credentialcalls/UIwrites; native negativeAPI401/403/fullauthstateequal thenlocalhostsametoken/keypositivecontrol.186passkey/108functional (five existing skips)/106helpers2958/Go-vet-nativeauth/hook/finalreview pass.015remains partial: secure guidance appears at authgate, not AddwithinSettings. No trust/propertyoverride or runtimechange/deployment/liveauth/TUI rows; initial explorationdelegate timedout, boundedreview passed.

- [x] Deploy556299c browser logout after CI36120572780: eight Linux/macOS test/build jobs green including183passkey+ledger.8090 PID3263425; six guarded live checks show unenrolled setup/no sign-out control, zero writes/errors and retained drafts. DB62/51/146, integrity/FKs OK, complete SQL equal except runtime lease (SHA256300ce679…53d). Production credentials/RP/policy unchanged; no live logout/TUI rows.

- [x] Add cookie-only explicit browser logout with exact Origin/transport/empty-JSON guards and transactional purpose/token/expiry validation, no fresh-proof requirement or other-session/factor mutation. Settings confirms POST plus native status before gate recheck; lost/false/failed replies offer read-only Check sign-in status, no replay.183passkey/42auth/216Settings/108functional (five existing skips)/106helpers2943/Go-vet-hook/web-auth race×3/review pass. Post-removal logout and accelerated2s real-clock expiry retain other browser, drafts/pills/selection; media byte equality and12hwait not claimed.024bounded with expiry substitution; no live logout or TUI rows.

- [x] Verify Settings duplicate path with real first/second creation and native exclusion list; crafted original none-attestation+fresh client-data finish returns exact post-verification409 without server collision seed. Original full credential/sessions unchanged, no success/second row/replay through refresh/re-entry, original key signs in.168passkey/107functional (five existing skips)/106helpers2918/Go-vet-nativeauth/hook/review pass.012bounded candidate, crafted-client substitution explicit/no physical duplicate or attestation-forgery/full mapping credit. Test/docs only/no deployment/live auth/TUI rows.

- [x] Deploycfef4d8 removal explanations after CI36113841805: eight Linux/macOS test/build jobs green (no Windows).8090 PID3145609; six guarded live browser/size checks zero writes/errors/draft loss. Live still unenrolled/passkeys disabled. DB62/51/146, integrity/FKs OK and full SQL equal except runtime lease (SHA256300ce679…53d). No production enrolment/policy change or TUI rows.

- [x] Derive removal reason/remaining_method under the atomic writer lock; preserve legacy error text409/errors.Is/wrapper and return no success result on failure. Settings renders fixed known reasons and post-confirmation fallback, clears details on next work, tolerates absent optional fields. Seven native/browser cases (six supported outline plus defensive disabled secret),165passkey/36auth/216Settings/107functional (five existing skips)/106helpers2911/Go-vet-hook/auth+web race×3/review pass.020pending-ownerTOTP still unsupported; seeded disabled secret is not that flow. No live auth/TUI rows.

- [x] Deploy07692d9 inventory-query refusal after Linux/macOS-only CI36110032772: eight test/build jobs green including144browser/ledger, no Windows jobs.8090 PID3062406, six guarded live Chromium/WebKit checks zero writes/errors/drafts retained. Still unenrolled/passkeys disabled; no production auth writes. DB62sessions/51turns/146messages, integrity/FKs OK, full SQL equal except runtime lease (SHA256300ce679…53d). Prior server-readiness failure preserved in evidence; runtime test timeout unchanged.

- [x] Reject unexpected passkey inventory queries instead of silently ignoring account selectors (four native200→400 regressions). Ten native authority cases cover missing-owner list, automation register, expired/revoked rename, foreign-origin remove and selector variants; complete auth-file equality/no inventory/cookie disclosure and owner-list control. Real-key browser query/refetch proof,144passkey/106functional (five existing skips)/106helpers2885/Go-vet-hook/auth+web race×3/review pass.017firstfive examples bounded; family mode still unsupported/fullMappingfalse. No live auth changes or TUI rows. Prior Linux/macOS-only CI36108460455 failed disposable server readiness before assertions; cause unestablished, no timeout changes.

- [x] Resolve auth test contention retry exhaustion as59db473 after failed CI36106302597: preserve25 writers and all token/revoke assertions; test-only10s elapsed budget/capped backoff, error propagation/exhaustion tests, no post-deadline retry. Auth race×10/Go-vet/review and native CI36107268256 green. Runtime locking unchanged; no deployment.

- [x] Remove Windows auth-state job and build/release artifact target from CI at the owner's request; retain Linux/macOS native auth and four Linux/macOS architecture builds. Remove obsolete .exe/ZIP workflow branches. Windows runtime code and optional local cross-build target remain. Condense the session sidebar to outstanding work; repository completion evidence retained.

- [x] Verify lost successful registration after pane switch and Settings close/reopen: automatic held native inventory GET, no invented row/empty state before release, new credential once, no auth POST replay through second return/full reload, exact committed auth state retained. Draft text/attachment pill survive (not byte equality).141 passkey/105 functional (five existing skips)/106 helpers2864/Go-vet-nativeauth/review pass;014explicit-return gap now bounded in criterion ledger, fullMappingfalse/global fixture limits retained. Test/docs only/no deployment/live auth/TUI rows.

- [x] Add validated passkey criterion ledger:26 scenarios/169 steps/4 Background steps/40 example rows/56 expanded cases, exact pinned source+named test declarations, bounded/substituted/partial/gap/manual/unsupported dispositions and zero-row browser-only terminal adaptation. No full-scenario mappings awarded. Correct014 Refresh≠return and020generic-reason gaps;017matrix/021lockcontention/024lifecycle explicit.106 helpers2854/Go-vet-build-hook/105functional (five existing skips) pass. Initial two stale anchor names corrected; review tightened declaration uniqueness,236IDs versus256expanded and disputed008 membership. Source semantics remain human-reviewed, not validator proof.

- [x] Verify standalone first-key enrolment waits for native finish, retains fresh TOTP sign-in and unchanged chat/no navigation-secret leakage; exactly-one-key passkey-only/no-TOTP fixture adds second distinct key with original record unchanged; confirmed removal rejects real removed-key assertion400/no cookie/protected401, surviving key signs in200 with identity retained.135 passkey/105 functional (five existing skips)/104 helpers1334/Go-vet-nativeauth/review pass.002/003/008 bounded candidates, no formal mapping/physical selection/TOTP-removal UI credit. Test/docs only/no deployment/live auth/TUI rows.

- [x] Verify seven Settings native registration failures: expiry/replay/foreign-session/altered origin-RP/invalid proof/revoked-after-create-before-finish. Native error/unchanged stored keys/retained-key fresh login; actual secret/challenge/proof canaries absent DOM/storage/URL/error. Foreign proof rejected in wrong browser then accepted unchanged in its own session.126 passkey/105 functional (five existing skips)/104 helpers1334/Go-vet-nativeauth/final review pass. Initial error-string/omitted-array/CDP pending-presence fixture failures corrected; foreign-proof weakness found by review tightened with positive control.018 stays partial: no physical prompt-open revocation or isolated origin/RP-check ordering. Test/docs only/no mappings/TUI/live auth changes.

- [x] Verify unavailable WebAuthn constructor/container/create/get and native TOTP-only policy leave controls explained/disabled with zero credential calls/auth writes and full auth-state equality across refresh/re-entry. Restored same wrapped APIs/policy permit real assertion+new enrolment as positive controls.105 passkey/105 functional (five existing skips)/104 helpers1334/Go-vet-nativeauth/review pass. Test/docs only;015 remains partial for actual insecure-host, capability overrides do not prove old-browser support; no new mapping/TUI rows/live auth change.

- [x] Verify Classic passkey IDs/dates/Never used against native records and exclude actual cookie/TOTP/user-handle/public-key/token-hash canaries from DOM/storage/URL/inventory. At390px verify80-code-point name, metadata/actions/long error horizontal bounds, labels/accessible names, Tab/Enter and trusted-tap rename/remove/cancel with zero writes; role status/alert and outside-focus guard.90 passkey executions (six explicitly repeat390px)/105 functional (five existing skips)/104 helpers1334/Go-vet-nativeauth/review pass. Initial visible-label locator mismatch corrected by checking both accessible name and input.labels. Test-only;001/016 stay partial for Visual/physical speech, no mappings/TUI rows/live auth changes.

- [x] Deploy verification-focus repair a9d49a6 after CI36093456846: Test,81-case passkey browser, three native OS auth jobs and five builds green. Restart8090 PID2772432; six guarded Chromium/WebKit×three-size checks retain drafts with zero API writes/errors. Live remains unenrolled/passkeys disabled. SQLite integrity/FKs and62sessions/51turns/146messages retained; complete SQL dump equal except runtime lease (normalised SHA256300ce679…53d). No production enrolment/policy change or TUI rows.

- [x] Verify passkey list/rename/remove failure truth and explicit native recovery for503/pre-delivery network errors; stale add/rename/remove remain unauthorised after button/Escape reauth cancellation and succeed only with new proof. Repair replaced verification-control focus using pane-local stable action IDs, retaining outside-focus/unmount guards. Nine focus failures red→green;81 passkey/36 auth/216 Settings/105 functional (five existing skips)/104 helpers1334/Go-vet-nativeauth/review pass. Whole-second fixture timestamps avoid equivalent Go RFC3339 formatting drift. No new frozen mapping, physical/native focus evidence or terminal rows; live auth unchanged.

- [x] Verify two Settings views concurrently removing the last two passkeys under passkey-only policy: hold both native requests, one200/one409, reconcile both inventories, no automatic retry, retain sessions/cookies and prove fresh sign-in with the surviving credential. Explicit retry returns last-factor409. The initial three failures exposed non-blocking writer-lock conflict before the factor check; scenario021 stays partial for that distinction. 54 passkey browser/104 functional (five existing skips)/104 helpers1334 assertions/Go-vet-nativeauth/review pass. Disposable stores; test/docs only, no frozen mapping, runtime change, restart or terminal controls.

- [x] Verify isolated Settings passkey unnamed-credential rename/reload, blank/whitespace/overlong/control/literal-HTML names with subsequent sign-in, cancellation with zero requests/focus return and a positive control, and stale second-browser add/rename/remove after first-browser reauthentication. 51 passkey and36 auth browser cases,104 functional (five existing skips),104 helpers/1334 assertions, Go/vet/native auth and criterion review pass. Test/documentation-only: no new frozen mapping, runtime deployment, live authentication changes or terminal rows. Canonical empty-name fixture repaired after three initial failures; interrupted rerun excluded and final51 passed. Functional build used an isolated /tmp binary after the live bin/gi symlink returned text-file-busy.

- [x] Decorate intact parenthesised HTTP(S) terminal URLs with safe OSC8 metadata, preserve targets through search, and give linked tool-body clicks precedence over block toggling. Visible text/drag-copy unchanged; credentials/control targets rejected, terminal-client activation only.6link PTYs+9selection/search/regular regressions/98standardfunctional/96helpers1175assertions/Go-vet-hook/TUI race x3 pass; no idle rows or browser mapping. Wrapped/complex URL spans, word selection and right-edge selection semantics remain separate. Older PTY scripts now accept isolated GI_TUI_BIN instead of building through the live binary symlink.

- [x] Verify Settings002 independently: uncached shell/header/navigation/General before held native snapshot, native values after release,503Retry recovery and exact persisted draft/media/session/no-write protection.216Settings browser/98standard functional (2seed-dependent cases skipped)/96helpers1175assertions/Go-vet-hook/review pass. Coverage96/236+30/42 unmapped140/12; test-only/no restart or terminal/pane-behaviour credit.

- [x] Persist recovered final-response marker from exact session/turn native stale-claim requeue events; successful shell and inference share stored/live blocks, ordinary/held/terminal-release recovery excluded. Guarded Post adapter places existing chips after timestamp. Timeline026 six-project geometry/reload/search/draft-media proof:54browser+1seededfunctional/98standardfunctional/96helpers1170assertions/Go-vet-hook/recovery race x3/sixTUI timing regressions pass. Coverage95/236+30/42 unmapped141/12; timeout semantics and terminal outcome display separate.

- [x] Project bounded stored HTTP(S) resource-link and text-only preview metadata through supplied Post cards; reject credentials/executable schemes, omit remote images, retain existing media/card handling. Timeline025 actual pointer/keyboard popup opener/referrer and reload/search/draft-media proof passes6projects;60browser +1seeded functional/98standard functional (seed case separately skipped)/95helpers1159assertions/Go-vet-hook pass. Coverage94/236+30/42 unmapped142/12; no URL discovery/fetch/general block safety or terminal credit.

- [x] Scope terminal tool-call blocks by turn+call ID and make retained terminal occurrences immutable against late starts and duplicate/conflicting ends. Missing-start timing stays unknown. Six fullscreen/regular timing PTYs match native event timestamps, freeze success/failure duration and retain draft/cursor through resize with zero idle growth;18media/session/model PTY regressions/98functional/6browser/92helpers1129assertions/Go-vet-hook/TUI race x3 pass. Existing headers/colours/layout unchanged; no eviction/restart deduplication or new browser mapping.

- [x] Project latest native tool occurrence from a consistent SQLite snapshot: call identity, bounded command/path/query preview, start/end and fixed terminal duration. Run-owned SSE invalidation refreshes the existing status area; interrupted calls retain unknown timing. Original027 passes6projects including held old reads, reload/session/draft-media, native shell completion/failure.60tool-queue-context+96compaction+48thought browser/98functional/92helpers1129assertions/Go-vet-hook/targeted race x3/sixTUI media PTYs pass. Coverage93/236+30/42 unmapped143/12; shared40 expandable lifecycle remains open, no terminal UI change.

- [x] Verify timeline027 native empty/whitespace assistant and API capability gates, plus independent original028/shared42 integrated code-copy and playback ownership. Real stored code bytes/trusted clipboard events, synchronous cancel/late callbacks, draft/media/reload/other-session protection;90browser/97functional/91helpers1115assertions/Go-vet-hook/review pass. Test-only isolated DB seed; coverage92/236+30/42 unmapped144/12. No physical-audio or terminal change.

- [x] Restore frozen Piclaw read-aloud via guarded Post adapter, <=1600-character Markdown-derived text, API capability checks and per-mount/session ownership. Timeline028 transfers/cancels playback with stale same-post callbacks fenced; error/pagehide/visibility/session draft-media safety.72browser/97functional/91helpers1103assertions/Go-vet-hook pass. Coverage90/236+29/42 unmapped146/13; browser OS speech boundary controlled in tests, no physical audio proof. Timeline027 empty-assistant gate and original028/shared42 integrated copy remain uncredited; no terminal controls or idle rows.

- [x] Verify Classic original008 independently against real loaded skills, description filtering, Slash commands-only grouping and shared composer-prefill path; separate Classic/shared17 executions preserve strict reporter semantics.18browser/96functional/87helpers1069assertions/Go-vet-hook pass; coverage89/236+29/42, unmapped147/13. Test-only; terminal load-only behaviour unchanged. Shared35 lacks positive local-estimate labelling proof; read-aloud cases lack implementation and stay unmapped.

- [x] Expose startup-loaded canonical /skill commands in browser Slash commands, preserving draft/media on insertion and expanding only the captured session's invocation. Rooted UTF-8 reads <=100KiB, startup hash checks, non-blocking FIFO rejection and unknown/stale recovery; shared17 passes six projects. Final12 combined browser cases (formerly18 split), session126/QuickActions138 regressions, functional96/helpers87 with1062 assertions, Go/vet/hook and targeted race x3 pass; bounded source review found no blocker. Coverage88/236+29/42, unmapped148/13. Supplied components unchanged; terminal /skill remains load-only, no terminal credit.

- [x] Align read-only Markdown tab code with configured mono font; verify glyph widths/native Appearance/reload/draft-media across6projects. Repair Safari swipe interception of horizontal table scrollers incl edge wheel;210browser+30swipe/95functional clean rerun/86helpers1047assertions/Go-vet-hook pass. One unchanged session-typeahead timing failure recorded. Shell009 stays unmapped (no editable editor); shared41 conflicts with Classic SVG source-only; coverage88/236+28/42 unchanged, no TUI rows/credit.

- [x] Reuse byte-identical multipart browser uploads within session/name/MIME under SQLite write transaction; native/JSON still create (reserved web hash stripped). Map shared39 attach-file cancel/retry/source unlink/reload same-ID/one-media delivery across six projects; supplementary DOM drop/paste not physical clipboard.198browser/94functional/85helpers1039assertions/Go-vet-hook/store-web race×3/twoStore16writers/four-process race/review pass. Coverage88/236+28/42 unmapped148/14; no general message idempotency, orphan GC, legacy upload migration or new TUI credit.

- [x] Add session-owned Cancel uploads for captured batches: native XHR abort/pre-dispatch fencing, exact IndexedDB draft/media recovery and explicit retry, other-session/newer-draft/send isolation.30 final focused/204 regression browser/94 functional/85 helpers1034assertions/Go-vet-hook; bounded review no blocker. Supplied composer unchanged via guarded build adapter. Shared39 remains unmapped: stored orphan cleanup and full source-removal/no-duplication contract not established; no new terminal credit.

- [x] Stage ≤6 terminal media refs per source session; /attachments and /detach expose pending-only review/removal, native admission claims recover rejection and hold uncertain DB reads. Native pre/post-INSERT fault tests prevent automatic duplicate attachment; no-model/directed sends preserve draft; regular-mode rejection clears thinking. Six fullscreen/regular PTYs preserve A/B Unicode cursor/multiline/resize/settings/session refs with exact stored bytes and zero idle rows; existing TUI/race×3/Go-vet-hook/93functional/82helpers pass. Process restart, queue-draft media restoration and cross-process idempotency remain separate; browser coverage88/236+27/42 unchanged.

- [x] Use supplied tab-store MRU/pin semantics in the read-only host; map workspace-011 with pinned-before-MRU and bulk-close/draft/media proof across six projects. Guarded menu adapter hides unsupported CSV popout, bounds touch targets and registers Escape before paint; Settings owns Gi tab shortcuts, browser shortcuts stay native. 186 browser/93 functional/82 helpers1004 assertions/Go-vet-hook pass; coverage88/236+27/42, unmapped148/15. No editing/dirty/pin persistence or terminal credit.

- [x] Pin and verify the [Piclaw read-only pane host subset](../internal/pane-host-subset.md): bounded context, capability rejection, resize/close callbacks, disposal and stale-response fences. Piclaw bfc34e4eb source hashes unchanged; 24 focused/174 regression browser cases, 92 functional, 80 support tests/963 assertions and Go/vet/hook checks pass. No new frozen mapping or terminal credit; editing/docks/transfers stay unsupported.

This checklist is organized by **subsystem** and grouped by **phase**.

---

## Piclaw Classic web parity — 2026-09-21

- [x] Alt-S reuses temporary regular-mode selector screen; resize clears only its visible surface, acceptance closes only after successful captured-generation switch, failures remain retryable.6session+6model PTYs/existingTUI/92functional/75helpers/Go-vet-build-hook/race×3/review pass; A/B Unicode drafts/cursor/multiline/history/exit and≤6rows/no idle growth. Push4797dad/restart8090 read-only6browser-size smoke/62sessions/integrityOK/zero writes-errors; captures attached. Browser87/236+27/42 unchanged.

- [x] Mount native read-only workspace tabs with bounded20K preview/late-read fence/disposal/native404 Retry; truthful guarded labels, touch close targets and narrow layout. Shell007 close does not activate background tabs; exact draft/media/refs, active/last/keyboard close and Settings focus race pass:18focused/168regression/92functional/75helpers/Go-vet-build-hook/bounded review. Pushf0db954/restart8090 guarded6size live existing-file preview/close/draft/zero API writes-errors/62sessions/integrityOK; screenshots attached. Classic87/236+shared27/42 unmapped149/15; no editor/dirty/pin/popout/TUI credit.

- [x] Verify/map Settings001 native menu/shortcut/header/navigation/General-first and dialog002 cached reopen<1s while real refresh held; exact draft/media/no writes/reload.12focused/234regression/91functional/72helpers/Go-vet-build-hook; criterion review with explicit header/nav assertions rerun. Classic86/236+shared27/42 unmapped150/15. Test-only/no restart/live62sessions/integrityOK; no pane functionality/offline/TUI cache credit.

- [x] Gi standalone scale visibility fallback shares navigator/display-mode capability with supplied viewport writer; native control/storage/cross-tab/reload/draft proof, unsupported-query cleanup.168browser/91functional/72helpers/Go-vet-build-hook/review pass. Pusha3f6c8a/restart8090 read-only6browser-size capability smoke/zero API writes-errors/62sessions/integrityOK. Frozen shell008 remains unmapped: wording lacks standalone visibility precondition; no physical-PWA/TUI scale credit.

- [x] Verify/map Classic shell002/003/005 independently: native hidden-file setting/event/reload, disabled workspace actions while hidden, composer available-width geometry across drawer/sidebar/resize/reload. Exact draft/media retention;18focused/162regression/90functional/68helpers/Go-vet-build-hook pass; reviewed. Classic84/236+shared27/42, unmapped152/15. Test-only, no restart; live61sessions/integrityOK. No URL chat-only, mobile-safe-area, scale or new terminal credit.

- [x] Terminal Alt-M: metadata-aware search, enabled-only navigation, explicit unavailable retry and generation-owned native selection in the existing ≤6-result selector. Temporary regular-mode screen prevents resize fragments entering history; Escape restores multiline draft/cursor and unchanged idle rows. Six fullscreen/regular PTYs, Go/vet/build-hook/race×3, existing TUI suites,90functional/68helpers pass. Push5ae2196/restart8090 read-only six-size browser smoke,61sessions/integrityOK/zero writes-errors; terminal captures attached. Web coverage unchanged81/236+27/42.

- [x] Fence stale terminal SSE frames with native turn IDs: captured Stop/reconnect, duplicate old idle/completion replay with held authority reads, newer Stop/preview retention and stale Stop409. Reconnect66/functional90/helpers68/Go-vet-build-hook/race×3/three-size TUI regressions pass. Push73613d4/restart8090: six-browser-size read-only shell/model/focus/draft smoke, zero writes/errors,61sessions/integrityOK. Native stale-frame proof stays isolated; no idle UI or frozen credit.
- [ ] Shared36 remains unmapped: engine cleanup automatically starts the next queued turn after cancellation, so the queue is not preserved unchanged. Resolve that policy explicitly before claiming the full contract. Shared35 capability/estimate evidence remains open.

- [x] Shared34: filtered native model search, enabled-only Arrow/Home/End/Page navigation and prefix-first typeahead through a guarded adapter. Native Enter/Space once, native editing, cancellation and exact durable draft/media/reference checks pass: models42/regressions294/functional89/helpers67/Go-vet-build-hook. Classic81/236 + shared27/42, unmapped155/15. Push9b745d1/restart8090: live six-browser-size filter/editing/navigation/Escape/draft checks, zero writes/errors, 61sessions/integrityOK. Supplied sources unchanged; terminal Alt-M acceptance independent.

- [x] Shared33 session-picker non-search typeahead: strict prefix-first enabled match/full-index mapping, actual entry focus for single Enter, native ID/name/model search, Arrow/Home/End/Page navigation and Escape/draft recovery. Sessions126/regression204/functional88/helpers64/Go-vet-build-hook; Classic81/236+shared26/42 unmapped155/16. Pushb39b182/restart8090 live6browser-size native prefix focus/Escape/draft/zero writes-errors/61sessions/integrityOK. Supplied bytes unchanged; terminal Alt-S separate.
- [ ] Resolve frozen command-prefill conflict before shared16 credit: Classic original007 requires replacing the draft, shared16 requires preserving it. No criterion or current replacement behaviour changed; failed activation/recovery acceptance also remains open.

- [x] Shared4 independent native idle typing/grouping/ranking: initial character filtered groups, exact/prefix/non-title multiple-result fallback, arrow wrap and Enter workspace activation once, exact durable draft/media bytes/file/message refs preserved. Browser258/functional87/helpers62/Go-vet-build-hook; Classic81/236+shared25/42 unmapped155/17. Test-only/no restart; shared16 slash/failure contract and terminal acceptance separate.

- [x] Shared15 compact Quick Actions Close: pointer/Tab Enter/Space/trusted touch use one-shot dismissal; exact drafts/refs/media bytes preserved, focused action rows activate once, layout-effect reopens reliably. Header geometry/no wrap; pinned source/CSS unchanged, separate Gi stylesheet. Browser252/functional87/helpers62/Go-vet-build-hook; Classic81/236+shared24/42 unmapped155/18; pushd0b209e/restart8090 live6browser-size mouse/touch/Enter/Space/Escape focus+draft/zero writes/errors/61sessions/integrityOK; captures attached, no idle/TUI rows.

- [x] Shared13/14 Quick Actions dismissal: focusable native Conversation region, captured-opener restoration, cancelled/fenced focus rAF and consumed trusted outside clicks; exact draft/media bytes/file/message refs, touch and Alt+Enter regression. Browser234/functional87/helpers61/Go-vet-build-hook; Classic81/236+shared23/42 unmapped155/19. Push5353392/restart8090; live6browser-size Escape/mouse/touch focus+draft/zero writes/errors/61sessions/integrityOK. Supplied sources unchanged/no TUI rows. Shared15 close control absent/uncredited.

- [x] Shared7/9/10 target event delivery: replace swallowing global Quick Actions guard with non-consuming predicate; Settings normal keys reach controls, background popups suspend keys/pointers, composing Escape/Tab wrap verified. Supplied sources unchanged via guarded adapter; browser216+108/functional86/helpers58/Go-vet-build-hook. Classic81/236+shared21/42 unmapped155/21; push4ae11da/restart8090 live6browser-size target keys/modal focus/draft retained, zero writes/errors,61sessions/integrityOK; terminal unchanged.

- [x] Shared5/6/11/12 independent native composer Unicode/editing, Settings filter suffix, session-ID typing/Escape and rejected timeline key metadata with positive real-key controls after every rejection. Combined192/functional85/helpers55/Go-vet-build-hook; Classic81/236+shared18/42 unmapped155/24. Test-only/no restart, unsupported editor/dismissal cases uncredited; terminal keyboard ownership separate.

- [x] Shared1/2 menu safety: fix outside-click Send activation and Escape focus loss via guarded dismissal-only build adapter; supplied source unchanged. Tab access, pointer/keyboard dismissal, trusted touch/next gesture and cancel-path checks; combined192/functional85/helpers55/Go-vet-build-hook. Classic81/236+shared14/42 unmapped155/28. Push8580f57/restart8090; live Chromium/WebKit three sizes, mouse/touch/Escape focus+draft retained, zero writes/errors,61sessions/integrityOK; no terminal UI change.

- [x] Shared3 native workspace show/hide and narrow backdrop: real pointer blocking over textarea/Send, exact stored text/media bytes/file/message refs and session/history preserved. Workspace24/functional85/helpers51/Go-vet-build-hook; Classic81/236+shared12/42 unmapped155/30. Test-only/no restart or Plan/TUI credit.
- [ ] Shared39 requires an explicit native upload Cancel action; progress and failed-upload recovery alone do not satisfy cancellation/retry/source-removal acceptance. Current XHR abort handler has no exposed cancel control.

- [x] Shared26 native session capability gating: delete absent for root/running/unknown-count selections with native405, archive/edit conflicts409, blank rename400/retry, persistent pin/archive/restore/unpin and actual child creation; draft/media preserved. Sessions120/functional85/helpers51/Go-vet-build-hook; Classic81/236+shared11/42 unmapped155/31. Test-only/no restart; terminal mutation submenus remain separate work.

- [x] Shared27 two native composer follow-ups: active-turn checks, held POST/SSE ID reconciliation, distinct request tokens, exact text/file/folder/message refs and media bytes, reload/FIFO consumption and research isolation. Queue48/functional85/helpers51/Go-vet-build-hook; Classic81/236+shared10/42 unmapped155/32. Test-only/no restart; compact terminal queue/media actions still require separate implementation and acceptance.

- [x] Shared25 coherent keyboard session switch: held native timeline/model/queue/activity/compaction reads cannot replace the selected history, durable queue IDs, context or draft/media. Context/shared36/session114/functional85/helpers51/Go-vet-build-hook; Classic81/236+shared9/42 unmapped155/33. Independent review passed; test-only/no restart. Alt-S adaptation stays bounded, terminal queue/media acceptance separate.

- [x] Shared31/32 native model-picker pointer/keyboard typeahead, real registry32K/200ctx and measured100tokens, held PATCH200 confirmation, retained text/media/file/exact message refs, reload/other-session isolation. Six-project context+shared30/functional85/helpers51/Go-vet-build-hook; Classic81/236+shared8/42 unmapped155/34. Test-only/no restart; Alt-M evidence separate/no idle rows.

- [x] Shared23/24 session-picker pointer/Enter opening: insertion/first-frame focus, both native triggers, composer-relative geometry/resize, native-ID search/Escape restore/draft and session unchanged. Dedicated12/fullsession114/functional85/helpers51/Go-vet-build-hook; Classic81/236+shared6/42 unmapped155/36. Test-only/no restart or new Alt-S credit. Mobile001 fixture now waits for native idle before tail-order assertion.

- [x] Shared29 durable adjacent queue reorder and native consumed-active DELETE409 reconciliation: other-session queue equality/two drafts/reload/final FIFO proof; dedicated6/fullqueue42/functional85/helpers51/Go-vet-build-hook. Classic81/236+shared4/42 unmapped155/38. Test-only, no restart/TUI credit; compact terminal queue actions remain design-only.

- [x] Status swipe mobile003: bounded conversation-host listener admits timeline/direct-child status only; native draft/thought links, selection/vertical/composer/Settings and excluded pen/default proof. Panels12/session102/steer18/functional85/helpers48/Go-vet-build-hook pass; Classic76/236+shared3/42 unmapped160/39. Push28996cf/restart8090 live status+4rapid transitions/draft retained/no writes/errors/61sessions/integrityOK; TUI unchanged/no idle rows.
- [x] Streaming preview correction: pass strings to supplied normalizer, key by session/native turn, reject mismatched tagged deltas, restore turn identity from activity; Gi host40% scroll bound. Thoughts002–005 pass; newline disclosure is Gi-derived only. Matrix36/status12/reconnect60/functional85/helpers49/Go-vet-build-hook pass; Classic80/236+shared3/42 unmapped156/39. Push965e819/restart8090, live status+rapid4/draft/no writes/errors, CSS200/61sessions/integrityOK; no TUI rows.
- [x] Thoughts001 wrapped disclosure: guarded build-time adapter + real collapsed scrollHeight/clientHeight, generic Show more (no invented counts), RO/font/stream/viewport and noRO ancestor fallback. Preview48/status12/functional85/helpers51/Go-vet-build-hook pass; Classic81/236+shared3/42 unmapped155/39; supplied sources unchanged. Pushbdf731e/restart8090 live status+rapid4/draft restored/no writes/errors/61sessions/integrityOK; no TUI rows.

- [x] Frozen mobile004 native pinned/active/ordinary order and archive/dedup acceptance: dedicated6/mixed30/session96/functional83/helpers47/Go-vet-build-hook; Classic75/236+shared3/42 unmapped161/39. Test-only, no runtime restart or new TUI credit. Gesture assertions wait for visible native catalogue activation.
- [x] Rapid swipe activation: host layout-effect listener replacement fixes previous-session closure on next-frame reverse; supplied helper/draft/generation code unchanged. Red→green6/browser102/functional84/helpers48/Go-vet-build-hook; held B timeline/media/drafts proof; pushbb92551/restart8090, live4transitions/draft restored/no writes/errors,61sessions/integrity/FK OK. No new parity/TUI credit.
- [x] Stabilise TestProcessSystemDirectWhileActiveSteersSameSession fixture: isolated temporary DB, setup gate and runner-lock cleanup join; original same-turn/role/metadata assertions intact. Repeat100/race50/fullGo-vet-build-hook/support48/functional84 pass; review found no issues. Test-only, no production scheduling or parity/TUI change.

- [x] Settings header/responsive004: pinned 860/720 client-width classes, header model filter/focus and per-visit Apply lock; shell90/regression216/functional83/helpers44/Go-vet-build-hook pass, Classic74/236+shared3/42 unmapped162/39. Pushed6df80e3/restart8090; live phone/tablet/desktop header filtering/focus/classes, 61sessions/integrity/FK OK, no writes/page errors. TUI Alt-M unchanged/no idle rows.
- [x] Model pane re-entry freshness: captured-session settlement invalidation/refetch, synchronous read generations, preserved newer drafts, separate read/action errors and explicit Refresh; browser234/functional83/helpers46/Go-vet-build-hook/HTTP graph pass, keyed-session review verified. Pushed78d1e68/restart8090; read-only live Refresh/query preservation at3sizes, no writes/page errors, 61sessions/integrity/FK OK. No broader settings008/009/010 or TUI credit.

- [x] Settings lazy modules: static General, hashed non-General panes on selection, loading/error/module cache with native data still fresh; bootstrap/hashed-app graph prevents duplicate init. Frozen dialog005/settings003 pass; shell72/regression186/compaction96/providers36/functional83/helpers44, Go/vet/build/hook. Stable bootstrap no-store, native HTTP graph/404/explicit reload proof and independent review pass. Classic73/236+shared3/42, unmapped163/39. Cached-data gap/TUI unchanged. Pushed0958a89/restarted8090; live one app/four deferred pane requests with cache-only revisits, no settings writes/page errors, 61sessions/integrity/FK OK.

- [x] Verify frozen Settings layering001–004 and dialog001/003/004: real workspace pointer blocking/recovery, portal geometry/dimming, rapid shortcut, held loading and typed numeric fields; fix first-paint Escape via layout effect. Frozen42/42 + shell/Gi132/132 + rapid tablet10/10, functional83/83, helpers43, Go/vet/build/hook pass. Classic71/236+shared3/42, unmapped165/39. Cached/lazy cases unmapped; no TUI credit. Push018fd83/restart8090: live rapid open/Escape, 61sessions/integrity/FK OK.

- [x] Gi Providers (`gi-settings-020`–`023`): OpenAI/Anthropic key save/confirmed remove, metadata-only OAuth/custom rows, private atomic native auth store shared by TUI logout; auth/origin/TLS-or-loopback-peer+Host/CAS/no-secret/failure/reopen proof. Local-provider six-project36, Settings/Models126, functional83, helpers43, Go/vet/build/hook/race×3/Darwin compile pass. No operator key writes, frozen credit or TUI rows. Push3829f55/restart8090: metadata/pane read-only smoke, 7 provider rows/61sessions/integrity/FK OK.

- [x] Gi automatic policy (`gi-settings-018`/`019`): shared atomic locked Pi settings writers, revision-bound save/restart semantics, unknown-field preservation and concurrent/process/unsafe-file/failure proof. Settings/Models126, compaction96, functional83, helpers43, Go/vet/build/hook/race×3, Darwin config compile and three-size TUI compaction pass. No provider credentials, auto-restart, frozen credit or idle rows. Pushf81a3ca/restart8090: read-only live policy/form match, unchanged threshold108000, 61sessions/integrity/FK OK.

- [x] Gi Compaction (`gi-settings-014`–`017`): effective engine policy read-only, token-bound Compact, matching-turn Stop, authoritative progress during delayed POST, read/action failure and close/switch guards. Native six-project compaction 96/96, Settings/Models108/108, functional83/83, helpers43, Go/vet/build/hook/web-turn race×3 pass. No policy writes, unsupported Piclaw controls, frozen credit or TUI rows. Push12de67b/restart8090: live read-only policy/pane smoke, 61 sessions, integrity/FK OK.

- [x] Gi identity (`gi-settings-012`/`013`): revision-checked atomic name save preserving config keys/avatars, rooted file checks, cross-process lock and restart-required General UI. Settings 72/72, Settings/Models 102/102, reviewed identity 18/18, functional 83/83, helpers 43, Go/vet/build/hook/config-web race ×3 pass. No avatar/credential edits, auto-restart, frozen credit or TUI chrome. Push ccc2e53/restart 8090: live GET/form matches without changing operator identity, 61 sessions, integrity/FK OK.

- [x] Gi Appearance (`gi-settings-009`–`011`): explicit browser-local save/reset, validated hex tint, storage denial/retry, reload/session/tab scope and dirty cross-tab drafts. Unchanged supplied renderer via narrow build export; settings/models 90/90, functional 83/83, helpers 43, Go/vet/build/hook pass. No server/TUI writes or frozen credit (64/236 + 3/42 unchanged). Push 7917cb8/restart 8090: live save/reload proof, 61 sessions, integrity/FK OK.

- [x] Gi-specific settings: 8 derived scenarios + 3 future proposals in `tests/features/settings/gi-settings.feature`; General read-only instance snapshot + explicit session Models edits. Settings 36/36, native catalogue/context 24/24, model/session regression 156/156, functional 82/82, helpers 40, Go/vet/build/hook and auth/model race ×3 pass. No frozen credit (64/236 + 3/42, unmapped 172/39), no TUI idle rows. Push 5179a19/restart 8090: live General/Models open and close, 61 sessions, integrity/FK OK; see `docs/internal/gi-settings-plan.md`.

- [x] Implement frozen `@ux-pwa-001`: linked `/manifest.json` declares any/maskable 192/512 PNGs and its `/static/icon-*.png` URLs resolve to embedded images. Go GET/HEAD, six-project browser + swipe regression 30/30, functional 81/81, 39 helpers, Go/vet/build/hook pass; Classic 64/236, shared 3/42, unmapped 172/39. Push de3349d/restart 8090: 61 sessions and manifest/icons HTTP 200, integrity/FK OK. Avatar-dependent `002/003/006` and Apple/favicons `004/005` need separate evidence; PWA installation has no TUI equivalent.

- [x] Verify frozen `@ux-mobile-006`: wire native Safari detection for horizontal wheel listener; Chrome and iOS do not navigate, desktop Safari positive control switches adjacent persisted session. Six-project swipe 24/24, picker 90/90, functional 80/80, 39 helpers, Go/vet/build/hook pass. Classic 63/236, shared 3/42, unmapped 173/39. Push d50e2f8/restart 8090: 61 sessions, integrity/FK OK. Browser-only gesture; terminal Alt-S separate; `002/003/004` unmapped.

- [x] Verify frozen `@ux-mobile-005`: a primarily vertical first move cancels a touch despite later horizontal movement, then a fresh eligible timeline gesture navigates the adjacent native candidate. Six-project browser and picker 84/84, functional 80/80, 39 helpers, Go/vet/build/hook pass. Classic 62/236, shared 3/42, unmapped 174/39. Push f88f1ce/restart 8090: 61 sessions, integrity/FK OK. Browser-only touch, terminal Alt-S evidence separate; `002/003/004/006` unmapped.

- [x] Verify frozen `@ux-mobile-001`: eligible horizontal finger gesture on supplied timeline selects adjacent persisted non-archived candidate and wraps from catalogue end, without active text selection; six projects, picker regression 78/78, functional 80/80, 39 helpers, Go/vet/build/hook pass. Classic 61/236, shared 3/42, unmapped 175/39. Push 0ef320a/restart 8090: 61 sessions, integrity/FK OK. Touch is browser-only; terminal Alt-S evidence separate.

- [ ] `@ux-workspace-005` remains unmapped: six-project probe cannot find the workspace-header menu because `internal/web/static/css/timeline-menu.css` hides `.workspace-header-left > .workspace-menu-wrap` in favour of the timeline menu. Frozen criterion requires the header menu; do not credit the replacement. Repair without duplicate idle chrome, then test Refresh/Reindex/hidden/create/upload and absence of in-pane file search. Terminal Alt-I evidence is separate.

- [x] Verify frozen `@ux-original-023`: real SSE outage/reconnect refreshes selected timeline/activity/queue; visible Stop posts captured active turn ID and authoritative state reports cancellation, leaving older run unable to retake control or lose draft. Six projects, reconnect regression 60/60, functional 80/80, 39 helpers, Go/vet/build/hook pass. Classic 60/236, shared 3/42, unmapped 176/39. Push 4ca897f/restart 8090: 61 sessions, integrity/FK OK; terminal `/cancel` evidence remains separate.

- [x] Verify frozen `@ux-original-022`: native sparse model without reasoning/context metadata renders its reported label only, unknown context shows `?` and unavailable provenance, and delayed superseded-chat response leaves target model/draft unchanged; six projects, model regression 36/36, functional 80/80, 39 helpers, Go/vet/build/hook pass. Classic 59/236, shared 3/42, unmapped 177/39. Push 8afcd36/restart 8090: 61 sessions, integrity/FK OK; terminal model/footer evidence remains separate.

- [x] Verify frozen `@ux-original-020` through supplied model picker: native captured-session PATCH, accepted model and context fields, persisted reload/keyboard selection, rejected 400 warning and untouched draft/media pass six projects; model regression 30/30, functional 80/80, 39 helpers, Go/vet/build/hook pass. Classic 58/236, shared 3/42, unmapped 178/39. Push 14f63b6/restart 8090: 61 sessions, integrity/FK OK. Terminal Alt-M and `/model` evidence remains separate.

- [x] Verify frozen `@ux-session-005`: Gi adapter wires supplied timeline swipe; native active-first/JID order and archived exclusion, real interactive and text-selection guards, draft preservation pass six projects; picker regression 72/72, session family 36/36, functional 80/80, helpers 39, Go/vet/build/hook pass. Classic 57/236, shared 3/42, unmapped 179/39. Push 4b6d88a/restart 8090: 61 sessions, integrity/FK OK. Touch remains browser-only; TUI Alt-S stays bounded with no idle rows.

- [x] Verify frozen `@ux-session-002`: native current, pinned, live active, same-tree, other and archived entries group in order through the unchanged picker; sole current row, preserved draft/chat pass six projects; picker regression 66/66, functional 80/80, 39 helpers, Go/vet/build/hook pass. Classic 56/236, shared 3/42, unmapped 180/39. Push a5a93f6/restart 8090: 61 sessions, integrity/FK OK. Terminal Alt-S keeps separate evidence and no added idle rows.

- [x] Verify frozen `@ux-session-004` through supplied picker and native child archive/restore PATCH: held accepted response, authoritative regrouping, transport failure and unchanged origin draft/chat pass six projects; picker regression 60/60, functional 80/80, helpers 39, Go/vet/build/hook pass. Classic 55/236, shared 3/42, unmapped 181/39. Push 11ab7de/restart 8090 (61 sessions, integrity/FK OK); terminal mutation submenus still design-only.

- [x] Verify `@ux-session-001` selected-chat timeline and superseded generation with persisted native posts, a held real HTTP response and visible picker: six-project pass, combined picker 54/54, 80 functional, 39 helpers, Go/vet/build/hook; two drafts isolated. Aggregated 54/236 Classic, 3/42 shared, 182/39 unmapped. Push 49f00d9/restart 8090 (61 sessions, integrity/FK OK); compact terminal session selection has separate evidence, no new terminal credit.

- [x] Verify frozen `@ux-session-003` through native session ID, persisted model and two-term handle search: six-project filtered keyboard navigation, draft and current chat unchanged until Enter; picker regression 48/48, 80 functional, 39 helpers, Go/vet/build/hook pass. Aggregated 53/236 Classic, 3/42 shared, 183/39 unmapped; push 10328bd/restart 8090 (61 sessions, integrity/FK OK). Status matching remains helper-only; terminal Alt-S acceptance is separate.

- [x] Verify frozen `@ux-session-006` with the supplied picker: Escape clears search/typeahead, restores both initiating triggers and activates no session; focused 6/6 and picker regression 42/42 pass with repeated reopen, persisted history and unsent draft. 80 functional/39 helpers, Go/vet/build/hook pass; aggregated 52/236 Classic, 3/42 shared, 184/39 unmapped. Push ad62a57/restart 8090 (61 sessions, integrity/FK OK); terminal Alt-S dismissal has separate evidence and gains no new credit.

- [x] Preserve exact stored latest-assistant source bytes in terminal `/copy` with truthful byte count and native/OSC 52 opt-in; six fullscreen/regular PTYs at 60×18/100×22/140×36 verify emitted bytes, clipboard-off fallback, draft and zero idle-row growth. Go/vet, TUI race×3, selection/regular/search/compaction/smoke/Gherkin and 39 helpers pass; ADR-0058. Selected-message deletion remains blocked on rendered-row identity, no feature-file credit.

- [x] Verify `@shared-37` as one integrated native copy/delete flow across all six browser projects: stored Markdown/code clipboard bytes, success/failure glyph reset, real busy-run 409 then 200 for the captured ID, unaffected other session/message reference/draft. Focused 6/6, 80 functional, 39 helpers, Go/vet/build/hook pass; aggregated 51/236 Classic, 3/42 shared, 185/39 unmapped. Push e05873d/restart 8090; 61 sessions/identities, HTTP 200, integrity/FK OK. Terminal source-copy and selected-message deletion remain separate zero-idle-row designs.

- [x] Add idle-only single-message native DELETE with atomic checkpoint version invalidation, retained audit/media and success-only timeline animation; focused 24 browser, full six-project 450 browser executions, 80 functional, 39 helpers, Go/vet/race and hook checks pass. Aggregated Classic 51/236, shared 2/42, unmapped 185/40; push dadd34b/start 8090 (61 sessions/61 identities, 146 messages/51 turns/317 events, SQLite integrity/FK OK). Reply/cascade and terminal credit excluded; ADR-0056.

- [x] Recover pre-canonical Gi session identities from retained legacy scope/aliases in one schema transaction; incomplete/colliding scope, late rollback, reopened current-schema missing-identity guards and race×3 pass. Copy and live 8090 return 200 for 61 sessions/61 scoped identities after restart; original session/message/turn/event table hashes match, 146 messages/51 turns/317 events preserved, SQLite integrity and FK checks pass. 80 functional, Go/vet/build/hook; push d3d3c6c/restart 8090. No frozen parity or TUI credit.

- [x] Upgrade go-ai to accepted upstream-v0.87.1/c2d0231 independently of stashed deletion WIP; retain all other pins including go-tui v0.18.2; 80 functional, 38 helpers, 30 model and 66 context/compaction browser executions, Go/vet/race, three-size terminal suites and Linux/macOS amd64/arm64 builds pass. No new frozen parity or terminal feature credit.

- [x] Verify native whole-message Markdown/rich clipboard bytes, fallback/denial/reset and late-session/draft/media preservation; fix absent API false success via narrow build adapter, supplied files unchanged; ADR-0055, 594 browser/80 functional/38 helpers/Go-vet-build-hook; copy/delete original-024/shared-37 remain unmapped, no terminal credit.

- [x] Import unchanged pinned Quick Actions with native capability catalogue, typing guard and session-owned compose prefill; original003/005/006/007, 564 browser/79 functional/36 helpers/web race×3/Go-vet-build-hook plus48 capture+24 transition stress; ADR-0054, frozen50/236+2/42/unmapped186/40; 004/008 and conflicting queue replacement remain gaps, TUI design only.

- [x] Close compose-005 with native XHR upload-byte progress and separate session-owned sending button state; real gates/bytes/overlap/rejection/reload/newer-draft proof, supplied files unchanged; ADR-0053, 516 browser/78 functional/34 helpers/Go-vet-build-hook; frozen46/236+2/42, unmapped190/40; terminal adaptation design only.

- [x] Repair legacy startup: atomically create tables, apply additive columns, then build indexes; rollback/idempotence tests, real-copy migration, 70/70 functional and Go/vet/race checks. Dev instance restarted on 8090; two-way backup comparisons preserve 51 turns, 146 messages, 317 events and existing session fields.

- [x] Vendor the frozen Vibes/Tau Classic corpus and shared interaction contract with hash checks.
- [x] Add a Gi matrix covering Chromium/WebKit at phone/tablet/desktop sizes; unmapped cases are not passes.
- [x] Map `@ux-original-001` and `002`, import the upstream TimelineMenu and drawer behavior, and pass all 12 executions.
- [x] Map `@ux-original-014` with real delayed-response/draft-switch checks and test New creating a distinct child (24/24 combined browser runs).
- [x] Fix pooled SQLite configuration exposed by the session matrix; existing functional web suite passes 70/70.
- [x] Map searchable picker/focus (`013`) and native keyboard navigation; combined matrix 36/36 with pinned helper provenance.
- [x] Prevent runner startup before submission-event persistence; reproduced regression and full 70/70 functional suite.
- [x] Verify native timeline table layout, trusted code-copy payloads and source-only SVG fences (timeline-023/024, original-029); host-only table/overflow override; 432/432 browser, 70/70 functional, Go/vet/hook and 29 helpers; ADR-0037. Compound link-preview/copy-delete/speech remain gaps.
- [x] Stabilise context-meter capability sampling by waiting for native compaction busy/claim reason to clear; exact tooltip/colour assertions retained, six-project suite passes twice; ADR-0047.
- [x] Verify original-026 native upload IDs, parser rejection/no submission, origin-owned merged recovery and explicit retry after partial success; 444/444 browser, 70/70 functional, Go/vet/hook/29 helpers; ADR-0038. No application changes. Compose-005 upload progress and terminal media remain gaps.
- [x] Upgrade go-ai to upstream-v0.87.0 (`c51fb076ad9f`), Go 1.26.8 and compatible Go/web dependencies; pair KaTeX JS/CSS/fonts and repair SSE/dispatcher test synchronisation. 444 browser, 71 functional, 29 helpers, Go/vet/race ×3, all three-size TUI/smoke/Gherkin and Linux/macOS cross-builds pass. See dependency-upgrade-20260922.md.
- [ ] Revisit go-tui ≥0.19 only after native regular-mode resize preserves completed history; 0.19.0/0.22.1 erase a response, so retain 0.18.2. Keep gVisor at Tailscale's required generated-source version.
- [x] Project native media into timeline/search and add authenticated metadata/raw lookup; verify lightbox 013–016 with trusted touch, draft/session/search/reload/404 recovery; 474 browser, 72 functional, Go/vet/build/hook, web race ×3, 31 helpers; ADR-0039. Supplied components unchanged; no terminal credit.
- [x] Wire supplied read-only workspace previews to bounded native metadata/text/raw routes, rooted filesystem access and safe response headers; workspace-008 Markdown/text/image/binary verified; 480 browser, 73 functional, 31 helpers, Go/vet/build/hook/web race ×3; ADR-0040. No editor/CRUD or terminal credit.
- [x] Wire native bounded subtree/hidden queries and host global-menu→explorer visibility bridge; workspace-004 root/expanded reload, on/off/on/reload/draft proof; 486 browser, 74 functional, 31 helpers, Go/vet/build/hook/web race ×3; ADR-0041. Legacy tree preserved; no reindex/CRUD/terminal credit.
- [x] Compare indexing at pinned Piclaw/Tau/Vibes revisions; preserve uncommitted full-rebuild prototype in stash; derive 15 proposal scenarios outside frozen corpus and test candidate scope/membership/chunk/FTS/lease SQL (not a runtime migration). See docs/internal/search/indexing-lineage-20260922.md. No new parity credit.
- [x] Install scoped workspace index v1 inside the core startup transaction with ledger/checksum/object validation; preserve legacy search/runtime tables; late-failure rollback, concurrent/reopen/FTS tests and copied dev DB 19-table hash/integrity rehearsal; Go/vet/build/hook, 74 functional, 32 helpers, store race ×3, focused 12 browser and all three-size TUI/smoke/Gherkin pass; ADR-0042, no new parity credit.
- [x] Add internal scoped configuration, workspace-wide fenced Begin/Renew/Fail/Commit, expired/replaced-owner guards, stable unchanged identities and atomic overlapping-scope cleanup on v1 tables; two-store/final-fence/write-failure/cancel/reopen/bounds tests; Go/vet/build/hook/74 functional/32 helpers/store race ×3; ADR-0043. No runtime scanner/API/UI or parity credit.
- [x] Implement rooted bounded complete-inventory scan and deterministic UTF-8 8KiB/128-line chunks; skills eligibility, byte/entry/depth limits, identity/hash changes, missing/unreadable roots, symlink/FIFO swaps, workspace/ancestor replacement and scan→commit identity/rollback/scoped cleanup tested; Go/vet/build/hook/74 functional/32 helpers/race×3/2.38M fuzz/Linux-macOS builds; ADR-0044, no worker/API/UI or parity credit.
- [x] Implement explicit lease-renewing Begin→ScanScope→Commit/Fail worker; join renewals before publication, separate bounded failure cleanup, native cancellation/renewal-write-failure/takeover and killed-child recovery proof; Go/vet/build/hook/74 functional/32 helpers/worker-store race×3; ADR-0045, no scheduler/API/UI or new parity credit.
- [x] Wire authenticated scoped lexical query/status/explicit reindex to worker plus host menu Refresh/Reindex; native scope/config/bounds/auth/failure/ownership proof and real missing-root browser failure/retry/draft recovery; 492 browser/75 functional/32 helpers/Go-vet-build-hook/search-store-indexer-web race×3; ADR-0046. No new frozen credit, background freshness or terminal UI.
- [x] Load workspaceIndex extraRoots/extraExtensions/optionalRoots at startup; preserve strict defaults/fingerprint, reject populated-root disappearance transactionally, retain results/drafts/reload/retry; 498 browser/76 functional/32 helpers/Go-vet-build-hook/config-search-web race×3; 17 derived proposals, no new frozen credit; ADR-0047.
- [x] Add atomic v2 invalidation revisions with unchanged v1/checksum/data, refresh capture/ack and pending-stale publication; two-store/event-commit races, failure/reopen, worker held-after-scan and copied dev DB 30-table/v1-ledger preservation; Go/vet/build/hook/76 functional/32 helpers/race×3/focused24 browser/TUI smoke+Gherkin; ADR-0048, no scheduler/watcher or new credit.
- [x] Add internal single-workspace fixed-scope scheduler: shared tickets, bounded contention/stale retry, peer-generation completion, final-read/request race protection and cancel/join shutdown; native tests, race suites×3 and scheduler×10, Go/vet/build/hook/76 functional/32 helpers; ADR-0049 and 20 derived proposals; no production caller, query/watcher activation or parity credit.
- [x] Wire explicit application-owned scheduler and POST coalescing, cancellation-isolated waits, startup/GET scan-free proof and joined primary/ACME HTTP/index shutdown; ADR-0050, native/race×10/full web-indexer-HTTP race×3/built-process SIGTERM-bind failure/77 functional/32 helpers/focused24 browser; 21 derived proposals, no watcher/query refresh/TUI or new parity credit.
- [x] Connect engine/web/script regular filesystem writes to atomic scoped invalidation before/after mutation; bounded failure/cancel paths, rooted no-symlink writes, explicit crash/alias limits; ADR-0051, Go/vet/build/hook/77 functional/32 helpers/focused24 browser/tools-store-web-turn race×3/Darwin arm64 cross-compile; no automatic refresh or new parity/TUI credit.
- [x] Add Alt-I five-row explicit index Status/Reindex panel, native joined lifetime and epoch/session guards; three-size fullscreen+regular PTY bytes/failure/retry/resize/draft/cursor/reader/busy/multiline/history/selector checks, TUI-indexer race×3, existing regular/search/smoke/Gherkin and Go-vet-build-hook/77 functional/32 helpers; ADR-0052, no idle rows or automatic freshness.
- [ ] Implement remaining index integration: external/shell/watch delivery, restart recovery, nonblocking background refresh and query consumers; keep provisional global rebuild stashed.
- [ ] Map the remaining 191 frozen scenarios and 40 shared interaction cases; full scope in `docs/internal/full-web-tui-parity-plan.md`.
- [x] Complete capability-gated rename, pin, archive and restore with persisted native metadata and failure-safe picker actions (`015`); 48/48 matrix, 70/70 functional, Go/race/Bun checks.
- [x] Persist per-session browser text/media/reference drafts and unacknowledged submissions; recover on reload without automatic resend (browser-local IndexedDB).
- [x] Capture background-send ownership, merge failed submissions with newer origin drafts and report storage failures (`ux-compose-001`, `002`, `003`, `006`); 102/102 matrix, 70/70 functional.
- [x] Honour explicit queued follow-ups during active turns; add persistent FIFO/reorder and queued-only cancel APIs with conflict guards.
- [x] Wire reorder/removal error recovery and session-owned late responses (`018`); 120/120 matrix, 70/70 functional, Go/race/Bun tests.
- [x] Complete `016` accepted/SSE queue reconciliation, disconnect transient clearing (`ux-reconnect-001`) and connection-generation guards; 138/138 matrix, 70/70 functional, Go/race/Bun checks.
- [ ] Complete remaining reconnect criteria separately; shared return/Steer pass, conflicting Classic replacement/idle-send cases remain unmapped.
- [x] Add authoritative session-local model commands/mutations with catalogue validation, persistence, failed-switch retention and late-response guards (`021`, compaction `008`); 168/168 matrix, 70/70 functional, Go/race/Bun checks.
- [x] Share validated model selection between web and TUI; persist terminal choices per session, restore on startup/switch, preserve editor state and settings bytes.
- [x] Verify bounded model selector/cycle/error paths at 60×18, 100×22 and 140×36 with zero added idle rows; live restart/next-turn checks, TUI smoke/Gherkin and Go/race tests.
- [x] Persist latest provider-request input/cache measurement separately from cumulative turn usage; expose session-scoped context and validate model fit without fabricated zeroes.
- [x] Render truthful context metadata/unknown state with capability-gated compaction; native/helper tests and `ux-context-002` browser acceptance (174/174 matrix; other context IDs remain open).
- [x] Implement shared queue return: merge latest origin text/media/refs, persist recovery before DELETE, retry without duplication and block DELETE on storage failure; Classic replacement conflict documented in ADR-0017.
- [x] Add separately reported shared-contract browser mappings; shared-28 passes six projects (192/192 executions overall), with incompatible Classic return cases still unmapped.
- [x] Complete run-bound atomic queue Steer (shared-30), disabled idle/unknown actions, at-most-once consumption and held recovery; 12/12 local-provider matrix, 192/192 existing matrix, 70/70 functional, Go/vet/race and 23 helper tests. See ADR-0018; Classic idle-send remains unmapped.
- [x] Verify measured context-fit rejection/accepted switch (compaction 006/007) through local-provider browser matrix, preserving drafts, native rejection and per-session usage; 12/12 context-fit, 216/216 combined browser, 70/70 functional, Go/vet and 23 helpers.
- [x] Verify context-meter formatting, tooltip data, clamped fill and boundary colours (context 001/005) through native local-provider measurements; 228/228 combined browser, 70/70 functional, Go/vet and 23 helpers. Three-size terminal footer/live regressions pass without UI changes; unsupported compaction cases stay unmapped.
- [x] Make automatic compaction lifecycle cancellation-safe and durable before UI wiring: atomic occurrence-keyed phase/event/summary updates, terminal suppression/failure outcomes, no cancelled-status overwrite. Go/vet, targeted race ×3, hook checks and 70/70 functional; ADR-0019. No browser parity credit or manual Compact capability added.
- [x] Wire automatic-compaction activity snapshot/Stop API and host status/meter adapters; compaction 001–005/context 004 pass across six projects with draft/media retention, reload, late-session and failed-stop acceptance. 276/276 combined browser, 70/70 functional, Go/vet/race and 24 helpers; ADR-0020. Manual Compact remains disabled.
- [x] Persist ID/fingerprint-owned eligible context checkpoints without deleting timeline; exact native projection/snapshot/prefix guards, atomic summary commit, cancellation/failure preservation and edit/delete fallback. 54/54 compaction, 282/282 combined browser, 70/70 functional, Go/vet/race, 24 helpers and three-size TUI regression; ADR-0021. No new frozen mapping.
- [x] Implement explicit idle-only manual Compact with atomic snapshot-token/claim admission, cancellable maintenance turn/no provider request and host callback preserving drafts/media; context-003, 66/66 compaction, 294/294 combined browser, 70/70 functional, Go/vet/race and 24 helpers. ADR-0022; terminal command remains informational.
- [x] Wire terminal /compact and draft-preserving Alt-C to shared maintenance admission; native scoped lifecycle/Stop, focused Escape fix, /compact info diagnostics. Live 60×18/100×22/140×36 gate/cancel/resize/reopen/zero-idle-row checks, unit/race, existing TUI suites and 70/70 functional pass; ADR-0023.
- [x] Verify reconnect refresh of native timeline/status/queue/context (002), fence pre-disconnect HTTP responses/errors, and surface real server asset-version drift without reload (004); 18/18 reconnect, 312/312 combined browser, 70/70 functional, Go/vet and 26 helpers. ADR-0024; search/pagination/full crash criteria stay open.
- [x] Implement bounded native message search/current-family-all scopes and supplied search composer; preserve draft/media and active view across real reconnect with query/session/connection guards; reconnect-003. 330/330 browser, 70/70 functional, Go/vet/race and 27 helpers; ADR-0025. Hashtag/paging and terminal search remain open.
- [x] Coordinate initial chat activation/SSE-ready refresh (reconnect-005) without pre-subscription snapshots or skipped reconnect; delayed readiness, failed initial read, A→B→A/native reconnect tests. 342/342 browser, 70/70 functional, Go/vet and 28 helpers; ADR-0026. All reconnect IDs mapped; broader crash/paging/hashtag gaps remain.
- [x] Implement stable bounded native message paging, promise-owned older loads and viewport anchoring; reconnect catches up across pages without replacing loaded history; old pages cannot overwrite search. 348/348 browser, 70/70 functional, Go/vet/race and 29 helpers; ADR-0027. No new frozen mapping.
- [x] Verify accepted-message refresh/visibility, native multi-upload ID/name/byte pairing, newer text/cursor, reader anchors and search/origin ownership; compose-007/009/010/011, 384/384 browser, 70/70 functional, Go/vet/hook and 29 helpers; ADR-0028.
- [x] Establish explicit folder action with unchanged explorer navigation; exact text/file/folder/message reference and references-only submission maps compose-008; durable session drafts/failure recovery, 402/402 browser, 70/70 functional, Go/vet/hook and 29 helpers; ADR-0035.
- [x] Reconcile completed-claim admission: atomic running-claim steering gate, distinct queued fallback with cleanup-owned release/FIFO, no terminal-run steering or early successor overlap. Held native hook/browser, two-connection race/rollback/capacity/maintenance tests; 408/408 browser, all TUI suites, 70/70 functional, Go/vet/race ×3 and 29 helpers; ADR-0036. No new frozen mapping.
- [ ] Verify terminal structured file/folder/message reference serialization and captured failure recovery using existing editor/commands without extra idle chrome; web compose-008 does not establish terminal credit.
- [x] Preserve terminal follow mode during editing and ordinary running/idle submission; native delayed-provider completion/newer cursor/same-row anchor/resize/explicit-follow tests at 60×18, 100×22 and 140×36 add zero idle rows. Go/vet, terminal race ×3, all existing TUI suites, 29 helpers and 70/70 functional tests; ADR-0029.
- [ ] Complete terminal routed acceptance/recovery, durable transcript-ID reconciliation and block anchors across wrapped reflow/scrollback eviction.
- [x] Adopt Pi dark user/pending/success/error bands without tool box rows; rendered-height fullscreen paging/bottom, focused Home/End with Ctrl variants for editor, Ctrl-O expansion and dock-wheel fallback. Three-size native ANSI/buffer tests, all TUI suites, Go/vet/race ×3, 29 helpers and 70/70 functional; ADR-0030.
- [x] Add opt-in regular-mode native terminal scrollback with five-row idle dock and no mouse/alternate-screen capture; stable final-output gates, fully expanded retained bands, multiline binding/layout fix and resize-geometry marker. Three-size ordered history/native copy/selection/draft/cursor/resize/session/exit/reopen tests, all TUI suites, Go/vet/race ×3, 29 helpers and 70/70 functional; ADR-0031.
- [x] Add fullscreen rendered-row search in a temporary independent editor, literal/case-insensitive/Unicode matching, highlight/next/previous, Escape restoration, and rendered user-prompt jumps. Three-size live/tool/resize/reopen/no-submit/zero-idle-row acceptance, all TUI suites, Go/vet/race ×3, 29 helpers and 70/70 functional; ADR-0032.
- [x] Correct Pi transcript spacing: user top/bottom padding, assistant leading blank, tool external blank plus colored padding; group Markdown continuation rows and update prompt offsets. Three-size buffer/PTY/scroll/search suites, all existing TUI suites, Go/vet/race ×3, 29 helpers and 70/70 functional; ADR-0033. Idle editor/footer unchanged.
- [x] Complete fullscreen padded-cell selection/copy/held-edge scroll with clipboard opt-out/OSC52 cap/native serialization, stale generation/layout guards, editor preservation and repaired plain tool click. Three-size native SGR/OSC52, all TUI suites, Go/vet/race ×3, 29 helpers and 70/70 functional; ADR-0034.
- [ ] Complete per-occurrence/cross-wrap search, pointer link precedence/word selection and mutation-stable anchors; light theme and retained-output/reflow/oversized-editor limits also need evidence.
- [ ] Complete remaining reconnect ownership and full browser compaction acceptance.
- [x] Implement per-session terminal editor/history state, generation-owned event/submit delivery and cancel-safe forwarding; unit/race coverage.
- [x] Bound the temporary terminal session selector to six results; live tmux verifies cancel/resize/zero added idle rows at 60×18, 100×22 and 140×36 (`make test-tui-sessions`).
- [ ] Complete remaining terminal adaptations in `docs/internal/web-session-parity.md` (mutation submenus/media/queue acceptance still separate).
- [ ] Continue component refreshes from a pinned upstream source where the suite demonstrates a gap.

Evidence and commands: [`../../tests/ux/README.md`](../../tests/ux/README.md). These results cover the mapped scenarios, not full Piclaw compliance.

## Phase 1 — minimal vertical slice

### Turn engine
- [x] Define append-only turn event model
- [x] Define turn state reconstruction from events
- [x] Persist turn start/progress/end events
- [x] Serialize turns per session
- [x] Allow concurrent turns across sessions
- [x] Implement queue/cancel state model
- [x] Implement cancellation state transitions (`running` → `cancelling` → `cancelled`)
- [x] Wire go-ai inference into turn engine
- [x] Load system prompt from workspace AGENTS.md
- [x] Stream inference tokens via go-ai `Stream()`
- [x] Broadcast Piclaw-compatible SSE events during inference
- [ ] Implement queue reorder
- [ ] Add centralized runtime budget config (tool calls per turn, retries, queue depth)

### Database/state model
- [x] Create SQLite schema baseline with JSON fields and expression indexes
- [x] Add sessions table
- [x] Add messages/content block tables
- [x] Add turn events/checkpoints tables
- [x] Add turns table with status tracking
- [x] Enable WAL and concurrency-safe pragmas
- [x] Auto-create default session on startup
- [ ] Add forks/ancestry table(s)
- [ ] Add schedules/background tasks tables
- [ ] Add attachments metadata tables
- [ ] Add settings/assets mirror tables

### Inference
- [x] Integrate go-ai as model/provider layer
- [x] Register builtin models (openai, openairesponses, anthropic)
- [x] Load auth tokens from `~/.pi/agent/auth.json`
- [x] GitHub Copilot token exchange (refresh → session token + enterprise endpoint detection)
- [x] Apply Copilot headers for IDE auth
- [x] Build conversation history from session messages
- [x] Streaming inference via `goai.Stream()`
- [x] Broadcast text deltas during streaming
- [ ] Token/cost tracking per turn
- [ ] Provider failover / retry with exponential backoff
- [ ] Context window overflow detection and compaction

### Web UI
- [x] Piclaw TypeScript source ported verbatim (199 files)
- [x] gi-specific `api.ts` adapter (same function signatures, maps to gi REST endpoints)
- [x] gi-specific `app.ts` entry point (uses Piclaw components verbatim)
- [x] Piclaw CSS stack served from `/css/styles.css`
- [x] Piclaw fonts vendored and embedded
- [x] Vendor bundles: preact-htm, marked, katex, beautiful-mermaid, codemirror
- [x] Load vendor scripts as ESM modules (scoped var declarations)
- [x] IIFE-wrapped app bundle to prevent global shadowing
- [x] Cache busters on all bundle URLs per server restart
- [x] `window.onerror` error boundary with backend log reporting
- [x] Import map for `#editor-vendor/codemirror`
- [x] Piclaw theme/tint runtime (presets, cycling, localStorage, system dark mode)
- [x] Timeline with Piclaw Post component (markdown rendering, avatars, file pills)
- [x] ComposeBox with Piclaw component (history, keyboard, slash commands, model picker)
- [x] AgentStatus with Piclaw component (draft/thought/plan panels)
- [x] WorkspaceExplorer with Piclaw component
- [x] TabStrip with Piclaw component
- [x] SSE streaming endpoint (`/sse/stream`) with Piclaw event model
- [x] Real SSEClient implementation (reconnection, heartbeat, event bindings)
- [x] Route-event-aware SSE subscriptions for `routing_decision`/`routing_incoming`
- [x] Frontend log bridge (`/api/frontend/log`)
- [x] Runtime config API (`/api/runtime/config`)
- [x] Workspace tree/file APIs (`/api/workspace/tree`, `/api/workspace/file`)
- [x] Route-event introspection API (`/api/sessions/{id}/route-events`) in frontend adapter
- [x] SSE-driven real-time timeline updates (wired but not yet consuming events in app.ts)
- [x] Streaming draft display in compose area

### Slash commands

The ComposeBox already has Piclaw's slash command autocomplete UI. The backend needs to handle them.

### System meters HUD

Direct port of Piclaw's `/meters` functionality. On by default until slash commands are implemented.

- [ ] Add `/api/system-metrics` endpoint (CPU, RAM, swap, RSS)
- [ ] Wire `SystemMetersHud` component in `app.ts` (already copied from Piclaw)
- [ ] Enable meters on by default (skip `/meters` toggle until slash commands exist)
- [ ] Poll metrics every 5s from the Go backend
- [ ] SSE `ui_meters` event for real-time updates

#### Session/model commands (Phase 1)
- [ ] `/model` — list available models or switch model
- [ ] `/cycle-model` — cycle to next available model
- [ ] `/thinking` — show or set thinking/effort level
- [ ] `/cycle-thinking` — cycle thinking level
- [ ] `/theme` — set UI theme
- [ ] `/tint` — tint default light/dark UI
- [ ] `/abort` — abort current response
- [ ] `/state` — show current session state
- [ ] `/stats` — show session token and cost stats
- [ ] `/context` — show context window usage
- [ ] `/last` — show last assistant response
- [ ] `/commands` — list available commands

#### Queue/steering commands (Phase 1)
- [ ] `/queue` — queue a follow-up message
- [ ] `/steer` — steer the current response
- [ ] `/abort-retry` — abort retry backoff

#### Session management commands (Phase 2)
- [ ] `/new-session` — start a new session
- [ ] `/session-name` — set or show the session name
- [ ] `/compact` — manually compact the session
- [ ] `/auto-compact` — toggle auto-compaction
- [ ] `/auto-retry` — toggle auto-retry
- [ ] `/fork` — fork from a previous message
- [ ] `/clone` — duplicate current branch into a new session
- [ ] `/tree` — list the session tree

#### Identity commands (Phase 2)
- [ ] `/agent-name` — set or show agent display name
- [ ] `/agent-avatar` — set or show agent avatar URL
- [ ] `/user-name` — set or show user display name
- [ ] `/user-avatar` — set or show user avatar URL

#### Tool commands (Phase 2)
- [ ] `/shell` — run a shell command and return output
- [ ] `/bash` — run a shell command and add output to context
- [ ] `/search` — search notes and skills in workspace
- [ ] `/skill:` — run a workspace skill

#### Auth/admin commands (Phase 3)
- [ ] `/login` — login to an AI model provider
- [ ] `/logout` — logout from a provider
- [ ] `/passkey` — manage passkeys
- [ ] `/totp` — show TOTP enrolment QR code
- [ ] `/restart` — restart the agent
- [ ] `/exit` — exit the process
- [ ] `/export-html` — export session to HTML
- [ ] `/tasks` — list scheduled tasks

### Pi/Piclaw config compatibility
- [x] Load `.piclaw/config.json` (assistant name/avatar, user name/avatar/background)
- [x] Load `.pi/settings.json` (provider, model, thinking level, enabled models)
- [x] Load `AGENTS.md` as system prompt
- [x] Load auth from `~/.pi/agent/auth.json`
- [x] Preserve Pi model/provider naming semantics

### Scripting engines
- [x] Goja (`js`) runtime implemented
- [x] Joker runtime implemented and bridge parity expanded
- [x] QuickJS deliberately out of scope (Goja is the supported JavaScript runtime)

### Testing
- [x] Go unit tests (store, turn, web, config)
- [x] Playwright base UX tests (13 tests)
- [x] Isolated test instance with dedicated DB and workspace (`make test-ux`)
- [x] Hook TDZ checker (`scripts/check-hook-tdz.ts`)
- [x] `go vet` / `go test ./...`
- [ ] CI pipeline (`.github/workflows/`)

### Developer tooling
- [x] Makefile with detached lifecycle management
- [x] `-bind`, `-port`, `-model`, `-log-file`, `-pid-file` CLI flags
- [x] `make help/bootstrap/deps/start/stop/restart/status/logs/run/clean`
- [x] `make test-ux` (isolated Playwright test instance)
- [x] `make build-web` / `make bun-checks` / `make check`
- [x] Bun build pipeline for vendor + app bundles

---

## Phase 2 — context, keychain, skills

### Context and recovery
- [ ] Implement compacted context rebuild from DB
- [ ] Carry forward last N messages on rollover
- [ ] Make N configurable
- [ ] Implement mid-turn compaction as same-turn recovery phase
- [ ] Emit UI hint for every compaction
- [ ] Checkpoint at compaction boundaries
- [ ] Implement provider retry checkpoints
- [ ] Preserve partial output on provider failure
- [ ] Auto-resume after clean restart
- [ ] No auto-resume after crash

### Keychain
- [ ] Port Piclaw encryption model
- [ ] Store encrypted secrets in SQLite
- [ ] Implement env/bootstrap unlock path
- [ ] Implement interactive unlock path
- [ ] Start in degraded mode when locked
- [ ] Emit unlock UI prompt when secret is needed
- [ ] Auto-resume blocked turn/tool after unlock
- [ ] Keep unlocked for process lifetime
- [ ] Support per-process unlock state

### Skills and hooks
- [ ] Define `SKILL.md` + frontmatter loader
- [ ] Implement SQLite managed VFS for skills/scripts/templates/packages
- [ ] Load skills from managed VFS first, then filesystem fallback where appropriate
- [ ] Migrate loose skill/script/template mirrors into the managed VFS
- [ ] Implement Joker-first hooks
- [ ] Add Go hook surface
- [ ] Implement required hook points
- [ ] Add packaged skill import baseline
- [ ] Investigate simple GitHub-sourced skill extraction flow

### Internal reference and self-extension docs
- [x] Create `docs/internal/` as the source tree for shipped internal reference docs
- [x] Add bootstrap docs for tools, scripting, VFS, and skills
- [ ] Document every built-in tool contract under `docs/internal/tools/`
- [ ] Document hook lifecycle and extension contracts under `docs/internal/hooks/`
- [ ] Embed `docs/internal/` as read-only runtime reference content (for example `vfs://reference/...`)
- [ ] Add a doc-maintenance check so new internal surfaces cannot land undocumented

---

## Phase 3 — UI parity and operational surfaces

### Web UI
- [x] Workspace browser (basic tree + file open)
- [x] Tab strip for open files
- [ ] Editor panes with CodeMirror (read-only pane exists, needs full editing)
- [ ] Diff view
- [ ] Split panes
- [x] File pills in timeline
- [ ] Interactive widgets
- [ ] Inline charting
- [ ] Search UI
- [ ] Schedule management commands
- [ ] Messages inspection UI

### TUI
- [ ] Evaluate `go-tui` fit/gaps
- [ ] Basic chat surface
- [ ] Good scrollback
- [ ] Expandable input
- [ ] Status/progress parity subset
- [ ] Forms support
- [ ] Advanced-terminal image preview path
- [ ] Session commands/search/schedules

### CLI
- [ ] Prompt submission
- [ ] Create/list/resume sessions
- [ ] Schedule commands
- [ ] Search/messages inspection
- [ ] Keychain management
- [ ] Direct script execution
- [ ] Maintenance/admin commands
- [ ] Prune state commands

---

## Phase 4 — storage, search, artifacts, observability

### Search and indexing
- [x] Decide hybrid workspace search direction: SQLite FTS5 + sqlite-vec + gte-go
- [x] Add ADR for workspace hybrid search (`adr/0008-workspace-hybrid-search.md`)
- [x] Add internal design doc for hybrid search/indexing (`docs/internal/search/README.md`)
- [x] Scaffold `internal/search/` package layout (service, rank, query, chunking, embed, vector, store, indexer)
- [x] Scaffold workspace search schema definitions (`workspace_documents`, `workspace_chunks`, `workspace_chunks_fts`, `workspace_index_meta`)
- [ ] Wire workspace search schema into main DB initialization
- [ ] Implement real `gte-go` embedder
- [ ] Implement real `sqlite-vec` backend
- [ ] Implement SQLite FTS strategy for workspace chunks
- [ ] Index messages
- [ ] Index notes/memory
- [ ] Index skills/scripts/templates/assets
- [ ] Index attachment filename/metadata/text
- [ ] Expose search via tool + CLI + UI

### Artifacts/attachments
- [ ] Define workspace-path artifact model
- [ ] Define selected DB-blob rules for images/generated outputs
- [ ] Implement attachment import/export/read flows
- [ ] Implement thumbnail generation table and cleanup
- [ ] Investigate image processing library choice
- [ ] Investigate Piclaw-compatible artifact metadata/FTS behavior

### Observability
- [ ] Define structured event schema
- [ ] Add counters tables/materialized aggregates as needed
- [ ] Record provider/tool/script/schedule/session events
- [ ] Add token/cost usage events
- [ ] Add compaction/rotation/recovery events
- [ ] Expose status views in UI/CLI

---

## Phase 5 — compatibility and hardening

### Pi/Piclaw compatibility
- [x] Preserve settings semantics
- [x] Preserve model/provider naming
- [ ] Preserve prompt template semantics
- [x] Preserve structured message/content model (via Piclaw components)
- [ ] Preserve intents/queue/steer conventions
- [ ] Preserve keychain env injection semantics

### Testing
- [x] Unit tests (Go)
- [ ] Fake-server provider tests
- [x] DB integration tests
- [ ] Golden tests for rendering/message projection
- [x] E2E web smoke tests (Playwright)
- [ ] E2E CLI smoke tests
- [x] E2E TUI smoke tests (tmux harness)
- [x] `go vet`
- [ ] Fuzzing

---

## Explicit investigation items

- [ ] Investigate exact Piclaw transcript/message export path for future import tools
- [ ] Investigate Piclaw artifact pill/open-editor behavior for parity
- [ ] Investigate Piclaw compaction and rollover implementation details for faithful compatibility
- [ ] Investigate `go-tui` capabilities vs required parity surface
- [ ] Investigate image processing library options for thumbnails/previews
- [ ] Investigate simple packaged-skill download/extract flow from GitHub URLs
- [x] Investigate embedded web asset pipeline with plain JS + Bun bundling only at build time
- [x] Port Piclaw TypeScript web source verbatim
- [x] Fullscreen selection includes the final content cell in forward/reverse drags; scrollbar-origin presses/drags stay with go-tui. Snapshot the text width separately from viewport width (which includes the scrollbar), keep interior half-open endpoints and linked-click hit testing, and leave regular terminal-owned selection unchanged. `make test-terminal-links` passes race x3 across ASCII/wide/combining suffixes, both scrollbar states and three widths; six `make test-tui-selection-edge BIN_DIR=/tmp/gi-media-bin` PTYs independently reset/check OSC52 for both directions, test scrollbar-origin drag exclusion, preserve draft/cursor/resize and add no idle rows. Fifteen selection/search/link/regular regression PTYs, Go/vet/Bun checks, 98 standard functional tests (two seed-dependent skips) and 96 helpers/1,175 assertions pass. Browser mappings stay 96/236 Classic + 30/42 shared. The final cell is an outer-boundary snap; precise last-cell-only selection, word selection and wrapped/complex links are separate gaps.
- [x] Verify non-iPad native image click/tap opens only the supplied lightbox in six browser projects; trusted touch, actual device identity, exact IndexedDB draft/media, stored image bytes, session/messages and reload survive without API writes. Independent review keeps frozen timeline-006 unmapped: Gi has no positive iPad annotator path to prove its device-gating clause. Sixty lightbox/rendering browser tests, 98 standard functional tests (two seed-dependent skips), 96 helpers/1,175 assertions and Go/vet/Bun checks pass. Concurrent build collision caused discarded blank-page failures; clean sequential reruns passed. Test-only, no restart or new TUI controls/credit; coverage stays 96/236 Classic + 30/42 shared.
- [x] Add per-occurrence fullscreen rendered-row search: precise whole-grapheme highlights, same-row next/previous wrap, unchanged-row occurrence retention after append and original style/link preservation. Three native search PTYs navigate 96 occurrences and check draft/cursor/resize/reader restoration; 18 edge-selection/selection/link/regular regression PTYs, focused race x3, 98 functional tests (two seed-dependent skips), 96 helpers/1,175 assertions and Go/vet/Bun checks pass. Existing search input/label only, no new idle rows; regular terminal-owned search unchanged. Match bytes map through Unicode lowercase to rendered cells; row-padding exclusion is unchanged. Cross-wrap, full case folding/normalisation and mutation-stable reflow anchors remain separate. Browser coverage stays 96/236 Classic + 30/42 shared.
- [x] Port/map frozen extra012 protected-recovery display validation from Piclaw 70d33bc93 through a guarded Post build adapter, with independent primary-failure validation for legacy blocks. Twenty-three native stored variants prove valid suppression/malformed visibility, native search responses, reload, exact draft/media bytes, unchanged session/history and zero API writes across six projects. Sixty recovery/copy/delete + 24 recovery/speech browser cases, seeded functional test, 98 standard functional tests (three fixture-dependent skips), 98 helpers/1,243 assertions and Go/vet/Bun checks pass; review finds no remaining blockers. Display-only: no execution authority, storage deletion or TUI credit. Coverage 97/236 Classic + 30/42 shared; extra013 stays unmapped.
- [x] Ship/map extra013 empty informational recovery-placeholder suppression using frozen typed outcome metadata and existing Post renderability; retain text/whitespace, warnings, media, rendered cards/fallbacks, submissions, references, resources and annotations. Twenty-one native variants pass across six projects with exact draft/media/session/history preservation. Restore pinned Adaptive Cards 3.0.6 SDK+MIT licence at the supplied lazy URL (GET/HEAD/hash test); rendering only, no action-mutation credit. Thirty recovery/speech + 102 recovery/copy/delete/lightbox browser tests, two seeded functional tests, 98 standard functional tests (four fixture-dependent skips), 100 helpers/1,293 assertions and Go/vet/Bun checks pass. Coverage98/236 Classic+30/42 shared, unmapped138/12; no TUI controls or idle rows added.
- [x] Reject unsupported browser Adaptive Card Submit asynchronously instead of resolving a false null-success; map frozen extra003 through real SDK pointer/keyboard retry, error notice/busy settlement, retained card inputs and exact composer/media bytes, unchanged native history/session/other session, zero network mutation and no success receipt. Thirty card/recovery/speech + 66 card/recovery/copy/delete browser cases, two seeded functional tests, 98 standard functional tests (five fixture-dependent skips), 101 helpers/1,307 assertions and Go/vet/Bun checks pass; independent review finds no blocker. Coverage99/236 Classic+30/42 shared, unmapped137/12. No submission backend/identity-validation credit or terminal controls.
- [x] Repair Windows CI compilation by isolating Unix process-group setup/termination from portable shell runtime and removing the turn engine's duplicate Unix-only kill. Windows uses bounded SystemRoot taskkill tree termination with Process.Kill fallback; cancellation closes output readers to avoid inherited-handle hangs, normal completion still drains output before Wait. Full Go tests/vet, shell cancellation/drain race x3, and five Linux/macOS/Windows cross-build targets pass. This restores Windows compilation; shell-backed execution still requires sh on PATH, not PowerShell.
- [x] Add compact Alt-S Right-arrow Pin/Unpin and Archive/Restore submenu, at most two actions with a human-readable target. Native capabilities hide archive for main/busy/claimed sessions; MutateSession rechecks the write. Captured target/source ownership, concurrent archive/queue rejection, storage-failure feedback, refreshed labels and Escape picker return preserve drafts/cursors/reader state. Six action + six session + six model + three regular PTYs pass at 60x18/100x22/140x36, including tall-list submenu cleanup and zero idle growth. Focused race x3, Go/vet/Bun, 98 functional tests, 101 helpers/1,307 assertions and all five cross-builds pass. Private tmux sockets and bounded resize waits fix test isolation/timing races. Rename and durable queue actions remain separate; browser coverage99/236+30/42 unchanged.
- [x] Stabilise CI's active-steering shell test after run36048713987 exposed first-turn completion before the second submission: private DB, real setup gate, bounded startup wait and runner-lock cleanup join replace timing assumptions. Same-turn/depth/two-completed-turn/no-claim assertions unchanged. Full Go/vet and50race repetitions pass; CI now includes the repeat target.
- [x] Add Rename to Alt-S actions using a four-row transient single-line editor and separate composer-safe buffer; native title validation, captured ownership, invalid/storage/archive-race retry, Escape/Ctrl-C cancel, editing/undo and mouse isolation. Rename is absent for archived sessions; submenu now has at most three actions. Six fullscreen/regular PTYs prove Unicode save/cancel/empty retry/160-rune horizontal window/resize/draft cursor/no idle growth, plus15picker-model-regular PTYs. Focused race x3,98functional/101helpers1307assertions/Go-vet-Bun/five cross-builds and bounded review pass. Existing rune-based editing retained (no new grapheme-motion claim); browser99/236+30/42 unchanged.

### Single-user browser sign-in
- [x] Gate application bootstrap on validated native auth policy; hide credentials on loading/failure and provide Retry. Remove the static placeholder before the client-only mount; install Settings opening controls before first paint.
- [x] Code-only TOTP via separate JSON `POST /api/auth/session`; host-only HttpOnly/SameSite Strict cookie, direct TLS-or-loopback-peer-and-Host, same-origin reads/writes and private no-store responses. Existing bearer/query API precedence is unchanged; browser responses never expose the token.
- [x] Verify native API/SSE admission, invalid/stale credentials, malformed policy, false-success cookie rejection, pending state, retry, reload and draft/media preservation: 18 auth + 264 regression browser cases across six projects; 99 functional / 5 skipped; 103 helpers / 1321 assertions; full Go/vet/hook, web/auth race ×3 and five cross-builds. Map only auth-002/004 (101/236 Classic, 30/42 shared).
- [x] Document direct TLS/localhost-only browser transport, no implicit proxy-header trust, no ongoing-stream expiry guarantee, no new terminal UI; family/passkey/enrolment/logout UI remain separate.
- [ ] Implement the separately pinned Piclaw multiple-passkey Settings addition (`b531ea3a8c`, 26 scenarios/outlines). Native WebAuthn, recent proof and atomic lockout-safe credential management are absent; upstream tags are not Gi credit.

### Native authentication persistence
- [x] Serialize enrolment/login/internal token revocation across managers/processes with `auth.lock`; stale enrolment candidates cannot overwrite an enrolled account. Conflicting HTTP mutations return 409; no token/cookie is issued before successful commit.
- [x] Replace snapshots with bounded private, synced temp files; preserve unknown top-level fields; reject unsafe paths, corrupt/oversized state and commit conflicts. Linux/macOS sync both parent and state directories; Windows uses Go's rooted replacement without a sudden-power-loss durability guarantee.
- [x] Test 24 concurrent logins plus revoke/read, four child writers, killed-lock release, single-winner enrolment, final-token revocation, unknown-field preservation, symlink and pathname swaps, and unchanged failed-write bytes. Auth race ×10, web/auth race ×3, 18 browser regressions, 99 functional / 5 skipped, 103 helpers / 1321 assertions, Go/vet/hook and five cross-builds pass. Native three-OS CI evidence follows.
- [ ] Browser logout/passkey credential removal still require their own API/UI and acceptance work; this internal revocation primitive grants no frozen parity credit or terminal rows.

### User-directed full UX audit and remediation (2026-09-24)
- [x] Review all 39 browser specs, 17 functional specs, 47 helper tests, 17 TUI test files, 20 terminal scripts and 37 feature files plus harnesses; record 200-file inventory and ranked evidence gaps in `docs/internal/ux-test-audit-2026-09-24.{md,csv}`. Review is source-level, not an all-suite rerun.
- [x] Compare deployed Gi/Piclaw composer and both pickers with mutation-blocked live probes; confirm responsive geometry/component differences and Escape-to-composer mismatch. Keep exact Return-to-start and workspace transition complaints open rather than disproving them with old tests.
- [ ] Add and enforce genuine clean-start/new-chat/Return, composer/slash/picker/Quick Actions and Settings user journeys; remove vacuous functional assertions.
- [ ] Reconcile component+CSS+host provenance with current Piclaw and add controlled visual/geometry oracles before claiming look-and-feel parity.
- [ ] Resolve Classic008/shared17 skill-prefill conflict; current Classic mapping is disputed, not newly verified.
- [ ] Wire every specialised suite/flag into an explicit all-suite manifest with provenance and skip accounting; establish CI UX gates.
- [ ] Reproduce and fix workspace-tab direction/transition semantics, touch and keyboard focus against the reference.
- [ ] Auth persistence `56079fe` remains undeployed: native macOS/Windows CI `36063764465` failed. User-directed UI remediation takes priority; do not deploy HEAD until failures are resolved or the slice is safely isolated.

### Workspace collapse motion (2026-09-25)
- [x] Reproduce the inherited desktop close defect with real animation frames: at 1440px the sidebar moves from x=0 to x=128 then x=262 while shrinking. Add a Gi-only CSS override that anchors the flex row, keeps a constant chat flex basis, animates chat margins and follows the sidebar edge with the toggle. Supplied components and CSS remain unchanged; drawer/editor layouts retain their rules.
- [x] Red→green motion test observes real pointer/keyboard close/open, in-flight reversal, synchronised geometry, reduced motion, exact draft/media bytes and zero admissions in Chromium/WebKit at 1024/1440/1920. Restore original phone/tablet/desktop viewport for drawer regression. 204 workspace/shell/Settings browser cases, 100 functional tests (five existing skips), Go/vet/hook checks pass; tightened frame synchronisation rerun passes six cases. Discard the earlier tool-timed-out partial run. Independent read-only review found no blockers; user-resized sidebar widths remain untested (host has no wired splitter hook).
- [ ] Deploy only after native auth CI failures are resolved. Retrieved CI36063764465 logs: Windows cross-process lock exclusivity test failed; macOS concurrent enrolment failed opening `auth.lock` with ENOENT. No auth changes in this CSS slice, no new frozen mappings, no terminal changes or idle rows.

### Native auth CI repair (2026-09-25)
- [x] Replace the test helper's empty `select {}` with a timer-backed wait so Go's deadlock detector cannot terminate the lock holder. Capture child output and require ten consecutive conflict observations before crash/release verification.
- [x] Separate exclusive first lock creation from existing-lock open; treat a disappearing lock pathname as a retryable conflict while retaining inode/type/root checks. Add sixteen concurrent first opens with same-inode and independent-handle exclusivity checks. Local auth race x10, web/auth race x3, Go/vet/hook, five cross-builds, eighteen auth browser tests and100 functional/five existing skips pass.
- [ ] Native macOS/Windows CI must validate the repair before deployment. Cross-builds and Linux results cannot establish native filesystem behaviour.
- [x] Native CI36071594429 for `6bc5c08` passed Linux/macOS/Windows auth-state jobs, main Test and all five builds. Deployed workspace repair `f33b466` plus auth repair `6bc5c08` via `make restart BIN_DIR=/tmp/gi-scheduler-bin BIND=0.0.0.0 PORT=8090`; PID2404942. Six guarded live motion probes1024/1440/1920×Chromium/WebKit and six phone/tablet/desktop status/draft checks passed with zero API writes/browser errors. Auth remains unenrolled; database integrity ok,62sessions/51turns/146messages, before/after dump hashes identical excluding runtime lease renewal only (`792e8258…ded00`). No new frozen or TUI credit.

### Public feature and parity documentation (2026-09-25)
- [x] Update README, docs indexes and `docs/feature-parity.md` with shipped browser/runtime features, bounded pi-tui adaptations, known workflow gaps and planned multi-passkey/tsnet/Iroh/MCP work. Replace blanket Piclaw/source compatibility claims; correct Go prerequisites and document unenrolled-instance exposure and Make binding defaults.
- [x] Mark May fit-gap and earlier session reports as historical; distinguish101/236 Classic +30/42 shared source mappings, disputedClassic008 and26 unimplemented additive passkey scenarios. Validate118 local Markdown link targets, frozen hashes,103 helper tests/1321 assertions and diff whitespace; independent read-only review found no blockers. Documentation-only, no runtime changes or deployment.

### Working multi-passkey enrolment requirement (2026-09-25)
- [x] Record the owner's explicit multi-passkey requirement and concrete acceptance gate in `tests/ux/features/additions/piclaw-2026-09-24/README.md`; preserve the pinned26-scenario contract byte-for-byte. Current Gi remains TOTP-only, with no WebAuthn implementation or new parity credit.
- [ ] Implement native pure-Go registration/assertion verification, stable configured RP/origins, durable per-credential records and session-scoped recent proof; bearer automation tokens cannot authorise browser credential management.
- [ ] Prove two independent browser authenticators can enrol without replacement and each sign in after restart; passkey-only second enrolment needs no TOTP. Cover rename/remove/revoked-key rejection, duplicate/replay/session/origin failures, uncertain finish/cancellation and atomic concurrent last-factor protection.
- [ ] Add Settings and login browser journeys with real WebAuthn cryptography, exact draft/media preservation and labelled virtual-authenticator evidence; record physical/synced credential and native prompt focus evidence separately. No terminal idle rows.
- [ ] Confirm the stable production HTTPS hostname/RP before owner enrolment or remote exposure; no production auth mutations during development. tsnet/Iroh and token-saving MCP remain separate requested workstreams.

### Browser-owner proof prerequisite (2026-09-25)
- [x] Persist browser-owner purpose and TOTP proof time separately from bearer/legacy sessions; retain ordinary legacy access while rejecting management authority. Add cookie-only proof status and reauth routes with direct TLS/loopback, same-origin, duplicate-cookie and explicit-credential rejection. Reauth changes only the captured session's proof, never its token or expiry.
- [x] Add an in-transaction owner/freshness guard for future credential mutations, boundary/restart/legacy/factor/revoke/concurrent tests, and six-project two-context browser API integration with retained drafts/media. Auth race x10, web/auth race x3,24 browser cases,101 functional/five existing skips, Go/vet/hook and five cross-builds pass. Correct timestamp test to compare instants and token identity (equivalent RFC3339 spellings differ). Independent security review found no blockers. Document older-writer metadata downgrade/fresh login requirement.
- [ ] WebAuthn ceremonies, credential APIs and Settings controls are still unimplemented. No additive passkey mapping, production enrolment or terminal chrome from this prerequisite; native CI/deployment follows separately.
- [x] CI36075923177 for `c09f788` passed native auth on Linux/macOS/Windows, main Test and all five builds. Deployed on8090 PID2457971; six guarded Chromium/WebKit phone/tablet/desktop checks passed with zero API writes/errors. Unenrolled proof GET refuses401, auth stays unenrolled, integrity ok62sessions/51turns/146messages. Before/after normalized dumps match (`792e8258…ded00`), excluding only runtime lease renewals. No production login/enrolment or WebAuthn support claimed.

### Native WebAuthn multi-key backend (2026-09-25)
- [x] Add pure-Go go-webauthn v0.18.2 verification with explicit startup RP/origins, required UV, durable bounded one-use challenges bound to browser-owner or login ceremony cookie, persistent per-RP keys/user handle, and native add/list/assertion/reauth/rename/remove APIs. Policy/freshness/revocation/current-key/last-factor rechecks share the commit lock. Duplicate and consumed challenges never replace keys; configuration loss cannot reopen a passkey account.
- [x] Verify12 Chromium API cases at3sizes: two independent credentials, restart/fresh cookie login each, passkey-only further enrolment, rename material equality, removed-key rejection, last-factor refusal, expiry/replay/session/revoke/origin/wrong-RP valid signature, cancellation, duplicate-ID and lost-finish-response reconciliation. Native race×10 and atomic two-removal tests pass;24TOTP browser regressions,102functional/five existing skips,Go/vet/hook,web/auth race×3 and5crossbuild pass. Fix test authenticator snapshots to retain advanced counters rather than weakening clone rejection. Focused independent review found no blockers.
- [x] Add `make test-ux-passkeys` as a required CI job before all build jobs. Document opt-in API subset and source-config boundaries; all26complete Settings scenarios remain unmapped. No TUI changes or production enrolment.
- [ ] Native CI/build/browser gate and disabled-config live deployment checks follow. Settings/login controls, policy configuration UI, owner bootstrap, physical/synced device checks and native focus/cancellation UX are not implemented by this slice.
- [x] CI36079153337 for901d746 passed Test, native auth onLinux/macOS/Windows, new native passkey browser job and all5builds. Deployed8090 PID2507187 with RP config confirmed absent. Six guarded responsive Chromium/WebKit checks zero API writes/errors; GETpasskeys403, auth remains unenrolled. Integrity ok62sessions51turns146messages and normalized dump SHA256792e8258…ded00 unchanged excluding runtime_leases. No live enrolment or full Settings scenario credit.

### Multi-passkey Settings and login (2026-09-25)
- [x] Add server capability policy and passkey login with confirmed cookie authority; lazy Authentication pane lists keys/times/IDs and provides native add/rename/remove/confirm, TOTP/passkey fresh proof, cancellation and explicit refresh. Unsupported origins/browser/policy are explained; no live policy changes. One-use helper serialises browser credentials, separates rejection from uncertain registration/local-orphan caveats and never auto-retries. Escape stays pane-local while busy/editing, blur does not cancel, unmount aborts and stale callbacks are fenced; focus restored on cancellation. Supplied components unchanged, zero TUI rows.
- [x] Pass21realChromium API+UI cases at3sizes,30auth browser cases,354Settings/QuickActions,103functional/five skips,104helpers/1334assertions,authrace×10/web×3,Go/vet/hook/5crossbuild. Two keys added via Settings and each logs in after restart; draft/media retained; passkey-only reauth/add/login, rename/remove/cancel, failed reads/writes, lost success, native cancellation/retry/unmount tested. Fix stale5section/4chunk test constants to6sections/5panes without weakening lazy/cache checks. Independent lifecycle/security review found no blockers; applied abort/focus follow-ups and reran21cases.
- [ ] Full26scenario mapping, Visual skin, policy configuration UI, initial owner bootstrap, physical/synced authenticator behaviour and native prompt focus remain open. NativeCI and guarded disabled-config deployment follow; do not enrol production credentials.
- [x] CI36083298328 for446baad passed native3OS auth,passkey API/UI browser gate,Test and5builds.Deployed8090PID2591087 with RP/origins absent;6guarded live Settings Authentication checks pass Chromium/WebKit3sizes:no writes/errors,disabledAdd explanation,draft/close/overflow. Auth stays unenrolled;DBintegrity/counts62/51/146/normalized SHA256792e8258…ded00 unchanged except lease. No production credentials enrolled.

### Lockout-safe authentication policy (2026-09-25)
- [x] Add owner-cookie-only policy GET/CAS POST, current-policy fresh proof, random revision and atomic target-factor check under the same auth.lock as removal. Legacy empty policy/revision reads either/initial; writes preserve sessions/credentials. Current-RP configured passkey or verified accepted TOTP required, bearer/query/cross-origin/stale proofs rejected. TOTP-only Settings keeps proof/policy controls and configured-RP inventory visible with passkey mutations disabled.
- [x] Add explicit policy selection/confirmation/cancel/retry, current-policy reauth and lost-response refresh.27passkeyChromium cases,36auth6projects,216Settings,104functional/five skips,104helpers1334assertions,authrace×10/web×3,Go/vet/hook/5crossbuild pass. Native multi-manager race and browser changed-policy-during-removal proof; independent final review no blockers. Last inventory UI adjustment reran27+36cases.
- [x] Review all26additive scenarios in `docs/internal/passkey-scenario-review.md`:004/014/022/025 candidate complete automated journeys; others partial/manual/unsupported with specific gaps. No new source mapping or frozen credit. Initial-owner bootstrap/Visual/physical/synced/nativeprompt/productionRP remain open; no live auth mutation or terminal rows.
- [ ] Native CI and guarded disabled-config deployment follow before shipping policy UI.
- [x] CI36086022919 for71f43c3 passed Test,native3OSauth,passkeybrowser gate,5builds.Deployed8090PID2640270 after RP-config-absent check;6guarded Chromium/WebKit3size Authentication checks no writes/errors/draft retained;policyGET401/authunenrolled. DBintegrity/counts62/51/146/normalized hash792e8258…ded00 unchanged except runtimelease.No production policy/enrolment change.
