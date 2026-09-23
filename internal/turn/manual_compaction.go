package turn

import (
	"context"
	"fmt"
	goai "github.com/rcarmo/go-ai"

	"github.com/rcarmo/gi/internal/compaction"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
)

// ManualCompactionState is advisory. Admission checks busy state and the exact
// history token again under the runner lock and in the store transaction.
func (e *Engine) ManualCompactionState(ctx context.Context, sessionID string) (map[string]any, error) {
	snapshot, err := e.store.ContextSnapshot(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	state := map[string]any{
		"available": false, "reason": "Not enough eligible context", "token": store.ContextToken(snapshot),
		// Expose the engine's effective policy, not a second config load or UI defaults.
		"policy": e.runtimeCfg.Compaction, "policy_scope": "startup",
	}
	busy, err := e.store.CompactionBusy(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if busy {
		state["reason"] = "Session has active or queued work"
		return state, nil
	}
	r := e.runner(sessionID)
	messages := r.projectContextSnapshot(ctx, sessionID, snapshot)
	if len(messages) < 2 {
		return state, nil
	}
	settings := e.runtimeCfg.Compaction
	prep := compaction.Prepare(messages, compaction.EstimateMessagesTokens(messages), settings.KeepRecentTokens, settings.ReserveTokens, settings.ThresholdTokens, settings.Strategy)
	// Reuse the exact projection/media guard used by automatic completion.
	boundary, err := r.compactionBoundary(ctx, sessionID, &goai.Context{Messages: messages}, snapshot, map[string]any{"messages_to_summarize": prep.MessagesToSummarize})
	if err != nil {
		return nil, err
	}
	if boundary == nil {
		state["reason"] = "Older media context cannot be compacted safely"
		return state, nil
	}
	state["available"] = true
	state["reason"] = ""
	return state, nil
}

func (e *Engine) SubmitManualCompaction(ctx context.Context, sessionID, expected string) (*SubmitResult, error) {
	r := e.runner(sessionID)
	r.mu.Lock()
	defer r.mu.Unlock()
	state, err := e.ManualCompactionState(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if state["available"] != true {
		return nil, fmt.Errorf("%w: %s", store.ErrQueueConflict, state["reason"])
	}
	if expected == "" || state["token"] != expected {
		return nil, store.ErrContextChanged
	}
	session, err := e.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	model := inference.SessionModel(session.State, inference.SessionModelChoice{Model: e.runtimeCfg.DefaultModel}).Model
	id := store.NowID("turn")
	if err = e.store.AdmitManualCompaction(ctx, sessionID, id, expected, model); err != nil {
		return nil, err
	}
	// Admission already committed the claim and submitted checkpoint together.
	runCtx, cancel := context.WithCancel(e.backgroundContext())
	active := &runningTurn{turnID: id, cancel: cancel}
	r.current = active
	go func() { r.mu.Lock(); r.mu.Unlock(); r.runTurn(e.store, sessionID, id, runCtx, cancel, active) }()
	e.PublishRuntimeTurnEvent("turn_submitted", sessionID, id, "", "running", "setup", map[string]any{"operation": "manual_compaction"})
	return &SubmitResult{TurnID: id, SessionID: sessionID, Status: "running"}, nil
}

func (r *sessionRunner) runManualCompaction(ctx context.Context, run *preparedTurnRun) {
	snapshot, err := r.store.ContextSnapshot(ctx, run.sessionID)
	if err == nil && store.ContextToken(snapshot) != run.turn.Metadata["context_token"] {
		err = store.ErrContextChanged
	}
	if err == nil {
		conv := &goai.Context{Messages: r.projectContextSnapshot(ctx, run.sessionID, snapshot)}
		if len(conv.Messages) < 2 {
			err = store.ErrContextChanged
		} else {
			err = r.compactSnapshot(ctx, run.sessionID, run.turnID, run.model, run.agentID, conv, snapshot, true)
		}
	}
	if err != nil {
		status, kind := "failed", "compaction_error"
		if ctx.Err() != nil || isCancellationError(err) {
			status = "cancelled"
			kind = ""
		}
		r.finishTurn(r.store, run.turnID, run.sessionID, run.agentID, run.model, status, fmt.Sprintf("Manual compaction: %v", err), kind)
		return
	}
	r.finishTurnOK(r.store, run.turnID, run.sessionID, run.agentID, run.model, 0)
}
