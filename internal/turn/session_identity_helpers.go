package turn

import (
	"context"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

func lookupSessionIdentityWithFallback(ctx, fallback context.Context, s *store.Store, sess *store.Session) *store.SessionIdentity {
	if s == nil || sess == nil || strings.TrimSpace(sess.ID) == "" {
		return nil
	}
	opCtx := coordinationContext(ctx, fallback)
	if opCtx == nil {
		return nil
	}
	identity, err := s.GetSessionIdentity(opCtx, sess.ID)
	if err != nil {
		return nil
	}
	return identity
}

func lookupSessionIdentity(ctx context.Context, s *store.Store, sess *store.Session) *store.SessionIdentity {
	return lookupSessionIdentityWithFallback(ctx, nil, s, sess)
}

func sessionAgentIDWithStoreFallback(ctx, fallback context.Context, s *store.Store, sess *store.Session) string {
	if identity := lookupSessionIdentityWithFallback(ctx, fallback, s, sess); identity != nil && strings.TrimSpace(identity.Scope.AgentID) != "" {
		return identity.Scope.AgentID
	}
	return sessionAgentID(sess)
}

func sessionAgentIDWithStore(ctx context.Context, s *store.Store, sess *store.Session) string {
	return sessionAgentIDWithStoreFallback(ctx, nil, s, sess)
}

func sessionChannelWithStoreFallback(ctx, fallback context.Context, s *store.Store, sess *store.Session) string {
	if identity := lookupSessionIdentityWithFallback(ctx, fallback, s, sess); identity != nil && strings.TrimSpace(identity.Scope.Channel) != "" {
		return identity.Scope.Channel
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
