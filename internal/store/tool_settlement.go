package store

import (
	"context"
	"fmt"
)

// SettleToolResult commits the replay result and its terminal boundary together.
// Callers must not publish completion or continue the provider before it succeeds.
func (s *Store) SettleToolResult(ctx context.Context, turnID, sessionID, eventType string, event map[string]any, messageID, content string, payload map[string]any) error {
	if eventType != "tool.finished" && eventType != "tool.failed" && eventType != "tool.skipped" {
		return fmt.Errorf("invalid tool settlement event %q", eventType)
	}
	eventJSON, err := marshalJSON(event)
	if err != nil {
		return err
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
	// Foreign keys validate each ID separately, not their ownership relation.
	var owner string
	if err := tx.QueryRowContext(ctx, `select session_id from turns where id=?`, turnID).Scan(&owner); err != nil {
		return err
	}
	if owner != sessionID {
		return fmt.Errorf("tool settlement session does not own turn")
	}
	if _, err = tx.ExecContext(ctx, `insert into turn_events(turn_id,session_id,seq,event_type,payload_json,created_at) values(?,?,coalesce((select max(seq)+1 from turn_events where turn_id=?),1),?,?,`+defaultNow+`)`, turnID, sessionID, turnID, eventType, eventJSON); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `insert into messages(id,session_id,role,content,payload_json,created_at) values(?,?,'tool_result',?,?,`+defaultNow+`)`, messageID, sessionID, content, resultJSON); err != nil {
		return err
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
