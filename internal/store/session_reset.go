package store

import (
	"context"
	"database/sql"
	"fmt"
)

// ResetContext appends a durable history head without deleting visible messages.
// Optional reset text is the first provider message under the new head.
func (s *Store) ResetContext(ctx context.Context, sessionID, markerID, text string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := resetContextTx(ctx, tx, sessionID, markerID, text); err != nil {
		return err
	}
	return tx.Commit()
}

const contextResetBoundarySQL = ` and (select row_id from message_rows where message_id=messages.id) > coalesce((select max(r.row_id) from message_rows r join messages reset on reset.id=r.message_id where reset.session_id=messages.session_id and json_extract(reset.payload_json,'$.kind')='context_reset'),0)`

func resetContextTx(ctx context.Context, tx *sql.Tx, sessionID, markerID, text string) error {
	var err error
	if _, err = tx.ExecContext(ctx, `insert into messages(id,session_id,role,content,payload_json,created_at) values(?,?,'system','','{"kind":"context_reset"}',`+defaultNow+`)`, markerID, sessionID); err != nil {
		return err
	}
	if text != "" {
		if _, err = tx.ExecContext(ctx, `insert into messages(id,session_id,role,content,payload_json,created_at) values(?,?,'user',?,'{"source":"reset"}',`+defaultNow+`)`, markerID+":prompt", sessionID, text); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `insert into context_checkpoints(session_id,version,summary,covered_json,created_at) values(?,1,'','[]',`+defaultNow+`) on conflict(session_id) do update set version=context_checkpoints.version+1,summary='',covered_json='[]',created_at=excluded.created_at`, sessionID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `update sessions set updated_at=`+defaultNow+` where id=?`, sessionID); err != nil {
		return err
	}
	return nil
}

// RequestContextReset serializes admission against durable active-turn claims.
func (s *Store) RequestContextReset(ctx context.Context, sessionID, text string) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `update sessions set id=id where id=?`, sessionID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if n != 1 {
		return false, sql.ErrNoRows
	}
	var active int
	if err := tx.QueryRowContext(ctx, `select count(*) from session_active_turns where session_id=?`, sessionID).Scan(&active); err != nil {
		return false, err
	}
	if active != 0 {
		var queued int
		if err := tx.QueryRowContext(ctx, `select count(*) from steering_queue where session_id=? and status='queued'`, sessionID).Scan(&queued); err != nil {
			return false, err
		}
		if queued >= 10 {
			return false, fmt.Errorf("steering queue is full")
		}
		if _, err := tx.ExecContext(ctx, `insert into steering_queue(session_id,role,content,payload_json,queue_mode,status,created_at,updated_at) values(?,'user',?,'{"kind":"context_reset"}','one-at-a-time','queued',`+defaultNow+`,`+defaultNow+`)`, sessionID, text); err != nil {
			return false, err
		}
	} else if err := resetContextTx(ctx, tx, sessionID, NowID("reset"), text); err != nil {
		return false, err
	}
	return active != 0, tx.Commit()
}

// ApplyPendingContextReset consumes only a reset at the FIFO head. Reset and
// acknowledgement (or conversion of its text into a follow-up) commit together.
func (s *Store) ApplyPendingContextReset(ctx context.Context, sessionID, turnID string) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `update sessions set id=id where id=?`, sessionID); err != nil {
		return false, err
	}
	var active string
	if err := tx.QueryRowContext(ctx, `select turn_id from session_active_turns where session_id=?`, sessionID).Scan(&active); err != nil {
		return false, err
	}
	if active != turnID {
		return false, ErrQueueConflict
	}
	var id int64
	var text, kind string
	err = tx.QueryRowContext(ctx, `select id,content,coalesce(json_extract(payload_json,'$.kind'),'') from steering_queue where session_id=? and status='queued' order by id limit 1`, sessionID).Scan(&id, &text, &kind)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if kind != "context_reset" {
		return false, nil
	}
	if err := resetContextTx(ctx, tx, sessionID, NowID("reset"), ""); err != nil {
		return false, err
	}
	status := "dequeued"
	if text != "" {
		status = "queued"
	}
	if _, err := tx.ExecContext(ctx, `update steering_queue set status=?,payload_json='{"source":"reset"}' where id=?`, status, id); err != nil {
		return false, err
	}
	return true, tx.Commit()
}
