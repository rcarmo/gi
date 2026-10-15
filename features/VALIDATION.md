# Gherkin validation ledger

All **21** tracked Gherkin files are in this tree: Gi-specific, terminal and search contracts. The shared browser scenarios are not copied here; they live only in the fixtures-vibes suite.

Browser compliance now uses the shared fixtures-vibes suite at `references/fixtures-vibes` (pinned to `f796ddf`, Piclaw 3.2.5 scenarios) and its scenario report; current results are in [the feature matrix](../docs/feature-parity.md#shared-browser-compliance). Gi keeps no copy of the Classic or shared Gherkin, and no upstream snapshots. Non-duplicated native browser probes moved from `tests/ux/` to `tests/web-regression/`; runtime, adapter and provenance checks remain under `tests/ux/`. See [the current browser guide](../tests/ux/README.md). Rows without a shared suite test have no shared browser evidence.

The earlier installed-reference checks and bounded Gi journeys are in [the web UX oracle matrix](../docs/implementation/audits/web-ux-oracle-matrix-2026-09-28.md). Full acceptance requires independent Piclaw (or Pi for terminal) observations for each clause and failure path.

| File | Group | Validation | Evidence / next gate |
| --- | --- | --- | --- |
| [features/gi/sessions/codex-request.feature](gi/sessions/codex-request.feature) | Gi-specific | Pending installed-reference comparison | Gi test results alone do not prove Piclaw behaviour. |
| [features/gi/sessions/gi-basic-send.feature](gi/sessions/gi-basic-send.feature) | Gi-specific | Pending installed-reference comparison | Gi test results alone do not prove Piclaw behaviour. |
| [features/gi/sessions/gi-preview.feature](gi/sessions/gi-preview.feature) | Gi-specific | Pending installed-reference comparison | Gi test results alone do not prove Piclaw behaviour. |
| [features/gi/sessions/gi-swipe.feature](gi/sessions/gi-swipe.feature) | Gi-specific | Pending installed-reference comparison | Gi test results alone do not prove Piclaw behaviour. |
| [features/gi/sessions/provider-retry.feature](gi/sessions/provider-retry.feature) | Gi-specific | Pending installed-reference comparison | Gi test results alone do not prove Piclaw behaviour. |
| [features/gi/settings/gi-settings.feature](gi/settings/gi-settings.feature) | Gi-specific | Pending installed-reference comparison | Gi test results alone do not prove Piclaw behaviour. |
| [features/gi/tui/inline-prose-layout.feature](gi/tui/inline-prose-layout.feature) | Gi-specific | Pending installed-reference comparison | Gi test results alone do not prove Piclaw behaviour. |
| [features/gi/tui/markdown-tables.feature](gi/tui/markdown-tables.feature) | Gi-specific | Pending installed-reference comparison | Gi test results alone do not prove Piclaw behaviour. |
| [features/gi/tui/pi-queue-input.feature](gi/tui/pi-queue-input.feature) | Gi-specific | Pending installed-reference comparison | Gi test results alone do not prove Piclaw behaviour. |
| [features/gi/tui/system-notices.feature](gi/tui/system-notices.feature) | Gi-specific | Pending installed-reference comparison | Gi test results alone do not prove Piclaw behaviour. |
| [features/gi/ux/gi-deviations.feature](gi/ux/gi-deviations.feature) | Gi-specific | Quick Actions 001/002 and SVG 003 bounded; other clauses pending | Installed prefill oracle 6/6, native palette 69+69 and skills 12/12; SVG installed probe 6/6 and native rendering 42/42 with adversarial/oversize cases. See `../docs/implementation/audits/quick-action-prefill-2026-09-28.md` and `../docs/implementation/audits/svg-fences-2026-09-28.md`. Queue and full contract remain open. |
| [features/search/workspace-index.feature](search/workspace-index.feature) | Search proposal | Pending installed reference | 23 proposal pickles; separate from frozen corpus. |
| [features/tui/assistant_basics.feature](tui/assistant_basics.feature) | Terminal | Pending full Pi 0.87.1 contract | Disposable shell Gherkin feature passed after replacing stale `you:` prefix assertion with an exact padded user-content row; full `make test-tui-gherkin` passed eight shell features plus dedicated Markdown24/inline-prose6. No full Pi terminal acceptance. |
| [features/tui/input_history.feature](tui/input_history.feature) | Terminal | Pending full Pi 0.87.1 contract | Some scoped PTY/oracle tests exist; no full terminal acceptance. |
| [features/tui/keyboard_behavior.feature](tui/keyboard_behavior.feature) | Terminal | Pending full Pi 0.87.1 contract | Some scoped PTY/oracle tests exist; no full terminal acceptance. |
| [features/tui/markdown/rendering.feature](tui/markdown/rendering.feature) | Terminal | Pending full Pi 0.87.1 contract | Some scoped PTY/oracle tests exist; no full terminal acceptance. |
| [features/tui/pi_like_workflows.feature](tui/pi_like_workflows.feature) | Terminal | Pending full Pi 0.87.1 contract | Some scoped PTY/oracle tests exist; no full terminal acceptance. |
| [features/tui/plugins.feature](tui/plugins.feature) | Terminal | Pending full Pi 0.87.1 contract | Some scoped PTY/oracle tests exist; no full terminal acceptance. |
| [features/tui/session_workflows.feature](tui/session_workflows.feature) | Terminal | Pending full Pi 0.87.1 contract | Some scoped PTY/oracle tests exist; no full terminal acceptance. |
| [features/tui/settings_and_approvals.feature](tui/settings_and_approvals.feature) | Terminal | Pending full Pi 0.87.1 contract | Some scoped PTY/oracle tests exist; no full terminal acceptance. |
| [features/tui/tool_controls.feature](tui/tool_controls.feature) | Terminal | Pending full Pi 0.87.1 contract | Some scoped PTY/oracle tests exist; no full terminal acceptance. |

Gi-specific and derived proposal contracts do not award Classic or Shared parity. No physical-device, provider, full-queue or WebKit-reload acceptance is implied by this ledger.
