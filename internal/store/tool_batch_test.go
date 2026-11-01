package store

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	goai "github.com/rcarmo/go-ai"
)

func toolBatchFixture(t *testing.T) (*Store, context.Context, *ToolBatchCheckpoint, goai.ToolCall) {
	t.Helper()
	s, ctx, _ := generationReadyFixture(t)
	generation := readyCheckpoint()
	generation.LastToolFailureSignature, generation.RepeatedToolFailures = "", 0
	var err error
	if err = s.SaveGenerationReady(ctx, "s", "t", generation); err != nil {
		t.Fatal(err)
	}
	call := goai.ToolCall{Type: "toolCall", ID: "wire-call", Name: "probe", Arguments: map[string]any{"value": json.Number("9007199254740993")}}
	if err = s.AdmitTurnEffect(ctx, "s", "t", "running", "inference.started", map[string]any{"request_id": "batch-provider", "phase": "inference", "iteration": 2, "checkpoint": true}); err != nil {
		t.Fatal(err)
	}
	_, err = s.SettleProviderAttempt(ctx, "t", "s", "batch-provider", ProviderReceipt{MessageID: "batch-message", Iteration: 2, Decision: "tools", Usage: &goai.Usage{Input: 3}, Payload: map[string]any{
		"kind": "tool_calls", "turn_id": "t", "tool_calls": []map[string]any{{"id": call.ID, "name": call.Name, "arguments": call.Arguments}},
		"provider_message": &goai.Message{Role: "assistant", Content: []goai.ContentBlock{{Type: "toolCall", ID: call.ID, Name: call.Name, Arguments: call.Arguments}}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := &ToolBatchCheckpoint{Version: 1, ProviderOccurrenceID: "batch-provider", CallIDs: []string{call.ID}, Generation: generation}
	if err = s.SaveToolBatch(ctx, "s", "t", checkpoint); err != nil {
		t.Fatal(err)
	}
	ctx = WithToolBatchSlot(ctx, "t", "batch-provider", 0, 2, "", 0)
	return s, ctx, checkpoint, call
}

func readToolBatch(t *testing.T, s *Store) *ToolBatchRecovery {
	t.Helper()
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	state, err := toolBatchCheckpointTx(context.Background(), tx, "s", "t")
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func startToolBatchCall(t *testing.T, s *Store, ctx context.Context, call goai.ToolCall) {
	t.Helper()
	if run, err := s.ClaimToolBatchHook(ctx, "s", "t", "batch-provider", "execution_start"); err != nil || !run {
		t.Fatalf("start hook = %v, %v", run, err)
	}
	if err := s.AdmitTurnEffect(ctx, "s", "t", "waiting_on_tools", "tool.preparing", map[string]any{"phase": "tool", "checkpoint": true, "tool_call_id": call.ID}); err != nil {
		t.Fatal(err)
	}
	admitToolBatchCall(t, s, ctx, call, call, "physical-1")
}

func admitToolBatchCall(t *testing.T, s *Store, ctx context.Context, original, prepared goai.ToolCall, occurrence string) {
	t.Helper()
	if err := s.AdmitTurnEffect(ctx, "s", "t", "waiting_on_tools", "tool.started", map[string]any{"phase": "tool", "checkpoint": true, "tool": prepared.Name, "tool_call_id": prepared.ID, "occurrence_id": occurrence, "original_call": original, "prepared_call": prepared, "replay_safe": true, "completes_turn": false}); err != nil {
		t.Fatal(err)
	}
}

func TestToolBatchFramingAndOwnedPreparedArguments(t *testing.T) {
	s, ctx, _, call := toolBatchFixture(t)
	if state := readToolBatch(t, s); state == nil || state.Calls[0].Preparing {
		t.Fatalf("untouched batch = %+v", state)
	}
	startToolBatchCall(t, s, ctx, call)
	state := readToolBatch(t, s)
	if state == nil || state.Calls[0].OccurrenceID != "physical-1" || !state.Calls[0].ReplaySafe || state.Calls[0].Prepared.Arguments["value"] != json.Number("9007199254740993") {
		t.Fatalf("prepared state = %+v", state)
	}
	call.Arguments["value"] = "mutated caller value"
	if state = readToolBatch(t, s); state == nil || state.Calls[0].Prepared.Arguments["value"] != json.Number("9007199254740993") {
		t.Fatalf("owned state changed: %+v", state)
	}
}

func TestToolBatchInterruptedReplayPlanAndSettledSuccessor(t *testing.T) {
	s, ctx, _, call := toolBatchFixture(t)
	startToolBatchCall(t, s, ctx, call)
	if err := s.SettleToolResult(ctx, "t", "s", "tool.interrupted", map[string]any{"occurrence_id": "physical-1", "replay_pending": true}, "", "", map[string]any{"usage_unknown": true}); err != nil {
		t.Fatal(err)
	}
	if state := readToolBatch(t, s); state == nil || !state.Calls[0].ReplayPending || state.Calls[0].Done {
		t.Fatalf("replay plan = %+v", state)
	}
	admitToolBatchCall(t, s, ctx, call, call, "physical-2")
	if state := readToolBatch(t, s); state == nil || state.Calls[0].ReplayPending || state.Calls[0].Done || state.Calls[0].OccurrenceID != "physical-2" {
		t.Fatalf("retry admitted without a receipt = %+v", state)
	}
	if err := s.SettleToolResult(ctx, "t", "s", "tool.finished", map[string]any{"occurrence_id": "physical-2", "failure_signature": "", "repeated_tool_failures": 0}, "tool-message", "done", map[string]any{"kind": "tool_result", "turn_id": "t", "tool_call_id": call.ID, "tool_name": call.Name}); err != nil {
		t.Fatal(err)
	}
	state := readToolBatch(t, s)
	if state == nil || !state.Calls[0].Done || state.Calls[0].OccurrenceID != "physical-2" {
		t.Fatalf("settled successor = %+v", state)
	}
	accounting, err := s.TurnAccounting(ctx, "t")
	if err != nil || accounting.Unknown != 1 {
		t.Fatalf("accounting = %+v, %v", accounting, err)
	}
}

func TestToolBatchHookAttemptsAreAtMostOnce(t *testing.T) {
	s, ctx, _, call := toolBatchFixture(t)
	startToolBatchCall(t, s, ctx, call)
	for i := 0; i < 2; i++ {
		run, err := s.ClaimToolBatchHook(ctx, "s", "t", "batch-provider", "result:0")
		if err != nil || run != (i == 0) {
			t.Fatalf("attempt %d = %v, %v", i, run, err)
		}
	}
	state := readToolBatch(t, s)
	if state == nil || !state.Calls[0].ResultHookStarted {
		t.Fatalf("result-policy state = %+v", state)
	}
	if err := s.SettleToolResult(ctx, "t", "s", "tool.interrupted", map[string]any{"occurrence_id": "physical-1", "replay_pending": true}, "", "", map[string]any{"usage_unknown": true}); err != nil {
		t.Fatal(err)
	}
	if state = readToolBatch(t, s); state != nil {
		t.Fatal("uncertain result policy authorised executor replay")
	}
}

func TestToolBatchRejectsUnprovenSuffix(t *testing.T) {
	for _, mode := range []string{"unknown", "foreign", "wrong-slot", "missing-safety", "changed-original", "receipt-missing", "receipt-without-event", "cancelled", "duplicate-preparing"} {
		t.Run(mode, func(t *testing.T) {
			s, ctx, _, call := toolBatchFixture(t)
			startToolBatchCall(t, s, ctx, call)
			var err error
			switch mode {
			case "unknown":
				err = s.AppendTurnEvent(ctx, "t", "s", "tool.future", map[string]any{})
			case "foreign":
				if _, err = s.CreateSession(context.Background(), "foreign", "foreign", nil); err != nil {
					t.Fatal(err)
				}
				_, err = s.db.Exec(`update turn_events set session_id='foreign' where event_type='tool.started'`)
			case "wrong-slot":
				_, err = s.db.Exec(`update turn_events set payload_json=json_set(payload_json,'$.batch_index',1) where event_type='tool.started'`)
			case "missing-safety":
				_, err = s.db.Exec(`update turn_events set payload_json=json_remove(payload_json,'$.replay_safe') where event_type='tool.started'`)
			case "changed-original":
				_, err = s.db.Exec(`update turn_events set payload_json=json_set(payload_json,'$.original_call.arguments.value','changed') where event_type='tool.started'`)
			case "receipt-missing":
				err = s.SettleToolResult(ctx, "t", "s", "tool.finished", map[string]any{"occurrence_id": "physical-1"}, "result", "done", map[string]any{})
				if err == nil {
					_, err = s.db.Exec(`delete from turn_receipts where occurrence_id='physical-1'`)
				}
			case "receipt-without-event":
				err = s.SettleToolResult(ctx, "t", "s", "tool.finished", map[string]any{"occurrence_id": "physical-1"}, "result", "done", map[string]any{})
				if err == nil {
					_, err = s.db.Exec(`delete from turn_events where event_type='tool.finished'`)
				}
			case "cancelled":
				err = s.SettleToolResult(ctx, "t", "s", "tool.cancelled", map[string]any{"occurrence_id": "physical-1"}, "", "", nil)
			case "duplicate-preparing":
				err = s.AdmitTurnEffect(ctx, "s", "t", "waiting_on_tools", "tool.preparing", map[string]any{"phase": "tool", "checkpoint": true, "tool_call_id": call.ID})
			}
			if err != nil {
				t.Fatal(err)
			}
			if readToolBatch(t, s) != nil {
				t.Fatal("unproven batch accepted")
			}
		})
	}
}

func TestToolBatchIntentAndHookFailuresPreventAdmission(t *testing.T) {
	s, ctx, p, _ := toolBatchFixture(t)
	if err := s.SaveToolBatch(ctx, "s", "t", p); err == nil {
		t.Fatal("duplicate batch accepted")
	}
	invalid := WithToolBatchSlot(ctx, "t", "batch-provider", 1, 2, "", 0)
	if err := s.AdmitTurnEffect(invalid, "s", "t", "waiting_on_tools", "tool.preparing", map[string]any{}); err == nil {
		t.Fatal("out-of-range slot accepted")
	}
	if _, err := s.db.Exec(`create trigger deny_batch_hook before insert on turn_events when new.event_type='tool.batch_hook_started' begin select raise(abort,'hook checkpoint unavailable');end;`); err != nil {
		t.Fatal(err)
	}
	if run, err := s.ClaimToolBatchHook(ctx, "s", "t", "batch-provider", "execution_start"); err == nil || run {
		t.Fatalf("failed hook attempt admitted: %v, %v", run, err)
	}
}

func staleToolBatchClaim(t *testing.T, s *Store) ActiveTurnClaim {
	t.Helper()
	if _, err := s.db.Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z' where session_id='s'`); err != nil {
		t.Fatal(err)
	}
	claims, err := s.ListStaleActiveTurnClaims(context.Background(), time.Now(), "s")
	if err != nil || len(claims) != 1 {
		t.Fatal(claims, err)
	}
	return claims[0]
}

func TestToolBatchAcquisitionFencesWithoutSettlingOrDispatching(t *testing.T) {
	s, old, _, call := toolBatchFixture(t)
	startToolBatchCall(t, s, old, call)
	claim := staleToolBatchClaim(t, s)
	state, err := s.AcquireToolBatchRecovery(context.Background(), claim, "new-tool-owner", time.Now())
	if err != nil || state == nil || state.Calls[0].Done || state.Calls[0].ReplayPending {
		t.Fatalf("acquired = %+v, %v", state, err)
	}
	if err = s.AdmitTurnEffect(old, "s", "t", "running", "inference.started", map[string]any{"request_id": "stale"}); err == nil {
		t.Fatal("old owner retained authority")
	}
	var receipts int
	if err = s.db.QueryRow(`select count(*) from turn_receipts where occurrence_id='physical-1'`).Scan(&receipts); err != nil || receipts != 0 {
		t.Fatal("acquisition fabricated tool outcome", receipts, err)
	}
	if again, err := s.AcquireToolBatchRecovery(context.Background(), claim, "other", time.Now()); err == nil || again != nil {
		t.Fatalf("stale claim acquired twice: %+v, %v", again, err)
	}
}

func TestToolBatchAcquisitionIsAtomic(t *testing.T) {
	for _, trigger := range []string{"before update of claim_token on session_active_turns", "before insert on turn_recovery_fences", "before insert on turn_events when new.event_type='tool.batch_recovered'"} {
		t.Run(trigger, func(t *testing.T) {
			s, ctx, _, call := toolBatchFixture(t)
			startToolBatchCall(t, s, ctx, call)
			claim := staleToolBatchClaim(t, s)
			if _, err := s.db.Exec(`create trigger deny_tool_recovery ` + trigger + ` begin select raise(abort,'tool recovery unavailable');end;`); err != nil {
				t.Fatal(err)
			}
			if state, err := s.AcquireToolBatchRecovery(context.Background(), claim, "blocked-owner", time.Now()); err == nil || state != nil {
				t.Fatalf("failed recovery acquired: %+v, %v", state, err)
			}
			var token string
			if err := s.db.QueryRow(`select claim_token from session_active_turns where session_id='s'`).Scan(&token); err != nil || token != claim.ClaimToken {
				t.Fatal("failed recovery changed owner", token, err)
			}
			if readToolBatch(t, s) == nil {
				t.Fatal("rollback damaged checkpoint")
			}
		})
	}
}

func TestToolBatchNoProgressRecoveryLimit(t *testing.T) {
	s, ctx, _, call := toolBatchFixture(t)
	startToolBatchCall(t, s, ctx, call)
	for i := 0; i <= generationReadyRecoveryLimit; i++ {
		claim := staleToolBatchClaim(t, s)
		state, err := s.AcquireToolBatchRecovery(context.Background(), claim, fmt.Sprintf("tool-owner-%d", i), time.Now())
		if err != nil || (state != nil) != (i < generationReadyRecoveryLimit) {
			t.Fatalf("recovery %d = %+v, %v", i, state, err)
		}
	}
}

func TestToolBatchReceivedUsageSurvivesPhysicalReplay(t *testing.T) {
	s, ctx, _, call := toolBatchFixture(t)
	startToolBatchCall(t, s, ctx, call)
	for _, input := range []int{5, 12} {
		if err := s.SaveToolUsage(ctx, "s", "t", "physical-1", goai.Usage{Input: input}, false, false); err != nil {
			t.Fatal(err)
		}
	}
	if state := readToolBatch(t, s); state == nil || state.Calls[0].ReceivedUsage.Input != 12 {
		t.Fatal("received usage missing", state)
	}
	if err := s.SettleToolResult(ctx, "t", "s", "tool.interrupted", map[string]any{"occurrence_id": "physical-1", "replay_pending": true}, "", "", map[string]any{"usage": goai.Usage{Input: 12}, "usage_unknown": true}); err != nil {
		t.Fatal(err)
	}
	admitToolBatchCall(t, s, ctx, call, call, "physical-2")
	if state := readToolBatch(t, s); state == nil || state.Calls[0].ReceivedUsage != nil {
		t.Fatal("usage carried into a different physical attempt", state)
	}
	if err := s.SaveToolUsage(ctx, "s", "t", "physical-2", goai.Usage{Input: 3}, false, true); err != nil {
		t.Fatal(err)
	}
	if err := s.SettleToolResult(ctx, "t", "s", "tool.finished", map[string]any{"occurrence_id": "physical-2"}, "observed-result", "done", map[string]any{"turn_id": "t", "tool_call_id": call.ID, "tool_name": call.Name, "usage": goai.Usage{Input: 3}}); err != nil {
		t.Fatal(err)
	}
	if state := readToolBatch(t, s); state == nil || !state.Calls[0].Done {
		t.Fatal("matching receipts rejected", state)
	}
	accounting, err := s.TurnAccounting(ctx, "t")
	if err != nil || accounting.Usage.Input != 23 || accounting.Unknown != 1 {
		t.Fatal(accounting, err)
	}
}

func TestToolBatchRejectsInvalidReceivedUsageHistory(t *testing.T) {
	for _, kind := range []string{"decreasing", "cleared_uncertainty", "after_return", "foreign_occurrence", "lost_receipt_usage", "false_known_receipt", "changed_complete_receipt"} {
		t.Run(kind, func(t *testing.T) {
			s, ctx, _, call := toolBatchFixture(t)
			startToolBatchCall(t, s, ctx, call)
			unknown, returned := kind == "cleared_uncertainty", kind == "after_return" || kind == "changed_complete_receipt"
			if err := s.SaveToolUsage(ctx, "s", "t", "physical-1", goai.Usage{Input: 5}, unknown, returned); err != nil {
				t.Fatal(err)
			}
			var err error
			switch kind {
			case "decreasing":
				err = s.SaveToolUsage(ctx, "s", "t", "physical-1", goai.Usage{Input: 4}, false, false)
			case "cleared_uncertainty":
				err = s.SaveToolUsage(ctx, "s", "t", "physical-1", goai.Usage{Input: 5}, false, false)
			case "after_return":
				err = s.SaveToolUsage(ctx, "s", "t", "physical-1", goai.Usage{Input: 5}, false, false)
			case "foreign_occurrence":
				err = s.SaveToolUsage(ctx, "s", "t", "foreign", goai.Usage{Input: 5}, false, false)
			default:
				input := 5
				if kind == "lost_receipt_usage" {
					input = 0
				}
				if kind == "changed_complete_receipt" {
					input = 6
				}
				err = s.SettleToolResult(ctx, "t", "s", "tool.interrupted", map[string]any{"occurrence_id": "physical-1", "replay_pending": false}, "bad-usage-result", "uncertain", map[string]any{"turn_id": "t", "tool_call_id": call.ID, "tool_name": call.Name, "usage": goai.Usage{Input: input}, "usage_unknown": kind != "false_known_receipt"})
			}
			if err != nil {
				t.Fatal(err)
			}
			if state := readToolBatch(t, s); state != nil {
				t.Fatal("invalid usage history authorised recovery", state)
			}
		})
	}
}

func TestToolBatchInterruptedProjectionDoesNotInventTiming(t *testing.T) {
	s, ctx, _, call := toolBatchFixture(t)
	startToolBatchCall(t, s, ctx, call)
	if err := s.SettleToolResult(ctx, "t", "s", "tool.interrupted", map[string]any{"phase": "tool", "checkpoint": true, "tool": call.Name, "tool_call_id": call.ID, "occurrence_id": "physical-1", "replay_pending": true}, "", "", map[string]any{"usage_unknown": true}); err != nil {
		t.Fatal(err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	activity, err := latestToolActivity(ctx, tx, "t", true)
	_ = tx.Rollback()
	if err != nil || activity["state"] != "interrupted" || activity["duration_ms"] != nil || activity["finished_at"] != "" {
		t.Fatal("invented execution timing", activity, err)
	}
	results, err := s.ListTerminalToolResults(ctx, "s")
	if err != nil || len(results) != 1 || results[0].Payload["status"] != "interrupted" || results[0].Payload["duration_ms"] != nil {
		t.Fatal("interrupted disposition missing or timed", results, err)
	}
}

func TestToolBatchPendingReplayMayBeDeclined(t *testing.T) {
	for _, previous := range []string{"physical-1", "foreign"} {
		t.Run(previous, func(t *testing.T) {
			s, ctx, _, call := toolBatchFixture(t)
			startToolBatchCall(t, s, ctx, call)
			if err := s.SettleToolResult(ctx, "t", "s", "tool.interrupted", map[string]any{"occurrence_id": "physical-1", "replay_pending": true}, "", "", map[string]any{"usage_unknown": true}); err != nil {
				t.Fatal(err)
			}
			if err := s.SettleToolResult(ctx, "t", "s", "tool.interrupted", map[string]any{"occurrence_id": "logical-interruption", "unadmitted": true, "previous_occurrence_id": previous}, "interrupted-message", "Outcome unknown", map[string]any{"turn_id": "t", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": true}); err != nil {
				t.Fatal(err)
			}
			state := readToolBatch(t, s)
			if previous == "foreign" {
				if state != nil {
					t.Fatal("unrelated previous occurrence accepted")
				}
			} else if state == nil || !state.Calls[0].Done {
				t.Fatalf("declined pending replay = %+v", state)
			}
		})
	}
}

func TestToolBatchRecoveryCapCarriesIntoGeneration(t *testing.T) {
	for _, terminal := range []string{"tool.interrupted", "tool.finished"} {
		t.Run(terminal, func(t *testing.T) {
			s, ctx, batch, call := toolBatchFixture(t)
			startToolBatchCall(t, s, ctx, call)
			var token string
			for i := 0; i < generationReadyRecoveryLimit; i++ {
				token = fmt.Sprintf("tool-owner-%d", i)
				if state, err := s.AcquireToolBatchRecovery(context.Background(), staleToolBatchClaim(t, s), token, time.Now()); err != nil || state == nil {
					t.Fatal(state, err)
				}
			}
			writer := WithTurnWriter(context.Background(), "s", "t", token)
			slot := WithToolBatchSlot(writer, "t", "batch-provider", 0, 2, "", 0)
			if err := s.SettleToolResult(slot, "t", "s", terminal, map[string]any{"occurrence_id": "physical-1"}, "tool-result", "result", map[string]any{"turn_id": "t", "tool_call_id": call.ID, "tool_name": call.Name, "usage_unknown": terminal == "tool.interrupted"}); err != nil {
				t.Fatal(err)
			}
			generation := *batch.Generation
			generation.Iteration++
			if err := s.SaveGenerationReady(writer, "s", "t", &generation); err != nil {
				t.Fatal(err)
			}
			state, err := s.AcquireGenerationReadyRecovery(context.Background(), staleToolBatchClaim(t, s), "next-generation-owner", time.Now())
			if err != nil || (state != nil) != (terminal == "tool.finished") {
				t.Fatalf("%s successor after recovery cap = %+v, %v", terminal, state, err)
			}
		})
	}
}
