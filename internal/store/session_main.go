package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func (s *Store) ResolveMainSessionID(ctx context.Context, agentID, channel, account string) (string, error) {
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
		return "", err
	}
	return sessionID, nil
}

func (s *Store) SetMainSession(ctx context.Context, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return sql.ErrNoRows
	}
	identity, err := s.RequireSessionIdentityRuntime(ctx, sessionID)
	if err != nil {
		return err
	}
	agentID := normalizeIdentityTupleValue(identity.AgentID, "gi")
	channel := normalizeIdentityTupleValue(identity.Channel, "gi")
	account := normalizeIdentityTupleValue(identity.Account, "default")
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("set main session begin tx: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		update session_identities
		set is_main_session = 0, updated_at = `+defaultNow+`
		where lower(agent_id) = ? and lower(channel) = ? and lower(account) = ? and session_id <> ? and is_main_session <> 0
	`, agentID, channel, account, sessionID); err != nil {
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
	if err := s.AttachChannelBindingForAllocation(ctx, sess.ID, in.Allocation); err != nil {
		return nil, false, err
	}
	reloaded, err := s.GetSession(ctx, sess.ID)
	if err != nil {
		return nil, false, err
	}
	return reloaded, created, nil
}
