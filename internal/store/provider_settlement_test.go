package store

import (
	"context"
	"math"
	"path/filepath"
	"testing"

	goai "github.com/rcarmo/go-ai"
)

func settlementFixture(t *testing.T) (*Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "receipts.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	ctx := context.Background()
	if _, err = s.CreateSession(ctx, "s", "session", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateTurnWithStatus(ctx, "t", "s", "running", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	return s, path
}
func sampleProviderReceipt() ProviderReceipt {
	return ProviderReceipt{MessageID: "m", Content: "answer", Payload: map[string]any{"turn_id": "t"}, Usage: &goai.Usage{Input: 7, Output: 2, TotalTokens: 9}, Measurement: map[string]any{"tokens": 7, "input": 7, "iteration": 1}, Decision: "reply_boundary", Iteration: 1}
}

func TestProviderSettlementAtomicReopen(t *testing.T) {
	for _, target := range []string{"messages", "turn_receipts", "turns", "turn_events", "sessions"} {
		t.Run(target, func(t *testing.T) {
			s, path := settlementFixture(t)
			ctx := context.Background()
			verb := "insert"
			if target == "turns" || target == "sessions" {
				verb = "update"
			}
			if _, err := s.db.Exec(`create trigger reject_receipt before ` + verb + ` on ` + target + ` begin select raise(abort,'reject receipt'); end`); err != nil {
				t.Fatal(err)
			}
			if _, err := s.SettleProviderAttempt(ctx, "t", "s", "request", sampleProviderReceipt()); err == nil {
				t.Fatal("fault hidden")
			}
			for _, table := range []string{"messages", "turn_receipts", "turn_events"} {
				var count int
				if err := s.db.QueryRow(`select count(*) from ` + table).Scan(&count); err != nil || count != 0 {
					t.Fatal("partial receipt", table, count, err)
				}
			}
			rec, _ := s.GetTurn(ctx, "t")
			if rec.Metadata["usage"] != nil {
				t.Fatal("accounting escaped rollback")
			}
			if _, err := s.db.Exec(`drop trigger reject_receipt`); err != nil {
				t.Fatal(err)
			}
			cancelled, cancel := context.WithCancel(ctx)
			cancel()
			if _, err := s.SettleProviderAttempt(cancelled, "t", "s", "request", sampleProviderReceipt()); err == nil {
				t.Fatal("cancelled admission succeeded")
			}
			first, err := s.SettleProviderAttempt(ctx, "t", "s", "request", sampleProviderReceipt())
			if err != nil || !first.Applied {
				t.Fatal(first, err)
			}
			retry := sampleProviderReceipt()
			retry.MessageID = "retry-message"
			retry.Usage.Input = 999
			again, err := s.SettleProviderAttempt(ctx, "t", "s", "request", retry)
			if err != nil || again.Applied || again.MessageID != "m" || again.Accounting.Usage.Input != 7 {
				t.Fatal("receipt not idempotent", again, err)
			}
			if err = s.Close(); err != nil {
				t.Fatal(err)
			}
			reopened, err := Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			total, err := reopened.TurnAccounting(ctx, "t")
			if err != nil || total.Usage.Input != 7 || total.Unknown != 0 {
				t.Fatal("receipt lost on reopen", total, err)
			}
			var count int
			if err = reopened.db.QueryRow(`select count(*) from messages`).Scan(&count); err != nil || count != 1 {
				t.Fatal(count, err)
			}
			measurement, err := reopened.LatestContextMeasurement(ctx, "s")
			if err != nil || measurement == nil || measurement.Tokens != 7 {
				t.Fatal(measurement, err)
			}
		})
	}
}

func TestReceiptAccountingUnknownAndNestedToolUsage(t *testing.T) {
	s, _ := settlementFixture(t)
	ctx := context.Background()
	if err := s.AppendTurnEvent(ctx, "t", "s", "inference.started", map[string]any{"request_id": "lost"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SettleProviderAttempt(ctx, "t", "s", "first", sampleProviderReceipt()); err != nil {
		t.Fatal(err)
	}
	payload := map[string]any{"turn_id": "t", "occurrence_id": "tool", "usage": goai.Usage{Input: 3, Output: 1, TotalTokens: 4}}
	for i := 0; i < 2; i++ {
		if err := s.SettleToolResult(ctx, "t", "s", "tool.finished", map[string]any{"occurrence_id": "tool"}, "tool-result", "nested response", payload); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.SettleProviderAttempt(ctx, "t", "s", "failed", ProviderReceipt{Decision: "error_boundary", Error: "disconnected"}); err != nil {
		t.Fatal(err)
	}
	total, err := s.TurnAccounting(ctx, "t")
	if err != nil || total.Usage.Input != 10 || total.Usage.TotalTokens != 13 || total.Unknown != 2 {
		t.Fatal(total, err)
	}
	if err = s.SettleToolResult(ctx, "t", "s", "tool.cancelled", map[string]any{"occurrence_id": "interrupted"}, "", "", map[string]any{"usage": goai.Usage{Input: 2}, "usage_unknown": true}); err != nil {
		t.Fatal(err)
	}
	total, err = s.TurnAccounting(ctx, "t")
	if err != nil || total.Usage.Input != 12 || total.Unknown != 3 {
		t.Fatal(total, err)
	}
}

func TestProviderSettlementOwnerAndInvalidUsage(t *testing.T) {
	s, _ := settlementFixture(t)
	ctx := context.Background()
	for _, owner := range []string{"foreign", "s"} {
		receipt := sampleProviderReceipt()
		if owner == "s" {
			receipt.Usage.Input = -1
		}
		if _, err := s.SettleProviderAttempt(ctx, "t", owner, "bad", receipt); err == nil {
			t.Fatal("invalid receipt adopted", owner)
		}
	}
	var n int
	if err := s.db.QueryRow(`select count(*) from turn_receipts`).Scan(&n); err != nil || n != 0 {
		t.Fatal(n, err)
	}
}

func TestReceiptAccountingOverflowRollsBackNewReceipt(t *testing.T) {
	for _, cost := range []bool{false, true} {
		t.Run(map[bool]string{false: "tokens", true: "cost"}[cost], func(t *testing.T) {
			s, _ := settlementFixture(t)
			ctx := context.Background()
			first := sampleProviderReceipt()
			first.Measurement = nil
			if cost {
				first.Usage.Cost.Total = math.MaxFloat64
			} else {
				first.Usage.Input = math.MaxInt
			}
			if _, err := s.SettleProviderAttempt(ctx, "t", "s", "first", first); err != nil {
				t.Fatal(err)
			}
			second := sampleProviderReceipt()
			second.MessageID = "second"
			if cost {
				second.Usage.Cost.Total = math.MaxFloat64
			}
			if _, err := s.SettleProviderAttempt(ctx, "t", "s", "second", second); err == nil {
				t.Fatal("overflow committed")
			}
			var count int
			if err := s.db.QueryRow(`select count(*) from turn_receipts`).Scan(&count); err != nil || count != 1 {
				t.Fatal(count, err)
			}
			a, err := s.TurnAccounting(ctx, "t")
			if err != nil || a.Usage.Input != first.Usage.Input || a.Usage.Cost.Total != first.Usage.Cost.Total {
				t.Fatal("prior accounting damaged", a, err)
			}
		})
	}
}

func TestLegacyUnkeyedAdmissionsRemainUsageUnknown(t *testing.T) {
	s, _ := settlementFixture(t)
	ctx := context.Background()
	for _, kind := range []string{"inference.started", "tool.started"} {
		if err := s.AppendTurnEvent(ctx, "t", "s", kind, nil); err != nil {
			t.Fatal(err)
		}
	}
	a, err := s.TurnAccounting(ctx, "t")
	if err != nil || a.Unknown != 2 {
		t.Fatal("legacy admission treated as known zero usage", a, err)
	}
}
