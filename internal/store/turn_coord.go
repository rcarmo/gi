package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ActiveTurnClaim struct {
	SessionID  string `json:"session_id"`
	TurnID     string `json:"turn_id"`
	WorkerID   string `json:"worker_id,omitempty"`
	ClaimToken string `json:"claim_token"`
	ClaimedAt  string `json:"claimed_at"`
	UpdatedAt  string `json:"updated_at"`
	Status     string `json:"status"`
	Phase      string `json:"phase"`
}

type SteeringMessage struct {
	ID        int64          `json:"id"`
	SessionID string         `json:"session_id"`
	TurnID    string         `json:"turn_id,omitempty"`
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	Payload   map[string]any `json:"payload,omitempty"`
	Media     []string       `json:"media,omitempty"`
	QueueMode string         `json:"queue_mode"`
	Status    string         `json:"status"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

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

func (s *Store) TouchSessionActiveTurn(ctx context.Context, sessionID, claimToken string) error {
	_, err := s.db.ExecContext(ctx, `update session_active_turns set updated_at = `+defaultNow+` where session_id = ? and claim_token = ?`, sessionID, claimToken)
	if err != nil {
		return fmt.Errorf("touch session active turn: %w", err)
	}
	return nil
}

func (s *Store) ListStaleActiveTurnClaims(ctx context.Context, olderThan time.Time, sessionID string) ([]ActiveTurnClaim, error) {
	cutoff := olderThan.UTC().Format(time.RFC3339Nano)
	query := `
		select sat.session_id, sat.turn_id, coalesce(sat.worker_id,''), sat.claim_token, sat.claimed_at, sat.updated_at,
		       t.status, t.phase
		from session_active_turns sat
		join turns t on t.id = sat.turn_id
		where sat.updated_at < ?
	`
	args := []any{cutoff}
	if sessionID != "" {
		query += ` and sat.session_id = ?`
		args = append(args, sessionID)
	}
	query += ` order by sat.updated_at asc, sat.session_id asc`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list stale active turn claims: %w", err)
	}
	defer rows.Close()
	var out []ActiveTurnClaim
	for rows.Next() {
		var item ActiveTurnClaim
		if err := rows.Scan(&item.SessionID, &item.TurnID, &item.WorkerID, &item.ClaimToken, &item.ClaimedAt, &item.UpdatedAt, &item.Status, &item.Phase); err != nil {
			return nil, fmt.Errorf("scan stale active turn claim: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stale active turn claims: %w", err)
	}
	return out, nil
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
		select ?, ?, ?, ?, ?, ?, ?, 'queued', `+defaultNow+`, `+defaultNow+`
		where (
			select count(*) from steering_queue where session_id = ? and status = 'queued'
		) < 10
	`, sessionID, nilIfEmpty(turnID), role, content, payloadJSON, mediaJSON, queueMode, sessionID)
	if err != nil {
		return 0, fmt.Errorf("enqueue steering: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("enqueue steering rows: %w", err)
	}
	if rows == 0 {
		return 0, fmt.Errorf("enqueue steering: steering queue is full")
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

func dequeueSteeringTx(ctx context.Context, tx *sql.Tx, sessionID string) ([]SteeringMessage, error) {
	row := tx.QueryRowContext(ctx, `
		select id, queue_mode
		from steering_queue
		where session_id = ? and status = 'queued'
		order by id asc
		limit 1
	`, sessionID)
	var firstID int64
	var queueMode string
	if err := row.Scan(&firstID, &queueMode); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("dequeue steering head: %w", err)
	}
	limit := 1
	if queueMode == "all" {
		limit = 1000
	}
	rows, err := tx.QueryContext(ctx, `
		select id, session_id, coalesce(turn_id,''), role, content, payload_json, media_json, queue_mode, status, created_at, updated_at
		from steering_queue
		where session_id = ? and status = 'queued'
		order by id asc
		limit ?
	`, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("dequeue steering rows: %w", err)
	}
	defer rows.Close()
	var out []SteeringMessage
	var ids []int64
	for rows.Next() {
		var item SteeringMessage
		var payloadJSON, mediaJSON string
		if err := rows.Scan(&item.ID, &item.SessionID, &item.TurnID, &item.Role, &item.Content, &payloadJSON, &mediaJSON, &item.QueueMode, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan steering row: %w", err)
		}
		payload, err := unmarshalJSONMap(payloadJSON)
		if err != nil {
			return nil, err
		}
		media, err := unmarshalJSONStringArray(mediaJSON)
		if err != nil {
			return nil, err
		}
		item.Payload = payload
		item.Media = media
		out = append(out, item)
		ids = append(ids, item.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate steering rows: %w", err)
	}
	for _, id := range ids {
		if _, err := tx.ExecContext(ctx, `update steering_queue set status = 'dequeued', updated_at = `+defaultNow+` where id = ?`, id); err != nil {
			return nil, fmt.Errorf("mark steering dequeued: %w", err)
		}
	}
	return out, nil
}

func steeringMessagesToContinuationMetadata(msgs []SteeringMessage) map[string]any {
	metadata := map[string]any{"continue": true}
	items := make([]map[string]any, 0, len(msgs))
	for _, msg := range msgs {
		items = append(items, map[string]any{
			"role":       msg.Role,
			"content":    msg.Content,
			"payload":    msg.Payload,
			"media":      msg.Media,
			"queue_mode": msg.QueueMode,
		})
	}
	metadata["initial_steering"] = items
	if len(msgs) > 0 && msgs[0].Payload != nil {
		for _, key := range []string{"intent", "model", "parent_turn_id", "source_session_id", "source_agent_id", "target_agent_id", "route_mode", "route_matched_by"} {
			if value, ok := msgs[0].Payload[key]; ok {
				metadata[key] = value
			}
		}
	}
	return metadata
}

func (s *Store) StageSteeringContinuation(ctx context.Context, sessionID, turnID string) (*Turn, []SteeringMessage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("stage steering continuation begin tx: %w", err)
	}
	defer tx.Rollback()
	row := tx.QueryRowContext(ctx, `select count(*) from turns where session_id = ? and status = 'queued'`, sessionID)
	var queuedCount int
	if err := row.Scan(&queuedCount); err != nil {
		return nil, nil, fmt.Errorf("stage steering continuation queued count: %w", err)
	}
	if queuedCount > 0 {
		return nil, nil, sql.ErrNoRows
	}
	msgs, err := dequeueSteeringTx(ctx, tx, sessionID)
	if err != nil {
		return nil, nil, err
	}
	if len(msgs) == 0 {
		return nil, nil, sql.ErrNoRows
	}
	metadata := steeringMessagesToContinuationMetadata(msgs)
	metadataJSON, err := marshalJSON(metadata)
	if err != nil {
		return nil, nil, err
	}
	if _, err := tx.ExecContext(ctx, `
		insert into turns (id, session_id, status, phase, prompt, metadata_json, created_at, updated_at)
		values (?, ?, 'queued', 'queued', '', ?, `+defaultNow+`, `+defaultNow+`)
	`, turnID, sessionID, metadataJSON); err != nil {
		return nil, nil, fmt.Errorf("stage steering continuation create turn: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("stage steering continuation commit: %w", err)
	}
	if err := s.SyncSessionQueueCount(ctx, sessionID); err != nil {
		return nil, nil, err
	}
	turnRec, err := s.GetTurn(ctx, turnID)
	if err != nil {
		return nil, nil, err
	}
	return turnRec, msgs, nil
}

func (s *Store) DequeueSteering(ctx context.Context, sessionID string) ([]SteeringMessage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("dequeue steering begin tx: %w", err)
	}
	defer tx.Rollback()
	out, err := dequeueSteeringTx(ctx, tx, sessionID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("dequeue steering commit: %w", err)
	}
	return out, nil
}

func isNotFound(err error) bool { return err == sql.ErrNoRows }
