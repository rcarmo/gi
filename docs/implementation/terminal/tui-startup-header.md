# TUI startup header

gi's startup header is adapted from Pi's built-in header and loaded-resource listing (pi-coding-agent interactive mode). It is implemented in `internal/tui/startup_header.go`.

- **Collapsed:**
  - The wordmark and version. A blue "g" (the gopher avatar's blue, RGB 64,128,192) and an "i" in the theme's text colour follow the avatar's blue/white/black.
  - gi's key hints (`Esc interrupt · Ctrl+C quit · / commands · ! bash · Ctrl+O more`) and "Press Ctrl+O to show full startup help…".
  - One line on what gi reads from Pi.
- **Expanded** (Ctrl+O, Pi's `app.tools.expand`; it toggles together with tool output): gi's full key list.
- **Loaded resources** (Pi's listing; collapsed lists names, expanded lists paths): the model scope line (`enabledModels`), then `[Context]` (AGENTS.md), `[Skills]`, `[Extensions]` and `[MCP]` (a gi addition).
- **`quietStartup`** (settings.json, project over user): `true` hides the header and the listing; `"header"` keeps the header and hides the listing (Pi 1.0).
- **Placement:** in fullscreen the header is a virtual first transcript block (never stored), so it scrolls with the chat, survives session switches, and works with windowing, search and selection. Regular mode prints the collapsed header to scrollback once. Test fixtures leave it off (`chatTUI.startupHeader`); the app turns it on.
- **Version** (`internal/version`, `gi --version`): an `-ldflags` `Version`, else the module version, with local pseudo-versions shown as `dev-<rev>[-dirty]`.
