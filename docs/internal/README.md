# Internal reference

Gi embeds this tree as the read-only `vfs://reference/...` documentation for agents and runtime developers. Feature implementation notes, plans and verification reports live in [implementation notes](../implementation/README.md). Current release status is in the [feature matrix](../feature-parity.md).

## Contract maintenance

Update the relevant reference when a change affects tools, scripting bridges, hooks, managed VFS behaviour, skill/package structure or other agent-visible extension points. Keep those reference paths stable. Put revision-specific findings and feature implementation history under `docs/implementation/`.

## Topic trees

- [Tools](tools/README.md) — built-in tool parameters, outputs and side effects
- [Scripting](scripting/README.md) — native Joker, Goja/QuickJS and bridge APIs
- [Hooks](hooks/README.md) — lifecycle and mutation rules
- [VFS](vfs/README.md) — namespaces, resolution and read-only reference URLs
- [Skills](skills/README.md) — discovery, packaging and conventions
- [Search](search/README.md) — scoped lexical indexing and the planned hybrid/vector design

## Runtime

- [Codemode scripts (gi reference)](codemode-scripts.md)
- [Codemode engine (`internal/codemode`)](codemode.md)
- [Compaction (#18)](compaction.md)
- [Configuration file lookup (#26)](config-files.md)
- [Connectivity hooks and route registration](connectivity-hooks.md)
- [Extension command registration semantics](extension-command-semantics.md)
- [Held-turn retry admission](held-turn-retry.md)
- [Keychain](keychain.md)
- [MCP client (`internal/mcp`)](mcp.md)
- [Shared media ingestion contract](media-ingestion-contract.md)
- [Bounded current-session message retrieval](message-retrieval.md)
- [Routing and route-event introspection](routing.md)
- [ADR: SQLite-backed coordinated runtime model](runtime-refactor-adr.md)
- [Runtime surfaces living index](runtime-surfaces.md)
- [Runtime target state](runtime-target-state.md)
- [Session import, export and sharing](session-import.md)
- [Session Plan backend](session-plan.md)
- [Session thinking selection](session-thinking.md)
- [Shell environment](shell-environment.md)
- [Sub-turn runtime contract](subturn-runtime.md)
- [System prompt (#32)](system-prompt.md)
- [Recorded tool execution duration](tool-duration.md)
- [Tool terminal provenance](tool-terminal-provenance.md)
- [Internal topic system design](topic-system.md)

## Terminal

- [Keybindings](keybindings.md)
- [Pi terminal and Piclaw queue contract](pi-tui-piclaw-queue-contract.md)
- [TUI clipboard and media](tui-clipboard-media.md)
- [Explicit terminal held retry](tui-held-retry.md)
- [Durable terminal attachment references](tui-media-journal.md)
- [TUI /login and /logout](tui-oauth-login.md)
- [TUI paste](tui-paste.md)
- [Gi TUI Pi-identical layout contract](tui-pi-layout-contract.md)
- [Explicit terminal queue commands](tui-queue-commands.md)
- [Gi TUI searchable selectors (PiSwift port)](tui-selectors.md)
- [TUI settings](tui-settings.md)
- [Gi TUI single-line status semantics](tui-status-line-semantics.md)
- [Terminal theme detection (issues #12, #31)](tui-terminal-theme.md)
- [Terminal plaintext journal prerequisite](tui-text-journal.md)
- [TUI /tree](tui-tree.md)

## Web

- [Browser-owner session proof](browser-auth-proof.md)
- [Browser-bound initial owner setup](browser-bootstrap.md)
- [Dashboard widget backend](dashboard-widgets.md)
- [Message deletion](message-deletion.md)
- [Piclaw read-only pane host](pane-host-subset.md)
- [Single-user passkey Settings contract](passkey-contract.md)
- [Passkey criterion ledger](passkey-criteria.md)
- [Passkey login cancellation boundary](passkey-login-cancellation.md)
- [Multi-passkey authentication](passkeys.md)
- [Revision-safe editor and Plan writes](revision-safe-writes.md)
- [`/btw`](side-prompt.md)
- [Cross-tab draft write fencing](web-cross-tab-drafts.md)
- [Basic HTTP delivery and run-control checks](web-delivery-recovery.md)
- [Basic web send on HTTP hosts](web-http-send.md)
- [Bounded web send receipts](web-send-receipts.md)
- [Web Stop preserves pending work](web-stop-queue.md)
- [Web terminal backend](web-terminal.md)
- [Web VNC backend](web-vnc.md)
- [Workspace editor backend](workspace-editor-backend.md)

## Development

- [Profiling gi](profiling.md)
- [Shipped internal reference system](reference-system.md)
- [Runtime package inventory](runtime-package-inventory.md)
