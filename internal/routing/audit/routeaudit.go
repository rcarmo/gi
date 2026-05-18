package audit

import (
	"context"

	"github.com/rcarmo/gi/internal/store"
)

type Options struct {
	PublishRuntimeRoutingEvent func(string, Event)
	Broadcast                  func(string, map[string]any)
}

func RecordDecision(ctx context.Context, st *store.Store, sourceSessionID, turnID string, metadata map[string]any, opts Options) error {
	sourceSession := stringValue(metadata["source_session_id"], sourceSessionID)
	targetSession := stringValue(metadata["target_session_id"], "")
	targetAgentID := stringValue(metadata["target_agent_id"], "")
	if targetAgentID == "" {
		return nil
	}
	routeMode := stringValue(metadata["route_mode"], stringValue(metadata["mode"], "prompt"))
	sourceAgent := stringValue(metadata["source_agent_id"], "")
	if sourceAgent == "" {
		sourceAgent = st.SessionAgentID(ctx, sourceSession)
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
	decision := decisionFromMetadata(sourceSession, targetSession, sourceAgent, targetAgentID, turnID, routeMode, matchedBy, routingPolicy, requestedAgent, metadata)
	if !boolValueOr(metadata["routing_enabled"], true) {
		return nil
	}
	routeEventID, err := RecordEvent(ctx, st, decision)
	if err != nil {
		return err
	}
	decision.ID = routeEventID
	if opts.PublishRuntimeRoutingEvent != nil {
		opts.PublishRuntimeRoutingEvent("routing_decision", decision)
	}
	if opts.Broadcast != nil {
		opts.Broadcast(sourceSession, map[string]any{
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
	}
	if targetSession != "" && targetSession != sourceSession {
		if opts.PublishRuntimeRoutingEvent != nil {
			opts.PublishRuntimeRoutingEvent("routing_incoming", decision)
		}
		if opts.Broadcast != nil {
			opts.Broadcast(targetSession, map[string]any{
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
	}
	return nil
}

func decisionFromMetadata(sourceSession, targetSession, sourceAgent, targetAgentID, turnID, routeMode, matchedBy, routingPolicy, requestedAgent string, metadata map[string]any) Event {
	decision := Event{
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

func stringValue(v any, fallback string) string {
	if s, ok := v.(string); ok {
		if s != "" {
			return s
		}
	}
	return fallback
}
func boolValue(v any) bool { b, ok := v.(bool); return ok && b }
func boolValueOr(v any, fallback bool) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return fallback
}
