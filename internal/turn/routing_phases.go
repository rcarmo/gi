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

func (e *Engine) preparePromptRouteResolution(ctx context.Context, in RunInput) (routing.ResolvedRoute, routing.InboundContext, string, bool, error) {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	if err := e.store.RequireSession(opCtx, in.SessionID); err != nil {
		return routing.ResolvedRoute{}, routing.InboundContext{}, "", false, err
	}
	targetAgentID, body, directed := routing.ParseDirectedPrompt(in.Prompt)
	promptBody := in.Prompt
	mentioned := false
	if directed {
		if body == "" {
			return routing.ResolvedRoute{}, routing.InboundContext{}, "", false, fmt.Errorf("directed prompt requires content after @%s", targetAgentID)
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
	return route, inbound, promptBody, directed, nil
}

func (e *Engine) preparePeerRouteResolution(ctx context.Context, sourceSessionID, targetAgentID, content, matchedBy string) (routing.ResolvedRoute, routing.InboundContext, error) {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	if err := e.store.RequireSession(opCtx, sourceSessionID); err != nil {
		return routing.ResolvedRoute{}, routing.InboundContext{}, err
	}
	inbound := inboundContextFromSessionIDWithFallback(opCtx, e.backgroundContext(), e.store, sourceSessionID)
	inbound.SenderID = e.store.SessionAgentID(opCtx, sourceSessionID)
	inbound.Mentioned = true
	inbound.Prompt = content
	route := e.routeResolver.ResolveRoute(inbound)
	route.AgentID = routing.NormalizeAgentID(targetAgentID)
	route.MatchedBy = matchedBy
	return route, inbound, nil
}

func (e *Engine) applyLocalRouteMetadata(ctx context.Context, in *RunInput, sourceSessionID, targetSessionID string, route routing.ResolvedRoute, created bool) {
	ctx = coordinationContext(ctx, e.backgroundContext())
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	in.Metadata["route_mode"] = "prompt"
	in.Metadata["route_matched_by"] = route.MatchedBy
	in.Metadata["target_agent_id"] = route.AgentID
	in.Metadata["target_session_id"] = targetSessionID
	in.Metadata["source_agent_id"] = e.store.SessionAgentID(coordinationContext(ctx, e.backgroundContext()), sourceSessionID)
	if route.MatchedBy != "" {
		in.Metadata["routing_policy"] = route.MatchedBy
	}
	in.Metadata["requested_agent_id"] = route.AgentID
	in.Metadata["source_session_id"] = sourceSessionID
	in.Metadata["route_created_session"] = created
	in.Metadata["routing_enabled"] = true
}

func (e *Engine) resolveExistingRouteSession(ctx context.Context, sourceSessionID string, route routing.ResolvedRoute, allocation gisession.Allocation) (*store.Session, error) {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	if sessionID, err := e.store.FindSessionByAllocation(opCtx, allocation); err == nil {
		return e.store.GetSession(opCtx, sessionID)
	} else if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if sessionID, err := e.store.FindChildSessionIDByParentAndAgent(opCtx, sourceSessionID, route.AgentID); err == nil {
		return e.store.GetSession(opCtx, sessionID)
	} else if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if sessionID, err := e.store.FindSiblingChildSessionIDByParentAndAgent(opCtx, sourceSessionID, route.AgentID); err == nil {
		return e.store.GetSession(opCtx, sessionID)
	} else if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return nil, nil
}

func (e *Engine) cloneRouteSession(ctx context.Context, sourceSessionID string, route routing.ResolvedRoute, allocation gisession.Allocation) (*store.Session, bool, error) {
	opCtx := coordinationContext(ctx, e.backgroundContext())
	state := map[string]any{"status": "idle", "queue_count": 0, "model": e.modelForAgent(route.AgentID), "provider": e.runtimeCfg.DefaultProvider, "thinking_level": e.runtimeCfg.DefaultThinkingLevel}
	cloned, created, err := e.store.ResolveOrCreateSessionFromAllocation(opCtx, store.ResolveOrCreateSessionFromAllocationInput{
		ID:              store.NowID("session"),
		ParentSessionID: sourceSessionID,
		Title:           "@" + route.AgentID,
		State:           state,
		Allocation:      allocation,
	})
	if err != nil {
		return nil, false, err
	}
	if !created {
		return cloned, false, nil
	}
	e.copyRouteSessionHistory(opCtx, sourceSessionID, cloned.ID)
	sourceAgentID := e.store.SessionAgentID(opCtx, sourceSessionID)
	warnStore("add forked-from message", e.store.AddMessage(opCtx, store.NowID("msg"), cloned.ID, "system", fmt.Sprintf("Forked from @%s", sourceAgentID), map[string]any{"kind": "fork", "source_session_id": sourceSessionID, "source_agent_id": sourceAgentID, "route_matched_by": route.MatchedBy, "clipped": true}))
	return cloned, true, nil
}

func inboundContextFromSessionIDWithFallback(ctx, fallback context.Context, s *store.Store, sessionID string) routing.InboundContext {
	opCtx := coordinationContext(ctx, fallback)
	inbound := routing.InboundContext{}
	if s != nil {
		inbound.Channel = s.SessionChannel(opCtx, sessionID)
		inbound.Account = s.SessionAccount(opCtx, sessionID)
	}
	identity := (*store.SessionIdentity)(nil)
	if s != nil {
		identity = s.SessionIdentityOrNil(opCtx, sessionID)
	}
	var scope *gisession.SessionScope
	if identity != nil {
		scope = &identity.Scope
	}
	if inbound.Channel == "" {
		inbound.Channel = "gi"
	}
	if inbound.Account == "" {
		inbound.Account = "default"
	}
	if strings.TrimSpace(sessionID) == "" || scope == nil || scope.Values == nil {
		if strings.TrimSpace(sessionID) != "" {
			inbound.ChatType = "direct"
			inbound.ChatID = sessionID
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
	if inbound.ChatID == "" && strings.TrimSpace(sessionID) != "" {
		inbound.ChatType = "direct"
		inbound.ChatID = sessionID
	}
	return inbound
}

func inboundContextFromSessionID(ctx context.Context, s *store.Store, sessionID string) routing.InboundContext {
	return inboundContextFromSessionIDWithFallback(ctx, nil, s, sessionID)
}

func splitScopedValue(raw, fallbackType string) (string, string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", false
	}
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
		return normalizedLowerString(parts[0]), normalizedLowerString(parts[1]), true
	}
	if strings.TrimSpace(fallbackType) == "" {
		return "", normalizedLowerString(raw), true
	}
	return normalizedLowerString(fallbackType), normalizedLowerString(raw), true
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
