package store

import (
	"context"
	"database/sql"
	"fmt"
)

// EnqueueActiveSteering admits generic live-turn steering and normalizes its
// session state in the same transaction as the running-claim check. The generic
// EnqueueSteering remains available for persisted continuation/recovery work.
func (s *Store) EnqueueActiveSteering(ctx context.Context, sessionID, turnID, role, content string, payload map[string]any, media []string, queueMode string) (int64, error) {
	payloadJSON, err := marshalJSON(payload)
	if err != nil {
		return 0, err
	}
	mediaJSON, err := marshalJSONArray(media)
	if err != nil {
		return 0, err
	}
	if queueMode == "" {
		queueMode = "one-at-a-time"
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `insert into steering_queue (session_id,turn_id,role,content,payload_json,media_json,queue_mode,status,created_at,updated_at)
 select ?,?,?,?,?,?,?,'queued',`+defaultNow+`,`+defaultNow+`
 where exists(select 1 from session_active_turns a join turns t on t.id=a.turn_id and t.session_id=a.session_id
 where a.session_id=? and a.turn_id=? and t.status='running' and coalesce(json_extract(t.metadata_json,'$.operation'),'')!='manual_compaction')
 and (select count(*) from steering_queue where session_id=? and status='queued')<10`, sessionID, turnID, role, content, payloadJSON, mediaJSON, queueMode, sessionID, turnID, sessionID)
	if err != nil {
		return 0, fmt.Errorf("enqueue active steering: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n == 0 {
		var active string
		err := tx.QueryRowContext(ctx, `select t.id from session_active_turns a join turns t on t.id=a.turn_id and t.session_id=a.session_id where a.session_id=? and a.turn_id=? and t.status='running' and coalesce(json_extract(t.metadata_json,'$.operation'),'')!='manual_compaction'`, sessionID, turnID).Scan(&active)
		if err == sql.ErrNoRows {
			return 0, ErrQueueConflict
		}
		if err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("enqueue steering: steering queue is full")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	// Completion may race admission, but cannot interleave these writes. Never
	// publish a running state for a terminal predecessor after admission failed.
	_, err = tx.ExecContext(ctx, `update sessions set state_json=json_set(state_json,'$.status','running','$.active_turn_id',?,
 '$.queue_count',(select count(*) from turns where session_id=? and status='queued')),
 updated_at=`+defaultNow+` where id=?`, turnID, sessionID, sessionID)
	if err != nil {
		return 0, err
	}
	_, err = tx.ExecContext(ctx, `update sessions set state_json=json_set(state_json,'$.model',(select json_extract(metadata_json,'$.model') from turns where id=?))
 where id=? and coalesce((select json_extract(metadata_json,'$.model') from turns where id=?),'')!=''`, turnID, sessionID, turnID)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}
