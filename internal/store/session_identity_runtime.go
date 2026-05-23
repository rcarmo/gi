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

func (s *Store) sessionIdentityRuntimeOrDefaults(ctx context.Context, sessionID string) SessionIdentityRuntime {
	if ctx == nil {
		ctx = context.Background()
	}
	sessionID = strings.TrimSpace(sessionID)
	if s == nil || sessionID == "" {
		return SessionIdentityRuntime{AgentID: defaultSessionAgentID, Channel: defaultSessionChannel, Account: defaultSessionAccount}
	}
	row := s.db.QueryRowContext(ctx, `
		select coalesce(agent_id,''), coalesce(channel,''), coalesce(account,''), coalesce(canonical_scope_signature,'')
		from session_identities
		where session_id = ?
	`, sessionID)
	identity := SessionIdentityRuntime{}
	if err := row.Scan(&identity.AgentID, &identity.Channel, &identity.Account, &identity.CanonicalScopeSignature); err != nil {
		identity.AgentID, identity.Channel, identity.Account = defaultSessionAgentID, defaultSessionChannel, defaultSessionAccount
		identity.CanonicalScopeSignature = ""
		return identity
	}
	identity.AgentID, identity.Channel, identity.Account = normalizeRuntimeIdentityTuple(identity.AgentID, identity.Channel, identity.Account)
	identity.CanonicalScopeSignature = strings.TrimSpace(identity.CanonicalScopeSignature)
	return identity
}

func (s *Store) SessionIdentityRuntime(ctx context.Context, sessionID string) SessionIdentityRuntime {
	return s.sessionIdentityRuntimeOrDefaults(ctx, sessionID)
}

func (s *Store) SessionAgentID(ctx context.Context, sessionID string) string {
	return s.sessionIdentityRuntimeOrDefaults(ctx, sessionID).AgentID
}

func (s *Store) SessionCanonicalScopeSignature(ctx context.Context, sessionID string) string {
	return s.sessionIdentityRuntimeOrDefaults(ctx, sessionID).CanonicalScopeSignature
}

func (s *Store) SessionChannel(ctx context.Context, sessionID string) string {
	return s.sessionIdentityRuntimeOrDefaults(ctx, sessionID).Channel
}

func (s *Store) SessionAccount(ctx context.Context, sessionID string) string {
	return s.sessionIdentityRuntimeOrDefaults(ctx, sessionID).Account
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
