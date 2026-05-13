package turn

import (
	"context"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/routing"
	"github.com/rcarmo/gi/internal/store"
)

func (e *Engine) SubmitPromptRouted(ctx context.Context, in RunInput) (*SubmitResult, error) {
	resolution, err := e.preparePromptRouteResolution(ctx, in)
	if err != nil {
		return nil, err
	}
	if err := e.resolveRoutedPromptTarget(ctx, resolution); err != nil {
		return nil, err
	}
	if resolution.target.ID != resolution.source.ID {
		return e.submitPeerRoutedPrompt(ctx, resolution.source, resolution.target, resolution.route, resolution.promptBody, in.Intent, in.Model, resolution.created, resolution.directed, in.ParentTurnID, in.Metadata)
	}
	in.SessionID = resolution.target.ID
	in.Prompt = resolution.promptBody
	e.applyLocalRouteMetadata(ctx, &in, resolution)
	return e.SubmitPrompt(ctx, in)
}

func (e *Engine) SubmitPeerMessage(ctx context.Context, sourceSessionID, targetAgentID, content, intent, model, parentTurnID string) (*SubmitResult, error) {
	return e.submitPeerMessageWithMetadata(ctx, sourceSessionID, targetAgentID, content, intent, model, parentTurnID, nil)
}

func (e *Engine) submitPeerMessageWithMetadata(ctx context.Context, sourceSessionID, targetAgentID, content, intent, model, parentTurnID string, extraMetadata map[string]any) (*SubmitResult, error) {
	resolution, err := e.preparePeerRouteResolution(ctx, sourceSessionID, targetAgentID, content, "peer-message")
	if err != nil {
		return nil, err
	}
	if err := e.resolveRoutedPromptTarget(ctx, resolution); err != nil {
		return nil, err
	}
	return e.submitPeerRoutedPrompt(ctx, resolution.source, resolution.target, resolution.route, content, intent, model, resolution.created, resolution.directed, parentTurnID, extraMetadata)
}

func (e *Engine) ResolveOrCreatePeerSession(ctx context.Context, sourceSessionID, targetAgentID string) (*store.Session, bool, error) {
	resolution, err := e.preparePeerRouteResolution(ctx, sourceSessionID, targetAgentID, "", "peer-session")
	if err != nil {
		return nil, false, err
	}
	if err := e.resolveRoutedPromptTarget(ctx, resolution); err != nil {
		return nil, false, err
	}
	return resolution.target, resolution.created, nil
}

func (e *Engine) ResolveOrCreateRouteSession(ctx context.Context, source *store.Session, route routing.ResolvedRoute, inbound routing.InboundContext) (*store.Session, bool, error) {
	if source == nil {
		return nil, false, fmt.Errorf("missing source session")
	}
	if normalizeAgentID(sessionAgentIDWithStore(ctx, e.store, source)) == normalizeAgentID(route.AgentID) {
		return source, false, nil
	}
	plan, err := e.prepareRouteSessionPlan(source, route, inbound)
	if err != nil {
		return nil, false, err
	}
	existing, err := e.resolveExistingRouteSession(ctx, plan)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}
	cloned, created, err := e.cloneRouteSession(ctx, plan)
	if err != nil {
		return nil, false, err
	}
	return cloned, created, nil
}

func (e *Engine) submitPeerRoutedPrompt(ctx context.Context, source, target *store.Session, route routing.ResolvedRoute, content, intent, model string, created, directed bool, parentTurnID string, extraMetadata map[string]any) (*SubmitResult, error) {
	opCtx := ctx
	if opCtx == nil || opCtx.Err() != nil {
		opCtx = e.backgroundContext()
	}
	sourceAgentID := sessionAgentIDWithStore(opCtx, e.store, source)
	routingContent := fmt.Sprintf("↪ routed to @%s: %s", route.AgentID, content)
	routingPayload := map[string]any{"kind": "routing", "target_agent_id": route.AgentID, "target_session_id": target.ID, "source_agent_id": sourceAgentID, "source_session_id": source.ID, "route_matched_by": route.MatchedBy, "clipped": true}
	warnStore("add routing message to source session", e.store.AddMessage(opCtx, store.NowID("msg"), source.ID, "system", routingContent, routingPayload))
	metadata := map[string]any{
		"source_session_id":     source.ID,
		"source_agent_id":       sourceAgentID,
		"target_session_id":     target.ID,
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
	result, err := e.SubmitPrompt(opCtx, RunInput{SessionID: target.ID, Prompt: content, Intent: intent, Model: model, ParentTurnID: parentTurnID, Metadata: metadata})
	if err != nil {
		return nil, err
	}
	result.SourceSessionID = source.ID
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
