package store

import (
	"context"
	"database/sql"
	"fmt"
	goai "github.com/rcarmo/go-ai"
	"testing"
	"time"
)

func parkedCheckpointFixture(t *testing.T, kind string) (*Store, context.Context) {
	t.Helper()
	var s *Store
	var old context.Context
	if kind == "tool" {
		var call goai.ToolCall
		s, old, _, call = toolBatchFixture(t)
		startToolBatchCall(t, s, old, call)
	} else {
		s, old, _ = generationReadyFixture(t)
		if err := s.SaveGenerationReady(old, "s", "t", readyCheckpoint()); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < generationReadyRecoveryLimit; i++ {
		claim := staleToolBatchClaim(t, s)
		token := fmt.Sprintf("recovery-%d", i)
		if kind == "tool" {
			if state, err := s.AcquireToolBatchRecovery(context.Background(), claim, token, time.Now()); err != nil || state == nil {
				t.Fatal(state, err)
			}
		} else if state, err := s.AcquireGenerationReadyRecovery(context.Background(), claim, token, time.Now()); err != nil || state == nil {
			t.Fatal(state, err)
		}
	}
	if kind == "provider" {
		writer := WithTurnWriter(context.Background(), "s", "t", fmt.Sprintf("recovery-%d", generationReadyRecoveryLimit-1))
		if err := s.AdmitTurnEffect(writer, "s", "t", "running", "inference.started", map[string]any{"request_id": "pending", "iteration": 2, "phase": "inference"}); err != nil {
			t.Fatal(err)
		}
	}
	claim := staleToolBatchClaim(t, s)
	if kind == "tool" {
		if state, err := s.AcquireToolBatchRecovery(context.Background(), claim, "over-limit", time.Now()); err != nil || state != nil {
			t.Fatal(state, err)
		}
	} else if state, err := s.AcquireGenerationReadyRecovery(context.Background(), claim, "over-limit", time.Now()); err != nil || state != nil {
		t.Fatal(state, err)
	}
	failure, disposition := "recovery_interrupted", "hold_for_retry_or_skip_after_interruption"
	if kind == "tool" {
		failure, disposition = "recovery_interrupted_tool_phase", "hold_for_retry_or_skip_after_tool_checkpoint"
	}
	if err := s.MarkTurnFailureWithFallbackErr(context.Background(), nil, "t", "s", failure, "review", "crash budget exhausted"); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateTurnStatusAndPhase(context.Background(), "t", "failed", "held_for_retry_or_skip"); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkTurnFinished(context.Background(), "t"); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendTurnEvent(context.Background(), "t", "s", "turn.recovered", map[string]any{"phase": "recovery", "checkpoint": true, "stale_claim": true, "previous_status": "running", "previous_phase": claim.Phase, "recovery_disposition": disposition}); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseSessionActiveTurn(context.Background(), "s", claim.ClaimToken); err != nil {
		t.Fatal(err)
	}
	return s, old
}

func TestCheckpointRetryReopensOriginalAndResetsOnlyCrashBudget(t *testing.T) {
	for _, kind := range []string{"generation", "provider", "tool"} {
		t.Run(kind, func(t *testing.T) {
			s, old := parkedCheckpointFixture(t, kind)
			ctx := context.Background()
			claim, handled, err := s.QueueHeldCheckpointRetry(ctx, "t", "explicit-owner", "operator retry")
			if err != nil || !handled || claim == nil {
				t.Fatal(claim, handled, err)
			}
			if _, err = s.GetTurnFailure(ctx, "t"); err != sql.ErrNoRows {
				t.Fatal("hold not atomically cleared", err)
			}
			if _, err = s.GetNextQueuedTurn(ctx, "s"); err != sql.ErrNoRows {
				t.Fatal("checkpoint entered ordinary prompt queue", err)
			}
			if err = s.AdmitTurnEffect(old, "s", "t", "running", "inference.started", map[string]any{"request_id": "stale"}); err == nil {
				t.Fatal("original writer revived")
			}
			duplicate, again, err := s.QueueHeldCheckpointRetry(ctx, "t", "duplicate-owner", "duplicate")
			if err != nil || !again || duplicate != nil {
				t.Fatal("duplicate admission", duplicate, again, err)
			}
			var seq int
			var database, path string
			if err = s.db.QueryRow(`pragma database_list`).Scan(&seq, &database, &path); err != nil {
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
			if duplicate, again, err = s.QueueHeldCheckpointRetry(ctx, "t", "after-restart", "duplicate after restart"); err != nil || !again || duplicate != nil {
				t.Fatal(duplicate, again, err)
			}
			claims, err := s.ListStaleActiveTurnClaims(ctx, time.Now().Add(-time.Hour), "s")
			if err != nil || len(claims) != 1 || claims[0].ClaimToken != claim.ClaimToken {
				t.Fatal("fresh checkpoint not resumable", claims, err)
			}
			var turns, retries int
			if err = s.db.QueryRow(`select count(*) from turns where session_id='s'`).Scan(&turns); err != nil {
				t.Fatal(err)
			}
			if err = s.db.QueryRow(`select count(*) from turn_events where turn_id='t' and event_type='turn.checkpoint_retry_requested'`).Scan(&retries); err != nil {
				t.Fatal(err)
			}
			if turns != 1 || retries != 1 {
				t.Fatal("prompt cloned or retry duplicated", turns, retries)
			}
			if kind == "tool" {
				if cp, err := s.AcquireToolBatchRecovery(ctx, *claim, "resumed-owner", time.Now().Add(-time.Hour)); err != nil || cp == nil {
					t.Fatal(cp, err)
				}
			} else if cp, err := s.AcquireGenerationReadyRecovery(ctx, *claim, "resumed-owner", time.Now().Add(-time.Hour)); err != nil || cp == nil || cp.Iteration != 2 || cp.RepeatedToolFailures != 1 {
				t.Fatal(cp, err)
			}
			for i := 0; i < generationReadyRecoveryLimit; i++ {
				fresh := staleToolBatchClaim(t, s)
				if kind == "tool" {
					if cp, err := s.AcquireToolBatchRecovery(ctx, fresh, fmt.Sprintf("new-crash-%d", i), time.Now()); err != nil || cp == nil {
						t.Fatal(cp, err)
					}
				} else if cp, err := s.AcquireGenerationReadyRecovery(ctx, fresh, fmt.Sprintf("new-crash-%d", i), time.Now()); err != nil || cp == nil {
					t.Fatal(cp, err)
				}
			}
		})
	}
}

func TestCheckpointRetryRejectsUnprovenFrontierAtomically(t *testing.T) {
	for _, mode := range []string{"unknown", "child", "child-link", "child-consumed", "terminal", "active", "fault-event", "fault-queue", "fault-provider-queue"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			kind := "tool"
			if mode == "fault-provider-queue" {
				kind = "provider"
			}
			s, _ := parkedCheckpointFixture(t, kind)
			var err error
			switch mode {
			case "unknown":
				err = s.AppendTurnEvent(context.Background(), "t", "s", "future.effect", nil)
			case "child":
				_, err = s.db.Exec(`update turns set metadata_json='{"parent_turn_id":"parent"}' where id='t'`)
			case "child-link", "child-consumed":
				if _, err = s.CreateSession(ctx, "P", "P", nil); err != nil {
					t.Fatal(err)
				}
				if _, err = s.CreateTurnWithStatus(ctx, "parent", "P", "running", "parent", nil); err != nil {
					t.Fatal(err)
				}
				_, err = s.CreateSubTurn(ctx, "parent", "P", "t", "s", "sync", 1, nil)
				if err == nil && mode == "child-consumed" {
					err = s.AppendTurnEvent(ctx, "t", "s", "turn.finished", map[string]any{"status": "failed"})
				}
			case "terminal":
				err = s.AppendTurnEvent(context.Background(), "t", "s", "turn.terminal_staged", map[string]any{"status": "failed"})
			case "active":
				_, err = s.db.Exec(`insert into session_active_turns(session_id,turn_id,claim_token,claimed_at,updated_at) values('s','t','other-owner',` + defaultNow + `,` + defaultNow + `)`)
			case "fault-event":
				_, err = s.db.Exec(`create trigger reject_retry before insert on turn_events when new.event_type='turn.checkpoint_retry_requested' begin select raise(abort,'reject retry');end`)
			case "fault-queue":
				_, err = s.db.Exec(`create trigger reject_retry before update on turns when new.phase='checkpoint_tool_batch' begin select raise(abort,'reject queue');end`)
			case "fault-provider-queue":
				_, err = s.db.Exec(`create trigger reject_retry before update on turns when new.phase='checkpoint_generation_ready' begin select raise(abort,'reject queue');end`)
			}
			if err != nil {
				t.Fatal(err)
			}
			if claim, handled, err := s.QueueHeldCheckpointRetry(context.Background(), "t", "retry-owner", "retry"); err == nil || !handled || claim != nil {
				t.Fatal("unsafe retry", claim, handled, err)
			}
			row, err := s.GetTurn(context.Background(), "t")
			if err != nil || row.Status != "failed" || row.Phase != "held_for_retry_or_skip" {
				t.Fatal(row, err)
			}
			failure, err := s.GetTurnFailure(context.Background(), "t")
			if err != nil || failure.HoldState != "review" {
				t.Fatal(failure, err)
			}
			var count int
			if err = s.db.QueryRow(`select count(*) from turn_events where turn_id='t' and event_type='turn.checkpoint_retry_requested'`).Scan(&count); err != nil || count != 0 {
				t.Fatal("failed admission reset budget", count, err)
			}
		})
	}
}

func TestCheckpointRetryConcurrentAdmission(t *testing.T) {
	s, _ := parkedCheckpointFixture(t, "tool")
	type outcome struct {
		claim   *ActiveTurnClaim
		handled bool
		err     error
	}
	results := make(chan outcome, 12)
	for i := 0; i < cap(results); i++ {
		go func(i int) {
			claim, handled, err := s.QueueHeldCheckpointRetry(context.Background(), "t", fmt.Sprintf("concurrent-%d", i), "retry")
			results <- outcome{claim, handled, err}
		}(i)
	}
	owners := 0
	for i := 0; i < cap(results); i++ {
		result := <-results
		if result.err != nil || !result.handled {
			t.Error(result)
		}
		if result.claim != nil {
			owners++
		}
	}
	if owners != 1 {
		t.Fatal("duplicate checkpoint owners", owners)
	}
}
