package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	gisession "github.com/rcarmo/gi/internal/session"
)

type SessionChannelBinding struct {
	ID             int64          `json:"id"`
	SessionID      string         `json:"session_id"`
	Channel        string         `json:"channel"`
	Account        string         `json:"account"`
	BindingType    string         `json:"binding_type"`
	RemoteIdentity string         `json:"remote_identity"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

func normalizeChannelBindingValue(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func channelBindingFromAllocation(alloc gisession.Allocation) (SessionChannelBinding, bool) {
	scope := alloc.Scope
	chat := ""
	if scope.Values != nil {
		chat = strings.TrimSpace(scope.Values["chat"])
	}
	if chat == "" {
		return SessionChannelBinding{}, false
	}
	return SessionChannelBinding{
		Channel:        normalizeChannelBindingValue(scope.Channel),
		Account:        normalizeIdentityTupleValue(scope.Account, "default"),
		BindingType:    "chat",
		RemoteIdentity: normalizeChannelBindingValue(chat),
		Metadata: map[string]any{
			"agent_id": scope.AgentID,
		},
	}, true
}

func (s *Store) AttachChannelBindingForAllocation(ctx context.Context, sessionID string, alloc gisession.Allocation) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return sql.ErrNoRows
	}
	identity, err := s.RequireSessionIdentityRuntime(ctx, sessionID)
	if err != nil {
		return err
	}
	if normalizeIdentityTupleValue(identity.AgentID, "gi") != normalizeIdentityTupleValue(alloc.Scope.AgentID, "gi") {
		return fmt.Errorf("attach channel binding: allocation agent %q does not match session agent %q", alloc.Scope.AgentID, identity.AgentID)
	}
	binding, ok := channelBindingFromAllocation(alloc)
	if !ok {
		return nil
	}
	binding.SessionID = sessionID
	return s.UpsertSessionChannelBinding(ctx, binding)
}

func (s *Store) UpsertSessionChannelBinding(ctx context.Context, binding SessionChannelBinding) error {
	binding.SessionID = strings.TrimSpace(binding.SessionID)
	binding.Channel = normalizeChannelBindingValue(binding.Channel)
	binding.Account = normalizeIdentityTupleValue(binding.Account, "default")
	binding.BindingType = normalizeChannelBindingValue(binding.BindingType)
	binding.RemoteIdentity = normalizeChannelBindingValue(binding.RemoteIdentity)
	if binding.SessionID == "" || binding.Channel == "" || binding.Account == "" || binding.BindingType == "" || binding.RemoteIdentity == "" {
		return sql.ErrNoRows
	}
	metadataJSON, err := marshalJSON(binding.Metadata)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		insert into session_channel_bindings (session_id, channel, account, binding_type, remote_identity, metadata_json, created_at, updated_at)
		values (?, ?, ?, ?, ?, ?, `+defaultNow+`, `+defaultNow+`)
		on conflict(channel, account, remote_identity) do update set
			session_id = excluded.session_id,
			binding_type = excluded.binding_type,
			metadata_json = excluded.metadata_json,
			updated_at = `+defaultNow+`
	`, binding.SessionID, binding.Channel, binding.Account, binding.BindingType, binding.RemoteIdentity, metadataJSON)
	if err != nil {
		return fmt.Errorf("upsert session channel binding: %w", err)
	}
	return nil
}

func (s *Store) ResolveSessionIDByChannelBinding(ctx context.Context, channel, account, remoteIdentity string) (string, error) {
	channel = normalizeChannelBindingValue(channel)
	account = normalizeIdentityTupleValue(account, "default")
	remoteIdentity = normalizeChannelBindingValue(remoteIdentity)
	if channel == "" || account == "" || remoteIdentity == "" {
		return "", sql.ErrNoRows
	}
	row := s.db.QueryRowContext(ctx, `
		select session_id
		from session_channel_bindings
		where channel = ? and account = ? and remote_identity = ?
	`, channel, account, remoteIdentity)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		return "", err
	}
	if _, err := s.RequireSessionIdentityRuntime(ctx, sessionID); err != nil {
		return "", err
	}
	return sessionID, nil
}

func (s *Store) ListSessionChannelBindings(ctx context.Context, sessionID string) ([]SessionChannelBinding, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, sql.ErrNoRows
	}
	rows, err := s.db.QueryContext(ctx, `
		select id, session_id, channel, account, binding_type, remote_identity, metadata_json, created_at, updated_at
		from session_channel_bindings
		where session_id = ?
		order by created_at asc, id asc
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SessionChannelBinding{}
	for rows.Next() {
		var item SessionChannelBinding
		var metadataJSON string
		if err := rows.Scan(&item.ID, &item.SessionID, &item.Channel, &item.Account, &item.BindingType, &item.RemoteIdentity, &metadataJSON, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.Metadata, err = unmarshalJSONMap(metadataJSON)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
