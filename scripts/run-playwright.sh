#!/usr/bin/env bash
set -euo pipefail

BUN=${BUN:-bun}

if command -v node >/dev/null 2>&1; then
  exec "$BUN" x playwright "$@"
fi

# Playwright's installed CLI has a `#!/usr/bin/env node` shebang. Bun can run it
# directly, so expose an ephemeral node-compatible command without modifying the
# host or repository.
shim=$(mktemp -d)
cleanup() { rm -rf "$shim"; }
trap cleanup EXIT
ln -s "$(command -v "$BUN")" "$shim/node"
PATH="$shim:$PATH" "$BUN" x playwright "$@"
