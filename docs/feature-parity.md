# Gi features and Piclaw parity

Updated: 2026-10-05, after runtime commit `423e8bf`. Gi's terminal follows Pi 1.0.1 (`@earendil-works/pi-coding-agent`) and its browser follows Piclaw 3.2.5. Browser behaviour is measured with the shared [fixtures-vibes](https://github.com/rcarmo/fixtures-vibes) suite.

Gi shares pinned Piclaw browser components and Pi/Piclaw configuration files, but has its own Go runtime, SQLite state and terminal UI. It does not replace Piclaw. The browser and the terminal share one turn engine; their interaction coverage is tracked separately.

## Status terms

| Status | Meaning |
|---|---|
| Implemented | Native behaviour exists with tests for the stated scope. Related Piclaw workflows may still differ. |
| Partial | A usable subset exists; the row names what is missing or different. |
| Scaffold | Configuration or internal types exist without an end-to-end user workflow. |
| Planned | No working implementation in Gi. |

## Shared browser compliance

`make fixtures-vibes` runs the suite pinned at `references/fixtures-vibes` (v0.2.0, `c28472f`) on Chromium and WebKit at phone, tablet and desktop sizes. `tests/fixtures-vibes/profile.json` declares the capabilities Gi claims, and `tests/fixtures-vibes/skips.json` lists every scenario Gi does not pass, with its reason. A failure that is not listed, a capability skip that is not listed, or a listed scenario that now passes fails the report gate.

The last full run started on 2026-10-04 at 20:46:28 UTC with Gi `5a68f4005a9fdab60d72808579c0c483536f07d1` and fixtures `6e49ae1dfb0c6ca900a9173ecc80817c12fd77cb`. It finished on 2026-10-05 after 388.5 minutes. Playwright recorded 1,233 passed tests, 280 failed, 71 skipped and zero flaky results across the six projects.

| Report | Scenarios | Passed | Failed (unlisted) | Skipped | Listed failing | No suite test yet |
|---|---:|---:|---:|---:|---:|---:|
| Original full4 gate | 316 | 185 | 1 | 3 | 49 | 78 |
| Same results, reconciled skips on 2026-10-05 | 316 | 185 | 0 | 3 | 50 | 78 |

The original gate failed because mandatory-core scenarios still used `capability-absent`, `@ux-extra-004/005` and `@ux-editor-007` had stale skips despite passing, and Vim scenario `@ux-editor-009` failed in all six projects without an issue-backed entry. The corrected skips file uses `not-implemented` for core workflows, removes those three passing entries and adds editor009 under [gi#46](https://github.com/rcarmo/gi/issues/46). Re-running only `make -C references/fixtures-vibes report` against the existing `compliance.json` gives `Gate: OK`; no browser tests were rerun. This gate accepts the issue-backed failure inventory, not feature completeness. Reports and the raw results are in `references/fixtures-vibes/test-results/`; an archived report is in [full4](internal/fixtures-vibes-full4.md).

The earlier 13:24 UTC run on `88e5b13` had a single `@ux-chat-lifecycle-009` WebKit `page.goto` engine error. It passed in the other five projects and in 10 of 10 repeats afterwards. Full4 passes that scenario. Focused checks also passed for `/theme` and `/tint` (90/90), avatar manifest icons (54/54), and confirmed cascade deletion with cancel/retry (36/36).

The current suite pin is v0.2.0 (`c28472f`); its Classic sources/static assets are unchanged from the accepted frontend handoff at `4259e82`. Independent focused checks on 2026-10-05 passed Plan/widgets/extra014 in all six projects (66/66), frontend units (189/189), and Gi functional tests (142 passed, 11 skipped). The eight Plan skips are removed; Plan and widgets are implemented. A new full six-project run on v0.2.0 is authorised and being prepared; its gate has no result yet. See [browser checks](internal/classic-ui-acceptance.md).

The `/btw` backend has isolated session-snapshot inference, abortable JSON/SSE routes, explicit retry and normal-prompt injection coverage. Ten focused native checks repeated three times and one isolated Chromium API journey pass. The pinned Classic app still needs its composer/panel/stream wiring; #40, capability claims and shared skips are unchanged. See [side-prompt contract](internal/side-prompt.md).

Scenarios Gi does not pass:

| Reason | Scenarios | Detail |
|---|---|---|
| Not implemented | 42 entries | `/btw` ([gi#40](https://github.com/rcarmo/gi/issues/40)), terminal ([gi#45](https://github.com/rcarmo/gi/issues/45)), CodeMirror/Vim editor ([gi#46](https://github.com/rcarmo/gi/issues/46)) and VNC ([gi#47](https://github.com/rcarmo/gi/issues/47)). |
| Capability absent | 2 entries | `@ux-shell-env-002/007`: Windows builds and shell detection exist, but no Windows host test ([gi#50](https://github.com/rcarmo/gi/issues/50)). |

`@ux-shell-008` is skipped by the suite in all six projects. Adaptive Cards, image annotation and text highlights are removed from this suite pin; their 14 former skip entries are gone.

Scenarios with no suite test have no shared browser evidence either way.

## Browser and runtime

| Area | Gi status and behaviour | Differences and open work |
|---|---|---|
| Runtime and distribution | Implemented: one pure-Go binary with embedded browser assets; SQLite/WAL sessions, messages, turn events and recovery; `go-ai` 1.0.1 inference. Interrupted turns are held for review, not replayed. | Piclaw extensions do not run in Gi. Bun is build-time only. |
| Chat and streaming | Implemented: prompt admission, SSE status/draft/thought updates, reconnect reconciliation, bounded timeline paging and scoped search. The timeline follows new replies while pinned to the bottom and keeps the reading position otherwise. | Conversation-level shortcuts and full visual equivalence with Piclaw are not verified. |
| Composer and drafts | Implemented: persistent browser-local text, media and references; failed-send recovery; file, folder and message references; upload progress, cancel and retry. Message references carry the numeric message row ID (`msg:42`). | Physical IME input is untested. |
| Sessions | Partial: selection, child sessions, grouping, search and typeahead, pin, rename, archive (except the last main session), restore and per-session drafts. | `/fork` and `/clone` create a new `@agentN` ([gi#20](https://github.com/rcarmo/gi/issues/20)). |
| Models and context | Implemented: session-local model and thinking selection, registry and context metadata, fit checks, and Piclaw 3.2.5's context meter. The meter can start compaction and shows an estimate after one. | The thinking regression tests predate Pi's effective default level and need updating. |
| Queue and Stop | Implemented: durable follow-ups, reorder and cancel, run-bound steering, return to draft and run-bound Stop with explicit Resume. See [contract](internal/web-stop-queue.md). | — |
| Compaction | Implemented: automatic and manual compaction with model-written summaries (Pi's cut point and prompts), persisted context checkpoints, progress and cancel. See [compaction](internal/compaction.md). | — |
| Timeline and media | Partial: Markdown, tables and code copy; image lightbox; stored media and resource links; tool timing; turn-outcome chips; single-message and confirmed cascade deletion; browser speech. | No iPad annotation or text highlights. Physical audio is unverified. See [cascade deletion](internal/message-deletion.md). |
| Cards and widgets | Implemented dashboard widgets: durable tool, authenticated artifact lookup, post-commit SSE, opaque-origin sandbox and session-scoped bridge. Adaptive Cards render; unsupported Submit actions are rejected visibly. | No agent card-posting API. See [widgets](internal/dashboard-widgets.md). |
| Workspace | Partial: rooted tree with hidden files, bounded previews, read-only tabs, folder hints, uploads, explicit lexical index/reindex; complete bounded edit reads, guarded Piclaw file aliases and client-scoped filesystem SSE. | Editable CodeMirror/Vim panes, popouts and docking await the shared frontend handoff ([gi#46](https://github.com/rcarmo/gi/issues/46)); backend [contract](internal/workspace-editor-backend.md) passes native/API tests. Plain code workspaces index nothing by default ([gi#22](https://github.com/rcarmo/gi/issues/22)). No vector search. |
| Web terminal | Implemented backend: native Linux/macOS PTY, WebSocket input/resize, owner isolation, bounded replay, reconnect, handoff and shutdown. See [contract](internal/web-terminal.md). | Shared frontend dock/tab/popout/zen integration awaits its owner ([gi#45](https://github.com/rcarmo/gi/issues/45)). Capability and skips stay unchanged until independent browser acceptance. Windows PTY is unsupported. |
| VNC | Implemented backend: explicit target policy, authenticated WebSocket/TCP bridge, bounded buffers/timeouts, owner/target-bound handoff and shutdown. See [contract](internal/web-vnc.md). | Disabled without operator target configuration. Viewer/tab/popout integration awaits the shared frontend owner ([gi#47](https://github.com/rcarmo/gi/issues/47)); capability/skips stay unchanged. Physical display behaviour is unverified. |
| Settings | Implemented: General (identity, upload limit), Models (filtered), Appearance (theme presets, tint, output padding; stored in the browser; also set by `/theme` and `/tint`), Keyboard (shortcut editing), Compaction, Providers (OpenAI and Anthropic API keys), Keychain, Environment and Authentication. | Providers cannot set up OAuth or custom providers; use `/login` in the terminal. |
| Keychain and shell environment | Implemented: Piclaw-format encrypted keychain, shell substitution of named secrets, and environment overrides applied to every shell path. See [keychain](internal/keychain.md) and [shell environment](internal/shell-environment.md). | — |
| MCP and codemode | Implemented: stdio and Streamable HTTP servers from Pi's `mcp.json`, OAuth sign-in, direct and deferred tools with `tool_search`, resources, the QuickJS-on-wazero codemode tool, `gi mcp` and `/mcp` in both UIs. See [MCP](internal/mcp.md) and [codemode](internal/codemode.md). | Servers that use provider authentication (`auth.provider`) are not supported ([gi#29](https://github.com/rcarmo/gi/issues/29)). Codemode has no dedicated renderer ([gi#30](https://github.com/rcarmo/gi/issues/30)). |
| Browser authentication | Partial: single-user TOTP sign-in, HttpOnly/Strict cookie, transport and origin checks, transactional auth storage, browser-owner proof, Settings logout and loopback-only initial owner setup. | No QR setup, family accounts or broader session management. |
| Multiple passkeys | Partial: pure-Go WebAuthn registration, login and re-authentication; add, list, rename and remove with fresh proof; last-factor protection; Settings and login controls tested with Chromium virtual authenticators. See [contract](internal/passkeys.md). | WebKit ceremonies and physical or synced keys are untested. |
| Skills, tools and scripting | Partial: native tools (including Pi's edit tool and the messages tool), embedded Joker/JavaScript bridges, process extensions and hooks, user and project skills, managed VFS and browser skill commands. | No general Pi or Piclaw package compatibility. |
| Operator integrations | Partial: backend routing, topics and inbound-work primitives; transactional per-session Plan storage/tool/API with SSE. | Editable Plan browser workflow passes independent acceptance. Scheduled tasks, SSH/Proxmox/Portainer and remote-agent surfaces are not ported. See [Plan](internal/session-plan.md). |

An instance with no enrolled owner permits application access. The CLI binds to loopback by default; `make start` binds to `0.0.0.0`. Use `BIND=127.0.0.1` until authentication is configured. Serving HTTPS does not enrol an owner.

## Terminal

The terminal follows Pi 1.0.1's layout, colours and keyboard handling, using Go widgets. Web overlays do not become permanent terminal panels.

| Area | Behaviour | Limits |
|---|---|---|
| Pi commands | `/settings`, `/model`, `/thinking`, `/scoped-models`, `/tree`, `/fork`, `/clone`, `/resume`, `/new`, `/name`, `/session`, `/compact`, `/copy`, `/export` (HTML or JSONL), `/import`, `/share` (secret gist), `/login`, `/logout`, `/hotkeys`, `/reload`, `/quit`, with Pi's argument completions. | `/bug`, `/changelog` and `/trust` have no Gi equivalent yet ([gi#16](https://github.com/rcarmo/gi/issues/16)). |
| Gi commands | `/abort`, `/queue`, `/retry`, `/draft`, `/attach`, `/attachments`, `/detach`, `/paste-image`, `/tools`, `/mcp`, `/codemode`, `/skills`, `/agents`, `/spawn`, `/switch`, `/send`, `/where`, `/plugins`, `/scrollback`. | — |
| Themes and keys | Pi's dark, light and terminal-derived themes, custom theme files with live reload, and Pi's default keybindings, kill ring, undo and jump. | — |
| Transcript | Fullscreen paging, tool folding, Pi-style message and tool bands, rendered search with occurrence navigation, prompt jumps and the "Jump to latest message" cue. | Search across wraps of prewrapped Markdown and inline code is incomplete. |
| Native scrollback | `-tui-mode regular` (or the `tuiMode` setting): terminal-owned history, selection and copy. | Not every fullscreen feature is available. |
| Editor | Cursor-following window capped at 30% of the rows, grapheme-aware wrapping, Pi's large-paste markers, unchanged draft bytes. | — |
| Files and copy | Durable pending media references, `/copy`, drag selection with edge scroll, word and line selection, OSC 52 clipboard by default, OSC 8 links. | Clipboard and link activation depend on the terminal emulator. |
| Authentication | Local terminal access uses the local runtime. `/login` handles provider OAuth and API keys. | Passkey management needs the browser. |

PTY tests cover fullscreen and regular modes at 60×18, 100×22 and 140×36. Protocol-byte checks for OSC 52 and OSC 8 do not prove every terminal emulator's behaviour. See the [TUI plan][tui] and the [clipboard and media contract][media].

## Requested integrations

| Workstream | Current state | Acceptance target |
|---|---|---|
| tsnet remote access | Scaffold: `internal/peering` wraps `tailscale.com/tsnet` and reports status. Nothing starts a tailnet listener. See [plan](internal/peering-tsnet-plan.md). | Opt-in tailnet HTTPS for the existing UI, API and SSE, with persistent node state, secret references and the existing app authentication. No public exposure by default. |
| Iroh inter-instance chat | Planned; Gi has no Iroh transport. | Explicit pairing, receiver-owned policies, signed bounded messages and files, durable retries and one-hop peer and agent addresses. Interoperability with Piclaw's remote-peer add-on needs confirmation. |

tsnet carries operator web access and Iroh carries peer messages. Neither grants the other's authority. The runtime stays pure Go, without CGO, native Rust libraries or sidecars.

Choose the production HTTPS hostname before enrolling real passkeys. A key registered for `localhost` generally cannot sign in at a later tailnet hostname.

## Verification

The six browser projects are Chromium and WebKit at phone, tablet and desktop sizes. Saved screenshots are diagnostic; they are not compared against a controlled Piclaw visual baseline.

| Command | Scope |
|---|---|
| `make check` | Go tests, vet, web build, hook checks and functional browser/API tests. |
| `make test-ux` | Functional browser/API tests on a fresh isolated instance. |
| `make fixtures-vibes` | Shared browser compliance and its report gate. |
| `make test-web-regression` | Gi-only browser race and recovery regressions. |
| `make test-ux-auth`, `make test-ux-passkeys` | Native TOTP and browser authentication; Chromium virtual-authenticator WebAuthn. |
| `make test-ux-journey`, `make test-ux-picker-geometry`, `make test-ux-slash`, `make test-ux-workspace-tabs` | Focused browser suites run in release CI. |
| `make test-tui-smoke`, `make test-tui-gherkin` | Terminal smoke and Gherkin checks; other PTY suites have their own targets. |
| `make check-cross-build` | Local Linux/macOS amd64/arm64 and Windows amd64 builds with `CGO_ENABLED=0`. |

CI runs only for `v*` release tags. It runs fixtures-vibes compliance, the focused browser suites, the passkey suite, native auth checks on Linux and macOS, and Linux/macOS/Windows amd64/arm64 builds. Windows binaries are built but not tested.

Dated plans and audits keep their original scope: the [UX audit][audit], the [full web/TUI plan][plan] and the [implementation checklist][checklist].

[audit]: internal/ux-test-audit-2026-09-24.md
[tui]: internal/tui-pi-parity-plan.md
[media]: internal/tui-clipboard-media.md
[checklist]: checklists/implementation.md
[plan]: internal/full-web-tui-parity-plan.md

## Shared front-end ownership

The front-end sources and embedded assets live in `references/fixtures-vibes/ui/classic`. Gi imports `giui.Static` and `giui.ThemeCatalogue` from that module. Front-end changes are made upstream and arrive through submodule updates. Go backends, integration tests and publication remain in Gi.

Widget and Plan browser workflows pass independent six-project acceptance ([widgets](internal/dashboard-widgets.md), [Plan](internal/session-plan.md)). The profile now claims `@cap-widgets` and `@cap-plan-sidebar`.
