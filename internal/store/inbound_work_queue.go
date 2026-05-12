package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type InboundWorkItem struct {
	ID                 int64          `json:"id"`
	SourceKind         string         `json:"source_kind"`
	SessionID          string         `json:"session_id,omitempty"`
	ExplicitSessionKey string         `json:"explicit_session_key,omitempty"`
	Envelope           map[string]any `json:"envelope,omitempty"`
	Status             string         `json:"status"`
	AttemptCount       int            `json:"attempt_count"`
	LastError          string         `json:"last_error,omitempty"`
	NextAttemptAt      string         `json:"next_attempt_at,omitempty"`
	ClaimedBy          string         `json:"claimed_by,omitempty"`
	ClaimedAt          string         `json:"claimed_at,omitempty"`
	CreatedAt          string         `json:"created_at"`
	UpdatedAt          string         `json:"updated_at"`
}

func (s *Store) EnqueueInboundWork(ctx context.Context, sourceKind, sessionID, explicitSessionKey string, envelope map[string]any) (*InboundWorkItem, error) {
	sourceKind = strings.TrimSpace(strings.ToLower(sourceKind))
	sessionID = strings.TrimSpace(sessionID)
	explicitSessionKey = strings.TrimSpace(strings.ToLower(explicitSessionKey))
	if sourceKind == "" {
		return nil, fmt.Errorf("enqueue inbound work: source kind is required")
	}
	envelopeJSON, err := marshalJSON(envelope)
	if err != nil {
		return nil, err
	}
	res, err := s.db.ExecContext(ctx, `
		insert into inbound_work_queue (source_kind, session_id, explicit_session_key, envelope_json, status, created_at, updated_at)
		values (?, ?, ?, ?, 'queued', `+defaultNow+`, `+defaultNow+`)
	`, sourceKind, nilIfEmpty(sessionID), explicitSessionKey, envelopeJSON)
	if err != nil {
		return nil, fmt.Errorf("enqueue inbound work: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("enqueue inbound work id: %w", err)
	}
	return s.GetInboundWork(ctx, id)
}

func (s *Store) GetInboundWork(ctx context.Context, id int64) (*InboundWorkItem, error) {
	row := s.db.QueryRowContext(ctx, `
		select id, source_kind, coalesce(session_id,''), explicit_session_key, envelope_json, status, attempt_count, last_error, coalesce(next_attempt_at,''), coalesce(claimed_by,''), coalesce(claimed_at,''), created_at, updated_at
		from inbound_work_queue
		where id = ?
	`, id)
	var item InboundWorkItem
	var envelopeJSON string
	if err := row.Scan(&item.ID, &item.SourceKind, &item.SessionID, &item.ExplicitSessionKey, &envelopeJSON, &item.Status, &item.AttemptCount, &item.LastError, &item.NextAttemptAt, &item.ClaimedBy, &item.ClaimedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	envelope, err := unmarshalJSONMap(envelopeJSON)
	if err != nil {
		return nil, err
	}
	item.Envelope = envelope
	return &item, nil
}

func (s *Store) ListInboundWork(ctx context.Context, status string, limit int) ([]InboundWorkItem, error) {
	status = strings.TrimSpace(strings.ToLower(status))
	if limit <= 0 {
		limit = 100
	}
	query := `
		select id, source_kind, coalesce(session_id,''), explicit_session_key, envelope_json, status, attempt_count, last_error, coalesce(next_attempt_at,''), coalesce(claimed_by,''), coalesce(claimed_at,''), created_at, updated_at
		from inbound_work_queue
	`
	args := []any{}
	if status != "" {
		query += ` where status = ?`
		args = append(args, status)
	}
	query += ` order by id asc limit ?`
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list inbound work: %w", err)
	}
	defer rows.Close()
	out := []InboundWorkItem{}
	for rows.Next() {
		var item InboundWorkItem
		var envelopeJSON string
		if err := rows.Scan(&item.ID, &item.SourceKind, &item.SessionID, &item.ExplicitSessionKey, &envelopeJSON, &item.Status, &item.AttemptCount, &item.LastError, &item.NextAttemptAt, &item.ClaimedBy, &item.ClaimedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		envelope, err := unmarshalJSONMap(envelopeJSON)
		if err != nil {
			return nil, err
		}
		item.Envelope = envelope
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list inbound work rows: %w", err)
	}
	return out, nil
}

func (s *Store) ClaimNextInboundWork(ctx context.Context, claimedBy string) (*InboundWorkItem, error) {
	claimedBy = strings.TrimSpace(claimedBy)
	if claimedBy == "" {
		claimedBy = "worker"
	}
	row := s.db.QueryRowContext(ctx, `
		update inbound_work_queue
		set status = 'claimed', claimed_by = ?, claimed_at = `+defaultNow+`, updated_at = `+defaultNow+`
		where id = (
			select id
			from inbound_work_queue
			where status in ('queued','retry')
			and (next_attempt_at is null or next_attempt_at = '' or next_attempt_at <= `+defaultNow+`)
			order by id asc
			limit 1
		)
		and status in ('queued','retry')
		returning id
	`, claimedBy)
	var id int64
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("claim inbound work: %w", err)
	}
	return s.GetInboundWork(ctx, id)
}

func (s *Store) UpdateInboundWorkStatus(ctx context.Context, id int64, status string) error {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		return fmt.Errorf("update inbound work status: status is required")
	}
	_, err := s.db.ExecContext(ctx, `
		update inbound_work_queue
		set status = ?,
			last_error = case when ? = 'completed' then '' else last_error end,
			next_attempt_at = case when ? = 'completed' then null else next_attempt_at end,
			claimed_by = case when ? = 'completed' then '' else claimed_by end,
			claimed_at = case when ? = 'completed' then null else claimed_at end,
			updated_at = `+defaultNow+`
		where id = ?
	`, status, status, status, status, status, id)
	if err != nil {
		return fmt.Errorf("update inbound work status: %w", err)
	}
	return nil
}

func (s *Store) RecordInboundWorkRetry(ctx context.Context, id int64, attemptCount int, errText string, delay time.Duration) error {
	errText = strings.TrimSpace(errText)
	if delay < 0 {
		delay = 0
	}
	_, err := s.db.ExecContext(ctx, `
		update inbound_work_queue
		set status = 'retry',
			attempt_count = ?,
			last_error = ?,
			next_attempt_at = strftime('%Y-%m-%dT%H:%M:%fZ','now', '+' || ? || ' seconds'),
			claimed_by = '',
			claimed_at = null,
			updated_at = `+defaultNow+`
		where id = ?
	`, attemptCount, errText, fmt.Sprintf("%.3f", delay.Seconds()), id)
	if err != nil {
		return fmt.Errorf("record inbound work retry: %w", err)
	}
	return nil
}

func (s *Store) RecordInboundWorkFailure(ctx context.Context, id int64, attemptCount int, errText string) error {
	errText = strings.TrimSpace(errText)
	_, err := s.db.ExecContext(ctx, `
		update inbound_work_queue
		set status = 'failed',
			attempt_count = ?,
			last_error = ?,
			next_attempt_at = null,
			claimed_by = '',
			claimed_at = null,
			updated_at = `+defaultNow+`
		where id = ?
	`, attemptCount, errText, id)
	if err != nil {
		return fmt.Errorf("record inbound work failure: %w", err)
	}
	return nil
}
