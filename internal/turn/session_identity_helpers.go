package turn

import (
	"context"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

func lookupSessionIdentityByIDWithFallback(ctx, fallback context.Context, s *store.Store, sessionID string) *store.SessionIdentity {
	if s == nil || strings.TrimSpace(sessionID) == "" {
		return nil
	}
	opCtx := coordinationContext(ctx, fallback)
	if opCtx == nil {
		return nil
	}
	identity, err := s.GetSessionIdentity(opCtx, sessionID)
	if err != nil {
		return nil
	}
	return identity
}

func lookupSessionIdentityByID(ctx context.Context, s *store.Store, sessionID string) *store.SessionIdentity {
	return lookupSessionIdentityByIDWithFallback(ctx, nil, s, sessionID)
}

func lookupSessionIdentityWithFallback(ctx, fallback context.Context, s *store.Store, sess *store.Session) *store.SessionIdentity {
	if sess == nil {
		return nil
	}
	return lookupSessionIdentityByIDWithFallback(ctx, fallback, s, sess.ID)
}

func lookupSessionIdentity(ctx context.Context, s *store.Store, sess *store.Session) *store.SessionIdentity {
	return lookupSessionIdentityWithFallback(ctx, nil, s, sess)
}

func sessionAgentIDWithStoreFallback(ctx, fallback context.Context, s *store.Store, sess *store.Session) string {
	if identity := lookupSessionIdentityWithFallback(ctx, fallback, s, sess); identity != nil && strings.TrimSpace(identity.Scope.AgentID) != "" {
		return identity.Scope.AgentID
	}
	if s != nil && sess != nil && strings.TrimSpace(sess.ID) != "" {
		return "agent"
	}
	return sessionAgentID(sess)
}

func sessionAgentIDWithStore(ctx context.Context, s *store.Store, sess *store.Session) string {
	return sessionAgentIDWithStoreFallback(ctx, nil, s, sess)
}

func sessionAgentIDForSessionIDWithFallback(ctx, fallback context.Context, s *store.Store, sessionID string) string {
	if identity := lookupSessionIdentityByIDWithFallback(ctx, fallback, s, sessionID); identity != nil && strings.TrimSpace(identity.Scope.AgentID) != "" {
		return identity.Scope.AgentID
	}
	if s != nil && strings.TrimSpace(sessionID) != "" {
		return "agent"
	}
	return "agent"
}

func sessionAgentIDForSessionID(ctx context.Context, s *store.Store, sessionID string) string {
	return sessionAgentIDForSessionIDWithFallback(ctx, nil, s, sessionID)
}

func sessionChannelWithStoreFallback(ctx, fallback context.Context, s *store.Store, sess *store.Session) string {
	if identity := lookupSessionIdentityWithFallback(ctx, fallback, s, sess); identity != nil && strings.TrimSpace(identity.Scope.Channel) != "" {
		return identity.Scope.Channel
	}
	if s != nil && sess != nil && strings.TrimSpace(sess.ID) != "" {
		return "gi"
	}
	return sessionChannel(sess)
}

func sessionChannelWithStore(ctx context.Context, s *store.Store, sess *store.Session) string {
	return sessionChannelWithStoreFallback(ctx, nil, s, sess)
}

func sessionAccountWithStoreFallback(ctx, fallback context.Context, s *store.Store, sess *store.Session) string {
	if identity := lookupSessionIdentityWithFallback(ctx, fallback, s, sess); identity != nil && strings.TrimSpace(identity.Scope.Account) != "" {
		return identity.Scope.Account
	}
	if s != nil && sess != nil && strings.TrimSpace(sess.ID) != "" {
		return "default"
	}
	return sessionAccount(sess)
}

func sessionAccountWithStore(ctx context.Context, s *store.Store, sess *store.Session) string {
	return sessionAccountWithStoreFallback(ctx, nil, s, sess)
}

func sessionAgentID(sess *store.Session) string {
	if sess != nil && sess.Scope != nil && sess.Scope.AgentID != "" {
		return sess.Scope.AgentID
	}
	return "agent"
}

func sessionChannel(sess *store.Session) string {
	if sess != nil && sess.Scope != nil && strings.TrimSpace(sess.Scope.Channel) != "" {
		return sess.Scope.Channel
	}
	return "gi"
}

func sessionAccount(sess *store.Session) string {
	if sess != nil && sess.Scope != nil && strings.TrimSpace(sess.Scope.Account) != "" {
		return sess.Scope.Account
	}
	return "default"
}

func normalizeAgentID(v string) string {
	return strings.TrimSpace(strings.TrimPrefix(normalizedLowerString(v), "@"))
}
