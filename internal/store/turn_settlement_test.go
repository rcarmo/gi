package store

import (
	"context"
	"testing"
)

func sampleTurnSettlement() TurnSettlement {
	return TurnSettlement{Status: "failed", Phase: "failed", MessageID: "terminal", Message: "Provider failed", MessagePayload: map[string]any{"kind": "chat", "turn_id": "t"}, FailureKind: "provider_error", Payload: map[string]any{"iterations": 1}}
}

func TestTurnSettlementAtomicReopen(t *testing.T) {
	for _, target := range []string{"turns", "subturns", "messages", "turn_failures", "turn_events", "sessions"} {
		t.Run(target, func(t *testing.T) {
			s, path := settlementFixture(t)
			ctx := context.Background()
			if _, err := s.CreateTurnWithStatus(ctx, "parent", "s", "running", "parent", nil); err != nil {
				t.Fatal(err)
			}
			if _, err := s.CreateSubTurn(ctx, "parent", "s", "t", "s", "sync", 1, nil); err != nil {
				t.Fatal(err)
			}
			if err := s.TouchSessionState(ctx, "s", map[string]any{"status": "running", "active_turn_id": "t"}); err != nil {
				t.Fatal(err)
			}
			if _, err := s.SettleProviderAttempt(ctx, "t", "s", "request", sampleProviderReceipt()); err != nil {
				t.Fatal(err)
			}
			verb := "insert"
			if target == "turns" || target == "subturns" || target == "sessions" {
				verb = "update"
			}
			if _, err := s.db.Exec(`create trigger reject_terminal before ` + verb + ` on ` + target + ` begin select raise(abort,'reject terminal'); end`); err != nil {
				t.Fatal(err)
			}
			if _, err := s.SettleRunningTurn(ctx, "t", "s", sampleTurnSettlement()); err == nil {
				t.Fatal("terminal fault hidden")
			}
			rec, err := s.GetTurn(ctx, "t")
			if err != nil || rec.Status != "running" || rec.FinishedAt != "" {
				t.Fatal("partial terminal", rec, err)
			}
			sub, err := s.GetSubTurnByChild(ctx, "t")
			if err != nil || sub.Status != "running" || sub.FinishedAt != "" {
				t.Fatal("partial child projection", sub, err)
			}
			var n int
			if err = s.db.QueryRow(`select count(*) from messages where id='terminal'`).Scan(&n); err != nil || n != 0 {
				t.Fatal("partial terminal message", n, err)
			}
			if err = s.db.QueryRow(`select count(*) from turn_events where event_type in ('turn.finished','inference.finished','turn.failure_marked')`).Scan(&n); err != nil || n != 0 {
				t.Fatal("partial terminal events", n, err)
			}
			if err = s.db.QueryRow(`select count(*) from turn_failures`).Scan(&n); err != nil || n != 0 {
				t.Fatal("partial failure projection", n, err)
			}
			if _, err = s.db.Exec(`drop trigger reject_terminal`); err != nil {
				t.Fatal(err)
			}
			first, err := s.SettleRunningTurn(ctx, "t", "s", sampleTurnSettlement())
			if err != nil || !first.Applied || first.Accounting.Usage.Input != 7 {
				t.Fatal(first, err)
			}
			again, err := s.SettleRunningTurn(ctx, "t", "s", sampleTurnSettlement())
			if err != nil || again.Applied {
				t.Fatal("terminal duplicated", again, err)
			}
			if err = s.Close(); err != nil {
				t.Fatal(err)
			}
			reopened, err := Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			rec, err = reopened.GetTurn(ctx, "t")
			if err != nil || rec.Status != "failed" || rec.FinishedAt == "" {
				t.Fatal("terminal lost after reopen", rec, err)
			}
			sub, err = reopened.GetSubTurnByChild(ctx, "t")
			if err != nil || sub.Status != "failed" || sub.FinishedAt == "" {
				t.Fatal("child terminal lost", sub, err)
			}
			if err = reopened.db.QueryRow(`select count(*) from turn_events where turn_id='t' and event_type='turn.finished'`).Scan(&n); err != nil || n != 1 {
				t.Fatal(n, err)
			}
		})
	}
}

func TestTurnSettlementCancellationAndProtectedFailure(t *testing.T) {
	t.Run("cancellation wins", func(t *testing.T) {
		s, _ := settlementFixture(t)
		ctx := context.Background()
		if _, err := s.SettleProviderAttempt(ctx, "t", "s", "request", sampleProviderReceipt()); err != nil {
			t.Fatal(err)
		}
		if err := s.UpdateTurnStatus(ctx, "t", "cancelling"); err != nil {
			t.Fatal(err)
		}
		out, err := s.SettleRunningTurn(ctx, "t", "s", TurnSettlement{Status: "completed", Phase: "completed", MessageID: "cancel"})
		if err != nil || !out.Applied || out.Status != "cancelled" || out.Phase != "aborted" || out.Message != "Turn cancelled" || out.FailureKind != "" || out.Accounting.Usage.Input != 7 {
			t.Fatal(out, err)
		}
	})
	t.Run("protected failure", func(t *testing.T) {
		s, _ := settlementFixture(t)
		ctx := context.Background()
		if err := s.UpsertTurnFailure(ctx, "t", "s", "held", "manual", "held"); err != nil {
			t.Fatal(err)
		}
		if _, err := s.SettleRunningTurn(ctx, "t", "s", TurnSettlement{Status: "completed", Phase: "completed"}); err == nil {
			t.Fatal("completed held turn")
		}
		rec, _ := s.GetTurn(ctx, "t")
		if rec.Status != "running" {
			t.Fatal(rec.Status)
		}
	})
	t.Run("ownership", func(t *testing.T) {
		s, _ := settlementFixture(t)
		if _, err := s.SettleRunningTurn(context.Background(), "t", "foreign", sampleTurnSettlement()); err == nil {
			t.Fatal("foreign session finalized turn")
		}
	})
}

func TestTurnSettlementDoesNotClearSuccessorSessionOwnership(t *testing.T) {
	s, _ := settlementFixture(t)
	ctx := context.Background()
	if err := s.TouchSessionState(ctx, "s", map[string]any{"status": "running", "active_turn_id": "successor"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SettleRunningTurn(ctx, "t", "s", sampleTurnSettlement()); err != nil {
		t.Fatal(err)
	}
	session, err := s.GetSession(ctx, "s")
	if err != nil || session.State["active_turn_id"] != "successor" || session.State["status"] != "running" {
		t.Fatal("new ownership cleared", session, err)
	}
}
