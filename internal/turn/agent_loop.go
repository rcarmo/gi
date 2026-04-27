package turn

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

// toolDefs returns the go-ai Tool definitions sent to the LLM.
func toolDefs() []goai.Tool {
	return []goai.Tool{
		{
			Name:        "tools",
			Description: "List available tools or get details about a specific tool. Use with no arguments to list all tools (names + short descriptions). Pass a tool name via the `name` argument to get its full schema and usage. Use `query` to filter tools by keyword.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"name":{"type":"string","description":"Exact tool name to get full details for"},"query":{"type":"string","description":"Filter tools by keyword in name or description"}}}`),
		},
		{
			Name:        "read",
			Description: "Read text content from a workspace file. Supports workspace-relative paths and vfs:// paths.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"Workspace-relative path or vfs://namespace/path"}},"required":["path"]}`),
		},
		{
			Name:        "write",
			Description: "Write text content to a workspace file. Creates parent directories for workspace paths.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"Workspace-relative path or vfs://namespace/path"},"content":{"type":"string","description":"File content to write"}},"required":["path","content"]}`),
		},
		{
			Name:        "shell",
			Description: "Execute a shell command and return stdout/stderr. Use for running tests, installing packages, searching files, etc.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"command":{"type":"string","description":"Shell command to execute"}},"required":["command"]}`),
		},
	}
}

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
func (r *sessionRunner) runAgentLoop(ctx context.Context, s *store.Store, turnID, sessionID, model, agentID string) {
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
		Tools:        toolDefs(),
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

	r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "title": "Thinking…", "status": "running", "turn_id": turnID})

	var totalUsage goai.Usage
	lastToolFailureSig := ""
	repeatedToolFailureCount := 0

	for iter := 0; iter < maxIter; iter++ {
		if ctx.Err() != nil {
			r.finishTurn(s, turnID, sessionID, agentID, "cancelled", "Turn cancelled")
			return
		}

		iterLabel := fmt.Sprintf("iter=%d/%d", iter+1, maxIter)
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

		if inferErr != nil {
			log.Printf("inference [%s] error: %v", iterLabel, inferErr)
			_ = s.AppendTurnEvent(ctx, turnID, sessionID, "inference.failed", map[string]any{"phase": "inference", "checkpoint": true, "error": inferErr.Error(), "iteration": iter + 1})
			r.finishTurn(s, turnID, sessionID, agentID, "failed", fmt.Sprintf("Inference error: %v", inferErr))
			return
		}
		if result == nil || result.Message == nil {
			log.Printf("inference [%s]: nil result", iterLabel)
			r.finishTurn(s, turnID, sessionID, agentID, "failed", "Inference returned no result")
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
			// Final response — no tool calls. Persist and finish.
			log.Printf("inference [%s]: final response (%d chars, %d iterations)", iterLabel, len(textContent), iter+1)
			r.persistUsage(s, turnID, sessionID, &totalUsage, iter+1)

			msgID := store.NowID("msg")
			_ = s.AddMessage(ctx, msgID, sessionID, "assistant", textContent, map[string]any{
				"kind": "chat", "source": "inference", "model": model,
				"turn_id": turnID, "agent_id": agentID, "iterations": iter + 1,
			})

			r.broadcastPost(sessionID, turnID, msgID, textContent, agentID)
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

		for _, call := range toolCalls {
			if ctx.Err() != nil {
				r.finishTurn(s, turnID, sessionID, agentID, "cancelled", "Turn cancelled during tool execution")
				return
			}

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
					r.finishTurn(s, turnID, sessionID, agentID, "failed", msg)
					return
				}
			} else {
				// Truncate very large results to avoid blowing context
				displayResult := toolResult
				if len(displayResult) > 100000 {
					displayResult = displayResult[:100000] + "\n... (truncated)"
				}
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
		}
		// Loop continues — next iteration will call LLM with tool results
	}

	// Budget exhausted
	log.Printf("inference: max iterations (%d) reached for turn %s", maxIter, turnID)
	r.persistUsage(s, turnID, sessionID, &totalUsage, maxIter)
	r.finishTurn(s, turnID, sessionID, agentID, "completed", fmt.Sprintf("Reached maximum iteration limit (%d). The task may be incomplete.", maxIter))
}

// executeTool dispatches a single tool call and returns the text result.
func (r *sessionRunner) executeTool(ctx context.Context, call goai.ToolCall, sessionID string) (string, error) {
	workspaceRoot := r.engine.runtimeCfg.WorkspaceRoot

	switch call.Name {
	case "tools":
		return executeToolsTool(call.Arguments)

	case "read":
		path, _ := call.Arguments["path"].(string)
		if path == "" {
			return "", fmt.Errorf("read: path is required")
		}
		resolved, err := tools.ResolveToolPath(workspaceRoot, path, false)
		if err != nil {
			return "", err
		}
		if resolved.IsVFS() {
			_, raw, err := r.store.GetVFSFileContent(ctx, resolved.VFSNamespace, resolved.VFSPath)
			if err != nil {
				return "", err
			}
			return string(raw), nil
		}
		content, err := os.ReadFile(resolved.WorkspacePath)
		if err != nil {
			return "", err
		}
		return string(content), nil

	case "write":
		path, _ := call.Arguments["path"].(string)
		content, _ := call.Arguments["content"].(string)
		if path == "" {
			return "", fmt.Errorf("write: path is required")
		}
		resolved, err := tools.ResolveToolPath(workspaceRoot, path, true)
		if err != nil {
			return "", err
		}
		if resolved.IsVFS() {
			_, err := r.store.SaveVFSFile(ctx, resolved.VFSNamespace, resolved.VFSPath, "text/plain", []byte(content), map[string]any{})
			if err != nil {
				return "", err
			}
			return "written", nil
		}
		dir := filepath.Dir(resolved.WorkspacePath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(resolved.WorkspacePath, []byte(content), 0o644); err != nil {
			return "", err
		}
		return "written", nil

	case "shell":
		command, _ := call.Arguments["command"].(string)
		if command == "" {
			return "", fmt.Errorf("shell: command is required")
		}
		cmd := exec.CommandContext(ctx, "sh", "-lc", command)
		cmd.Dir = workspaceRoot
		out, err := cmd.CombinedOutput()
		output := string(out)
		if err != nil {
			return output, fmt.Errorf("exit: %w", err)
		}
		return output, nil

	default:
		return "", fmt.Errorf("unknown tool: %s", call.Name)
	}
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
	_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "turn.finished", map[string]any{
		"phase": "turn", "checkpoint": true, "status": "completed", "iterations": iterations,
	})
	_ = s.UpdateTurnStatus(context.Background(), turnID, "completed")
	_ = s.TouchSessionState(context.Background(), sessionID, map[string]any{"status": "idle", "active_turn_id": nil})
	r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "title": "", "status": "idle"})
}

// finishTurn persists a terminal status and optional system message.
func (r *sessionRunner) finishTurn(s *store.Store, turnID, sessionID, agentID, status, systemMsg string) {
	if systemMsg != "" {
		msgID := store.NowID("msg")
		_ = s.AddMessage(context.Background(), msgID, sessionID, "assistant", systemMsg, map[string]any{
			"kind": "chat", "source": "system", "turn_id": turnID, "agent_id": agentID,
		})
		if status == "completed" || status == "failed" {
			r.broadcastPost(sessionID, turnID, msgID, systemMsg, agentID)
		}
	}
	_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "turn.finished", map[string]any{
		"phase": "turn", "checkpoint": true, "status": status,
	})
	_ = s.UpdateTurnStatus(context.Background(), turnID, status)
	_ = s.TouchSessionState(context.Background(), sessionID, map[string]any{"status": "idle", "active_turn_id": nil})
	r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "title": "", "status": "idle"})
}

// --- tools tool implementation ---

// toolEntry is a compact representation of a tool for the registry.
type toolEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters,omitempty"`
}

// executeToolsTool handles the "tools" meta-tool: list, search, and inspect.
func executeToolsTool(args map[string]any) (string, error) {
	name, _ := args["name"].(string)
	query, _ := args["query"].(string)

	allTools := toolDefs()

	// Detail mode: return full schema for a specific tool
	if name != "" {
		for _, t := range allTools {
			if t.Name == name {
				var params any
				_ = json.Unmarshal(t.Parameters, &params)
				entry := toolEntry{Name: t.Name, Description: t.Description, Parameters: params}
				b, _ := json.MarshalIndent(entry, "", "  ")
				return string(b), nil
			}
		}
		return "", fmt.Errorf("tool not found: %s", name)
	}

	// List/search mode: return compact summaries
	var entries []toolEntry
	for _, t := range allTools {
		if query != "" {
			lq := strings.ToLower(query)
			if !strings.Contains(strings.ToLower(t.Name), lq) && !strings.Contains(strings.ToLower(t.Description), lq) {
				continue
			}
		}
		entries = append(entries, toolEntry{Name: t.Name, Description: t.Description})
	}

	if len(entries) == 0 {
		return "No tools matched the query.", nil
	}

	// Format as compact text list
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d tool(s):\n", len(entries))
	for _, e := range entries {
		fmt.Fprintf(&sb, "- %s: %s\n", e.Name, e.Description)
	}
	sb.WriteString("\nUse tools({name: \"<tool>\"}) for full parameter schema.")
	return sb.String(), nil
}
