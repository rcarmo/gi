# Classic /theme and /tint 001–015: native command gap

`features/ux/classic/compose/theme-tint.feature` requires commands sent
through the Classic composer, timeline responses and visual changes. Gi has a
browser-local Appearance Settings pane, but it is a different interaction:
`web/src/gi-settings-appearance.ts` exposes explicit Save/Reset;
`web/src/gi-appearance.ts` persists `gi_browser_appearance_v1`. It does not
write the legacy `piclaw_theme`/`piclaw_tint` keys. The native quick-action
catalogue (`internal/web/quick_actions.go`) advertises `/model`, `/compact`
and loaded skills, not `/theme` or `/tint`. The Go turn/web handlers have no
`/theme` or `/tint` command path. No `@ux-theme-001`–`015` Gi browser test is
tagged.

| Frozen IDs | Missing interaction | Existing evidence that cannot substitute |
|---|---|---|
| `001`, `004`, `010`, `011` | Composer commands list themes/usage or reject invalid theme/tint with timeline feedback and unchanged appearance. | Native Appearance Settings validation does not submit composer commands or yield those responses. |
| `002`, `003`, `005`, `013`–`015` | `/theme` switches to/from ristretto, updates root attributes, CSS variables, legacy storage and timeline; round-trip matches the starting colours. | A saved Settings preset changes browser-local appearance, but uses different storage, explicit Save and no command timeline. |
| `006`–`009`, `012` | `/tint` sets named/hex colours, off, default-theme interaction, visual differences and refresh persistence through Classic command state. | Settings accepts `#RGB`/`#RRGGBB` for the default preset, not the full Classic command semantics or named-colour path. |

This is a **native capability gap for all 15 frozen scenarios**. It does not
negate `@gi-settings-009`/`010` or their own Appearance Settings tests.
`tests/ux/oracle/piclaw-theme-command-probe.ts` pins the installed 3.2.4
version and Classic source-map hash, then calls its
`src/channels/web/theming/ui-theme-commands.ts` parser directly. Nine inputs
cover theme list, ristretto/default, invalid `dark`, tint usage, hex/named/off
and invalid tint. All parser status/message/payload assertions passed. The
installed `src/channels/web/handlers/agent.ts` calls this parser, persists a
successful theme payload via `setServerUiThemeConfig` and broadcasts
`ui_theme`. `make test-piclaw-theme-handler` calls that installed handler with
an in-memory database and temporary config. `/theme ristretto` and `/tint
orange` each return HTTP 200 `ui_only`, emit `ui_theme`, write global
`piclaw-ui` extension KV and the temporary legacy config, and send a
forced-root timeline response. Invalid tint and the no-argument theme list
send responses without changing stored appearance.
`tests/ux/oracle/piclaw-theme-ui-probe.mjs` uses the shipped Classic UI with
a disposable `/agent/default/message` response and SSE events fed by the
installed parser. In Chromium and WebKit desktop, composer submission sent
`/theme ristretto` then `/tint #e11d48`; the `ui_theme` fixture updated root
attributes and legacy localStorage, and the tint response appeared in the
fixture timeline. `make test-piclaw-theme-combined` joins the installed
handler and shipped UI in Chromium and WebKit desktop with an in-memory DB
and temporary config. Both passed `/theme ristretto`, `/tint orange`, `/tint
off` and invalid tint: HTTP 200 `ui_only`, matching timeline posts, three
`ui_theme` events for the successful changes, matching legacy localStorage,
global extension KV and temporary config, and unchanged appearance after the
invalid command. The fixture never called an agent executor.

Neither fixture exercises the production router/authentication, a live chat
or browser reload. Named tint is covered in the joined probe, but physical
colour matching and reload persistence have not been checked. The frozen
Gherkin stays a historical contract; these bounded oracle probes and Gi
Settings grant no Gi command parity or physical/pixel credit.
