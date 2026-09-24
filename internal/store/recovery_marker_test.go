package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestRecoveryMarkerRequiresNativeRequeueAndExactSessionTurn(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateTurn(ctx, "t", "A", "request", nil); err != nil {
		t.Fatal(err)
	}
	check := func(session, turn string, want int) {
		t.Helper()
		m, err := s.RecoveryMarker(ctx, session, turn)
		if err != nil {
			t.Fatal(err)
		}
		if want == 0 {
			if m != nil {
				t.Fatal(m)
			}
			return
		}
		if m["attempts_used"] != want || m["recovered"] != true || m["type"] != "recovery_marker" {
			t.Fatal(m)
		}
	}
	check("A", "t", 0)
	for _, disposition := range []string{"release_terminal", "hold_for_retry_or_skip_after_tool_checkpoint", "abort_cancelling"} {
		if err = s.AppendTurnEvent(ctx, "t", "A", "turn.recovered", map[string]any{"recovery_disposition": disposition}); err != nil {
			t.Fatal(err)
		}
	}
	check("A", "t", 0)
	for _, d := range []string{"requeue_interrupted_turn", "requeue_after_compaction_checkpoint"} {
		if err = s.AppendTurnEvent(ctx, "t", "A", "turn.recovered", map[string]any{"recovery_disposition": d}); err != nil {
			t.Fatal(err)
		}
	}
	check("A", "t", 3)
	check("other", "t", 0)
	check("A", "other", 0)
}
