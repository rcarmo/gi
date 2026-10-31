package store

import (
	"context"
	"fmt"
)

// AdmitTurnEffect is the host-entry boundary: a committed intent and phase
// transition must both exist before calling a provider or executor. The active
// claim is checked in the same transaction so a committed cancellation wins.
func (s *Store) AdmitTurnEffect(ctx context.Context, sessionID, turnID, phase, eventType string, payload map[string]any) error {
	switch eventType {
	case "inference.started", "inference.request_prepared", "tool.started":
	default:
		return fmt.Errorf("invalid effect intent %q", eventType)
	}
	raw, err := marshalJSON(payload)
	if err != nil {
		return err
	}
	if len(raw) > 64<<20 {
		return fmt.Errorf("effect intent exceeds 64 MiB")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `update turns set phase=?,updated_at=`+defaultNow+` where id=? and session_id=? and status='running' and exists(select 1 from session_active_turns where session_id=? and turn_id=?)`, phase, turnID, sessionID, sessionID, turnID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return context.Canceled
	}
	if _, err = tx.ExecContext(ctx, `insert into turn_events(turn_id,session_id,seq,event_type,payload_json,created_at) values(?,?,coalesce((select max(seq)+1 from turn_events where turn_id=?),1),?,?,`+defaultNow+`)`, turnID, sessionID, turnID, eventType, raw); err != nil {
		return err
	}
	return tx.Commit()
}
