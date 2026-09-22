package turn

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
)

func TestPromptAfterDisplayIdleGetsDistinctTurnBeforeCleanup(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "completion-admission", "Test", map[string]any{"model": "test-model", "status": "idle"}); err != nil {
		t.Fatal(err)
	}
	cfg := config.RuntimeConfig{WorkspaceRoot: t.TempDir(), DefaultModel: "test-model"}
	cfg.Hooks.TimeoutMS = 15000
	e := NewWithRuntimeConfig(s, cfg, "")
	defer e.Close()
	reached, release := make(chan struct{}), make(chan struct{})
	var once, unblock sync.Once
	defer unblock.Do(func() { close(release) })
	_, err := e.RegisterHook(HookSessionState, "hold-completed-claim", func(ctx context.Context, req HookRequest) (HookResponse, error) {
		if req.SessionID == "completion-admission" && req.SessionStatus == "idle" {
			once.Do(func() {
				close(reached)
				select {
				case <-release:
				case <-ctx.Done():
				}
			})
		}
		return HookResponse{}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	first, err := e.SubmitPrompt(ctx, RunInput{SessionID: "completion-admission", Prompt: "first", Model: "test-model"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-reached:
	case <-time.After(5 * time.Second):
		t.Fatal("terminal hook not reached")
	}
	activity, err := s.SessionActivity(ctx, first.SessionID)
	if err != nil || activity["status"] != "idle" {
		t.Fatal(activity, err)
	}
	active, _, err := s.GetSessionActiveTurn(ctx, first.SessionID)
	if err != nil || active != first.TurnID {
		t.Fatal("fixture did not retain old claim", active, err)
	}
	second, err := e.SubmitPrompt(ctx, RunInput{SessionID: first.SessionID, Prompt: "second explicit", Model: "test-model"})
	if err != nil {
		t.Fatal(err)
	}
	if second.TurnID == first.TurnID || !second.Queued || second.Status != "queued" {
		t.Fatalf("post-completion prompt attached to exhausted run: %+v", second)
	}
	if depth, err := s.SteeringQueueLength(ctx, first.SessionID); err != nil || depth != 0 {
		t.Fatal("terminal steering accepted", depth, err)
	}
	if active, _, _ := s.GetSessionActiveTurn(ctx, first.SessionID); active != first.TurnID {
		t.Fatal("old claim prematurely released")
	}
	unblock.Do(func() { close(release) })
	waitForCondition(t, 5*time.Second, func() bool {
		turn, err := s.GetTurn(ctx, second.TurnID)
		return err == nil && turn.Status == "completed"
	}, "second turn completes")
	waitForCondition(t, 5*time.Second, func() bool { _, _, err := s.GetSessionActiveTurn(ctx, first.SessionID); return err != nil }, "claim cleanup")
	turns, err := s.ListTurns(ctx, first.SessionID)
	if err != nil || len(turns) != 2 {
		t.Fatal(turns, err)
	}
	messages, err := s.ListMessages(ctx, first.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, m := range messages {
		if m.Role == "user" {
			counts[m.Content]++
		}
	}
	if counts["first"] != 1 || counts["second explicit"] != 1 || len(counts) != 2 {
		t.Fatal("user prompt duplicated/lost", counts)
	}
}

func TestTerminalClaimAdmissionsRemainFIFOAndDoNotSteer(t *testing.T) {
	for _, status := range []string{"completed", "failed", "cancelled", "aborted", "cancelling"} {
		for _, maintenance := range []bool{false, true} {
			if maintenance && status == "cancelling" {
				continue
			} // live maintenance remains a hard conflict
			t.Run(fmt.Sprintf("%s/maintenance=%v", status, maintenance), func(t *testing.T) {
				s := openTestStore(t)
				defer s.Close()
				ctx := context.Background()
				s.CreateSession(ctx, "A", "A", map[string]any{"model": "test-model", "status": "idle"})
				metadata := map[string]any{"model": "test-model"}
				if maintenance {
					metadata["operation"] = "manual_compaction"
				}
				s.CreateTurnWithStatus(ctx, "old", "A", status, "old", metadata)
				s.ClaimSessionActiveTurn(ctx, "A", "old", "test", "old")
				e := New(s)
				defer e.Close()
				runner := e.runner("A")
				old := &runningTurn{turnID: "old"}
				runner.current = old
				var ids []string
				for _, prompt := range []string{"one", "two", "three"} {
					result, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: prompt, Model: "test-model"})
					if err != nil {
						t.Fatal(err)
					}
					if !result.Queued || result.TurnID == "old" {
						t.Fatal("not distinct queued admission", result)
					}
					ids = append(ids, result.TurnID)
				}
				if n, _ := s.SteeringQueueLength(ctx, "A"); n != 0 {
					t.Fatal("terminal steering rows", n)
				}
				if active, _, _ := s.GetSessionActiveTurn(ctx, "A"); active != "old" {
					t.Fatal("claim stolen")
				}
				if status == "cancelling" {
					s.UpdateTurnStatusAndPhase(ctx, "old", "cancelled", "cancelled")
				}
				runner.cleanupTurnRun("A", "old", old)
				waitForCondition(t, 5*time.Second, func() bool { last, err := s.GetTurn(ctx, ids[2]); return err == nil && last.Status == "completed" }, "FIFO drain")
				waitForCondition(t, 5*time.Second, func() bool { _, _, err := s.GetSessionActiveTurn(ctx, "A"); return err != nil }, "drain cleanup")
				messages, _ := s.ListMessages(ctx, "A")
				var prompts []string
				for _, message := range messages {
					if message.Role == "user" {
						prompts = append(prompts, message.Content)
					}
				}
				if !reflect.DeepEqual(prompts, []string{"one", "two", "three"}) {
					t.Fatal("queue order/loss", prompts)
				}
			})
		}
	}
}

func TestLaunchConflictWithTerminalClaimRetainsDistinctAdmission(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	s.CreateSession(ctx, "A", "A", map[string]any{"status": "idle", "model": "test-model"})
	s.CreateTurnWithStatus(ctx, "exhausted", "A", "completed", "old", map[string]any{"model": "test-model"})
	e := New(s)
	defer e.Close()
	var once sync.Once
	e.beforeLaunchClaimHook = func(ctx context.Context, sid, tid string) {
		once.Do(func() {
			if ok, err := s.ClaimSessionActiveTurn(ctx, sid, "exhausted", "foreign", "foreign"); err != nil || !ok {
				t.Error("claim injection", ok, err)
			}
		})
	}
	result, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "preserved new prompt", Model: "test-model"})
	if err != nil {
		t.Fatal(err)
	}
	if result.TurnID == "exhausted" || !result.Queued {
		t.Fatal(result)
	}
	if n, _ := s.SteeringQueueLength(ctx, "A"); n != 0 {
		t.Fatal("launch fallback steered terminal run")
	}
	e.runner("A").cleanupTurnRun("A", "foreign", nil)
	waitForCondition(t, 5*time.Second, func() bool {
		record, err := s.GetTurn(ctx, result.TurnID)
		return err == nil && record.Status == "completed"
	}, "conflict fallback completes")
	waitForCondition(t, 5*time.Second, func() bool { _, _, err := s.GetSessionActiveTurn(ctx, "A"); return err != nil }, "fallback cleanup")
	messages, _ := s.ListMessages(ctx, "A")
	count := 0
	for _, m := range messages {
		if m.Role == "user" && m.Content == "preserved new prompt" {
			count++
		}
	}
	if count != 1 {
		t.Fatal("fallback lost or duplicated", count)
	}
}
