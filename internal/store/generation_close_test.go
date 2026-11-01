package store

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func generationCloseFixture(t *testing.T, mode string) (*Store, context.Context) {
	t.Helper()
	if mode != "ready" {
		return interruptedProviderFixture(t, mode == "partial")
	}
	s, ctx, _ := generationReadyFixture(t)
	if err := s.SaveGenerationReady(ctx, "s", "t", readyCheckpoint()); err != nil {
		t.Fatal(err)
	}
	return s, ctx
}

func TestGenerationCloseReopenFenceAndResume(t *testing.T) {
	for _, mode := range []string{"ready", "silent", "partial"} {
		t.Run(mode, func(t *testing.T) {
			s, oldCtx := generationCloseFixture(t, mode)
			ctx := context.Background()
			cp, err := s.ReserveGenerationClose(ctx, "s", "t", "epoch", "closing")
			if err != nil || cp == nil || cp.Iteration != 2 {
				t.Fatal(cp, err)
			}
			turn, err := s.GetTurn(ctx, "t")
			if err != nil || turn.Status != "running" || turn.FinishedAt != "" {
				t.Fatal("reservation published terminal/queue", turn, err)
			}
			if err = s.AdmitTurnEffect(oldCtx, "s", "t", "running", "inference.started", map[string]any{"request_id": "stale"}); err == nil {
				t.Fatal("old writer retained authority")
			}
			if err = s.CommitGenerationClose(ctx, "s", "t", "closing"); err != nil {
				t.Fatal(err)
			}
			turn, err = s.GetTurn(ctx, "t")
			if err != nil || turn.Status != "queued" || turn.Phase != ClosedGenerationReadyPhase || turn.FinishedAt != "" {
				t.Fatal(turn, err)
			}
			if _, err = s.GetNextQueuedTurn(ctx, "s"); err != sql.ErrNoRows {
				t.Fatal("ordinary queue saw checkpoint", err)
			}
			var path string
			if err = s.db.QueryRow(`select file from pragma_database_list where name='main'`).Scan(&path); err != nil || path == "" {
				t.Fatal(path, err)
			}
			if err = s.Close(); err != nil {
				t.Fatal(err)
			}
			s, err = Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			claims, err := s.ListStaleActiveTurnClaims(ctx, time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC), "s")
			if err != nil || len(claims) != 1 {
				t.Fatal("closed checkpoint must be immediately discoverable", claims, err)
			}
			cp, err = s.AcquireGenerationReadyRecovery(ctx, claims[0], "resumed", time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC))
			if err != nil || cp == nil || cp.Lifecycle.Message != "once only" || cp.LastToolFailureSignature != "prior" {
				t.Fatal(cp, err)
			}
			a, err := s.TurnAccounting(ctx, "t")
			wantUnknown, wantInput := 0, 5
			if mode != "ready" {
				wantUnknown = 1
			}
			if mode == "partial" {
				wantInput += 17
			}
			if err != nil || a.Unknown != wantUnknown || a.Usage.Input != wantInput {
				t.Fatal(a, err)
			}
			var crashed, resumed int
			if err = s.db.QueryRow(`select count(*) from turn_events where turn_id='t' and event_type='generation.recovered'`).Scan(&crashed); err != nil || crashed != 0 {
				t.Fatal("graceful restart counted as crash", crashed, err)
			}
			if err = s.db.QueryRow(`select count(*) from turn_events where turn_id='t' and event_type='generation.close_resumed'`).Scan(&resumed); err != nil || resumed != 1 {
				t.Fatal(resumed, err)
			}
			// A crash before the new host prepares anything still has a valid frontier.
			cp, err = s.AcquireGenerationReadyRecovery(ctx, staleGenerationClaim(t, s), "crash-resume", time.Now())
			if err != nil || cp == nil {
				t.Fatal(cp, err)
			}
			a, err = s.TurnAccounting(ctx, "t")
			if err != nil || a.Unknown != wantUnknown || a.Usage.Input != wantInput {
				t.Fatal("receipt duplicated", a, err)
			}
		})
	}
}

func TestGenerationCloseDoesNotResetCrashBudget(t *testing.T) {
	s, _ := generationCloseFixture(t, "ready")
	ctx := context.Background()
	for range generationReadyRecoveryLimit {
		if err := s.AppendTurnEvent(ctx, "t", "s", "generation.recovered", map[string]any{"phase": "generation_ready", "iteration": 2}); err != nil {
			t.Fatal(err)
		}
	}
	if cp, err := s.ReserveGenerationClose(ctx, "s", "t", "epoch", "closing"); err != nil || cp == nil {
		t.Fatal(cp, err)
	}
	if err := s.CommitGenerationClose(ctx, "s", "t", "closing"); err != nil {
		t.Fatal(err)
	}
	claims, err := s.ListStaleActiveTurnClaims(ctx, time.Now(), "s")
	if err != nil || len(claims) != 1 {
		t.Fatal(claims, err)
	}
	if cp, err := s.AcquireGenerationReadyRecovery(ctx, claims[0], "resumed", time.Now()); err != nil || cp == nil {
		t.Fatal("graceful restart consumed crash allowance", cp, err)
	}
	if cp, err := s.AcquireGenerationReadyRecovery(ctx, staleGenerationClaim(t, s), "crashed", time.Now()); err != nil || cp != nil {
		t.Fatal("graceful restart erased crash allowance", cp, err)
	}
}

func TestGenerationCloseCommitFaultsRollback(t *testing.T) {
	for _, target := range []string{"receipt", "event", "message", "phase", "claim", "session"} {
		t.Run(target, func(t *testing.T) {
			s, _ := generationCloseFixture(t, "partial")
			ctx := context.Background()
			if cp, err := s.ReserveGenerationClose(ctx, "s", "t", "epoch", "closing"); err != nil || cp == nil {
				t.Fatal(cp, err)
			}
			table, action := "turn_receipts", "insert"
			switch target {
			case "event":
				table = "turn_events"
			case "message":
				table, action = "messages", "update"
			case "phase":
				table, action = "turns", "update"
			case "claim":
				table, action = "session_active_turns", "update"
			case "session":
				table, action = "sessions", "update"
			}
			if _, err := s.db.Exec(fmt.Sprintf(`create trigger reject_close before %s on %s begin select raise(abort,'close rejected');end`, action, table)); err != nil {
				t.Fatal(err)
			}
			if err := s.CommitGenerationClose(ctx, "s", "t", "closing"); err == nil {
				t.Fatal("fault accepted")
			}
			var receipts int
			if err := s.db.QueryRow(`select count(*) from turn_receipts where occurrence_id='pending'`).Scan(&receipts); err != nil || receipts != 0 {
				t.Fatal("partial commit", receipts, err)
			}
			turn, err := s.GetTurn(ctx, "t")
			if err != nil || turn.Status != "running" || turn.Phase != "running" {
				t.Fatal(turn, err)
			}
			var worker string
			if err = s.db.QueryRow(`select worker_id from session_active_turns where session_id='s'`).Scan(&worker); err != nil || worker != "generation-closing" {
				t.Fatal(worker, err)
			}
			if _, err = s.db.Exec(`drop trigger reject_close`); err != nil {
				t.Fatal(err)
			}
			if err = s.CommitGenerationClose(ctx, "s", "t", "closing"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestGenerationCloseCancellationAndUnknownBoundaries(t *testing.T) {
	for _, mode := range []string{"cancel-after-reserve", "unknown", "tool", "foreign", "invalid"} {
		t.Run(mode, func(t *testing.T) {
			s, oldCtx := generationCloseFixture(t, "ready")
			ctx := context.Background()
			switch mode {
			case "unknown":
				if err := s.AppendTurnEvent(ctx, "t", "s", "unknown", map[string]any{"iteration": 2}); err != nil {
					t.Fatal(err)
				}
			case "tool":
				if err := s.AdmitTurnEffect(oldCtx, "s", "t", "waiting_on_tools", "tool.started", map[string]any{"occurrence_id": "pending-tool"}); err != nil {
					t.Fatal(err)
				}
			case "foreign":
				if _, err := s.db.Exec(`update session_active_turns set claim_token='foreign'`); err != nil {
					t.Fatal(err)
				}
			case "invalid":
				if _, err := s.db.Exec(`update turn_events set payload_json=json_set(payload_json,'$.continuation.version',9) where event_type='generation.ready'`); err != nil {
					t.Fatal(err)
				}
			}
			cp, err := s.ReserveGenerationClose(ctx, "s", "t", "epoch", "closing")
			if mode == "foreign" {
				if err == nil {
					t.Fatal("foreign owner accepted")
				}
				return
			}
			if mode != "cancel-after-reserve" {
				if err != nil || cp != nil {
					t.Fatal("unproven boundary accepted", cp, err)
				}
				return
			}
			if err != nil || cp == nil {
				t.Fatal(cp, err)
			}
			if _, err = s.db.Exec(`update turns set status='cancelling',phase='cancelling' where id='t'`); err != nil {
				t.Fatal(err)
			}
			if err = s.CommitGenerationClose(ctx, "s", "t", "closing"); err != ErrQueueConflict {
				t.Fatal("cancel overwritten", err)
			}
		})
	}
}

func TestGenerationCloseCannotAcquireWholePromptClaim(t *testing.T) {
	s, _ := generationCloseFixture(t, "ready")
	ctx := context.Background()
	if cp, err := s.ReserveGenerationClose(ctx, "s", "t", "epoch", "closing"); err != nil || cp == nil {
		t.Fatal(cp, err)
	}
	if err := s.CommitGenerationClose(ctx, "s", "t", "closing"); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseSessionActiveTurn(ctx, "s", "closing"); err != nil {
		t.Fatal(err)
	}
	if claimed, err := s.ClaimSessionActiveTurn(ctx, "s", "t", "queue", "whole-prompt"); err != nil || claimed {
		t.Fatal("released checkpoint became prompt authority", claimed, err)
	}
}

func TestGenerationCloseReservationFaultsRetainOldAuthority(t *testing.T) {
	for _, table := range []string{"session_active_turns", "turn_recovery_fences"} {
		t.Run(table, func(t *testing.T) {
			s, oldCtx := generationCloseFixture(t, "ready")
			ctx := context.Background()
			action := "update"
			if table == "turn_recovery_fences" {
				action = "insert"
			}
			closeRetryTrigger(t, s, action, table, "")
			if _, err := s.ReserveGenerationClose(ctx, "s", "t", "epoch", "closing"); err == nil {
				t.Fatal("reservation fault ignored")
			}
			_, token, err := s.GetSessionActiveTurn(ctx, "s")
			if err != nil || token != "epoch" {
				t.Fatal("failed reservation rotated authority", token, err)
			}
			if _, err = s.db.Exec(`drop trigger reject_close`); err != nil {
				t.Fatal(err)
			}
			if err = s.AdmitTurnEffect(oldCtx, "s", "t", "running", "inference.started", map[string]any{"request_id": "still-owned"}); err != nil {
				t.Fatal("failed reservation fenced owner", err)
			}
		})
	}
}

func TestGenerationCloseCrashDuringReservation(t *testing.T) {
	for _, mode := range []string{"ready", "partial"} {
		t.Run(mode, func(t *testing.T) {
			s, oldCtx := generationCloseFixture(t, mode)
			ctx := context.Background()
			if cp, err := s.ReserveGenerationClose(ctx, "s", "t", "epoch", "closing"); err != nil || cp == nil {
				t.Fatal(cp, err)
			}
			claims, err := s.ListStaleActiveTurnClaims(ctx, time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC), "s")
			if err != nil || len(claims) != 0 {
				t.Fatal("unjoined reservation published as ready", claims, err)
			}
			var path string
			if err = s.db.QueryRow(`select file from pragma_database_list where name='main'`).Scan(&path); err != nil {
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
			claim := staleGenerationClaim(t, s)
			if claim.WorkerID != "generation-closing" {
				t.Fatal(claim)
			}
			if cp, err := s.AcquireGenerationReadyRecovery(ctx, claim, "restarted", time.Now()); err != nil || cp == nil {
				t.Fatal(cp, err)
			}
			if err = s.AdmitTurnEffect(oldCtx, "s", "t", "running", "inference.started", map[string]any{"request_id": "stale"}); err == nil {
				t.Fatal("old authority survived restart")
			}
			a, err := s.TurnAccounting(ctx, "t")
			want := 0
			if mode == "partial" {
				want = 1
			}
			if err != nil || a.Unknown != want {
				t.Fatal(a, err)
			}
			var queued int
			if err = s.db.QueryRow(`select count(*) from turn_events where event_type='turn.checkpoint_queued'`).Scan(&queued); err != nil || queued != 0 {
				t.Fatal("crash fabricated graceful completion", queued, err)
			}
		})
	}
}
