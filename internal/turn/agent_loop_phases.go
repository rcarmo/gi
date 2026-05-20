package turn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/compaction"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/logutil"
	"github.com/rcarmo/gi/internal/routing"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
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
	return hookAbortError{reason: reason, hard: store.BoolValue(resp.Payload["hard_abort"])}
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
	compaction.MaybeCompactContext(ctx, compaction.RuntimeRequest{SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Settings: r.engine.runtimeCfg.Compaction}, convCtx, compaction.RuntimeOps{BackgroundContext: r.engine.backgroundContext, BeforeCompact: func(ctx context.Context, payload map[string]any, messages []goai.Message) (compaction.HookDecision, error) {
		resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookSessionBeforeCompact, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Payload: payload, Messages: messages})
		return compaction.HookDecision{Cancel: resp.Cancel, Block: resp.Block, Payload: resp.Payload}, err
	}, AfterCompact: func(ctx context.Context, payload map[string]any) {
		_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookSessionCompact, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Payload: payload})
	}, UpdateTurnStatusAndPhase: r.store.UpdateTurnStatusAndPhase, AppendTurnEvent: r.store.AppendTurnEvent, TouchSessionActiveTurn: r.store.TouchSessionActiveTurn, AddMessage: r.store.AddMessage, Broadcast: r.engine.broadcast, Warn: logutil.WarnIfErr})
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
	logutil.WarnIfErr("update turn running phase", s.UpdateTurnStatusAndPhase(ctx, turnID, "running", "running"))
	r.emitTurnStateHook(ctx, sessionID, turnID, agentID, model, "running", "running", map[string]any{"reason": "provider_iteration", "iteration": iter})
	logutil.WarnIfErr("append inference.started event", s.AppendTurnEvent(ctx, turnID, sessionID, "inference.started", map[string]any{"phase": "inference", "model": model, "iteration": iter, "checkpoint": true}))
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
				logutil.WarnIfErr("append hook-responded tool.finished event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.finished", map[string]any{
					"phase":         "tool",
					"tool":          call.Name,
					"checkpoint":    true,
					"tool_call_id":  call.ID,
					"output_length": len(injectedResult),
					"source":        "hook",
					"hook_phase":    "tool_call",
				}))
				goai.AppendToolResult(convCtx, call.ID, call.Name, displayResult, false)
				logutil.WarnIfErr("add injected tool_result message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", displayResult, map[string]any{"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": false, "turn_id": turnID, "source": "hook", "hook_phase": "tool_call"}))
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
				reason := store.StringValue(resp.Reason, "tool call blocked")
				r.engine.PublishRuntimeHookDecisionEvent("hook_deny", HookRequest{Name: HookToolCall, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "tool_call", "reason": reason})
				logutil.WarnIfErr("append hook-denied tool.skipped event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.skipped", map[string]any{
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
				logutil.WarnIfErr("add blocked tool_result message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", errText, map[string]any{"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": true, "turn_id": turnID, "skipped": true, "skip_reason": reason, "hook_phase": "tool_call"}))
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
				reason := store.StringValue(resp.Reason, "tool not approved")
				r.engine.PublishRuntimeHookDecisionEvent("hook_deny", HookRequest{Name: HookApproveTool, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "approve_tool", "reason": reason})
				logutil.WarnIfErr("append approve-denied tool.skipped event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.skipped", map[string]any{
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
				logutil.WarnIfErr("add denied tool_result message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", errText, map[string]any{"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": true, "turn_id": turnID, "skipped": true, "skip_reason": reason, "hook_phase": "approve_tool"}))
				continue
			}
			if resp.ToolCall != nil {
				r.engine.PublishRuntimeHookDecisionEvent("hook_modify", HookRequest{Name: HookApproveTool, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Iteration: iter, ToolCall: &call}, map[string]any{"phase": "approve_tool", "modified_tool": resp.ToolCall.Name, "modified_tool_call_id": resp.ToolCall.ID})
				call = *resp.ToolCall
			}
		}

		logutil.WarnIfErr("update turn waiting_on_tools phase", s.UpdateTurnStatusAndPhase(ctx, turnID, "running", "waiting_on_tools"))
		r.emitTurnStateHook(ctx, sessionID, turnID, agentID, model, "running", "waiting_on_tools", map[string]any{"reason": "tool_execution", "tool": call.Name, "iteration": iter})
		r.engine.PublishRuntimeToolEvent("tool_started", sessionID, turnID, agentID, call.Name, call.ID, iter, nil, map[string]any{"phase": "tool"})
		logutil.WarnIfErr("append tool.started event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.started", map[string]any{
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
			logutil.WarnIfErr("append tool.failed event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.failed", map[string]any{
				"phase": "tool", "tool": call.Name, "checkpoint": true,
				"tool_call_id": call.ID, "error": toolErr.Error(),
			}))
			errText := fmt.Sprintf("Error: %v", toolErr)
			goai.AppendToolResult(convCtx, call.ID, call.Name, errText, true)
			logutil.WarnIfErr("add errored tool_result message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", errText, map[string]any{
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
			logutil.WarnIfErr("append tool.finished event", s.AppendTurnEvent(ctx, turnID, sessionID, "tool.finished", map[string]any{
				"phase": "tool", "tool": call.Name, "checkpoint": true,
				"tool_call_id": call.ID, "output_length": len(toolResult),
			}))
			goai.AppendToolResult(convCtx, call.ID, call.Name, displayResult, false)
			logutil.WarnIfErr("add successful tool_result message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", displayResult, map[string]any{
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
		logutil.WarnIfErr("append steering final checkpoint", s.AppendTurnEvent(bgCtx, turnID, sessionID, "steering.final_checkpoint", map[string]any{"phase": "steering", "checkpoint": true, "staged_turn_id": stagedTurnID}))
	}
}

type preparedTurnRun struct {
	turn            *store.Turn
	sessionID       string
	turnID          string
	prompt          string
	intent          string
	model           string
	agentID         string
	initialSteering []store.SteeringMessage
}

func (r *sessionRunner) appendCleanupHandoffFailure(ctx context.Context, sessionID, turnID, stage string, err error) {
	if strings.TrimSpace(sessionID) == "" || strings.TrimSpace(turnID) == "" || err == nil {
		return
	}
	logutil.WarnIfErr("append cleanup handoff failure event", r.store.AppendTurnEvent(ctx, turnID, sessionID, "turn.cleanup_handoff_failed", map[string]any{
		"phase":      "cleanup",
		"checkpoint": true,
		"stage":      stage,
		"error":      err.Error(),
	}))
}

func (r *sessionRunner) cleanupTurnRun(sessionID, claimToken string, active *runningTurn) {
	ctx := r.engine.backgroundContext()
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.store.ReleaseSessionActiveTurn(ctx, sessionID, claimToken); err != nil {
		log.Printf("turn coordination: release session active turn failed: %v", err)
		r.appendCleanupHandoffFailure(ctx, sessionID, claimToken, "release_active_claim", err)
		if r.current == active {
			r.current = nil
		}
		return
	}
	if r.current == active {
		r.current = nil
	}
	if err := r.store.SyncSessionQueueCount(ctx, sessionID); err != nil {
		log.Printf("turn coordination: sync session queue count failed: %v", err)
		r.appendCleanupHandoffFailure(ctx, sessionID, claimToken, "sync_queue_count", err)
		return
	}
	if hook := r.engine.beforeCleanupNextWorkHook; hook != nil {
		hook(ctx, sessionID)
	}
	launched, err := r.engine.startNextQueuedTurnLocked(ctx, r, sessionID)
	if err != nil {
		log.Printf("turn coordination: launch queued turn failed: %v", err)
	} else if launched {
		if strings.TrimSpace(claimToken) != "" {
			logutil.WarnIfErr("append cleanup handoff event", r.store.AppendTurnEvent(ctx, claimToken, sessionID, "turn.cleanup_handoff", map[string]any{"phase": "cleanup", "checkpoint": true, "handoff": "next_queued_turn"}))
		}
	} else {
		if _, _, err := r.store.GetSessionActiveTurn(ctx, sessionID); err == sql.ErrNoRows {
			if _, err := r.engine.continueQueuedSteeringLocked(ctx, r, sessionID); err != nil {
				log.Printf("steering continuation: %v", err)
				r.appendCleanupHandoffFailure(ctx, sessionID, claimToken, "continue_queued_steering", err)
			}
		} else if err != nil {
			log.Printf("turn coordination: inspect active turn after cleanup failed: %v", err)
		}
	}
	queueCount, err := r.store.CountQueuedTurns(ctx, sessionID)
	if err != nil {
		log.Printf("turn coordination: count queued turns after cleanup failed: %v", err)
		return
	}
	activeTurnID, _, err := r.store.GetSessionActiveTurn(ctx, sessionID)
	hasActiveTurn := err == nil
	if err != nil && err != sql.ErrNoRows {
		log.Printf("turn coordination: inspect active turn during cleanup normalization failed: %v", err)
		return
	}
	if hasActiveTurn {
		if err := r.engine.normalizeRunningSessionState(ctx, sessionID, activeTurnID, false, ""); err != nil {
			log.Printf("turn coordination: normalize running session state after cleanup failed: %v", err)
		}
		return
	}
	if queueCount > 0 {
		if err := r.engine.normalizeInactiveSessionState(ctx, sessionID, "queued", "", false); err != nil {
			log.Printf("turn coordination: normalize queued session state after cleanup failed: %v", err)
		}
		return
	}
	if err := r.engine.normalizeInactiveSessionState(ctx, sessionID, "idle", "", false); err != nil {
		log.Printf("turn coordination: normalize idle session state after cleanup failed: %v", err)
	}
}

func (r *sessionRunner) setupTurnRun(ctx context.Context, s *store.Store, sessionID, turnID string) (*preparedTurnRun, error) {
	turnRec, err := s.GetTurn(ctx, turnID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(sessionID) == "" {
		sessionID = turnRec.SessionID
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	initialSteering := steeringMessagesFromMetadata(turnRec.Metadata)
	prompt := turnRec.Prompt
	if strings.TrimSpace(prompt) == "" && len(initialSteering) > 0 {
		prompt = steeringPromptForShell(initialSteering)
	}
	intent := store.StringValue(turnRec.Metadata["intent"], "prompt")
	agentID, model := r.resolveTurnAgentAndModel(ctx, s, turnRec, sessionID, prompt)
	if hook := r.engine.beforeSetupHook; hook != nil {
		hook(ctx, sessionID, turnID)
	}
	if hook := r.engine.beforeSetupErrorHook; hook != nil {
		if err := hook(ctx, sessionID, turnID); err != nil {
			return nil, err
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	logutil.WarnIfErr("touch session state running", s.TouchSessionState(ctx, sessionID, map[string]any{"active_turn_id": turnID, "model": model, "status": "running"}))
	r.engine.PublishRuntimeSessionEvent("session_running", sessionID, agentID, "running", map[string]any{"reason": "setup", "active_turn_id": turnID, "turn_id": turnID, "turn_status": "running", "turn_phase": "setup", "model": model})
	r.emitSessionStateHookOnly(ctx, sessionID, agentID, model, "running", map[string]any{"reason": "setup", "active_turn_id": turnID, "turn_id": turnID, "turn_status": "running", "turn_phase": "setup"})
	r.engine.PublishRuntimeTurnEvent("turn_started", sessionID, turnID, agentID, "running", "setup", map[string]any{"reason": "setup", "model": model})
	r.emitTurnStateHookOnly(ctx, sessionID, turnID, agentID, model, "running", "setup", map[string]any{"reason": "setup"})
	userPayload := map[string]any{"kind": "chat", "intent": intent, "turn_id": turnID}
	for _, key := range []string{"source_session_id", "source_agent_id", "target_agent_id", "routed_from_prompt", "ingress_kind", "ingress_source_kind", "ingress_source_id", "ingress_role", "ingress_label", "ingress_session_key"} {
		if value, ok := turnRec.Metadata[key]; ok {
			userPayload[key] = value
		}
	}
	if strings.TrimSpace(prompt) != "" && len(initialSteering) == 0 {
		logutil.WarnIfErr("add user prompt message", s.AddMessage(ctx, store.NowID("msg"), sessionID, "user", prompt, userPayload))
	}
	startedPayload := map[string]any{"phase": "turn", "prompt": prompt, "intent": intent, "model": model, "checkpoint": true}
	for _, key := range []string{"source_session_id", "source_agent_id", "target_agent_id", "routed_from_prompt", "parent_turn_id", "route_mode", "route_matched_by", "ingress_kind", "ingress_source_kind", "ingress_source_id", "ingress_role", "ingress_label", "ingress_session_key"} {
		if value, ok := turnRec.Metadata[key]; ok {
			startedPayload[key] = value
		}
	}
	logutil.WarnIfErr("append turn.started event", s.AppendTurnEvent(ctx, turnID, sessionID, "turn.started", startedPayload))
	return &preparedTurnRun{
		turn:            turnRec,
		sessionID:       sessionID,
		turnID:          turnID,
		prompt:          prompt,
		intent:          intent,
		model:           model,
		agentID:         agentID,
		initialSteering: initialSteering,
	}, nil
}

func (r *sessionRunner) resolveTurnAgentAndModel(ctx context.Context, s *store.Store, turnRec *store.Turn, sessionID, prompt string) (string, string) {
	opCtx := store.CoordinationContext(ctx, r.engine.backgroundContext())
	model := store.StringValue(turnRec.Metadata["model"], "bootstrap")
	agentID := s.SessionAgentID(opCtx, sessionID)
	agentModel := r.engine.modelForAgent(agentID)
	if strings.TrimSpace(model) == "" {
		model = agentModel
	}
	if model == agentModel {
		history, _ := s.ListMessages(opCtx, sessionID)
		routingHistory := make([]routing.HistoryMessage, 0, len(history))
		for _, msg := range history {
			routingHistory = append(routingHistory, routing.HistoryMessage{Payload: msg.Payload})
		}
		selected, usedLight, score := r.engine.modelRouter.SelectModel(prompt, routingHistory, agentModel)
		if usedLight && strings.TrimSpace(selected) != "" {
			model = selected
			turnRec.Metadata["route_model_score"] = score
			turnRec.Metadata["route_used_light_model"] = usedLight
		}
	}
	return agentID, model
}

func (r *sessionRunner) runPreparedTurn(ctx context.Context, s *store.Store, run *preparedTurnRun) {
	if run.model != "bootstrap" && run.model != "test-model" && run.model != "" {
		r.runAgentLoop(ctx, s, run.turnID, run.sessionID, run.model, run.agentID, run.initialSteering)
		return
	}
	r.runShellTurn(ctx, s, run)
}

func (r *sessionRunner) runShellTurn(ctx context.Context, s *store.Store, run *preparedTurnRun) {
	if len(run.initialSteering) > 0 {
		r.persistSteeringMessages(ctx, run.sessionID, run.turnID, run.initialSteering)
	}
	r.engine.PublishRuntimeToolEvent("tool_started", run.sessionID, run.turnID, run.agentID, "shell", "", 0, nil, map[string]any{"phase": "tool", "command": []string{"sh", "-lc", "printf 'Gi received: %s' \"$GI_PROMPT\""}})
	logutil.WarnIfErr("append shell tool.started event", s.AppendTurnEvent(ctx, run.turnID, run.sessionID, "tool.started", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "command": []string{"sh", "-lc", "printf 'Gi received: %s' \"$GI_PROMPT\""}}))

	out, runErr, cancelled := tools.RunShellPrompt(ctx, run.prompt, func(cmd *exec.Cmd) {
		r.mu.Lock()
		if r.current != nil && r.current.turnID == run.turnID {
			r.current.cmdMu.Lock()
			r.current.cmd = cmd
			r.current.cmdMu.Unlock()
		}
		r.mu.Unlock()
	}, func(delta string) {
		if strings.TrimSpace(delta) == "" {
			return
		}
		r.engine.broadcast(run.sessionID, map[string]any{
			"type":     "agent_draft_delta",
			"chat_jid": "gi:" + run.sessionID,
			"delta":    delta,
			"turn_id":  run.turnID,
		})
	})
	if cancelled {
		bgCtx := r.engine.backgroundContext()
		r.appendFinalSteeringCheckpoint(s, run.turnID, run.sessionID)
		logutil.WarnIfErr("append turn.cancelled event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "turn.cancelled", map[string]any{"phase": "cancel", "checkpoint": true, "reason": "cancelled", "status": "cancelled", "turn_phase": "aborted", "failure_kind": ""}))
		logutil.WarnIfErr("append turn.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "turn.finished", map[string]any{"phase": "turn", "checkpoint": true, "status": "cancelled", "reason": "cancelled", "failure_kind": ""}))
		logutil.WarnIfErr("update turn status cancelled", s.UpdateTurnStatus(bgCtx, run.turnID, "cancelled"))
		r.engine.PublishRuntimeTurnEvent("turn_terminal", run.sessionID, run.turnID, run.agentID, "cancelled", "aborted", map[string]any{"reason": "cancelled", "failure_kind": ""})
		r.emitTurnStateHookOnly(bgCtx, run.sessionID, run.turnID, run.agentID, run.model, "cancelled", "aborted", map[string]any{"reason": "cancelled", "failure_kind": ""})
		r.propagateChildSubTurnCancellation(bgCtx, run.turnID, "cancelled", "")
		r.publishSubTurnLifecycle(bgCtx, run.turnID, "cancelled")
		msgID := store.NowID("msg")
		logutil.WarnIfErr("add turn cancelled system message", s.AddMessage(bgCtx, msgID, run.sessionID, "system", "Turn cancelled", map[string]any{"kind": "status", "turn_id": run.turnID, "clipped": true}))
		r.broadcastSystemPost(run.sessionID, run.turnID, msgID, "Turn cancelled")
		logutil.WarnIfErr("touch session idle after cancel", s.TouchSessionState(bgCtx, run.sessionID, map[string]any{"status": "idle", "active_turn_id": nil}))
		r.engine.PublishRuntimeSessionEvent("session_idle", run.sessionID, run.agentID, "idle", map[string]any{"reason": "turn_terminal", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "cancelled", "turn_phase": "aborted", "failure_kind": "", "model": run.model})
		r.emitSessionStateHookOnly(bgCtx, run.sessionID, run.agentID, run.model, "idle", map[string]any{"reason": "turn_terminal", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "cancelled", "turn_phase": "aborted", "failure_kind": ""})
		return
	}
	if runErr != nil {
		bgCtx := r.engine.backgroundContext()
		r.appendFinalSteeringCheckpoint(s, run.turnID, run.sessionID)
		logutil.WarnIfErr("turn failure mark", s.MarkTurnFailureWithFallbackErr(bgCtx, nil, run.turnID, run.sessionID, "shell_error", "none", runErr.Error()))
		r.engine.PublishRuntimeToolEvent("tool_failed", run.sessionID, run.turnID, run.agentID, "shell", "", 0, runErr, map[string]any{"phase": "tool"})
		logutil.WarnIfErr("append shell tool.failed event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "tool.failed", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "error": runErr.Error(), "failure_kind": "shell_error"}))
		msgID := store.NowID("msg")
		logutil.WarnIfErr("add shell failure system message", s.AddMessage(bgCtx, msgID, run.sessionID, "system", fmt.Sprintf("Shell tool failed: %v", runErr), map[string]any{"kind": "status", "turn_id": run.turnID, "source": "system", "failure_kind": "shell_error"}))
		r.broadcastSystemPost(run.sessionID, run.turnID, msgID, fmt.Sprintf("Shell tool failed: %v", runErr))
		logutil.WarnIfErr("append turn.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "turn.finished", map[string]any{"phase": "turn", "checkpoint": true, "status": "failed", "reason": "shell_error", "failure_kind": "shell_error"}))
		logutil.WarnIfErr("update turn status failed", s.UpdateTurnStatus(bgCtx, run.turnID, "failed"))
		r.engine.PublishRuntimeTurnEvent("turn_terminal", run.sessionID, run.turnID, run.agentID, "failed", "failed", map[string]any{"reason": "shell_error", "failure_kind": "shell_error"})
		r.emitTurnStateHookOnly(bgCtx, run.sessionID, run.turnID, run.agentID, run.model, "failed", "failed", map[string]any{"reason": "shell_error", "failure_kind": "shell_error"})
		r.propagateChildSubTurnCancellation(bgCtx, run.turnID, "failed", "shell_error")
		r.publishSubTurnLifecycle(bgCtx, run.turnID, "failed")
		logutil.WarnIfErr("touch session idle after failure", s.TouchSessionState(bgCtx, run.sessionID, map[string]any{"status": "idle", "active_turn_id": nil}))
		r.engine.PublishRuntimeSessionEvent("session_idle", run.sessionID, run.agentID, "idle", map[string]any{"reason": "turn_terminal", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "failed", "turn_phase": "failed", "failure_kind": "shell_error", "model": run.model})
		r.emitSessionStateHookOnly(bgCtx, run.sessionID, run.agentID, run.model, "idle", map[string]any{"reason": "turn_terminal", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "failed", "turn_phase": "failed", "failure_kind": "shell_error"})
		return
	}
	bgCtx := r.engine.backgroundContext()
	r.appendFinalSteeringCheckpoint(s, run.turnID, run.sessionID)
	r.engine.PublishRuntimeToolEvent("tool_finished", run.sessionID, run.turnID, run.agentID, "shell", "", 0, nil, map[string]any{"phase": "tool", "output_length": len(out)})
	logutil.WarnIfErr("append shell tool.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "tool.finished", map[string]any{"phase": "tool", "tool": "shell", "checkpoint": true, "output_length": len(out)}))
	msgID := store.NowID("msg")
	logutil.WarnIfErr("add shell assistant message", s.AddMessage(bgCtx, msgID, run.sessionID, "assistant", out, map[string]any{"kind": "chat", "source": "shell", "turn_id": run.turnID, "agent_id": run.agentID}))
	r.engine.broadcast(run.sessionID, map[string]any{
		"type": "new_post", "id": msgID, "chat_jid": "gi:" + run.sessionID,
		"content": out, "timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"sender": "agent", "is_bot_message": true,
		"data": map[string]any{"type": "agent_response", "content": out, "agent_id": run.agentID},
	})
	completionPayload := map[string]any{"reason": "completed", "completion_kind": "response"}
	logutil.WarnIfErr("append turn.finished event", s.AppendTurnEvent(bgCtx, run.turnID, run.sessionID, "turn.finished", map[string]any{"phase": "turn", "checkpoint": true, "status": "completed", "reason": "completed", "failure_kind": "", "completion_kind": "response"}))
	logutil.WarnIfErr("update turn status completed", s.UpdateTurnStatus(bgCtx, run.turnID, "completed"))
	r.engine.PublishRuntimeTurnEvent("turn_completed", run.sessionID, run.turnID, run.agentID, "completed", "completed", completionPayload)
	r.emitTurnStateHookOnly(bgCtx, run.sessionID, run.turnID, run.agentID, run.model, "completed", "completed", completionPayload)
	r.propagateChildSubTurnCancellation(bgCtx, run.turnID, "completed", "")
	r.publishSubTurnLifecycle(bgCtx, run.turnID, "completed")
	logutil.WarnIfErr("touch session idle after completion", s.TouchSessionState(bgCtx, run.sessionID, map[string]any{"status": "idle", "active_turn_id": nil}))
	sessionCompletionPayload := map[string]any{"reason": "turn_completed", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "completed", "turn_phase": "completed", "failure_kind": "", "model": run.model, "completion_kind": "response"}
	r.engine.PublishRuntimeSessionEvent("session_idle", run.sessionID, run.agentID, "idle", sessionCompletionPayload)
	r.emitSessionStateHookOnly(bgCtx, run.sessionID, run.agentID, run.model, "idle", map[string]any{"reason": "turn_completed", "active_turn_id": nil, "turn_id": run.turnID, "turn_status": "completed", "turn_phase": "completed", "failure_kind": "", "completion_kind": "response"})
}
