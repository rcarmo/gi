package store

import (
	"context"
	"fmt"
	"strings"
)

// SearchMessages is a bounded literal substring search. Family ancestry and
// descendants use UNION (not ALL) so malformed cycles cannot recurse forever.
func (s *Store) SearchMessages(ctx context.Context, sessionID, query, scope string, limit, offset int) ([]Message, error) {
	return s.searchMessages(ctx, sessionID, query, scope, limit, offset, false)
}
func (s *Store) SearchConversationMessages(ctx context.Context, sessionID, query, scope string, limit, offset int) ([]Message, error) {
	return s.searchMessages(ctx, sessionID, query, scope, limit, offset, true)
}
func (s *Store) searchMessages(ctx context.Context, sessionID, query, scope string, limit, offset int, conversation bool) ([]Message, error) {
	query = strings.TrimSpace(query)
	if query == "" || len(query) > 512 || limit < 1 || limit > 100 || offset < 0 || offset > 10000 {
		return nil, fmt.Errorf("invalid search bounds")
	}
	if scope != "current" && scope != "root" && scope != "all" {
		return nil, fmt.Errorf("invalid search scope")
	}
	content, visible := "content", "1"
	if conversation {
		content, visible = conversationContentSQL, conversationVisibleSQL
	}
	displayID := "0"
	if conversation {
		displayID = `(select row_id from message_rows where message_id=m.id)`
	}
	replyTo := "''"
	if conversation {
		replyTo = conversationReplyToSQL("m")
	}
	payload := "m.payload_json"
	if conversation {
		payload = "json_remove(m.payload_json,'$.provider_message','$.provider_partial','$.provider_content_blocks','$.added_tool_names')"
	}
	rows, err := s.db.QueryContext(ctx, `with recursive
 ancestors(id,parent_session_id) as (select id,parent_session_id from sessions where id=? union select s.id,s.parent_session_id from sessions s join ancestors a on a.parent_session_id=s.id),
 roots(id) as (select id from ancestors where parent_session_id is null or not exists(select 1 from sessions p where p.id=ancestors.parent_session_id)),
 family(id) as (select id from roots union select s.id from sessions s join family f on s.parent_session_id=f.id)
 select m.id,m.session_id,m.role,`+content+`,`+payload+`,m.created_at,coalesce(`+displayID+`,0),`+replyTo+` from messages m
 where (?='all' or (?='current' and m.session_id=?) or (?='root' and m.session_id in(select id from family)))
 and (`+visible+`) and instr(lower(`+content+`),lower(?))>0
 order by m.created_at desc,m.id desc limit ? offset ?`, sessionID, scope, scope, sessionID, scope, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Message{}
	for rows.Next() {
		var m Message
		var raw string
		if err = rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &raw, &m.CreatedAt, &m.DisplayRowID, &m.ReplyToID); err != nil {
			return nil, err
		}
		m.Payload, err = unmarshalJSONMap(raw)
		if err != nil {
			return nil, err
		}
		if conversation {
			m.Payload = conversationPayload(m.Payload)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
