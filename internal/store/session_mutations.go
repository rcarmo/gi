package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrSessionMutationInvalid = errors.New("invalid session mutation")
var ErrSessionMutationConflict = errors.New("session mutation conflicts with current state")

// SessionMutation changes display metadata only, never routing identity or history.
// Archive is a reversible picker state, not a delete or an agent shutdown.
type SessionMutation struct {
	Action string `json:"action"`
	Title  string `json:"title,omitempty"`
	Pinned *bool  `json:"pinned,omitempty"`
}

func (s *Store) MutateSession(ctx context.Context, sessionID string, mutation SessionMutation) error {
	switch mutation.Action {
	case "rename":
		mutation.Title = strings.TrimSpace(mutation.Title)
		if mutation.Title == "" || utf8.RuneCountInString(mutation.Title) > 160 || strings.ContainsAny(mutation.Title, "\r\n\x00") {
			return fmt.Errorf("%w: title must be 1–160 characters on one line", ErrSessionMutationInvalid)
		}
	case "pin":
		if mutation.Pinned == nil {
			return fmt.Errorf("%w: pinned must be a boolean", ErrSessionMutationInvalid)
		}
	case "archive", "restore":
	default:
		return fmt.Errorf("%w: unknown action %q", ErrSessionMutationInvalid, mutation.Action)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var title, raw string
	var parent sql.NullString
	if err := tx.QueryRowContext(ctx, `select title, parent_session_id, state_json from sessions where id = ?`, sessionID).Scan(&title, &parent, &raw); err != nil {
		return err
	}
	state := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	archived, _ := state["archived_at"].(string)
	if archived != "" && (mutation.Action == "rename" || mutation.Action == "pin") {
		return fmt.Errorf("%w: restore the session before editing it", ErrSessionMutationConflict)
	}
	patch := map[string]any{}
	switch mutation.Action {
	case "rename":
		title = mutation.Title
	case "pin":
		patch["pinned"] = *mutation.Pinned
	case "archive":
		if !parent.Valid || parent.String == "" {
			return fmt.Errorf("%w: main sessions cannot be archived", ErrSessionMutationConflict)
		}
		var busy bool
		if err := tx.QueryRowContext(ctx, `select exists(select 1 from turns where session_id = ? and status in ('queued', 'running')) or exists(select 1 from session_active_turns where session_id = ?)`, sessionID, sessionID).Scan(&busy); err != nil {
			return err
		}
		if busy {
			return fmt.Errorf("%w: session has running or queued work", ErrSessionMutationConflict)
		}
		if archived == "" {
			patch["archived_at"] = now
		}
	case "restore":
		patch["archived_at"] = nil // json_patch removes the key.
	}
	encoded, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `update sessions set title = ?, state_json = json_patch(coalesce(state_json, '{}'), ?), updated_at = ? where id = ?`, title, string(encoded), now, sessionID); err != nil {
		return err
	}
	return tx.Commit()
}
