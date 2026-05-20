package queue

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

func EnqueueInboundWork(ctx context.Context, db *sql.DB, sourceKind, sessionID, explicitSessionKey string, envelope map[string]any) (*InboundWorkItem, error) {
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
	res, err := db.ExecContext(ctx, `
		insert into inbound_work_queue (source_kind, session_id, explicit_session_key, envelope_json, status, created_at, updated_at)
		values (?, ?, ?, ?, '`+statusQueued+`', `+defaultNow+`, `+defaultNow+`)
	`, sourceKind, nilIfEmpty(sessionID), explicitSessionKey, envelopeJSON)
	if err != nil {
		return nil, fmt.Errorf("enqueue inbound work: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("enqueue inbound work id: %w", err)
	}
	return GetInboundWork(ctx, db, id)
}

func GetInboundWork(ctx context.Context, db *sql.DB, id int64) (*InboundWorkItem, error) {
	row := db.QueryRowContext(ctx, `
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

func ListInboundWork(ctx context.Context, db *sql.DB, status string, limit int) ([]InboundWorkItem, error) {
	return ListInboundWorkFiltered(ctx, db, status, limit, nil)
}

func ListInboundWorkFiltered(ctx context.Context, db *sql.DB, status string, limit int, eligible *bool) ([]InboundWorkItem, error) {
	status = strings.TrimSpace(strings.ToLower(status))
	if limit <= 0 {
		limit = 100
	}
	query := `
		select id, source_kind, coalesce(session_id,''), explicit_session_key, envelope_json, status, attempt_count, last_error, coalesce(next_attempt_at,''), coalesce(claimed_by,''), coalesce(claimed_at,''), created_at, updated_at
		from inbound_work_queue
	`
	where := []string{}
	args := []any{}
	if status != "" {
		where = append(where, `status = ?`)
		args = append(args, status)
	}
	if eligible != nil {
		clause := `(status = '` + statusQueued + `' or (status = '` + statusRetry + `' and (next_attempt_at is null or next_attempt_at = '' or next_attempt_at <= ` + defaultNow + `)))`
		if *eligible {
			where = append(where, clause)
		} else {
			where = append(where, `not `+clause)
		}
	}
	if len(where) > 0 {
		query += ` where ` + strings.Join(where, ` and `)
	}
	query += ` order by id asc limit ?`
	args = append(args, limit)
	rows, err := db.QueryContext(ctx, query, args...)
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

func CountInboundWorkByStatus(ctx context.Context, db *sql.DB) (map[string]int, error) {
	rows, err := db.QueryContext(ctx, `select status, count(*) from inbound_work_queue group by status`)
	if err != nil {
		return nil, fmt.Errorf("count inbound work by status: %w", err)
	}
	defer rows.Close()
	counts := map[string]int{}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan inbound work status counts: %w", err)
		}
		counts[status] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inbound work status counts: %w", err)
	}
	return counts, nil
}

func CountEligibleInboundWork(ctx context.Context, db *sql.DB) (int, error) {
	row := db.QueryRowContext(ctx, `
		select count(*)
		from inbound_work_queue
		where status = '`+statusQueued+`'
		or (status = '`+statusRetry+`' and (next_attempt_at is null or next_attempt_at = '' or next_attempt_at <= `+defaultNow+`))
	`)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count eligible inbound work: %w", err)
	}
	return count, nil
}

func ClaimNextInboundWork(ctx context.Context, db *sql.DB, claimedBy string) (*InboundWorkItem, error) {
	claimedBy = strings.TrimSpace(claimedBy)
	if claimedBy == "" {
		claimedBy = defaultClaimedByWorker
	}
	row := db.QueryRowContext(ctx, `
		update inbound_work_queue
		set status = '`+statusClaimed+`', claimed_by = ?, claimed_at = `+defaultNow+`, updated_at = `+defaultNow+`
		where id = (
			select id
			from inbound_work_queue
			where status in ('`+statusQueued+`','`+statusRetry+`')
			and (next_attempt_at is null or next_attempt_at = '' or next_attempt_at <= `+defaultNow+`)
			order by id asc
			limit 1
		)
		and status in ('`+statusQueued+`','`+statusRetry+`')
		returning id
	`, claimedBy)
	var id int64
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("claim inbound work: %w", err)
	}
	return GetInboundWork(ctx, db, id)
}

func UpdateInboundWorkStatus(ctx context.Context, db *sql.DB, id int64, status string) error {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		return fmt.Errorf("update inbound work status: status is required")
	}
	_, err := db.ExecContext(ctx, `
		update inbound_work_queue
		set status = ?,
			last_error = case when ? = '`+statusCompleted+`' then '' else last_error end,
			next_attempt_at = case when ? = '`+statusCompleted+`' then null else next_attempt_at end,
			claimed_by = case when ? = '`+statusCompleted+`' then '' else claimed_by end,
			claimed_at = case when ? = '`+statusCompleted+`' then null else claimed_at end,
			updated_at = `+defaultNow+`
		where id = ?
	`, status, status, status, status, status, id)
	if err != nil {
		return fmt.Errorf("update inbound work status: %w", err)
	}
	return nil
}

func RecordInboundWorkRetry(ctx context.Context, db *sql.DB, id int64, attemptCount int, errText string, delay time.Duration) error {
	errText = strings.TrimSpace(errText)
	if delay < 0 {
		delay = 0
	}
	_, err := db.ExecContext(ctx, `
		update inbound_work_queue
		set status = '`+statusRetry+`',
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

func RecordInboundWorkFailure(ctx context.Context, db *sql.DB, id int64, attemptCount int, errText string) error {
	errText = strings.TrimSpace(errText)
	_, err := db.ExecContext(ctx, `
		update inbound_work_queue
		set status = '`+statusFailed+`',
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

func RequeueInboundWork(ctx context.Context, db *sql.DB, id int64, resetAttempts bool) (*InboundWorkItem, error) {
	attemptExpr := `attempt_count`
	if resetAttempts {
		attemptExpr = `0`
	}
	res, err := db.ExecContext(ctx, `
		update inbound_work_queue
		set status = '`+statusQueued+`',
			attempt_count = `+attemptExpr+`,
			last_error = '',
			next_attempt_at = null,
			claimed_by = '',
			claimed_at = null,
			updated_at = `+defaultNow+`
		where id = ? and status in ('`+statusFailed+`','`+statusRetry+`')
	`, id)
	if err != nil {
		return nil, fmt.Errorf("requeue inbound work: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("requeue inbound work rows: %w", err)
	}
	if rows == 0 {
		item, getErr := GetInboundWork(ctx, db, id)
		if getErr != nil {
			return nil, fmt.Errorf("requeue inbound work: %w", getErr)
		}
		return nil, fmt.Errorf("requeue inbound work: item %d is not requeueable from status %q", id, item.Status)
	}
	return GetInboundWork(ctx, db, id)
}

func DiscardInboundWork(ctx context.Context, db *sql.DB, id int64) (*InboundWorkItem, error) {
	res, err := db.ExecContext(ctx, `
		update inbound_work_queue
		set status = '`+statusDiscarded+`',
			next_attempt_at = null,
			claimed_by = '',
			claimed_at = null,
			updated_at = `+defaultNow+`
		where id = ? and status in ('`+statusQueued+`','`+statusRetry+`','`+statusFailed+`')
	`, id)
	if err != nil {
		return nil, fmt.Errorf("discard inbound work: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("discard inbound work rows: %w", err)
	}
	if rows == 0 {
		item, getErr := GetInboundWork(ctx, db, id)
		if getErr != nil {
			return nil, fmt.Errorf("discard inbound work: %w", getErr)
		}
		return nil, fmt.Errorf("discard inbound work: item %d is not discardable from status %q", id, item.Status)
	}
	return GetInboundWork(ctx, db, id)
}

