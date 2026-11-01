package store

import (
	"context"
	"database/sql"
	"fmt"
)

const turnTerminalReceiptQuery = `select exists(select 1 from turns t join turn_events e on e.turn_id=t.id where t.id=? and coalesce(t.finished_at,'')<>'' and t.status in ('completed','failed','cancelled','aborted') and ((e.event_type='turn.finished' and json_extract(e.payload_json,'$.status')=t.status) or (e.event_type='turn.cancelled' and t.status='cancelled' and json_extract(e.payload_json,'$.queued')=1)))`

func validateTerminalClaimTx(ctx context.Context, tx *sql.Tx, turnID, sessionID, token string) error {
	if token == "" {
		return nil
	}
	// A disappearing legacy claim cancels its host; that host must still be
	// able to retire the abandoned turn. A recovery takeover, unlike mere
	// release, permanently revokes the previous writer even after cleanup.
	var replaced bool
	if err := tx.QueryRowContext(ctx, `select
	 exists(select 1 from session_active_turns where session_id=? and turn_id=? and claim_token<>?)
	 or exists(select 1 from turn_recovery_fences where turn_id=? and (session_id<>? or claim_token<>?))`, sessionID, turnID, token, turnID, sessionID, token).Scan(&replaced); err != nil {
		return err
	}
	if replaced {
		return ErrQueueConflict
	}
	return nil
}

func (s *Store) HasTurnTerminalReceipt(ctx context.Context, turnID string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, turnTerminalReceiptQuery, turnID).Scan(&exists)
	return exists, err
}

// StageTurnTerminal records the parent's decision while retaining its live
// claim. Child admission is fenced until settlement commits. The decision is
// a recovery checkpoint; it never authorises replay of a provider or tool.
func (s *Store) StageTurnTerminal(ctx context.Context, turnID, sessionID string, in TurnSettlement) (status string, pending bool, err error) {
	switch in.Status {
	case "completed", "failed", "cancelled", "aborted":
	default:
		return "", false, fmt.Errorf("invalid terminal status %q", in.Status)
	}
	ctx, err = admitToolSettlement(ctx)
	if err != nil {
		return "", false, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback()
	if err = validateTerminalClaimTx(ctx, tx, turnID, sessionID, in.ClaimToken); err != nil {
		return "", false, err
	}
	var owner, previousPhase string
	if err = tx.QueryRowContext(ctx, `select session_id,status,phase from turns where id=?`, turnID).Scan(&owner, &status, &previousPhase); err != nil {
		return "", false, err
	}
	if owner != sessionID {
		return "", false, fmt.Errorf("terminal decision session does not own turn")
	}
	if status != "running" && status != "cancelling" {
		var terminal bool
		if err = tx.QueryRowContext(ctx, turnTerminalReceiptQuery, turnID).Scan(&terminal); err != nil {
			return "", false, err
		}
		if terminal {
			return status, false, nil
		}
		return "", false, ErrQueueConflict
	}
	if status == "cancelling" {
		status = "cancelled"
	} else {
		status = in.Status
	}
	if _, err = tx.ExecContext(ctx, `update turns set phase='settling',updated_at=`+defaultNow+` where id=?`, turnID); err != nil {
		return "", false, err
	}
	if err = receiptEventTx(ctx, tx, turnID, sessionID, "turn.terminal_staged", map[string]any{
		"phase": "settling", "checkpoint": true, "status": status, "previous_phase": previousPhase,
		"decision": map[string]any{"status": in.Status, "phase": in.Phase, "message_id": in.MessageID, "message": in.Message, "message_payload": in.MessagePayload, "failure_kind": in.FailureKind, "payload": in.Payload},
	}); err != nil {
		return "", false, err
	}
	if err = tx.Commit(); err != nil {
		return "", false, err
	}
	return status, true, nil
}

// ClaimAgentEnd records a once-only lifecycle attempt under the current writer.
// A terminal-only recovery host must not repeat an already attempted callback.
func (s *Store) ClaimAgentEnd(ctx context.Context, sessionID, turnID string) (bool, error) {
	ctx, err := admitToolSettlement(ctx)
	if err != nil {
		return false, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `update turns set updated_at=updated_at where id=? and session_id=?`, turnID, sessionID)
	if err != nil {
		return false, err
	}
	if n, err := result.RowsAffected(); err != nil || n != 1 {
		if err != nil {
			return false, err
		}
		return false, ErrQueueConflict
	}
	if err = validateTurnWriterTx(ctx, tx, sessionID, turnID); err != nil {
		return false, err
	}
	var attempted bool
	if err = tx.QueryRowContext(ctx, `select exists(select 1 from turn_events where turn_id=? and event_type='turn.agent_end_started')`, turnID).Scan(&attempted); err != nil {
		return false, err
	}
	if attempted {
		return false, nil
	}
	if err = receiptEventTx(ctx, tx, turnID, sessionID, "turn.agent_end_started", map[string]any{"phase": "agent_end", "checkpoint": true}); err != nil {
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
