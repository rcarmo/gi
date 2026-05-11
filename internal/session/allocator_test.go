package session

import (
	"testing"

	"github.com/rcarmo/gi/internal/routing"
)

func TestAllocateRouteSessionBuildsScopedKey(t *testing.T) {
	alloc := AllocateRouteSession(AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "gi", Account: "default", ChatType: "direct", ChatID: "chat-1", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}},
	})
	if alloc.Scope.AgentID != "support" || alloc.Scope.Values["chat"] != "direct:chat-1" || alloc.Scope.Values["sender"] != "rui" {
		t.Fatalf("unexpected allocation scope: %#v", alloc.Scope)
	}
	if alloc.SessionKey == "" || len(alloc.SessionAliases) == 0 {
		t.Fatalf("expected session key and aliases, got %#v", alloc)
	}
}

func TestNormalizeAllocationIdentityLinksCanonicalizesSenderAndDerivedKey(t *testing.T) {
	alloc := Allocation{
		Scope: SessionScope{
			Version:    ScopeVersionV1,
			AgentID:    "support",
			Channel:    "slack",
			Account:    "workspace",
			Dimensions: []string{"chat", "sender"},
			Values: map[string]string{
				"chat":   "group:thread-7",
				"sender": "slack:ruicarmo",
			},
		},
		IdentityLinks: map[string][]string{"rui": {"slack:ruicarmo"}},
	}
	alloc.SessionKey = BuildSessionKey(alloc.Scope)
	alloc.SessionAliases = []string{"agent:support:slack:chat:group:thread-7:sender:slack:ruicarmo", "slack:group:thread-7"}
	normalized := NormalizeAllocationIdentityLinks(alloc)
	if normalized.Scope.Values["sender"] != "rui" {
		t.Fatalf("expected canonical sender identity, got %#v", normalized.Scope)
	}
	if normalized.SessionKey != BuildSessionKey(normalized.Scope) {
		t.Fatalf("expected derived session key to be recomputed, got %#v", normalized)
	}
	foundCanonicalAlias := false
	for _, alias := range normalized.SessionAliases {
		if alias == "agent:support:slack:chat:group:thread-7:sender:rui" {
			foundCanonicalAlias = true
		}
	}
	if !foundCanonicalAlias {
		t.Fatalf("expected canonical sender alias in %#v", normalized.SessionAliases)
	}
}

func TestAllocateDefaultSessionUsesRouteCompatibleChatAlias(t *testing.T) {
	alloc := AllocateDefaultSession("support", "gi", "default", "chat-1")
	if alloc.Scope.Values["chat"] != "direct:chat-1" {
		t.Fatalf("expected direct-scoped default chat value, got %#v", alloc.Scope)
	}
	foundCanonicalAlias := false
	foundChannelAlias := false
	for _, alias := range alloc.SessionAliases {
		if alias == "agent:support:gi:chat:direct:chat-1" {
			foundCanonicalAlias = true
		}
		if alias == "gi:direct:chat-1" {
			foundChannelAlias = true
		}
	}
	if !foundCanonicalAlias || !foundChannelAlias {
		t.Fatalf("expected route-compatible aliases, got %#v", alloc.SessionAliases)
	}
}
