#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
source "$ROOT/scripts/project-test-env.sh"
ARTIFACT_DIR="${ARTIFACT_DIR:-$GI_TEST_RUN_ROOT/results/tui-smoke}"
SESSION="gi-tui-smoke-$$"
TEST_DIR="${TEST_DIR:-$GI_TEST_RUN_ROOT/tui-smoke}"
gi_test_path "$ARTIFACT_DIR"
gi_test_path "$TEST_DIR"
GI_BIN=${GI_BIN:-$PROJECT_TMP_ROOT/build/$GI_WORKTREE/gi}
DB="$TEST_DIR/gi.db"
WORKSPACE="$TEST_DIR/workspace"
OVERLAY_UPPER="$TEST_DIR/overlay-upper"
OVERLAY_WORK="$TEST_DIR/overlay-work"
MOUNTED=none
# Read-only source the isolated workspace is layered over (never written).
SMOKE_LOWER="${SMOKE_LOWER:-/workspace}"

cleanup() {
  tmux kill-session -t "$SESSION" >/dev/null 2>&1 || true
  if mountpoint -q "$WORKSPACE" 2>/dev/null; then
    case "$MOUNTED" in
      kernel) sudo -n umount "$WORKSPACE" >/dev/null 2>&1 || true ;;
      fuse) fusermount3 -u "$WORKSPACE" >/dev/null 2>&1 || fusermount -u "$WORKSPACE" >/dev/null 2>&1 || true ;;
    esac
  fi
  rm -rf "$TEST_DIR"
}
trap cleanup EXIT

rm -rf "$ARTIFACT_DIR" "$TEST_DIR"
mkdir -p "$ARTIFACT_DIR" "$WORKSPACE" "$OVERLAY_UPPER" "$OVERLAY_WORK"
# Isolate the workspace from $SMOKE_LOWER. Prefer a kernel overlay; fall back
# where overlayfs, root or the lower dir are unavailable (containers, VMs
# without /workspace). The smoke only needs its seeded files, so every mode
# is equivalent for the assertions below; the chosen mode is recorded.
setup_workspace() {
  local opts="lowerdir=$SMOKE_LOWER,upperdir=$OVERLAY_UPPER,workdir=$OVERLAY_WORK"
  if [[ -d "$SMOKE_LOWER" ]]; then
    if grep -qw overlay /proc/filesystems 2>/dev/null &&
      sudo -n mount -t overlay overlay -o "$opts" "$WORKSPACE" 2>/dev/null; then
      MOUNTED=kernel; echo "kernel-overlay lower=$SMOKE_LOWER"; return
    fi
    if command -v fuse-overlayfs >/dev/null 2>&1 && [[ -e /dev/fuse ]] &&
      fuse-overlayfs -o "$opts" "$WORKSPACE" 2>/dev/null; then
      MOUNTED=fuse; echo "fuse-overlay lower=$SMOKE_LOWER"; return
    fi
    if [[ -n "${SMOKE_COPY_LOWER:-}" ]]; then
      cp -a --reflink=auto "$SMOKE_LOWER/." "$WORKSPACE/"
      echo "copy lower=$SMOKE_LOWER"; return
    fi
  fi
  echo "scratch (lower=$SMOKE_LOWER unavailable or not layerable)"
}
WORKSPACE_MODE=$(setup_workspace)
# setup_workspace ran in a subshell; recover the mount kind for cleanup.
case "$WORKSPACE_MODE" in kernel-overlay*) MOUNTED=kernel ;; fuse-overlay*) MOUNTED=fuse ;; esac
echo "$WORKSPACE_MODE" > "$ARTIFACT_DIR/workspace-mode.txt"
echo "smoke workspace: $WORKSPACE_MODE"
mkdir -p "$WORKSPACE/.pi"

cat > "$WORKSPACE/.pi/settings.json" <<'JSON'
{"defaultProvider":"test","defaultModel":"test-model","defaultThinkingLevel":"low","enabledModels":["test-model"]}
JSON
cat > "$WORKSPACE/AGENTS.md" <<'MD'
You are Gi Test.
MD

cd "$ROOT"
mkdir -p "$(dirname "$GI_BIN")"
if [[ ! -x "$GI_BIN" ]]; then
  go build -o "$GI_BIN" ./cmd/gi
fi

# Start detached with a fixed terminal size for deterministic mouse coordinates.
# User agent dirs are test-local, so the host's mcp.json, auth and skills
# (~/.gi/agent, ~/.pi/agent) cannot change what the smoke test sees.
AGENT_DIR="$TEST_DIR/agent"
mkdir -p "$AGENT_DIR/gi" "$AGENT_DIR/pi"
tmux new-session -d -x 100 -y 20 -s "$SESSION" "cd '$ROOT' && GI_CODING_AGENT_DIR='$AGENT_DIR/gi' PI_CODING_AGENT_DIR='$AGENT_DIR/pi' '$GI_BIN' -db '$DB' -workspace '$WORKSPACE'"
for _ in 1 2 3 4 5; do
  sleep 1
  tmux capture-pane -pe -t "$SESSION":0 > "$ARTIFACT_DIR/01-start.txt"
  if grep -q "%/" "$ARTIFACT_DIR/01-start.txt"; then
    break
  fi
done

if ! grep -q "%/" "$ARTIFACT_DIR/01-start.txt"; then
  echo "TUI did not render bottom-band session counters" >&2
  exit 1
fi
# An empty session shows gi's startup header (or, with quietStartup true,
# the empty-transcript placeholder).
if ! sed 's/\x1b\[[0-9;]*m//g' "$ARTIFACT_DIR/01-start.txt" | grep -Eq "\(no messages yet\)|Ctrl\+O to show full startup help"; then
  echo "TUI did not render the empty transcript" >&2
  exit 1
fi
if [[ ! -f "$DB" ]]; then
  echo "TUI did not create its session database" >&2
  exit 1
fi
sqlite3 "$DB" 'select count(*) from sessions;' > "$ARTIFACT_DIR/01-session-count.txt"

# Submit a prompt and verify it round-trips through the store.
tmux send-keys -t "$SESSION":0 "hello from tmux" Enter
for _ in 1 2 3 4 5 6; do
  sleep 1
  sqlite3 -separator '|' "$DB" 'select role, content from messages order by created_at asc, id asc;' > "$ARTIFACT_DIR/02-messages-after-submit.txt"
  if grep -q 'assistant|Gi received: hello from tmux' "$ARTIFACT_DIR/02-messages-after-submit.txt"; then
    break
  fi
done
if ! grep -q 'user|hello from tmux' "$ARTIFACT_DIR/02-messages-after-submit.txt"; then
  echo "TUI did not persist submitted user input" >&2
  exit 1
fi
if ! grep -q 'assistant|Gi received: hello from tmux' "$ARTIFACT_DIR/02-messages-after-submit.txt"; then
  echo "TUI did not persist assistant response" >&2
  exit 1
fi

# Idle Escape keeps the editor focused (Pi): typing lands in the editor and
# nothing is submitted until Enter.
tmux send-keys -t "$SESSION":0 Escape
sleep 1
BEFORE_COUNT=$(sqlite3 "$DB" 'select count(*) from messages;')
tmux send-keys -t "$SESSION":0 -l "still focused draft"
sleep 1
tmux capture-pane -p -t "$SESSION":0 > "$ARTIFACT_DIR/03-after-escape.txt"
AFTER_ESCAPE_COUNT=$(sqlite3 "$DB" 'select count(*) from messages;')
if ! grep -q 'still focused draft' "$ARTIFACT_DIR/03-after-escape.txt"; then
  echo "TUI editor lost focus after idle Escape" >&2
  exit 1
fi
if [[ "$BEFORE_COUNT" != "$AFTER_ESCAPE_COUNT" ]]; then
  echo "TUI submitted input without Enter" >&2
  exit 1
fi
tmux send-keys -t "$SESSION":0 C-u
sleep 1

# A multi-line bracketed paste lands in the editor as one edit (Pi); its
# newlines must not submit. Enter then sends all lines as one message.
BEFORE_PASTE=$(sqlite3 "$DB" 'select count(*) from messages;')
tmux set-buffer -b gi-paste "$(printf 'pasted line one\npasted line two\npasted line three')"
tmux paste-buffer -p -b gi-paste -t "$SESSION":0
sleep 1
tmux capture-pane -p -t "$SESSION":0 > "$ARTIFACT_DIR/03b-after-paste.txt"
if [[ "$(sqlite3 "$DB" 'select count(*) from messages;')" != "$BEFORE_PASTE" ]]; then
  echo "TUI submitted a pasted line (bracketed paste not handled)" >&2
  exit 1
fi
if ! grep -q 'pasted line three' "$ARTIFACT_DIR/03b-after-paste.txt"; then
  echo "TUI editor did not show the pasted lines" >&2
  exit 1
fi
tmux send-keys -t "$SESSION":0 Enter
for _ in 1 2 3 4 5; do
  sleep 1
  if [[ "$(sqlite3 "$DB" "select count(*) from messages where role='user' and content = 'pasted line one' || char(10) || 'pasted line two' || char(10) || 'pasted line three';")" == 1 ]]; then
    break
  fi
done
if [[ "$(sqlite3 "$DB" "select count(*) from messages where role='user' and content = 'pasted line one' || char(10) || 'pasted line two' || char(10) || 'pasted line three';")" != 1 ]]; then
  sqlite3 -separator '|' "$DB" 'select role, content from messages;' > "$ARTIFACT_DIR/03c-paste-messages.txt"
  echo "TUI did not submit the paste as one message" >&2
  exit 1
fi

# Exercise transcript scrolling keys before resizing.
tmux send-keys -t "$SESSION":0 PageUp
sleep 1
tmux capture-pane -pe -t "$SESSION":0 > "$ARTIFACT_DIR/04-after-pageup.txt"
tmux send-keys -t "$SESSION":0 End
sleep 1
tmux capture-pane -pe -t "$SESSION":0 > "$ARTIFACT_DIR/05-after-end.txt"
if ! grep -q 'Gi received: pasted line one' "$ARTIFACT_DIR/05-after-end.txt"; then
  echo "TUI did not restore transcript bottom after End" >&2
  exit 1
fi

# Resize and verify the session stays alive after interaction.
tmux resize-window -t "$SESSION":0 -x 120 -y 24
sleep 1
tmux has-session -t "$SESSION"
tmux capture-pane -pe -t "$SESSION":0 > "$ARTIFACT_DIR/06-after-resize.txt"
if ! grep -q '%/' "$ARTIFACT_DIR/06-after-resize.txt"; then
  echo "TUI lost the footer context meter after resize" >&2
  exit 1
fi
if ! grep -q 'Gi received: pasted line one' "$ARTIFACT_DIR/06-after-resize.txt"; then
  echo "TUI did not render transcript content after resize" >&2
  exit 1
fi

# Persist final server-side state for inspection.
sqlite3 "$DB" 'select id, title, state_json from sessions;' > "$ARTIFACT_DIR/07-session-state.txt" 2>/dev/null || true

echo "TUI smoke test passed. Artifacts: $ARTIFACT_DIR"
