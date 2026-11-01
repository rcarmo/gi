package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ProviderRetryCheckpoint resumes a committed failed-attempt successor. It
// carries lifecycle values and budget position, never a dispatched request.
type ProviderRetryCheckpoint struct {
	OccurrenceID string
	Wait         *ProviderRetryWait
	Resumed      bool // wake committed before the owner stopped; do not wake again
}

// AcquireProviderRetryRecovery rotates a stale incarnation only while its
// current retry checkpoint remains complete and recoverable. Nil leaves legacy
// checkpoints on the existing conservative hold path. No effect is dispatched.
func (s *Store) AcquireProviderRetryRecovery(ctx context.Context, claim ActiveTurnClaim, newToken string, cutoff time.Time) (*ProviderRetryCheckpoint, error) {
	return s.acquireProviderRetryCheckpoint(ctx, claim, newToken, cutoff, false)
}

// ReserveProviderRetryClose seals the current retry frontier before cancelling
// its host. This is NOT a requeue: the caller must keep the new claim alive,
// join all owned work, and only then call CommitProviderRetryClose.
func (s *Store) ReserveProviderRetryClose(ctx context.Context, sessionID, turnID, oldToken, newToken string) (*ProviderRetryCheckpoint, error) {
	if oldToken == "" {
		return nil, ErrQueueConflict
	}
	return s.acquireProviderRetryCheckpoint(ctx, ActiveTurnClaim{SessionID: sessionID, TurnID: turnID, ClaimToken: oldToken}, newToken, time.Time{}, true)
}

func (s *Store) acquireProviderRetryCheckpoint(ctx context.Context, claim ActiveTurnClaim, newToken string, cutoff time.Time, closing bool) (*ProviderRetryCheckpoint, error) {
	if newToken == "" || newToken == claim.ClaimToken {
		return nil, fmt.Errorf("recovery requires a fresh claim incarnation")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var status, phase, token, updated, worker string
	if err = tx.QueryRowContext(ctx, `select t.status,t.phase,a.claim_token,a.updated_at,coalesce(a.worker_id,'') from turns t join session_active_turns a on a.turn_id=t.id and a.session_id=t.session_id where t.id=? and t.session_id=?`, claim.TurnID, claim.SessionID).Scan(&status, &phase, &token, &updated, &worker); err != nil {
		return nil, err
	}
	closed := status == "queued" && IsClosedProviderRetryPhase(phase) && worker == "closed-checkpoint"
	if closing && closed {
		return nil, nil
	}
	if !closed && (status != "running" || (phase != "retry_wait" && phase != "running")) {
		return nil, nil
	}
	boundary := "inference.retry_scheduled"
	resumed := phase == "running" || phase == ClosedProviderRetryReadyPhase
	if resumed {
		boundary = "inference.retry_resumed"
	}
	stamp, err := time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return nil, err
	}
	if token != claim.ClaimToken || (!closing && (updated != claim.UpdatedAt || (!closed && !stamp.Before(cutoff)))) {
		return nil, ErrQueueConflict
	}
	checkpoint, err := providerRetryCheckpointTx(ctx, tx, claim.SessionID, claim.TurnID, boundary, resumed)
	if closing && errors.Is(err, ErrQueueConflict) {
		return nil, nil
	} // a later effect is not a retry checkpoint
	if err != nil || checkpoint == nil {
		return nil, err
	}
	event, reason, newWorker := "turn.recovery_admitted", "resume_provider_retry", "retry-recovery"
	if closing {
		event, reason, newWorker = "turn.checkpoint_close_started", "graceful_close", "retry-closing"
	}
	result, err := tx.ExecContext(ctx, `update session_active_turns set claim_token=?,worker_id=?,claimed_at=`+defaultNow+`,updated_at=`+defaultNow+` where session_id=? and turn_id=? and claim_token=? and updated_at=?`, newToken, newWorker, claim.SessionID, claim.TurnID, claim.ClaimToken, updated)
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
	if closed {
		nextPhase := "retry_wait"
		if resumed {
			nextPhase = "running"
		}
		if _, err = tx.ExecContext(ctx, `update turns set status='running',phase=?,updated_at=`+defaultNow+` where id=? and session_id=?`, nextPhase, claim.TurnID, claim.SessionID); err != nil {
			return nil, err
		}
		if err = retryCheckpointSessionStateTx(ctx, tx, claim.SessionID, "running", claim.TurnID); err != nil {
			return nil, err
		}
	}
	if err = receiptEventTx(ctx, tx, claim.TurnID, claim.SessionID, event, map[string]any{"phase": "recovery", "checkpoint": true, "reason": reason, "recovery_disposition": "resume_provider_retry", "previous_phase": phase, "graceful": closing || closed}); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return checkpoint, nil
}

// Shared eligibility reader: callers choose ownership/adoption rules, but cannot
// infer a continuation from phase alone or from a superseded receipt.
func providerRetryCheckpointTx(ctx context.Context, tx *sql.Tx, sessionID, turnID, boundary string, resumed bool) (*ProviderRetryCheckpoint, error) {
	var occurrence string
	err := tx.QueryRowContext(ctx, `select json_extract(payload_json,'$.request_id') from turn_events where turn_id=? and session_id=? and event_type=? order by seq desc limit 1`, turnID, sessionID, boundary).Scan(&occurrence)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err = validateProviderRetryBoundaryTx(ctx, tx, turnID, occurrence, boundary); err != nil {
		return nil, err
	}
	var raw string
	if err = tx.QueryRowContext(ctx, `select decision_json from turn_receipts where turn_id=? and occurrence_id=? and kind='provider'`, turnID, occurrence).Scan(&raw); err != nil {
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
		return nil, nil
	}
	if decision.Next != "retry_wait" || decision.Wait == nil {
		return nil, nil
	}
	if err = decision.Wait.validate(); err != nil {
		return nil, nil
	}
	if resumed && time.Now().Before(decision.Wait.RetryAt) {
		return nil, nil
	}
	if decision.Wait.Iteration < 1 || decision.Wait.Iteration != decision.Iteration || decision.Wait.Lifecycle == nil {
		return nil, nil
	}
	var held bool
	if err = tx.QueryRowContext(ctx, `select exists(select 1 from turn_failures where turn_id=? and (hold_state<>'none' or coalesce(resolution_state,'')<>''))`, turnID).Scan(&held); err != nil {
		return nil, err
	}
	if held {
		return nil, ErrQueueConflict
	}
	return &ProviderRetryCheckpoint{OccurrenceID: occurrence, Wait: decision.Wait, Resumed: resumed}, nil
}
