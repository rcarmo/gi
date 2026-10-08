#!/usr/bin/env bash
set -euo pipefail
ROOT=$(cd "$(dirname "$0")/.." && pwd)
source "$ROOT/scripts/project-test-env.sh"
scratch="$GI_TEST_RUN_ROOT/path-tests"
gi_test_path "$scratch"
mkdir -p "$scratch"
trap 'rm -rf "$scratch"' EXIT
resolver="$ROOT/scripts/project-tmp.sh"
check() {
  local expected=$1; shift
  local actual
  actual=$(env -u PROJECT_TMP_ROOT -u PROJECT_TMP_BASE -u PROJECT_ORIGINAL_TMPDIR -u TMPDIR -u CI -u RUNNER_TEMP PROJECT_NAME=gi "$@" bash "$resolver")
  [[ "$actual" == "$expected" ]] || { echo "$actual != $expected" >&2; exit 1; }
}
check "$scratch/base/gi" PROJECT_TMP_BASE="$scratch/base"
check "$scratch/root/gi" PROJECT_TMP_ROOT="$scratch/root/gi"
check "$scratch/base/gi" PROJECT_TMP_BASE="$scratch/base" PROJECT_TMP_ROOT="$scratch/base/gi"
check "$scratch/runner/gi" CI=1 RUNNER_TEMP="$scratch/runner" PROJECT_ORIGINAL_TMPDIR="$scratch/original"
check "$scratch/original/gi" CI=1 PROJECT_ORIGINAL_TMPDIR="$scratch/original"
actual=$(env -u PROJECT_TMP_ROOT -u PROJECT_TMP_BASE bash -c 'source "$1"; project_tmp_resolve gi "$2"' _ "$resolver" "$scratch/generic")
[[ "$actual" == "$scratch/generic/gi" ]]
actual=$(env -u PROJECT_TMP_ROOT -u PROJECT_TMP_BASE CI=0 bash -c 'source "$1"; project_tmp_resolve gi /dev/null/unusable' _ "$resolver")
[[ "$actual" == /tmp/gi ]]
check /tmp/gi CI=1
for args in 'PROJECT_TMP_BASE=relative' "PROJECT_TMP_ROOT=$scratch/not-gi" "PROJECT_TMP_BASE=$scratch/base PROJECT_TMP_ROOT=$scratch/other/gi"; do
  if env -u PROJECT_TMP_ROOT -u PROJECT_TMP_BASE $args PROJECT_NAME=gi bash "$resolver" >/dev/null 2>&1; then echo 'invalid override accepted' >&2; exit 1; fi
done
mkdir -p "$scratch/escape"
ln -s "$scratch/escape" "$scratch/link"
if env -u PROJECT_TMP_ROOT PROJECT_TMP_BASE="$scratch/link" PROJECT_NAME=gi bash "$resolver" >/dev/null 2>&1; then echo 'symlink accepted' >&2; exit 1; fi
if gi_test_path "$scratch/link/child" >/dev/null 2>&1; then echo 'helper symlink accepted' >&2; exit 1; fi
if gi_test_path "$PROJECT_TMP_ROOT/cache" >/dev/null 2>&1; then echo 'out-of-run cleanup accepted' >&2; exit 1; fi
# A second helper inherits the resolved project root, never TMPDIR/gi nesting.
actual=$(PROJECT_TMP_ROOT="$PROJECT_TMP_ROOT" GI_TEST_RUN_ROOT="$GI_TEST_RUN_ROOT" bash -c 'source "$1"; printf "%s" "$PROJECT_TMP_ROOT"' _ "$ROOT/scripts/project-test-env.sh")
[[ "$actual" == "$PROJECT_TMP_ROOT" ]]
# Fixture guard rejects outside-root and symlink paths before touching them.
for bad in "$scratch/escape" "$scratch/link" "$scratch/../fixtures-gi-escape"; do
  if FIXTURES_ROOT="$bad" FIXTURES_RUN_ROOT="$scratch" FIXTURE_MODEL_URL=http://127.0.0.1:1 FIXTURE_MODEL_ID=fixture-1 sh "$ROOT/tests/fixtures-vibes/prepare.sh" >/dev/null 2>&1; then echo 'unsafe fixture accepted' >&2; exit 1; fi
done
echo 'Project workspace/override/CI/generic/non-nesting and isolation checks passed'
