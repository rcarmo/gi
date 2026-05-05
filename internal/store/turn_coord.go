package store

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *Store) ClaimSessionActiveTurn(ctx context.Context, sessionID, turnID, workerID, claimToken string) (bool, error) {
	res, err := s.db.ExecContext(ctx, `
		insert into session_active_turns (session_id, turn_id, worker_id, claim_token, claimed_at, updated_at)
		values (?, ?, ?, ?, `+defaultNow+`, `+defaultNow+`)
		on conflict(session_id) do nothing
	`, sessionID, turnID, nilIfEmpty(workerID), claimToken)
	if err != nil {
		return false, fmt.Errorf("claim session active turn: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("claim session active turn rows: %w", err)
	}
	return rows > 0, nil
}

func (s *Store) ReleaseSessionActiveTurn(ctx context.Context, sessionID, claimToken string) error {
	if claimToken == "" {
		_, err := s.db.ExecContext(ctx, `delete from session_active_turns where session_id = ?`, sessionID)
		if err != nil {
			return fmt.Errorf("release session active turn: %w", err)
		}
		return nil
	}
	_, err := s.db.ExecContext(ctx, `delete from session_active_turns where session_id = ? and claim_token = ?`, sessionID, claimToken)
	if err != nil {
		return fmt.Errorf("release session active turn: %w", err)
	}
	return nil
}

func (s *Store) GetSessionActiveTurn(ctx context.Context, sessionID string) (turnID string, claimToken string, err error) {
	row := s.db.QueryRowContext(ctx, `select turn_id, claim_token from session_active_turns where session_id = ?`, sessionID)
	if err := row.Scan(&turnID, &claimToken); err != nil {
		return "", "", err
	}
	return turnID, claimToken, nil
}

func (s *Store) EnqueueSteering(ctx context.Context, sessionID, turnID, role, content string, payload map[string]any, media []string, queueMode string) (int64, error) {
	payloadJSON, err := marshalJSON(payload)
	if err != nil {
		return 0, err
	}
	mediaJSON, err := marshalJSONArray(media)
	if err != nil {
		return 0, err
	}
	if queueMode == "" {
		queueMode = "one-at-a-time"
	}
	res, err := s.db.ExecContext(ctx, `
		insert into steering_queue (session_id, turn_id, role, content, payload_json, media_json, queue_mode, status, created_at, updated_at)
		values (?, ?, ?, ?, ?, ?, ?, 'queued', `+defaultNow+`, `+defaultNow+`)
	`, sessionID, nilIfEmpty(turnID), role, content, payloadJSON, mediaJSON, queueMode)
	if err != nil {
		return 0, fmt.Errorf("enqueue steering: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("enqueue steering id: %w", err)
	}
	return id, nil
}

func (s *Store) SteeringQueueLength(ctx context.Context, sessionID string) (int, error) {
	row := s.db.QueryRowContext(ctx, `select count(*) from steering_queue where session_id = ? and status = 'queued'`, sessionID)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("steering queue length: %w", err)
	}
	return count, nil
}

func isNotFound(err error) bool { return err == sql.ErrNoRows }
