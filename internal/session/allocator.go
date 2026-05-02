package session

import (
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/routing"
)

type Allocation struct {
	Scope          SessionScope `json:"scope"`
	SessionKey     string       `json:"session_key"`
	SessionAliases []string     `json:"session_aliases"`
}

type AllocationInput struct {
	AgentID       string
	Context       routing.InboundContext
	SessionPolicy routing.SessionPolicy
}

func AllocateDefaultSession(agentID, channel, account, logicalChatID string) Allocation {
	agentID = normalize(agentID, "gi")
	channel = normalize(channel, "gi")
	account = normalize(account, "default")
	logicalChatID = normalize(logicalChatID, "default")

	scope := SessionScope{
		Version:    ScopeVersionV1,
		AgentID:    agentID,
		Channel:    channel,
		Account:    account,
		Dimensions: []string{"chat"},
		Values: map[string]string{
			"chat": logicalChatID,
		},
	}
	alias := "agent:" + agentID + ":" + channel + ":chat:" + logicalChatID
	return Allocation{
		Scope:          scope,
		SessionKey:     BuildSessionKey(scope),
		SessionAliases: []string{alias, channel + ":" + logicalChatID},
	}
}

func AllocateRouteSession(input AllocationInput) Allocation {
	scope := buildSessionScope(input)
	alias := "agent:" + routing.NormalizeAgentID(input.AgentID) + ":" + strings.ToLower(scope.Channel)
	for _, dimension := range scope.Dimensions {
		alias += ":" + dimension + ":" + strings.ToLower(strings.TrimSpace(scope.Values[dimension]))
	}
	aliases := []string{alias}
	if chat := strings.TrimSpace(scope.Values["chat"]); chat != "" {
		aliases = append(aliases, strings.ToLower(scope.Channel)+":"+chat)
	}
	return Allocation{Scope: scope, SessionKey: BuildSessionKey(scope), SessionAliases: uniqueAliases(aliases)}
}

func buildSessionScope(input AllocationInput) SessionScope {
	inbound := input.Context
	scope := SessionScope{Version: ScopeVersionV1, AgentID: routing.NormalizeAgentID(input.AgentID), Channel: strings.ToLower(strings.TrimSpace(inbound.Channel)), Account: routing.NormalizeAccountID(inbound.Account)}
	if scope.Channel == "" {
		scope.Channel = "gi"
	}
	dimensions := make([]string, 0, len(input.SessionPolicy.Dimensions))
	values := make(map[string]string, len(input.SessionPolicy.Dimensions))
	for _, dimension := range input.SessionPolicy.Dimensions {
		switch dimension {
		case "space":
			if spaceID := strings.TrimSpace(inbound.SpaceID); spaceID != "" {
				spaceType := strings.ToLower(strings.TrimSpace(inbound.SpaceType))
				if spaceType == "" {
					spaceType = "space"
				}
				dimensions = append(dimensions, "space")
				values["space"] = fmt.Sprintf("%s:%s", spaceType, strings.ToLower(spaceID))
			}
		case "chat":
			if chatID := strings.TrimSpace(inbound.ChatID); chatID != "" {
				chatType := strings.ToLower(strings.TrimSpace(inbound.ChatType))
				if chatType == "" {
					chatType = "direct"
				}
				dimensions = append(dimensions, "chat")
				values["chat"] = fmt.Sprintf("%s:%s", chatType, strings.ToLower(chatID))
			}
		case "topic":
			if topicID := strings.TrimSpace(inbound.TopicID); topicID != "" {
				dimensions = append(dimensions, "topic")
				values["topic"] = "topic:" + strings.ToLower(topicID)
			}
		case "sender":
			if senderID := canonicalSessionIdentityID(inbound.Channel, inbound.SenderID, input.SessionPolicy.IdentityLinks); senderID != "" {
				dimensions = append(dimensions, "sender")
				values["sender"] = senderID
			}
		}
	}
	if len(dimensions) > 0 {
		scope.Dimensions = dimensions
		scope.Values = values
	}
	return scope
}

func canonicalSessionIdentityID(channel, rawID string, identityLinks map[string][]string) string {
	normalizedID := strings.TrimSpace(rawID)
	if normalizedID == "" {
		return ""
	}
	if linked := resolveLinkedPeerID(identityLinks, channel, normalizedID); linked != "" {
		normalizedID = linked
	}
	return strings.ToLower(normalizedID)
}

func resolveLinkedPeerID(identityLinks map[string][]string, channel, peerID string) string {
	if len(identityLinks) == 0 {
		return ""
	}
	peerID = strings.TrimSpace(peerID)
	if peerID == "" {
		return ""
	}
	candidates := map[string]bool{}
	rawCandidate := strings.ToLower(peerID)
	if rawCandidate != "" {
		candidates[rawCandidate] = true
	}
	channel = strings.ToLower(strings.TrimSpace(channel))
	if channel != "" {
		candidates[fmt.Sprintf("%s:%s", channel, rawCandidate)] = true
	}
	if idx := strings.Index(rawCandidate, ":"); idx > 0 && idx < len(rawCandidate)-1 {
		candidates[rawCandidate[idx+1:]] = true
	}
	for canonical, ids := range identityLinks {
		canonicalName := strings.TrimSpace(canonical)
		if canonicalName == "" {
			continue
		}
		for _, id := range ids {
			normalized := strings.ToLower(strings.TrimSpace(id))
			if normalized != "" && candidates[normalized] {
				return canonicalName
			}
		}
	}
	return ""
}

func uniqueAliases(aliases []string) []string {
	if len(aliases) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(aliases))
	seen := make(map[string]struct{}, len(aliases))
	for _, alias := range aliases {
		alias = strings.TrimSpace(strings.ToLower(alias))
		if alias == "" {
			continue
		}
		if _, ok := seen[alias]; ok {
			continue
		}
		seen[alias] = struct{}{}
		normalized = append(normalized, alias)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func normalize(v, fallback string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return fallback
	}
	return v
}
