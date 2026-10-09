# Settings title correction

The user's report concerned the redundant **Gi Settings** dialog title, not an extra menu entry. Installed Piclaw 3.2.4 renders `t('settings.title')`; its English translation is **Settings**. Gi's native dialog now uses that title. Native sections, lazy loading, focus restoration, keyboard ownership and mutation controls are unchanged.

`make test-piclaw-settings-title` reads the installed source map and verifies both the oracle wording and Gi's adapter. This is a source-level title check, not independent browser acceptance of all Settings behaviour.

The menu-entry functional test and rapid-shortcut test assert the exact title. Existing dependent browser locators and the active Gi feature file use the same accessible dialog name. Protected supplied source and historical contracts are unchanged.

Verification:

- Settings shell/native settings: 108 Chromium and 108 WebKit cases passed across phone, tablet and desktop projects. The first combined 216-case run hit the 285-second tool limit at case 194; that incomplete run is retained and not counted as passing. Both browser groups were then rerun in full, without increasing test timeouts.
- `make check`: Go/vet/build/hooks and 142 functional tests passed; 11 functional tests skipped.
- `make ux-parity-inventory`: 204 support tests / 7,747 assertions passed. No new mappings.

No deployment, restart or live-chat writes. Logs: `/workspace/tmp/gi-settings-title*.log`. Message-reference display labels and the rest of the web audit remain separate work.
