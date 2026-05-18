package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/routing"
	gisession "github.com/rcarmo/gi/internal/session"
)

func TestStoreSessionTurnMessageFlow(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	session, err := s.CreateSession(ctx, "session_1", "Test", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if session.State["model"] != "bootstrap" {
		t.Fatalf("unexpected session state: %#v", session.State)
	}
	if session.Scope == nil || session.Scope.AgentID != "gi" || session.Scope.Values["chat"] != "direct:session_1" {
		t.Fatalf("unexpected session scope: %#v", session.Scope)
	}
	if len(session.Aliases) == 0 {
		t.Fatalf("expected session aliases, got %#v", session.Aliases)
	}

	turn, err := s.CreateTurn(ctx, "turn_1", session.ID, "hello", map[string]any{"intent": "prompt"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if turn.Status != "running" {
		t.Fatalf("unexpected turn status: %s", turn.Status)
	}

	if err := s.AppendTurnEvent(ctx, turn.ID, session.ID, "turn.started", map[string]any{"phase": "turn", "checkpoint": true}); err != nil {
		t.Fatalf("append turn event: %v", err)
	}
	if err := s.AddMessage(ctx, "msg_1", session.ID, "user", "hello", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("add message: %v", err)
	}

	events, err := s.ListTurnEvents(ctx, turn.ID)
	if err != nil {
		t.Fatalf("list turn events: %v", err)
	}
	if len(events) != 1 || events[0].Payload["phase"] != "turn" {
		t.Fatalf("unexpected events: %#v", events)
	}

	msgs, err := s.ListMessages(ctx, session.ID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Content != "hello" {
		t.Fatalf("unexpected messages: %#v", msgs)
	}
}

func TestStoreResolvesSessionByOpaqueKeyAndAlias(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	alloc := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "gi", Account: "default", ChatType: "direct", ChatID: "chat-1", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}},
	})
	sess, err := s.CreateSessionWithMetadata(ctx, "session_route", "", "@support", map[string]any{"status": "idle"}, &alloc.Scope, alloc.SessionAliases)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	byKey, err := s.GetSessionByOpaqueKey(ctx, alloc.SessionKey)
	if err != nil {
		t.Fatalf("get session by opaque key: %v", err)
	}
	if byKey.ID != sess.ID {
		t.Fatalf("unexpected session by key: %#v", byKey)
	}

	byAlias, err := s.GetSessionByAlias(ctx, alloc.SessionAliases[0])
	if err != nil {
		t.Fatalf("get session by alias: %v", err)
	}
	if byAlias.ID != sess.ID {
		t.Fatalf("unexpected session by alias: %#v", byAlias)
	}

	byAlloc, err := s.FindSessionByAllocation(ctx, alloc)
	if err != nil {
		t.Fatalf("find session by allocation: %v", err)
	}
	if byAlloc.ID != sess.ID {
		t.Fatalf("unexpected allocation resolution: %#v", byAlloc)
	}
}

func TestStoreFindSessionByAllocationUsesSessionIdentityInsteadOfSessionScopeJSON(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "slack", Account: "workspace", ChatType: "group", ChatID: "thread-7", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}},
	})
	sess, err := s.CreateSessionWithMetadata(ctx, "session_route_identity_truth", "", "@support", map[string]any{"status": "idle"}, &alloc.Scope, alloc.SessionAliases)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update sessions set scope_json = ? where id = ?`, `{"version":1,"agent_id":"wrong","channel":"wrong","account":"wrong","dimensions":["chat"],"values":{"chat":"direct:wrong"}}`, sess.ID); err != nil {
		t.Fatalf("mutate legacy scope json: %v", err)
	}
	byAlloc, err := s.FindSessionByAllocation(ctx, alloc)
	if err != nil {
		t.Fatalf("find session by allocation after stale scope mutation: %v", err)
	}
	if byAlloc.ID != sess.ID {
		t.Fatalf("unexpected allocation resolution after stale scope mutation: %#v", byAlloc)
	}
	loaded, err := s.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("reload session: %v", err)
	}
	if loaded.Scope == nil || loaded.Scope.AgentID != "wrong" {
		t.Fatalf("expected reloaded session scope json to be stale test fixture, got %#v", loaded.Scope)
	}
}

func TestStoreResolveSessionByKeyOrAlias(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_key_alias")
	sess, err := s.CreateSessionWithMetadata(ctx, "session_key_alias", "", "@agent", map[string]any{"status": "idle"}, &alloc.Scope, alloc.SessionAliases)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	for _, key := range []string{sess.ID, alloc.SessionKey, strings.ToUpper(alloc.SessionAliases[0])} {
		resolved, err := s.ResolveSessionByKeyOrAlias(ctx, key)
		if err != nil {
			t.Fatalf("resolve session by key or alias %q: %v", key, err)
		}
		if resolved.ID != sess.ID {
			t.Fatalf("unexpected resolution for %q: %#v", key, resolved)
		}
	}
}

func TestStoreGetSessionIdentity(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "slack", Account: "workspace", ChatType: "group", ChatID: "thread-7", SpaceType: "room", SpaceID: "eng", TopicID: "builds", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"space", "chat", "topic", "sender"}},
	})
	sess, err := s.CreateSessionWithMetadata(ctx, "session_identity", "", "@support", map[string]any{"status": "idle"}, &alloc.Scope, alloc.SessionAliases)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	identity, err := s.GetSessionIdentity(ctx, sess.ID)
	if err != nil {
		t.Fatalf("get session identity: %v", err)
	}
	if identity.SessionID != sess.ID || identity.Scope.AgentID != "support" || identity.Scope.Channel != "slack" || identity.Scope.Account != "workspace" {
		t.Fatalf("unexpected session identity header: %#v", identity)
	}
	if identity.Scope.Values["space"] != "room:eng" || identity.Scope.Values["chat"] != "group:thread-7" || identity.Scope.Values["topic"] != "topic:builds" || identity.Scope.Values["sender"] != "rui" {
		t.Fatalf("unexpected session identity scope values: %#v", identity.Scope)
	}
	if identity.CanonicalScopeSignature == "" || identity.OpaqueSessionKey == "" || len(identity.Aliases) == 0 {
		t.Fatalf("expected canonical identity metadata, got %#v", identity)
	}
}

func TestStoreListSessionIdentities(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	allocA := gisession.AllocateDefaultSession("agent", "gi", "default", "session_list_a")
	if _, err := s.CreateSessionWithMetadata(ctx, "session_list_a", "", "@agent", map[string]any{"status": "idle"}, &allocA.Scope, allocA.SessionAliases); err != nil {
		t.Fatalf("create session a: %v", err)
	}
	allocB := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "agent1",
		Context:       routing.InboundContext{Channel: "slack", Account: "workspace", ChatType: "group", ChatID: "thread-7", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}},
	})
	if _, err := s.CreateSessionWithMetadata(ctx, "session_list_b", "session_list_a", "@agent1", map[string]any{"status": "idle"}, &allocB.Scope, allocB.SessionAliases); err != nil {
		t.Fatalf("create session b: %v", err)
	}
	identities, err := s.ListSessionIdentities(ctx)
	if err != nil {
		t.Fatalf("list session identities: %v", err)
	}
	found := map[string]SessionIdentity{}
	for _, identity := range identities {
		found[identity.SessionID] = identity
	}
	if found["session_list_a"].Scope.AgentID != "agent" || found["session_list_b"].Scope.AgentID != "agent1" {
		t.Fatalf("unexpected identities: %#v", found)
	}
	if found["session_list_b"].Scope.Values["chat"] != "group:thread-7" || found["session_list_b"].Scope.Values["sender"] != "rui" {
		t.Fatalf("unexpected identity dimensions: %#v", found["session_list_b"])
	}
	if len(found["session_list_a"].Aliases) == 0 || len(found["session_list_b"].Aliases) == 0 {
		t.Fatalf("expected aliases in listed identities, got %#v", found)
	}
}

func TestStoreAliasManagementAPIs(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_alias_api")
	sess, err := s.CreateSessionWithMetadata(ctx, "session_alias_api", "", "@agent", map[string]any{"status": "idle", "model": "bootstrap"}, &alloc.Scope, alloc.SessionAliases)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	aliases, err := s.ListSessionAliases(ctx, sess.ID)
	if err != nil {
		t.Fatalf("list session aliases: %v", err)
	}
	if len(aliases) != len(alloc.SessionAliases) {
		t.Fatalf("unexpected initial aliases: %#v", aliases)
	}
	resolved, err := s.ResolveSessionByAlias(ctx, aliases[0])
	if err != nil {
		t.Fatalf("resolve session by alias: %v", err)
	}
	if resolved.ID != sess.ID {
		t.Fatalf("unexpected alias resolution: %#v", resolved)
	}
	updatedAliases := []string{"agent:agent:gi:chat:direct:session_alias_api", "custom:Team-Alpha", "custom:team-alpha"}
	if err := s.UpdateSessionAliases(ctx, sess.ID, updatedAliases); err != nil {
		t.Fatalf("update session aliases: %v", err)
	}
	aliases, err = s.ListSessionAliases(ctx, sess.ID)
	if err != nil {
		t.Fatalf("list updated aliases: %v", err)
	}
	if len(aliases) != 2 || aliases[0] != "agent:agent:gi:chat:direct:session_alias_api" || aliases[1] != "custom:team-alpha" {
		t.Fatalf("unexpected updated aliases: %#v", aliases)
	}
	resolved, err = s.ResolveSessionByAlias(ctx, "custom:TEAM-alpha")
	if err != nil {
		t.Fatalf("resolve updated alias: %v", err)
	}
	if resolved.ID != sess.ID {
		t.Fatalf("unexpected updated alias resolution: %#v", resolved)
	}
	reloaded, err := s.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("reload session: %v", err)
	}
	if reloaded.Title != "@agent" || reloaded.State["status"] != "idle" || reloaded.State["model"] != "bootstrap" {
		t.Fatalf("unexpected session after alias update: %#v", reloaded)
	}
	if reloaded.Scope == nil || reloaded.Scope.AgentID != "agent" || reloaded.Scope.Values["chat"] != "direct:session_alias_api" {
		t.Fatalf("unexpected scope after alias update: %#v", reloaded.Scope)
	}
	if len(reloaded.Aliases) != 2 || reloaded.Aliases[1] != "custom:team-alpha" {
		t.Fatalf("unexpected stored aliases after update: %#v", reloaded.Aliases)
	}
}

func TestStoreResolveOrCreateSessionFromAllocationPreservesExplicitSessionKey(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "slack", Account: "workspace", ChatType: "group", ChatID: "thread-7", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}},
	})
	derivedKey := alloc.SessionKey
	explicitKey := gisession.BuildOpaqueSessionKey("explicit:session:key")
	if explicitKey == derivedKey {
		t.Fatalf("expected explicit key to differ from derived key: %q", explicitKey)
	}
	alloc.SessionKey = explicitKey
	sess, created, err := s.ResolveOrCreateSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: "session_alloc_explicit", Title: "@support", State: map[string]any{"status": "idle"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("resolve or create session from allocation with explicit key: %v", err)
	}
	if !created {
		t.Fatalf("expected session creation for explicit key path")
	}
	identity, err := s.GetSessionIdentity(ctx, sess.ID)
	if err != nil {
		t.Fatalf("get session identity: %v", err)
	}
	if identity.OpaqueSessionKey != explicitKey {
		t.Fatalf("expected explicit opaque key %q, got %#v", explicitKey, identity)
	}
	resolved, err := s.ResolveSessionByCanonicalKey(ctx, explicitKey)
	if err != nil {
		t.Fatalf("resolve session by explicit canonical key: %v", err)
	}
	if resolved.ID != sess.ID {
		t.Fatalf("unexpected explicit-key resolution: %#v", resolved)
	}
	reused, created, err := s.ResolveOrCreateSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: "session_alloc_explicit_other", Title: "@support2", State: map[string]any{"status": "other"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("resolve existing explicit-key allocation session: %v", err)
	}
	if created || reused.ID != sess.ID {
		t.Fatalf("expected existing explicit-key allocation session reuse, got session=%#v created=%v", reused, created)
	}
}

func TestStoreResolveOrCreateSessionFromAllocationConcurrentSameScope(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "slack", Account: "workspace", ChatType: "group", ChatID: "thread-7", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}},
	})
	const workers = 12
	results := make(chan string, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sess, _, err := s.ResolveOrCreateSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: fmt.Sprintf("session_alloc_concurrent_%d", i), Title: "@support", State: map[string]any{"status": "idle"}, Allocation: alloc})
			if err != nil {
				errs <- err
				return
			}
			results <- sess.ID
		}(i)
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent resolve-or-create: %v", err)
		}
	}
	unique := map[string]bool{}
	for id := range results {
		unique[id] = true
	}
	if len(unique) != 1 {
		t.Fatalf("expected one session from concurrent resolve-or-create, got %#v", unique)
	}
	sessions, err := s.ListSessions(ctx)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected exactly one persisted session, got %#v", sessions)
	}
}

func TestStoreResolveOrCreateSessionFromAllocation(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "slack", Account: "workspace", ChatType: "group", ChatID: "thread-7", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}},
	})
	created, wasCreated, err := s.ResolveOrCreateSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: "session_alloc_api", Title: "@support", State: map[string]any{"status": "idle"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("resolve or create session from allocation: %v", err)
	}
	if !wasCreated || created.ID != "session_alloc_api" {
		t.Fatalf("expected created allocation session, got session=%#v created=%v", created, wasCreated)
	}
	reused, wasCreated, err := s.ResolveOrCreateSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: "session_alloc_api_other", Title: "@support2", State: map[string]any{"status": "other"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("resolve existing allocation session: %v", err)
	}
	if wasCreated || reused.ID != created.ID {
		t.Fatalf("expected existing allocation session reuse, got session=%#v created=%v", reused, wasCreated)
	}
}

func TestStoreResolveSessionByAllocationCollapsesIdentityLinksAtStoreBoundary(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	canonicalAlloc := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "slack", Account: "workspace", ChatType: "group", ChatID: "thread-7", SenderID: "ruicarmo"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}, IdentityLinks: map[string][]string{"rui": {"slack:ruicarmo", "ruicarmo"}}},
	})
	sess, err := s.CreateSessionWithMetadata(ctx, "session_identity_link", "", "@support", map[string]any{"status": "idle"}, &canonicalAlloc.Scope, canonicalAlloc.SessionAliases)
	if err != nil {
		t.Fatalf("create canonical linked session: %v", err)
	}
	staleAlloc := gisession.Allocation{
		Scope: gisession.SessionScope{
			Version:    gisession.ScopeVersionV1,
			AgentID:    "support",
			Channel:    "slack",
			Account:    "workspace",
			Dimensions: []string{"chat", "sender"},
			Values:     map[string]string{"chat": "group:thread-7", "sender": "slack:ruicarmo"},
		},
		SessionKey:     gisession.BuildSessionKey(gisession.SessionScope{Version: gisession.ScopeVersionV1, AgentID: "support", Channel: "slack", Account: "workspace", Dimensions: []string{"chat", "sender"}, Values: map[string]string{"chat": "group:thread-7", "sender": "slack:ruicarmo"}}),
		SessionAliases: []string{"agent:support:slack:chat:group:thread-7:sender:slack:ruicarmo", "slack:group:thread-7"},
		IdentityLinks:  map[string][]string{"rui": {"slack:ruicarmo", "ruicarmo"}},
	}
	resolved, err := s.ResolveSessionByAllocation(ctx, staleAlloc)
	if err != nil {
		t.Fatalf("resolve allocation with store-boundary identity-link collapse: %v", err)
	}
	if resolved.ID != sess.ID {
		t.Fatalf("expected identity-link allocation collapse to reuse canonical session, got %#v", resolved)
	}
}

func TestStoreSetAndResolveMainSession(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	allocA := gisession.AllocateDefaultSession("agent", "gi", "default", "session_main_a")
	sessA, err := s.CreateSessionWithMetadata(ctx, "session_main_a", "", "@agent", map[string]any{"status": "idle"}, &allocA.Scope, allocA.SessionAliases)
	if err != nil {
		t.Fatalf("create session a: %v", err)
	}
	allocB := gisession.AllocateDefaultSession("agent", "gi", "default", "session_main_b")
	sessB, err := s.CreateSessionWithMetadata(ctx, "session_main_b", "", "@agent", map[string]any{"status": "idle"}, &allocB.Scope, allocB.SessionAliases)
	if err != nil {
		t.Fatalf("create session b: %v", err)
	}
	if err := s.SetMainSession(ctx, sessA.ID); err != nil {
		t.Fatalf("set main session a: %v", err)
	}
	mainSess, err := s.ResolveMainSession(ctx, "agent", "gi", "default")
	if err != nil {
		t.Fatalf("resolve main session a: %v", err)
	}
	if mainSess.ID != sessA.ID {
		t.Fatalf("expected session a as main, got %#v", mainSess)
	}
	if err := s.SetMainSession(ctx, sessB.ID); err != nil {
		t.Fatalf("set main session b: %v", err)
	}
	mainSess, err = s.ResolveMainSession(ctx, "agent", "gi", "default")
	if err != nil {
		t.Fatalf("resolve main session b: %v", err)
	}
	if mainSess.ID != sessB.ID {
		t.Fatalf("expected session b as main, got %#v", mainSess)
	}
	identityA, err := s.GetSessionIdentity(ctx, sessA.ID)
	if err != nil {
		t.Fatalf("get session a identity: %v", err)
	}
	identityB, err := s.GetSessionIdentity(ctx, sessB.ID)
	if err != nil {
		t.Fatalf("get session b identity: %v", err)
	}
	if identityA.IsMainSession || !identityB.IsMainSession {
		t.Fatalf("expected only session b to be main, got a=%v b=%v", identityA.IsMainSession, identityB.IsMainSession)
	}
}

func TestStoreFindSiblingChildSessionByParentAndAgentUsesDirectLookup(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	rootAlloc := gisession.AllocateDefaultSession("agent", "gi", "default", "root_parent_lookup")
	root, err := s.CreateSessionWithMetadata(ctx, "root_parent_lookup", "", "@agent", map[string]any{"status": "idle"}, &rootAlloc.Scope, rootAlloc.SessionAliases)
	if err != nil {
		t.Fatalf("create root: %v", err)
	}
	childAAlloc := gisession.AllocateDefaultSession("agentA", "gi", "default", "child_lookup_a")
	childA, err := s.CreateSessionWithMetadata(ctx, "child_lookup_a", root.ID, "@agentA", map[string]any{"status": "idle"}, &childAAlloc.Scope, childAAlloc.SessionAliases)
	if err != nil {
		t.Fatalf("create child a: %v", err)
	}
	childBAlloc := gisession.AllocateDefaultSession("agentB", "gi", "default", "child_lookup_b")
	childB, err := s.CreateSessionWithMetadata(ctx, "child_lookup_b", root.ID, "@agentB", map[string]any{"status": "idle"}, &childBAlloc.Scope, childBAlloc.SessionAliases)
	if err != nil {
		t.Fatalf("create child b: %v", err)
	}
	resolved, err := s.FindSiblingChildSessionByParentAndAgent(ctx, childB.ID, "agentA")
	if err != nil {
		t.Fatalf("find sibling child by parent and agent: %v", err)
	}
	if resolved.ID != childA.ID {
		t.Fatalf("expected agentA sibling child, got %#v", resolved)
	}
}

func TestStoreResolveOrCreateMainSessionFromAllocation(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateDefaultSession("agent", "gi", "default", "session_main_create")
	sess, created, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: "session_main_create", Title: "@agent", State: map[string]any{"status": "idle"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("resolve or create main session: %v", err)
	}
	if !created {
		t.Fatalf("expected session creation for main session")
	}
	mainSess, err := s.ResolveMainSession(ctx, "agent", "gi", "default")
	if err != nil {
		t.Fatalf("resolve main session: %v", err)
	}
	if mainSess.ID != sess.ID {
		t.Fatalf("expected created session to be main, got %#v", mainSess)
	}
	identity, err := s.GetSessionIdentity(ctx, sess.ID)
	if err != nil {
		t.Fatalf("get created session identity: %v", err)
	}
	if !identity.IsMainSession {
		t.Fatalf("expected created session to be marked main: %#v", identity)
	}
}

func TestStoreResolveOrCreateMainSessionRegistersPrimaryChannelBinding(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "slack", Account: "workspace", ChatType: "group", ChatID: "thread-7", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}},
	})
	sess, created, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: "session_binding_main", Title: "@support", State: map[string]any{"status": "idle"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("resolve or create main session: %v", err)
	}
	if !created {
		t.Fatalf("expected created main session")
	}
	bindings, err := s.ListSessionChannelBindings(ctx, sess.ID)
	if err != nil {
		t.Fatalf("list session channel bindings: %v", err)
	}
	if len(bindings) != 1 {
		t.Fatalf("expected one primary channel binding, got %#v", bindings)
	}
	if bindings[0].Channel != "slack" || bindings[0].Account != "workspace" || bindings[0].BindingType != "chat" || bindings[0].RemoteIdentity != "group:thread-7" {
		t.Fatalf("unexpected primary channel binding: %#v", bindings[0])
	}
	resolved, err := s.ResolveSessionByChannelBinding(ctx, "slack", "workspace", "group:thread-7")
	if err != nil {
		t.Fatalf("resolve session by primary channel binding: %v", err)
	}
	if resolved.ID != sess.ID {
		t.Fatalf("unexpected binding resolution: %#v", resolved)
	}
}

func TestStoreResolveOrCreateSessionFromAllocationContinuesSessionAcrossChannels(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	baseAlloc := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "slack", Account: "workspace", ChatType: "group", ChatID: "thread-7", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}},
	})
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: "session_continue_xchan", Title: "@support", State: map[string]any{"status": "idle"}, Allocation: baseAlloc})
	if err != nil {
		t.Fatalf("resolve or create base main session: %v", err)
	}
	altAlloc := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "discord", Account: "guild", ChatType: "direct", ChatID: "user-42", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}},
	})
	continued, created, err := s.ResolveOrCreateSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: "session_continue_xchan_other", Title: "@support2", State: map[string]any{"status": "other"}, Allocation: altAlloc, ContinueSessionID: sess.ID})
	if err != nil {
		t.Fatalf("continue session across channels: %v", err)
	}
	if created || continued.ID != sess.ID {
		t.Fatalf("expected explicit cross-channel continuation to reuse same session, got session=%#v created=%v", continued, created)
	}
	bindings, err := s.ListSessionChannelBindings(ctx, sess.ID)
	if err != nil {
		t.Fatalf("list continued session bindings: %v", err)
	}
	if len(bindings) < 2 {
		t.Fatalf("expected alternate channel binding to be attached, got %#v", bindings)
	}
	resolved, err := s.ResolveSessionByAllocation(ctx, altAlloc)
	if err != nil {
		t.Fatalf("resolve alternate allocation after continuation: %v", err)
	}
	if resolved.ID != sess.ID {
		t.Fatalf("expected alternate allocation to reuse continued session, got %#v", resolved)
	}
}

func TestStoreResolveSessionByAllocationUsesAlternateChannelBinding(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "slack", Account: "workspace", ChatType: "group", ChatID: "thread-7", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}},
	})
	sess, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: "session_binding_alt", Title: "@support", State: map[string]any{"status": "idle"}, Allocation: alloc})
	if err != nil {
		t.Fatalf("resolve or create main session: %v", err)
	}
	if err := s.UpsertSessionChannelBinding(ctx, SessionChannelBinding{SessionID: sess.ID, Channel: "discord", Account: "guild", BindingType: "chat", RemoteIdentity: "direct:user-42", Metadata: map[string]any{"agent_id": "support"}}); err != nil {
		t.Fatalf("upsert alternate channel binding: %v", err)
	}
	altAlloc := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "discord", Account: "guild", ChatType: "direct", ChatID: "user-42", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}},
	})
	resolved, err := s.ResolveSessionByAllocation(ctx, altAlloc)
	if err != nil {
		t.Fatalf("resolve allocation by alternate channel binding: %v", err)
	}
	if resolved.ID != sess.ID {
		t.Fatalf("expected alternate channel binding to reuse same session, got %#v", resolved)
	}
}

func TestStoreResolveOrCreateSessionFromAllocationIsolatesForumTopics(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	allocA := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "discord", Account: "guild", ChatType: "forum", ChatID: "support", TopicID: "topic-a", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "topic", "sender"}},
	})
	sessA, createdA, err := s.ResolveOrCreateSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: "session_forum_topic_a", Title: "@support", State: map[string]any{"status": "idle"}, Allocation: allocA})
	if err != nil {
		t.Fatalf("resolve forum topic a: %v", err)
	}
	if !createdA {
		t.Fatalf("expected topic a session creation")
	}
	allocAAgain := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "discord", Account: "guild", ChatType: "forum", ChatID: "support", TopicID: "topic-a", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "topic", "sender"}},
	})
	reusedA, createdAAgain, err := s.ResolveOrCreateSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: "session_forum_topic_a_reuse", Title: "@support", State: map[string]any{"status": "idle"}, Allocation: allocAAgain})
	if err != nil {
		t.Fatalf("resolve forum topic a reuse: %v", err)
	}
	if createdAAgain || reusedA.ID != sessA.ID {
		t.Fatalf("expected same topic to reuse same session, got session=%#v created=%v", reusedA, createdAAgain)
	}
	allocB := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "discord", Account: "guild", ChatType: "forum", ChatID: "support", TopicID: "topic-b", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "topic", "sender"}},
	})
	sessB, createdB, err := s.ResolveOrCreateSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: "session_forum_topic_b", Title: "@support", State: map[string]any{"status": "idle"}, Allocation: allocB})
	if err != nil {
		t.Fatalf("resolve forum topic b: %v", err)
	}
	if !createdB || sessB.ID == sessA.ID {
		t.Fatalf("expected different forum topic to isolate session, got a=%#v b=%#v created=%v", sessA, sessB, createdB)
	}
	chatOnlyA := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "discord", Account: "guild", ChatType: "forum", ChatID: "support", TopicID: "topic-x", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}},
	})
	chatOnlySessA, createdChatOnlyA, err := s.ResolveOrCreateSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: "session_forum_chat_only_a", Title: "@support", State: map[string]any{"status": "idle"}, Allocation: chatOnlyA})
	if err != nil {
		t.Fatalf("resolve forum chat-only topic x: %v", err)
	}
	if !createdChatOnlyA {
		t.Fatalf("expected chat-only session creation")
	}
	chatOnlyB := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "discord", Account: "guild", ChatType: "forum", ChatID: "support", TopicID: "topic-y", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}},
	})
	chatOnlySessB, createdChatOnlyB, err := s.ResolveOrCreateSessionFromAllocation(ctx, ResolveOrCreateSessionFromAllocationInput{ID: "session_forum_chat_only_b", Title: "@support", State: map[string]any{"status": "idle"}, Allocation: chatOnlyB})
	if err != nil {
		t.Fatalf("resolve forum chat-only topic y: %v", err)
	}
	if createdChatOnlyB || chatOnlySessB.ID != chatOnlySessA.ID {
		t.Fatalf("expected chat-only policy to reuse same forum session, got a=%#v b=%#v created=%v", chatOnlySessA, chatOnlySessB, createdChatOnlyB)
	}
}

func TestStoreFindSessionByAllocationFallsBackToCanonicalSignature(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	alloc := gisession.AllocateRouteSession(gisession.AllocationInput{
		AgentID:       "support",
		Context:       routing.InboundContext{Channel: "gi", Account: "default", ChatType: "direct", ChatID: "chat-1", SenderID: "rui"},
		SessionPolicy: routing.SessionPolicy{Dimensions: []string{"chat", "sender"}},
	})
	sess, err := s.CreateSessionWithMetadata(ctx, "session_route_sig", "", "@support", map[string]any{"status": "idle"}, &alloc.Scope, alloc.SessionAliases)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update session_identities set opaque_session_key = ? where session_id = ?`, "other-key", sess.ID); err != nil {
		t.Fatalf("mutate opaque key: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `delete from session_aliases where session_id = ?`, sess.ID); err != nil {
		t.Fatalf("delete aliases: %v", err)
	}
	found, err := s.FindSessionByAllocation(ctx, alloc)
	if err != nil {
		t.Fatalf("find session by allocation with signature fallback: %v", err)
	}
	if found.ID != sess.ID {
		t.Fatalf("unexpected signature fallback session: %#v", found)
	}
}

func TestStoreRecordsHookInvocations(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	session, err := s.CreateSession(ctx, "session_hook_inv", "HookInv", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	turn, err := s.CreateTurnWithStatus(ctx, "turn_hook_inv", session.ID, "running", "hello", map[string]any{"intent": "prompt"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	id, err := s.RecordHookInvocation(ctx, turn.ID, session.ID, "tool_call", "tool_call", "test-hook", "modify", map[string]any{"trace": map[string]any{"id": "hook_1"}}, map[string]any{"action": "modify"}, "", 17)
	if err != nil {
		t.Fatalf("record hook invocation: %v", err)
	}
	got, err := s.GetHookInvocation(ctx, id)
	if err != nil {
		t.Fatalf("get hook invocation: %v", err)
	}
	if got.HookName != "tool_call" || got.HookSource != "test-hook" || got.DurationMS != 17 {
		t.Fatalf("unexpected hook invocation: %#v", got)
	}
	if got.Request["trace"].(map[string]any)["id"] != "hook_1" {
		t.Fatalf("expected request trace in hook invocation: %#v", got)
	}
	items, err := s.ListHookInvocationsByTurn(ctx, turn.ID)
	if err != nil {
		t.Fatalf("list hook invocations by turn: %v", err)
	}
	if len(items) != 1 || items[0].ID != id {
		t.Fatalf("unexpected turn hook invocations: %#v", items)
	}
	sessionItems, err := s.ListHookInvocationsBySession(ctx, session.ID)
	if err != nil {
		t.Fatalf("list hook invocations by session: %v", err)
	}
	if len(sessionItems) != 1 || sessionItems[0].ID != id {
		t.Fatalf("unexpected session hook invocations: %#v", sessionItems)
	}
}

func TestStoreClaimsActiveTurnOncePerSession(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	sess, err := s.CreateSession(ctx, "session_claim", "Test", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_claim_1", sess.ID, "running", "hello", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create turn 1: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_claim_2", sess.ID, "queued", "hello again", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create turn 2: %v", err)
	}

	claimed, err := s.ClaimSessionActiveTurn(ctx, sess.ID, "turn_claim_1", "runner", "claim1")
	if err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	if !claimed {
		t.Fatal("expected first claim to succeed")
	}
	claimed, err = s.ClaimSessionActiveTurn(ctx, sess.ID, "turn_claim_2", "runner", "claim2")
	if err != nil {
		t.Fatalf("claim second active turn: %v", err)
	}
	if claimed {
		t.Fatal("expected second claim to fail while first is active")
	}
	if err := s.ReleaseSessionActiveTurn(ctx, sess.ID, "claim1"); err != nil {
		t.Fatalf("release active turn: %v", err)
	}
	if _, _, err := s.GetSessionActiveTurn(ctx, sess.ID); err == nil || err != sql.ErrNoRows {
		t.Fatalf("expected no active turn after release, got %v", err)
	}
}

func TestTouchSessionActiveTurnRejectsMissingClaim(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if err := s.TouchSessionActiveTurn(ctx, "missing-session", "missing-claim"); err == nil || !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected missing active turn touch to return sql.ErrNoRows, got %v", err)
	}
}

func TestTouchSessionStateRejectsMissingSession(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if err := s.TouchSessionState(ctx, "missing-session", map[string]any{"status": "running"}); err == nil || !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected missing session touch to return sql.ErrNoRows, got %v", err)
	}
}

func TestTurnLifecycleWritersRejectMissingTurn(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	checks := []struct {
		name string
		run  func() error
	}{
		{name: "update turn status and phase", run: func() error { return s.UpdateTurnStatusAndPhase(ctx, "missing-turn", "running", "setup") }},
		{name: "mark turn claimed", run: func() error { return s.MarkTurnClaimed(ctx, "missing-turn", "runner") }},
		{name: "reset turn claim", run: func() error { return s.ResetTurnClaim(ctx, "missing-turn") }},
		{name: "mark turn finished", run: func() error { return s.MarkTurnFinished(ctx, "missing-turn") }},
	}
	for _, check := range checks {
		if err := check.run(); err == nil || !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected %s to return sql.ErrNoRows, got %v", check.name, err)
		}
	}
}

func TestTurnFailureMarkersClearOnRequeueAndCompletion(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	sess, err := s.CreateSession(ctx, "session_failure", "Test", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_failure", sess.ID, "failed", "hello", map[string]any{"intent": "prompt"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if err := s.UpsertTurnFailure(ctx, turnRec.ID, sess.ID, "provider_error", "none", "provider failed"); err != nil {
		t.Fatalf("upsert failure marker: %v", err)
	}
	if _, err := s.GetTurnFailure(ctx, turnRec.ID); err != nil {
		t.Fatalf("get failure marker: %v", err)
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, turnRec.ID, "queued", "queued"); err != nil {
		t.Fatalf("requeue turn: %v", err)
	}
	if _, err := s.GetTurnFailure(ctx, turnRec.ID); err == nil {
		t.Fatal("expected requeue to clear failure marker")
	}
	if err := s.UpsertTurnFailure(ctx, turnRec.ID, sess.ID, "provider_error", "none", "provider failed again"); err != nil {
		t.Fatalf("upsert second failure marker: %v", err)
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, turnRec.ID, "completed", "completed"); err != nil {
		t.Fatalf("complete turn: %v", err)
	}
	if _, err := s.GetTurnFailure(ctx, turnRec.ID); err == nil {
		t.Fatal("expected completion to clear failure marker")
	}
}

func TestSteeringDequeueRespectsQueueMode(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	_, err = s.CreateSession(ctx, "session_steering", "Test", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_steering", "", "user", "one", map[string]any{"intent": "prompt"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering one: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_steering", "", "user", "two", map[string]any{"intent": "prompt"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering two: %v", err)
	}
	msgs, err := s.DequeueSteering(ctx, "session_steering")
	if err != nil {
		t.Fatalf("dequeue steering: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Content != "one" {
		t.Fatalf("expected one-at-a-time dequeue of first message, got %#v", msgs)
	}
	msgs, err = s.DequeueSteering(ctx, "session_steering")
	if err != nil {
		t.Fatalf("dequeue second steering: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Content != "two" {
		t.Fatalf("expected second queued steering message, got %#v", msgs)
	}
	if _, err := s.EnqueueSteering(ctx, "session_steering", "", "user", "all-1", map[string]any{"intent": "prompt"}, nil, "all"); err != nil {
		t.Fatalf("enqueue steering all-1: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_steering", "", "user", "all-2", map[string]any{"intent": "prompt"}, nil, "all"); err != nil {
		t.Fatalf("enqueue steering all-2: %v", err)
	}
	msgs, err = s.DequeueSteering(ctx, "session_steering")
	if err != nil {
		t.Fatalf("dequeue all steering: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected all-mode dequeue to drain both messages, got %#v", msgs)
	}
}

func TestSteeringQueueOverflowReturnsError(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	_, err = s.CreateSession(ctx, "session_steering_overflow", "Test", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	for i := 0; i < 10; i++ {
		if _, err := s.EnqueueSteering(ctx, "session_steering_overflow", "", "user", fmt.Sprintf("msg-%d", i), map[string]any{"intent": "prompt"}, nil, "one-at-a-time"); err != nil {
			t.Fatalf("enqueue steering %d: %v", i, err)
		}
	}
	if _, err := s.EnqueueSteering(ctx, "session_steering_overflow", "", "user", "overflow", map[string]any{"intent": "prompt"}, nil, "one-at-a-time"); err == nil {
		t.Fatal("expected steering queue overflow error")
	}
	if depth, err := s.SteeringQueueLength(ctx, "session_steering_overflow"); err != nil {
		t.Fatalf("steering queue length: %v", err)
	} else if depth != 10 {
		t.Fatalf("expected queue depth to stay capped at 10, got %d", depth)
	}
}

func TestChatVFSVirtualTreeAndMessageDocument(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	session, err := s.CreateSession(ctx, "session_chat_vfs", "Chat VFS", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := s.AddMessage(ctx, "msg_chat_vfs", session.ID, "user", "hello from chat vfs", map[string]any{"kind": "chat"}); err != nil {
		t.Fatalf("add message: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_chat_vfs", session.ID, "completed", "prompt text", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	rootEntries, err := s.ListVFSChildren(ctx, "chat", "")
	if err != nil {
		t.Fatalf("list chat root: %v", err)
	}
	hasSessionsDir := false
	for _, entry := range rootEntries {
		if entry.Path == "sessions" && entry.IsDir {
			hasSessionsDir = true
		}
	}
	if !hasSessionsDir {
		t.Fatalf("expected sessions dir in chat root: %#v", rootEntries)
	}
	_, raw, err := s.GetVFSFileContent(ctx, "chat", "sessions/session_chat_vfs/messages/msg_chat_vfs.md")
	if err != nil {
		t.Fatalf("get chat message doc: %v", err)
	}
	text := string(raw)
	if !strings.Contains(text, "kind: \"chat/message\"") || !strings.Contains(text, "hello from chat vfs") {
		t.Fatalf("unexpected chat vfs message doc: %q", text)
	}
	if _, err := s.SaveVFSFile(ctx, "chat", "sessions/session_chat_vfs/messages/evil.md", "text/plain", []byte("nope"), nil); err == nil {
		t.Fatal("expected read-only chat namespace write failure")
	}
}

func TestSubTurnLifecycle(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	parentSession, err := s.CreateSession(ctx, "session_parent", "Parent", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	childSession, err := s.CreateSession(ctx, "session_child", "Child", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create child session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_parent", parentSession.ID, "running", "parent", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_child", childSession.ID, "queued", "child", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create child turn: %v", err)
	}
	sub, err := s.CreateSubTurn(ctx, "turn_parent", parentSession.ID, "turn_child", childSession.ID, "sync", 1, map[string]any{"origin": "test"})
	if err != nil {
		t.Fatalf("create subturn: %v", err)
	}
	if sub.ParentTurnID != "turn_parent" || sub.ChildTurnID != "turn_child" || sub.Status != "running" {
		t.Fatalf("unexpected created subturn: %#v", sub)
	}
	runningCount, err := s.CountRunningSubTurnsByParent(ctx, "turn_parent")
	if err != nil {
		t.Fatalf("count running subturns: %v", err)
	}
	if runningCount != 1 {
		t.Fatalf("expected one running subturn, got %d", runningCount)
	}
	if err := s.UpdateSubTurnStatusByChild(ctx, "turn_child", "completed"); err != nil {
		t.Fatalf("update subturn status by child: %v", err)
	}
	stored, err := s.GetSubTurnByChild(ctx, "turn_child")
	if err != nil {
		t.Fatalf("get subturn by child: %v", err)
	}
	if stored.Status != "completed" || stored.FinishedAt == "" {
		t.Fatalf("unexpected updated subturn: %#v", stored)
	}
	runningCount, err = s.CountRunningSubTurnsByParent(ctx, "turn_parent")
	if err != nil {
		t.Fatalf("count running subturns after completion: %v", err)
	}
	if runningCount != 0 {
		t.Fatalf("expected no running subturns after completion, got %d", runningCount)
	}
	items, err := s.ListSubTurnsByParent(ctx, "turn_parent")
	if err != nil {
		t.Fatalf("list subturns by parent: %v", err)
	}
	if len(items) != 1 || items[0].ChildTurnID != "turn_child" {
		t.Fatalf("unexpected parent listing: %#v", items)
	}
}

func TestCreateSubTurnRejectsInvalidDeliveryMode(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_invalid_delivery", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_invalid_delivery", "Child", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_parent_invalid_delivery", "session_parent_invalid_delivery", "running", "parent", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_child_invalid_delivery", "session_child_invalid_delivery", "queued", "child", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create child turn: %v", err)
	}
	if _, err := s.CreateSubTurn(ctx, "turn_parent_invalid_delivery", "session_parent_invalid_delivery", "turn_child_invalid_delivery", "session_child_invalid_delivery", "fanout", 1, map[string]any{}); err == nil {
		t.Fatal("expected invalid subturn delivery mode error")
	}
}

func TestUpdateSubTurnMetadataByChild(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_metadata_patch", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_metadata_patch", "Child", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_parent_metadata_patch", "session_parent_metadata_patch", "running", "parent", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_child_metadata_patch", "session_child_metadata_patch", "running", "child", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create child turn: %v", err)
	}
	if _, err := s.CreateSubTurn(ctx, "turn_parent_metadata_patch", "session_parent_metadata_patch", "turn_child_metadata_patch", "session_child_metadata_patch", "async", 1, map[string]any{"origin": "test"}); err != nil {
		t.Fatalf("create subturn: %v", err)
	}
	if err := s.UpdateSubTurnMetadataByChild(ctx, "turn_child_metadata_patch", map[string]any{"orphaned": true, "orphan_reason": "test"}); err != nil {
		t.Fatalf("patch subturn metadata: %v", err)
	}
	sub, err := s.GetSubTurnByChild(ctx, "turn_child_metadata_patch")
	if err != nil {
		t.Fatalf("lookup patched subturn: %v", err)
	}
	if sub.Metadata["origin"] != "test" || sub.Metadata["orphaned"] != true || sub.Metadata["orphan_reason"] != "test" {
		t.Fatalf("unexpected patched metadata: %#v", sub.Metadata)
	}
}

func TestTouchSessionStateConcurrentPatchMerge(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	sess, err := s.CreateSession(ctx, "session_state_concurrent", "State", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	const workers = 16
	var wg sync.WaitGroup
	errCh := make(chan error, workers)
	for i := 0; i < workers; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := fmt.Sprintf("k%d", i)
			if err := s.TouchSessionState(ctx, sess.ID, map[string]any{key: i}); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("touch session state concurrent patch: %v", err)
		}
	}
	updated, err := s.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("get updated session: %v", err)
	}
	for i := 0; i < workers; i++ {
		key := fmt.Sprintf("k%d", i)
		if _, ok := updated.State[key]; !ok {
			t.Fatalf("missing merged key %q in session state: %#v", key, updated.State)
		}
	}
}

func TestUpdateSubTurnMetadataByChildConcurrentPatchMerge(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_parent_metadata_concurrent", "Parent", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create parent session: %v", err)
	}
	if _, err := s.CreateSession(ctx, "session_child_metadata_concurrent", "Child", map[string]any{"model": "bootstrap"}); err != nil {
		t.Fatalf("create child session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_parent_metadata_concurrent", "session_parent_metadata_concurrent", "running", "parent", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create parent turn: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_child_metadata_concurrent", "session_child_metadata_concurrent", "running", "child", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("create child turn: %v", err)
	}
	if _, err := s.CreateSubTurn(ctx, "turn_parent_metadata_concurrent", "session_parent_metadata_concurrent", "turn_child_metadata_concurrent", "session_child_metadata_concurrent", "async", 1, map[string]any{"origin": "test"}); err != nil {
		t.Fatalf("create subturn: %v", err)
	}
	const workers = 12
	var wg sync.WaitGroup
	errCh := make(chan error, workers)
	for i := 0; i < workers; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := fmt.Sprintf("patch_%d", i)
			if err := s.UpdateSubTurnMetadataByChild(ctx, "turn_child_metadata_concurrent", map[string]any{key: i}); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("update subturn metadata concurrent patch: %v", err)
		}
	}
	sub, err := s.GetSubTurnByChild(ctx, "turn_child_metadata_concurrent")
	if err != nil {
		t.Fatalf("get patched subturn: %v", err)
	}
	if sub.Metadata["origin"] != "test" {
		t.Fatalf("missing origin metadata: %#v", sub.Metadata)
	}
	for i := 0; i < workers; i++ {
		key := fmt.Sprintf("patch_%d", i)
		if _, ok := sub.Metadata[key]; !ok {
			t.Fatalf("missing merged subturn metadata key %q: %#v", key, sub.Metadata)
		}
	}
}

func TestHoldAndResolveTurnFailurePhase(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	sess, err := s.CreateSession(ctx, "session_hold", "Test", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_hold", sess.ID, "failed", "hello", map[string]any{"intent": "prompt"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if err := s.UpsertTurnFailure(ctx, turnRec.ID, sess.ID, "provider_error", "none", "provider failed"); err != nil {
		t.Fatalf("upsert failure marker: %v", err)
	}
	if err := s.HoldTurnFailure(ctx, turnRec.ID, "review", "needs choice"); err != nil {
		t.Fatalf("hold turn failure: %v", err)
	}
	heldTurn, err := s.GetTurn(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("get held turn: %v", err)
	}
	if heldTurn.Phase != "held_for_retry_or_skip" {
		t.Fatalf("expected held phase, got %#v", heldTurn)
	}
	failureRec, err := s.GetTurnFailure(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("get held failure row: %v", err)
	}
	if failureRec.HoldState != "review" {
		t.Fatalf("expected review hold state, got %#v", failureRec)
	}
	if err := s.ResolveTurnFailure(ctx, turnRec.ID, "skipped", "skip requested", ""); err != nil {
		t.Fatalf("resolve turn failure: %v", err)
	}
	resolvedTurn, err := s.GetTurn(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("get resolved turn: %v", err)
	}
	if resolvedTurn.Phase != "failed" {
		t.Fatalf("expected resolved phase to return to failed, got %#v", resolvedTurn)
	}
	failureRec, err = s.GetTurnFailure(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("get resolved failure row: %v", err)
	}
	if failureRec.HoldState != "none" || failureRec.ResolutionState != "skipped" {
		t.Fatalf("unexpected resolved failure row: %#v", failureRec)
	}
}

func TestStoreInboundWorkQueueConcurrentClaimSingleWinner(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	queued, err := s.EnqueueInboundWork(ctx, "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "hello"})
	if err != nil {
		t.Fatalf("enqueue inbound work: %v", err)
	}
	const workers = 8
	claimedIDs := make(chan int64, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			item, err := s.ClaimNextInboundWork(ctx, fmt.Sprintf("worker-%d", i))
			if err == sql.ErrNoRows {
				return
			}
			if err != nil {
				errs <- err
				return
			}
			claimedIDs <- item.ID
		}(i)
	}
	wg.Wait()
	close(claimedIDs)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent claim inbound work: %v", err)
		}
	}
	claims := []int64{}
	for id := range claimedIDs {
		claims = append(claims, id)
	}
	if len(claims) != 1 || claims[0] != queued.ID {
		t.Fatalf("expected exactly one successful claim for %d, got %#v", queued.ID, claims)
	}
	item, err := s.GetInboundWork(ctx, queued.ID)
	if err != nil {
		t.Fatalf("get claimed inbound work: %v", err)
	}
	if item.Status != "claimed" || item.ClaimedBy == "" {
		t.Fatalf("expected claimed inbound work after concurrent claim, got %#v", item)
	}
}

func TestStoreInboundWorkRetryScheduling(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	queued, err := s.EnqueueInboundWork(ctx, "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "retry me"})
	if err != nil {
		t.Fatalf("enqueue inbound work: %v", err)
	}
	if err := s.RecordInboundWorkRetry(ctx, queued.ID, 1, "temporary failure", 30*time.Millisecond); err != nil {
		t.Fatalf("record inbound work retry: %v", err)
	}
	retried, err := s.GetInboundWork(ctx, queued.ID)
	if err != nil {
		t.Fatalf("get retried inbound work: %v", err)
	}
	if retried.Status != "retry" || retried.AttemptCount != 1 || retried.LastError != "temporary failure" || retried.NextAttemptAt == "" || retried.ClaimedBy != "" || retried.ClaimedAt != "" {
		t.Fatalf("unexpected retried inbound work: %#v", retried)
	}
	if _, err := s.ClaimNextInboundWork(ctx, "worker-too-early"); err != sql.ErrNoRows {
		t.Fatalf("expected no eligible retry claim before backoff expires, got %v", err)
	}
	time.Sleep(40 * time.Millisecond)
	claimed, err := s.ClaimNextInboundWork(ctx, "worker-after-backoff")
	if err != nil {
		t.Fatalf("claim inbound work after backoff: %v", err)
	}
	if claimed.ID != queued.ID || claimed.Status != "claimed" || claimed.AttemptCount != 1 || claimed.LastError != "temporary failure" {
		t.Fatalf("unexpected claimed retry item: %#v", claimed)
	}
}

func TestStoreRequeueInboundWork(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	queuedA, err := s.EnqueueInboundWork(ctx, "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "requeue me"})
	if err != nil {
		t.Fatalf("enqueue inbound work a: %v", err)
	}
	if err := s.RecordInboundWorkFailure(ctx, queuedA.ID, 3, "bad input"); err != nil {
		t.Fatalf("record failure a: %v", err)
	}
	requeued, err := s.RequeueInboundWork(ctx, queuedA.ID, false)
	if err != nil {
		t.Fatalf("requeue inbound work: %v", err)
	}
	if requeued.Status != "queued" || requeued.AttemptCount != 3 || requeued.LastError != "" || requeued.NextAttemptAt != "" || requeued.ClaimedBy != "" || requeued.ClaimedAt != "" {
		t.Fatalf("unexpected requeued inbound work: %#v", requeued)
	}
	queuedB, err := s.EnqueueInboundWork(ctx, "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "requeue and reset"})
	if err != nil {
		t.Fatalf("enqueue inbound work b: %v", err)
	}
	if err := s.RecordInboundWorkFailure(ctx, queuedB.ID, 2, "bad input again"); err != nil {
		t.Fatalf("record failure b: %v", err)
	}
	reset, err := s.RequeueInboundWork(ctx, queuedB.ID, true)
	if err != nil {
		t.Fatalf("requeue inbound work with reset: %v", err)
	}
	if reset.AttemptCount != 0 || reset.Status != "queued" {
		t.Fatalf("expected attempt count reset on requeue, got %#v", reset)
	}
	claimed, err := s.ClaimNextInboundWork(ctx, "worker")
	if err != nil {
		t.Fatalf("claim requeued work: %v", err)
	}
	if claimed.ID != queuedA.ID {
		t.Fatalf("expected first requeued item to become claimable again, got %#v", claimed)
	}
}

func TestStoreDiscardInboundWork(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	queued, err := s.EnqueueInboundWork(ctx, "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "discard queued"})
	if err != nil {
		t.Fatalf("enqueue queued inbound work: %v", err)
	}
	discardedQueued, err := s.DiscardInboundWork(ctx, queued.ID)
	if err != nil {
		t.Fatalf("discard queued inbound work: %v", err)
	}
	if discardedQueued.Status != "discarded" || discardedQueued.NextAttemptAt != "" || discardedQueued.ClaimedBy != "" || discardedQueued.ClaimedAt != "" {
		t.Fatalf("unexpected discarded queued item: %#v", discardedQueued)
	}
	failed, err := s.EnqueueInboundWork(ctx, "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "discard failed"})
	if err != nil {
		t.Fatalf("enqueue failed inbound work: %v", err)
	}
	if err := s.RecordInboundWorkFailure(ctx, failed.ID, 3, "bad input"); err != nil {
		t.Fatalf("record failure: %v", err)
	}
	discardedFailed, err := s.DiscardInboundWork(ctx, failed.ID)
	if err != nil {
		t.Fatalf("discard failed inbound work: %v", err)
	}
	if discardedFailed.Status != "discarded" {
		t.Fatalf("unexpected discarded failed item: %#v", discardedFailed)
	}
}

func TestStageSteeringContinuationReturnsQueuedTurnWithoutReload(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "session_stage_store_turn", "Test", map[string]any{"model": "bootstrap", "status": "idle"}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "turn_stage_store_prev", "session_stage_store_turn", "completed", "prev", map[string]any{"intent": "prompt", "model": "bootstrap"}); err != nil {
		t.Fatalf("create previous turn: %v", err)
	}
	if _, err := s.EnqueueSteering(ctx, "session_stage_store_turn", "turn_stage_store_prev", "user", "continue", map[string]any{"intent": "continue", "model": "bootstrap"}, nil, "one-at-a-time"); err != nil {
		t.Fatalf("enqueue steering: %v", err)
	}
	turnRec, msgs, err := s.StageSteeringContinuation(ctx, "session_stage_store_turn", "turn_stage_store_new")
	if err != nil {
		t.Fatalf("stage steering continuation: %v", err)
	}
	if turnRec == nil || turnRec.ID != "turn_stage_store_new" || turnRec.SessionID != "session_stage_store_turn" || turnRec.Status != "queued" || turnRec.Phase != "queued" {
		t.Fatalf("unexpected staged turn record: %#v", turnRec)
	}
	if len(msgs) != 1 || msgs[0].Content != "continue" {
		t.Fatalf("unexpected staged steering messages: %#v", msgs)
	}
	if got, err := s.CountQueuedTurns(ctx, "session_stage_store_turn"); err != nil {
		t.Fatalf("count queued turns: %v", err)
	} else if got != 1 {
		t.Fatalf("expected queued count 1 after stage, got %d", got)
	}
	persisted, err := s.GetTurn(ctx, "turn_stage_store_new")
	if err != nil {
		t.Fatalf("get persisted staged turn: %v", err)
	}
	if persisted.Status != "queued" || persisted.Phase != "queued" {
		t.Fatalf("unexpected persisted staged turn: %#v", persisted)
	}
}

func TestDeleteTurnIsIdempotentWhenMissing(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	if err := s.DeleteTurn(ctx, "turn_missing"); err != nil {
		t.Fatalf("delete missing turn should be idempotent: %v", err)
	}
}

func TestStoreInboundDispatcherLeaseSingleOwner(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	acquired, err := s.AcquireInboundDispatcherLease(ctx, "owner-a", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire first lease: %v", err)
	}
	if !acquired {
		t.Fatal("expected first lease acquisition to succeed")
	}
	acquired, err = s.AcquireInboundDispatcherLease(ctx, "owner-b", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire competing lease: %v", err)
	}
	if acquired {
		t.Fatal("expected competing lease acquisition to fail while current lease is live")
	}
	time.Sleep(120 * time.Millisecond)
	acquired, err = s.AcquireInboundDispatcherLease(ctx, "owner-b", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire expired lease: %v", err)
	}
	if !acquired {
		t.Fatal("expected lease acquisition after expiry to succeed")
	}
}

func TestStoreDiscardInboundWorkRejectsNonDiscardableState(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	queued, err := s.EnqueueInboundWork(ctx, "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "complete me"})
	if err != nil {
		t.Fatalf("enqueue inbound work: %v", err)
	}
	if err := s.UpdateInboundWorkStatus(ctx, queued.ID, "completed"); err != nil {
		t.Fatalf("mark inbound work completed: %v", err)
	}
	if _, err := s.DiscardInboundWork(ctx, queued.ID); err == nil {
		t.Fatal("expected discard of completed item to fail")
	}
}

func TestStoreRequeueInboundWorkRejectsNonRequeueableState(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	queued, err := s.EnqueueInboundWork(ctx, "ipc", "", "", map[string]any{"kind": "prompt", "prompt": "still queued"})
	if err != nil {
		t.Fatalf("enqueue inbound work: %v", err)
	}
	if _, err := s.RequeueInboundWork(ctx, queued.ID, false); err == nil {
		t.Fatal("expected requeue of queued item to fail")
	}
}

func TestStoreInboundWorkQueueLifecycle(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	sess, err := s.CreateSession(ctx, "session_inbound", "Inbound", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	queuedA, err := s.EnqueueInboundWork(ctx, "ipc", sess.ID, "", map[string]any{"kind": "prompt", "prompt": "hello"})
	if err != nil {
		t.Fatalf("enqueue inbound work a: %v", err)
	}
	queuedB, err := s.EnqueueInboundWork(ctx, "system", "", "opaque-session-key", map[string]any{"kind": "continue"})
	if err != nil {
		t.Fatalf("enqueue inbound work b: %v", err)
	}
	items, err := s.ListInboundWork(ctx, "queued", 10)
	if err != nil {
		t.Fatalf("list queued inbound work: %v", err)
	}
	if len(items) != 2 || items[0].ID != queuedA.ID || items[1].ID != queuedB.ID {
		t.Fatalf("unexpected queued inbound work list: %#v", items)
	}
	claimed, err := s.ClaimNextInboundWork(ctx, "worker-1")
	if err != nil {
		t.Fatalf("claim inbound work: %v", err)
	}
	if claimed.ID != queuedA.ID || claimed.Status != "claimed" || claimed.ClaimedBy != "worker-1" {
		t.Fatalf("unexpected claimed inbound work: %#v", claimed)
	}
	if err := s.UpdateInboundWorkStatus(ctx, claimed.ID, "completed"); err != nil {
		t.Fatalf("update inbound work status: %v", err)
	}
	completed, err := s.GetInboundWork(ctx, claimed.ID)
	if err != nil {
		t.Fatalf("get completed inbound work: %v", err)
	}
	if completed.Status != "completed" || completed.ClaimedBy != "" || completed.ClaimedAt != "" {
		t.Fatalf("expected completed inbound work with cleared claim state, got %#v", completed)
	}
	remaining, err := s.ClaimNextInboundWork(ctx, "worker-2")
	if err != nil {
		t.Fatalf("claim second inbound work: %v", err)
	}
	if remaining.ID != queuedB.ID || remaining.ExplicitSessionKey != "opaque-session-key" || remaining.SourceKind != "system" {
		t.Fatalf("unexpected second claimed inbound work: %#v", remaining)
	}
}

func TestCloneSessionCreatesChildAgentSession(t *testing.T) {
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	source, err := s.CreateSession(ctx, "session_root", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create source session: %v", err)
	}
	if err := s.AddMessage(ctx, "msg_root", source.ID, "user", "hello", map[string]any{"intent": "prompt"}); err != nil {
		t.Fatalf("add source message: %v", err)
	}

	child, err := s.CloneSession(ctx, source.ID, "session_child", "@agent1", "agent1")
	if err != nil {
		t.Fatalf("clone session: %v", err)
	}
	if child.ParentSessionID != source.ID {
		t.Fatalf("expected parent session %s, got %s", source.ID, child.ParentSessionID)
	}
	if child.Scope == nil || child.Scope.AgentID != "agent1" {
		t.Fatalf("unexpected child scope: %#v", child.Scope)
	}
	msgs, err := s.ListMessages(ctx, child.ID)
	if err != nil {
		t.Fatalf("list child messages: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Content != "hello" {
		t.Fatalf("unexpected cloned messages: %#v", msgs)
	}
}
