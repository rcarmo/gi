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
CURRENT_FEATURE=""
CURRENT_STEP=""
STARTED_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)

capture_pane() {
  if ! tmux has-session -t "$SESSION" >/dev/null 2>&1; then
    return 1
  fi
  tmux capture-pane -p -t "$SESSION":0 2>/dev/null | tr -d '\r'
}

write_failure_summary() {
  local rc="$1"
  mkdir -p "$ARTIFACT_DIR"
  {
    echo "# TUI Gherkin failure summary"
    echo
    echo "- status: failed"
    echo "- exit_code: $rc"
    echo "- feature: ${CURRENT_FEATURE:-unknown}"
    echo "- step: ${CURRENT_STEP:-unknown}"
    echo "- started_at: $STARTED_AT"
    echo "- failed_at: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
    echo
    echo "## Last screen"
    echo '```'
    if ! capture_pane; then
      echo "tmux session not running"
    fi
    echo '```'
    echo
    echo "## Messages"
    echo '```'
    sqlite3 -separator '|' "$DB" 'select role, content from messages order by created_at asc, id asc;' 2>/dev/null || true
    echo '```'
  } > "$ARTIFACT_DIR/failure-summary.md"
}

cleanup() {
  local rc=$?
  if [[ "$rc" != "0" ]]; then
    write_failure_summary "$rc" || true
  fi
  tmux kill-session -t "$SESSION" >/dev/null 2>&1 || true
  rm -rf "$TEST_DIR"
  exit "$rc"
}
trap cleanup EXIT

capture() {
  STEP=$((STEP+1))
  capture_pane > "$ARTIFACT_DIR/$(printf '%02d' "$STEP")-screen.txt"
}

slugify() {
  basename "$1" .feature | tr -c '[:alnum:]_-' '-'
}

dump_feature_state() {
  local feature="$1"
  local slug
  slug=$(slugify "$feature")
  if [[ -f "$DB" ]]; then
    sqlite3 -header -column "$DB" 'select id, title, parent_session_id, state_json from sessions order by created_at asc;' > "$ARTIFACT_DIR/${slug}-sessions.txt" 2>/dev/null || true
    sqlite3 -header -column "$DB" 'select role, content, created_at from messages order by created_at asc, id asc;' > "$ARTIFACT_DIR/${slug}-messages.txt" 2>/dev/null || true
    sqlite3 "$DB" .dump > "$ARTIFACT_DIR/${slug}.sqlite.sql" 2>/dev/null || true
  fi
}

write_report() {
  local status="$1"
  {
    echo "# TUI Gherkin report"
    echo
    echo "- status: $status"
    echo "- started_at: $STARTED_AT"
    echo "- finished_at: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
    echo "- feature_dir: $FEATURE_DIR"
    echo "- artifact_dir: $ARTIFACT_DIR"
    echo
    echo "## Features"
    for feature in "${features[@]}"; do
      local slug
      slug=$(slugify "$feature")
      echo "- $(basename "$feature")"
      echo "  - sessions: ${slug}-sessions.txt"
      echo "  - messages: ${slug}-messages.txt"
      echo "  - sqlite_dump: ${slug}.sqlite.sql"
    done
    echo
    echo "## Pane captures"
    find "$ARTIFACT_DIR" -maxdepth 1 -name '*-screen.txt' -printf '- %f\n' | sort
  } > "$ARTIFACT_DIR/report.md"
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

prepare_tui_workspace() {
  tmux kill-session -t "$SESSION" >/dev/null 2>&1 || true
  rm -rf "$TEST_DIR"
  mkdir -p "$ARTIFACT_DIR" "$WORKSPACE/.pi" "$WORKSPACE/.piclaw" "$TEST_DIR/.pi" "$TEST_DIR/.piclaw"
  cat > "$WORKSPACE/.pi/settings.json" <<'JSON'
{"defaultProvider":"test","defaultModel":"test-model","defaultThinkingLevel":"low","enabledModels":["test-model"]}
JSON
  cp "$WORKSPACE/.pi/settings.json" "$TEST_DIR/.pi/settings.json"
  cat > "$WORKSPACE/.piclaw/config.json" <<'JSON'
{"assistant":{"assistantName":"Gi"},"user":{"userName":"Tester"}}
JSON
  cp "$WORKSPACE/.piclaw/config.json" "$TEST_DIR/.piclaw/config.json"
  cat > "$WORKSPACE/AGENTS.md" <<'MD'
You are Gi TUI Test.
MD
  cp "$WORKSPACE/AGENTS.md" "$TEST_DIR/AGENTS.md"
  cd "$ROOT"
}

wait_for_tui_ready() {
  for _ in 1 2 3 4 5 6 7 8 9 10; do
    sleep 0.4
    if capture_pane | grep -Fq "m0/t0"; then
      return 0
    fi
  done
  echo "TUI did not become ready" >&2
  capture_pane >&2 || true
  exit 1
}

start_tui() {
  prepare_tui_workspace
  tmux new-session -d -x 120 -y 28 -s "$SESSION" "cd '$ROOT' && ./bin/gi -tui -db '$DB' -workspace '$WORKSPACE'"
  wait_for_tui_ready
}

start_tui_default() {
  prepare_tui_workspace
  tmux new-session -d -x 120 -y 28 -s "$SESSION" "cd '$TEST_DIR' && '$ROOT/bin/gi'"
  wait_for_tui_ready
}

type_and_enter() {
  tmux send-keys -t "$SESSION":0 -l "$1"
  tmux send-keys -t "$SESSION":0 C-m
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
  CURRENT_STEP="$line"
  case "$line" in
    "Given a fresh gi TUI workspace") ;;
    "When I start the gi TUI in tmux") start_tui ;;
    "When I start gi without arguments in tmux") start_tui_default ;;
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
    "Then the database should contain an assistant message "*|"And the database should contain an assistant message "*)
      local text=${line#*assistant message \"}; text=${text%\"}; db_should_contain assistant "$text" ;;
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
  CURRENT_FEATURE="$feature"
  echo "Running $feature"
  while IFS= read -r line; do
    run_step "$line"
  done < "$feature"
  dump_feature_state "$feature"
done

write_report passed

echo "TUI Gherkin tests passed. Artifacts: $ARTIFACT_DIR"
