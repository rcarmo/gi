package store

import (
	"context"
	"encoding/json"
)

// SessionDisplayCapabilities is a read hint for the compact terminal picker.
// MutateSession remains the transactional authority and must recheck each write.
type SessionDisplayCapabilities struct {
	Pinned     bool
	CanPin     bool
	CanRename  bool
	CanArchive bool
	CanRestore bool
}

func (s *Store) SessionDisplayCapabilities(ctx context.Context, id string) (SessionDisplayCapabilities, error) {
	var raw string
	var child, busy bool
	err := s.db.QueryRowContext(ctx, `select state_json, coalesce(parent_session_id,'') <> '',
 exists(select 1 from turns where session_id=sessions.id and status in ('running','queued'))
 or exists(select 1 from session_active_turns where session_id=sessions.id)
 from sessions where id=?`, id).Scan(&raw, &child, &busy)
	if err != nil {
		return SessionDisplayCapabilities{}, err
	}
	var state map[string]any
	if err = json.Unmarshal([]byte(raw), &state); err != nil {
		return SessionDisplayCapabilities{}, err
	}
	archived, _ := state["archived_at"].(string)
	pinned, _ := state["pinned"].(bool)
	return SessionDisplayCapabilities{Pinned: pinned, CanPin: archived == "", CanRename: archived == "", CanArchive: archived == "" && child && !busy, CanRestore: archived != ""}, nil
}
