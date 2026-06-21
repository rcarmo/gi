# TUI OAuth / login parity

Status: implemented as a status + credential-management surface backed by `auth.json`.

## What Gi does

PiSwift exposes `/login` and `/logout` to drive an interactive browser OAuth flow.
Gi's runtime authenticates providers through `~/.pi/agent/auth.json` (shared with
the inference layer), so the TUI surfaces and manages that credential state rather
than running a separate in-TUI browser flow.

### `/login [provider]`

- With no argument: lists every registered OAuth provider plus any extra
  `auth.json` entries, marking which are authenticated and the credential kind
  (`oauth`, `api-key`, `credential`). Also prints the `auth.json` path.
- With a provider id: reports whether that provider is authenticated, and if not,
  points at the credentials file to populate (then `/reload`).

### `/logout <provider>`

- Removes the provider's entry from `auth.json` (no-op with a clear message when
  there is nothing stored). Writes the file back with `0o600` perms.

## Backing helpers (`internal/inference`)

- `ListAuthStatus() []AuthStatus` — merges `oauth.ListProviders()` with the
  on-disk `auth.json` entries; never performs network calls.
- `RemoveAuthEntry(provider) (bool, error)` — deletes a credential entry and
  rewrites `auth.json`.
- `AuthFilePath()` — canonical `~/.pi/agent/auth.json` location.

## Constraints

- No network calls in the status/listing path (safe default for tmux/scripts/CI).
- Lives entirely in the transcript/bottom band; never adds top chrome.
- Tests: `TestListAuthStatusAndRemoveEntry` (inference) and
  `TestLoginAndLogoutCommands` (tui) use an isolated `HOME` with a temp
  `auth.json`.
