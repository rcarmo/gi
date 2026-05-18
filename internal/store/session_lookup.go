package store

import (
	"context"
	"database/sql"
	"strings"
)

func (s *Store) FindChildSessionByParentAndAgent(ctx context.Context, parentSessionID, agentID string) (*Session, error) {
	parentSessionID = strings.TrimSpace(parentSessionID)
	agentID = strings.TrimSpace(strings.ToLower(agentID))
	if parentSessionID == "" || agentID == "" {
		return nil, sql.ErrNoRows
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
		return nil, err
	}
	return s.GetSession(ctx, sessionID)
}

func (s *Store) FindSiblingChildSessionByParentAndAgent(ctx context.Context, siblingSessionID, agentID string) (*Session, error) {
	siblingSessionID = strings.TrimSpace(siblingSessionID)
	agentID = strings.TrimSpace(strings.ToLower(agentID))
	if siblingSessionID == "" || agentID == "" {
		return nil, sql.ErrNoRows
	}
	row := s.db.QueryRowContext(ctx, `
		select child.id
		from sessions sibling
		join sessions child on child.parent_session_id = sibling.parent_session_id
		join session_identities si on si.session_id = child.id
		where sibling.id = ? and sibling.parent_session_id <> '' and lower(si.agent_id) = ?
		order by child.updated_at desc, child.created_at desc
		limit 1
	`, siblingSessionID, agentID)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		return nil, err
	}
	return s.GetSession(ctx, sessionID)
}
