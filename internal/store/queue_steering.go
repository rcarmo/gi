package store

import (
	"context"
	"database/sql"
	"encoding/json"
)

// SteerQueuedTurn moves an unclaimed queue row into a run-bound steering entry.
// The durable row is kept for audit/recovery; the transaction is the admission
// boundary, so a failed or repeated request cannot duplicate delivery.
func (s *Store) SteerQueuedTurn(ctx context.Context, sessionID, queuedID, activeID string) error {
	if activeID == "" || queuedID == activeID {
		return ErrQueueConflict
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `update turns set status = 'steering', phase = 'steering', updated_at = `+defaultNow+`
 where id = ? and session_id = ? and status = 'queued'
 and not exists (select 1 from session_active_turns where turn_id = turns.id)
 and exists (select 1 from session_active_turns a join turns t on t.id = a.turn_id
 where a.session_id = ? and a.turn_id = ? and t.session_id = ? and t.status = 'running')
 and (select count(*) from steering_queue where session_id = ? and status = 'queued') < 10`,
		queuedID, sessionID, sessionID, activeID, sessionID, sessionID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrQueueConflict
	}
	var prompt, metadata string
	if err = tx.QueryRowContext(ctx, `select prompt, metadata_json from turns where id = ?`, queuedID).Scan(&prompt, &metadata); err != nil {
		return err
	}
	payload, err := unmarshalJSONMap(metadata)
	if err != nil {
		return err
	}
	if payload == nil {
		payload = map[string]any{}
	}
	payload["intent"] = "steer"
	payload["kind"] = "steering"
	payload["active_turn_id"] = activeID
	payload["source_queue_id"] = queuedID
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	// Keep structured media refs in payload; do not replace them with an empty
	// string-only media list. The ordinary queue row remains the recovery source.
	if _, err = tx.ExecContext(ctx, `insert into steering_queue (session_id,turn_id,role,content,payload_json,media_json,queue_mode,status,source_queue_id,created_at,updated_at)
 values (?,?,'user',?,?,'[]','one-at-a-time','queued',?,`+defaultNow+`,`+defaultNow+`)`, sessionID, activeID, prompt, string(raw), queuedID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `update sessions set state_json = json_set(state_json,'$.queue_count', (select count(*) from turns where session_id = ? and status = 'queued')) where id = ?`, sessionID, sessionID); err != nil {
		return err
	}
	return tx.Commit()
}

// restoreBoundSteeringTx runs before releasing a claim. Unconsumed queue Steer
// must never be picked up by another run or a generic steering continuation.
func restoreBoundSteeringTx(ctx context.Context, tx *sql.Tx, sessionID, claimToken string) error {
	_, err := tx.ExecContext(ctx, `update turns set status = 'queued', phase = 'steer_returned', updated_at = `+defaultNow+`
 where session_id = ? and status = 'steering' and id in (
 select q.source_queue_id from steering_queue q join session_active_turns a on a.turn_id = q.turn_id and a.session_id = q.session_id
 where q.session_id = ? and q.status in ('queued','claimed') and q.source_queue_id is not null and (? = '' or a.claim_token = ?))`, sessionID, sessionID, claimToken, claimToken)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `update steering_queue set status = 'returned', updated_at = `+defaultNow+`
 where session_id = ? and status in ('queued','claimed') and source_queue_id is not null and turn_id in
 (select turn_id from session_active_turns where session_id = ? and (? = '' or claim_token = ?))`, sessionID, sessionID, claimToken, claimToken)
	return err
}

// PersistBoundSteering acknowledges a claimed entry together with its user
// message. If the process/run ends between dequeue and here, release recovers
// the original queued row instead of silently losing it.
func (s *Store) PersistBoundSteering(ctx context.Context, sessionID, turnID string, msg SteeringMessage, payload map[string]any) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `update steering_queue set status='dequeued',updated_at=`+defaultNow+` where id=? and session_id=? and turn_id=? and source_queue_id is not null and status='claimed'
 and exists(select 1 from session_active_turns a join turns t on t.id=a.turn_id where a.session_id=? and a.turn_id=? and t.status='running')`, msg.ID, sessionID, turnID, sessionID, turnID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrQueueConflict
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `insert into messages(id,session_id,role,content,payload_json,created_at) values(?,?,'user',?,?,`+defaultNow+`)`, NowID("msg"), sessionID, msg.Content, string(raw)); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `update turns set status='cancelled',phase='steered',finished_at=`+defaultNow+`,updated_at=`+defaultNow+` where status='steering' and id=(select source_queue_id from steering_queue where id=?)`, msg.ID); err != nil {
		return err
	}
	return tx.Commit()
}
