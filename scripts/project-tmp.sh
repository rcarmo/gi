#!/usr/bin/env bash
# Vendored project-owned scratch resolver; never removes files.
set -euo pipefail
if [[ -z "${PROJECT_ORIGINAL_TMPDIR+x}" ]]; then
  export PROJECT_ORIGINAL_TMPDIR="${TMPDIR:-}"
fi
project_is_ci() {
  case "${CI:-}" in ''|0|false|FALSE) ;; *) return 0;; esac
  case "${GITHUB_ACTIONS:-}:${GITLAB_CI:-}:${TF_BUILD:-}:${CIRCLECI:-}" in *true*|*True*|*TRUE*) return 0;; esac
  return 1
}
project_path_usable() {
  local path="$1" parent ancestor
  case "$path" in /*) ;; *) return 1;; esac
  case "/${path#/}/" in */../*|*/./*) return 1;; esac
  [[ ! -L "$path" ]] || return 1
  ancestor="${path%/*}"
  while [[ -n "$ancestor" && "$ancestor" != / ]]; do
    [[ ! -L "$ancestor" || "$ancestor" == /workspace ]] || return 1
    ancestor="${ancestor%/*}"
  done
  if [[ -e "$path" ]]; then [[ -d "$path" && -O "$path" && -w "$path" && -x "$path" ]] || return 1; fi
  parent="$path"
  while [[ ! -e "$parent" && ! -L "$parent" ]]; do parent="${parent%/*}"; [[ -n "$parent" ]] || parent=/; done
  [[ -d "$parent" && -w "$parent" && -x "$parent" ]] || return 1
}
project_tmp_resolve() {
  local project="$1" workspace_base="${2:-/workspace/tmp}" base candidate explicit=''
  case "$project" in ''|*[!A-Za-z0-9._-]*|.*|-*) echo 'Invalid project name' >&2; return 1;; esac
  if [[ -n "${PROJECT_TMP_BASE+x}" ]]; then
    [[ -n "$PROJECT_TMP_BASE" ]] || { echo 'Empty PROJECT_TMP_BASE' >&2; return 1; }
    explicit="${PROJECT_TMP_BASE%/}/$project"
    project_path_usable "$explicit" || { echo 'PROJECT_TMP_BASE must be usable and absolute' >&2; return 1; }
  fi
  if [[ -n "${PROJECT_TMP_ROOT+x}" ]]; then
    candidate="${PROJECT_TMP_ROOT%/}"
    [[ "${candidate##*/}" == "$project" ]] && project_path_usable "$candidate" || { echo 'PROJECT_TMP_ROOT must be absolute, usable and project-named' >&2; return 1; }
    [[ -z "$explicit" || "$explicit" == "$candidate" ]] || { echo 'Conflicting PROJECT_TMP_BASE and PROJECT_TMP_ROOT' >&2; return 1; }
    printf '%s\n' "$candidate"; return
  fi
  if [[ -n "$explicit" ]]; then printf '%s\n' "$explicit"; return; fi
  local bases=()
  if project_is_ci; then bases=("${RUNNER_TEMP:-}" "$PROJECT_ORIGINAL_TMPDIR" /tmp); else bases=("$workspace_base" /tmp); fi
  for base in "${bases[@]}"; do
    [[ -n "$base" ]] || continue
    candidate="${base%/}/$project"
    if project_path_usable "$candidate"; then printf '%s\n' "$candidate"; return; fi
  done
  echo 'No usable project temporary root' >&2; return 1
}
if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  root="$(project_tmp_resolve "${PROJECT_NAME:-gi}")"
  printf '%s\n' "$root"
fi
