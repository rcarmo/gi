#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
FEATURE_DIR="${FEATURE_DIR:-$ROOT/features/tui}"
ARTIFACT_DIR="${ARTIFACT_DIR:-$ROOT/test-results/tui-gherkin}"
TEST_DIR="${TEST_DIR:-$ROOT/.gi-tui-gherkin}"
SESSION="gi-tui-gherkin-$$"
DB="$TEST_DIR/gi.db"
WORKSPACE="$TEST_DIR/workspace"
STEP=0

cleanup() {
  tmux kill-session -t "$SESSION" >/dev/null 2>&1 || true
  rm -rf "$TEST_DIR"
}
trap cleanup EXIT

capture() {
  STEP=$((STEP+1))
  tmux capture-pane -pe -t "$SESSION":0 > "$ARTIFACT_DIR/$(printf '%02d' "$STEP")-screen.txt"
}

screen_should_contain() {
  local text="$1"
  for _ in 1 2 3 4 5; do
    sleep 0.5
    capture
    if grep -Fq "$text" "$ARTIFACT_DIR/$(printf '%02d' "$STEP")-screen.txt"; then
      return 0
    fi
  done
  echo "screen did not contain: $text" >&2
  tail -80 "$ARTIFACT_DIR/$(printf '%02d' "$STEP")-screen.txt" >&2 || true
  exit 1
}

db_should_contain() {
  local role="$1" text="$2"
  for _ in 1 2 3 4 5 6 7 8; do
    sleep 0.5
    sqlite3 -separator '|' "$DB" 'select role, content from messages order by created_at asc, id asc;' > "$ARTIFACT_DIR/messages.txt" 2>/dev/null || true
    if grep -Fq "$role|$text" "$ARTIFACT_DIR/messages.txt"; then
      return 0
    fi
  done
  echo "database did not contain $role message: $text" >&2
  cat "$ARTIFACT_DIR/messages.txt" >&2 || true
  exit 1
}

message_count_should_be() {
  local expected="$1"
  local got
  got=$(sqlite3 "$DB" 'select count(*) from messages;' 2>/dev/null || echo 0)
  if [[ "$got" != "$expected" ]]; then
    echo "message count was $got, expected $expected" >&2
    sqlite3 -separator '|' "$DB" 'select role, content from messages order by created_at asc, id asc;' >&2 2>/dev/null || true
    exit 1
  fi
}

message_occurrence_should_be() {
  local expected="$1" role="$2" text="$3"
  local got
  for _ in 1 2 3 4 5 6 7 8; do
    sleep 0.5
    got=$(sqlite3 "$DB" "select count(*) from messages where role = '$role' and content = '$text';" 2>/dev/null || echo 0)
    if [[ "$got" == "$expected" ]]; then
      return 0
    fi
  done
  echo "$role message occurrence count for '$text' was $got, expected $expected" >&2
  sqlite3 -separator '|' "$DB" 'select role, content from messages order by created_at asc, id asc;' >&2 2>/dev/null || true
  exit 1
}

start_tui() {
  tmux kill-session -t "$SESSION" >/dev/null 2>&1 || true
  rm -rf "$TEST_DIR"
  mkdir -p "$ARTIFACT_DIR" "$WORKSPACE/.pi" "$WORKSPACE/.piclaw"
  cat > "$WORKSPACE/.pi/settings.json" <<'JSON'
{"defaultProvider":"test","defaultModel":"test-model","defaultThinkingLevel":"low","enabledModels":["test-model"]}
JSON
  cat > "$WORKSPACE/.piclaw/config.json" <<'JSON'
{"assistant":{"assistantName":"Gi"},"user":{"userName":"Tester"}}
JSON
  cat > "$WORKSPACE/AGENTS.md" <<'MD'
You are Gi TUI Test.
MD
  cd "$ROOT"
  mkdir -p bin
  go build -o bin/gi ./cmd/gi
  tmux new-session -d -x 120 -y 28 -s "$SESSION" "cd '$ROOT' && ./bin/gi -tui -db '$DB' -workspace '$WORKSPACE'"
}

type_and_enter() {
  tmux send-keys -t "$SESSION":0 "$1" Enter
  sleep 0.5
}

press_key() {
  local key="$1"
  case "$key" in
    "Ctrl-D") tmux send-keys -t "$SESSION":0 C-d ;;
    "Ctrl-P") tmux send-keys -t "$SESSION":0 C-p ;;
    "Ctrl-N") tmux send-keys -t "$SESSION":0 C-n ;;
    *) tmux send-keys -t "$SESSION":0 "$key" ;;
  esac
  sleep 0.5
}

resize_terminal() {
  local size="$1"
  local w=${size%x*}
  local h=${size#*x}
  tmux resize-window -t "$SESSION":0 -x "$w" -y "$h"
  sleep 0.5
}

tmux_should_be_alive() {
  tmux has-session -t "$SESSION" >/dev/null
}

tmux_should_exit() {
  for _ in 1 2 3 4 5; do
    sleep 0.5
    if ! tmux has-session -t "$SESSION" >/dev/null 2>&1; then
      return 0
    fi
  done
  echo "tmux session did not exit" >&2
  exit 1
}

run_step() {
  local line="$1"
  line="${line#    }"
  line="${line#  }"
  case "$line" in
    "Given a fresh gi TUI workspace") ;;
    "When I start the gi TUI in tmux") start_tui ;;
    "Then the screen should contain "*|"And the screen should contain "*)
      local text=${line#*should contain \"}; text=${text%\"}; screen_should_contain "$text" ;;
    "When I type "*" and press Enter"|"And I type "*" and press Enter")
      local text=${line#*I type \"}; text=${text%\" and press Enter}; type_and_enter "$text" ;;
    "When I press "*|"And I press "*)
      local key=${line#*I press }; press_key "$key" ;;
    "When I resize the terminal to "*)
      local size=${line#When I resize the terminal to }; resize_terminal "$size" ;;
    "Then the tmux session should be alive"|"And the tmux session should be alive") tmux_should_be_alive ;;
    "Then the tmux session should exit"|"And the tmux session should exit") tmux_should_exit ;;
    "Then the database message count should be "*)
      local expected=${line#Then the database message count should be }; message_count_should_be "$expected" ;;
    "Then the database should contain "*" user messages "*)
      local rest=${line#Then the database should contain }; local expected=${rest%% user messages*}; local text=${line#*user messages \"}; text=${text%\"}; message_occurrence_should_be "$expected" user "$text" ;;
    "And the database should contain "*" user messages "*)
      local rest=${line#And the database should contain }; local expected=${rest%% user messages*}; local text=${line#*user messages \"}; text=${text%\"}; message_occurrence_should_be "$expected" user "$text" ;;
    "Then the database should contain a user message "*)
      local text=${line#Then the database should contain a user message \"}; text=${text%\"}; db_should_contain user "$text" ;;
    "And the database should contain an assistant message "*)
      local text=${line#And the database should contain an assistant message \"}; text=${text%\"}; db_should_contain assistant "$text" ;;
    Feature:*|Scenario:*|""|"The "*) ;;
    *) echo "unsupported Gherkin step: $line" >&2; exit 1 ;;
  esac
}

rm -rf "$ARTIFACT_DIR"
mkdir -p "$ARTIFACT_DIR"
shopt -s nullglob
features=("$FEATURE_DIR"/*.feature)
if [[ ${#features[@]} -eq 0 ]]; then
  echo "no feature files in $FEATURE_DIR" >&2
  exit 1
fi
for feature in "${features[@]}"; do
  echo "Running $feature"
  while IFS= read -r line; do
    run_step "$line"
  done < "$feature"
done

echo "TUI Gherkin tests passed. Artifacts: $ARTIFACT_DIR"
