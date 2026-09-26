# Gi features and Piclaw parity

Updated: 2026-09-25. This describes the repository after the workspace-motion and
native auth persistence repairs. Multi-passkey APIs and Settings/login controls now have browser integration tests;
lockout-safe policy controls are implemented; physical devices, remote networking
and MCP have work outstanding.

Gi shares pinned Piclaw browser components and configuration conventions, but has
its own Go runtime, SQLite state and terminal UI. It is not a drop-in Piclaw
replacement. The browser and terminal share the turn engine; their interaction
coverage is tracked separately.

## Reading the status

| Status | Meaning |
|---|---|
| Implemented | Native behaviour exists with tests for the stated scope. This does not cover every related Piclaw workflow. |
| Partial | A usable subset exists; the row names the missing or differing behaviour. |
| Scaffold | Configuration or internal types exist without an end-to-end user workflow. |
| Planned | Requested behaviour has no working implementation in Gi. |

Historical slice reports contain their own test totals. Those runs are not one
combined, current full-suite pass. The [UX audit][audit] reviewed 200 files and
found missing user journeys, weak assertions and tests tied to Gi's older layout.

## Browser and runtime

| Area | Gi status and available behaviour | Piclaw parity limits / next work |
|---|---|---|
| Runtime and distribution | Implemented: one pure-Go binary, embedded browser assets, SQLite/WAL sessions, messages, turn events and recovery; `go-ai` inference. | Gi owns its runtime and storage. Piclaw extensions are not automatically compatible. Bun is build-time only. |
| Chat and streaming | Implemented: native prompt admission, SSE status/draft/thought updates, reconnect reconciliation, bounded timeline paging and scoped search. Startup/new-chat focus and explicit loading-retry repairs have an empty-store, six-project browser gate. | The [first-Return journeys](internal/startup-return-journeys.md) verify exact session/turn identity. Conversation-level shortcut ownership and current-Piclaw visual equivalence remain open. |
| Composer and drafts | Partial: persistent browser-local text/media/references, failed-send recovery, file/folder/message references, upload progress/cancel/retry and byte-identical upload reuse. | Composer inline padding now follows the pinned Classic reference; session-strip structure and full visual styling still differ. Native slash catalogue, Tab/Enter/Escape and keyboard ownership now have a six-project CI suite. Physical IME and prefill-policy conflicts remain open. |
| Sessions | Partial: selection, child creation, grouping/search/typeahead, capability-gated pin/rename/archive/restore and draft isolation. | Session picker outer bounds now follow the pinned Classic reference, including fixed mobile panels. Row structure/styles and complete session-management UI remain incomplete. |
| Models and context | Partial: session-local model selection, registry/context metadata, fit checks, usage meter and model commands. | Model picker outer geometry follows the pinned reference with an explicit intermediate-width containment correction. Catalogue structure and remaining workflows/local-estimate labelling are incomplete; unavailable metadata stays unknown. |
| Queue and Stop | Partial: durable browser follow-ups, reorder/cancel, run-bound steering, queue return-to-draft, reconciliation and run-bound Stop. | Web Stop now preserves pending work behind durable explicit Resume (local race/HTTP/reconnect gates pass); generic/TUI cancellation still advances. Shared-36 mapping, product CI and deployment review remain pending. |
| Compaction | Implemented: automatic/manual native compaction, persisted context checkpoints, progress/cancel and shared browser/terminal engine behaviour. | Broad Settings parity and every upstream compaction workflow are not complete. |
| Timeline and media | Partial: Markdown/tables/code copy, image lightbox, stored media/resource links, tool timing, recovered-response labels, idle single-message deletion and browser speech controls. | No iPad annotation workflow. Speech tests use a controlled browser boundary; physical audio is not verified. |
| Cards and widgets | Partial: supplied Adaptive Cards rendering and truthful rejection of unsupported Submit. | Accepted card actions and the full agent-authored widget/attachment tool surface are not implemented. |
| Workspace | Partial: rooted tree/hidden files, bounded previews, read-only tabs, retained preview/conversation switching, keyboard/touch tab navigation, MRU/pinning/context actions and explicit scoped lexical index/reindex. | No document editing, dirty-buffer save, popouts or docking. Automatic external-change freshness and vector search are not complete. |
| Workspace motion | Implemented: desktop collapse stays left-anchored; chat and toggle interpolate together, including rapid reversal and reduced motion. | Deliberate correction of an inherited Piclaw CSS defect. It does not close broader workspace-tab workflow gaps. |
| Settings | Partial: General identity, session Models, local Appearance, compaction controls/policy and guarded OpenAI/Anthropic API-key management. | Other/custom provider setup and browser OAuth flows, richer panes and remaining focus paths are incomplete. |
| Browser authentication | Partial: single-user TOTP sign-in, HttpOnly/Strict cookie, transport/origin checks, transactional auth persistence and browser-owner proof. Settings logout revokes only this browser, confirms native status and reconciles uncertain results without replay; drafts and other sessions stay intact. | Initial owner Settings setup uses a manual TOTP key with a ten-minute display lifetime, atomic owner/session creation and explicit uncertain-response recovery, restricted to loopback. QR/physical setup, broader session-management UI and family mode are not implemented. Opt-in WebAuthn APIs are tested separately. |
| Multiple passkeys | Partial: pure-Go WebAuthn registration/login/reauth APIs, RP-scoped storage, add/list/rename/remove with fresh proof and atomic last-factor protection. Real Chromium API tests verify independent credentials after restart and further passkey-only enrolment. | Settings add/list/rename/remove/reauth and login controls pass Chromium virtual-authenticator journeys. Policy UI now enforces fresh current-policy proof, revision checks and write-time usable factors. WebKit ceremonies, Visual skin and physical-device validation are outstanding. The [26-scenario review](internal/passkey-scenario-review.md) records remaining gaps; formal mappings are outstanding. |
| Skills, tools and scripting | Partial: native tools, embedded Joker/JavaScript bridges, process extensions/hooks, skills, managed VFS and browser skill commands. | No general Piclaw/Pi package or extension compatibility. Classic skill-prefill semantics are disputed against the shared contract. |
| Operator integrations | Partial backend routing/topics/inbound-work primitives. | The complete Piclaw plan, scheduled-task, dashboard, SSH/Proxmox/Portainer and remote-agent operator surfaces are not ported. |

An unenrolled instance permits application access. CLI binding defaults to
loopback; `make start` defaults to `0.0.0.0`. Use `BIND=127.0.0.1` until access and
authentication are configured. Serving HTTPS does not itself enrol the owner.

## Terminal adaptations

The terminal target is pi-tui's compact appearance and keyboard ownership, using
Go widgets. Web overlays do not become permanent terminal panels.

| Interaction | Implemented adaptation | Limits |
|---|---|---|
| Transcript | Fullscreen paging, tool folding, Pi-style message/tool bands, rendered search with occurrence navigation, validated single-paragraph soft-wrap matches and prompt jumps. | Cross-wrap for prewrapped Markdown/inline-code layouts, mutation-stable reflow/eviction anchors and light-theme parity need further work. |
| Native scrollback | Opt-in `-tui-mode regular`; terminal-owned history, selection and copy, with five idle editor/footer rows. | Fullscreen features are not all available in regular mode. Retained completion output follows documented retention limits. |
| Editor | Cursor-following visible window capped at 30% of terminal rows (five-line default minimum, clamped to available space), grapheme/cell wrapping and unchanged draft bytes. | Full-draft layout cost, richer grapheme editing and extremely short fixed-widget layouts need further work. |
| Session/model choice | `Alt-S` / `Alt-M`, at most six visible results, native selection and draft/cursor preservation. Alt-M adds known context inline only when the full key and suffix fit; blocked reasons take priority. | Regular-mode selectors use a temporary alternate screen. Remaining picker differences are tracked independently of browser tests. |
| Session actions | Bounded Pin/Unpin, Archive/Restore and Rename controls, including native capability checks. | No permanent action panel and no full session-management parity. |
| Queue | On-demand `/queue` pages six durable rows; guarded removal and Steer by IDs, plus before/after moves against the last current-visit snapshot. | No retry UI or persisted queued text drafts; Stop-policy differences remain. |
| Compaction/index | `Alt-C` and `/compact`; `Alt-I` transient explicit index actions. | No background index-progress panel or automatic freshness guarantee. |
| Files | Durable session-local pending media refs, `/attachments`, reference-only `/detach`, native stored-byte admission and held unresolved claims after restart. | Plain-text/queue-draft restart recovery, automatic replay of ambiguous claims and general clipboard-image/drag-drop parity are not provided. |
| Copy and links | `/copy` for latest assistant source, fullscreen drag/copy/edge scroll, word/line multiclick for eligible text, opt-in native/OSC52 clipboard, safe OSC8 targets for supported complete parenthesised links, including long tokens wrapped by the terminal renderer. | Host clipboard/URL activation depends on the terminal. Complex inline-code/table links, physical terminal activation, cross-wrap word selection and locale-specific segmentation parity are incomplete. |
| Authentication | Local terminal access uses the local runtime; no idle login panel. | Passkey management must use authenticated browser Settings. No terminal emulation of WebAuthn or bypass of recent proof. |

PTY tests cover fullscreen/regular cases at 60x18, 100x22 and 140x36, including
resize, cursor/draft preservation and idle footprint. Protocol-byte checks for
OSC52/OSC8 do not prove every terminal emulator's behaviour. See the [TUI plan][tui]
and [clipboard/media contract][media].

## Requested integrations

| Workstream | Current state | Acceptance target |
|---|---|---|
| tsnet remote access | Scaffold: `internal/peering` wraps `tailscale.com/tsnet` and exposes status. No runtime start/listener wiring provides remote UI access. | Opt-in tailnet HTTPS for the existing UI/API/SSE, persistent node state, secret references, joined shutdown and existing app authentication. No public exposure/Funnel by default. |
| Iroh inter-instance chat | Planned; no Iroh transport in Gi. `tmc/go-iroh` is a candidate requiring version-pinned interoperability tests. | Explicit pairing, receiver-owned policies, signed bounded messages/files, epochs/revocation, durable retries/deduplication and one-hop peer/agent addresses. Gi-to-Piclaw interoperability scope still needs confirmation. |
| Token-saving MCP | Planned; no MCP client/gateway in Gi. The official Go SDK is the first transport candidate. | One model-visible gateway, compact paginated search/list, explicit schema describe, lazy stdio/Streamable HTTP connections, auth/config-aware metadata cache and bounded recoverable outputs. Measure actual model request size and total discovery/call cost. |
| Multi-passkey management | Partial: native backend plus Settings/login controls; [contract](internal/passkeys.md). Lockout-safe policy controls exist; initial owner bootstrap and physical-device validation are outstanding. | Native Go WebAuthn verification, session-bound five-minute proof, one-use challenges, two-key independent sign-in after restart and concurrent last-key protection. Virtual and physical authenticator results reported separately. |

These are separate integrations. tsnet carries operator web access; Iroh carries
peer messages; MCP connects tools. None may silently grant the authority of
another. The runtime requirement is pure Go, without CGO, native Rust libraries
or sidecars. An explicitly configured external MCP stdio server may use its own
runtime.

Piclaw references are its remote-peer add-on 0.3.4 and shipping
`pi-mcp-adapter` 2.15.0. Gi's MCP design intentionally omits schemas from search by
default and avoids the adapter's first-cache connect-all behaviour. OAuth,
resources/prompts, Apps, sampling and elicitation need separate scope and tests;
implementing `tools/call` alone will not establish full adapter parity.

The [browser-owner proof prerequisite](internal/browser-auth-proof.md) distinguishes
browser sessions from bearer/legacy tokens and refreshes five-minute TOTP proof
for one session without extending login expiry. The native passkey backend now
and Settings controls use this authority boundary.

Choose the production HTTPS hostname/RP before enrolling real passkeys. A key
registered for localhost generally cannot sign in at a future tailnet hostname.
Development must use disposable accounts and origins, without changing the
operator's live authentication policy.

## Frozen feature coverage

| Contract | Inventory | Source mappings | Unmapped |
|---|---:|---:|---:|
| Piclaw Classic, pinned `70d33bc93ab540845bbcf5f80503ca8125c71594` | 236 IDs / 256 expanded cases | 101 IDs | 135 IDs |
| Shared Tau/Vibes interaction contract | 42 cases | 30 | 12 |
| Additional single-user passkey Settings contract | 26 scenarios/outlines | 0 | 26 |

These are mappings in `tests/ux/support/catalogue.mjs`, not a pass percentage or a
complete current run. Classic `@ux-original-008` is included in the 101 but disputed:
its skill test preserves a draft under the shared contract, whereas Classic's
inherited prefill path requires replacement. Conflicting criteria stay visible;
the frozen sources are not edited to make Gi pass.

The additive passkey contract is pinned separately and retains upstream
`@implemented` / `@browser-verified` tags. Those tags describe Piclaw, not Gi.
[Its README][passkeys] gives provenance, the required two-authenticator journey
and manual-device limits.

### Verification and remaining work

The most recent bounded workspace/Settings regression passed 204 browser cases;
the subsequent tightened motion check passed six cases. Functional validation
passed 100 tests with five fixture-dependent skips, and the auth suite passed 18.
Native CI run 36071594429 passed the main test job, auth tests on Linux/macOS/Windows
and five platform builds. These are distinct runs, with details in the
[implementation checklist][checklist]. No combined full-matrix result is claimed.

The six browser projects are Chromium and WebKit at phone, tablet and desktop
sizes. Motion-specific tests temporarily use 1024/1440/1920px widths and restore
the original viewport. Most saved screenshots are diagnostic artifacts; they do
not compare against a controlled current-Piclaw visual baseline.

| Command | Scope |
|---|---|
| `make check` | Go tests, vet, web build, hook checks and functional browser/API tests. |
| `make test-ux` | Fresh isolated functional instance; excludes `tests/ux/`. |
| `make ux-parity-inventory` | Frozen hashes and scenario inventory. |
| `make test-ux-parity` | Default six-project browser suite; specialised suites require their own flags/targets. |
| `make test-ux-auth` | Isolated native TOTP/browser authentication. |
| `make test-ux-journey` | Empty-store startup/new-chat/Return and explicit recovery; Chromium/WebKit at three sizes, required in CI. |
| `make test-ux-picker-geometry` | Pinned composer/picker outer geometry and dismissal; required browser CI step. |
| `make test-ux-slash` | Native slash catalogue and keyboard ownership with Quick Actions/Settings/search; required browser CI step. |
| `make test-ux-workspace-tabs` | Read-only tab keyboard/touch/lifecycle and conversation return; required browser CI step. |
| `make test-ux-passkeys` | Chromium virtual-authenticator WebAuthn API and Settings/login tests at three sizes; required in CI. |
| `make test-tui-smoke test-tui-gherkin` | Terminal smoke and Gherkin checks; specialised PTY suites are separate Make targets. |
| `make check-cross-build` | Optional local Linux/macOS amd64/arm64 and Windows amd64 builds with `CGO_ENABLED=0`; Windows is excluded from CI. |

CI gates builds on the isolated passkey browser suite, startup/Return Chromium/WebKit
journeys, pinned picker geometry, slash-key ownership, read-only workspace tabs and native Linux/macOS auth checks. It builds Linux/macOS amd64/arm64 artifacts. Windows CI tests, builds and
release artifacts were removed at the owner's request; local cross-build support
remains. CI does not run the full browser UX matrix. Remaining priorities include
remaining current-Piclaw composer/picker structure and styling, conversation shortcuts,
Settings focus, full Piclaw editor/workspace equivalence beyond read-only transitions,
disputed command/skill-prefill policies, and an
explicit full-suite runner with skip accounting. See the [suite guide][ux] and
[full web/TUI plan][plan].

[audit]: internal/ux-test-audit-2026-09-24.md
[tui]: internal/tui-pi-parity-plan.md
[media]: internal/tui-clipboard-media.md
[passkeys]: ../tests/ux/features/additions/piclaw-2026-09-24/README.md
[checklist]: checklists/implementation.md
[ux]: ../tests/ux/README.md
[plan]: internal/full-web-tui-parity-plan.md
