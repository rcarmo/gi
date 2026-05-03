package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type RouteEvent struct {
	ID             int64          `json:"id"`
	TurnID         string         `json:"turn_id,omitempty"`
	SourceSession  string         `json:"source_session_id"`
	TargetSession  string         `json:"target_session_id,omitempty"`
	SourceAgentID  string         `json:"source_agent_id,omitempty"`
	TargetAgentID  string         `json:"target_agent_id"`
	Mode           string         `json:"mode"`
	MatchedBy      string         `json:"matched_by,omitempty"`
	RoutingPolicy  string         `json:"routing_policy,omitempty"`
	RequestedAgent string         `json:"requested_agent_id,omitempty"`
	Metadata       map[string]any `json:"metadata"`
	CreatedAt      string         `json:"created_at"`
}

func (s *Store) RecordRouteEvent(ctx context.Context, event RouteEvent) (int64, error) {
	if event.SourceSession == "" {
		return 0, fmt.Errorf("record route event: missing source_session_id")
	}
	if event.TargetAgentID == "" {
		return 0, fmt.Errorf("record route event: missing target_agent_id")
	}
	payloadJSON, err := marshalJSON(event.Metadata)
	if err != nil {
		return 0, err
	}
	var id int64
	req := `insert into routing_events (turn_id, source_session_id, target_session_id, source_agent_id, target_agent_id, mode, matched_by, routing_policy, requested_agent_id, metadata_json, created_at)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ` + defaultNow + `)`
	_, err = s.db.ExecContext(ctx, req,
		nilIfEmpty(event.TurnID),
		event.SourceSession,
		nilIfEmpty(event.TargetSession),
		nilIfEmpty(event.SourceAgentID),
		event.TargetAgentID,
		nilIfEmpty(event.Mode),
		nilIfEmpty(event.MatchedBy),
		nilIfEmpty(event.RoutingPolicy),
		nilIfEmpty(event.RequestedAgent),
		payloadJSON,
	)
	if err != nil {
		return 0, fmt.Errorf("record route event: %w", err)
	}
	row := s.db.QueryRowContext(ctx, `select last_insert_rowid()`)
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("record route event id: %w", err)
	}
	if event.SourceSession != "" {
		_, _ = s.db.ExecContext(ctx, `update sessions set updated_at = `+defaultNow+` where id = ?`, event.SourceSession)
	}
	if event.TargetSession != "" {
		_, _ = s.db.ExecContext(ctx, `update sessions set updated_at = `+defaultNow+` where id = ?`, event.TargetSession)
	}
	return id, nil
}

func (s *Store) ListRouteEvents(ctx context.Context, sessionID string) ([]RouteEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
		select id, turn_id, source_session_id, target_session_id, source_agent_id, target_agent_id, mode, matched_by, routing_policy, requested_agent_id, metadata_json, created_at
		from routing_events
		where source_session_id = ? or target_session_id = ?
		order by created_at desc, id desc
	`, sessionID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list route events: %w", err)
	}
	defer rows.Close()

	var out []RouteEvent
	for rows.Next() {
		var item RouteEvent
		var payloadJSON string
		var mode, matchedBy, policy, reqAgent sql.NullString
		if err := rows.Scan(&item.ID, &item.TurnID, &item.SourceSession, &item.TargetSession, &item.SourceAgentID, &item.TargetAgentID, &mode, &matchedBy, &policy, &reqAgent, &payloadJSON, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Mode = mode.String
		item.MatchedBy = matchedBy.String
		item.RoutingPolicy = policy.String
		item.RequestedAgent = reqAgent.String
		item.Metadata, err = unmarshalJSONMap(payloadJSON)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) GetRouteEvent(ctx context.Context, id int64) (*RouteEvent, error) {
	row := s.db.QueryRowContext(ctx, `
		select id, turn_id, source_session_id, target_session_id, source_agent_id, target_agent_id, mode, matched_by, routing_policy, requested_agent_id, metadata_json, created_at
		from routing_events
		where id = ?
	`, id)
	var item RouteEvent
	var payloadJSON string
	var mode, matchedBy, policy, reqAgent sql.NullString
	if err := row.Scan(&item.ID, &item.TurnID, &item.SourceSession, &item.TargetSession, &item.SourceAgentID, &item.TargetAgentID, &mode, &matchedBy, &policy, &reqAgent, &payloadJSON, &item.CreatedAt); err != nil {
		return nil, err
	}
	item.Mode = mode.String
	item.MatchedBy = matchedBy.String
	item.RoutingPolicy = policy.String
	item.RequestedAgent = reqAgent.String
	var err error
	item.Metadata, err = unmarshalJSONMap(payloadJSON)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func nilIfEmpty(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}
