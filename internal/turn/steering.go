package turn

import (
	"context"
	"database/sql"
	"strings"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

const skippedDueToQueuedUserMessage = "Skipped due to queued user message."

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
			Role:      stringValue(m["role"], "user"),
			Content:   stringValue(m["content"], ""),
			Payload:   payload,
			Media:     media,
			QueueMode: stringValue(m["queue_mode"], "one-at-a-time"),
		})
	}
	return out
}

func (e *Engine) submitSteeringPrompt(ctx context.Context, sessionID, activeTurnID string, in RunInput) (*SubmitResult, error) {
	payload := map[string]any{"intent": in.Intent, "model": in.Model, "kind": "steering", "active_turn_id": activeTurnID}
	if in.ParentTurnID != "" {
		payload["parent_turn_id"] = in.ParentTurnID
	}
	for k, v := range in.Metadata {
		payload[k] = v
	}
	queueMode := stringValue(in.Metadata["steering_mode"], "one-at-a-time")
	if _, err := e.store.EnqueueSteering(ctx, sessionID, activeTurnID, "user", in.Prompt, payload, nil, queueMode); err != nil {
		return nil, err
	}
	_ = e.store.AppendTurnEvent(ctx, activeTurnID, sessionID, "steering.enqueued", map[string]any{
		"phase":      "steering",
		"checkpoint": true,
		"content":    in.Prompt,
		"queue_mode": queueMode,
	})
	return &SubmitResult{TurnID: activeTurnID, SessionID: sessionID, Status: "running", Queued: false}, nil
}

func (e *Engine) continueQueuedSteering(ctx context.Context, sessionID string) (bool, error) {
	msgs, err := e.store.DequeueSteering(ctx, sessionID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if len(msgs) == 0 {
		return false, nil
	}
	metadata := map[string]any{
		"initial_steering": steeringMessagesToMetadata(msgs),
		"continue":         true,
	}
	if len(msgs) > 0 {
		if msgs[0].Payload != nil {
			for _, key := range []string{"intent", "model", "parent_turn_id", "source_session_id", "source_agent_id", "target_agent_id", "route_mode", "route_matched_by"} {
				if value, ok := msgs[0].Payload[key]; ok {
					metadata[key] = value
				}
			}
		}
	}
	_, err = e.SubmitPrompt(ctx, RunInput{
		SessionID: sessionID,
		Prompt:    "",
		Intent:    stringValue(metadata["intent"], "continue"),
		Model:     stringValue(metadata["model"], ""),
		Metadata:  metadata,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *sessionRunner) dequeueSteeringMessages(ctx context.Context, sessionID string) ([]store.SteeringMessage, error) {
	msgs, err := r.store.DequeueSteering(ctx, sessionID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (r *sessionRunner) injectSteeringMessages(ctx context.Context, sessionID, turnID string, convCtx *goai.Context, msgs []store.SteeringMessage) int {
	if len(msgs) == 0 {
		return 0
	}
	totalContentLen := 0
	for _, msg := range msgs {
		role := strings.TrimSpace(strings.ToLower(msg.Role))
		if role == "" {
			role = "user"
		}
		switch role {
		case "assistant":
			convCtx.Messages = append(convCtx.Messages, goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: msg.Content}}})
		default:
			convCtx.Messages = append(convCtx.Messages, goai.UserMessage(msg.Content))
		}
		payload := map[string]any{"kind": "chat", "intent": stringValue(msg.Payload["intent"], "prompt"), "turn_id": turnID, "steering": true}
		for k, v := range msg.Payload {
			payload[k] = v
		}
		_ = r.store.AddMessage(ctx, store.NowID("msg"), sessionID, role, msg.Content, payload)
		totalContentLen += len(msg.Content)
	}
	_ = r.store.AppendTurnEvent(ctx, turnID, sessionID, "steering.injected", map[string]any{
		"phase":             "steering",
		"checkpoint":        true,
		"count":             len(msgs),
		"total_content_len": totalContentLen,
	})
	r.engine.broadcast(sessionID, map[string]any{"type": "steering_injected", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "count": len(msgs)})
	return len(msgs)
}

func (r *sessionRunner) skipRemainingToolCalls(ctx context.Context, sessionID, turnID string, convCtx *goai.Context, toolCalls []goai.ToolCall, start int) {
	for i := start; i < len(toolCalls); i++ {
		call := toolCalls[i]
		goai.AppendToolResult(convCtx, call.ID, call.Name, skippedDueToQueuedUserMessage, true)
		_ = r.store.AppendTurnEvent(ctx, turnID, sessionID, "tool.skipped", map[string]any{
			"phase":        "tool",
			"checkpoint":   true,
			"tool":         call.Name,
			"tool_call_id": call.ID,
			"reason":       "queued user steering message",
		})
		_ = r.store.AddMessage(ctx, store.NowID("msg"), sessionID, "tool_result", skippedDueToQueuedUserMessage, map[string]any{
			"kind":         "tool_result",
			"tool_call_id": call.ID,
			"tool_name":    call.Name,
			"is_error":     true,
			"turn_id":      turnID,
			"skipped":      true,
			"skip_reason":  "queued user steering message",
		})
		r.engine.broadcast(sessionID, map[string]any{"type": "tool_skipped", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tool": call.Name, "reason": "queued user steering message"})
	}
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
