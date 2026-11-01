package store

import (
	"context"
	"fmt"
)

// AdmitTurnEffect is the host-entry boundary: a committed intent and phase
// transition must both exist before calling a provider or executor. The active
// claim is checked in the same transaction so a committed cancellation wins.
func (s *Store) AdmitTurnEffect(ctx context.Context, sessionID, turnID, phase, eventType string, payload map[string]any) error {
	switch eventType {
	case "inference.started", "inference.request_prepared", "tool.started", "generation.ready", "tool.batch_ready", "tool.preparing", "generation.preparing":
	default:
		return fmt.Errorf("invalid effect intent %q", eventType)
	}
	token, err := turnWriterToken(ctx, sessionID, turnID)
	if err != nil {
		return err
	}
	if eventType == "tool.started" || eventType == "tool.preparing" {
		payload = toolBatchEvent(ctx, turnID, payload)
	}
	raw, err := marshalJSON(payload)
	if err != nil {
		return err
	}
	if len(raw) > 64<<20 {
		return fmt.Errorf("effect intent exceeds 64 MiB")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	phaseGuard := ""
	if eventType == "tool.batch_ready" {
		phaseGuard = " and phase in ('running','waiting_on_tools')"
	} else if eventType == "tool.preparing" {
		phaseGuard = " and phase='waiting_on_tools'"
	}
	if eventType == "generation.ready" || eventType == "generation.preparing" {
		phaseGuard = " and phase in ('setup','running','waiting_on_tools','generation_ready')"
	}
	result, err := tx.ExecContext(ctx, `update turns set phase=?,updated_at=`+defaultNow+` where id=? and session_id=? and status='running' and phase not in ('settling','cancelling')`+phaseGuard+` and exists(select 1 from session_active_turns where session_id=? and turn_id=? and (?='' or claim_token=?))`, phase, turnID, sessionID, sessionID, turnID, token, token)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return context.Canceled
	}
	// Take the writer lock before inspecting the frontier. Upgrading a read
	// transaction can deadlock against a concurrently settling child in SQLite.
	if eventType == "tool.batch_ready" {
		checkpoint, err := decodeToolBatch(raw)
		if err != nil {
			return err
		}
		if err = checkGenerationEffectsSettledTx(ctx, tx, turnID); err != nil {
			return err
		}
		if err = checkToolBatchReceiptTx(ctx, tx, sessionID, turnID, checkpoint); err != nil {
			return err
		}
		var count int
		if err = tx.QueryRowContext(ctx, `select count(*) from turn_events where turn_id=? and event_type='tool.batch_ready' and json_extract(payload_json,'$.continuation.provider_occurrence_id')=?`, turnID, checkpoint.ProviderOccurrenceID).Scan(&count); err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("duplicate tool batch")
		}
	}
	if eventType == "tool.started" || eventType == "tool.preparing" {
		if err := checkToolBatchSlotTx(ctx, tx, sessionID, turnID); err != nil {
			return err
		}
	}
	if eventType == "generation.ready" || eventType == "generation.preparing" {
		if err := checkGenerationEffectsSettledTx(ctx, tx, turnID); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `insert into turn_events(turn_id,session_id,seq,event_type,payload_json,created_at) values(?,?,coalesce((select max(seq)+1 from turn_events where turn_id=?),1),?,?,`+defaultNow+`)`, turnID, sessionID, turnID, eventType, raw); err != nil {
		return err
	}
	return tx.Commit()
}
