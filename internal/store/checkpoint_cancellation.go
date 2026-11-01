package store

import "context"

// StageClosedCheckpointCancellation is control admission, not executor recovery.
// A closed reservation proves its prior host joined. Rotate its writer and stage
// cancellation atomically so a crash can only resume terminal cleanup, never SDK
// execution. The staged payload retains the kind of tool cleanup still required.
func (s *Store) StageClosedCheckpointCancellation(ctx context.Context, sessionID, turnID, token string) (*TurnSettlement, error) {
	if token == "" {
		return nil, ErrQueueConflict
	}
	ctx, err := admitToolSettlement(ctx)
	if err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `update session_active_turns set updated_at=updated_at where session_id=? and turn_id=?`, sessionID, turnID); err != nil {
		return nil, err
	}
	var status, phase, worker, old string
	if err = tx.QueryRowContext(ctx, `select t.status,t.phase,a.worker_id,a.claim_token from turns t join session_active_turns a on a.turn_id=t.id and a.session_id=t.session_id where t.id=? and t.session_id=?`, turnID, sessionID).Scan(&status, &phase, &worker, &old); err != nil {
		return nil, err
	}
	if status != "queued" || worker != "closed-checkpoint" || !IsClosedCheckpointPhase(phase) || old == "" || token == old {
		return nil, ErrQueueConflict
	}
	switch phase {
	case ClosedToolBatchPhase:
		cp, readErr := toolBatchCheckpointTx(ctx, tx, sessionID, turnID)
		if readErr != nil {
			return nil, readErr
		}
		if cp == nil {
			return nil, ErrQueueConflict
		}
	case ClosedGenerationReadyPhase:
		cp, _, readErr := generationCheckpointTx(ctx, tx, sessionID, turnID, "generation_ready")
		if readErr != nil {
			return nil, readErr
		}
		if cp == nil {
			return nil, ErrQueueConflict
		}
	default:
		boundary, resumed := "inference.retry_scheduled", false
		if phase == ClosedProviderRetryReadyPhase {
			boundary, resumed = "inference.retry_resumed", true
		}
		cp, readErr := providerRetryCheckpointTx(ctx, tx, sessionID, turnID, boundary, resumed)
		if readErr != nil {
			return nil, readErr
		}
		if cp == nil {
			return nil, ErrQueueConflict
		}
	}
	if _, err = tx.ExecContext(ctx, `update session_active_turns set claim_token=?,worker_id='terminal-cancellation',claimed_at=`+defaultNow+`,updated_at=`+defaultNow+` where session_id=? and turn_id=?`, token, sessionID, turnID); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `insert into turn_recovery_fences(turn_id,session_id,claim_token) values(?,?,?) on conflict(turn_id) do update set session_id=excluded.session_id,claim_token=excluded.claim_token`, turnID, sessionID, token); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `update turns set status='cancelling',phase='settling',updated_at=`+defaultNow+` where id=? and session_id=?`, turnID, sessionID); err != nil {
		return nil, err
	}
	decision := &TurnSettlement{Status: "cancelled", Phase: "cancelled", ClaimToken: token, MessageID: NowID("msg"), Message: "Turn cancelled", MessagePayload: map[string]any{"kind": "chat", "source": "system", "turn_id": turnID}, Payload: map[string]any{"checkpoint_cancellation": phase, "reason": "checkpoint_cancel"}}
	if err = receiptEventTx(ctx, tx, turnID, sessionID, "turn.terminal_staged", map[string]any{
		"phase": "settling", "checkpoint": true, "status": "cancelled", "previous_phase": phase,
		"decision": map[string]any{"status": decision.Status, "phase": decision.Phase, "message_id": decision.MessageID, "message": decision.Message, "message_payload": decision.MessagePayload, "payload": decision.Payload},
	}); err != nil {
		return nil, err
	}
	if err = retryCheckpointSessionStateTx(ctx, tx, sessionID, "running", turnID); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return decision, nil
}

// LoadCancelledToolBatchCheckpoint reads the validated journal, including settled slots.
// It grants no execution authority. Terminal cleanup uses it after an owner crash.
func (s *Store) LoadCancelledToolBatchCheckpoint(ctx context.Context, sessionID, turnID string) (*ToolBatchRecovery, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	return readToolBatchCheckpointTx(ctx, tx, sessionID, turnID, true)
}
