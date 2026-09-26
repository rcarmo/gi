package store

import (
	"context"
	"encoding/json"
	"errors"
	"unicode"
	"unicode/utf8"
)

func ValidWebSendToken(token string) bool {
	if token == "" || len(token) > 128 || !utf8.ValidString(token) {
		return false
	}
	for _, r := range token {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

// Record only AFTER synchronous native Submit returned success: provisional
// turn INSERTs can still be rolled back. No UPDATE/dedup by token is allowed.
func (s *Store) RecordWebSendReceipt(ctx context.Context, sourceSessionID, token string, result any) error {
	if !ValidWebSendToken(token) {
		return errors.New("invalid client_request_id")
	}
	raw, err := json.Marshal(result)
	if err != nil {
		return err
	}
	if len(raw) > 4096 {
		return errors.New("send receipt too large")
	}
	_, err = s.db.ExecContext(ctx, `insert into web_send_receipts(source_session_id,client_request_id,result_json,created_at) values(?,?,?,`+defaultNow+`)`, sourceSessionID, token, string(raw))
	return err
}

// Bound by an indexed source+token lookup and two rows, never chat history.
// Missing/duplicate receipts are unknown. This API never starts or retries work.
func (s *Store) GetWebSendReceipt(ctx context.Context, sourceSessionID, token string) (json.RawMessage, bool, error) {
	if !ValidWebSendToken(token) {
		return nil, false, errors.New("invalid client_request_id")
	}
	rows, err := s.db.QueryContext(ctx, `select result_json from web_send_receipts where source_session_id=? and client_request_id=? order by id limit 2`, sourceSessionID, token)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	var first json.RawMessage
	count := 0
	for rows.Next() {
		var raw string
		if err = rows.Scan(&raw); err != nil {
			return nil, false, err
		}
		count++
		if count == 1 {
			first = json.RawMessage(raw)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, false, err
	}
	if count != 1 {
		return nil, false, nil
	}
	if !json.Valid(first) {
		return nil, false, errors.New("invalid stored send receipt")
	}
	return first, true, nil
}
