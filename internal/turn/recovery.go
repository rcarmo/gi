package turn

import (
	"context"
	"database/sql"
	"errors"
	"github.com/rcarmo/gi/internal/logutil"
	"log"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

var activeTurnHeartbeatInterval = 5 * time.Second

const interruptedTurnStaleAfter = 30 * time.Second

func (e *Engine) appendRecoveryFailureEvent(ctx context.Context, claim store.ActiveTurnClaim, err error) {
	if e == nil || e.store == nil || err == nil || strings.TrimSpace(claim.SessionID) == "" || strings.TrimSpace(claim.TurnID) == "" {
		return
	}
	payload := map[string]any{
		"phase":                "recovery",
		"checkpoint":           true,
		"reason":               "recovery_failed",
		"error":                err.Error(),
		"previous_status":      claim.Status,
		"previous_phase":       claim.Phase,
		"recovery_disposition": recoveryDispositionForClaim(claim),
		"stale_claim":          true,
	}
	logutil.WarnIfErr("append recovery failure event", e.store.AppendTurnEvent(ctx, claim.TurnID, claim.SessionID, "turn.recovery_failed", payload))
	runner := e.runner(claim.SessionID)
	agentID, model := "", ""
	if turnRec, turnErr := e.store.GetTurn(ctx, claim.TurnID); turnErr == nil {
		agentID, model = runner.resolveTurnAgentAndModel(ctx, e.store, turnRec, claim.SessionID, turnRec.Prompt)
	}
	e.PublishRuntimeTurnEvent("turn_recovery_failed", claim.SessionID, claim.TurnID, agentID, claim.Status, claim.Phase, cloneMap(payload))
	runner.emitTurnStateHook(ctx, claim.SessionID, claim.TurnID, agentID, model, claim.Status, claim.Phase, cloneMap(payload))
	runner.emitSessionStateHook(ctx, claim.SessionID, agentID, model, "running", map[string]any{
		"reason":               "recovery_failed",
		"error":                err.Error(),
		"recovery_disposition": recoveryDispositionForClaim(claim),
		"stale_claim":          true,
		"active_turn_id":       claim.TurnID,
		"turn_id":              claim.TurnID,
		"turn_status":          claim.Status,
		"turn_phase":           claim.Phase,
	})
}

func (e *Engine) recoverInterruptedTurns(ctx context.Context, sessionID string) (bool, error) {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	claims, err := e.store.ListStaleActiveTurnClaims(opCtx, time.Now().Add(-interruptedTurnStaleAfter), sessionID)
	if err != nil {
		return false, err
	}
	type recoveryScanCounts struct {
		recovered int
		failed    int
	}
	recovered := false
	var firstErr error
	sessionsToRestart := map[string]bool{}
	sessionCounts := map[string]*recoveryScanCounts{}
	failedClaimsBySession := map[string][]store.ActiveTurnClaim{}
	for _, claim := range claims {
		counts := sessionCounts[claim.SessionID]
		if counts == nil {
			counts = &recoveryScanCounts{}
			sessionCounts[claim.SessionID] = counts
		}
		if err := e.recoverInterruptedTurn(opCtx, claim); err != nil {
			log.Printf("turn recovery: recover %s/%s failed: %v", claim.SessionID, claim.TurnID, err)
			e.appendRecoveryFailureEvent(opCtx, claim, err)
			failedClaimsBySession[claim.SessionID] = append(failedClaimsBySession[claim.SessionID], claim)
			counts.failed++
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		counts.recovered++
		recovered = true
		queueCount, err := e.store.CountQueuedTurns(opCtx, claim.SessionID)
		if err != nil {
			return recovered, err
		}
		if queueCount > 0 {
			sessionsToRestart[claim.SessionID] = true
		}
	}
	for failedSessionID, counts := range sessionCounts {
		if counts.failed > 0 {
			e.emitRecoveryScanFailureSessionState(opCtx, failedSessionID, counts.recovered, counts.failed)
			for _, claim := range failedClaimsBySession[failedSessionID] {
				logutil.WarnIfErr("append recovery scan summary event", e.store.AppendTurnEvent(opCtx, claim.TurnID, failedSessionID, "turn.recovery_scan_failed", map[string]any{
					"phase":                 "recovery",
					"checkpoint":            true,
					"reason":                "recovery_scan_failed",
					"recovery_disposition":  recoveryDispositionForClaim(claim),
					"recovered_claim_count": counts.recovered,
					"failed_claim_count":    counts.failed,
				}))
			}
		}
	}
	for recoveredSessionID := range sessionsToRestart {
		if err := e.startNextQueuedTurn(opCtx, recoveredSessionID); err != nil {
			e.emitRecoveryRestartFailureSessionState(opCtx, recoveredSessionID, err)
			return recovered, err
		}
	}
	if firstErr != nil {
		return recovered, firstErr
	}
	return recovered, nil
}

func (e *Engine) emitRecoveryScanFailureSessionState(ctx context.Context, sessionID string, recoveredCount, failedCount int) {
	if e == nil || e.store == nil || strings.TrimSpace(sessionID) == "" || failedCount == 0 {
		return
	}
	runner := e.runner(sessionID)
	agentID, model := "", ""
	activeTurnID, _, activeErr := e.store.GetSessionActiveTurn(ctx, sessionID)
	if activeErr != nil && activeErr != sql.ErrNoRows {
		return
	}
	queueCount, err := e.store.CountQueuedTurns(ctx, sessionID)
	if err != nil {
		return
	}
	status := "idle"
	activeTurnValue := any(nil)
	if activeErr == nil {
		status = "running"
		activeTurnValue = activeTurnID
		if turnRec, turnErr := e.store.GetTurn(ctx, activeTurnID); turnErr == nil {
			agentID, model = runner.resolveTurnAgentAndModel(ctx, e.store, turnRec, sessionID, turnRec.Prompt)
		}
	} else if queueCount > 0 {
		status = "queued"
	}
	if model == "" {
		if sessRec, sessErr := e.store.GetSession(ctx, sessionID); sessErr == nil {
			model = store.StringValue(sessRec.State["model"], "")
		}
	}
	runner.emitSessionStateHook(ctx, sessionID, agentID, model, status, map[string]any{
		"reason":                "recovery_scan_failed",
		"active_turn_id":        activeTurnValue,
		"queue_count":           queueCount,
		"recovered_claim_count": recoveredCount,
		"failed_claim_count":    failedCount,
	})
}

func (e *Engine) emitRecoveryRestartFailureSessionState(ctx context.Context, sessionID string, err error) {
	if e == nil || e.store == nil || strings.TrimSpace(sessionID) == "" || err == nil {
		return
	}
	runner := e.runner(sessionID)
	agentID, model := "", ""
	activeTurnID, _, activeErr := e.store.GetSessionActiveTurn(ctx, sessionID)
	if activeErr != nil && activeErr != sql.ErrNoRows {
		return
	}
	queueCount, countErr := e.store.CountQueuedTurns(ctx, sessionID)
	if countErr != nil {
		return
	}
	status := "idle"
	activeTurnValue := any(nil)
	if activeErr == nil {
		status = "running"
		activeTurnValue = activeTurnID
		if turnRec, turnErr := e.store.GetTurn(ctx, activeTurnID); turnErr == nil {
			agentID, model = runner.resolveTurnAgentAndModel(ctx, e.store, turnRec, sessionID, turnRec.Prompt)
		}
	} else if queueCount > 0 {
		status = "queued"
	}
	if model == "" {
		if sessRec, sessErr := e.store.GetSession(ctx, sessionID); sessErr == nil {
			model = store.StringValue(sessRec.State["model"], "")
		}
	}
	payload := map[string]any{
		"phase":          "recovery",
		"checkpoint":     true,
		"reason":         "recovery_restart_failed",
		"error":          err.Error(),
		"active_turn_id": activeTurnValue,
		"queue_count":    queueCount,
	}
	if queuedTurn, queuedErr := e.store.GetNextQueuedTurn(ctx, sessionID); queuedErr == nil {
		logutil.WarnIfErr("append recovery restart failure event", e.store.AppendTurnEvent(ctx, queuedTurn.ID, sessionID, "turn.recovery_restart_failed", cloneMap(payload)))
	}
	runner.emitSessionStateHook(ctx, sessionID, agentID, model, status, payload)
}

func recoveryDispositionForClaim(claim store.ActiveTurnClaim) string {
	switch claim.Phase {
	case "waiting_on_tools":
		return "hold_for_retry_or_skip_after_tool_checkpoint"
	case "cancelling":
		return "abort_cancelling"
	case "completed", "failed", "aborted", "cancelled":
		return "release_terminal"
	default:
		if claim.Phase == "compacting" {
			return "requeue_after_compaction_checkpoint"
		}
		return "requeue_interrupted_turn"
	}
}

func (e *Engine) recoverInterruptedTurn(ctx context.Context, claim store.ActiveTurnClaim) error {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	disposition := recoveryDispositionForClaim(claim)
	status := claim.Status
	phase := claim.Phase
	markFinished := false

	switch claim.Phase {
	case "waiting_on_tools":
		status = "failed"
		phase = "held_for_retry_or_skip"
		markFinished = true
	case "cancelling":
		status = "aborted"
		phase = "aborted"
		markFinished = true
	case "completed", "failed", "aborted", "cancelled":
		// Terminal turn with a stale claim: just release the claim.
	default:
		status = "queued"
		phase = "queued"
	}

	payload := map[string]any{
		"phase":                "recovery",
		"checkpoint":           true,
		"reason":               "recovery",
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
			if err := e.store.MarkTurnFailureWithFallbackErr(e.backgroundContext(), nil, claim.TurnID, claim.SessionID, "recovery_interrupted_tool_phase", "review", "Recovered stale turn that was interrupted while waiting on tool results"); err != nil {
				return err
			}
		}
		if err := e.store.UpdateTurnStatusAndPhase(opCtx, claim.TurnID, status, phase); err != nil {
			return err
		}
		if markFinished {
			if err := e.store.MarkTurnFinished(opCtx, claim.TurnID); err != nil {
				return err
			}
		}
	}
	logutil.WarnIfErr("turn recovered event append", e.store.AppendTurnEvent(opCtx, claim.TurnID, claim.SessionID, "turn.recovered", payload))
	if err := e.store.ReleaseSessionActiveTurn(opCtx, claim.SessionID, claim.ClaimToken); err != nil {
		return err
	}
	if err := e.store.SyncSessionQueueCount(opCtx, claim.SessionID); err != nil {
		return err
	}
	queueCount, err := e.store.CountQueuedTurns(opCtx, claim.SessionID)
	if err != nil {
		return err
	}
	sessionStatus := "idle"
	if queueCount > 0 {
		sessionStatus = "queued"
	}
	if err := e.store.TouchSessionState(opCtx, claim.SessionID, map[string]any{"active_turn_id": nil, "status": sessionStatus}); err != nil {
		return err
	}
	turnRec, err := e.store.GetTurn(opCtx, claim.TurnID)
	if err != nil {
		return err
	}
	runner := e.runner(claim.SessionID)
	agentID, model := runner.resolveTurnAgentAndModel(opCtx, e.store, turnRec, claim.SessionID, turnRec.Prompt)
	e.PublishRuntimeTurnEvent("turn_recovered", claim.SessionID, claim.TurnID, agentID, status, phase, cloneMap(payload))
	runner.emitTurnStateHook(opCtx, claim.SessionID, claim.TurnID, agentID, model, status, phase, cloneMap(payload))
	runner.emitSessionStateHook(opCtx, claim.SessionID, agentID, model, sessionStatus, map[string]any{"reason": "recovery", "recovery_disposition": disposition, "stale_claim": true, "active_turn_id": nil, "turn_id": claim.TurnID, "turn_status": status, "turn_phase": phase})
	return nil
}

func (e *Engine) startNextQueuedTurnLocked(ctx context.Context, runner *sessionRunner, sessionID string) (bool, error) {
	if sessionID == "" {
		return false, nil
	}
	coordCtx := coordinationContext(ctx, e.backgroundContext())
	if activeTurnID, _, err := e.store.GetSessionActiveTurn(coordCtx, sessionID); err == nil {
		if err := e.normalizeRunningSessionState(coordCtx, sessionID, activeTurnID, true, ""); err != nil {
			return false, err
		}
		return false, nil
	} else if err != sql.ErrNoRows {
		return false, err
	}
	next, err := e.store.GetNextQueuedTurn(coordCtx, sessionID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if activeTurnID, _, err := e.store.GetSessionActiveTurn(coordCtx, sessionID); err == nil {
		if err := e.normalizeRunningSessionState(coordCtx, sessionID, activeTurnID, true, ""); err != nil {
			return false, err
		}
		return false, nil
	} else if err != sql.ErrNoRows {
		return false, err
	}
	launched, err := e.launchTurnLocked(coordCtx, runner, sessionID, next.ID)
	if err != nil {
		return false, err
	}
	return launched, nil
}

func (e *Engine) startNextQueuedTurn(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	runner := e.runner(sessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	_, err := e.startNextQueuedTurnLocked(ctx, runner, sessionID)
	return err
}

func (r *sessionRunner) heartbeatActiveTurn(ctx context.Context, sessionID, claimToken string, cancel context.CancelFunc) {
	bgCtx := r.engine.backgroundContext()
	ticker := time.NewTicker(activeTurnHeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.store.TouchSessionActiveTurn(bgCtx, sessionID, claimToken); err != nil {
				log.Printf("turn heartbeat: %v", err)
				if errors.Is(err, sql.ErrNoRows) {
					cancel()
					return
				}
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
