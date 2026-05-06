package store

import (
	"context"
	"database/sql"
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
	if session.Scope == nil || session.Scope.AgentID != "gi" || session.Scope.Values["chat"] != "session_1" {
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
