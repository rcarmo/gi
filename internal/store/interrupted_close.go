package store

import (
	"context"
	"database/sql"
	"errors"
)

// InterruptedClosingWorker identifies a shutdown with no proven continuation.
// A successor must hold it even if the last recorded phase normally requeues.
const InterruptedClosingWorker = "interruption-closing"

// ReserveInterruptedClose fences ordinary writes while retaining ownership for
// bounded receipt progress. It never authorizes replay of an unproven frontier.
func (s *Store) ReserveInterruptedClose(ctx context.Context, sessionID, turnID, oldToken, token string) (bool, error) {
	ctx = CoordinationContext(ctx, nil)
	if oldToken == "" || token == "" || oldToken == token {
		return false, ErrQueueConflict
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `update turns set updated_at=updated_at where id=? and session_id=?`, turnID, sessionID); err != nil {
		return false, err
	}
	var status, phase, owner string
	err = tx.QueryRowContext(ctx, `select t.status,t.phase,a.claim_token from turns t join session_active_turns a on a.turn_id=t.id and a.session_id=t.session_id where t.id=? and t.session_id=?`, turnID, sessionID).Scan(&status, &phase, &owner)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if owner != oldToken {
		return false, ErrQueueConflict
	}
	if status != "running" && status != "queued" {
		return false, nil
	}
	// Let a previously staged terminal decision finish; shutdown must not
	// replace a completed result or an explicit user cancellation with a hold.
	var staged bool
	if err = tx.QueryRowContext(ctx, `select exists(select 1 from turn_events where turn_id=? and event_type='turn.terminal_staged')`, turnID).Scan(&staged); err != nil {
		return false, err
	}
	if staged {
		return false, nil
	}
	if _, err = tx.ExecContext(ctx, `update session_active_turns set claim_token=?,worker_id=?,updated_at=`+defaultNow+` where session_id=? and turn_id=? and claim_token=?`, token, InterruptedClosingWorker, sessionID, turnID, oldToken); err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, `insert into turn_recovery_fences(turn_id,session_id,claim_token) values(?,?,?) on conflict(turn_id) do update set session_id=excluded.session_id,claim_token=excluded.claim_token`, turnID, sessionID, token); err != nil {
		return false, err
	}
	if err = receiptEventTx(ctx, tx, turnID, sessionID, "turn.interrupted_close_reserved", map[string]any{"phase": phase, "checkpoint": true, "replay_authorized": false}); err != nil {
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

// CommitInterruptedClose parks an unproven continuation after its owned hosts
// and closing heartbeat have joined. No root terminal receipt is created.
func (s *Store) CommitInterruptedClose(ctx context.Context, sessionID, turnID, token string, failure TurnFailure) error {
	ctx = CoordinationContext(ctx, nil)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `update turns set updated_at=updated_at where id=? and session_id=?`, turnID, sessionID); err != nil {
		return err
	}
	var owner, worker string
	if err = tx.QueryRowContext(ctx, `select claim_token,worker_id from session_active_turns where session_id=? and turn_id=?`, sessionID, turnID).Scan(&owner, &worker); err != nil {
		return err
	}
	if owner != token || worker != InterruptedClosingWorker {
		return ErrQueueConflict
	}
	if err = sealClosingProviderProgressTx(WithTurnWriter(ctx, sessionID, turnID, token), tx, sessionID, turnID); err != nil {
		return err
	}
	if err = upsertTurnFailure(ctx, tx, turnID, sessionID, failure.FailureKind, "review", failure.Summary); err != nil {
		return err
	}
	if err = receiptEventTx(ctx, tx, turnID, sessionID, "turn.failure_marked", map[string]any{"phase": "recovery", "checkpoint": true, "failure_kind": failure.FailureKind, "hold_state": "review", "summary": failure.Summary}); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `update turns set status='failed',phase='held_for_retry_or_skip',finished_at=coalesce(finished_at,`+defaultNow+`),updated_at=`+defaultNow+` where id=? and session_id=?`, turnID, sessionID); err != nil {
		return err
	}
	if err = receiptEventTx(ctx, tx, turnID, sessionID, "turn.recovered", map[string]any{"phase": "held_for_retry_or_skip", "recovered": true, "held_for_retry": true, "source": "graceful_close", "code": "interrupted_by_close", "message": failure.Summary}); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `delete from session_active_turns where session_id=? and turn_id=? and claim_token=?`, sessionID, turnID, token); err != nil {
		return err
	}
	if err = retryCheckpointSessionStateTx(ctx, tx, sessionID, "idle", ""); err != nil {
		return err
	}
	return tx.Commit()
}

// The shared join barrier has ended these physical hosts. Seal captured progress
// even if Close won the writer fence just before normal receipt settlement.
func sealClosingProviderProgressTx(ctx context.Context, tx *sql.Tx, sessionID, turnID string) error {
	rows, err := tx.QueryContext(ctx, `select distinct json_extract(e.payload_json,'$.request_id'),coalesce(json_extract(e.payload_json,'$.iteration'),0) from turn_events e
where e.turn_id=? and e.session_id=? and e.event_type='inference.started'
and not exists(select 1 from turn_receipts r where r.turn_id=e.turn_id and r.kind='provider' and r.occurrence_id=json_extract(e.payload_json,'$.request_id'))
and exists(select 1 from messages m where m.session_id=e.session_id and json_extract(m.payload_json,'$.turn_id')=e.turn_id and json_extract(m.payload_json,'$.request_id')=json_extract(e.payload_json,'$.request_id') and json_extract(m.payload_json,'$.partial')=1)`, turnID, sessionID)
	if err != nil {
		return err
	}
	defer rows.Close()
	type pending struct {
		id        string
		iteration int
	}
	var attempts []pending
	for rows.Next() {
		var p pending
		if err = rows.Scan(&p.id, &p.iteration); err != nil {
			return err
		}
		attempts = append(attempts, p)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}
	for _, p := range attempts {
		if err = interruptProviderTx(ctx, tx, sessionID, turnID, p.id, p.iteration); err != nil {
			return err
		}
	}
	return nil
}
