package turn

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestAutomaticCompactionNativeLifecycle(t *testing.T) {
	for _, kind := range []string{"complete", "cancel", "suppress", "persistence"} {
		t.Run(kind, func(t *testing.T) {
			s := openTestStore(t)
			defer s.Close()
			ctx := context.Background()
			if _, err := s.CreateSession(ctx, "A", "A", map[string]any{"model": "mock-stream"}); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 6; i++ {
				if err := s.AddMessage(ctx, store.NowID("msg"), "A", "user", strings.Repeat("history ", 60), nil); err != nil {
					t.Fatal(err)
				}
			}
			e := New(s)
			defer e.Close()
			e.runtimeCfg.Compaction = config.CompactionSettings{Enabled: true, ThresholdTokens: 20, KeepRecentTokens: 20}
			entered := make(chan string, 1)
			if _, err := e.RegisterHook(HookSessionBeforeCompact, "gate", func(ctx context.Context, req HookRequest) (HookResponse, error) {
				entered <- req.TurnID
				if kind == "cancel" {
					<-ctx.Done()
					return HookResponse{}, ctx.Err()
				}
				if kind == "suppress" {
					return HookResponse{Block: true}, nil
				}
				return HookResponse{Payload: map[string]any{"summary": "retained decisions"}}, nil
			}); err != nil {
				t.Fatal(err)
			}
			if kind == "persistence" {
				if _, err := s.DB().Exec(`create trigger reject_compaction before insert on messages when json_extract(new.payload_json,'$.kind')='compaction' begin select raise(abort,'injected summary failure'); end`); err != nil {
					t.Fatal(err)
				}
			}
			var calls atomic.Int32
			withStreamWithToolsStub(t, func(ctx context.Context, _ string, conv *goai.Context, _ func(map[string]any)) (*inference.StreamResult, error) {
				calls.Add(1)
				if kind == "complete" && !strings.Contains(goai.GetTextContent(&conv.Messages[0]), "retained decisions") {
					t.Error("provider did not get compacted context")
				}
				return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "done"}}}}, nil
			})
			res, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "next", Model: "mock-stream"})
			if err != nil {
				t.Fatal(err)
			}
			select {
			case <-entered:
			case <-time.After(3 * time.Second):
				t.Fatal("compaction hook not entered")
			}
			if kind == "cancel" {
				if err := e.CancelTurn(ctx, "A", res.TurnID); err != nil {
					t.Fatal(err)
				}
			}
			wanted := "completed"
			event := "compaction.completed"
			switch kind {
			case "cancel":
				wanted = "cancelled"
				event = "compaction.cancelled"
			case "suppress":
				event = "compaction.suppressed"
			case "persistence":
				wanted = "failed"
				event = "compaction.failed"
			}
			waitForCondition(t, 3*time.Second, func() bool { turn, err := s.GetTurn(ctx, res.TurnID); return err == nil && turn.Status == wanted }, "compaction terminal turn")
			waitForCondition(t, 3*time.Second, func() bool { r := e.runner("A"); r.mu.Lock(); defer r.mu.Unlock(); return r.current == nil }, "runner cleanup")
			events, err := s.ListTurnEvents(ctx, res.TurnID)
			if err != nil {
				t.Fatal(err)
			}
			started, ended := 0, 0
			for _, ev := range events {
				if ev.Type == "compaction.started" {
					started++
				}
				if ev.Type == event {
					ended++
				}
				if kind != "complete" && ev.Type == "compaction.completed" {
					t.Fatal("false completed checkpoint")
				}
			}
			if started != 1 || ended != 1 {
				t.Fatal(started, ended, events)
			}
			msgs, err := s.ListMessages(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			summaries := 0
			for _, m := range msgs {
				if m.Payload["kind"] == "compaction" {
					summaries++
				}
			}
			expected := 1
			if kind != "complete" {
				expected = 0
			}
			if summaries != expected {
				t.Fatal(summaries)
			}
			if (kind == "cancel" || kind == "persistence") && calls.Load() != 0 {
				t.Fatal("provider called after cancellation/persistence failure", calls.Load())
			}
		})
	}
}

func TestProviderEntryRejectsCancelledClaim(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	s.CreateSession(ctx, "A", "A", nil)
	e := New(s)
	defer e.Close()
	s.CreateTurn(ctx, "run", "A", "prompt", nil)
	s.ClaimSessionActiveTurn(ctx, "A", "run", "runner", "run")
	e.RegisterHook(HookBeforeProviderRequest, "cancel", func(context.Context, HookRequest) (HookResponse, error) {
		return HookResponse{}, s.UpdateTurnStatusAndPhase(ctx, "run", "cancelling", "cancelling")
	})
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		t.Fatal("provider invoked after durable cancel")
		return nil, nil
	})
	_, err := e.runner("A").runProviderIteration(ctx, s, "run", "A", "mock-stream", "agent", 1, 5, &goai.Context{})
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	turn, _ := s.GetTurn(ctx, "run")
	if turn.Status != "cancelling" || turn.Phase != "cancelling" {
		t.Fatal(turn)
	}
}
