# MCP client (`internal/mcp`)

Status: the client core (#25, phase 1), tool exposure (phase 2) and
`tool_search` (phase 3) are in place. Codemode, `/mcp` and OAuth are later
phases
([plan](../implementation/plans/mcp-codemode-plan.md)). The TUI and web server call
`Engine.EnableMCP()` at startup; engines built for tests never read the user's
`mcp.json`.

## Configuration

The format is Pi's `mcp.json`, read from:

1. user level: `~/.gi/agent/mcp.json`, otherwise `~/.pi/agent/mcp.json`;
2. project level: `<workspace>/.gi/mcp.json`, otherwise
   `<workspace>/.pi/mcp.json`. **This is read only for trusted projects.** gi
   has no project trust yet (#16), so project MCP configs are ignored for now.

For each file the first existing location wins; the `.gi` and `.pi` files are
not merged (#26). A project entry replaces a user entry with the same name, and
a project `autoEnableCodemode` overrides the user value.

A project entry without `command`, `url` or `type` is a **project override**
(Pi 1.0.1): it sets only `enabled`, `exposure` and `toolExposure` of the user
server with the same name, and the rest of the user entry (including
credentials the project could not set) is kept. `toolExposure` replaces the
user map. An override without a user server, or with another key, is an
error. The server keeps scope `global` and records the override file
(`ServerConfig.Override`); `Config.ProjectConfig` is the trusted project's
mcp.json. Golden: `scripts/golden-mcp-overrides.mjs` runs Pi's own
`loadMcpConfig` and `updateMcpServerConfig`; `TestProjectOverridesMatchPi`.
Until project trust exists (#16) no project file is read, so overrides and the
manager's per-project actions stay inactive in practice.

```json
{ "mcpServers": { "internal-tools": { "enabled": false } } }
```

```json
{
  "autoEnableCodemode": true,
  "mcpServers": {
    "fs":   {"command": "npx", "args": ["-y", "@modelcontextprotocol/server-filesystem", "."], "env": {"TOKEN": "${MY_TOKEN}"}, "cwd": "."},
    "docs": {"url": "https://example.com/mcp", "headers": {"Authorization": "Bearer ${DOCS_TOKEN}"}, "timeout": 30,
             "exposure": "deferred", "toolExposure": {"search": "direct", "delete_*": "hidden"}, "description": "Product docs"}
  }
}
```

Rules, as in Pi:

- **Transport:** `command` selects stdio and `url` selects streamable HTTP.
  An optional `type` must be `stdio`, `http` or `streamable-http`. `sse` is
  rejected.
- **Names:** server names use letters, digits, `_` and `-`. Names that differ
  only in `-` and `_` are the same server, so a second one is rejected.
- **Common fields:**
  - `timeout`: seconds, default 60.
  - `enabled`: `false` keeps the entry without connecting.
  - `description`: what the server offers.
  - `exposure`: `codemode` (the default), `deferred`, `direct` or `hidden`;
    `codemode-deferred` is an alias of `codemode`.
  - `toolExposure`: per-tool overrides. An exact tool name wins, otherwise
    the first matching `*` pattern.
- **Values:** `env` and `headers` values expand `${VAR}`. A value that is
  wholly `!command` runs the command and uses its trimmed output. Expansion
  happens only when connecting. A leading `~/` expands in `command`, `args`
  and `cwd`, and a relative `cwd` resolves from the workspace.
- **`oauth`** (url servers) is validated as Pi's `validateOAuth`, with Pi's
  messages: `clientId`, `clientSecret`, `scope` strings, `callbackPort`,
  a loopback `callbackUrl`, a non-empty `clientName`, `clientRegistration`
  (`dcr` or `cimd`; `cimd` excludes `clientId` and `clientName` and needs
  a `/callback` path on localhost or 127.0.0.1) and `authServerMetadataUrl`.
  Golden: `scripts/golden-mcp-cimd.mjs`, `TestOAuthSettingsAndCIMDMatchPi`.
- **`auth.provider`** reuses current provider credentials from `auth.json`; it is separate from MCP OAuth. See [provider authentication](#provider-authentication).
- **Invalid entries** are reported (`Config.Errors`) and skipped; the other
  servers still connect.

## Provider authentication

```json
{
  "mcpServers": {
    "radius-tools": {
      "url": "https://example.com/mcp",
      "auth": { "provider": "radius" }
    }
  }
}
```

The provider must already have credentials in the normal Gi/Pi `auth.json`. API-key entries use `key` or legacy `apiKey`; OAuth entries use the registered go-ai provider's token derivation. An unregistered provider can use a stored, unexpired `access` or `token`. MCP never sends refresh credentials or starts a provider sign-in.

Each HTTP request, including stream reconnects and session cleanup, reads the current selected entry. Rotation applies without reconnecting; removed credentials fail before sending a request. Expiring registered OAuth credentials refresh under Pi's shared `auth.json.lock`, re-reading the entry to preserve concurrent rotation/logout, unrelated providers and provider-specific fields. Credential-lock acquisition and supported refreshes honour cancellation. A provider that needs refresh but lacks context-aware refresh fails instead of blocking an MCP call indefinitely. Valid tokens use a read-only path without creating credential locks or rewriting the file.

Provider-authenticated URLs require HTTPS or HTTP on literal loopback IPs or `localhost`. User information and URL fragments are rejected. `auth.provider` cannot coexist with `oauth` or an `Authorization` header and applies only to HTTP servers. Empty/control-character tokens are rejected. Requests must keep the configured scheme, host, path and query; redirects and Host overrides are blocked. Configured non-authentication headers remain available. Provider credential/refresh errors and HTTP transport errors return generic messages without credential contents.

These servers return false from `UsesOAuth()`: `gi mcp login|logout` and `/mcp` OAuth actions cannot create or delete MCP credentials for them. Manage the provider's sign-in through its normal provider controls. Tests use fake credentials and local MCP servers; no live provider sign-in is required.

`env -u PI_CODING_AGENT_DIR -u GI_CODING_AGENT_DIR make test-mcp-provider-auth` runs the provider-auth checks three times with race detection where supported. It covers secure URL validation, current-token rotation/revocation, redirects and endpoint/Host changes, error redaction, cancelled refresh/lock waits, concurrent refresh, provider-specific fields, read-only valid tokens and CLI OAuth exclusion. Go 1.27.1 verification also passed 2,167 tests across 37 packages and vet. Browser/UI fixtures are unchanged.

Focused captures are short: CPU samples are too sparse for throughput comparisons. Fresh MCP server/schema setup and HTTP JSON decoder buffers dominate allocations; provider-token reads stay bounded to the credential file's 1 MiB limit. The valid OAuth path avoids write-lock acquisition and rewriting; concurrent callers re-read after locking so one successful refresh serves the others. No equivalent-workload speed improvement is asserted. CPU, `alloc_space` and `alloc_objects` were inspected; raw captures and disposable logs were deleted after analysis.

## Lifecycle

- **Connecting:**
  - `Manager` connects lazily on first use, or in the background with
    `ConnectAll`.
  - Concurrent callers share one session per server.
  - HTTP connections retry transient failures (408, 429, 5xx, network) twice.
- **Calls:**
  - Tool calls are never retried, because the server may already have
    performed them.
  - Each request has the server's `timeout`.
  - Results with `isError` are returned as results, not errors.
- **Failures:**
  - A transport failure drops the session, and the next call reconnects.
  - Server-side JSON-RPC errors, cancellation and timeouts keep the session.
- **Tool lists** are paginated and cached until the server announces a change.
  New tools then appear and withdrawn ones become unreachable.
- **stdio servers:**
  - They inherit gi's environment plus the expanded `env`.
  - Each runs in its own process group.
  - Stopping one closes its stdin, sends SIGTERM to the group, waits up to
    2 s, then sends SIGKILL to anything left, so wrapper children (`npx`,
    `uvx`) do not survive.
- **Logging:**
  - Server logging notifications are written to `mcp.log` as
    `<time> [<server>] <level> <logger>: <message>`, and rotated to
    `mcp.log.1` past 5 MB.
  - Servers on protocol 2026-07-28 get the level per request through
    `_meta`; older servers through `logging/setLevel`.
- **Startup report:** once every enabled server has finished its first
  connection attempt, the engine posts Pi's `reportProblems` message
  (`ProblemReport` in `internal/mcp/report.go`): config errors, failed servers
  and servers that need a sign-in, then "Run /mcp to fix." Config errors are
  reported even when no server is left to start.
  - Notices go to `Engine.SetMCPNotifier`. Notices posted before a notifier is
    set are kept and delivered when it is set. The TUI shows warnings as
    `Warning: …` lines (Pi colours them; gi does not yet). `gi -web` has no
    notifier and only logs them.
- **Shutdown:** `Engine.Close` cancels background work before closing the
  servers, so a server still connecting does not delay exit.
- **Status:** `Status()` reports each server's state (`disabled`,
  `disconnected`, `connecting`, `connected` or `failed`), its error, tool
  count, exposure, source file, instructions and stderr tail.

## How tools reach the model

- **Naming:** tools are named like Pi's: `mcp__<server>__<tool>`, with every
  character outside `[A-Za-z0-9_]` replaced by `_`.
  - A name over 64 characters, or one that collides, gets `_` plus 8 hex
    characters of SHA-256 of `server\0tool`.
  - Tools of one server that sanitize to the same name all get the suffix.
  - The server's namespace is `mcp__<server>` with `-` replaced by `_`.
- **`direct` tools:** registered in the tool registry like built-in tools. Hooks,
  permissions, events and audit apply, and the source is `mcp:<server>`. A
  `readOnlyHint` annotation marks the tool as `read`.
  - Before the first prompt's tool set is admitted, admission waits up to 10 s
    (Pi's `startupWaitMs`), outside the runner lock, for servers that can expose
    direct tools. Only the first prompt of a run waits. If a server is still
    connecting, Pi's notice "MCP servers are still connecting; their tools
    become available once connected." is posted, and its tools are declared
    once it connects. As in Pi, tools loaded with `tool_search` earlier in a
    resumed session are declared once their server has registered them; the
    first prompt does not wait for them.
  - When a server announces a changed tool list, its tools are re-registered,
    and withdrawn tools are unregistered.
- **`codemode` and `deferred` tools** are registered as *deferred*:
  executable, but not declared to the model until loaded. Admission's default
  tool set skips them.
- **`tool_search`** (Pi's, #25 phase 3) is registered whenever an enabled
  server can give tools `deferred` exposure. It is decided from the config,
  before servers connect.
  - **Ranking:** a port of Pi's BM25 ranker, using Pi's tokenizer, stop words
    and naive stemming. The search text is the name, the description, schema
    descriptions, property names, and the server namespace with its
    description and instructions. It searches deferred tools the session has
    not loaded yet.
  - **Loading:** matches are recorded in the session state (`loaded_tools`),
    so they survive restarts and resume, and are copied to forks and clones.
  - **Result:** pi's text (`Loaded N tools. They are available from your next
    call:` followed by `- name: first description line`). The tool-result
    message carries `AddedToolNames`, and the definitions join the running
    turn's tools, so the next model call declares them. go-ai uses the marker
    to load them at that point on providers with deferred tools; others get
    them in the normal tool list.
  - **Later turns** declare and allow every tool the session has loaded.
  - **Errors and defaults** are pi's: `query must not be empty`, `limit must
    be a positive integer`, a default limit of 8, and `No matching tools
    found.` when nothing matches.
- **`hidden` tools** are unreachable.
- **System prompt:** servers with codemode or deferred tools are listed in an
  `<mcp_servers>` section, using Pi's renderer (intro line, `- mcp__<server>
  (codemode|tool_search): <summary>`, 4096-character budget).
  - **Placement, as in Pi:** it is one of the system prompt's sections
    ([system-prompt.md](system-prompt.md)). Its value when the session starts
    leads the conversation; a later change (for example, a server connecting
    and its instructions becoming available) is added once, as a section
    update before the prompt that introduced it, so earlier messages stay
    cached.
- **Results** follow Pi's `convertMcpResult`:
  - Text and embedded text resources pass through.
  - Resource links name `read_mcp_resource`.
  - Empty content falls back to `structuredContent` as JSON.
  - An `isError` result becomes a tool error.
  - Text over 20 KB keeps its start and end around `…N chars truncated…`,
    with Pi's warning header.
  - **gi differences:**
    - The full text, and any binary resource, is saved to
      `vfs://mcp-output/<session>/<id><ext>`, where the `read` tool can open
      it. Pi uses temp files. Saved outputs are pruned after 7 days (checked at MCP
      start and every 6 hours); Pi leaves its temp files to the OS.
  - **Images** (image blocks and embedded image resources) are attached to
    the tool result as image blocks, as in Pi. go-ai replaces them with
    placeholders for models without image input, and the stored transcript
    adds an `[image <mime>, <size>]` line. Any tool can attach images through
    `ToolRuntime.AttachImage`.
- **Resource tools:** `list_mcp_resources`, `list_mcp_resource_templates` and
  `read_mcp_resource` are registered while an enabled, non-hidden server with
  resources has direct exposure.
  - Without a `server` argument, the list tools return every resource from
    every server; with one, they return a page plus `nextCursor`.
  - Listing and reading are retried once after a transient error.

## `gi mcp` CLI (`internal/mcp/cli.go`)

`gi mcp` is a port of `pi mcp` that works without starting a session.

- **Commands:** `add`, `remove`, `list [--json]`, `login` and `logout`. The options and messages are Pi's.
- **`add`:**
  - Writes the server entry in Pi's shape and key order to the user `mcp.json`, or with `-l`/`--local` to the project's.
  - Keeps the rest of the file and its indentation (an ordered JSON edit followed by re-indenting).
  - Validates the entry with the same rules as `LoadConfig`.
- **`remove`:** deletes an entry. When the server is defined in the other scope, the error says so.
- **`list`:**
  - Connects to every enabled server and prints its state, its tools (marking tools whose exposure differs from the server's), its resource counts and any errors.
  - A server with a project override shows `project override: <file>` (Pi 1.0.1).
  - Exits 1 when an entry is invalid or an enabled server does not connect.
- **Project config:** gi does not read project MCP configuration until it has project trust (#16), so the project file is reported as ignored.
- **Sign-in:** `login` and `logout` report that OAuth is not supported yet (#25 phase 6c).

## `/mcp` in the TUI (`internal/tui/mcp_command.go`)

- `/mcp` opens Pi's manager (`internal/tui/mcp_manager.go`, a port of
  `McpManagerView` and the `manage` loop):
  - The server list: servers needing attention first (needs sign-in, failed,
    disconnected, connecting, connected, disabled), each with its state,
    exposure and scope (`global` for the user mcp.json, `project`); config
    errors above the list.
  - A server's menu: its endpoint, scope and source, state and errors, and
    the actions Pi offers for its state: Sign in, Tools, Reconnect, Sign out,
    Exposure, Disable (Enable when disabled).
  - Tools: the server's tools with their first description line, marked
    `[exposure]` when `toolExposure` overrides the server's exposure.
  - Exposure: codemode, deferred or direct, saved to the mcp.json that defines
    the server, or to the project file that overrides it (other content and
    indentation kept, `Manager.UpdateServer`, `UpdateServerConfig`); a
    connected server's tools are registered again.
  - Enable/Disable are saved the same way; disabling closes the connection and
    withdraws the tools, enabling connects.
  - In a trusted project, a user server without an override also offers
    "Enable in this project"/"Disable in this project", which add an override
    to the project mcp.json (Pi 1.0.1). An override keeps the default values
    it writes (`"enabled": true`, `"exposure": "codemode"`), since they
    replace the user server's. Overridden servers show as `global, project
    override`, with `project override: <file>` in their details. Test:
    `TestMCPManagerProjectOverride`.
  - Menus are rebuilt from the engine's state on every frame and redraw when
    a server changes (`Engine.SetMCPChangeListener`); slow actions show Pi's
    status screen ("Reconnecting…") until they finish.
  - Up/Down wrap, Enter selects, Escape or Ctrl+C goes back (closes the list).
  - Golden: `scripts/golden-mcp-manager.mjs` renders Pi's own
    `McpManagerView`; `TestMCPManagerRenderMatchesPi`, and
    `TestMCPManagerManagesServers` against real servers.
  - Sign in runs Pi's `signInWithUi` in the manager: "Contacting the
    authorization server…", then the sign-in screen (the authorization URL
    with Pi's `AuthUrlComponent` hints, Ctrl+X copies it, and an input for
    the URL the browser was redirected to), then "Connecting…". Escape cancels;
    a failure shows on the server's screen. Golden: the sign-in screen in
    `scripts/golden-mcp-manager.mjs`.
  - Pi's `overridden:` notices (extension-registered servers) do not apply.
- `/mcp reconnect [server]` drops the connection, connects again and re-registers the server's tools (`Manager.Reconnect`, `Engine.MCPReconnect`).
  - Without a name it picks the only enabled server, or the only failed or disconnected one; otherwise it asks for a name.
- Argument completion (Pi's `getArgumentCompletions` for `/mcp`): after `/mcp `
  the slash menu lists `login`, `logout` and `reconnect`, then the eligible
  servers with their state (OAuth servers for `login`/`logout`, enabled servers
  for `reconnect`). Tab or Enter completes an argument without submitting.
  - The slash menu follows pi-tui's Editor: an open list refreshes on every
    edit; a closed one opens when `/` is typed at the start, when a letter,
    digit, `.`, `-`, `_` or CJK character is typed in a slash command, or when
    a character is deleted in one (also after Esc). Programmatic changes
    (history, drafts, paste) do not open it.
  - Golden: `scripts/golden-slash-autocomplete.mjs` replays pi-tui's own Editor
    with Pi's `/mcp` completion function; `TestSlashArgumentCompletionMatchesPi`.
  - Difference: Go has no Unicode `Script_Extensions`, so CJK punctuation
    shared between scripts (`、`, `ー`) does not open the list.
  - Pi's `/model`, `/thinking` and `/login` argument completions are not
    ported yet (#17).
- Server picker (Pi's `pickServer`): without a name, `login`/`logout` and
  `reconnect` use the only eligible server, else the only preferred one (the
  one needing sign-in; the failed or disconnected one), else ask with Pi's
  `ctx.ui.select` dialog ("MCP server"). Cancelling does nothing. An unknown
  name is reported as `No MCP server named "x".`
  - The dialog (`internal/tui/select_dialog.go`) ports
    `ExtensionSelectorComponent`: title, options, key hints between borders,
    word-wrapped like pi-tui `Text`; Up/Down or k/j move without wrapping,
    Enter selects, Escape or Ctrl+C cancels. Golden:
    `scripts/golden-select-dialog.mjs`, `TestSelectDialogMatchesPi`. The
    dialog is shown where gi shows its other Pi selectors.

## `/mcp` in the web UI (`internal/web/mcp_command.go`)

Pi's `/mcp` without a TUI. The prompt endpoint intercepts it like `/model`;
no turn runs. Replies are system messages in the session's timeline (kind
`mcp`, shown through `new_post`), which are not model context.

- `/mcp` replies with Pi's `formatStatus` (`mcp.FormatStatus`): one line per
  server with state, tool count and exposure, connection errors indented
  below, then config errors.
- `/mcp login|logout|reconnect [server]` use the TUI's server choice
  (`Engine.MCPPickServer`, Pi's `pickServer`) and messages. Reconnect and
  sign-in reply when they finish.
- Sign-in posts the authorization link (the server does not open a browser).
  If the browser cannot reach gi's loopback callback, `/mcp login <server>
  <redirect URL>` hands the URL it was sent to to the waiting sign-in.
- Advertised in `/api/quick-actions` when the server has a turn engine.
- Differences: Pi asks for a server with `ctx.ui.select` and for the redirect
  URL with `ctx.ui.input`; the web UI has neither, so an ambiguous name
  replies with the choices and the redirect URL is a `/mcp login` argument.
  The manager is TUI-only, as in Pi.
- Test: `TestWebMCPCommand` (status, sign-in with a pasted redirect,
  reconnect, sign-out, model context untouched); `TestFormatStatusMatchesPi`.

## OAuth (`internal/mcp/oauth.go`)

This is a port of Pi's MCP OAuth (pi-coding-agent `extensions/mcp/oauth.js`, pi-mcp `oauth/{flow,discovery,provider}.js`).

- **Which servers:** HTTP servers without `auth.provider` or an `Authorization` header (`ServerConfig.UsesOAuth`).
- **Credentials:** stored in Pi's `mcp-auth.json`, so gi and Pi share sign-ins. Entries are keyed per server as `<mcp namespace>|<url>` (Pi 1.0), so servers sharing a URL keep separate accounts; entries keyed by URL alone move to the first server that loads them. Writes are guarded by a `<file>.lock` directory; refreshes use per-server `mcp-auth-refresh-<hash>.lock` locks, as in Pi.
- **Connecting:**
  - The stored access token is sent, refreshed first when it is within 30 s of expiry.
  - After a 401 (or a 403 `insufficient_scope`), one refresh is shared by concurrent requests, then the request is retried once.
  - When the user must sign in, the server's state becomes `needs-auth`; connections never open a browser.
- **Sign-in** (`gi mcp login`): discovery (protected resource metadata, authorization server metadata with issuer check), dynamic client registration, PKCE S256 and a loopback callback on `127.0.0.1/callback`.
  - In a terminal, the redirect URL can also be pasted. It must match the sign-in's redirect URI (origin and path), and so must the callback's path.
  - **`oauth.clientRegistration: "cimd"`** (Pi 1.0.1): no registration; the client ID is pi's Client ID Metadata Document on pi.dev (gi identifies as pi, whose documents list its loopback callback). The authorization server must advertise `client_id_metadata_document_supported` and the `none` token endpoint auth method. With RFC 9207 `iss` support the document is `https://pi.dev/oauth/client.json` with the usual redirect URI; without it, `https://pi.dev/oauth/<id>/client.json` and `/callback/<id>`, where `<id>` is 12 base64url characters of SHA-256 of the server URL (Pi's `callbackId`), so responses from different authorization servers cannot be mixed up. The document is not stored, and a previously registered client is replaced. Test: `TestCLIOAuthClientIDMetadataDocument`.
  - `invalid_client` and `invalid_grant` errors reset credentials and retry, as in Pi.
  - **Issuer check (RFC 9207):** an authorization response whose `iss` does not name the authorization server, or that lacks `iss` when the server advertises it, is rejected before the code exchange.
  - **`oauth.authServerMetadataUrl`:** replaces discovery for servers that advertise a wrong authorization server or none. The document is trusted as configured and not cached.
  - **Step-up sign-in** (`insufficient_scope`): requests the granted scopes plus the challenged ones. Saved tokens record their scope (the requested scope when the response omits it, the grant's scope after a refresh).
  - **Empty or `null` optional token fields** count as absent, and `expires_in` may be a numeric string.
- **Sign-out** (`gi mcp logout`) deletes the stored credentials.
- **TUI** (`/mcp login|logout [server]`):
  - **Picking the server:** as in Pi: the named server, else the only OAuth server, else the only one needing sign-in.
  - **Signing in** (Pi 1.0.1): the manager opens on its sign-in flow (above) and the browser opens; it closes with the result.
    - Escape on the sign-in screen cancels (`Sign-in cancelled.`).
    - Success reconnects the server (`Signed in to MCP server "x" (N tools).`).
  - **Logout** deletes the credentials and leaves the server in `needs-auth`.
  - **`/mcp`** lists such servers as `x: needs sign-in, run /mcp login x`.
