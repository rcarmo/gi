package turn

import (
	"context"
	"database/sql"
	"strings"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

const skippedDueToQueuedUserMessage = "Skipped due to queued user message."

func normalizeSteeringRole(role string) string {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "assistant":
		return "assistant"
	case "system":
		return "system"
	default:
		return "user"
	}
}

func persistedSteeringChatRole(role string) string {
	if normalizeSteeringRole(role) == "assistant" {
		return "assistant"
	}
	return "user"
}

func steeringMessagesToMetadata(msgs []store.SteeringMessage) []map[string]any {
	out := make([]map[string]any, 0, len(msgs))
	for _, msg := range msgs {
		out = append(out, map[string]any{
			"role":       msg.Role,
			"content":    msg.Content,
			"payload":    msg.Payload,
			"media":      msg.Media,
			"queue_mode": msg.QueueMode,
		})
	}
	return out
}

func steeringMessagesFromMetadata(metadata map[string]any) []store.SteeringMessage {
	if metadata == nil {
		return nil
	}
	raw, ok := metadata["initial_steering"]
	if !ok || raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]store.SteeringMessage, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		payload, _ := m["payload"].(map[string]any)
		media := []string{}
		if rawMedia, ok := m["media"].([]any); ok {
			for _, v := range rawMedia {
				if s, ok := v.(string); ok && s != "" {
					media = append(media, s)
				}
			}
		}
		out = append(out, store.SteeringMessage{
			Role:      normalizeSteeringRole(stringValue(m["role"], "user")),
			Content:   stringValue(m["content"], ""),
			Payload:   payload,
			Media:     media,
			QueueMode: stringValue(m["queue_mode"], "one-at-a-time"),
		})
	}
	return out
}

func (e *Engine) submitSteeringPrompt(ctx context.Context, sessionID, activeTurnID string, in RunInput) (*SubmitResult, error) {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	if strings.TrimSpace(sessionID) != "" && strings.TrimSpace(activeTurnID) != "" {
		e.normalizeRunningSessionState(opCtx, sessionID, activeTurnID, true, "")
	}
	payload := map[string]any{"intent": in.Intent, "model": in.Model, "kind": "steering", "active_turn_id": activeTurnID}
	if in.ParentTurnID != "" {
		payload["parent_turn_id"] = in.ParentTurnID
	}
	for k, v := range in.Metadata {
		payload[k] = v
	}
	media := steeringMediaFromMetadata(in.Metadata)
	queueMode := stringValue(in.Metadata["steering_mode"], "one-at-a-time")
	steeringRole := normalizeSteeringRole(stringValue(in.Metadata["ingress_role"], "user"))
	if _, err := e.store.EnqueueSteering(opCtx, sessionID, activeTurnID, steeringRole, in.Prompt, payload, media, queueMode); err != nil {
		warnStore("append steering.rejected event", e.store.AppendTurnEvent(e.backgroundContext(), activeTurnID, sessionID, "steering.rejected", map[string]any{
			"phase":       "steering",
			"checkpoint":  true,
			"reason":      err.Error(),
			"queue_mode":  queueMode,
			"content":     in.Prompt,
			"media_count": len(media),
		}))
		e.broadcast(sessionID, map[string]any{
			"type":        "steering_rejected",
			"chat_jid":    "gi:" + sessionID,
			"turn_id":     activeTurnID,
			"queue_mode":  queueMode,
			"content_len": len(in.Prompt),
			"media_count": len(media),
			"reason":      err.Error(),
		})
		return nil, err
	}
	warnStore("append steering.enqueued event", e.store.AppendTurnEvent(opCtx, activeTurnID, sessionID, "steering.enqueued", map[string]any{
		"phase":       "steering",
		"checkpoint":  true,
		"content":     in.Prompt,
		"queue_mode":  queueMode,
		"media_count": len(media),
	}))
	e.broadcast(sessionID, map[string]any{
		"type":        "steering_enqueued",
		"chat_jid":    "gi:" + sessionID,
		"turn_id":     activeTurnID,
		"queue_mode":  queueMode,
		"content_len": len(in.Prompt),
		"media_count": len(media),
	})
	return &SubmitResult{TurnID: activeTurnID, SessionID: sessionID, Status: "running", Queued: false}, nil
}

func steeringMetadataFromMessages(msgs []store.SteeringMessage) map[string]any {
	metadata := map[string]any{
		"initial_steering": steeringMessagesToMetadata(msgs),
		"continue":         true,
	}
	if len(msgs) > 0 && msgs[0].Payload != nil {
		for _, key := range []string{"intent", "model", "parent_turn_id", "source_session_id", "source_agent_id", "target_agent_id", "route_mode", "route_matched_by"} {
			if value, ok := msgs[0].Payload[key]; ok {
				metadata[key] = value
			}
		}
	}
	return metadata
}

func (e *Engine) stageQueuedSteeringContinuation(ctx context.Context, sessionID string) (bool, string, error) {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	turnID := store.NowID("turn")
	turnRec, msgs, err := e.store.StageSteeringContinuation(opCtx, sessionID, turnID)
	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	metadata := turnRec.Metadata
	submittedPayload := map[string]any{"phase": "queue", "intent": stringValue(metadata["intent"], "continue"), "queued": true, "checkpoint": true, "continue": true}
	warnStore("append continued turn.submitted event", e.store.AppendTurnEvent(opCtx, turnID, sessionID, "turn.submitted", submittedPayload))
	e.PublishRuntimeTurnEvent("turn_submitted", sessionID, turnID, "", "queued", "queued", submittedPayload)
	if err := e.normalizeInactiveSessionState(opCtx, sessionID, "queued", stringValue(metadata["model"], ""), true); err != nil {
		return false, "", err
	}
	warnStore("append steering.continue_staged event", e.store.AppendTurnEvent(opCtx, turnID, sessionID, "steering.continue_staged", map[string]any{"phase": "steering", "checkpoint": true, "count": len(msgs)}))
	e.broadcast(sessionID, map[string]any{"type": "steering_continue_staged", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "count": len(msgs)})
	return true, turnID, nil
}

func (e *Engine) continueQueuedSteeringLocked(ctx context.Context, runner *sessionRunner, sessionID string) (bool, error) {
	staged, turnID, err := e.stageQueuedSteeringContinuation(ctx, sessionID)
	if err != nil {
		return false, err
	}
	if !staged {
		return false, nil
	}
	coordCtx := coordinationContext(ctx, e.backgroundContext())
	launched, err := e.startNextQueuedTurnLocked(coordCtx, runner, sessionID)
	if err != nil {
		return false, err
	}
	if launched {
		if err := e.store.SyncSessionQueueCount(coordCtx, sessionID); err != nil {
			return false, err
		}
		warnStore("append steering.continued event", e.store.AppendTurnEvent(coordCtx, turnID, sessionID, "steering.continued", map[string]any{"phase": "steering", "checkpoint": true, "handoff": "launched"}))
		e.broadcast(sessionID, map[string]any{"type": "steering_continued", "chat_jid": "gi:" + sessionID, "turn_id": turnID})
		return true, nil
	}
	activeTurnID, _, err := e.store.GetSessionActiveTurn(coordCtx, sessionID)
	if err == nil {
		if err := e.normalizeRunningSessionState(coordCtx, sessionID, activeTurnID, true, ""); err != nil {
			return false, err
		}
		warnStore("append steering.continued event", e.store.AppendTurnEvent(coordCtx, turnID, sessionID, "steering.continued", map[string]any{"phase": "steering", "checkpoint": true, "handoff": "active_claim"}))
		return true, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return true, nil
}

func (e *Engine) continueQueuedSteering(ctx context.Context, sessionID string) (bool, error) {
	runner := e.runner(sessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	return e.continueQueuedSteeringLocked(ctx, runner, sessionID)
}

func (e *Engine) ContinueSession(ctx context.Context, sessionID string) (bool, error) {
	runner := e.runner(sessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	coordCtx := coordinationContext(ctx, e.backgroundContext())
	if activeTurnID, _, err := e.store.GetSessionActiveTurn(coordCtx, sessionID); err == nil {
		if err := e.normalizeRunningSessionState(coordCtx, sessionID, activeTurnID, true, ""); err != nil {
			return false, err
		}
		return false, nil
	} else if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	launched, err := e.startNextQueuedTurnLocked(coordCtx, runner, sessionID)
	if err != nil {
		return false, err
	}
	if launched {
		return true, nil
	}
	if activeTurnID, _, err := e.store.GetSessionActiveTurn(coordCtx, sessionID); err == nil {
		if err := e.normalizeRunningSessionState(coordCtx, sessionID, activeTurnID, true, ""); err != nil {
			return false, err
		}
		return true, nil
	} else if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	continued, err := e.continueQueuedSteeringLocked(coordCtx, runner, sessionID)
	if err != nil {
		return false, err
	}
	if continued {
		return true, nil
	}
	if err := e.normalizeInactiveSessionState(coordCtx, sessionID, "idle", "", true); err != nil {
		return false, err
	}
	return false, nil
}

func (r *sessionRunner) dequeueSteeringMessages(ctx context.Context, sessionID string) ([]store.SteeringMessage, error) {
	msgs, err := r.store.DequeueSteering(ctx, sessionID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(msgs) > 0 {
		r.engine.broadcast(sessionID, map[string]any{"type": "steering_dequeued", "chat_jid": "gi:" + sessionID, "count": len(msgs)})
	}
	return msgs, nil
}

func (r *sessionRunner) persistSteeringMessages(ctx context.Context, sessionID, turnID string, msgs []store.SteeringMessage) int {
	if len(msgs) == 0 {
		return 0
	}
	totalContentLen := 0
	for _, msg := range msgs {
		role := persistedSteeringChatRole(msg.Role)
		payload := map[string]any{"kind": "chat", "intent": stringValue(msg.Payload["intent"], "prompt"), "turn_id": turnID, "steering": true, "steering_role": normalizeSteeringRole(msg.Role)}
		for k, v := range msg.Payload {
			payload[k] = v
		}
		if len(msg.Media) > 0 {
			payload["media"] = append([]string(nil), msg.Media...)
		}
		warnStore("add steering message", r.store.AddMessage(ctx, store.NowID("msg"), sessionID, role, msg.Content, payload))
		totalContentLen += len(msg.Content)
	}
	warnStore("append steering.injected event", r.store.AppendTurnEvent(ctx, turnID, sessionID, "steering.injected", map[string]any{
		"phase":             "steering",
		"checkpoint":        true,
		"count":             len(msgs),
		"total_content_len": totalContentLen,
	}))
	r.engine.broadcast(sessionID, map[string]any{"type": "steering_injected", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "count": len(msgs), "media_count": steeringMediaCount(msgs)})
	return len(msgs)
}

func (r *sessionRunner) injectSteeringMessages(ctx context.Context, sessionID, turnID string, convCtx *goai.Context, msgs []store.SteeringMessage) int {
	if len(msgs) == 0 {
		return 0
	}
	for _, msg := range msgs {
		role := normalizeSteeringRole(msg.Role)
		content := msg.Content
		if len(msg.Media) > 0 {
			if strings.TrimSpace(content) == "" {
				content = "[user provided media attachments]"
			} else {
				content += "\n\n[media attachments included]"
			}
		}
		switch role {
		case "assistant":
			convCtx.Messages = append(convCtx.Messages, goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: content}}})
		default:
			convCtx.Messages = append(convCtx.Messages, goai.UserMessage(content))
		}
	}
	return r.persistSteeringMessages(ctx, sessionID, turnID, msgs)
}

func (r *sessionRunner) skipRemainingToolCalls(ctx context.Context, sessionID, turnID string, convCtx *goai.Context, toolCalls []goai.ToolCall, start int) {
	for i := start; i < len(toolCalls); i++ {
		call := toolCalls[i]
		goai.AppendToolResult(convCtx, call.ID, call.Name, skippedDueToQueuedUserMessage, true)
		warnStore("append tool.skipped event", r.store.AppendTurnEvent(ctx, turnID, sessionID, "tool.skipped", map[string]any{
			"phase":        "tool",
			"checkpoint":   true,
			"tool":         call.Name,
			"tool_call_id": call.ID,
			"reason":       "queued user steering message",
		}))
		warnStore("add skipped tool_result message", r.store.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", skippedDueToQueuedUserMessage, map[string]any{
			"kind":         "tool_result",
			"tool_call_id": call.ID,
			"tool_name":    call.Name,
			"is_error":     true,
			"turn_id":      turnID,
			"skipped":      true,
			"skip_reason":  "queued user steering message",
		}))
		r.engine.broadcast(sessionID, map[string]any{"type": "tool_skipped", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "reason": "queued user steering message"})
		r.engine.PublishRuntimeToolEvent("tool_skipped", sessionID, turnID, "", call.Name, call.ID, 0, nil, map[string]any{"reason": "queued user steering message", "phase": "tool"})
	}
}

func steeringMediaFromMetadata(metadata map[string]any) []string {
	if metadata == nil {
		return nil
	}
	raw, ok := metadata["media"]
	if !ok || raw == nil {
		return nil
	}
	switch m := raw.(type) {
	case []string:
		out := make([]string, 0, len(m))
		for _, item := range m {
			if strings.TrimSpace(item) != "" {
				out = append(out, item)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(m))
		for _, item := range m {
			s, ok := item.(string)
			if !ok || strings.TrimSpace(s) == "" {
				continue
			}
			out = append(out, s)
		}
		return out
	default:
		return nil
	}
}

func steeringMediaCount(msgs []store.SteeringMessage) int {
	total := 0
	for _, msg := range msgs {
		total += len(msg.Media)
	}
	return total
}

func steeringPromptForShell(msgs []store.SteeringMessage) string {
	parts := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		if strings.TrimSpace(msg.Content) != "" {
			parts = append(parts, msg.Content)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}
