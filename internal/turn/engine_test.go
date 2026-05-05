package turn

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

func openTestStore(t *testing.T) *store.Store {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	s, err := store.Open(dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return s
}

func TestSubmitPromptQueuesSecondTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_1", "Test", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)

	first, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_1", Prompt: "one", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit first: %v", err)
	}
	second, err := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_1", Prompt: "two", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit second: %v", err)
	}
	if first.Queued {
		t.Fatalf("first should not be queued: %#v", first)
	}
	if !second.Queued || second.Status != "queued" {
		t.Fatalf("second should be queued: %#v", second)
	}
	sess, err := s.GetSession(ctx, "session_1")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if got := sess.State["queue_count"]; got != float64(1) && got != 1 {
		t.Fatalf("expected queue_count=1 while queued, got %#v", got)
	}
	queuedTurn, err := s.GetTurn(ctx, second.TurnID)
	if err != nil {
		t.Fatalf("get queued turn: %v", err)
	}
	if queuedTurn.Phase != "queued" {
		t.Fatalf("expected queued phase, got %#v", queuedTurn)
	}

	time.Sleep(2500 * time.Millisecond)
	turns, err := s.ListTurns(ctx, "session_1")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected 2 turns, got %d", len(turns))
	}
	if turns[0].Status != "completed" || turns[1].Status != "completed" {
		t.Fatalf("unexpected turn statuses: %#v", turns)
	}
	if turns[0].Phase != "completed" || turns[1].Phase != "completed" {
		t.Fatalf("unexpected turn phases: %#v", turns)
	}
	sess, err = s.GetSession(ctx, "session_1")
	if err != nil {
		t.Fatalf("get session after completion: %v", err)
	}
	if got := sess.State["queue_count"]; got != float64(0) && got != 0 {
		t.Fatalf("expected queue_count=0 after completion, got %#v", got)
	}
}

func TestCancelQueuedTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, _ = s.CreateSession(ctx, "session_1", "Test", map[string]any{"model": "bootstrap"})
	engine := New(s)
	_, _ = engine.SubmitPrompt(ctx, RunInput{SessionID: "session_1", Prompt: "one", Model: "bootstrap"})
	second, _ := engine.SubmitPrompt(ctx, RunInput{SessionID: "session_1", Prompt: "two", Model: "bootstrap"})
	if err := engine.CancelTurn(ctx, "session_1", second.TurnID); err != nil {
		t.Fatalf("cancel queued: %v", err)
	}
	turnRec, err := s.GetTurn(ctx, second.TurnID)
	if err != nil {
		t.Fatalf("get turn: %v", err)
	}
	if turnRec.Status != "cancelled" {
		t.Fatalf("expected cancelled, got %s", turnRec.Status)
	}
}

func TestSubmitPromptRoutedCreatesChildAgentSession(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	root, err := s.CreateSession(ctx, "session_root", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	engine := New(s)
	result, err := engine.SubmitPromptRouted(ctx, RunInput{SessionID: root.ID, Prompt: "@agent1 hello there", Model: "bootstrap"})
	if err != nil {
		t.Fatalf("submit routed prompt: %v", err)
	}
	if !result.Routed || result.TargetAgentID != "agent1" || !result.CreatedSession {
		t.Fatalf("unexpected routed result: %#v", result)
	}
	time.Sleep(1500 * time.Millisecond)
	sessions, err := s.ListSessions(ctx)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	var child *store.Session
	for i := range sessions {
		if sessions[i].ID == result.SessionID {
			child = &sessions[i]
		}
	}
	if child == nil || child.ParentSessionID != root.ID {
		t.Fatalf("unexpected child session: %#v", child)
	}
	msgs, err := s.ListMessages(ctx, child.ID)
	if err != nil {
		t.Fatalf("list child messages: %v", err)
	}
	found := false
	for _, msg := range msgs {
		if msg.Role == "assistant" && msg.Payload["agent_id"] == "agent1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected assistant reply from @agent1, got %#v", msgs)
	}
}

func TestRunShellStreamsDraftChunks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var chunks []string
	out, err, cancelled := runShell(ctx, "hello", nil, func(delta string) {
		chunks = append(chunks, delta)
	})
	if cancelled || err != nil {
		t.Fatalf("unexpected shell result: out=%q err=%v cancelled=%v", out, err, cancelled)
	}
	if out != "Gi received: hello" {
		t.Fatalf("unexpected output: %q", out)
	}
	if len(chunks) == 0 {
		t.Fatal("expected streamed chunks")
	}
	if got := chunks[0]; got == "" {
		t.Fatalf("expected non-empty first chunk: %#v", chunks)
	}
}

func TestSubmitPeerMessageUsesExistingTargetSession(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	source, err := s.CreateSession(ctx, "session_root", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	target, err := s.CloneSession(ctx, source.ID, "session_child", "@agent1", "agent1")
	if err != nil {
		t.Fatalf("create target: %v", err)
	}
	engine := New(s)
	result, err := engine.SubmitPeerMessage(ctx, source.ID, "agent1", "hello from peer", "prompt", "bootstrap", "")
	if err != nil {
		t.Fatalf("submit peer message: %v", err)
	}
	if result.SessionID != target.ID || result.CreatedSession {
		t.Fatalf("unexpected peer result: %#v", result)
	}
}
