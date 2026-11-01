package turn

import (
	"context"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func (r *sessionRunner) settleSkippedTool(ctx context.Context, sessionID, turnID, agentID string, iter int, conv *goai.Context, call goai.ToolCall, reason, text, hookPhase string) error {
	occurrence := store.NowID("tool")
	text, diagnostics := boundToolOutput(text)
	event := map[string]any{"phase": "tool", "checkpoint": true, "tool": call.Name, "tool_call_id": call.ID, "occurrence_id": occurrence, "reason": reason, "hook_phase": hookPhase}
	payload := withOutputDiagnostics(map[string]any{"kind": "tool_result", "tool_call_id": call.ID, "tool_name": call.Name, "is_error": true, "turn_id": turnID, "occurrence_id": occurrence, "skipped": true, "skip_reason": reason, "hook_phase": hookPhase}, diagnostics)
	if err := r.store.SettleToolResult(ctx, turnID, sessionID, "tool.skipped", event, store.NowID("msg"), text, payload); err != nil {
		return err
	}
	wire := toolResultWireCall(ctx, turnID, call)
	goai.AppendToolResult(conv, wire.ID, wire.Name, text, true)
	r.engine.broadcast(sessionID, map[string]any{"type": "tool_skipped", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "reason": reason})
	r.engine.PublishRuntimeToolEvent("tool_skipped", sessionID, turnID, agentID, call.Name, call.ID, iter, nil, event)
	return nil
}
