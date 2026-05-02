#!/usr/bin/env bash
set -euo pipefail

GI_BIN="$(pwd)/bin/gi"
DB="$(pwd)/.gi-run/gi.db"
PORT=8091
WORKSPACE=/workspace
RESULTS_FILE="/tmp/tool-matrix-results.md"

# Test matrix: provider/model
MODELS=(
  # openai-responses (Copilot)
  "github-copilot/gpt-5-mini"
  "github-copilot/gpt-5.4-mini"
  # openai-completions (Copilot)
  "github-copilot/gpt-4.1"
  "github-copilot/gpt-4o"
  "github-copilot/gemini-2.5-pro"
  "github-copilot/grok-code-fast-1"
  # anthropic-messages (Copilot) — known 404 on Copilot proxy, included for tracking
  "github-copilot/claude-sonnet-4"
  # openai-codex-responses
  "openai-codex/gpt-5.1-codex-mini"
  "openai-codex/gpt-5.4-mini"
  "openai-codex/gpt-5.3-codex-spark"
)

PROMPT='Use the shell tool to run this exact command: echo TOOLTEST_OK'
TIMEOUT_SECS=40

echo "# Tool-use matrix test — $(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"
echo "| Model | API | Iter1 | Iter2 | Result | Time |" >> "$RESULTS_FILE"
echo "|-------|-----|-------|-------|--------|------|" >> "$RESULTS_FILE"

pass=0
fail=0
skip=0

for MODEL in "${MODELS[@]}"; do
  echo "=== Testing: $MODEL ==="
  
  # Start gi
  "$GI_BIN" -bind 127.0.0.1 -port $PORT -model "$MODEL" -db "$DB" -workspace "$WORKSPACE" 2>/tmp/gi-matrix.log &
  GIPID=$!
  sleep 2
  
  if ! kill -0 $GIPID 2>/dev/null; then
    echo "  SKIP: gi failed to start"
    echo "| \`$MODEL\` | — | — | — | ⏭ SKIP (start failed) | — |" >> "$RESULTS_FILE"
    skip=$((skip+1))
    continue
  fi
  
  # Create session + submit prompt
  SESS=$(curl -sf -X POST "http://127.0.0.1:$PORT/api/sessions" \
    -H 'Content-Type: application/json' \
    -d '{"title":"matrix-test"}' | python3 -c "import json,sys; print(json.load(sys.stdin)['id'])" 2>/dev/null || echo "")
  
  if [ -z "$SESS" ]; then
    echo "  SKIP: session creation failed"
    echo "| \`$MODEL\` | — | — | — | ⏭ SKIP (no session) | — |" >> "$RESULTS_FILE"
    kill $GIPID 2>/dev/null; wait $GIPID 2>/dev/null || true
    skip=$((skip+1))
    continue
  fi
  
  START_T=$(date +%s)
  
  curl -sf -X POST "http://127.0.0.1:$PORT/api/sessions/$SESS/prompt" \
    -H 'Content-Type: application/json' \
    -d "{\"prompt\":\"$PROMPT\",\"model\":\"$MODEL\"}" >/dev/null 2>&1
  
  # Wait for completion
  COMPLETED=false
  for i in $(seq 1 $TIMEOUT_SECS); do
    sleep 1
    STATUS=$(curl -sf "http://127.0.0.1:$PORT/api/sessions/$SESS/turns" 2>/dev/null | \
      python3 -c "import json,sys; turns=json.load(sys.stdin).get('turns',[]); print(turns[-1]['status'] if turns else 'none')" 2>/dev/null || echo "error")
    if [ "$STATUS" = "completed" ] || [ "$STATUS" = "failed" ]; then
      COMPLETED=true
      break
    fi
  done
  
  END_T=$(date +%s)
  ELAPSED=$((END_T - START_T))
  
  # Extract results from log
  ITER1=$(grep -o 'iter=1/64.*stop="[^"]*"' /tmp/gi-matrix.log | tail -1 | grep -o 'stop="[^"]*"' || echo "—")
  ITER2=$(grep -o 'iter=2/64.*stop="[^"]*"\|iter=2/64.*error:.*' /tmp/gi-matrix.log | tail -1 | head -c 80 || echo "—")
  
  # Check if tool result contains our marker
  HAS_MARKER=$(curl -sf "http://127.0.0.1:$PORT/api/sessions/$SESS/messages" 2>/dev/null | \
    python3 -c "
import json,sys
msgs = json.load(sys.stdin).get('messages',[])
found = any('TOOLTEST_OK' in m.get('content','') for m in msgs if m.get('role') == 'tool_result')
print('yes' if found else 'no')
" 2>/dev/null || echo "unknown")
  
  # Determine API type from log
  API_TYPE=$(grep -o 'openai-responses\|openai-completions\|anthropic-messages\|openai-codex-responses' /tmp/gi-matrix.log | tail -1 || echo "?")
  if [ -z "$API_TYPE" ]; then
    PROVIDER=$(echo "$MODEL" | cut -d/ -f1)
    MNAME=$(echo "$MODEL" | cut -d/ -f2)
    API_TYPE="$PROVIDER"
  fi
  
  if [ "$COMPLETED" = "true" ] && [ "$STATUS" = "completed" ] && [ "$HAS_MARKER" = "yes" ]; then
    RESULT="✅ PASS"
    pass=$((pass+1))
    echo "  PASS (${ELAPSED}s)"
  elif [ "$COMPLETED" = "true" ] && [ "$STATUS" = "completed" ]; then
    RESULT="⚠️ COMPLETE (no marker)"
    pass=$((pass+1))
    echo "  COMPLETE but no marker (${ELAPSED}s)"
  elif [ "$COMPLETED" = "true" ] && [ "$STATUS" = "failed" ]; then
    RESULT="❌ FAIL"
    fail=$((fail+1))
    ERR=$(grep -o 'error:.*' /tmp/gi-matrix.log | tail -1 | head -c 60 || echo "unknown")
    echo "  FAIL: $ERR (${ELAPSED}s)"
  else
    RESULT="⏱ TIMEOUT"
    fail=$((fail+1))
    echo "  TIMEOUT after ${TIMEOUT_SECS}s"
  fi
  
  echo "| \`$MODEL\` | $API_TYPE | $ITER1 | ${ITER2:0:60} | $RESULT | ${ELAPSED}s |" >> "$RESULTS_FILE"
  
  # Cleanup
  kill $GIPID 2>/dev/null; wait $GIPID 2>/dev/null || true
  > /tmp/gi-matrix.log
done

echo "" >> "$RESULTS_FILE"
echo "**Summary:** $pass pass, $fail fail, $skip skip out of ${#MODELS[@]} models" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"
echo "=== Done: $pass pass, $fail fail, $skip skip ==="
cat "$RESULTS_FILE"
