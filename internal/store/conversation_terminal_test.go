package store

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

func TestConversationTerminalCodemodePagination(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "terminal-pages.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	for _, session := range []string{"s", "other"} {
		if _, err := s.CreateSession(ctx, session, session, nil); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.CreateTurn(ctx, "t", "s", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	if err := s.AddMessage(ctx, "user", "s", "user", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		p := map[string]any{"tool": "codemode", "tool_call_id": "reused", "occurrence_id": fmt.Sprintf("o%d", i), "duration_ms": 0}
		for _, kind := range []string{"tool.started", "tool.cancelled", "tool.aborted"} {
			if err := s.AppendTurnEvent(ctx, "t", "s", kind, p); err != nil {
				t.Fatal(err)
			}
		}
	}
	// Equal timestamps exercise the ID tie-breaker for synthetic and real rows.
	if _, err := s.db.Exec(`update messages set created_at='2026-01-01T00:00:00Z'; update turn_events set created_at='2026-01-01T00:00:00Z'`); err != nil {
		t.Fatal(err)
	}
	var before string
	seen := map[string]bool{}
	for {
		page, err := s.PageConversationMessages(ctx, "s", before, "", 2)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range page.Messages {
			if seen[m.ID] {
				t.Fatalf("duplicate %s", m.ID)
			}
			seen[m.ID] = true
			if m.Role == "tool_result" && (m.Payload["status"] != "cancelled" || m.Payload["kind"] != "tool_result") {
				t.Fatalf("bad projection: %+v", m)
			}
		}
		if !page.HasMore {
			break
		}
		before = page.Before
	}
	if len(seen) != 6 {
		t.Fatalf("want user plus five occurrences: %v", seen)
	}
	// Forward paging must use the same merged relation.
	first, err := s.PageConversationMessages(ctx, "s", "", "", 100)
	if err != nil {
		t.Fatal(err)
	}
	next, err := s.PageConversationMessages(ctx, "s", "", encodeMessageCursor(first.Messages[0]), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(next.Messages) != 2 || !next.HasMore {
		t.Fatalf("forward page: %+v", next)
	}
	other, err := s.PageConversationMessages(ctx, "other", "", "", 10)
	if err != nil || len(other.Messages) != 0 {
		t.Fatalf("cross-session leak: %+v %v", other, err)
	}
	raw, err := s.ListMessages(ctx, "s")
	if err != nil || len(raw) != 1 {
		t.Fatalf("projection entered raw history: %+v %v", raw, err)
	}
	// A persisted result suppresses only its matching occurrence.
	if err := s.AddMessage(ctx, "saved", "s", "tool_result", "Script failed", map[string]any{"kind": "tool_result", "tool_name": "codemode", "turn_id": "t", "occurrence_id": "o0"}); err != nil {
		t.Fatal(err)
	}
	page, err := s.PageConversationMessages(ctx, "s", "", "", 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Messages) != 6 {
		t.Fatalf("persisted result duplicated terminal: %+v", page.Messages)
	}
	for _, m := range page.Messages {
		if m.ID == "terminal:t:o0" {
			t.Fatal("stored occurrence still projected")
		}
	}
}
