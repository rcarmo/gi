package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rcarmo/gi/internal/store/internalx"
)

func turnPhaseForStatus(status string) string {
	switch status {
	case "queued":
		return "queued"
	case "running":
		return "setup"
	case "cancelling":
		return "cancelling"
	case "completed":
		return "completed"
	case "failed":
		return "failed"
	case "aborted", "cancelled":
		return "aborted"
	default:
		return status
	}
}

func (s *Store) CreateTurnWithStatus(ctx context.Context, id, sessionID, status, prompt string, metadata map[string]any) (*Turn, error) {
	metadataJSON, err := marshalJSON(metadata)
	if err != nil {
		return nil, err
	}
	phase := turnPhaseForStatus(status)
	_, err = s.db.ExecContext(ctx, `
		insert into turns (id, session_id, status, phase, prompt, metadata_json, created_at, updated_at)
		values (?, ?, ?, ?, ?, ?, `+defaultNow+`, `+defaultNow+`)
	`, id, sessionID, status, phase, prompt, metadataJSON)
	if err != nil {
		return nil, fmt.Errorf("create turn with status: %w", err)
	}
	if err := s.SyncSessionQueueCount(ctx, sessionID); err != nil {
		return nil, err
	}
	return s.GetTurn(ctx, id)
}

func (s *Store) DeleteTurn(ctx context.Context, turnID string) error {
	var sessionID string
	row := s.db.QueryRowContext(ctx, `select session_id from turns where id = ?`, turnID)
	if err := row.Scan(&sessionID); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("delete turn lookup: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `delete from turns where id = ?`, turnID); err != nil {
		return fmt.Errorf("delete turn: %w", err)
	}
	if err := s.SyncSessionQueueCount(ctx, sessionID); err != nil {
		return err
	}
	return nil
}

func (s *Store) ListTurns(ctx context.Context, sessionID string) ([]Turn, error) {
	rows, err := s.db.QueryContext(ctx, `
		select id, session_id, status, phase, prompt, metadata_json, coalesce(claimed_by,''), coalesce(claimed_at,''), coalesce(started_at,''), coalesce(finished_at,''), created_at, updated_at
		from turns where session_id = ?
		order by created_at asc, id asc
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list turns: %w", err)
	}
	defer rows.Close()
	var out []Turn
	for rows.Next() {
		var item Turn
		var metadataJSON string
		if err := rows.Scan(&item.ID, &item.SessionID, &item.Status, &item.Phase, &item.Prompt, &metadataJSON, &item.ClaimedBy, &item.ClaimedAt, &item.StartedAt, &item.FinishedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.Metadata, err = unmarshalJSONMap(metadataJSON)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) GetNextQueuedTurn(ctx context.Context, sessionID string) (*Turn, error) {
	row := s.db.QueryRowContext(ctx, `
		select id, session_id, status, phase, prompt, metadata_json, coalesce(claimed_by,''), coalesce(claimed_at,''), coalesce(started_at,''), coalesce(finished_at,''), created_at, updated_at
		from turns where session_id = ? and status = 'queued'
		order by created_at asc, id asc
		limit 1
	`, sessionID)
	var item Turn
	var metadataJSON string
	if err := row.Scan(&item.ID, &item.SessionID, &item.Status, &item.Phase, &item.Prompt, &metadataJSON, &item.ClaimedBy, &item.ClaimedAt, &item.StartedAt, &item.FinishedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	var err error
	item.Metadata, err = unmarshalJSONMap(metadataJSON)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) UpdateTurnStatusAndPhase(ctx context.Context, turnID, status, phase string) error {
	res, err := s.db.ExecContext(ctx, `update turns set status = ?, phase = ?, updated_at = `+defaultNow+` where id = ?`, status, phase, turnID)
	if err != nil {
		return fmt.Errorf("update turn status and phase: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update turn status and phase rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("update turn status and phase: %w", sql.ErrNoRows)
	}
	turnRec, err := s.GetTurn(ctx, turnID)
	if err == nil {
		if status == "queued" || status == "running" || status == "completed" {
			if err := s.ClearTurnFailure(ctx, turnID); err != nil {
				return err
			}
		}
		if err := s.SyncSessionQueueCount(ctx, turnRec.SessionID); err != nil {
			return err
		}
	}
	if err := s.UpdateSubTurnStatusByChild(ctx, turnID, status); err != nil {
		return err
	}
	return nil
}

func (s *Store) MarkTurnClaimed(ctx context.Context, turnID, claimedBy string) error {
	res, err := s.db.ExecContext(ctx, `update turns set claimed_by = ?, claimed_at = coalesce(claimed_at, `+defaultNow+`), started_at = coalesce(started_at, `+defaultNow+`), updated_at = `+defaultNow+` where id = ?`, internalx.NilIfEmpty(claimedBy), turnID)
	if err != nil {
		return fmt.Errorf("mark turn claimed: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark turn claimed rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("mark turn claimed: %w", sql.ErrNoRows)
	}
	return nil
}

func (s *Store) ResetTurnClaim(ctx context.Context, turnID string) error {
	res, err := s.db.ExecContext(ctx, `update turns set claimed_by = null, claimed_at = null, started_at = null, updated_at = `+defaultNow+` where id = ?`, turnID)
	if err != nil {
		return fmt.Errorf("reset turn claim: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("reset turn claim rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("reset turn claim: %w", sql.ErrNoRows)
	}
	return nil
}

func (s *Store) MarkTurnFinished(ctx context.Context, turnID string) error {
	res, err := s.db.ExecContext(ctx, `update turns set finished_at = coalesce(finished_at, `+defaultNow+`), updated_at = `+defaultNow+` where id = ?`, turnID)
	if err != nil {
		return fmt.Errorf("mark turn finished: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark turn finished rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("mark turn finished: %w", sql.ErrNoRows)
	}
	return nil
}

func (s *Store) CountQueuedTurns(ctx context.Context, sessionID string) (int, error) {
	row := s.db.QueryRowContext(ctx, `select count(*) from turns where session_id = ? and status = 'queued'`, sessionID)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count queued turns: %w", err)
	}
	return count, nil
}

func (s *Store) SyncSessionQueueCount(ctx context.Context, sessionID string) error {
	count, err := s.CountQueuedTurns(ctx, sessionID)
	if err != nil {
		return err
	}
	return s.TouchSessionState(ctx, sessionID, map[string]any{"queue_count": count})
}

func (s *Store) TouchSessionState(ctx context.Context, sessionID string, patch map[string]any) error {
	if len(patch) == 0 {
		return nil
	}
	patchJSON, err := marshalJSON(patch)
	if err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, `
		update sessions
		set state_json = json_patch(coalesce(nullif(state_json, ''), '{}'), json(?)),
		    updated_at = `+defaultNow+`
		where id = ?
	`, patchJSON, sessionID)
	if err != nil {
		return fmt.Errorf("touch session state: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("touch session state rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("touch session state: %w", sql.ErrNoRows)
	}
	return nil
}
