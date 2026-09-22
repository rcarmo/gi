package store

import "context"

// AdmitManualCompaction never queues behind other work or redirects to steering.
func (s *Store) AdmitManualCompaction(ctx context.Context, sessionID, turnID, expected, model string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var busy bool
	if err = tx.QueryRowContext(ctx, `select exists(select 1 from session_active_turns where session_id=?) or exists(select 1 from turns where session_id=? and status in ('queued','steering','running','cancelling')) or exists(select 1 from steering_queue where session_id=? and status in ('queued','claimed'))`, sessionID, sessionID, sessionID).Scan(&busy); err != nil {
		return err
	}
	if busy {
		return ErrQueueConflict
	}
	snapshot, err := readContextSnapshot(ctx, tx, sessionID)
	if err != nil {
		return err
	}
	if expected == "" || ContextToken(snapshot) != expected {
		return ErrContextChanged
	}
	metadata, err := marshalJSON(map[string]any{"operation": "manual_compaction", "context_token": expected, "model": model, "intent": "compact"})
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `insert into turns(id,session_id,status,phase,prompt,metadata_json,created_at,updated_at,queue_position,claimed_by,claimed_at) values(?,?,'running','setup','',?,`+defaultNow+`,`+defaultNow+`,(select coalesce(max(queue_position),0)+1 from turns where session_id=?),'runner',`+defaultNow+`)`, turnID, sessionID, metadata, sessionID)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `insert into session_active_turns(session_id,turn_id,worker_id,claim_token,claimed_at,updated_at) values(?,?,'runner',?,`+defaultNow+`,`+defaultNow+`)`, sessionID, turnID, turnID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `insert into turn_events(turn_id,session_id,seq,event_type,payload_json,created_at) values(?,?,1,'turn.submitted','{"phase":"queue","checkpoint":true,"intent":"compact","operation":"manual_compaction"}',`+defaultNow+`)`, turnID, sessionID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `update sessions set state_json=json_set(state_json,'$.status','running','$.active_turn_id',?),updated_at=`+defaultNow+` where id=?`, turnID, sessionID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) CompactionBusy(ctx context.Context, sessionID string) (bool, error) {
	var busy bool
	err := s.db.QueryRowContext(ctx, `select exists(select 1 from session_active_turns where session_id=?) or exists(select 1 from turns where session_id=? and status in ('queued','steering','running','cancelling')) or exists(select 1 from steering_queue where session_id=? and status in ('queued','claimed'))`, sessionID, sessionID, sessionID).Scan(&busy)
	return busy, err
}
