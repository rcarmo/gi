package routing

import "testing"

func TestResolveRouteUsesDispatchRule(t *testing.T) {
	mentioned := true
	resolver := NewRouteResolver(
		AgentsConfig{
			List: []AgentConfig{{ID: "main", Default: true}, {ID: "support"}},
			Dispatch: &DispatchConfig{Rules: []DispatchRule{{
				Name:  "mentions-support",
				Agent: "support",
				When:  DispatchSelector{Channel: "gi", Mentioned: &mentioned},
			}}},
		},
		SessionConfig{Dimensions: []string{"chat"}},
	)
	resolved := resolver.ResolveRoute(InboundContext{Channel: "gi", Account: "default", ChatType: "direct", ChatID: "chat-1", Mentioned: true})
	if resolved.AgentID != "support" {
		t.Fatalf("expected support, got %s", resolved.AgentID)
	}
	if resolved.MatchedBy != "dispatch.rule:mentions-support" {
		t.Fatalf("unexpected matchedBy: %s", resolved.MatchedBy)
	}
}

func TestResolveRouteFallsBackToDefaultAgent(t *testing.T) {
	resolver := NewRouteResolver(AgentsConfig{List: []AgentConfig{{ID: "main", Default: true}, {ID: "support"}}}, SessionConfig{Dimensions: []string{"chat"}})
	resolved := resolver.ResolveRoute(InboundContext{Channel: "gi", Account: "default", ChatType: "direct", ChatID: "chat-1"})
	if resolved.AgentID != "main" || resolved.MatchedBy != "default" {
		t.Fatalf("unexpected route: %#v", resolved)
	}
}

func TestNormalizeSessionDimensionsFiltersInvalidEntries(t *testing.T) {
	out := normalizeSessionDimensions([]string{"chat", "sender", "bogus", "CHAT", "topic"})
	if len(out) != 3 || out[0] != "chat" || out[1] != "sender" || out[2] != "topic" {
		t.Fatalf("unexpected dimensions: %#v", out)
	}
}
