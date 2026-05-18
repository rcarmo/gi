package routing

import (
	"fmt"
	"strings"
)

func PreparePromptRoutedInput(prompt string, inbound InboundContext, resolver *RouteResolver) (ResolvedRoute, string, bool, error) {
	targetAgentID, body, directed := ParseDirectedPrompt(prompt)
	promptBody := prompt
	mentioned := false
	if directed {
		if body == "" {
			return ResolvedRoute{}, "", false, fmt.Errorf("directed prompt requires content after @%s", targetAgentID)
		}
		promptBody = body
		mentioned = true
	}
	inbound.Mentioned = mentioned
	inbound.Prompt = promptBody
	route := resolver.ResolveRoute(inbound)
	if directed && targetAgentID != "" {
		route.AgentID = NormalizeAgentID(targetAgentID)
		route.MatchedBy = "mention"
	}
	return route, promptBody, directed, nil
}

func PreparePeerRoutedInput(targetAgentID, matchedBy, content string, inbound InboundContext, resolver *RouteResolver) ResolvedRoute {
	inbound.Mentioned = true
	inbound.Prompt = content
	route := resolver.ResolveRoute(inbound)
	route.AgentID = NormalizeAgentID(targetAgentID)
	route.MatchedBy = matchedBy
	return route
}

func ApplyPromptRouteMetadata(metadata map[string]any, sourceSessionID, targetSessionID, sourceAgentID string, route ResolvedRoute, created bool) map[string]any {
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["route_mode"] = "prompt"
	metadata["route_matched_by"] = route.MatchedBy
	metadata["target_agent_id"] = route.AgentID
	metadata["target_session_id"] = targetSessionID
	metadata["source_agent_id"] = sourceAgentID
	if route.MatchedBy != "" {
		metadata["routing_policy"] = route.MatchedBy
	}
	metadata["requested_agent_id"] = route.AgentID
	metadata["source_session_id"] = sourceSessionID
	metadata["route_created_session"] = created
	metadata["routing_enabled"] = true
	return metadata
}

func ModelForAgent(agentID string, agents AgentsConfig, defaultModel string) string {
	agentID = NormalizeAgentID(agentID)
	for _, agent := range agents.List {
		if NormalizeAgentID(agent.ID) == agentID {
			if strings.TrimSpace(agent.Model) != "" {
				return agent.Model
			}
		}
	}
	if strings.TrimSpace(defaultModel) != "" {
		return defaultModel
	}
	return "bootstrap"
}
