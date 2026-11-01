package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	goai "github.com/rcarmo/go-ai"
)

func interruptedProviderFixture(t *testing.T, partial bool) (*Store, context.Context) {
	t.Helper()
	s, ctx, _ := generationReadyFixture(t)
	if err := s.SaveGenerationReady(ctx, "s", "t", readyCheckpoint()); err != nil {
		t.Fatal(err)
	}
	admitInterruptedTestRequest(t, s, ctx, "pending")
	if partial {
		message := goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: "committed prefix"}, {Type: "thinking", Thinking: "committed thought", ThinkingSignature: "retained signature"}, {Type: "toolCall", ID: "not-admitted", Name: "must-not-run", Arguments: map[string]any{"prefix": true}}}, Usage: &goai.Usage{Input: 17}}
		if err := s.SaveProviderPartial(ctx, "t", "s", "pending", "tampered UI draft", map[string]any{"provider_message": message, "thinking_blocks": []map[string]any{{"thinking": "committed thought", "thinking_signature": "retained signature"}}}); err != nil {
			t.Fatal(err)
		}
	}
	return s, ctx
}
func admitInterruptedTestRequest(t *testing.T, s *Store, ctx context.Context, id string) {
	t.Helper()
	for _, kind := range []string{"inference.started", "inference.request_prepared"} {
		if err := s.AdmitTurnEffect(ctx, "s", "t", "running", kind, map[string]any{"request_id": id, "iteration": 2, "phase": "inference", "intent": map[string]any{"messages": "old request must never be dispatched"}}); err != nil {
			t.Fatal(err)
		}
	}
}
func staleGenerationClaim(t *testing.T, s *Store) ActiveTurnClaim {
	t.Helper()
	if _, err := s.db.Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
		t.Fatal(err)
	}
	claims, err := s.ListStaleActiveTurnClaims(context.Background(), time.Now(), "s")
	if err != nil || len(claims) != 1 {
		t.Fatal(claims, err)
	}
	return claims[0]
}
func TestInterruptedProviderRecoverySealsUncertainReceipt(t *testing.T) {
	for _, partial := range []bool{false, true} {
		t.Run(fmt.Sprint(partial), func(t *testing.T) {
			s, oldCtx := interruptedProviderFixture(t, partial)
			cp, err := s.AcquireGenerationReadyRecovery(context.Background(), staleGenerationClaim(t, s), "new", time.Now())
			if err != nil || cp == nil || cp.Iteration != 2 || cp.MaxIterations != 4 || cp.LastToolFailureSignature != "prior" {
				t.Fatal(cp, err)
			}
			row, err := s.GetTurn(context.Background(), "t")
			if err != nil || row.Phase != "generation_ready" {
				t.Fatal(row, err)
			}
			accounting, err := s.TurnAccounting(context.Background(), "t")
			want := 5
			if partial {
				want += 17
			}
			if err != nil || accounting.Unknown != 1 || accounting.Usage.Input != want || accounting.Iterations != 2 {
				t.Fatal(accounting, err)
			}
			var decision string
			if err = s.db.QueryRow(`select json_extract(decision_json,'$.next') from turn_receipts where turn_id='t' and occurrence_id='pending'`).Scan(&decision); err != nil || decision != "interrupted_boundary" {
				t.Fatal(decision, err)
			}
			var interrupted, settled int
			if err = s.db.QueryRow(`select count(*) from turn_events where turn_id='t' and event_type='inference.interrupted'`).Scan(&interrupted); err != nil || interrupted != 1 {
				t.Fatal(interrupted, err)
			}
			if err = s.db.QueryRow(`select count(*) from turn_events where turn_id='t' and event_type='inference.settled'`).Scan(&settled); err != nil || settled != 1 {
				t.Fatal("interruption recorded as progress", settled, err)
			}
			if partial {
				record, err := s.ProviderPartial(context.Background(), "t", "s", "pending")
				if err != nil || record == nil || record.Content != "committed prefix" {
					t.Fatal("UI draft trusted or prefix lost", record, err)
				}
				var adopted bool
				if err = s.db.QueryRow(`select json_extract(payload_json,'$.adopted') from messages where id=?`, ProviderMessageID("pending")).Scan(&adopted); err != nil || adopted {
					t.Fatal("interrupted prefix adopted", adopted, err)
				}
			}
			if err = s.SaveProviderPartial(oldCtx, "t", "s", "pending", "stale", map[string]any{}); err == nil {
				t.Fatal("stale writer mutated sealed prefix")
			}
			// A crash after acquisition, before fresh request preparation, remains safe.
			again, err := s.AcquireGenerationReadyRecovery(context.Background(), staleGenerationClaim(t, s), "newer", time.Now())
			if err != nil || again == nil {
				t.Fatal(again, err)
			}
			accounting, err = s.TurnAccounting(context.Background(), "t")
			if err != nil || accounting.Unknown != 1 || accounting.Usage.Input != want {
				t.Fatal("duplicate interrupted receipt", accounting, err)
			}
		})
	}
}
func TestInterruptedProviderRecoveryAtomicRollback(t *testing.T) {
	for _, mode := range []string{"receipt", "message", "interrupted-event", "ready-event", "phase", "recovery-event", "session"} {
		t.Run(mode, func(t *testing.T) {
			s, _ := interruptedProviderFixture(t, true)
			statement := ""
			switch mode {
			case "receipt":
				statement = `create trigger reject_interrupt before insert on turn_receipts when new.occurrence_id='pending' begin select raise(abort,'receipt');end`
			case "message":
				statement = `create trigger reject_interrupt before update on messages begin select raise(abort,'message');end`
			case "interrupted-event":
				statement = `create trigger reject_interrupt before insert on turn_events when new.event_type='inference.interrupted' begin select raise(abort,'interrupted');end`
			case "ready-event":
				statement = `create trigger reject_interrupt before insert on turn_events when new.event_type='generation.ready' begin select raise(abort,'ready');end`
			case "phase":
				statement = `create trigger reject_interrupt before update on turns when new.phase='generation_ready' begin select raise(abort,'phase');end`
			case "recovery-event":
				statement = `create trigger reject_interrupt before insert on turn_events when new.event_type='generation.recovered' begin select raise(abort,'recovered');end`
			case "session":
				statement = `create trigger reject_interrupt before update on sessions begin select raise(abort,'session');end`
			}
			if _, err := s.db.Exec(statement); err != nil {
				t.Fatal(err)
			}
			cp, err := s.AcquireGenerationReadyRecovery(context.Background(), staleGenerationClaim(t, s), "new", time.Now())
			if err == nil || cp != nil {
				t.Fatal("fault authorised successor", cp, err)
			}
			_, token, err := s.GetSessionActiveTurn(context.Background(), "s")
			if err != nil || token != "epoch" {
				t.Fatal("fault changed owner", token, err)
			}
			row, err := s.GetTurn(context.Background(), "t")
			if err != nil || row.Phase != "running" {
				t.Fatal(row, err)
			}
			var count int
			if err = s.db.QueryRow(`select count(*) from turn_receipts where turn_id='t' and occurrence_id='pending'`).Scan(&count); err != nil || count != 0 {
				t.Fatal("partial transaction committed", count, err)
			}
			if err = s.db.QueryRow(`select count(*) from turn_recovery_fences where turn_id='t'`).Scan(&count); err != nil || count != 0 {
				t.Fatal("fault fenced owner", count, err)
			}
		})
	}
}
func TestInterruptedProviderRecoveryRejectsOtherBoundaries(t *testing.T) {
	for _, mode := range []string{"tool", "second-request", "retry", "settled", "mismatched-iteration", "mismatched-prepared", "foreign-session", "unsupported-prepared-phase", "duplicate-prepared", "malformed-header", "string-iteration", "object-request-id"} {
		t.Run(mode, func(t *testing.T) {
			s, ctx := interruptedProviderFixture(t, false)
			var err error
			switch mode {
			case "tool":
				err = s.AppendTurnEvent(ctx, "t", "s", "tool.started", map[string]any{"iteration": 2, "replay_safe": true, "occurrence_id": "tool"})
			case "second-request":
				err = s.AdmitTurnEffect(ctx, "s", "t", "running", "inference.started", map[string]any{"request_id": "other", "iteration": 2, "phase": "inference"})
			case "retry":
				err = s.AppendTurnEvent(ctx, "t", "s", "inference.retry_wake", map[string]any{"iteration": 2})
			case "settled":
				receipt := retryReceipt()
				receipt.RetryWait = nil
				receipt.Iteration = 2
				_, err = s.SettleProviderAttempt(ctx, "t", "s", "pending", receipt)
			case "mismatched-iteration":
				_, err = s.db.Exec(`update turn_events set payload_json=json_set(payload_json,'$.iteration',3) where event_type='inference.started' and json_extract(payload_json,'$.request_id')='pending'`)
			case "mismatched-prepared":
				_, err = s.db.Exec(`update turn_events set payload_json=json_set(payload_json,'$.request_id','foreign') where event_type='inference.request_prepared'`)
			case "foreign-session":
				if _, err = s.CreateSession(ctx, "other", "other", nil); err == nil {
					_, err = s.db.Exec(`update turn_events set session_id='other' where event_type='inference.request_prepared'`)
				}
			case "unsupported-prepared-phase":
				_, err = s.db.Exec(`update turn_events set payload_json=json_set(payload_json,'$.phase','unknown') where event_type='inference.request_prepared'`)
			case "duplicate-prepared":
				err = s.AppendTurnEvent(ctx, "t", "s", "inference.request_prepared", map[string]any{"request_id": "pending", "phase": "inference", "iteration": 2})
			case "malformed-header":
				_, err = s.db.Exec(`update turn_events set payload_json='[]' where event_type='inference.request_prepared'`)
			case "string-iteration":
				_, err = s.db.Exec(`update turn_events set payload_json=json_set(payload_json,'$.iteration','2') where event_type='inference.request_prepared'`)
			case "object-request-id":
				_, err = s.db.Exec(`update turn_events set payload_json=json_set(payload_json,'$.request_id',json('{}')) where event_type='inference.request_prepared'`)
			}
			if err != nil {
				t.Fatal(err)
			}
			cp, err := s.AcquireGenerationReadyRecovery(context.Background(), staleGenerationClaim(t, s), "new", time.Now())
			if err != nil || cp != nil {
				t.Fatal("unsafe boundary recovered", cp, err)
			}
		})
	}
}
func TestInterruptedProviderRecoveryDoesNotResetNoProgressLimit(t *testing.T) {
	s, ctx, _ := generationReadyFixture(t)
	if err := s.SaveGenerationReady(ctx, "s", "t", readyCheckpoint()); err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= generationReadyRecoveryLimit; i++ {
		admitInterruptedTestRequest(t, s, ctx, fmt.Sprintf("pending-%d", i))
		token := NowID("owner")
		cp, err := s.AcquireGenerationReadyRecovery(context.Background(), staleGenerationClaim(t, s), token, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		if i == generationReadyRecoveryLimit {
			if cp != nil {
				t.Fatal("interruption bookkeeping reset progress")
			}
			return
		}
		if cp == nil {
			t.Fatal("valid interruption held", i)
		}
		ctx = WithTurnWriter(context.Background(), "s", "t", token)
		if err = s.SaveGenerationReady(ctx, "s", "t", readyCheckpoint()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInterruptedProviderIntentOnlyRemainsUncertain(t *testing.T) {
	s, _ := interruptedProviderFixture(t, false)
	if _, err := s.db.Exec(`delete from turn_events where turn_id='t' and event_type='inference.request_prepared'`); err != nil {
		t.Fatal(err)
	}
	cp, err := s.AcquireGenerationReadyRecovery(context.Background(), staleGenerationClaim(t, s), "new", time.Now())
	if err != nil || cp == nil {
		t.Fatal(cp, err)
	}
	a, err := s.TurnAccounting(context.Background(), "t")
	if err != nil || a.Unknown != 1 {
		t.Fatal("unobserved admission treated as known zero", a, err)
	}
}

func TestInterruptedProviderReceiptCannotClaimKnownOutcome(t *testing.T) {
	s, ctx := interruptedProviderFixture(t, false)
	for _, receipt := range []ProviderReceipt{{Decision: "interrupted_boundary"}, {Decision: "interrupted_boundary", Unknown: true, RetryWait: retryReceipt().RetryWait}} {
		if _, err := s.SettleProviderAttempt(ctx, "t", "s", "pending", receipt); err == nil {
			t.Fatal("uncertain outcome weakened")
		}
	}
	var count int
	if err := s.db.QueryRow(`select count(*) from turn_receipts where turn_id='t' and occurrence_id='pending'`).Scan(&count); err != nil || count != 0 {
		t.Fatal("invalid receipt committed", count, err)
	}
}
