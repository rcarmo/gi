package store

import (
	"context"
	"fmt"
)

// ReserveToolBatchClose fences the executor while retaining the native batch.
// Only bounded output/usage recorders may use the new incarnation. The owner
// must join every callback and closing lease before publishing queued work.
func (s *Store) ReserveToolBatchClose(ctx context.Context, sessionID, turnID, oldToken, token string) (*ToolBatchRecovery, error) {
	if oldToken == "" || token == "" || token == oldToken {
		return nil, fmt.Errorf("tool close requires a fresh claim incarnation")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// Serialize before reading the journal: SDK progress can still be arriving.
	if _, err = tx.ExecContext(ctx, `update session_active_turns set updated_at=updated_at where session_id=? and turn_id=?`, sessionID, turnID); err != nil {
		return nil, err
	}
	var current, status, phase string
	if err = tx.QueryRowContext(ctx, `select a.claim_token,t.status,t.phase from session_active_turns a join turns t on t.id=a.turn_id and t.session_id=a.session_id where a.session_id=? and a.turn_id=?`, sessionID, turnID).Scan(&current, &status, &phase); err != nil {
		return nil, err
	}
	if current != oldToken {
		return nil, ErrQueueConflict
	}
	if status != "running" || phase != "waiting_on_tools" {
		return nil, nil
	}
	checkpoint, err := toolBatchCheckpointTx(ctx, tx, sessionID, turnID)
	if err != nil || checkpoint == nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `update session_active_turns set claim_token=?,worker_id='tool-closing',claimed_at=`+defaultNow+`,updated_at=`+defaultNow+` where session_id=? and turn_id=? and claim_token=?`, token, sessionID, turnID, oldToken); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `insert into turn_recovery_fences(turn_id,session_id,claim_token) values(?,?,?) on conflict(turn_id) do update set session_id=excluded.session_id,claim_token=excluded.claim_token`, turnID, sessionID, token); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return checkpoint, nil
}

// CommitToolBatchClose requires a quiescent owner, not just a cancelled context.
// No result, terminal lifecycle or crash-budget reset is invented by shutdown.
func (s *Store) CommitToolBatchClose(ctx context.Context, sessionID, turnID, token string) error {
	if token == "" {
		return ErrQueueConflict
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `update session_active_turns set updated_at=updated_at where session_id=? and turn_id=?`, sessionID, turnID); err != nil {
		return err
	}
	var current, status, phase, worker string
	if err = tx.QueryRowContext(ctx, `select a.claim_token,t.status,t.phase,coalesce(a.worker_id,'') from session_active_turns a join turns t on t.id=a.turn_id and t.session_id=a.session_id where a.session_id=? and a.turn_id=?`, sessionID, turnID).Scan(&current, &status, &phase, &worker); err != nil {
		return err
	}
	if current != token || worker != "tool-closing" || status != "running" || phase != "waiting_on_tools" {
		return ErrQueueConflict
	}
	state, err := toolBatchCheckpointTx(ctx, tx, sessionID, turnID)
	if err != nil {
		return err
	}
	if state == nil {
		return ErrQueueConflict
	}
	if _, err = tx.ExecContext(ctx, `update turns set status='queued',phase=?,updated_at=`+defaultNow+` where id=? and session_id=?`, ClosedToolBatchPhase, turnID, sessionID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `update session_active_turns set worker_id='closed-checkpoint',updated_at=`+defaultNow+` where session_id=? and turn_id=? and claim_token=?`, sessionID, turnID, token); err != nil {
		return err
	}
	if err = retryCheckpointSessionStateTx(ctx, tx, sessionID, "queued", ""); err != nil {
		return err
	}
	if err = receiptEventTx(ctx, tx, turnID, sessionID, "turn.checkpoint_queued", map[string]any{"phase": ClosedToolBatchPhase, "iteration": state.Checkpoint.Generation.Iteration, "batch_id": state.Checkpoint.ProviderOccurrenceID, "checkpoint": true, "reason": "graceful_close"}); err != nil {
		return err
	}
	return tx.Commit()
}
