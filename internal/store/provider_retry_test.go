package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	goai "github.com/rcarmo/go-ai"
)

func retryReceiptFixture(t *testing.T) (*Store, context.Context, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "retry.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "s", "s", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "t", "s", "running", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "s", "t", "worker", "epoch"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	ctx = WithTurnWriter(ctx, "s", "t", "epoch")
	if err := s.AdmitTurnEffect(ctx, "s", "t", "running", "inference.started", map[string]any{"request_id": "request"}); err != nil {
		t.Fatal(err)
	}
	return s, ctx, path
}
func retryReceipt() ProviderReceipt {
	return ProviderReceipt{Decision: "error_boundary", Error: "HTTP 503", Usage: &goai.Usage{Input: 5, Output: 2, TotalTokens: 7}, Iteration: 1,
		RetryWait: &ProviderRetryWait{Attempt: 1, MaxAttempts: 3, DelayMS: 2000, RetryAt: time.Now().Add(2 * time.Second).UTC(), FailureCategory: "network", Iteration: 1, Model: "fixture/model", Title: "Provider failed — retrying"}}
}
func TestProviderRetryCheckpointIsAtomic(t *testing.T) {
	for _, trigger := range []string{
		`before insert on turn_receipts`,
		`before insert on turn_events when new.event_type='inference.settled'`,
		`before insert on turn_events when new.event_type='inference.retry_scheduled'`,
		`before update of phase on turns when new.phase='retry_wait'`,
	} {
		t.Run(trigger, func(t *testing.T) {
			s, ctx, _ := retryReceiptFixture(t)
			if _, err := s.db.Exec(`create trigger reject_retry ` + trigger + ` begin select raise(abort,'retry checkpoint rejected'); end`); err != nil {
				t.Fatal(err)
			}
			if _, err := s.SettleProviderAttempt(ctx, "t", "s", "request", retryReceipt()); err == nil {
				t.Fatal("failed checkpoint committed")
			}
			var n int
			if err := s.db.QueryRow(`select count(*) from turn_receipts`).Scan(&n); err != nil || n != 0 {
				t.Fatal(n, err)
			}
			if err := s.db.QueryRow(`select count(*) from turn_events where event_type in ('inference.settled','inference.retry_scheduled')`).Scan(&n); err != nil || n != 0 {
				t.Fatal(n, err)
			}
			row, err := s.GetTurn(ctx, "t")
			if err != nil || row.Phase != "running" {
				t.Fatal(row, err)
			}
			accounting, err := s.TurnAccounting(ctx, "t")
			if err != nil || accounting.Usage.TotalTokens != 0 || accounting.Unknown != 1 {
				t.Fatal(accounting, err)
			}
		})
	}
}
func TestProviderRetryCheckpointSurvivesReopenAndDuplicate(t *testing.T) {
	s, ctx, path := retryReceiptFixture(t)
	receipt := retryReceipt()
	first, err := s.SettleProviderAttempt(ctx, "t", "s", "request", receipt)
	if err != nil || first.Decision != "retry_wait" || first.RetryWait == nil {
		t.Fatal(first, err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	changed := retryReceipt()
	changed.RetryWait.RetryAt = time.Now().Add(time.Hour)
	duplicate, err := reopened.SettleProviderAttempt(ctx, "t", "s", "request", changed)
	if err != nil || duplicate.Applied || duplicate.RetryWait == nil || !duplicate.RetryWait.RetryAt.Equal(receipt.RetryWait.RetryAt) {
		t.Fatal("duplicate replaced deadline", duplicate, err)
	}
	if duplicate.Accounting.Usage.TotalTokens != 7 || duplicate.Accounting.Unknown != 0 {
		t.Fatal(duplicate.Accounting)
	}
	row, err := reopened.GetTurn(ctx, "t")
	if err != nil || row.Phase != "retry_wait" {
		t.Fatal(row, err)
	}
	var n int
	if err := reopened.db.QueryRow(`select count(*) from turn_events where event_type='inference.retry_scheduled'`).Scan(&n); err != nil || n != 1 {
		t.Fatal(n, err)
	}
}
func TestProviderRetryCheckpointCannotReopenOrOutrunAdmission(t *testing.T) {
	for _, state := range []string{"staged", "later_inference", "later_tool", "unclaimed", "cancelling"} {
		t.Run(state, func(t *testing.T) {
			s, ctx, _ := retryReceiptFixture(t)
			switch state {
			case "staged":
				if _, err := s.db.Exec(`update turns set phase='settling' where id='t'`); err != nil {
					t.Fatal(err)
				}
			case "later_inference":
				if err := s.AdmitTurnEffect(ctx, "s", "t", "running", "inference.started", map[string]any{"request_id": "later"}); err != nil {
					t.Fatal(err)
				}
			case "later_tool":
				if err := s.AdmitTurnEffect(ctx, "s", "t", "waiting_on_tools", "tool.started", map[string]any{"occurrence_id": "later"}); err != nil {
					t.Fatal(err)
				}
			case "unclaimed":
				if err := s.ReleaseSessionActiveTurn(ctx, "s", "epoch"); err != nil {
					t.Fatal(err)
				}
			case "cancelling":
				if err := s.UpdateTurnStatusAndPhase(ctx, "t", "cancelling", "cancelling"); err != nil {
					t.Fatal(err)
				}
			}
			settled, err := s.SettleProviderAttempt(ctx, "t", "s", "request", retryReceipt())
			if state != "cancelling" {
				if err == nil {
					t.Fatal("invalid retry successor admitted", settled)
				}
				return
			}
			if err != nil || settled.Decision != "cancelled" || settled.RetryWait != nil || settled.Accounting.Usage.TotalTokens != 7 {
				t.Fatal("cancellation lost accounting or scheduled retry", settled, err)
			}
			var n int
			if err := s.db.QueryRow(`select count(*) from turn_events where event_type='inference.retry_scheduled'`).Scan(&n); err != nil || n != 0 {
				t.Fatal(n, err)
			}
		})
	}
}

func TestProviderRetryWakeRequiresDeadlineOwnershipAndAtomicEvent(t *testing.T) {
	for _, state := range []string{"ready", "future", "cancelled", "revoked", "released", "fault_event", "fault_phase"} {
		t.Run(state, func(t *testing.T) {
			s, ctx, _ := retryReceiptFixture(t)
			receipt := retryReceipt()
			receipt.RetryWait.RetryAt = time.Now().Add(-time.Millisecond)
			if state == "future" {
				receipt.RetryWait.RetryAt = time.Now().Add(time.Hour)
			}
			if _, err := s.SettleProviderAttempt(ctx, "t", "s", "request", receipt); err != nil {
				t.Fatal(err)
			}
			switch state {
			case "cancelled":
				if err := s.UpdateTurnStatusAndPhase(ctx, "t", "cancelling", "cancelling"); err != nil {
					t.Fatal(err)
				}
			case "revoked":
				if _, err := s.db.Exec(`update session_active_turns set claim_token='new' where session_id='s'`); err != nil {
					t.Fatal(err)
				}
			case "released":
				if err := s.ReleaseSessionActiveTurn(ctx, "s", "epoch"); err != nil {
					t.Fatal(err)
				}
			case "fault_event":
				if _, err := s.db.Exec(`create trigger reject_wake before insert on turn_events when new.event_type='inference.retry_resumed' begin select raise(abort,'wake rejected'); end`); err != nil {
					t.Fatal(err)
				}
			case "fault_phase":
				if _, err := s.db.Exec(`create trigger reject_wake before update of phase on turns when new.phase='running' begin select raise(abort,'wake rejected'); end`); err != nil {
					t.Fatal(err)
				}
			}
			err := s.ResumeProviderRetry(ctx, "s", "t", "request")
			if state == "ready" {
				if err != nil {
					t.Fatal(err)
				}
				if err := s.ResumeProviderRetry(ctx, "s", "t", "request"); err == nil {
					t.Fatal("wait resumed twice")
				}
			} else if err == nil {
				t.Fatal("invalid wake admitted")
			}
			var n int
			if err := s.db.QueryRow(`select count(*) from turn_events where event_type='inference.retry_resumed'`).Scan(&n); err != nil {
				t.Fatal(err)
			}
			if (state == "ready" && n != 1) || (state != "ready" && n != 0) {
				t.Fatal("wake event not atomic", n)
			}
			row, err := s.GetTurn(context.Background(), "t")
			if err != nil {
				t.Fatal(err)
			}
			expected := "retry_wait"
			if state == "ready" {
				expected = "running"
			}
			if state == "cancelled" {
				expected = "cancelling"
			}
			if row.Phase != expected {
				t.Fatal("wake changed phase", row)
			}
		})
	}
}

func TestProviderRetryAccountingScratchDoesNotCarrySparseFields(t *testing.T) {
	s, ctx, _ := retryReceiptFixture(t)
	first := retryReceipt()
	first.RetryWait = nil
	first.Usage = &goai.Usage{Input: 2, Output: 3, TotalTokens: 5, CacheRead: 6}
	if _, err := s.SettleProviderAttempt(ctx, "t", "s", "request", first); err != nil {
		t.Fatal(err)
	}
	if err := s.AdmitTurnEffect(ctx, "s", "t", "running", "inference.started", map[string]any{"request_id": "request-z"}); err != nil {
		t.Fatal(err)
	}
	second := retryReceipt()
	second.RetryWait = nil
	second.Usage = &goai.Usage{Input: 1}
	if _, err := s.SettleProviderAttempt(ctx, "t", "s", "request-z", second); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`update turn_receipts set usage_json='{"input":1}' where occurrence_id='request-z'`); err != nil {
		t.Fatal(err)
	}
	a, err := s.TurnAccounting(ctx, "t")
	if err != nil || a.Usage.Input != 3 || a.Usage.Output != 3 || a.Usage.TotalTokens != 5 || a.Usage.CacheRead != 6 {
		t.Fatal("sparse receipt inherited previous usage", a, err)
	}
}
