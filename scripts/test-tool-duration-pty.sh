#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/project-test-env.sh"
root="$GI_TEST_RUN_ROOT/duration-pty"
gi_test_path "$root"
mkdir -p "$root/agent" "$root/workspace"
session="gi-duration-$$"
cleanup() {
 tmux kill-session -t "$session" 2>/dev/null || true
 rm -rf -- "$root"
}
trap cleanup EXIT
export PI_CODING_AGENT_DIR="$root/agent" GI_CODING_AGENT_DIR="$root/agent"
export GI_DURATION_PTY_DIR="$root/workspace" TERM=xterm-256color
printf '{"quietStartup":true,"theme":"dark"}\n' > "$root/agent/settings.json"
go test -c -o "$root/test-bin" ./internal/tui
tmux new-session -d -s "$session" -x 100 -y 24 "'$root/test-bin' -test.run '^TestRecordedDurationPTYFixture$'; printf '\\nPTY-EXIT:%s\\n' \"\$?\"; sleep 2"
for _ in $(seq 1 100); do
 tmux capture-pane -t "$session" -p > "$root/screen"
 if grep -q 'Took 1.5s' "$root/screen" && grep -q 'recorded output' "$root/screen"; then break; fi
 sleep .1
done
grep -q 'Took 1.5s' "$root/screen"
grep -q 'recorded output' "$root/screen"
tmux send-keys -t "$session" C-d
for _ in $(seq 1 100); do
 tmux capture-pane -t "$session" -p > "$root/screen" 2>/dev/null || break
 grep -q 'PTY-EXIT:0' "$root/screen" && { echo 'Recorded Took PTY: PASS'; exit 0; }
 sleep .05
done
echo 'TUI did not exit cleanly' >&2
exit 1
