package turn

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	goai "github.com/rcarmo/go-ai"
)

func TestHookTimeoutDetachesInputAndCloseJoinsHost(t *testing.T) {
	s := openTestStore(t)
	e := New(s)
	e.runtimeCfg.Hooks.TimeoutMS = 10
	release := make(chan struct{})
	var once sync.Once
	defer func() { once.Do(func() { close(release) }); _ = e.Close(); _ = s.Close() }()
	request := HookRequest{Name: HookBeforeProviderRequest, Payload: map[string]any{"nested": map[string]any{"number": 17}}, Messages: []goai.Message{{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "toolCall", Arguments: map[string]any{"command": "original"}}}}}, Tools: []goai.Tool{{Name: "tool", Parameters: json.RawMessage(`{"type":"object"}`)}}}
	result, _, err := e.invokeHookHandler(context.Background(), request, registeredHook{source: "test", handler: func(ctx context.Context, req HookRequest) (HookResponse, error) {
		if _, ok := req.Payload["nested"].(map[string]any)["number"].(int); !ok {
			return HookResponse{}, fmt.Errorf("scalar type changed")
		}
		<-release
		req.Payload["nested"].(map[string]any)["number"] = 99
		req.Messages[0].Content[0].Arguments["command"] = "late mutation"
		req.Tools[0].Parameters[0] = 'X'
		return HookResponse{Cancel: true, Payload: req.Payload}, nil
	}})
	if err == nil || result.Cancel {
		t.Fatal("expired hook kept result authority", result, err)
	}
	closed := make(chan error, 2)
	go func() { closed <- e.Close() }()
	go func() { closed <- e.Close() }()
	select {
	case err := <-closed:
		t.Fatal("close escaped timed-out callback", err)
	case <-time.After(50 * time.Millisecond):
	}
	once.Do(func() { close(release) })
	for i := 0; i < 2; i++ {
		select {
		case err := <-closed:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("callback not joined")
		}
	}
	if request.Payload["nested"].(map[string]any)["number"] != 17 || request.Messages[0].Content[0].Arguments["command"] != "original" || string(request.Tools[0].Parameters) != `{"type":"object"}` {
		t.Fatal("late hook mutated live input", request)
	}
}
func TestHookResponseIsDetachedAndBudgetRespectsLongerParent(t *testing.T) {
	s := openTestStore(t)
	e := New(s)
	defer s.Close()
	defer e.Close()
	e.runtimeCfg.Hooks.TimeoutMS = 100
	parent, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	parentDeadline, _ := parent.Deadline()
	original := HookResponse{Payload: map[string]any{"typed": map[string]string{"key": "original"}, "number": 17}, ToolCall: &goai.ToolCall{Arguments: map[string]any{"command": "original"}}}
	observed := make(chan time.Time, 1)
	response, _, err := e.invokeHookHandler(parent, HookRequest{Name: HookBeforeProviderRequest}, registeredHook{source: "test", handler: func(ctx context.Context, _ HookRequest) (HookResponse, error) {
		deadline, _ := ctx.Deadline()
		observed <- deadline
		return original, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !(<-observed).Before(parentDeadline) {
		t.Fatal("parent deadline disabled configured hook budget")
	}
	original.Payload["typed"].(map[string]string)["key"] = "late"
	original.ToolCall.Arguments["command"] = "late"
	if response.Payload["typed"].(map[string]string)["key"] != "original" || response.Payload["number"] != 17 || response.ToolCall.Arguments["command"] != "original" {
		t.Fatal("response alias escaped", response)
	}
}
func TestHookRejectsCapabilitiesAndCyclesWithoutEnteringHandler(t *testing.T) {
	s := openTestStore(t)
	e := New(s)
	defer s.Close()
	defer e.Close()
	cyclic := map[string]any{}
	cyclic["self"] = cyclic
	for _, payload := range []map[string]any{{"callback": func() {}}, {"channel": make(chan struct{})}, cyclic} {
		called := false
		_, _, err := e.invokeHookHandler(context.Background(), HookRequest{Name: HookToolCall, Payload: payload}, registeredHook{source: "test", handler: func(context.Context, HookRequest) (HookResponse, error) { called = true; return HookResponse{}, nil }})
		if err == nil || called {
			t.Fatal("unrevocable capability admitted", err, called)
		}
	}
}
func TestHookHostJoinsBeforeTurnTerminalAndClaimRelease(t *testing.T) {
	s := openTestStore(t)
	e := New(s)
	ctx := context.Background()
	e.runtimeCfg.Hooks.TimeoutMS = 10
	e.runtimeCfg.Hooks.OnTimeout = "continue"
	release := make(chan struct{})
	var once sync.Once
	defer func() { once.Do(func() { close(release) }); _ = e.Close(); _ = s.Close() }()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := e.RegisterHook(HookBeforeAgentStart, "ignores-cancel", func(_ context.Context, req HookRequest) (HookResponse, error) {
		<-release
		if len(req.Messages) > 0 {
			req.Messages[0].Content[0].Text = "late request mutation"
		}
		return HookResponse{Cancel: true}, nil
	}); err != nil {
		t.Fatal(err)
	}
	original := streamWithToolsWithHooks
	defer func() { streamWithToolsWithHooks = original }()
	observed := make(chan struct{}, 1)
	streamWithToolsWithHooks = func(context.Context, string, *goai.Context, func(map[string]any), *inference.StreamHooks) (*inference.StreamResult, error) {
		observed <- struct{}{}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "owned reply"}}}}, nil
	}
	run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "original prompt", Model: "provider/model"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-observed:
	case <-time.After(time.Second):
		t.Fatal("timeout-continue did not permit inference")
	}
	// The owner has a result, but a timed-out host still owns its detached scope.
	time.Sleep(20 * time.Millisecond)
	turn, err := s.GetTurn(ctx, run.TurnID)
	if err != nil || isTerminalSubTurnStatus(turn.Status) {
		t.Fatal("terminal committed before callback joined", turn, err)
	}
	if active, _, err := s.GetSessionActiveTurn(ctx, "A"); err != nil || active != run.TurnID {
		t.Fatal("claim escaped live callback", active, err)
	}
	once.Do(func() { close(release) })
	waitForCondition(t, time.Second, func() bool { turn, err := s.GetTurn(ctx, run.TurnID); return err == nil && turn.Status == "completed" }, "callback joined before completion")
	messages, err := s.ListMessages(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	for _, message := range messages {
		if message.Role == "user" && message.Content != "original prompt" {
			t.Fatal("late mutation retained authority", message)
		}
	}
}

func TestHookTurnSealFencesFirstLateInvocation(t *testing.T) {
	s := openTestStore(t)
	e := New(s)
	defer s.Close()
	defer e.Close()
	e.joinTurnHooks("finished", true)
	called := false
	_, _, err := e.invokeHookHandler(context.Background(), HookRequest{Name: HookTurnState, TurnID: "finished"}, registeredHook{source: "late", handler: func(context.Context, HookRequest) (HookResponse, error) { called = true; return HookResponse{}, nil }})
	if err == nil || called {
		t.Fatal("late callback bypassed empty sealed scope", err, called)
	}
	e.forgetTurnHooks("finished")
}
