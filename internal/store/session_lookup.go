package store

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

func (s *Store) ListSessionAgentIDs(ctx context.Context) (map[string]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	rows, err := s.db.QueryContext(ctx, `
		select s.id, coalesce(nullif(trim(si.agent_id), ''), ?)
		from sessions s
		left join session_identities si on si.session_id = s.id
		order by s.created_at asc, s.id asc
	`, defaultSessionAgentID)
	if err != nil {
		return nil, fmt.Errorf("list session agent ids: %w", err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var sessionID, agentID string
		if err := rows.Scan(&sessionID, &agentID); err != nil {
			return nil, err
		}
		sessionID = strings.TrimSpace(sessionID)
		agentID = strings.TrimSpace(agentID)
		if sessionID == "" {
			continue
		}
		if agentID == "" {
			agentID = defaultSessionAgentID
		}
		out[sessionID] = agentID
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list session agent ids rows: %w", err)
	}
	return out, nil
}

func (s *Store) ListSessionIDs(ctx context.Context) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	rows, err := s.db.QueryContext(ctx, `select id from sessions order by created_at asc, id asc`)
	if err != nil {
		return nil, fmt.Errorf("list session ids: %w", err)
	}
	defer rows.Close()
	out := []string{}
	seen := map[string]struct{}{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list session ids rows: %w", err)
	}
	sort.Strings(out)
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
