package store

import (
	"context"
	"database/sql"
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

type ResolveOrCreateSessionFromAllocationInput struct {
	ID              string
	ParentSessionID string
	Title           string
	State           map[string]any
	Allocation      gisession.Allocation
}

func normalizeIdentityTupleValue(value, fallback string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return fallback
	}
	return value
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
	return upsertSessionAliasesTx(ctx, tx, sessionID, aliases)
}

func scanSessionIdentityRows(rows *sql.Rows) ([]SessionIdentity, error) {
	identities := make([]SessionIdentity, 0)
	for rows.Next() {
		var identity SessionIdentity
		var isMain int
		if err := rows.Scan(&identity.SessionID, &identity.Scope.AgentID, &identity.Scope.Channel, &identity.Scope.Account, &identity.Scope.Version, &identity.CanonicalScopeSignature, &identity.OpaqueSessionKey, &isMain); err != nil {
			return nil, err
		}
		identity.IsMainSession = isMain != 0
		identity.Scope.Values = map[string]string{}
		identities = append(identities, identity)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return identities, nil
}

func (s *Store) hydrateSessionIdentityDetails(ctx context.Context, identities []SessionIdentity) error {
	if len(identities) == 0 {
		return nil
	}
	bySessionID := make(map[string]*SessionIdentity, len(identities))
	for i := range identities {
		bySessionID[identities[i].SessionID] = &identities[i]
	}
	dimRows, err := s.db.QueryContext(ctx, `
		select session_id, dimension_name, dimension_value
		from session_identity_dimensions
		order by session_id asc, ordinal asc
	`)
	if err != nil {
		return err
	}
	defer dimRows.Close()
	for dimRows.Next() {
		var sessionID, name, value string
		if err := dimRows.Scan(&sessionID, &name, &value); err != nil {
			return err
		}
		identity := bySessionID[sessionID]
		if identity == nil {
			continue
		}
		identity.Scope.Dimensions = append(identity.Scope.Dimensions, name)
		identity.Scope.Values[name] = value
	}
	if err := dimRows.Err(); err != nil {
		return err
	}
	aliasRows, err := s.db.QueryContext(ctx, `
		select session_id, alias
		from session_aliases
		order by session_id asc, alias asc
	`)
	if err != nil {
		return err
	}
	defer aliasRows.Close()
	for aliasRows.Next() {
		var sessionID, alias string
		if err := aliasRows.Scan(&sessionID, &alias); err != nil {
			return err
		}
		identity := bySessionID[sessionID]
		if identity == nil {
			continue
		}
		identity.Aliases = append(identity.Aliases, alias)
	}
	if err := aliasRows.Err(); err != nil {
		return err
	}
	return nil
}

func (s *Store) hydrateSingleSessionIdentityDetails(ctx context.Context, identity *SessionIdentity) error {
	if identity == nil || strings.TrimSpace(identity.SessionID) == "" {
		return nil
	}
	identity.Scope.Values = map[string]string{}
	dimRows, err := s.db.QueryContext(ctx, `
		select dimension_name, dimension_value
		from session_identity_dimensions
		where session_id = ?
		order by ordinal asc
	`, identity.SessionID)
	if err != nil {
		return err
	}
	defer dimRows.Close()
	for dimRows.Next() {
		var name, value string
		if err := dimRows.Scan(&name, &value); err != nil {
			return err
		}
		identity.Scope.Dimensions = append(identity.Scope.Dimensions, name)
		identity.Scope.Values[name] = value
	}
	if err := dimRows.Err(); err != nil {
		return err
	}
	aliasRows, err := s.db.QueryContext(ctx, `
		select alias
		from session_aliases
		where session_id = ?
		order by alias asc
	`, identity.SessionID)
	if err != nil {
		return err
	}
	defer aliasRows.Close()
	for aliasRows.Next() {
		var alias string
		if err := aliasRows.Scan(&alias); err != nil {
			return err
		}
		identity.Aliases = append(identity.Aliases, alias)
	}
	if err := aliasRows.Err(); err != nil {
		return err
	}
	return nil
}

func (s *Store) GetSessionIdentity(ctx context.Context, sessionID string) (*SessionIdentity, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, sql.ErrNoRows
	}
	rows, err := s.db.QueryContext(ctx, `
		select session_id, agent_id, channel, account, scope_version, canonical_scope_signature, opaque_session_key, is_main_session
		from session_identities
		where session_id = ?
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	identities, err := scanSessionIdentityRows(rows)
	if err != nil {
		return nil, err
	}
	if len(identities) == 0 {
		return nil, sql.ErrNoRows
	}
	if err := s.hydrateSingleSessionIdentityDetails(ctx, &identities[0]); err != nil {
		return nil, err
	}
	return &identities[0], nil
}

func (s *Store) ListSessionIdentities(ctx context.Context) ([]SessionIdentity, error) {
	rows, err := s.db.QueryContext(ctx, `
		select session_id, agent_id, channel, account, scope_version, canonical_scope_signature, opaque_session_key, is_main_session
		from session_identities
		order by session_id asc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	identities, err := scanSessionIdentityRows(rows)
	if err != nil {
		return nil, err
	}
	if err := s.hydrateSessionIdentityDetails(ctx, identities); err != nil {
		return nil, err
	}
	return identities, nil
}
