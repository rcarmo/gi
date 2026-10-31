package store

import (
	"context"
	"path/filepath"
	"testing"

	goai "github.com/rcarmo/go-ai"
)

func TestToolSettlementAtomicReopen(t *testing.T) {
	for _, target := range []string{"messages", "turn_events", "turn_receipts", "turns", "sessions"} {
		t.Run(target, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "settle.db")
			s, err := Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { s.Close() }()
			ctx := context.Background()
			if _, err := s.CreateSession(ctx, "s", "session", nil); err != nil {
				t.Fatal(err)
			}
			if _, err := s.CreateTurn(ctx, "t", "s", "prompt", nil); err != nil {
				t.Fatal(err)
			}
			verb := "insert"
			if target == "turns" || target == "sessions" {
				verb = "update"
			}
			if _, err := s.db.Exec(`create trigger fail_settlement before ` + verb + ` on ` + target + ` begin select raise(abort,'settlement rejected'); end`); err != nil {
				t.Fatal(err)
			}
			settle := func(ctx context.Context) error {
				return s.SettleToolResult(ctx, "t", "s", "tool.finished", map[string]any{"occurrence_id": "o", "tool_call_id": "call"}, "m", "result", map[string]any{"kind": "tool_result", "occurrence_id": "o", "turn_id": "t", "usage": goai.Usage{Input: 3}})
			}
			if err := settle(ctx); err == nil {
				t.Fatal("expected injected failure")
			}
			for _, table := range []string{"messages", "turn_events"} {
				var count int
				if err := s.db.QueryRow(`select count(*) from ` + table).Scan(&count); err != nil {
					t.Fatal(err)
				}
				if count != 0 {
					t.Fatalf("partial settlement in %s", table)
				}
			}
			a, err := s.TurnAccounting(ctx, "t")
			if err != nil || a.Usage.Input != 0 {
				t.Fatal("tool accounting escaped rollback", a, err)
			}
			if _, err := s.db.Exec(`drop trigger fail_settlement`); err != nil {
				t.Fatal(err)
			}
			cancelled, cancel := context.WithCancel(ctx)
			cancel()
			if err := settle(cancelled); err == nil {
				t.Fatal("cancelled admission succeeded")
			}
			if err := settle(ctx); err != nil {
				t.Fatal(err)
			}
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
			s, err = Open(path)
			if err != nil {
				t.Fatal(err)
			}
			a, err = s.TurnAccounting(ctx, "t")
			if err != nil || a.Usage.Input != 3 {
				t.Fatal("tool usage missing after reopen", a, err)
			}
			if err := settle(ctx); err != nil {
				t.Fatal("idempotent tool retry", err)
			}
			for _, table := range []string{"messages", "turn_events", "turn_receipts"} {
				var count int
				if err := s.db.QueryRow(`select count(*) from ` + table).Scan(&count); err != nil {
					t.Fatal(err)
				}
				if count != 1 {
					t.Fatalf("settlement missing after reopen in %s: %d", table, count)
				}
			}
		})
	}
}

func TestToolSettlementRejectsCrossSessionTurn(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "settlement-owner.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	for _, id := range []string{"owner", "foreign"} {
		if _, err := s.CreateSession(ctx, id, id, nil); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.CreateTurn(ctx, "owned-turn", "owner", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	if err := s.SettleToolResult(ctx, "owned-turn", "foreign", "tool.finished", nil, "foreign-result", "secret", nil); err == nil {
		t.Fatal("cross-session settlement accepted")
	}
	for _, table := range []string{"turn_events", "messages"} {
		var count int
		if err := s.db.QueryRow(`select count(*) from ` + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("cross-session partial write: %s", table)
		}
	}
}
