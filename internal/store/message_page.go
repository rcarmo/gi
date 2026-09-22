package store

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
)

var ErrMessageCursor = errors.New("invalid message cursor")

type messageCursor struct {
	Session string `json:"s"`
	Time    string `json:"t"`
	ID      string `json:"i"`
}
type MessagePage struct {
	Messages []Message `json:"messages"`
	HasMore  bool      `json:"has_more"`
	Before   string    `json:"before"`
	After    string    `json:"after"`
}

func encodeMessageCursor(m Message) string {
	raw, _ := json.Marshal(messageCursor{m.SessionID, m.CreatedAt, m.ID})
	return base64.RawURLEncoding.EncodeToString(raw)
}
func (s *Store) PageMessages(ctx context.Context, sessionID, before, after string, limit int) (MessagePage, error) {
	out := MessagePage{Messages: []Message{}}
	if limit < 1 || limit > 100 || (before != "" && after != "") {
		return out, ErrMessageCursor
	}
	cursor := before
	if after != "" {
		cursor = after
	}
	var c messageCursor
	if len(cursor) > 4096 {
		return out, ErrMessageCursor
	}
	if cursor != "" {
		raw, err := base64.RawURLEncoding.DecodeString(cursor)
		if err != nil || len(raw) > 2048 {
			return out, ErrMessageCursor
		}
		if json.Unmarshal(raw, &c) != nil || c.Session != sessionID || c.Time == "" || c.ID == "" {
			return out, ErrMessageCursor
		}
	}
	direction, order := "<", "desc"
	if after != "" {
		direction, order = ">", "asc"
	}
	query := `select id,session_id,role,content,payload_json,created_at from messages where session_id=?`
	args := []any{sessionID}
	if cursor != "" {
		query += ` and (created_at,id) ` + direction + ` (?,?)`
		args = append(args, c.Time, c.ID)
	}
	query += ` order by created_at ` + order + `,id ` + order + ` limit ?`
	args = append(args, limit+1)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Message
		var raw string
		if err = rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &raw, &m.CreatedAt); err != nil {
			return out, err
		}
		m.Payload, err = unmarshalJSONMap(raw)
		if err != nil {
			return out, err
		}
		out.Messages = append(out.Messages, m)
	}
	if err = rows.Err(); err != nil {
		return out, err
	}
	if len(out.Messages) > limit {
		out.HasMore = true
		out.Messages = out.Messages[:limit]
	}
	if after == "" {
		for l, r := 0, len(out.Messages)-1; l < r; l, r = l+1, r-1 {
			out.Messages[l], out.Messages[r] = out.Messages[r], out.Messages[l]
		}
	}
	if len(out.Messages) > 0 {
		out.Before = encodeMessageCursor(out.Messages[0])
		out.After = encodeMessageCursor(out.Messages[len(out.Messages)-1])
	}
	return out, nil
}
