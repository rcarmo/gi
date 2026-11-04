# TUI alignment with Pi 1.5.0

The current target is the **published `@earendil-works/pi-tui` and
`@earendil-works/pi-coding-agent` 1.5.0 packages** (November 2, 2026).
This supersedes the version target in [the historical 1.3 audit](tui-pi-1.3-alignment.md),
not its outstanding limitations. Gi remains a Go renderer with SQLite-backed
sessions, not a TypeScript Pi runtime. Do not claim exhaustive parity.

## Changes in this update

- **Alt+Up edits pending text** at application level, independent of editor
  focus. Still-pending steering text is restored first, then follow-ups, then
  the current draft. Withdrawal and durable draft save share a SQLite writer
  transaction. The current run is **not aborted**. Queue-backed steering also
  cancels its original turn, preventing recovery from delivering it again.
  Claimed work cannot be withdrawn; conflicts leave draft/delivery unchanged.
  Pending attachments and maintenance operations remain guarded rather than
  being silently converted to text. Escape keeps its separate atomic abort
  contract. Windows/WSL use Alt+Q as Pi does.
- **PageUp/PageDown belong to open lists**, not the transcript: model, scoped
  models, login, MCP, fork, settings/theme submenu and slash autocomplete.
  Paging clamps at list edges; arrows keep their existing wrap semantics.
  Autocomplete retains paging in regular mode. Models unavailable under Gi's
  credential/context checks remain unselectable.
- **Model rows show useful aliases** in parentheses, but omit names that merely
  restate the ID with different case/separators. On narrow unavailable rows,
  the rejection reason takes precedence over an optional alias. The selected
  model's detail still shows its name. Search scores contiguous matches as
  well as greedy subsequences, and keeps provider identity ahead of proxy IDs.
- **Double-click token selection** joins `/ - + \\ ~ _ .` on Unicode word
  boundaries, matching Pi's iTerm2-style path/token character set. Gi uses UAX
  segmentation, not ICU's language-aware `Intl.Segmenter`; CJK dictionary word
  grouping is not claimed identical.
- **Colored blocks** retain one full-width colored padding row above/below and
  one uncolored separator between adjacent bands, including user→user and
  tool→user. Fullscreen projection/search/selection and regular scrollback use
  the same spacing helper. Tool state changes do not change band padding.

## User guide and self-documentation

[Using Gi's TUI](tui.md) is shipped as `vfs://reference/tui.md`, with command,
model, queue, recovery, clipboard and configuration guidance. The reference
index links the detailed contracts. The system prompt points to these embedded
references: the assistant need not guess Pi behavior or read a separate Pi
installation to explain Gi. [Prompt caching](prompt-caching.md) documents the
separate caching repair and provider limitations.

## Reproducible checks

`make test-tui-pi150-reference` installs isolated exact published packages,
checks package identities/versions, regenerates navigation/editor/hotkey/scoped
model fixtures into scratch, compares them and removes the installation.
Legacy `test-tui-pi110-reference` and `test-tui-pi130-reference` names are aliases
to the current target, not claims that the current fixture remains 1.3.
`generate-pi150-reference PI_NODE_MODULES=<absolute node_modules>` is the explicit
regeneration target; the comparison gate never rewrites checked-in goldens.

Native tests cover both modes at 60×18, 100×22 and 140×36 for list ownership and
colored band geometry, including tool statuses/expansion. Queue tests verify
atomic rollback, media guards, stale/claimed ownership and no duplicate delivery.
The versioned fixtures come from upstream, not from Gi-generated expected rows.
Verification: all native packages passed before the request to restrict further
runs; the complete TUI package also passed independently. The isolated published
reference comparison passed (navigation, editor, four platform hotkey fixtures,
scoped model selector). Focused real tmux/SQLite queue-edit checks passed in both
fullscreen and regular mode: steering then two follow-ups then draft, active run
retained, queue entries withdrawn, and no redelivery after the held tool completed.
The binary was rebuilt locally with tmpfs caches. No further full suites were
started after the focused-only instruction; broad browser/OS acceptance and live
provider performance are not claimed.

## Deliberate limits / follow-up

Gi has fixed Pi-default keys, not Pi's `keybindings.json` remapping. Native
extensions, media edit-all, custom dialog stacks, all terminal protocols, the
full Pi 1.4/1.5 Markdown/render-performance change list and every OS-specific
integration have not been independently requalified here. Existing Gi durable
recovery/ownership guards are preserved. No global Pi install, live provider
account, production database or running server is changed by the tests.
