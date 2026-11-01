package turn

import (
	"context"
	"errors"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

// Parser failures are terminal results without execution, approval or tool-call
// hooks. Their fresh occurrence must not terminate a previously reused call ID.
func (r *sessionRunner) rejectInvalidToolArguments(ctx context.Context, s *store.Store, sessionID, turnID, agentID string, iter int, conv *goai.Context, call goai.ToolCall) error {
	text := "Invalid tool arguments: " + call.ArgumentsError
	occurrence := store.NowID("tool")
	payload := map[string]any{"kind": "tool_result", "tool_name": call.Name, "tool_call_id": call.ID, "turn_id": turnID, "occurrence_id": occurrence, "is_error": true, "reason": "invalid_arguments"}
	event := map[string]any{"phase": "tool", "tool": call.Name, "tool_call_id": call.ID, "occurrence_id": occurrence, "reason": "invalid_arguments", "error": text}
	if err := s.SettleToolResult(ctx, turnID, sessionID, "tool.failed", event, store.NowID("msg"), text, payload); err != nil {
		return err
	}
	wire := toolResultWireCall(ctx, turnID, call)
	goai.AppendToolResult(conv, wire.ID, wire.Name, text, true)
	r.engine.PublishRuntimeToolEvent("tool_failed", sessionID, turnID, agentID, call.Name, call.ID, iter, errors.New(text), event)
	return nil
}
