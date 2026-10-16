package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestAppendRollsBackWhenParentUpdateFails(t *testing.T) {
	for _, kind := range []string{"message", "event"} {
		t.Run(kind, func(t *testing.T) {
			s, err := Open(filepath.Join(t.TempDir(), "atomic.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			ctx := context.Background()
			if _, err := s.CreateSession(ctx, "s", "session", nil); err != nil {
				t.Fatal(err)
			}
			if _, err := s.CreateTurn(ctx, "t", "s", "prompt", nil); err != nil {
				t.Fatal(err)
			}
			parent, rows := "sessions", "messages"
			if kind == "event" {
				parent, rows = "turns", "turn_events"
			}
			if _, err := s.db.Exec(`create trigger fail_parent before update on ` + parent + ` begin select raise(abort, 'parent update failed'); end`); err != nil {
				t.Fatal(err)
			}
			appendRow := func() error {
				if kind == "message" {
					return s.AddMessage(ctx, "m", "s", "assistant", "result", nil)
				}
				return s.AppendTurnEvent(ctx, "t", "s", "tool.finished", map[string]any{"occurrence_id": "o"})
			}
			if err := appendRow(); err == nil {
				t.Fatal("expected parent update failure")
			}
			var count int
			if err := s.db.QueryRow(`select count(*) from ` + rows).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 0 {
				t.Fatalf("failed append committed %d rows", count)
			}
			if _, err := s.db.Exec(`drop trigger fail_parent`); err != nil {
				t.Fatal(err)
			}
			if err := appendRow(); err != nil {
				t.Fatalf("retry after rollback: %v", err)
			}
			if err := s.db.QueryRow(`select count(*) from ` + rows).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("retry committed %d rows", count)
			}
		})
	}
}
