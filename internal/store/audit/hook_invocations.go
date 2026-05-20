package audit

import (
	"context"
	"database/sql"
	"fmt"
)

type HookInvocation struct {
	ID         int64          `json:"id"`
	TurnID     string         `json:"turn_id,omitempty"`
	SessionID  string         `json:"session_id,omitempty"`
	HookName   string         `json:"hook_name"`
	HookPhase  string         `json:"hook_phase"`
	HookSource string         `json:"hook_source"`
	Action     string         `json:"action"`
	Request    map[string]any `json:"request,omitempty"`
	Response   map[string]any `json:"response,omitempty"`
	ErrorText  string         `json:"error_text,omitempty"`
	DurationMS int            `json:"duration_ms"`
	CreatedAt  string         `json:"created_at"`
}

func RecordHookInvocation(ctx context.Context, db *sql.DB, turnID, sessionID, hookName, hookPhase, hookSource, action string, request any, response any, errorText string, durationMS int) (int64, error) {
	if hookName == "" {
		return 0, fmt.Errorf("record hook invocation: missing hook_name")
	}
	if hookPhase == "" {
		hookPhase = hookName
	}
	if action == "" {
		action = defaultHookActionContinue
	}
	requestJSON, err := marshalJSON(request)
	if err != nil {
		return 0, err
	}
	responseJSON, err := marshalJSON(response)
	if err != nil {
		return 0, err
	}
	_, err = db.ExecContext(ctx, `
		insert into hook_invocations (turn_id, session_id, hook_name, hook_phase, hook_source, action, request_json, response_json, error_text, duration_ms, created_at)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, `+defaultNow+`)
	`, nilIfEmpty(turnID), nilIfEmpty(sessionID), hookName, hookPhase, nilIfEmpty(hookSource), action, requestJSON, responseJSON, errorText, durationMS)
	if err != nil {
		return 0, fmt.Errorf("record hook invocation: %w", err)
	}
	var id int64
	row := db.QueryRowContext(ctx, `select last_insert_rowid()`)
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("record hook invocation id: %w", err)
	}
	return id, nil
}

func ListHookInvocationsByTurn(ctx context.Context, db *sql.DB, turnID string) ([]HookInvocation, error) {
	rows, err := db.QueryContext(ctx, `
		select id, coalesce(turn_id,''), coalesce(session_id,''), hook_name, hook_phase, coalesce(hook_source,''), action, request_json, response_json, error_text, duration_ms, created_at
		from hook_invocations
		where turn_id = ?
		order by id asc
	`, turnID)
	if err != nil {
		return nil, fmt.Errorf("list hook invocations by turn: %w", err)
	}
	defer rows.Close()
	var out []HookInvocation
	for rows.Next() {
		var item HookInvocation
		var requestJSON, responseJSON string
		if err := rows.Scan(&item.ID, &item.TurnID, &item.SessionID, &item.HookName, &item.HookPhase, &item.HookSource, &item.Action, &requestJSON, &responseJSON, &item.ErrorText, &item.DurationMS, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Request, err = unmarshalJSONMap(requestJSON)
		if err != nil {
			return nil, err
		}
		item.Response, err = unmarshalJSONMap(responseJSON)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func ListHookInvocationsBySession(ctx context.Context, db *sql.DB, sessionID string) ([]HookInvocation, error) {
	rows, err := db.QueryContext(ctx, `
		select id, coalesce(turn_id,''), coalesce(session_id,''), hook_name, hook_phase, coalesce(hook_source,''), action, request_json, response_json, error_text, duration_ms, created_at
		from hook_invocations
		where session_id = ?
		order by id asc
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list hook invocations by session: %w", err)
	}
	defer rows.Close()
	var out []HookInvocation
	for rows.Next() {
		var item HookInvocation
		var requestJSON, responseJSON string
		if err := rows.Scan(&item.ID, &item.TurnID, &item.SessionID, &item.HookName, &item.HookPhase, &item.HookSource, &item.Action, &requestJSON, &responseJSON, &item.ErrorText, &item.DurationMS, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Request, err = unmarshalJSONMap(requestJSON)
		if err != nil {
			return nil, err
		}
		item.Response, err = unmarshalJSONMap(responseJSON)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func GetHookInvocation(ctx context.Context, db *sql.DB, id int64) (*HookInvocation, error) {
	row := db.QueryRowContext(ctx, `
		select id, coalesce(turn_id,''), coalesce(session_id,''), hook_name, hook_phase, coalesce(hook_source,''), action, request_json, response_json, error_text, duration_ms, created_at
		from hook_invocations
		where id = ?
	`, id)
	var item HookInvocation
	var requestJSON, responseJSON string
	if err := row.Scan(&item.ID, &item.TurnID, &item.SessionID, &item.HookName, &item.HookPhase, &item.HookSource, &item.Action, &requestJSON, &responseJSON, &item.ErrorText, &item.DurationMS, &item.CreatedAt); err != nil {
		return nil, err
	}
	var err error
	item.Request, err = unmarshalJSONMap(requestJSON)
	if err != nil {
		return nil, err
	}
	item.Response, err = unmarshalJSONMap(responseJSON)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

