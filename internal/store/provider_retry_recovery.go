package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// ProviderRetryCheckpoint resumes a committed failed-attempt successor. It
// carries lifecycle values and budget position, never a dispatched request.
type ProviderRetryCheckpoint struct {
	OccurrenceID string
	Wait         *ProviderRetryWait
}

// AcquireProviderRetryRecovery rotates a stale incarnation only while its
// current retry checkpoint remains complete and recoverable. Nil leaves legacy
// checkpoints on the existing conservative hold path. No effect is dispatched.
func (s *Store) AcquireProviderRetryRecovery(ctx context.Context, claim ActiveTurnClaim, newToken string, cutoff time.Time) (*ProviderRetryCheckpoint, error) {
	if newToken == "" || newToken == claim.ClaimToken {
		return nil, fmt.Errorf("recovery requires a fresh claim incarnation")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var status, phase, token, updated string
	if err = tx.QueryRowContext(ctx, `select t.status,t.phase,a.claim_token,a.updated_at from turns t join session_active_turns a on a.turn_id=t.id and a.session_id=t.session_id where t.id=? and t.session_id=?`, claim.TurnID, claim.SessionID).Scan(&status, &phase, &token, &updated); err != nil {
		return nil, err
	}
	if status != "running" || phase != "retry_wait" {
		return nil, nil
	}
	stamp, err := time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return nil, err
	}
	if token != claim.ClaimToken || updated != claim.UpdatedAt || !stamp.Before(cutoff) {
		return nil, ErrQueueConflict
	}
	var occurrence string
	err = tx.QueryRowContext(ctx, `select json_extract(payload_json,'$.request_id') from turn_events where turn_id=? and session_id=? and event_type='inference.retry_scheduled' order by seq desc limit 1`, claim.TurnID, claim.SessionID).Scan(&occurrence)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err = validateProviderRetryFrontierTx(ctx, tx, claim.TurnID, occurrence); err != nil {
		return nil, err
	}
	var raw string
	if err = tx.QueryRowContext(ctx, `select decision_json from turn_receipts where turn_id=? and occurrence_id=? and kind='provider'`, claim.TurnID, occurrence).Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var decision struct {
		Next      string             `json:"next"`
		Iteration int                `json:"iteration"`
		Wait      *ProviderRetryWait `json:"retry_wait"`
	}
	if err = json.Unmarshal([]byte(raw), &decision); err != nil {
		return nil, nil // corrupt/incomplete policy is held, never guessed
	}
	if decision.Next != "retry_wait" || decision.Wait == nil {
		return nil, nil
	}
	if err = decision.Wait.validate(); err != nil {
		return nil, nil
	}
	if decision.Wait.Iteration < 1 || decision.Wait.Iteration != decision.Iteration || decision.Wait.Lifecycle == nil {
		return nil, nil
	}
	var held bool
	if err = tx.QueryRowContext(ctx, `select exists(select 1 from turn_failures where turn_id=? and (hold_state<>'none' or coalesce(resolution_state,'')<>''))`, claim.TurnID).Scan(&held); err != nil {
		return nil, err
	}
	if held {
		return nil, ErrQueueConflict
	}
	result, err := tx.ExecContext(ctx, `update session_active_turns set claim_token=?,worker_id='retry-recovery',claimed_at=`+defaultNow+`,updated_at=`+defaultNow+` where session_id=? and turn_id=? and claim_token=? and updated_at=?`, newToken, claim.SessionID, claim.TurnID, claim.ClaimToken, claim.UpdatedAt)
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
	if _, err = tx.ExecContext(ctx, `insert into turn_recovery_fences(turn_id,session_id,claim_token) values(?,?,?) on conflict(turn_id) do update set session_id=excluded.session_id,claim_token=excluded.claim_token`, claim.TurnID, claim.SessionID, newToken); err != nil {
		return nil, err
	}
	if err = receiptEventTx(ctx, tx, claim.TurnID, claim.SessionID, "turn.recovery_admitted", map[string]any{"phase": "recovery", "checkpoint": true, "reason": "resume_provider_retry", "recovery_disposition": "resume_provider_retry", "previous_phase": phase}); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &ProviderRetryCheckpoint{OccurrenceID: occurrence, Wait: decision.Wait}, nil
}
