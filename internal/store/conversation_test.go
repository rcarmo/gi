package store

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestConversationPagesAndSearchDoNotExposeToolHistory(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "conversation.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	for _, id := range []string{"A", "B"} {
		if _, err = s.CreateSession(ctx, id, id, nil); err != nil {
			t.Fatal(err)
		}
	}
	add := func(id, role, text string, payload map[string]any) {
		t.Helper()
		if err := s.AddMessage(ctx, id, "A", role, text, payload); err != nil {
			t.Fatal(err)
		}
		if _, err := s.DB().Exec(`update messages set created_at='2026-01-01T00:00:00Z' where id=?`, id); err != nil {
			t.Fatal(err)
		}
	}
	add("000", "user", "visible question", nil)
	add("001", "assistant", "Let me inspect.\n[tool_call: shell]\n[tool_call: skills]", map[string]any{"kind": "tool_calls"})
	for i := 2; i < 80; i++ {
		add(fmt.Sprintf("%03d", i), "tool_result", "private output needle", nil)
	}
	add("080", "assistant", "[tool_call: shell]", map[string]any{"kind": "tool_calls"})
	add("081", "assistant", "literal [tool_call: this is prose]", nil)
	add("082", "assistant", "stored opaque marker", map[string]any{"kind": "tool_calls", "display_text": "explicit assistant prose"})
	add("083", "system", "provider failed", nil)
	add("084", "alien", "unknown role", nil)
	add("085", "system", "Queued prompt: private queue notice", map[string]any{"kind": "queue"})
	page, err := s.PageConversationMessages(ctx, "A", "", "", 2)
	if err != nil {
		t.Fatal(err)
	}
	if !page.HasMore || len(page.Messages) != 2 || page.Messages[0].ID != "082" || page.Messages[1].Role != "system" {
		t.Fatalf("bad latest page %#v", page)
	}
	second, err := s.PageConversationMessages(ctx, "A", page.Before, "", 2)
	if err != nil {
		t.Fatal(err)
	}
	if !second.HasMore || len(second.Messages) != 2 || second.Messages[0].ID != "001" || second.Messages[0].Content != "Let me inspect." || second.Messages[1].ID != "081" {
		t.Fatalf("raw tools disrupted paging %#v", second)
	}
	third, err := s.PageConversationMessages(ctx, "A", second.Before, "", 2)
	if err != nil || third.HasMore || len(third.Messages) != 1 || third.Messages[0].ID != "000" {
		t.Fatal(third, err)
	}
	forward, err := s.PageConversationMessages(ctx, "A", "", third.After, 100)
	if err != nil || len(forward.Messages) != 4 {
		t.Fatal(forward, err)
	}
	if _, err = s.PageConversationMessages(ctx, "B", page.Before, "", 2); err != ErrMessageCursor {
		t.Fatal("cross-session cursor accepted", err)
	}
	for _, q := range []string{"private output", "shell", "opaque marker", "unknown role", "private queue notice"} {
		rows, err := s.SearchConversationMessages(ctx, "A", q, "current", 10, 0)
		if err != nil || len(rows) != 0 {
			t.Fatal(q, rows, err)
		}
	}
	rows, err := s.SearchConversationMessages(ctx, "A", "explicit assistant", "current", 10, 0)
	if err != nil || len(rows) != 1 || rows[0].ID != "082" {
		t.Fatal(rows, err)
	}
	raw, err := s.ListMessages(ctx, "A")
	if err != nil || len(raw) != 86 {
		t.Fatal(len(raw), err)
	}
	if raw[len(raw)-1].Content != "Queued prompt: private queue notice" {
		t.Fatal("queue notice missing from raw history")
	}
	if !strings.Contains(raw[1].Content, "[tool_call: shell]") {
		t.Fatal("raw history modified")
	}
	rawSearch, err := s.SearchMessages(ctx, "A", "private output", "current", 100, 0)
	if err != nil || len(rawSearch) != 78 {
		t.Fatal(len(rawSearch), err)
	}
}

func TestConversationNeverTreatsUserTextAsInternalToolMetadata(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "users.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err = s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	for i, kind := range []string{"tool_calls", "tool_result"} {
		if err = s.AddMessage(ctx, fmt.Sprint(i), "A", "user", "literal [tool_call: example]", map[string]any{"kind": kind, "display_text": "not the user's words"}); err != nil {
			t.Fatal(err)
		}
	}
	page, err := s.PageConversationMessages(ctx, "A", "", "", 10)
	if err != nil || len(page.Messages) != 2 {
		t.Fatal(page, err)
	}
	for _, m := range page.Messages {
		if m.Content != "literal [tool_call: example]" {
			t.Fatal("user prose changed", m)
		}
	}
}

func TestCompactionSummaryIsDurableButNeverConversationOutput(t *testing.T) {
	s := compactionStore(t)
	ctx := context.Background()
	// Exercise the real persistence path used by manual and automatic compaction.
	for _, trigger := range []string{"manual", "automatic"} {
		seq, err := s.BeginCompaction(ctx, "A", "run", map[string]any{"trigger": trigger})
		if err != nil {
			t.Fatal(err)
		}
		if out, err := s.FinishCompaction(ctx, "A", "run", seq, "completed", "private compacted state", nil); err != nil || out != "completed" {
			t.Fatal(out, err)
		}
	}
	if err := s.AddMessage(ctx, "answer", "A", "assistant", "normal continuation", nil); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ { // Initial load and reconnect/reload use the same projection.
		page, err := s.PageConversationMessages(ctx, "A", "", "", 1)
		if err != nil || len(page.Messages) != 1 || page.Messages[0].ID != "answer" || page.HasMore {
			t.Fatalf("page: %+v %v", page, err)
		}
		prior, err := s.PageConversationMessages(ctx, "A", page.Before, "", 1)
		if err != nil || len(prior.Messages) != 0 {
			t.Fatalf("older page: %+v %v", prior, err)
		}
		matches, err := s.SearchConversationMessages(ctx, "A", "private compacted", "current", 10, 0)
		if err != nil || len(matches) != 0 {
			t.Fatalf("search leaked summary: %+v %v", matches, err)
		}
	}
	raw, err := s.ListMessages(ctx, "A")
	if err != nil || len(raw) != 3 {
		t.Fatal(raw, err)
	}
	summaries := 0
	for _, m := range raw {
		if m.Payload["kind"] == "compaction" {
			summaries++
			if m.Content != "private compacted state" {
				t.Fatal("durable summary lost", m)
			}
		}
	}
	if summaries != 2 {
		t.Fatal("durable summaries missing", raw)
	}
}
