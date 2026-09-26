package store

import (
	"context"
	"database/sql"
)

// StopWebActiveTurn fences queued web work behind an explicit resume. The stop
// is admitted only for the current claimed active turn and creates a durable
// hold row when queued work still exists for the session.
func (s *Store) StopWebActiveTurn(ctx context.Context, sessionID, turnID, claimToken string) error {
	if sessionID == "" || turnID == "" || claimToken == "" {
		return ErrQueueConflict
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `update turns
		set status='cancelling', phase='cancelling', updated_at=`+defaultNow+`
		where id=? and session_id=? and status in ('running','cancelling')
		and exists(
			select 1 from session_active_turns a
			where a.session_id=? and a.turn_id=? and a.claim_token=?
		)
		and not exists(
			select 1 from web_queue_holds h
			where h.session_id=? and h.stop_turn_id<>?
		)`, turnID, sessionID, sessionID, turnID, claimToken, sessionID, turnID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrQueueConflict
	}
	var shouldHold bool
	if err := tx.QueryRowContext(ctx, `select case
		when exists(select 1 from turns where session_id=? and (status='steering' or (status='queued' and phase<>'steer_returned' and coalesce(json_extract(metadata_json,'$.operation'),'')<>'manual_compaction'))) 
			or exists(select 1 from steering_queue where session_id=? and status in ('queued','claimed'))
		then 1 else 0 end`, sessionID, sessionID).Scan(&shouldHold); err != nil {
		return err
	}
	if shouldHold {
		if _, err := tx.ExecContext(ctx, `insert into web_queue_holds(session_id, stop_turn_id, created_at)
			values(?,?,`+defaultNow+`)
			on conflict(session_id) do nothing`, sessionID, turnID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) WebQueueHold(ctx context.Context, sessionID string) (string, error) {
	row := s.db.QueryRowContext(ctx, `select stop_turn_id from web_queue_holds where session_id=?`, sessionID)
	var stopTurnID string
	if err := row.Scan(&stopTurnID); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return stopTurnID, nil
}

// ResumeWebQueue retires the exact hold only after a new claim is durable, or
// after every pending item has been removed. Pre-launch errors leave it intact.
func (s *Store) ResumeWebQueue(ctx context.Context, sessionID, stopTurnID string) error {
	if sessionID == "" || stopTurnID == "" {
		return ErrQueueConflict
	}
	result, err := s.db.ExecContext(ctx, `delete from web_queue_holds
		where session_id=? and stop_turn_id=?
		and (exists(select 1 from session_active_turns where session_id=? and turn_id<>?)
 or (not exists(select 1 from session_active_turns where session_id=?)
 and not exists(select 1 from turns where session_id=? and (status='steering' or (status='queued' and phase<>'steer_returned' and coalesce(json_extract(metadata_json,'$.operation'),'')<>'manual_compaction'))) 
 and not exists(select 1 from steering_queue where session_id=? and status in ('queued','claimed'))))`, sessionID, stopTurnID, sessionID, stopTurnID, sessionID, sessionID, sessionID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrQueueConflict
	}
	return nil
}

// ClaimWebResumedTurn bypasses only the captured hold, without deleting it.
func (s *Store) ClaimWebResumedTurn(ctx context.Context, sessionID, turnID, workerID, claimToken, stopTurnID string) (bool, error) {
	if stopTurnID == "" {
		return false, ErrQueueConflict
	}
	res, err := s.db.ExecContext(ctx, `insert into session_active_turns(session_id,turn_id,worker_id,claim_token,claimed_at,updated_at)
 select ?,?,?,?,`+defaultNow+`,`+defaultNow+`
 where exists(select 1 from web_queue_holds where session_id=? and stop_turn_id=?)
 and exists(select 1 from turns where id=? and session_id=? and status='queued' and phase<>'steer_returned' and coalesce(json_extract(metadata_json,'$.operation'),'')<>'manual_compaction')
 on conflict(session_id) do nothing`, sessionID, turnID, workerID, claimToken, sessionID, stopTurnID, turnID, sessionID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 1, err
}
