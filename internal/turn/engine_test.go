package turn

import (
	"context"
	"fmt"
	"sync"
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
	activeTurnID, _, err := s.GetSessionActiveTurn(ctx, "session_1")
	if err != nil {
		t.Fatalf("get active turn: %v", err)
	}
	if activeTurnID != first.TurnID {
		t.Fatalf("expected active turn %s, got %s", first.TurnID, activeTurnID)
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
	if _, _, err := s.GetSessionActiveTurn(ctx, "session_1"); err == nil {
		t.Fatal("expected no active turn after completion")
	}
}

func TestConcurrentSubmitAcrossEnginesQueuesCleanly(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_1", "Test", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	engineA := New(s)
	engineB := New(s)

	results := make(chan *SubmitResult, 2)
	errs := make(chan error, 2)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i, engine := range []*Engine{engineA, engineB} {
		wg.Add(1)
		go func(idx int, eng *Engine) {
			defer wg.Done()
			<-start
			res, err := eng.SubmitPrompt(ctx, RunInput{SessionID: "session_1", Prompt: fmt.Sprintf("prompt-%d", idx+1), Model: "bootstrap"})
			if err != nil {
				errs <- err
				return
			}
			results <- res
		}(i, engine)
	}
	close(start)
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected submit error: %v", err)
		}
	}
	var submitted []*SubmitResult
	queuedCount := 0
	for res := range results {
		submitted = append(submitted, res)
		if res.Queued {
			queuedCount++
		}
	}
	if len(submitted) != 2 {
		t.Fatalf("expected 2 submit results, got %d", len(submitted))
	}
	if queuedCount != 1 {
		t.Fatalf("expected exactly one queued turn, got %d: %#v", queuedCount, submitted)
	}

	time.Sleep(2500 * time.Millisecond)
	turns, err := s.ListTurns(ctx, "session_1")
	if err != nil {
		t.Fatalf("list turns: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected 2 turns, got %d", len(turns))
	}
	for _, turnRec := range turns {
		if turnRec.Status != "completed" {
			t.Fatalf("expected completed turns, got %#v", turns)
		}
	}
}

func TestStartupRecoveryRequeuesCompactingTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_recover_compact", "Test", map[string]any{"model": "bootstrap", "status": "running"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_recover_compact", "session_recover_compact", "running", "hello", map[string]any{"intent": "prompt"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, turnRec.ID, "running", "compacting"); err != nil {
		t.Fatalf("set compacting phase: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_recover_compact", turnRec.ID, "runner", turnRec.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update session_active_turns set updated_at = '2000-01-01T00:00:00Z' where session_id = ?`, "session_recover_compact"); err != nil {
		t.Fatalf("age active turn claim: %v", err)
	}

	_ = New(s)

	recovered, err := s.GetTurn(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("get recovered turn: %v", err)
	}
	if recovered.Status != "queued" || recovered.Phase != "queued" {
		t.Fatalf("expected queued recovered turn, got %#v", recovered)
	}
	if _, _, err := s.GetSessionActiveTurn(ctx, "session_recover_compact"); err == nil {
		t.Fatal("expected stale active claim to be released")
	}
	events, err := s.ListTurnEvents(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("list recovery events: %v", err)
	}
	found := false
	for _, event := range events {
		if event.Type == "turn.recovered" && event.Payload["recovery_disposition"] == "requeue_after_compaction_checkpoint" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected compaction recovery event, got %#v", events)
	}
}

func TestStartupRecoveryFailsToolPhaseTurn(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	_, err := s.CreateSession(ctx, "session_recover_tool", "Test", map[string]any{"model": "bootstrap", "status": "running"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	turnRec, err := s.CreateTurnWithStatus(ctx, "turn_recover_tool", "session_recover_tool", "running", "hello", map[string]any{"intent": "prompt"})
	if err != nil {
		t.Fatalf("create turn: %v", err)
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, turnRec.ID, "running", "waiting_on_tools"); err != nil {
		t.Fatalf("set waiting_on_tools phase: %v", err)
	}
	if _, err := s.ClaimSessionActiveTurn(ctx, "session_recover_tool", turnRec.ID, "runner", turnRec.ID); err != nil {
		t.Fatalf("claim active turn: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx, `update session_active_turns set updated_at = '2000-01-01T00:00:00Z' where session_id = ?`, "session_recover_tool"); err != nil {
		t.Fatalf("age active turn claim: %v", err)
	}

	_ = New(s)

	recovered, err := s.GetTurn(ctx, turnRec.ID)
	if err != nil {
		t.Fatalf("get recovered turn: %v", err)
	}
	if recovered.Status != "failed" || recovered.Phase != "failed" {
		t.Fatalf("expected failed recovered turn, got %#v", recovered)
	}
	if recovered.FinishedAt == "" {
		t.Fatalf("expected finished_at on failed recovered turn, got %#v", recovered)
	}
	if _, _, err := s.GetSessionActiveTurn(ctx, "session_recover_tool"); err == nil {
		t.Fatal("expected stale active claim to be released")
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
