package store

import (
	"context"
	"database/sql"
	"errors"
)

var ErrMessageDeleteBusy = errors.New("Session has active or queued work")
var ErrMessageDeleteProtected = errors.New("Only user and assistant messages can be deleted")

// DeleteMessage removes one flat timeline row, never its audit turn/events or
// stored media. A checkpoint reset in the same transaction advances its version
// so a concurrent snapshot cannot publish a summary containing removed content.
func (s *Store) DeleteMessage(ctx context.Context, sessionID, messageID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var role string
	if err = tx.QueryRowContext(ctx, "SELECT role FROM messages WHERE session_id=? AND id=?", sessionID, messageID).Scan(&role); err != nil {
		return err
	}
	if role != "user" && role != "assistant" {
		return ErrMessageDeleteProtected
	}
	var busy bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM session_active_turns WHERE session_id=?) OR EXISTS(SELECT 1 FROM turns WHERE session_id=? AND status IN ('running','queued','cancelling'))`, sessionID, sessionID).Scan(&busy); err != nil {
		return err
	}
	if busy {
		return ErrMessageDeleteBusy
	}
	result, err := tx.ExecContext(ctx, "DELETE FROM messages WHERE session_id=? AND id=?", sessionID, messageID)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return sql.ErrNoRows
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO context_checkpoints(session_id,version,summary,covered_json,created_at) VALUES(?,1,'','[]',`+defaultNow+`)
 ON CONFLICT(session_id) DO UPDATE SET version=version+1,summary='',covered_json='[]',created_at=excluded.created_at`, sessionID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE sessions SET updated_at="+defaultNow+" WHERE id=?", sessionID); err != nil {
		return err
	}
	return tx.Commit()
}
