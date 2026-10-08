#!/bin/sh
# Disposable fixture-only build. The production binary never registers this model.
set -eu
: "${FIXTURES_ROOT:?}"
: "${FIXTURE_MODEL_URL:?}"
: "${FIXTURE_MODEL_ID:?}"
[ "$FIXTURE_MODEL_ID" = 'fixture-1' ] || { echo 'Unexpected fixture model' >&2; exit 1; }
# The suite owns an isolated directory beneath the resolved project run root.
: "${FIXTURES_RUN_ROOT:?}"
case "$FIXTURES_ROOT" in *'/../'*|*'/./'*|*'//'*) echo 'Unsafe fixture traversal' >&2; exit 1 ;; esac
case "$FIXTURES_ROOT" in "$FIXTURES_RUN_ROOT"/fixtures-gi-*) ;; *) echo 'Refusing fixture outside the owned run root' >&2; exit 1;; esac
# Reject symlink/ownership escapes before any writes, including parent components.
current=$FIXTURES_ROOT
while [ "$current" != / ]; do
  [ ! -L "$current" ] && [ -d "$current" ] && [ -O "$current" ] || { echo "Unsafe fixture path: $current" >&2; exit 1; }
  [ "$current" != "$FIXTURES_RUN_ROOT" ] || break
  current=$(dirname "$current")
done
mkdir -p "$FIXTURES_ROOT/bin" "$FIXTURES_ROOT/workspace/.pi" "$FIXTURES_ROOT/workspace/.piclaw" "$FIXTURES_ROOT/workspace/.gi/skills/proof" "$FIXTURES_ROOT/home/.pi/agent"
# Project-owned scratch paths are long. Keep the fixture prompt short so a
# command-echo assertion measures input, not wrapping caused by directory names.
# This is isolated HOME only; production shells retain their normal startup.
printf '%s\n' 'PS1="$ "' > "$FIXTURES_ROOT/home/.bashrc"
cp "$(dirname "$0")/../ux/fixtures/skills/.gi/skills/proof/SKILL.md" "$FIXTURES_ROOT/workspace/.gi/skills/proof/SKILL.md"
mkdir -p "$FIXTURES_ROOT/workspace/.gi/skills/review"
printf '%s\n' '---' 'name: review' 'description: Examine the vermilion verification marker' '---' 'Canonical skill body: VERMILION_NATIVE_SKILL_V1.' > "$FIXTURES_ROOT/workspace/.gi/skills/review/SKILL.md"
# Shared editor specs open a visible workspace README.
printf '%s\n' '# Fixture workspace' '' 'Sample workspace file.' > "$FIXTURES_ROOT/workspace/README.md"
# Shared compose specs reference a visible (non-root) workspace folder.
mkdir -p "$FIXTURES_ROOT/workspace/notes"
printf '%s\n' '# Notes' '' 'Fixture folder.' > "$FIXTURES_ROOT/workspace/notes/index.md"
printf '%s\n' '{"defaultProvider":"fixture-vibes","defaultModel":"fixture-vibes/fixture-1","defaultThinkingLevel":"low","enabledModels":["fixture-vibes/fixture-1","fixture-vibes/fixture-2"],"maxIterations":4,"inboundWork":{"enabled":false}}' > "$FIXTURES_ROOT/workspace/.pi/settings.json"
printf '%s\n' '{"assistant":{"assistantName":"Gi Fixture"},"user":{"userName":"Fixture User"}}' > "$FIXTURES_ROOT/workspace/.piclaw/config.json"
printf '%s\n' '{"fixture-vibes":{"type":"api_key","apiKey":"fixture-only"}}' > "$FIXTURES_ROOT/home/.pi/agent/auth.json"
# The Makefile builds once before Playwright starts. Replacement workers only
# copy that fixture build, so compilation cannot exhaust their setup timeout.
if [ -n "${GI_FIXTURE_BIN:-}" ]; then
  cp "$GI_FIXTURE_BIN" "$FIXTURES_ROOT/bin/gi"
else
  cd "$(dirname "$0")/../.."
  go build -tags fixtures_vibes -o "$FIXTURES_ROOT/bin/gi" ./cmd/gi
fi
