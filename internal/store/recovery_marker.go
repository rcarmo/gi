package store

import "context"

// RecoveryMarker describes a native stale-claim requeue, never a provider retry,
// timeout or held tool checkpoint. Callers attach it only to final replies.
func (s *Store) RecoveryMarker(ctx context.Context, sessionID, turnID string) (map[string]any, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `select count(*) from turn_events where session_id=? and turn_id=? and event_type='turn.recovered'
 and json_extract(payload_json,'$.recovery_disposition') in ('requeue_interrupted_turn','requeue_after_compaction_checkpoint')`, sessionID, turnID).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}
	return map[string]any{"type": "recovery_marker", "recovered": true, "recovery_kind": "native_interrupted_turn", "attempts_used": count + 1}, nil
}
