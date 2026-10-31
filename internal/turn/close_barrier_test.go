package turn

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func TestCloseBarrierWaitsForCancellationIgnoringProvider(t *testing.T) {
	for _, concurrent := range []bool{false, true} {
		t.Run(map[bool]string{false: "past-grace", true: "concurrent-close"}[concurrent], func(t *testing.T) {
			s := openTestStore(t)
			e := New(s)
			ctx := context.Background()
			release := make(chan struct{})
			var releaseOnce sync.Once
			defer func() { releaseOnce.Do(func() { close(release) }); _ = e.Close(); _ = s.Close() }()
			if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
				t.Fatal(err)
			}
			entered, cancelled := make(chan struct{}), make(chan struct{})
			original := streamWithToolsWithHooks
			defer func() { streamWithToolsWithHooks = original }()
			streamWithToolsWithHooks = func(requestCtx context.Context, _ string, _ *goai.Context, _ func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
				close(entered)
				<-requestCtx.Done()
				close(cancelled)
				<-release
				message := testProviderPartial()
				message.Content[0].Text = "retained after host join"
				if err := hooks.OnProgress(message); err != nil {
					return nil, err
				}
				return nil, requestCtx.Err()
			}
			run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "join host", Model: "provider/model"})
			if err != nil {
				t.Fatal(err)
			}
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("host not entered")
			}
			closed := make(chan error, 2)
			go func() { closed <- e.Close() }()
			select {
			case <-cancelled:
			case <-time.After(time.Second):
				t.Fatal("owner not signalled")
			}
			delay := shutdownGrace + 100*time.Millisecond
			callers := 1
			if concurrent {
				callers = 2
				go func() { closed <- e.Close() }()
				delay = 100 * time.Millisecond
			}
			select {
			case err := <-closed:
				t.Fatal("close released an owned host", err)
			case <-time.After(delay):
			}
			if e.backgroundContext().Err() != nil {
				t.Fatal("settlement context revoked before host join")
			}
			releaseOnce.Do(func() { close(release) })
			for i := 0; i < callers; i++ {
				select {
				case err := <-closed:
					if err != nil {
						t.Fatal(err)
					}
				case <-time.After(2 * time.Second):
					t.Fatal("close did not join final settlement")
				}
			}
			turn, err := s.GetTurn(ctx, run.TurnID)
			if err != nil || turn.Status != "cancelled" {
				t.Fatal(turn, err)
			}
			messages, err := s.ListMessages(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			retained := false
			for _, message := range messages {
				if message.Role == "assistant" && message.Content == "retained after host join" && message.Payload["adopted"] == false {
					retained = true
				}
			}
			if !retained {
				t.Fatal("close revoked durable writer before host settled", messages)
			}
		})
	}
}

func TestCloseBarrierWaitsForCancellationIgnoringTool(t *testing.T) {
	s := openTestStore(t)
	e := New(s)
	ctx := context.Background()
	release := make(chan struct{})
	var once sync.Once
	defer func() { once.Do(func() { close(release) }); _ = e.Close(); _ = s.Close() }()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	entered, cancelled := make(chan struct{}), make(chan struct{})
	if err := e.RegisterTool(tools.RegisteredTool{Name: "ignores-cancel", Executor: func(toolCtx context.Context, rt tools.ToolRuntime, _ goai.ToolCall) (string, error) {
		close(entered)
		<-toolCtx.Done()
		close(cancelled)
		<-release
		rt.AddUsage(goai.Usage{Input: 17})
		return "owned tool stopped", toolCtx.Err()
	}}); err != nil {
		t.Fatal(err)
	}
	original := streamWithToolsWithHooks
	defer func() { streamWithToolsWithHooks = original }()
	streamWithToolsWithHooks = func(context.Context, string, *goai.Context, func(map[string]any), *inference.StreamHooks) (*inference.StreamResult, error) {
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonToolUse, Content: []goai.ContentBlock{{Type: "toolCall", ID: "tool", Name: "ignores-cancel", Arguments: map[string]any{}}}}}, nil
	}
	run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "join tool", Model: "provider/model"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("tool not entered")
	}
	closed := make(chan error, 2)
	go func() { closed <- e.Close() }()
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("tool not signalled")
	}
	go func() { closed <- e.Close() }()
	select {
	case err := <-closed:
		t.Fatal("close escaped owned tool", err)
	case <-time.After(100 * time.Millisecond):
	}
	once.Do(func() { close(release) })
	for i := 0; i < 2; i++ {
		select {
		case err := <-closed:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("tool settlement not joined")
		}
	}
	total, err := s.TurnAccounting(ctx, run.TurnID)
	if err != nil || total.Usage.Input != 17 {
		t.Fatal("tool receipt lost at shutdown", total, err)
	}
}
