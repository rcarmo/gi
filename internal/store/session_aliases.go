package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

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

func (s *Store) requireIdentityRuntimeSessionID(ctx context.Context, sessionID string) (string, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return "", sql.ErrNoRows
	}
	if _, err := s.RequireSessionIdentityRuntime(ctx, sessionID); err != nil {
		return "", err
	}
	return sessionID, nil
}

func (s *Store) ResolveSessionIDByOpaqueKey(ctx context.Context, opaqueKey string) (string, error) {
	opaqueKey = strings.TrimSpace(strings.ToLower(opaqueKey))
	if opaqueKey == "" {
		return "", sql.ErrNoRows
	}
	row := s.db.QueryRowContext(ctx, `select session_id from session_identities where opaque_session_key = ?`, opaqueKey)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		return "", err
	}
	return s.requireIdentityRuntimeSessionID(ctx, sessionID)
}

func (s *Store) ResolveSessionIDByCanonicalScopeSignature(ctx context.Context, signature string) (string, error) {
	signature = strings.TrimSpace(strings.ToLower(signature))
	if signature == "" {
		return "", sql.ErrNoRows
	}
	row := s.db.QueryRowContext(ctx, `select session_id from session_identities where canonical_scope_signature = ?`, signature)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		return "", err
	}
	return s.requireIdentityRuntimeSessionID(ctx, sessionID)
}

func (s *Store) ResolveSessionIDByAlias(ctx context.Context, alias string) (string, error) {
	alias = strings.TrimSpace(strings.ToLower(alias))
	if alias == "" {
		return "", sql.ErrNoRows
	}
	row := s.db.QueryRowContext(ctx, `select session_id from session_aliases where alias = ?`, alias)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		return "", err
	}
	return s.requireIdentityRuntimeSessionID(ctx, sessionID)
}

func (s *Store) ResolveSessionIDByKeyOrAlias(ctx context.Context, key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", sql.ErrNoRows
	}
	if sessionID, err := s.requireIdentityRuntimeSessionID(ctx, key); err == nil {
		return sessionID, nil
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if sessionID, err := s.ResolveSessionIDByOpaqueKey(ctx, key); err == nil {
		return sessionID, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if sessionID, err := s.ResolveSessionIDByAlias(ctx, key); err == nil {
		return sessionID, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	return "", sql.ErrNoRows
}
