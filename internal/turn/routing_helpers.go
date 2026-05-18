package turn

import (
	"context"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

func parseDirectedPrompt(prompt string) (string, string, bool) {
	trimmed := strings.TrimSpace(prompt)
	if !strings.HasPrefix(trimmed, "@") {
		return "", prompt, false
	}
	rest := trimmed[1:]
	end := strings.IndexAny(rest, " \t\r\n:")
	if end < 0 {
		return normalizeAgentID(rest), "", true
	}
	target := normalizeAgentID(rest[:end])
	body := strings.TrimSpace(rest[end:])
	body = strings.TrimPrefix(body, ":")
	body = strings.TrimSpace(body)
	return target, body, target != ""
}

func (e *Engine) recordRouteDecision(ctx context.Context, sourceSessionID, turnID string, metadata map[string]any) error {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	sourceSession := stringValue(metadata["source_session_id"], sourceSessionID)
	targetSession := stringValue(metadata["target_session_id"], "")
	targetAgentID := stringValue(metadata["target_agent_id"], "")
	if targetAgentID == "" {
		return nil
	}
	routeMode := stringValue(metadata["route_mode"], stringValue(metadata["mode"], "prompt"))
	sourceAgent := stringValue(metadata["source_agent_id"], "")
	if sourceAgent == "" {
		sourceAgent = sessionAgentIDForSessionID(opCtx, e.store, sourceSession)
	}
	routingPolicy := stringValue(metadata["routing_policy"], "")
	matchedBy := stringValue(metadata["route_matched_by"], "")
	requestedAgent := stringValue(metadata["requested_agent_id"], "")
	if requestedAgent == "" {
		requestedAgent = targetAgentID
	}
	if targetSession == "" {
		targetSession = sourceSession
	}
	decision := routeDecisionFromMetadata(sourceSession, targetSession, sourceAgent, targetAgentID, turnID, routeMode, matchedBy, routingPolicy, requestedAgent, metadata)
	if !boolValueOr(metadata["routing_enabled"], true) {
		return nil
	}
	routeEventID, err := e.store.RecordRouteEvent(opCtx, decision)
	if err != nil {
		return err
	}
	decision.ID = routeEventID
	e.PublishRuntimeRoutingEvent("routing_decision", decision)
	e.broadcast(sourceSession, map[string]any{
		"type":            "routing_decision",
		"chat_jid":        "gi:" + sourceSession,
		"turn_id":         turnID,
		"source_session":  sourceSession,
		"target_session":  targetSession,
		"source_agent_id": sourceAgent,
		"target_agent_id": targetAgentID,
		"mode":            routeMode,
		"matched_by":      matchedBy,
		"created_session": boolValue(metadata["route_created_session"]),
	})
	if targetSession != "" && targetSession != sourceSession {
		incoming := decision
		e.PublishRuntimeRoutingEvent("routing_incoming", incoming)
		e.broadcast(targetSession, map[string]any{
			"type":            "routing_incoming",
			"chat_jid":        "gi:" + targetSession,
			"source_session":  sourceSession,
			"target_session":  targetSession,
			"source_agent_id": sourceAgent,
			"target_agent_id": targetAgentID,
			"turn_id":         turnID,
			"mode":            routeMode,
		})
	}
	return nil
}

func routeDecisionFromMetadata(sourceSession, targetSession, sourceAgent, targetAgentID, turnID, routeMode, matchedBy, routingPolicy, requestedAgent string, metadata map[string]any) store.RouteEvent {
	decision := store.RouteEvent{
		TurnID:         turnID,
		SourceSession:  sourceSession,
		TargetSession:  targetSession,
		SourceAgentID:  sourceAgent,
		TargetAgentID:  targetAgentID,
		Mode:           routeMode,
		MatchedBy:      matchedBy,
		RoutingPolicy:  routingPolicy,
		RequestedAgent: requestedAgent,
		Metadata: map[string]any{
			"routed_from_prompt": boolValue(metadata["routed_from_prompt"]),
			"created_session":    boolValue(metadata["route_created_session"]),
			"routing_enabled":    boolValueOr(metadata["routing_enabled"], true),
		},
	}
	meta := decision.Metadata
	for k, v := range metadata {
		if k == "routed_from_prompt" || k == "route_created_session" || k == "routing_enabled" || k == "route_matched_by" || k == "routing_policy" || k == "target_agent_id" || k == "source_agent_id" || k == "target_session_id" || k == "route_mode" || k == "requested_agent_id" {
			continue
		}
		meta[k] = v
	}
	return decision
}
