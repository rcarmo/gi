# Gi web UI — 2026-09-21

Gi has a Piclaw-derived Preact web shell backed by Go HTTP APIs, SSE and SQLite. Many copied Piclaw controls have no working Gi adapter yet. Backend capabilities and browser feature support differ substantially.

Later on 2026-09-21, the session slice added guarded selection, page-local per-session drafts and real child-chat creation. Its combined parity matrix passed 24/24 runs (3 frozen IDs plus a child-creation regression). A reproduced pooled-SQLite configuration defect was fixed in `bd30518`; the existing web suite then passed 70/70. The initial-run counts below are historical. See [web-session-parity.md](web-session-parity.md) for current limits and proposed terminal adaptations.

The subsequent searchable-picker slice (`013`) passed 36/36 matrix executions, bringing mapped coverage to 4/236 frozen IDs (232 unmapped). The full web suite passed 70/70 after fixing a reproduced submission-event ordering race. Scope is now explicitly the entire Classic corpus plus all 42 shared-contract cases; see [full-web-tui-parity-plan.md](full-web-tui-parity-plan.md). Terminal designs still need separate implementation evidence.

The native mutation slice (`015`) then added persisted rename/pin/archive/restore and failure-safe controls: 48/48 matrix runs, 5/236 frozen IDs mapped, 231 unmapped, and 70/70 functional web tests. Archive is reversible picker metadata, not deletion or an agent shutdown. [ADR-0009](../adr/0009-session-picker-mutations.md) records the native API and limits.

On 2026-09-22, browser-local IndexedDB drafts and captured-send recovery added four compose mappings: 102/102 matrix executions, 9/236 frozen IDs mapped, 227 unmapped, and 70/70 functional tests. Real uploaded bytes and file/message references survive reload. [ADR-0011](../adr/0011-browser-draft-recovery.md) records storage, cross-tab and uncertain-delivery limits. Terminal session isolation has separate evidence in [ADR-0010](../adr/0010-terminal-session-selection.md).

The following queue slice added explicit queued follow-ups, persistent reorder and queued-only cancellation (`018`): 120/120 matrix executions, 10/236 frozen IDs mapped, 226 unmapped, and 70/70 functional tests. Queue return/Steer and shared queue cases remain open. [ADR-0012](../adr/0012-queued-followup-order.md) records API and race handling.

Queue SSE reconciliation and disconnect cleanup (`016`, reconnect `001`) then passed 138/138 matrix executions: 12/236 Classic IDs mapped, 224 unmapped, and 70/70 functional tests. [ADR-0013](../adr/0013-queue-sse-reconciliation.md) covers connection/source generations, real-stream tests and remaining reconnect limits.

Authoritative session-local model selection and native `/model` commands (`021`, compaction `008`) passed 168/168 matrix executions: 14/236 Classic IDs, 222 unmapped, 70/70 functional tests. Measured context compatibility and shared model criteria remain open. See [ADR-0014](../adr/0014-session-model-selection.md).

## Implemented and wired

- Embedded JavaScript/CSS, identity/avatar configuration, themes and system meters.
- A default browser chat, persisted message timeline, Markdown/code rendering, compose input and prompt submission.
- Session-scoped SSE subscription and periodic timeline refresh.
- Backend session creation, forks, explicit peer routing, turn cancellation and event history; some of these are exposed through the browser's session controls.
- Provider/model listing and runtime model-selection API. This is not yet proven to preserve independent models per browser session.
- Workspace tree/file reads and file creation endpoints.
- HTTP authentication middleware, TOTP login and TLS/ACME server options. These concern access to Gi, not provider OAuth login.
- Model tools through `/api/tools/execute`, including read-only `vfs://reference` documentation.

## First parity correction

The frozen Piclaw Classic corpus used by Vibes/Tau is now at `tests/ux/features/classic/`. It contains 24 files, 236 tagged scenarios and 256 expanded cases. The separate shared Vibes/Tau contract is also copied byte-for-byte (42 expanded cases).

Gi previously had only the explorer's file-actions menu. The global TimelineMenu, language selector, recent-file and display-scale helpers were imported unchanged from Piclaw `70d33bc93ab540845bbcf5f80503ca8125c71594`. Source hashes and upstream MIT licence are in `web/upstream/`. The Gi adapter now mounts that menu and the Piclaw narrow-screen workspace drawer/backdrop. CSS excerpts retain source line references.

`@ux-original-001` and `002` pass in Chromium and WebKit at phone, tablet and desktop sizes: 12/12 executions. They check menu dismissal, workspace toggling, preserved draft/session and absence of accidental submission. They do not exercise every menu action. Settings, terminal/VNC, recent-file editor behaviour, other session actions and the shared contract still need their own mappings.

## Incomplete browser features

| Area | Current limitation |
|---|---|
| Multi-session/agent/chat | Backend identities, session/fork records and switch callbacks exist. The app centres on a hard-coded `web` agent and one `gi_session_id`. Creating another `@web` main session can resolve to the existing main session. Independent drafts, queues, model/context state and late-response rejection need implementation and browser evidence. |
| Session actions | Rename/archive/restore helpers return null; other session callbacks are no-ops. A visible control does not establish persistence. |
| Queue and steer | The queue adapter returns an empty list; reorder/steer handlers are stubs. Backend queue capabilities do not reach the browser stack. |
| Attachments | Media storage/API endpoints exist, but `uploadMedia` returns null and send helpers ignore media IDs. Browser attachment delivery is incomplete. |
| Workspace editor | Tab headers can open; the editor host is not connected to a working pane implementation. Rename/move/delete/upload/reindex/preview helpers are stubs. |
| Terminal/VNC | Component sources exist; app callbacks do not open working native sessions. |
| Timeline actions | Delete is a no-op; thread retrieval is empty; card actions and several widget callbacks are not connected. |
| Search | Search adapter filters loaded messages, but the app disables the search control. |
| Plan / Quick Actions | No complete native web integration or parity mappings yet. |
| Auth/settings | Named secret references resolve from injected environment variables. No browser keychain management or interactive provider OAuth flow. |
| Notifications/approvals | Browser push subscription and approval-response helpers are stubs. |

Principal evidence: `web/src/app.ts`, `web/src/api.ts`, `internal/web/{server,sse,workspace,tools,auth,runtime}.go` and `tests/ux/README.md`.

## Validation and next work

- Full Go suite passed after correcting a reproduced shell stdout-drain race; regression failed before the fix with 128 of 26,013 bytes and passes afterward. Commit `49c6cac`.
- `make vet` and `make bun-checks` passed.
- Frozen source hash/inventory tests passed.
- New mapped parity matrix: 12/12 passed; 2/236 scenarios mapped, 234 unmapped. No blanket retries, fallback submits or forced clicks.
- Existing `make test-ux`: 58/70 passed, 12 failed in chat/SSE/compose/turn cases. This suite is not green. Fixing pipe ordering alone did not resolve these browser/runtime failures.

Next priority: map session-picker cases `@ux-original-013`–`015`, then implement coherent multi-session state and failure-safe mutations. Queue/model/context/draft isolation and reconnect races must be checked through visible controls and native persistence. Do not infer full Piclaw compliance from the first shell cases.
