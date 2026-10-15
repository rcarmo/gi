package store

import (
	"context"
)

// ResetContext appends a durable history head without deleting visible messages.
// Optional reset text is the first provider message under the new head.
func (s *Store) ResetContext(ctx context.Context, sessionID, markerID, text string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
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
	return tx.Commit()
}

const contextResetBoundarySQL = ` and (select row_id from message_rows where message_id=messages.id) > coalesce((select max(r.row_id) from message_rows r join messages reset on reset.id=r.message_id where reset.session_id=messages.session_id and json_extract(reset.payload_json,'$.kind')='context_reset'),0)`
