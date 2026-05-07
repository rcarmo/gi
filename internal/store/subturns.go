package store

import (
	"context"
	"fmt"
	"strings"
)

func (s *Store) CreateSubTurn(ctx context.Context, parentTurnID, parentSessionID, childTurnID, childSessionID, deliveryMode string, depth int, metadata map[string]any) (*SubTurn, error) {
	deliveryMode = strings.ToLower(strings.TrimSpace(deliveryMode))
	if deliveryMode == "" {
		deliveryMode = "sync"
	}
	if deliveryMode != "sync" && deliveryMode != "async" {
		return nil, fmt.Errorf("create subturn: invalid delivery mode: %s", deliveryMode)
	}
	if depth <= 0 {
		depth = 1
	}
	metadataJSON, err := marshalJSON(metadata)
	if err != nil {
		return nil, err
	}
	_, err = s.db.ExecContext(ctx, `
		insert into subturns (
			parent_turn_id, parent_session_id, child_turn_id, child_session_id,
			delivery_mode, status, depth, metadata_json, created_at, updated_at
		)
		values (?, ?, ?, ?, ?, 'running', ?, ?, `+defaultNow+`, `+defaultNow+`)
	`, parentTurnID, parentSessionID, childTurnID, childSessionID, deliveryMode, depth, metadataJSON)
	if err != nil {
		return nil, fmt.Errorf("create subturn: %w", err)
	}
	return s.GetSubTurnByChild(ctx, childTurnID)
}

func (s *Store) GetSubTurnByChild(ctx context.Context, childTurnID string) (*SubTurn, error) {
	row := s.db.QueryRowContext(ctx, `
		select id, parent_turn_id, parent_session_id, child_turn_id, child_session_id,
			delivery_mode, status, depth, metadata_json, created_at, updated_at, coalesce(finished_at,'')
		from subturns
		where child_turn_id = ?
	`, childTurnID)
	var item SubTurn
	var metadataJSON string
	if err := row.Scan(&item.ID, &item.ParentTurnID, &item.ParentSessionID, &item.ChildTurnID, &item.ChildSessionID, &item.DeliveryMode, &item.Status, &item.Depth, &metadataJSON, &item.CreatedAt, &item.UpdatedAt, &item.FinishedAt); err != nil {
		return nil, err
	}
	var err error
	item.Metadata, err = unmarshalJSONMap(metadataJSON)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) ListSubTurnsByParent(ctx context.Context, parentTurnID string) ([]SubTurn, error) {
	rows, err := s.db.QueryContext(ctx, `
		select id, parent_turn_id, parent_session_id, child_turn_id, child_session_id,
			delivery_mode, status, depth, metadata_json, created_at, updated_at, coalesce(finished_at,'')
		from subturns
		where parent_turn_id = ?
		order by created_at asc, id asc
	`, parentTurnID)
	if err != nil {
		return nil, fmt.Errorf("list subturns: %w", err)
	}
	defer rows.Close()
	var out []SubTurn
	for rows.Next() {
		var item SubTurn
		var metadataJSON string
		if err := rows.Scan(&item.ID, &item.ParentTurnID, &item.ParentSessionID, &item.ChildTurnID, &item.ChildSessionID, &item.DeliveryMode, &item.Status, &item.Depth, &metadataJSON, &item.CreatedAt, &item.UpdatedAt, &item.FinishedAt); err != nil {
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

func (s *Store) UpdateSubTurnStatusByChild(ctx context.Context, childTurnID, status string) error {
	if status == "" {
		return nil
	}
	setFinished := status == "completed" || status == "failed" || status == "aborted" || status == "cancelled"
	query := `update subturns set status = ?, updated_at = ` + defaultNow
	if setFinished {
		query += `, finished_at = coalesce(finished_at, ` + defaultNow + `)`
	}
	query += ` where child_turn_id = ?`
	_, err := s.db.ExecContext(ctx, query, status, childTurnID)
	if err != nil {
		return fmt.Errorf("update subturn status: %w", err)
	}
	return nil
}

func (s *Store) CountRunningSubTurnsByParent(ctx context.Context, parentTurnID string) (int, error) {
	row := s.db.QueryRowContext(ctx, `
		select count(*)
		from subturns
		where parent_turn_id = ? and status = 'running'
	`, parentTurnID)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count running subturns: %w", err)
	}
	return count, nil
}

func (s *Store) UpdateSubTurnMetadataByChild(ctx context.Context, childTurnID string, patch map[string]any) error {
	if strings.TrimSpace(childTurnID) == "" || len(patch) == 0 {
		return nil
	}
	normalized := make(map[string]any, len(patch))
	for k, v := range patch {
		if strings.TrimSpace(k) == "" {
			continue
		}
		normalized[k] = v
	}
	if len(normalized) == 0 {
		return nil
	}
	patchJSON, err := marshalJSON(normalized)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		update subturns
		set metadata_json = json_patch(coalesce(nullif(metadata_json, ''), '{}'), json(?)),
		    updated_at = `+defaultNow+`
		where child_turn_id = ?
	`, patchJSON, childTurnID)
	if err != nil {
		return fmt.Errorf("update subturn metadata: %w", err)
	}
	return nil
}
