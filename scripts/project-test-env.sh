#!/usr/bin/env bash
# Source from owned helpers before creating scratch or launching child tools.
GI_SCRIPT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
export PROJECT_NAME=gi
export PROJECT_ORIGINAL_TMPDIR=${PROJECT_ORIGINAL_TMPDIR-${TMPDIR-}}
PROJECT_TMP_ROOT=$(bash "$GI_SCRIPT_ROOT/scripts/project-tmp.sh") || return 1
export PROJECT_TMP_ROOT
export GI_WORKTREE=${GI_WORKTREE:-$(basename "$GI_SCRIPT_ROOT")}
export GI_TEST_RUN_ROOT=${GI_TEST_RUN_ROOT:-$PROJECT_TMP_ROOT/runs/tests/$GI_WORKTREE/$(date -u +%Y%m%dT%H%M%SZ)-$$}

# A destructive test helper may touch only an owned, non-symlink descendant of
# this invocation's run directory. Check every existing component, not just leaf.
gi_test_path() {
  local path=$1 current
  case "$path" in "$GI_TEST_RUN_ROOT"|"$GI_TEST_RUN_ROOT"/*) ;; *) echo "Refusing test path outside run: $path" >&2; return 1 ;; esac
  case "$path" in *'/../'*|*'/./'*|*/..|*/.|*'//'*) echo "Unsafe test path: $path" >&2; return 1 ;; esac
  current=$path
  while [[ "$current" != / ]]; do
    [[ ! -L "$current" ]] || { echo "Symlink test path: $current" >&2; return 1; }
    if [[ -e "$current" ]]; then
      [[ -d "$current" && -O "$current" ]] || { echo "Unowned/non-directory test path: $current" >&2; return 1; }
    fi
    [[ "$current" != "$PROJECT_TMP_ROOT" ]] || break
    current=$(dirname "$current")
  done
}
case "$GI_TEST_RUN_ROOT" in "$PROJECT_TMP_ROOT"/runs/*) ;; *) echo 'GI_TEST_RUN_ROOT must be beneath the project runs directory' >&2; return 1 ;; esac
gi_test_path "$GI_TEST_RUN_ROOT" || return 1
export TMPDIR="$GI_TEST_RUN_ROOT/tmp" TMP="$GI_TEST_RUN_ROOT/tmp" TEMP="$GI_TEST_RUN_ROOT/tmp"
export BUN_INSTALL_CACHE_DIR="$PROJECT_TMP_ROOT/cache/bun" npm_config_cache="$PROJECT_TMP_ROOT/cache/npm" XDG_CACHE_HOME="$PROJECT_TMP_ROOT/cache/xdg"
export GOCACHE="$PROJECT_TMP_ROOT/cache/go-build" GOMODCACHE="$PROJECT_TMP_ROOT/cache/go-mod" GOTMPDIR="$TMPDIR/go"
export PLAYWRIGHT_BROWSERS_PATH=${PLAYWRIGHT_BROWSERS_PATH:-$PROJECT_TMP_ROOT/cache/ms-playwright}
mkdir -p "$TMPDIR" "$GOTMPDIR" "$BUN_INSTALL_CACHE_DIR" "$npm_config_cache" "$XDG_CACHE_HOME" "$GOCACHE" "$GOMODCACHE"
