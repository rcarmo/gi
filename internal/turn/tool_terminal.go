package turn

import (
	"github.com/rcarmo/gi/internal/logutil"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

// A turn ending is not evidence that a particular tool completed. Persist the
// execution boundary itself, even when the provider/tool context is cancelled.
// Failure leaves reconstruction unknown rather than inventing a terminal time.
func (r *sessionRunner) persistStoppedTool(s *store.Store, sessionID, turnID string, call goai.ToolCall, occurrenceID, state string, durationMS int64, extras ...*toolExtras) {
	payload := map[string]any{"usage": nil, "usage_unknown": true}
	if len(extras) > 0 && extras[0] != nil {
		extras[0].mu.Lock()
		payload["usage"] = extras[0].usage
		extras[0].mu.Unlock()
	}
	err := s.SettleToolResult(r.engine.backgroundContext(), turnID, sessionID, "tool."+state, map[string]any{
		"phase": "tool", "tool": call.Name, "checkpoint": true, "tool_call_id": call.ID, "occurrence_id": occurrenceID, "duration_ms": durationMS,
	}, "", "", payload)
	if err != nil {
		logutil.WarnIfErr("settle stopped tool event", err)
		return
	}
	r.engine.PublishRuntimeToolEvent("tool_"+state, sessionID, turnID, "", call.Name, call.ID, 0, nil, map[string]any{"phase": "tool", "occurrence_id": occurrenceID, "duration_ms": durationMS})
	r.engine.broadcast(sessionID, map[string]any{"type": "tool_activity_changed", "chat_jid": "gi:" + sessionID, "turn_id": turnID})
}
