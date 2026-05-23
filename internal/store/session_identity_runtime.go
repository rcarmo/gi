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

type SessionIdentityRuntime struct {
	AgentID                 string
	Channel                 string
	Account                 string
	CanonicalScopeSignature string
}

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

func normalizeRuntimeIdentityTuple(agentID, channel, account string) (string, string, string) {
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

func (s *Store) sessionIdentityRuntime(ctx context.Context, sessionID string) (SessionIdentityRuntime, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	sessionID = strings.TrimSpace(sessionID)
	if s == nil || sessionID == "" {
		return SessionIdentityRuntime{}, sql.ErrNoRows
	}
	row := s.db.QueryRowContext(ctx, `
		select coalesce(agent_id,''), coalesce(channel,''), coalesce(account,''), coalesce(canonical_scope_signature,'')
		from session_identities
		where session_id = ?
	`, sessionID)
	identity := SessionIdentityRuntime{}
	if err := row.Scan(&identity.AgentID, &identity.Channel, &identity.Account, &identity.CanonicalScopeSignature); err != nil {
		return SessionIdentityRuntime{}, err
	}
	identity.AgentID, identity.Channel, identity.Account = normalizeRuntimeIdentityTuple(identity.AgentID, identity.Channel, identity.Account)
	identity.CanonicalScopeSignature = strings.TrimSpace(identity.CanonicalScopeSignature)
	return identity, nil
}

func (s *Store) RequireSessionIdentityRuntime(ctx context.Context, sessionID string) (SessionIdentityRuntime, error) {
	return s.sessionIdentityRuntime(ctx, sessionID)
}

