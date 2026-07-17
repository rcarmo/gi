#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
ARTIFACT_DIR="${ARTIFACT_DIR:-$ROOT/test-results/tui-smoke}"
SESSION="gi-tui-smoke-$$"
TEST_DIR="${TEST_DIR:-$ROOT/.gi-tui-test}"
DB="$TEST_DIR/gi.db"
WORKSPACE="$TEST_DIR/workspace"
OVERLAY_UPPER="$TEST_DIR/overlay-upper"
OVERLAY_WORK="$TEST_DIR/overlay-work"
MOUNTED=0

cleanup() {
  tmux kill-session -t "$SESSION" >/dev/null 2>&1 || true
  if [[ "$MOUNTED" == "1" ]] && mountpoint -q "$WORKSPACE"; then
    sudo umount "$WORKSPACE" >/dev/null 2>&1 || true
  fi
  rm -rf "$TEST_DIR"
}
trap cleanup EXIT

rm -rf "$ARTIFACT_DIR" "$TEST_DIR"
mkdir -p "$ARTIFACT_DIR" "$WORKSPACE" "$OVERLAY_UPPER" "$OVERLAY_WORK"
sudo mount -t overlay overlay -o lowerdir=/workspace,upperdir="$OVERLAY_UPPER",workdir="$OVERLAY_WORK" "$WORKSPACE"
MOUNTED=1
mkdir -p "$WORKSPACE/.pi"

cat > "$WORKSPACE/.pi/settings.json" <<'JSON'
{"defaultProvider":"test","defaultModel":"test-model","defaultThinkingLevel":"low","enabledModels":["test-model"]}
JSON
cat > "$WORKSPACE/AGENTS.md" <<'MD'
You are Gi Test.
MD

cd "$ROOT"
mkdir -p bin
if [[ ! -x bin/gi ]]; then
  go build -o bin/gi ./cmd/gi
fi

# Start detached with a fixed terminal size for deterministic mouse coordinates.
tmux new-session -d -x 100 -y 20 -s "$SESSION" "cd '$ROOT' && ./bin/gi -tui -db '$DB' -workspace '$WORKSPACE'"
for _ in 1 2 3 4 5; do
  sleep 1
  tmux capture-pane -pe -t "$SESSION":0 > "$ARTIFACT_DIR/01-start.txt"
  if grep -q "m0/t0" "$ARTIFACT_DIR/01-start.txt"; then
    break
  fi
done

if ! grep -q "m0/t0" "$ARTIFACT_DIR/01-start.txt"; then
  echo "TUI did not render bottom-band session counters" >&2
  exit 1
fi
if ! grep -q "(no messages yet)" "$ARTIFACT_DIR/01-start.txt"; then
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

# Blur the input and verify random typing is ignored until focus is restored by navigation.
tmux send-keys -t "$SESSION":0 Escape
sleep 1
BEFORE_COUNT=$(sqlite3 "$DB" 'select count(*) from messages;')
tmux send-keys -t "$SESSION":0 "ignored while blurred" Enter
sleep 1
AFTER_BLUR_COUNT=$(sqlite3 "$DB" 'select count(*) from messages;')
if [[ "$BEFORE_COUNT" != "$AFTER_BLUR_COUNT" ]]; then
  echo "TUI accepted input even though the input should have been blurred" >&2
  exit 1
fi

# Exercise transcript scrolling keys before resizing.
tmux send-keys -t "$SESSION":0 PageUp
sleep 1
tmux capture-pane -pe -t "$SESSION":0 > "$ARTIFACT_DIR/03-after-pageup.txt"
tmux send-keys -t "$SESSION":0 End
sleep 1
tmux capture-pane -pe -t "$SESSION":0 > "$ARTIFACT_DIR/04-after-end.txt"
if ! grep -q 'Gi received: hello from tmux' "$ARTIFACT_DIR/04-after-end.txt"; then
  echo "TUI did not restore transcript bottom after End" >&2
  exit 1
fi

# Resize and verify the session stays alive after interaction.
tmux resize-window -t "$SESSION":0 -x 120 -y 24
sleep 1
tmux has-session -t "$SESSION"
tmux capture-pane -pe -t "$SESSION":0 > "$ARTIFACT_DIR/05-after-resize.txt"
if ! grep -Eq 'm[0-9]+/t[0-9]+' "$ARTIFACT_DIR/05-after-resize.txt"; then
  echo "TUI lost bottom-band session counters after resize" >&2
  exit 1
fi
if ! grep -q 'Gi received: hello from tmux' "$ARTIFACT_DIR/05-after-resize.txt"; then
  echo "TUI did not render transcript content after resize" >&2
  exit 1
fi

# Persist final server-side state for inspection.
sqlite3 "$DB" 'select id, title, state_json from sessions;' > "$ARTIFACT_DIR/06-session-state.txt" 2>/dev/null || true

echo "TUI smoke test passed. Artifacts: $ARTIFACT_DIR"
