package store

import (
	"context"
	"encoding/json"
)

// ListTerminalToolResults projects recorded execution boundaries for display only.
// These rows never enter persisted messages or the provider conversation. A
// missing terminal boundary is deliberately not reconstructed from turn state.
func (s *Store) ListTerminalToolResults(ctx context.Context, sessionID string) ([]Message, error) {
	rows, err := s.db.QueryContext(ctx, `select e.id,e.turn_id,e.event_type,e.payload_json,e.created_at
 from turn_events e where e.session_id=? and e.event_type in ('tool.finished','tool.failed','tool.cancelled','tool.aborted')
 and json_type(e.payload_json,'$.duration_ms') in ('integer','real')
 and exists(select 1 from turn_events a where a.turn_id=e.turn_id and a.seq<e.seq and a.event_type='tool.started'
 and json_extract(a.payload_json,'$.occurrence_id')=json_extract(e.payload_json,'$.occurrence_id')
 and coalesce(json_extract(a.payload_json,'$.tool_call_id'),'')=coalesce(json_extract(e.payload_json,'$.tool_call_id'),''))
 and not exists(select 1 from turn_events b where b.turn_id=e.turn_id and b.seq<e.seq
 and b.event_type in ('tool.finished','tool.failed','tool.cancelled','tool.aborted')
 and json_extract(b.payload_json,'$.occurrence_id')=json_extract(e.payload_json,'$.occurrence_id')
 and coalesce(json_extract(b.payload_json,'$.tool_call_id'),'')=coalesce(json_extract(e.payload_json,'$.tool_call_id'),''))
 and not exists(select 1 from messages m where m.session_id=e.session_id
 and coalesce(json_extract(m.payload_json,'$.turn_id'),'')=e.turn_id
 and json_extract(m.payload_json,'$.occurrence_id')=json_extract(e.payload_json,'$.occurrence_id'))
 order by e.created_at,e.id`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Message
	for rows.Next() {
		var id int64
		var turnID, kind, raw, created string
		if err := rows.Scan(&id, &turnID, &kind, &raw, &created); err != nil {
			return nil, err
		}
		var p map[string]any
		if err := json.Unmarshal([]byte(raw), &p); err != nil {
			return nil, err
		}
		occurrence, ok := p["occurrence_id"].(string)
		if !ok || occurrence == "" {
			continue
		}
		p["tool_name"] = p["tool"]
		p["turn_id"] = turnID
		p["status"] = kind[len("tool."):]
		if kind == "tool.finished" {
			p["status"] = "ok"
		}
		p["is_error"] = kind != "tool.finished"
		result = append(result, Message{ID: "terminal:" + turnID + ":" + occurrence, SessionID: sessionID, Role: "tool_result", Content: "Tool execution " + kind[len("tool."):], CreatedAt: created, Payload: p})
	}
	return result, rows.Err()
}
