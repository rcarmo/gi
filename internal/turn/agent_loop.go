package turn

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/inference"
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

	sysPrompt := r.engine.systemPrompt
	if sysPrompt == "" {
		sysPrompt = "You are a helpful coding assistant."
	}

	// Build go-ai conversation context from stored messages.
	// We only reconstruct text user/assistant history here;
	// the tool-use loop keeps its own messages in memory.
	msgs, _ := s.ListMessages(ctx, sessionID)
	convCtx := &goai.Context{
		SystemPrompt: sysPrompt,
		Tools:        r.engine.toolDefs(),
	}
	for _, m := range msgs {
		switch m.Role {
		case "user":
			convCtx.Messages = append(convCtx.Messages, goai.UserMessage(m.Content))
		case "assistant":
			convCtx.Messages = append(convCtx.Messages, goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: m.Content}}})
			// Deliberately skip tool_result — those are part of the current turn's loop,
			// not persistent history that should be replayed.
		}
	}

	if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookBeforeAgentStart, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, SystemPrompt: convCtx.SystemPrompt, Messages: convCtx.Messages, Tools: convCtx.Tools}); err != nil {
		log.Printf("hook before_agent_start error: %v", err)
	} else {
		if resp.SystemPrompt != "" {
			convCtx.SystemPrompt = resp.SystemPrompt
		}
		if resp.Messages != nil {
			convCtx.Messages = resp.Messages
		}
		if resp.Tools != nil {
			convCtx.Tools = resp.Tools
		}
		if strings.TrimSpace(resp.Message) != "" {
			convCtx.Messages = append([]goai.Message{goai.UserMessage(resp.Message)}, convCtx.Messages...)
		}
	}
	_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookAgentStart, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model})
	agentEndReason := "completed"
	defer func() {
		_, _ = r.engine.emitHook(context.Background(), HookRequest{Name: HookAgentEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Payload: map[string]any{"reason": agentEndReason}})
	}()

	r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "title": "Thinking…", "status": "running", "turn_id": turnID})

	var totalUsage goai.Usage
	lastToolFailureSig := ""
	repeatedToolFailureCount := 0
	pendingSteering := append([]store.SteeringMessage(nil), initialSteering...)

	for iter := 0; iter < maxIter; iter++ {
		if ctx.Err() != nil {
			r.finishTurn(s, turnID, sessionID, agentID, "cancelled", "Turn cancelled", "")
			return
		}

		iterLabel := fmt.Sprintf("iter=%d/%d", iter+1, maxIter)
		if steerMsgs, err := r.dequeueSteeringMessages(ctx, sessionID); err != nil {
			log.Printf("steering dequeue error: %v", err)
		} else if len(steerMsgs) > 0 {
			pendingSteering = append(pendingSteering, steerMsgs...)
		}
		if len(pendingSteering) > 0 {
			r.injectSteeringMessages(ctx, sessionID, turnID, convCtx, pendingSteering)
			pendingSteering = nil
		}
		_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookTurnStart, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter + 1})
		if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookContext, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter + 1, SystemPrompt: convCtx.SystemPrompt, Messages: convCtx.Messages, Tools: convCtx.Tools}); err != nil {
			log.Printf("hook context error: %v", err)
		} else {
			if resp.SystemPrompt != "" {
				convCtx.SystemPrompt = resp.SystemPrompt
			}
			if resp.Messages != nil {
				convCtx.Messages = resp.Messages
			}
			if resp.Tools != nil {
				convCtx.Tools = resp.Tools
			}
		}
		r.maybeCompactContext(ctx, sessionID, turnID, agentID, model, convCtx)
		_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookBeforeProviderRequest, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter + 1, Payload: map[string]any{"model": model, "messages": len(convCtx.Messages), "tools": len(convCtx.Tools)}})
		_ = s.UpdateTurnStatusAndPhase(ctx, turnID, "running", "running")
		_ = s.AppendTurnEvent(ctx, turnID, sessionID, "inference.started", map[string]any{"phase": "inference", "model": model, "iteration": iter + 1, "checkpoint": true})
		log.Printf("inference [%s]: calling %s", iterLabel, model)

		r.engine.broadcast(sessionID, map[string]any{
			"type": "agent_status", "chat_jid": "gi:" + sessionID,
			"title": fmt.Sprintf("Thinking… (%d)", iter+1), "status": "running", "turn_id": turnID,
		})

		result, inferErr := inference.StreamWithTools(ctx, model, convCtx, func(ev map[string]any) {
			ev["chat_jid"] = "gi:" + sessionID
			ev["turn_id"] = turnID
			ev["iteration"] = iter + 1
			switch ev["type"] {
			case "text_delta":
				_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookMessageUpdate, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter + 1, Payload: map[string]any{"delta": ev["delta"]}})
				ev["type"] = "agent_draft_delta"
				r.engine.broadcast(sessionID, ev)
			case "thinking_delta":
				ev["type"] = "agent_thought_delta"
				r.engine.broadcast(sessionID, ev)
			case "tool_call_start":
				r.engine.broadcast(sessionID, map[string]any{
					"type": "agent_status", "chat_jid": "gi:" + sessionID,
					"title": fmt.Sprintf("Tool: %s", ev["name"]), "status": "running", "turn_id": turnID,
				})
			case "error":
				r.engine.broadcast(sessionID, ev)
			}
		})

		_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookAfterProviderResponse, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter + 1, Payload: map[string]any{"ok": inferErr == nil}})
		if inferErr != nil {
			log.Printf("inference [%s] error: %v", iterLabel, inferErr)
			_ = s.AppendTurnEvent(ctx, turnID, sessionID, "inference.failed", map[string]any{"phase": "inference", "checkpoint": true, "error": inferErr.Error(), "iteration": iter + 1})
			r.finishTurn(s, turnID, sessionID, agentID, "failed", fmt.Sprintf("Inference error: %v", inferErr), "provider_error")
			return
		}
		if result == nil || result.Message == nil {
			log.Printf("inference [%s]: nil result", iterLabel)
			r.finishTurn(s, turnID, sessionID, agentID, "failed", "Inference returned no result", "provider_invalid_result")
			return
		}

		// Accumulate usage
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

		// Append the full assistant message to context for the next iteration
		goai.AppendAssistantMessage(convCtx, assistantMsg)

		// Some providers (notably Codex) may return stop="stop" while still emitting tool calls.
		needsToolExecution := goai.NeedsToolExecution(assistantMsg) || len(toolCalls) > 0
		if !needsToolExecution {
			if steerMsgs, err := r.dequeueSteeringMessages(ctx, sessionID); err != nil {
				log.Printf("steering dequeue error after direct response: %v", err)
			} else if len(steerMsgs) > 0 {
				pendingSteering = append(pendingSteering, steerMsgs...)
				continue
			}
			// Final response — no tool calls. Persist and finish.
			log.Printf("inference [%s]: final response (%d chars, %d iterations)", iterLabel, len(textContent), iter+1)
			r.persistUsage(s, turnID, sessionID, &totalUsage, iter+1)

			msgID := store.NowID("msg")
			_ = s.AddMessage(ctx, msgID, sessionID, "assistant", textContent, map[string]any{
				"kind": "chat", "source": "inference", "model": model,
				"turn_id": turnID, "agent_id": agentID, "iterations": iter + 1,
			})

			r.broadcastPost(sessionID, turnID, msgID, textContent, agentID)
			_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookMessageEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter + 1, Payload: map[string]any{"chars": len(textContent)}})
			_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookTurnEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter + 1, Payload: map[string]any{"status": "completed"}})
			r.finishTurnOK(s, turnID, sessionID, iter+1)
			return
		}

		// Tool calls detected — execute them
		log.Printf("inference [%s]: %d tool call(s)", iterLabel, len(toolCalls))

		// Store the assistant message with tool calls for audit trail
		toolCallSummary := textContent
		for _, tc := range toolCalls {
			if toolCallSummary != "" {
				toolCallSummary += "\n"
			}
			toolCallSummary += fmt.Sprintf("[tool_call: %s]", tc.Name)
		}
		_ = s.AddMessage(ctx, store.NowID("msg"), sessionID, "assistant", toolCallSummary, map[string]any{
			"kind": "tool_calls", "source": "inference", "model": model,
			"turn_id": turnID, "agent_id": agentID,
		})

		_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookToolExecutionStart, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter + 1, Payload: map[string]any{"count": len(toolCalls)}})
		skipRemainingTools := false
		for i, call := range toolCalls {
			if ctx.Err() != nil {
				r.finishTurn(s, turnID, sessionID, agentID, "cancelled", "Turn cancelled during tool execution", "")
				return
			}

			if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookToolCall, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter + 1, ToolCall: &call, Payload: map[string]any{"tool": call.Name, "tool_call_id": call.ID, "arguments": call.Arguments}}); err != nil {
				log.Printf("hook tool_call error: %v", err)
			} else if resp.Block {
				toolErr := fmt.Errorf("blocked by hook: %s", stringValue(resp.Reason, "tool call blocked"))
				errText := fmt.Sprintf("Error: %v", toolErr)
				goai.AppendToolResult(convCtx, call.ID, call.Name, errText, true)
				_ = s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", errText, map[string]any{"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": true, "turn_id": turnID})
				continue
			} else if resp.ToolCall != nil {
				call = *resp.ToolCall
			}
			if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookApproveTool, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter + 1, ToolCall: &call, Payload: map[string]any{"tool": call.Name, "tool_call_id": call.ID, "arguments": call.Arguments}}); err != nil {
				log.Printf("hook approve_tool error: %v", err)
			} else if resp.Block {
				toolErr := fmt.Errorf("blocked by hook: %s", stringValue(resp.Reason, "tool not approved"))
				errText := fmt.Sprintf("Error: %v", toolErr)
				goai.AppendToolResult(convCtx, call.ID, call.Name, errText, true)
				_ = s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", errText, map[string]any{"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": true, "turn_id": turnID})
				continue
			} else if resp.ToolCall != nil {
				call = *resp.ToolCall
			}

			_ = s.UpdateTurnStatusAndPhase(ctx, turnID, "running", "waiting_on_tools")
			_ = s.AppendTurnEvent(ctx, turnID, sessionID, "tool.started", map[string]any{
				"phase": "tool", "tool": call.Name, "checkpoint": true,
				"tool_call_id": call.ID, "iteration": iter + 1,
			})

			r.engine.broadcast(sessionID, map[string]any{
				"type": "agent_status", "chat_jid": "gi:" + sessionID,
				"title": fmt.Sprintf("Running: %s", call.Name), "status": "running", "turn_id": turnID,
			})

			toolResult, toolErr := r.executeTool(ctx, call, sessionID)

			if toolErr != nil {
				log.Printf("tool [%s] error: %v", call.Name, toolErr)
				r.engine.broadcast(sessionID, map[string]any{"type": "tool_failed", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "error": toolErr.Error()})
				_ = s.AppendTurnEvent(ctx, turnID, sessionID, "tool.failed", map[string]any{
					"phase": "tool", "tool": call.Name, "checkpoint": true,
					"tool_call_id": call.ID, "error": toolErr.Error(),
				})
				errText := fmt.Sprintf("Error: %v", toolErr)
				goai.AppendToolResult(convCtx, call.ID, call.Name, errText, true)
				_ = s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", errText, map[string]any{
					"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": true, "turn_id": turnID,
				})
				lastToolFailureSig, repeatedToolFailureCount = nextRepeatedToolFailureCount(lastToolFailureSig, repeatedToolFailureCount, call, toolErr)
				if repeatedToolFailureCount >= repeatedToolFailureLimit {
					msg := fmt.Sprintf("Aborting after %d repeated identical tool failures: %v", repeatedToolFailureCount, toolErr)
					log.Printf("tool [%s] repeated failure guard tripped: %s", call.Name, msg)
					r.persistUsage(s, turnID, sessionID, &totalUsage, iter+1)
					r.finishTurn(s, turnID, sessionID, agentID, "failed", msg, "repeated_tool_failure")
					return
				}
			} else {
				if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookToolResult, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter + 1, ToolCall: &call, ToolResult: toolResult, Payload: map[string]any{"tool": call.Name, "tool_call_id": call.ID, "is_error": false}}); err != nil {
					log.Printf("hook tool_result error: %v", err)
				} else if resp.ToolResult != nil {
					toolResult = *resp.ToolResult
				}
				// Truncate very large results to avoid blowing context
				displayResult := toolResult
				if len(displayResult) > 100000 {
					displayResult = displayResult[:100000] + "\n... (truncated)"
				}
				r.engine.broadcast(sessionID, map[string]any{"type": "tool_finished", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "output_length": len(toolResult)})
				_ = s.AppendTurnEvent(ctx, turnID, sessionID, "tool.finished", map[string]any{
					"phase": "tool", "tool": call.Name, "checkpoint": true,
					"tool_call_id": call.ID, "output_length": len(toolResult),
				})
				goai.AppendToolResult(convCtx, call.ID, call.Name, displayResult, false)
				_ = s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", displayResult, map[string]any{
					"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": false, "turn_id": turnID,
				})
				lastToolFailureSig = ""
				repeatedToolFailureCount = 0
			}
			if steerMsgs, err := r.dequeueSteeringMessages(ctx, sessionID); err != nil {
				log.Printf("steering dequeue error after tool: %v", err)
			} else if len(steerMsgs) > 0 {
				pendingSteering = append(pendingSteering, steerMsgs...)
				if i+1 < len(toolCalls) {
					r.skipRemainingToolCalls(ctx, sessionID, turnID, convCtx, toolCalls, i+1)
					skipRemainingTools = true
				}
				break
			}
		}
		_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookToolExecutionEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter + 1, Payload: map[string]any{"count": len(toolCalls)}})
		if skipRemainingTools || len(pendingSteering) > 0 {
			continue
		}
		_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookTurnEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter + 1, Payload: map[string]any{"status": "tools"}})
		// Loop continues — next iteration will call LLM with tool results
	}

	// Budget exhausted
	log.Printf("inference: max iterations (%d) reached for turn %s", maxIter, turnID)
	r.persistUsage(s, turnID, sessionID, &totalUsage, maxIter)
	r.finishTurn(s, turnID, sessionID, agentID, "completed", fmt.Sprintf("Reached maximum iteration limit (%d). The task may be incomplete.", maxIter), "")
}

// executeTool dispatches a single tool call and returns the text result.
func (r *sessionRunner) executeTool(ctx context.Context, call goai.ToolCall, sessionID string) (string, error) {
	tool, ok := r.engine.tools.Get(call.Name)
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", call.Name)
	}
	return tool.Executor(ctx, ToolRuntime{
		Engine:        r.engine,
		Runner:        r,
		Store:         r.store,
		SessionID:     sessionID,
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
	_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "inference.finished", map[string]any{
		"phase": "inference", "checkpoint": true, "usage": usageMap, "iterations": iterations,
	})
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
func (r *sessionRunner) finishTurnOK(s *store.Store, turnID, sessionID string, iterations int) {
	if staged, stagedTurnID, err := r.engine.stageQueuedSteeringContinuation(context.Background(), sessionID); err != nil {
		log.Printf("steering final checkpoint: %v", err)
	} else if staged {
		_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "steering.final_checkpoint", map[string]any{"phase": "steering", "checkpoint": true, "staged_turn_id": stagedTurnID})
	}
	_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "turn.finished", map[string]any{
		"phase": "turn", "checkpoint": true, "status": "completed", "iterations": iterations,
	})
	_ = s.UpdateTurnStatusAndPhase(context.Background(), turnID, "completed", "completed")
	_ = s.MarkTurnFinished(context.Background(), turnID)
	_ = s.TouchSessionState(context.Background(), sessionID, map[string]any{"status": "idle", "active_turn_id": nil})
	r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "title": "", "status": "idle"})
}

// finishTurn persists a terminal status and optional system message.
func (r *sessionRunner) finishTurn(s *store.Store, turnID, sessionID, agentID, status, systemMsg, failureKind string) {
	if staged, stagedTurnID, err := r.engine.stageQueuedSteeringContinuation(context.Background(), sessionID); err != nil {
		log.Printf("steering final checkpoint: %v", err)
	} else if staged {
		_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "steering.final_checkpoint", map[string]any{"phase": "steering", "checkpoint": true, "staged_turn_id": stagedTurnID})
	}
	if systemMsg != "" {
		msgID := store.NowID("msg")
		_ = s.AddMessage(context.Background(), msgID, sessionID, "assistant", systemMsg, map[string]any{
			"kind": "chat", "source": "system", "turn_id": turnID, "agent_id": agentID,
		})
		if status == "completed" || status == "failed" {
			r.broadcastPost(sessionID, turnID, msgID, systemMsg, agentID)
		}
	}
	if failureKind != "" {
		markTurnFailure(s, turnID, sessionID, failureKind, systemMsg)
	}
	_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "turn.finished", map[string]any{
		"phase": "turn", "checkpoint": true, "status": status,
	})
	_ = s.UpdateTurnStatusAndPhase(context.Background(), turnID, status, terminalPhaseForStatus(status))
	_ = s.MarkTurnFinished(context.Background(), turnID)
	_ = s.TouchSessionState(context.Background(), sessionID, map[string]any{"status": "idle", "active_turn_id": nil})
	r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "title": "", "status": "idle"})
}
