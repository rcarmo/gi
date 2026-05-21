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

func (s *Store) SessionAgentID(ctx context.Context, sessionID string) string {
	if identity := s.SessionIdentityOrNil(ctx, sessionID); identity != nil && strings.TrimSpace(identity.Scope.AgentID) != "" {
		return identity.Scope.AgentID
	}
	return defaultSessionAgentID
}

func (s *Store) SessionChannel(ctx context.Context, sessionID string) string {
	if identity := s.SessionIdentityOrNil(ctx, sessionID); identity != nil && strings.TrimSpace(identity.Scope.Channel) != "" {
		return identity.Scope.Channel
	}
	return defaultSessionChannel
}

func (s *Store) SessionAccount(ctx context.Context, sessionID string) string {
	if identity := s.SessionIdentityOrNil(ctx, sessionID); identity != nil && strings.TrimSpace(identity.Scope.Account) != "" {
		return identity.Scope.Account
	}
	return defaultSessionAccount
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
