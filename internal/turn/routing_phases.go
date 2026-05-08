package turn

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/routing"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
)

type routedPromptResolution struct {
	source     *store.Session
	target     *store.Session
	route      routing.ResolvedRoute
	inbound    routing.InboundContext
	promptBody string
	directed   bool
	created    bool
}

type routeSessionPlan struct {
	source     *store.Session
	route      routing.ResolvedRoute
	inbound    routing.InboundContext
	allocation gisession.Allocation
}

func (e *Engine) preparePromptRouteResolution(ctx context.Context, in RunInput) (*routedPromptResolution, error) {
	source, err := e.store.GetSession(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}
	targetAgentID, body, directed := parseDirectedPrompt(in.Prompt)
	promptBody := in.Prompt
	mentioned := false
	if directed {
		if body == "" {
			return nil, fmt.Errorf("directed prompt requires content after @%s", targetAgentID)
		}
		promptBody = body
		mentioned = true
	}
	inbound := routing.InboundContext{
		Channel:   sessionChannel(source),
		Account:   sessionAccount(source),
		ChatType:  "direct",
		ChatID:    source.ID,
		SenderID:  "user",
		Mentioned: mentioned,
		Prompt:    promptBody,
	}
	route := e.routeResolver.ResolveRoute(inbound)
	if directed && targetAgentID != "" {
		route.AgentID = routing.NormalizeAgentID(targetAgentID)
		route.MatchedBy = "mention"
	}
	return &routedPromptResolution{
		source:     source,
		route:      route,
		inbound:    inbound,
		promptBody: promptBody,
		directed:   directed,
	}, nil
}

func (e *Engine) preparePeerRouteResolution(ctx context.Context, sourceSessionID, targetAgentID, content, matchedBy string) (*routedPromptResolution, error) {
	source, err := e.store.GetSession(ctx, sourceSessionID)
	if err != nil {
		return nil, err
	}
	inbound := routing.InboundContext{
		Channel:   sessionChannel(source),
		Account:   sessionAccount(source),
		ChatType:  "direct",
		ChatID:    source.ID,
		SenderID:  sessionAgentID(source),
		Mentioned: true,
		Prompt:    content,
	}
	route := e.routeResolver.ResolveRoute(inbound)
	route.AgentID = routing.NormalizeAgentID(targetAgentID)
	route.MatchedBy = matchedBy
	return &routedPromptResolution{
		source:     source,
		route:      route,
		inbound:    inbound,
		promptBody: content,
		directed:   true,
	}, nil
}

func (e *Engine) resolveRoutedPromptTarget(ctx context.Context, resolution *routedPromptResolution) error {
	target, created, err := e.ResolveOrCreateRouteSession(ctx, resolution.source, resolution.route, resolution.inbound)
	if err != nil {
		return err
	}
	resolution.target = target
	resolution.created = created
	return nil
}

func (e *Engine) applyLocalRouteMetadata(in *RunInput, resolution *routedPromptResolution) {
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	in.Metadata["route_mode"] = "prompt"
	in.Metadata["route_matched_by"] = resolution.route.MatchedBy
	in.Metadata["target_agent_id"] = resolution.route.AgentID
	in.Metadata["target_session_id"] = resolution.target.ID
	in.Metadata["source_agent_id"] = sessionAgentID(resolution.source)
	if resolution.route.MatchedBy != "" {
		in.Metadata["routing_policy"] = resolution.route.MatchedBy
	}
	in.Metadata["requested_agent_id"] = resolution.route.AgentID
	in.Metadata["source_session_id"] = resolution.source.ID
	in.Metadata["route_created_session"] = resolution.created
	in.Metadata["routing_enabled"] = true
}

func (e *Engine) prepareRouteSessionPlan(source *store.Session, route routing.ResolvedRoute, inbound routing.InboundContext) (*routeSessionPlan, error) {
	if source == nil {
		return nil, fmt.Errorf("missing source session")
	}
	return &routeSessionPlan{
		source:  source,
		route:   route,
		inbound: inbound,
		allocation: gisession.AllocateRouteSession(gisession.AllocationInput{
			AgentID:       route.AgentID,
			Context:       inbound,
			SessionPolicy: route.SessionPolicy,
		}),
	}, nil
}

func (e *Engine) resolveExistingRouteSession(ctx context.Context, plan *routeSessionPlan) (*store.Session, error) {
	if existing, err := e.store.FindSessionByAllocation(ctx, plan.allocation); err == nil {
		return existing, nil
	} else if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if existing, err := e.store.FindChildSessionByParentAndAgent(ctx, plan.source.ID, plan.route.AgentID); err == nil {
		return existing, nil
	} else if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if strings.TrimSpace(plan.source.ParentSessionID) != "" {
		if existing, err := e.store.FindChildSessionByParentAndAgent(ctx, plan.source.ParentSessionID, plan.route.AgentID); err == nil {
			return existing, nil
		} else if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
	}
	return nil, nil
}

func (e *Engine) cloneRouteSession(ctx context.Context, plan *routeSessionPlan) (*store.Session, error) {
	state := map[string]any{"status": "idle", "queue_count": 0, "model": e.modelForAgent(plan.route.AgentID), "provider": e.runtimeCfg.DefaultProvider, "thinking_level": e.runtimeCfg.DefaultThinkingLevel}
	cloned, err := e.store.CreateSessionWithMetadata(ctx, store.NowID("session"), plan.source.ID, "@"+plan.route.AgentID, state, &plan.allocation.Scope, plan.allocation.SessionAliases)
	if err != nil {
		return nil, err
	}
	e.copyRouteSessionHistory(ctx, plan.source.ID, cloned.ID)
	warnStore("add forked-from message", e.store.AddMessage(ctx, store.NowID("msg"), cloned.ID, "system", fmt.Sprintf("Forked from @%s", sessionAgentID(plan.source)), map[string]any{"kind": "fork", "source_session_id": plan.source.ID, "source_agent_id": sessionAgentID(plan.source), "route_matched_by": plan.route.MatchedBy, "clipped": true}))
	return cloned, nil
}

func (e *Engine) copyRouteSessionHistory(ctx context.Context, sourceSessionID, targetSessionID string) {
	messages, err := e.store.ListMessages(ctx, sourceSessionID)
	if err != nil {
		return
	}
	for _, msg := range messages {
		payload := map[string]any{}
		for k, v := range msg.Payload {
			payload[k] = v
		}
		payload["forked_from_message_id"] = msg.ID
		warnStore("copy message to cloned session", e.store.AddMessage(ctx, store.NowID("msg"), targetSessionID, msg.Role, msg.Content, payload))
	}
}
