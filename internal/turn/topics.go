package turn

import (
	"strings"

	"github.com/rcarmo/gi/internal/routing/audit"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
)

// Topics returns the engine-wide normalized topic bus.
func (e *Engine) Topics() *topics.Bus { return e.topics }

func (e *Engine) startTopicBridge() {
	if e == nil || e.topics == nil || e.connectivity == nil || e.connectivity.Bus() == nil || e.bgCtx == nil {
		return
	}
	ch, _ := e.connectivity.Bus().Subscribe(e.bgCtx, "*", 64)
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
	payload := cloneMap(extra)
	if payload == nil {
		payload = map[string]any{}
	}
	payload["id"] = item.ID
	payload["status"] = item.Status
	payload["source_kind"] = item.SourceKind
	payload["session_id"] = item.SessionID
	payload["explicit_session_key"] = item.ExplicitSessionKey
	payload["attempt_count"] = item.AttemptCount
	payload["last_error"] = item.LastError
	payload["next_attempt_at"] = item.NextAttemptAt
	payload["claimed_by"] = item.ClaimedBy
	payload["claimed_at"] = item.ClaimedAt
	payload["created_at"] = item.CreatedAt
	payload["updated_at"] = item.UpdatedAt
	e.publishRuntimeTopicEvent("runtime.inbound_work", item.SessionID, "", "notice", eventType, payload)
}

func (e *Engine) PublishRuntimeDispatcherEvent(eventType string, payload map[string]any) {
	e.publishRuntimeTopicEvent("runtime.dispatcher", "", "", "notice", eventType, payload)
}

func (e *Engine) PublishRuntimeHookEvent(eventType string, req HookRequest, source string, action string, durationMS int, err error) {
	payload := runtimeHookPayload(req, nil)
	payload["source"] = strings.TrimSpace(source)
	payload["action"] = strings.TrimSpace(action)
	payload["duration_ms"] = durationMS
	payload["trace"] = map[string]any{
		"id":         req.Trace.ID,
		"emitted_at": req.Trace.EmittedAt,
	}
	if err != nil {
		payload["error"] = err.Error()
	}
	e.publishRuntimeTopicEvent("runtime.hook", req.SessionID, req.AgentID, "notice", eventType, payload)
}

func (e *Engine) PublishRuntimeHookDecisionEvent(eventType string, req HookRequest, payload map[string]any) {
	e.publishRuntimeTopicEvent("runtime.hook", req.SessionID, req.AgentID, "notice", eventType, runtimeHookPayload(req, payload))
}

func (e *Engine) PublishRuntimeTurnEvent(eventType, sessionID, turnID, agentID, status, phase string, payload map[string]any) {
	body := cloneMap(payload)
	if body == nil {
		body = map[string]any{}
	}
	body["turn_id"] = strings.TrimSpace(turnID)
	body["status"] = strings.TrimSpace(status)
	body["phase"] = strings.TrimSpace(phase)
	e.publishRuntimeTopicEvent("runtime.turn", sessionID, agentID, "notice", eventType, body)
}

func (e *Engine) PublishRuntimeSessionEvent(eventType, sessionID, agentID, status string, payload map[string]any) {
	body := cloneMap(payload)
	if body == nil {
		body = map[string]any{}
	}
	body["status"] = strings.TrimSpace(status)
	e.publishRuntimeTopicEvent("runtime.session", sessionID, agentID, "notice", eventType, body)
}

func (e *Engine) PublishRuntimeToolEvent(eventType, sessionID, turnID, agentID, toolName, toolCallID string, iteration int, err error, payload map[string]any) {
	body := cloneMap(payload)
	if body == nil {
		body = map[string]any{}
	}
	body["turn_id"] = strings.TrimSpace(turnID)
	body["tool"] = strings.TrimSpace(toolName)
	body["tool_call_id"] = strings.TrimSpace(toolCallID)
	if iteration > 0 {
		body["iteration"] = iteration
	}
	envelopeType := "notice"
	if err != nil {
		body["error"] = err.Error()
		envelopeType = "error"
	}
	e.publishRuntimeTopicEvent("runtime.tool", sessionID, agentID, envelopeType, eventType, body)
}

func (e *Engine) PublishRuntimeRoutingEvent(eventType string, decision audit.Event) {
	body := cloneMap(decision.Metadata)
	if body == nil {
		body = map[string]any{}
	}
	body["route_event_id"] = decision.ID
	body["turn_id"] = strings.TrimSpace(decision.TurnID)
	body["source_session_id"] = strings.TrimSpace(decision.SourceSession)
	body["target_session_id"] = strings.TrimSpace(decision.TargetSession)
	body["source_agent_id"] = strings.TrimSpace(decision.SourceAgentID)
	body["target_agent_id"] = strings.TrimSpace(decision.TargetAgentID)
	body["mode"] = strings.TrimSpace(decision.Mode)
	body["matched_by"] = strings.TrimSpace(decision.MatchedBy)
	body["routing_policy"] = strings.TrimSpace(decision.RoutingPolicy)
	body["requested_agent_id"] = strings.TrimSpace(decision.RequestedAgent)
	body["created_at"] = strings.TrimSpace(decision.CreatedAt)
	sessionID := strings.TrimSpace(decision.SourceSession)
	if strings.TrimSpace(eventType) == "routing_incoming" && strings.TrimSpace(decision.TargetSession) != "" {
		sessionID = strings.TrimSpace(decision.TargetSession)
	}
	e.publishRuntimeTopicEvent("runtime.routing", sessionID, strings.TrimSpace(decision.TargetAgentID), "notice", eventType, body)
}

func runtimeHookPayload(req HookRequest, payload map[string]any) map[string]any {
	body := cloneMap(payload)
	if body == nil {
		body = map[string]any{}
	}
	body["hook"] = req.Name
	body["session_id"] = req.SessionID
	body["turn_id"] = req.TurnID
	body["agent_id"] = req.AgentID
	body["iteration"] = req.Iteration
	body["turn_status"] = req.TurnStatus
	body["turn_phase"] = req.TurnPhase
	body["session_status"] = req.SessionStatus
	if req.ToolCall != nil {
		body["tool"] = req.ToolCall.Name
		body["tool_call_id"] = req.ToolCall.ID
	}
	return body
}

func (e *Engine) publishRuntimeTopicEvent(topic, sessionID, agentID, envelopeType, eventType string, payload map[string]any) {
	if e == nil || e.topics == nil {
		return
	}
	body := cloneMap(payload)
	if body == nil {
		body = map[string]any{}
	}
	body["type"] = strings.TrimSpace(eventType)
	e.publishTopicEvent(topics.Envelope{
		Topic:     strings.TrimSpace(topic),
		SessionID: strings.TrimSpace(sessionID),
		AgentID:   strings.TrimSpace(agentID),
		Source:    "runtime",
		Type:      strings.TrimSpace(firstNonEmpty(envelopeType, "notice")),
		Payload:   body,
	})
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
