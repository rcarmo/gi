#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/project-test-env.sh"
root="$GI_TEST_RUN_ROOT/fixtures/startup-hooks"
gi_test_path "$root"
mkdir -p "$root"
socket="$root/tmux.sock"
cleanup() { tmux -S "$socket" kill-server 2>/dev/null || true; }
trap cleanup EXIT
bin="${GI_STARTUP_PTY_BIN:?Make must provide the fixture binary}"
for spec in fullscreen:0: regular:0: fullscreen:1: fullscreen:0:true; do
  IFS=: read -r mode debug quiet <<< "$spec"
  name="$mode-$debug-${quiet:-normal}"
  dir="$root/$name"
  mkdir -p "$dir"
  args=(env PI_OFFLINE=1 PI_PROGRAM_STATUS=0 GI_STARTUP_PTY_DIR="$dir" GI_STARTUP_PTY_MODE="$mode" GI_STARTUP_PTY_DEBUG="$debug" GI_STARTUP_PTY_QUIET="$quiet" "$bin" -test.run '^TestStartupAndHookPTYFixture$' -test.timeout=30s)
  if [[ "${PROFILING:-0}" == 1 ]]; then args+=(-test.cpuprofile "$dir/cpu.pprof" -test.memprofile "$dir/mem.pprof"); fi
  printf -v command '%q ' "${args[@]}"
  tmux -S "$socket" -f /dev/null new-session -d -s "$name" -x 100 -y 35 "$command"
  tmux -S "$socket" set-option -g remain-on-exit on
  ready=0
  for _ in $(seq 1 200); do
    screen=$(tmux -S "$socket" capture-pane -p -t "$name" -S -100)
    if grep -q 'bootstrap' <<< "$screen"; then ready=1; break; fi
    sleep .05
  done
  [[ "$ready" == 1 ]] || { printf '%s\n' "$screen"; echo 'startup not ready'; exit 1; }
  if [[ "$quiet" == true ]]; then
    ! grep -q '█▀▀▀' <<< "$screen" || { echo 'quiet startup still has logo'; exit 1; }
  else
    grep -q '█▀▀▀ gi ' <<< "$screen"
    grep -q '█▄██' <<< "$screen"
  fi
  touch "$dir/fire-hooks"
  delivered=0
  for _ in $(seq 1 200); do
    screen=$(tmux -S "$socket" capture-pane -p -t "$name" -S -100)
    if grep -q 'fixture denial remains visible' <<< "$screen"; then delivered=1; break; fi
    sleep .05
  done
  [[ "$delivered" == 1 ]] || { printf '%s\n' "$screen"; echo 'hook denial missing'; exit 1; }
  for title in 'Hook invoked' 'Hook modified' 'Hook responded directly'; do
    if [[ "$debug" == 1 ]]; then grep -q "$title" <<< "$screen";
    else ! grep -q "$title" <<< "$screen" || { echo "unexpected routine audit: $title"; exit 1; }; fi
  done
  tmux -S "$socket" send-keys -t "$name" C-d
  exited=0
  for _ in $(seq 1 200); do
    if [[ "$(tmux -S "$socket" display-message -p -t "$name" '#{pane_dead}')" == 1 ]]; then exited=1; break; fi
    sleep .05
  done
  [[ "$exited" == 1 ]]
  [[ "$(tmux -S "$socket" display-message -p -t "$name" '#{pane_dead_status}')" == 0 ]]
  tmux -S "$socket" kill-session -t "$name"
  echo "PASS startup/hooks PTY: $spec"
done
