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

func normalizeSessionAliases(aliases []string) []string {
	if len(aliases) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(aliases))
	seen := make(map[string]struct{}, len(aliases))
	for _, alias := range aliases {
		alias = strings.TrimSpace(strings.ToLower(alias))
		if alias == "" {
			continue
		}
		if _, ok := seen[alias]; ok {
			continue
		}
		seen[alias] = struct{}{}
		normalized = append(normalized, alias)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func upsertSessionAliasesTx(ctx context.Context, tx *sql.Tx, sessionID string, aliases []string) error {
	if _, err := tx.ExecContext(ctx, `delete from session_aliases where session_id = ?`, sessionID); err != nil {
		return fmt.Errorf("clear session aliases: %w", err)
	}
	for _, alias := range normalizeSessionAliases(aliases) {
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
	if err := s.hydrateSessionIdentityDetails(ctx, identities); err != nil {
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

func (s *Store) ResolveMainSession(ctx context.Context, agentID, channel, account string) (*Session, error) {
	agentID = normalizeIdentityTupleValue(agentID, "gi")
	channel = normalizeIdentityTupleValue(channel, "gi")
	account = normalizeIdentityTupleValue(account, "default")
	row := s.db.QueryRowContext(ctx, `
		select session_id
		from session_identities
		where lower(agent_id) = ? and lower(channel) = ? and lower(account) = ? and is_main_session <> 0
		order by updated_at desc, created_at desc
		limit 1
	`, agentID, channel, account)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		return nil, err
	}
	return s.GetSession(ctx, sessionID)
}

func (s *Store) SetMainSession(ctx context.Context, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return sql.ErrNoRows
	}
	identity, err := s.GetSessionIdentity(ctx, sessionID)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("set main session begin tx: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		update session_identities
		set is_main_session = 0, updated_at = `+defaultNow+`
		where lower(agent_id) = ? and lower(channel) = ? and lower(account) = ? and session_id <> ? and is_main_session <> 0
	`, normalizeIdentityTupleValue(identity.Scope.AgentID, "gi"), normalizeIdentityTupleValue(identity.Scope.Channel, "gi"), normalizeIdentityTupleValue(identity.Scope.Account, "default"), sessionID); err != nil {
		return fmt.Errorf("clear prior main sessions: %w", err)
	}
	result, err := tx.ExecContext(ctx, `
		update session_identities
		set is_main_session = 1, updated_at = `+defaultNow+`
		where session_id = ?
	`, sessionID)
	if err != nil {
		return fmt.Errorf("set main session: %w", err)
	}
	if rowsAffected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("set main session rows affected: %w", err)
	} else if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("set main session commit: %w", err)
	}
	return nil
}

func (s *Store) ResolveOrCreateMainSessionFromAllocation(ctx context.Context, in ResolveOrCreateSessionFromAllocationInput) (*Session, bool, error) {
	sess, created, err := s.ResolveOrCreateSessionFromAllocation(ctx, in)
	if err != nil {
		return nil, false, err
	}
	if err := s.SetMainSession(ctx, sess.ID); err != nil {
		return nil, false, err
	}
	reloaded, err := s.GetSession(ctx, sess.ID)
	if err != nil {
		return nil, false, err
	}
	return reloaded, created, nil
}

func (s *Store) ResolveSessionByCanonicalKey(ctx context.Context, opaqueKey string) (*Session, error) {
	return s.GetSessionByOpaqueKey(ctx, opaqueKey)
}

func (s *Store) ResolveSessionByAlias(ctx context.Context, alias string) (*Session, error) {
	return s.GetSessionByAlias(ctx, alias)
}

func (s *Store) ListSessionAliases(ctx context.Context, sessionID string) ([]string, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, sql.ErrNoRows
	}
	rows, err := s.db.QueryContext(ctx, `
		select alias
		from session_aliases
		where session_id = ?
		order by alias asc
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	aliases := make([]string, 0)
	for rows.Next() {
		var alias string
		if err := rows.Scan(&alias); err != nil {
			return nil, err
		}
		aliases = append(aliases, alias)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return aliases, nil
}

func (s *Store) UpdateSessionAliases(ctx context.Context, sessionID string, aliases []string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return sql.ErrNoRows
	}
	aliases = normalizeSessionAliases(aliases)
	aliasesJSON, err := marshalJSONArray(aliases)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("update session aliases begin tx: %w", err)
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `
		update sessions
		set aliases_json = ?, updated_at = `+defaultNow+`
		where id = ?
	`, aliasesJSON, sessionID)
	if err != nil {
		return fmt.Errorf("update session aliases: %w", err)
	}
	if rowsAffected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("update session aliases rows affected: %w", err)
	} else if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	if err := upsertSessionAliasesTx(ctx, tx, sessionID, aliases); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("update session aliases commit: %w", err)
	}
	return nil
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
	if sess, err := s.ResolveSessionByCanonicalKey(ctx, key); err == nil {
		return sess, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return s.ResolveSessionByAlias(ctx, key)
}

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
	if sess, resolveErr := s.ResolveSessionByAllocation(context.Background(), in.Allocation); resolveErr == nil {
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
