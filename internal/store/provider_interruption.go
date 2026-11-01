package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// generationPendingRequestTx recognises only the request admitted by this
// generation frontier. Settled responses, retries, tools and unknown suffixes
// require their own continuation routes; they cannot be promoted into inference.
func generationPendingRequestTx(ctx context.Context, tx *sql.Tx, sessionID, turnID string, seq int64, iteration int) (string, bool, error) {
	rows, err := tx.QueryContext(ctx, `select event_type,session_id,payload_json from turn_events where turn_id=? and seq>? order by seq`, turnID, seq)
	if err != nil {
		return "", false, err
	}
	defer rows.Close()
	requestID := ""
	prepared := false
	for rows.Next() {
		var kind, session, raw string
		if err = rows.Scan(&kind, &session, &raw); err != nil {
			return "", false, err
		}
		var header struct {
			RequestID string `json:"request_id"`
			Iteration int    `json:"iteration"`
			Phase     string `json:"phase"`
		}
		// Invalid data is a held checkpoint, not a transient database error.
		// Decode only identity fields; never restore the old request payload.
		if session != sessionID || json.Unmarshal([]byte(raw), &header) != nil || header.Iteration != iteration {
			return "", false, nil
		}
		switch kind {
		case "generation.recovered":
			if requestID != "" || header.Phase != "generation_ready" {
				return "", false, nil
			}
		case "inference.started":
			if requestID != "" || strings.TrimSpace(header.RequestID) == "" || header.Phase != "inference" {
				return "", false, nil
			}
			requestID = header.RequestID
		case "inference.request_prepared":
			if requestID == "" || header.RequestID != requestID || header.Phase != "inference" || prepared {
				return "", false, nil
			}
			prepared = true
		default:
			return "", false, nil
		}
	}
	if err = rows.Err(); err != nil {
		return "", false, err
	}
	if err = rows.Close(); err != nil {
		return "", false, err
	}
	if requestID != "" {
		var settled bool
		if err = tx.QueryRowContext(ctx, `select exists(select 1 from turn_receipts where turn_id=? and occurrence_id=?)`, turnID, requestID).Scan(&settled); err != nil {
			return "", false, err
		}
		if settled {
			return "", false, nil
		}
	}
	return requestID, true, nil
}

// interruptProviderTx retains only committed observed data. It neither invents
// a terminal provider response nor authorises replay of the old request.
func interruptProviderTx(ctx context.Context, tx *sql.Tx, sessionID, turnID, requestID string, iteration int) error {
	const reason = "Provider outcome unknown after process interruption"
	receipt := ProviderReceipt{Decision: "interrupted_boundary", Unknown: true, Iteration: iteration, Error: reason}
	record, err := providerPartial(ctx, tx, turnID, sessionID, requestID)
	if err != nil {
		return err
	}
	if record != nil {
		if record.Message.Role != "assistant" {
			return fmt.Errorf("interrupted provider checkpoint is not assistant output")
		}
		receipt.MessageID = ProviderMessageID(requestID)
		receipt.Content = providerPartialText(record)
		receipt.Payload = record.Payload
		receipt.Payload["receipt_error"] = reason
		receipt.Payload["interrupted"] = true
		receipt.Usage = record.Message.Usage
		if ValidateReceiptUsage(receipt.Usage) != nil {
			receipt.Usage = nil
		}
	}
	if err := validateProviderReceipt(requestID, receipt); err != nil {
		return err
	}
	settled, err := settleProviderAttemptTx(ctx, tx, turnID, sessionID, requestID, receipt)
	if err != nil {
		return err
	}
	if !settled.Applied {
		return ErrQueueConflict
	}
	return nil
}

func providerPartialText(record *ProviderPartialRecord) string {
	var text strings.Builder
	for _, block := range record.Message.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	return text.String()
}
