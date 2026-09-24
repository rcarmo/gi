package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// latestToolActivity projects one persisted occurrence, never a log of outputs.
// Joining the end by call ID and sequence prevents equal tool names from mixing.
func latestToolActivity(ctx context.Context, tx *sql.Tx, turnID string, claimed bool) (map[string]any, error) {
	var startRaw, started, status string
	var ended, endType sql.NullString
	var seq int64
	err := tx.QueryRowContext(ctx, `select a.seq,a.payload_json,a.created_at,b.created_at,b.event_type,t.status
 from turns t join turn_events a on a.id=(select id from turn_events where turn_id=t.id and event_type='tool.started' order by seq desc limit 1)
 left join turn_events b on b.id=(select id from turn_events where turn_id=t.id and seq>a.seq and event_type in ('tool.finished','tool.failed')
 and coalesce(json_extract(payload_json,'$.tool_call_id'),'')=coalesce(json_extract(a.payload_json,'$.tool_call_id'),'') order by seq asc limit 1)
 where t.id=?`, turnID).Scan(&seq, &startRaw, &started, &ended, &endType, &status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err = json.Unmarshal([]byte(startRaw), &payload); err != nil {
		return nil, err
	}
	name, _ := payload["tool"].(string)
	callID, _ := payload["tool_call_id"].(string)
	if callID == "" {
		callID = fmt.Sprintf("%s:%d", turnID, seq)
	}
	preview, _ := payload["preview"].(string)
	if preview == "" {
		preview = ToolActivityPreview(payload)
	} else {
		preview = ToolActivityPreview(map[string]any{"command": preview})
	}
	state := "running"
	finished := ""
	if ended.Valid {
		finished = ended.String
		if endType.String == "tool.failed" {
			state = "failed"
		} else {
			state = "completed"
		}
	} else if !claimed || status != "running" && status != "cancelling" {
		// A lost claim proves inactivity, not the moment the tool stopped.
		// Do not invent terminal timing from unrelated turn updates.
		state = "interrupted"
	}
	result := map[string]any{"turn_id": turnID, "tool_call_id": callID, "start_seq": seq, "name": name, "preview": preview, "state": state, "started_at": started, "finished_at": finished, "duration_ms": nil}
	if finished != "" {
		a, e1 := time.Parse(time.RFC3339Nano, started)
		b, e2 := time.Parse(time.RFC3339Nano, finished)
		if e1 == nil && e2 == nil {
			ms := b.Sub(a).Milliseconds()
			if ms < 0 {
				ms = 0
			}
			result["duration_ms"] = ms
		}
	}
	return result, nil
}

// ToolActivityPreview retains a small command/path/query, not arbitrary arguments.
func ToolActivityPreview(args map[string]any) string {
	preview := ""
	if args != nil {
		for _, key := range []string{"command", "path", "query"} {
			switch value := args[key].(type) {
			case string:
				preview = value
			case []any:
				var parts []string
				for _, p := range value {
					if str, ok := p.(string); ok {
						parts = append(parts, str)
					}
				}
				preview = strings.Join(parts, " ")
			}
			if preview != "" {
				break
			}
		}
	}
	// Bounded plain text, never executable markup. Do not expose full tool output.
	preview = strings.Join(strings.Fields(preview), " ")
	if utf8.RuneCountInString(preview) > 200 {
		preview = string([]rune(preview)[:200]) + "…"
	}
	return preview
}
