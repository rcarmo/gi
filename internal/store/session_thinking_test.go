package store

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
)

func sessionThinkingState() map[string]any {
	return map[string]any{
		"status":            "idle",
		"queue_count":       0,
		"model":             "bootstrap",
		"provider":          "test",
		"selected_model":    "bootstrap",
		"selected_provider": "test",
		"thinking_level":    "low",
		"thinking_model":    "test/bootstrap",
	}
}

func openSessionThinkingStore(t *testing.T, path string) *Store {
	t.Helper()
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func loadSessionThinking(t *testing.T, s *Store, id string) *Session {
	t.Helper()
	sess, err := s.GetSession(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return sess
}

func requireNoSessionTurns(t *testing.T, s *Store, id string) {
	t.Helper()
	turns, err := s.ListTurns(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if len(turns) != 0 {
		t.Fatalf("unexpected turns: %+v", turns)
	}
}

func requireThinkingSelection(t *testing.T, sess *Session, model, level string) {
	t.Helper()
	if sess.State["thinking_model"] != model || sess.State["thinking_level"] != level {
		t.Fatalf("unexpected thinking state: %#v", sess.State)
	}
}

func TestSessionThinkingTokenIsSessionScopedAndFencesModelThinkingState(t *testing.T) {
	ctx := context.Background()
	s := openSessionThinkingStore(t, filepath.Join(t.TempDir(), "thinking.db"))
	defer s.Close()
	for _, id := range []string{"A", "B"} {
		if _, err := s.CreateSession(ctx, id, id, sessionThinkingState()); err != nil {
			t.Fatal(err)
		}
	}

	a := loadSessionThinking(t, s, "A")
	b := loadSessionThinking(t, s, "B")
	tokenA := SessionThinkingToken(a.ID, a.State)
	tokenB := SessionThinkingToken(b.ID, b.State)
	if tokenA == tokenB {
		t.Fatal("thinking token must be session scoped")
	}
	if err := s.SelectSessionThinking(ctx, "B", tokenA, "test/bootstrap", "medium"); !errors.Is(err, ErrContextChanged) {
		t.Fatal("accepted foreign session token", err)
	}

	if err := s.TouchSessionState(ctx, "A", map[string]any{"status": "running", "queue_count": 3}); err != nil {
		t.Fatal(err)
	}
	a = loadSessionThinking(t, s, "A")
	if SessionThinkingToken(a.ID, a.State) != tokenA {
		t.Fatal("unrelated session state changed thinking token")
	}
	if err := s.SelectSessionThinking(ctx, "A", tokenA, "test/bootstrap", "medium"); err != nil {
		t.Fatal(err)
	}
	a = loadSessionThinking(t, s, "A")
	requireThinkingSelection(t, a, "test/bootstrap", "medium")
	if SessionThinkingToken(a.ID, a.State) == tokenA {
		t.Fatal("thinking mutation did not advance thinking token")
	}
	if err := s.SelectSessionThinking(ctx, "A", tokenA, "test/bootstrap", "high"); !errors.Is(err, ErrContextChanged) {
		t.Fatal("accepted stale thinking token after thinking change", err)
	}

	tokenA = SessionThinkingToken(a.ID, a.State)
	if err := s.TouchSessionState(ctx, "A", map[string]any{"model": "other-model"}); err != nil {
		t.Fatal(err)
	}
	a = loadSessionThinking(t, s, "A")
	if SessionThinkingToken(a.ID, a.State) == tokenA {
		t.Fatal("model mutation did not advance thinking token")
	}
	if err := s.SelectSessionThinking(ctx, "A", tokenA, "test/bootstrap", "low"); !errors.Is(err, ErrContextChanged) {
		t.Fatal("accepted stale thinking token after model change", err)
	}

	requireThinkingSelection(t, loadSessionThinking(t, s, "B"), "test/bootstrap", "low")
	requireNoSessionTurns(t, s, "A")
	requireNoSessionTurns(t, s, "B")
}

func TestSelectSessionThinkingConcurrentStoresCASAndReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "thinking-cas.db")
	a := openSessionThinkingStore(t, path)
	defer a.Close()
	b := openSessionThinkingStore(t, path)
	defer b.Close()
	if _, err := a.CreateSession(ctx, "A", "A", sessionThinkingState()); err != nil {
		t.Fatal(err)
	}
	base := loadSessionThinking(t, a, "A")
	token := SessionThinkingToken(base.ID, base.State)

	type result struct {
		level string
		err   error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	var wg sync.WaitGroup
	for _, candidate := range []struct {
		store *Store
		level string
	}{{a, "medium"}, {b, "high"}} {
		wg.Add(1)
		go func(candidate struct {
			store *Store
			level string
		}) {
			defer wg.Done()
			<-start
			results <- result{level: candidate.level, err: candidate.store.SelectSessionThinking(ctx, "A", token, "test/bootstrap", candidate.level)}
		}(candidate)
	}
	close(start)
	wg.Wait()
	close(results)

	wins := 0
	winner := ""
	for res := range results {
		switch {
		case res.err == nil:
			wins++
			winner = res.level
		case errors.Is(res.err, ErrContextChanged):
		default:
			t.Fatal(res.err)
		}
	}
	if wins != 1 || winner == "" {
		t.Fatalf("expected exactly one CAS winner, got wins=%d winner=%q", wins, winner)
	}

	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	c := openSessionThinkingStore(t, path)
	defer c.Close()
	sess := loadSessionThinking(t, c, "A")
	requireThinkingSelection(t, sess, "test/bootstrap", winner)
	requireNoSessionTurns(t, c, "A")
}

func TestSelectSessionThinkingRollbackTriggerLeavesPersistedSelectionIntact(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "thinking-rollback.db")
	s := openSessionThinkingStore(t, path)
	defer s.Close()
	if _, err := s.CreateSession(ctx, "A", "A", sessionThinkingState()); err != nil {
		t.Fatal(err)
	}
	base := loadSessionThinking(t, s, "A")
	if err := s.SelectSessionThinking(ctx, "A", SessionThinkingToken(base.ID, base.State), "test/bootstrap", "medium"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().Exec(`create trigger reject_session_thinking before update of state_json on sessions begin select raise(abort,'thinking failed'); end`); err != nil {
		t.Fatal(err)
	}
	current := loadSessionThinking(t, s, "A")
	err := s.SelectSessionThinking(ctx, "A", SessionThinkingToken(current.ID, current.State), "test/bootstrap", "high")
	if err == nil || errors.Is(err, ErrContextChanged) {
		t.Fatal(err)
	}
	requireThinkingSelection(t, loadSessionThinking(t, s, "A"), "test/bootstrap", "medium")
	requireNoSessionTurns(t, s, "A")
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s = openSessionThinkingStore(t, path)
	defer s.Close()
	requireThinkingSelection(t, loadSessionThinking(t, s, "A"), "test/bootstrap", "medium")
	requireNoSessionTurns(t, s, "A")
}

func TestSessionThinkingSelectionRevisionRejectsABA(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "aba.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	row, err := s.CreateSession(ctx, "A", "A", map[string]any{"selected_model": "p/a", "thinking_model": "p/a", "thinking_level": "low"})
	if err != nil {
		t.Fatal(err)
	}
	old := SessionThinkingToken("A", row.State)
	for _, level := range []string{"high", "low"} {
		row, _ = s.GetSession(ctx, "A")
		if err = s.SelectSessionThinking(ctx, "A", SessionThinkingToken("A", row.State), "p/a", level); err != nil {
			t.Fatal(err)
		}
	}
	if err = s.SelectSessionThinking(ctx, "A", old, "p/a", "high"); !errors.Is(err, ErrContextChanged) {
		t.Fatal("ABA accepted", err)
	}
	row, _ = s.GetSession(ctx, "A")
	old = SessionThinkingToken("A", row.State)
	for _, model := range []string{"p/b", "p/a"} {
		if err = s.TouchSessionState(ctx, "A", map[string]any{"selected_model": model}); err != nil {
			t.Fatal(err)
		}
	}
	if err = s.SelectSessionThinking(ctx, "A", old, "p/a", "high"); !errors.Is(err, ErrContextChanged) {
		t.Fatal("model ABA accepted", err)
	}
}

func TestSessionThinkingLegacyWriterClearsWebBinding(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "legacy.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.CreateSession(ctx, "A", "A", map[string]any{"thinking_model": "p/a", "thinking_level": "low"})
	patch := map[string]any{"thinking_level": "high"}
	if err = s.TouchSessionState(ctx, "A", patch); err != nil {
		t.Fatal(err)
	}
	a, _ := s.GetSession(ctx, "A")
	if a.State["thinking_model"] != "" || a.State["thinking_level"] != "high" {
		t.Fatal(a.State)
	}
	if _, mutated := patch["thinking_model"]; mutated {
		t.Fatal("mutated caller patch")
	}
}
