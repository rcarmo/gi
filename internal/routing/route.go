package routing

import (
	"fmt"
	"strings"
)

type RouteResolver struct {
	agents  AgentsConfig
	session SessionConfig
}

func NewRouteResolver(agents AgentsConfig, session SessionConfig) *RouteResolver {
	return &RouteResolver{agents: agents, session: session}
}

func (r *RouteResolver) ResolveRoute(inbound InboundContext) ResolvedRoute {
	channel := strings.ToLower(strings.TrimSpace(inbound.Channel))
	accountID := NormalizeAccountID(inbound.Account)
	identityLinks := cloneIdentityLinks(r.session.IdentityLinks)
	view := buildDispatchView(inbound, identityLinks)

	if rule := r.matchDispatchRule(view); rule != nil {
		return ResolvedRoute{
			AgentID:       r.pickAgentID(rule.Agent),
			Channel:       channel,
			AccountID:     accountID,
			SessionPolicy: r.sessionPolicy(rule),
			MatchedBy:     matchedByForRule(rule),
		}
	}
	return ResolvedRoute{
		AgentID:       r.pickAgentID(r.resolveDefaultAgentID()),
		Channel:       channel,
		AccountID:     accountID,
		SessionPolicy: r.sessionPolicy(nil),
		MatchedBy:     "default",
	}
}

func (r *RouteResolver) pickAgentID(agentID string) string {
	trimmed := strings.TrimSpace(agentID)
	if trimmed == "" {
		return NormalizeAgentID(r.resolveDefaultAgentID())
	}
	normalized := NormalizeAgentID(trimmed)
	if len(r.agents.List) == 0 {
		return normalized
	}
	for _, a := range r.agents.List {
		if NormalizeAgentID(a.ID) == normalized {
			return normalized
		}
	}
	return NormalizeAgentID(r.resolveDefaultAgentID())
}

func (r *RouteResolver) resolveDefaultAgentID() string {
	if len(r.agents.List) == 0 {
		return DefaultAgentID
	}
	for _, a := range r.agents.List {
		if a.Default && strings.TrimSpace(a.ID) != "" {
			return NormalizeAgentID(a.ID)
		}
	}
	if strings.TrimSpace(r.agents.List[0].ID) != "" {
		return NormalizeAgentID(r.agents.List[0].ID)
	}
	return DefaultAgentID
}

func (r *RouteResolver) sessionPolicy(rule *DispatchRule) SessionPolicy {
	dimensions := r.session.Dimensions
	if rule != nil && len(rule.SessionDimensions) > 0 {
		dimensions = rule.SessionDimensions
	}
	return SessionPolicy{Dimensions: normalizeSessionDimensions(dimensions), IdentityLinks: cloneIdentityLinks(r.session.IdentityLinks)}
}

func normalizeSessionDimensions(dimensions []string) []string {
	if len(dimensions) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(dimensions))
	seen := make(map[string]struct{}, len(dimensions))
	for _, dimension := range dimensions {
		dimension = strings.ToLower(strings.TrimSpace(dimension))
		switch dimension {
		case "space", "chat", "topic", "sender":
		default:
			continue
		}
		if _, ok := seen[dimension]; ok {
			continue
		}
		seen[dimension] = struct{}{}
		normalized = append(normalized, dimension)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func cloneIdentityLinks(src map[string][]string) map[string][]string {
	if len(src) == 0 {
		return nil
	}
	cloned := make(map[string][]string, len(src))
	for canonical, ids := range src {
		dup := make([]string, len(ids))
		copy(dup, ids)
		cloned[canonical] = dup
	}
	return cloned
}

type dispatchView struct {
	Channel   string
	Account   string
	Space     string
	Chat      string
	Topic     string
	Sender    string
	Mentioned bool
}

func (r *RouteResolver) matchDispatchRule(view dispatchView) *DispatchRule {
	if r == nil || r.agents.Dispatch == nil || len(r.agents.Dispatch.Rules) == 0 {
		return nil
	}
	for i := range r.agents.Dispatch.Rules {
		rule := &r.agents.Dispatch.Rules[i]
		if !selectorHasAnyConstraint(rule.When) {
			continue
		}
		if ruleMatchesView(*rule, view) {
			return rule
		}
	}
	return nil
}

func ruleMatchesView(rule DispatchRule, view dispatchView) bool {
	when := normalizeDispatchSelector(rule.When)
	if when.Channel != "" && when.Channel != view.Channel {
		return false
	}
	if when.Account != "" && when.Account != view.Account {
		return false
	}
	if when.Space != "" && when.Space != view.Space {
		return false
	}
	if when.Chat != "" && when.Chat != view.Chat {
		return false
	}
	if when.Topic != "" && when.Topic != view.Topic {
		return false
	}
	if when.Sender != "" && when.Sender != view.Sender {
		return false
	}
	if when.Mentioned != nil && *when.Mentioned != view.Mentioned {
		return false
	}
	return true
}

func matchedByForRule(rule *DispatchRule) string {
	if rule == nil {
		return "default"
	}
	name := strings.TrimSpace(rule.Name)
	if name == "" {
		return "dispatch.rule"
	}
	return "dispatch.rule:" + strings.ToLower(name)
}

func buildDispatchView(inbound InboundContext, identityLinks map[string][]string) dispatchView {
	view := dispatchView{Channel: strings.ToLower(strings.TrimSpace(inbound.Channel)), Account: NormalizeAccountID(inbound.Account), Mentioned: inbound.Mentioned}
	if spaceID := strings.TrimSpace(inbound.SpaceID); spaceID != "" {
		spaceType := strings.ToLower(strings.TrimSpace(inbound.SpaceType))
		if spaceType == "" {
			spaceType = "space"
		}
		view.Space = fmt.Sprintf("%s:%s", spaceType, strings.ToLower(spaceID))
	}
	if chatID := strings.TrimSpace(inbound.ChatID); chatID != "" {
		chatType := strings.ToLower(strings.TrimSpace(inbound.ChatType))
		if chatType == "" {
			chatType = "direct"
		}
		view.Chat = fmt.Sprintf("%s:%s", chatType, strings.ToLower(chatID))
	}
	if topicID := strings.TrimSpace(inbound.TopicID); topicID != "" {
		view.Topic = "topic:" + strings.ToLower(topicID)
	}
	view.Sender = canonicalDispatchSenderID(inbound.Channel, inbound.SenderID, identityLinks)
	return view
}

func normalizeDispatchSelector(selector DispatchSelector) DispatchSelector {
	selector.Channel = strings.ToLower(strings.TrimSpace(selector.Channel))
	selector.Account = NormalizeAccountID(selector.Account)
	selector.Space = strings.ToLower(strings.TrimSpace(selector.Space))
	selector.Chat = strings.ToLower(strings.TrimSpace(selector.Chat))
	selector.Topic = strings.ToLower(strings.TrimSpace(selector.Topic))
	selector.Sender = strings.ToLower(strings.TrimSpace(selector.Sender))
	return selector
}

func selectorHasAnyConstraint(selector DispatchSelector) bool {
	return strings.TrimSpace(selector.Channel) != "" || strings.TrimSpace(selector.Account) != "" || strings.TrimSpace(selector.Space) != "" || strings.TrimSpace(selector.Chat) != "" || strings.TrimSpace(selector.Topic) != "" || strings.TrimSpace(selector.Sender) != "" || selector.Mentioned != nil
}

func canonicalDispatchSenderID(channel, rawID string, identityLinks map[string][]string) string {
	normalizedID := strings.TrimSpace(rawID)
	if normalizedID == "" {
		return ""
	}
	if linked := resolveLinkedDispatchID(identityLinks, channel, normalizedID); linked != "" {
		normalizedID = linked
	}
	return strings.ToLower(normalizedID)
}

func resolveLinkedDispatchID(identityLinks map[string][]string, channel, peerID string) string {
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
