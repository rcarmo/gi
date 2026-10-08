#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/project-test-env.sh"
BUN=${BUN:-bun}
export PLAYWRIGHT_JSON_OUTPUT_NAME=${PLAYWRIGHT_JSON_OUTPUT_NAME:-$GI_TEST_RUN_ROOT/results/results.json}
mkdir -p "$(dirname "$PLAYWRIGHT_JSON_OUTPUT_NAME")"

if command -v node >/dev/null 2>&1; then
  if [[ "${PROFILING:-0}" == 1 ]]; then
    mkdir -p "$GI_TEST_RUN_ROOT/profiles/node"
    export NODE_OPTIONS="${NODE_OPTIONS:-} --cpu-prof --cpu-prof-dir=$GI_TEST_RUN_ROOT/profiles/node --heap-prof --heap-prof-dir=$GI_TEST_RUN_ROOT/profiles/node"
  fi
  cli="$GI_SCRIPT_ROOT/node_modules/@playwright/test/cli.js"
  for arg in "$@"; do
    if [[ "$arg" == *playwright.fixtures.config.ts* ]]; then cli="$GI_SCRIPT_ROOT/references/fixtures-vibes/node_modules/@playwright/test/cli.js"; fi
  done
  exec node "$cli" "$@"
fi

# Playwright's installed CLI has a `#!/usr/bin/env node` shebang. Bun can run it
# directly, so expose an ephemeral node-compatible command without modifying the
# host or repository.
shim=$(mktemp -d "$TMPDIR/node-shim.XXXXXX")
cleanup() { rm -rf "$shim"; }
trap cleanup EXIT
ln -s "$(command -v "$BUN")" "$shim/node"
PATH="$shim:$PATH" "$BUN" x playwright "$@"
