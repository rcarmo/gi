package store

import (
	"context"
	"errors"
	"testing"
)

func TestTerminalStageAtomicReopenAndChildFence(t *testing.T) {
	s, path := settlementFixture(t)
	ctx := context.Background()
	decision := sampleTurnSettlement()
	if _, err := s.db.Exec(`create trigger reject_stage before insert on turn_events when new.event_type='turn.terminal_staged' begin select raise(abort,'reject stage');end`); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.StageTurnTerminal(ctx, "t", "s", decision); err == nil {
		t.Fatal("failed stage accepted")
	}
	before, err := s.GetTurn(ctx, "t")
	if err != nil || before.Phase == "settling" || before.Status != "running" {
		t.Fatal("stage rollback leaked", before, err)
	}
	if _, err := s.db.Exec(`drop trigger reject_stage`); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.StageTurnTerminal(ctx, "t", "other", decision); err == nil {
		t.Fatal("foreign decision accepted")
	}
	status, pending, err := s.StageTurnTerminal(ctx, "t", "s", decision)
	if err != nil || !pending || status != "failed" {
		t.Fatal(status, pending, err)
	}
	staged, err := s.GetTurn(ctx, "t")
	if err != nil || staged.Status != "running" || staged.Phase != "settling" || staged.FinishedAt != "" {
		t.Fatal("stage published terminal state", staged, err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "late", "s", "queued", "late", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateSubTurn(ctx, "t", "s", "late", "s", "async", 1, nil); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("child passed settling fence", err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	events, err := reopened.ListTurnEvents(ctx, "t")
	if err != nil {
		t.Fatal(err)
	}
	var stages int
	for _, event := range events {
		if event.Type == "turn.finished" {
			t.Fatal("stage committed finish")
		}
		if event.Type == "turn.terminal_staged" {
			stages++
			body := event.Payload["decision"].(map[string]any)
			if body["message"] != decision.Message || body["failure_kind"] != decision.FailureKind {
				t.Fatal("checkpoint lost terminal decision", body)
			}
		}
	}
	if stages != 1 {
		t.Fatal("stage count", stages)
	}
	decision.ChildrenStatus = "failed"
	settled, err := reopened.SettleRunningTurn(ctx, "t", "s", decision)
	if err != nil || !settled.Applied {
		t.Fatal(settled, err)
	}
	receipt, err := reopened.HasTurnTerminalReceipt(ctx, "t")
	if err != nil || !receipt {
		t.Fatal("missing terminal receipt", err)
	}
	_, pending, err = reopened.StageTurnTerminal(ctx, "t", "s", decision)
	if err != nil || pending {
		t.Fatal("terminal decision replayed", pending, err)
	}
}
func TestTerminalSettlementRequiresUpdatedChildDecisionAfterCancel(t *testing.T) {
	s, _ := settlementFixture(t)
	ctx := context.Background()
	decision := TurnSettlement{Status: "completed", Phase: "completed"}
	status, _, err := s.StageTurnTerminal(ctx, "t", "s", decision)
	if err != nil {
		t.Fatal(err)
	}
	decision.ChildrenStatus = status
	if err := s.UpdateTurnStatusAndPhase(ctx, "t", "cancelling", "cancelling"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SettleRunningTurn(ctx, "t", "s", decision); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("cancel published before critical-child join", err)
	}
	receipt, err := s.HasTurnTerminalReceipt(ctx, "t")
	if err != nil || receipt {
		t.Fatal("guard committed a terminal receipt", receipt, err)
	}
	status, pending, err := s.StageTurnTerminal(ctx, "t", "s", decision)
	if err != nil || !pending || status != "cancelled" {
		t.Fatal(status, pending, err)
	}
	decision.ChildrenStatus = status
	decision.MessageID = "cancelled"
	settled, err := s.SettleRunningTurn(ctx, "t", "s", decision)
	if err != nil || !settled.Applied || settled.Status != "cancelled" {
		t.Fatal(settled, err)
	}
}
