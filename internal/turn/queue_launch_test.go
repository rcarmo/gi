package turn

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestLaunchTurnLockedSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_launch_ctx", "Test", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_launch_ctx", "session_launch_ctx", "queued", "hello", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create queued turn: %v", err)
	}
	engine := New(s)
	defer engine.Close()
	runner := engine.runner("session_launch_ctx")
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	launched, err := engine.launchTurnLocked(cancelCtx, runner, "session_launch_ctx", queuedTurn.ID)
	if err != nil {
		t.Fatalf("launch turn with canceled caller context: %v", err)
	}
	if !launched {
		t.Fatalf("expected queued turn to launch despite canceled caller context")
	}
	waitForCondition(t, 2*time.Second, func() bool {
		turnRec, err := s.GetTurn(ctx, queuedTurn.ID)
		return err == nil && strings.TrimSpace(turnRec.FinishedAt) != ""
	}, "launched turn completion with canceled caller context")
	turnRec, err := s.GetTurn(ctx, queuedTurn.ID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if turnRec.Status != "completed" {
		t.Fatalf("expected launched turn to complete, got %#v", turnRec)
	}
	// FinishedAt is written before cleanupTurnRun releases the claim. Wait for
	// that cleanup boundary before asserting the exact absence of active work.
	waitForCondition(t, 2*time.Second, func() bool {
		_, _, err := s.GetSessionActiveTurn(ctx, "session_launch_ctx")
		return err == sql.ErrNoRows
	}, "launched turn claim cleanup")
	if _, _, err := s.GetSessionActiveTurn(ctx, "session_launch_ctx"); err != sql.ErrNoRows {
		t.Fatalf("expected no lingering active turn after launched completion, got err=%v", err)
	}
}

func TestConvertLaunchConflictToSteeringSurvivesCanceledCallerContext(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_claim_conflict_ctx", "Test", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_existing_active_ctx", "session_claim_conflict_ctx", "running", "already running", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create existing active turn: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_transient_ctx", "session_claim_conflict_ctx", "queued", "steer me", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create transient queued turn: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_claim_conflict_ctx", activeTurn.ID, "runner", activeTurn.ID); err != nil {
		t.Fatalf("claim existing active turn: %v", err)
	}
	if err := s.TouchSessionState(ctx, "session_claim_conflict_ctx", map[string]any{"active_turn_id": activeTurn.ID, "status": "running"}); err != nil {
		t.Fatalf("touch session state: %v", err)
	}
	engine := New(s)
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	res, steered, err := engine.convertLaunchConflictToSteering(cancelCtx, queuedTurn.ID, RunInput{SessionID: "session_claim_conflict_ctx", Prompt: "steer me", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("convert launch conflict with canceled caller context: %v", err)
	}
	if !steered || res == nil || res.TurnID != activeTurn.ID || res.Status != "running" {
		t.Fatalf("expected steering fallback to existing active turn, got steered=%v res=%#v", steered, res)
	}
	turns, err := s.ListTurns(ctx, "session_claim_conflict_ctx")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 1 || turns[0].ID != activeTurn.ID {
		t.Fatalf("expected transient queued turn removed after canceled-context steering fallback, got %#v", turns)
	}
	if depth, err := s.SteeringQueueLength(ctx, "session_claim_conflict_ctx"); err != nil {
		t.Fatalf("steering queue length: %v", err)
	} else if depth != 1 {
		t.Fatalf("expected steering queue depth 1 after canceled-context claim-conflict fallback, got %d", depth)
	}
}

func TestConvertLaunchConflictToSteeringPreservesSuccessWhenTransientTurnCleanupFails(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_claim_conflict_cleanup", "Test", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	activeTurn, err := s.CreateTurnWithStatus(ctx, "turn_existing_active_cleanup", "session_claim_conflict_cleanup", "running", "already running", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create existing active turn: %v", err)
	}
	queuedTurn, err := s.CreateTurnWithStatus(ctx, "turn_transient_cleanup", "session_claim_conflict_cleanup", "queued", "steer me", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatalf("create transient queued turn: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_claim_conflict_cleanup", activeTurn.ID, "runner", activeTurn.ID); err != nil {
		t.Fatalf("claim existing active turn: %v", err)
	}
	if err := s.TouchSessionState(ctx, "session_claim_conflict_cleanup", map[string]any{"active_turn_id": activeTurn.ID, "status": "running"}); err != nil {
		t.Fatalf("touch session state: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `
		create trigger fail_delete_transient_turn_cleanup
		before delete on turns
		for each row when old.id = 'turn_transient_cleanup'
		begin
			select raise(fail, 'delete blocked for test');
		end;
	`); err != nil {
		t.Fatalf("create delete-failure trigger: %v", err)
	}
	engine := New(s)
	res, steered, err := engine.convertLaunchConflictToSteering(ctx, queuedTurn.ID, RunInput{SessionID: "session_claim_conflict_cleanup", Prompt: "steer me", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("expected steering fallback success despite transient cleanup failure, got %v", err)
	}
	if !steered || res == nil || res.TurnID != activeTurn.ID || res.Status != "running" {
		t.Fatalf("expected steering fallback to existing active turn, got steered=%v res=%#v", steered, res)
	}
}
