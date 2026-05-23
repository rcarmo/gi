package store

import (
	"context"
	"database/sql"
	"strings"
)

const (
	defaultSessionAgentID = "agent"
	defaultSessionChannel = "gi"
	defaultSessionAccount = "default"
)

func (s *Store) SessionIdentityDimensions(ctx context.Context, sessionID string) map[string]string {
	if ctx == nil {
		ctx = context.Background()
	}
	sessionID = strings.TrimSpace(sessionID)
	if s == nil || sessionID == "" {
		return map[string]string{}
	}
	rows, err := s.db.QueryContext(ctx, `
		select coalesce(dimension_name,''), coalesce(dimension_value,'')
		from session_identity_dimensions
		where session_id = ?
		order by ordinal asc
	`, sessionID)
	if err != nil {
		return map[string]string{}
	}
	defer rows.Close()
	values := map[string]string{}
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			return map[string]string{}
		}
		name = strings.TrimSpace(strings.ToLower(name))
		value = strings.TrimSpace(strings.ToLower(value))
		if name == "" || value == "" {
			continue
		}
		values[name] = value
	}
	if err := rows.Err(); err != nil {
		return map[string]string{}
	}
	return values
}

func (s *Store) sessionIdentityTupleOrDefaults(ctx context.Context, sessionID string) (agentID, channel, account string) {
	if ctx == nil {
		ctx = context.Background()
	}
	sessionID = strings.TrimSpace(sessionID)
	if s == nil || sessionID == "" {
		return defaultSessionAgentID, defaultSessionChannel, defaultSessionAccount
	}
	row := s.db.QueryRowContext(ctx, `
		select coalesce(agent_id,''), coalesce(channel,''), coalesce(account,'')
		from session_identities
		where session_id = ?
	`, sessionID)
	if err := row.Scan(&agentID, &channel, &account); err != nil {
		return defaultSessionAgentID, defaultSessionChannel, defaultSessionAccount
	}
	agentID = strings.TrimSpace(agentID)
	channel = strings.TrimSpace(channel)
	account = strings.TrimSpace(account)
	if agentID == "" {
		agentID = defaultSessionAgentID
	}
	if channel == "" {
		channel = defaultSessionChannel
	}
	if account == "" {
		account = defaultSessionAccount
	}
	return agentID, channel, account
}

func (s *Store) SessionAgentID(ctx context.Context, sessionID string) string {
	agentID, _, _ := s.sessionIdentityTupleOrDefaults(ctx, sessionID)
	return agentID
}

func (s *Store) SessionCanonicalScopeSignature(ctx context.Context, sessionID string) string {
	if ctx == nil {
		ctx = context.Background()
	}
	sessionID = strings.TrimSpace(sessionID)
	if s == nil || sessionID == "" {
		return ""
	}
	row := s.db.QueryRowContext(ctx, `
		select coalesce(canonical_scope_signature,'')
		from session_identities
		where session_id = ?
	`, sessionID)
	var signature string
	if err := row.Scan(&signature); err != nil {
		return ""
	}
	return strings.TrimSpace(signature)
}

func (s *Store) SessionChannel(ctx context.Context, sessionID string) string {
	_, channel, _ := s.sessionIdentityTupleOrDefaults(ctx, sessionID)
	return channel
}

func (s *Store) SessionAccount(ctx context.Context, sessionID string) string {
	_, _, account := s.sessionIdentityTupleOrDefaults(ctx, sessionID)
	return account
}

func (s *Store) RequireSession(ctx context.Context, sessionID string) error {
	exists, err := s.SessionExists(ctx, sessionID)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}
	return nil
}
