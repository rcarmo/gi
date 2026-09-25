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
| Chat and streaming | Implemented: native prompt admission, SSE status/draft/thought updates, reconnect reconciliation, bounded timeline paging and scoped search. | Fresh-chat/Return is an unresolved user report. Existing-session submission tests do not close that journey. |
| Composer and drafts | Partial: persistent browser-local text/media/references, failed-send recovery, file/folder/message references, upload progress/cancel/retry and byte-identical upload reuse. | Composer layout differs from current Piclaw; direct slash-menu and Quick Actions integration need stronger end-to-end tests. |
| Sessions | Partial: selection, child creation, grouping/search/typeahead, capability-gated pin/rename/archive/restore and draft isolation. | Picker geometry differs from current Piclaw. Session deletion and the complete session-management UI are not covered. |
| Models and context | Partial: session-local model selection, registry/context metadata, fit checks, usage meter and model commands. | Missing local-estimate labelling and remaining model/picker workflows; unavailable metadata stays unknown. |
| Queue and Stop | Partial: durable browser follow-ups, reorder/cancel, run-bound steering, queue return-to-draft, reconciliation and run-bound Stop. | Stop can advance the next queued turn, conflicting with shared-36's unchanged-queue criterion. That case is unmapped. |
| Compaction | Implemented: automatic/manual native compaction, persisted context checkpoints, progress/cancel and shared browser/terminal engine behaviour. | Broad Settings parity and every upstream compaction workflow are not complete. |
| Timeline and media | Partial: Markdown/tables/code copy, image lightbox, stored media/resource links, tool timing, recovered-response labels, idle single-message deletion and browser speech controls. | No iPad annotation workflow. Speech tests use a controlled browser boundary; physical audio is not verified. |
| Cards and widgets | Partial: supplied Adaptive Cards rendering and truthful rejection of unsupported Submit. | Accepted card actions and the full agent-authored widget/attachment tool surface are not implemented. |
| Workspace | Partial: rooted tree/hidden files, bounded previews, read-only tabs, MRU/pinning/context actions and explicit scoped lexical index/reindex. | No document editing, dirty-buffer save, popouts or docking. Automatic external-change freshness and vector search are not complete. |
| Workspace motion | Implemented: desktop collapse stays left-anchored; chat and toggle interpolate together, including rapid reversal and reduced motion. | Deliberate correction of an inherited Piclaw CSS defect. It does not close broader workspace-tab workflow gaps. |
| Settings | Partial: General identity, session Models, local Appearance, compaction controls/policy and guarded OpenAI/Anthropic API-key management. | Other/custom provider setup and browser OAuth flows, richer panes and remaining focus paths are incomplete. |
| Browser authentication | Partial: single-user code-only TOTP sign-in, HttpOnly/SameSite Strict cookie, transport/origin checks, transactional cross-process auth persistence and browser-owner session proof APIs. | Initial enrolment is an API workflow restricted to loopback. No initial-owner enrolment UI, complete logout/session UI or family mode. Opt-in WebAuthn APIs are tested separately. |
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
| Transcript | Fullscreen paging, tool folding, Pi-style message/tool bands, rendered-row search with occurrence navigation and prompt jumps. | Cross-wrap search, mutation-stable reflow/eviction anchors and light-theme parity need further work. |
| Native scrollback | Opt-in `-tui-mode regular`; terminal-owned history, selection and copy, with five idle editor/footer rows. | Fullscreen features are not all available in regular mode. Retained completion output follows documented retention limits. |
| Session/model choice | `Alt-S` / `Alt-M`, at most six visible results, native selection and draft/cursor preservation. | Regular-mode selectors use a temporary alternate screen. Remaining picker differences are tracked independently of browser tests. |
| Session actions | Bounded Pin/Unpin, Archive/Restore and Rename controls, including native capability checks. | No permanent action panel and no full session-management parity. |
| Compaction/index | `Alt-C` and `/compact`; `Alt-I` transient explicit index actions. | No background index-progress panel or automatic freshness guarantee. |
| Files | Process-local pending media refs, `/attachments`, pending-only `/detach`, native stored-byte admission. | No restart-persistent pending attachments, complete durable queue-draft recovery or general clipboard-image/drag-drop parity. |
| Copy and links | `/copy` for latest assistant source, fullscreen drag/copy/edge scroll, opt-in native/OSC52 clipboard, safe OSC8 targets for supported intact links. | Host clipboard/URL activation depends on the terminal. Wrapped/complex links and word-selection details are incomplete. |
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
| `make test-ux-passkeys` | Chromium virtual-authenticator WebAuthn API and Settings/login tests at three sizes; required in CI. |
| `make test-tui-smoke test-tui-gherkin` | Terminal smoke and Gherkin checks; specialised PTY suites are separate Make targets. |
| `make check-cross-build` | Optional local Linux/macOS amd64/arm64 and Windows amd64 builds with `CGO_ENABLED=0`; Windows is excluded from CI. |

CI gates builds on the isolated passkey browser suite and native Linux/macOS auth
checks. It builds Linux/macOS amd64/arm64 artifacts. Windows CI tests, builds and
release artifacts were removed at the owner's request; local cross-build support
remains. CI does not run the general browser UX matrix. Remaining priorities include the fresh-chat/Return
journey, current-Piclaw composer/picker geometry, slash/Quick Actions and Settings
focus, workspace tab transitions, the disputed skill-prefill mapping, and an
explicit full-suite runner with skip accounting. See the [suite guide][ux] and
[full web/TUI plan][plan].

[audit]: internal/ux-test-audit-2026-09-24.md
[tui]: internal/tui-pi-parity-plan.md
[media]: internal/tui-clipboard-media.md
[passkeys]: ../tests/ux/features/additions/piclaw-2026-09-24/README.md
[checklist]: checklists/implementation.md
[ux]: ../tests/ux/README.md
[plan]: internal/full-web-tui-parity-plan.md
