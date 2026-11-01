package store

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestProviderRetryLifecycleValidationAndDetachment(t *testing.T) {
	valid := &ProviderLifecyclePolicy{Version: 1, Messages: json.RawMessage(`[{"role":"user","content":[{"type":"text","text":"policy"}]}]`), Tools: json.RawMessage(`[{"name":"owned"}]`), RegisteredTools: map[string]bool{"owned": true}}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	clone := valid.Clone()
	clone.Messages[0] = '!'
	clone.Tools[0] = '!'
	clone.RegisteredTools["owned"] = false
	if err := valid.Validate(); err != nil || !valid.RegisteredTools["owned"] {
		t.Fatal("clone retained caller authority", valid, err)
	}
	for _, bad := range []*ProviderLifecyclePolicy{nil, {Version: 2}, {Version: 1, Messages: json.RawMessage(`null`)}, {Version: 1, Messages: json.RawMessage(`{}`)}, {Version: 1, Tools: json.RawMessage(`[{"name":"missing"}]`)}, {Version: 1, Tools: json.RawMessage(`[]`), RegisteredTools: map[string]bool{"extra": true}}} {
		if err := bad.Validate(); err == nil {
			t.Fatal("invalid lifecycle accepted", bad)
		}
	}
}

func TestProviderRetryRecoveryReopenDeadlineAndFence(t *testing.T) {
	s, oldCtx, path := retryReceiptFixture(t)
	receipt := retryReceipt()
	receipt.RetryWait.RetryAt = time.Now().Add(time.Hour)
	receipt.RetryWait.Lifecycle = &ProviderLifecyclePolicy{Version: 1, Message: "once only", Tools: json.RawMessage(`[]`)}
	result, err := s.SettleProviderAttempt(oldCtx, "t", "s", "request", receipt)
	if err != nil {
		t.Fatal(err)
	}
	result.RetryWait.Lifecycle.Message = "mutated output"
	if receipt.RetryWait.Lifecycle.Message != "once only" {
		t.Fatal("returned policy aliases input")
	}
	if _, err = s.db.Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	claims, err := reopened.ListStaleActiveTurnClaims(context.Background(), time.Now(), "s")
	if err != nil || len(claims) != 1 {
		t.Fatal(claims, err)
	}
	checkpoint, err := reopened.AcquireProviderRetryRecovery(context.Background(), claims[0], "new-epoch", time.Now())
	if err != nil || checkpoint == nil {
		t.Fatal(checkpoint, err)
	}
	if checkpoint.OccurrenceID != "request" || !checkpoint.Wait.RetryAt.Equal(receipt.RetryWait.RetryAt) || checkpoint.Wait.Lifecycle.Message != "once only" {
		t.Fatal("checkpoint changed across reopen", checkpoint)
	}
	if err = reopened.ResumeProviderRetry(oldCtx, "s", "t", "request"); err == nil {
		t.Fatal("old incarnation woke checkpoint")
	}
	current := WithTurnWriter(context.Background(), "s", "t", "new-epoch")
	if err = reopened.ResumeProviderRetry(current, "s", "t", "request"); err == nil {
		t.Fatal("recovery bypassed absolute deadline")
	}
	row, err := reopened.GetTurn(current, "t")
	if err != nil || row.Phase != "retry_wait" || row.Status != "running" {
		t.Fatal(row, err)
	}
	if err = reopened.ReleaseSessionActiveTurn(current, "s", "new-epoch"); err != nil {
		t.Fatal(err)
	}
	if err = reopened.AppendTurnEffectEvent(oldCtx, "t", "s", "tool.output", nil); err == nil {
		t.Fatal("old writer revived after release")
	}
}

func TestProviderRetryRecoveryRejectsInvalidOrChangedState(t *testing.T) {
	for _, state := range []string{"legacy", "live", "refreshed", "replaced", "later_effect", "invalid_policy", "cancelled", "fault_claim", "fault_fence", "fault_event"} {
		t.Run(state, func(t *testing.T) {
			s, ctx, _ := retryReceiptFixture(t)
			receipt := retryReceipt()
			receipt.RetryWait.Lifecycle = &ProviderLifecyclePolicy{Version: 1}
			if state == "legacy" {
				receipt.RetryWait.Lifecycle = nil
			}
			if _, err := s.SettleProviderAttempt(ctx, "t", "s", "request", receipt); err != nil {
				t.Fatal(err)
			}
			if _, err := s.db.Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
				t.Fatal(err)
			}
			claims, err := s.ListStaleActiveTurnClaims(context.Background(), time.Now(), "s")
			if err != nil || len(claims) != 1 {
				t.Fatal(claims, err)
			}
			claim := claims[0]
			cutoff := time.Now()
			switch state {
			case "live":
				cutoff = time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)
			case "refreshed":
				if err = s.TouchSessionActiveTurn(ctx, "s", "epoch"); err != nil {
					t.Fatal(err)
				}
			case "replaced":
				_, err = s.db.Exec(`update session_active_turns set claim_token='foreign'`)
			case "later_effect":
				err = s.AppendTurnEvent(ctx, "t", "s", "tool.started", nil)
			case "invalid_policy":
				_, err = s.db.Exec(`update turn_receipts set decision_json=json_set(decision_json,'$.retry_wait.lifecycle.version',9)`)
			case "cancelled":
				err = s.UpdateTurnStatusAndPhase(ctx, "t", "cancelling", "cancelling")
			case "fault_claim":
				_, err = s.db.Exec(`create trigger reject_recovery before update on session_active_turns begin select raise(abort,'claim rejected');end`)
			case "fault_fence":
				_, err = s.db.Exec(`create trigger reject_recovery before insert on turn_recovery_fences begin select raise(abort,'fence rejected');end`)
			case "fault_event":
				_, err = s.db.Exec(`create trigger reject_recovery before insert on turn_events when new.event_type='turn.recovery_admitted' begin select raise(abort,'event rejected');end`)
			}
			if err != nil {
				t.Fatal(err)
			}
			checkpoint, err := s.AcquireProviderRetryRecovery(context.Background(), claim, "new", cutoff)
			if checkpoint != nil {
				t.Fatal("invalid checkpoint recovered", state, checkpoint)
			}
			if state != "legacy" && state != "cancelled" && state != "invalid_policy" && err == nil {
				t.Fatal("invalid ownership accepted", state)
			}
			_, token, readErr := s.GetSessionActiveTurn(context.Background(), "s")
			want := "epoch"
			if state == "replaced" {
				want = "foreign"
			}
			if readErr != nil || token != want {
				t.Fatal("partial takeover escaped rollback", token, readErr)
			}
			var fences int
			if err = s.db.QueryRow(`select count(*) from turn_recovery_fences`).Scan(&fences); err != nil || fences != 0 {
				t.Fatal("fence escaped rollback", fences, err)
			}
		})
	}
}

func TestProviderRetryRecoveryRejectsCrossSession(t *testing.T) {
	s, ctx, _ := retryReceiptFixture(t)
	receipt := retryReceipt()
	receipt.RetryWait.Lifecycle = &ProviderLifecyclePolicy{Version: 1}
	if _, err := s.SettleProviderAttempt(ctx, "t", "s", "request", receipt); err != nil {
		t.Fatal(err)
	}
	_, err := s.AcquireProviderRetryRecovery(context.Background(), ActiveTurnClaim{SessionID: "other", TurnID: "t", ClaimToken: "epoch"}, "new", time.Now())
	if err == nil {
		t.Fatal("cross-session takeover accepted")
	}
}
