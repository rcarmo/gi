package store

import (
	"context"
	"fmt"
)

// RequestTurnCancellation linearises the abort mark and its checkpoint against
// terminal settlement. A late Cancel cannot reopen a committed terminal turn.
func (s *Store) RequestTurnCancellation(ctx context.Context, turnID, sessionID string) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	var owner, status string
	if err = tx.QueryRowContext(ctx, `select session_id,status from turns where id=?`, turnID).Scan(&owner, &status); err != nil {
		return false, err
	}
	if owner != sessionID {
		return false, fmt.Errorf("cancellation session does not own turn")
	}
	if status == "cancelling" {
		return false, nil
	}
	if status != "running" {
		return false, ErrQueueConflict
	}
	result, err := tx.ExecContext(ctx, `update turns set status='cancelling',phase='cancelling',updated_at=`+defaultNow+` where id=? and status='running'`, turnID)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if n != 1 {
		return false, ErrQueueConflict
	}
	if _, err = tx.ExecContext(ctx, `update subturns set status='cancelling',updated_at=`+defaultNow+` where child_turn_id=?`, turnID); err != nil {
		return false, err
	}
	if err = receiptEventTx(ctx, tx, turnID, sessionID, "turn.cancelling", map[string]any{"phase": "cancel", "checkpoint": true, "reason": "cancel_requested", "status": "cancelling", "turn_phase": "cancelling", "failure_kind": ""}); err != nil {
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
