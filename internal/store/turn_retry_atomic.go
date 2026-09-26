package store

import (
	"context"
	"database/sql"
	"fmt"
)

// HeldRetryAdmission is passed by the retry owner, never derived from arbitrary
// SubmitPrompt metadata. The opaque token is checked inside the writer transaction.
type HeldRetryAdmission struct{ OriginalTurnID, Token, Summary string }
type RetrySubTurn struct {
	ParentTurnID, ParentSessionID, DeliveryMode string
	Depth                                       int
	Metadata                                    map[string]any
}

// AdmitHeldRetry commits all durable admission state together. No queued turn
// or subturn can be observed while it is still eligible for setup rollback.
func (s *Store) AdmitHeldRetry(ctx context.Context, retry HeldRetryAdmission, id, sessionID, prompt string, metadata map[string]any, sub *RetrySubTurn) error {
	if retry.Token == "" || metadata["retry_of_turn_id"] != retry.OriginalTurnID || metadata["retry_admission_token"] != retry.Token {
		return ErrFailureConflict
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var matches int
	err = tx.QueryRowContext(ctx, `select count(*) from turn_failures f join turns t on t.id=f.turn_id where f.turn_id=? and f.session_id=? and t.session_id=f.session_id and f.resolution_state='retry_pending' and f.hold_state<>'none' and f.retry_admission_token=? and f.retry_admission_version=1 and t.status in ('failed','aborted','cancelled')`, retry.OriginalTurnID, sessionID, retry.Token).Scan(&matches)
	if err != nil {
		return err
	}
	if matches != 1 {
		return ErrFailureConflict
	}
	if err = tx.QueryRowContext(ctx, `select count(*) from turns where json_extract(metadata_json,'$.retry_of_turn_id')=? and json_extract(metadata_json,'$.retry_admission_token')=?`, retry.OriginalTurnID, retry.Token).Scan(&matches); err != nil {
		return err
	}
	if matches != 0 {
		return ErrFailureConflict
	}
	if err = insertTurn(ctx, tx, id, sessionID, "queued", prompt, metadata); err != nil {
		return err
	}
	if sub != nil {
		if err = insertSubTurn(ctx, tx, sub.ParentTurnID, sub.ParentSessionID, id, sessionID, sub.DeliveryMode, sub.Depth, sub.Metadata); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `update turn_failures set hold_state='none',resolution_state='retried',resolution_summary=?,resolved_turn_id=?,resolved_at=`+defaultNow+`,updated_at=`+defaultNow+` where turn_id=?`, retry.Summary, id, retry.OriginalTurnID); err != nil {
		return fmt.Errorf("resolve atomic retry: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `update turns set phase=case status when 'cancelled' then 'aborted' else status end,updated_at=`+defaultNow+` where id=? and phase='held_for_retry_or_skip'`, retry.OriginalTurnID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `update sessions set state_json=json_set(coalesce(nullif(state_json,''),'{}'),'$.queue_count',(select count(*) from turns where session_id=? and status='queued')),updated_at=`+defaultNow+` where id=?`, sessionID, sessionID); err != nil {
		return err
	}
	return tx.Commit()
}

// ReleaseUnadmittedHeldRetry is an explicit recovery action, not an automatic
// retry. Release and admission serialize on the same writer lock. Once released,
// an old caller cannot admit with this token. Legacy reservations are refused.
func (s *Store) ReleaseUnadmittedHeldRetry(ctx context.Context, sessionID, turnID, token string) error {
	if token == "" {
		return ErrFailureConflict
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var version int
	err = tx.QueryRowContext(ctx, `select retry_admission_version from turn_failures where turn_id=? and session_id=? and resolution_state='retry_pending' and retry_admission_token=? and hold_state<>'none'`, turnID, sessionID, token).Scan(&version)
	if err == sql.ErrNoRows {
		return ErrFailureConflict
	}
	if err != nil {
		return err
	}
	if version != 1 {
		return ErrRetryPending
	}
	var count int
	if err = tx.QueryRowContext(ctx, `select count(*) from turns where json_extract(metadata_json,'$.retry_of_turn_id')=? and json_extract(metadata_json,'$.retry_admission_token')=?`, turnID, token).Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return ErrRetryPending
	}
	if _, err = tx.ExecContext(ctx, `update turn_failures set resolution_state='',resolution_summary='',retry_admission_token='',retry_admission_version=0,updated_at=`+defaultNow+` where turn_id=?`, turnID); err != nil {
		return err
	}
	return tx.Commit()
}
