package turn

import (
	"context"
	"testing"
	"time"
)

func TestToolOutputReporterPersistsBeforeNotificationAndFlushesFinal(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	defer e.Close()

	ctx := context.Background()
	_, _ = s.CreateSession(ctx, "a", "a", nil)
	_, _ = s.CreateTurn(ctx, "t", "a", "prompt", nil)
	_ = s.AppendTurnEvent(ctx, "t", "a", "tool.started", map[string]any{"tool": "shell", "tool_call_id": "call", "occurrence_id": "occ"})
	r := &sessionRunner{store: s, engine: e}
	report, closeOutput := r.toolOutputReporter("t", "a", "call", "occ")
	defer closeOutput()
	if err := report("first", false); err != nil {
		t.Fatal(err)
	}
	if err := report("first\nsecond", false); err != nil {
		t.Fatal(err)
	}
	if err := report("first\nsecond", true); err != nil {
		t.Fatal(err)
	}
	a, err := s.SessionActivity(ctx, "a")
	if err != nil {
		t.Fatal(err)
	}
	if a["tool"].(map[string]any)["output_preview"] != "first\nsecond" {
		t.Fatal(a)
	}
	var count int
	s.DB().QueryRow(`select count(*) from turn_events where event_type='tool.output'`).Scan(&count)
	if count != 2 {
		t.Fatal(count)
	}
	// A failed preview write stays failed; a final flush cannot disguise it.
	if _, err = s.DB().Exec(`CREATE TRIGGER reject_output BEFORE INSERT ON turn_events WHEN NEW.event_type='tool.output' BEGIN SELECT RAISE(ABORT,'fixture output failure'); END`); err != nil {
		t.Fatal(err)
	}
	failed, closeFailed := r.toolOutputReporter("t", "a", "call", "occ")
	defer closeFailed()
	if failed("new", true) == nil || failed("", true) == nil {
		t.Fatal("persistence failure lost")
	}
}

func TestToolOutputProgressCoalescesAndJoinsFinalReplacement(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	defer e.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "a", "a", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, "t", "a", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	report, closeOutput := e.runner("a").toolOutputReporter("t", "a", "call", "occ")
	defer closeOutput()
	if err := report("first", false); err != nil {
		t.Fatal(err)
	}
	if err := report("pending latest", false); err != nil {
		t.Fatal(err)
	}
	waitForCondition(t, time.Second, func() bool {
		var count int
		s.DB().QueryRow(`select count(*) from turn_events where event_type='tool.output'`).Scan(&count)
		return count == 2
	}, "coalesced progress written without another producer update")
	if err := report("", true); err != nil {
		t.Fatal(err)
	}
	closeOutput()
	if err := report("late update after terminal", false); err != nil {
		t.Fatal(err)
	}
	events, err := s.ListTurnEvents(ctx, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 || events[2].Payload["output_preview"] != "" || events[2].Payload["final"] != true {
		t.Fatal("empty terminal did not replace progress", events)
	}
}
