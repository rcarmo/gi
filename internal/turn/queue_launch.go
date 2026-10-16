package turn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/rcarmo/gi/internal/logutil"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/store/internalx"
)

func (e *Engine) cancelTurn(ctx context.Context, sessionID, turnID string, queuedOnly, activeOnly bool) error {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	turn, err := e.store.GetTurn(opCtx, turnID)
	if err != nil {
		return err
	}
	turnSessionID := turn.SessionID
	if strings.TrimSpace(sessionID) != "" && turnSessionID != sessionID {
		return fmt.Errorf("turn %s does not belong to session %s", turnID, sessionID)
	}
	runner := e.runner(turnSessionID)
	agentID, model := runner.resolveTurnAgentAndModel(opCtx, e.store, turn, turnSessionID, turn.Prompt)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if activeOnly {
		turn, err = e.store.GetTurn(opCtx, turnID)
		if err != nil {
			return err
		}
		activeID, _, claimErr := e.store.GetSessionActiveTurn(opCtx, turnSessionID)
		if claimErr != nil || activeID != turnID || runner.current == nil || runner.current.turnID != turnID || (turn.Status != "running" && turn.Status != "cancelling") {
			return store.ErrQueueConflict
		}
	}
	if queuedOnly {
		turn, err = e.store.GetTurn(opCtx, turnID)
		if err != nil {
			return err
		}
		if turn.Status == "cancelled" {
			return nil
		}
		if turn.Status != "queued" || (runner.current != nil && runner.current.turnID == turnID) {
			return store.ErrQueueConflict
		}
	}
	if runner.current != nil && runner.current.turnID == turnID {
		if err := e.store.AppendTurnEvent(opCtx, turnID, turnSessionID, "turn.cancelling", map[string]any{"phase": "cancel", "checkpoint": true, "reason": "cancel_requested", "status": "cancelling", "turn_phase": "cancelling", "failure_kind": ""}); err != nil {
			return err
		}
		if err := e.store.UpdateTurnStatusAndPhase(opCtx, turnID, "cancelling", "cancelling"); err != nil {
			return err
		}
		e.PublishRuntimeTurnEvent("turn_cancelling", turnSessionID, turnID, agentID, "cancelling", "cancelling", map[string]any{"reason": "cancel_requested", "failure_kind": ""})
		runner.emitTurnStateHook(opCtx, turnSessionID, turnID, agentID, model, "cancelling", "cancelling", map[string]any{"reason": "cancel_requested", "failure_kind": ""})
		runner.emitSessionStateHook(opCtx, turnSessionID, agentID, model, "running", map[string]any{"reason": "cancel_requested", "failure_kind": "", "active_turn_id": turnID, "turn_id": turnID, "turn_status": "cancelling", "turn_phase": "cancelling"})
		// The shell runtime owns platform-specific termination and output-reader
		// teardown. Cancel once here rather than racing a second process kill.
		runner.current.cancel()
		return nil
	}
	if turn.Status == "queued" {
		if queuedOnly {
			if err := e.store.CancelQueuedTurn(opCtx, turnSessionID, turnID); err != nil {
				return err
			}
		} else if err := e.store.UpdateTurnStatusAndPhase(opCtx, turnID, "cancelled", "aborted"); err != nil {
			return err
		}
		if err := e.store.MarkTurnFinished(opCtx, turnID); err != nil {
			return err
		}
		runner.emitTurnStateHook(opCtx, turnSessionID, turnID, agentID, model, "cancelled", "aborted", map[string]any{"reason": "queued_cancel"})
		e.PublishRuntimeTurnEvent("turn_terminal", turnSessionID, turnID, agentID, "cancelled", "aborted", map[string]any{"reason": "queued_cancel", "failure_kind": ""})
		logutil.WarnIfErr("sync queue count after queued cancel", e.store.SyncSessionQueueCount(opCtx, turnSessionID))
		queueCount, err := e.store.CountQueuedTurns(opCtx, turnSessionID)
		if err != nil {
			return err
		}
		activeTurnID, _, err := e.store.GetSessionActiveTurn(opCtx, turnSessionID)
		hasActiveTurn := err == nil
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		sessionStatus := "idle"
		activeTurnValue := any(nil)
		if hasActiveTurn {
			sessionStatus = "running"
			activeTurnValue = activeTurnID
		} else if queueCount > 0 {
			sessionStatus = "queued"
		}
		logutil.WarnIfErr("touch session state after queued cancel", e.store.TouchSessionState(opCtx, turnSessionID, map[string]any{"status": sessionStatus, "active_turn_id": activeTurnValue}))
		runner.emitSessionStateHook(opCtx, turnSessionID, agentID, model, sessionStatus, map[string]any{"reason": "queued_cancel", "failure_kind": "", "turn_id": turnID, "turn_status": "cancelled", "turn_phase": "aborted", "active_turn_id": activeTurnValue})
		if sessionStatus == "idle" {
			e.PublishRuntimeSessionEvent("session_idle", turnSessionID, agentID, "idle", map[string]any{"reason": "turn_terminal", "active_turn_id": nil, "failure_kind": "", "turn_id": turnID, "turn_status": "cancelled", "turn_phase": "aborted", "model": model})
		}
		if err := e.store.AppendTurnEvent(opCtx, turnID, turnSessionID, "turn.cancelled", map[string]any{"phase": "cancel", "checkpoint": true, "queued": true, "reason": "queued_cancel", "status": "cancelled", "turn_phase": "aborted", "failure_kind": ""}); err != nil {
			return err
		}
		if sessionStatus == "running" {
			if err := e.store.AppendTurnEvent(opCtx, turnID, turnSessionID, "turn.cleanup_handoff", map[string]any{"phase": "cancel", "checkpoint": true, "reason": "queued_cancel", "handoff": "active_turn", "active_turn_id": activeTurnValue}); err != nil {
				return err
			}
		}
		if sessionStatus == "queued" {
			if err := e.store.AppendTurnEvent(opCtx, turnID, turnSessionID, "turn.cleanup_handoff", map[string]any{"phase": "cancel", "checkpoint": true, "reason": "queued_cancel", "handoff": "next_queued_turn"}); err != nil {
				return err
			}
			if _, err := e.startNextQueuedTurnLocked(opCtx, runner, turnSessionID); err != nil {
				return err
			}
		}
		if sessionStatus == "idle" {
			if err := e.store.AppendTurnEvent(opCtx, turnID, turnSessionID, "turn.cleanup_handoff", map[string]any{"phase": "cancel", "checkpoint": true, "reason": "queued_cancel", "handoff": "idle_session"}); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("turn not cancellable")
}

func (e *Engine) convertLaunchConflictToSteering(ctx context.Context, turnID string, in RunInput) (*SubmitResult, bool, error) {
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	activeTurnID, _, err := e.store.GetSessionActiveTurn(opCtx, in.SessionID)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	res, err := e.submitSteeringPrompt(opCtx, in.SessionID, activeTurnID, in)
	if err != nil {
		// Keep the already-persisted queued turn as a fallback rather than dropping the prompt.
		logutil.WarnIfErr("append launch-conflict fallback preserved event", e.store.AppendTurnEvent(opCtx, turnID, in.SessionID, "turn.cleanup_handoff", map[string]any{"phase": "launch", "checkpoint": true, "reason": "launch_claim_conflict", "handoff": "queued_fallback_preserved", "active_turn_id": activeTurnID}))
		return nil, false, nil
	}
	logutil.WarnIfErr("append launch-conflict steering handoff event", e.store.AppendTurnEvent(opCtx, activeTurnID, in.SessionID, "turn.cleanup_handoff", map[string]any{"phase": "launch", "checkpoint": true, "reason": "launch_claim_conflict", "handoff": "active_turn_steering", "replaced_turn_id": turnID}))
	if err := e.store.DeleteTurn(opCtx, turnID); err != nil {
		log.Printf("turn coordination: delete transient queued turn after steering fallback failed: %v", err)
	}
	// Running-state normalization was committed with steering admission.
	return res, true, nil
}

func (r *sessionRunner) resolveTurnIdentityForFinalize(ctx context.Context, s *store.Store, sessionID, turnID string) (string, string) {
	opCtx := store.CoordinationContext(ctx, r.engine.backgroundContext())
	turnRec, err := s.GetTurn(opCtx, turnID)
	if err != nil {
		return "", ""
	}
	if strings.TrimSpace(sessionID) == "" {
		sessionID = turnRec.SessionID
	}
	return r.resolveTurnAgentAndModel(opCtx, s, turnRec, sessionID, turnRec.Prompt)
}

func (e *Engine) runner(sessionID string) *sessionRunner {
	if v, ok := e.sessions.Load(sessionID); ok {
		return v.(*sessionRunner)
	}
	coordKey := sessionCoordKey{store: e.store, sessionID: sessionID}
	coordVal, _ := sharedSessionCoords.LoadOrStore(coordKey, &sharedSessionCoord{})
	runner := &sessionRunner{sharedSessionCoord: coordVal.(*sharedSessionCoord), store: e.store, engine: e}
	actual, _ := e.sessions.LoadOrStore(sessionID, runner)
	return actual.(*sessionRunner)
}

func (e *Engine) launchTurnLocked(ctx context.Context, runner *sessionRunner, sessionID, turnID string) (bool, error) {
	return e.launchQueueActionLocked(ctx, runner, sessionID, turnID, false, "")
}

func (e *Engine) launchQueueActionLocked(ctx context.Context, runner *sessionRunner, sessionID, turnID string, selectedSteer bool, endedID string) (bool, error) {
	if e.closing.Load() {
		return false, nil // shutting down: leave queued work for the next start
	}
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if hook := e.beforeLaunchClaimHook; hook != nil {
		hook(opCtx, sessionID, turnID)
	}
	claimToken, err := store.NewActiveTurnClaimToken()
	if err != nil {
		return false, err
	}
	var claimed bool
	if endedID != "" {
		claimed, err = e.store.ClaimEndedQueueSteer(opCtx, sessionID, turnID, "runner", claimToken, endedID)
	} else if selectedSteer {
		claimed, err = e.store.ClaimIdleQueueAction(opCtx, sessionID, turnID, "runner", claimToken)
	} else {
		claimed, err = e.store.ClaimSessionActiveTurn(opCtx, sessionID, turnID, "runner", claimToken)
	}
	if err != nil {
		return false, err
	}
	if !claimed {
		return false, nil
	}
	runCtx, cancel := context.WithCancel(e.backgroundContext())
	active := &runningTurn{turnID: turnID, claimToken: claimToken, cancel: cancel}
	runner.current = active
	claimedTurn := false
	rollbackPhase := "queued"
	releaseClaim := func(restoreQueued bool) error {
		var cleanupErrs []error
		if restoreQueued {
			if err := e.store.UpdateTurnStatusAndPhase(e.backgroundContext(), turnID, "queued", rollbackPhase); err != nil {
				cleanupErrs = append(cleanupErrs, fmt.Errorf("rollback turn status to queued: %w", err))
			}
			if claimedTurn {
				if err := e.store.ResetTurnClaim(e.backgroundContext(), turnID); err != nil {
					cleanupErrs = append(cleanupErrs, fmt.Errorf("reset turn claim after launch failure: %w", err))
				}
			}
			if err := e.store.SyncSessionQueueCount(e.backgroundContext(), sessionID); err != nil {
				cleanupErrs = append(cleanupErrs, fmt.Errorf("sync queue count after launch rollback: %w", err))
			}
			if err := e.store.TouchSessionState(e.backgroundContext(), sessionID, map[string]any{"status": "queued", "active_turn_id": nil}); err != nil {
				cleanupErrs = append(cleanupErrs, fmt.Errorf("touch session queued after launch rollback: %w", err))
			}
		}
		var releaseErr error
		if endedID != "" {
			releaseErr = e.store.ReleaseFailedEndedSteer(e.backgroundContext(), sessionID, claimToken, endedID)
		} else {
			releaseErr = e.store.ReleaseSessionActiveTurn(e.backgroundContext(), sessionID, claimToken)
		}
		if releaseErr != nil {
			cleanupErrs = append(cleanupErrs, fmt.Errorf("release active claim after launch failure: %w", releaseErr))
		}
		if runner.current == active {
			runner.current = nil
		}
		cancel()
		return errors.Join(cleanupErrs...)
	}
	// A queued-only cancellation may have committed before this claim.
	claimedRecord, getErr := e.store.GetTurn(opCtx, turnID)
	if getErr != nil || claimedRecord.Status != "queued" || (claimedRecord.Phase == "steer_returned" && !selectedSteer) {
		if cleanupErr := releaseClaim(false); cleanupErr != nil {
			return false, cleanupErr
		}
		if getErr != nil {
			return false, getErr
		}
		return false, store.ErrQueueConflict
	}
	if selectedSteer {
		rollbackPhase = claimedRecord.Phase
	}
	if err := e.store.MarkTurnClaimed(opCtx, turnID, "runner"); err != nil {
		if cleanupErr := releaseClaim(false); cleanupErr != nil {
			return false, errors.Join(err, cleanupErr)
		}
		return false, err
	}
	claimedTurn = true
	if err := e.store.UpdateTurnStatusAndPhase(opCtx, turnID, "running", "setup"); err != nil {
		if cleanupErr := releaseClaim(true); cleanupErr != nil {
			return false, errors.Join(err, cleanupErr)
		}
		return false, err
	}
	if hook := e.beforeLaunchSessionStateErrorHook; hook != nil {
		if err := hook(opCtx, sessionID, turnID); err != nil {
			if cleanupErr := releaseClaim(true); cleanupErr != nil {
				return false, errors.Join(err, cleanupErr)
			}
			return false, err
		}
	}
	sessionState := map[string]any{"active_turn_id": turnID, "status": "running"}
	if turnRec, turnErr := e.store.GetTurn(opCtx, turnID); turnErr == nil {
		if model := strings.TrimSpace(internalx.StringValue(turnRec.Metadata["model"], "")); model != "" {
			sessionState["model"] = model
		}
	}
	if err := e.store.TouchSessionState(opCtx, sessionID, sessionState); err != nil {
		if cleanupErr := releaseClaim(true); cleanupErr != nil {
			return false, errors.Join(err, cleanupErr)
		}
		return false, err
	}

	if selectedSteer {
		payload := map[string]any{"kind": "chat", "intent": "prompt", "turn_id": turnID}
		copySelectedMetadata(payload, claimedRecord.Metadata, ingressMetadataKeys)
		if media, ok := claimedRecord.Metadata["media"]; ok {
			payload["media"] = media
		}
		if err := e.store.PersistIdleQueuePrompt(opCtx, sessionID, turnID, claimToken, claimedRecord.Prompt, payload); err != nil {
			return false, errors.Join(err, releaseClaim(true))
		}
	}
	logutil.WarnIfErr("sync queue count after launch", e.store.SyncSessionQueueCount(opCtx, sessionID))
	e.runs.Add(1)
	go func() {
		defer e.runs.Done()
		// Submission and queued launch callers hold runner.mu until their
		// durable admission events are written. Do not emit turn.started (or
		// run tools) before turn.submitted has a sequence number.
		runner.mu.Lock()
		runner.mu.Unlock()
		runner.runTurn(e.store, sessionID, turnID, runCtx, cancel, active)
	}()
	return true, nil
}
