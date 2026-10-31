package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

func (s *Store) HasStagedTerminalCheckpoint(ctx context.Context, turnID string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `select exists(select 1 from turn_events where turn_id=? and event_type='turn.terminal_staged')`, turnID).Scan(&exists)
	return exists, err
}

// AcquireStagedTerminalRecovery adopts only an existing terminal decision. It
// never rewinds an inference/tool phase or authorises an external dispatch.
// A nil result means the turn has no eligible staged-terminal checkpoint.
func (s *Store) AcquireStagedTerminalRecovery(ctx context.Context, claim ActiveTurnClaim, newToken string, cutoff time.Time, abort bool) (*TurnSettlement, error) {
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
	if phase != "settling" && phase != "cancelling" {
		return nil, nil
	}
	if status != "running" && status != "cancelling" {
		return nil, nil
	}
	var raw string
	err = tx.QueryRowContext(ctx, `select payload_json from turn_events where turn_id=? and session_id=? and event_type='turn.terminal_staged' order by seq desc limit 1`, claim.TurnID, claim.SessionID).Scan(&raw)
	if err == sql.ErrNoRows && phase != "settling" {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load staged terminal decision: %w", err)
	}
	var saved struct {
		Decision struct {
			Status         string         `json:"status"`
			Phase          string         `json:"phase"`
			MessageID      string         `json:"message_id"`
			Message        string         `json:"message"`
			MessagePayload map[string]any `json:"message_payload"`
			FailureKind    string         `json:"failure_kind"`
			Payload        map[string]any `json:"payload"`
		} `json:"decision"`
	}
	if err = json.Unmarshal([]byte(raw), &saved); err != nil {
		return nil, err
	}
	in := TurnSettlement{Status: saved.Decision.Status, Phase: saved.Decision.Phase, MessageID: saved.Decision.MessageID, Message: saved.Decision.Message, MessagePayload: saved.Decision.MessagePayload, FailureKind: saved.Decision.FailureKind, Payload: saved.Decision.Payload, ClaimToken: newToken}
	switch in.Status {
	case "completed", "failed", "cancelled", "aborted":
	default:
		return nil, fmt.Errorf("invalid staged terminal status %q", in.Status)
	}
	wantPhase := in.Status
	if wantPhase == "aborted" {
		wantPhase = "cancelled"
	}
	if in.Phase != wantPhase {
		return nil, fmt.Errorf("invalid staged terminal phase %q for %s", in.Phase, in.Status)
	}
	if in.Message != "" && in.MessageID == "" {
		return nil, fmt.Errorf("staged terminal message lacks identity")
	}
	if id, ok := in.MessagePayload["turn_id"].(string); ok && id != "" && id != claim.TurnID {
		return nil, fmt.Errorf("staged terminal message belongs to another turn")
	}
	stamp, err := time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return nil, err
	}
	if token != claim.ClaimToken || updated != claim.UpdatedAt || !stamp.Before(cutoff) {
		return nil, ErrQueueConflict
	}
	result, err := tx.ExecContext(ctx, `update session_active_turns set claim_token=?,worker_id='terminal-recovery',claimed_at=`+defaultNow+`,updated_at=`+defaultNow+` where session_id=? and turn_id=? and claim_token=? and updated_at=?`, newToken, claim.SessionID, claim.TurnID, claim.ClaimToken, claim.UpdatedAt)
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
	if abort && status == "running" {
		if _, err = tx.ExecContext(ctx, `update turns set status='cancelling',updated_at=`+defaultNow+` where id=?`, claim.TurnID); err != nil {
			return nil, err
		}
		if err = receiptEventTx(ctx, tx, claim.TurnID, claim.SessionID, "turn.cancelling", map[string]any{"phase": "cancel", "checkpoint": true, "status": "cancelling", "reason": "stale_abort_requested"}); err != nil {
			return nil, err
		}
	}
	if err = receiptEventTx(ctx, tx, claim.TurnID, claim.SessionID, "turn.recovery_admitted", map[string]any{"phase": "recovery", "checkpoint": true, "reason": "resume_terminal_decision", "recovery_disposition": "resume_terminal_decision", "previous_phase": phase}); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &in, nil
}
