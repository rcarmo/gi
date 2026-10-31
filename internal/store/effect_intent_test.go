package store

import (
	"context"
	"testing"
)

func TestEffectIntentAdmissionRollsBackPhaseAndBoundary(t *testing.T) {
	for _, table := range []string{"turns", "turn_events"} {
		t.Run(table, func(t *testing.T) {
			s := compactionStore(t)
			ctx := context.Background()
			before, err := s.GetTurn(ctx, "run")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.DB().Exec(`create trigger reject_intent before ` + map[string]string{"turns": "update", "turn_events": "insert"}[table] + ` on ` + table + ` begin select raise(abort,'intent rejected'); end`); err != nil {
				t.Fatal(err)
			}
			if err := s.AdmitTurnEffect(ctx, "A", "run", "waiting_on_tools", "tool.started", map[string]any{"prepared_call": "fixture"}); err == nil {
				t.Fatal("rejected intent admitted")
			}
			row, err := s.GetTurn(ctx, "run")
			if err != nil || row.Phase != before.Phase {
				t.Fatal(row, err)
			}
			events, err := s.ListTurnEvents(ctx, "run")
			if err != nil || len(events) != 0 {
				t.Fatal(events, err)
			}
		})
	}
}

func TestEffectIntentAdmissionRejectsCancellationAndForeignOwnership(t *testing.T) {
	s := compactionStore(t)
	ctx := context.Background()
	if err := s.AdmitTurnEffect(ctx, "foreign", "run", "waiting_on_tools", "tool.started", nil); err == nil {
		t.Fatal("foreign intent admitted")
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, "run", "cancelling", "cancelling"); err != nil {
		t.Fatal(err)
	}
	if err := s.AdmitTurnEffect(ctx, "A", "run", "waiting_on_tools", "tool.started", nil); err == nil {
		t.Fatal("cancelled intent admitted")
	}
	events, err := s.ListTurnEvents(ctx, "run")
	if err != nil || len(events) != 0 {
		t.Fatal(events, err)
	}
}
