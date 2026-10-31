package turn

import (
	"context"
	"fmt"
	"strings"

	goai "github.com/rcarmo/go-ai"

	"github.com/rcarmo/gi/internal/compaction"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
)

// CompactionPolicy is the immutable policy loaded when this engine started.
func (e *Engine) CompactionPolicy() config.CompactionSettings {
	policy := e.runtimeCfg.Compaction
	policy.Enabled = e.autoCompaction()
	return policy
}

// SetAutoCompaction is Pi's setAutoCompactionEnabled: it applies to the
// running engine at once.
func (e *Engine) SetAutoCompaction(enabled bool) {
	v := int32(2)
	if enabled {
		v = 1
	}
	e.autoCompactionOverride.Store(v)
}

func (e *Engine) autoCompaction() bool {
	switch e.autoCompactionOverride.Load() {
	case 1:
		return true
	case 2:
		return false
	}
	return e.runtimeCfg.Compaction.Enabled
}

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
	if prep.MessagesToSummarize == 0 {
		// Everything fits in the kept recent window: nothing to compact yet.
		return state, nil
	}
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
	return e.SubmitManualCompactionWithInstructions(ctx, sessionID, expected, "")
}

// MaxCompactionInstructions bounds /compact custom instructions.
const MaxCompactionInstructions = 4000

// SubmitManualCompactionWithInstructions is Pi's /compact [instructions]:
// the instructions reach the before-compact hook and the stored summary.
func (e *Engine) SubmitManualCompactionWithInstructions(ctx context.Context, sessionID, expected, instructions string) (*SubmitResult, error) {
	instructions = strings.TrimSpace(instructions)
	if len(instructions) > MaxCompactionInstructions {
		return nil, fmt.Errorf("compaction instructions exceed %d bytes", MaxCompactionInstructions)
	}
	r := e.runner(sessionID)
	r.mu.Lock()
	defer r.mu.Unlock()
	e.launchMu.Lock()
	defer e.launchMu.Unlock()
	if e.closing.Load() {
		return nil, fmt.Errorf("engine closing")
	}
	state, err := e.ManualCompactionState(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	// A stale token means the history moved on, whatever the new state allows.
	if expected == "" || state["token"] != expected {
		return nil, store.ErrContextChanged
	}
	if state["available"] != true {
		return nil, fmt.Errorf("%w: %s", store.ErrQueueConflict, state["reason"])
	}
	session, err := e.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	model := inference.SessionModel(session.State, inference.SessionModelChoice{Model: e.runtimeCfg.DefaultModel}).Model
	id := store.NowID("turn")
	claimToken, err := store.NewActiveTurnClaimToken()
	if err != nil {
		return nil, err
	}
	if err = e.store.AdmitManualCompactionWithInstructions(ctx, sessionID, id, claimToken, expected, model, instructions); err != nil {
		return nil, err
	}
	// Admission already committed the claim and submitted checkpoint together.
	runCtx, cancel := context.WithCancel(e.backgroundContext())
	active := &runningTurn{turnID: id, claimToken: claimToken, cancel: cancel}
	r.current = active
	e.runs.Add(1)
	go func() {
		defer e.runs.Done()
		r.mu.Lock()
		r.mu.Unlock()
		r.runTurn(e.store, sessionID, id, runCtx, cancel, active)
	}()
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
			instructions, _ := run.turn.Metadata["custom_instructions"].(string)
			err = r.compactSnapshotWithInstructions(ctx, run.sessionID, run.turnID, run.model, run.agentID, conv, snapshot, true, instructions)
		}
	}
	if err != nil {
		// Piclaw reports the outcome in the timeline ("Compaction cancelled").
		status, kind, notice := "failed", "compaction_error", fmt.Sprintf("Compaction failed: %v", err)
		if ctx.Err() != nil || isCancellationError(err) {
			status, kind, notice = "cancelled", "", "Compaction cancelled."
		}
		r.finishTurn(r.store, run.turnID, run.sessionID, run.agentID, run.model, status, notice, kind)
		return
	}
	r.finishTurnOK(r.store, run.turnID, run.sessionID, run.agentID, run.model, 0)
}
