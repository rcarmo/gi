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
	messages, err := e.store.ListMessages(opCtx, sourceSessionID)
	if err == nil {
		for _, msg := range messages {
			payload := map[string]any{}
			for k, v := range msg.Payload {
				payload[k] = v
			}
			payload["forked_from_message_id"] = msg.ID
			warnStore("copy message to cloned session", e.store.AddMessage(opCtx, store.NowID("msg"), cloned.ID, msg.Role, msg.Content, payload))
		}
	}
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
