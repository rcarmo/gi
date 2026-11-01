package turn

import (
	"context"
	"time"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func (r *sessionRunner) claimToolBatchHook(ctx context.Context, s *store.Store, req HookRequest, key string) (bool, error) {
	batchID := store.ToolBatchID(ctx, req.TurnID)
	if batchID == "" {
		return true, nil
	}
	return s.ClaimToolBatchHook(ctx, req.SessionID, req.TurnID, batchID, key)
}

// Boundary-hook responses were never policy inputs. Persist the attempt first,
// but retain the existing treatment of handler errors and ignored responses.
func (r *sessionRunner) emitToolBatchBoundary(ctx context.Context, s *store.Store, req HookRequest, key string) error {
	run, err := r.claimToolBatchHook(ctx, s, req, key)
	if err != nil {
		return err
	}
	if run {
		_, _ = r.engine.emitHook(ctx, req)
	}
	return nil
}

func toolResultWireCall(ctx context.Context, turnID string, call goai.ToolCall) goai.ToolCall {
	if id, name := store.ToolBatchWireCall(ctx, turnID); id != "" {
		call.ID, call.Name = id, name
	}
	return call
}

func (e *Engine) recoverToolBatch(ctx context.Context, claim store.ActiveTurnClaim) (bool, error) {
	closed := claim.Status == "queued" && claim.Phase == store.ClosedToolBatchPhase
	if !closed && (claim.Status != "running" || claim.Phase != "waiting_on_tools") {
		return false, nil
	}
	if !e.providerRecoveryEnabled.Load() {
		candidate, err := e.store.HasToolBatchCheckpoint(ctx, claim.SessionID, claim.TurnID)
		return candidate, err
	}
	runner := e.runner(claim.SessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if runner.current != nil || runner.hosts[claim.TurnID] != nil {
		return true, store.ErrQueueConflict
	}
	e.launchMu.Lock()
	defer e.launchMu.Unlock()
	if e.closing.Load() {
		return true, context.Canceled
	}
	token, err := store.NewActiveTurnClaimToken()
	if err != nil {
		return true, err
	}
	state, err := e.store.AcquireToolBatchRecovery(ctx, claim, token, time.Now().Add(-interruptedTurnStaleAfter))
	if err != nil {
		return true, err
	}
	if state == nil {
		return false, nil
	}
	runCtx, cancel := context.WithCancel(e.backgroundContext())
	active := &runningTurn{turnID: claim.TurnID, claimToken: token, cancel: cancel, recoveredToolBatch: state}
	runner.current = active
	runner.registerHostFenceLocked(active, runCtx)
	e.runs.Add(1)
	go func() {
		defer e.runs.Done()
		defer runner.finishHostFence(active)
		runner.mu.Lock()
		runner.mu.Unlock()
		runner.runTurn(e.store, claim.SessionID, claim.TurnID, runCtx, cancel, active)
	}()
	return true, nil
}

// resolveInterruptedToolCall closes uncertainty without rerunning preparation.
// A replay plan is not a successful model result and retains unknown billing.
// Revalidation of an already committed plan may instead produce a zero-execution
// interrupted disposition; the old physical receipt remains unknown.
func (r *sessionRunner) resolveInterruptedToolCall(ctx context.Context, s *store.Store, sessionID, turnID, agentID string, iteration int, saved store.ToolBatchCall, conv *goai.Context) (bool, tools.RegisteredTool, error) {
	current, _ := r.engine.tools.GetRegistered(saved.Prepared.Name)
	replay := saved.OccurrenceID != "" && preparedToolReplaySafe(saved, current)
	if replay {
		turn, err := s.GetTurn(ctx, turnID)
		if err != nil {
			return false, tools.RegisteredTool{}, err
		}
		replay = toolAllowedByMetadata(turn.Metadata, saved.Prepared.Name) || r.engine.sessionLoadedTool(ctx, sessionID, saved.Prepared.Name)
	}
	if replay {
		if !saved.ReplayPending {
			err := s.SettleToolResult(ctx, turnID, sessionID, "tool.interrupted", map[string]any{"phase": "tool", "checkpoint": true, "tool": saved.Prepared.Name, "tool_call_id": saved.Prepared.ID, "occurrence_id": saved.OccurrenceID, "replay_pending": true}, "", "", map[string]any{"usage": saved.ReceivedUsage, "usage_unknown": !saved.CallbackReturned || saved.ReceivedUsageUnknown, "interrupted": true, "adopted": false, "context_contribution": false})
			if err != nil {
				return false, tools.RegisteredTool{}, err
			}
			r.engine.PublishRuntimeToolEvent("tool_interrupted", sessionID, turnID, agentID, saved.Prepared.Name, saved.Prepared.ID, iteration, nil, map[string]any{"occurrence_id": saved.OccurrenceID, "interrupted": true, "outcome_unknown": true, "replay_pending": true})
		}
		return true, current, nil
	}
	text := "Tool outcome unknown after interruption. It may already have produced effects. Verify the outcome before repeating effects; the previous execution and preparation hooks were not replayed."
	if saved.OccurrenceID == "" {
		text = "Tool preparation was interrupted before executor admission. Preparation hooks were not replayed; verify any preparation-hook effects before retrying."
	}
	if saved.CallbackReturned {
		text = "Execution returned without a committed tool result. Its executor was not replayed. Verify effects before retrying."
	}
	if saved.ResultHookStarted {
		text = "The execution callback returned, but its result policy was interrupted. Verify effects before retrying; neither the executor nor its result policy was replayed."
	}
	diagnostic := map[string]any{"severity": "error", "code": "interrupted", "message": text}
	if saved.CommittedOutput != "" {
		text = saved.CommittedOutput + "\n\n<harness>\n[error] " + text + "\n</harness>"
	}
	occurrence := saved.OccurrenceID
	physical := occurrence != "" && !saved.ReplayPending
	event := map[string]any{"phase": "tool", "checkpoint": true, "tool": saved.Original.Name, "tool_call_id": saved.Original.ID, "replay_pending": false}
	if !physical {
		occurrence = store.NowID("tool")
		event["unadmitted"], event["previous_occurrence_id"] = true, saved.OccurrenceID
	}
	event["occurrence_id"] = occurrence
	result := map[string]any{"kind": "tool_result", "turn_id": turnID, "tool_call_id": saved.Original.ID, "tool_name": saved.Original.Name, "is_error": true, "interrupted": true, "outcome_unknown": saved.OccurrenceID != "", "usage_unknown": physical}
	result["occurrence_id"] = occurrence
	result["diagnostics"] = []map[string]any{diagnostic}
	if saved.OutputTruncated {
		result["output_truncated"] = true
	}
	if physical {
		event["tool"], event["tool_call_id"] = saved.Prepared.Name, saved.Prepared.ID
		result["usage"] = saved.ReceivedUsage
		result["usage_unknown"] = !saved.CallbackReturned || saved.ReceivedUsageUnknown
	}
	if err := s.SettleToolResult(ctx, turnID, sessionID, "tool.interrupted", event, store.NowID("msg"), text, result); err != nil {
		return false, tools.RegisteredTool{}, err
	}
	goai.AppendToolResult(conv, saved.Original.ID, saved.Original.Name, text, true)
	r.engine.PublishRuntimeToolEvent("tool_interrupted", sessionID, turnID, agentID, event["tool"].(string), event["tool_call_id"].(string), iteration, nil, map[string]any{"occurrence_id": occurrence, "interrupted": true, "outcome_unknown": saved.OccurrenceID != "", "replay_pending": false})
	r.engine.broadcast(sessionID, map[string]any{"type": "tool_failed", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": saved.Original.Name, "tool_call_id": saved.Original.ID, "occurrence_id": occurrence, "error": text, "interrupted": true})
	return false, tools.RegisteredTool{}, nil
}

// As in pi-durable's execute phase, replay the committed arguments directly:
// neither preparation nor validation against a replacement schema runs again.
// Gi additionally keeps its already-attempted result hooks once-only.
func preparedToolReplaySafe(saved store.ToolBatchCall, current tools.RegisteredTool) bool {
	return saved.ReplaySafe && current.ReplaySafe && !saved.ResultHookStarted && saved.Prepared.ArgumentsError == "" && current.Name == saved.Prepared.Name && current.Executor != nil
}
