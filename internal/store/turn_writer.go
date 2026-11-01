package store

import (
	"context"
	"database/sql"
)

type turnWriterKey struct{}
type turnWriter struct{ sessionID, turnID, token string }

// WithTurnWriter binds a native host to one claim incarnation. Context copies
// retain the binding; cancellation and claim revocation remain separate gates.
// Unbound contexts retain the privileged, legacy Store API behaviour.
func WithTurnWriter(ctx context.Context, sessionID, turnID, token string) context.Context {
	return context.WithValue(ctx, turnWriterKey{}, turnWriter{sessionID, turnID, token})
}
func turnWriterToken(ctx context.Context, sessionID, turnID string) (string, error) {
	if ctx == nil {
		return "", nil
	}
	writer, scoped := ctx.Value(turnWriterKey{}).(turnWriter)
	if !scoped {
		return "", nil
	}
	if writer.sessionID != sessionID || writer.turnID != turnID || writer.token == "" {
		return "", ErrQueueConflict
	}
	return writer.token, nil
}
func validateTurnWriterTx(ctx context.Context, tx *sql.Tx, sessionID, turnID string) error {
	token, err := turnWriterToken(ctx, sessionID, turnID)
	if err != nil {
		return err
	}
	return validateTerminalClaimTx(ctx, tx, turnID, sessionID, token)
}

// AppendTurnEffectEvent fences an owned progress/boundary write in the same
// transaction as its journal entry. Administrative events use AppendTurnEvent.
func (s *Store) AppendTurnEffectEvent(ctx context.Context, turnID, sessionID, eventType string, payload map[string]any) error {
	if eventType == "tool.output" {
		payload = toolBatchEvent(ctx, turnID, payload)
	}
	ctx, err := admitToolSettlement(ctx)
	if err != nil {
		return err
	}
	token, err := turnWriterToken(ctx, sessionID, turnID)
	if err != nil {
		return err
	}
	raw, err := marshalJSON(payload)
	if err != nil {
		return err
	}
	// One conditional insert retains atomic fencing without allocating a
	// transaction and two Go query results for each coalesced output update.
	result, err := s.db.ExecContext(ctx, `insert into turn_events(turn_id,session_id,seq,event_type,payload_json,created_at)
	 select ?,?,coalesce((select max(seq)+1 from turn_events where turn_id=?),1),?,?,`+defaultNow+`
	 from turns t where t.id=? and t.session_id=? and (?='' or (
	 not exists(select 1 from session_active_turns a where a.session_id=t.session_id and a.turn_id=t.id and a.claim_token<>?)
	 and not exists(select 1 from turn_recovery_fences f where f.turn_id=t.id and (f.session_id<>t.session_id or f.claim_token<>?))))`, turnID, sessionID, turnID, eventType, raw, turnID, sessionID, token, token, token)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrQueueConflict
	}
	return nil
}

// RecordTurnPromptSections commits an owned preparation's prompt-policy record
// only while that exact turn claim is running. A superseded builder must not
// replace the successor's session policy while its provider admission is fenced.
func (s *Store) RecordTurnPromptSections(ctx context.Context, sessionID, turnID string, sections any) error {
	token, err := turnWriterToken(ctx, sessionID, turnID)
	if err != nil {
		return err
	}
	raw, err := marshalJSON(map[string]any{"system_prompt_sections": sections})
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `update sessions set state_json=json_patch(coalesce(nullif(state_json,''),'{}'),json(?)),updated_at=`+defaultNow+` where id=? and exists(select 1 from turns t join session_active_turns a on a.session_id=t.session_id and a.turn_id=t.id where t.id=? and t.session_id=sessions.id and t.status='running' and (?='' or a.claim_token=?))`, raw, sessionID, turnID, token, token)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrQueueConflict
	}
	return nil
}
