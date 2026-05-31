# Gi / PiClaw fit-gap

Date: 2026-05-27

Scope: compare Gi's current runtime, web, TUI, tools, extension, media, and CI posture against the PiClaw environment used by the coding-agent harness. This is an implementation planning note, not a compatibility promise.

## Executive summary

Gi now fits PiClaw's core coding-agent shape in the most important backend areas: durable sessions, turns, tool execution, routing, topics, media storage, web/TUI entrypoints, and CI single-binary delivery. The biggest gaps are not raw runtime capability; they are PiClaw's richer operator-facing surfaces: web timeline affordances, agent-side downloadable attachments, adaptive cards/widgets, plan/sidebar state, integrated model/thinking switching as agent tools, and polished cross-session orchestration UI.

Gi should not clone PiClaw's process model. The right target is to keep Gi's SQLite/WAL runtime truth and add selected PiClaw-grade operator affordances where they improve day-to-day use.

## Fit areas

### Runtime truth and coordination

Status: strong fit, Gi intentionally differs internally.

Gi has:

- SQLite/WAL-backed sessions, messages, turns, subturns, steering, media, hooks, topics, route events, inbound work, and active-turn claims.
- Durable active-turn coordination with stale-claim recovery and queued-turn handoff.
- Store-backed session identity, aliases, channel bindings, and main-session resolution.
- Durable inbound work queue with dispatcher lease ownership and retry/failure state.

PiClaw fit:

- PiClaw's user-visible behavior expects agent work to survive UI/request boundaries and be inspectable. Gi now fits that expectation.

Intentional difference:

- Gi uses relational SQLite as truth; PiClaw's surrounding harness has its own message/timeline database and active tool orchestration. Gi should not replace its store with PiClaw's timeline store.

### Tool execution

Status: partial-to-good fit.

Gi has:

- Built-in tool registry and active tool scope.
- Shell, file, VFS, search, scripting, skills, and extension tools.
- Tool lifecycle persistence and runtime topics.
- Existing `/api/tools/execute` shape preserved.

PiClaw fit:

- The basic tool-call model is compatible: tools return JSON-safe results and errors, and tool lifecycle is observable.

Gaps:

- PiClaw has a broader operator tool surface: attachment delivery, plan updates, model/thinking switching, keychain/env helpers, scheduled tasks, dashboard widgets, adaptive cards, SSH profile routing, and addon profiles.
- Gi has pieces of this, but not the same integrated operator affordance set.

### Topics, SSE, and live UI updates

Status: good fit.

Gi has:

- Canonical bounded topic bus with monotonic sequence numbers.
- `/sse/topics` with session/agent filtering and reconnect gap metadata.
- Runtime families for turns, sessions, tools, hooks, routing, inbound work, dispatcher, compaction, steering, and subturns.
- TUI topic-native consumption for key families.

PiClaw fit:

- This matches PiClaw's expectation that live UI surfaces can subscribe to state changes without polling full state every time.

Gaps:

- No durable replay API equivalent to PiClaw's message timeline diff/search/get tooling.
- Topic payloads are selective, not full-row mirrors.

### Media and attachments

Status: medium fit.

Gi has:

- Session-scoped `media` table.
- Hash/detected MIME metadata.
- Upload/list/get web API.
- TUI `/attach <path> [prompt]` fallback.
- Provider-safe image projection.

PiClaw fit:

- The media reference contract now resembles PiClaw's attachment-safe workflow: binary is stored out-of-band and referenced from messages/turns.

Gaps:

- No first-class user download card equivalent to PiClaw's `attach_file` response behavior.
- No web timeline inline image/file cards yet.
- No drag/drop or clipboard image paste in TUI.
- No attachment safety policy around deletion/moves comparable to PiClaw timeline attachment protections.

### Web runtime

Status: partial fit.

Gi has:

- Web server, session APIs, prompt submission, tool endpoints, runtime config, inbound work controls, SSE, topic SSE, auth surfaces, and media endpoints.

PiClaw fit:

- Gi can serve as a web-managed agent runtime.

Gaps:

- PiClaw's web UI is richer as an operator console:
  - message timeline with rich content blocks;
  - adaptive cards;
  - dashboard widgets;
  - downloadable attachments;
  - plan sidebar;
  - editor popouts;
  - cross-session chat controls;
  - integrated model/thinking controls.
- Gi's web UI is more application-specific and less harness-like.

### TUI

Status: improving fit with Pi/PiClaw terminal ergonomics.

Gi has recently moved to Pi's steady-state row contract:

- no top chrome;
- transcript-first rendering;
- bottom editor band with separator, editor row, separator;
- path/branch row;
- single bottom status/notification row;
- simplified `/help` and `/model`;
- `/attach` media fallback;
- scrollback via PageUp/PageDown/Home/End/mouse wheel.

PiClaw/Pi fit:

- The steady-state physical row order now matches Pi's layout contract.

Gaps:

- Needs broader screenshot regression coverage across prompt/tool/queued/error/scrolled states.
- No native image paste/drag-drop.
- No extension-provided editor replacement/widgets/overlays.
- No Pi-style startup resource summary with progressive disclosure (`ctrl+o more`) beyond simplified `/help`.
- Status line lacks cost/context window metrics.

### Extensions and scripting

Status: good backend fit, partial UX fit.

Gi has:

- JS and Joker bridge surfaces.
- Extension discovery/loading.
- Command registration and TUI dispatch.
- Topic APIs in both bridges.
- State/runtime namespaces.
- Mounted process hook support.

PiClaw fit:

- Extension commands, topics, and scripting bridge concepts align well.

Gaps:

- No npm-like Pi package manager equivalent.
- No extension-provided rich web/TUI widgets.
- No theme package surface comparable to Pi/PiClaw theming.
- Joker ergonomics still lag JS for callback-style command handlers.

### Skills and prompt templates

Status: medium fit.

Gi has:

- Skill discovery/loading.
- `/skills`, `/skill:name`, and related docs.

Gaps:

- No full prompt-template system equivalent to Pi.
- No packaged skill/theme/template installation workflow.
- No embedded reference browser comparable to Pi's doc/package model.

### Model/provider management

Status: partial fit.

Gi has:

- Runtime config for provider/model/thinking.
- `/model`, `/scoped-models`, `/thinking`.
- CI and local config persistence.

Gaps:

- No PiClaw-style agent tool for model/thinking switching.
- No unified provider login/subscription flow.
- No rich model registry browser.
- No status-line cost/context usage yet.

### CI/distribution

Status: good initial fit.

Gi now has:

- GitHub Actions CI.
- Web asset build.
- `go test ./...` and `go vet ./...`.
- Single-binary artifacts for Linux, macOS, and Windows.
- Release asset publishing on `v*` tags.

Gaps:

- No signed/checksummed release manifest yet.
- No package manager install/update flow.
- No smoke test that boots generated artifacts.

## Prioritized gaps

### P0: make the TUI actually Pi-identical in the steady-state layout

Why: user has explicitly asked for identical layout, not approximate.

Work:

- Keep no top chrome.
- Preserve Pi bottom band exactly:
  - separator;
  - editor row;
  - separator;
  - cwd/path + branch;
  - single notification/status line.
- Make status line carry transient notifications without adding rows.
- Add screenshot regression captures for empty, prompt, tool, queued, error, and scrollback states.

### P1: add status metrics to the single bottom line

Why: Pi's footer is useful because it exposes cost/context/model state without visual clutter.

Work:

- Add token/context counters if available from provider usage.
- Add best-effort cost placeholder or zero-cost indicator if no pricing data exists.
- Add queue/steering status only when non-zero.
- Keep one physical line with truncation and right-aligned model/thinking.

### P1: web/timeline attachment cards

Why: Gi now stores media but lacks PiClaw-grade operator download/preview affordances.

Work:

- Render media refs as web cards.
- Add download links for `/api/sessions/{id}/media/{media_id}`.
- Show MIME/size/hash metadata.
- Keep transcript text deterministic.

### P1: plan/sidebar equivalent

Why: PiClaw plan state is useful for long-running work and avoids burying task status in transcript messages.

Work:

- Add a small durable `plans`/`plan_items` table or session metadata projection.
- Expose web and TUI `/plan` commands.
- Publish `runtime.plan` topics.

### P2: richer PiClaw-like web operator affordances

Work:

- Adaptive-card-like structured prompts or forms.
- Dashboard/widget-like HTML panels.
- File attachment/download helper for generated artifacts.
- Editor popout/open-file equivalent.

### P2: extension UX parity

Work:

- Extension-provided web/TUI widgets.
- Extension-provided status/footer segments with strict one-line budget.
- Package/theme discovery/install flow.

### P2: release hardening

Work:

- Checksums for CI artifacts.
- Smoke-test built binaries.
- Release notes generation.
- Optional package-manager installation path.

## Non-goals

- Replacing Gi's SQLite/WAL runtime truth with PiClaw's message timeline store.
- Adding top chrome back to the TUI.
- Adding multi-line notifications/status in the bottom area.
- Pixel-cloning Pi internals if the same layout can be achieved with Gi's existing Go TUI stack.

## Recommendation

Continue with a focused PiClaw/Pi UX convergence track:

1. Lock the TUI layout contract with screenshot tests and fixtures.
2. Implement one-line status metrics and transient notifications.
3. Add web attachment cards/download affordances.
4. Add a durable plan/sidebar equivalent.
5. Then revisit richer extension widgets and package/theme ergonomics.
