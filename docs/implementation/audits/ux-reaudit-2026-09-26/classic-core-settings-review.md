# Classic Settings 001–032: mounted Gi section audit

Gi mounts six sections in `web/src/gi-settings.ts`: General, Models,
Appearance, Compaction, Providers and Authentication. General has a cached
read-only runtime snapshot and a separate saved display-name control. The
other five sections lazy-load through `gi-settings-lazy.ts`. The frozen
`features/ux/classic/canonical/core-settings.feature` describes a larger
Piclaw Settings surface, so a matching section name alone earns no clause
acceptance. `make test-piclaw-settings-shell` now runs the installed 3.2.4
Classic shell with disposable settings reads: six Chromium/WebKit viewport
fixtures opened one dialog, exposed General while `/agent/settings-data` was
held, then dismissed with Escape and backdrop while retaining an unsent draft.
The released bundle already has the dialog module available in this ordinary
open sequence; holding data did not show the import-time loading shell.
This fixture does not establish a cold lazy-import loading state, section
mutation, production Settings routes, live authentication or physical input.

| Frozen IDs | Mounted Gi behaviour and missing clause |
|---|---|
| `001`–`004` | `settings-shell.spec.mjs` tags all four; focused **24/24** across six projects. `002` cold loading, `003` lazy cache and `004` header focus/element-width layout are source-backed. `001`'s cached reopen and one portal pass, while separate `settings-layering-review.md` covers backdrop/underlying pointer action: **qualified combined evidence**, not a single tagged all-clause test. |
| `005`–`006` | Gi Compaction exposes a read-only active automatic policy, saved instance policy for next restart, current progress, Compact now and Stop turn. It has no Classic per-chat watchdog/backoff list, clear-suppression action, probe row, remote-native/tool-result controls or dense full policy matrix. **Partial/native gap.** |
| `007` | Gi Providers filters rows and saves/removes editable OpenAI/Anthropic API keys; OAuth/token/custom setup is read-only via native tooling. It lacks the Classic per-provider OAuth/custom/reconfigure flow. **Partial/native gap.** |
| `008`–`010` | Gi Models performs session-scoped catalogue, model and thinking mutations with stale-read guards, context-fit blocking and at most 50 filtered options. It lacks enabledModels scope toggling and the Classic grouped master-detail/provider/publisher/family/variant/sort/pin controls and explicit provider-settings/Compact-context actions. **Partial; no tagged Classic IDs.** |
| `011`–`017` | Budget, Scheduled Tasks, Environment, encrypted Keychain, Add-ons, Keyboard and Workspace settings sections are absent from the mounted six-section dialog. Other tools and workspace controls cannot stand in for these Settings interactions. **Native section gaps.** |
| `018` | Browser-local Appearance saves preset and hex tint via `gi_browser_appearance_v1`, without server output-pad control or merging returned instance appearance settings. **Partial;** see `theme-command-gap.md` for separate composer commands. |
| `019`–`023` | General does not debounce/save Classic general/session snapshots or provide avatar/meters/widget-token controls. No Classic Sessions pane is mounted. **Native Settings workflow gaps.** |
| `024`–`032` | Recordings, Tools, Quick Actions Settings, Editor and Developer panes are not mounted. The Quick Actions palette and read-only workspace preview do not implement their Settings forms. **Native section gaps.** |

`@ux-settings-001`–`004` focused run passed 24/24; no other frozen ID has a
direct tagged Gi browser journey. Several Gi-only settings tests cover their
own narrower controls, not the missing Classic clauses. Production Piclaw
Settings routes, physical input and deployed Gi were not tested here. No
production code or frozen Gherkin changed.
