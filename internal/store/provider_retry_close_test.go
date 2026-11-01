package store

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestProviderRetryCloseReservationCommitAndRecovery(t *testing.T) {
	for _, mode := range []string{"sleep", "woken", "legacy", "later", "foreign", "aborting", "reserve_claim", "reserve_fence", "reserve_event", "commit_turn", "commit_claim", "commit_event", "commit_session", "abort_after_reserve"} {
		t.Run(mode, func(t *testing.T) {
			s, old, _ := retryReceiptFixture(t)
			receipt := retryReceipt()
			receipt.RetryWait.RetryAt = time.Now().Add(time.Hour)
			if mode != "legacy" {
				receipt.RetryWait.Lifecycle = &ProviderLifecyclePolicy{Version: 1, Message: "once"}
			}
			if mode == "woken" {
				receipt.RetryWait.RetryAt = time.Now().Add(-time.Millisecond)
			}
			if _, err := s.SettleProviderAttempt(old, "t", "s", "request", receipt); err != nil {
				t.Fatal(err)
			}
			if mode == "woken" {
				if err := s.ResumeProviderRetry(old, "s", "t", "request"); err != nil {
					t.Fatal(err)
				}
			}
			switch mode {
			case "later":
				if err := s.AppendTurnEvent(old, "t", "s", "tool.started", nil); err != nil {
					t.Fatal(err)
				}
			case "foreign":
				if _, err := s.db.Exec(`update session_active_turns set claim_token='foreign'`); err != nil {
					t.Fatal(err)
				}
			case "aborting":
				if _, err := s.RequestTurnCancellation(context.Background(), "t", "s"); err != nil {
					t.Fatal(err)
				}
			case "reserve_claim":
				closeRetryTrigger(t, s, "update", "session_active_turns", "")
			case "reserve_fence":
				closeRetryTrigger(t, s, "insert", "turn_recovery_fences", "")
			case "reserve_event":
				closeRetryTrigger(t, s, "insert", "turn_events", "when new.event_type='turn.checkpoint_close_started'")
			}
			checkpoint, err := s.ReserveProviderRetryClose(context.Background(), "s", "t", "epoch", "close-epoch")
			rejected := mode == "legacy" || mode == "later" || mode == "foreign" || mode == "aborting" || strings.HasPrefix(mode, "reserve_")
			if rejected {
				if checkpoint != nil {
					t.Fatal("invalid shutdown checkpoint reserved", checkpoint)
				}
				if (mode == "foreign" || strings.HasPrefix(mode, "reserve_")) && err == nil {
					t.Fatal("reservation error lost", mode)
				}
				_, token, readErr := s.GetSessionActiveTurn(context.Background(), "s")
				want := "epoch"
				if mode == "foreign" {
					want = "foreign"
				}
				if readErr != nil || token != want {
					t.Fatal("failed reservation rotated owner", token, readErr)
				}
				var fences int
				if err = s.db.QueryRow(`select count(*) from turn_recovery_fences`).Scan(&fences); err != nil || fences != 0 {
					t.Fatal("failed reservation left fence", fences, err)
				}
				return
			}
			if err != nil || checkpoint == nil || checkpoint.Resumed != (mode == "woken") {
				t.Fatal(checkpoint, err)
			}
			row, err := s.GetTurn(context.Background(), "t")
			if err != nil || row.Status != "running" {
				t.Fatal("reservation published before join", row, err)
			}
			if err = s.AppendTurnEffectEvent(old, "t", "s", "tool.output", nil); err == nil {
				t.Fatal("old writer survived reservation")
			}
			switch mode {
			case "commit_turn":
				closeRetryTrigger(t, s, "update", "turns", "")
			case "commit_claim":
				closeRetryTrigger(t, s, "update", "session_active_turns", "")
			case "commit_event":
				closeRetryTrigger(t, s, "insert", "turn_events", "when new.event_type='turn.checkpoint_queued'")
			case "commit_session":
				closeRetryTrigger(t, s, "update", "sessions", "")
			case "abort_after_reserve":
				if _, err = s.RequestTurnCancellation(context.Background(), "t", "s"); err != nil {
					t.Fatal(err)
				}
			}
			err = s.CommitProviderRetryClose(context.Background(), "s", "t", "close-epoch")
			if strings.HasPrefix(mode, "commit_") || mode == "abort_after_reserve" {
				if err == nil {
					t.Fatal("failed close was accepted")
				}
				row, _ = s.GetTurn(context.Background(), "t")
				want := "running"
				if mode == "abort_after_reserve" {
					want = "cancelling"
				}
				if row.Status != want {
					t.Fatal("failed close changed status/abort mark", row)
				}
				var queued int
				if e := s.db.QueryRow(`select count(*) from turn_events where event_type='turn.checkpoint_queued'`).Scan(&queued); e != nil || queued != 0 {
					t.Fatal("partial close event committed", queued, e)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			row, err = s.GetTurn(context.Background(), "t")
			if err != nil || row.Status != "queued" || !IsClosedProviderRetryPhase(row.Phase) || row.FinishedAt != "" {
				t.Fatal(row, err)
			}
			if next, err := s.GetNextQueuedTurn(context.Background(), "s"); err != sql.ErrNoRows || next != nil {
				t.Fatal("checkpoint entered whole-prompt queue", next, err)
			}
			claims, err := s.ListStaleActiveTurnClaims(context.Background(), time.Now().Add(-time.Hour), "s")
			if err != nil || len(claims) != 1 {
				t.Fatal("closed checkpoint waited for crash timeout", claims, err)
			}
			recovered, err := s.AcquireProviderRetryRecovery(context.Background(), claims[0], "resumed-epoch", time.Now().Add(-time.Hour))
			if err != nil || recovered == nil {
				t.Fatal(recovered, err)
			}
			if recovered.Resumed != checkpoint.Resumed || recovered.Wait.Lifecycle.Message != "once" || !recovered.Wait.RetryAt.Equal(receipt.RetryWait.RetryAt) {
				t.Fatal("close changed continuation", recovered)
			}
			if err = s.AppendTurnEffectEvent(WithTurnWriter(context.Background(), "s", "t", "close-epoch"), "t", "s", "tool.output", nil); err == nil {
				t.Fatal("close controller survived recovery fence")
			}
		})
	}
}

func closeRetryTrigger(t *testing.T, s *Store, operation, table, condition string) {
	t.Helper()
	if _, err := s.db.Exec(`create trigger reject_close before ` + operation + ` on ` + table + ` ` + condition + ` begin select raise(abort,'close rejected');end`); err != nil {
		t.Fatal(err)
	}
}

func TestProviderRetryClosedCheckpointCannotAcquireWholePromptClaim(t *testing.T) {
	s, ctx, _ := retryReceiptFixture(t)
	receipt := retryReceipt()
	receipt.RetryWait.Lifecycle = &ProviderLifecyclePolicy{Version: 1}
	if _, err := s.SettleProviderAttempt(ctx, "t", "s", "request", receipt); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReserveProviderRetryClose(context.Background(), "s", "t", "epoch", "close"); err != nil {
		t.Fatal(err)
	}
	if err := s.CommitProviderRetryClose(context.Background(), "s", "t", "close"); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseSessionActiveTurn(context.Background(), "s", "close"); err != nil {
		t.Fatal(err)
	}
	if claimed, err := s.ClaimSessionActiveTurn(context.Background(), "s", "t", "queue", "whole-prompt"); err != nil || claimed {
		t.Fatal("released checkpoint became whole-prompt authority", claimed, err)
	}
}
