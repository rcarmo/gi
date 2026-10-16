# Full TUI alignment with Pi 1.3

Target: released pi-mono `71023aec1546d6d677a65669c5665c0927ce0dd0` (1.3.0). User requested full alignment on 16 October 2026. This is a behavioural alignment, not replacing Go with the TypeScript renderer. No full parity claim until the crosswalk and acceptance are complete.

Branch: review/tui-pi-1.3-alignment, based on merged main76baa1a2. Release-gate repairs are separate in PR58; v0.1.1 is an immutable failed release attempt and no new tag is authorised by passing only a TUI subset.

## Coverage required

| Area | Upstream source surface | Gi surface | State |
|---|---|---|---|
| Editor/input | tui components/editor,input; coding-agent custom-editor | multiline_input, editor_layout, keybindings, app_keys | Audit cursor, selection, paste, undo, Unicode, history, submit |
| Fullscreen/regular terminal | tui-alt-screen,tui,terminal,stdin-buffer | chat, terminal_output, go-tui fork | Audit resize, retained scrollback, redraw, focus, protocol/IME |
| Transcript rendering | markdown,text,truncate; assistant/user/tool messages | markdown*, transcript*, tool rendering | Audit width/theme invalidation, whitespace, outputPad, thinking/tool expansion |
| Overlays | tui overlay handles and interactive selectors | selectors, settings, workspace index, menus | Unhandled wheel slice implemented; ownership/closure/focus audit pending |
| Settings/keymaps/theme | settings-list,keybindings,theme | settings_list/menu, themes, keybindings | Full defaults/options/error crosswalk pending |
| Runtime UX | interactive.ts, footer, queue, compaction, session tools | chat, turn/store adapters | Audit visible state, steering/followup/cancel/retry, active-tool resume |
| Terminal integrations | terminal-image, OSC/status/cursor/clipboard | terminal theme/output, status, clipboard | Source/contract audit and supported-platform matrix pending |
| Acceptance | upstream tui and interactive tests | Go tests and real tmux harnesses | Per-contract evidence and final profiled suite pending |

Existing wheel acceleration, transcript caching, tool timing, same-session reset and parser-error work must be reused, not reimplemented. Historical docs contain stale go-tui0.18.2 restrictions; current runtime-only0.22.1 fork carries fixes, so verify its actual behaviour instead of blindly downgrading.

## First completed slice

Pi fix11f99114 routes wheel input not consumed by a keyboard overlay to the fullscreen transcript. Gi previously swallowed all mouse events for rename/workspace overlays. It now routes wheel up/down through existing transcript scrolling without closing overlays or mutating drafts. Focused overlay and input-region tests pass race ×3 under Go1.27.2 with isolated project test roots. This is development verification, not full pre-release profiling or PTY acceptance.

## Definition of done

Inventory all upstream tui source modules and relevant interactive components; map each user-visible contract to Gi implementation/tests, or document an explicit host/platform incompatibility for user review. Implement gaps without weakening security, cancellation, history or layout assertions. Verify fullscreen and regular modes at60×18,100×22,140×36, including Unicode, resize, streaming, overlays, cancellation and session changes. Profile CPU/alloc_space/alloc_objects during final pre-release tests, analyse and immediately dispose captures. Full alignment remains in progress.
