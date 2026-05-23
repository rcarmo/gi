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
	sourceIdentity, err := st.RequireSessionIdentityRuntime(ctx, sourceSessionID)
	if err != nil {
		return "", false, err
	}
	if routing.NormalizeAgentID(sourceIdentity.AgentID) == routing.NormalizeAgentID(route.AgentID) {
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
		messages, err := st.ListMessages(ctx, sourceSessionID)
		if err != nil {
			return "", false, err
		}
		for _, msg := range messages {
			payload := map[string]any{}
			for k, v := range msg.Payload {
				payload[k] = v
			}
			payload["forked_from_message_id"] = msg.ID
			if err := st.AddMessage(ctx, store.NowID("msg"), cloned.ID, msg.Role, msg.Content, payload); err != nil {
				return "", false, err
			}
		}
		sourceAgentID := sourceIdentity.AgentID
		if err := st.AddMessage(ctx, store.NowID("msg"), cloned.ID, "system", fmt.Sprintf("Forked from @%s", sourceAgentID), map[string]any{"kind": "fork", "source_session_id": sourceSessionID, "source_agent_id": sourceAgentID, "route_matched_by": route.MatchedBy, "clipped": true}); err != nil {
			return "", false, err
		}
	}
	return cloned.ID, created, nil
}

func inboundContextFromIdentitySnapshot(sessionID string, snapshot store.SessionIdentitySnapshot) routing.InboundContext {
	inbound := routing.InboundContext{Channel: snapshot.Runtime.Channel, Account: snapshot.Runtime.Account}
	dimensions := snapshot.Dimensions
	if strings.TrimSpace(sessionID) == "" || len(dimensions) == 0 {
		if strings.TrimSpace(sessionID) != "" {
			inbound.ChatType = "direct"
			inbound.ChatID = sessionID
		}
		return inbound
	}
	if raw := strings.TrimSpace(dimensions["space"]); raw != "" {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			inbound.SpaceType = normalize(parts[0])
			inbound.SpaceID = normalize(parts[1])
		} else {
			inbound.SpaceType = "space"
			inbound.SpaceID = normalize(raw)
		}
	}
	if raw := strings.TrimSpace(dimensions["chat"]); raw != "" {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			inbound.ChatType = normalize(parts[0])
			inbound.ChatID = normalize(parts[1])
		} else {
			inbound.ChatType = "direct"
			inbound.ChatID = normalize(raw)
		}
	}
	if raw := strings.TrimSpace(dimensions["topic"]); raw != "" {
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

func RequireInboundContextFromSession(ctx context.Context, st *store.Store, sessionID string) (routing.InboundContext, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if st == nil {
		return routing.InboundContext{}, sql.ErrNoRows
	}
	snapshot, err := st.RequireSessionIdentitySnapshot(ctx, sessionID)
	if err != nil {
		return routing.InboundContext{}, err
	}
	return inboundContextFromIdentitySnapshot(sessionID, snapshot), nil
}

func normalize(v string) string { return strings.ToLower(strings.TrimSpace(v)) }
