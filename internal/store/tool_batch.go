package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	goai "github.com/rcarmo/go-ai"
	"strconv"
	"strings"
)

// ToolBatchCheckpoint pins the accepted loop/lifecycle policy and the exact
// provider receipt supplying this batch. Calls come from that receipt's owned
// provider snapshot, never from a partial response or the display projection.
type ToolBatchCheckpoint struct {
	Version              int                        `json:"version"`
	ProviderOccurrenceID string                     `json:"provider_occurrence_id"`
	CallIDs              []string                   `json:"call_ids"`
	Generation           *GenerationReadyCheckpoint `json:"generation"`
}

func (p *ToolBatchCheckpoint) Validate() error {
	if p == nil || p.Version != 1 || strings.TrimSpace(p.ProviderOccurrenceID) == "" {
		return fmt.Errorf("unsupported tool batch checkpoint")
	}
	if len(p.CallIDs) == 0 {
		return fmt.Errorf("empty tool batch")
	}
	seen := make(map[string]bool, len(p.CallIDs))
	for _, id := range p.CallIDs {
		if strings.TrimSpace(id) == "" || seen[id] {
			return fmt.Errorf("ambiguous tool batch call identity")
		}
		seen[id] = true
	}
	return p.Generation.Validate()
}

type toolBatchSlotKey struct{}
type toolBatchSlot struct {
	TurnID           string
	BatchID          string
	Index            int
	Iteration        int
	FailureSignature string
	RepeatedFailures int
	WireCallID       string
	WireToolName     string
}

// WithToolBatch identifies a native batch without granting a per-call slot.
func WithToolBatch(ctx context.Context, turnID, batchID string) context.Context {
	return context.WithValue(ctx, toolBatchSlotKey{}, toolBatchSlot{TurnID: turnID, BatchID: batchID, Index: -1})
}

func ToolBatchID(ctx context.Context, turnID string) string {
	if ctx == nil {
		return ""
	}
	slot, _ := ctx.Value(toolBatchSlotKey{}).(toolBatchSlot)
	if slot.TurnID != turnID {
		return ""
	}
	return slot.BatchID
}

// WithToolBatchSlot carries ordinal provenance even if a preparation hook
// changes the tool's name, ID or arguments. Nested turns cannot inherit it.
func WithToolBatchSlot(ctx context.Context, turnID, batchID string, index int, iteration int, failureSignature string, repeatedFailures int, wire ...goai.ToolCall) context.Context {
	slot := toolBatchSlot{TurnID: turnID, BatchID: batchID, Index: index, Iteration: iteration, FailureSignature: failureSignature, RepeatedFailures: repeatedFailures}
	if len(wire) > 0 {
		slot.WireCallID, slot.WireToolName = wire[0].ID, wire[0].Name
	}
	return context.WithValue(ctx, toolBatchSlotKey{}, slot)
}

func toolBatchEvent(ctx context.Context, turnID string, event map[string]any) map[string]any {
	if ctx == nil {
		return event
	}
	slot, _ := ctx.Value(toolBatchSlotKey{}).(toolBatchSlot)
	if slot.TurnID != turnID || slot.BatchID == "" || slot.Index < 0 {
		return event
	}
	owned := make(map[string]any, len(event)+4)
	for key, value := range event {
		owned[key] = value
	}
	owned["batch_id"], owned["batch_index"], owned["iteration"] = slot.BatchID, slot.Index, slot.Iteration
	if _, ok := owned["failure_signature"]; !ok {
		owned["failure_signature"] = slot.FailureSignature
	}
	if _, ok := owned["repeated_tool_failures"]; !ok {
		owned["repeated_tool_failures"] = slot.RepeatedFailures
	}
	return owned
}

// SaveToolBatch is the first boundary after a committed tools response. All
// previous physical effects must have receipts before any new call is prepared.
func (s *Store) SaveToolBatch(ctx context.Context, sessionID, turnID string, p *ToolBatchCheckpoint) error {
	if err := p.Validate(); err != nil {
		return err
	}
	return s.AdmitTurnEffect(ctx, sessionID, turnID, "waiting_on_tools", "tool.batch_ready", map[string]any{"phase": "tool_batch", "iteration": p.Generation.Iteration, "checkpoint": true, "continuation": p})
}

func checkToolBatchReceiptTx(ctx context.Context, tx *sql.Tx, sessionID, turnID string, p *ToolBatchCheckpoint) error {
	var next, rawCalls string
	var iteration int
	var adopted sql.NullBool
	if err := tx.QueryRowContext(ctx, `select json_extract(r.decision_json,'$.next'),json_extract(r.decision_json,'$.iteration'),json_extract(m.payload_json,'$.adopted'),json_extract(m.payload_json,'$.tool_calls') from turn_receipts r join messages m on m.id=r.message_id and m.session_id=? where r.turn_id=? and r.occurrence_id=? and r.kind='provider'`, sessionID, turnID, p.ProviderOccurrenceID).Scan(&next, &iteration, &adopted, &rawCalls); err != nil {
		return err
	}
	if next != "tools" || iteration != p.Generation.Iteration || !adopted.Valid || !adopted.Bool {
		return fmt.Errorf("tool batch requires an adopted matching provider receipt")
	}
	var calls []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(rawCalls), &calls); err != nil {
		return err
	}
	if len(calls) != len(p.CallIDs) {
		return fmt.Errorf("tool batch call count mismatch")
	}
	for i, call := range calls {
		if call.ID != p.CallIDs[i] {
			return fmt.Errorf("tool batch call order mismatch")
		}
	}
	var latest string
	if err := tx.QueryRowContext(ctx, `select json_extract(payload_json,'$.request_id') from turn_events where turn_id=? and session_id=? and event_type='inference.settled' order by seq desc limit 1`, turnID, sessionID).Scan(&latest); err != nil {
		return err
	}
	if latest != p.ProviderOccurrenceID {
		return fmt.Errorf("tool batch does not reference the latest settled provider")
	}
	return nil
}

// ClaimToolBatchHook records an at-most-once hook attempt before invoking it.
// A crash after this commit leaves an uncertain attempt, never permission to
// replay the hook. Keys distinguish batch boundaries and individual results.
func (s *Store) ClaimToolBatchHook(ctx context.Context, sessionID, turnID, batchID, key string) (bool, error) {
	if batchID == "" {
		return true, nil
	}
	if key != "execution_start" && key != "execution_end" && key != "turn_end" {
		index, err := strconv.Atoi(strings.TrimPrefix(key, "result:"))
		slot, _ := ctx.Value(toolBatchSlotKey{}).(toolBatchSlot)
		if err != nil || !strings.HasPrefix(key, "result:") || index < 0 || slot.TurnID != turnID || slot.BatchID != batchID || slot.Index != index {
			return false, fmt.Errorf("invalid tool batch hook key")
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	// Lock before reading: concurrent child settlement must not force a WAL
	// read-to-write upgrade while claiming the once-only hook attempt.
	if _, err = tx.ExecContext(ctx, `update turns set updated_at=updated_at where id=? and session_id=?`, turnID, sessionID); err != nil {
		return false, err
	}
	if err = validateTurnWriterTx(ctx, tx, sessionID, turnID); err != nil {
		return false, err
	}
	var status, phase string
	if err = tx.QueryRowContext(ctx, `select status,phase from turns where id=? and session_id=?`, turnID, sessionID).Scan(&status, &phase); err != nil {
		return false, err
	}
	if status == "running" {
		if phase != "waiting_on_tools" {
			return false, ErrQueueConflict
		}
	} else if status != "completed" && status != "failed" && status != "aborted" && status != "cancelled" {
		return false, ErrQueueConflict
	}
	var batchCount, iteration int
	if err = tx.QueryRowContext(ctx, `select count(*),coalesce(max(json_extract(payload_json,'$.iteration')),0) from turn_events where turn_id=? and session_id=? and event_type='tool.batch_ready' and json_extract(payload_json,'$.continuation.provider_occurrence_id')=? and seq=(select max(seq) from turn_events where turn_id=? and event_type='tool.batch_ready')`, turnID, sessionID, batchID, turnID).Scan(&batchCount, &iteration); err != nil {
		return false, err
	}
	if batchCount != 1 {
		return false, fmt.Errorf("missing or ambiguous tool batch")
	}
	var count int
	if err = tx.QueryRowContext(ctx, `select count(*) from turn_events where turn_id=? and session_id=? and event_type='tool.batch_hook_started' and json_extract(payload_json,'$.batch_id')=? and json_extract(payload_json,'$.key')=?`, turnID, sessionID, batchID, key).Scan(&count); err != nil {
		return false, err
	}
	if count != 0 {
		return false, nil
	}
	payload := toolBatchEvent(ctx, turnID, map[string]any{"phase": "tool_batch", "checkpoint": true, "batch_id": batchID, "key": key, "iteration": iteration})
	if err = receiptEventTx(ctx, tx, turnID, sessionID, "tool.batch_hook_started", payload); err != nil {
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func decodeToolBatch(raw string) (*ToolBatchCheckpoint, error) {
	var saved struct {
		Continuation *ToolBatchCheckpoint `json:"continuation"`
	}
	if err := json.Unmarshal([]byte(raw), &saved); err != nil {
		return nil, err
	}
	if err := saved.Continuation.Validate(); err != nil {
		return nil, err
	}
	return saved.Continuation, nil
}

func checkToolBatchSlotTx(ctx context.Context, tx *sql.Tx, sessionID, turnID string) error {
	if ctx == nil {
		return nil
	}
	slot, _ := ctx.Value(toolBatchSlotKey{}).(toolBatchSlot)
	if slot.TurnID != turnID || slot.BatchID == "" || slot.Index < 0 {
		return nil
	}
	var raw string
	if err := tx.QueryRowContext(ctx, `select payload_json from turn_events where turn_id=? and session_id=? and event_type='tool.batch_ready' order by seq desc limit 1`, turnID, sessionID).Scan(&raw); err != nil {
		return err
	}
	checkpoint, err := decodeToolBatch(raw)
	if err != nil {
		return err
	}
	if checkpoint.ProviderOccurrenceID != slot.BatchID || slot.Index >= len(checkpoint.CallIDs) || slot.Iteration != checkpoint.Generation.Iteration || slot.RepeatedFailures < 0 || (slot.WireCallID != "" && slot.WireCallID != checkpoint.CallIDs[slot.Index]) {
		return fmt.Errorf("tool slot does not belong to current batch")
	}
	return nil
}

// ToolBatchWireCall keeps model results paired with the provider's original
// identity even when an authorised preparation hook changes execution identity.
func ToolBatchWireCall(ctx context.Context, turnID string) (string, string) {
	if ctx == nil {
		return "", ""
	}
	slot, _ := ctx.Value(toolBatchSlotKey{}).(toolBatchSlot)
	if slot.TurnID != turnID || slot.Index < 0 {
		return "", ""
	}
	return slot.WireCallID, slot.WireToolName
}

// SaveToolUsage records received cumulative usage without adopting tool output.
func (s *Store) SaveToolUsage(ctx context.Context, sessionID, turnID, occurrence string, usage goai.Usage, unknown, returned bool) error {
	if ToolBatchID(ctx, turnID) == "" {
		return nil
	}
	if occurrence == "" {
		return fmt.Errorf("missing tool usage occurrence")
	}
	if err := ValidateReceiptUsage(&usage); err != nil {
		return err
	}
	payload := toolBatchEvent(ctx, turnID, map[string]any{"phase": "tool", "checkpoint": true, "occurrence_id": occurrence, "received_usage": usage, "usage_unknown": unknown, "callback_returned": returned})
	return s.AppendTurnEffectEvent(ctx, turnID, sessionID, "tool.usage_received", payload)
}

// HasToolBatchCheckpoint identifies a candidate for the runtime-ready barrier;
// presence is not validation or permission to replay it.
func (s *Store) HasToolBatchCheckpoint(ctx context.Context, sessionID, turnID string) (bool, error) {
	var present bool
	err := s.db.QueryRowContext(ctx, `select exists(select 1 from turn_events where session_id=? and turn_id=? and event_type='tool.batch_ready')`, sessionID, turnID).Scan(&present)
	return present, err
}
