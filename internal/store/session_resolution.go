package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	gisession "github.com/rcarmo/gi/internal/session"
)

func (s *Store) ResolveSessionByAllocation(ctx context.Context, alloc gisession.Allocation) (*Session, error) {
	return s.FindSessionByAllocation(ctx, alloc)
}

func (s *Store) ResolveOrCreateSessionFromAllocation(ctx context.Context, in ResolveOrCreateSessionFromAllocationInput) (*Session, bool, error) {
	if sess, err := s.ResolveSessionByAllocation(ctx, in.Allocation); err == nil {
		return sess, false, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}
	created, err := s.CreateSessionWithMetadata(ctx, in.ID, in.ParentSessionID, in.Title, in.State, &in.Allocation.Scope, in.Allocation.SessionAliases)
	if err == nil {
		return created, true, nil
	}
	if sess, resolveErr := s.ResolveSessionByAllocation(ctx, in.Allocation); resolveErr == nil {
		return sess, false, nil
	}
	return nil, false, err
}

func sessionMatchesAllocationScope(sess *Session, scope gisession.SessionScope) bool {
	if sess == nil || sess.Scope == nil {
		return false
	}
	return gisession.CanonicalScopeSignature(*sess.Scope) == gisession.CanonicalScopeSignature(scope)
}

func (s *Store) FindSessionByAllocation(ctx context.Context, alloc gisession.Allocation) (*Session, error) {
	if sess, err := s.GetSessionByOpaqueKey(ctx, alloc.SessionKey); err == nil {
		return sess, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	for _, alias := range alloc.SessionAliases {
		sess, err := s.GetSessionByAlias(ctx, alias)
		if err == nil {
			if sessionMatchesAllocationScope(sess, alloc.Scope) {
				return sess, nil
			}
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	if signature := gisession.CanonicalScopeSignature(alloc.Scope); strings.TrimSpace(signature) != "" {
		sess, err := s.GetSessionByCanonicalScopeSignature(ctx, signature)
		if err == nil {
			return sess, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	return nil, sql.ErrNoRows
}
