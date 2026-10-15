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
	return s.pageMessages(ctx, sessionID, before, after, limit, false)
}

func (s *Store) PageConversationMessages(ctx context.Context, sessionID, before, after string, limit int) (MessagePage, error) {
	return s.pageMessages(ctx, sessionID, before, after, limit, true)
}

func (s *Store) pageMessages(ctx context.Context, sessionID, before, after string, limit int, conversation bool) (MessagePage, error) {
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
	content := "content"
	if conversation {
		content = conversationContentSQL
	}
	displayID := "0"
	if conversation {
		displayID = `(select row_id from message_rows where message_id=messages.id)`
	}
	replyTo := "''"
	if conversation {
		replyTo = conversationReplyToSQL("messages")
	}
	source, prefix := "messages", ""
	args := []any{}
	if conversation {
		// Merge display-only terminal events before cursor filtering and LIMIT.
		// Provider history and the raw message APIs never query this relation.
		prefix = `with terminal_events as (` + terminalToolEventsSQL + `), conversation_messages as (
 select id,session_id,role,content,payload_json,created_at from messages
 union all
 select 'terminal:' || turn_id || ':' || json_extract(payload_json,'$.occurrence_id'),
 ?, 'tool_result', 'Tool execution ' || substr(event_type,6),
 json_set(payload_json,'$.kind','tool_result','$.tool_name','codemode','$.turn_id',turn_id,
 '$.status',case when event_type='tool.cancelled' then 'cancelled' else 'error' end,
 '$.is_error',json('true'),'$.source','terminal_projection'),created_at
 from terminal_events where json_extract(payload_json,'$.tool')='codemode'
 and event_type in ('tool.cancelled','tool.aborted')
 and json_type(payload_json,'$.occurrence_id')='text' and json_extract(payload_json,'$.occurrence_id')!=''
 ) `
		source = "conversation_messages as messages"
		args = append(args, sessionID, sessionID)
	}
	query := prefix + `select id,session_id,role,` + content + `,payload_json,created_at,coalesce(` + displayID + `,0),` + replyTo + ` from ` + source + ` where session_id=?`
	if conversation {
		query += ` and (` + conversationVisibleSQL + `)`
	}
	args = append(args, sessionID)
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
		if err = rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &raw, &m.CreatedAt, &m.DisplayRowID, &m.ReplyToID); err != nil {
			return out, err
		}
		m.Payload, err = unmarshalJSONMap(raw)
		if err != nil {
			return out, err
		}
		if conversation {
			m.Payload = conversationPayload(m.Payload)
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
