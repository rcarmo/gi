package turn

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

const repeatedToolFailureLimit = 4

func toolFailureSignature(call goai.ToolCall, err error) string {
	if err == nil {
		return ""
	}
	argKeys := make([]string, 0, len(call.Arguments))
	for k := range call.Arguments {
		argKeys = append(argKeys, k)
	}
	sort.Strings(argKeys)
	parts := make([]string, 0, len(argKeys))
	for _, k := range argKeys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, call.Arguments[k]))
	}
	return fmt.Sprintf("%s|%s|%s", call.Name, strings.Join(parts, ","), err.Error())
}

func nextRepeatedToolFailureCount(lastSig string, lastCount int, call goai.ToolCall, err error) (string, int) {
	sig := toolFailureSignature(call, err)
	if sig == "" {
		return "", 0
	}
	if sig == lastSig {
		return sig, lastCount + 1
	}
	return sig, 1
}

// runAgentLoop runs the core tool-use loop: call LLM, execute any tool calls,
// feed results back, repeat until the LLM produces a final text response or
// the iteration budget is exhausted.
func (r *sessionRunner) runAgentLoop(ctx context.Context, s *store.Store, turnID, sessionID, model, agentID string, initialSteering []store.SteeringMessage) {
	maxIter := r.engine.runtimeCfg.MaxIterations
	if maxIter <= 0 {
		maxIter = 64
	}

	convCtx := r.assembleAgentContext(ctx, s, turnID, sessionID, model, agentID)
	agentEndReason := "completed"
	defer func() {
		_, _ = r.engine.emitHook(context.Background(), HookRequest{Name: HookAgentEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Payload: map[string]any{"reason": agentEndReason}})
	}()

	r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "title": "Thinking…", "status": "running", "turn_id": turnID})

	var totalUsage goai.Usage
	lastToolFailureSig := ""
	repeatedToolFailureCount := 0
	pendingSteering := append([]store.SteeringMessage(nil), initialSteering...)

	for iter := 1; iter <= maxIter; iter++ {
		if ctx.Err() != nil {
			agentEndReason = "cancelled"
			r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled", "")
			return
		}

		pendingSteering = r.prepareAgentIteration(ctx, sessionID, turnID, model, agentID, iter, convCtx, pendingSteering)
		result, inferErr := r.runProviderIteration(ctx, s, turnID, sessionID, model, agentID, iter, maxIter, convCtx)
		iterLabel := fmt.Sprintf("iter=%d/%d", iter, maxIter)
		if inferErr != nil {
			if ctx.Err() != nil || isCancellationError(inferErr) {
				agentEndReason = "cancelled"
				r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled", "")
				return
			}
			var abortErr hookAbortError
			if errors.As(inferErr, &abortErr) {
				warnStore("append inference.aborted event", s.AppendTurnEvent(ctx, turnID, sessionID, "inference.aborted", map[string]any{"phase": "inference", "checkpoint": true, "error": abortErr.Error(), "iteration": iter, "hard_abort": abortErr.hard}))
				agentEndReason = "aborted"
				r.finishTurn(s, turnID, sessionID, agentID, model, "aborted", abortErr.Error(), "hook_abort")
				return
			}
			log.Printf("inference [%s] error: %v", iterLabel, inferErr)
			warnStore("append inference.failed event", s.AppendTurnEvent(ctx, turnID, sessionID, "inference.failed", map[string]any{"phase": "inference", "checkpoint": true, "error": inferErr.Error(), "iteration": iter}))
			agentEndReason = "failed"
			r.finishTurn(s, turnID, sessionID, agentID, model, "failed", fmt.Sprintf("Inference error: %v", inferErr), "provider_error")
			return
		}
		if result == nil || result.Message == nil {
			if ctx.Err() != nil {
				agentEndReason = "cancelled"
				r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled", "")
				return
			}
			log.Printf("inference [%s]: nil result", iterLabel)
			agentEndReason = "failed"
			r.finishTurn(s, turnID, sessionID, agentID, model, "failed", "Inference returned no result", "provider_invalid_result")
			return
		}

		if result.Usage != nil {
			totalUsage.Input += result.Usage.Input
			totalUsage.Output += result.Usage.Output
			totalUsage.TotalTokens += result.Usage.TotalTokens
			totalUsage.CacheRead += result.Usage.CacheRead
			totalUsage.CacheWrite += result.Usage.CacheWrite
			totalUsage.Cost.Input += result.Usage.Cost.Input
			totalUsage.Cost.Output += result.Usage.Cost.Output
			totalUsage.Cost.Total += result.Usage.Cost.Total
		}

		assistantMsg := result.Message
		textContent := goai.GetTextContent(assistantMsg)
		toolCalls := goai.GetToolCalls(assistantMsg)
		log.Printf("inference [%s]: stop=%q toolCalls=%d text=%d", iterLabel, assistantMsg.StopReason, len(toolCalls), len(textContent))
		goai.AppendAssistantMessage(convCtx, assistantMsg)

		needsToolExecution := goai.NeedsToolExecution(assistantMsg) || len(toolCalls) > 0
		if !needsToolExecution {
			if steerMsgs, err := r.dequeueSteeringMessages(ctx, sessionID); err != nil {
				log.Printf("steering dequeue error after direct response: %v", err)
			} else if len(steerMsgs) > 0 {
				pendingSteering = append(pendingSteering, steerMsgs...)
				continue
			}
			log.Printf("inference [%s]: final response (%d chars, %d iterations)", iterLabel, len(textContent), iter)
			agentEndReason = "completed"
			r.persistUsage(s, turnID, sessionID, &totalUsage, iter)

			msgID := store.NowID("msg")
			warnStore("add assistant inference message", s.AddMessage(ctx, msgID, sessionID, "assistant", textContent, map[string]any{
				"kind": "chat", "source": "inference", "model": model,
				"turn_id": turnID, "agent_id": agentID, "iterations": iter,
			}))

			r.broadcastPost(sessionID, turnID, msgID, textContent, agentID)
			_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookMessageEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"chars": len(textContent)}})
			_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookTurnEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"status": "completed"}})
			r.finishTurnOK(s, turnID, sessionID, agentID, model, iter)
			return
		}

		log.Printf("inference [%s]: %d tool call(s)", iterLabel, len(toolCalls))
		toolCallSummary := textContent
		for _, tc := range toolCalls {
			if toolCallSummary != "" {
				toolCallSummary += "\n"
			}
			toolCallSummary += fmt.Sprintf("[tool_call: %s]", tc.Name)
		}
		warnStore("add assistant tool_calls summary", s.AddMessage(ctx, store.NowID("msg"), sessionID, "assistant", toolCallSummary, map[string]any{
			"kind": "tool_calls", "source": "inference", "model": model,
			"turn_id": turnID, "agent_id": agentID,
		}))

		outcome := r.executeToolCallsPhase(ctx, s, turnID, sessionID, model, agentID, iter, convCtx, toolCalls, pendingSteering, lastToolFailureSig, repeatedToolFailureCount, &totalUsage)
		if outcome.terminated {
			return
		}
		pendingSteering = outcome.pendingSteering
		lastToolFailureSig = outcome.lastToolFailureSig
		repeatedToolFailureCount = outcome.repeatedToolFailureCount
		if outcome.skipRemainingTools || len(pendingSteering) > 0 {
			continue
		}
		_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookTurnEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"status": "tools"}})
	}

	log.Printf("inference: max iterations (%d) reached for turn %s", maxIter, turnID)
	agentEndReason = "completed"
	r.persistUsage(s, turnID, sessionID, &totalUsage, maxIter)
	r.finishTurn(s, turnID, sessionID, agentID, model, "completed", fmt.Sprintf("Reached maximum iteration limit (%d). The task may be incomplete.", maxIter), "")
}

// executeTool dispatches a single tool call and returns the text result.
func (r *sessionRunner) executeTool(ctx context.Context, call goai.ToolCall, sessionID, turnID string) (string, error) {
	if strings.TrimSpace(turnID) != "" {
		turnRec, err := r.store.GetTurn(ctx, turnID)
		if err != nil {
			return "", err
		}
		if !toolAllowedByMetadata(turnRec.Metadata, call.Name) {
			return "", fmt.Errorf("tool not allowed in this turn: %s", call.Name)
		}
	}
	tool, ok := r.engine.tools.GetRegistered(call.Name)
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", call.Name)
	}
	return tool.Executor(ctx, ToolRuntime{
		Engine:        r.engine,
		Runner:        r,
		Store:         r.store,
		SessionID:     sessionID,
		TurnID:        turnID,
		WorkspaceRoot: r.engine.runtimeCfg.WorkspaceRoot,
	}, call)
}

// persistUsage records cumulative usage for the turn.
func (r *sessionRunner) persistUsage(s *store.Store, turnID, sessionID string, usage *goai.Usage, iterations int) {
	usageMap := map[string]any{
		"input": usage.Input, "output": usage.Output,
		"total":      usage.TotalTokens,
		"cache_read": usage.CacheRead, "cache_write": usage.CacheWrite,
		"cost_input": usage.Cost.Input, "cost_output": usage.Cost.Output,
		"cost_total": usage.Cost.Total,
		"iterations": iterations,
	}
	warnStore("append inference.finished event", s.AppendTurnEvent(context.Background(), turnID, sessionID, "inference.finished", map[string]any{
		"phase": "inference", "checkpoint": true, "usage": usageMap, "iterations": iterations,
	}))
	log.Printf("inference: usage input=%d output=%d total=%d cost=%.6f iterations=%d",
		usage.Input, usage.Output, usage.TotalTokens, usage.Cost.Total, iterations)
}

// broadcastPost sends a new_post SSE event for the final assistant message.
func (r *sessionRunner) broadcastPost(sessionID, turnID, msgID, content, agentID string) {
	r.engine.broadcast(sessionID, map[string]any{
		"type": "new_post", "id": msgID, "chat_jid": "gi:" + sessionID,
		"content": content, "timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"sender": "agent", "is_bot_message": true,
		"data": map[string]any{"type": "agent_response", "content": content, "agent_id": agentID},
	})
	r.engine.broadcast(sessionID, map[string]any{"type": "agent_response", "chat_jid": "gi:" + sessionID, "id": msgID})
}

// finishTurnOK marks a turn as successfully completed.
func (r *sessionRunner) finishTurnOK(s *store.Store, turnID, sessionID, agentID, model string, iterations int) {
	r.appendFinalSteeringCheckpoint(s, turnID, sessionID)
	warnStore("append turn.finished event", s.AppendTurnEvent(context.Background(), turnID, sessionID, "turn.finished", map[string]any{
		"phase": "turn", "checkpoint": true, "status": "completed", "iterations": iterations,
	}))
	warnStore("update turn status and phase completed", s.UpdateTurnStatusAndPhase(context.Background(), turnID, "completed", "completed"))
	warnStore("mark turn finished", s.MarkTurnFinished(context.Background(), turnID))
	r.engine.PublishRuntimeTurnEvent("turn_completed", sessionID, turnID, agentID, "completed", "completed", map[string]any{"reason": "completed", "iterations": iterations})
	r.emitTurnStateHookOnly(context.Background(), sessionID, turnID, agentID, model, "completed", "completed", map[string]any{"reason": "completed", "iterations": iterations})
	r.propagateChildSubTurnCancellation(context.Background(), turnID, "completed", "")
	r.publishSubTurnLifecycle(context.Background(), turnID, "completed")
	warnStore("touch session idle", s.TouchSessionState(context.Background(), sessionID, map[string]any{"status": "idle", "active_turn_id": nil}))
	r.engine.PublishRuntimeSessionEvent("session_idle", sessionID, agentID, "idle", map[string]any{"reason": "turn_completed", "turn_id": turnID, "model": model})
	r.emitSessionStateHookOnly(context.Background(), sessionID, agentID, model, "idle", map[string]any{"reason": "turn_completed"})
	r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "title": "", "status": "idle"})
}

// finishTurn persists a terminal status and optional system message.
func (r *sessionRunner) finishTurn(s *store.Store, turnID, sessionID, agentID, model, status, systemMsg, failureKind string) {
	r.appendFinalSteeringCheckpoint(s, turnID, sessionID)
	if systemMsg != "" {
		msgID := store.NowID("msg")
		warnStore("add terminal system message", s.AddMessage(context.Background(), msgID, sessionID, "assistant", systemMsg, map[string]any{
			"kind": "chat", "source": "system", "turn_id": turnID, "agent_id": agentID,
		}))
		if status == "completed" || status == "failed" {
			r.broadcastPost(sessionID, turnID, msgID, systemMsg, agentID)
		}
	}
	if failureKind != "" {
		markTurnFailure(s, turnID, sessionID, failureKind, systemMsg)
	}
	warnStore("append turn.finished event", s.AppendTurnEvent(context.Background(), turnID, sessionID, "turn.finished", map[string]any{
		"phase": "turn", "checkpoint": true, "status": status,
	}))
	phase := terminalPhaseForStatus(status)
	warnStore("update turn status and phase terminal", s.UpdateTurnStatusAndPhase(context.Background(), turnID, status, phase))
	warnStore("mark turn finished", s.MarkTurnFinished(context.Background(), turnID))
	turnEventType := "turn_terminal"
	sessionIdleReason := "turn_terminal"
	if status == "completed" && failureKind == "" {
		turnEventType = "turn_completed"
		sessionIdleReason = "turn_completed"
	}
	r.engine.PublishRuntimeTurnEvent(turnEventType, sessionID, turnID, agentID, status, phase, map[string]any{"reason": firstNonEmpty(failureKind, status), "failure_kind": failureKind})
	r.emitTurnStateHookOnly(context.Background(), sessionID, turnID, agentID, model, status, phase, map[string]any{"reason": firstNonEmpty(failureKind, status)})
	r.propagateChildSubTurnCancellation(context.Background(), turnID, status, failureKind)
	r.publishSubTurnLifecycle(context.Background(), turnID, status)
	warnStore("touch session idle", s.TouchSessionState(context.Background(), sessionID, map[string]any{"status": "idle", "active_turn_id": nil}))
	r.engine.PublishRuntimeSessionEvent("session_idle", sessionID, agentID, "idle", map[string]any{"reason": sessionIdleReason, "turn_id": turnID, "turn_status": status, "model": model})
	r.emitSessionStateHookOnly(context.Background(), sessionID, agentID, model, "idle", map[string]any{"reason": sessionIdleReason, "turn_status": status})
	r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "title": "", "status": "idle"})
}

func (r *sessionRunner) publishSubTurnLifecycle(ctx context.Context, childTurnID, status string) {
	if strings.TrimSpace(childTurnID) == "" {
		return
	}
	sub, err := r.store.GetSubTurnByChild(ctx, childTurnID)
	if err != nil {
		return
	}
	payload := map[string]any{
		"type":           "subturn_status",
		"chat_jid":       "gi:" + sub.ParentSessionID,
		"parent_turn_id": sub.ParentTurnID,
		"parent_session": sub.ParentSessionID,
		"child_turn_id":  sub.ChildTurnID,
		"child_session":  sub.ChildSessionID,
		"status":         status,
		"depth":          sub.Depth,
		"delivery_mode":  sub.DeliveryMode,
	}
	r.engine.broadcast(sub.ParentSessionID, payload)
	if sub.ChildSessionID != sub.ParentSessionID {
		mirror := cloneMap(payload)
		mirror["chat_jid"] = "gi:" + sub.ChildSessionID
		r.engine.broadcast(sub.ChildSessionID, mirror)
	}
	if !isTerminalSubTurnStatus(status) {
		return
	}
	summary := r.subTurnResultSummary(ctx, sub.ChildSessionID, sub.ChildTurnID)
	if sub.DeliveryMode == "async" {
		orphaned := false
		if parentTurn, err := r.store.GetTurn(ctx, sub.ParentTurnID); err == nil {
			orphaned = isTerminalSubTurnStatus(parentTurn.Status)
		}
		eventType := "subturn_result_ready"
		if orphaned {
			eventType = "subturn_orphaned"
			warnStore("update async subturn orphan metadata", r.store.UpdateSubTurnMetadataByChild(ctx, sub.ChildTurnID, map[string]any{
				"orphaned":      true,
				"orphaned_at":   time.Now().UTC().Format(time.RFC3339Nano),
				"orphan_reason": "parent_turn_completed_before_async_result_consumption",
			}))
			if sub.ParentSessionID != sub.ChildSessionID {
				content := fmt.Sprintf("Async sub-turn %s finished with status %s after parent turn %s had already ended.", sub.ChildTurnID, status, sub.ParentTurnID)
				if strings.TrimSpace(summary) != "" {
					content += "\n\n" + summary
				}
				warnStore("add async orphan result message", r.store.AddMessage(ctx, store.NowID("msg"), sub.ParentSessionID, "system", content, map[string]any{
					"kind":             "subturn_orphan_result",
					"parent_turn_id":   sub.ParentTurnID,
					"child_turn_id":    sub.ChildTurnID,
					"child_session_id": sub.ChildSessionID,
					"status":           status,
					"delivery_mode":    "async",
					"orphaned":         true,
					"summary":          summary,
				}))
			}
		}
		r.engine.broadcast(sub.ParentSessionID, map[string]any{
			"type":           eventType,
			"chat_jid":       "gi:" + sub.ParentSessionID,
			"parent_turn_id": sub.ParentTurnID,
			"parent_session": sub.ParentSessionID,
			"child_turn_id":  sub.ChildTurnID,
			"child_session":  sub.ChildSessionID,
			"status":         status,
			"summary":        summary,
			"delivery_mode":  "async",
			"orphaned":       orphaned,
		})
		return
	}
	if sub.ParentSessionID != sub.ChildSessionID {
		content := fmt.Sprintf("Sub-turn %s finished with status %s.", sub.ChildTurnID, status)
		if strings.TrimSpace(summary) != "" {
			content += "\n\n" + summary
		}
		warnStore("add sync subturn result message", r.store.AddMessage(ctx, store.NowID("msg"), sub.ParentSessionID, "system", content, map[string]any{
			"kind":             "subturn_result",
			"parent_turn_id":   sub.ParentTurnID,
			"child_turn_id":    sub.ChildTurnID,
			"child_session_id": sub.ChildSessionID,
			"status":           status,
			"delivery_mode":    "sync",
			"summary":          summary,
		}))
	}
	r.engine.broadcast(sub.ParentSessionID, map[string]any{
		"type":           "subturn_result_delivered",
		"chat_jid":       "gi:" + sub.ParentSessionID,
		"parent_turn_id": sub.ParentTurnID,
		"parent_session": sub.ParentSessionID,
		"child_turn_id":  sub.ChildTurnID,
		"child_session":  sub.ChildSessionID,
		"status":         status,
		"summary":        summary,
		"delivery_mode":  "sync",
	})
}

func isTerminalSubTurnStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "completed", "failed", "aborted", "cancelled":
		return true
	default:
		return false
	}
}

func (r *sessionRunner) subTurnResultSummary(ctx context.Context, childSessionID, childTurnID string) string {
	if strings.TrimSpace(childSessionID) == "" {
		return ""
	}
	msgs, err := r.store.ListMessages(ctx, childSessionID)
	if err != nil {
		return ""
	}
	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		if msg.Role != "assistant" {
			continue
		}
		msgTurnID := stringValue(msg.Payload["turn_id"], "")
		if msgTurnID != "" && msgTurnID != childTurnID {
			continue
		}
		text := strings.TrimSpace(msg.Content)
		if text == "" {
			continue
		}
		if len(text) > 500 {
			return text[:500] + "..."
		}
		return text
	}
	return ""
}
