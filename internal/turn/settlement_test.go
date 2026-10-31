package turn

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func TestCommittedUsageSurvivesLaterProviderFailureOrCancellation(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		for _, known := range []bool{false, true} {
			t.Run(fmt.Sprintf("cancel=%v/known=%v", cancelled, known), func(t *testing.T) {
				ctx := context.Background()
				s := openTestStore(t)
				defer s.Close()
				e := New(s)
				defer e.Close()
				if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
					t.Fatal(err)
				}
				if err := e.RegisterTool(tools.RegisteredTool{Name: "account-model", Executor: func(_ context.Context, rt tools.ToolRuntime, _ goai.ToolCall) (string, error) {
					rt.AddUsage(goai.Usage{Input: 3, Output: 1, TotalTokens: 4, Cost: goai.CostBreakdown{Total: .02}})
					return "nested model answer", nil
				}}); err != nil {
					t.Fatal(err)
				}
				var requests atomic.Int32
				original := streamWithToolsWithHooks
				defer func() { streamWithToolsWithHooks = original }()
				streamWithToolsWithHooks = func(requestCtx context.Context, _ string, _ *goai.Context, _ func(map[string]any), _ *inference.StreamHooks) (*inference.StreamResult, error) {
					if requests.Add(1) == 1 {
						return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonToolUse, Content: []goai.ContentBlock{{Type: "toolCall", ID: "c", Name: "account-model", Arguments: map[string]any{}}}}, Usage: &goai.Usage{Input: 7, Output: 2, TotalTokens: 9, Cost: goai.CostBreakdown{Total: .1}}}, nil
					}
					var result *inference.StreamResult
					if known {
						result = &inference.StreamResult{Usage: &goai.Usage{Input: 5, Output: 1, TotalTokens: 6, Cost: goai.CostBreakdown{Total: .03}}}
					}
					if cancelled {
						if err := e.CancelActiveTurn(ctx, "A", e.LiveActiveTurn("A")); err != nil {
							return nil, err
						}
						return result, requestCtx.Err()
					}
					return result, fmt.Errorf("invalid request after prior successful work")
				}
				run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "account across failures", Model: "provider/model"})
				if err != nil {
					t.Fatal(err)
				}
				status := "failed"
				if cancelled {
					status = "cancelled"
				}
				waitForCondition(t, 3*time.Second, func() bool { rec, err := s.GetTurn(ctx, run.TurnID); return err == nil && rec.Status == status }, "terminal settlement")
				total, err := s.TurnAccounting(ctx, run.TurnID)
				if err != nil {
					t.Fatal(err)
				}
				wantInput, wantTotal, wantUnknown, wantCost := 10, 13, 1, .12
				if known {
					wantInput = 15
					wantTotal = 19
					wantUnknown = 0
					wantCost = .15
				}
				if total.Usage.Input != wantInput || total.Usage.TotalTokens != wantTotal || total.Unknown != wantUnknown || math.Abs(total.Usage.Cost.Total-wantCost) > 1e-12 {
					t.Fatal("prior/nested usage omitted or counted twice", total)
				}
				events, err := s.ListTurnEvents(ctx, run.TurnID)
				if err != nil {
					t.Fatal(err)
				}
				finished := 0
				for _, ev := range events {
					if ev.Type == "inference.finished" {
						finished++
						usage, _ := ev.Payload["usage"].(map[string]any)
						if usage["input"] != float64(wantInput) || ev.Payload["accounting_source"] != "committed_receipts" {
							t.Fatal("terminal total not derived from receipts", ev.Payload)
						}
					}
				}
				if finished != 1 || requests.Load() != 2 {
					t.Fatal("duplicate accounting or requests", finished, requests.Load())
				}
				waitForCondition(t, time.Second, func() bool {
					r := e.runner("A")
					r.mu.Lock()
					defer r.mu.Unlock()
					_, _, err := s.GetSessionActiveTurn(ctx, "A")
					return err != nil
				}, "cleanup after terminal settlement")
			})
		}
	}
}

func TestFailedTerminalSettlementPublishesNothing(t *testing.T) {
	s, e := intentFixture(t)
	ctx := context.Background()
	if _, err := s.DB().Exec(`create trigger reject_terminal before insert on messages when new.role='system' begin select raise(abort,'terminal fault'); end`); err != nil {
		t.Fatal(err)
	}
	events := e.Subscribe("A")
	defer e.Unsubscribe("A", events)
	e.runner("A").finishTurn(s, "t", "A", "agent", "provider/model", "failed", "failure", "provider_error")
	rec, err := s.GetTurn(ctx, "t")
	if err != nil || rec.Status != "running" {
		t.Fatal("terminal escaped rollback", rec, err)
	}
	if len(events) != 0 {
		t.Fatal("uncommitted terminal published", <-events)
	}
	persisted, err := s.ListTurnEvents(ctx, "t")
	if err != nil {
		t.Fatal(err)
	}
	for _, ev := range persisted {
		if ev.Type == "turn.finished" || ev.Type == "inference.finished" || ev.Type == "turn.failure_marked" {
			t.Fatal("partial terminal event", ev)
		}
	}
}

func TestCancellationRetainsLateReceiptWithoutAdoptingIt(t *testing.T) {
	s, e := intentFixture(t)
	ctx := context.Background()
	if err := s.UpdateTurnStatus(ctx, "t", "cancelling"); err != nil {
		t.Fatal(err)
	}
	message := replayAnswer("late answer")
	message.StopReason = goai.StopReasonStop
	result := &inference.StreamResult{OccurrenceID: "late", Message: &message, Usage: &goai.Usage{Input: 7}}
	err := e.runner("A").settleProviderResponse(s, "t", "A", "provider/model", 1, result, "late-message", "late answer", map[string]any{"provider_message": message}, "reply_boundary")
	if err != context.Canceled {
		t.Fatal("late receipt adopted", err)
	}
	raw, err := s.ListMessages(ctx, "A")
	if err != nil || len(raw) != 1 {
		t.Fatal("late receipt lost", raw, err)
	}
	saved, ok := decodeAssistantReplay(raw[0].Payload["provider_message"])
	if !ok || saved.StopReason != goai.StopReasonStop || raw[0].Payload["adopted"] != false {
		t.Fatal("remote receipt mutated", raw[0])
	}
	snapshot, err := s.ContextSnapshot(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	if projected := e.runner("A").projectContextSnapshot(ctx, "A", snapshot); len(projected) != 0 {
		t.Fatal("unadopted response contributes to context", projected)
	}
	total, err := s.TurnAccounting(ctx, "t")
	if err != nil || total.Usage.Input != 7 {
		t.Fatal("late billing receipt lost", total, err)
	}
}

func TestCompletionHookIsProposalUntilCommit(t *testing.T) {
	s, e := intentFixture(t)
	ctx := context.Background()
	if _, err := s.DB().Exec(`create trigger reject_terminal before insert on turn_events when new.event_type='turn.finished' begin select raise(abort,'terminal fault'); end`); err != nil {
		t.Fatal(err)
	}
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		m := replayAnswer("answer")
		m.StopReason = goai.StopReasonStop
		return &inference.StreamResult{Message: &m}, nil
	})
	proposals := 0
	ended := ""
	if _, err := e.RegisterHook(HookTurnEnd, "completion-observer", func(_ context.Context, req HookRequest) (HookResponse, error) {
		proposals++
		rec, err := s.GetTurn(ctx, "t")
		if err != nil || rec.Status != "running" || req.Payload["settled"] != false {
			t.Error("hook falsely claimed committed completion", rec, req.Payload, err)
		}
		return HookResponse{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.RegisterHook(HookAgentEnd, "end-observer", func(_ context.Context, req HookRequest) (HookResponse, error) {
		ended, _ = req.Payload["reason"].(string)
		return HookResponse{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	e.runner("A").runAgentLoop(ctx, s, "t", "A", "provider/model", "agent", nil)
	if proposals != 1 || ended != "unsettled" {
		t.Fatal("proposal/failed settlement confused with completion", proposals, ended)
	}
	rec, err := s.GetTurn(ctx, "t")
	if err != nil || rec.Status != "running" {
		t.Fatal("terminal escaped rollback", rec, err)
	}
}

func TestNestedUsageOverflowRetainsKnownSubtotalAndUncertainty(t *testing.T) {
	extras := &toolExtras{}
	add := extras.addUsageFunc()
	add(goai.Usage{Input: math.MaxInt})
	add(goai.Usage{Input: 7})
	add(goai.Usage{Output: -1})
	payload := extras.withReplay(map[string]any{}, "result", false)
	usage := payload["usage"].(goai.Usage)
	if usage.Input != math.MaxInt || usage.Output != 0 || payload["usage_unknown"] != true || payload["usage_error"] == "" {
		t.Fatal("invalid nested usage accepted as a known total", payload)
	}
}
