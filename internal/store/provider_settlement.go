package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"

	goai "github.com/rcarmo/go-ai"
)

// ProviderReceipt adopts one observed response and a durable successor boundary.
// A nil Usage means unknown remote usage, not a known zero-token request.
// Reply boundaries check steering before finalization; tools boundaries dispatch
// only the calls in the committed message. Recovery never infers a settled
// response again; interrupted attempts retain explicit usage uncertainty.
type ProviderReceipt struct {
	MessageID   string
	Content     string
	Payload     map[string]any
	Usage       *goai.Usage
	Unknown     bool
	Measurement map[string]any
	Decision    string
	Error       string
	Iteration   int
	RetryWait   *ProviderRetryWait
}

type ReceiptSettlement struct {
	Applied    bool
	MessageID  string
	Decision   string
	Accounting TurnAccounting
	RetryWait  *ProviderRetryWait
}

type TurnAccounting struct {
	Usage      goai.Usage
	Unknown    int
	Iterations int
}

func (a TurnAccounting) Map() map[string]any {
	u := a.Usage
	return map[string]any{"input": u.Input, "output": u.Output, "total": u.TotalTokens, "cache_read": u.CacheRead, "cache_write": u.CacheWrite, "cache_write_1h": u.CacheWrite1h, "reasoning": u.Reasoning, "cost_input": u.Cost.Input, "cost_output": u.Cost.Output, "cost_cache_read": u.Cost.CacheRead, "cost_cache_write": u.Cost.CacheWrite, "cost_total": u.Cost.Total, "unknown_receipts": a.Unknown, "iterations": a.Iterations}
}

// ValidateReceiptUsage rejects unrepresentable or negative accounting values.
func ValidateReceiptUsage(u *goai.Usage) error {
	if u == nil {
		return nil
	}
	for _, n := range []int{u.Input, u.Output, u.TotalTokens, u.CacheRead, u.CacheWrite, u.CacheWrite1h, u.Reasoning} {
		if n < 0 {
			return fmt.Errorf("negative receipt usage")
		}
	}
	for _, n := range []float64{u.Cost.Input, u.Cost.Output, u.Cost.CacheRead, u.Cost.CacheWrite, u.Cost.Total} {
		if n < 0 || math.IsNaN(n) || math.IsInf(n, 0) {
			return fmt.Errorf("invalid receipt cost")
		}
	}
	return nil
}

// AddReceiptUsage leaves total unchanged if either input or the sum is invalid.
func AddReceiptUsage(total *goai.Usage, u goai.Usage) error {
	if total == nil {
		return fmt.Errorf("nil receipt total")
	}
	if err := ValidateReceiptUsage(total); err != nil {
		return err
	}
	if err := ValidateReceiptUsage(&u); err != nil {
		return err
	}
	next := *total
	dst := []*int{&next.Input, &next.Output, &next.TotalTokens, &next.CacheRead, &next.CacheWrite, &next.CacheWrite1h, &next.Reasoning}
	src := []int{u.Input, u.Output, u.TotalTokens, u.CacheRead, u.CacheWrite, u.CacheWrite1h, u.Reasoning}
	for i, n := range src {
		if n > math.MaxInt-*dst[i] {
			return fmt.Errorf("receipt token total overflow")
		}
		*dst[i] += n
	}
	next.Cost.Input += u.Cost.Input
	next.Cost.Output += u.Cost.Output
	next.Cost.CacheRead += u.Cost.CacheRead
	next.Cost.CacheWrite += u.Cost.CacheWrite
	next.Cost.Total += u.Cost.Total
	if err := ValidateReceiptUsage(&next); err != nil {
		return err
	}
	*total = next
	return nil
}

func receiptAccountingTx(ctx context.Context, tx *sql.Tx, turnID string) (TurnAccounting, error) {
	var out TurnAccounting
	rows, err := tx.QueryContext(ctx, `select usage_json,usage_unknown from turn_receipts where turn_id=? order by occurrence_id`, turnID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var raw sql.NullString
		var unknown bool
		if err := rows.Scan(&raw, &unknown); err != nil {
			return out, err
		}
		if unknown || !raw.Valid {
			out.Unknown++
		}
		if !raw.Valid {
			continue
		}
		var u goai.Usage
		if err := json.Unmarshal([]byte(raw.String), &u); err != nil {
			return out, err
		}
		if err := ValidateReceiptUsage(&u); err != nil {
			return out, err
		}
		if err := AddReceiptUsage(&out.Usage, u); err != nil {
			return out, err
		}
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	if err := rows.Close(); err != nil {
		return out, err
	}
	// A durable admission without a matching receipt may have reached the remote
	// provider. Keep that uncertainty visible after a crash, without billing it.
	var lost int
	err = tx.QueryRowContext(ctx, `select count(distinct e.event_type || ':' || coalesce(json_extract(e.payload_json,'$.request_id'),json_extract(e.payload_json,'$.occurrence_id'),'legacy-event:'||e.seq)) from turn_events e where e.turn_id=? and e.event_type in ('inference.started','tool.started') and not exists(select 1 from turn_receipts r where r.turn_id=e.turn_id and r.occurrence_id=coalesce(json_extract(e.payload_json,'$.request_id'),json_extract(e.payload_json,'$.occurrence_id'),'legacy-event:'||e.seq))`, turnID).Scan(&lost)
	if err != nil {
		return out, err
	}
	out.Unknown += lost
	err = tx.QueryRowContext(ctx, `select coalesce(max(json_extract(decision_json,'$.iteration')),0) from turn_receipts where turn_id=? and kind='provider'`, turnID).Scan(&out.Iterations)
	return out, err
}

func receiptEventTx(ctx context.Context, tx *sql.Tx, turnID, sessionID, kind string, payload map[string]any) error {
	raw, err := marshalJSON(payload)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `insert into turn_events(turn_id,session_id,seq,event_type,payload_json,created_at) values(?,?,coalesce((select max(seq)+1 from turn_events where turn_id=?),1),?,?,`+defaultNow+`)`, turnID, sessionID, turnID, kind, raw)
	return err
}

func storeAccountingTx(ctx context.Context, tx *sql.Tx, turnID string) (TurnAccounting, error) {
	a, err := receiptAccountingTx(ctx, tx, turnID)
	if err != nil {
		return a, err
	}
	raw, err := marshalJSON(a.Map())
	if err != nil {
		return a, err
	}
	_, err = tx.ExecContext(ctx, `update turns set metadata_json=json_set(coalesce(nullif(metadata_json,''),'{}'),'$.usage',json(?)),updated_at=`+defaultNow+` where id=?`, raw, turnID)
	return a, err
}

func insertReceiptTx(ctx context.Context, tx *sql.Tx, turnID, occurrenceID, kind, messageID string, usage *goai.Usage, unknown bool, decision map[string]any) error {
	if err := ValidateReceiptUsage(usage); err != nil {
		return err
	}
	var usageJSON any
	if usage != nil {
		raw, err := marshalJSON(usage)
		if err != nil {
			return err
		}
		usageJSON = raw
	}
	decisionJSON, err := marshalJSON(decision)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `insert into turn_receipts(turn_id,occurrence_id,kind,message_id,usage_json,usage_unknown,decision_json,created_at) values(?,?,?,?,?,?,?,`+defaultNow+`)`, turnID, occurrenceID, kind, messageID, usageJSON, unknown, decisionJSON)
	return err
}

func (s *Store) SettleProviderAttempt(ctx context.Context, turnID, sessionID, occurrenceID string, receipt ProviderReceipt) (ReceiptSettlement, error) {
	if err := validateProviderReceipt(occurrenceID, receipt); err != nil {
		return ReceiptSettlement{}, err
	}
	ctx, err := admitToolSettlement(ctx)
	if err != nil {
		return ReceiptSettlement{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ReceiptSettlement{}, err
	}
	defer tx.Rollback()
	out, err := settleProviderAttemptTx(ctx, tx, turnID, sessionID, occurrenceID, receipt)
	if err != nil || !out.Applied {
		return out, err
	}
	if err := tx.Commit(); err != nil {
		return ReceiptSettlement{}, err
	}
	return out, nil
}

func validateProviderReceipt(occurrenceID string, receipt ProviderReceipt) error {
	if occurrenceID == "" {
		return fmt.Errorf("provider settlement requires an occurrence")
	}
	switch receipt.Decision {
	case "reply_boundary", "tools", "error_boundary", "interrupted_boundary":
	default:
		return fmt.Errorf("invalid provider successor %q", receipt.Decision)
	}
	if err := receipt.RetryWait.validate(); err != nil {
		return err
	}
	if receipt.RetryWait != nil && (receipt.Decision != "error_boundary" || receipt.RetryWait.Iteration != receipt.Iteration) {
		return fmt.Errorf("retry requires the matching failed provider receipt")
	}
	if receipt.RetryWait != nil && (receipt.Content != "" || receipt.Payload["partial"] == true) {
		return fmt.Errorf("retry cannot repeat an observed partial")
	}
	if (receipt.Decision == "tools" || receipt.Decision == "reply_boundary") && receipt.MessageID == "" {
		return fmt.Errorf("adoption requires a message identity")
	}
	if receipt.Decision == "interrupted_boundary" && (!receipt.Unknown || receipt.RetryWait != nil) {
		return fmt.Errorf("interrupted provider outcome must remain uncertain")
	}
	return ValidateReceiptUsage(receipt.Usage)
}

// settleProviderAttemptTx requires a validated receipt. Recovery uses it to fence
// and interrupt in one transaction; no successor is authorised before commit.
func settleProviderAttemptTx(ctx context.Context, tx *sql.Tx, turnID, sessionID, occurrenceID string, receipt ProviderReceipt) (ReceiptSettlement, error) {
	var out ReceiptSettlement
	if err := validateTurnWriterTx(ctx, tx, sessionID, turnID); err != nil {
		return out, err
	}
	var owner, status, phase string
	if err := tx.QueryRowContext(ctx, `select session_id,status,phase from turns where id=?`, turnID).Scan(&owner, &status, &phase); err != nil {
		return out, err
	}
	if owner != sessionID {
		return out, fmt.Errorf("provider settlement session does not own turn")
	}
	var kind, decisionJSON string
	err := tx.QueryRowContext(ctx, `select kind,coalesce(message_id,''),decision_json from turn_receipts where turn_id=? and occurrence_id=?`, turnID, occurrenceID).Scan(&kind, &out.MessageID, &decisionJSON)
	if err == nil {
		if kind != "provider" {
			return out, fmt.Errorf("receipt occurrence belongs to %s", kind)
		}
		var decision map[string]any
		if err = json.Unmarshal([]byte(decisionJSON), &decision); err != nil {
			return out, err
		}
		out.Decision, _ = decision["next"].(string)
		if raw, ok := decision["retry_wait"]; ok {
			encoded, marshalErr := json.Marshal(raw)
			if marshalErr != nil {
				return out, marshalErr
			}
			if err = json.Unmarshal(encoded, &out.RetryWait); err != nil {
				return out, err
			}
		}
		out.Accounting, err = receiptAccountingTx(ctx, tx, turnID)
		return out, err
	}
	if err != sql.ErrNoRows {
		return out, err
	}
	if status != "running" && status != "cancelling" {
		return out, ErrQueueConflict
	}
	if receipt.RetryWait != nil && status == "running" {
		if phase == "settling" || phase == "cancelling" {
			return out, ErrQueueConflict
		}
		var current bool
		err = tx.QueryRowContext(ctx, `select exists(select 1 from turn_events i where i.turn_id=? and i.event_type='inference.started' and json_extract(i.payload_json,'$.request_id')=?
		and not exists(select 1 from turn_events n where n.turn_id=i.turn_id and n.seq>i.seq and n.event_type in ('inference.started','tool.started')))
		and exists(select 1 from session_active_turns where session_id=? and turn_id=?)`, turnID, occurrenceID, sessionID, turnID).Scan(&current)
		if err != nil {
			return out, err
		}
		if !current {
			return out, ErrQueueConflict
		}
	}
	decision := map[string]any{"next": receipt.Decision, "iteration": receipt.Iteration, "error": receipt.Error}
	if status == "cancelling" {
		decision["next"] = "cancelled"
	} else if receipt.RetryWait != nil {
		decision["next"] = "retry_wait"
		decision["retry_wait"] = receipt.RetryWait
	}
	if receipt.MessageID != "" {
		messagePayload := make(map[string]any, len(receipt.Payload)+1)
		for k, v := range receipt.Payload {
			messagePayload[k] = v
		}
		messagePayload["turn_id"] = turnID
		messagePayload["request_id"] = occurrenceID
		messagePayload["adopted"] = status == "running" && (receipt.Decision == "tools" || receipt.Decision == "reply_boundary")
		payload, err := marshalJSON(messagePayload)
		if err != nil {
			return out, err
		}
		written, err := tx.ExecContext(ctx, `insert into messages(id,session_id,role,content,payload_json,created_at) values(?,?,'assistant',?,?,`+defaultNow+`) on conflict(id) do update set content=excluded.content,payload_json=excluded.payload_json where messages.session_id=excluded.session_id and json_extract(messages.payload_json,'$.partial')=1 and json_extract(messages.payload_json,'$.request_id')=? and json_extract(messages.payload_json,'$.turn_id')=?`, receipt.MessageID, sessionID, receipt.Content, payload, occurrenceID, turnID)
		if err != nil {
			return out, err
		}
		if n, err := written.RowsAffected(); err != nil {
			return out, err
		} else if n != 1 {
			return out, fmt.Errorf("provider message identity conflict")
		}

	}
	if err = insertReceiptTx(ctx, tx, turnID, occurrenceID, "provider", receipt.MessageID, receipt.Usage, receipt.Unknown, decision); err != nil {
		return out, err
	}
	out.Accounting, err = storeAccountingTx(ctx, tx, turnID)
	if err != nil {
		return out, err
	}
	if receipt.Measurement != nil {
		if err = receiptEventTx(ctx, tx, turnID, sessionID, "context.measured", receipt.Measurement); err != nil {
			return out, err
		}
	}
	event := map[string]any{"request_id": occurrenceID, "message_id": receipt.MessageID, "decision": decision, "usage": out.Accounting.Map(), "checkpoint": true, "phase": "inference"}
	eventType := "inference.settled"
	if receipt.Decision == "interrupted_boundary" {
		eventType = "inference.interrupted"
	}
	if err = receiptEventTx(ctx, tx, turnID, sessionID, eventType, event); err != nil {
		return out, err
	}
	if decision["next"] == "retry_wait" {
		payload := receipt.RetryWait.Payload()
		payload["request_id"] = occurrenceID
		if err = receiptEventTx(ctx, tx, turnID, sessionID, "inference.retry_scheduled", payload); err != nil {
			return out, err
		}
		if _, err = tx.ExecContext(ctx, `update turns set phase='retry_wait',updated_at=`+defaultNow+` where id=? and session_id=? and status='running'`, turnID, sessionID); err != nil {
			return out, err
		}
	}
	if _, err = tx.ExecContext(ctx, `update sessions set updated_at=`+defaultNow+` where id=?`, sessionID); err != nil {
		return out, err
	}
	out.Applied = true
	out.MessageID = receipt.MessageID
	out.Decision, _ = decision["next"].(string)
	if out.Decision == "retry_wait" {
		p := *receipt.RetryWait
		p.Lifecycle = receipt.RetryWait.Lifecycle.Clone()
		out.RetryWait = &p
	}
	return out, nil
}

// TurnAccounting rebuilds totals from committed occurrences, never event totals.
func (s *Store) TurnAccounting(ctx context.Context, turnID string) (TurnAccounting, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return TurnAccounting{}, err
	}
	defer tx.Rollback()
	return receiptAccountingTx(ctx, tx, turnID)
}
