package routedsession

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/routing"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
)

type ResolveOptions struct {
	ModelForAgent   func(string) string
	DefaultProvider string
	DefaultThinking string
}

func ResolveOrCreate(ctx context.Context, st *store.Store, sourceSessionID string, route routing.ResolvedRoute, inbound routing.InboundContext, opts ResolveOptions) (string, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(sourceSessionID) == "" {
		return "", false, fmt.Errorf("missing source session")
	}
	if routing.NormalizeAgentID(st.SessionAgentID(ctx, sourceSessionID)) == routing.NormalizeAgentID(route.AgentID) {
		return sourceSessionID, false, nil
	}
	allocation := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       route.AgentID,
		Context:       inbound,
		SessionPolicy: route.SessionPolicy,
	})
	if sessionID, err := st.FindSessionByAllocation(ctx, allocation); err == nil {
		return sessionID, false, nil
	} else if err != nil && err != sql.ErrNoRows {
		return "", false, err
	}
	if sessionID, err := st.FindChildSessionIDByParentAndAgent(ctx, sourceSessionID, route.AgentID); err == nil {
		return sessionID, false, nil
	} else if err != nil && err != sql.ErrNoRows {
		return "", false, err
	}
	if sessionID, err := st.FindSiblingChildSessionIDByParentAndAgent(ctx, sourceSessionID, route.AgentID); err == nil {
		return sessionID, false, nil
	} else if err != nil && err != sql.ErrNoRows {
		return "", false, err
	}
	model := ""
	if opts.ModelForAgent != nil {
		model = opts.ModelForAgent(route.AgentID)
	}
	state := map[string]any{"status": "idle", "queue_count": 0, "model": model, "provider": opts.DefaultProvider, "thinking_level": opts.DefaultThinking}
	cloned, created, err := st.ResolveOrCreateSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{
		ID:              store.NowID("session"),
		ParentSessionID: sourceSessionID,
		Title:           "@" + route.AgentID,
		State:           state,
		Allocation:      allocation,
	})
	if err != nil {
		return "", false, err
	}
	if created {
		if messages, err := st.ListMessages(ctx, sourceSessionID); err == nil {
			for _, msg := range messages {
				payload := map[string]any{}
				for k, v := range msg.Payload {
					payload[k] = v
				}
				payload["forked_from_message_id"] = msg.ID
				_ = st.AddMessage(ctx, store.NowID("msg"), cloned.ID, msg.Role, msg.Content, payload)
			}
		}
		sourceAgentID := st.SessionAgentID(ctx, sourceSessionID)
		_ = st.AddMessage(ctx, store.NowID("msg"), cloned.ID, "system", fmt.Sprintf("Forked from @%s", sourceAgentID), map[string]any{"kind": "fork", "source_session_id": sourceSessionID, "source_agent_id": sourceAgentID, "route_matched_by": route.MatchedBy, "clipped": true})
	}
	return cloned.ID, created, nil
}

func InboundContextFromSession(ctx context.Context, st *store.Store, sessionID string) routing.InboundContext {
	if ctx == nil {
		ctx = context.Background()
	}
	inbound := routing.InboundContext{}
	if st != nil {
		inbound.Channel = st.SessionChannel(ctx, sessionID)
		inbound.Account = st.SessionAccount(ctx, sessionID)
	}
	identity := (*store.SessionIdentity)(nil)
	if st != nil {
		identity = st.SessionIdentityOrNil(ctx, sessionID)
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
	if raw := strings.TrimSpace(scope.Values["space"]); raw != "" {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			inbound.SpaceType = normalize(parts[0])
			inbound.SpaceID = normalize(parts[1])
		} else {
			inbound.SpaceType = "space"
			inbound.SpaceID = normalize(raw)
		}
	}
	if raw := strings.TrimSpace(scope.Values["chat"]); raw != "" {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			inbound.ChatType = normalize(parts[0])
			inbound.ChatID = normalize(parts[1])
		} else {
			inbound.ChatType = "direct"
			inbound.ChatID = normalize(raw)
		}
	}
	if raw := strings.TrimSpace(scope.Values["topic"]); raw != "" {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			inbound.TopicID = normalize(parts[1])
		} else {
			inbound.TopicID = normalize(raw)
		}
	}
	if inbound.ChatID == "" && strings.TrimSpace(sessionID) != "" {
		inbound.ChatType = "direct"
		inbound.ChatID = sessionID
	}
	return inbound
}

func normalize(v string) string { return strings.ToLower(strings.TrimSpace(v)) }
