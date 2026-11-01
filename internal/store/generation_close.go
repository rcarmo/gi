package store

import (
	"context"
	"fmt"
)

const ClosedGenerationReadyPhase = "checkpoint_generation_ready"

// IsClosedCheckpointPhase excludes checkpoint reservations from ordinary prompt
// launch. Only a runtime-ready recovery route may adopt them.
func IsClosedCheckpointPhase(phase string) bool {
	return IsClosedProviderRetryPhase(phase) || phase == ClosedGenerationReadyPhase || phase == ClosedToolBatchPhase
}

// ReserveGenerationClose revokes the host's writer without publishing queued
// work. The controller must keep the replacement claim alive, join every owned
// host and callback, then call CommitGenerationClose. A crash during this wait
// retains the original verified frontier for stale-claim recovery.
func (s *Store) ReserveGenerationClose(ctx context.Context, sessionID, turnID, oldToken, token string) (*GenerationReadyCheckpoint, error) {
	if oldToken == "" || token == "" || token == oldToken {
		return nil, fmt.Errorf("generation close requires a fresh claim incarnation")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var current, status, phase string
	if err = tx.QueryRowContext(ctx, `select a.claim_token,t.status,t.phase from session_active_turns a join turns t on t.id=a.turn_id and t.session_id=a.session_id where a.session_id=? and a.turn_id=?`, sessionID, turnID).Scan(&current, &status, &phase); err != nil {
		return nil, err
	}
	if current != oldToken {
		return nil, ErrQueueConflict
	}
	if status != "running" || (phase != "generation_ready" && phase != "running") {
		return nil, nil
	}
	checkpoint, _, err := generationCheckpointTx(ctx, tx, sessionID, turnID, phase)
	if err != nil || checkpoint == nil {
		return nil, err
	}
	result, err := tx.ExecContext(ctx, `update session_active_turns set claim_token=?,worker_id='generation-closing',claimed_at=`+defaultNow+`,updated_at=`+defaultNow+` where session_id=? and turn_id=? and claim_token=?`, token, sessionID, turnID, oldToken)
	if err != nil {
		return nil, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, ErrQueueConflict
	}
	if _, err = tx.ExecContext(ctx, `insert into turn_recovery_fences(turn_id,session_id,claim_token) values(?,?,?) on conflict(turn_id) do update set session_id=excluded.session_id,claim_token=excluded.claim_token`, turnID, sessionID, token); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return checkpoint, nil
}

// CommitGenerationClose publishes a resumable checkpoint only after the
// controller has joined its hosts and closing lease. Cancellation or a changed
// frontier wins. No prompt, lifecycle terminal or crash admission is fabricated.
func (s *Store) CommitGenerationClose(ctx context.Context, sessionID, turnID, token string) error {
	if token == "" {
		return ErrQueueConflict
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var current, status, phase, worker string
	if err = tx.QueryRowContext(ctx, `select a.claim_token,t.status,t.phase,coalesce(a.worker_id,'') from session_active_turns a join turns t on t.id=a.turn_id and t.session_id=a.session_id where a.session_id=? and a.turn_id=?`, sessionID, turnID).Scan(&current, &status, &phase, &worker); err != nil {
		return err
	}
	if current != token || worker != "generation-closing" || status != "running" || (phase != "generation_ready" && phase != "running") {
		return ErrQueueConflict
	}
	checkpoint, requestID, err := generationCheckpointTx(ctx, tx, sessionID, turnID, phase)
	if err != nil {
		return err
	}
	if checkpoint == nil {
		return ErrQueueConflict
	}
	if requestID != "" {
		if err = interruptProviderTx(WithTurnWriter(ctx, sessionID, turnID, token), tx, sessionID, turnID, requestID, checkpoint.Iteration); err != nil {
			return err
		}
	}
	if err = checkGenerationEffectsSettledTx(ctx, tx, turnID); err != nil {
		return err
	}
	if err = receiptEventTx(ctx, tx, turnID, sessionID, "generation.ready", map[string]any{"continuation": checkpoint, "graceful": true, "interrupted_request_id": requestID}); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `update turns set status='queued',phase=?,updated_at=`+defaultNow+` where id=? and session_id=?`, ClosedGenerationReadyPhase, turnID, sessionID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `update session_active_turns set worker_id='closed-checkpoint',updated_at=`+defaultNow+` where session_id=? and turn_id=? and claim_token=?`, sessionID, turnID, token); err != nil {
		return err
	}
	if err = retryCheckpointSessionStateTx(ctx, tx, sessionID, "queued", ""); err != nil {
		return err
	}
	if err = receiptEventTx(ctx, tx, turnID, sessionID, "turn.checkpoint_queued", map[string]any{"phase": ClosedGenerationReadyPhase, "iteration": checkpoint.Iteration, "checkpoint": true, "reason": "graceful_close"}); err != nil {
		return err
	}
	return tx.Commit()
}
