package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	gisession "github.com/rcarmo/gi/internal/session"
)

type SessionIdentity struct {
	SessionID               string                 `json:"session_id"`
	Scope                   gisession.SessionScope `json:"scope"`
	Aliases                 []string               `json:"aliases,omitempty"`
	CanonicalScopeSignature string                 `json:"canonical_scope_signature"`
	OpaqueSessionKey        string                 `json:"opaque_session_key"`
	IsMainSession           bool                   `json:"is_main_session"`
}

func (s *Store) upsertSessionIdentityTx(ctx context.Context, tx *sql.Tx, sessionID string, scope *gisession.SessionScope, aliases []string) error {
	if scope == nil || strings.TrimSpace(sessionID) == "" {
		return nil
	}
	signature := gisession.CanonicalScopeSignature(*scope)
	opaqueKey := gisession.BuildSessionKey(*scope)
	if strings.TrimSpace(signature) == "" || strings.TrimSpace(opaqueKey) == "" {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		insert into session_identities (session_id, agent_id, channel, account, scope_version, canonical_scope_signature, opaque_session_key, is_main_session, created_at, updated_at)
		values (?, ?, ?, ?, ?, ?, ?, 0, `+defaultNow+`, `+defaultNow+`)
		on conflict(session_id) do update set
			agent_id = excluded.agent_id,
			channel = excluded.channel,
			account = excluded.account,
			scope_version = excluded.scope_version,
			canonical_scope_signature = excluded.canonical_scope_signature,
			opaque_session_key = excluded.opaque_session_key,
			updated_at = `+defaultNow+`
	`, sessionID, scope.AgentID, scope.Channel, scope.Account, scope.Version, signature, opaqueKey)
	if err != nil {
		return fmt.Errorf("upsert session identity: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `delete from session_identity_dimensions where session_id = ?`, sessionID); err != nil {
		return fmt.Errorf("clear session identity dimensions: %w", err)
	}
	for ordinal, dimension := range scope.Dimensions {
		value := strings.TrimSpace(scope.Values[dimension])
		if strings.TrimSpace(dimension) == "" || value == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			insert into session_identity_dimensions (session_id, dimension_name, dimension_value, ordinal)
			values (?, ?, ?, ?)
		`, sessionID, strings.ToLower(strings.TrimSpace(dimension)), strings.ToLower(value), ordinal); err != nil {
			return fmt.Errorf("insert session identity dimension: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `delete from session_aliases where session_id = ?`, sessionID); err != nil {
		return fmt.Errorf("clear session aliases: %w", err)
	}
	for _, alias := range aliases {
		alias = strings.TrimSpace(strings.ToLower(alias))
		if alias == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			insert into session_aliases (alias, session_id, alias_kind, created_at, updated_at)
			values (?, ?, 'compat', `+defaultNow+`, `+defaultNow+`)
			on conflict(alias) do update set
				session_id = excluded.session_id,
				alias_kind = excluded.alias_kind,
				updated_at = `+defaultNow+`
		`, alias, sessionID); err != nil {
			return fmt.Errorf("upsert session alias: %w", err)
		}
	}
	return nil
}

func (s *Store) GetSessionIdentity(ctx context.Context, sessionID string) (*SessionIdentity, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, sql.ErrNoRows
	}
	row := s.db.QueryRowContext(ctx, `
		select agent_id, channel, account, scope_version, canonical_scope_signature, opaque_session_key, is_main_session
		from session_identities
		where session_id = ?
	`, sessionID)
	identity := SessionIdentity{SessionID: sessionID}
	var isMain int
	if err := row.Scan(&identity.Scope.AgentID, &identity.Scope.Channel, &identity.Scope.Account, &identity.Scope.Version, &identity.CanonicalScopeSignature, &identity.OpaqueSessionKey, &isMain); err != nil {
		return nil, err
	}
	identity.IsMainSession = isMain != 0
	identity.Scope.Values = map[string]string{}
	rows, err := s.db.QueryContext(ctx, `
		select dimension_name, dimension_value
		from session_identity_dimensions
		where session_id = ?
		order by ordinal asc
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		identity.Scope.Dimensions = append(identity.Scope.Dimensions, name)
		identity.Scope.Values[name] = value
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	aliasRows, err := s.db.QueryContext(ctx, `
		select alias
		from session_aliases
		where session_id = ?
		order by alias asc
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer aliasRows.Close()
	for aliasRows.Next() {
		var alias string
		if err := aliasRows.Scan(&alias); err != nil {
			return nil, err
		}
		identity.Aliases = append(identity.Aliases, alias)
	}
	if err := aliasRows.Err(); err != nil {
		return nil, err
	}
	return &identity, nil
}

func (s *Store) GetSessionByOpaqueKey(ctx context.Context, opaqueKey string) (*Session, error) {
	opaqueKey = strings.TrimSpace(strings.ToLower(opaqueKey))
	if opaqueKey == "" {
		return nil, sql.ErrNoRows
	}
	row := s.db.QueryRowContext(ctx, `select session_id from session_identities where opaque_session_key = ?`, opaqueKey)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		return nil, err
	}
	return s.GetSession(ctx, sessionID)
}

func (s *Store) GetSessionByCanonicalScopeSignature(ctx context.Context, signature string) (*Session, error) {
	signature = strings.TrimSpace(strings.ToLower(signature))
	if signature == "" {
		return nil, sql.ErrNoRows
	}
	row := s.db.QueryRowContext(ctx, `select session_id from session_identities where canonical_scope_signature = ?`, signature)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		return nil, err
	}
	return s.GetSession(ctx, sessionID)
}

func (s *Store) GetSessionByAlias(ctx context.Context, alias string) (*Session, error) {
	alias = strings.TrimSpace(strings.ToLower(alias))
	if alias == "" {
		return nil, sql.ErrNoRows
	}
	row := s.db.QueryRowContext(ctx, `select session_id from session_aliases where alias = ?`, alias)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		return nil, err
	}
	return s.GetSession(ctx, sessionID)
}

func (s *Store) ResolveSessionByKeyOrAlias(ctx context.Context, key string) (*Session, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, sql.ErrNoRows
	}
	if sess, err := s.GetSession(ctx, key); err == nil {
		return sess, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if sess, err := s.GetSessionByOpaqueKey(ctx, key); err == nil {
		return sess, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return s.GetSessionByAlias(ctx, key)
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
