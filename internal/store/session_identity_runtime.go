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

func (s *Store) SessionIdentityOrNil(ctx context.Context, sessionID string) *SessionIdentity {
	if s == nil || strings.TrimSpace(sessionID) == "" {
		return nil
	}
	identity, err := s.GetSessionIdentity(ctx, sessionID)
	if err != nil {
		return nil
	}
	return identity
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
