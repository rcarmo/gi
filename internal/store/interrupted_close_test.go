package store

import (
	"database/sql"
	"errors"
	"testing"

	goai "github.com/rcarmo/go-ai"
)

func TestInterruptedCloseFencesAndAtomicallyHolds(t *testing.T) {
	for _, phase := range []string{"setup", "preparing", "inference"} {
		t.Run(phase, func(t *testing.T) {
			s, ctx, path := generationReadyFixture(t)
			_, old, err := s.GetSessionActiveTurn(ctx, "s")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.DB().Exec(`update turns set phase=? where id='t'`, phase); err != nil {
				t.Fatal(err)
			}
			ok, err := s.ReserveInterruptedClose(ctx, "s", "t", old, "closing")
			if err != nil || !ok {
				t.Fatal(ok, err)
			}
			if err = s.AdmitTurnEffect(ctx, "s", "t", "inference", "inference.started", map[string]any{"request_id": "stale"}); err == nil {
				t.Fatal("stale writer accepted")
			}
			row, err := s.GetTurn(ctx, "t")
			if err != nil || row.Status != "running" || row.FinishedAt != "" {
				t.Fatal("published before join", row, err)
			}
			if _, err = s.DB().Exec(`create trigger reject_hold before update on turns when new.phase='held_for_retry_or_skip' begin select raise(abort,'hold rejected');end`); err != nil {
				t.Fatal(err)
			}
			failure := TurnFailure{FailureKind: "recovery_interrupted", Summary: "unproven shutdown continuation"}
			if err = s.CommitInterruptedClose(ctx, "s", "t", "closing", failure); err == nil {
				t.Fatal("commit fault lost")
			}
			if _, err = s.GetTurnFailure(ctx, "t"); !errors.Is(err, sql.ErrNoRows) {
				t.Fatal("failure escaped rollback", err)
			}
			var count int
			if err = s.DB().QueryRow(`select count(*) from turn_events where turn_id='t' and event_type='turn.recovered'`).Scan(&count); err != nil || count != 0 {
				t.Fatal(count, err)
			}
			if _, err = s.DB().Exec(`drop trigger reject_hold`); err != nil {
				t.Fatal(err)
			}
			if err = s.CommitInterruptedClose(ctx, "s", "t", "wrong", failure); !errors.Is(err, ErrQueueConflict) {
				t.Fatal("foreign committer", err)
			}
			if err = s.CommitInterruptedClose(ctx, "s", "t", "closing", failure); err != nil {
				t.Fatal(err)
			}
			if err = s.Close(); err != nil {
				t.Fatal(err)
			}
			s, err = Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			row, err = s.GetTurn(ctx, "t")
			if err != nil || row.Status != "failed" || row.Phase != "held_for_retry_or_skip" || row.FinishedAt == "" {
				t.Fatal(row, err)
			}
			hold, err := s.GetTurnFailure(ctx, "t")
			if err != nil || hold.HoldState != "review" || hold.FailureKind != failure.FailureKind {
				t.Fatal(hold, err)
			}
			if receipt, err := s.HasTurnTerminalReceipt(ctx, "t"); err != nil || receipt {
				t.Fatal("invented terminal receipt", receipt, err)
			}
			if _, _, err = s.GetSessionActiveTurn(ctx, "s"); !errors.Is(err, sql.ErrNoRows) {
				t.Fatal("claim retained after join", err)
			}
			if _, err = s.GetNextQueuedTurn(ctx, "s"); !errors.Is(err, sql.ErrNoRows) {
				t.Fatal("whole prompt queued", err)
			}
		})
	}
}

func TestInterruptedClosePreservesUserCancellation(t *testing.T) {
	s, ctx, _ := generationReadyFixture(t)
	_, old, err := s.GetSessionActiveTurn(ctx, "s")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().Exec(`update turns set status='cancelling',phase='cancelling' where id='t'`); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ReserveInterruptedClose(ctx, "s", "t", old, "closing"); err != nil || ok {
		t.Fatal(ok, err)
	}
	_, token, err := s.GetSessionActiveTurn(ctx, "s")
	if err != nil || token != old {
		t.Fatal(token, err)
	}
}

func TestInterruptedCloseSealsFencedProviderOutcomeAtomically(t *testing.T) {
	s, ctx, _ := generationReadyFixture(t)
	base, err := s.TurnAccounting(ctx, "t")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.AppendTurnEvent(ctx, "t", "s", "inference.started", map[string]any{"request_id": "pending", "iteration": 2}); err != nil {
		t.Fatal(err)
	}
	payload := partialPayload("committed prefix")
	payload["provider_received_usage"] = &goai.Usage{Input: 23}
	if err = s.SaveProviderPartial(ctx, "t", "s", "pending", "committed prefix", payload); err != nil {
		t.Fatal(err)
	}
	_, old, err := s.GetSessionActiveTurn(ctx, "s")
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ReserveInterruptedClose(ctx, "s", "t", old, "closing"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if _, err = s.DB().Exec(`create trigger reject_hold before update on turns when new.phase='held_for_retry_or_skip' begin select raise(abort,'hold rejected');end`); err != nil {
		t.Fatal(err)
	}
	failure := TurnFailure{FailureKind: "recovery_interrupted", Summary: "joined fenced outcome"}
	if err = s.CommitInterruptedClose(ctx, "s", "t", "closing", failure); err == nil {
		t.Fatal("hold fault lost")
	}
	after, err := s.TurnAccounting(ctx, "t")
	if err != nil || after.Usage.Input != base.Usage.Input {
		t.Fatal("receipt escaped failed hold transaction", after, err)
	}
	if _, err = s.DB().Exec(`drop trigger reject_hold`); err != nil {
		t.Fatal(err)
	}
	if err = s.CommitInterruptedClose(ctx, "s", "t", "closing", failure); err != nil {
		t.Fatal(err)
	}
	after, err = s.TurnAccounting(ctx, "t")
	if err != nil || after.Usage.Input != base.Usage.Input+23 || after.Unknown != base.Unknown+1 {
		t.Fatal("fenced received outcome lost", base, after, err)
	}
}
