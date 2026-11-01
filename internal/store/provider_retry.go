package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// RepeatedToolFailureLimit bounds the loop state carried across checkpoints.
const RepeatedToolFailureLimit = 4

// ProviderRetryWait is a committed successor of one failed provider receipt,
// not permission to resend its request. A successor must prepare a fresh intent.
type ProviderRetryWait struct {
	Attempt                  int                      `json:"attempt"`
	MaxAttempts              int                      `json:"max_attempts"`
	DelayMS                  int64                    `json:"delay_ms"`
	RetryAt                  time.Time                `json:"retry_at"`
	FailureCategory          string                   `json:"failure_category"`
	Iteration                int                      `json:"iteration"`
	Model                    string                   `json:"model"`
	Title                    string                   `json:"title"`
	Lifecycle                *ProviderLifecyclePolicy `json:"lifecycle,omitempty"`
	LastToolFailureSignature string                   `json:"last_tool_failure_signature,omitempty"`
	RepeatedToolFailures     int                      `json:"repeated_tool_failures,omitempty"`
}

func (p *ProviderRetryWait) validate() error {
	if p == nil {
		return nil
	}
	if p.Attempt < 1 || p.Attempt > 100 || p.MaxAttempts < p.Attempt || p.MaxAttempts > 100 || p.DelayMS < 0 || p.RetryAt.IsZero() || p.Iteration < 0 || p.Model == "" {
		return fmt.Errorf("invalid provider retry checkpoint")
	}
	if p.RepeatedToolFailures < 0 || p.RepeatedToolFailures >= RepeatedToolFailureLimit || (p.RepeatedToolFailures > 0 && p.LastToolFailureSignature == "") || (p.RepeatedToolFailures == 0 && p.LastToolFailureSignature != "") {
		return fmt.Errorf("invalid provider retry tool-failure budget")
	}
	if p.Lifecycle != nil {
		if err := p.Lifecycle.Validate(); err != nil {
			return err
		}
	}
	switch p.FailureCategory {
	case "network", "timeout", "rate_limit":
	default:
		return fmt.Errorf("invalid provider retry category")
	}
	return nil
}

// ResumeProviderRetry advances only the committed wait after its deadline.
// Phase and wake event commit together; this does not dispatch a provider.
func (s *Store) ResumeProviderRetry(ctx context.Context, sessionID, turnID, occurrenceID string) error {
	if err := ctx.Err(); err != nil {
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
	var raw, phase, status, owner string
	err = tx.QueryRowContext(ctx, `select r.decision_json,t.phase,t.status,t.session_id from turn_receipts r join turns t on t.id=r.turn_id where r.turn_id=? and r.occurrence_id=? and r.kind='provider'`, turnID, occurrenceID).Scan(&raw, &phase, &status, &owner)
	if err != nil {
		return err
	}
	if owner != sessionID || phase != "retry_wait" || status != "running" {
		return ErrQueueConflict
	}
	var decision struct {
		Next      string             `json:"next"`
		RetryWait *ProviderRetryWait `json:"retry_wait"`
	}
	if err = json.Unmarshal([]byte(raw), &decision); err != nil {
		return err
	}
	if decision.Next != "retry_wait" || decision.RetryWait == nil {
		return ErrQueueConflict
	}
	if err = decision.RetryWait.validate(); err != nil {
		return err
	}
	// A turn can wait more than once. A valid old receipt must not wake a
	// newer delay or outrun a subsequent provider/tool/terminal boundary.
	if err = validateProviderRetryFrontierTx(ctx, tx, turnID, occurrenceID); err != nil {
		return err
	}
	if time.Now().Before(decision.RetryWait.RetryAt) {
		return ErrQueueConflict
	}
	result, err := tx.ExecContext(ctx, `update turns set phase='running',updated_at=`+defaultNow+` where id=? and exists(select 1 from session_active_turns where session_id=? and turn_id=?)`, turnID, sessionID, turnID)
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
	if err = receiptEventTx(ctx, tx, turnID, sessionID, "inference.retry_resumed", map[string]any{"request_id": occurrenceID, "iteration": decision.RetryWait.Iteration, "phase": "inference", "checkpoint": true}); err != nil {
		return err
	}
	return tx.Commit()
}

func (p *ProviderRetryWait) Payload() map[string]any {
	return map[string]any{"attempt": p.Attempt, "max_attempts": p.MaxAttempts, "delay_ms": p.DelayMS, "retry_at": p.RetryAt.UTC().Format(time.RFC3339Nano), "failure_category": p.FailureCategory, "iteration": p.Iteration, "model": p.Model, "title": p.Title, "phase": "retry_wait", "checkpoint": true}
}

func validateProviderRetryFrontierTx(ctx context.Context, tx *sql.Tx, turnID, occurrenceID string) error {
	var current bool
	err := tx.QueryRowContext(ctx, `select exists(
 select 1 from turn_events scheduled
 where scheduled.turn_id=? and scheduled.event_type='inference.retry_scheduled'
 and json_extract(scheduled.payload_json,'$.request_id')=?
 and not exists(select 1 from turn_events later
 where later.turn_id=scheduled.turn_id and later.seq>scheduled.seq
 and later.event_type in ('inference.started','inference.request_prepared','inference.settled','inference.retry_scheduled','inference.retry_resumed','tool.started','tool.finished','tool.failed','tool.cancelled','tool.aborted','tool.skipped','tool.output','turn.terminal_staged','turn.finished'))
 )`, turnID, occurrenceID).Scan(&current)
	if err != nil {
		return err
	}
	if !current {
		return ErrQueueConflict
	}
	return nil
}

// CheckProviderRetryRunning is a lightweight pre-preparation check, not an
// effect admission fence. The builder still reads fresh metadata; dispatch and
// settlement retain their transactional claim/status checks.
func (s *Store) CheckProviderRetryRunning(ctx context.Context, sessionID, turnID string) error {
	var owner, status string
	if err := s.db.QueryRowContext(ctx, `select session_id,status from turns where id=?`, turnID).Scan(&owner, &status); err != nil {
		return err
	}
	if owner != sessionID || status != "running" {
		return context.Canceled
	}
	return nil
}
