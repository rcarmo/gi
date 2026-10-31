package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ProviderRetryWait is a committed successor of one failed provider receipt,
// not permission to resend its request. A successor must prepare a fresh intent.
type ProviderRetryWait struct {
	Attempt         int       `json:"attempt"`
	MaxAttempts     int       `json:"max_attempts"`
	DelayMS         int64     `json:"delay_ms"`
	RetryAt         time.Time `json:"retry_at"`
	FailureCategory string    `json:"failure_category"`
	Iteration       int       `json:"iteration"`
	Model           string    `json:"model"`
	Title           string    `json:"title"`
}

func (p *ProviderRetryWait) validate() error {
	if p == nil {
		return nil
	}
	if p.Attempt < 1 || p.Attempt > 100 || p.MaxAttempts < p.Attempt || p.MaxAttempts > 100 || p.DelayMS < 0 || p.RetryAt.IsZero() || p.Iteration < 0 || p.Model == "" {
		return fmt.Errorf("invalid provider retry checkpoint")
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
