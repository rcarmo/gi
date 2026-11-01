package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	goai "github.com/rcarmo/go-ai"
)

// SettleToolResult commits the replay result and its terminal boundary together.
// Callers must not publish completion or continue the provider before it succeeds.
func (s *Store) SettleToolResult(ctx context.Context, turnID, sessionID, eventType string, event map[string]any, messageID, content string, payload map[string]any) error {
	return s.settleToolResult(ctx, turnID, sessionID, eventType, event, messageID, content, payload, "tool_result")
}

// SettleToolResponse commits a terminal tool's user-facing assistant response
// with its receipt. Empty output needs no history row. Parent settlement still
// joins required children and commits its own terminal projection separately.
func (s *Store) SettleToolResponse(ctx context.Context, turnID, sessionID, eventType string, event map[string]any, messageID, content string, payload map[string]any) error {
	if messageID == "" && content != "" {
		return fmt.Errorf("tool response requires a message identity")
	}
	return s.settleToolResult(ctx, turnID, sessionID, eventType, event, messageID, content, payload, "assistant")
}

func (s *Store) settleToolResult(ctx context.Context, turnID, sessionID, eventType string, event map[string]any, messageID, content string, payload map[string]any, messageRole string) error {
	if eventType != "tool.finished" && eventType != "tool.failed" && eventType != "tool.skipped" && eventType != "tool.cancelled" && eventType != "tool.aborted" && eventType != "tool.interrupted" {
		return fmt.Errorf("invalid tool settlement event %q", eventType)
	}
	if eventType == "tool.interrupted" && ToolBatchID(ctx, turnID) == "" {
		return fmt.Errorf("tool interruption requires a native batch")
	}
	occurrence, _ := event["occurrence_id"].(string)
	if occurrence == "" {
		occurrence, _ = payload["occurrence_id"].(string)
	}
	if occurrence == "" {
		occurrence = "tool-message:" + messageID
	}
	if occurrence == "tool-message:" {
		return fmt.Errorf("tool settlement requires an occurrence")
	}
	if supplied, _ := event["occurrence_id"].(string); supplied == "" {
		owned := make(map[string]any, len(event)+1)
		for key, value := range event {
			owned[key] = value
		}
		owned["occurrence_id"] = occurrence
		event = owned
	}
	event = toolBatchEvent(ctx, turnID, event)
	if eventType == "tool.finished" && ToolBatchID(ctx, turnID) != "" {
		if slot, _ := ctx.Value(toolBatchSlotKey{}).(toolBatchSlot); slot.Index >= 0 {
			event["failure_signature"], event["repeated_tool_failures"] = "", 0
		}
	}
	eventJSON, err := marshalJSON(event)
	if err != nil {
		return err
	}
	if wireID, wireName := ToolBatchWireCall(ctx, turnID); wireID != "" && messageRole == "tool_result" {
		owned := make(map[string]any, len(payload)+2)
		for key, value := range payload {
			owned[key] = value
		}
		owned["tool_call_id"], owned["tool_name"] = wireID, wireName
		payload = owned
	}
	resultJSON, err := marshalJSON(payload)
	if err != nil {
		return err
	}
	// Admission is the cancellation boundary. Once admitted, settle both rows
	// even if the caller cancels while SQLite is committing the result.
	ctx, err = admitToolSettlement(ctx)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = validateTurnWriterTx(ctx, tx, sessionID, turnID); err != nil {
		return err
	}
	// Foreign keys validate each ID separately, not their ownership relation.
	var owner, status string
	if err := tx.QueryRowContext(ctx, `select session_id,status from turns where id=?`, turnID).Scan(&owner, &status); err != nil {
		return err
	}
	if owner != sessionID {
		return fmt.Errorf("tool settlement session does not own turn")
	}
	if messageRole == "assistant" && status == "cancelling" {
		// Cancellation can commit before the host's context is cancelled.
		// Retain the observed response without adopting it into model context.
		partial := make(map[string]any, len(payload)+4)
		for k, v := range payload {
			partial[k] = v
		}
		partial["partial"], partial["adopted"], partial["context_contribution"], partial["stop_reason"] = true, false, false, goai.StopReasonAborted
		resultJSON, err = marshalJSON(partial)
		if err != nil {
			return err
		}
	}
	var existingKind string
	if err := tx.QueryRowContext(ctx, `select kind from turn_receipts where turn_id=? and occurrence_id=?`, turnID, occurrence).Scan(&existingKind); err == nil {
		if existingKind != "tool" {
			return fmt.Errorf("tool occurrence belongs to %s", existingKind)
		}
		return nil
	} else if err != sql.ErrNoRows {
		return err
	}
	if err = checkToolBatchSlotTx(ctx, tx, sessionID, turnID); err != nil {
		return err
	}
	usage := &goai.Usage{}
	if value, ok := payload["usage"]; ok {
		if value == nil {
			usage = nil
		} else {
			raw, err := marshalJSON(value)
			if err != nil {
				return err
			}
			if err = json.Unmarshal([]byte(raw), usage); err != nil {
				return err
			}
		}
	}
	unknown, _ := payload["usage_unknown"].(bool)
	if err := insertReceiptTx(ctx, tx, turnID, occurrence, "tool", messageID, usage, unknown, map[string]any{"next": "tool_boundary", "event_type": eventType}); err != nil {
		return err
	}
	if _, err := storeAccountingTx(ctx, tx, turnID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `insert into turn_events(turn_id,session_id,seq,event_type,payload_json,created_at) values(?,?,coalesce((select max(seq)+1 from turn_events where turn_id=?),1),?,?,`+defaultNow+`)`, turnID, sessionID, turnID, eventType, eventJSON); err != nil {
		return err
	}
	if messageID != "" {
		if _, err = tx.ExecContext(ctx, `insert into messages(id,session_id,role,content,payload_json,created_at) values(?,?,?,?,?,`+defaultNow+`)`, messageID, sessionID, messageRole, content, resultJSON); err != nil {
			return err
		}
	} else if messageRole != "assistant" && eventType != "tool.cancelled" && eventType != "tool.aborted" {
		pending, _ := event["replay_pending"].(bool)
		if eventType != "tool.interrupted" || !pending {
			return fmt.Errorf("tool result requires a message identity")
		}
	}
	if _, err = tx.ExecContext(ctx, `update turns set updated_at=`+defaultNow+` where id=?`, turnID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `update sessions set updated_at=`+defaultNow+` where id=?`, sessionID); err != nil {
		return err
	}
	return tx.Commit()
}

func admitToolSettlement(ctx context.Context) (context.Context, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil tool settlement context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return context.WithoutCancel(ctx), nil
}
