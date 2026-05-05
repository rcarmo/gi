package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rcarmo/gi/internal/session"
	_ "modernc.org/sqlite"
)

const defaultNow = "strftime('%Y-%m-%dT%H:%M:%fZ','now')"

type Store struct {
	db *sql.DB
}

type Session struct {
	ID              string                `json:"id"`
	ParentSessionID string                `json:"parent_session_id,omitempty"`
	Title           string                `json:"title"`
	State           map[string]any        `json:"state"`
	Scope           *session.SessionScope `json:"scope,omitempty"`
	Aliases         []string              `json:"aliases,omitempty"`
	CreatedAt       string                `json:"created_at"`
	UpdatedAt       string                `json:"updated_at"`
}

type Message struct {
	ID        string         `json:"id"`
	SessionID string         `json:"session_id"`
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	Payload   map[string]any `json:"payload"`
	CreatedAt string         `json:"created_at"`
}

type Turn struct {
	ID         string         `json:"id"`
	SessionID  string         `json:"session_id"`
	Status     string         `json:"status"`
	Phase      string         `json:"phase,omitempty"`
	Prompt     string         `json:"prompt"`
	Metadata   map[string]any `json:"metadata"`
	ClaimedBy  string         `json:"claimed_by,omitempty"`
	ClaimedAt  string         `json:"claimed_at,omitempty"`
	StartedAt  string         `json:"started_at,omitempty"`
	FinishedAt string         `json:"finished_at,omitempty"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
}

type TurnEvent struct {
	Seq       int64          `json:"seq"`
	TurnID    string         `json:"turn_id"`
	SessionID string         `json:"session_id"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload"`
	CreatedAt string         `json:"created_at"`
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := configure(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := initSchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) DB() *sql.DB { return s.db }

func configure(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA busy_timeout=5000;",
		"PRAGMA temp_store=MEMORY;",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("configure sqlite (%s): %w", pragma, err)
		}
	}
	return nil
}

func (s *Store) CreateSession(ctx context.Context, id, title string, state map[string]any) (*Session, error) {
	alloc := session.AllocateDefaultSession("gi", "gi", "default", id)
	return s.CreateSessionWithMetadata(ctx, id, "", title, state, &alloc.Scope, alloc.SessionAliases)
}

func (s *Store) CreateSessionWithMetadata(ctx context.Context, id, parentSessionID, title string, state map[string]any, scope *session.SessionScope, aliases []string) (*Session, error) {
	stateJSON, err := marshalJSON(state)
	if err != nil {
		return nil, err
	}
	scopeJSON, err := marshalJSON(scope)
	if err != nil {
		return nil, err
	}
	aliasesJSON, err := marshalJSONArray(aliases)
	if err != nil {
		return nil, err
	}
	var parent any
	if parentSessionID != "" {
		parent = parentSessionID
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("create session begin tx: %w", err)
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
		insert into sessions (id, parent_session_id, title, state_json, scope_json, aliases_json, created_at, updated_at)
		values (?, ?, ?, ?, ?, ?, `+defaultNow+`, `+defaultNow+`)
	`, id, parent, title, stateJSON, scopeJSON, aliasesJSON)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	if err := s.upsertSessionIdentityTx(ctx, tx, id, scope, aliases); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("create session commit: %w", err)
	}
	return s.GetSession(ctx, id)
}

func (s *Store) GetSession(ctx context.Context, id string) (*Session, error) {
	row := s.db.QueryRowContext(ctx, `
		select id, parent_session_id, title, state_json, scope_json, aliases_json, created_at, updated_at
		from sessions where id = ?
	`, id)
	var out Session
	var parent sql.NullString
	var stateJSON, scopeJSON, aliasesJSON string
	if err := row.Scan(&out.ID, &parent, &out.Title, &stateJSON, &scopeJSON, &aliasesJSON, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return nil, err
	}
	if parent.Valid {
		out.ParentSessionID = parent.String
	}
	state, err := unmarshalJSONMap(stateJSON)
	if err != nil {
		return nil, err
	}
	out.State = state
	out.Scope, err = unmarshalSessionScope(scopeJSON)
	if err != nil {
		return nil, err
	}
	out.Aliases, err = unmarshalJSONStringArray(aliasesJSON)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *Store) ListSessions(ctx context.Context) ([]Session, error) {
	rows, err := s.db.QueryContext(ctx, `
		select id, parent_session_id, title, state_json, scope_json, aliases_json, created_at, updated_at
		from sessions
		order by updated_at desc, created_at desc
	`)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()
	var out []Session
	for rows.Next() {
		var item Session
		var parent sql.NullString
		var stateJSON, scopeJSON, aliasesJSON string
		if err := rows.Scan(&item.ID, &parent, &item.Title, &stateJSON, &scopeJSON, &aliasesJSON, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if parent.Valid {
			item.ParentSessionID = parent.String
		}
		item.State, err = unmarshalJSONMap(stateJSON)
		if err != nil {
			return nil, err
		}
		item.Scope, err = unmarshalSessionScope(scopeJSON)
		if err != nil {
			return nil, err
		}
		item.Aliases, err = unmarshalJSONStringArray(aliasesJSON)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) AddMessage(ctx context.Context, id, sessionID, role, content string, payload map[string]any) error {
	payloadJSON, err := marshalJSON(payload)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		insert into messages (id, session_id, role, content, payload_json, created_at)
		values (?, ?, ?, ?, ?, `+defaultNow+`)
	`, id, sessionID, role, content, payloadJSON)
	if err != nil {
		return fmt.Errorf("add message: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `update sessions set updated_at = `+defaultNow+` where id = ?`, sessionID)
	return err
}

func (s *Store) ListMessages(ctx context.Context, sessionID string) ([]Message, error) {
	rows, err := s.db.QueryContext(ctx, `
		select id, session_id, role, content, payload_json, created_at
		from messages where session_id = ?
		order by created_at asc, id asc
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()
	var out []Message
	for rows.Next() {
		var item Message
		var payloadJSON string
		if err := rows.Scan(&item.ID, &item.SessionID, &item.Role, &item.Content, &payloadJSON, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Payload, err = unmarshalJSONMap(payloadJSON)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) CreateTurn(ctx context.Context, id, sessionID, prompt string, metadata map[string]any) (*Turn, error) {
	metadataJSON, err := marshalJSON(metadata)
	if err != nil {
		return nil, err
	}
	_, err = s.db.ExecContext(ctx, `
		insert into turns (id, session_id, status, phase, prompt, metadata_json, created_at, updated_at)
		values (?, ?, 'running', 'setup', ?, ?, `+defaultNow+`, `+defaultNow+`)
	`, id, sessionID, prompt, metadataJSON)
	if err != nil {
		return nil, fmt.Errorf("create turn: %w", err)
	}
	return s.GetTurn(ctx, id)
}

func (s *Store) GetTurn(ctx context.Context, id string) (*Turn, error) {
	row := s.db.QueryRowContext(ctx, `select id, session_id, status, phase, prompt, metadata_json, coalesce(claimed_by,''), coalesce(claimed_at,''), coalesce(started_at,''), coalesce(finished_at,''), created_at, updated_at from turns where id = ?`, id)
	var out Turn
	var metadataJSON string
	if err := row.Scan(&out.ID, &out.SessionID, &out.Status, &out.Phase, &out.Prompt, &metadataJSON, &out.ClaimedBy, &out.ClaimedAt, &out.StartedAt, &out.FinishedAt, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return nil, err
	}
	var err error
	out.Metadata, err = unmarshalJSONMap(metadataJSON)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *Store) AppendTurnEvent(ctx context.Context, turnID, sessionID, eventType string, payload map[string]any) error {
	payloadJSON, err := marshalJSON(payload)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		insert into turn_events (turn_id, session_id, seq, event_type, payload_json, created_at)
		values (
			?, ?,
			coalesce((select max(seq) + 1 from turn_events where turn_id = ?), 1),
			?, ?, `+defaultNow+`
		)
	`, turnID, sessionID, turnID, eventType, payloadJSON)
	if err != nil {
		return fmt.Errorf("append turn event: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `update turns set updated_at = `+defaultNow+` where id = ?`, turnID)
	return err
}

func (s *Store) ListTurnEvents(ctx context.Context, turnID string) ([]TurnEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
		select seq, turn_id, session_id, event_type, payload_json, created_at
		from turn_events where turn_id = ?
		order by seq asc
	`, turnID)
	if err != nil {
		return nil, fmt.Errorf("list turn events: %w", err)
	}
	defer rows.Close()
	var out []TurnEvent
	for rows.Next() {
		var item TurnEvent
		var payloadJSON string
		if err := rows.Scan(&item.Seq, &item.TurnID, &item.SessionID, &item.Type, &payloadJSON, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Payload, err = unmarshalJSONMap(payloadJSON)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) UpdateTurnStatus(ctx context.Context, turnID, status string) error {
	phase := turnPhaseForStatus(status)
	if err := s.UpdateTurnStatusAndPhase(ctx, turnID, status, phase); err != nil {
		return fmt.Errorf("update turn status: %w", err)
	}
	if status == "completed" || status == "failed" || status == "aborted" || status == "cancelled" {
		_ = s.MarkTurnFinished(ctx, turnID)
	}
	return nil
}

func (s *Store) SetSessionState(ctx context.Context, sessionID string, state map[string]any) error {
	stateJSON, err := marshalJSON(state)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `update sessions set state_json = ?, updated_at = `+defaultNow+` where id = ?`, stateJSON, sessionID)
	return err
}

func marshalJSON(v any) (string, error) {
	if v == nil {
		return "{}", nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("marshal json: %w", err)
	}
	return string(b), nil
}

func unmarshalJSONMap(raw string) (map[string]any, error) {
	if raw == "" {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("unmarshal json map: %w", err)
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

func marshalJSONArray(v []string) (string, error) {
	if len(v) == 0 {
		return "[]", nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("marshal json array: %w", err)
	}
	return string(b), nil
}

func unmarshalJSONStringArray(raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("unmarshal json string array: %w", err)
	}
	return out, nil
}

func unmarshalSessionScope(raw string) (*session.SessionScope, error) {
	if raw == "" || raw == "{}" {
		return nil, nil
	}
	var out session.SessionScope
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("unmarshal session scope: %w", err)
	}
	if out.Version == 0 {
		return nil, nil
	}
	return &out, nil
}

func (s *Store) CloneSession(ctx context.Context, sourceSessionID, newID, newTitle, newAgentID string) (*Session, error) {
	source, err := s.GetSession(ctx, sourceSessionID)
	if err != nil {
		return nil, err
	}
	state := map[string]any{}
	for k, v := range source.State {
		state[k] = v
	}
	state["forked_from"] = sourceSessionID
	state["status"] = "idle"
	state["active_turn_id"] = nil
	logicalChatID := newID
	alloc := session.AllocateDefaultSession(newAgentID, "gi", "default", logicalChatID)
	cloned, err := s.CreateSessionWithMetadata(ctx, newID, sourceSessionID, newTitle, state, &alloc.Scope, alloc.SessionAliases)
	if err != nil {
		return nil, err
	}
	messages, err := s.ListMessages(ctx, sourceSessionID)
	if err != nil {
		return nil, err
	}
	for _, msg := range messages {
		payload := map[string]any{}
		for k, v := range msg.Payload {
			payload[k] = v
		}
		payload["forked_from_message_id"] = msg.ID
		if err := s.AddMessage(ctx, NowID("msg"), cloned.ID, msg.Role, msg.Content, payload); err != nil {
			return nil, err
		}
	}
	return s.GetSession(ctx, cloned.ID)
}

func NowID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
