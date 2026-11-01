package store

import (
	"context"
	"path/filepath"
	"testing"

	goai "github.com/rcarmo/go-ai"
)

func TestToolSettlementAtomicReopen(t *testing.T)         { testToolSettlementAtomicReopen(t, false) }
func TestToolResponseSettlementAtomicReopen(t *testing.T) { testToolSettlementAtomicReopen(t, true) }
func testToolSettlementAtomicReopen(t *testing.T, response bool) {
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
				apply := s.SettleToolResult
				if response {
					apply = s.SettleToolResponse
				}
				return apply(ctx, "t", "s", "tool.finished", map[string]any{"occurrence_id": "o", "tool_call_id": "call"}, "m", "result", map[string]any{"kind": "tool_result", "occurrence_id": "o", "turn_id": "t", "usage": goai.Usage{Input: 3}})
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
			var role string
			want := "tool_result"
			if response {
				want = "assistant"
			}
			if err := s.db.QueryRow(`select role from messages where id='m'`).Scan(&role); err != nil || role != want {
				t.Fatal("wrong settled message role", role, err)
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

func TestToolResponseSettlementCancellationWinsAdoption(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "response.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err = s.CreateSession(ctx, "s", "session", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateTurn(ctx, "t", "s", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	if err = s.UpdateTurnStatus(ctx, "t", "cancelling"); err != nil {
		t.Fatal(err)
	}
	payload := map[string]any{"kind": "chat", "turn_id": "t", "source": "shell"}
	if err = s.SettleToolResponse(ctx, "t", "s", "tool.finished", map[string]any{"occurrence_id": "o"}, "m", "observed", payload); err != nil {
		t.Fatal(err)
	}
	messages, err := s.ListMessages(ctx, "s")
	if err != nil || len(messages) != 1 {
		t.Fatal(messages, err)
	}
	m := messages[0]
	if m.Content != "observed" || m.Role != "assistant" || m.Payload["partial"] != true || m.Payload["adopted"] != false || m.Payload["stop_reason"] != string(goai.StopReasonAborted) {
		t.Fatal("cancelled response adopted", m)
	}
	if _, ok := payload["partial"]; ok {
		t.Fatal("caller envelope mutated", payload)
	}
	if err = s.SettleToolResponse(ctx, "t", "s", "tool.finished", map[string]any{"occurrence_id": "empty"}, "", "lost output", nil); err == nil {
		t.Fatal("accepted output without history identity")
	}
	if err = s.SettleToolResponse(ctx, "t", "s", "tool.finished", map[string]any{"occurrence_id": "empty"}, "", "", nil); err != nil {
		t.Fatal("empty output requires no chat row", err)
	}
	messages, err = s.ListMessages(ctx, "s")
	if err != nil || len(messages) != 1 {
		t.Fatal("empty response created history", messages, err)
	}
}
