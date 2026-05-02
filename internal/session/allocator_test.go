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
