package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func (s *Store) ListSessionIDs(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `select id from sessions order by created_at asc, id asc`)
	if err != nil {
		return nil, fmt.Errorf("list session ids: %w", err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list session ids rows: %w", err)
	}
	return out, nil
}

func (s *Store) FindChildSessionIDByParentAndAgent(ctx context.Context, parentSessionID, agentID string) (string, error) {
	parentSessionID = strings.TrimSpace(parentSessionID)
	agentID = strings.TrimSpace(strings.ToLower(agentID))
	if parentSessionID == "" || agentID == "" {
		return "", sql.ErrNoRows
	}
	row := s.db.QueryRowContext(ctx, `
		select s.id
		from sessions s
		join session_identities si on si.session_id = s.id
		where s.parent_session_id = ? and lower(si.agent_id) = ?
		order by s.updated_at desc, s.created_at desc
		limit 1
	`, parentSessionID, agentID)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		return "", err
	}
	return sessionID, nil
}

func (s *Store) FindSiblingChildSessionIDByParentAndAgent(ctx context.Context, siblingSessionID, agentID string) (string, error) {
	siblingSessionID = strings.TrimSpace(siblingSessionID)
	agentID = strings.TrimSpace(strings.ToLower(agentID))
	if siblingSessionID == "" || agentID == "" {
		return "", sql.ErrNoRows
	}
	row := s.db.QueryRowContext(ctx, `
		select child.id
		from sessions sibling
		join sessions child on child.parent_session_id = sibling.parent_session_id
		join session_identities si on si.session_id = child.id
		where sibling.id = ? and sibling.parent_session_id <> '' and child.id <> sibling.id and lower(si.agent_id) = ?
		order by child.updated_at desc, child.created_at desc
		limit 1
	`, siblingSessionID, agentID)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		return "", err
	}
	return sessionID, nil
}
