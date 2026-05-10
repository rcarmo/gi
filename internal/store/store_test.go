package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"testing"

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
