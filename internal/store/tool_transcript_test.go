package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestStoppedToolResultsAreDisplayOnlyAndFirstTerminalWins(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	s.CreateSession(ctx, "s", "s", nil)
	s.CreateTurn(ctx, "t", "s", "prompt", nil)
	event := func(kind, occ string) {
		t.Helper()
		if err := s.AppendTurnEvent(ctx, "t", "s", kind, map[string]any{"tool": "shell", "tool_call_id": "same", "occurrence_id": occ, "duration_ms": 1500}); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.AddMessage(ctx, "old-result", "s", "tool_result", "ok", map[string]any{"turn_id": "t", "occurrence_id": "old"}); err != nil {
		t.Fatal(err)
	}
	event("tool.started", "old")
	event("tool.finished", "old")
	event("tool.cancelled", "old")
	event("tool.started", "new")
	event("tool.cancelled", "new")
	event("tool.aborted", "new")
	event("tool.started", "interrupted")
	rows, err := s.ListTerminalToolResults(ctx, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Payload["status"] != "cancelled" || rows[0].Payload["duration_ms"] != float64(1500) {
		t.Fatal(rows)
	}
	msgs, _ := s.ListMessages(ctx, "s")
	if len(msgs) != 1 {
		t.Fatal("display results entered model history", msgs)
	}
	other, _ := s.ListTerminalToolResults(ctx, "other")
	if len(other) != 0 {
		t.Fatal("session leak", other)
	}
}

func TestTerminalToolProjectionBootstrapAndIdentity(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "s", "s", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, "t", "s", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	event := func(kind, call string, ms any) {
		t.Helper()
		if err := s.AppendTurnEvent(ctx, "t", "s", kind, map[string]any{"tool": "shell", "tool_call_id": call, "occurrence_id": "bootstrap", "duration_ms": ms}); err != nil {
			t.Fatal(err)
		}
	}
	event("tool.started", "", nil)
	event("tool.cancelled", "wrong", 999)
	event("tool.finished", "", 0)
	rows, err := s.ListTerminalToolResults(ctx, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Payload["status"] != "ok" || rows[0].Payload["duration_ms"] != float64(0) {
		t.Fatal(rows)
	}
	if err := s.AddMessage(ctx, "result", "s", "tool_result", "ok", map[string]any{"turn_id": "t", "occurrence_id": "bootstrap"}); err != nil {
		t.Fatal(err)
	}
	rows, err = s.ListTerminalToolResults(ctx, "s")
	if err != nil || len(rows) != 0 {
		t.Fatal(rows, err)
	}
}
