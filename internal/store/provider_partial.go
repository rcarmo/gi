package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	goai "github.com/rcarmo/go-ai"
)

// ProviderMessageID keeps a streaming checkpoint and its final receipt on the
// same timeline row. It is derived from Gi's locally admitted request occurrence.
func ProviderMessageID(occurrenceID string) string {
	// Retain NowID's sortable message shape for millisecond timestamp ties.
	return "msg_" + strings.TrimPrefix(occurrenceID, "inference_")
}

// SaveProviderPartial replaces one bounded snapshot, without appending deltas to
// the event journal. The caller owns detachment and bounds before admission.
func (s *Store) SaveProviderPartial(ctx context.Context, turnID, sessionID, occurrenceID, content string, payload map[string]any) error {
	if occurrenceID == "" {
		return fmt.Errorf("partial requires a provider occurrence")
	}
	ctx, err := admitToolSettlement(ctx)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = validateTurnWriterTx(ctx, tx, sessionID, turnID); err != nil {
		return err
	}
	var owner, status string
	if err = tx.QueryRowContext(ctx, `select session_id,status from turns where id=?`, turnID).Scan(&owner, &status); err != nil {
		return err
	}
	if owner != sessionID {
		return fmt.Errorf("partial session does not own turn")
	}
	var settled bool
	if err = tx.QueryRowContext(ctx, `select exists(select 1 from turn_receipts where turn_id=? and occurrence_id=?)`, turnID, occurrenceID).Scan(&settled); err != nil {
		return err
	}
	if settled {
		return nil
	}
	if status != "running" && status != "cancelling" {
		return ErrQueueConflict
	}
	var admitted bool
	if err = tx.QueryRowContext(ctx, `select exists(select 1 from turn_events where turn_id=? and session_id=? and event_type='inference.started' and json_extract(payload_json,'$.request_id')=?)`, turnID, sessionID, occurrenceID).Scan(&admitted); err != nil {
		return err
	}
	if !admitted {
		return fmt.Errorf("partial has no committed provider intent")
	}
	copied := make(map[string]any, len(payload)+5)
	for k, v := range payload {
		copied[k] = v
	}
	copied["turn_id"] = turnID
	copied["request_id"] = occurrenceID
	copied["partial"] = true
	copied["adopted"] = false
	copied["kind"] = "chat"
	copied["source"] = "inference"
	raw, err := marshalJSON(copied)
	if err != nil {
		return err
	}
	if len(raw) > 512<<10 || len(content) > 64<<10 {
		return fmt.Errorf("provider partial exceeds checkpoint bounds")
	}
	result, err := tx.ExecContext(ctx, `insert into messages(id,session_id,role,content,payload_json,created_at) values(?,?,'assistant',?,?,`+defaultNow+`) on conflict(id) do update set content=excluded.content,payload_json=excluded.payload_json where messages.session_id=excluded.session_id and json_extract(messages.payload_json,'$.partial')=1 and json_extract(messages.payload_json,'$.request_id')=? and json_extract(messages.payload_json,'$.turn_id')=?`, ProviderMessageID(occurrenceID), sessionID, content, raw, occurrenceID, turnID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("partial message identity conflict")
	}
	if _, err = tx.ExecContext(ctx, `update sessions set updated_at=`+defaultNow+` where id=?`, sessionID); err != nil {
		return err
	}
	return tx.Commit()
}

// ProviderPartialRecord is a committed snapshot with its truncation metadata.
type ProviderPartialRecord struct {
	Message *goai.Message
	Payload map[string]any
	Content string
}

// ProviderPartial returns the committed provider snapshot, never UI draft text.
func (s *Store) ProviderPartial(ctx context.Context, turnID, sessionID, occurrenceID string) (*ProviderPartialRecord, error) {
	var raw, content string
	err := s.db.QueryRowContext(ctx, `select payload_json,content from messages where id=? and session_id=? and json_extract(payload_json,'$.turn_id')=? and json_extract(payload_json,'$.request_id')=? and json_extract(payload_json,'$.partial')=1`, ProviderMessageID(occurrenceID), sessionID, turnID, occurrenceID).Scan(&raw, &content)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal([]byte(raw), &fields); err != nil {
		return nil, err
	}
	var message goai.Message
	if err = json.Unmarshal(fields["provider_message"], &message); err != nil {
		return nil, err
	}
	payload := make(map[string]any, len(fields))
	for k, v := range fields {
		payload[k] = v
	}
	return &ProviderPartialRecord{Message: &message, Payload: payload, Content: content}, nil
}

// SealProviderPartial retains the last committed snapshot as an interrupted
// receipt. The coordinator must first join the owner or establish its death;
// this method does not release claims, dispatch work, or replay a request.
func (s *Store) SealProviderPartial(ctx context.Context, turnID, sessionID, occurrenceID, reason string) (ReceiptSettlement, error) {
	record, err := s.ProviderPartial(ctx, turnID, sessionID, occurrenceID)
	if err != nil || record == nil {
		return ReceiptSettlement{}, err
	}
	payload := record.Payload
	payload["receipt_error"] = reason
	payload["partial"] = true
	usage := record.Message.Usage
	if ValidateReceiptUsage(usage) != nil {
		usage = nil
	}
	return s.SettleProviderAttempt(ctx, turnID, sessionID, occurrenceID, ProviderReceipt{MessageID: ProviderMessageID(occurrenceID), Content: record.Content, Payload: payload, Usage: usage, Unknown: true, Decision: "error_boundary", Error: reason})
}
