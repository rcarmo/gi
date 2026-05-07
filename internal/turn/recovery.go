package turn

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

const (
	activeTurnHeartbeatInterval = 5 * time.Second
	interruptedTurnStaleAfter   = 30 * time.Second
)

func (e *Engine) recoverInterruptedTurns(ctx context.Context, sessionID string) (bool, error) {
	claims, err := e.store.ListStaleActiveTurnClaims(ctx, time.Now().Add(-interruptedTurnStaleAfter), sessionID)
	if err != nil {
		return false, err
	}
	recovered := false
	for _, claim := range claims {
		if err := e.recoverInterruptedTurn(ctx, claim); err != nil {
			log.Printf("turn recovery: recover %s/%s failed: %v", claim.SessionID, claim.TurnID, err)
			continue
		}
		recovered = true
	}
	return recovered, nil
}

func (e *Engine) recoverInterruptedTurn(ctx context.Context, claim store.ActiveTurnClaim) error {
	disposition := "release_terminal"
	status := claim.Status
	phase := claim.Phase
	markFinished := false

	switch claim.Phase {
	case "waiting_on_tools":
		disposition = "hold_for_retry_or_skip_after_tool_checkpoint"
		status = "failed"
		phase = "held_for_retry_or_skip"
		markFinished = true
	case "cancelling":
		disposition = "abort_cancelling"
		status = "aborted"
		phase = "aborted"
		markFinished = true
	case "completed", "failed", "aborted":
		// Terminal turn with a stale claim: just release the claim.
	default:
		if claim.Phase == "compacting" {
			disposition = "requeue_after_compaction_checkpoint"
		} else {
			disposition = "requeue_interrupted_turn"
		}
		status = "queued"
		phase = "queued"
	}

	payload := map[string]any{
		"phase":                "recovery",
		"checkpoint":           true,
		"previous_status":      claim.Status,
		"previous_phase":       claim.Phase,
		"recovery_disposition": disposition,
		"stale_claim":          true,
	}
	if staleFor, err := staleDurationSeconds(claim.UpdatedAt); err == nil {
		payload["stale_for_seconds"] = staleFor
	}

	if disposition != "release_terminal" {
		if disposition == "hold_for_retry_or_skip_after_tool_checkpoint" {
			markTurnFailureWithHold(e.store, claim.TurnID, claim.SessionID, "recovery_interrupted_tool_phase", "review", "Recovered stale turn that was interrupted while waiting on tool results")
		}
		if err := e.store.UpdateTurnStatusAndPhase(ctx, claim.TurnID, status, phase); err != nil {
			return err
		}
		if markFinished {
			if err := e.store.MarkTurnFinished(ctx, claim.TurnID); err != nil {
				return err
			}
		}
	}
	warnStore("turn recovered event append", e.store.AppendTurnEvent(ctx, claim.TurnID, claim.SessionID, "turn.recovered", payload))
	if err := e.store.ReleaseSessionActiveTurn(ctx, claim.SessionID, claim.ClaimToken); err != nil {
		return err
	}
	if err := e.store.SyncSessionQueueCount(ctx, claim.SessionID); err != nil {
		return err
	}
	queueCount, err := e.store.CountQueuedTurns(ctx, claim.SessionID)
	if err != nil {
		return err
	}
	sessionStatus := "idle"
	if queueCount > 0 {
		sessionStatus = "queued"
	}
	return e.store.TouchSessionState(ctx, claim.SessionID, map[string]any{"active_turn_id": nil, "status": sessionStatus})
}

func (e *Engine) startNextQueuedTurn(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	if _, _, err := e.store.GetSessionActiveTurn(ctx, sessionID); err == nil {
		return nil
	} else if err != sql.ErrNoRows {
		return err
	}
	next, err := e.store.GetNextQueuedTurn(ctx, sessionID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	runner := e.runner(sessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if _, _, err := e.store.GetSessionActiveTurn(ctx, sessionID); err == nil {
		return nil
	} else if err != sql.ErrNoRows {
		return err
	}
	launched, err := e.launchTurnLocked(ctx, runner, sessionID, next.ID)
	if err != nil {
		return err
	}
	if !launched {
		return nil
	}
	return nil
}

func (r *sessionRunner) heartbeatActiveTurn(ctx context.Context, sessionID, claimToken string) {
	ticker := time.NewTicker(activeTurnHeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.store.TouchSessionActiveTurn(context.Background(), sessionID, claimToken); err != nil {
				log.Printf("turn heartbeat: %v", err)
			}
		}
	}
}

func staleDurationSeconds(updatedAt string) (float64, error) {
	ts, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return 0, err
	}
	return time.Since(ts).Seconds(), nil
}
