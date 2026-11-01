package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const generationReadyRecoveryLimit = 5

// GenerationReadyCheckpoint follows completed preparation and settled effects.
// It retains lifecycle policy and loop budgets, never a prepared provider request.
type GenerationReadyCheckpoint struct {
	Version                  int                      `json:"version"`
	Iteration                int                      `json:"iteration"`
	MaxIterations            int                      `json:"max_iterations"`
	Model                    string                   `json:"model"`
	Lifecycle                *ProviderLifecyclePolicy `json:"lifecycle"`
	LastToolFailureSignature string                   `json:"last_tool_failure_signature,omitempty"`
	RepeatedToolFailures     int                      `json:"repeated_tool_failures,omitempty"`
}

func (p *GenerationReadyCheckpoint) Validate() error {
	if p == nil || p.Version != 1 || p.Iteration < 1 || p.MaxIterations < p.Iteration || strings.TrimSpace(p.Model) == "" || p.RepeatedToolFailures < 0 || p.RepeatedToolFailures >= RepeatedToolFailureLimit || (p.RepeatedToolFailures == 0) != (p.LastToolFailureSignature == "") {
		return fmt.Errorf("invalid generation-ready continuation")
	}
	return p.Lifecycle.Validate()
}

func (s *Store) SaveGenerationReady(ctx context.Context, sessionID, turnID string, checkpoint *GenerationReadyCheckpoint) error {
	if err := checkpoint.Validate(); err != nil {
		return err
	}
	return s.AdmitTurnEffect(ctx, sessionID, turnID, "generation_ready", "generation.ready", map[string]any{"continuation": checkpoint})
}

// Unsettled effects must never be hidden behind a later ready checkpoint. Prior
// ready frontiers already verified their prefix; inspect only the new suffix.
func checkGenerationEffectsSettledTx(ctx context.Context, tx *sql.Tx, turnID string) error {
	var unresolved int
	err := tx.QueryRowContext(ctx, `select count(*) from turn_events e
 where e.turn_id=? and e.seq>coalesce((select max(seq) from turn_events where turn_id=? and event_type='generation.ready'),0)
 and ((e.event_type='inference.started' and not exists(select 1 from turn_receipts r where r.turn_id=e.turn_id and r.kind='provider' and r.occurrence_id=json_extract(e.payload_json,'$.request_id')))
 or (e.event_type='tool.started' and not exists(select 1 from turn_receipts r where r.turn_id=e.turn_id and r.kind='tool' and r.occurrence_id=json_extract(e.payload_json,'$.occurrence_id'))))`, turnID, turnID).Scan(&unresolved)
	if err != nil {
		return err
	}
	if unresolved != 0 {
		return fmt.Errorf("generation has unsettled effects")
	}
	return nil
}

// AcquireGenerationReadyRecovery verifies the entire frontier and fences the
// dead owner in the same transaction. A single unsettled provider admission is
// sealed as uncertain before authorising a fresh generation. Unknown events and
// legacy policies hold.
func (s *Store) AcquireGenerationReadyRecovery(ctx context.Context, claim ActiveTurnClaim, token string, olderThan time.Time) (*GenerationReadyCheckpoint, error) {
	if token == "" || token == claim.ClaimToken {
		return nil, fmt.Errorf("generation recovery requires a fresh claim token")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var current ActiveTurnClaim
	err = tx.QueryRowContext(ctx, `select sat.claim_token,sat.updated_at,t.status,t.phase,coalesce(sat.worker_id,'') from session_active_turns sat join turns t on t.id=sat.turn_id and t.session_id=sat.session_id where sat.session_id=? and sat.turn_id=?`, claim.SessionID, claim.TurnID).Scan(&current.ClaimToken, &current.UpdatedAt, &current.Status, &current.Phase, &current.WorkerID)
	if err != nil {
		return nil, err
	}
	stamp, err := time.Parse(time.RFC3339Nano, current.UpdatedAt)
	if err != nil {
		return nil, err
	}
	closed := current.Status == "queued" && current.Phase == ClosedGenerationReadyPhase && current.WorkerID == "closed-checkpoint"
	if current.ClaimToken != claim.ClaimToken || current.UpdatedAt != claim.UpdatedAt || (!closed && !stamp.Before(olderThan)) {
		return nil, ErrQueueConflict
	}
	if !closed && (current.Status != "running" || (current.Phase != "generation_ready" && current.Phase != "running")) {
		return nil, nil
	}
	phase := current.Phase
	if closed {
		phase = "generation_ready"
	}
	checkpoint, requestID, err := generationCheckpointTx(ctx, tx, claim.SessionID, claim.TurnID, phase)
	if err != nil || checkpoint == nil {
		return nil, err
	}
	// Rewriting preparation is not progress. Graceful restart is not a crash.
	var recoveries int
	if err = tx.QueryRowContext(ctx, `select count(*) from turn_events where turn_id=? and event_type in ('generation.recovered','tool.batch_recovered') and seq>coalesce((select max(seq) from turn_events where turn_id=? and event_type in ('inference.settled','tool.finished','tool.failed','tool.skipped','tool.cancelled','tool.aborted')),0)`, claim.TurnID, claim.TurnID).Scan(&recoveries); err != nil {
		return nil, err
	}
	if !closed && recoveries >= generationReadyRecoveryLimit {
		return nil, nil
	}
	result, err := tx.ExecContext(ctx, `update session_active_turns set claim_token=?,worker_id='recovery',claimed_at=`+defaultNow+`,updated_at=`+defaultNow+` where session_id=? and turn_id=? and claim_token=? and updated_at=?`, token, claim.SessionID, claim.TurnID, claim.ClaimToken, claim.UpdatedAt)
	if err != nil {
		return nil, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if count != 1 {
		return nil, ErrQueueConflict
	}
	if _, err = tx.ExecContext(ctx, `insert into turn_recovery_fences(turn_id,session_id,claim_token) values(?,?,?) on conflict(turn_id) do update set claim_token=excluded.claim_token`, claim.TurnID, claim.SessionID, token); err != nil {
		return nil, err
	}
	if requestID != "" {
		writerCtx := WithTurnWriter(ctx, claim.SessionID, claim.TurnID, token)
		if err = interruptProviderTx(writerCtx, tx, claim.SessionID, claim.TurnID, requestID, checkpoint.Iteration); err != nil {
			return nil, err
		}
		if err = checkGenerationEffectsSettledTx(ctx, tx, claim.TurnID); err != nil {
			return nil, err
		}
		if err = receiptEventTx(ctx, tx, claim.TurnID, claim.SessionID, "generation.ready", map[string]any{"continuation": checkpoint, "interrupted_request_id": requestID}); err != nil {
			return nil, err
		}
		if _, err = tx.ExecContext(ctx, `update turns set phase='generation_ready',updated_at=`+defaultNow+` where id=? and session_id=? and status='running'`, claim.TurnID, claim.SessionID); err != nil {
			return nil, err
		}
	}
	eventType := "generation.recovered"
	if closed {
		eventType = "generation.close_resumed"
		if _, err = tx.ExecContext(ctx, `update turns set status='running',phase='generation_ready',updated_at=`+defaultNow+` where id=? and session_id=?`, claim.TurnID, claim.SessionID); err != nil {
			return nil, err
		}
		if err = retryCheckpointSessionStateTx(ctx, tx, claim.SessionID, "running", claim.TurnID); err != nil {
			return nil, err
		}
	}
	if err = receiptEventTx(ctx, tx, claim.TurnID, claim.SessionID, eventType, map[string]any{"phase": "generation_ready", "checkpoint": true, "iteration": checkpoint.Iteration, "previous_claim": claim.ClaimToken, "interrupted_request_id": requestID}); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return checkpoint, nil
}

// Shared eligibility reader for recovery and Close. Never infer a checkpoint
// from phase alone, or hide a newer settled response/tool behind an old frontier.
func generationCheckpointTx(ctx context.Context, tx *sql.Tx, sessionID, turnID, phase string) (*GenerationReadyCheckpoint, string, error) {
	var seq int64
	var raw string
	err := tx.QueryRowContext(ctx, `select seq,payload_json from turn_events where turn_id=? and session_id=? and event_type='generation.ready' order by seq desc limit 1`, turnID, sessionID).Scan(&seq, &raw)
	if err == sql.ErrNoRows {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	var saved struct {
		Continuation *GenerationReadyCheckpoint `json:"continuation"`
	}
	if json.Unmarshal([]byte(raw), &saved) != nil || saved.Continuation.Validate() != nil {
		return nil, "", nil
	}
	requestID, eligible, err := generationPendingRequestTx(ctx, tx, sessionID, turnID, seq, saved.Continuation.Iteration)
	if err != nil {
		return nil, "", err
	}
	if !eligible || (phase == "running") != (requestID != "") {
		return nil, "", nil
	}
	var held int
	if err = tx.QueryRowContext(ctx, `select count(*) from turn_failures where turn_id=? and (hold_state<>'none' or coalesce(resolution_state,'')<>'')`, turnID).Scan(&held); err != nil {
		return nil, "", err
	}
	if held != 0 {
		return nil, "", nil
	}
	return saved.Continuation, requestID, nil
}
