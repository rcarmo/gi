package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	gisession "github.com/rcarmo/gi/internal/session"
)

func (s *Store) ResolveOrCreateSessionFromAllocation(ctx context.Context, in ResolveOrCreateSessionFromAllocationInput) (*Session, bool, error) {
	in.Allocation = gisession.NormalizeAllocationIdentityLinks(in.Allocation)
	if continueSessionID := strings.TrimSpace(in.ContinueSessionID); continueSessionID != "" {
		if exists, err := s.SessionExists(ctx, continueSessionID); err != nil {
			return nil, false, err
		} else if !exists {
			return nil, false, sql.ErrNoRows
		}
		if err := s.AttachChannelBindingForAllocation(ctx, continueSessionID, in.Allocation); err != nil {
			return nil, false, err
		}
		resolved, err := s.GetSession(ctx, continueSessionID)
		if err != nil {
			return nil, false, err
		}
		return resolved, false, nil
	}
	if sessionID, err := s.FindSessionByAllocation(ctx, in.Allocation); err == nil {
		resolved, err := s.GetSession(ctx, sessionID)
		if err != nil {
			return nil, false, err
		}
		return resolved, false, nil
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}
	created, err := s.createSessionWithMetadataAndOpaqueKey(ctx, in.ID, in.ParentSessionID, in.Title, in.State, &in.Allocation.Scope, in.Allocation.SessionAliases, in.Allocation.SessionKey)
	if err == nil {
		return created, true, nil
	}
	if sessionID, resolveErr := s.FindSessionByAllocation(ctx, in.Allocation); resolveErr == nil {
		resolved, err := s.GetSession(ctx, sessionID)
		if err != nil {
			return nil, false, err
		}
		return resolved, false, nil
	}
	return nil, false, err
}

func sessionMatchesAllocationScope(signature string, scope gisession.SessionScope) bool {
	signature = strings.TrimSpace(signature)
	return signature != "" && signature == gisession.CanonicalScopeSignature(scope)
}

func (s *Store) findSessionIDByAllocation(ctx context.Context, alloc gisession.Allocation) (string, error) {
	runtimeBySessionID := map[string]SessionIdentityRuntime{}
	lookupRuntime := func(sessionID string) (SessionIdentityRuntime, error) {
		sessionID = strings.TrimSpace(sessionID)
		if cached, ok := runtimeBySessionID[sessionID]; ok {
			return cached, nil
		}
		runtime, err := s.RequireSessionIdentityRuntime(ctx, sessionID)
		if err != nil {
			return SessionIdentityRuntime{}, err
		}
		runtimeBySessionID[sessionID] = runtime
		return runtime, nil
	}
	if sessionID, err := s.ResolveSessionIDByOpaqueKey(ctx, alloc.SessionKey); err == nil {
		return sessionID, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if binding, ok := channelBindingFromAllocation(alloc); ok {
		sessionID, err := s.ResolveSessionIDByChannelBinding(ctx, binding.Channel, binding.Account, binding.RemoteIdentity)
		if err == nil {
			identity, identityErr := lookupRuntime(sessionID)
			if identityErr == nil && strings.EqualFold(strings.TrimSpace(identity.AgentID), strings.TrimSpace(alloc.Scope.AgentID)) {
				return sessionID, nil
			}
			if identityErr != nil && !errors.Is(identityErr, sql.ErrNoRows) {
				return "", identityErr
			}
		} else if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}
	for _, alias := range alloc.SessionAliases {
		sessionID, err := s.ResolveSessionIDByAlias(ctx, alias)
		if err == nil {
			identity, identityErr := lookupRuntime(sessionID)
			if identityErr == nil && sessionMatchesAllocationScope(identity.CanonicalScopeSignature, alloc.Scope) {
				return sessionID, nil
			}
			if identityErr != nil && !errors.Is(identityErr, sql.ErrNoRows) {
				return "", identityErr
			}
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}
	if signature := gisession.CanonicalScopeSignature(alloc.Scope); strings.TrimSpace(signature) != "" {
		sessionID, err := s.ResolveSessionIDByCanonicalScopeSignature(ctx, signature)
		if err == nil {
			return sessionID, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}
	return "", sql.ErrNoRows
}

func (s *Store) FindSessionByAllocation(ctx context.Context, alloc gisession.Allocation) (string, error) {
	return s.findSessionIDByAllocation(ctx, gisession.NormalizeAllocationIdentityLinks(alloc))
}
