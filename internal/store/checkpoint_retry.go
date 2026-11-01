package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

const ClosedToolBatchPhase = "checkpoint_tool_batch"

// checkpointRetryControl pairs a recovery hold with explicit admission.
// A skip marker permits validation for terminal cleanup; its resolved failure
// still blocks ordinary checkpoint readers and does not reset the retry budget.
// An unmatched hold must never make a frontier eligible for automatic recovery.
func checkpointRetryControl(kind, raw, phase string, parked *int) (bool, bool) {
	if kind != "turn.failure_marked" && kind != "turn.recovered" && kind != "turn.checkpoint_retry_requested" && kind != "turn.checkpoint_skip_requested" {
		return false, true
	}
	var h struct {
		Phase          string `json:"phase"`
		PreviousPhase  string `json:"previous_phase"`
		PreviousStatus string `json:"previous_status"`
		Disposition    string `json:"recovery_disposition"`
		Checkpoint     bool   `json:"checkpoint"`
		Stale          bool   `json:"stale_claim"`
		Reason         string `json:"reason"`
		Token          string `json:"admission_token"`
		FailureKind    string `json:"failure_kind"`
		HoldState      string `json:"hold_state"`
	}
	if json.Unmarshal([]byte(raw), &h) != nil {
		return true, false
	}
	if kind == "turn.failure_marked" {
		expected := "recovery_interrupted"
		if phase == "waiting_on_tools" {
			expected = "recovery_interrupted_tool_phase"
		}
		if *parked != 0 || h.Phase != "recovery" || !h.Checkpoint || h.HoldState != "review" || h.FailureKind != expected {
			return true, false
		}
		*parked = 1
	} else if kind == "turn.recovered" {
		expected := "hold_for_retry_or_skip_after_interruption"
		validPhase := h.PreviousPhase == "generation_ready" || h.PreviousPhase == "running"
		if phase == "waiting_on_tools" {
			expected = "hold_for_retry_or_skip_after_tool_checkpoint"
			validPhase = h.PreviousPhase == phase
		}
		if *parked != 1 || h.Phase != "recovery" || !h.Checkpoint || !h.Stale || h.PreviousStatus != "running" || !validPhase || h.Disposition != expected {
			return true, false
		}
		*parked = 2
	} else {
		reason := "explicit_retry"
		if kind == "turn.checkpoint_skip_requested" {
			reason = "explicit_skip"
		}
		if *parked != 2 || h.Phase != phase || !h.Checkpoint || h.Reason != reason || h.Token == "" {
			return true, false
		}
		*parked = 0
	}
	return true, true
}

// QueueHeldCheckpointRetry reopens only a proven native frontier, never its
// prompt. Clearing the hold, fencing, budget reset and checkpoint queue admission
// commit together. Unsupported native frontiers remain held, not prompt-retried.
func (s *Store) QueueHeldCheckpointRetry(ctx context.Context, turnID, token, summary string) (*ActiveTurnClaim, bool, error) {
	claim, _, handled, err := s.admitHeldCheckpoint(ctx, turnID, token, summary, false)
	return claim, handled, err
}

// StageHeldCheckpointSkip acknowledges an abandoned native invocation without
// retrying its prompt or changing its failed outcome into cancellation.
func (s *Store) StageHeldCheckpointSkip(ctx context.Context, turnID, token, summary string) (*TurnSettlement, bool, error) {
	_, decision, handled, err := s.admitHeldCheckpoint(ctx, turnID, token, summary, true)
	return decision, handled, err
}

func (s *Store) admitHeldCheckpoint(ctx context.Context, turnID, token, summary string, skip bool) (claim *ActiveTurnClaim, decision *TurnSettlement, handled bool, err error) {
	if token == "" {
		return nil, nil, true, ErrQueueConflict
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, false, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `update turns set updated_at=updated_at where id=?`, turnID); err != nil {
		return nil, nil, false, err
	}
	var session, status, phase, parent string
	if err = tx.QueryRowContext(ctx, `select session_id,status,phase,coalesce(nullif(json_extract(metadata_json,'$.parent_turn_id'),''),(select parent_turn_id from subturns where child_turn_id=turns.id),'') from turns where id=?`, turnID).Scan(&session, &status, &phase, &parent); err != nil {
		return nil, nil, false, err
	}
	var native bool
	if err = tx.QueryRowContext(ctx, `select exists(select 1 from turn_events where turn_id=? and event_type in ('generation.ready','tool.batch_ready'))`, turnID).Scan(&native); err != nil || !native {
		return nil, nil, false, err
	}
	var failure, hold, resolution string
	err = tx.QueryRowContext(ctx, `select failure_kind,hold_state,coalesce(resolution_state,'') from turn_failures where turn_id=?`, turnID).Scan(&failure, &hold, &resolution)
	if err == sql.ErrNoRows {
		// A repeated request returns the original admission, including after restart.
		var admitted bool
		if err = tx.QueryRowContext(ctx, `select exists(select 1 from turn_events where turn_id=? and event_type='turn.checkpoint_retry_requested')`, turnID).Scan(&admitted); err != nil {
			return nil, nil, true, err
		}
		if admitted {
			return nil, nil, true, nil
		}
		return nil, nil, false, nil
	}
	if err != nil {
		return nil, nil, true, err
	}
	if resolution != "" {
		return nil, nil, false, nil
	}
	if hold == "none" {
		return nil, nil, false, nil
	}
	if skip {
		var finished bool
		if err = tx.QueryRowContext(ctx, turnTerminalReceiptQuery, turnID).Scan(&finished); err != nil {
			return nil, nil, true, err
		}
		if finished {
			return nil, nil, false, nil
		}
	}
	if status != "failed" || phase != "held_for_retry_or_skip" || (failure != "recovery_interrupted" && failure != "recovery_interrupted_tool_phase") {
		return nil, nil, true, ErrFailureConflict
	}
	// A consumed child terminal cannot be revived behind its parent's back.
	if parent != "" && !skip {
		return nil, nil, true, ErrFailureConflict
	}
	var terminal bool
	if err = tx.QueryRowContext(ctx, `select exists(select 1 from turn_events where turn_id=? and event_type in ('turn.finished','turn.cancelled','turn.terminal_staged'))`, turnID).Scan(&terminal); err != nil {
		return nil, nil, true, err
	}
	if terminal {
		return nil, nil, true, ErrFailureConflict
	}
	var previousPhase string
	if err = tx.QueryRowContext(ctx, `select json_extract(payload_json,'$.previous_phase') from turn_events where turn_id=? and session_id=? and event_type='turn.recovered' order by seq desc limit 1`, turnID, session).Scan(&previousPhase); err != nil {
		if err == sql.ErrNoRows {
			err = ErrFailureConflict
		}
		return nil, nil, true, err
	}
	frontier, queuedPhase := "generation_ready", ClosedGenerationReadyPhase
	if previousPhase == "waiting_on_tools" {
		frontier, queuedPhase = "waiting_on_tools", ClosedToolBatchPhase
	}
	if skip {
		_, err = tx.ExecContext(ctx, `update turn_failures set hold_state='none' where turn_id=?`, turnID)
	} else {
		_, err = tx.ExecContext(ctx, `delete from turn_failures where turn_id=?`, turnID)
	}
	if err != nil {
		return nil, nil, true, err
	}
	event, reason := "turn.checkpoint_retry_requested", "explicit_retry"
	if skip {
		event, reason = "turn.checkpoint_skip_requested", "explicit_skip"
	}
	if err = receiptEventTx(ctx, tx, turnID, session, event, map[string]any{"phase": frontier, "checkpoint": true, "reason": reason, "admission_token": token, "summary": summary}); err != nil {
		return nil, nil, true, err
	}
	// Settlement of a pending provider belongs to this original invocation.
	// The intermediate running state is invisible until the queued admission
	// commits; any validation or queue conflict rolls it back with the hold.
	if _, err = tx.ExecContext(ctx, `update turns set status='running',phase=? where id=? and session_id=?`, previousPhase, turnID, session); err != nil {
		return nil, nil, true, err
	}
	if frontier == "waiting_on_tools" {
		checkpoint, readErr := toolBatchCheckpointTx(ctx, tx, session, turnID)
		if readErr != nil {
			return nil, nil, true, readErr
		}
		if checkpoint == nil {
			return nil, nil, true, ErrFailureConflict
		}
	} else {
		checkpoint, requestID, readErr := generationCheckpointTx(ctx, tx, session, turnID, previousPhase)
		if readErr != nil {
			return nil, nil, true, readErr
		}
		if checkpoint == nil {
			return nil, nil, true, ErrFailureConflict
		}
		if requestID != "" {
			if _, err = tx.ExecContext(ctx, `insert into turn_recovery_fences(turn_id,session_id,claim_token) values(?,?,?) on conflict(turn_id) do update set session_id=excluded.session_id,claim_token=excluded.claim_token`, turnID, session, token); err != nil {
				return nil, nil, true, err
			}
			if err = interruptProviderTx(WithTurnWriter(ctx, session, turnID, token), tx, session, turnID, requestID, checkpoint.Iteration); err != nil {
				return nil, nil, true, err
			}
		}
		if err = checkGenerationEffectsSettledTx(ctx, tx, turnID); err != nil {
			return nil, nil, true, err
		}
		if !skip {
			if err = receiptEventTx(ctx, tx, turnID, session, "generation.ready", map[string]any{"continuation": checkpoint, "explicit_retry": true}); err != nil {
				return nil, nil, true, err
			}
		}
	}
	if skip {
		decision = &TurnSettlement{Status: "failed", Phase: "failed", MessageID: NowID("msg"), Message: "Held checkpoint skipped. Verify interrupted effects before repeating them.", MessagePayload: map[string]any{"kind": "chat", "source": "system", "turn_id": turnID}, Payload: map[string]any{"checkpoint_skip": queuedPhase, "resolution_summary": summary}, ClaimToken: token}
		if _, err = tx.ExecContext(ctx, `update turn_failures set resolution_state='skipped',resolution_summary=?,resolved_at=`+defaultNow+`,updated_at=`+defaultNow+` where turn_id=?`, summary, turnID); err != nil {
			return nil, nil, true, err
		}
		if err = receiptEventTx(ctx, tx, turnID, session, "turn.terminal_staged", map[string]any{"phase": "settling", "checkpoint": true, "status": "failed", "previous_phase": previousPhase, "decision": map[string]any{"status": decision.Status, "phase": decision.Phase, "message_id": decision.MessageID, "message": decision.Message, "message_payload": decision.MessagePayload, "payload": decision.Payload}}); err != nil {
			return nil, nil, true, err
		}
	}
	worker, active := "closed-checkpoint", ""
	if skip {
		worker, active = "terminal-skip", turnID
	}
	result, err := tx.ExecContext(ctx, `insert into session_active_turns(session_id,turn_id,worker_id,claim_token,claimed_at,updated_at) values(?,?,?,?,`+defaultNow+`,`+defaultNow+`) on conflict(session_id) do nothing`, session, turnID, worker, token)
	if err != nil {
		return nil, nil, true, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, nil, true, err
	}
	if n != 1 {
		return nil, nil, true, ErrQueueConflict
	}
	if _, err = tx.ExecContext(ctx, `insert into turn_recovery_fences(turn_id,session_id,claim_token) values(?,?,?) on conflict(turn_id) do update set session_id=excluded.session_id,claim_token=excluded.claim_token`, turnID, session, token); err != nil {
		return nil, nil, true, err
	}
	status = "queued"
	if skip {
		status, queuedPhase = "running", "settling"
	}
	if _, err = tx.ExecContext(ctx, `update turns set status=?,phase=?,finished_at=null,updated_at=`+defaultNow+` where id=? and session_id=?`, status, queuedPhase, turnID, session); err != nil {
		return nil, nil, true, err
	}
	if err = retryCheckpointSessionStateTx(ctx, tx, session, status, active); err != nil {
		return nil, nil, true, err
	}
	claim = &ActiveTurnClaim{SessionID: session, TurnID: turnID, WorkerID: worker, ClaimToken: token, Status: status, Phase: queuedPhase}
	if err = tx.QueryRowContext(ctx, `select claimed_at,updated_at from session_active_turns where session_id=?`, session).Scan(&claim.ClaimedAt, &claim.UpdatedAt); err != nil {
		return nil, nil, true, fmt.Errorf("checkpoint retry claim: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, nil, true, err
	}
	return claim, decision, true, nil
}
