package store

import (
	"context"
	"testing"
	"time"
)

func generationReadyFixture(t *testing.T) (*Store, context.Context, string) {
	t.Helper()
	s, ctx, path := retryReceiptFixture(t)
	receipt := retryReceipt()
	receipt.RetryWait = nil
	receipt.Decision = "reply_boundary"
	receipt.MessageID = "settled-response"
	receipt.Content = "settled response"
	if _, err := s.SettleProviderAttempt(ctx, "t", "s", "request", receipt); err != nil {
		t.Fatal(err)
	}
	return s, ctx, path
}
func readyCheckpoint() *GenerationReadyCheckpoint {
	return &GenerationReadyCheckpoint{Version: 1, Iteration: 2, MaxIterations: 4, Model: "provider/model", Lifecycle: &ProviderLifecyclePolicy{Version: 1, Message: "once only"}, LastToolFailureSignature: "prior", RepeatedToolFailures: 1}
}
func TestGenerationReadyReopenAndFence(t *testing.T) {
	s, oldCtx, path := generationReadyFixture(t)
	saved := readyCheckpoint()
	if err := s.SaveGenerationReady(oldCtx, "s", "t", saved); err != nil {
		t.Fatal(err)
	}
	saved.Lifecycle.Message = "mutated"
	if _, err := s.db.Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	claims, err := s.ListStaleActiveTurnClaims(context.Background(), time.Now(), "s")
	if err != nil || len(claims) != 1 {
		t.Fatal(claims, err)
	}
	cp, err := s.AcquireGenerationReadyRecovery(context.Background(), claims[0], "new", time.Now())
	if err != nil || cp == nil || cp.Lifecycle.Message != "once only" || cp.Iteration != 2 || cp.MaxIterations != 4 || cp.RepeatedToolFailures != 1 || cp.LastToolFailureSignature != "prior" {
		t.Fatal(cp, err)
	}
	if err = s.AdmitTurnEffect(oldCtx, "s", "t", "running", "inference.started", map[string]any{"request_id": "stale"}); err == nil {
		t.Fatal("stale writer dispatched")
	}
	current := WithTurnWriter(context.Background(), "s", "t", "new")
	if err = s.AdmitTurnEffect(current, "s", "t", "running", "inference.started", map[string]any{"request_id": "next"}); err != nil {
		t.Fatal(err)
	}
	if _, err = s.AcquireGenerationReadyRecovery(context.Background(), claims[0], "other", time.Now()); err == nil {
		t.Fatal("stale claim reacquired")
	}
}
func TestGenerationReadyRejectsUnknownOrUnsettledState(t *testing.T) {
	for _, mode := range []string{"provider", "tool", "legacy-tool", "legacy-provider", "unsupported", "missing-policy", "invalid-policy", "later-effect", "later-unknown", "cancelled", "refreshed", "replaced", "live", "fault-claim", "fault-fence", "fault-event"} {
		t.Run(mode, func(t *testing.T) {
			s, ctx, _ := generationReadyFixture(t)
			if mode == "provider" || mode == "tool" || mode == "legacy-tool" || mode == "legacy-provider" {
				kind, payload := "inference.started", map[string]any{"request_id": "unsettled"}
				if mode == "tool" || mode == "legacy-tool" {
					kind = "tool.started"
					payload = map[string]any{"occurrence_id": "unsettled"}
				}
				if mode == "legacy-tool" || mode == "legacy-provider" {
					payload = nil
				}
				if err := s.AdmitTurnEffect(ctx, "s", "t", "running", kind, payload); err != nil {
					t.Fatal(err)
				}
				if err := s.SaveGenerationReady(ctx, "s", "t", readyCheckpoint()); err == nil {
					t.Fatal("unsettled effect hidden by checkpoint")
				}
				return
			}
			if err := s.SaveGenerationReady(ctx, "s", "t", readyCheckpoint()); err != nil {
				t.Fatal(err)
			}
			if _, err := s.db.Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
				t.Fatal(err)
			}
			claims, err := s.ListStaleActiveTurnClaims(context.Background(), time.Now(), "s")
			if err != nil || len(claims) != 1 {
				t.Fatal(claims, err)
			}
			cutoff := time.Now()
			switch mode {
			case "unsupported":
				_, err = s.db.Exec(`update turn_events set payload_json=json_set(payload_json,'$.continuation.version',9) where event_type='generation.ready'`)
			case "missing-policy":
				_, err = s.db.Exec(`update turn_events set payload_json=json_remove(payload_json,'$.continuation.lifecycle') where event_type='generation.ready'`)
			case "invalid-policy":
				_, err = s.db.Exec(`update turn_events set payload_json=json_set(payload_json,'$.continuation.lifecycle.version',9) where event_type='generation.ready'`)
			case "later-effect":
				err = s.AppendTurnEvent(ctx, "t", "s", "tool.started", nil)
			case "later-unknown":
				err = s.AppendTurnEvent(ctx, "t", "s", "unknown", nil)
			case "cancelled":
				err = s.UpdateTurnStatusAndPhase(ctx, "t", "cancelling", "cancelling")
			case "refreshed":
				err = s.TouchSessionActiveTurn(ctx, "s", "epoch")
			case "replaced":
				_, err = s.db.Exec(`update session_active_turns set claim_token='foreign'`)
			case "live":
				cutoff = time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)
			case "fault-claim":
				_, err = s.db.Exec(`create trigger reject_gen before update on session_active_turns begin select raise(abort,'claim rejected');end`)
			case "fault-fence":
				_, err = s.db.Exec(`create trigger reject_gen before insert on turn_recovery_fences begin select raise(abort,'fence rejected');end`)
			case "fault-event":
				_, err = s.db.Exec(`create trigger reject_gen before insert on turn_events when new.event_type='generation.recovered' begin select raise(abort,'event rejected');end`)
			}
			if err != nil {
				t.Fatal(err)
			}
			cp, err := s.AcquireGenerationReadyRecovery(context.Background(), claims[0], "new", cutoff)
			if cp != nil {
				t.Fatal("unsafe state recovered", cp)
			}
			if (mode == "refreshed" || mode == "replaced" || mode == "live" || mode == "fault-claim" || mode == "fault-fence" || mode == "fault-event") && err == nil {
				t.Fatal("ownership failure ignored")
			}
			_, token, err := s.GetSessionActiveTurn(context.Background(), "s")
			want := "epoch"
			if mode == "replaced" {
				want = "foreign"
			}
			if err != nil || token != want {
				t.Fatal("failed transaction changed owner", token, err)
			}
		})
	}
}

func TestGenerationReadyRecoveryRequiresReceiptProgress(t *testing.T) {
	for _, progress := range []bool{false, true} {
		t.Run(map[bool]string{false: "unchanged", true: "receipt-progress"}[progress], func(t *testing.T) {
			s, ctx, _ := generationReadyFixture(t)
			for i := 0; i < generationReadyRecoveryLimit+1; i++ {
				if progress && i == generationReadyRecoveryLimit {
					if err := s.AdmitTurnEffect(ctx, "s", "t", "running", "inference.started", map[string]any{"request_id": "progress"}); err != nil {
						t.Fatal(err)
					}
					receipt := retryReceipt()
					receipt.RetryWait = nil
					if _, err := s.SettleProviderAttempt(ctx, "t", "s", "progress", receipt); err != nil {
						t.Fatal(err)
					}
				}
				if err := s.SaveGenerationReady(ctx, "s", "t", readyCheckpoint()); err != nil {
					t.Fatal(err)
				}
				if _, err := s.db.Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
					t.Fatal(err)
				}
				claims, err := s.ListStaleActiveTurnClaims(context.Background(), time.Now(), "s")
				if err != nil || len(claims) != 1 {
					t.Fatal(claims, err)
				}
				token := NowID("epoch")
				cp, err := s.AcquireGenerationReadyRecovery(context.Background(), claims[0], token, time.Now())
				if err != nil {
					t.Fatal(err)
				}
				if !progress && i == generationReadyRecoveryLimit {
					if cp != nil {
						t.Fatal("rewritten preparation reset interruption budget")
					}
					return
				}
				if cp == nil {
					t.Fatal("receipt progress did not permit recovery", i)
				}
				ctx = WithTurnWriter(context.Background(), "s", "t", token)
			}
		})
	}
}

func TestGenerationReadyCheckpointWriteIsAtomic(t *testing.T) {
	for _, mode := range []string{"event-fault", "cancelled", "retry-wait"} {
		t.Run(mode, func(t *testing.T) {
			s, ctx, _ := generationReadyFixture(t)
			phase := "running"
			var err error
			switch mode {
			case "event-fault":
				_, err = s.db.Exec(`create trigger reject_ready before insert on turn_events when new.event_type='generation.ready' begin select raise(abort,'ready rejected');end`)
			case "cancelled":
				phase = "cancelling"
				err = s.UpdateTurnStatusAndPhase(ctx, "t", "cancelling", phase)
			case "retry-wait":
				phase = "retry_wait"
				err = s.UpdateTurnStatusAndPhase(ctx, "t", "running", phase)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err = s.SaveGenerationReady(ctx, "s", "t", readyCheckpoint()); err == nil {
				t.Fatal("invalid frontier committed")
			}
			row, err := s.GetTurn(ctx, "t")
			if err != nil || row.Phase != phase {
				t.Fatal("failed checkpoint changed phase", row, err)
			}
			var count int
			if err = s.db.QueryRow(`select count(*) from turn_events where turn_id='t' and event_type='generation.ready'`).Scan(&count); err != nil || count != 0 {
				t.Fatal("failed checkpoint left event", count, err)
			}
		})
	}
}

func TestGenerationReadyConcurrentRecoverers(t *testing.T) {
	s, ctx, path := generationReadyFixture(t)
	if err := s.SaveGenerationReady(ctx, "s", "t", readyCheckpoint()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
		t.Fatal(err)
	}
	claims, err := s.ListStaleActiveTurnClaims(context.Background(), time.Now(), "s")
	if err != nil || len(claims) != 1 {
		t.Fatal(claims, err)
	}
	other, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	results := make(chan bool, 2)
	for _, db := range []*Store{s, other} {
		go func(db *Store) {
			token, err := NewActiveTurnClaimToken()
			if err != nil {
				results <- false
				return
			}
			cp, err := db.AcquireGenerationReadyRecovery(context.Background(), claims[0], token, time.Now())
			results <- err == nil && cp != nil
		}(db)
	}
	admitted := 0
	for range 2 {
		if <-results {
			admitted++
		}
	}
	if admitted != 1 {
		t.Fatal("competing owners admitted", admitted)
	}
}
