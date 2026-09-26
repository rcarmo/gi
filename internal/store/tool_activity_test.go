package store

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestToolActivityOccurrenceIdentityBoundsAndTerminalTiming(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateSession(ctx, "B", "B", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateTurn(ctx, "t", "A", "request", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "t", "test", "claim-t"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	appendEvent := func(kind string, p map[string]any) {
		t.Helper()
		if err := s.AppendTurnEvent(ctx, "t", "A", kind, p); err != nil {
			t.Fatal(err)
		}
	}
	get := func() map[string]any {
		t.Helper()
		a, err := s.SessionActivity(ctx, "A")
		if err != nil {
			t.Fatal(err)
		}
		return a["tool"].(map[string]any)
	}
	appendEvent("tool.started", map[string]any{"tool": "shell", "tool_call_id": "a", "preview": strings.Repeat("β", 300)})
	first := get()
	if first["state"] != "running" || first["tool_call_id"] != "a" || len([]rune(first["preview"].(string))) != 201 {
		t.Fatal(first)
	}
	appendEvent("tool.started", map[string]any{"tool": "shell", "tool_call_id": "b", "preview": "echo new"})
	appendEvent("tool.finished", map[string]any{"tool": "shell", "tool_call_id": "a"})
	if got := get(); got["state"] != "running" || got["tool_call_id"] != "b" {
		t.Fatal(got)
	}
	appendEvent("tool.failed", map[string]any{"tool": "shell", "tool_call_id": "b"})
	final := get()
	if final["state"] != "failed" || final["duration_ms"] == nil || final["finished_at"] == "" {
		t.Fatal(final)
	}
	if got := get(); got["duration_ms"] != final["duration_ms"] {
		t.Fatal("duration drift", got)
	}
	if b, err := s.SessionActivity(ctx, "B"); err != nil || b["tool"] != nil {
		t.Fatal(b, err)
	}
	if _, err = s.CreateTurn(ctx, "new", "A", "next", nil); err != nil {
		t.Fatal(err)
	}
	if err = s.ReleaseSessionActiveTurn(ctx, "A", "claim-t"); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "new", "test", "claim-new"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if a, err := s.SessionActivity(ctx, "A"); err != nil || a["tool"] != nil {
		t.Fatal(a, err)
	}
	if err = s.AppendTurnEvent(ctx, "new", "A", "tool.started", map[string]any{"tool": "shell", "command": []string{"echo", "legacy"}}); err != nil {
		t.Fatal(err)
	}
	if got := get(); got["tool_call_id"] == "" || got["preview"] != "echo legacy" {
		t.Fatal(got)
	}
	if err = s.ReleaseSessionActiveTurn(ctx, "A", "claim-new"); err != nil {
		t.Fatal(err)
	}
	if got := get(); got["state"] != "interrupted" || got["duration_ms"] != nil || got["finished_at"] != "" {
		t.Fatal(got)
	}
}
func TestToolPreviewExcludesOutputAndArbitraryArguments(t *testing.T) {
	if got := ToolActivityPreview(map[string]any{"password": "secret", "output": "hidden"}); got != "" {
		t.Fatal(got)
	}
	if got := ToolActivityPreview(map[string]any{"command": "echo\n β\t ok"}); got != "echo β ok" {
		t.Fatal(got)
	}
}

func TestToolActivityExplicitStoppedOccurrenceRejectsLateReusedCall(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "tool-terminal.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err = s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateTurnWithStatus(ctx, "t", "A", "running", "", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "t", "owner", "claim"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	event := func(kind, occurrence string) {
		t.Helper()
		if err := s.AppendTurnEvent(ctx, "t", "A", kind, map[string]any{"tool": "same", "tool_call_id": "reused", "occurrence_id": occurrence}); err != nil {
			t.Fatal(err)
		}
	}
	snapshot := func() map[string]any {
		t.Helper()
		a, err := s.SessionActivity(ctx, "A")
		if err != nil {
			t.Fatal(err)
		}
		return a["tool"].(map[string]any)
	}
	event("tool.started", "old")
	event("tool.finished", "old")
	event("tool.started", "new")
	event("tool.cancelled", "old")
	event("tool.finished", "")
	if got := snapshot(); got["state"] != "running" || got["duration_ms"] != nil {
		t.Fatal(got)
	}
	event("tool.cancelled", "new")
	if _, err := s.DB().Exec(`update turn_events set created_at=case event_type when 'tool.started' then '2026-01-01T00:00:00Z' else '2026-01-01T00:00:02.250Z' end where json_extract(payload_json,'$.occurrence_id')='new'`); err != nil {
		t.Fatal(err)
	}
	got := snapshot()
	if got["state"] != "cancelled" || got["duration_ms"] != int64(2250) || got["occurrence_id"] != "new" {
		t.Fatal(got)
	}
	event("tool.finished", "new")
	event("tool.aborted", "new")
	if later := snapshot(); later["state"] != "cancelled" || later["duration_ms"] != got["duration_ms"] {
		t.Fatal(later)
	}
	event("tool.started", "third")
	event("tool.aborted", "third")
	if got := snapshot(); got["state"] != "aborted" || got["duration_ms"] == nil {
		t.Fatal(got)
	}
	event("tool.started", "unknown")
	if err := s.AppendTurnEvent(ctx, "t", "A", "turn.finished", map[string]any{"status": "cancelled"}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateTurnStatus(ctx, "t", "cancelled"); err != nil {
		t.Fatal(err)
	}
	if got := snapshot(); got["state"] != "interrupted" || got["duration_ms"] != nil || got["finished_at"] != "" {
		t.Fatal(got)
	}
}
