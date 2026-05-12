package turn

import (
	"context"
	"strings"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
)

// Topics returns the engine-wide normalized topic bus.
func (e *Engine) Topics() *topics.Bus { return e.topics }

func (e *Engine) startTopicBridge() {
	if e == nil || e.topics == nil || e.connectivity == nil || e.connectivity.Bus() == nil {
		return
	}
	ch, _ := e.connectivity.Bus().Subscribe(context.Background(), "*", 64)
	go func() {
		for ev := range ch {
			topic := strings.TrimSpace(ev.Topic)
			if topic == "" {
				topic = "event"
			}
			if !strings.HasPrefix(topic, "connectivity.") {
				topic = "connectivity." + topic
			}
			e.topics.Publish(topics.Envelope{
				Topic:     topic,
				SessionID: ev.SessionID,
				AgentID:   ev.AgentID,
				Source:    firstNonEmpty(ev.Source, "connectivity"),
				Type:      "event",
				Payload: map[string]any{
					"id":        ev.ID,
					"route_id":  ev.RouteID,
					"transport": ev.Transport,
					"topic":     ev.Topic,
					"payload":   ev.Payload,
				},
				Timestamp: ev.Timestamp,
			})
		}
	}()
}

func (e *Engine) publishTopicEvent(env topics.Envelope) {
	if e == nil || e.topics == nil {
		return
	}
	e.topics.Publish(env)
}

func (e *Engine) PublishTopicEvent(env topics.Envelope) {
	e.publishTopicEvent(env)
}

func (e *Engine) publishTopicFromBroadcast(sessionID string, ev map[string]any) {
	if e == nil || e.topics == nil || ev == nil {
		return
	}
	evType, _ := ev["type"].(string)
	topic, envelopeType := topicForBroadcastEvent(evType)
	if topic == "" {
		return
	}
	agentID, _ := ev["agent_id"].(string)
	e.publishTopicEvent(topics.Envelope{
		Topic:     topic,
		SessionID: sessionID,
		AgentID:   agentID,
		Source:    "turn",
		Type:      envelopeType,
		Payload:   cloneMap(ev),
	})
}

func topicForBroadcastEvent(evType string) (topic string, envelopeType string) {
	switch strings.TrimSpace(evType) {
	case "agent_status":
		return "turn.status", "status"
	case "agent_draft_delta":
		return "turn.draft", "delta"
	case "agent_thought_delta":
		return "turn.thought", "delta"
	case "tool_finished":
		return "turn.tool.end", "result"
	case "tool_failed":
		return "turn.tool.end", "error"
	case "tool_skipped":
		return "turn.tool.end", "notice"
	case "steering_enqueued", "steering_dequeued", "steering_continue_staged", "steering_continued", "steering_injected", "steering_rejected":
		return "session.steering", "notice"
	case "subturn_created", "subturn_status", "subturn_result_ready", "subturn_result_delivered", "subturn_orphaned", "subturn_cancel_requested":
		return "turn.subturn", "notice"
	case "compaction":
		return "session.compaction", "notice"
	case "routing_decision", "routing_incoming":
		return "session.routing", "notice"
	case "new_post", "agent_response":
		return "turn.response", "result"
	default:
		if strings.TrimSpace(evType) == "" {
			return "", ""
		}
		return "event." + strings.ReplaceAll(evType, "_", "."), "notice"
	}
}

func (e *Engine) PublishRuntimeInboundWorkEvent(eventType string, item *store.InboundWorkItem, extra map[string]any) {
	if e == nil || e.topics == nil || item == nil {
		return
	}
	payload := map[string]any{
		"type":                 strings.TrimSpace(eventType),
		"id":                   item.ID,
		"status":               item.Status,
		"source_kind":          item.SourceKind,
		"session_id":           item.SessionID,
		"explicit_session_key": item.ExplicitSessionKey,
		"attempt_count":        item.AttemptCount,
		"last_error":           item.LastError,
		"next_attempt_at":      item.NextAttemptAt,
		"claimed_by":           item.ClaimedBy,
		"claimed_at":           item.ClaimedAt,
		"created_at":           item.CreatedAt,
		"updated_at":           item.UpdatedAt,
	}
	for k, v := range extra {
		payload[k] = v
	}
	e.publishTopicEvent(topics.Envelope{
		Topic:     "runtime.inbound_work",
		SessionID: item.SessionID,
		Source:    "runtime",
		Type:      "notice",
		Payload:   payload,
	})
}

func (e *Engine) PublishRuntimeDispatcherEvent(eventType string, payload map[string]any) {
	if e == nil || e.topics == nil {
		return
	}
	body := cloneMap(payload)
	if body == nil {
		body = map[string]any{}
	}
	body["type"] = strings.TrimSpace(eventType)
	e.publishTopicEvent(topics.Envelope{
		Topic:   "runtime.dispatcher",
		Source:  "runtime",
		Type:    "notice",
		Payload: body,
	})
}

func (e *Engine) PublishRuntimeHookEvent(eventType string, req HookRequest, source string, action string, durationMS int, err error) {
	if e == nil || e.topics == nil {
		return
	}
	payload := map[string]any{
		"type":           strings.TrimSpace(eventType),
		"hook":           req.Name,
		"session_id":     req.SessionID,
		"turn_id":        req.TurnID,
		"agent_id":       req.AgentID,
		"source":         strings.TrimSpace(source),
		"action":         strings.TrimSpace(action),
		"duration_ms":    durationMS,
		"iteration":      req.Iteration,
		"turn_status":    req.TurnStatus,
		"turn_phase":     req.TurnPhase,
		"session_status": req.SessionStatus,
		"trace": map[string]any{
			"id":         req.Trace.ID,
			"emitted_at": req.Trace.EmittedAt,
		},
	}
	if err != nil {
		payload["error"] = err.Error()
	}
	e.publishTopicEvent(topics.Envelope{
		Topic:     "runtime.hook",
		SessionID: req.SessionID,
		AgentID:   req.AgentID,
		Source:    "runtime",
		Type:      "notice",
		Payload:   payload,
	})
}

func (e *Engine) PublishRuntimeTurnEvent(eventType, sessionID, turnID, agentID, status, phase string, payload map[string]any) {
	if e == nil || e.topics == nil {
		return
	}
	body := cloneMap(payload)
	if body == nil {
		body = map[string]any{}
	}
	body["type"] = strings.TrimSpace(eventType)
	body["turn_id"] = strings.TrimSpace(turnID)
	body["status"] = strings.TrimSpace(status)
	body["phase"] = strings.TrimSpace(phase)
	e.publishTopicEvent(topics.Envelope{Topic: "runtime.turn", SessionID: strings.TrimSpace(sessionID), AgentID: strings.TrimSpace(agentID), Source: "runtime", Type: "notice", Payload: body})
}

func (e *Engine) PublishRuntimeSessionEvent(eventType, sessionID, agentID, status string, payload map[string]any) {
	if e == nil || e.topics == nil {
		return
	}
	body := cloneMap(payload)
	if body == nil {
		body = map[string]any{}
	}
	body["type"] = strings.TrimSpace(eventType)
	body["status"] = strings.TrimSpace(status)
	e.publishTopicEvent(topics.Envelope{Topic: "runtime.session", SessionID: strings.TrimSpace(sessionID), AgentID: strings.TrimSpace(agentID), Source: "runtime", Type: "notice", Payload: body})
}

func (e *Engine) PublishRuntimeToolEvent(eventType, sessionID, turnID, agentID, toolName, toolCallID string, iteration int, err error, payload map[string]any) {
	if e == nil || e.topics == nil {
		return
	}
	body := cloneMap(payload)
	if body == nil {
		body = map[string]any{}
	}
	body["type"] = strings.TrimSpace(eventType)
	body["turn_id"] = strings.TrimSpace(turnID)
	body["tool"] = strings.TrimSpace(toolName)
	body["tool_call_id"] = strings.TrimSpace(toolCallID)
	if iteration > 0 {
		body["iteration"] = iteration
	}
	if err != nil {
		body["error"] = err.Error()
	}
	envelopeType := "notice"
	if err != nil {
		envelopeType = "error"
	}
	e.publishTopicEvent(topics.Envelope{Topic: "runtime.tool", SessionID: strings.TrimSpace(sessionID), AgentID: strings.TrimSpace(agentID), Source: "runtime", Type: envelopeType, Payload: body})
}

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
