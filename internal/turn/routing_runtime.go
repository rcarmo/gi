package turn

import (
	"context"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/routing"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
)

func (e *Engine) SubmitPromptRouted(ctx context.Context, in RunInput) (*SubmitResult, error) {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	if err := e.store.RequireSession(opCtx, in.SessionID); err != nil {
		return nil, err
	}
	targetAgentID, body, directed := routing.ParseDirectedPrompt(in.Prompt)
	promptBody := in.Prompt
	mentioned := false
	if directed {
		if body == "" {
			return nil, fmt.Errorf("directed prompt requires content after @%s", targetAgentID)
		}
		promptBody = body
		mentioned = true
	}
	inbound := inboundContextFromSessionIDWithFallback(opCtx, e.backgroundContext(), e.store, in.SessionID)
	inbound.SenderID = "user"
	inbound.Mentioned = mentioned
	inbound.Prompt = promptBody
	route := e.routeResolver.ResolveRoute(inbound)
	if directed && targetAgentID != "" {
		route.AgentID = routing.NormalizeAgentID(targetAgentID)
		route.MatchedBy = "mention"
	}
	targetSessionID, created, err := e.ResolveOrCreateRouteSession(opCtx, in.SessionID, route, inbound)
	if err != nil {
		return nil, err
	}
	if targetSessionID != in.SessionID {
		return e.submitPeerRoutedPrompt(opCtx, in.SessionID, targetSessionID, route, promptBody, in.Intent, in.Model, created, directed, in.ParentTurnID, in.Metadata)
	}
	sourceSessionID := in.SessionID
	in.SessionID = targetSessionID
	in.Prompt = promptBody
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	in.Metadata["route_mode"] = "prompt"
	in.Metadata["route_matched_by"] = route.MatchedBy
	in.Metadata["target_agent_id"] = route.AgentID
	in.Metadata["target_session_id"] = targetSessionID
	in.Metadata["source_agent_id"] = e.store.SessionAgentID(opCtx, sourceSessionID)
	if route.MatchedBy != "" {
		in.Metadata["routing_policy"] = route.MatchedBy
	}
	in.Metadata["requested_agent_id"] = route.AgentID
	in.Metadata["source_session_id"] = sourceSessionID
	in.Metadata["route_created_session"] = created
	in.Metadata["routing_enabled"] = true
	return e.SubmitPrompt(opCtx, in)
}

func (e *Engine) SubmitPeerMessage(ctx context.Context, sourceSessionID, targetAgentID, content, intent, model, parentTurnID string) (*SubmitResult, error) {
	return e.submitPeerMessageWithMetadata(ctx, sourceSessionID, targetAgentID, content, intent, model, parentTurnID, nil)
}

func (e *Engine) submitPeerMessageWithMetadata(ctx context.Context, sourceSessionID, targetAgentID, content, intent, model, parentTurnID string, extraMetadata map[string]any) (*SubmitResult, error) {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	if err := e.store.RequireSession(opCtx, sourceSessionID); err != nil {
		return nil, err
	}
	inbound := inboundContextFromSessionIDWithFallback(opCtx, e.backgroundContext(), e.store, sourceSessionID)
	inbound.SenderID = e.store.SessionAgentID(opCtx, sourceSessionID)
	inbound.Mentioned = true
	inbound.Prompt = content
	route := e.routeResolver.ResolveRoute(inbound)
	route.AgentID = routing.NormalizeAgentID(targetAgentID)
	route.MatchedBy = "peer-message"
	targetSessionID, created, err := e.ResolveOrCreateRouteSession(opCtx, sourceSessionID, route, inbound)
	if err != nil {
		return nil, err
	}
	return e.submitPeerRoutedPrompt(opCtx, sourceSessionID, targetSessionID, route, content, intent, model, created, true, parentTurnID, extraMetadata)
}

func (e *Engine) ResolveOrCreatePeerSessionID(ctx context.Context, sourceSessionID, targetAgentID string) (string, error) {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	if err := e.store.RequireSession(opCtx, sourceSessionID); err != nil {
		return "", err
	}
	inbound := inboundContextFromSessionIDWithFallback(opCtx, e.backgroundContext(), e.store, sourceSessionID)
	inbound.SenderID = e.store.SessionAgentID(opCtx, sourceSessionID)
	inbound.Mentioned = true
	route := e.routeResolver.ResolveRoute(inbound)
	route.AgentID = routing.NormalizeAgentID(targetAgentID)
	route.MatchedBy = "peer-session"
	targetSessionID, _, err := e.ResolveOrCreateRouteSession(opCtx, sourceSessionID, route, inbound)
	if err != nil {
		return "", err
	}
	return targetSessionID, nil
}

func (e *Engine) ResolveOrCreateRouteSession(ctx context.Context, sourceSessionID string, route routing.ResolvedRoute, inbound routing.InboundContext) (string, bool, error) {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	if strings.TrimSpace(sourceSessionID) == "" {
		return "", false, fmt.Errorf("missing source session")
	}
	if routing.NormalizeAgentID(e.store.SessionAgentID(opCtx, sourceSessionID)) == routing.NormalizeAgentID(route.AgentID) {
		return sourceSessionID, false, nil
	}
	allocation := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       route.AgentID,
		Context:       inbound,
		SessionPolicy: route.SessionPolicy,
	})
	existing, err := e.resolveExistingRouteSession(opCtx, sourceSessionID, route, allocation)
	if err != nil {
		return "", false, err
	}
	if existing != nil {
		return existing.ID, false, nil
	}
	cloned, created, err := e.cloneRouteSession(opCtx, sourceSessionID, route, allocation)
	if err != nil {
		return "", false, err
	}
	return cloned.ID, created, nil
}

func (e *Engine) submitPeerRoutedPrompt(ctx context.Context, sourceSessionID, targetSessionID string, route routing.ResolvedRoute, content, intent, model string, created, directed bool, parentTurnID string, extraMetadata map[string]any) (*SubmitResult, error) {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	sourceAgentID := e.store.SessionAgentID(opCtx, sourceSessionID)
	routingContent := fmt.Sprintf("↪ routed to @%s: %s", route.AgentID, content)
	routingPayload := map[string]any{"kind": "routing", "target_agent_id": route.AgentID, "target_session_id": targetSessionID, "source_agent_id": sourceAgentID, "source_session_id": sourceSessionID, "route_matched_by": route.MatchedBy, "clipped": true}
	warnStore("add routing message to source session", e.store.AddMessage(opCtx, store.NowID("msg"), sourceSessionID, "system", routingContent, routingPayload))
	metadata := map[string]any{
		"source_session_id":     sourceSessionID,
		"source_agent_id":       sourceAgentID,
		"target_session_id":     targetSessionID,
		"target_agent_id":       route.AgentID,
		"requested_agent_id":    route.AgentID,
		"routing_policy":        route.MatchedBy,
		"route_matched_by":      route.MatchedBy,
		"routed_from_prompt":    directed,
		"route_mode":            "peer-message",
		"route_created_session": created,
		"routing_enabled":       true,
	}
	if strings.TrimSpace(parentTurnID) != "" {
		metadata["parent_turn_id"] = parentTurnID
	}
	for k, v := range extraMetadata {
		metadata[k] = v
	}
	result, err := e.SubmitPrompt(opCtx, RunInput{SessionID: targetSessionID, Prompt: content, Intent: intent, Model: model, ParentTurnID: parentTurnID, Metadata: metadata})
	if err != nil {
		return nil, err
	}
	result.SourceSessionID = sourceSessionID
	result.TargetAgentID = route.AgentID
	result.Routed = true
	result.CreatedSession = created
	return result, nil
}

func (e *Engine) modelForAgent(agentID string) string {
	agentID = routing.NormalizeAgentID(agentID)
	for _, agent := range e.runtimeCfg.Agents.List {
		if routing.NormalizeAgentID(agent.ID) == agentID {
			if strings.TrimSpace(agent.Model) != "" {
				return agent.Model
			}
		}
	}
	if strings.TrimSpace(e.runtimeCfg.DefaultModel) != "" {
		return e.runtimeCfg.DefaultModel
	}
	return "bootstrap"
}
