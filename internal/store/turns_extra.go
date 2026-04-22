package store

import (
	"context"
	"fmt"
)

func (s *Store) CreateTurnWithStatus(ctx context.Context, id, sessionID, status, prompt string, metadata map[string]any) (*Turn, error) {
	metadataJSON, err := marshalJSON(metadata)
	if err != nil {
		return nil, err
	}
	_, err = s.db.ExecContext(ctx, `
		insert into turns (id, session_id, status, prompt, metadata_json, created_at, updated_at)
		values (?, ?, ?, ?, ?, `+defaultNow+`, `+defaultNow+`)
	`, id, sessionID, status, prompt, metadataJSON)
	if err != nil {
		return nil, fmt.Errorf("create turn with status: %w", err)
	}
	return s.GetTurn(ctx, id)
}

func (s *Store) ListTurns(ctx context.Context, sessionID string) ([]Turn, error) {
	rows, err := s.db.QueryContext(ctx, `
		select id, session_id, status, prompt, metadata_json, created_at, updated_at
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
		if err := rows.Scan(&item.ID, &item.SessionID, &item.Status, &item.Prompt, &metadataJSON, &item.CreatedAt, &item.UpdatedAt); err != nil {
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
		select id, session_id, status, prompt, metadata_json, created_at, updated_at
		from turns where session_id = ? and status = 'queued'
		order by created_at asc, id asc
		limit 1
	`, sessionID)
	var item Turn
	var metadataJSON string
	if err := row.Scan(&item.ID, &item.SessionID, &item.Status, &item.Prompt, &metadataJSON, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	var err error
	item.Metadata, err = unmarshalJSONMap(metadataJSON)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) TouchSessionState(ctx context.Context, sessionID string, patch map[string]any) error {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	state := session.State
	if state == nil {
		state = map[string]any{}
	}
	for k, v := range patch {
		state[k] = v
	}
	return s.SetSessionState(ctx, sessionID, state)
}
