package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	gisession "github.com/rcarmo/gi/internal/session"
)

func (s *Store) ResolveSessionByAllocation(ctx context.Context, alloc gisession.Allocation) (*Session, error) {
	return s.FindSessionByAllocation(ctx, gisession.NormalizeAllocationIdentityLinks(alloc))
}

func (s *Store) ResolveOrCreateSessionFromAllocation(ctx context.Context, in ResolveOrCreateSessionFromAllocationInput) (*Session, bool, error) {
	in.Allocation = gisession.NormalizeAllocationIdentityLinks(in.Allocation)
	if continueSessionID := strings.TrimSpace(in.ContinueSessionID); continueSessionID != "" {
		sess, err := s.GetSession(ctx, continueSessionID)
		if err != nil {
			return nil, false, err
		}
		if err := s.AttachChannelBindingForAllocation(ctx, sess.ID, in.Allocation); err != nil {
			return nil, false, err
		}
		return sess, false, nil
	}
	if sess, err := s.ResolveSessionByAllocation(ctx, in.Allocation); err == nil {
		return sess, false, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}
	created, err := s.createSessionWithMetadataAndOpaqueKey(ctx, in.ID, in.ParentSessionID, in.Title, in.State, &in.Allocation.Scope, in.Allocation.SessionAliases, in.Allocation.SessionKey)
	if err == nil {
		return created, true, nil
	}
	if sess, resolveErr := s.ResolveSessionByAllocation(ctx, in.Allocation); resolveErr == nil {
		return sess, false, nil
	}
	return nil, false, err
}

func sessionMatchesAllocationScope(identity *SessionIdentity, scope gisession.SessionScope) bool {
	if identity == nil {
		return false
	}
	return strings.TrimSpace(identity.CanonicalScopeSignature) != "" && identity.CanonicalScopeSignature == gisession.CanonicalScopeSignature(scope)
}

func (s *Store) FindSessionByAllocation(ctx context.Context, alloc gisession.Allocation) (*Session, error) {
	if sess, err := s.GetSessionByOpaqueKey(ctx, alloc.SessionKey); err == nil {
		return sess, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if binding, ok := channelBindingFromAllocation(alloc); ok {
		sess, err := s.ResolveSessionByChannelBinding(ctx, binding.Channel, binding.Account, binding.RemoteIdentity)
		if err == nil {
			identity, identityErr := s.GetSessionIdentity(ctx, sess.ID)
			if identityErr == nil && strings.EqualFold(strings.TrimSpace(identity.Scope.AgentID), strings.TrimSpace(alloc.Scope.AgentID)) {
				return sess, nil
			}
		} else if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	for _, alias := range alloc.SessionAliases {
		sess, err := s.GetSessionByAlias(ctx, alias)
		if err == nil {
			identity, identityErr := s.GetSessionIdentity(ctx, sess.ID)
			if identityErr == nil && sessionMatchesAllocationScope(identity, alloc.Scope) {
				return sess, nil
			}
			if identityErr != nil && !errors.Is(identityErr, sql.ErrNoRows) {
				return nil, identityErr
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
