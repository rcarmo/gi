package store

import (
	"context"
	"database/sql"
)

// SessionActivity reads phase, claim and the latest compaction occurrence in one
// SQLite statement. Lifecycle events merely invalidate this authoritative view.
func (s *Store) SessionActivity(ctx context.Context, sessionID string) (map[string]any, error) {
	result := map[string]any{"chat_jid": "gi:" + sessionID, "status": "idle", "turn_id": "", "compaction": nil}
	var id, status, phase string
	var claimed bool
	var payload, eventType, at sql.NullString
	var seq sql.NullInt64
	err := s.db.QueryRowContext(ctx, `select t.id,t.status,t.phase,e.event_type,e.seq,e.payload_json,e.created_at,exists(select 1 from session_active_turns where session_id=t.session_id and turn_id=t.id)
 from turns t
 left join turn_events e on e.id=(select id from turn_events where turn_id=t.id and session_id=t.session_id and event_type in ('compaction.started','compaction.completed','compaction.cancelled','compaction.failed','compaction.suppressed') order by seq desc limit 1)
 where t.session_id=? and t.id=coalesce((select turn_id from session_active_turns where session_id=t.session_id),
 (select id from turns where session_id=t.session_id and status not in ('queued','steering') order by created_at desc,id desc limit 1))`, sessionID).Scan(&id, &status, &phase, &eventType, &seq, &payload, &at, &claimed)
	if err == sql.ErrNoRows {
		return result, nil
	}
	if err != nil {
		return nil, err
	}
	result["turn_id"] = id
	result["status"] = status
	if !claimed || status != "running" && status != "cancelling" {
		result["status"] = "idle"
	}
	result["phase"] = phase
	if eventType.Valid {
		details, err := unmarshalJSONMap(payload.String)
		if err != nil {
			return nil, err
		}
		details["event_type"] = eventType.String
		details["seq"] = seq.Int64
		details["timestamp"] = at.String
		details["turn_id"] = id
		details["active"] = claimed && (status == "running" || status == "cancelling") && eventType.String == "compaction.started" && phase == "compacting"
		result["compaction"] = details
	}
	return result, nil
}
