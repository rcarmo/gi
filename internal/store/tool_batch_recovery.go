package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	goai "github.com/rcarmo/go-ai"
)

// ToolBatchCall separates a wire call from its already-approved execution.
// ReplayPending means a previous physical attempt has an uncertain receipt,
// not that its executor or preparation hooks may run without revalidation.
type ToolBatchCall struct {
	Original             goai.ToolCall
	Prepared             goai.ToolCall
	Preparing            bool
	OccurrenceID         string
	ReplaySafe           bool
	CompletesTurn        bool
	ResultHookStarted    bool
	ReceivedUsage        *goai.Usage
	ReceivedUsageUnknown bool
	CallbackReturned     bool
	CommittedOutput      string
	OutputTruncated      bool
	Done                 bool
	ReplayPending        bool
	ResultEvent          string
	FailureSignature     string
	RepeatedFailures     int
}

type ToolBatchRecovery struct {
	Checkpoint *ToolBatchCheckpoint
	Calls      []ToolBatchCall
	Hooks      map[string]bool
}

type toolBatchEventHeader struct {
	BatchID    string `json:"batch_id"`
	Index      *int   `json:"batch_index"`
	Iteration  int    `json:"iteration"`
	Phase      string `json:"phase"`
	Checkpoint bool   `json:"checkpoint"`
	Decision   *struct {
		Status  string         `json:"status"`
		Phase   string         `json:"phase"`
		Payload map[string]any `json:"payload"`
	} `json:"decision"`
	Key                  string          `json:"key"`
	Stage                string          `json:"stage"`
	Reason               string          `json:"reason"`
	ReceivedUsage        *goai.Usage     `json:"received_usage"`
	UsageUnknown         *bool           `json:"usage_unknown"`
	CallbackReturned     bool            `json:"callback_returned"`
	Final                *bool           `json:"final"`
	OutputPreview        *string         `json:"output_preview"`
	OutputTruncated      bool            `json:"output_truncated"`
	Tool                 string          `json:"tool"`
	ToolCallID           string          `json:"tool_call_id"`
	OccurrenceID         string          `json:"occurrence_id"`
	OriginalCall         json.RawMessage `json:"original_call"`
	PreparedCall         json.RawMessage `json:"prepared_call"`
	ReplaySafe           *bool           `json:"replay_safe"`
	CompletesTurn        *bool           `json:"completes_turn"`
	ReplayPending        bool            `json:"replay_pending"`
	Unadmitted           bool            `json:"unadmitted"`
	PreviousOccurrenceID string          `json:"previous_occurrence_id"`
	FailureSignature     *string         `json:"failure_signature"`
	RepeatedFailures     *int            `json:"repeated_tool_failures"`
}

func decodeToolCall(raw json.RawMessage) (goai.ToolCall, error) {
	var call goai.ToolCall
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	err := decoder.Decode(&call)
	return call, err
}

// ContentBlock omits empty arguments in provider-message JSON. ToolCall does
// not. Compare their empty objects semantically without changing prepared calls.
func sameBatchToolCall(a, b goai.ToolCall) bool {
	if len(a.Arguments) == 0 && len(b.Arguments) == 0 {
		a.Arguments, b.Arguments = nil, nil
	}
	return reflect.DeepEqual(a, b)
}

// toolBatchCheckpointTx accepts only the append-only native batch grammar.
// A phase or an isolated tool intent is not sufficient continuation evidence.
func toolBatchCheckpointTx(ctx context.Context, tx *sql.Tx, sessionID, turnID string) (*ToolBatchRecovery, error) {
	return readToolBatchCheckpointTx(ctx, tx, sessionID, turnID, false)
}

func readToolBatchCheckpointTx(ctx context.Context, tx *sql.Tx, sessionID, turnID string, terminalCleanup bool) (*ToolBatchRecovery, error) {
	var seq int64
	var raw, owner string
	err := tx.QueryRowContext(ctx, `select seq,session_id,payload_json from turn_events where turn_id=? and event_type='tool.batch_ready' order by seq desc limit 1`, turnID).Scan(&seq, &owner, &raw)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	checkpoint, err := decodeToolBatch(raw)
	if err != nil || owner != sessionID {
		return nil, nil
	}
	if err = checkToolBatchReceiptTx(ctx, tx, sessionID, turnID, checkpoint); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var held int
	if err = tx.QueryRowContext(ctx, `select count(*) from turn_failures where turn_id=? and (hold_state<>'none' or (coalesce(resolution_state,'')<>'' and not (? and resolution_state='skipped')))`, turnID, terminalCleanup).Scan(&held); err != nil {
		return nil, err
	}
	if held != 0 {
		return nil, nil
	}
	var payload string
	if err = tx.QueryRowContext(ctx, `select m.payload_json from turn_receipts r join messages m on m.id=r.message_id and m.session_id=? where r.turn_id=? and r.kind='provider' and r.occurrence_id=? and m.role='assistant'`, sessionID, turnID, checkpoint.ProviderOccurrenceID).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var saved struct {
		Message *goai.Message `json:"provider_message"`
	}
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	if decoder.Decode(&saved) != nil || saved.Message == nil || saved.Message.Role != "assistant" {
		return nil, nil
	}
	calls := goai.GetToolCalls(saved.Message)
	if len(calls) != len(checkpoint.CallIDs) {
		return nil, nil
	}
	state := &ToolBatchRecovery{Checkpoint: checkpoint, Calls: make([]ToolBatchCall, len(calls)), Hooks: map[string]bool{}}
	for i, call := range calls {
		if call.ID != checkpoint.CallIDs[i] || strings.TrimSpace(call.Name) == "" {
			return nil, nil
		}
		state.Calls[i].Original = call
	}
	rows, err := tx.QueryContext(ctx, `select event_type,session_id,payload_json from turn_events where turn_id=? and seq>? order by seq`, turnID, seq)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type expectedReceipt struct {
		kind         string
		message      bool
		index        int
		usage        *goai.Usage
		usageUnknown bool
		returned     bool
	}
	expected := map[string]expectedReceipt{}
	admitted := map[string]bool{}
	physicalSlots := map[string]int{}
	cursor := 0
	parked := 0
	terminalReason := "checkpoint_cancel"
	cancelling := false
	budgetSignature, budgetFailures := checkpoint.Generation.LastToolFailureSignature, checkpoint.Generation.RepeatedToolFailures
	for rows.Next() {
		var kind, session, body string
		if err = rows.Scan(&kind, &session, &body); err != nil {
			return nil, err
		}
		if session != sessionID {
			return nil, nil
		}
		if cancelling && kind != "turn.terminal_staged" && kind != "turn.recovery_admitted" && kind != "turn.cleanup_handoff_failed" && kind != "tool.interrupted" {
			return nil, nil
		}
		if control, valid := checkpointRetryControl(kind, body, "waiting_on_tools", &parked); control {
			if !valid {
				return nil, nil
			}
			continue
		}
		if parked != 0 {
			return nil, nil
		}
		var h toolBatchEventHeader
		if json.Unmarshal([]byte(body), &h) != nil {
			return nil, nil
		}
		if kind == "turn.terminal_staged" {
			if !terminalCleanup || h.Phase != "settling" || !h.Checkpoint || h.Decision == nil {
				return nil, nil
			}
			skip := h.Decision.Status == "failed" && h.Decision.Phase == "failed" && h.Decision.Payload["checkpoint_skip"] == ClosedToolBatchPhase
			cancel := h.Decision.Status == "cancelled" && h.Decision.Phase == "cancelled" && h.Decision.Payload["checkpoint_cancellation"] == ClosedToolBatchPhase
			if !skip && !cancel {
				return nil, nil
			}
			if skip {
				terminalReason = "checkpoint_skip"
			}
			cancelling = true
			continue
		}
		if kind == "turn.recovery_admitted" && cancelling {
			if h.Phase != "recovery" || !h.Checkpoint || h.Reason != "resume_terminal_decision" {
				return nil, nil
			}
			continue
		}
		if kind == "turn.cleanup_handoff_failed" && cancelling {
			if h.Phase != "cleanup" || !h.Checkpoint || h.Stage != "terminal_settlement_pending" {
				return nil, nil
			}
			continue
		}
		if h.BatchID != checkpoint.ProviderOccurrenceID || h.Iteration != checkpoint.Generation.Iteration {
			return nil, nil
		}
		if kind == "tool.output" {
			index, ok := physicalSlots[h.OccurrenceID]
			if !ok || h.Index == nil || *h.Index != index || h.ToolCallID != state.Calls[index].Prepared.ID || h.Final == nil {
				return nil, nil
			}
			if h.OutputPreview == nil || len(*h.OutputPreview) > 3*(12*1024+6) {
				return nil, nil
			}
			call := &state.Calls[index]
			// Before executor return these are committed SDK reports, not the
			// returned result awaiting policy. Retain only that published tail.
			if call.OccurrenceID == h.OccurrenceID && !call.CallbackReturned {
				call.CommittedOutput, call.OutputTruncated = *h.OutputPreview, h.OutputTruncated
			}
			// Partial output never authorises success or resets the crash cap.
			continue
		}
		switch kind {
		case "turn.checkpoint_queued":
			if h.Phase != ClosedToolBatchPhase || !h.Checkpoint || h.Reason != "graceful_close" {
				return nil, nil
			}
			continue
		case "tool.batch_recovered", "tool.batch_resumed":
			if h.Phase != "tool_batch" || !h.Checkpoint {
				return nil, nil
			}
			continue
		case "tool.batch_hook_started":
			if h.Phase != "tool_batch" || !h.Checkpoint || state.Hooks[h.Key] {
				return nil, nil
			}
			switch h.Key {
			case "execution_start":
				if cursor != 0 || state.Calls[0].Preparing || state.Hooks["execution_end"] {
					return nil, nil
				}
			case "execution_end":
				if !state.Hooks["execution_start"] || cursor != len(calls) {
					return nil, nil
				}
			case "turn_end":
				if cursor != len(calls) {
					return nil, nil
				}
			default:
				index, e := strconv.Atoi(strings.TrimPrefix(h.Key, "result:"))
				if e != nil || !strings.HasPrefix(h.Key, "result:") || h.Index == nil || *h.Index != index || index != cursor || index >= len(calls) {
					return nil, nil
				}
				call := &state.Calls[index]
				if call.OccurrenceID == "" || call.Done || call.ReplayPending || call.ResultHookStarted || state.Hooks["execution_end"] {
					return nil, nil
				}
				call.ResultHookStarted = true
			}
			state.Hooks[h.Key] = true
			continue
		}
		if h.Index == nil || *h.Index != cursor || cursor >= len(calls) || state.Hooks["turn_end"] {
			return nil, nil
		}
		call := &state.Calls[cursor]
		if h.FailureSignature == nil || h.RepeatedFailures == nil || *h.RepeatedFailures < 0 {
			return nil, nil
		}
		switch kind {
		case "tool.usage_received":
			if h.Phase != "tool" || !h.Checkpoint || !call.Preparing || call.ReplayPending || call.OccurrenceID == "" || h.OccurrenceID != call.OccurrenceID || h.ReceivedUsage == nil || h.UsageUnknown == nil || ValidateReceiptUsage(h.ReceivedUsage) != nil || !toolUsageExtends(h.ReceivedUsage, call.ReceivedUsage) || (call.ReceivedUsageUnknown && !*h.UsageUnknown) || (call.CallbackReturned && !h.CallbackReturned) || *h.FailureSignature != budgetSignature || *h.RepeatedFailures != budgetFailures {
				return nil, nil
			}
			call.ReceivedUsage, call.ReceivedUsageUnknown, call.CallbackReturned = h.ReceivedUsage, *h.UsageUnknown, h.CallbackReturned
		case "tool.preparing":
			if *h.FailureSignature != budgetSignature || *h.RepeatedFailures != budgetFailures {
				return nil, nil
			}
			if h.Phase != "tool" || !h.Checkpoint || call.Preparing || h.ToolCallID != call.Original.ID || state.Hooks["execution_end"] {
				return nil, nil
			}
			call.Preparing = true
		case "tool.started":
			if *h.FailureSignature != budgetSignature || *h.RepeatedFailures != budgetFailures {
				return nil, nil
			}
			if h.Phase != "tool" || !h.Checkpoint || !call.Preparing || (call.OccurrenceID != "" && !call.ReplayPending) || h.OccurrenceID == "" || admitted[h.OccurrenceID] || h.ReplaySafe == nil || h.CompletesTurn == nil || state.Hooks["execution_end"] {
				return nil, nil
			}
			original, e := decodeToolCall(h.OriginalCall)
			if e != nil || !sameBatchToolCall(original, call.Original) {
				return nil, nil
			}
			prepared, e := decodeToolCall(h.PreparedCall)
			if e != nil || prepared.Name != h.Tool || prepared.ID != h.ToolCallID {
				return nil, nil
			}
			if call.OccurrenceID != "" && (!reflect.DeepEqual(prepared, call.Prepared) || !*h.ReplaySafe || call.ResultHookStarted) {
				return nil, nil
			}
			call.Prepared = prepared
			call.OccurrenceID = h.OccurrenceID
			call.ReplaySafe = *h.ReplaySafe
			call.CompletesTurn = *h.CompletesTurn
			call.ReplayPending = false
			call.ReceivedUsage, call.ReceivedUsageUnknown, call.CallbackReturned = nil, false, false
			call.CommittedOutput, call.OutputTruncated = "", false
			admitted[h.OccurrenceID] = true
			physicalSlots[h.OccurrenceID] = cursor
		case "tool.finished", "tool.failed", "tool.skipped", "tool.interrupted":
			cancelUnstarted := cancelling && kind == "tool.interrupted" && h.Reason == terminalReason && h.Unadmitted && call.OccurrenceID == ""
			if (!call.Preparing && !cancelUnstarted) || h.OccurrenceID == "" || (cancelling && h.Reason != terminalReason) {
				return nil, nil
			}
			if kind == "tool.interrupted" && h.Unadmitted {
				if h.ReplayPending || admitted[h.OccurrenceID] || (call.OccurrenceID == "" && h.PreviousOccurrenceID != "") || (call.OccurrenceID != "" && (!call.ReplayPending || h.PreviousOccurrenceID != call.OccurrenceID)) {
					return nil, nil
				}
			} else if call.OccurrenceID != "" {
				if h.OccurrenceID != call.OccurrenceID || call.ReplayPending {
					return nil, nil
				}
			} else if kind != "tool.skipped" && kind != "tool.finished" && !(kind == "tool.failed" && h.Reason == "invalid_arguments") {
				return nil, nil
			}
			if _, duplicate := expected[h.OccurrenceID]; duplicate {
				return nil, nil
			}
			switch kind {
			case "tool.finished":
				if *h.RepeatedFailures != 0 || *h.FailureSignature != "" {
					return nil, nil
				}
			case "tool.failed":
				if call.OccurrenceID == "" && h.Reason == "invalid_arguments" {
					if *h.FailureSignature != budgetSignature || *h.RepeatedFailures != budgetFailures {
						return nil, nil
					}
					break
				}
				next := 1
				if *h.FailureSignature == budgetSignature {
					next = budgetFailures + 1
				}
				if *h.FailureSignature == "" || *h.RepeatedFailures != next {
					return nil, nil
				}
			default:
				if *h.FailureSignature != budgetSignature || *h.RepeatedFailures != budgetFailures {
					return nil, nil
				}
			}
			budgetSignature, budgetFailures = *h.FailureSignature, *h.RepeatedFailures
			expected[h.OccurrenceID] = expectedReceipt{kind: kind, message: !h.ReplayPending, index: cursor, usage: call.ReceivedUsage, usageUnknown: call.ReceivedUsageUnknown, returned: call.CallbackReturned}
			if !admitted[h.OccurrenceID] {
				receipt := expected[h.OccurrenceID]
				receipt.usage, receipt.usageUnknown, receipt.returned = nil, false, false
				expected[h.OccurrenceID] = receipt
			}
			call.FailureSignature = *h.FailureSignature
			call.RepeatedFailures = *h.RepeatedFailures
			if kind == "tool.interrupted" && h.ReplayPending {
				if !call.ReplaySafe || call.ResultHookStarted {
					return nil, nil
				}
				call.ReplayPending = true
			} else {
				call.Done = true
				call.ReplayPending = false
				call.ResultEvent = kind
				cursor++
			}
		default:
			// Cancellation, abort, later providers, retries, foreign or future stages
			// require their own terminal/continuation policy; never infer progress.
			return nil, nil
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	if terminalCleanup && !cancelling {
		return nil, nil
	}
	if parked != 0 {
		return nil, nil
	}
	receiptRows, err := tx.QueryContext(ctx, `select r.occurrence_id,r.kind,r.decision_json,r.message_id,m.session_id,m.role,json_extract(m.payload_json,'$.turn_id'),r.usage_json,r.usage_unknown from turn_receipts r left join messages m on m.id=r.message_id where r.turn_id=?`, turnID)
	if err != nil {
		return nil, err
	}
	defer receiptRows.Close()
	for receiptRows.Next() {
		var occurrence, kind, decision string
		var message, messageSession, messageRole, messageTurn, usageRaw sql.NullString
		var usageUnknown bool
		if err = receiptRows.Scan(&occurrence, &kind, &decision, &message, &messageSession, &messageRole, &messageTurn, &usageRaw, &usageUnknown); err != nil {
			return nil, err
		}
		expectedReceipt, wanted := expected[occurrence]
		if !wanted {
			if admitted[occurrence] {
				return nil, nil
			}
			continue
		}
		var saved struct {
			Event string `json:"event_type"`
		}
		if kind != "tool" || json.Unmarshal([]byte(decision), &saved) != nil || saved.Event != expectedReceipt.kind || (message.Valid && message.String != "") != expectedReceipt.message {
			return nil, nil
		}
		if expectedReceipt.message && (!messageSession.Valid || messageSession.String != sessionID || !messageTurn.Valid || messageTurn.String != turnID || (messageRole.String != "tool_result" && !(messageRole.String == "assistant" && state.Calls[expectedReceipt.index].CompletesTurn))) {
			return nil, nil
		}
		if expectedReceipt.usage != nil {
			var usage goai.Usage
			if !usageRaw.Valid || json.Unmarshal([]byte(usageRaw.String), &usage) != nil || ValidateReceiptUsage(&usage) != nil || !toolUsageExtends(&usage, expectedReceipt.usage) {
				return nil, nil
			}
			if expectedReceipt.returned && (!reflect.DeepEqual(usage, *expectedReceipt.usage) || usageUnknown != expectedReceipt.usageUnknown) {
				return nil, nil
			}
		}
		if expectedReceipt.kind == "tool.interrupted" && admitted[occurrence] && !expectedReceipt.returned && !usageUnknown {
			return nil, nil
		}
		delete(expected, occurrence)
	}
	if err = receiptRows.Err(); err != nil {
		return nil, err
	}
	if len(expected) != 0 {
		return nil, nil
	}
	return state, nil
}

// AcquireToolBatchRecovery fences only a proven native per-call frontier. It
// does not settle unknown tools or grant replay permission; the current runtime
// must separately validate the pinned execution and both safety declarations.
func (s *Store) AcquireToolBatchRecovery(ctx context.Context, claim ActiveTurnClaim, token string, olderThan time.Time) (*ToolBatchRecovery, error) {
	if token == "" || token == claim.ClaimToken {
		return nil, fmt.Errorf("tool recovery requires a fresh claim token")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// Lock before the reader: required children may still be settling in WAL.
	locked, err := tx.ExecContext(ctx, `update session_active_turns set updated_at=updated_at where session_id=? and turn_id=? and claim_token=? and updated_at=?`, claim.SessionID, claim.TurnID, claim.ClaimToken, claim.UpdatedAt)
	if err != nil {
		return nil, err
	}
	count, err := locked.RowsAffected()
	if err != nil {
		return nil, err
	}
	if count != 1 {
		return nil, ErrQueueConflict
	}
	var status, phase, updated, worker string
	if err = tx.QueryRowContext(ctx, `select t.status,t.phase,sat.updated_at,coalesce(sat.worker_id,'') from session_active_turns sat join turns t on t.id=sat.turn_id and t.session_id=sat.session_id where sat.session_id=? and sat.turn_id=?`, claim.SessionID, claim.TurnID).Scan(&status, &phase, &updated, &worker); err != nil {
		return nil, err
	}
	stamp, err := time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return nil, err
	}
	closed := status == "queued" && phase == ClosedToolBatchPhase && worker == "closed-checkpoint"
	if !closed && (!stamp.Before(olderThan) || status != "running" || phase != "waiting_on_tools") {
		return nil, nil
	}
	state, err := toolBatchCheckpointTx(ctx, tx, claim.SessionID, claim.TurnID)
	if err != nil || state == nil {
		return nil, err
	}
	var recoveries int
	if err = tx.QueryRowContext(ctx, `select count(*) from turn_events where turn_id=? and event_type in ('generation.recovered','tool.batch_recovered') and seq>coalesce((select max(seq) from turn_events where turn_id=? and event_type in ('inference.settled','tool.finished','tool.failed','tool.skipped','turn.checkpoint_retry_requested')),0)`, claim.TurnID, claim.TurnID).Scan(&recoveries); err != nil {
		return nil, err
	}
	if !closed && recoveries >= generationReadyRecoveryLimit {
		return nil, nil
	}
	if _, err = tx.ExecContext(ctx, `update session_active_turns set claim_token=?,worker_id='recovery',claimed_at=`+defaultNow+`,updated_at=`+defaultNow+` where session_id=? and turn_id=?`, token, claim.SessionID, claim.TurnID); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `insert into turn_recovery_fences(turn_id,session_id,claim_token) values(?,?,?) on conflict(turn_id) do update set claim_token=excluded.claim_token`, claim.TurnID, claim.SessionID, token); err != nil {
		return nil, err
	}
	eventType := "tool.batch_recovered"
	if closed {
		eventType = "tool.batch_resumed"
		if _, err = tx.ExecContext(ctx, `update turns set status='running',phase='waiting_on_tools',updated_at=`+defaultNow+` where id=? and session_id=?`, claim.TurnID, claim.SessionID); err != nil {
			return nil, err
		}
		if err = retryCheckpointSessionStateTx(ctx, tx, claim.SessionID, "running", claim.TurnID); err != nil {
			return nil, err
		}
	}
	if err = receiptEventTx(ctx, tx, claim.TurnID, claim.SessionID, eventType, map[string]any{"phase": "tool_batch", "checkpoint": true, "batch_id": state.Checkpoint.ProviderOccurrenceID, "iteration": state.Checkpoint.Generation.Iteration, "previous_claim": claim.ClaimToken}); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return state, nil
}

func toolUsageExtends(next, previous *goai.Usage) bool {
	if next == nil {
		return false
	}
	if previous == nil {
		return true
	}
	for _, pair := range [][2]int{{next.Input, previous.Input}, {next.Output, previous.Output}, {next.TotalTokens, previous.TotalTokens}, {next.CacheRead, previous.CacheRead}, {next.CacheWrite, previous.CacheWrite}, {next.CacheWrite1h, previous.CacheWrite1h}, {next.Reasoning, previous.Reasoning}} {
		if pair[0] < pair[1] {
			return false
		}
	}
	for _, pair := range [][2]float64{{next.Cost.Input, previous.Cost.Input}, {next.Cost.Output, previous.Cost.Output}, {next.Cost.CacheRead, previous.Cost.CacheRead}, {next.Cost.CacheWrite, previous.Cost.CacheWrite}, {next.Cost.Total, previous.Cost.Total}} {
		if pair[0] < pair[1] {
			return false
		}
	}
	return true
}
