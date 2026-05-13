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
	opCtx := ctx
	if opCtx == nil || opCtx.Err() != nil {
		opCtx = e.backgroundContext()
	}
	source, err := e.store.GetSession(opCtx, in.SessionID)
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
	inbound := inboundContextFromSession(opCtx, e.store, source)
	inbound.SenderID = "user"
	inbound.Mentioned = mentioned
	inbound.Prompt = promptBody
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
	opCtx := ctx
	if opCtx == nil || opCtx.Err() != nil {
		opCtx = e.backgroundContext()
	}
	source, err := e.store.GetSession(opCtx, sourceSessionID)
	if err != nil {
		return nil, err
	}
	inbound := inboundContextFromSession(opCtx, e.store, source)
	inbound.SenderID = sessionAgentIDWithStore(opCtx, e.store, source)
	inbound.Mentioned = true
	inbound.Prompt = content
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

func (e *Engine) applyLocalRouteMetadata(ctx context.Context, in *RunInput, resolution *routedPromptResolution) {
	if ctx == nil || ctx.Err() != nil {
		ctx = e.backgroundContext()
	}
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	in.Metadata["route_mode"] = "prompt"
	in.Metadata["route_matched_by"] = resolution.route.MatchedBy
	in.Metadata["target_agent_id"] = resolution.route.AgentID
	in.Metadata["target_session_id"] = resolution.target.ID
	in.Metadata["source_agent_id"] = sessionAgentIDWithStore(ctx, e.store, resolution.source)
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
	opCtx := ctx
	if opCtx == nil || opCtx.Err() != nil {
		opCtx = e.backgroundContext()
	}
	if existing, err := e.store.ResolveSessionByAllocation(opCtx, plan.allocation); err == nil {
		return existing, nil
	} else if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if existing, err := e.store.FindChildSessionByParentAndAgent(opCtx, plan.source.ID, plan.route.AgentID); err == nil {
		return existing, nil
	} else if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if strings.TrimSpace(plan.source.ParentSessionID) != "" {
		if existing, err := e.store.FindChildSessionByParentAndAgent(opCtx, plan.source.ParentSessionID, plan.route.AgentID); err == nil {
			return existing, nil
		} else if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
	}
	return nil, nil
}

func (e *Engine) cloneRouteSession(ctx context.Context, plan *routeSessionPlan) (*store.Session, bool, error) {
	opCtx := ctx
	if opCtx == nil || opCtx.Err() != nil {
		opCtx = e.backgroundContext()
	}
	state := map[string]any{"status": "idle", "queue_count": 0, "model": e.modelForAgent(plan.route.AgentID), "provider": e.runtimeCfg.DefaultProvider, "thinking_level": e.runtimeCfg.DefaultThinkingLevel}
	cloned, created, err := e.store.ResolveOrCreateSessionFromAllocation(opCtx, store.ResolveOrCreateSessionFromAllocationInput{
		ID:              store.NowID("session"),
		ParentSessionID: plan.source.ID,
		Title:           "@" + plan.route.AgentID,
		State:           state,
		Allocation:      plan.allocation,
	})
	if err != nil {
		return nil, false, err
	}
	if !created {
		return cloned, false, nil
	}
	e.copyRouteSessionHistory(opCtx, plan.source.ID, cloned.ID)
	sourceAgentID := sessionAgentIDWithStore(opCtx, e.store, plan.source)
	warnStore("add forked-from message", e.store.AddMessage(opCtx, store.NowID("msg"), cloned.ID, "system", fmt.Sprintf("Forked from @%s", sourceAgentID), map[string]any{"kind": "fork", "source_session_id": plan.source.ID, "source_agent_id": sourceAgentID, "route_matched_by": plan.route.MatchedBy, "clipped": true}))
	return cloned, true, nil
}

func inboundContextFromSession(ctx context.Context, s *store.Store, sess *store.Session) routing.InboundContext {
	inbound := routing.InboundContext{
		Channel: sessionChannelWithStore(ctx, s, sess),
		Account: sessionAccountWithStore(ctx, s, sess),
	}
	identity := lookupSessionIdentity(ctx, s, sess)
	var scope *gisession.SessionScope
	if identity != nil {
		scope = &identity.Scope
	} else if sess != nil {
		scope = sess.Scope
	}
	if sess == nil || scope == nil || scope.Values == nil {
		if sess != nil {
			inbound.ChatType = "direct"
			inbound.ChatID = sess.ID
		}
		return inbound
	}
	if spaceType, spaceID, ok := splitScopedValue(scope.Values["space"], "space"); ok {
		inbound.SpaceType = spaceType
		inbound.SpaceID = spaceID
	}
	if chatType, chatID, ok := splitScopedValue(scope.Values["chat"], "direct"); ok {
		inbound.ChatType = chatType
		inbound.ChatID = chatID
	}
	if _, topicID, ok := splitScopedValue(scope.Values["topic"], "topic"); ok {
		inbound.TopicID = topicID
	}
	if inbound.ChatID == "" && sess != nil {
		inbound.ChatType = "direct"
		inbound.ChatID = sess.ID
	}
	return inbound
}

func splitScopedValue(raw, fallbackType string) (string, string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", false
	}
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
		return strings.ToLower(strings.TrimSpace(parts[0])), strings.ToLower(strings.TrimSpace(parts[1])), true
	}
	if strings.TrimSpace(fallbackType) == "" {
		return "", strings.ToLower(raw), true
	}
	return strings.ToLower(strings.TrimSpace(fallbackType)), strings.ToLower(raw), true
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
