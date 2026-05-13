package turn

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

type toolExecutionOutcome struct {
	terminated               bool
	skipRemainingTools       bool
	pendingSteering          []store.SteeringMessage
	lastToolFailureSig       string
	repeatedToolFailureCount int
}

var streamWithToolsWithHooks = inference.StreamWithToolsWithHooks

func isCancellationError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

type hookAbortError struct {
	reason string
	hard   bool
}

func (e hookAbortError) Error() string {
	if strings.TrimSpace(e.reason) == "" {
		if e.hard {
			return "turn hard-aborted by hook"
		}
		return "turn aborted by hook"
	}
	return e.reason
}

func hookAbortFromResponse(resp HookResponse, fallback string) error {
	if !resp.Cancel {
		return nil
	}
	reason := strings.TrimSpace(resp.Reason)
	if reason == "" {
		reason = fallback
	}
	return hookAbortError{reason: reason, hard: boolValue(resp.Payload["hard_abort"])}
}

func directToolResultFromHook(resp HookResponse) (string, bool) {
	if resp.ToolResult != nil {
		return *resp.ToolResult, true
	}
	if resp.Handled && strings.TrimSpace(resp.Message) != "" {
		return resp.Message, true
	}
	return "", false
}

func providerRequestReplacementFromHook(resp HookResponse) (any, bool) {
	if resp.Payload == nil {
		return nil, false
	}
	replacement, ok := resp.Payload["request"]
	if !ok || replacement == nil {
		return nil, false
	}
	return replacement, true
}

func (r *sessionRunner) assembleAgentContext(ctx context.Context, s *store.Store, turnID, sessionID, model, agentID string) *goai.Context {
	sysPrompt := r.engine.systemPrompt
	if sysPrompt == "" {
		sysPrompt = "You are a helpful coding assistant."
	}

	var turnMetadata map[string]any
	if turnRec, err := s.GetTurn(ctx, turnID); err == nil {
		turnMetadata = turnRec.Metadata
	}

	msgs, _ := s.ListMessages(ctx, sessionID)
	convCtx := &goai.Context{
		SystemPrompt: sysPrompt,
		Tools:        r.engine.toolDefsForMetadata(turnMetadata),
	}
	for _, m := range msgs {
		switch m.Role {
		case "user":
			convCtx.Messages = append(convCtx.Messages, goai.UserMessage(m.Content))
		case "assistant":
			convCtx.Messages = append(convCtx.Messages, goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: m.Content}}})
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
	return convCtx
}

func (r *sessionRunner) prepareAgentIteration(ctx context.Context, sessionID, turnID, model, agentID string, iter int, convCtx *goai.Context, pendingSteering []store.SteeringMessage) []store.SteeringMessage {
	if steerMsgs, err := r.dequeueSteeringMessages(ctx, sessionID); err != nil {
		log.Printf("steering dequeue error: %v", err)
	} else if len(steerMsgs) > 0 {
		pendingSteering = append(pendingSteering, steerMsgs...)
	}
	if len(pendingSteering) > 0 {
		r.injectSteeringMessages(ctx, sessionID, turnID, convCtx, pendingSteering)
		pendingSteering = nil
	}
	_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookTurnStart, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter})
	if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookContext, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, SystemPrompt: convCtx.SystemPrompt, Messages: convCtx.Messages, Tools: convCtx.Tools}); err != nil {
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
	return pendingSteering
}

func (r *sessionRunner) runProviderIteration(ctx context.Context, s *store.Store, turnID, sessionID, model, agentID string, iter, maxIter int, convCtx *goai.Context) (*inference.StreamResult, error) {
	requestCtx := &goai.Context{SystemPrompt: convCtx.SystemPrompt, Messages: append([]goai.Message(nil), convCtx.Messages...), Tools: append([]goai.Tool(nil), convCtx.Tools...)}
	if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookBeforeProviderRequest, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, SystemPrompt: convCtx.SystemPrompt, Messages: convCtx.Messages, Tools: convCtx.Tools, Payload: map[string]any{"model": model, "messages": len(convCtx.Messages), "tools": len(convCtx.Tools), "stage": "context"}}); err != nil {
		log.Printf("hook before_provider_request error: %v", err)
	} else {
		if abortErr := hookAbortFromResponse(resp, "aborted before provider request by hook"); abortErr != nil {
			return nil, abortErr
		}
		if resp.SystemPrompt != "" {
			convCtx.SystemPrompt = resp.SystemPrompt
			requestCtx.SystemPrompt = resp.SystemPrompt
		}
		if resp.Messages != nil {
			convCtx.Messages = resp.Messages
			requestCtx.Messages = append([]goai.Message(nil), resp.Messages...)
		}
		if resp.Tools != nil {
			convCtx.Tools = resp.Tools
			requestCtx.Tools = append([]goai.Tool(nil), resp.Tools...)
		}
		if strings.TrimSpace(resp.Message) != "" {
			requestCtx.Messages = append([]goai.Message{goai.UserMessage(resp.Message)}, requestCtx.Messages...)
		}
	}
	warnStore("update turn running phase", s.UpdateTurnStatusAndPhase(ctx, turnID, "running", "running"))
	r.emitTurnStateHook(ctx, sessionID, turnID, agentID, model, "running", "running", map[string]any{"reason": "provider_iteration", "iteration": iter})
	warnStore("append inference.started event", s.AppendTurnEvent(ctx, turnID, sessionID, "inference.started", map[string]any{"phase": "inference", "model": model, "iteration": iter, "checkpoint": true}))
	iterLabel := fmt.Sprintf("iter=%d/%d", iter, maxIter)
	log.Printf("inference [%s]: calling %s", iterLabel, model)

	r.engine.broadcast(sessionID, map[string]any{
		"type": "agent_status", "chat_jid": "gi:" + sessionID,
		"title": fmt.Sprintf("Thinking… (%d)", iter), "status": "running", "turn_id": turnID,
	})

	responseObserved := false
	result, inferErr := streamWithToolsWithHooks(ctx, model, requestCtx, func(ev map[string]any) {
		ev["chat_jid"] = "gi:" + sessionID
		ev["turn_id"] = turnID
		ev["iteration"] = iter
		switch ev["type"] {
		case "text_delta":
			_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookMessageUpdate, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"delta": ev["delta"]}})
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
	}, &inference.StreamHooks{
		OnPayload: func(payload any, modelDef *goai.Model) (any, error) {
			hookPayload := map[string]any{"ok": true, "request": payload, "stage": "payload"}
			if modelDef != nil {
				hookPayload["provider"] = string(modelDef.Provider)
				hookPayload["api"] = string(modelDef.Api)
				hookPayload["model_id"] = modelDef.ID
			}
			resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookBeforeProviderRequest, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, SystemPrompt: requestCtx.SystemPrompt, Messages: requestCtx.Messages, Tools: requestCtx.Tools, Payload: hookPayload})
			if err != nil {
				return nil, err
			}
			if abortErr := hookAbortFromResponse(resp, "aborted before provider request send by hook"); abortErr != nil {
				return nil, abortErr
			}
			if replacement, ok := providerRequestReplacementFromHook(resp); ok {
				return replacement, nil
			}
			return payload, nil
		},
		OnResponse: func(status int, headers map[string]string, modelDef *goai.Model) {
			responseObserved = true
			payload := map[string]any{"ok": true, "status": status, "headers": headers}
			if modelDef != nil {
				payload["provider"] = string(modelDef.Provider)
				payload["api"] = string(modelDef.Api)
				payload["model_id"] = modelDef.ID
			}
			if _, err := r.engine.emitHook(ctx, HookRequest{Name: HookAfterProviderResponse, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: payload}); err != nil {
				log.Printf("hook after_provider_response error: %v", err)
			}
		},
	})
	if !responseObserved {
		if _, err := r.engine.emitHook(ctx, HookRequest{Name: HookAfterProviderResponse, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"ok": inferErr == nil}}); err != nil {
			log.Printf("hook after_provider_response error: %v", err)
		}
	}
	return result, inferErr
}

func (r *sessionRunner) executeToolCallsPhase(ctx context.Context, s *store.Store, turnID, sessionID, model, agentID string, iter int, convCtx *goai.Context, toolCalls []goai.ToolCall, pendingSteering []store.SteeringMessage, lastToolFailureSig string, repeatedToolFailureCount int, totalUsage *goai.Usage) toolExecutionOutcome {
	outcome := toolExecutionOutcome{
		pendingSteering:          pendingSteering,
		lastToolFailureSig:       lastToolFailureSig,
		repeatedToolFailureCount: repeatedToolFailureCount,
	}
	_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookToolExecutionStart, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"count": len(toolCalls)}})
	defer func() {
		_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookToolExecutionEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, Payload: map[string]any{"count": len(toolCalls)}})
	}()
	for i, call := range toolCalls {
		if ctx.Err() != nil {
			r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled during tool execution", "")
			outcome.terminated = true
			return outcome
		}

		if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookToolCall, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, ToolCall: &call, Payload: map[string]any{"tool": call.Name, "tool_call_id": call.ID, "arguments": call.Arguments}}); err != nil {
			log.Printf("hook tool_call error: %v", err)
		} else {
			if abortErr := hookAbortFromResponse(resp, fmt.Sprintf("tool %s aborted by hook", call.Name)); abortErr != nil {
				r.engine.PublishRuntimeHookDecisionEvent("hook_abort", HookRequest{Name: HookToolCall, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "tool_call", "reason": abortErr.Error()})
				r.finishTurn(s, turnID, sessionID, agentID, model, "aborted", abortErr.Error(), "hook_abort")
				outcome.terminated = true
				return outcome
			}
			if resp.ToolCall != nil {
				r.engine.PublishRuntimeHookDecisionEvent("hook_modify", HookRequest{Name: HookToolCall, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "tool_call", "modified_tool": resp.ToolCall.Name, "modified_tool_call_id": resp.ToolCall.ID})
				call = *resp.ToolCall
			}
			if injectedResult, ok := directToolResultFromHook(resp); ok {
				r.engine.PublishRuntimeHookDecisionEvent("hook_respond", HookRequest{Name: HookToolCall, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "tool_call", "source": "hook", "output_length": len(injectedResult)})
				displayResult := injectedResult
				if len(displayResult) > 100000 {
					displayResult = displayResult[:100000] + "\n... (truncated)"
				}
				r.engine.broadcast(sessionID, map[string]any{"type": "tool_finished", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "output_length": len(injectedResult)})
				r.engine.PublishRuntimeToolEvent("tool_finished", sessionID, turnID, agentID, call.Name, call.ID, iter, nil, map[string]any{"phase": "tool", "output_length": len(injectedResult), "source": "hook", "hook_phase": "tool_call"})
				warnStore("append hook-responded tool.finished event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.finished", map[string]any{
					"phase":         "tool",
					"tool":          call.Name,
					"checkpoint":    true,
					"tool_call_id":  call.ID,
					"output_length": len(injectedResult),
					"source":        "hook",
					"hook_phase":    "tool_call",
				}))
				goai.AppendToolResult(convCtx, call.ID, call.Name, displayResult, false)
				warnStore("add injected tool_result message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", displayResult, map[string]any{"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": false, "turn_id": turnID, "source": "hook", "hook_phase": "tool_call"}))
				outcome.lastToolFailureSig = ""
				outcome.repeatedToolFailureCount = 0
				if steerMsgs, err := r.dequeueSteeringMessages(ctx, sessionID); err != nil {
					log.Printf("steering dequeue error after hook tool response: %v", err)
				} else if len(steerMsgs) > 0 {
					outcome.pendingSteering = append(outcome.pendingSteering, steerMsgs...)
					if i+1 < len(toolCalls) {
						r.skipRemainingToolCalls(ctx, sessionID, turnID, convCtx, toolCalls, i+1)
						outcome.skipRemainingTools = true
					}
				}
				if outcome.skipRemainingTools || len(outcome.pendingSteering) > 0 {
					return outcome
				}
				continue
			}
			if resp.Block {
				reason := stringValue(resp.Reason, "tool call blocked")
				r.engine.PublishRuntimeHookDecisionEvent("hook_deny", HookRequest{Name: HookToolCall, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "tool_call", "reason": reason})
				warnStore("append hook-denied tool.skipped event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.skipped", map[string]any{
					"phase":        "tool",
					"checkpoint":   true,
					"tool":         call.Name,
					"tool_call_id": call.ID,
					"reason":       reason,
					"hook_phase":   "tool_call",
				}))
				r.engine.broadcast(sessionID, map[string]any{"type": "tool_skipped", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "reason": reason})
				r.engine.PublishRuntimeToolEvent("tool_skipped", sessionID, turnID, agentID, call.Name, call.ID, iter, nil, map[string]any{"reason": reason, "phase": "tool", "hook_phase": "tool_call"})
				toolErr := fmt.Errorf("blocked by hook: %s", reason)
				errText := fmt.Sprintf("Error: %v", toolErr)
				goai.AppendToolResult(convCtx, call.ID, call.Name, errText, true)
				warnStore("add blocked tool_result message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", errText, map[string]any{"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": true, "turn_id": turnID, "skipped": true, "skip_reason": reason, "hook_phase": "tool_call"}))
				continue
			}
		}
		if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookApproveTool, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, ToolCall: &call, Payload: map[string]any{"tool": call.Name, "tool_call_id": call.ID, "arguments": call.Arguments}}); err != nil {
			log.Printf("hook approve_tool error: %v", err)
		} else {
			if abortErr := hookAbortFromResponse(resp, fmt.Sprintf("tool %s approval aborted by hook", call.Name)); abortErr != nil {
				r.engine.PublishRuntimeHookDecisionEvent("hook_abort", HookRequest{Name: HookApproveTool, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "approve_tool", "reason": abortErr.Error()})
				r.finishTurn(s, turnID, sessionID, agentID, model, "aborted", abortErr.Error(), "hook_abort")
				outcome.terminated = true
				return outcome
			}
			if resp.Block {
				reason := stringValue(resp.Reason, "tool not approved")
				r.engine.PublishRuntimeHookDecisionEvent("hook_deny", HookRequest{Name: HookApproveTool, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "approve_tool", "reason": reason})
				warnStore("append approve-denied tool.skipped event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.skipped", map[string]any{
					"phase":        "tool",
					"checkpoint":   true,
					"tool":         call.Name,
					"tool_call_id": call.ID,
					"reason":       reason,
					"hook_phase":   "approve_tool",
				}))
				r.engine.broadcast(sessionID, map[string]any{"type": "tool_skipped", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "reason": reason})
				r.engine.PublishRuntimeToolEvent("tool_skipped", sessionID, turnID, agentID, call.Name, call.ID, iter, nil, map[string]any{"reason": reason, "phase": "tool", "hook_phase": "approve_tool"})
				toolErr := fmt.Errorf("blocked by hook: %s", reason)
				errText := fmt.Sprintf("Error: %v", toolErr)
				goai.AppendToolResult(convCtx, call.ID, call.Name, errText, true)
				warnStore("add denied tool_result message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", errText, map[string]any{"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": true, "turn_id": turnID, "skipped": true, "skip_reason": reason, "hook_phase": "approve_tool"}))
				continue
			}
			if resp.ToolCall != nil {
				r.engine.PublishRuntimeHookDecisionEvent("hook_modify", HookRequest{Name: HookApproveTool, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "approve_tool", "modified_tool": resp.ToolCall.Name, "modified_tool_call_id": resp.ToolCall.ID})
				call = *resp.ToolCall
			}
		}

		warnStore("update turn waiting_on_tools phase", s.UpdateTurnStatusAndPhase(ctx, turnID, "running", "waiting_on_tools"))
		r.emitTurnStateHook(ctx, sessionID, turnID, agentID, model, "running", "waiting_on_tools", map[string]any{"reason": "tool_execution", "tool": call.Name, "iteration": iter})
		r.engine.PublishRuntimeToolEvent("tool_started", sessionID, turnID, agentID, call.Name, call.ID, iter, nil, map[string]any{"phase": "tool"})
		warnStore("append tool.started event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.started", map[string]any{
			"phase": "tool", "tool": call.Name, "checkpoint": true,
			"tool_call_id": call.ID, "iteration": iter,
		}))

		r.engine.broadcast(sessionID, map[string]any{
			"type": "agent_status", "chat_jid": "gi:" + sessionID,
			"title": fmt.Sprintf("Running: %s", call.Name), "status": "running", "turn_id": turnID,
		})

		toolResult, toolErr := r.executeTool(ctx, call, sessionID, turnID)
		if toolErr != nil {
			if ctx.Err() != nil || isCancellationError(toolErr) {
				r.finishTurn(s, turnID, sessionID, agentID, model, "cancelled", "Turn cancelled during tool execution", "")
				outcome.terminated = true
				return outcome
			}
			log.Printf("tool [%s] error: %v", call.Name, toolErr)
			r.engine.broadcast(sessionID, map[string]any{"type": "tool_failed", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "error": toolErr.Error()})
			r.engine.PublishRuntimeToolEvent("tool_failed", sessionID, turnID, agentID, call.Name, call.ID, iter, toolErr, map[string]any{"phase": "tool"})
			warnStore("append tool.failed event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.failed", map[string]any{
				"phase": "tool", "tool": call.Name, "checkpoint": true,
				"tool_call_id": call.ID, "error": toolErr.Error(),
			}))
			errText := fmt.Sprintf("Error: %v", toolErr)
			goai.AppendToolResult(convCtx, call.ID, call.Name, errText, true)
			warnStore("add errored tool_result message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", errText, map[string]any{
				"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": true, "turn_id": turnID,
			}))
			outcome.lastToolFailureSig, outcome.repeatedToolFailureCount = nextRepeatedToolFailureCount(outcome.lastToolFailureSig, outcome.repeatedToolFailureCount, call, toolErr)
			if outcome.repeatedToolFailureCount >= repeatedToolFailureLimit {
				msg := fmt.Sprintf("Aborting after %d repeated identical tool failures: %v", outcome.repeatedToolFailureCount, toolErr)
				log.Printf("tool [%s] repeated failure guard tripped: %s", call.Name, msg)
				r.persistUsage(s, turnID, sessionID, totalUsage, iter)
				r.finishTurn(s, turnID, sessionID, agentID, model, "failed", msg, "repeated_tool_failure")
				outcome.terminated = true
				return outcome
			}
		} else {
			if resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookToolResult, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Iteration: iter, ToolCall: &call, ToolResult: toolResult, Payload: map[string]any{"tool": call.Name, "tool_call_id": call.ID, "is_error": false}}); err != nil {
				log.Printf("hook tool_result error: %v", err)
			} else {
				if abortErr := hookAbortFromResponse(resp, fmt.Sprintf("tool %s result aborted by hook", call.Name)); abortErr != nil {
					r.engine.PublishRuntimeHookDecisionEvent("hook_abort", HookRequest{Name: HookToolResult, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "tool_result", "reason": abortErr.Error()})
					r.finishTurn(s, turnID, sessionID, agentID, model, "aborted", abortErr.Error(), "hook_abort")
					outcome.terminated = true
					return outcome
				}
				if resp.ToolResult != nil {
					r.engine.PublishRuntimeHookDecisionEvent("hook_modify", HookRequest{Name: HookToolResult, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "tool_result", "output_length": len(*resp.ToolResult)})
					toolResult = *resp.ToolResult
				}
			}
			displayResult := toolResult
			if len(displayResult) > 100000 {
				displayResult = displayResult[:100000] + "\n... (truncated)"
			}
			r.engine.broadcast(sessionID, map[string]any{"type": "tool_finished", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "output_length": len(toolResult)})
			r.engine.PublishRuntimeToolEvent("tool_finished", sessionID, turnID, agentID, call.Name, call.ID, iter, nil, map[string]any{"phase": "tool", "output_length": len(toolResult)})
			warnStore("append tool.finished event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.finished", map[string]any{
				"phase": "tool", "tool": call.Name, "checkpoint": true,
				"tool_call_id": call.ID, "output_length": len(toolResult),
			}))
			goai.AppendToolResult(convCtx, call.ID, call.Name, displayResult, false)
			warnStore("add successful tool_result message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", displayResult, map[string]any{
				"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": false, "turn_id": turnID,
			}))
			outcome.lastToolFailureSig = ""
			outcome.repeatedToolFailureCount = 0
		}
		if steerMsgs, err := r.dequeueSteeringMessages(ctx, sessionID); err != nil {
			log.Printf("steering dequeue error after tool: %v", err)
		} else if len(steerMsgs) > 0 {
			outcome.pendingSteering = append(outcome.pendingSteering, steerMsgs...)
			if i+1 < len(toolCalls) {
				r.skipRemainingToolCalls(ctx, sessionID, turnID, convCtx, toolCalls, i+1)
				outcome.skipRemainingTools = true
			}
			return outcome
		}
	}
	return outcome
}

func (r *sessionRunner) appendFinalSteeringCheckpoint(s *store.Store, turnID, sessionID string) {
	bgCtx := r.engine.backgroundContext()
	if staged, stagedTurnID, err := r.engine.stageQueuedSteeringContinuation(bgCtx, sessionID); err != nil {
		log.Printf("steering final checkpoint: %v", err)
	} else if staged {
		warnStore("append steering final checkpoint", s.AppendTurnEvent(bgCtx, turnID, sessionID, "steering.final_checkpoint", map[string]any{"phase": "steering", "checkpoint": true, "staged_turn_id": stagedTurnID}))
	}
}
