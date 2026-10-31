package store

import (
	"context"
	"fmt"
)

// TurnSettlement commits the terminal projection as one transaction. Runtime
// events/hooks may be published only if Applied is true after this call.
type TurnSettlement struct {
	Applied        bool
	Status         string
	Phase          string
	MessageID      string
	Message        string
	MessagePayload map[string]any
	FailureKind    string
	Payload        map[string]any
	Accounting     TurnAccounting
	// A changed cancellation decision requires the coordinator to join the
	// newly required critical children before publishing terminal cleanup.
	ChildrenStatus string
}

func (s *Store) SettleRunningTurn(ctx context.Context, turnID, sessionID string, in TurnSettlement) (TurnSettlement, error) {
	out := in
	out.Applied = false
	switch in.Status {
	case "completed", "failed", "cancelled", "aborted":
	default:
		return out, fmt.Errorf("invalid terminal status %q", in.Status)
	}
	ctx, err := admitToolSettlement(ctx)
	if err != nil {
		return out, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	var owner, status string
	if err = tx.QueryRowContext(ctx, `select session_id,status from turns where id=?`, turnID).Scan(&owner, &status); err != nil {
		return out, err
	}
	if owner != sessionID {
		return out, fmt.Errorf("turn settlement session does not own turn")
	}
	if status != "running" && status != "cancelling" {
		var finished int
		if err = tx.QueryRowContext(ctx, `select count(*) from turn_events where turn_id=? and event_type='turn.finished'`, turnID).Scan(&finished); err != nil {
			return out, err
		}
		if finished > 0 {
			out.Status = status
			out.Accounting, err = receiptAccountingTx(ctx, tx, turnID)
			return out, err
		}
		return out, ErrQueueConflict
	}
	if status == "cancelling" {
		out.Status = "cancelled"
		out.Phase = "aborted"
		out.FailureKind = ""
		out.Message = "Turn cancelled"
		out.Payload = nil
		out.MessagePayload = map[string]any{"kind": "chat", "source": "system", "turn_id": turnID, "agent_id": in.MessagePayload["agent_id"], "content_blocks": []any{map[string]any{"type": "turn_outcome_marker", "kind": "cancelled", "severity": "warning", "label": "cancelled"}}}
	}
	if in.ChildrenStatus != "" && out.Status != in.ChildrenStatus {
		return out, ErrQueueConflict
	}
	if out.Message != "" && out.MessageID == "" {
		return out, fmt.Errorf("terminal message requires an identity")
	}
	result, err := tx.ExecContext(ctx, `update turns set status=?,phase=?,finished_at=coalesce(finished_at,`+defaultNow+`),updated_at=`+defaultNow+` where id=? and status=? and (?<>'completed' or not exists(select 1 from turn_failures f where f.turn_id=turns.id and (f.hold_state<>'none' or coalesce(f.resolution_state,'')<>'')))`, out.Status, out.Phase, turnID, status, out.Status)
	if err != nil {
		return out, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return out, err
	}
	if n != 1 {
		return out, ErrFailureConflict
	}
	if out.Status == "completed" {
		if _, err = tx.ExecContext(ctx, `delete from turn_failures where turn_id=? and hold_state='none' and coalesce(resolution_state,'')=''`, turnID); err != nil {
			return out, err
		}
	}
	if _, err = tx.ExecContext(ctx, `update subturns set status=?,finished_at=coalesce(finished_at,`+defaultNow+`),updated_at=`+defaultNow+` where child_turn_id=?`, out.Status, turnID); err != nil {
		return out, err
	}
	if out.Message != "" {
		raw, err := marshalJSON(out.MessagePayload)
		if err != nil {
			return out, err
		}
		if _, err = tx.ExecContext(ctx, `insert into messages(id,session_id,role,content,payload_json,created_at) values(?,?,'system',?,?,`+defaultNow+`)`, out.MessageID, sessionID, out.Message, raw); err != nil {
			return out, err
		}
	}
	if out.FailureKind != "" {
		if err = upsertTurnFailure(ctx, tx, turnID, sessionID, out.FailureKind, "none", out.Message); err != nil {
			return out, err
		}
		if err = receiptEventTx(ctx, tx, turnID, sessionID, "turn.failure_marked", map[string]any{"phase": "recovery", "checkpoint": true, "failure_kind": out.FailureKind, "hold_state": "none", "summary": out.Message}); err != nil {
			return out, err
		}
	}
	out.Accounting, err = storeAccountingTx(ctx, tx, turnID)
	if err != nil {
		return out, err
	}
	if err = receiptEventTx(ctx, tx, turnID, sessionID, "inference.finished", map[string]any{"phase": "inference", "checkpoint": true, "usage": out.Accounting.Map(), "accounting_source": "committed_receipts", "iterations": out.Accounting.Iterations}); err != nil {
		return out, err
	}
	finished := map[string]any{}
	for k, v := range out.Payload {
		finished[k] = v
	}
	finished["phase"] = "turn"
	finished["checkpoint"] = true
	finished["status"] = out.Status
	finished["reason"] = out.Status
	finished["failure_kind"] = out.FailureKind
	if out.FailureKind != "" {
		finished["reason"] = out.FailureKind
	}
	if err = receiptEventTx(ctx, tx, turnID, sessionID, "turn.finished", finished); err != nil {
		return out, err
	}
	// Never clear a newer turn's live session ownership. Queue count and terminal
	// markers nevertheless belong to the same commit as this outcome.
	_, err = tx.ExecContext(ctx, `update sessions set state_json=case when coalesce(json_extract(coalesce(nullif(state_json,''),'{}'),'$.active_turn_id'),'') in ('',?) then json_patch(coalesce(nullif(state_json,''),'{}'),json_object('status','idle','active_turn_id',null,'queue_count',(select count(*) from turns where session_id=? and status='queued'))) else json_set(coalesce(nullif(state_json,''),'{}'),'$.queue_count',(select count(*) from turns where session_id=? and status='queued')) end,updated_at=`+defaultNow+` where id=?`, turnID, sessionID, sessionID, sessionID)
	if err != nil {
		return out, err
	}
	if err = tx.Commit(); err != nil {
		return out, err
	}
	out.Applied = true
	return out, nil
}
