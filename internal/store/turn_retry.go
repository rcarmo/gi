package store

import (
	"context"
	"errors"
	"fmt"
)

var ErrRetryPending = errors.New("retry admission unresolved; retry rechecks stored admission without resending")
var ErrFailureConflict = errors.New("failure state changed; refresh before retry or skip")

// Reserve the existing held terminal failure atomically. The resolution turn ID
// is kept separate from the opaque token until a follow-on turn is durable.
func (s *Store) ReserveHeldRetry(ctx context.Context, turnID, token string) (bool, error) {
	return s.reserveHeldRetry(ctx, turnID, token, 0)
}

// ReserveAtomicHeldRetry opts into fenced, all-or-nothing turn admission.
func (s *Store) ReserveAtomicHeldRetry(ctx context.Context, turnID, token string) (bool, error) {
	return s.reserveHeldRetry(ctx, turnID, token, 1)
}

func (s *Store) reserveHeldRetry(ctx context.Context, turnID, token string, version int) (bool, error) {
	if token == "" {
		return false, ErrFailureConflict
	}
	result, err := s.db.ExecContext(ctx, `UPDATE turn_failures SET resolution_state='retry_pending',resolution_summary='',retry_admission_token=?,retry_admission_version=?,resolved_at=NULL,updated_at=`+defaultNow+` WHERE turn_id=? AND hold_state<>'none' AND coalesce(resolution_state,'')='' AND EXISTS(SELECT 1 FROM turns t WHERE t.id=turn_failures.turn_id AND t.session_id=turn_failures.session_id AND t.status IN ('failed','aborted','cancelled'))`, token, version, turnID)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n == 1, err
}

// ReconcileHeldRetry only observes admission after SubmitPrompt's rollback
// boundary. Absence after a restart never releases the reservation.
func (s *Store) ReconcileHeldRetry(ctx context.Context, turnID, token, summary string) (string, error) {
	return s.reconcileHeldRetry(ctx, turnID, token, summary, false, false)
}

// FinishHeldRetry is for the caller that reserved the token, only after its
// synchronous SubmitPrompt has returned. Only an error permits releasing an
// absent receipt; another/restarted caller must use ReconcileHeldRetry.
func (s *Store) FinishHeldRetry(ctx context.Context, turnID, token, summary string, submitFailed bool) (string, error) {
	return s.reconcileHeldRetry(ctx, turnID, token, summary, true, submitFailed)
}

func (s *Store) reconcileHeldRetry(ctx context.Context, turnID, token, summary string, submitReturned, releaseAbsent bool) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var session, state, current, resolved string
	var version int
	if err = tx.QueryRowContext(ctx, `SELECT session_id,coalesce(resolution_state,''),retry_admission_token,coalesce(resolved_turn_id,''),retry_admission_version FROM turn_failures WHERE turn_id=?`, turnID).Scan(&session, &state, &current, &resolved, &version); err != nil {
		return "", err
	}
	if state == "retried" && current == token {
		return resolved, nil
	}
	if state != "retry_pending" || current != token {
		return "", ErrFailureConflict
	}
	// Do not silently choose among duplicate receipts or accept a token reused
	// in another session. Corrupt/ambiguous evidence remains blocked.
	var admitted, admittedSession string
	var count int
	err = tx.QueryRowContext(ctx, `SELECT count(*), coalesce(min(id),''), coalesce(min(session_id),'') FROM turns WHERE json_extract(metadata_json,'$.retry_of_turn_id')=? AND json_extract(metadata_json,'$.retry_admission_token')=?`, turnID, token).Scan(&count, &admitted, &admittedSession)
	if err != nil {
		return "", err
	}
	if count > 1 || (count == 1 && admittedSession != session) {
		return "", ErrFailureConflict
	}
	if count == 0 {
		if !releaseAbsent {
			return "", ErrRetryPending
		}
		_, err = tx.ExecContext(ctx, `UPDATE turn_failures SET resolution_state='',resolution_summary='',retry_admission_token='',retry_admission_version=0,updated_at=`+defaultNow+` WHERE turn_id=? AND resolution_state='retry_pending' AND retry_admission_token=?`, turnID, token)
	} else {
		// New-protocol admissions always resolve in the insert transaction.
		// A pending row plus receipt cannot be inferred safe (e.g. legacy writer).
		if version != 0 {
			return "", ErrRetryPending
		}
		if !submitReturned {
			var confirmed int
			// turn.submitted is written after subturn setup/rollback completes.
			// A visible INSERT alone may still be deleted by that live caller.
			if err = tx.QueryRowContext(ctx, `select count(*) from turn_events where turn_id = ? and event_type = 'turn.submitted'`, admitted).Scan(&confirmed); err != nil {
				return "", err
			}
			if confirmed == 0 {
				return "", ErrRetryPending
			}
		}
		_, err = tx.ExecContext(ctx, `UPDATE turn_failures SET hold_state='none',resolution_state='retried',resolution_summary=?,resolved_turn_id=?,resolved_at=`+defaultNow+`,updated_at=`+defaultNow+` WHERE turn_id=? AND resolution_state='retry_pending' AND retry_admission_token=?`, summary, admitted, turnID, token)
		if err == nil {
			_, err = tx.ExecContext(ctx, `UPDATE turns SET phase=CASE status WHEN 'cancelled' THEN 'aborted' ELSE status END,updated_at=`+defaultNow+` WHERE id=? AND phase='held_for_retry_or_skip'`, turnID)
		}
	}
	if err != nil {
		return "", fmt.Errorf("reconcile retry: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return admitted, nil
}
