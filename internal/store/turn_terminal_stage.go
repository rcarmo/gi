package store

import (
	"context"
	"fmt"
)

const turnTerminalReceiptQuery = `select exists(select 1 from turns t join turn_events e on e.turn_id=t.id where t.id=? and coalesce(t.finished_at,'')<>'' and t.status in ('completed','failed','cancelled','aborted') and ((e.event_type='turn.finished' and json_extract(e.payload_json,'$.status')=t.status) or (e.event_type='turn.cancelled' and t.status='cancelled' and json_extract(e.payload_json,'$.queued')=1)))`

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
