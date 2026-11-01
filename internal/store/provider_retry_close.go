package store

import (
	"context"
	"database/sql"
)

const (
	ClosedProviderRetryWaitPhase  = "checkpoint_retry_wait"
	ClosedProviderRetryReadyPhase = "checkpoint_retry_ready"
)

func IsClosedProviderRetryPhase(phase string) bool {
	return phase == ClosedProviderRetryWaitPhase || phase == ClosedProviderRetryReadyPhase
}

// CommitProviderRetryClose is only called after the reserving engine has joined
// the main host, progress writers and owned callbacks. Keep the claim as a
// checkpoint reservation, never as authority for a whole-prompt queue launch.
// A cancellation mark or changed frontier wins over graceful requeue.
func (s *Store) CommitProviderRetryClose(ctx context.Context, sessionID, turnID, token string) error {
	if token == "" {
		return ErrQueueConflict
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status, phase, worker, current string
	if err = tx.QueryRowContext(ctx, `select t.status,t.phase,coalesce(a.worker_id,''),a.claim_token from turns t join session_active_turns a on a.turn_id=t.id and a.session_id=t.session_id where t.id=? and t.session_id=?`, turnID, sessionID).Scan(&status, &phase, &worker, &current); err != nil {
		return err
	}
	if current != token || worker != "retry-closing" || status != "running" || (phase != "running" && phase != "retry_wait") {
		return ErrQueueConflict
	}
	boundary := "inference.retry_scheduled"
	resumed := phase == "running"
	nextPhase := ClosedProviderRetryWaitPhase
	if resumed {
		boundary = "inference.retry_resumed"
		nextPhase = ClosedProviderRetryReadyPhase
	}
	checkpoint, err := providerRetryCheckpointTx(ctx, tx, sessionID, turnID, boundary, resumed)
	if err != nil {
		return err
	}
	if checkpoint == nil {
		return ErrQueueConflict
	}
	if _, err = tx.ExecContext(ctx, `update turns set status='queued',phase=?,updated_at=`+defaultNow+` where id=? and session_id=?`, nextPhase, turnID, sessionID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `update session_active_turns set worker_id='closed-checkpoint',updated_at=`+defaultNow+` where session_id=? and turn_id=? and claim_token=?`, sessionID, turnID, token); err != nil {
		return err
	}
	if err = retryCheckpointSessionStateTx(ctx, tx, sessionID, "queued", ""); err != nil {
		return err
	}
	if err = receiptEventTx(ctx, tx, turnID, sessionID, "turn.checkpoint_queued", map[string]any{"phase": nextPhase, "checkpoint": true, "reason": "graceful_close", "request_id": checkpoint.OccurrenceID}); err != nil {
		return err
	}
	return tx.Commit()
}

func retryCheckpointSessionStateTx(ctx context.Context, tx *sql.Tx, sessionID, status, active string) error {
	var activeValue any
	if active != "" {
		activeValue = active
	}
	_, err := tx.ExecContext(ctx, `update sessions set state_json=json_set(coalesce(nullif(state_json,''),'{}'),'$.status',?,'$.active_turn_id',?,'$.queue_count',(select count(*) from turns where session_id=? and status='queued')),updated_at=`+defaultNow+` where id=?`, status, activeValue, sessionID, sessionID)
	return err
}
