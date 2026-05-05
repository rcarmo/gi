#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="${1:-$ROOT/artifacts/tui-audit}"
mkdir -p "$OUT_DIR"
export PATH=/home/linuxbrew/.linuxbrew/bin:$PATH
export TMPDIR=/workspace/tmp

cd "$ROOT"
go build -o bin/gi ./cmd/gi

make_workspace() {
  local dir="$1"
  mkdir -p "$dir/.pi" "$dir/.piclaw"
  cat > "$dir/.pi/settings.json" <<'JSON'
{"defaultProvider":"opencode-zen","defaultModel":"opencode-zen/minimax-m2.5-free","defaultThinkingLevel":"low","enabledModels":["opencode-zen/minimax-m2.5-free"],"tuiScrollbackLimit":250}
JSON
  cat > "$dir/.piclaw/config.json" <<'JSON'
{"assistant":{"assistantName":"Gi"},"user":{"userName":"Rui"}}
JSON
  cat > "$dir/AGENTS.md" <<'MD'
You are Gi screenshot audit test.
MD
}

seed_markdown_db() {
  local db="$1"
  cat > "$OUT_DIR/seed_markdown.go" <<'EOF'
package main
import (
  "context"
  "os"
  "github.com/rcarmo/gi/internal/store"
)
func main(){
  db:=os.Args[1]
  s,err:=store.Open(db)
  if err!=nil { panic(err) }
  defer s.Close()
  ctx:=context.Background()
  sess,err:=s.CreateSession(ctx, "session_md", "@agent", map[string]any{"status":"idle","model":"opencode-zen/minimax-m2.5-free","provider":"opencode-zen","thinking_level":"low"})
  if err!=nil { panic(err) }
  msg := "# Markdown demo\n\n- streaming model output\n- runtime controls\n\n| Name | Role | Value |\n| --- | --- | --- |\n| Alice | admin | 42 |\n| Bob | user | 7 |\n\n> Responsive table fallback on narrow layouts."
  if err:=s.AddMessage(ctx, store.NowID("msg"), sess.ID, "assistant", msg, map[string]any{"kind":"chat"}); err!=nil { panic(err) }
}
EOF
  go run "$OUT_DIR/seed_markdown.go" "$db"
  rm -f "$OUT_DIR/seed_markdown.go"
}

capture_session() {
  local name="$1"; shift
  local cols="$1"; shift
  local rows="$1"; shift
  local cmd="$1"; shift
  local session="audit-${name}-$$"
  tmux new-session -d -x "$cols" -y "$rows" -s "$session" "$cmd"
  sleep 3
  tmux capture-pane -pe -t "$session":0 > "$OUT_DIR/${name}.ansi.txt"
  tmux capture-pane -p -t "$session":0 > "$OUT_DIR/${name}.txt"
  tmux kill-session -t "$session" || true
}

GI_WS_BASIC="$OUT_DIR/workspace-basic"
GI_DB_BASIC="$OUT_DIR/gi-basic.db"
make_workspace "$GI_WS_BASIC"

GI_WS_MD="$OUT_DIR/workspace-md"
GI_DB_MD="$OUT_DIR/gi-md.db"
make_workspace "$GI_WS_MD"
seed_markdown_db "$GI_DB_MD"

capture_session "gi-wide" 100 22 "cd '$ROOT' && ./bin/gi -tui -db '$GI_DB_BASIC' -workspace '$GI_WS_BASIC'"
capture_session "gi-narrow" 60 18 "cd '$ROOT' && ./bin/gi -tui -db '$GI_DB_BASIC' -workspace '$GI_WS_BASIC'"
capture_session "gi-markdown" 100 22 "cd '$ROOT' && ./bin/gi -tui -db '$GI_DB_MD' -workspace '$GI_WS_MD'"
capture_session "pi-wide" 100 22 "cd '$GI_WS_BASIC' && pi"

echo "$OUT_DIR"
