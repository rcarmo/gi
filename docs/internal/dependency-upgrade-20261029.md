# Dependency upgrade — 29 October 2026

Local branch: `review/dependency-upgrade-20261029`. No publication or deployment.

## Changes

- Go 1.27.2 → 1.27.3, required by Tailscale's patch release. Updated the
  Makefile toolchain and race-cache key with the module directive.
- go-ai 1.3.0 → 1.4.0.
- Goja 2cb8274eed06 → f81816846776 (dated pseudo-versions in go.mod).
- x/net 0.61.0 → 0.62.0; SQLite 1.60.1 → 1.61.0;
  Tailscale 1.104.1 → 1.104.4. Tidied their required transitive dependencies.
- Root and actual frontend manifests/lockfiles updated: Preact 11.1.0,
  Playwright 1.64.1, marked 18.1.2, KaTeX 0.19.1, Babel runtime 8.0.10,
  and available CodeMirror patch releases. Rebuilt the tracked frontend bundle.

## Verification

Passed: Go tests for scripting, codemode, store, auth and web; selected turn
provider/retry/MCP/tool-search checks; CLI and frontend builds; 46 focused
frontend unit tests; one WebKit tablet model-focus/Enter regression (6.9s).
No full local matrix, Actions run, profiling or release verification performed.

Browser setup initially lacked the matching WebKit binary. After installing it,
a run inherited PI_CODING_AGENT_DIR and used the wrong fixture credentials.
The correctly isolated run passed with both agent-directory overrides unset.
Those setup failures are not product regressions. No assertions changed.

## Deliberate exclusions

`go-tui` remains the locally patched v0.22.1 replacement. Upstream v0.23.0 is
available, but updating the require line would not update the replacement.
Its terminal/paste/redraw extensions need a separate fork integration; this
upgrade does not discard them. Vendored Piclaw source is unchanged. Transitive
modules are not independently bumped beyond the selected dependencies' needs.
